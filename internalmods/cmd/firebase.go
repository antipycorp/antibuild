// +build module,firebase
// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"os"

	modFirebase "gitlab.com/antipy/antibuild/cli/internalmods/firebase"
)

//Start starts the module indicated by build-flags
func Start() {
	modFirebase.Start(os.Stdin, os.Stdout)
}
