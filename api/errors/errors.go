// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package errors

import "encoding/gob"

//Error is an error used for communication over the protocol
type Error struct {
	Message string
	Code    int
}

const (
	//CodeFatal is for when the module had a fatal error
	CodeFatal = iota
	//CodeError is for when the module had an error
	CodeError
	//CodeInfo is for when the module wants to send info
	CodeInfo
	//CodeDebug is for when the module wants to send debug info
	CodeDebug
	//CodeProtocolFailure is for when the communication has failed
	CodeProtocolFailure
	//CodeInvalidResponse is for when the response datatype is not valid
	CodeInvalidResponse
	//exlusive highest value of error codes
	maxCodeValue
)

func init() {
	gob.Register(Error{})
}

//New returns a new error
func New(err string, code int) Error {
	return Error{
		Message: err,
		Code:    code,
	}
}

func (e Error) Error() string {
	return e.Message
}
