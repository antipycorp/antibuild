// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package host

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"gitlab.com/antipy/antibuild/module/protocol"
)

type (
	connection struct {
		send chan protocol.Response
	}
	ModuleHost struct {
		connections map[protocol.ID]*connection
		lock        sync.RWMutex
	}
)

var con *protocol.Connection

func New() *ModuleHost {
	var moduleHost ModuleHost
	moduleHost.lock = sync.RWMutex{}
	moduleHost.connections = make(map[protocol.ID]*connection)
	return &moduleHost
}

//Start starts the Initites protocol for a given io.Reader and io.Writer.
func (m *ModuleHost) Start(in io.Reader, out io.Writer) error {
	con = protocol.OpenConnection(in, out)
	_, err := con.Init(true)
	if err != nil {
		return err
	}

	go func() {
		for {
			resp := con.GetResponse()
			conn := m.getCon(resp.ID)
			conn.send <- resp
		}
	}()
	return nil
}

func (m *ModuleHost) addConnection(id protocol.ID) {
	connection := connection{}
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

	con.Send(protocol.GetMethods, protocol.ReceiveMethods{}, id)
	m.addConnection(id)
	resp := m.awaitResponse(id)
	fmt.Println(resp)
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
func (m *ModuleHost) ExcecuteMethod(function string, args []interface{}) (interface{}, error) {
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {
		return nil, errors.New("could not generate random ID")
	}

	var payload protocol.ExecuteMethod
	payload.Function = function
	payload.Args = args
	con.Send(protocol.ComExecute, payload, id)
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

func (m *ModuleHost) getCon(id protocol.ID) *connection {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.connections[id]
}

func (m *ModuleHost) setCon(id protocol.ID, con *connection) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.connections[id] = con
}

func (m *ModuleHost) remCon(id protocol.ID, con connection) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.connections, id)
}
