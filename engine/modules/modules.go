// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/blang/semver"
	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (

	//DataLoader is a module that loads data
	DataLoader interface {
		GetPipe(string) Pipe
	}

	//DataParser is a module that parses loaded data
	DataParser interface {
		GetPipe(string) Pipe
	}

	//DPP is a function thats able to post-process data
	DPP interface {
		GetPipe(string) Pipe
	}

	//SPP is a function thats able to post-process data
	SPP interface {
		GetPipe(string) Pipe
	}

	//Iterator is a function thats able to post-process data
	Iterator interface {
		GetIterations(string) []string
	}
)

const (
	//STDRepo is the standard repo for antibuild modules
	STDRepo = "build.antipy.com/std"
)

var (
	//ErrModuleFailedStarting means a module failed to start
	ErrModuleFailedStarting = errors.NewError("module failed to start", 1)
	//ErrModuleFailedObtainStdin means we could not obtain stdin
	ErrModuleFailedObtainStdin = errors.NewError("failed to obtain stdin", 2)
	//ErrModuleFailedObtainStdout means we could not obtain stdout
	ErrModuleFailedObtainStdout = errors.NewError("failed to obtain stdout", 3)
	//ErrModuleFailedObtainFunctions means we could not obtain the registered functions
	ErrModuleFailedObtainFunctions = errors.NewError("failed to obtain registered functions", 4)
	//ErrModuleFailedConfigure means we could not configure module
	ErrModuleFailedConfigure = errors.NewError("failed to configure module", 5)

	//TemplateFunctions are all the template functions defined by modules
	TemplateFunctions = template.FuncMap{}

	//DataLoaders are all the module data loaders
	DataLoaders = make(map[string]DataLoader)
	//DataParsers are all the module file parsers
	DataParsers = make(map[string]DataParser)
	//DataPostProcessors are all the module data post processors
	DataPostProcessors = make(map[string]DPP)
	//SPPs are all the module site post processors
	SPPs = make(map[string]SPP)

	//Iterators are all the module iterators
	Iterators = make(map[string]Iterator)

	loadedModules = make(map[string]*Module)
)

// LoadModules loads all modules listed in the deps.
// For hotloading individual modules use LoadModule, but this is safe to use after startup
func LoadModules(moduleRoot string, modules Modules, log host.Logger) (moduleHost map[string]*host.ModuleHost, err errors.Error) {
	deps := modules.Dependencies
	configs := modules.Config
	moduleHost = make(map[string]*host.ModuleHost, len(deps))

	for identifier, meta := range deps {
		//if any problems come up uncomment this, but this is alreaady checked in LoadModule and better so that should be used
		/* if _, ok := loadedModules[identifier]; ok { // check if the module is still loaded,
			if loadedModules[identifier].Repository == meta.Repository && loadedModules[identifier].Version == meta.Version { // if the version is the same leave it be
				//dont remove the module if all the meta is still the same
				continue
			}
			remModule(identifier, moduleHost)
		} */

		err := LoadModule(moduleRoot, identifier, meta, moduleHost, configs[identifier], log)
		if err != nil {
			return nil, err
		}
	}

	return
}

//LoadModule Loads a specific module and is menth for hotloading, this
//should not be used for initial setup. For initial setup use LoadModules.
func LoadModule(moduleRoot string, identifier string, meta *Module, moduleHost map[string]*host.ModuleHost, cfg map[string]interface{}, log host.Logger) errors.Error {
	if _, ok := loadedModules[identifier]; ok {
		if loadedModules[identifier].Repository == meta.Repository && loadedModules[identifier].Version == meta.Version {
			return nil
		}

		remModule(identifier, moduleHost)
	}

	defer func() {
		loadedModules[identifier] = meta
	}()

	stdout, stdin, versionLoaded, err := loadModule(identifier, meta, moduleRoot, log)
	if err != nil {
		return err
	}

	var errr error
	moduleHost[identifier], errr = host.Start(stdout, stdin, log, versionLoaded)
	if errr != nil {
		return ErrModuleFailedStarting.SetRoot(errr.Error())
	}

	err = setupModule(identifier, moduleHost[identifier], cfg)
	if err != nil {
		return err
	}

	loadedModules[identifier] = meta

	return nil
}

func loadModule(name string, meta *Module, path string, log host.Logger) (io.Reader, io.Writer, string, errors.Error) {
	log.Debugf("Loading module %s from %s at %s version", name, meta.Repository, meta.Version)

	if v, ok := InternalModules[name]; ok && (meta.Repository == v.repository) {
		//this guarantees we have an internal module will be availabel thus no error, els load an exernal module
		if MatchesInternalModule(name, meta.Version, v.repository) == HaveSameVersion {
			return LoadInternalModule(meta, name, log)
		}

		//TODO: switch to mustParse, we should make sure that the version is correct during config parsing
		internalVersion, err := semver.Parse(v.version)
		if err == nil {
			requestedVersion, err := semver.Parse(meta.Version)
			if err == nil {
				//TODO: make this also check if the major version is the same, if so make it a warning and say a compatibel veriosn is available
				if internalVersion.GT(requestedVersion) {
					log.Infof("Module %s has a more up to date internal version available: %s", name, v.version)
				}
			}
		}

	}

	//prepare command and get nesecary data
	module := exec.Command(filepath.Join(path, "abm_"+name+"@"+meta.Version))

	stdin, err := module.StdinPipe()
	if nil != err {
		return nil, nil, "", ErrModuleFailedObtainStdin.SetRoot(err.Error())
	}

	stdout, err := module.StdoutPipe()
	if nil != err {
		return nil, nil, "", ErrModuleFailedObtainStdout.SetRoot(err.Error())
	}

	module.Stderr = os.Stderr

	//start module and initaite connection
	if errr := module.Start(); errr != nil {
		return nil, nil, "", ErrModuleFailedStarting.SetRoot(errr.Error())
	}

	return stdout, stdin, meta.Version, nil
}

func setupModule(identifier string, moduleHost *host.ModuleHost, config map[string]interface{}) errors.Error {
	methods, errr := moduleHost.AskMethods()
	if errr != nil {
		return ErrModuleFailedObtainFunctions.SetRoot(errr.Error())
	}

	//registers all functions modules can possibly suply
	for _, function := range methods["templateFunctions"] {
		TemplateFunctions[identifier+"_"+function] = getTemplateFunction(function, moduleHost).Run
	}

	for _, function := range methods["dataLoaders"] {
		DataLoaders[identifier+"_"+function] = getDataLoader(function, moduleHost)
	}

	for _, function := range methods["dataParsers"] {
		DataParsers[identifier+"_"+function] = getDataParser(function, moduleHost)
	}

	for _, function := range methods["dataPostProcessors"] {
		DataPostProcessors[identifier+"_"+function] = getDataPostProcessor(function, moduleHost)
	}

	for _, function := range methods["sitePostProcessors"] {
		SPPs[identifier+"_"+function] = getSitePostProcessor(function, moduleHost)
	}

	for _, function := range methods["iterators"] {
		Iterators[identifier+"_"+function] = getIterator(function, moduleHost)
	}

	if config != nil {
		output, err := moduleHost.ExcecuteMethod("internal_config", []interface{}{
			config,
		})
		if err != nil || output != "module: ready" {
			return ErrModuleFailedConfigure.SetRoot(err.Error())
		}
	}

	return nil
}

func remModule(identifier string, hosts map[string]*host.ModuleHost) {
	hosts[identifier].Kill()
	delete(hosts, identifier)
}
