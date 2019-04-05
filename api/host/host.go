// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package host

import (
	"io"
	"math/rand"
	"sync"

	"gitlab.com/antipy/antibuild/cli/api/errors"
	// Registers types for GOB
	_ "gitlab.com/antipy/antibuild/cli/api/internal"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
)

type (
	command struct {
		send chan protocol.Response
	}

	//ModuleHost is the host for a module with all the data
	ModuleHost struct {
		commands map[protocol.ID]*command
		lock     sync.RWMutex
		con      *protocol.Connection
		Logger
	}

	//Logger is a simple logger only used for debug output
	Logger interface {
		Fatal(string)
		Fatalf(string, ...interface{})
		Error(string)
		Errorf(string, ...interface{})
		Info(string)
		Infof(string, ...interface{})
		Debug(string)
		Debugf(string, ...interface{})
	}

	logger struct{}
)

var (
	//ErrFailedToGenID is for when it failed generating an ID
	ErrFailedToGenID = errors.New("could not generate random ID", errors.CodeFatal)
	//ErrInvalidResponse is for when the response is of a wrong type
	ErrInvalidResponse = errors.New("return datatype is incorrect", errors.CodeInvalidResponse)
)

//Start starts the Initites protocol for a given io.Reader and io.Writer.
func Start(in io.Reader, out io.Writer, log Logger) (moduleHost *ModuleHost, err error) {
	moduleHost = &ModuleHost{}

	moduleHost.lock = sync.RWMutex{}
	moduleHost.commands = make(map[protocol.ID]*command)
	moduleHost.con = protocol.OpenConnection(in, out)

	moduleHost.Logger = log

	if moduleHost.Logger == nil {
		moduleHost.Logger = logger{}
	}

	_, err = moduleHost.con.Init(true)

	if err != nil {
		return
	}

	go func() {
		for {
			resp := moduleHost.con.GetResponse()
			conn := moduleHost.getCon(resp.ID)
			if conn == nil {
				return
			}
			conn.send <- resp
		}
	}()
	return
}

//Kill kills the module
func (m *ModuleHost) Kill() {
	var id [10]byte
	rand.Read(id[:]) //even if the ID sucks and isnt complete it doesnt matter since we dont expect a response

	m.con.Send(protocol.KillCommand, protocol.Kill{}, id)
}

// AskMethods asks for the methods a moduleHost can handle, it returns a methods type
func (m *ModuleHost) AskMethods() (protocol.Methods, error) {

	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, ErrFailedToGenID
	}

	m.addConnection(id)
	m.con.Send(protocol.GetMethods, protocol.ReceiveMethods{}, id)

	resp, errs := m.awaitResponse(id)
	m.remCon(id)

	var v protocol.Methods
	var ok bool
	if v, ok = resp.(protocol.Methods); !ok {
		return nil, ErrInvalidResponse
	}

	for _, err := range errs {
		switch err.Code {
		case errors.CodeDebug:
			m.Logger.Debug(err.Error())
		case errors.CodeInfo:
			m.Logger.Info(err.Error())
		case errors.CodeError:
			m.Logger.Error(err.Error())
		case errors.CodeFatal:
			m.Logger.Fatal(err.Error())
		}
	}

	return v, nil
}

// ExcecuteMethod asks for the methods a moduleHost can handle, it returns a methods type
func (m *ModuleHost) ExcecuteMethod(function string, args []interface{}) (interface{}, error) {
	//fmt.Println("doing execute method")
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, ErrFailedToGenID
	}

	var payload protocol.ExecuteMethod
	payload.Function = function
	payload.Args = args

	m.addConnection(id)
	m.con.Send(protocol.ComExecute, payload, id)

	resp, errs := m.awaitResponse(id)
	m.remCon(id)

	for _, err := range errs {
		switch err.Code {
		case errors.CodeInfo:
			m.Logger.Debug(err.Error())
		case errors.CodeError:
			m.Logger.Error(err.Error())
		case errors.CodeFatal:
			m.Logger.Fatal(err.Error())
		}
	}

	return resp, nil
}

func (m *ModuleHost) addConnection(id protocol.ID) {
	connection := command{}
	connection.send = make(chan protocol.Response)
	m.setCon(id, &connection)
}

func (m *ModuleHost) awaitResponse(id protocol.ID) (interface{}, []errors.Error) {
	con := m.getCon(id)
	resp := <-con.send
	return resp.Data, resp.Log
}

func (m *ModuleHost) getCon(id protocol.ID) *command {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.commands[id]
}

func (m *ModuleHost) setCon(id protocol.ID, con *command) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.commands[id] = con
}

func (m *ModuleHost) remCon(id protocol.ID) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.commands, id)
}

func (l logger) Fatal(s string) {
	return
}
func (l logger) Fatalf(s string, v ...interface{}) {
	return
}

func (l logger) Error(s string) {
	return
}
func (l logger) Errorf(s string, v ...interface{}) {
	return
}

func (l logger) Info(s string) {
	return
}
func (l logger) Infof(s string, v ...interface{}) {
	return
}

func (l logger) Debug(s string) {
	return
}
func (l logger) Debugf(s string, v ...interface{}) {
	return
}
