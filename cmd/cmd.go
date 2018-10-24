package cmd


import(
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder"
	"gitlab.com/antipy/antibuild/cli/cmd/modules"
	"gitlab.com/antipy/antibuild/cli/cmd/new"
)

var (
	configFileDevelopCmd string
	portDevelopCmd       string

	// newCmd represents the new command
	developCmd = &cobra.Command{
		Use:   "develop",
		Short: "Develop a project using the config file file",
		Long:  `Develop a Antibuild project and export into the output folder. Will also install any dependencies.`,
		Run: func(cmd *cobra.Command, args []string) {
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
			builder.Start(false, false, configFileBuildCmd, true, "")
		},
	}
)
//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command){
	developCmd.Flags().StringVarP(&configFileDevelopCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")
	developCmd.Flags().StringVarP(&portDevelopCmd, "port", "p", "8080", "The port that is used to host the development server.")
	buildCmd.Flags().StringVarP(&configFileBuildCmd, "config", "c", "config.json", "Config file that should be used for building. If not specified will use config.json")

cmd.AddCommand(developCmd, buildCmd)
new.SetCommands(cmd)
modules.SetCommands(cmd)
}