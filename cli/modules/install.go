// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/compile"
	"gitlab.com/antipy/antibuild/cli/internal/download"
	"gitlab.com/antipy/antibuild/cli/internal/errors"

	tm "github.com/lucacasonato/goterm"
	globalConfig "gitlab.com/antipy/antibuild/cli/configuration/global"
)

type (
	// ModuleEntry is a single entry for a module repository file
	ModuleEntry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      struct {
			Type         string `json:"type"`
			URL          string `json:"url"`
			SubDirectory string `json:"subdirectory"`
		} `json:"source"`
		CompiledStatic  map[string]map[string]map[string]string `json:"compiled_static"`
		CompiledDynamic struct {
			URL          string              `json:"url"`
			Vesions      []string            `json:"versions"`
			OSArchCombos map[string][]string `json:"os_arch_combos"`
		} `json:"compiled_dynamic"`
		LatestVersion string `json:"latest"`
	}
)

const (
	stdRepo = modules.STDRepo

	// NoRepositorySpecified is when you dont pass the -m flag
	NoRepositorySpecified = ""
)

var (
	//ErrNotExist means the module does not exist in the repository
	ErrNotExist = errors.NewError("module does not exist in the repository", 1)
	//ErrNoLatestSpecified means the module binary download failed
	ErrNoLatestSpecified = errors.NewError("the module repository did not specify a latest version", 2)
	//ErrFailedModuleBinaryDownload means the module binary download failed
	ErrFailedModuleBinaryDownload = errors.NewError("failed downloading module binary from repository server", 3)
	//ErrUnkownSourceRepositoryType means that source repository type was not recognized
	ErrUnkownSourceRepositoryType = errors.NewError("source repository code is unknown", 10)
	//ErrFailedModuleBuild means that the module could not be built
	ErrFailedModuleBuild = errors.NewError("failed to build the module from repository source", 21)
)

//InstallModule installs a module
func InstallModule(name string, version string, repoURL string, filePrefix string) (*modules.Module, errors.Error) {
	var repoURLs []string

	if repoURL == NoRepositorySpecified {
		err := globalConfig.LoadDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not load global config file: "+err.Error(), tm.RED) + "\n")
			tm.FlushAll()
		}

		repoURLs = []string{
			stdRepo,
		}
		repoURLs = append(repoURLs, globalConfig.DefaultGlobalConfig.Repositories...)
	} else {
		repoURLs = []string{
			repoURL,
		}
	}

	for _, rURL := range repoURLs {
		repo := &ModuleRepository{}

		err := repo.Download(rURL)
		if err != nil {
			return nil, err
		}

		if me, ok := (*repo)[name]; ok {
			if version == "latest" {
				if me.LatestVersion == "" {
					return nil, ErrNoLatestSpecified
				}

				version = me.LatestVersion
			}

			if modules.MatchesInternalModule(name, version, rURL) == modules.HaveSameVersion {
				tm.Print(tm.Color("Module is available "+tm.Bold("internally"), tm.BLUE) + tm.Color(". There is no need to download.", tm.BLUE) + "\n")
				tm.FlushAll()
				return &modules.Module{Repository: rURL, Version: version}, nil
			}

			installedVersion, err := me.Install(version, filepath.Join(filePrefix, "abm_"+name))
			if err != nil {
				if err.GetCode() == ErrNotExist.GetCode() {
					continue
				}
				return nil, err
			}

			return &modules.Module{Repository: rURL, Version: installedVersion}, nil
		}
	}

	return nil, ErrNotExist.SetRoot("module does not exist in any repository")
}

//Install installs a module from a module entry
func (me ModuleEntry) Install(version string, targetFile string) (string, errors.Error) {
	targetFile += "@" + version

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// static compiled modules
	if _, ok := me.CompiledStatic[version]; ok {
		if _, ok := me.CompiledStatic[version][goos]; ok {
			if _, ok := me.CompiledStatic[version][goos][goarch]; ok {
				err := download.File(targetFile, me.CompiledStatic[version][goos][goarch], true)
				if err != nil {
					return "", ErrFailedModuleBinaryDownload.SetRoot(err.Error())
				}

				return version, nil
			}
		}
	}

	// dynamic compiled modules
	if me.CompiledDynamic.URL != "" && internal.Contains(me.CompiledDynamic.Vesions, version) {
		if _, ok := me.CompiledDynamic.OSArchCombos[goos]; ok && internal.Contains(me.CompiledDynamic.OSArchCombos[goos], goarch) {
			url := strings.ReplaceAll(me.CompiledDynamic.URL, "{{version}}", version)
			url = strings.ReplaceAll(url, "{{os}}", goos)
			url = strings.ReplaceAll(url, "{{arch}}", goarch)

			tm.Print(tm.Color("Using "+tm.Bold(url), tm.BLUE) + tm.Color(" for download.", tm.BLUE) + "\n")
			tm.FlushAll()
			err := download.File(targetFile, url, true)
			if err != nil {
				return "", ErrFailedModuleBinaryDownload.SetRoot(err.Error())
			}

			return version, nil
		}
	}

	// local compiled modules
	dir, err := ioutil.TempDir("", "abm_"+me.Name+"@"+version)
	if err != nil {
		panic(err)
	}

	switch me.Source.Type {
	case "git":
		v := version
		if me.Source.SubDirectory != "" {
			v = me.Source.SubDirectory + "/" + version
		}

		err = download.Git(dir, me.Source.URL, v)
		if err != nil {
			return "", ErrFailedGitRepositoryDownload.SetRoot(err.Error())
		}

		dir = filepath.Join(dir, filepath.Base(me.Source.URL))
	default:
		return "", ErrUnkownSourceRepositoryType.SetRoot(me.Source.Type + " is not a known source repository type")
	}

	dir = filepath.Join(dir, me.Source.SubDirectory)
	err = compile.FromSource(dir, targetFile)
	if err != nil {
		return "", ErrFailedModuleBuild.SetRoot(err.Error())
	}

	return version, nil
}
