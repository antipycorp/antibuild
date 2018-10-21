// +build module,noescape

// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"os"

	modNoESC "gitlab.com/antipy/antibuild/cli/internalmods/noescape"
)

//Start starts the module indicated by build-flags
func Start() {
	modNoESC.Start(os.Stdin, os.Stdout)
}
