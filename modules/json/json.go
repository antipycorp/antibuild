// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package json

import (
	"encoding/json"
	"io"

	abm "gitlab.com/antipy/antibuild/api/client"
)

//Start starts the file module
func Start(in io.Reader, out io.Writer) {
	module := abm.Register("json")

	module.FileParserRegister("json", parseJSON)

	module.CustomStart(in, out)
}

func parseJSON(w abm.FPRequest, r abm.Response) {
	if w.Data == nil {
		r.AddInvalid(abm.InvalidInput)
		return
	}

	var jsonData map[string]interface{}

	err := json.Unmarshal(w.Data, &jsonData)
	if err != nil {
		r.AddErr(err.Error())
		return
	}
	r.AddData(jsonData)
}
