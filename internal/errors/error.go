// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package errors

import (
	"fmt"
	"strconv"
)

type (
	//Error is an ierror struct used in antibuild
	Error interface {
		Error() string
		SetRoot(string) Error
		GetRoot() string
		GetCode() string
	}
	ierror struct {
		RootCause string
		message   string
		//Is an ID for the ierror, diferent packages can use the same code, as long as the general cause for the ierror can be derived.
		//"0" is reserved for imported errors
		code string
	}
)

func (e ierror) Error() string {
	if e.message == "" {
		return e.RootCause
	}

	if e.RootCause == "" {
		return e.message
	}

	return fmt.Sprintf("%s: %s", e.message, e.RootCause)
}

//SetRoot sets the root cause and returns a new ierror
func (e ierror) SetRoot(rootcause string) Error {
	e.RootCause = rootcause
	return &e
}

func (e ierror) GetRoot() string {
	return e.RootCause
}

func (e ierror) GetCode() string {
	return e.code
}

//NewError returns a new ierror, RootCause should be set yourself
func NewError(message string, code int) Error {
	return &ierror{
		message: message,
		code:    strconv.Itoa(code),
	}
}

//Import imports an stderr into the Error type
func Import(err error) Error {
	return &ierror{
		RootCause: err.Error(),
		code:      err.Error(),
	}
}
