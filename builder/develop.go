package builder

import (
	"fmt"
	"io"
	"io/ioutil"
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

				info, err := os.Lstat(config.Folders.Static)
				if err != nil {
					fmt.Println("couldn't move files form static to out: ", err.Error())
				}

				genCopy(config.Folders.Static, config.Folders.Output, info)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}

			startParse(configLocation)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func genCopy(src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return dirCopy(src, dest, info)
	}
	return fileCopy(src, dest, info)
}

func fileCopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dirCopy(srcdir, destdir string, info os.FileInfo) error {

	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := genCopy(cs, cd, content); err != nil {
			return err
		}
	}
	return nil
}
