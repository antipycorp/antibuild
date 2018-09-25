// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"

	"gitlab.com/antipy/antibuild/cli"
)

const version = "v0.2.0"

func main() {
	fmt.Println(version)
	cli.Execute()
}
