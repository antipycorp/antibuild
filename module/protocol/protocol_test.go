// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package protocol

import (
	"fmt"
	"io"
	"sync"
	"testing"
)

var (
	host   *Connection
	client *Connection
)

func TestProtocol(t *testing.T) {

	t.Run("Open", testOpen)
	t.Run("Init", testInit)
	t.Run("round", testRoundTrip)
}

func testOpen(t *testing.T) {
	hin, cout := io.Pipe()
	cin, hout := io.Pipe()

	host = OpenConnection(hin, hout)
	client = OpenConnection(cin, cout)
}

func testInit(t *testing.T) {
	wait := sync.WaitGroup{}
	go func() {
		wait.Add(1)
		i, err := host.Init(true)
		if err != nil {
			t.Error(err)
		}
		if i != int(version) {
			t.Error(i)
		}
		wait.Done()
	}()
	i, err := client.Init(false)
	if err != nil {
		t.Error(err)
	}
	if i != int(version) {
		t.Error(i)
	}
	wait.Wait()
	fmt.Println("init done!")
}
func testRoundTrip(t *testing.T) {
	wait := sync.WaitGroup{}
	id := ID{2}
	function := "TestExecute"
	go func() {
		wait.Add(1)
		p := ExecuteMethod{Function: function}
		host.Send(ComExecute, p, id)
		resp := host.GetResponse()

		if resp.ID != id {
			t.Error("received wrong ID back")
		}
		v, succ := resp.Data.(bool)
		if !succ {
			t.Error("received wrong data type back")
		}
		if v != true {
			t.Error("received wrong value back")
		}
		wait.Done()
	}()
	message := client.Receive()
	if message.ID != id {
		t.Error("received wrong ID")
	}
	if message.Command != function {
		t.Error("received wrong command")
	}
	message.Respond(true)
	wait.Wait()
}
