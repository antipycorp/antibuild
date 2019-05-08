// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"os"

	tm "github.com/buger/goterm"
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules"
	"gitlab.com/antipy/antibuild/cli/ui"
)

var fallbackUI = ui.UI{
	HostingEnabled: false,
	PrettyLog:      true,
}

var configFile string
var repositoryFile = modules.STDRepo

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
	Long:  `Adds and downloads a module. Uses the standard repository (` + modules.STDRepo + `) by default. Use -m to change.`,
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

		newModule, errr := modules.ParseModuleString(args[0])
		if errr != nil {
			tm.Print(tm.Color("Module is not valid.", tm.RED) +
				"\n \n")
			tm.Flush()
			return
		}

		tm.Print(tm.Color("Downloading "+tm.Bold(newModule.Repository), tm.BLUE) + tm.Color(" from repository "+tm.Bold(repositoryFile), tm.BLUE) + "\n")
		tm.Flush()

		installedVersion, err := modules.InstallModule(newModule.Repository, newModule.Version, repositoryFile, cfg.Folders.Modules)

		checkModuleErr(err)
		if err != nil {
			return
		}

		cfg.Modules.Dependencies[newModule.Repository] = &modules.Module{
			Repository: repositoryFile,
			Version:    installedVersion,
		}

		err = config.SaveConfig(configFile, cfg)
		if err != nil {
			tm.Print(tm.Color("Config could not be saved.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.Flush()
			return
		}

		tm.Print(tm.Color("Finished downloading "+tm.Bold(newModule.Repository), tm.GREEN) + "\n \n")
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

		if cfg.Modules.Dependencies[newModule].Repository == "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" can not be removed because it is not installed!"), tm.RED))
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

		for moduleName, module := range cfg.Modules.Dependencies {
			tm.Print(tm.Color("Downloading "+tm.Bold(moduleName), tm.BLUE) + tm.Color(" from repository "+tm.Bold(repositoryFile), tm.BLUE) + "\n")
			tm.Flush()

			installedVersion, err := modules.InstallModule(moduleName, module.Version, module.Repository, cfg.Folders.Modules)
			checkModuleErr(err)
			if err != nil {
				return
			}

			cfg.Modules.Dependencies[moduleName] = &modules.Module{
				Repository: repositoryFile,
				Version:    installedVersion,
			}

			err = config.SaveConfig(configFile, cfg)
			if err != nil {
				tm.Print(tm.Color("Config could not be saved.", tm.RED) +
					"This error message might help: " +
					tm.Color(err.Error(), tm.WHITE) +
					"\n \n")
				tm.Flush()
				return
			}

			tm.Print(tm.Color("Finished downloading "+tm.Bold(moduleName), tm.GREEN) + "\n \n")
			tm.Flush()
		}
	},
}

func checkModuleErr(err errors.Error) {
	if err != nil {
		switch err.GetCode() {
		case modules.ErrFailedModuleBinaryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module."), tm.RED) + "\n" +
				"\n" +
				"   The module you are trying to download has a pre-built binary for your architecture and os but it failed to download. The server might be down. \n" +
				"\n")
		case modules.ErrNotExist.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Module is not found."), tm.RED) + "\n" +
				"\n" +
				"   The module you requested is not listed in the module repository specified.\nIs the name of the module spelled correctly?\n" +
				"\n")
		case modules.ErrFailedModuleRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to query the repository."), tm.RED) + "\n" +
				"\n" +
				"   The repository that was specified, or any in the config file, are not valid repositories. Make sure you specified the correct url.\n" +
				"\n")
		case modules.ErrFailedGitRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download git repository for module."), tm.RED) + "\n" +
				"\n" +
				"   The source code could not be cloned from repository the git repository. Do you have Git installed?\n" +
				"\n")
		case modules.ErrFailedModuleBuild.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to build module form source."), tm.RED) + "\n" +
				"\n" +
				"   The module could not be built from repository source. Make sure you have Go installed.\n" +
				"\n")
		case modules.ErrUnkownSourceRepositoryType.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("The repository is invalid."), tm.RED) + "\n" +
				"\n" +
				"   The source type is not supported in this version of antibuild.\n" +
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

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	modulesInstallCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesAddCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesAddCMD.Flags().StringVarP(&repositoryFile, "modules", "m", modules.STDRepo, "The module repository to use.")
	modulesRemoveCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	modulesCMD.AddCommand(modulesInstallCMD)
	modulesCMD.AddCommand(modulesAddCMD)
	modulesCMD.AddCommand(modulesRemoveCMD)

	cmd.AddCommand(modulesCMD)
}
