package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Sphinxql(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Sphinxql ("select * from lj; select * from lj" )
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}