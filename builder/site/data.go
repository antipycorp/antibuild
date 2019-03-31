// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"
	"fmt"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	data struct {
		shouldRange     string
		loader          string
		loaderArguments string
		parser          string
		parserArguments string
		postProcessors  []dataPostProcessor
	}

	dataPostProcessor struct {
		postProcessor          string
		postProcessorArguments string
	}
)

var (
	//ErrNoDataLoaderFound means a data loader does not have a function
	ErrNoDataLoaderFound = errors.NewError("could not get data loader information", 1)
	//ErrNoDataParserFound means a data parser does not have a function
	ErrNoDataParserFound = errors.NewError("could not get data parser information", 2)
	//ErrNoDataPostProcessorFound means a data post processor does not have a function
	ErrNoDataPostProcessorFound = errors.NewError("could not get data post processor information", 3)
	//ErrNoRangeVariable means a data range attribute does not specify a variable
	ErrNoRangeVariable = errors.NewError("could not get variable for range", 3)
)

func (df *data) MarshalJSON() ([]byte, error) {
	return []byte("\"" + df.String() + "\""), nil
}

func (df *data) String() string {
	out := ""

	out += "[" + df.loader
	if df.loaderArguments != "" {
		out += ":" + df.loaderArguments
	}
	out += "]"

	out += "[" + df.parser
	if df.parserArguments != "" {
		out += ":" + df.parserArguments
	}
	out += "]"

	for _, dpp := range df.postProcessors {
		out += "[" + dpp.postProcessor
		if dpp.postProcessorArguments != "" {
			out += ":" + dpp.postProcessorArguments
		}
		out += "]"
	}

	return out
}

func (df *data) UnmarshalJSON(data []byte) error {
	data = data[1 : len(data)-1]

	//get the data from for the dataLoader
	i1 := bytes.Index(data, []byte("["))
	i2 := bytes.Index(data, []byte("]"))

	loaderData := data[i1+1 : i2]
	data = data[i2+1:] //data is used for parser

	{
		//get all the arguments for the loader
		sep := bytes.Split(loaderData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoDataLoaderFound.SetRoot(string(loaderData))
		}

		if string(sep[0]) == "range" {
			if len(sep) < 2 { //only if bigger than 2 this is available
				return ErrNoRangeVariable.SetRoot(string(loaderData))
			}

			df.shouldRange = string(sep[1])

			//get the data from for the dataLoader
			i1 := bytes.Index(data, []byte("["))
			i2 := bytes.Index(data, []byte("]"))

			loaderData = data[i1+1 : i2]
			data = data[i2+1:] //data is used for parser
		}
	}

	{
		//get all the arguments for the loader
		sep := bytes.Split(loaderData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoDataLoaderFound.SetRoot(string(loaderData))
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

	//get the data from for the dataParser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	parserData := data[i1+1 : i2]
	data = data[i2+1:] //keep this in place for potential fututre extentions

	fmt.Println(string(data))
	{
		//get all the arguments for the dataParser
		sep := bytes.Split(parserData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoDataParserFound.SetRoot(string(parserData))
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

	for len(data) > 0 {
		i1 = bytes.Index(data, []byte("["))
		i2 = bytes.Index(data, []byte("]"))

		postProcessorData := data[i1+1 : i2]
		data = data[i2+1:] //keep this in place for potential fututre extentions

		fmt.Println(string(data))

		{
			var dpp = dataPostProcessor{}

			//get all the arguments for the dataPostProcessor
			sep := bytes.Split(postProcessorData, []byte(":"))
			if len(sep) == 0 {
				return ErrNoDataPostProcessorFound.SetRoot(string(postProcessorData))
			}

			var postProcessor = make([]byte, len(sep[0]))
			copy(postProcessor, sep[0])

			//length is not 0 thus everything is fine
			if len(bytes.Split(postProcessor, []byte("_"))) == 1 {
				postProcessor = append(postProcessor, append([]byte("_"), postProcessor...)...)
			}

			dpp.postProcessor = string(postProcessor)
			dpp.postProcessorArguments = ""

			if len(sep) >= 2 { //only if bigger than 2 this is available
				dpp.postProcessorArguments = string(sep[1])
			}

			df.postProcessors = append(df.postProcessors, dpp)
		}
	}

	return nil
}

func getLine() {

}
