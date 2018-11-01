// +build module,language

// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cmd

import (
	"os"

	modLang "gitlab.com/antipy/antibuild/cli/internalmods/language"
)

//Start starts the module indicated by build-flags
func Start() {
	modLang.Start(os.Stdin, os.Stdout)
}
