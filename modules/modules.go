package modules

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/site"
)

type (
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

	internalMods = map[string]internalMod{}

	loadedModules = make(map[string]string)
)

//LoadModules communicates with modules to load them.
//Although this should be used for initial setup, for hoatloading modules use LoadModule.
func LoadModules(moduleRoot string, deps map[string]string, configs map[string]ModuleConfig, log host.Logger) (moduleHost map[string]*host.ModuleHost) {
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
		moduleHost[identifier], err = host.Start(stdout, stdin, log)
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
func LoadModule(moduleRoot string, identifier string, version string, moduleHost map[string]*host.ModuleHost, config ModuleConfig, log host.Logger) {
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
	moduleHost[identifier], err = host.Start(stdout, stdin, log)
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
		templateFunctions[identifier+"_"+function] = getTemplateFunction(function, moduleHost).Run
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
