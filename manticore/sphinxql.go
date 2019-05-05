package manticore

import (
	"encoding/binary"
	"fmt"
)

func buildSphinxqlRequest(cmd string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(cmd)
	}
}

func parseSphinxqlAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		var rss []sqlresult
		for true {
			var rs sqlresult
			if rs.parseChain (answer) {
				rss = append (rss, rs)
			} else {
				break
			}
		}
		return rss
	}
}

func (buf *apibuf) getnextSqlchunk() (bool, apibuf) {
	for true {
		if len(*buf) == 0 {
			break
		}
		_, lng := buf.getMysqlPacketHead()
//		fmt.Printf("packet %d, len %d\n", packetID, lng)
		res := (*buf)[:lng]
		*buf = (*buf)[lng:]
		if len(res)>0 {
			return true, res
		}
	}
	return false,nil
}

func (rs *sqlresult) parseChain (source *apibuf) bool {

	have, buf := source.getnextSqlchunk()
	if !have {
		return false
	}
	firstbyte := buf[0]
	switch firstbyte {
	case byte(packetOk):
		_ = buf.getByte()
		rs.parseOK(&buf)
		return true
	case byte(packetError):
		_ = buf.getByte()
		rs.parseError(&buf)
		return true
	}
	if buf.isEOF() {
		rs.warnings, _ = buf.parseEOF()
		return true
	}
	rs.schema.ncolumns = buf.getMysqlInt()
	fmt.Printf("Resultset of %d columns\n", rs.schema.ncolumns)
	rs.schema.parseschema (source)
	have, buf = source.getnextSqlchunk()
	rs.warnings, _ = buf.parseEOF()

	for true {
		if !rs.schema.parserow(source) {
			break
		}
	}
	return true
}

type packetType byte

const (
	packetOk packetType = iota
	packetField
	packetResultset
	packetEOF packetType = 0xFE
	packetError packetType = 0xFF
	)

func (t packetType) String() string {
	switch t {
	case packetOk : return "OK"
	case packetEOF : return "EOF"
	case packetError : return "ERROR"
	case packetField : return "FIELD"
	}
	return fmt.Sprintf("UNKNWN(%d)", t)
}

type fieldType byte

const (
	colDecimal fieldType = 0
	colLong fieldType = 3
	colFloat fieldType = 4
	colLonglong fieldType = 8
	colString fieldType = 254
)

type sqlfield struct {
	name string
	length uint32
	tp fieldType
	unsigned bool
}

type sqlresultset struct {
	ncolumns int
	columns []sqlfield
	rows [][]interface{}
}

type sqlresult struct {
	msg string
	warnings uint16
	errorCode uint16
	rowsAffected int
	schema sqlresultset
}

func (buf apibuf) parseEOF () (uint16, bool) {
	_ = buf.getByte()
	warnings := buf.getLsbWord()
	status := buf.getLsbWord()
	return warnings, status&8!=0
}

func (buf apibuf) isEOF () bool {
	return buf[0]==byte(packetEOF) && len(buf)>=5
}



func (rs *sqlresultset) parseschema (source *apibuf) {

	rs.columns = make ([]sqlfield, rs.ncolumns)
	for i:=0; i< rs.ncolumns; i++ {
		valid, buf := source.getnextSqlchunk()
		if !valid {
			return
		}
		_ = buf.getMysqlStrLen() // "def"
		_ = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		rs.columns[i].name = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		_ = buf.getByte() // = 12, filter
		_ = buf.getWord() // = 0x0021, utf8
		rs.columns[i].length = buf.getLsbDword()
		rs.columns[i].tp = fieldType(buf.getByte())
		rs.columns[i].unsigned = buf.getWord()!=0
	}
}

func (rs *sqlresultset) parserow (source *apibuf) bool {

	row := make ([]interface{}, rs.ncolumns)
	valid, buf := source.getnextSqlchunk()
	if buf.isEOF() || !valid {
		return false
	}
	for i:=0; i<rs.ncolumns; i++ {
		if buf[0] == 0xFB {
			row[i] = nil
			_ = buf.getByte()
		} else {
			row[i] = buf.getMysqlStrLen()
		}
	}
	rs.rows = append(rs.rows,row)
	return true
}



func (rs *sqlresult) parseOK(buf *apibuf)  {
	rs.rowsAffected = buf.getMysqlInt()
	_ = buf.getMysqlInt() // last_insert_id
	_ = buf.getLsbWord() // status
	rs.warnings = buf.getLsbWord()
	rs.msg = buf.getMysqlStrEof()
	fmt.Printf("OK rows %d, warnings %d, msg %s\n", rs.rowsAffected, rs.warnings, rs.msg)
}

func (rs *sqlresult) parseError(buf *apibuf) {
	rs.errorCode = buf.getLsbWord()
	rs.msg = buf.getMysqlStrEof()
	fmt.Printf("ERROR code %d, msg %s\n", rs.errorCode, rs.msg)
}

func (rs *sqlresult) parseEOF(buf *apibuf) {
	fmt.Println("EOF")
}




func (buf *apibuf) getMysqlInt() int {

	res := int(buf.getByte())
	if res < 251 {
		return res
	}

	if res==252 {
		res = int((*buf)[0]) | int((*buf)[1])<<8
		*buf = (*buf)[2:]
	}

	if res==253 {
		res = int((*buf)[0]) | int((*buf)[1])<<8 | int((*buf)[2])<<16
		*buf = (*buf)[3:]
	}

	if res==254 {
		res = int((*buf)[0]) | int((*buf)[1])<<8 | int((*buf)[2])<<16 | int((*buf)[3])<<24
		*buf = (*buf)[8:]
	}
	return res
}

func (buf *apibuf) getMysqlStrEof() string {
	result := string(*buf)
	*buf = (*buf)[:]
	return result
}

func (buf *apibuf) getMysqlStrLen() string {
	lng := buf.getMysqlInt()
	result := string((*buf)[:lng])
	*buf = (*buf)[lng:]
	return result
}

func (buf *apibuf) getMysqlPacketHead() (byte, uint32) {

	packlen := uint32((*buf)[0]) | uint32((*buf)[1])<<8 | uint32((*buf)[2])<<16
	id := (*buf)[3]
	*buf = (*buf)[4:]
	return id, packlen
}

func (buf *apibuf) getLsbWord() uint16 {
	val := binary.LittleEndian.Uint16(*buf)
	*buf = (*buf)[2:]
	return val
}

func (buf *apibuf) getLsbDword() uint32 {
	val := binary.LittleEndian.Uint32(*buf)
	*buf = (*buf)[4:]
	return val
}