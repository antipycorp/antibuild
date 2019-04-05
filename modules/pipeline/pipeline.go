// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package pipeline

import (
	"gitlab.com/antipy/antibuild/api/file"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	//Pipe is a type used for pipes
	Pipe func(string) errors.Error
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
func ExecPipeline(data interface{}, retdata interface{}, pipes ...Pipe) errors.Error {
	var f file.File
	if data == nil {
		var err error
		f, err = file.NewFile([]byte(""))
		if err != nil {
			return ErrFailedCreateFile.SetRoot(err.Error())
		}
	} else {
		var err error
		f, err = file.NewFile(data)
		if err != nil {
			return ErrFailedCreateFile.SetRoot(err.Error())
		}
	}

	defer f.Cleanup()
	fileName := f.GetRef()
	for _, pipe := range pipes {
		err := pipe(fileName)
		if err != nil {
			return err
		}
	}

	f.Retreive(retdata)
	return nil
}
