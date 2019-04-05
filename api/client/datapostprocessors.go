// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/cli/api/errors"
	"gitlab.com/antipy/antibuild/cli/api/file"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
)

/*
	DATA POST PROCESSORS
*/
type (
	//DPPRequest is the request with data and meta from the module caller.
	DPPRequest struct {
		Data     map[interface{}]interface{}
		Variable string
	}

	//DPPResponse is the response to the module API that will be used to respond to the client
	DPPResponse struct {
		Data map[interface{}]interface{}
		Log  []errors.Error
		file file.File
	}

	//DPPTest is a object that is used to test a data post processor.
	DPPTest struct {
		Request  DPPRequest
		Response *DPPResponse
	}

	//DataPostProcessor is a function that processes some data
	DataPostProcessor func(DPPRequest, Response)

	//dataPostProcessor is a object that stores the data post processor function and its tests.
	dataPostProcessor struct {
		Execute DataPostProcessor
	}
)

func dataPostProcessorsHandle(command string, r protocol.Token, m *Module) {
	if m.dataPostProcessors[command].Execute == nil {
		r.Respond(nil, ErrInvalidCommand)
		return
	}

	var objectInput tt.Data

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
	f.Retreive(&objectInput)

	var variable string
	if variable, ok = r.Data[1].(string); ok != true {
		r.Respond(nil, ErrInvalidInput)
		return
	}

	var request = DPPRequest{
		Data:     map[interface{}]interface{}(objectInput),
		Variable: variable,
	}

	var response = &DPPResponse{
		file: f,
	}

	m.dataPostProcessors[command].Execute(request, response)

	if r.Respond(response.Data, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//DataPostProcessorRegister registers a new file post processor with specified identifier to the module.
func (m *Module) DataPostProcessorRegister(identifer string, function DataPostProcessor) {
	dataPostProcessorRegister(m, identifer, function)
	return
}

func dataPostProcessorRegister(m *Module, identifer string, function DataPostProcessor) {
	if identifer == "" {
		panic("module: dataPostProcessor: identifer is not defined")
	}

	if m.dataPostProcessors == nil {
		panic("module: dataPostProcessor: initalization of module was not correct")
	}

	if _, ok := m.dataPostProcessors[identifer]; ok {
		panic("module: dataPostProcessor: dataPostProcessor with this identifier is already registered")
	}

	if function == nil {
		panic("module: dataPostProcessor: function is not defined")
	}

	m.dataPostProcessors[identifer] = dataPostProcessor{
		Execute: function,
	}
}

//AddDebug adds a debug message to the log
func (dpp *DPPResponse) AddDebug(message string) {
	dpp.Log = append(dpp.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (dpp *DPPResponse) AddInfo(message string) {
	dpp.Log = append(dpp.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (dpp *DPPResponse) AddError(message string) {
	dpp.Log = append(dpp.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (dpp *DPPResponse) AddFatal(message string) {
	dpp.Log = append(dpp.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (dpp *DPPResponse) AddData(data interface{}) bool {
	var d tt.Data
	var ok bool
	if d, ok = data.(map[interface{}]interface{}); !ok {
		dpp.AddFatal("return data is not valid")
		return false
	}

	err := dpp.file.Update(d)
	if err != nil {
		dpp.AddFatal("failed to update file: " + err.Error())
		return false
	}

	return true
}
