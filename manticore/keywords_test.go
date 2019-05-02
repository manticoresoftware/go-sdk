package manticore

import (
	"fmt"
	"testing"
	"time"
)

func TestClient_BuildKeywords(t *testing.T) {
	cl := NewClient()

	cl.SetServer("localhost")
	cl.SetConnectTimeout(1 * time.Second)

	kwds, err := cl.BuildKeywords("martin luthers king", "lj", false)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(kwds)
	}
}

func ExampleClient_BuildKeywords_withoutHits() {
	cl := NewClient()

	keywords, err := cl.BuildKeywords("this.is.my query", "lj", false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(keywords)
	}
	// Output:
	// [{Tok: 'this',	Norm: 'this',	Qpos: 1; docs/hits 0/0}
	//  {Tok: 'is',	Norm: 'is',	Qpos: 2; docs/hits 0/0}
	//  {Tok: 'my',	Norm: 'my',	Qpos: 3; docs/hits 0/0}
	//  {Tok: 'query',	Norm: 'query',	Qpos: 4; docs/hits 0/0}
	// ]
}

func ExampleClient_BuildKeywords_withHits() {
	cl := NewClient()

	keywords, err := cl.BuildKeywords("this.is.my query", "lj", true)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(keywords)
	}
	// Output:
	// [{Tok: 'this',	Norm: 'this',	Qpos: 1; docs/hits 1629922/3905279}
	//  {Tok: 'is',	Norm: 'is',	Qpos: 2; docs/hits 1901345/6052344}
	//  {Tok: 'my',	Norm: 'my',	Qpos: 3; docs/hits 1981048/7549917}
	//  {Tok: 'query',	Norm: 'query',	Qpos: 4; docs/hits 1235/1474}
	// ]
}
