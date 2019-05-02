package manticore

import "fmt"

func ExampleEscapeString() {

	escaped := EscapeString("escaping-sample@query/string")
	fmt.Println(escaped)
	// Output:
	// escaping\-sample\@query\/string
}
