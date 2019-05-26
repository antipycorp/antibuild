// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package repositories

import (
	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	globalConfig "gitlab.com/antipy/antibuild/cli/configuration/global"
)

// AddCommandRun is the cobra command
func AddCommandRun(command *cobra.Command, args []string) {
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
}
