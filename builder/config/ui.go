// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package config

import "io"

type (
	log struct {
		File        string `json:"file"`
		PrettyPrint bool   `json:"pretty_print"`
		LogDebug    bool   `json:"log_debug"`
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
		ShouldLogDebug(bool)
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
