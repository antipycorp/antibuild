// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"io"
	"os"
	"strings"

	"gitlab.com/antipy/antibuild/cli/api/errors"
	//initiates nessecary GOB registers
	_ "gitlab.com/antipy/antibuild/cli/api/internal"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
)

type (
	//Module is the collection of registered events that the module API should react to.
	Module struct {
		name string

		configFunction func(map[string]interface{}) *errors.Error

		templateFunctions  map[string]templateFunction
		dataLoaders        map[string]dataLoader
		dataParsers        map[string]dataParser
		dataPostProcessors map[string]dataPostProcessor
		sitePostProcessors map[string]sitePostProcessor
		iterators          map[string]iterator
	}
	//Response is a response to a request
	Response interface {
		AddInfo(message string)
		AddError(message string)
		AddFatal(message string)
		AddData(data interface{}) bool
	}
)

const (
	//keys in the config map
	keyTemplateFunctions  = "templateFunctions"
	keyDataLoaders        = "dataLoaders"
	keyDataParsers        = "dataParsers"
	keyDataPostProcessors = "dataPostProcessors"
	keySitePostProcessors = "sitePostProcessors"
	keyIterators          = "iterators"

	//InvalidCommand is the errors that occurs when a function is called that is not registered with the module.
	InvalidCommand = "module: the provided command does not exist"

	//InvalidInput is the errors that occurs when a function is called with data that is not correct, valid or applicable.
	InvalidInput = "module: the provided data is invalid"

	//Failed is the errors that occurs when the module experiences an internal errors.
	Failed = "module: internal processing errors"

	//NoConfig is the errors that occurs when the module has not recieved its configuration when needed.
	NoConfig = "module: config has not been recieved yet"

	//ModuleReady tells the host that the config worked
	ModuleReady = "module: ready"
)

var (
	//ErrInvalidCommand is the errors that occurs when a function is called that is not registered with the module.
	ErrInvalidCommand = errors.New(InvalidCommand, 1)

	//ErrInvalidInput is the errors that occurs when a function is called with data that is not correct, valid or applicable.
	ErrInvalidInput = errors.New(InvalidInput, 1)

	//ErrFailed is the errors that occurs when the module experiences an internal errors.
	ErrFailed = errors.New(Failed, 1)

	//ErrNoConfig is the errors that occurs when the module has not recieved its configuration when needed.
	ErrNoConfig = errors.New(NoConfig, 1)
)

/*
	Module Management Functions
*/

//Register registers a new module with its meta information.
func Register(name string) (module *Module) {
	return register(name)
}

func register(name string) *Module {
	module := new(Module)

	if name == "" {
		panic("module: name is not defined")
	}

	module.name = name

	module.templateFunctions = make(map[string]templateFunction)
	module.dataLoaders = make(map[string]dataLoader)
	module.dataParsers = make(map[string]dataParser)
	module.dataPostProcessors = make(map[string]dataPostProcessor)
	module.sitePostProcessors = make(map[string]sitePostProcessor)
	module.iterators = make(map[string]iterator)

	return module
}

/*
	Module Start
*/

//Start listenes to messages from host and responds if possible. Should be called AFTER registering all functions to the module.
func (m *Module) Start() {
	start(m, os.Stdin, os.Stdout)
}

//CustomStart just like Start but with custom read adn writers
func (m *Module) CustomStart(in io.Reader, out io.Writer) {
	start(m, in, out)
}

func start(m *Module, in io.Reader, out io.Writer) {
	con := protocol.OpenConnection(in, out)
	con.ID = m.name
	con.Init(false)
	for {
		r := con.Receive()
		if len(r.Data) == 1 {
			if r.Data[0] == protocol.ErrEOF {
				os.Exit(1)
			}
		}
		if r.Command == protocol.KillCommand {
			return
		} else if r.Command == protocol.GetMethods {
			getMethodsHandle(r, m)
		}
		commandSplit := strings.SplitN(r.Command, "_", 2)

		switch commandSplit[0] {
		case "internal":
			internalHandle(commandSplit[1], r, m)
		case keyTemplateFunctions:
			templateFunctionsHandle(commandSplit[1], r, m)
		case keyDataLoaders:
			dataLoadersHandle(commandSplit[1], r, m)
		case keyDataParsers:
			dataParsersHandle(commandSplit[1], r, m)
		case keyDataPostProcessors:
			dataPostProcessorsHandle(commandSplit[1], r, m)
		case keySitePostProcessors:
			sitePostProcessorsHandle(commandSplit[1], r, m)
		case keyIterators:
			iteratorsHandle(commandSplit[1], r, m)
		}
	}
}

func getMethodsHandle(r protocol.Token, m *Module) {
	var (
		templateFunctions  = make([]string, len(m.templateFunctions))
		dataLoaders        = make([]string, len(m.dataLoaders))
		dataParsers        = make([]string, len(m.dataParsers))
		dataPostProcessors = make([]string, len(m.dataPostProcessors))
		sitePostProcessors = make([]string, len(m.sitePostProcessors))
		iterators          = make([]string, len(m.iterators))
	)

	for key := range m.templateFunctions {
		templateFunctions = append(templateFunctions, key)
	}

	for key := range m.dataLoaders {
		dataLoaders = append(dataLoaders, key)
	}

	for key := range m.dataParsers {
		dataParsers = append(dataParsers, key)
	}

	for key := range m.dataPostProcessors {
		dataPostProcessors = append(dataPostProcessors, key)
	}

	for key := range m.sitePostProcessors {
		sitePostProcessors = append(sitePostProcessors, key)
	}

	for key := range m.iterators {
		iterators = append(iterators, key)
	}

	r.Respond(protocol.Methods{
		keyTemplateFunctions:  templateFunctions,
		keyDataLoaders:        dataLoaders,
		keyDataParsers:        dataParsers,
		keyDataPostProcessors: dataPostProcessors,
		keySitePostProcessors: sitePostProcessors,
		keyIterators:          iterators,
	})
}

func internalHandle(command string, r protocol.Token, m *Module) {
	switch command {
	case "config":
		if m.configFunction == nil {
			r.Respond(ModuleReady)
			return
		}

		var ok bool

		var objectInput map[string]interface{}
		if objectInput, ok = r.Data[0].(map[string]interface{}); ok != true {
			r.Respond(ErrInvalidInput)
			return
		}

		err := m.configFunction(objectInput)
		if err != nil {
			r.Respond(err)
			return
		}

		r.Respond(ModuleReady)
	case "testMethods":
		r.Respond(testMethods(m))
	}
}

func testMethods(m *Module) bool {
	for _, templateFunction := range m.templateFunctions {
		var response = &TFResponse{}

		templateFunction.Execute(templateFunction.Test.Request, response)

		for _, log := range response.Log {
			if log.Code == errors.CodeFatal {
				return false
			}
		}

		if response.Data != templateFunction.Test.Response.Data {
			return false
		}
	}

	return true
}

//ConfigFunctionRegister is the function that will be called that handles the config of the client.
func (m *Module) ConfigFunctionRegister(function func(map[string]interface{}) *errors.Error) {
	configFunctionRegister(m, function)
}

func configFunctionRegister(m *Module, function func(map[string]interface{}) *errors.Error) {
	if function == nil {
		panic("module: configFunctionRegister: function is not defined")
	}

	m.configFunction = function
}
