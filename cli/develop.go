// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/engine"
)

func developCommandRun(command *cobra.Command, args []string) {
	configfile := *command.Flags().StringP("config", "c", "config.json",
		"Config file that should be used for building. If not specified will use config.json")
	port := *command.Flags().StringP("port", "p", "8080",
		"The port that is used to host the development server.")

	engine.Start(true, true, configfile, true, port)
}
