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
		start func(io.Reader, io.Writer)
		name  string
	}

	//ModuleConfig contains the config for modules
	ModuleConfig struct {
		Config map[string]interface{}
	}
)

var (
	templateFunctions = site.TemplateFunctions

	dataLoaders        = &site.DataLoaders
	dataParsers        = &site.DataParsers
	dataPostProcessors = &site.DataPostProcessors
	sitePostProcessors = &site.SPPs

	iterators = &site.Iterators

	internalMods = map[string]internalMod{
		"file": internalMod{
			name:  "file",
			start: abm_file.Handler,
		},
		"json": internalMod{
			name:  "json",
			start: abm_json.Handler,
		},
		"language": internalMod{
			name:  "language",
			start: abm_language.Handler,
		},
		"markdown": internalMod{
			name:  "markdown",
			start: abm_markdown.Handler,
		},
		"math": internalMod{
			name:  "math",
			start: abm_math.Handler,
		},
		"noescape": internalMod{
			name:  "noescape",
			start: abm_noescape.Handler,
		},
		"util": internalMod{
			name:  "util",
			start: abm_util.Handler,
		},
		"yaml": internalMod{
			name:  "yaml",
			start: abm_yaml.Handler,
		},
	}

	loadedModules = make(map[string]string)

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
func LoadModules(moduleRoot string, deps map[string]string, configs map[string]ModuleConfig, log host.Logger) (moduleHost map[string]*host.ModuleHost, err errors.Error) {
	moduleHost = make(map[string]*host.ModuleHost, len(deps))

	for identifier, version := range deps {
		if _, ok := loadedModules[identifier]; ok { //check if the module is still loaded,
			if loadedModules[identifier] == version { //if the version is the same leave it be
				continue
			}
			remModule(identifier, moduleHost) //else remove the old version and continue with loading the new version
		}

		stdout, stdin, err := loadModule(identifier, version, moduleRoot)
		if err != nil {
			return nil, err
		}

		var errr error
		moduleHost[identifier], errr = host.Start(stdout, stdin, log)
		if errr != nil {
			return nil, ErrModuleFailedStarting.SetRoot(errr.Error())
		}

		setupModule(identifier, moduleHost[identifier], configs[identifier])
		loadedModules[identifier] = version
	}
	return
}

//LoadModule Loads a specific module and is menth for hotloading, this
//should not be used for initial setup. For initial setup use LoadModules.
func LoadModule(moduleRoot string, identifier string, version string, moduleHost map[string]*host.ModuleHost, config ModuleConfig, log host.Logger) errors.Error {
	if _, ok := loadedModules[identifier]; ok {
		if loadedModules[identifier] == version {
			return nil
		}
		remModule(identifier, moduleHost)
	}

	defer func() {
		loadedModules[identifier] = version
	}()

	stdout, stdin, err := loadModule(identifier, version, moduleRoot)
	if err != nil {
		return err
	}

	var errr error
	moduleHost[identifier], errr = host.Start(stdout, stdin, log)
	if errr != nil {
		return ErrModuleFailedStarting.SetRoot(errr.Error())
	}

	setupModule(identifier, moduleHost[identifier], config)
	loadedModules[identifier] = version

	return nil
}

func remModule(identifier string, hosts map[string]*host.ModuleHost) {
	hosts[identifier].Kill()
	delete(hosts, identifier)
}

func loadModule(name, version, path string) (io.Reader, io.Writer, errors.Error) {
	fmt.Printf("Loading module: %s@%s\n", name, version)

	if v, ok := internalMods[name]; ok {
		in, stdin := io.Pipe()
		stdout, out := io.Pipe()

		in2 := bufio.NewReader(in)
		stdout2 := bufio.NewReader(stdout)

		go v.start(in2, out)

		return stdout2, stdin, nil
	}

	//prepare command and get nesecary data
	module := exec.Command(filepath.Join(path, "abm_"+name))

	stdin, err := module.StdinPipe()
	if nil != err {
		return nil, nil, ErrModuleFailedObtainStdin.SetRoot(err.Error())
	}

	stdout, err := module.StdoutPipe()
	if nil != err {
		return nil, nil, ErrModuleFailedObtainStdout.SetRoot(err.Error())
	}

	module.Stderr = os.Stderr

	//start module and initaite connection
	if errr := module.Start(); errr != nil {
		return nil, nil, ErrModuleFailedStarting.SetRoot(errr.Error())
	}
	return stdout, stdin, nil
}

func setupModule(identifier string, moduleHost *host.ModuleHost, config ModuleConfig) errors.Error {
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

	if config.Config != nil {
		output, err := moduleHost.ExcecuteMethod("internal_config", []interface{}{
			config.Config,
		})
		if err != nil || output != "module: ready" {
			return ErrModuleFailedConfigure.SetRoot(err.Error())
		}
	}

	return nil
}
