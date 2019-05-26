// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package net

import (
	"log"
	"net"
	"net/http"
	"time"

	//pprof should only work when the host is on, otherwise its not gonna be used anyways
	_ "net/http/pprof"
	"os"

	tm "github.com/lucacasonato/goterm"
)

type host struct {
	http.Server
}

type handler struct {
	handler http.Handler
}

var (
	server   host
	shutdown chan int
)

//HostDebug host the /debug/pprof endpoint locally on port 5000
//panics if it cant host the server
func HostDebug() {
	debug := http.Server{
		Addr:        ":5000",
		Handler:     http.DefaultServeMux,
		ReadTimeout: time.Millisecond * 500,
	}
	go func() {
		err := debug.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

//HostLocally hosts output folder
func HostLocally(output, port string) {
	if shutdown == nil {
		shutdown = make(chan int, 1)
	} else if len(shutdown) != 0 {
		<-shutdown
	}

	//make sure there is a port set
	addr := ":" + port
	if addr == ":" {
		addr = ":8080"
	}

	//host a static file server from the output folder
	mux := http.NewServeMux()
	//mux.HandleFunc("/__/websocket", ws.Handle)
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(output))))
	server.Server = http.Server{
		Addr:         addr,
		Handler:      handler{mux},
		ReadTimeout:  time.Millisecond * 500,
		WriteTimeout: time.Millisecond * 500,
	}
	server.ErrorLog = log.New(os.Stdout, "", 0)
	server.RegisterOnShutdown(handleShutdown)

	//start the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		if sysErr, ok := err.(*net.OpError).Err.(*os.SyscallError); ok && sysErr.Err.Error() == "address already in use" {
			tm.Clear()
			tm.MoveCursor(1, 1)
			tm.Printf("The port %s is already being used by a different program."+"\n\n", tm.Bold(tm.Color(port, tm.RED)))
			tm.FlushAll()
			os.Exit(0)
		}
		panic(err)
	}
}

func handleShutdown() {
	shutdown <- 1
}

func (hndl handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("server", "antibuildmodulehost")
	hndl.handler.ServeHTTP(w, r)
}
