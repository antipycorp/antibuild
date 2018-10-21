package site

import (
	"bytes"
	"errors"
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

	loader := data[i1+1 : i2]
	data = data[i2+1:]

	//get all the arguments for the loader
	sep := bytes.Split(loader, []byte(":"))
	if len(sep) >= 2 {
		df.loader = string(sep[0])
		df.loaderArguments = string(sep[1])
	} else if len(sep) == 1 {
		df.loader = string(sep[0])
		df.loaderArguments = ""
	} else {
		return errors.New("could not get file loader information")
	}

	//get the data from for the fileParser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	parser := data[i1+1 : i2]
	data = data[i2+1:]

	//get all the arguments for the fileParser
	sep = bytes.Split(parser, []byte(":"))
	if len(sep) >= 2 {
		df.parser = string(sep[0])
		df.parserArguments = string(sep[1])
	} else if len(sep) == 1 {
		df.parser = string(sep[0])
		df.parserArguments = ""
	} else {
		return errors.New("could not get file parser information")
	}

	return nil
}
