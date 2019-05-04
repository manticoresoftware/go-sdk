package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Sphinxql(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Sphinxql ("show meta" )
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}