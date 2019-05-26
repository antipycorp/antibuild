// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/cmd/modules"
	"gitlab.com/antipy/antibuild/cli/cmd/modules/repositories"
)

// $ antibuild
var antibuildCommand = &cobra.Command{
	Use:   "antibuild",
	Short: "A fast and simple static site generator with module support.",
	Long: `Antibuild is a static site generator that can use dynamic datasets
and simple or advanced modules for endless configurability.

To start a new antibuild project run "antibuild new"
Antibuild is written in Golang and can be extended by modules written in Golang.
To get started with modules go to https://antibuild.io/modules.`,
}

// $ antibuild version
var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of antibuild",
	Run:   versionCommandRun,
}

// $ antibuild new
var newCommand = &cobra.Command{
	Use:   "new",
	Short: "Make a new antibuild project",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Run:   newCommandRun,
}

// $ antibuild build
var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Build a project",
	Long:  `Build a Antibuild project and export into the output folder.`,
	Run:   buildCommandRun,
}

// $ antibuild develop
var developCommand = &cobra.Command{
	Use:   "develop",
	Short: "Develop a project using the config file",
	Long:  `Develop a Antibuild project and export into the output folder.`,
	Run:   developCommandRun,
}

// $ antibuild modules
var modulesCommand = &cobra.Command{
	Use:     "modules",
	Aliases: []string{"m"},
	Short:   "Manage your antibuild modules",
	Long:    `Used to manage your modules for antibuild. Run a subcommand to get more info.`,
}

// $ antibuild modules add {module_id}
var modulesAddCommand = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Get a module",
	Long:    `Adds and downloads a module. Uses the standard repository by default. Will use repos in the global config if not found in std. Use -m to force a repo.`,
	Args:    cobra.ExactArgs(1),
	Run:     modules.AddCommandRun,
}

// $ antibuild modules remove {module_id}
var modulesRemoveCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"r"},
	Short:   "Remove a module",
	Long:    `Removes and deletes a module.`,
	Args:    cobra.ExactArgs(1),
	Run:     modules.RemoveCommandRun,
}

// $ antibuild modules install
var modulesInstallCommand = &cobra.Command{
	Use:     "install",
	Aliases: []string{"i"},
	Short:   "Install all modules defined in the config file.",
	Long:    `Will install all modules defined in the config file at the right versions and OS/ARCH.`,
	Run:     modules.InstallCommandRun,
}

// $ antibuild modules repositories
var repositoriesCommand = &cobra.Command{
	Use:     "repositories",
	Aliases: []string{"repos"},
	Short:   "Manage your antibuild module repositories",
	Long:    `Used to manage your module repositories for antibuild.`,
}

// $ antibuild modules repositories list
var repositoriesListCommand = &cobra.Command{
	Use:     "list",
	Aliases: []string{"list"},
	Short:   "List all repositories in the global antibuild config file.",
	Run:     repositories.ListCommandRun,
}

// $ antibuild modules repositories add {repo_url}
var repositoriesAddCommand = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a repository to the global antibuild config file.",
	Long:    `Adds a repository that is used to try to pull modules from when adding modules.`,
	Args:    cobra.ExactArgs(1),
	Run:     repositories.AddCommandRun,
}

// $ antibuild modules repositories remove {repo_url}
var repositoriesRemoveCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"r"},
	Short:   "Remove a repository from the global antibuild config file.",
	Args:    cobra.ExactArgs(1),
	Run:     repositories.RemoveCommandRun,
}

// Run the cli
func Run() {
	// helper commands
	antibuildCommand.AddCommand(versionCommand, newCommand)

	// building commands
	antibuildCommand.AddCommand(buildCommand, developCommand)

	// module commands
	repositoriesCommand.AddCommand(repositoriesListCommand, repositoriesAddCommand, repositoriesRemoveCommand)
	modulesCommand.AddCommand(repositoriesCommand)
	modulesCommand.AddCommand(modulesAddCommand, modulesRemoveCommand, modulesInstallCommand)
	antibuildCommand.AddCommand(modulesCommand)

	antibuildCommand.Execute()
}
