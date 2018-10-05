package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use: "config",
	Aliases: []string{
		"c",
	},
	Short: "Manage your antibuild config",
	Long:  `Used to manage your config for antibuild. Run a subcommand to get more info.`,
}

// configInstallCmd represents the config install command
var configFoldersCmd = &cobra.Command{
	Use: "folders",
	Aliases: []string{
		"f",
	},
	Short: "Manage the folders of a module",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os := runtime.GOOS
		arch := runtime.GOARCH
		module := "abm_" + args[0]
		if !((os == "linux" && (arch == "amd64" || arch == "arm" || arch == "arm64")) || (os == "darwin" && (arch == "amd64")) || (os == "windows" && (arch == "amd64"))) {
			log.Fatal("Your OS or ARCH isnt supported for auto module install.")
		}

		fmt.Println("Getting https://build.antipy.com/cli/config/" + os + "/" + arch + "/" + module)
		err := downloadFile(".config/"+module, "https://build.antipy.com/cli/config/"+os+"/"+arch+"/"+module)
		if err != nil {
			if err == errFileNotExist {
				log.Fatal("That module doesnt exist.")
			}

			log.Fatal(err)
		}
	},
}
