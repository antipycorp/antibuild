// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

func main() {
	module := abm.Register("math")

	module.TemplateFunctionRegister("add", add, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			1,
			2,
		}}, Response: &abm.TFResponse{
			Data: 3,
		},
	})

	module.TemplateFunctionRegister("subtract", subtract, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			3,
			1,
		}}, Response: &abm.TFResponse{
			Data: 2,
		},
	})
	module.TemplateFunctionRegister("multiply", multiply, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			3,
			2,
		}}, Response: &abm.TFResponse{
			Data: 6,
		},
	})

	module.TemplateFunctionRegister("divide", divide, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			6,
			2,
		}}, Response: &abm.TFResponse{
			Data: 3,
		},
	})

	module.TemplateFunctionRegister("modulo", modulo, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			5,
			2,
		}}, Response: &abm.TFResponse{
			Data: 1,
		},
	})

	module.TemplateFunctionRegister("power", power, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			3,
			3,
		}}, Response: &abm.TFResponse{
			Data: 27,
		},
	})

	module.Start()
}

func add(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := args[0] + args[1]

	r.Data = result
	return
}

func subtract(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := args[0] - args[1]

	r.Data = result
	return
}

func multiply(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := args[0] * args[1]

	r.Data = result
	return
}

func divide(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := args[0] / args[1]

	r.Data = result
	return
}

func modulo(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := args[0] % args[1]

	r.Data = result
	return
}

func power(w abm.TFRequest, r *abm.TFResponse) {
	var args = make([]int, len(w.Data))
	var err bool

	for i, data := range w.Data {
		if args[i], err = data.(int); err == false {
			r.Error = abm.ErrInvalidInput
			return
		}
	}

	result := 1

	for index := 0; index < args[1]; index++ {
		result = result * args[0]
	}

	r.Data = result
	return
}
