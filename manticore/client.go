package manticore

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Client represents connection to manticore daemon. It provides set of public API functions
type Client struct {
	host, dialmethod     string
	port                 uint16
	conn                 net.Conn
	connected, connError bool
	lastWarning          string
	buf                  apibuf
	timeout              time.Duration
	maxAlloc             int
}

// NewClient creates default connector, which points to 'localhost:9312', has zero timeout and 8M maxalloc.
// Defaults may be changed later by invoking `SetServer()`, `SetMaxAlloc()`
func NewClient() Client {
	return Client{
		"localhost", "tcp",
		SphinxPort,
		nil,
		false, false,
		"",
		nil,
		0,
		8 * 1024 * 1024,
	}
}

// getByteBuf provides byte buffer. One and same buffer will be reused on each call, so that no GC will
// be in use/
func (cl *Client) getByteBuf(size int) *apibuf {
	cl.buf.resizeBuf(size, cl.maxAlloc)
	return &cl.buf
}

func (cl *Client) getOutBuf() *apibuf {
	return cl.getByteBuf(0)
}

// gracefully return error
func (cl *Client) confail(err error) error {
	cl.connError = err != nil
	return err
}

func (cl *Client) failclose(err error) error {
	if err != nil {
		_ = cl.conn.Close()
		cl.conn = nil
		cl.connected = false
	}
	return err
}

func (client *Client) eof() bool {

	if !client.connected {
		return true
	}
	_ = client.conn.SetReadDeadline(time.Now())
	var one []byte
	if _, err := client.conn.Read(one); err == io.EOF {
		client.connected = false
		return true
	}
	_ = client.conn.SetReadDeadline(time.Time{})
	return false
}

/// connect to searchd server
func (cl *Client) connect() error {

	// we are in persistent connection mode, so we have a socket
	// however, need to check whether it's still alive
	if cl.connected {
		if cl.eof() { // connection timed out
			_ = cl.conn.Close()
			cl.conn = nil
		} else { // connection alive; no more actions need
			return cl.confail(nil)
		}
	}

	address := cl.host
	if cl.dialmethod == "tcp" {
		sPort := fmt.Sprintf("%d", cl.port)
		address = net.JoinHostPort(address, sPort)
	}

	var err error
	// connect
	if cl.timeout != 0 {
		cl.conn, err = net.DialTimeout(cl.dialmethod, address, cl.timeout)
	} else {
		cl.conn, err = net.Dial(cl.dialmethod, address)
	}

	if err != nil {
		return cl.confail(err)
	}

	cl.connected = true

	// send my version
	// this is a subtle part. we must do it before (!) reading back from searchd.
	// because otherwise under some conditions (reported on FreeBSD for instance)
	// TCP stack could throttle write-write-read pattern because of Nagle.
	// send handshake, retrieve answer and check it
	handshake := apibuf(make([]byte, 0, 4))
	handshake.putUint(cphinxClientVersion)

	_, err = cl.conn.Write(handshake)
	if err == nil {
		buf := cl.getByteBuf(4)
		_, err = cl.conn.Read(*buf)
		if err == nil {
			ver := buf.getDword()
			if ver == cphinxSearchdProto {
				return cl.confail(nil)
			}
			err = errors.New(fmt.Sprintf("Wrong version num received: %d", ver))
		}
	}

	// error happened, return it
	return cl.failclose(err)
}

/// read raw answer (with fixed size) to buf
func (cl *Client) readRawAnswer(buf []byte, size int) (int, error) {
	const MAX_CHUNK_SIZE = 16 * 1024
	nbytes := 0
	for {
		chunkSize := MAX_CHUNK_SIZE
		bytesRemaining := size - nbytes
		if bytesRemaining < chunkSize {
			chunkSize = bytesRemaining
		}
		n, e := cl.conn.Read(buf[nbytes:nbytes+chunkSize])
		if e != nil {
			return n, e
		}
		nbytes += n
		if (nbytes < size) {
			continue
		}
		break
	}

	if (nbytes > size) {
		return nbytes, errors.New("Logical error in Client.read()!")
	}

	return nbytes, nil
}

/// get and check response packet from searchd server
func (cl *Client) getResponse(client_ver uCommandVersion) (apibuf, error) {
	rawrecv := cl.getByteBuf(8)
	nbytes, err := cl.conn.Read(*rawrecv)

	if err != nil {
		return nil, err
	} else if nbytes == 0 {
		return nil, errors.New("received zero-sized searchd response")
	}

	uStat := ESearchdstatus(rawrecv.getWord())
	uVer := uCommandVersion(rawrecv.getWord())
	iReplySize := rawrecv.getInt()

	rawanswer := cl.getByteBuf(iReplySize)
	nbytes, err = cl.readRawAnswer(*rawanswer, iReplySize)
	if err != nil {
		return nil, err
	}

	if nbytes != iReplySize {
		if nbytes == 0 {
			return nil, errors.New("received zero-sized searchd response")
		}
		return nil, errors.New(
			fmt.Sprintf("failed to read searchd response (status=%d, ver=%d, len=%d, read=%d)",
				uStat, uVer, iReplySize, nbytes))
	}

	switch uStat {
	case StatusError:
		return *rawanswer, errors.New(fmt.Sprintf("searchd error: %s", rawanswer.getString()))
	case StatusRetry:
		return *rawanswer, errors.New(fmt.Sprintf("temporary searchd error: %s", rawanswer.getString()))
	case StatusWarning:
		cl.lastWarning = rawanswer.getString()
	case StatusOk:
		break
	default:
		return *rawanswer, errors.New(fmt.Sprintf("unknown status code '%d'", uStat))
	}

	// check version
	if uVer < client_ver {
		cl.lastWarning = fmt.Sprintf("searchd command v.%v older than cl's v.%v, some options might not work",
			uVer, client_ver)
	}
	return *rawanswer, nil
}

func (cl *Client) netQuery(command eSearchdcommand, builder func(*apibuf), parser func(*apibuf) interface{}) (interface{}, error) {

	// connect (if necessary)
	err := cl.connect()
	if err != nil {
		return nil, err
	}

	// build packet
	buf := cl.getOutBuf()
	tPos := buf.apiCommand(command)
	if builder != nil {
		builder(buf)
	}
	buf.finishAPIPacket(tPos)

	// send query
	_, err = cl.conn.Write(cl.buf)
	if err != nil {
		return nil, err
	}

	if parser == nil {
		return nil, nil
	}

	// get response
	var answer apibuf
	answer, err = cl.getResponse(searchdcommandv[command])

	if err != nil {
		return nil, err
	}

	// parse response
	if answer != nil {
		return parser(&answer), nil
	}
	return nil, nil
}

// common case when payload has only one value, which is boolean as DWORD
func buildBoolRequest(val bool) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putBoolDword(val)
	}
}

// common case when payload has only one value, which is DWORD
func buildDwordRequest(val uint32) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putDword(val)
	}
}

// common case when answer contains the only integer
func parseDwordAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		res := answer.getDword()
		return res
	}
}
