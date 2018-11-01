// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package file

import (
	"fmt"
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

func loadFile(w abm.FLRequest, r *abm.FLResponse) {
	fmt.Println("new f request")

	if w.Variable == "" {
		r.Error = abm.ErrInvalidInput
		return
	}

	file, err := ioutil.ReadFile(w.Variable)
	if err != nil {
		r.Error = err
		return
	}
	//fmt.Println("new file:", string(file))

	r.Data = file
}
