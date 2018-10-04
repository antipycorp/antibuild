package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/spf13/cobra"
)

// modulesCmd represents the modules command
var modulesCmd = &cobra.Command{
	Use: "modules",
	Aliases: []string{
		"m",
	},
	Short: "Manage your antibuild modules",
	Long:  `Used to manage your modules for antibuild. Run a subcommand to get more info.`,
}

// modulesInstallCmd represents the modules install command
var modulesInstallCmd = &cobra.Command{
	Use: "install",
	Aliases: []string{
		"i",
	},
	Short: "Install a module",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os := runtime.GOOS
		arch := runtime.GOARCH
		module := "abm_" + args[0]
		if !((os == "linux" && (arch == "amd64" || arch == "arm" || arch == "arm64")) || (os == "darwin" && (arch == "amd64")) || (os == "windows" && (arch == "amd64"))) {
			log.Fatal("Your OS or ARCH isnt supported for auto module install.")
		}

		fmt.Println("Getting https://build.antipy.com/cli/modules/" + os + "/" + arch + "/" + module)
		err := downloadFile(".modules/"+module, "https://build.antipy.com/cli/modules/"+os+"/"+arch+"/"+module)
		if err != nil {
			if err == errFileNotExist {
				log.Fatal("That module doesnt exist.")
			}

			log.Fatal(err)
		}
	},
}
