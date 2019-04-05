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
	FILE PARSERS
*/

type (

	//DPRequest is the request with data and meta from the module caller.
	DPRequest struct {
		Data     []byte
		Variable string
	}

	//DPResponse is the response to the module API that will be used to respond to the client
	DPResponse struct {
		Log  []errors.Error
		file file.File
	}

	//DPTest is a object that is used to test a data parser.
	DPTest struct {
		Request  DPRequest
		Response *DPResponse
	}

	//DataParser is a function that is able to parse data
	DataParser func(DPRequest, Response)

	//dataParser is a object that stores a DataParser and its tests.
	dataParser struct {
		Execute DataParser
	}
)

func dataParsersHandle(command string, r protocol.Token, m *Module) {
	if m.dataParsers[command].Execute == nil {
		r.Respond(nil, ErrInvalidCommand)
		return
	}

	var fileInput []byte

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
	err = f.Retreive(&fileInput)
	if err != nil {
		r.Respond(nil, ErrFailed)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); !ok {
		r.Respond(nil, ErrInvalidInput)
		return
	}

	var request = DPRequest{
		Data:     fileInput,
		Variable: variable,
	}

	var response = &DPResponse{
		file: f,
	}

	m.dataParsers[command].Execute(request, response)

	if r.Respond(nil, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//DataParserRegister registers a new file parser with specified identifier to the module.
func (m *Module) DataParserRegister(identifer string, function DataParser) {
	dataParserRegister(m, identifer, function)
	return
}

func dataParserRegister(m *Module, identifer string, function DataParser) {
	if identifer == "" {
		panic("module: dataParserRegister: identifer is not defined")
	}

	if m.dataParsers == nil {
		panic("module: dataParserRegister: initalization of module was not correct")
	}

	if _, ok := m.dataParsers[identifer]; ok {
		panic("module: dataParserRegister: dataParser with this identifier is already registered")
	}

	if function == nil {
		panic("module: dataParserRegister: function is not defined")
	}

	m.dataParsers[identifer] = dataParser{
		Execute: function,
	}
}

//AddDebug adds a debug message to the log
func (dp *DPResponse) AddDebug(message string) {
	dp.Log = append(dp.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (dp *DPResponse) AddInfo(message string) {
	dp.Log = append(dp.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (dp *DPResponse) AddError(message string) {
	dp.Log = append(dp.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (dp *DPResponse) AddFatal(message string) {
	dp.Log = append(dp.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (dp *DPResponse) AddData(data interface{}) bool {
	var d tt.Data
	var ok bool
	if d, ok = data.(map[interface{}]interface{}); !ok {
		dp.AddFatal("return data is not valid")
		return false
	}

	err := dp.file.Update(d)
	if err != nil {
		dp.AddFatal("failed to update file: " + err.Error())
		return false
	}

	return true
}
