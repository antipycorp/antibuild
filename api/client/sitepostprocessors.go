// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"gitlab.com/antipy/antibuild/cli/api/errors"
	"gitlab.com/antipy/antibuild/cli/api/file"
	"gitlab.com/antipy/antibuild/cli/api/protocol"
	"gitlab.com/antipy/antibuild/cli/builder/site"
)

/*
	SITE POST PROCESSORS
*/

type (

	//SPPRequest is the request with data and meta from the module caller.
	SPPRequest struct {
		Data     []*site.Site
		Variable string
	}

	//SPPResponse is the response to the module API that will be used to respond to the client
	SPPResponse struct {
		Log  []errors.Error
		file file.File
	}
	//SitePortProcessor is a function that processes a site
	SitePortProcessor func(SPPRequest, Response)

	//sitePostProcessor is a object that stores the site post processor function and its tests.
	sitePostProcessor struct {
		Execute SitePortProcessor
	}
)

func sitePostProcessorsHandle(command string, r protocol.Token, m *Module) {
	if m.sitePostProcessors[command].Execute == nil {
		r.Respond(nil, ErrInvalidCommand)
		return
	}

	var objectInput []*site.Site

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

	err = f.Retreive(&objectInput)
	if err != nil {
		r.Respond(nil, ErrFailed)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); !ok {
		r.Respond(nil, ErrInvalidInput)
		return
	}

	var request = SPPRequest{
		Data:     objectInput,
		Variable: variable,
	}

	var response = &SPPResponse{
		file: f,
	}

	m.sitePostProcessors[command].Execute(request, response)

	if r.Respond(nil, response.Log...) != nil {
		r.Respond(nil, errors.New("failed to send data", errors.CodeInvalidResponse))
	}
}

//SitePostProcessorRegister registers a new site post processor with specified identifier to the module.
func (m *Module) SitePostProcessorRegister(identifer string, function SitePortProcessor) {
	sitePostProcessorRegister(m, identifer, function)
	return
}

func sitePostProcessorRegister(m *Module, identifer string, function SitePortProcessor) {
	if identifer == "" {
		panic("module: sitePostProcessor: identifer is not defined")
	}

	if m.sitePostProcessors == nil {
		panic("module: sitePostProcessor: initalization of module was not correct")
	}

	if _, ok := m.sitePostProcessors[identifer]; ok {
		panic("module: sitePostProcessor: sitePostProcessor with this identifier is already registered")
	}

	if function == nil {
		panic("module: sitePostProcessor: function is not defined")
	}

	m.sitePostProcessors[identifer] = sitePostProcessor{
		Execute: function,
	}

}

//AddDebug adds a debug message to the log
func (spp *SPPResponse) AddDebug(message string) {
	spp.Log = append(spp.Log, errors.New(message, errors.CodeDebug))
}

//AddInfo adds an info message to the log
func (spp *SPPResponse) AddInfo(message string) {
	spp.Log = append(spp.Log, errors.New(message, errors.CodeInfo))
}

//AddError adds an error in the log
func (spp *SPPResponse) AddError(message string) {
	spp.Log = append(spp.Log, errors.New(message, errors.CodeError))
}

//AddFatal adds a fatal error to the log
func (spp *SPPResponse) AddFatal(message string) {
	spp.Log = append(spp.Log, errors.New(message, errors.CodeFatal))
}

//AddData adds the data to the response
func (spp *SPPResponse) AddData(data interface{}) bool {
	if _, ok := data.([]*site.Site); !ok {
		spp.AddFatal("return data is not valid")
		return false
	}

	err := spp.file.Update(data)
	if err != nil {
		spp.AddFatal("failed to update file: " + err.Error())
	}
	return true
}
