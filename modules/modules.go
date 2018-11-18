package modules

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	modFile "gitlab.com/antipy/antibuild/cli/modules/file"
	modJSON "gitlab.com/antipy/antibuild/cli/modules/json"
	modLang "gitlab.com/antipy/antibuild/cli/modules/language"
	modNoESC "gitlab.com/antipy/antibuild/cli/modules/noescape"
)

type (
	fileLoader struct {
		host    *host.ModuleHost
		command string
	}

	fileParser struct {
		host    *host.ModuleHost
		command string
	}

	filePostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	sitePostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	templateFunction struct {
		host    *host.ModuleHost
		command string
	}

	internalMod struct {
		version string
		start   func(io.Reader, io.Writer)
		name    string
	}
	//ModuleConfig contains the config for modules
	ModuleConfig struct {
		Config map[string]interface{}
	}
)

var (
	templateFunctions = site.TemplateFunctions

	fileLoaders        = &site.FileLoaders
	fileParsers        = &site.FileParsers
	filePostProcessors = &site.FilePostProcessors
	sitePostProcessors = &site.SPPs

	internalMods = map[string]internalMod{
		"file": internalMod{
			version: "0.0.1",
			name:    "file",
			start:   modFile.Start,
		},
		"json": internalMod{
			version: "0.0.1",
			name:    "json",
			start:   modJSON.Start,
		},
		"language": internalMod{
			version: "0.0.2",
			name:    "language",
			start:   modLang.Start,
		},
		"noescape": internalMod{
			version: "0.0.1",
			name:    "noescape",
			start:   modNoESC.Start,
		},
	}

	loadedModules = make(map[string]string)
)

//LoadModules communicates with modules to load them.
//Although this should be used for initial setup, for hoatloading modules use LoadModule.
func LoadModules(moduleRoot string, deps map[string]string, configs map[string]ModuleConfig, logger host.Logger) (moduleHost map[string]*host.ModuleHost) {
	moduleHost = make(map[string]*host.ModuleHost, len(deps))

	for identifier, version := range deps {
		if _, ok := loadedModules[identifier]; ok { //check if the module is still loaded,
			if loadedModules[identifier] == version { //if the version is the same leave it be
				continue
			}
			remModule(identifier, moduleHost) //else remove the old version and continue with loading the new version
		}

		stdout, stdin := loadModule(identifier, version, moduleRoot)
		if stdout == nil || stdin == nil {
			return
		}
		var err error
		moduleHost[identifier], err = host.Start(stdout, stdin, logger)
		if err != nil {
			panic(err)
		}
		setupModule(identifier, moduleHost[identifier], configs[identifier])
		loadedModules[identifier] = version
	}
	return
}

//LoadModule Loads a specific module and is menth for hotloading, this
//should not be used for initial setup. For initial setup use LoadModules.
func LoadModule(moduleRoot string, identifier string, version string, moduleHost map[string]*host.ModuleHost, config ModuleConfig, logger host.Logger) {
	if _, ok := loadedModules[identifier]; ok {
		if loadedModules[identifier] == version {
			return
		}
		remModule(identifier, moduleHost)
	}

	defer func() {
		loadedModules[identifier] = version
	}()

	stdout, stdin := loadModule(identifier, version, moduleRoot)
	if stdout == nil || stdin == nil {
		return
	}
	var err error
	moduleHost[identifier], err = host.Start(stdout, stdin, logger)
	if err != nil {
		panic(err)
	}
	setupModule(identifier, moduleHost[identifier], config)
	loadedModules[identifier] = version
}

func remModule(identifier string, hosts map[string]*host.ModuleHost) {
	hosts[identifier].Kill()
	delete(hosts, identifier)
}

func loadModule(name, version, path string) (io.Reader, io.Writer) {

	fmt.Printf("Loading module: %s@%s\n", name, version)

	if v, ok := internalMods[name]; ok {
		if v.version == version {

			in, stdin := io.Pipe()
			stdout, out := io.Pipe()

			in2 := bufio.NewReader(in)
			stdout2 := bufio.NewReader(stdout)

			go v.start(in2, out)

			return stdout2, stdin
		}
	}

	//prepare command and get nesecary data
	module := exec.Command(filepath.Join(path, "abm_"+name))

	stdin, err := module.StdinPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}

	stdout, err := module.StdoutPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}

	module.Stderr = os.Stderr

	//start module and initaite connection
	if err := module.Start(); err != nil {
		panic(err)
	}
	return stdout, stdin
}

func setupModule(identifier string, moduleHost *host.ModuleHost, config ModuleConfig) {
	methods, err := moduleHost.AskMethods()
	if err != nil {
		panic(err)
	}

	//registers all functions modules can possibly suply
	for _, function := range methods["templateFunctions"] {
		templateFunctions[identifier+"_"+function] = getTemplateFunction(function, moduleHost)
	}

	for _, function := range methods["fileLoaders"] {
		(*fileLoaders)[identifier+"_"+function] = getFileLoader(function, moduleHost)
	}

	for _, function := range methods["fileParsers"] {
		(*fileParsers)[identifier+"_"+function] = getFileParser(function, moduleHost)
	}

	for _, function := range methods["filePostProcessors"] {
		(*filePostProcessors)[identifier+"_"+function] = getFilePostProcessor(function, moduleHost)
	}

	for _, function := range methods["sitePostProcessors"] {
		(*sitePostProcessors)[identifier+"_"+function] = getSitePostProcessor(function, moduleHost)
	}

	if config.Config != nil {
		output, err := moduleHost.ExcecuteMethod("internal_config", []interface{}{
			config.Config,
		})
		if err != nil || output != "module: ready" {
			panic("couldnt send config: " + err.Error())
		}
	}
}

func getTemplateFunction(command string, host *host.ModuleHost) *templateFunction {
	return &templateFunction{
		host:    host,
		command: command,
	}
}

func (f *templateFunction) Load(data ...interface{}) []byte {
	output, err := f.host.ExcecuteMethod("templateFunctions_"+f.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []byte
	var ok bool
	if outputFinal, ok = output.([]byte); ok != true {
		panic("fileLoader_" + f.command + " did not return a []byte")
	}

	return outputFinal
}

func getFileLoader(command string, host *host.ModuleHost) *fileLoader {
	return &fileLoader{
		host:    host,
		command: command,
	}
}

func (f *fileLoader) Load(variable string) []byte {
	data := []interface{}{
		variable,
	}

	output, err := f.host.ExcecuteMethod("fileLoaders_"+f.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []byte
	var ok bool
	if outputFinal, ok = output.([]byte); ok != true {
		panic("fileLoader_" + f.command + " did not return a []byte")
	}

	return outputFinal
}

func getFileParser(command string, host *host.ModuleHost) *fileParser {
	return &fileParser{
		host:    host,
		command: command,
	}
}

func (f *fileParser) Parse(data []byte, variable string) map[string]interface{} {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := f.host.ExcecuteMethod("fileParsers_"+f.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal map[string]interface{}
	var ok bool
	if outputFinal, ok = output.(map[string]interface{}); ok != true {
		panic("fileParser_" + f.command + " did not return a map[string]interface{}")
	}

	return outputFinal
}

func getFilePostProcessor(command string, host *host.ModuleHost) *filePostProcessor {
	return &filePostProcessor{
		host:    host,
		command: command,
	}
}

func (f *filePostProcessor) Process(data map[string]interface{}, variable string) map[string]interface{} {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := f.host.ExcecuteMethod("filePostProcessors_"+f.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal map[string]interface{}
	var ok bool
	if outputFinal, ok = output.(map[string]interface{}); ok != true {
		panic("filePostProcessors_" + f.command + " did not return a map[string]interface{}")
	}

	return outputFinal
}

func getSitePostProcessor(command string, host *host.ModuleHost) sitePostProcessor {
	return sitePostProcessor{
		host:    host,
		command: command,
	}
}

func (s sitePostProcessor) Process(data []*site.Site, variable string) []*site.Site {
	sendData := []interface{}{
		data,
		variable,
	}

	output, err := s.host.ExcecuteMethod("sitePostProcessors_"+s.command, sendData)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []*site.Site
	var ok bool
	if outputFinal, ok = output.([]*site.Site); ok != true {
		panic("sitePostProcessors_" + s.command + " did not return a []*site.Site")
	}

	return outputFinal
}

//UnmarshalJSON unmarschals the json into a moduleconfig
func (mc *ModuleConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &mc.Config); err != nil {
		return err
	}
	return nil
}

//MarshalJSON marschals the data into json
func (mc *ModuleConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(mc.Config)
}
