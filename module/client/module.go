// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"errors"
	"os"
	"strings"

	"gitlab.com/antipy/antibuild/module/protocol"
)

type (
	//Module is the collection of registered events that the module API should react to.
	Module struct {
		name string

		templateFunctions map[string]TemplateFunction
	}

	//TFRequest is the request with data and meta from the module caller.
	TFRequest struct {
		Data interface{}
	}

	//TFResponse is the response to the module API that will be used to respond to the client
	TFResponse struct {
		Error error
		Data  interface{}
	}

	//TFTest is a object that is used to test a function.
	TFTest struct {
		Request  TFRequest
		Response *TFResponse
	}

	//TemplateFunction is a object that stores the template function and its tests.
	TemplateFunction struct {
		Function func(TFRequest, *TFResponse)
		Test     *TFTest
	}
)

var (
	//ErrInvalidCommand is the error that occurs when a function is called that is not registered with the module.
	ErrInvalidCommand = errors.New("module: the provided command does not exist")

	//ErrInvalidInput is the error that occurs when a function is called with data that is not correct, valid or applicable.
	ErrInvalidInput = errors.New("module: the provided data is invalid")

	//ErrFailed is the error that occurs when the module experiences an internal error.
	ErrFailed = errors.New("module: internal processing error")

	con *protocol.Connection
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

	module.templateFunctions = make(map[string]TemplateFunction)

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
	con = protocol.OpenConnection(os.Stdin, os.Stdout)
	con.Init(false)

	for {
		r := con.Receive()

		commandSplit := strings.SplitN(r.Command, "_", 2)
		//json.NewEncoder(os.Stderr).Encode(commandSplit)

		switch commandSplit[0] {
		case "internal":
			//json.NewEncoder(os.Stderr).Encode(r)
			internalHandle(commandSplit[1], r, m)
		case "templateFunctions":
			templateFunctionsHandle(commandSplit[1], r, m)
		}
	}
}

func internalHandle(command string, r protocol.Token, m *Module) {
	//fmt.Fprintf(os.Stderr, "internal handle!")
	switch command {
	case "getMethods":
		var functions = make([]string, len(m.templateFunctions))

		for key := range m.templateFunctions {
			functions = append(functions, key)
		}

		r.Respond(protocol.Methods{
			"templateFunctions": functions,
		})

	case "testTemplateFunctions":
		r.Respond(testTemplateFunctions(m))
	}
}

func templateFunctionsHandle(command string, r protocol.Token, m *Module) {
	if m.templateFunctions[command].Function == nil {
		r.Respond(ErrInvalidCommand)
		return
	}

	var request = TFRequest{
		Data: r.Data,
	}

	var response = &TFResponse{}

	m.templateFunctions[command].Function(request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(r.Data)
}

func testMethods(m *Module) bool {
	for _, templateFunction := range m.templateFunctions {
		var response = &TFResponse{}

		templateFunction.Function(templateFunction.Test.Request, response)

		if response.Error != nil {
			return false
		}

		if response.Data != templateFunction.Test.Response.Data {
			return false
		}
	}

	return true
}

/*
	Module Part Registration Functions
*/

//TemplateFunctionRegister registers a new template function with identifier "identifier" to the module.
func (m *Module) TemplateFunctionRegister(identifer string, function func(TFRequest, *TFResponse), test *TFTest) {
	templateFunctionRegister(m, identifer, function, test)
	return
}

func templateFunctionRegister(m *Module, identifer string, function func(TFRequest, *TFResponse), test *TFTest) {
	if identifer == "" {
		panic("module: templateFunctionRegister: identifer is not defined")
	}

	if m.templateFunctions == nil {
		panic("module: templateFunctionRegister: initalization of module was not correct")
	}

	if m.templateFunctions[identifer].Function != nil {
		panic("module: templateFunctionRegister: templateFunction with this identifier is already registered")
	}

	if function == nil {
		panic("module: templateFunctionRegister: function is not defined")
	}

	if test == nil {
		panic("module: templateFunctionRegister: test is not defined")
	}

	if test.Request.Data == nil {
		panic("module: templateFunctionRegister: test request data is not defined")
	}

	if test.Response.Data == nil {
		panic("module: templateFunctionRegister: test response data is not defined")
	}

	m.templateFunctions[identifer] = TemplateFunction{
		Function: function,
		Test:     test,
	}

	return
}
