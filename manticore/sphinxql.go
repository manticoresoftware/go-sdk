package manticore

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

func buildSphinxqlRequest(cmd string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(cmd)
	}
}

func parseSphinxqlAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		var rss []Sqlresult
		for true {
			var rs Sqlresult
			if rs.parseChain(answer) {
				rss = append(rss, rs)
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
		if len(res) > 0 {
			return true, res
		}
	}
	return false, nil
}

func (rs *Sqlresult) parseChain(source *apibuf) bool {

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
		rs.Warnings, _ = buf.parseEOF()
		return true
	}
	ncolumns := buf.getMysqlInt()
	//	fmt.Printf("Resultset of %d columns\n", ncolumns)
	rs.Schema.parseschema(source, ncolumns)
	have, buf = source.getnextSqlchunk()
	rs.Warnings, _ = buf.parseEOF()

	for true {
		if !rs.parserow(source) {
			break
		}
	}
	return true
}

type packetType byte

const (
	packetOk packetType = iota
	packetField
	packetEOF   packetType = 0xFE
	packetError packetType = 0xFF
)

func (t packetType) String() string {
	switch t {
	case packetOk:
		return "OK"
	case packetEOF:
		return "EOF"
	case packetError:
		return "ERROR"
	case packetField:
		return "FIELD"
	}
	return fmt.Sprintf("UNKNWN(%d)", t)
}

type fieldType byte

const (
	colDecimal  fieldType = 0
	colLong     fieldType = 3
	colFloat    fieldType = 4
	colLonglong fieldType = 8
	colString   fieldType = 254
)

func (t fieldType) String() string {
	switch t {
	case colDecimal:
		return "decimal"
	case colLong:
		return "long"
	case colFloat:
		return "float"
	case colLonglong:
		return "longlong"
	case colString:
		return "string"
	}
	return fmt.Sprintf("unknwn(%d)", t)
}

type sqlfield struct {
	Name     string
	Length   uint32
	Tp       fieldType
	Unsigned bool
}

type SqlSchema []sqlfield

type SqlResultset [][]interface{}

type SqlMsg string

func (r SqlMsg) String() string {
	if r[0] == '#' {
		code := r[1:6]
		return fmt.Sprintf("(%v): %v", code, r[6:])
	}
	return string(r)
}

type Sqlresult struct {
	Msg          SqlMsg
	Warnings     uint16
	ErrorCode    uint16
	RowsAffected int
	Schema       SqlSchema
	Rows         SqlResultset
}

func (r Sqlresult) String() string {
	if r.ErrorCode != 0 {
		return fmt.Sprintf("ERROR %d %v", r.ErrorCode, r.Msg)
	}

	if r.Schema != nil {
		line := ""
		for _, col := range r.Schema {
			line += fmt.Sprintf("%v\t", col.Name)
		}
		line = line[:len(line)-1] + "\n"
		for _, row := range r.Rows {
			for i := 0; i < len(r.Schema); i++ {
				line += fmt.Sprintf("%v\t", row[i])
			}
			line = line[:len(line)-1] + "\n"
		}
		line += fmt.Sprintf("%v rows in set", len(r.Rows))
		if r.Warnings != 0 {
			line += fmt.Sprintf(", %d warnings", r.Warnings)
		}
		line += "\n"
		return line
	}

	return fmt.Sprintf("Query OK, %v rows affected", r.RowsAffected)
}

func (buf apibuf) parseEOF() (uint16, bool) {
	_ = buf.getByte()
	warnings := buf.getLsbWord()
	status := buf.getLsbWord()
	return warnings, status&8 != 0
}

func (buf apibuf) isEOF() bool {
	return buf[0] == byte(packetEOF) && len(buf) >= 5
}

func (rs *SqlSchema) parseschema(source *apibuf, ncolumns int) {

	*rs = make([]sqlfield, ncolumns)
	for i := 0; i < ncolumns; i++ {
		valid, buf := source.getnextSqlchunk()
		if !valid {
			return
		}
		_ = buf.getMysqlStrLen() // "def"
		_ = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		(*rs)[i].Name = buf.getMysqlStrLen()
		_ = buf.getMysqlStrLen()
		_ = buf.getByte() // = 12, filter
		_ = buf.getWord() // = 0x0021, utf8
		(*rs)[i].Length = buf.getLsbDword()
		(*rs)[i].Tp = fieldType(buf.getByte())
		(*rs)[i].Unsigned = buf.getWord() != 0
	}
}

func (rs *Sqlresult) parserow(source *apibuf) bool {

	ncolumns := len(rs.Schema)
	row := make([]interface{}, ncolumns)
	valid, buf := source.getnextSqlchunk()
	if buf.isEOF() || !valid {
		return false
	}
	for i := 0; i < ncolumns; i++ {
		if buf[0] == 0xFB {
			row[i] = nil
			_ = buf.getByte()
		} else {
			strValue := buf.getMysqlStrLen()
			switch rs.Schema[i].Tp {
			case colDecimal, colLong:
				if rs.Schema[i].Unsigned {
					ui, _ := strconv.ParseUint(strValue, 10, 32)
					row[i] = uint32(ui)
				} else {
					ii, _ := strconv.ParseInt(strValue, 10, 32)
					row[i] = int32(ii)
				}
			case colFloat:
				fl, _ := strconv.ParseFloat(strValue, 32)
				row[i] = float32(fl)
			case colLonglong:
				if rs.Schema[i].Unsigned {
					ui, _ := strconv.ParseUint(strValue, 10, 64)
					row[i] = uint64(ui)
				} else {
					ii, _ := strconv.ParseInt(strValue, 10, 64)
					row[i] = int64(ii)
				}
			default:
				row[i] = strValue
			}
		}
	}
	rs.Rows = append(rs.Rows, row)
	return true
}

func (rs *Sqlresult) parseOK(buf *apibuf) {
	rs.RowsAffected = buf.getMysqlInt()
	_ = buf.getMysqlInt() // last_insert_id
	_ = buf.getLsbWord()  // status
	rs.Warnings = buf.getLsbWord()
	rs.Msg = SqlMsg(buf.getMysqlStrEof())
	//	fmt.Printf("OK rows %d, Warnings %d, Msg %s\n", rs.RowsAffected, rs.Warnings, rs.Msg)
}

func (rs *Sqlresult) parseError(buf *apibuf) {
	rs.ErrorCode = buf.getLsbWord()
	rs.Msg = SqlMsg(buf.getMysqlStrEof())
	//	fmt.Printf("ERROR code %d, Msg %s\n", rs.ErrorCode, rs.Msg)
}

func (buf *apibuf) getMysqlInt() int {

	res := int(buf.getByte())
	if res < 251 {
		return res
	}

	if res == 252 {
		res = int((*buf)[0]) | int((*buf)[1])<<8
		*buf = (*buf)[2:]
	}

	if res == 253 {
		res = int((*buf)[0]) | int((*buf)[1])<<8 | int((*buf)[2])<<16
		*buf = (*buf)[3:]
	}

	if res == 254 {
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
