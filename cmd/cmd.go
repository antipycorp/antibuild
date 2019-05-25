// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder"
	"gitlab.com/antipy/antibuild/cli/cmd/modules"
	"gitlab.com/antipy/antibuild/cli/cmd/new"
)

var (
	// developCMD represents the new command
	developCMD = &cobra.Command{
		Use:   "develop",
		Short: "Develop a project using the config file file",
		Long:  `Develop a Antibuild project and export into the output folder. Will also install any dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			configFileDevelopCMD := *cmd.Flags().StringP("config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
			portDevelopCMD := *cmd.Flags().StringP("port", "p", "8080", "The port that is used to host the development server.")

			builder.Start(true, true, configFileDevelopCMD, true, portDevelopCMD)
		},
	}

	// newCMD represents the new command
	buildCMD = &cobra.Command{
		Use:   "build",
		Short: "Build a project",
		Long:  `Build a Antibuild project and export into the output folder. Will also install any dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			configFileBuildCMD := *cmd.Flags().StringP("config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
			builder.Start(false, false, configFileBuildCMD, true, "")
		},
	}
)

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	cmd.AddCommand(developCMD, buildCMD)
	new.SetCommands(cmd)
	modules.SetCommands(cmd)
}
