package modules

import (
	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	globalConfig "gitlab.com/antipy/antibuild/cli/configuration/global"
)

// reposCMD represents the repositories command
var reposCMD = &cobra.Command{
	Use: "repositories",
	Aliases: []string{
		"repos",
	},
	Short: "Manage your antibuild module repositories",
	Long:  `Used to manage your module repositories for antibuild. Run a subcommand to get more info.`,
}

// reposListCMD represents the module repositories list command
var reposListCMD = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"list",
	},
	Short: "List all repositories in the global antibuild config file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := globalConfig.LoadDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not load global config file: "+err.Error()+"\n", tm.RED))
			tm.FlushAll()
			return
		}

		for _, repo := range globalConfig.DefaultGlobalConfig.Repositories {
			tm.Print(repo + "\n")
		}

		tm.FlushAll()
	},
}

// reposAddCMD represents the module repositories add command
var reposAddCMD = &cobra.Command{
	Use: "add",
	Aliases: []string{
		"a",
	},
	Short: "Add a repository to the global antibuild config file.",
	Long:  `Adds a repository that is used to try to pull modules from when adding modules.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := globalConfig.LoadDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not load global config file: "+err.Error(), tm.RED) + "\n")
			tm.FlushAll()
			return
		}

		for _, repo := range globalConfig.DefaultGlobalConfig.Repositories {
			if repo == args[0] {
				tm.Print(tm.Color("This repository is already added.", tm.RED) + "\n")
				tm.FlushAll()
				return
			}
		}

		globalConfig.DefaultGlobalConfig.Repositories = append(globalConfig.DefaultGlobalConfig.Repositories, args[0])
		err = globalConfig.SaveDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not save global config file: "+err.Error(), tm.RED) + "\n")
			tm.FlushAll()
			return
		}

		tm.Print(tm.Color("Done.", tm.GREEN) + "\n")
		tm.FlushAll()
	},
}

// reposRemoveCMD represents the module repositories remove command
var reposRemoveCMD = &cobra.Command{
	Use: "remove",
	Aliases: []string{
		"r",
	},
	Short: "Remove a repository from the global antibuild config file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := globalConfig.LoadDefaultGlobal()
		if err != nil {
			tm.Print(tm.Color("Could not load global config file: "+err.Error(), tm.RED) + "\n")
			tm.FlushAll()
			return
		}

		for i, repo := range globalConfig.DefaultGlobalConfig.Repositories {
			if repo == args[0] {
				globalConfig.DefaultGlobalConfig.Repositories = append(
					globalConfig.DefaultGlobalConfig.Repositories[:i],
					globalConfig.DefaultGlobalConfig.Repositories[i+1:]...)

				err = globalConfig.SaveDefaultGlobal()
				if err != nil {
					tm.Print(tm.Color("Could not save global config file: "+err.Error(), tm.RED) + "\n")
					tm.FlushAll()
					return
				}

				tm.Print(tm.Color("Done.", tm.GREEN) + "\n")
				tm.FlushAll()
				return
			}
		}

		tm.Print(tm.Color("This repository is not in the global config.", tm.RED) + "\n")
		tm.FlushAll()
	},
}
