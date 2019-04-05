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
	DATA LOADERS
*/

type (
	//DLRequest is the request with data and meta from the module caller.
	DLRequest struct {
		Variable string
	}

	//DLResponse is the response to the module API that will be used to respond to the client
	DLResponse struct {
		Log  []errors.Error
		file file.File
	}

	//DLTest is a object that is used to test a file loader.
	DLTest struct {
		Request  DLRequest
		Response *DLResponse
	}
	//DataLoader is a function that loads files
	DataLoader func(DLRequest, Response)

	//dataLoader is a object that stores a DataLoader and tests.
	dataLoader struct {
		Execute DataLoader
	}
)

func dataLoadersHandle(command string, r protocol.Token, m *Module) {
	if m.dataLoaders[command].Execute == nil {
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

	var request = DLRequest{
		Variable: variable,
	}

	var response = &DLResponse{
		file: f,
	}

	m.dataLoaders[command].Execute(request, response)

	if r.Respond(nil, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//DataLoaderRegister registers a new file loader with specified identifier to the module.
func (m *Module) DataLoaderRegister(identifer string, function DataLoader) {
	dataLoaderRegister(m, identifer, function)
	return
}

func dataLoaderRegister(m *Module, identifer string, function DataLoader) {
	if identifer == "" {
		panic("module: dataLoaderRegister: identifer is not defined")
	}

	if m.dataLoaders == nil {
		panic("module: dataLoaderRegister: initalization of module was not correct")
	}

	if _, ok := m.dataLoaders[identifer]; ok {
		panic("module: dataLoaderRegister: dataLoader with this identifier is already registered")

	}

	if function == nil {
		panic("module: dataLoaderRegister: function is not defined")
	}

	m.dataLoaders[identifer] = dataLoader{
		Execute: function,
	}
}

//AddDebug adds a debug message to the log
func (dl *DLResponse) AddDebug(message string) {
	dl.Log = append(dl.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (dl *DLResponse) AddInfo(message string) {
	dl.Log = append(dl.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (dl *DLResponse) AddError(message string) {
	dl.Log = append(dl.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (dl *DLResponse) AddFatal(message string) {
	dl.Log = append(dl.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (dl *DLResponse) AddData(data interface{}) bool {
	if _, ok := data.([]byte); !ok {
		dl.AddFatal("return data is not valid")

		return false
	}

	err := dl.file.Update(data)
	if err != nil {
		dl.AddFatal("failed to update file: " + err.Error())
		return false
	}
	return true
}
