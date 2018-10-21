package builder

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	modFile "gitlab.com/antipy/antibuild/cli/internalmods/file"
	modJSON "gitlab.com/antipy/antibuild/cli/internalmods/json"
	modLang "gitlab.com/antipy/antibuild/cli/internalmods/language"
	modNoESC "gitlab.com/antipy/antibuild/cli/internalmods/noescape"
	"gitlab.com/antipy/antibuild/api/host"
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
	internalMod struct {
		version string
		start   func(io.Reader, io.Writer)
		name    string
	}
)

var (
	loadedModules = false

	templateFunctions = site.TemplateFunctions

	fileLoaders        = &site.FileLoaders
	fileParsers        = &site.FileParsers
	filePostProcessors = &site.FilePostProcessors

	sitePostProcessors = make(map[string]sitePostProcessor)
	internalMods       = map[string]internalMod{
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
			version: "0.0.1",
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

//communicates with modules to load them
func loadModules(config *Config) {
	//make a refrence to keep all module (data) in
	config.moduleHost = make(map[string]*host.ModuleHost, len(config.Modules.Dependencies))

	for identifier, version := range config.Modules.Dependencies {
		fmt.Printf("Loading module: %s@%s\n", identifier, version)
		stdout, stdin := loadModule(identifier, version, config.Folders.Modules)
		var err error
		config.moduleHost[identifier], err = host.Start(stdout, stdin)
		if err != nil {
			panic(err)
		}

		methods, err := config.moduleHost[identifier].AskMethods()
		if err != nil {
			panic(err)
		}

		//registers all functions modules can possibly suply
		for _, function := range methods["templateFunctions"] {
			templateFunctions[identifier+"_"+function] = moduleTemplateFunctionDefinition(identifier, function, config)
		}

		for _, function := range methods["fileLoaders"] {
			(*fileLoaders)[identifier+"_"+function] = getFileLoader(function, config.moduleHost[identifier])
		}

		for _, function := range methods["fileParsers"] {
			(*fileParsers)[identifier+"_"+function] = getFileParser(function, config.moduleHost[identifier])
		}

		for _, function := range methods["filePostProcessors"] {
			(*filePostProcessors)[identifier+"_"+function] = getFilePostProcessor(function, config.moduleHost[identifier])
		}

		for _, function := range methods["sitePostProcessors"] {
			sitePostProcessors[identifier+"_"+function] = getSitePostProcessor(function, config.moduleHost[identifier])
		}

		if config.Modules.Config[identifier] != nil {
			output, err := config.moduleHost[identifier].ExcecuteMethod("internal_config", []interface{}{
				config.Modules.Config[identifier],
			})
			if err != nil || output != "module: ready" {
				panic("couldnt send config: " + err.Error())
			}
		}
	}
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

//generates the function that is called when a template function in executed.
func moduleTemplateFunctionDefinition(module string, command string, config *Config) func(data ...interface{}) interface{} {
	return func(data ...interface{}) interface{} {
		//send the data to the module and wait for response
		output, err := config.moduleHost[module].ExcecuteMethod("templateFunctions_"+command, data)
		if err != nil {
			panic("execute methods: " + err.Error())
		}

		//return the data
		return output
	}
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
