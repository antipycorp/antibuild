// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package file

import (
	"io"
	"io/ioutil"

	abm "gitlab.com/antipy/antibuild/api/client"
)

//Start starts the file module
func Start(in io.Reader, out io.Writer) {
	module := abm.Register("file")

	module.FileLoaderRegister("file", loadFile)

	module.CustomStart(in, out)
}

func loadFile(w abm.FLRequest, r abm.Response) {
	if w.Variable == "" {
		r.AddInvalid(abm.InvalidInput)
		return
	}

	file, err := ioutil.ReadFile(w.Variable)
	if err != nil {
		r.AddErr(err.Error())
		return
	}
	r.AddData(file)
}
