package manticore

import (
	"fmt"
	"testing"
)

func TestClient_Uvar(t *testing.T) {

	cl := NewClient()
	_, _ = cl.Open()

	// values are 1) not sorted; 2) have dupe.
	err := cl.Uvar("@foo", []uint64{7811237, 7811235, 7811235, 7811233, 7811236})
	if err != nil {
		fmt.Println(err.Error())
	}

	q := NewSearch("", "lj", "")
	q.AddFilterUservar("id", "@foo", false)

	// expect to receive 4 rows
	foo, err := cl.RunQuery(q)

	if err != nil {
		fmt.Println(err.Error())
		if foo != nil {
			fmt.Println(foo.Error)
		}
	} else {
		fmt.Println(foo)
	}
}
