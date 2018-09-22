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
	"regexp"
	"strings"

	"net/http"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/russross/blackfriday.v2"
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
		Slug            string   `json:"Slug"`
		Templates       []string `json:"Templates"`
		JSONFiles       []string `json:"JSONfiles"`
		Languages       []string `json:"languages"`
		Sites           []site   `json:"sites"`
		TemplateFolder  string   `json:"templateroot"`
		JSONFolder      string   `json:"jsonroot"`
		OUTFolder       string   `json:"outroot"`
		Static          string   `json:"staticroot"`
		DefaultLanguage string   `json:"defaultlanguage"`
		language        string
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

const version = "v0.1.1"

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

func executeTemplate() (*site, error) {
	var config site
	JSONFile, err := os.Open(configLocation)
	defer JSONFile.Close()
	if err != nil {
		return nil, errors.New("Could not open the layout file: " + err.Error())
	}

	dec := json.NewDecoder(JSONFile)
	err = dec.Decode(&config)
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
	err = config.execute(nil)
	if err != nil {
		fmt.Println("failled parsing config file")
	}

	if config.TemplateFolder == "" {
		return &config, errNoTemplate
	}
	if config.JSONFolder == "" {
		return &config, errNoJSON
	}
	if config.OUTFolder == "" {
		return &config, errNoOutput
	}
	return &config, err
}

func (s *site) execute(parent *site) error {
	if parent != nil {
		if s.JSONFiles != nil {
			s.JSONFiles = append(parent.JSONFiles, s.JSONFiles...)
		} else {
			s.JSONFiles = make([]string, len(parent.JSONFiles))
			copy(s.JSONFiles, parent.JSONFiles)
		}
		if s.Templates != nil {
			s.Templates = append(parent.Templates, s.Templates...)
		} else {
			s.Templates = make([]string, len(parent.Templates))
			copy(s.Templates, parent.Templates)
		}
		if s.Languages != nil {
			s.Languages = append(parent.Languages, s.Languages...)
		} else {
			s.Languages = make([]string, len(parent.Languages))
			copy(s.Languages, parent.Languages)
		}
		s.Slug = parent.Slug + s.Slug
		if parent.OUTFolder != "" {
			s.OUTFolder = parent.OUTFolder
		}
		if parent.TemplateFolder != "" {
			s.TemplateFolder = parent.TemplateFolder
		}
		if parent.JSONFolder != "" {
			s.JSONFolder = parent.JSONFolder
		}
		if parent.DefaultLanguage != "" {
			s.DefaultLanguage = parent.DefaultLanguage
		}
	}

	if s.Static != "" && s.OUTFolder != "" {
		fmt.Println("copying static files")
		info, err := os.Lstat(s.Static)
		if err != nil {
			return err
		}
		genCopy(s.Static, s.OUTFolder, info)
	}

	for jIndex, jsonfile := range s.JSONFiles {
		if strings.Contains(jsonfile, "*") {
			return parseStar(s, jIndex)
		}
	}

	if s.Sites != nil {
		for _, site := range s.Sites {
			err := site.execute(s)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if s.Languages != nil && s.Sites == nil && s.language == "" {
		for _, lang := range s.Languages {
			site := s.copy()
			site.language = lang
			err := site.execute(nil)
			if err != nil {
				return fmt.Errorf("could not execute %s the for lang %s:", site.Slug, lang, err)
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

func parseStar(s *site, jIndex int) error {
	jsonPath := filepath.Dir(filepath.Join(s.JSONFolder, s.JSONFiles[jIndex]))
	jsonfile := strings.Replace(filepath.Base(s.JSONFiles[jIndex]), "*", "([^/]*)", -1)
	re := regexp.MustCompile(jsonfile)
	var matches [][][]string
	//fmt.Println(jsonfile)
	//fmt.Println(jsonPath)
	err := filepath.Walk(jsonPath, func(path string, file os.FileInfo, err error) error {
		if path == jsonPath {
			return nil
		}
		//fmt.Println(file.Name())
		//fmt.Println(jsonfile)
		if file.IsDir() {
			return filepath.SkipDir
		}
		if ok, _ := regexp.MatchString(jsonfile, file.Name()); ok {
			matches = append(matches, re.FindAllStringSubmatch(file.Name(), -1))
		}
		return nil
	})
	if err != nil {
		return nil
	}
	//fmt.Println(matches)
	for _, file := range matches {
		site := s.copy()
		for _, match := range file {
			site.Slug = strings.Replace(site.Slug, "*", match[1], 1)
			site.JSONFiles[jIndex] = strings.Replace(site.JSONFiles[jIndex], "*", match[1], 1)
		}
		err := site.execute(nil)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *site) gatherJSON(jsonImput *jsonImput) error {
	fmt.Println("gathering JSON files for: ", s.Slug)

	for _, jsonLocation := range s.JSONFiles {
		jsonPath := filepath.Join(s.JSONFolder, jsonLocation)

		JSONFile, err := os.Open(jsonPath)
		defer JSONFile.Close()
		if err != nil {
			return err
		}

		dec := json.NewDecoder(JSONFile)
		err = dec.Decode(&jsonImput)
		if s.language != "" {
			if v, ok := jsonImput.Data[s.language].(map[string]interface{}); ok {
				jsonImput.Data = v
			} // else: if it cant find the language just use the whole json to allow for languageless jsonfiles
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *site) gatherTemplates() (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(s.TemplateFolder, s.Templates[i])
	}

	template, err := template.New("").Funcs(fn).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func (s *site) executeTemplate(template *template.Template, jsonImput jsonImput) error {
	OUTPath := filepath.Join(s.OUTFolder, s.Slug)
	if s.language != "" && s.DefaultLanguage != s.language {
		OUTPath = filepath.Join(s.OUTFolder, s.language, s.Slug)
	}

	err := os.MkdirAll(filepath.Dir(OUTPath), 0766)
	if err != nil {
		return errors.New("Couldn't create directory: " + err.Error())
	}

	//fmt.Println(s.Slug)
	OUTFile, err := os.Create(OUTPath)
	if err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	//enc.Encode(jsonImput)

	err = template.ExecuteTemplate(OUTFile, "html", jsonImput.Data)
	if err != nil {
		return errors.New("Could not parse: " + err.Error())
	}
	return nil
}

func (s *site) copy() site {
	newSite := *s
	for i, site := range s.Sites {
		newSite.Sites[i] = site.copy()
	}
	newSite.JSONFiles = make([]string, len(s.JSONFiles))
	copy(newSite.JSONFiles, s.JSONFiles)

	newSite.Templates = make([]string, len(s.Templates))
	copy(newSite.Templates, s.Templates)

	return newSite
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
