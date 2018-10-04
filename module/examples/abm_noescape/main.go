// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"html/template"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("noescape")

	module.TemplateFunctionRegister("html", html, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			"<h1>Test</h1>",
		}}, Response: &abm.TFResponse{
			Data: template.HTML("<h1>Test</h1>"),
		},
	})

	module.TemplateFunctionRegister("js", js, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			"console.log(\"Test\")",
		}}, Response: &abm.TFResponse{
			Data: template.JS("console.log(\"Test\")"),
		},
	})

	module.Start()
}

func html(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]string, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(string); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := template.HTML(args[0])

	r.Data = result
	return
}

func js(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]string, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(string); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := template.JS(args[0])

	r.Data = result
	return
}
