// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package protocol

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"sync"
)

type (
	//Methods is a map of available commands to list of functions allowed to be called
	Methods map[string][]string

	//GetMethods is the type used as payload for GetAll
	GetMethods struct{}
	//Version is the version type used for transmission of the version number
	Version int

	executeMethod struct {
		Function string
		Args     []interface{}
	}

	payload interface {
		excecute(ID) Token
	}

	message struct {
		Command string
		Payload payload
		ID      ID
	}

	//Response is the response for a given command
	Response struct {
		ID   ID
		Data interface{}
	}

	//Token is a token used to receive data and send it back to the host
	Token struct {
		Command string
		Data    []interface{}
		ID      ID
	}

	//ID is the type used for identification of the messages
	ID [10]byte
)

var (
	outInit sync.Once
	//Out is the variable where the Guest/Host writes to
	Out       = io.Writer(os.Stdout)
	writer    *gob.Encoder
	writeLock = sync.RWMutex{}

	inInit sync.Once
	//In is the variable where the Guest/Host reads from
	In       = io.Reader(os.Stdin)
	reader   *gob.Decoder
	readLock = sync.RWMutex{}

	tokenGetTemplateFunctions = Token{Command: "getTemplateFunctions"}
	tokenReturnVars           = Token{Command: "return vars"}

	IDError = [10]byte{0}
)

func init() {
	gob.RegisterName("message", message{})
	gob.RegisterName("getmethods", GetMethods{})
}

//Init initiates the protocol with a version exchange, returns 0 as version when a protocol violation happens
func Init(isHost bool) (int, error) {
	if isHost {
		Send("GetVersion", version, verifyVersionID)
		resp := GetResponse()
		if resp.ID != verifyVersionID {
			return 0, ErrProtocoolViolation
		}
		v, ok := resp.Data.(Version)
		if !ok {
			return 0, ErrProtocoolViolation
		}
		if v < version {
			return int(v), errors.New("Guest is using an older version of the API")
		}
		if v > version {
			return int(v), errors.New("Guest is using a newer version of the API")
		}

		return int(v), nil
	}
	message := Receive()
	if message.ID != verifyVersionID {
		return 0, ErrProtocoolViolation
	}
	if len(message.Data) != 0 {
		return 0, ErrProtocoolViolation
	}
	v, ok := message.Data[0].(Version)
	if !ok {
		return 0, ErrProtocoolViolation
	}
	if v < version {
		return int(v), errors.New("Host is using an older version of the API")
	}
	if v > version {
		return int(v), errors.New("Host is using a newer version of the API")
	}

	message.Respond(version)
	return int(v), nil
}

//Receive waits for a command from the host to excecute
func Receive() Token {
	var command message

	getMessage(&command)
	return command.excecute()
}

//GetResponse waits for a response from the client
func GetResponse() Response {
	var resp Response

	getMessage(&resp)
	return resp
}

//Send sends a command to the guest
func Send(command string, payload payload, id ID) {
	outInit.Do(initOut)

	var message message

	message.Command = command
	message.Payload = payload
	message.ID = id
	writeLock.Lock()
	writer.Encode(command)
	writeLock.Unlock()

}

func getMessage(message interface{}) {
	inInit.Do(initIn)
	readLock.Lock()
	reader.Decode(message)
	readLock.Unlock()
}

func (m message) excecute() Token {
	return m.Payload.excecute(m.ID)
}

func (gm GetMethods) excecute(id ID) Token {
	ret := tokenGetTemplateFunctions
	ret.ID = id
	return ret
}

func (v Version) excecute(id ID) Token {
	ret := tokenGetVersion
	ret.ID = id
	ret.Data = make([]interface{}, 1)
	ret.Data[0] = v
	return ret
}

func (gm executeMethod) excecute(id ID) Token {
	ret := Token{Command: gm.Function}
	ret.ID = id
	ret.Data = gm.Args
	return ret
}

//Respond sends the given data back to the host
func (t *Token) Respond(data interface{}) {
	var resp Response
	resp.Data = data
	resp.ID = t.ID
	writer.Encode(resp)
}

func initOut() {
	writer = gob.NewEncoder(Out)
}

func initIn() {
	reader = gob.NewDecoder(In)
}
