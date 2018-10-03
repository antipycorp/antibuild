// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package protocol

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type (
	//Methods is a map of available commands to list of functions allowed to be called
	Methods map[string][]string

	//ReceiveMethods is the type used as payload for GetAll
	ReceiveMethods struct{}
	//Version is the version type used for transmission of the version number
	Version int

	ExecuteMethod struct {
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
		con     *Connection
	}

	Connection struct {
		in     io.Reader
		inInit sync.Once
		reader *gob.Decoder

		out     io.Writer
		outInit sync.Once
		writer  *gob.Encoder

		rlock sync.Mutex
		wlock sync.Mutex
	}

	//ID is the type used for identification of the messages
	ID [10]byte
)

const (
	GetMethods = "internal_getMethods"
	CommandErr = "failled in receiving the command"

	ComExecute = "ExecuteMethod"
	ComVersion = "getVersion"
)

var (
	tokenGetVersion = Token{Command: ComVersion}
	tokenGetMethods = Token{Command: GetMethods}

	EOF = errors.New("could not get response")

	//version ID used for verifying versioning
	verifyVersionID = ID{1}
	version         = Version(1)

	//ErrProtocoolViolation is the error thrown whenever a protocol violation occurs
	ErrProtocoolViolation = errors.New("the protocol is violated by the opposite party, either the version is incompatible or the module is not a module")
)

func init() {
	gob.RegisterName("message", message{})
	gob.RegisterName("getMethods", ReceiveMethods{})
	gob.RegisterName("version", version)
	gob.RegisterName("id", verifyVersionID)
	gob.RegisterName("methds", Methods{})
	gob.RegisterName("exec", ExecuteMethod{})
}

func OpenConnection(in io.Reader, out io.Writer) *Connection {
	con := Connection{}
	con.in = in
	con.out = out
	con.inInit = sync.Once{}
	con.outInit = sync.Once{}
	con.rlock = sync.Mutex{}
	con.wlock = sync.Mutex{}
	return &con
}

//Init initiates the protocol with a version exchange, returns 0 as version when a protocol violation happens
func (c *Connection) Init(isHost bool) (int, error) {
	if isHost {
		c.Send(ComVersion, version, verifyVersionID)
		resp := c.GetResponse()

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
	message := c.Receive()
	if message.ID != verifyVersionID {
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
func (c *Connection) Receive() Token {
	var command message

	err := c.getMessage(&command)
	if err != nil {
		return Token{Data: []interface{}{EOF}}
	}

	token := command.excecute()
	token.con = c
	return token
}

//GetResponse waits for a response from the client
func (c *Connection) GetResponse() Response {
	var resp Response
	err := c.getMessage(&resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		resp.Data = errors.New("could not get response")
	}

	return resp
}

//Send sends a command to the guest
func (c *Connection) Send(command string, payload payload, id ID) {
	c.outInit.Do(initOut(c))

	var message message

	message.Command = command
	message.Payload = payload
	message.ID = id
	c.wlock.Lock()
	err := c.writer.Encode(message)
	c.wlock.Unlock()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not send message:", err)
	}

}

func (c *Connection) getMessage(m interface{}) error {
	c.inInit.Do(initIn(c))

	c.rlock.Lock()
	err := c.reader.Decode(m)
	c.rlock.Unlock()

	if err != nil {
		if err.Error() != "EOF" {
			fmt.Fprintf(os.Stderr, "could not get message: %v", err)
		}
		return err
	}
	return nil
}

func (m message) excecute() Token {
	return m.Payload.excecute(m.ID)
}

func (gm ReceiveMethods) excecute(id ID) Token {
	ret := tokenGetMethods
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

func (gm ExecuteMethod) excecute(id ID) Token {
	ret := Token{Command: gm.Function}
	ret.ID = id
	ret.Data = gm.Args
	return ret
}

//Respond sends the given data back to the host
func (t *Token) Respond(data interface{}) error {

	if t.con == nil {
		return errors.New("connection is nil")
	}
	t.con.outInit.Do(initOut(t.con))

	var resp Response
	resp.Data = data
	resp.ID = t.ID

	t.con.wlock.Lock()
	err := t.con.writer.Encode(resp)
	t.con.wlock.Unlock()

	if err != nil {
		fmt.Fprintf(os.Stderr, "failled sending:%v\n", err)
		return err
	}

	return nil
}

func initOut(c *Connection) func() {
	return func() {
		c.writer = gob.NewEncoder(c.out)
	}
}
func initIn(c *Connection) func() {
	return func() {
		c.reader = gob.NewDecoder(c.in)
	}
}
