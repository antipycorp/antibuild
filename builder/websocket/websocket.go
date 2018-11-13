package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	transcoder interface {
		encode(interface{}) error
		//decode into a pointer not a struct
		decode(interface{}) error
		init(*websocket.Conn)
	}
	jsonTranscoder struct {
		conn *websocket.Conn
	}
	request struct {
		Action string
		ID     []byte
	}
	response struct {
		Error string
		ID    []byte
	}
	//Connection is the type used to represent a connection
	Connection struct {
		*websocket.Conn
		transcoder
	}
	//Action is an action activatable by WS connections
	Action func() error
)

const (
	//JSONComunication is the constant that indicates a JSON sommunication
	JSONComunication = "JSON"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	taken       = false
	transcoders = map[string]transcoder{
		JSONComunication: &jsonTranscoder{},
	}
	//actions are the functions that can be called by WS clients
	actions = make(map[string]Action)
	cons    = make(map[ID]*Connection)
)

//SendUpdate sends an update request to all the WS clients
func SendUpdate() {
	for _, v := range cons {
		v.encode("UPDATE!")
	}
}

//AddAction adds an action to a map of actions, overwrites already set functions
func AddAction(name string, action Action) {
	actions[name] = action
}

//Handle handles the websocket connection
func Handle(w http.ResponseWriter, r *http.Request) {
	defer func() { taken = false }()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var com Connection
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		if messageType == websocket.TextMessage {
			com = setup(conn, string(p))
			if com != (Connection{}) {
				break
			}
			continue
		}
	}
	for {
		var message request
		if com.decode(&message) != nil {
			log.Println(err)
			return
		}
		if tr, ok := actions[message.Action]; ok {
			err := tr()
			response := response{
				err.Error(),
				message.ID,
			}
			com.encode(response)
		}
	}
}

func setup(conn *websocket.Conn, transid string) (com Connection) {
	if tr, ok := transcoders[transid]; ok {

		com.transcoder = tr
		com.setCon(conn)
	genID:
		cID := genID()
		if _, exists := cons[cID]; exists {
			goto genID
		}
		cons[cID] = &com
		conn.WriteMessage(websocket.TextMessage, cID[:])
		return
	}
	conn.WriteMessage(websocket.TextMessage, []byte(JSONComunication))

	return
}

func transCoder(transcoder string) transcoder {
	if tr, ok := transcoders[transcoder]; ok {
		return tr
	}
	return nil
}

func (jt *jsonTranscoder) decode(v interface{}) error {
	return jt.conn.ReadJSON(v)
}

func (jt *jsonTranscoder) encode(v interface{}) error {
	return jt.conn.WriteJSON(v)
}

func (jt *jsonTranscoder) init(conn *websocket.Conn) {
	jt.conn = conn
}

func (con *Connection) setCon(conn *websocket.Conn) error {
	con.Conn = conn
	con.transcoder.init(conn)
	return nil
}
