package protocol

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"sync"
)

type (
	//Methods is a map of available commands to list of functions allowed to be called
	Methods    map[string][]string
	getMethods struct{}

	executeMethod struct {
		Function string
		Args     []interface{}
	}

	payload interface {
		excecute([]byte) Token
	}

	message struct {
		Command string
		Payload payload
		ID      []byte
	}

	response struct {
		ID   []byte
		Data interface{}
	}
	//Token is a token used to receive data and send it back to the host
	Token struct {
		Command string
		Data    []interface{}
		ID      []byte
	}
)

var (
	outInit sync.Once
	out     = os.Stdout
	writer  *gob.Encoder

	inInit sync.Once
	in     = os.Stdin
	reader *gob.Decoder

	tokenGetMessages = Token{Command: "getmessages"}
	tokenReturnVars  = Token{Command: "return vars"}
)

func init() {
	gob.RegisterName("message", message{})
	gob.RegisterName("getmethods", getMethods{})

}

//Receive waits for a command from the host to excecute
func Receive() Token {
	inInit.Do(initIn)
	var command message
	reader.Decode(&command)
	return command.excecute()
}

func (m message) excecute() Token {
	return m.Payload.excecute(m.ID)
}

func (gm getMethods) excecute(id []byte) Token {
	ret := tokenGetMessages
	ret.ID = id
	return ret
}

func (gm executeMethod) excecute(id []byte) Token {
	ret := Token{Command: gm.Function}
	ret.ID = id
	ret.Data = gm.Args
	return ret
}

//Respond sends the given data back to the host
func (t *Token) Respond(data interface{}) {
	var resp response
	resp.Data = data
	resp.ID = t.ID
	writer.Encode(resp)
}

// AskMethods asks for the methods a module can handle, it returns a methods type
func AskMethods() (Methods, error) {
	outInit.Do(initOut)

	var command message
	var m getMethods

	command.Command = "GetAll"
	command.Payload = m

	writer.Encode(command)
	resp := waitResponse(command.ID)
	if resp == nil {
		if v, ok := resp.(error); ok {
			return nil, v
		}
		return nil, errors.New("could not receive error")
	}
	if v, ok := resp.(Methods); ok {
		return v, nil
	}
	return nil, errors.New("return datatype is incorrect")
}

func waitResponse(id []byte) interface{} {
	inInit.Do(initIn)
	var resp response
	reader.Decode(&resp)

	if !bytes.Equal(resp.ID, id) {
		return nil
	}
	return resp.Data
}

func initOut() {
	writer = gob.NewEncoder(out)
}

func initIn() {
	reader = gob.NewDecoder(in)
}
