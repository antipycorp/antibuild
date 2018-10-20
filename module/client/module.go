// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package module

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/antipy/antibuild/cli/builder/site"

	//_ "gitlab.com/antipy/antibuild/cli/module/internal"
	"gitlab.com/antipy/antibuild/cli/module/protocol"
)

type (
	//Module is the collection of registered events that the module API should react to.
	Module struct {
		name string

		configFunction func(map[string]interface{}) error

		templateFunctions  map[string]TemplateFunction
		fileLoaders        map[string]FileLoader
		fileParsers        map[string]FileParser
		filePostProcessors map[string]FilePostProcessor
		sitePostProcessors map[string]SitePostProcessor
	}

	/*
		TEMPLATE FUNCTIONS
	*/

	//TFRequest is the request with data and meta from the module caller.
	TFRequest struct {
		Data []interface{}
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

	/*
		FILE LOADERS
	*/

	//FLRequest is the request with data and meta from the module caller.
	FLRequest struct {
		Variable string
	}

	//FLResponse is the response to the module API that will be used to respond to the client
	FLResponse struct {
		Error error
		Data  []byte
	}

	//FLTest is a object that is used to test a file loader.
	FLTest struct {
		Request  FLRequest
		Response *FLResponse
	}

	//FileLoader is a object that stores the file loader function and its tests.
	FileLoader struct {
		Function func(FLRequest, *FLResponse)
	}

	/*
		FILE PARSERS
	*/

	//FPRequest is the request with data and meta from the module caller.
	FPRequest struct {
		Data     []byte
		Variable string
	}

	//FPResponse is the response to the module API that will be used to respond to the client
	FPResponse struct {
		Error error
		Data  map[string]interface{}
	}

	//FPTest is a object that is used to test a file parser.
	FPTest struct {
		Request  FPRequest
		Response *FPResponse
	}

	//FileParser is a object that stores the file parser function and its tests.
	FileParser struct {
		Function func(FPRequest, *FPResponse)
	}

	/*
		FILE POST PROCESSORS
	*/

	//FPPRequest is the request with data and meta from the module caller.
	FPPRequest struct {
		Data     map[string]interface{}
		Variable string
	}

	//FPPResponse is the response to the module API that will be used to respond to the client
	FPPResponse struct {
		Error error
		Data  map[string]interface{}
	}

	//FPPTest is a object that is used to test a file post processor.
	FPPTest struct {
		Request  FPPRequest
		Response *FPPResponse
	}

	//FilePostProcessor is a object that stores the file post processor function and its tests.
	FilePostProcessor struct {
		Function func(FPPRequest, *FPPResponse)
	}

	/*
		SITE POST PROCESSORS
	*/

	//SPPRequest is the request with data and meta from the module caller.
	SPPRequest struct {
		Data     []*site.Site
		Variable string
	}

	//SPPResponse is the response to the module API that will be used to respond to the client
	SPPResponse struct {
		Error error
		Data  []*site.Site
	}

	//SitePostProcessor is a object that stores the site post processor function and its tests.
	SitePostProcessor struct {
		Function func(SPPRequest, *SPPResponse)
	}
)

var (
	//ErrInvalidCommand is the error that occurs when a function is called that is not registered with the module.
	ErrInvalidCommand = errors.New("module: the provided command does not exist")

	//ErrInvalidInput is the error that occurs when a function is called with data that is not correct, valid or applicable.
	ErrInvalidInput = errors.New("module: the provided data is invalid")

	//ErrFailed is the error that occurs when the module experiences an internal error.
	ErrFailed = errors.New("module: internal processing error")

	//ErrNoConfig is the error that occurs when the module has not recieved its configuration when needed.
	ErrNoConfig = errors.New("module: config has not been recieved yet")

	//ModuleReady tells the host that the config worked
	ModuleReady = "module: ready"
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
	module.fileLoaders = make(map[string]FileLoader)
	module.fileParsers = make(map[string]FileParser)
	module.filePostProcessors = make(map[string]FilePostProcessor)
	module.sitePostProcessors = make(map[string]SitePostProcessor)

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
			if r.Data[0] == protocol.EOF {
				fmt.Println("exiting now")
				os.Exit(1)
			}
		}
		commandSplit := strings.SplitN(r.Command, "_", 2)

		switch commandSplit[0] {
		case "internal":
			internalHandle(commandSplit[1], r, m)
		case "templateFunctions":
			templateFunctionsHandle(commandSplit[1], r, m)
		case "fileLoaders":
			fileLoadersHandle(commandSplit[1], r, m)
		case "fileParsers":
			fileParsersHandle(commandSplit[1], r, m)
		case "filePostProcessors":
			filePostProcessorsHandle(commandSplit[1], r, m)
		case "sitePostProcessors":
			sitePostProcessorsHandle(commandSplit[1], r, m)
		}

	}
}

func internalHandle(command string, r protocol.Token, m *Module) {
	switch command {
	case "getMethods":
		var templateFunctions []string
		var fileLoaders []string
		var fileParsers []string
		var filePostProcessors []string
		var sitePostProcessors []string

		for key := range m.templateFunctions {
			templateFunctions = append(templateFunctions, key)
		}

		for key := range m.fileLoaders {
			fileLoaders = append(fileLoaders, key)
		}

		for key := range m.fileParsers {
			fileParsers = append(fileParsers, key)
		}

		for key := range m.filePostProcessors {
			filePostProcessors = append(filePostProcessors, key)
		}

		for key := range m.sitePostProcessors {
			sitePostProcessors = append(sitePostProcessors, key)
		}

		r.Respond(protocol.Methods{
			"templateFunctions":  templateFunctions,
			"fileLoaders":        fileLoaders,
			"fileParsers":        fileParsers,
			"filePostProcessors": filePostProcessors,
			"sitePostProcessors": sitePostProcessors,
		})

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

	r.Respond(response.Data)
}

func fileLoadersHandle(command string, r protocol.Token, m *Module) {
	if m.fileLoaders[command].Function == nil {
		fmt.Fprintf(os.Stderr, "does not exist: %s", m.name)
		r.Respond(ErrInvalidCommand)
		return
	}

	var ok bool

	var variable string
	if variable, ok = r.Data[0].(string); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var request = FLRequest{
		Variable: variable,
	}

	var response = &FLResponse{}

	m.fileLoaders[command].Function(request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(response.Data)
}

func fileParsersHandle(command string, r protocol.Token, m *Module) {
	if m.fileParsers[command].Function == nil {
		fmt.Fprintf(os.Stderr, "does not exist: %s", m.name)
		r.Respond(ErrInvalidCommand)
		return
	}

	var ok bool

	var fileInput []byte
	if fileInput, ok = r.Data[0].([]byte); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var request = FPRequest{
		Data:     fileInput,
		Variable: variable,
	}

	var response = &FPResponse{}

	m.fileParsers[command].Function(request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(response.Data)
}

func filePostProcessorsHandle(command string, r protocol.Token, m *Module) {
	if m.filePostProcessors[command].Function == nil {
		r.Respond(ErrInvalidCommand)
		return
	}

	var ok bool

	var objectInput map[string]interface{}
	if objectInput, ok = r.Data[0].(map[string]interface{}); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var request = FPPRequest{
		Data:     objectInput,
		Variable: variable,
	}

	var response = &FPPResponse{}

	m.filePostProcessors[command].Function(request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(response.Data)
}

func sitePostProcessorsHandle(command string, r protocol.Token, m *Module) {
	if m.sitePostProcessors[command].Function == nil {
		r.Respond(ErrInvalidCommand)
		return
	}

	var ok bool

	var objectInput []*site.Site
	if objectInput, ok = r.Data[0].([]*site.Site); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var variable string
	if variable, ok = r.Data[1].(string); ok != true {
		r.Respond(ErrInvalidInput)
		return
	}

	var request = SPPRequest{
		Data:     objectInput,
		Variable: variable,
	}

	var response = &SPPResponse{}

	m.sitePostProcessors[command].Function(request, response)

	if response.Error != nil {
		r.Respond(response.Error)
		return
	}

	r.Respond(response.Data)
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

//ConfigFunctionRegister is the function that will be called that handles the config of the client.
func (m *Module) ConfigFunctionRegister(function func(map[string]interface{}) error) {
	configFunctionRegister(m, function)
	return
}

func configFunctionRegister(m *Module, function func(map[string]interface{}) error) {
	if function == nil {
		panic("module: configFunctionRegister: function is not defined")
	}

	m.configFunction = function

	return
}

//FileLoaderRegister registers a new file loader with identifier "identifier" to the module.
func (m *Module) FileLoaderRegister(identifer string, function func(FLRequest, *FLResponse)) {
	fileLoaderRegister(m, identifer, function)
	return
}

func fileLoaderRegister(m *Module, identifer string, function func(FLRequest, *FLResponse)) {
	if identifer == "" {
		panic("module: fileLoaderRegister: identifer is not defined")
	}

	if m.fileLoaders == nil {
		panic("module: fileLoaderRegister: initalization of module was not correct")
	}

	if m.fileLoaders[identifer].Function != nil {
		panic("module: fileLoaderRegister: fileLoader with this identifier is already registered")
	}

	if function == nil {
		panic("module: fileLoaderRegister: function is not defined")
	}

	m.fileLoaders[identifer] = FileLoader{
		Function: function,
	}

	return
}

//FileParserRegister registers a new file parser with identifier "identifier" to the module.
func (m *Module) FileParserRegister(identifer string, function func(FPRequest, *FPResponse)) {
	fileParserRegister(m, identifer, function)
	return
}

func fileParserRegister(m *Module, identifer string, function func(FPRequest, *FPResponse)) {
	if identifer == "" {
		panic("module: fileParserRegister: identifer is not defined")
	}

	if m.fileParsers == nil {
		panic("module: fileParserRegister: initalization of module was not correct")
	}

	if m.fileParsers[identifer].Function != nil {
		panic("module: fileParserRegister: fileParser with this identifier is already registered")
	}

	if function == nil {
		panic("module: fileParserRegister: function is not defined")
	}

	m.fileParsers[identifer] = FileParser{
		Function: function,
	}

	return
}

//FilePostProcessor registers a new file post processor with identifier "identifier" to the module.
func (m *Module) FilePostProcessor(identifer string, function func(FPPRequest, *FPPResponse)) {
	filePostProcessor(m, identifer, function)
	return
}

func filePostProcessor(m *Module, identifer string, function func(FPPRequest, *FPPResponse)) {
	if identifer == "" {
		panic("module: filePostProcessor: identifer is not defined")
	}

	if m.filePostProcessors == nil {
		panic("module: filePostProcessor: initalization of module was not correct")
	}

	if m.filePostProcessors[identifer].Function != nil {
		panic("module: filePostProcessor: filePostProcessor with this identifier is already registered")
	}

	if function == nil {
		panic("module: filePostProcessor: function is not defined")
	}

	m.filePostProcessors[identifer] = FilePostProcessor{
		Function: function,
	}

	return
}

//SitePostProcessor registers a new site post processor with identifier "identifier" to the module.
func (m *Module) SitePostProcessor(identifer string, function func(SPPRequest, *SPPResponse)) {
	sitePostProcessor(m, identifer, function)
	return
}

func sitePostProcessor(m *Module, identifer string, function func(SPPRequest, *SPPResponse)) {
	if identifer == "" {
		panic("module: sitePostProcessor: identifer is not defined")
	}

	if m.sitePostProcessors == nil {
		panic("module: sitePostProcessor: initalization of module was not correct")
	}

	if m.sitePostProcessors[identifer].Function != nil {
		panic("module: sitePostProcessor: sitePostProcessor with this identifier is already registered")
	}

	if function == nil {
		panic("module: sitePostProcessor: function is not defined")
	}

	m.sitePostProcessors[identifer] = SitePostProcessor{
		Function: function,
	}

	return
}
