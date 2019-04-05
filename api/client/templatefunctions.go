// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"gitlab.com/antipy/antibuild/cli/api/errors"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
)

type (

	//TFRequest is the request with data and meta from the module caller.
	TFRequest struct {
		Data []interface{}
	}

	//TFResponse is the response to the module API that will be used to respond to the client
	TFResponse struct {
		Data interface{}
		Log  []errors.Error
	}

	//TFTest is a object that is used to test a function.
	TFTest struct {
		Request  TFRequest
		Response *TFResponse
	}

	//TemplateFunction is a function thats usable in golang html templates
	TemplateFunction func(TFRequest, Response)

	//templateFunction is a object that stores the template function and its tests.
	templateFunction struct {
		Execute TemplateFunction
		Test    *TFTest
	}
)

//TemplateFunctionRegister registers a new template function with specified identifier to the module.
func (m *Module) TemplateFunctionRegister(identifer string, function TemplateFunction, test *TFTest) {
	templateFunctionRegister(m, identifer, function, test)
	return
}

func templateFunctionRegister(m *Module, identifer string, function TemplateFunction, test *TFTest) {
	if identifer == "" {
		panic("module: templateFunctionRegister: identifer is not defined")
	}

	if _, ok := m.templateFunctions[identifer]; ok {
		panic("module: templateFunctionRegister: TemplateFunction with this identifier is already registered")
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

	m.templateFunctions[identifer] = templateFunction{
		Execute: function,
		Test:    test,
	}
}

func templateFunctionsHandle(command string, r protocol.Token, m *Module) {
	if m.templateFunctions[command].Execute == nil {
		r.Respond(nil, ErrInvalidCommand)
		return
	}

	var request = TFRequest{
		Data: r.Data,
	}

	var response = &TFResponse{}

	m.templateFunctions[command].Execute(request, response)

	if r.Respond(response.Data, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//AddDebug adds a debug message to the log
func (tf *TFResponse) AddDebug(message string) {
	tf.Log = append(tf.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (tf *TFResponse) AddInfo(message string) {
	tf.Log = append(tf.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (tf *TFResponse) AddError(message string) {
	tf.Log = append(tf.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (tf *TFResponse) AddFatal(message string) {
	tf.Log = append(tf.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (tf *TFResponse) AddData(data interface{}) bool {
	tf.Data = data
	return true
}
