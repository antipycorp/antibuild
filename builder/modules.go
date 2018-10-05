package builder

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/module/host"
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
	fileProcessor struct {
		host    *host.ModuleHost
		command string
	}
)

var (
	loadedModules = false

	templateFunctions = template.FuncMap{}
	fileLoaders       = &site.FileLoaders
	fileParsers       = &site.FileParsers
	fileProcessors    = &site.FileProcessors
)

//communicates with modules to load them
func loadModules(config *Config) {
	//make a refrence to keep all module (data) in
	config.moduleHost = make(map[string]*host.ModuleHost, len(config.Modules.Dependencies))

	for identifier, version := range config.Modules.Dependencies {
		fmt.Printf("Loading module: %s@%s\n", identifier, version)

		//prepare command and get nesecary data
		module := exec.Command(filepath.Join(config.Folders.Modules, "abm_"+identifier))

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
			(*fileProcessors)[identifier+"_"+function] = getFileProcessor(function, config.moduleHost[identifier])
		}
	}
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
func getFileProcessor(command string, host *host.ModuleHost) *fileProcessor {
	return &fileProcessor{
		host:    host,
		command: command,
	}
}

func (f *fileProcessor) Proces(data map[string]interface{}, variable string) map[string]interface{} {
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

//generates the function that is called when a file post processor in executed.
func moduleFilePostProcessorDefinition(module string, command string, config *Config) func(data map[string]interface{}, variable string) map[string]interface{} {
	return func(data map[string]interface{}, variable string) map[string]interface{} {
		//make an array to send to the client
		sendData := []interface{}{
			data,
			variable,
		}

		//send the data to the module and wait for response
		output, err := config.moduleHost[module].ExcecuteMethod("filePostProcessors_"+command, sendData)
		if err != nil {
			panic("execute methods: " + err.Error())
		}

		//check if return type is correct
		var outputFinal map[string]interface{}
		var ok bool
		if outputFinal, ok = output.(map[string]interface{}); ok != true {
			panic("filePostProcessors_" + command + " did not return a map[string]interface{}")
		}

		//return data
		return outputFinal
	}
}
