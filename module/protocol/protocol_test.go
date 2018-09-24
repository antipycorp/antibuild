package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

func TestProtocol(t *testing.T) {
	//t.Run("Receive", TestReceive)

	t.Run("Receive", TestReceive)
}

func TestReceive(t *testing.T) {

	var in bytes.Buffer  // Stand-in for the network.
	var out bytes.Buffer // Stand-in for the network.

	writer = gob.NewEncoder(&out)
	reader = gob.NewDecoder(&in)

	var command message
<<<<<<< HEAD
	var methods GetMethods
=======
	var methods getMethods
>>>>>>> ebe3b2a495d0338e065b3484550fc1b8a199b64a
	command.Command = "GetAll"
	command.Payload = methods

	enc := gob.NewEncoder(&in)
	err := enc.Encode(&command)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(Receive())
}
