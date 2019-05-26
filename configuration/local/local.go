// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package local

import (
	"encoding/json"
	"io"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/engine/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	//Config is the config struct
	Config struct {
		LogConfig  log                         `json:"logging"`
		Folders    Folder                      `json:"folders"`
		Modules    modules.Modules             `json:"modules"`
		Pages      *site.ConfigSite            `json:"pages"`
		ModuleHost map[string]*host.ModuleHost `json:"-"`
		UILogger   UIlogger                    `json:"-"`
	}
	//Folder is the part of the config file that handles folders
	Folder struct {
		Templates string `json:"templates"`
		Static    string `json:"static"`
		Output    string `json:"output"`
		Modules   string `json:"modules"`
	}

	// This all has to do with the UI part of this program

	log struct {
		File        string `json:"file"`
		PrettyPrint bool   `json:"pretty_print"`
		EnableDebug bool   `json:"enable_debug"`
	}

	//UIlogger combines a UI and a logger
	UIlogger interface {
		ui
		Logger
	}

	ui interface {
		ShowResult()
	}

	//Logger is the logger we use
	Logger interface {
		Info(string)
		Infof(string, ...interface{})
		Error(string)
		Errorf(string, ...interface{})
		Fatal(string)
		Fatalf(string, ...interface{})
		Debug(string)
		Debugf(string, ...interface{})
		ShouldEnableDebug(bool)
	}

	uiLoggerSetter interface {
		logfileSetter
		prettylogSetter
		UIlogger
	}

	logfileSetter interface {
		SetLogfile(io.Writer)
	}

	prettylogSetter interface {
		SetPrettyPrint(bool)
	}
)

var (
	//ErrFailedOpen is when the template failed building
	ErrFailedOpen = errors.NewError("could not open the config file", 1)

	//ErrFailedParse is for a failure moving the static folder
	ErrFailedParse = errors.NewError("could not parse the config file", 2)

	//ErrFailedWrite is for a failure in gathering files.
	ErrFailedWrite = errors.NewError("could not write the config file", 3)

	//ErrNoTemplateFolder is for a failure in gathering files.
	ErrNoTemplateFolder = errors.NewError("template folder not set", 4)

	//ErrNoOutputFolder is for a failure in gathering files.
	ErrNoOutputFolder = errors.NewError("output folder not set", 5)

	//ErrFailedCreateLog is for a failure in gathering files.
	ErrFailedCreateLog = errors.NewError("could not open log file", 6)

	//ErrDependencyWrongFormat means a wrong format for a dependency
	ErrDependencyWrongFormat = errors.NewError("dependency must be in the format 'json' or 'json@1.0.0'", 101)
)

func (l *log) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '{': //if it starts with a { its and object and thus should be parsable as a whole
		cfgl := struct {
			File        string `json:"file"`
			PrettyPrint bool   `json:"pretty_print"`
			EnableDebug bool   `json:"enable_debug"`
		}{}

		if err := json.Unmarshal(data, &cfgl); err != nil {
			return err
		}

		*l = cfgl //converts cfg to a proper configLog
	default: //else just parse it add a string
		if err := json.Unmarshal(data, &l.File); err != nil {
			return err
		}
	}
	return nil
}
