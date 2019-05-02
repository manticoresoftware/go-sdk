package manticore

import (
	"fmt"
	"testing"
)

//status, err := cl.Status ()
//  foreach ( $status as $row )
//    print join ( ": ", $row ) . "\n";


func TestClient_Status_global(t *testing.T) {

	cl := NewClient()
	foo, err := cl.Status(false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for key, line := range (foo) {
			fmt.Printf("%v:\t%v\n", key, line)
		}
	}
}