// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	tm "github.com/buger/goterm"
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	cmdInternal "gitlab.com/antipy/antibuild/cli/cmd/internal"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/ui"
)

var fallbackUI = ui.UI{
	HostingEnabled: false,
	PrettyLog:      true,
}

var moduleList = make(map[string]map[string]cmdInternal.ModuleRepositoryEntry)
var moduleListLoaded = make(map[string]bool)

var (
	//ErrFailedModuleBinaryDownload means the module binary download failed
	ErrFailedModuleBinaryDownload = errors.NewError("failed downloading module binary from repository server", 1)
	//ErrFailedModuleRepositoryListDownload means the module repository list download failed
	ErrFailedModuleRepositoryListDownload = errors.NewError("failed downloading the module repository list", 2)
	//ErrUnknownModule means that the module could not be found in the module repository list
	ErrUnknownModule = errors.NewError("module was not found in module repository list", 3)
	//ErrUnkownSourceRepositoryType means that source repository type was not recognized
	ErrUnkownSourceRepositoryType = errors.NewError("source repository code is unknown", 10)
	//ErrFailedGitRepositoryDownload means that the git repository could not be cloned
	ErrFailedGitRepositoryDownload = errors.NewError("failed to clone the git repository", 11)
	//ErrFailedModuleBuild means that the module could not be built
	ErrFailedModuleBuild = errors.NewError("failed to build the module from repository source", 21)
	//ErrFileSystem means that something withh the filesystem has gone wrong
	ErrFileSystem = errors.NewError("failled to ineract with the filesystem", 31)
)

var configFile string
var repositoryFile string

// modulesCMD represents the modules command
var modulesCMD = &cobra.Command{
	Use: "modules",
	Aliases: []string{
		"m",
	},
	Short: "Manage your antibuild modules",
	Long:  `Used to manage your modules for antibuild. Run a subcommand to get more info.`,
}

// modulesAddCMD represents the modules add command
var modulesAddCMD = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Get a module",
	Long:  `Adds and downloads a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		newModule := args[0]

		tm.Print(tm.Color("Downloading "+tm.Bold(newModule), tm.BLUE) + tm.Color(" from repository "+tm.Bold(repositoryFile), tm.BLUE) + "\n")
		tm.Flush()

		err = installModule(newModule, repositoryFile)
		checkModuleErr(err)
		if err != nil {
			return
		}

		cfg.Modules.Dependencies[newModule] = repositoryFile

		err = config.SaveConfig(configFile, cfg)
		if err != nil {
			tm.Print(tm.Color("Config could not be saved.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		tm.Print(tm.Color("Finished downloading "+tm.Bold(newModule), tm.GREEN) + "\n \n")
		tm.Flush()

		return
	},
}

// modulesRemoveCMD represents the modules remove command
var modulesRemoveCMD = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
	},
	Short: "Remove a module",
	Long:  `Removes and deletes a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		newModule := args[0]

		if cfg.Modules.Dependencies[newModule] == "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" can not be removed because it is not part of this project!"), tm.RED))
			tm.Flush()
			return
		}

		delete(cfg.Modules.Dependencies, newModule)

		err = config.SaveConfig(configFile, cfg)
		if err != nil {
			tm.Print(tm.Color("Config could not be saved.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		errDefault := os.Remove(".modules/abm_" + newModule)
		if errDefault != nil {
			return
		}

		tm.Print(tm.Color("Removed the module "+tm.Bold(newModule), tm.GREEN) + "\n \n")
		tm.Flush()

		return
	},
}

// modulesInstallCMD represents the modules install command
var modulesInstallCMD = &cobra.Command{
	Use: "install",
	Aliases: []string{
		"i",
	},
	Short: "Install all modules defined in the config file.",
	Long:  `Will install all modules defined in the config file at the right versions and OS/ARCH.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		for moduleName, moduleRepository := range cfg.Modules.Dependencies {
			tm.Print(tm.Color("Downloading "+tm.Bold(moduleName), tm.BLUE) + tm.Color(" from repository "+tm.Bold(repositoryFile), tm.BLUE) + "\n")
			tm.Flush()

			err := installModule(moduleName, moduleRepository)
			checkModuleErr(err)

			tm.Print(tm.Color("Finished downloading "+tm.Bold(moduleName), tm.GREEN) + "\n \n")
			tm.Flush()
		}
	},
}

func installModule(moduleName string, moduleRepository string) errors.Error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	module := "abm_" + moduleName
	targetFile := ".modules/" + module

	if _, err := os.Stat(".modules/"); os.IsNotExist(err) {
		err = os.MkdirAll(".modules/", 0755)
		if err != nil {
			return ErrFileSystem.SetRoot(err.Error())
		}
	}

	err := updateModuleList(moduleRepository)
	if err != nil {
		return ErrFailedModuleRepositoryListDownload.SetRoot(err.Error())
	}

	var moduleInfo cmdInternal.ModuleRepositoryEntry
	var ok bool

	if moduleInfo, ok = moduleList[moduleRepository][moduleName]; !ok {
		return ErrUnknownModule.SetRoot("")
	}

	if _, ok := moduleInfo.Compiled[goos]; ok {
		if _, ok := moduleInfo.Compiled[goos][goarch]; ok {
			err = internal.DownloadFile(targetFile, moduleInfo.Compiled[goos][goarch], true)
			if err != nil {
				return ErrFailedModuleBinaryDownload.SetRoot(err.Error())
			}

			return nil
		}
	}

	dir, err := ioutil.TempDir("", module)
	if err != nil {
		panic(err)
	}

	switch moduleInfo.Source.Type {
	case "git":
		err = internal.DownloadGit(dir, moduleInfo.Source.URL)
		if err != nil {
			return ErrFailedGitRepositoryDownload.SetRoot(err.Error())
		}

		dir = filepath.Join(dir, filepath.Base(moduleInfo.Source.URL))

		break
	default:
		return ErrUnkownSourceRepositoryType.SetRoot(moduleInfo.Source.Type + " is not a known source repository type")
	}

	dir = filepath.Join(dir, moduleInfo.Source.SubDirectory)
	err = internal.CompileFromSource(dir, targetFile)
	if err != nil {
		return ErrFailedModuleBuild.SetRoot(err.Error())
	}
	return nil
}

func checkModuleErr(err errors.Error) {
	if err != nil {
		switch err.GetCode() {
		case ErrFailedModuleBinaryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module."), tm.RED) + "\n" +
				"\n" +
				"   The module you are trying to download has a pre-built binary for your architecture and os but it failed to download. The server might be down. \n" +
				"\n")
		case ErrFailedModuleRepositoryListDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module repository list."), tm.RED) + "\n" +
				"\n" +
				"   The module repository list could not be downloaded. The server might be down.\n" +
				"\n")
		case ErrUnknownModule.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Module does not exist."), tm.RED) + "\n" +
				"\n" +
				"   The module you requested is not listed in the module repository list. Is the name of the module spelled correctly?\n" +
				"\n")
		case ErrUnkownSourceRepositoryType.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module."), tm.RED) + "\n" +
				"\n" +
				"   The repository type that was supplied in the module repository list is not valid. Are you using an old version of Antibuild?\n" +
				"\n")
		case ErrFailedGitRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download git repository for module."), tm.RED) + "\n" +
				"\n" +
				"   The source code could not be cloned from repository the git repository. Do you have Git installed?\n" +
				"\n")
		case ErrFailedModuleBuild.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to build module form source."), tm.RED) + "\n" +
				"\n" +
				"   The module could not be built from repository source. Make sure you have Go installed.\n" +
				"\n")
		default:
			tm.Print("" +
				tm.Color(tm.Bold("Unknown error."), tm.RED) + "\n" +
				"\n" +
				"We could not directly identify the error. Does this help?\n" +
				"		" + err.Error() + "\n" +
				"\n" +
				"If that doesnt help please look on our site " + tm.Color("https://build.antipy.com/", tm.BLUE) + "\n" +
				"\n")
		}

		tm.Flush()
	}
}

func updateModuleList(moduleRepository string) error {
	if _, ok := moduleListLoaded[moduleRepository]; !ok {
		moduleListLoaded[moduleRepository] = false
	}

	if !moduleListLoaded[moduleRepository] {
		moduleListLoaded[moduleRepository] = true

		var err error
		moduleList[moduleRepository], err = cmdInternal.GetModuleRepository(moduleRepository)
		if err != nil {
			return err
		}
	}

	return nil
}

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	modulesInstallCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesAddCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesAddCMD.Flags().StringVarP(&repositoryFile, "modules", "m", "https://build.antipy.com/dl/modules.json", "The module repository list file to use. Default is \"https://build.antipy.com/dl/modules.json\"")
	modulesRemoveCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	modulesCMD.AddCommand(modulesInstallCMD)
	modulesCMD.AddCommand(modulesAddCMD)
	modulesCMD.AddCommand(modulesRemoveCMD)

	cmd.AddCommand(modulesCMD)
}
