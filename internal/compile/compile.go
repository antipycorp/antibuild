// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package compile

import (
	"os"
	"os/exec"
	"path/filepath"
)

// FromSource compiles a go program
func FromSource(path string, outFile string) error {
	compilecmd := exec.Command("go", "build", "-o", outFile, filepath.Join(path, "main.go"))
	compilecmd.Stderr = os.Stderr
	err := compilecmd.Run()
	if err != nil {
		return err
	}

	return nil
}
