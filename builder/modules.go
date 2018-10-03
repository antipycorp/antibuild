package builder

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/module/host"
)

var (
	loadedModules = false

	templateFunctions  = template.FuncMap{}
	fileLoaders        = make(map[string]func(variable string) []byte)
	fileParsers        = make(map[string]func(data []byte, variable string) map[string]interface{})
	filePostProcessors = make(map[string]func(data map[string]interface{}, variable string) map[string]interface{})
)

//communicates with modules to load them
func loadModules(config *config) {
	//make a refrence to keep all module (data) in
	config.moduleHost = make(map[string]*host.ModuleHost, len(config.Modules.Dependencies))

	//loop over the modules
	for identifier, version := range config.Modules.Dependencies {
		fmt.Printf("Loading module: %s@%s\n", identifier, version)

		//prepare to open the module
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

		//start the module
		if err := module.Start(); err != nil {
			panic(err)
		}

		//start the host for the module (does version check ect)
		config.moduleHost[identifier], err = host.Start(stdout, stdin)
		if err != nil {
			panic(err)
		}

		//gets all the things the module can do
		methods, err := config.moduleHost[identifier].AskMethods()
		if err != nil {
			panic(err)
		}

		//registers all templateFunctions
		for _, function := range methods["templateFunctions"] {
			templateFunctions[identifier+"_"+function] = moduleTemplateFunctionDefinition(identifier, function, config)
		}

		//register all fileLoaders
		for _, function := range methods["fileLoaders"] {
			fileLoaders[identifier+"_"+function] = moduleFileLoaderDefinition(identifier, function, config)
		}

		//register all file parsers
		for _, function := range methods["fileParsers"] {
			fileParsers[identifier+"_"+function] = moduleFileParserDefinition(identifier, function, config)
		}

		//register all filePostProcessors
		for _, function := range methods["filePostProcessors"] {
			filePostProcessors[identifier+"_"+function] = moduleFilePostProcessorDefinition(identifier, function, config)
		}
	}
}

//generates the function that is called when a template function in executed.
func moduleTemplateFunctionDefinition(module string, command string, config *config) func(data ...interface{}) interface{} {
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

//generates the function that is called when a file loader in executed.
func moduleFileLoaderDefinition(module string, command string, config *config) func(variable string) []byte {
	return func(variable string) []byte {
		//make an array to send to the client
		data := []interface{}{
			variable,
		}

		//send the data to the module and wait for response
		output, err := config.moduleHost[module].ExcecuteMethod("fileLoaders_"+command, data)
		if err != nil {
			panic("execute methods: " + err.Error())
		}

		//check if return type is correct
		var outputFinal []byte
		var ok bool
		if outputFinal, ok = output.([]byte); ok != true {
			panic("fileLoader_" + command + " did not return a []byte")
		}

		//return data
		return outputFinal
	}
}

//generates the function that is called when a file parser in executed.
func moduleFileParserDefinition(module string, command string, config *config) func(data []byte, variable string) map[string]interface{} {
	return func(data []byte, variable string) map[string]interface{} {
		//make an array to send to the client
		sendData := []interface{}{
			data,
			variable,
		}

		//send the data to the module and wait for response
		output, err := config.moduleHost[module].ExcecuteMethod("fileParsers_"+command, sendData)
		if err != nil {
			panic("execute methods: " + err.Error())
		}

		//check if return type is correct
		var outputFinal map[string]interface{}
		var ok bool
		if outputFinal, ok = output.(map[string]interface{}); ok != true {
			panic("fileParser_" + command + " did not return a map[string]interface{}")
		}

		//return data
		return outputFinal
	}
}

//generates the function that is called when a file post processor in executed.
func moduleFilePostProcessorDefinition(module string, command string, config *config) func(data map[string]interface{}, variable string) map[string]interface{} {
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
