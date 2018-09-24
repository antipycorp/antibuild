package protocol

import (
	"encoding/gob"
	"io"
	"os"
	"sync"
)

type (
	//Methods is a map of available commands to list of functions allowed to be called
	Methods map[string][]string

	//GetMethods is the type used as payload for GetAll
	GetMethods struct{}

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
	outInit   sync.Once
	Out       = io.Writer(os.Stdout)
	writer    *gob.Encoder
	writeLock = sync.RWMutex{}

	inInit   sync.Once
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
