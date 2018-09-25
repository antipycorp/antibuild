// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/module/client"
)

func main() {
	module := abm.Register("arithmetic")

	module.TemplateFunctionRegister("add", func(w abm.TFRequest, r *abm.TFResponse) {
		var args []int
		var err bool

		if args, err = w.Data.([]int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}

		sum := args[0] + args[1]

		r.Data = sum
		return
	}, &abm.TFTest{
		Request: abm.TFRequest{
			Data: []int{
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
