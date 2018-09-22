package protocol

import (
	"encoding/gob"
	"os"
)

type (
	getMethods struct{}

	executeMethod struct {
		Function string
		Args     []interface{}
	}

	payload interface {
		excecute([]byte) token
	}

	message struct {
		Command string
		Payload payload
		ID      []byte
	}

	token struct {
		Command string
		Data    []interface{}
		ID      []byte
	}
	response struct {
		ID   []byte
		Data interface{}
	}
)

var (
	out    = os.Stdout
	writer *gob.Encoder

	in     = os.Stdin
	reader *gob.Decoder

	tokenGetMessages = token{Command: "getmessages"}
)

func init() {
	gob.RegisterName("message", message{})
	gob.RegisterName("getmethods", getMethods{})

	writer = gob.NewEncoder(out)
	reader = gob.NewDecoder(in)
}

func Receive() token {
	var command message
	reader.Decode(&command)
	return command.excecute()
}

func (m message) excecute() token {
	return m.Payload.excecute(m.ID)
}

func (gm getMethods) excecute(id []byte) token {
	ret := tokenGetMessages
	ret.ID = id
	return ret
}

func (gm executeMethod) excecute(id []byte) token {
	ret := token{Command: gm.Function}
	ret.ID = id
	ret.Data = gm.Args
	return ret
}

func (t *token) Respond(data interface{}) {
	var resp response
	resp.Data = data
	resp.ID = t.ID
	writer.Encode(resp)
}
