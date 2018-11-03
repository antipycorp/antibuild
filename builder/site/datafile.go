package site

import (
	"bytes"
	"errors"
	"fmt"
)

//!!! need to implement post processors !!!
type (
	datafile struct {
		loader          string
		loaderArguments string
		parser          string
		parserArguments string
		postProcessor   string //TODO Is not yet used and will be probably be changed
	}
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

	//get all the arguments for the loader
	sep := bytes.Split(loaderData, []byte(":"))
	if len(sep) == 0 {
		return errors.New("could not get file parser information")
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

	//get the data from for the fileParser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	parserData := data[i1+1 : i2]
	data = data[i2+1:] //data should not be empty after this, but might contain leftover data.

	//get all the arguments for the fileParser
	sep = bytes.Split(parserData, []byte(":"))
	if len(sep) == 0 {
		return errors.New("could not get file parser information")
	}

	var parser = make([]byte, len(sep[0]))
	copy(parser, sep[0])

	//length is not 0 thus everything is fine
	if len(bytes.Split(parser, []byte("_"))) == 1 {
		parser = append(parser, append([]byte("_"), parser...)...)
	}
	//fmt.Println(bytes.Split(sep[0], []byte("_")))

	df.parser = string(parser)
	df.parserArguments = ""

	if len(sep) >= 2 { //only if bigger than 2 this is available
		df.parserArguments = string(sep[1])
	}
	fmt.Println(df.parser)
	fmt.Println(df)
	return nil
}
