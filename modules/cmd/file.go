// +build module,file

// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"os"

	modFile "gitlab.com/antipy/antibuild/cli/internalmods/file"
)

//Start starts the module indicated by build-flags
func Start() {
	modFile.Start(os.Stdin, os.Stdout)
}
