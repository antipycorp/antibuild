// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package host

import (
	"errors"
	"io"
	"math/rand"
	"sync"

	"gitlab.com/antipy/antibuild/module/protocol"
)

type (
	command struct {
		send chan protocol.Response
	}
	ModuleHost struct {
		commands map[protocol.ID]*command
		lock     sync.RWMutex
		con      *protocol.Connection
	}
)

//Start starts the Initites protocol for a given io.Reader and io.Writer.
func Start(in io.Reader, out io.Writer) (moduleHost *ModuleHost, err error) {
	moduleHost.lock = sync.RWMutex{}
	moduleHost.commands = make(map[protocol.ID]*command)

	moduleHost.con = protocol.OpenConnection(in, out)
	_, err = moduleHost.con.Init(true)
	if err != nil {
		return
	}

	go func() {
		for {
			resp := moduleHost.con.GetResponse()
			conn := moduleHost.getCon(resp.ID)
			conn.send <- resp
		}
	}()
	return
}

func (m *ModuleHost) addConnection(id protocol.ID) {
	connection := command{}
	connection.send = make(chan protocol.Response)
	m.setCon(id, &connection)
}

// AskMethods asks for the methods a moduleHost can handle, it returns a methods type
func (m *ModuleHost) AskMethods() (protocol.Methods, error) {

	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, errors.New("could not generate random ID")
	}

	m.con.Send(protocol.GetMethods, protocol.ReceiveMethods{}, id)
	m.addConnection(id)
	resp := m.awaitResponse(id)
	if resp == nil {
		return nil, errors.New("could not receive error")
	}
	if v, ok := resp.(error); ok {
		return nil, v
	}
	if v, ok := resp.(protocol.Methods); ok {
		return v, nil
	}
	return nil, errors.New("return datatype is incorrect")
}

// ExcecuteMethod asks for the methods a moduleHost can handle, it returns a methods type
func (m *ModuleHost) ExcecuteMethod(function string, args interface{}) (interface{}, error) {
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, errors.New("could not generate random ID")
	}

	var payload protocol.ExecuteMethod
	payload.Function = function
	payload.Args = args

	m.con.Send(protocol.ComExecute, payload, id)
	m.addConnection(id)
	resp := m.awaitResponse(id)
	if resp == nil {
		return nil, errors.New("could not receive error")
	}
	if v, ok := resp.(error); ok {
		return nil, v
	}
	return nil, errors.New("return datatype is incorrect")
}

func (m *ModuleHost) awaitResponse(id protocol.ID) interface{} {
	con := m.getCon(id)
	resp := <-con.send
	return resp.Data
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

func (m *ModuleHost) remCon(id protocol.ID, con command) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.commands, id)
}
