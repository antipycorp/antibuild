// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("arithmetic")

	module.TemplateFunctionRegister("add", func(w abm.TFRequest, r *abm.TFResponse) {
		var args = make([]int, len(w.Data))
		var err bool

		for i, data := range w.Data {
			if args[i], err = data.(int); err == false {
				r.Error = abm.ErrInvalidInput
				return
			}
		}

		sum := args[0] + args[1]

		r.Data = sum
		return
	}, &abm.TFTest{
		Request: abm.TFRequest{
			Data: []interface{}{
				1,
				2,
			},
		},
		Response: &abm.TFResponse{
			Data: 3,
		},
	})

	module.Start()
}
