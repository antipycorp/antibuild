package builder

import (
	"log"
	"net/http"
	"os"

	ws "gitlab.com/antipy/antibuild/cli/builder/websocket"
)

type host struct {
	http.Server
}

type handler struct {
	handler http.Handler
}

var (
	server host
)

//locally hosts output folder
func hostLocally(output, port string) {
	//make sure there is a port set
	addr := ":" + port
	if addr == ":" {
		addr = ":8080"
	}

	//host a static file server from the output folder
	mux := http.NewServeMux()
	mux.HandleFunc("/websocket", ws.Handle)
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(output))))
	server.Server = http.Server{Addr: addr, Handler: handler{mux}}
	server.ErrorLog = log.New(os.Stdout, "", 0)

	//start the server
	panic(server.ListenAndServe())
}

func (hndl handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("server", "antibuildmodulehost")
	hndl.handler.ServeHTTP(w, r)
}
