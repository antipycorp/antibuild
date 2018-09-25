// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/builder"
)

var configFileBuildCmd string

// newCmd represents the new command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a project using the " + configFileBuildCmd + " file",
	Long:  `Build a Antibuild project and export into the output folder. Will also install any dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Start(false, false, configFileBuildCmd, true)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&configFileBuildCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	rootCmd.AddCommand(buildCmd)
}
