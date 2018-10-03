// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"encoding/json"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("json")

	module.FileParserRegister("json", parseJSON, &abm.FPTest{
		Request: abm.FPRequest{
			Data: []byte("{\"test\": \"test\"}"),
		},
		Response: &abm.FPResponse{
			Data: map[string]interface{}{
				"test": "test",
			},
		},
	})

	module.Start()
}

func parseJSON(w abm.FPRequest, r *abm.FPResponse) {
	if w.Data == nil {
		r.Error = abm.ErrInvalidInput
		return
	}

	var jsonData map[string]interface{}

	err := json.Unmarshal(w.Data, &jsonData)
	if err != nil {
		r.Error = err
		return
	}

	r.Data = jsonData
}
