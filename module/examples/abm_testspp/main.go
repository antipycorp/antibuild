// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"
	"os"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

var fileName string

func main() {
	module := abm.Register("testspp")

	module.ConfigFunctionRegister(func(input map[string]interface{}) error {
		var ok bool

		if fileName, ok = input["file_name"].(string); !ok {
			if fileName == "" {
				return abm.ErrInvalidInput
			}

		}

		return nil
	})
	module.SitePostProcessor("testspp", testApp)

	module.Start()
}

func testApp(w abm.SPPRequest, r *abm.SPPResponse) {
	var siteData = w.Data

	if fileName == "" {
		r.Error = abm.ErrNoConfig
		return
	}

	newSite := *siteData[0]
	newSite.Slug = fileName

	siteData = append(siteData, &newSite)

	for _, data := range siteData {
		fmt.Fprint(os.Stderr, data, "\n")
	}

	r.Data = siteData
}
