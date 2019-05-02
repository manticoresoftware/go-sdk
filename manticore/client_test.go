package manticore

import (
	"fmt"
	"testing"
)

func TestSetServer(t *testing.T) {

	cl := NewClient()

	// /var/log -> unixsocket on /var/log
	cl.SetServer("/var/log")

	if cl.dialmethod != "unix" {
		t.Errorf("method  is not as expected unix, got %s", cl.dialmethod)
	}

	if cl.host != "/var/log" {
		t.Errorf("host is not as expected, got %s", cl.host)
	}

	// unix:///bla -> unixsocket /bla
	cl.SetServer("unix:///tmp.sock")

	if cl.dialmethod != "unix" {
		t.Errorf("method  is not as expected unix, got %s", cl.dialmethod)
	}

	if cl.host != "/tmp.sock" {
		t.Errorf("host is not as expected, got %s", cl.host)
	}

	// path starting from not / - tcp connect to sock, port
	cl.SetServer("google.com")

	if cl.dialmethod != "tcp" {
		t.Errorf("path is not as expected tcp, got %s", cl.dialmethod)
	}

	if cl.host != "google.com" {
		t.Errorf("host is not as expected, got %s", cl.host)
	}

	if cl.port != 0 {
		t.Errorf("port is not as expected, got %d", cl.port)
	}

	cl.SetServer("google.com", 9999)

	if cl.dialmethod != "tcp" {
		t.Errorf("path is not as expected tcp, got %s", cl.dialmethod)
	}

	if cl.host != "google.com" {
		t.Errorf("host is not as expected, got %s", cl.host)
	}

	if cl.port != 9999 {
		t.Errorf("port is not as expected, got %d", cl.port)
	}
}

func ExampleClient_SetServer_unixsocket() {
	cl := NewClient()
	cl.SetServer("/var/log")

	fmt.Println(cl.dialmethod)
	fmt.Println(cl.host)
	// Output:
	// unix
	// /var/log
}

func ExampleClient_SetServer_tcpsocket() {
	cl := NewClient()
	cl.SetServer("google.com", 9999)

	fmt.Println(cl.dialmethod)
	fmt.Println(cl.host)
	fmt.Println(cl.port)
	// Output:
	// tcp
	// google.com
	// 9999
}


