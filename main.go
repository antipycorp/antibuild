package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"net/http"

	"github.com/fsnotify/fsnotify"
)

type (
	combined struct {
		Layout interface{}
		Page   interface{}
	}
	jsonImput struct {
		Data map[string]interface{} `json:"Data"`
	}

	site struct {
		Slug      string   `json:"Slug"`
		Templates []string `json:"Templates"`
		JSONFiles []string `json:"JSONfiles"`
		Sites     []site   `json:"sites"`
	}
)

var (
	comms = map[string]func(string){
		"--templates": setTemplate,
		"--json":      setJSON,
		"--out":       setOUT,
	}
	flags = map[string]func(){
		"--development": setRefresh,
		"--host":        setHost,
	}

	isTemplateSet  bool
	folderTemplate string

	isJSONSet  bool
	folderJSON string

	isOUTSet  bool
	folderOUT string

	isRefreshEnabled bool
	isHost           bool
)

var (
	server http.Server
	fn     = template.FuncMap{
		"noescape": noescape,
	}
)

func main() {
	for i, comm := range os.Args {
		if _, ok := comms[comm]; ok {
			comms[comm](os.Args[i+1])
		} else if _, ok := flags[comm]; ok {
			flags[comm]()
		}
	}

	if isTemplateSet && isJSONSet && isOUTSet {
		err := executeTemplate()
		if err != nil {
			fmt.Println("failled building templates: ", err.Error())
		}

		if isHost {
			addr := ":" + os.Getenv("PORT")
			if addr == ":" {
				addr = ":8080"
			}

			r := http.StripPrefix("/", http.FileServer(http.Dir(folderOUT)))
			server = http.Server{Addr: addr, Handler: r}
			server.ErrorLog = log.New(os.Stdout, "", 0)

			go server.ListenAndServe()
			defer server.Shutdown(nil)
		}

		if isRefreshEnabled {
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				fmt.Println("could not open a file watcher: ", err)
			}
			err = filepath.Walk(folderJSON, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})
			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}

			err = filepath.Walk(folderTemplate, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})

			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}
			for {
				select {
				case _, ok := <-watcher.Events:
					if !ok {
						return
					}
					err := executeTemplate()
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

func executeTemplate() error {
	err := os.RemoveAll(folderOUT)
	if err != nil {
		fmt.Println("could not remove files: ", err, " old html will be left in place")
	}
	var config site
	JSONFile, err := os.Open(filepath.Join(folderJSON, "/config.json"))
	defer JSONFile.Close()
	if err != nil {
		return errors.New("Could not open the layout file: " + err.Error())
	}

	dec := json.NewDecoder(JSONFile)
	err = dec.Decode(&config)
	if err != nil {
		return err
	}

	fmt.Println("------ START ------")
	for _, site := range config.Sites {
		err := site.execute(config.JSONFiles, config.Templates, config.Slug)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *site) execute(JSONFiles, templates []string, slug string) error {
	if s.JSONFiles != nil {
		s.JSONFiles = append(JSONFiles, s.JSONFiles...)
	} else {
		s.JSONFiles = make([]string, len(JSONFiles))
		copy(s.JSONFiles, JSONFiles)
	}
	if s.Templates != nil {
		s.Templates = append(templates, s.Templates...)
	} else {
		s.Templates = make([]string, len(templates))
		copy(s.Templates, templates)
	}
	s.Slug = slug + s.Slug

	if s.Sites != nil {
		for _, site := range s.Sites {
			err := site.execute(s.JSONFiles, s.Templates, s.Slug)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var jsonImput jsonImput

	err := s.gatherJSON(&jsonImput)
	if err != nil {
		return err
	}
	template, err := s.gatherTemplates()
	if err != nil {
		return err
	}
	err = s.executeTemplate(template, jsonImput)
	if err != nil {
		return err
	}

	return nil
}

func (s *site) gatherJSON(jsonImput *jsonImput) error {
	fmt.Println("gathering JSON files for: ", s.Slug)
	for _, jsonLocation := range s.JSONFiles {
		jsonPath := filepath.Join(folderJSON, jsonLocation)

		JSONFile, err := os.Open(jsonPath)
		defer JSONFile.Close()
		if err != nil {
			return err
		}

		dec := json.NewDecoder(JSONFile)
		err = dec.Decode(&jsonImput)
		if err != nil {
			return err
		}
		fmt.Println(jsonLocation)

	}
	return nil
}

func (s *site) gatherTemplates() (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(folderTemplate, s.Templates[i])
	}

	OUTPath := filepath.Join(folderOUT, s.Slug)

	err := os.MkdirAll(filepath.Dir(OUTPath), 0766)
	if err != nil {
		return nil, errors.New("Couldn't create directory: " + err.Error())
	}

	template, err := template.New("").Funcs(fn).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func (s *site) executeTemplate(template *template.Template, jsonImput jsonImput) error {
	OUTPath := filepath.Join(folderOUT, s.Slug)
	fmt.Println(s.Slug)
	OUTFile, err := os.Create(OUTPath)
	if err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}
	err = template.ExecuteTemplate(OUTFile, "html", jsonImput.Data)
	if err != nil {
		return errors.New("Could not parse: " + err.Error())
	}
	return nil
}

func (ji *jsonImput) UnmarshalJSON(data []byte) error {
	var input map[string]interface{}
	err := json.Unmarshal(data, &input)
	if err != nil {
		return err
	}

	if ji.Data == nil {
		ji.Data = make(map[string]interface{})
	}

	for name, in := range input {
		ji.Data[name] = in
	}
	return nil
}

func setTemplate(templateFolder string) {
	isTemplateSet = true
	folderTemplate = templateFolder
}

func setJSON(JSONFolder string) {
	isJSONSet = true
	folderJSON = JSONFolder
}

func setOUT(OUTFolder string) {
	isOUTSet = true
	folderOUT = OUTFolder
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
