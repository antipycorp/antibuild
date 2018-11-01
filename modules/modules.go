package modules

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	modFile "gitlab.com/antipy/antibuild/cli/modules/file"
	modJSON "gitlab.com/antipy/antibuild/cli/modules/json"
	modLang "gitlab.com/antipy/antibuild/cli/modules/language"
	modNoESC "gitlab.com/antipy/antibuild/cli/modules/noescape"
)

type (
	fileLoader struct {
		host    *host.ModuleHost
		command string
	}

	fileParser struct {
		host    *host.ModuleHost
		command string
	}

	filePostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	sitePostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	templateFunction struct {
		host    *host.ModuleHost
		command string
	}

	internalMod struct {
		version string
		start   func(io.Reader, io.Writer)
		name    string
	}

	ModuleConfig struct {
		Config map[string]interface{}
	}
)

var (
	loadedModules = false

	templateFunctions = site.TemplateFunctions

	fileLoaders        = &site.FileLoaders
	fileParsers        = &site.FileParsers
	filePostProcessors = &site.FilePostProcessors
	sitePostProcessors = &site.SitePostProcessors

	internalMods = map[string]internalMod{
		"file": internalMod{
			version: "0.0.1",
			name:    "file",
			start:   modFile.Start,
		},
		"json": internalMod{
			version: "0.0.1",
			name:    "json",
			start:   modJSON.Start,
		},
		"language": internalMod{
			version: "0.0.2",
			name:    "language",
			start:   modLang.Start,
		},
		"noescape": internalMod{
			version: "0.0.1",
			name:    "noescape",
			start:   modNoESC.Start,
		},
	}
)

//LoadModules communicates with modules to load them
func LoadModules(moduleRoot string, deps map[string]string, configs map[string]ModuleConfig) (moduleHost map[string]*host.ModuleHost) {
	if loadedModules { //TODO doesnt support hotloadding modules
		return nil
	}

	moduleHost = make(map[string]*host.ModuleHost, len(deps))

	for identifier, version := range deps {
		fmt.Printf("Loading module: %s@%s\n", identifier, version)
		stdout, stdin := loadModule(identifier, version, moduleRoot)
		var err error
		moduleHost[identifier], err = host.Start(stdout, stdin)
		if err != nil {
			panic(err)
		}

		methods, err := moduleHost[identifier].AskMethods()
		if err != nil {
			panic(err)
		}

		//registers all functions modules can possibly suply
		for _, function := range methods["templateFunctions"] {
			templateFunctions[identifier+"_"+function] = getTemplateFunction(function, moduleHost[identifier])
		}

		for _, function := range methods["fileLoaders"] {
			(*fileLoaders)[identifier+"_"+function] = getFileLoader(function, moduleHost[identifier])
		}

		for _, function := range methods["fileParsers"] {
			(*fileParsers)[identifier+"_"+function] = getFileParser(function, moduleHost[identifier])
		}

		for _, function := range methods["filePostProcessors"] {
			(*filePostProcessors)[identifier+"_"+function] = getFilePostProcessor(function, moduleHost[identifier])
		}

		for _, function := range methods["sitePostProcessors"] {
			(*sitePostProcessors)[identifier+"_"+function] = getSitePostProcessor(function, moduleHost[identifier])
		}

		if configs[identifier].Config != nil {
			output, err := moduleHost[identifier].ExcecuteMethod("internal_config", []interface{}{
				configs[identifier].Config,
			})
			if err != nil || output != "module: ready" {
				panic("couldnt send config: " + err.Error())
			}
		}
	}
	return
}

func loadModule(name, version, path string) (io.Reader, io.Writer) {
	fmt.Printf("Loading module: %s@%s\n", name, version)

	if v, ok := internalMods[name]; ok {
		if v.version == version {
			//var in, stdout io.Reader
			//var out, stdin io.Writer

			in, stdin := io.Pipe()
			stdout, out := io.Pipe()
			in2 := bufio.NewReader(in)
			stdout2 := bufio.NewReader(stdout)

			go v.start(in2, out)

			return stdout2, stdin
		}
	}

	//prepare command and get nesecary data
	module := exec.Command(filepath.Join(path, "abm_"+name))

	stdin, err := module.StdinPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	stdout, err := module.StdoutPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}

	module.Stderr = os.Stderr

	//start module and initaite connection
	if err := module.Start(); err != nil {
		panic(err)
	}
	return stdout, stdin
}

func getTemplateFunction(command string, host *host.ModuleHost) *templateFunction {
	return &templateFunction{
		host:    host,
		command: command,
	}
}

func (f *templateFunction) Load(data ...interface{}) []byte {
	output, err := f.host.ExcecuteMethod("templateFunctions_"+f.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []byte
	var ok bool
	if outputFinal, ok = output.([]byte); ok != true {
		panic("fileLoader_" + f.command + " did not return a []byte")
	}

	return outputFinal
}

func getFileLoader(command string, host *host.ModuleHost) *fileLoader {
	return &fileLoader{
		host:    host,
		command: command,
	}
}

func (f *fileLoader) Load(variable string) []byte {
	data := []interface{}{
		variable,
	}

	output, err := f.host.ExcecuteMethod("fileLoaders_"+f.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []byte
	var ok bool
	if outputFinal, ok = output.([]byte); ok != true {
		panic("fileLoader_" + f.command + " did not return a []byte")
	}

	return outputFinal
}

func getFileParser(command string, host *host.ModuleHost) *fileParser {
	return &fileParser{
		host:    host,
		command: command,
	}
}

func (f *fileParser) Parse(data []byte, variable string) map[string]interface{} {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := f.host.ExcecuteMethod("fileParsers_"+f.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal map[string]interface{}
	var ok bool
	if outputFinal, ok = output.(map[string]interface{}); ok != true {
		panic("fileParser_" + f.command + " did not return a map[string]interface{}")
	}

	return outputFinal
}

func getFilePostProcessor(command string, host *host.ModuleHost) *filePostProcessor {
	return &filePostProcessor{
		host:    host,
		command: command,
	}
}

func (f *filePostProcessor) Process(data map[string]interface{}, variable string) map[string]interface{} {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := f.host.ExcecuteMethod("filePostProcessors_"+f.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal map[string]interface{}
	var ok bool
	if outputFinal, ok = output.(map[string]interface{}); ok != true {
		panic("filePostProcessors_" + f.command + " did not return a map[string]interface{}")
	}

	return outputFinal
}

func getSitePostProcessor(command string, host *host.ModuleHost) sitePostProcessor {
	return sitePostProcessor{
		host:    host,
		command: command,
	}
}

func (s sitePostProcessor) Process(data []*site.Site, variable string) []*site.Site {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := s.host.ExcecuteMethod("sitePostProcessors_"+s.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []*site.Site
	var ok bool
	if outputFinal, ok = output.([]*site.Site); ok != true {
		panic("sitePostProcessors_" + s.command + " did not return a []*site.Site")
	}

	return outputFinal
}

func (mc *ModuleConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &mc.Config); err != nil {
		return err
	}
	return nil
}

func (mc *ModuleConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(mc.Config)
}
