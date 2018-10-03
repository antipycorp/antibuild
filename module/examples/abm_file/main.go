// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"io/ioutil"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("file")

	module.FileLoaderRegister("file", loadFile)

	module.Start()
}

func loadFile(w abm.FLRequest, r *abm.FLResponse) {
	if w.Variable == "" {
		r.Error = abm.ErrInvalidInput
		return
	}

	file, err := ioutil.ReadFile(w.Variable)
	if err != nil {
		r.Error = err
		return
	}

	r.Data = file
}
