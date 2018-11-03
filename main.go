// +build !module

// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/cmd"
	"os"
)

const version = "v0.5.4"

var (
	// rootCMD represents the base command when called without any subcommands
	rootCMD = &cobra.Command{
		Use:   "antibuild",
		Short: "A fast and simple static site generator with module support.",
		Long: `Antibuild is a static site generator that can use dynamic datasets and simple or advanced modules for endless configurability.

To start a new antibuild project run "antibuild new"
Antibuild is written in Golang and can be extended by modules written in Golang. To get started with modules go to https://antibuild.io/modules.`,
	}
	versionCMD = &cobra.Command{
		Use:   "version",
		Short: "Prints the version of antibuild and exits succesfully.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
			os.Exit(0)
		},
	}
)

func main() {
	rootCMD.AddCommand(versionCMD)
	cmd.SetCommands(rootCMD)
	rootCMD.Execute()
}
