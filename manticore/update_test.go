package manticore

import (
	"fmt"
	"testing"
)

func TestClient_UpdateAttributes(t *testing.T) {

	cl := NewClient()

	upd, err := cl.UpdateAttributes("lj", []string{"channel_id"}, map[DocID][]interface{}{5000000:{1}, 5000011:{11}}, UpdateInt, false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(upd)
	}
}

func TestClient_UpdateAttributes_many(t *testing.T) {

	cl := NewClient()

	upd, err := cl.UpdateAttributes("lj", []string{"channel_id","published"}, map[DocID][]interface{}{5000000:{1,2}, 5000011:{3,4}}, UpdateInt, false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(upd)
	}
}