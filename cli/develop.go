// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder"
)

var configFileDevelopCmd string

// newCmd represents the new command
var developCmd = &cobra.Command{
	Use:   "develop",
	Short: "Develop a project using the" + configFileDevelopCmd + "file",
	Long:  `Develop a Antibuild project and export into the output folder. Will also install any dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Start(true, true, configFileDevelopCmd, true)
	},
}

func init() {
	developCmd.Flags().StringVarP(&configFileDevelopCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	rootCmd.AddCommand(developCmd)
}
