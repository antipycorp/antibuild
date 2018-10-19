// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

//Start starts the file module
func Start(in io.Reader, out io.Writer) {
	module := abm.Register("json")

	module.FileParserRegister("json", parseJSON)

	module.CustomStart(in, out)
}

func parseJSON(w abm.FPRequest, r *abm.FPResponse) {
	if w.Data == nil {
		r.Error = abm.ErrInvalidInput
		fmt.Fprintln(os.Stderr, abm.ErrInvalidInput)
		return
	}

	var jsonData map[string]interface{}

	err := json.Unmarshal(w.Data, &jsonData)
	if err != nil {
		r.Error = err
		fmt.Fprintln(os.Stderr, err)
		return
	}

	r.Data = jsonData
}
