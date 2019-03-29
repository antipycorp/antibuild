// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	iterator struct {
		iterator          string
		iteratorArguments string
	}
)

var (
	//ErrNoIteratorFound is when the template failed building
	ErrNoIteratorFound = errors.NewError("could not get iterator information", 3)
)

func (i *iterator) MarshalJSON() ([]byte, error) {
	return []byte("\"" + i.String() + "\""), nil
}

func (i *iterator) String() string {
	out := ""

	out += "[" + i.iterator
	if i.iteratorArguments != "" {
		out += ":" + i.iteratorArguments
	}
	out += "]"

	return out
}

func (i *iterator) UnmarshalJSON(data []byte) error {
	//get the data from for the fileLoader
	i1 := bytes.Index(data, []byte("["))
	i2 := bytes.Index(data, []byte("]"))

	iteratorData := data[i1+1 : i2]

	{
		//get all the arguments for the loader
		sep := bytes.Split(iteratorData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoFileLoaderFound.SetRoot(string(iteratorData))
		}
		var loader = make([]byte, len(sep[0]))
		copy(loader, sep[0])

		if len(bytes.Split(sep[0], []byte("_"))) == 1 {
			loader = append(sep[0], append([]byte("_"), sep[0]...)...)
		}

		i.iterator = string(loader)
		i.iteratorArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			i.iteratorArguments = string(sep[1])
		}
	}

	return nil
}
