// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	// Data is info about data from the config
	Data struct {
		ShouldRange     string
		Loader          string
		LoaderArguments string
		Parser          string
		ParserArguments string
		PostProcessors  []DataPostProcessor
	}

	// DataPostProcessor is info about a data post processor from the config
	DataPostProcessor struct {
		PostProcessor          string
		PostProcessorArguments string
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

// MarshalJSON marshalls the data
func (df *Data) MarshalJSON() ([]byte, error) {
	return []byte("\"" + df.String() + "\""), nil
}

func (df *Data) String() string {
	out := ""

	out += "[" + df.Loader
	if df.LoaderArguments != "" {
		out += ":" + df.LoaderArguments
	}
	out += "]"

	out += "[" + df.Parser
	if df.ParserArguments != "" {
		out += ":" + df.ParserArguments
	}
	out += "]"

	for _, dpp := range df.PostProcessors {
		out += "[" + dpp.PostProcessor
		if dpp.PostProcessorArguments != "" {
			out += ":" + dpp.PostProcessorArguments
		}
		out += "]"
	}

	return out
}

// UnmarshalJSON unmarshalls the data
func (df *Data) UnmarshalJSON(data []byte) error {
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

			df.ShouldRange = string(sep[1])

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
		df.Loader = string(loader)
		df.LoaderArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			df.LoaderArguments = string(sep[1])
		}
	}

	//get the data from for the dataParser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	parserData := data[i1+1 : i2]
	data = data[i2+1:] //keep this in place for potential fututre extentions

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

		df.Parser = string(parser)
		df.ParserArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			df.ParserArguments = string(sep[1])
		}
	}

	for len(data) > 0 {
		i1 = bytes.Index(data, []byte("["))
		i2 = bytes.Index(data, []byte("]"))

		postProcessorData := data[i1+1 : i2]
		data = data[i2+1:] //keep this in place for potential fututre extentions

		{
			var dpp = DataPostProcessor{}

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

			dpp.PostProcessor = string(postProcessor)
			dpp.PostProcessorArguments = ""

			if len(sep) >= 2 { //only if bigger than 2 this is available
				dpp.PostProcessorArguments = string(sep[1])
			}

			df.PostProcessors = append(df.PostProcessors, dpp)
		}
	}

	return nil
}

func getLine() {

}
