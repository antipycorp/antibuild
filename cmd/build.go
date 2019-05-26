// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/engine"
)

func buildCommandRun(command *cobra.Command, args []string) {
	configfile := *command.Flags().StringP("config", "c", "config.json",
		"Config file that should be used for building. If not specified will use config.json")
	engine.Start(false, false, configfile, true, "")
}
