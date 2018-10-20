package site

import (
	"bytes"
	"errors"
)

//!!! need to implement post processors !!!
type (
	datafile struct {
		loader        string
		file          string
		parser        string
		parseArg      string
		postprocessor string //TODO Is not yet used and will be probably be changed
	}
)

func (df *datafile) MarshalJSON() ([]byte, error) {
	return []byte(df.String()), nil
}
func (df *datafile) String() string {
	return "[" + df.loader + ":" + df.file + "][" + df.parser + "]"
}

func (df *datafile) UnmarshalJSON(data []byte) error {
	//get the data from for the fileloader
	i1 := bytes.Index(data, []byte("["))
	i2 := bytes.Index(data, []byte("]"))

	loader := data[i1+1 : i2]
	data = data[i2+1:]
	//get all the arguments for the loader
	sep := bytes.Split(loader, []byte(":"))
	if len(sep) >= 2 {
		df.loader = string(sep[0])
		df.file = string(sep[1])
	} else if len(sep) == 1 {
		df.loader = string(sep[0])
		df.file = ""
	} else {
		return errors.New("could not get file loader information")
	}

	//get the data from for the fileparser
	i1 = bytes.Index(data, []byte("["))
	i2 = bytes.Index(data, []byte("]"))

	loader = data[i1+1 : i2]
	data = data[i2+1:]
	//get all the arguments for the loader
	sep = bytes.Split(loader, []byte(":"))
	if len(sep) >= 2 {
		df.loader = string(sep[0])
		df.file = string(sep[1])
	} else if len(sep) == 1 {
		df.loader = string(sep[0])
		df.file = ""
	} else {
		return errors.New("could not get file loader information")
	}

	return nil
}
