package main

import (
	"reflect"

	abm "gitlab.com/antipy/antibuild/module"
)

func main() {
	module := abm.Register("arithmetic")

	module.TemplateFunctionRegister("add", func(w abm.Request, r *abm.Response) {
		if reflect.TypeOf(w.Data).String() != "[]interface{}" {
			r.Error = abm.ErrInvalidInput
			return
		}

		args := w.Data.([]interface{})

		if reflect.TypeOf(args[0]).String() != "int" || reflect.TypeOf(args[1]).String() != "int" {
			r.Error = abm.ErrInvalidInput
			return
		}

		sum := args[0].(int) + args[1].(int)
		r.Data = sum
		return
	})
}
