// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package repositories

import (
	tm "github.com/lucacasonato/goterm"
	"github.com/spf13/cobra"
	globalConfig "gitlab.com/antipy/antibuild/cli/configuration/global"
)

// ListCommandRun is the cobra command
func ListCommandRun(command *cobra.Command, args []string) {
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
}
