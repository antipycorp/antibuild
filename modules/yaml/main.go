// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package yaml

import (
	"io"

	abm "gitlab.com/antipy/antibuild/api/client"
	yaml "gopkg.in/yaml.v2"
)

//Start starts the yaml module
func Start(in io.Reader, out io.Writer) {
	module := abm.Register("yaml")

	module.FileParserRegister("yaml", parseYAML)

	module.CustomStart(in, out)
}

func parseYAML(w abm.FPRequest, r abm.Response) {
	if w.Data == nil {
		r.AddInvalid(abm.InvalidInput)
		return
	}

	var yamlData map[interface{}]interface{}

	err := yaml.Unmarshal(w.Data, &yamlData)
	if err != nil {
		r.AddErr(err.Error())
		return
	}

	r.AddData(yamlData)
	return
}
