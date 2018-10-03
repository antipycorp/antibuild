// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder"
)

var configFileDevelopCmd string
var port string

// newCmd represents the new command
var developCmd = &cobra.Command{
	Use:   "develop",
	Short: "Develop a project using the" + configFileDevelopCmd + "file",
	Long:  `Develop a Antibuild project and export into the output folder. Will also install any dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		builder.Start(true, true, configFileDevelopCmd, true, port)
	},
}

func init() {
	developCmd.Flags().StringVarP(&configFileDevelopCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	developCmd.Flags().StringVarP(&port, "port", "p", "8080", "The port that should be used to host on.")

	rootCmd.AddCommand(developCmd)
}
