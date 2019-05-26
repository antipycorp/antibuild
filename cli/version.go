// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const releaseType = "alpha"
const version = "0.13.0"

func versionCommandRun(command *cobra.Command, arguments []string) {
	fmt.Printf("antibuild %s/%s\n", releaseType, version)
}
