package modules

import (
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

const (
	//NoRepoSpecified should be used as a module repo when none is specified, will use the config repos and STDRepo
	NoRepoSpecified = "UNSPECIFIED"
	STDRepo         = "https://build.antipy.com/dl/modules.json"
)

var (
	//ErrNotExist means the module does not exist in the repository
	ErrNotExist = errors.NewError("module does not exist the repository:", 1)
	//ErrNotExistAll means the module does not exist in the repository specified, not in the ones specified in the config, and not in the STDRepo
	ErrNotExistAll = errors.NewError("module does not exist the repositories:", 2)
	//ErrFailedModuleBinaryDownload means the module binary download failed
	ErrFailedModuleBinaryDownload = errors.NewError("failed downloading module binary from repository server", 2)
	//ErrUnkownSourceRepositoryType means that source repository type was not recognized
	ErrUnkownSourceRepositoryType = errors.NewError("source repository code is unknown", 10)
	//ErrFailedGitRepositoryDownload means that the git repository could not be cloned
	ErrFailedGitRepositoryDownload = errors.NewError("failed to clone the git repository", 11)
	//ErrFailedModuleBuild means that the module could not be built
	ErrFailedModuleBuild = errors.NewError("failed to build the module from repository source", 21)
	//ErrFailedModuleRepositoryDownload means the module repository list download failed
	ErrFailedModuleRepositoryDownload = errors.NewError("failed downloading the module repository list", 2)
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
	Compiled map[string]map[string]string
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
func (m ModuleRepository) Install(name, version, targetFile string) errors.Error {
	if v, ok := m[name]; ok {
		return v.Install(version, targetFile)
	}
	return ErrNotExist
}

//Install installs a module from a module entry
func (me ModuleEntry) Install(version, targetFile string) errors.Error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if _, ok := me.Compiled[goos]; ok {
		if _, ok := me.Compiled[goos][goarch]; ok {
			err := internal.DownloadFile(targetFile, me.Compiled[goos][goarch], true)
			if err != nil {
				return ErrFailedModuleBinaryDownload.SetRoot(err.Error())
			}

			return nil
		}
	}
	dir, err := ioutil.TempDir("", "abm_"+me.Name)
	if err != nil {
		panic(err)
	}

	switch me.Source.Type {
	case "git":
		err = internal.DownloadGit(dir, me.Source.URL)
		if err != nil {
			return ErrFailedGitRepositoryDownload.SetRoot(err.Error())
		}

		dir = filepath.Join(dir, filepath.Base(me.Source.URL))

		break
	default:
		return ErrUnkownSourceRepositoryType.SetRoot(me.Source.Type + " is not a known source repository type")
	}

	dir = filepath.Join(dir, me.Source.SubDirectory)
	err = internal.CompileFromSource(dir, targetFile)
	if err != nil {
		return ErrFailedModuleBuild.SetRoot(err.Error())
	}
	return nil

}

//InstallModule installs a module
func InstallModule(name, version, repoURL string, config *Modules) errors.Error {
	repo := &ModuleRepository{}
	var err errors.Error
	if repoURL != NoRepoSpecified {
		err := repo.Download(repoURL)
		err = repo.Install(name, version, "amb_"+name)
		if err != nil {
			return err
		}
		goto postSetup
	}

	for _, v := range config.Repositories {
		err := repo.Download(v)
		err = repo.Install(name, version, "amb_"+name)
		if err != nil {
			if err.GetCode() == ErrNotExist.GetCode() {
				continue
			}
			return err
		}
		goto postSetup
	}
	repoURL = STDRepo
	err = repo.Download(repoURL)
	err = repo.Install(name, version, "amb_"+name)
	if err != nil {
		if err.GetCode() == ErrNotExist.GetCode() {
			return ErrNotExistAll.SetRoot(name)
		}

		return err
	}

postSetup:
	isExist := false
	for _, v := range config.Repositories {
		if v == repoURL {
			isExist = true
			break
		}
	}
	if !isExist {
		config.Repositories = append(config.Repositories, repoURL)
	}
	return nil
}
