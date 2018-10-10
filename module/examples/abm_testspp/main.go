// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"
	"os"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("testspp")

	module.SitePostProcessor("testspp", testApp)

	module.Start()
}

func testApp(w abm.SPPRequest, r *abm.SPPResponse) {
	var siteData = w.Data

	newSite := *siteData[0]
	newSite.Slug = "/ilikecats.html"

	siteData = append(siteData, &newSite)

	for _, data := range siteData {
		fmt.Fprint(os.Stderr, data, "\n")
	}

	r.Data = siteData
}
