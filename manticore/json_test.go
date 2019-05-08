package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Json_search(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Json("search", "index=lj&match=luther&select=id,channel_id&limit=20")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_Json_sqlapi(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Json("sql", "query=select * from lj where match ('luther')")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_Json_json_search(t *testing.T) {

	cl := NewClient()

	foo, err := cl.Json("json/search", `{"index":"lj","query":{"match":{"title":"luther"}}}`)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}
