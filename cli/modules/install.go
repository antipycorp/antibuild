// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"gitlab.com/antipy/antibuild/cli/cli/modules/repositories"
	"gitlab.com/antipy/antibuild/cli/internal/errors"

	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
)

// InstallCommandRun is the cobra command
func InstallCommandRun(command *cobra.Command, args []string) {
	configfile := *command.Flags().StringP("config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	cfg, err := localConfig.GetConfig(configfile)
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

		installedModule, err := InstallModule(moduleName, module.Version, module.Repository, cfg.Folders.Modules)
		checkModuleErr(err)
		if err != nil {
			return
		}

		cfg.Modules.Dependencies[moduleName] = installedModule

		err = localConfig.SaveConfig(configfile, cfg)
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
}

func checkModuleErr(err errors.Error) {
	if err != nil {
		switch err.GetCode() {
		case ErrFailedModuleBinaryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download module."), tm.RED) + "\n" +
				"\n" +
				"The module you are trying to download has a pre-built binary for your architecture and os but it failed to download. The server might be down. \n" +
				"   More info: " + err.GetRoot() + " \n" +
				"\n")
		case ErrNotExist.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Module is not found."), tm.RED) + "\n" +
				"\n" +
				"   The module you requested is not listed in the module repository specified.\nIs the name of the module spelled correctly?\n" +
				"\n")
		case repositories.ErrFailedModuleRepositoryDownload.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to query the repository."), tm.RED) + "\n" +
				"\n" +
				"   The repository that was specified, or any in the config file, are not valid repositories. Make sure you specified the correct url.\n" +
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
		case ErrUnkownSourceRepositoryType.GetCode():
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
