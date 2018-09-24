package module

import (
	"errors"
	"strings"

	"gitlab.com/antipy/antibuild/module/protocol"
)

type (
	//Module is the collection of registered events that the module API should react to.
	Module struct {
		name string

		templateFunctions map[string]func(Request, *Response)
	}

	//Request is the request with data and meta from the module caller.
	Request struct {
		Data []interface{}
	}

	//Response is the response to the module API that will be used to respond to the client
	Response struct {
		Error error
		Data  interface{}
	}
)

var (
	//ErrInvalidCommand is the error that occurs when a function is called that is not registered with the module.
	ErrInvalidCommand = errors.New("module: the provided command does not exist")

	//ErrInvalidInput is the error that occurs when a function is called with data that is not correct, valid or applicable.
	ErrInvalidInput = errors.New("module: the provided data is invalid")

	//ErrFailed is the error that occurs when the module experiences an internal error.
	ErrFailed = errors.New("module: internal processing error")
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

	module.templateFunctions = make(map[string]func(Request, *Response))

	return module
}

/*
	Module Start
*/

//Start listenes to messages from host and responds if possible. Should be called AFTER registering all functions to the module.
func (m *Module) Start() {
	start(m)
}

func start(m *Module) {
	//protocol.Init(false)

	for {
		r := protocol.Receive()

		commandSplit := strings.SplitN(r.Command, "_", 1)

		switch commandSplit[0] {
		case "internal":
			internalHandle(commandSplit[1], r, m)
		case "templateFunctions":
			templateFunctionsHandle(commandSplit[1], r, m)
		}
	}
}

func internalHandle(command string, r protocol.Token, m *Module) {
	switch command {
	case "getTemplateFunctions":
		var functions = make([]string, len(m.templateFunctions))

		for key := range m.templateFunctions {
			functions = append(functions, key)
		}

		r.Respond(protocol.Methods{
			"templateFunctions": functions,
		})
	}
}

func templateFunctionsHandle(command string, r protocol.Token, m *Module) {
	if m.templateFunctions[command] == nil {
		r.Respond(ErrInvalidCommand)
		return
	}

	var request = Request{
		Data: r.Data,
	}
	var response = &Response{}

	m.templateFunctions[command](request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(r.Data)
}

/*
	Module Part Registration Functions
*/

//TemplateFunctionRegister registers a new template function with identifier "identifier" to the module.
func (m *Module) TemplateFunctionRegister(identifer string, function templateFunction) {
	templateFunctionRegister(m, identifer, function)
	return
}

func templateFunctionRegister(m *Module, identifer string, function templateFunction) {
	if identifer == "" {
		panic("module: templateFunctionRegister: identifer is not defined")
	}

	if m.templateFunctions == nil {
		panic("module: templateFunctionRegister: initalization of module was not correct")
	}

	if m.templateFunctions[identifer] != nil {
		panic("module: templateFunctionRegister: templateFunction with this identifier is already registered")
	}

	if function == nil {
		panic("module: templateFunctionRegister: function is not defined")
	}

	m.templateFunctions[identifer] = function

	return
}
