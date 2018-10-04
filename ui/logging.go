package ui

import (
	"fmt"
	"os"
	"time"
)

var (
	logfile *os.File
)

var (
	//DataFolder is an error
	DataFolder = Error{
		extensive: "The data folder is not set or valid",
		short:     "config/data",
	}

	//TemplateFolder is an error
	TemplateFolder = Error{
		extensive: "The template folder is not set or valid",
		short:     "config/template",
	}

	//OutputFolder is an error
	OutputFolder = Error{
		extensive: "The output folder is not set or valid",
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

//LogImportant an important an error and shows it on screen
func (ui *UI) LogImportant(err Error, page string, line string, data []interface{}) {
	fmt.Fprintf(logfile, "[%v] %v: %v [page: %v; line %v]", time.Now().String(), err.short, fmt.Sprintf(err.extensive, data), page, line)
	ui.showError(err, page, line, data)
}

//LogInfo an information that only gets logged to the log file
func (ui *UI) LogInfo(info string) {
	fmt.Fprintf(logfile, "[%v] %v", time.Now().String(), info)
}
