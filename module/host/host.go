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
)

var (
	connections map[protocol.ID]*connection
	lock        = sync.RWMutex{}
)

//Start starts the Initites protocol for a given io.Reader and io.Writer.
func Start(in io.Reader, out io.Writer) {
	protocol.In = in
	protocol.Out = out
	go func() {
		resp := protocol.GetResponse()
		conn := getCon(resp.ID)
		conn.send <- resp
	}()
}

func addConnection(id protocol.ID) {
	connection := connection{}
	connection.send = make(chan protocol.Response)
	setCon(id, &connection)
}

// AskMethods asks for the methods a module can handle, it returns a methods type
func AskMethods() (protocol.Methods, error) {
	var id [10]byte
	_, err := rand.Read(id[:])
	if err != nil {

	}

	protocol.Send("GetAll", protocol.GetMethods{}, id)
	addConnection(id)
	resp := awaitResponse(id)
	if resp == nil {
		if v, ok := resp.(error); ok {
			return nil, v
		}
		return nil, errors.New("could not receive error")
	}
	if v, ok := resp.(protocol.Methods); ok {
		return v, nil
	}
	return nil, errors.New("return datatype is incorrect")
}

func awaitResponse(id protocol.ID) interface{} {
	con := getCon(id)
	resp := <-con.send
	return resp.Data
}

func getCon(id protocol.ID) *connection {
	lock.RLock()
	defer lock.RUnlock()
	return connections[id]
}

func setCon(id protocol.ID, con *connection) {
	lock.Lock()
	defer lock.Unlock()
	connections[id] = con
}

func remCon(id protocol.ID, con connection) {
	lock.Lock()
	defer lock.Unlock()
	delete(connections, id)
}
