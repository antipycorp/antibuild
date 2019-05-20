// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	"gitlab.com/antipy/antibuild/cli/builder/config"

	tm "github.com/buger/goterm"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

const (
	// STDRepo is the default module repository
	STDRepo = "build.antipy.com/std"
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
	//ErrFailedGitRepositoryDownload means that the git repository could not be cloned
	ErrFailedGitRepositoryDownload = errors.NewError("failed to clone the git repository", 11)
	//ErrFailedModuleBuild means that the module could not be built
	ErrFailedModuleBuild = errors.NewError("failed to build the module from repository source", 21)
	//ErrFailedModuleRepositoryDownload means the module repository list download failed
	ErrFailedModuleRepositoryDownload = errors.NewError("failed downloading the module repository list", 22)
	//ErrFailedGlobalConfigLoad means loading the global config failed
	ErrFailedGlobalConfigLoad = errors.NewError("failed to load global config", 31)
)

// ModuleEntry is a single entry for a module repository file
type ModuleEntry struct {
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

//ModuleRepository is a repository for modules
type ModuleRepository map[string]ModuleEntry

//Download downloads a json file into a module repository
func (m *ModuleRepository) Download(url string) errors.Error {
	err := internal.DownloadJSON(url, m)
	if err != nil {
		return ErrFailedModuleRepositoryDownload.SetRoot(err.Error())
	}

	return nil
}

//Install installs a module from a repository
func (m ModuleRepository) Install(name string, version string, targetFile string) (string, errors.Error) {
	if v, ok := m[name]; ok {
		return v.Install(version, targetFile)
	}

	return "", ErrNotExist
}

//Install installs a module from a module entry
func (me ModuleEntry) Install(version string, targetFile string) (string, errors.Error) {
	if version == "latest" {
		if me.LatestVersion == "" {
			return "", ErrNoLatestSpecified
		}

		version = me.LatestVersion
	}

	targetFile += "@" + version

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// static compiled modules
	if _, ok := me.CompiledStatic[version]; ok {
		if _, ok := me.CompiledStatic[version][goos]; ok {
			if _, ok := me.CompiledStatic[version][goos][goarch]; ok {
				err := internal.DownloadFile(targetFile, me.CompiledStatic[version][goos][goarch], true)
				if err != nil {
					return "", ErrFailedModuleBinaryDownload.SetRoot(err.Error())
				}

				return version, nil
			}
		}
	}

	// dynamic compiled modules
	if me.CompiledDynamic.URL != "" && contains(me.CompiledDynamic.Vesions, version) {
		if _, ok := me.CompiledDynamic.OSArchCombos[goos]; ok && contains(me.CompiledDynamic.OSArchCombos[goos], goarch) {
			url := strings.ReplaceAll(me.CompiledDynamic.URL, "{{version}}", version)
			url = strings.ReplaceAll(url, "{{os}}", goos)
			url = strings.ReplaceAll(url, "{{arch}}", goarch)

			tm.Print(tm.Color("Using "+tm.Bold(url), tm.BLUE) + tm.Color(" for download.", tm.BLUE) + "\n")
			tm.Flush()
			err := internal.DownloadFile(targetFile, url, true)
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

		err = internal.DownloadGit(dir, me.Source.URL, v)
		if err != nil {
			return "", ErrFailedGitRepositoryDownload.SetRoot(err.Error())
		}

		dir = filepath.Join(dir, filepath.Base(me.Source.URL))

		break
	default:
		return "", ErrUnkownSourceRepositoryType.SetRoot(me.Source.Type + " is not a known source repository type")
	}

	dir = filepath.Join(dir, me.Source.SubDirectory)
	err = internal.CompileFromSource(dir, targetFile)
	if err != nil {
		return "", ErrFailedModuleBuild.SetRoot(err.Error())
	}

	return version, nil
}

//InstallModule installs a module
func InstallModule(name string, version string, repoURL string, filePrefix string) (*config.Module, errors.Error) {
	var repoURLs []string

	if repoURL == "" {
		err := config.LoadDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not load global config file: "+err.Error(), tm.RED) + "\n")
			tm.Flush()
		}

		repoURLs = []string{
			STDRepo,
		}
		repoURLs = append(repoURLs, config.DefaultGlobalConfig.Repositories...)
	} else {
		repoURLs = []string{
			repoURL,
		}
	}

	for _, rURL := range repoURLs {
		repo := &ModuleRepository{}
		var err errors.Error

		err = repo.Download(rURL)
		if err != nil {
			return nil, err
		}

		if internal, ok := InternalModules[name]; ok && internal.version == version && internal.repository == rURL {
			if _, ok := (*repo)[name]; ok {
				tm.Print(tm.Color("Module is available "+tm.Bold("internally"), tm.BLUE) + tm.Color(". There is no need to download.", tm.BLUE) + "\n")
				tm.Flush()
				return &config.Module{Repository: rURL, Version: version}, nil
			}

			continue
		}

		installedVersion, err := repo.Install(name, version, filepath.Join(filePrefix, "abm_"+name))
		if err != nil {
			if err.GetCode() == ErrNotExist.GetCode() {
				continue
			}
			return nil, err
		}

		return &config.Module{Repository: rURL, Version: installedVersion}, nil
	}

	return nil, ErrNotExist.SetRoot("module does not exist in any repository")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
