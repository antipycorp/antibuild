package net

import (
	"log"
	"net/http"
	"time"

	//pprof should only work when the host is on, otherwise its not gonna be used anyways
	_ "net/http/pprof"
	"os"

	ws "gitlab.com/antipy/antibuild/cli/net/websocket"
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

//HostLocally hosts output folder
func HostLocally(output, port string) {
	//make sure there is a port set
	addr := ":" + port
	if addr == ":" {
		addr = ":8080"
	}
	if os.Getenv("DEBUG") == "1" {
		debug := http.Server{
			Addr:        ":5000",
			Handler:     http.DefaultServeMux,
			ReadTimeout: time.Millisecond * 500,
		}
		go debug.ListenAndServe()
	}

	//host a static file server from the output folder
	mux := http.NewServeMux()
	mux.HandleFunc("/__/websocket", ws.Handle)
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(output))))
	server.Server = http.Server{
		Addr:         addr,
		Handler:      handler{mux},
		ReadTimeout:  time.Millisecond * 500,
		WriteTimeout: time.Millisecond * 500,
	}
	server.ErrorLog = log.New(os.Stdout, "", 0)

	//start the server
	panic(server.ListenAndServe())
}

func (hndl handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("server", "antibuildmodulehost")
	hndl.handler.ServeHTTP(w, r)
}
