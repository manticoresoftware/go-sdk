package manticore

func buildJsonRequest(endpoint, request string) func(*apibuf) {
	return func(buf *apibuf) {
		buf.putString(endpoint)
		buf.putString(request)
	}
}

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
