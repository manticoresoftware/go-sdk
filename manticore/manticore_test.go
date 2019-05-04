package manticore

import (
	"fmt"
	"testing"
)

func ExampleEscapeString() {

	escaped := EscapeString("escaping-sample@query/string")
	fmt.Println(escaped)
	// Output:
	// escaping\-sample\@query\/string
}


func TestClient_Ping(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Ping (123456789 )
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

