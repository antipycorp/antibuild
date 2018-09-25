// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "antibuild",
	Short: "A fast and simple static site generator with module support.",
	Long: `Antibuild is a static site generator that can use dynamic datasets and simple or advanced modules for endless configurability.

To start a new antibuild project run "antibuild new"
Antibuild is written in Golang and can be extended by modules written in Golang. To get started with modules go to https://antibuild.io/modules.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
