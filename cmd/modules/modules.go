// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"os"

	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	modulesRepo "gitlab.com/antipy/antibuild/cli/cli/modules"
	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

// This should be extended to also allow for what we do here but in better form.
// We should now directly talk to the terminal here
//var fallbackUI = ui.UI{
//	HostingEnabled: false,
//	PrettyLog:      true,
//}

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
	Long: `Adds and downloads a module. Uses the standard repository (` + modules.STDRepo +
		`) by default. Will use repos in the global config if not found in std. Use -m to force a repo.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := *cmd.Flags().StringP("config", "c", "config.json",
			"Config file that should be used for building. If not specified will use config.json")
		repositoryFile := *cmd.Flags().StringP("modules", "m", modulesRepo.NoRepositorySpecified,
			"The module repository to use.")
		cfg, err := localConfig.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.FlushAll()
			return
		}

		newModule, err := modules.ParseModuleString(args[0])
		if err != nil {
			tm.Print(tm.Color("Module is not valid.", tm.RED) +
				"\n \n")
			tm.FlushAll()
			return
		}

		tm.Print(tm.Color("Installing "+tm.Bold(newModule.Repository), tm.BLUE) + "\n")
		tm.FlushAll()

		installedModule, err := modulesRepo.InstallModule(newModule.Repository,
			newModule.Version, repositoryFile, cfg.Folders.Modules)

		checkModuleErr(err)
		if err != nil {
			return
		}

		cfg.Modules.Dependencies[newModule.Repository] = installedModule

		err = localConfig.SaveConfig(configFile, cfg)
		if err != nil {
			tm.Print(tm.Color("Config could not be saved.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.FlushAll()
			return
		}

		tm.Print(tm.Color("Finished installing "+tm.Bold(newModule.Repository), tm.GREEN) +
			tm.Color(" at version "+tm.Bold(installedModule.Version), tm.GREEN) + tm.Color(" from "+tm.Bold(installedModule.Repository), tm.GREEN) + "\n \n")
		tm.FlushAll()
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
		configFile := *cmd.Flags().StringP("config", "c", "config.json",
			"Config file that should be used for building. If not specified will use config.json")

		cfg, err := localConfig.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.FlushAll()
			return
		}

		newModule := args[0]

		if cfg.Modules.Dependencies[newModule].Repository == "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" can not be removed because it is not installed!"), tm.RED))
			tm.FlushAll()
			return
		}

		delete(cfg.Modules.Dependencies, newModule)

		err = localConfig.SaveConfig(configFile, cfg)
		if err != nil {
			tm.Print(tm.Color("Config could not be saved.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.FlushAll()
			return
		}

		errDefault := os.Remove(".modules/abm_" + newModule)
		if errDefault != nil {
			return
		}

		tm.Print(tm.Color("Removed the module "+tm.Bold(newModule), tm.GREEN) + "\n \n")
		tm.FlushAll()
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
		configFile := *cmd.Flags().StringP("config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

		cfg, err := localConfig.GetConfig(configFile)
		if err != nil {
			tm.Print(tm.Color("Config is not valid.", tm.RED) +
				"This error message might help: " +
				tm.Color(err.Error(), tm.WHITE) +
				"\n \n")
			tm.FlushAll()
			return
		}

		for moduleName, module := range cfg.Modules.Dependencies {
			tm.Print(tm.Color("Downloading "+tm.Bold(moduleName), tm.BLUE) +
				tm.Color(" from repository "+tm.Bold(module.Repository), tm.BLUE) + "\n")
			tm.FlushAll()

			installedModule, err := modulesRepo.InstallModule(moduleName, module.Version, module.Repository, cfg.Folders.Modules)
			checkModuleErr(err)
			if err != nil {
				return
			}

			cfg.Modules.Dependencies[moduleName] = installedModule

			err = localConfig.SaveConfig(configFile, cfg)
			if err != nil {
				tm.Print(tm.Color("Config could not be saved.", tm.RED) +
					"This error message might help: " +
					tm.Color(err.Error(), tm.WHITE) +
					"\n \n")
				tm.FlushAll()
				return
			}

			tm.Print(tm.Color("Finished downloading "+tm.Bold(moduleName), tm.GREEN) + "\n \n")
			tm.FlushAll()
		}
	},
}

func checkModuleErr(err errors.Error) {
	if err != nil {
		switch err.GetCode() {
		case modulesRepo.ErrFailedModuleBinaryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module."), tm.RED) + "\n" +
				"\n" +
				"The module you are trying to download has a pre-built binary for your architecture and os but it failed to download. The server might be down. \n" +
				"   More info: " + err.GetRoot() + " \n" +
				"\n")
		case modulesRepo.ErrNotExist.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Module is not found."), tm.RED) + "\n" +
				"\n" +
				"   The module you requested is not listed in the module repository specified.\nIs the name of the module spelled correctly?\n" +
				"\n")
		case modulesRepo.ErrFailedModuleRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to query the repository."), tm.RED) + "\n" +
				"\n" +
				"   The repository that was specified, or any in the config file, are not valid repositories. Make sure you specified the correct url.\n" +
				"\n")
		case modulesRepo.ErrFailedGitRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download git repository for module."), tm.RED) + "\n" +
				"\n" +
				"   The source code could not be cloned from repository the git repository. Do you have Git installed?\n" +
				"\n")
		case modulesRepo.ErrFailedModuleBuild.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to build module form source."), tm.RED) + "\n" +
				"\n" +
				"   The module could not be built from repository source. Make sure you have Go installed.\n" +
				"\n")
		case modulesRepo.ErrUnkownSourceRepositoryType.GetCode():
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

		tm.FlushAll()
	}
}

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	reposCMD.AddCommand(reposListCMD)
	reposCMD.AddCommand(reposAddCMD)
	reposCMD.AddCommand(reposRemoveCMD)

	modulesCMD.AddCommand(reposCMD)
	modulesCMD.AddCommand(modulesInstallCMD)
	modulesCMD.AddCommand(modulesAddCMD)
	modulesCMD.AddCommand(modulesRemoveCMD)

	cmd.AddCommand(modulesCMD)
}
