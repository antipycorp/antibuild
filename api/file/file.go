// Copyright © 2018 - 2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package file

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
)

type (
	//File is a module file used for communicating with modules
	File struct {
		file *os.File
	}
)

var (
	tmproot = os.TempDir() + "/abm/modules"
)

func init() {
	if _, err := os.Stat(tmproot); os.IsNotExist(err) {
		err = os.MkdirAll(tmproot, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func Import(name string) (f File, err error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0600)
	if err != nil {
		return
	}
	f = File{
		file: file,
	}
	return
}

func NewFile(data interface{}) (f File, err error) {
	file, err := ioutil.TempFile(tmproot, "*")
	if err != nil {
		return
	}
	f = File{
		file: file,
	}
	if data == nil {
		return
	}
	err = writeFile(f.file, data)
	if err != nil {
		cleanup(f.file)
		return
	}
	return
}
func (f *File) Update(data interface{}) error {
	f.reset()

	err := writeFile(f.file, data)
	if err != nil {
		cleanup(f.file)
		return err
	}
	return nil
}

func (f *File) Retreive(data interface{}) error {
	f.reset()
	if v, ok := data.(*[]byte); ok {

		buf := bytes.NewBuffer(*v)
		_, err := buf.ReadFrom(f.file)
		*v = buf.Bytes()

		return err
	}

	dec := gob.NewDecoder(f.file)
	err := dec.Decode(data)

	if err != nil {
		cleanup(f.file)
		return err
	}

	return nil
}

func (f *File) GetRef() string {
	return f.file.Name()
}

func (f *File) reset() {
	f.file.Seek(0, 0)
}

func (f *File) Close() {
	if f.file != nil {
		f.file.Close()
	}
}
func (f *File) Cleanup() {
	if f.file != nil {
		cleanup(f.file)
	}
}

func cleanup(file *os.File) {
	file.Close()
	os.Remove(file.Name())
	file = nil
}

func writeFile(file *os.File, data interface{}) error {
	err := file.Truncate(0)
	if err != nil {
		return err
	}

	if v, ok := data.([]byte); ok {
		_, err := file.Write(v)
		return err
	}
	enc := gob.NewEncoder(file)
	err = enc.Encode(data)
	return err
}
