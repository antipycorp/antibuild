// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("testspp")

	module.SitePostProcessor("testspp", parseYAML)

	module.Start()
}

func testApp(w abm.SPPRequest, r *abm.SPPResponse) {
	var siteData = w.Data

	siteData
		r.Data = siteData
}
