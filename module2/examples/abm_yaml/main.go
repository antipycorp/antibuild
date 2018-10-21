// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/cli/module/client"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	module := abm.Register("yaml")

	module.FileParserRegister("yaml", parseYAML)
	module.Start()
}

func parseYAML(w abm.FPRequest, r *abm.FPResponse) {
	if w.Data == nil {
		r.Error = abm.ErrInvalidInput
		return
	}

	var yamlData map[interface{}]interface{}

	err := yaml.Unmarshal(w.Data, &yamlData)
	if err != nil {
		r.Error = err
		return
	}

	var yamlOutData = make(map[string]interface{})

	for index, data := range yamlData {
		strIndex, err := index.(string)
		if err == true {
			yamlOutData[strIndex] = data
		}
	}

	r.Data = yamlOutData
}
