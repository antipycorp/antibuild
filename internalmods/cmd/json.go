// +build module,json
// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"os"

	modJSON "gitlab.com/antipy/antibuild/cli/internalmods/json"
)

//Start starts the module indicated by build-flags
func Start() {
	modJSON.Start(os.Stdin, os.Stdout)
}
