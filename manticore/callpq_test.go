package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Sphinxql_tbls(t *testing.T) {

	cl := NewClient()
	cl.SetServer("", 6712)

	foo, err := cl.Sphinxql("select * from pq; select * from pq1")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

//CALL PQ ('META:multi', ('[{"title":"angry test", "gid":3 },
// {"title":"filter test doc2", "gid":13}]'),
// 1 as docs, 1 as verbose, 1 as docs_json, 1 as query, 'gid' as docs_id)

func TestClient_CallPQ(t *testing.T) {
	cl := NewClient()
	cl.SetServer("", 6712)

	pq := NewSearchPqOptions()
	pq.Flags = NeedDocs | Verbose | NeedQuery

	resp, err := cl.CallPQ("pq", []string{"angry test", "filter test doc2"}, pq)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(resp)
	}
}

func TestClient_Sphinxql_callpq(t *testing.T) {

	cl := NewClient()
	cl.SetServer("", 6712)

	_, _ = cl.Open()
	foo, err := cl.Sphinxql(`call pq ('pq', ('angry test','filter test doc2'), 1 as docs, 1 as verbose, 1 as query, 0 as docs_json)`)
	foo1, err := cl.Sphinxql(`show meta`)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
		fmt.Println(foo1)
	}
}
