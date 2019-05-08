package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Sphinxql_selectmeta(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Sphinxql("select channel_id+1.3 x, * from lj; show meta")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_Sphinxql_status(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Sphinxql("show status")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}
