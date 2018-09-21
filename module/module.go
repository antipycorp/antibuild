package module

import "errors"

type (
	//Module is the collection of registered events that the module API should react to.
	Module struct {
		name string

		RegisterTemplateFunction func(identifier string, function func(Request) *Response)
	}

	//Request is the request with data and meta from the module caller.
	Request struct {
		Error error
		Data  interface{}
	}

	//Response is the response to the module API that will be used to respond to the client
	Response struct {
		Error error
		Data  interface{}
	}
)

var (
	//ErrInvalidInput is the error that occurs when a function is called with data that is not correct, valid or applicable.
	ErrInvalidInput = errors.New("module: the provided data is invalid")

	//ErrFailed is the error that occurs when the module experiences an internal error.
	ErrFailed = errors.New("module: internal processing error")
)
