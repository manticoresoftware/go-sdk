package manticore

func buildSphinxqlRequest(cmd string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(cmd)
	}
}

func parseSphinxqlAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		resp := []byte(*answer)
		return resp
	}
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