// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/module/client"
)

func main() {
	module := abm.Register("arithmetic")

	module.TemplateFunctionRegister("add", func(w abm.Request, r *abm.Response) {
		args := w.Data

		// since all data coming in is of type interface and we expect 2 intergers, we have to convert the types
		a1, ok1 := args[0].(int)
		a2, ok2 := args[1].(int)
		if !ok1 || !ok2 {
			r.Error = abm.ErrInvalidInput
			return
		}

		sum := a1 + a2
		r.Data = sum
		return
	})
}
