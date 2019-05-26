// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"os"

	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
)

// RemoveCommandRun is the cobra command
func RemoveCommandRun(command *cobra.Command, args []string) {
	configfile := *command.Flags().StringP("config", "c", "config.json",
		"Config file that should be used for building. If not specified will use config.json")

	cfg, err := localConfig.GetConfig(configfile)
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

	err = localConfig.SaveConfig(configfile, cfg)
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
}
