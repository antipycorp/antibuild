package modules

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"

	tm "github.com/buger/goterm"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

const (
	// STDRepo is the default module repository
	STDRepo = "https://build.antipy.com/dl/modules.json"
)

var (
	//ErrNotExist means the module does not exist in the repository
	ErrNotExist = errors.NewError("module does not exist the repository", 1)
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
	Compiled        map[string]map[string]map[string]string `json:"compiled"`
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

	goos := runtime.GOOS
	goarch := runtime.GOARCH

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

	if _, ok := me.Compiled[version]; ok {
		if _, ok := me.Compiled[version][goos]; ok {
			if _, ok := me.Compiled[version][goos][goarch]; ok {
				err := internal.DownloadFile(targetFile, me.Compiled[version][goos][goarch], true)
				if err != nil {
					return "", ErrFailedModuleBinaryDownload.SetRoot(err.Error())
				}

				return version, nil
			}
		}
	}
	dir, err := ioutil.TempDir("", "abm_"+me.Name)
	if err != nil {
		panic(err)
	}

	switch me.Source.Type {
	case "git":
		err = internal.DownloadGit(dir, me.Source.URL, version)
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
func InstallModule(name string, version string, repoURL string, filePrefix string) (string, errors.Error) {
	if version == "internal" {
		tm.Print(tm.Color("Module is "+tm.Bold("internal"), tm.BLUE) + tm.Color(". There is no need to download.", tm.BLUE) + "\n")
		tm.Flush()
		return "internal", nil
	}

	repo := &ModuleRepository{}
	var err errors.Error

	err = repo.Download(repoURL)
	if err != nil {
		return "", err
	}

	installedVersion, err := repo.Install(name, version, filepath.Join(filePrefix, "abm_"+name))
	if err != nil {
		return "", err
	}

	return installedVersion, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
