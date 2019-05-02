package manticore

func parseStatusAnswer() func(*apibuf) interface{} {
	return func(answer *apibuf) interface{} {
		nrows := answer.getInt()
		status := make(map[string]string)
		_ = answer.getInt() // n of cols, always 2
		for j := 0; j < nrows; j++ {
			key := answer.getString()
			status[key] = answer.getString()
		}
		return status
	}
}

