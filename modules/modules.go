// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"

	abm_file "gitlab.com/antipy/antibuild/std/file/handler"
	abm_json "gitlab.com/antipy/antibuild/std/json/handler"
	abm_language "gitlab.com/antipy/antibuild/std/language/handler"
	abm_markdown "gitlab.com/antipy/antibuild/std/markdown/handler"
	abm_math "gitlab.com/antipy/antibuild/std/math/handler"
	abm_noescape "gitlab.com/antipy/antibuild/std/noescape/handler"
	abm_util "gitlab.com/antipy/antibuild/std/util/handler"
	abm_yaml "gitlab.com/antipy/antibuild/std/yaml/handler"
)

type (
	internalMod struct {
		start      func(io.Reader, io.Writer)
		version    string
		repository string
	}
)

var (
	templateFunctions = site.TemplateFunctions

	dataLoaders        = &site.DataLoaders
	dataParsers        = &site.DataParsers
	dataPostProcessors = &site.DataPostProcessors
	sitePostProcessors = &site.SPPs

	iterators = &site.Iterators

	// InternalModules that the are integrated into the antibuild binary
	InternalModules = map[string]internalMod{
		"file": internalMod{
			start:      abm_file.Handler,
			version:    abm_file.Version,
			repository: STDRepo,
		},
		"json": internalMod{
			start:      abm_json.Handler,
			version:    abm_json.Version,
			repository: STDRepo,
		},
		"language": internalMod{
			start:      abm_language.Handler,
			version:    abm_language.Version,
			repository: STDRepo,
		},
		"markdown": internalMod{
			start:      abm_markdown.Handler,
			version:    abm_markdown.Version,
			repository: STDRepo,
		},
		"math": internalMod{
			start:      abm_math.Handler,
			version:    abm_math.Version,
			repository: STDRepo,
		},
		"noescape": internalMod{
			start:      abm_noescape.Handler,
			version:    abm_noescape.Version,
			repository: STDRepo,
		},
		"util": internalMod{
			start:      abm_util.Handler,
			version:    abm_util.Version,
			repository: STDRepo,
		},
		"yaml": internalMod{
			start:      abm_yaml.Handler,
			version:    abm_yaml.Version,
			repository: STDRepo,
		},
	}

	loadedModules = make(map[string]*config.Module)

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
)

//LoadModules communicates with modules to load them.
//Although this should be used for initial setup, for hoatloading modules use LoadModule.
func LoadModules(moduleRoot string, modules config.Modules, log host.Logger) (moduleHost map[string]*host.ModuleHost, err errors.Error) {
	deps := modules.Dependencies
	configs := modules.Config
	moduleHost = make(map[string]*host.ModuleHost, len(deps))

	for identifier, meta := range deps {
		if _, ok := loadedModules[identifier]; ok { // check if the module is still loaded,
			if loadedModules[identifier].Repository == meta.Repository && loadedModules[identifier].Version == meta.Version { // if the version is the same leave it be
				continue
			}

			remModule(identifier, moduleHost) // else remove the old version and continue with loading the new version
		}

		stdout, stdin, versionLoaded, err := loadModule(identifier, meta, moduleRoot)
		if err != nil {
			return nil, err
		}

		var errr error
		moduleHost[identifier], errr = host.Start(stdout, stdin, log, versionLoaded)
		if errr != nil {
			return nil, ErrModuleFailedStarting.SetRoot(errr.Error())
		}

		err = setupModule(identifier, moduleHost[identifier], configs[identifier])
		if err != nil {
			return nil, err
		}

		loadedModules[identifier] = meta
	}

	return
}

//LoadModule Loads a specific module and is menth for hotloading, this
//should not be used for initial setup. For initial setup use LoadModules.
func LoadModule(moduleRoot string, identifier string, meta *config.Module, moduleHost map[string]*host.ModuleHost, config map[string]interface{}, log host.Logger) errors.Error {
	if _, ok := loadedModules[identifier]; ok {
		if loadedModules[identifier].Repository == meta.Repository && loadedModules[identifier].Version == meta.Version {
			return nil
		}

		remModule(identifier, moduleHost)
	}

	defer func() {
		loadedModules[identifier] = meta
	}()

	stdout, stdin, versionLoaded, err := loadModule(identifier, meta, moduleRoot)
	if err != nil {
		return err
	}

	var errr error
	moduleHost[identifier], errr = host.Start(stdout, stdin, log, versionLoaded)
	if errr != nil {
		return ErrModuleFailedStarting.SetRoot(errr.Error())
	}

	err = setupModule(identifier, moduleHost[identifier], config)
	if err != nil {
		return err
	}

	loadedModules[identifier] = meta

	return nil
}

func remModule(identifier string, hosts map[string]*host.ModuleHost) {
	hosts[identifier].Kill()
	delete(hosts, identifier)
}

func loadModule(name string, meta *config.Module, path string) (io.Reader, io.Writer, string, errors.Error) {
	//TODO: make this a log.debug thing
	fmt.Printf("Loading module %s from %s at %s version\n", name, meta.Repository, meta.Version)

	if v, ok := InternalModules[name]; ok && (meta.Version == v.version) {
		in, stdin := io.Pipe()
		stdout, out := io.Pipe()

		in2 := bufio.NewReader(in)
		stdout2 := bufio.NewReader(stdout)

		go v.start(in2, out)

		return stdout2, stdin, v.version, nil
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
		templateFunctions[identifier+"_"+function] = getTemplateFunction(function, moduleHost).Run
	}

	for _, function := range methods["dataLoaders"] {
		(*dataLoaders)[identifier+"_"+function] = getDataLoader(function, moduleHost)
	}

	for _, function := range methods["dataParsers"] {
		(*dataParsers)[identifier+"_"+function] = getDataParser(function, moduleHost)
	}

	for _, function := range methods["dataPostProcessors"] {
		(*dataPostProcessors)[identifier+"_"+function] = getDataPostProcessor(function, moduleHost)
	}

	for _, function := range methods["sitePostProcessors"] {
		(*sitePostProcessors)[identifier+"_"+function] = getSitePostProcessor(function, moduleHost)
	}

	for _, function := range methods["iterators"] {
		(*iterators)[identifier+"_"+function] = getIterator(function, moduleHost)
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
