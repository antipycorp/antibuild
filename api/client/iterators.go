// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"gitlab.com/antipy/antibuild/cli/api/errors"
	"gitlab.com/antipy/antibuild/cli/api/file"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
)

/*
	ITERATORS
*/

type (
	//ITRequest is the request with data and meta from the module caller.
	ITRequest struct {
		Variable string
	}

	//ITResponse is the response to the module API that will be used to respond to the client
	ITResponse struct {
		Log  []errors.Error
		file file.File
	}

	//ITTest is a object that is used to test an iterator.
	ITTest struct {
		Request  ITRequest
		Response *ITResponse
	}
	//Iterator is a function that loads files
	Iterator func(ITRequest, Response)

	//iterator is a object that stores a Iterator and tests.
	iterator struct {
		Execute Iterator
	}
)

func iteratorsHandle(command string, r protocol.Token, m *Module) {
	if m.iterators[command].Execute == nil {
		r.Respond(nil, ErrInvalidCommand)
		return
	}

	var fileName string
	var ok bool
	if fileName, ok = r.Data[0].(string); !ok {
		r.Respond(nil, ErrInvalidInput)
		return
	}

	f, err := file.Import(fileName)
	defer f.Close()

	if err != nil {
		r.Respond(nil, ErrFailed)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); !ok {
		r.Respond(nil, ErrInvalidInput)
		return
	}

	var request = ITRequest{
		Variable: variable,
	}

	var response = &ITResponse{
		file: f,
	}

	m.iterators[command].Execute(request, response)

	if r.Respond(nil, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//IteratorRegister registers a new iterator with specified identifier to the module.
func (m *Module) IteratorRegister(identifer string, function Iterator) {
	iteratorRegister(m, identifer, function)
	return
}

func iteratorRegister(m *Module, identifer string, function Iterator) {
	if identifer == "" {
		panic("module: iteratorRegister: identifer is not defined")
	}

	if m.iterators == nil {
		panic("module: iteratorRegister: initalization of module was not correct")
	}

	if _, ok := m.iterators[identifer]; ok {
		panic("module: iteratorRegister: iterator with this identifier is already registered")

	}

	if function == nil {
		panic("module: iteratorRegister: function is not defined")
	}

	m.iterators[identifer] = iterator{
		Execute: function,
	}
}

//AddDebug adds a debug message to the log
func (it *ITResponse) AddDebug(message string) {
	it.Log = append(it.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (it *ITResponse) AddInfo(message string) {
	it.Log = append(it.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (it *ITResponse) AddError(message string) {
	it.Log = append(it.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (it *ITResponse) AddFatal(message string) {
	it.Log = append(it.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (it *ITResponse) AddData(data interface{}) bool {
	if _, ok := data.([]string); !ok {
		it.AddFatal("return data is not valid")

		return false
	}

	err := it.file.Update(data)
	if err != nil {
		it.AddFatal("failed to update file: " + err.Error())
		return false
	}
	return true
}
