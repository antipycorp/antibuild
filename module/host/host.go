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
	connection struct {
		send chan protocol.Response
	}
	Module struct {
		connections map[protocol.ID]*connection
		lock        sync.RWMutex
	}
)

func New() *Module {
	var module Module
	module.lock = sync.RWMutex{}
	module.connections = make(map[protocol.ID]*connection)
	return &module
}

//Start starts the Initites protocol for a given io.Reader and io.Writer.
func (m *Module) Start(in io.Reader, out io.Writer) {
	protocol.In = in
	protocol.Out = out
	protocol.Init(true)
	go func() {
		resp := protocol.GetResponse()
		conn := m.getCon(resp.ID)
		conn.send <- resp
	}()
}

func (m *Module) addConnection(id protocol.ID) {
	connection := connection{}
	connection.send = make(chan protocol.Response)
	m.setCon(id, &connection)
}

// AskMethods asks for the methods a module can handle, it returns a methods type
func (m *Module) AskMethods() (protocol.Methods, error) {
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, errors.New("could not generate random ID")
	}

	protocol.Send(protocol.GetAll, protocol.GetMethods{}, id)
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

// ExcecuteFunction asks for the methods a module can handle, it returns a methods type
func (m *Module) ExcecuteFunction(function string, args []interface{}) (interface{}, error) {
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, errors.New("could not generate random ID")
	}

	var payload protocol.ExecuteMethod
	payload.Function = function
	payload.Args = args
	protocol.Send(protocol.Execute, payload, id)
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

func (m *Module) awaitResponse(id protocol.ID) interface{} {
	con := m.getCon(id)
	resp := <-con.send
	return resp.Data
}

func (m *Module) getCon(id protocol.ID) *connection {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.connections[id]
}

func (m *Module) setCon(id protocol.ID, con *connection) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.connections[id] = con
}

func (m *Module) remCon(id protocol.ID, con connection) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.connections, id)
}
