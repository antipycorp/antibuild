// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"

	"gitlab.com/antipy/antibuild/cli/cli"
)

const version = "v0.3.0"

func main() {
	fmt.Println("Antibuild by Antipy  ", version)
	cli.Execute()
}
