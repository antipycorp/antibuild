// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/cli/modules/repositories"
	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
	"gitlab.com/antipy/antibuild/cli/engine/modules"
)

// AddCommandRun is the cobra command
func AddCommandRun(command *cobra.Command, args []string) {
	configFile := *command.Flags().StringP("config", "c", "config.json",
		"Config file that should be used for building. If not specified will use config.json")
	repositoryFile := *command.Flags().StringP("modules", "m", repositories.NoRepositorySpecified,
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

	installedModule, err := InstallModule(newModule.Repository,
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
}
