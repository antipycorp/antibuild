// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package file

import (
	"testing"
)

var (
	data = map[string]string{
		"1": "1",
		"2": "2",
		"3": "3",
		"4": "4",
	}
)

func TestNewFile(t *testing.T) {
	_, err := NewFile(data)

	if err != nil {
		t.FailNow()
	}
}

func TestImport(t *testing.T) {
	file, err := NewFile(data)

	if err != nil {
		t.FailNow()
	}

	_, err = Import(file.GetRef())
	if err != nil {
		t.FailNow()
	}
}

func TestWriteRetreive(t *testing.T) {
	file, err := NewFile(data)

	if err != nil {
		t.FailNow()
	}
	ret := make(map[string]string)

	err = file.Retreive(&ret)

	if err != nil {
		t.FailNow()
	}
}
