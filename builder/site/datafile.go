// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	datafile struct {
		loader          string
		loaderArguments string
		parser          string
		parserArguments string
	}
)

var (
	//ErrNoFileLoaderFound is when the template failled building
	ErrNoFileLoaderFound = errors.NewError("could not get file loader information", 1)
	//ErrNoFileParserFound is for a faillure moving the static folder
	ErrNoFileParserFound = errors.NewError("could not get file parser information", 2)
)

func (df *datafile) MarshalJSON() ([]byte, error) {
	return []byte(df.String()), nil
}

func (df *datafile) String() string {
	return "[" + df.loader + ":" + df.loaderArguments + "][" + df.parser + ":" + df.parserArguments + "]"
}

func (df *datafile) UnmarshalJSON(data []byte) error {
	//get the data from for the fileLoader
	i1 := bytes.Index(data, []byte("["))
	i2 := bytes.Index(data, []byte("]"))

	loaderData := data[i1+1 : i2]
	data = data[i2+1:] //data is used for parser

	{
		//get all the arguments for the loader
		sep := bytes.Split(loaderData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoFileLoaderFound.SetRoot(string(loaderData))
		}
		var loader = make([]byte, len(sep[0]))
		copy(loader, sep[0])
		//length is not 0 thus everything is fine
		if len(bytes.Split(sep[0], []byte("_"))) == 1 {
			loader = append(sep[0], append([]byte("_"), sep[0]...)...)
		}
		df.loader = string(loader)
		df.loaderArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			df.loaderArguments = string(sep[1])
		}
	}

	//get the data from for the fileParser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	parserData := data[i1+1 : i2]
	data = data[i2+1:] //keep this in place for potential fututre extentions

	{
		//get all the arguments for the fileParser
		sep := bytes.Split(parserData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoFileParserFound.SetRoot(string(parserData))
		}

		var parser = make([]byte, len(sep[0]))
		copy(parser, sep[0])

		//length is not 0 thus everything is fine
		if len(bytes.Split(parser, []byte("_"))) == 1 {
			parser = append(parser, append([]byte("_"), parser...)...)
		}

		df.parser = string(parser)
		df.parserArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			df.parserArguments = string(sep[1])
		}
	}
	return nil
}
