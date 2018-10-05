package builder

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var (
	server http.Server
)

//locally hosts output folder
func hostLocally(config *Config, port string) {
	//make sure there is a port set
	addr := ":" + port
	if addr == ":" {
		addr = ":8080"
	}

	//host a static file server from the output folder
	r := http.StripPrefix("/", http.FileServer(http.Dir(config.Folders.Output)))
	server = http.Server{Addr: addr, Handler: r}
	server.ErrorLog = log.New(os.Stdout, "", 0)

	//start the server
	go panic(server.ListenAndServe())
	defer server.Shutdown(nil)
}

//watches files and folders and rebuilds when things change
func buildOnRefresh(config *Config, configLocation string) {
	//initalze watchers
	watcher, err := fsnotify.NewWatcher()
	staticWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("could not open a file watcher: ", err)
	}

	//add data folder to watcher
	err = filepath.Walk(config.Folders.Data, func(path string, file os.FileInfo, err error) error {
		return watcher.Add(path)
	})
	if err != nil {
		fmt.Println("failled walking over data folder: ", err)
	}

	//add template folder to watcher
	err = filepath.Walk(config.Folders.Templates, func(path string, file os.FileInfo, err error) error {
		return watcher.Add(path)
	})
	if err != nil {
		fmt.Println("failled walking over template folder: ", err)
	}

	//add static folder to staticWatcher
	err = filepath.Walk(config.Folders.Static, func(path string, file os.FileInfo, err error) error {
		return staticWatcher.Add(path)
	})
	if err != nil {
		fmt.Println("failled walking over static folder: ", err)
	}

	//listen for staticWatcher events
	go func() {
		for {
			select {
			case _, ok := <-staticWatcher.Events:
				if !ok {
					return
				}

				//check for folder
				info, err := os.Lstat(config.Folders.Static)
				if err != nil {
					fmt.Println("couldn't move files form static to out: ", err.Error())
				}

				//copy static folder
				genCopy(config.Folders.Static, config.Folders.Output, info)
			case err, ok := <-watcher.Errors:
				//handle staticWatcher errors
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	//listen for watcher events
	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}

			//start the parse
			startParse(configLocation)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			//handle static watch errors
			log.Println("error:", err)
		}
	}
}
