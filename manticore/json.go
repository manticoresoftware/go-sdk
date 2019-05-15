package manticore

func buildJsonRequest(endpoint, request string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(endpoint)
		buf.putString(request)
	}
}

/*
JsonAnswer encapsulates answer to Json command.

`Endpoint` - endpoint to which request was directed

`Answer` - string, containing the answer. In opposite to true HTTP connection, here only string mesages given,
no numeric error codes.

*/
type JsonAnswer struct {
	Endpoint string
	Answer   string
}

func parseJsonAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		endpoint := answer.getString()
		blob := answer.getString()
		return JsonAnswer{endpoint, blob}
	}
}
