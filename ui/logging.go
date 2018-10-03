package ui

import (
	"os"
)

var (
	logfile *os.File
)

var (
	//DataFolder is an error
	DataFolder = Error{
		extensive: "The data folder is not set",
		short:     "config/data",
	}

	//TemplateFolder is an error
	TemplateFolder = Error{
		extensive: "The template folder is not set",
		short:     "config/template",
	}

	//OutputFolder is an error
	OutputFolder = Error{
		extensive: "The output folder is not set",
		short:     "config/output",
	}
)

func init() {
	var err error
	logfile, err = os.Create("antibuild.log")
	if err != nil {
		panic("Couldn't start logger file.   " + err.Error())
	}
}

//Log logs an error
func (ui *UI) Log(err Error, page string, line string, data []interface{}) {

}
