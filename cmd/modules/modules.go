package modules

import (
	"fmt"
	"os"
	"runtime"

	tm "github.com/buger/goterm"
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/ui"
)

var fallbackUI = ui.UI{
	HostingEnabled: false,
	PrettyLog: true,
}

var(
		//ErrFailledDownload is when the template failled building
		ErrFailledDownload = errors.NewError("failled downloading file", 1)
		//ErrArchNotSupported is for a faillure moving the static folder
		ErrArchNotSupported = errors.NewError("arch or os not supported", 2)	
)

var configFile string

// modulesCMD represents the modules command
var modulesCMD = &cobra.Command{
	Use: "modules",
	Aliases: []string{
		"m",
	},
	Short: "Manage your antibuild modules",
	Long:  `Used to manage your modules for antibuild. Run a subcommand to get more info.`,
}

// modulesAddCMD represents the modules add command
var modulesAddCMD = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Get a module",
	Long:  `Adds and downloads a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil{
			fallbackUI.Fatal("could not get the config")
			fallbackUI.ShowResult()
			return
		}

		newModule := args[0]

		if cfg.Modules.Dependencies[newModule] != "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" is already installed!"), tm.RED))
			tm.Flush()
			return
		}

		cfg.Modules.Dependencies[newModule] = "0.0.1"

		config.SaveConfig(configFile, cfg)

		err = installModule(newModule)
		checkModuleErr(err)
		tm.Print(tm.Color(tm.Bold("Downloading "+newModule+" at version "+cfg.Modules.Dependencies[newModule]+"..."), tm.BLUE))
		tm.Flush()

		tm.Print(tm.Color(tm.Bold("Finished downloading "+newModule+"\n \n"), tm.GREEN))
		tm.Flush()

		return
	},
}

// modulesRemoveCMD represents the modules remove command
var modulesRemoveCMD = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
	},
	Short: "Remove a module",
	Long:  `Removes and deletes a module.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil{
			fallbackUI.Fatal("could not get the config")
			fallbackUI.ShowResult()
			return
		}
		newModule := args[0]

		if cfg.Modules.Dependencies[newModule] == "" {
			tm.Print(tm.Color(tm.Bold("The module "+newModule+" can not be removed because it is not part of this project!"), tm.RED))
			tm.Flush()
			return
		}

		delete(cfg.Modules.Dependencies, newModule)

		config.SaveConfig(configFile, cfg)

		err = errors.Import(os.Remove(".modules/abm_" + newModule))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tm.Print(tm.Color(tm.Bold("Finished removing "+newModule+"\n \n"), tm.GREEN))
		tm.Flush()

		return
	},
}

// modulesInstallCMD represents the modules install command
var modulesInstallCMD = &cobra.Command{
	Use: "install",
	Aliases: []string{
		"i",
	},
	Short: "Install all modules defined in the config file.",
	Long:  `Will install all modules defined in the config file at the right versions and OS/ARCH.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.GetConfig(configFile)
		if err != nil{
			fallbackUI.Fatal("could not get the config")
			fallbackUI.ShowResult()
			return
		}

		for moduleName, version := range cfg.Modules.Dependencies {
			tm.Print(tm.Color(tm.Bold("Downloading "+moduleName+" at version "+version+"..."), tm.BLUE))
			tm.Flush()

			err := installModule(moduleName)
			checkModuleErr(err)

			tm.Print(tm.Color(tm.Bold("Finished downloading "+moduleName+"\n \n"), tm.GREEN))
			tm.Flush()
		}
	},
}

func installModule(moduleName string) errors.Error {
	os := runtime.GOOS
	arch := runtime.GOARCH
	module := "abm_" + moduleName

	if !((os == "linux" && (arch == "amd64" || arch == "arm" || arch == "arm64")) || (os == "darwin" && (arch == "amd64")) || (os == "windows" && (arch == "amd64"))) {
		return ErrArchNotSupported.SetRoot("you are using an os/arch combination that isn't supported")
	}

	err := internal.DownloadFile(".modules/"+module, "https://build.antipy.com/cli/modules/"+os+"/"+arch+"/"+module, true)
	if err != nil {
		return ErrFailledDownload.SetRoot(err.Error())
	}

	return nil
}

func checkModuleErr(err errors.Error) {
	if err != nil {
		switch err.GetCode() {
		case ErrArchNotSupported.GetCode():
			tm.Print("" +
				tm.Color(tm.Bold("Failed to download modules."), tm.RED) + "\n" +
				"\n" +
				"   Your os or arch is not suppported. To learn how to compile a module for your os and arch visit " + tm.Color("https://build.antipy.com/documentation", tm.BLUE) + "\n" +
				"")
		case ErrFailledDownload.GetCode():
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

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command){
	modulesInstallCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesAddCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	modulesRemoveCMD.Flags().StringVarP(&configFile, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

	modulesCMD.AddCommand(modulesInstallCMD)
	modulesCMD.AddCommand(modulesAddCMD)
	modulesCMD.AddCommand(modulesRemoveCMD)

	cmd.AddCommand(modulesCMD)
}