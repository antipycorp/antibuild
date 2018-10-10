// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"gitlab.com/antipy/antibuild/cli/builder/site"
	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("testspp")

	module.SitePostProcessor("testspp", testApp)

	module.Start()
}

func testApp(w abm.SPPRequest, r *abm.SPPResponse) {
	var siteData = w.Data

	siteData.Sites = append(siteData.Sites, &site.Site{})
	r.Data = siteData
}
