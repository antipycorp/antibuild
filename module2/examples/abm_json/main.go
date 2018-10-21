// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"encoding/json"
	"fmt"
	"os"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("json")

	module.FileParserRegister("json", parseJSON)

	module.Start()
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
