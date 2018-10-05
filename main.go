// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder"
)

const version = "v0.4.0"

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "antibuild",
		Short: "A fast and simple static site generator with module support.",
		Long: `Antibuild is a static site generator that can use dynamic datasets and simple or advanced modules for endless configurability.

To start a new antibuild project run "antibuild new"
Antibuild is written in Golang and can be extended by modules written in Golang. To get started with modules go to https://antibuild.io/modules.`,
	}

	configFileDevelopCmd string
	portDevelopCmd       string

	// newCmd represents the new command
	developCmd = &cobra.Command{
		Use:   "develop",
		Short: "Develop a project using the config file file",
		Long:  `Develop a Antibuild project and export into the output folder. Will also install any dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
			builder.Start(true, true, configFileDevelopCmd, true, portDevelopCmd)
		},
	}

	configFileBuildCmd string

	// newCmd represents the new command
	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build a project using the " + configFileBuildCmd + " file",
		Long:  `Build a Antibuild project and export into the output folder. Will also install any dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
			builder.Start(false, false, configFileBuildCmd, true, "")
		},
	}
)

func main() {
	developCmd.Flags().StringVarP(&configFileDevelopCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	developCmd.Flags().StringVarP(&portDevelopCmd, "port", "p", "8080", "The port that is used to host the development server.")
	buildCmd.Flags().StringVarP(&configFileBuildCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	modulesCmd.AddCommand(modulesInstallCmd)
	modulesCmd.AddCommand(modulesAddCmd)
	modulesCmd.AddCommand(modulesRemoveCmd)
	rootCmd.AddCommand(developCmd, buildCmd, newCmd, modulesCmd)
	rootCmd.Execute()
}
