package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Query_default(t *testing.T) {
	cl := NewClient()
	foo, err := cl.Query("query")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_Query_index(t *testing.T) {
	cl := NewClient()
	foo, err := cl.Query("query", "lj")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_Query_unixsocket(t *testing.T) {
	cl := NewClient()

	cl.SetServer("/work/lj/sphinxapi")
	q := NewSearch("luther", "lj", "")
	q.SetSortMode(SortAttrAsc, "published")
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestClient_RunPreparedQueries(t *testing.T) {
	cl := NewClient()

	queries := []Search{
		NewSearch("luther", "lj", ""),
		NewSearch("martin luther", "lj", ""),
	}
	foo, err := cl.RunQueries(queries)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestSearch_SetFieldWeights(t *testing.T) {
	cl := NewClient()

	q := NewSearch("luther", "lj", "")
	q.FieldWeights = map[string]int32{"title": 1000, "content": 10}
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestSearch_AddFilter(t *testing.T) {

	q := NewSearch("query", "lj", "")
	q.AddFilter("channel_id", []int64{537345, 536802, 538617}, false)

	cl := NewClient()
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestSearch_AddFilter_exclude(t *testing.T) {

	q := NewSearch("query", "lj", "")
	q.AddFilter("channel_id", []int64{537345, 536802, 538617}, true)

	cl := NewClient()
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestSearch_AddFilterFloatRange(t *testing.T) {

	q := NewSearch("query", "lj", "")
	q.AddFilterFloatRange("channel_id", 10000.0, 200000.0, false)

	cl := NewClient()
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

func TestSearch_AddFilterUservar(t *testing.T) {

	var q Search
	q.AddFilterUservar("foo", "bar", false)
	if len(q.filters) != 1 {
		t.Errorf("Wrong len of filters: %d", len(q.filters))
	}
}

func TestSearch_AddFilterExpression(t *testing.T) {
	q := NewSearch("query", "lj", "")
	q.SelectClause = "channel_id*10 as cchh, channel_id"
	q.AddFilterExpression("channel_id*10<1000", false)

	cl := NewClient()
	_, erro := cl.Open()
	if erro!=nil {
		fmt.Println(erro.Error())
	}

	foo, err := cl.RunQuery(q)
	bar, err1 := cl.Status(false)
	_,_ = cl.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}

	if err1 != nil {
		fmt.Println(err1.Error())
	} else {
		for key, line := range (bar) {
			fmt.Printf("%v:\t%v\n", key, line)
		}
	}
}

func TestSearch_SetSortMode(t *testing.T) {
	q := NewSearch("query", "lj", "")
	q.SelectClause = "channel_id*10 as cchh, channel_id*5 as cchhh"
	q.SetSortMode(SortExtended, "cchh DESC, cchhh DESC")

	cl := NewClient()
	foo, err := cl.RunQuery(q)

	if err != nil {
		fmt.Println(err.Error())
		if foo!=nil {
			fmt.Println(foo.Error)
		}
	} else {
		fmt.Println(foo)
	}
}

//  q := NewSearch("some common query terms", "index", "")
//	q.SelectClause = "id, slow_rank() as slow, fast_rank as fast"
//  q.SetSortMode( SortExpr, "fast DESC, slow DESC"
//	q.AddFilterExpression("channel_id*10<1000", false)


func ExampleQflags() {
	fl := QflagJsonQuery
	fmt.Println(fl)
	// Output:
	// 2048
}

func TestSearch_SetOuterSelect(t *testing.T) {
	q := NewSearch("query", "lj", "")
	q.SelectClause = "channel_id*10 as cchh, channel_id"
	q.SetOuterSelect("cchh asc", 0, 3)

	cl := NewClient()
	foo, err := cl.RunQuery(q)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(foo)
	}
}

