// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	//Pipe is a type used for pipes
	Pipe func([]byte) ([]byte, errors.Error)
)

var (
	//ErrFailedCreateFile is for the pipeline failed at creating a new file
	ErrFailedCreateFile = errors.NewError("Failed to create a new temporary file", 0)
	//ErrFailedExecPipe is for the pipeline failed at one of the pipes
	ErrFailedExecPipe = errors.NewError("Failed to execute a pipe in the pipeline", 1)
)

// ExecPipeline is a pipeline executer, data is the input data into the first function,
// retdata is a pointer to where you want the return data to be. Make sure it ais a pointer.
// A pipeline is a set of fuctions that eacht take a file location and process everything based on that.
// For example in modules.go there are .GetPipe methods for all module types, these take the variable
// and return a pipe. At the start of the pipe the provided data is put in, at the end the data is read
// from the same file.
func ExecPipeline(data []byte, pipes ...Pipe) ([]byte, errors.Error) {
	for _, pipe := range pipes {
		var err errors.Error
		data, err = pipe(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
