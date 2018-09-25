// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"net/http"

	"github.com/fsnotify/fsnotify"
	"gitlab.com/antipy/antibuild/module/host"
)

type (
	combined struct {
		Layout interface{}
		Page   interface{}
	}

	jsonDataFile struct {
		Data map[string]interface{} `json:"Data"`
	}

	config struct {
		Folders    configFolder  `json:"folders"`
		Modules    configModules `json:"modules"`
		Pages      site          `json:"pages"`
		moduleHost *host.ModuleHost
	}

	configFolder struct {
		Templates string `json:"templates"`
		Data      string `json:"data"`
		Static    string `json:"static"`
		Output    string `json:"output"`
		Modules   string `json:"modules"`
	}

	configModules struct {
		Dependencies map[string]string                 `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config"`
	}

	site struct {
		Slug            string   `json:"slug"`
		Templates       []string `json:"templates"`
		Data            []string `json:"data"`
		Languages       []string `json:"languages"`
		DefaultLanguage string   `json:"language_default"`

		Sites []site `json:"sites"`

		language string
	}
)

var (
	server http.Server
	fn     = template.FuncMap{
		"noescape":  noescape,
		"typeof":    typeof,
		"increment": increment,
	}
	errNoTemplate = errors.New("the template folder is not set")
	errNoData     = errors.New("the data folder is not set")
	errNoOutput   = errors.New("the output folder is not set")

	loadedModules = false
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool) {
	if isConfigSet {
		config, parseErr := startParse(configLocation)

		if parseErr != nil {
			panic(fmt.Sprintf("could not get the output folder from config.json: %s", parseErr))
		}

		if isHost {

			addr := ":" + os.Getenv("PORT")
			if addr == ":" {
				addr = ":8080"
			}

			r := http.StripPrefix("/", http.FileServer(http.Dir(config.Folders.Output)))
			server = http.Server{Addr: addr, Handler: r}
			server.ErrorLog = log.New(os.Stdout, "", 0)

			go server.ListenAndServe()
			defer server.Shutdown(nil)
		}

		if isRefreshEnabled {
			if parseErr == errNoTemplate || parseErr == errNoData || parseErr == errNoOutput {
				panic("could not get one of: template, data or output from config.json")
			}
			watcher, err := fsnotify.NewWatcher()
			staticWatcher, err := fsnotify.NewWatcher()

			if err != nil {
				fmt.Println("could not open a file watcher: ", err)
			}
			err = filepath.Walk(config.Folders.Data, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})
			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}

			err = filepath.Walk(config.Folders.Templates, func(path string, file os.FileInfo, err error) error {
				return watcher.Add(path)
			})
			if err != nil {
				fmt.Println("failled walking over all folders: ", err)
			}

			err = filepath.Walk(config.Folders.Static, func(path string, file os.FileInfo, err error) error {
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
						info, err := os.Lstat(config.Folders.Static)
						if err != nil {
							fmt.Println("Couldnt move files form static to out: ", err.Error())
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
	}
}

func startParse(configLocation string) (*config, error) {

	config, configErr := parseConfig(configLocation)
	if configErr != nil {
		fmt.Println(configErr.Error())
		return config, configErr
	}

	//var outb, inb bytes.Buffer

	config.moduleHost = host.New()
	module := exec.Command("abm_arithmetic")

	stdin, err := module.StdinPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}
	stdout, err := module.StdoutPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}
	module.Stderr = os.Stderr

	if err := module.Start(); err != nil {
		fmt.Println("process failled")
		return nil, err
	}

	err = config.moduleHost.Start(stdout, stdin)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(config.moduleHost.AskMethods())

	templateErr := executeTemplate(config)
	if templateErr != nil {
		fmt.Println("failed building templates: ", templateErr.Error())
		return config, templateErr
	}

	return config, nil
}

func loadModules(config *config) {
	config.moduleHost = host.New()

	for identifier, version := range config.Modules.Dependencies {
		fmt.Printf("Loading module: %s@%s\n", identifier, version)

		module := exec.Command(filepath.Join(config.Folders.Modules, "abm_"+identifier))

		stdin, err := module.StdinPipe()
		if nil != err {
			log.Fatalf("Error obtaining stdin: %s", err.Error())
		}

		/*stdout, err := module.StdoutPipe()
		if nil != err {
			log.Fatalf("Error obtaining stdout: %s", err.Error())
		}*/

		if err := module.Start(); err != nil {
			panic(err)
		}

		err = config.moduleHost.Start(os.Stdout, stdin)
		if err != nil {
			panic(err)
		}

		fmt.Println(config.moduleHost.AskMethods())
	}
}

func parseConfig(configLocation string) (*config, error) {
	var config config

	configFile, err := os.Open(configLocation)
	defer configFile.Close()
	if err != nil {
		return nil, errors.New("could not open the config file: " + err.Error())
	}

	dec := json.NewDecoder(configFile)
	err = dec.Decode(&config)
	if err != nil {
		return &config, err
	}

	if config.Folders.Templates == "" {
		return &config, errNoTemplate
	}

	if config.Folders.Data == "" {
		return &config, errNoData
	}

	if config.Folders.Output == "" {
		return &config, errNoOutput
	}

	return &config, nil
}

func executeTemplate(config *config) (err error) {
	if config.Folders.Output == "" {
		err = os.RemoveAll(config.Folders.Output)
	}

	if err != nil {
		fmt.Println("could not remove files: ", err, ". Old html will be left in place")
	}

	fmt.Println("------ START ------")

	err = config.Pages.execute(nil, config)
	if err != nil {
		fmt.Println("failled parsing config file")
	}

	return
}

func (s *site) execute(parent *site, config *config) error {
	fmt.Println("new site")
	if parent != nil {
		if s.Data != nil {
			s.Data = append(parent.Data, s.Data...)
		} else {
			s.Data = make([]string, len(parent.Data))
			copy(s.Data, parent.Data)
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

		if parent.DefaultLanguage != "" {
			s.DefaultLanguage = parent.DefaultLanguage
		}
	}

	if config.Folders.Static != "" && config.Folders.Output != "" {
		fmt.Println("copying static files")
		info, err := os.Lstat(config.Folders.Static)
		if err != nil {
			return err
		}
		genCopy(config.Folders.Static, config.Folders.Output, info)
	}

	for jIndex, jsonfile := range s.Data {
		if strings.Contains(jsonfile, "*") {
			return parseStar(s, config, jIndex)
		}
	}
	fmt.Println(len(s.Sites))

	if len(s.Sites) != 0 {
		for _, site := range s.Sites {
			err := site.execute(s, config)
			if err != nil {
				return err
			}
		}
		return nil
	}
	fmt.Println(len(s.Languages))
	if len(s.Languages) != 0 && len(s.Sites) == 0 && s.language == "" {
		for _, lang := range s.Languages {
			site := s.copy()
			site.language = lang
			err := site.execute(nil, config)
			if err != nil {
				return fmt.Errorf("could not execute %s the for lang %s: %s", site.Slug, lang, err)
			}
		}
		return nil
	}

	var jsonDataFile jsonDataFile

	err := s.gatherJSON(&jsonDataFile, config)
	if err != nil {
		return err
	}
	template, err := s.gatherTemplates(config)
	if err != nil {
		return err
	}
	err = s.executeTemplate(template, jsonDataFile, config)
	if err != nil {
		return err
	}

	return nil
}

func parseStar(s *site, config *config, jIndex int) error {
	jsonPath := filepath.Dir(filepath.Join(config.Folders.Data, s.Data[jIndex]))
	jsonFile := strings.Replace(filepath.Base(s.Data[jIndex]), "*", "([^/]*)", -1)
	re := regexp.MustCompile(jsonFile)

	var matches [][][]string

	err := filepath.Walk(jsonPath, func(path string, file os.FileInfo, err error) error {
		if path == jsonPath {
			return nil
		}

		if file.IsDir() {
			return filepath.SkipDir
		}

		if ok, _ := regexp.MatchString(jsonFile, file.Name()); ok {
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
			site.Data[jIndex] = strings.Replace(site.Data[jIndex], "*", match[1], 1)
		}
		err := site.execute(nil, config)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *site) gatherJSON(jsonImput *jsonDataFile, config *config) error {
	fmt.Println("gathering JSON files for: ", s.Slug)

	for _, jsonLocation := range s.Data {
		jsonPath := filepath.Join(config.Folders.Data, jsonLocation)

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

func (s *site) gatherTemplates(config *config) (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(config.Folders.Templates, s.Templates[i])
	}

	template, err := template.New("").Funcs(fn).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func (s *site) executeTemplate(template *template.Template, jsonImput jsonDataFile, config *config) error {
	OUTPath := filepath.Join(config.Folders.Output, s.Slug)
	if s.language != "" && s.DefaultLanguage != s.language {
		OUTPath = filepath.Join(config.Folders.Output, s.language, s.Slug)
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
	newSite.Data = make([]string, len(s.Data))
	copy(newSite.Data, s.Data)

	newSite.Templates = make([]string, len(s.Templates))
	copy(newSite.Templates, s.Templates)

	return newSite
}

func (ji *jsonDataFile) UnmarshalJSON(data []byte) error {
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

func noescape(str string) template.HTML {
	return template.HTML(str)
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
