package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"net/http"

	"gitlab.com/antipy/antibuild/cli/site"

	"github.com/fsnotify/fsnotify"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type (
	combined struct {
		Layout interface{}
		Page   interface{}
	}
	jsonImput struct {
		Data map[string]interface{} `json:"Data"`
	}
)

var (
	comms = map[string]func(string){
		"--config": setConfig,
	}
	flags = map[string]func(){
		"--development": setRefresh,
		"--host":        setHost,
	}

	isRefreshEnabled bool
	isHost           bool

	configLocation string
	isConfigSet    bool
)

var (
	server http.Server
	fn     = template.FuncMap{
		"noescape":  noescape,
		"mdprocess": mdprocess,
		"typeof":    typeof,
		"increment": increment,
	}
	errNoTemplate = errors.New("the template folder is not set")
	errNoJSON     = errors.New("the json folder is not set")
	errNoOutput   = errors.New("the output folder is not set")
)

const version = "v0.2.1"

func main() {
	fmt.Println(version)
	for i, comm := range os.Args {
		if _, ok := comms[comm]; ok {
			comms[comm](os.Args[i+1])
		} else if _, ok := flags[comm]; ok {
			flags[comm]()
		}
	}

	if isConfigSet {
		config, templateErr := executeTemplate()
		if templateErr != nil {
			fmt.Println("failled building templates: ", templateErr.Error())
		}

		if isHost {
			if templateErr == errNoOutput {
				panic("could not get the output folder from config.json")
			}

			addr := ":" + os.Getenv("PORT")
			if addr == ":" {
				addr = ":8080"
			}

			r := http.StripPrefix("/", http.FileServer(http.Dir(config.OUTFolder)))
			server = http.Server{Addr: addr, Handler: r}
			server.ErrorLog = log.New(os.Stdout, "", 0)

			go server.ListenAndServe()
			defer server.Shutdown(nil)
		}

		if isRefreshEnabled {
			if templateErr == errNoTemplate || templateErr == errNoJSON || templateErr == errNoOutput {
				panic("could not get one of: template, json or output from config.json")
			}
			watcher, err := fsnotify.NewWatcher()
			staticWatcher, err := fsnotify.NewWatcher()

			if err != nil {
				fmt.Println("could not open a file watcher: ", err)
			}
			err = filepath.Walk(config.JSONFolder, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})
			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}

			err = filepath.Walk(config.TemplateFolder, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})
			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}

			err = filepath.Walk(config.Static, func(path string, file os.FileInfo, err error) error {
				return staticWatcher.Add(path)
			})

			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}
			go func() {
				for {
					select {
					case _, ok := <-staticWatcher.Events:
						if !ok {
							return
						}
						fmt.Println("copying over files")
						info, err := os.Lstat(config.Static)
						if err != nil {
							fmt.Println("Couldnt move files form static to out: ", err.Error())
						}
						genCopy(config.Static, config.OUTFolder, info)
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
					config, err = executeTemplate()
					if err != nil {
						fmt.Println("failled building templates: ", err.Error())
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}
	}
}

func executeTemplate() (*site.Site, error) {
	var config site.Site
	JSONFile, err := os.Open(configLocation)
	defer JSONFile.Close()
	if err != nil {
		return nil, errors.New("Could not open the layout file: " + err.Error())
	}

	dec := json.NewDecoder(JSONFile)
	config.Sites = make([]*site.Site, 1)
	err = dec.Decode(&config.Sites[0])
	if err != nil {
		return &config, err
	}

	if config.OUTFolder == "" {
		err = os.RemoveAll(config.OUTFolder)
	}

	if err != nil {
		fmt.Println("could not remove files: ", err, " old html will be left in place")
	}

	fmt.Println("------ START ------")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(config)

	sitemap := site.SiteMap{}
	config.SiteMap = &sitemap
	err = config.Unfold(nil)
	if err != nil {
		fmt.Println("failled parsing config file")
	}
	enc.Encode(config)
	site.TemplateFunctions = &fn
	err = config.Execute()
	if err != nil {
		fmt.Println("failled executing")
	}

	if config.Sites[0].TemplateFolder == "" {
		return config.Sites[0], errNoTemplate
	}
	if config.Sites[0].JSONFolder == "" {
		return config.Sites[0], errNoJSON
	}
	if config.Sites[0].OUTFolder == "" {
		return config.Sites[0], errNoOutput
	}
	return config.Sites[0], err
}

func setConfig(config string) {
	configLocation = config
	isConfigSet = true
}

func setRefresh() {
	isRefreshEnabled = true
}
func setHost() {
	isHost = true
}

func noescape(str string) template.HTML {
	return template.HTML(str)
}

func mdprocess(md string) template.HTML {
	return template.HTML(string(blackfriday.Run([]byte(md), blackfriday.WithExtensions(blackfriday.HardLineBreak))))
}

func increment(no int) int {
	no++
	return no
}

func typeof(thing interface{}) string {
	fmt.Println(reflect.TypeOf(thing))
	return "Check Console"
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
			// If any error, exit immediately
			return err
		}
	}
	return nil
}
