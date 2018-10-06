package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	tm "github.com/buger/goterm"
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

// modulesAddCmd represents the modules add command
var modulesAddCmd = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Get a module",
	Long:  `Adds and downloads a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfig()
		newModule := args[0]

		if config.Modules.Dependencies[newModule] != "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" is already installed!"), tm.RED))
			tm.Flush()
			return
		}

		config.Modules.Dependencies[newModule] = "0.0.1"

		saveConfig(config)

		err := installModule(newModule)
		checkModuleErr(err)
		tm.Print(tm.Color(tm.Bold("Downloading "+newModule+" at version "+config.Modules.Dependencies[newModule]+"..."), tm.BLUE))
		tm.Flush()

		tm.Print(tm.Color(tm.Bold("Finished downloading "+newModule+"\n \n"), tm.GREEN))
		tm.Flush()

		return
	},
}

// modulesRemoveCmd represents the modules remove command
var modulesRemoveCmd = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
	},
	Short: "Remove a module",
	Long:  `Removes and deletes a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfig()
		newModule := args[0]

		if config.Modules.Dependencies[newModule] == "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" can not be removed because it is not part of this project!"), tm.RED))
			tm.Flush()
			return
		}

		delete(config.Modules.Dependencies, newModule)

		saveConfig(config)

		err := os.Remove(".modules/abm_" + newModule)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tm.Print(tm.Color(tm.Bold("Finished removing "+newModule+"\n \n"), tm.GREEN))
		tm.Flush()

		return
	},
}

// modulesInstallCmd represents the modules install command
var modulesInstallCmd = &cobra.Command{
	Use: "install",
	Aliases: []string{
		"i",
	},
	Short: "Install all modules defined in the config file.",
	Long:  `Will install all modules defined in the config file at the right versions and OS/ARCH.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfig()

		for moduleName, version := range config.Modules.Dependencies {
			tm.Print(tm.Color(tm.Bold("Downloading "+moduleName+" at version "+version+"..."), tm.BLUE))
			tm.Flush()

			err := installModule(moduleName)
			checkModuleErr(err)

			tm.Print(tm.Color(tm.Bold("Finished downloading "+moduleName+"\n \n"), tm.GREEN))
			tm.Flush()
		}
	},
}

var errArchOrOsNotSupported = errors.New("arch or os not supported")

func installModule(moduleName string) error {
	os := runtime.GOOS
	arch := runtime.GOARCH
	module := "abm_" + moduleName

	if !((os == "linux" && (arch == "amd64" || arch == "arm" || arch == "arm64")) || (os == "darwin" && (arch == "amd64")) || (os == "windows" && (arch == "amd64"))) {
		return errArchOrOsNotSupported
	}

	err := downloadFile(".modules/"+module, "https://build.antipy.com/cli/modules/"+os+"/"+arch+"/"+module, true)
	if err != nil {
		return err
	}

	return nil
}

func checkModuleErr(err error) {
	if err != nil {
		switch err {
		case errArchOrOsNotSupported:
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download modules."), tm.RED) + "\n" +
				"\n" +
				"   Your os or arch is not suppported. To learn how to compile a module for your os and arch visit " + tm.Color("https://build.antipy.com/documentation", tm.BLUE) + "\n" +
				"")
		case errFileNotExist:
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download modules."), tm.RED) + "\n" +
				"\n" +
				"   The module you are trying to download does not exist in the repository. Please check on " + tm.Color("https://build.antipy.com/modules", tm.BLUE) + " if you got the right module.\n" +
				"")
		default:
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download modules."), tm.RED) + "\n" +
				"\n" +
				"We could not directly identify the error. Does this help?\n" +
				"   " + err.Error() + "\n" +
				"\n" +
				"If that doesnt help please look on our site " + tm.Color("https://build.antipy.com/", tm.BLUE) + "\n" +
				"")
		}

		tm.Flush()

	}
}
