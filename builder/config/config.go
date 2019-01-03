// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package config

import (
	"encoding/json"
	"io"
	"os"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules"

	"gitlab.com/antipy/antibuild/cli/builder/site"
)

type (
	//Config is the config struct
	Config struct {
		LogConfig  log              `json:"logfile"`
		Folders    Folder           `json:"folders"`
		Modules    Modules          `json:"modules"`
		Pages      *site.ConfigSite `json:"pages"`
		ModuleHost map[string]*host.ModuleHost
		UILogger   UIlogger
	}
	//Folder is the part of the config file that handles folders
	Folder struct {
		Templates string `json:"templates"`
		Static    string `json:"static"`
		Output    string `json:"output"`
		Modules   string `json:"modules"`
	}

	//Modules is the part of the config file that handles modules
	Modules struct {
		Dependencies map[string]string               `json:"dependencies"`
		Config       map[string]modules.ModuleConfig `json:"config"`
		SPPs         []string                        `json:"spps"`
	}
)

var (
	//ErrFailledOpen is when the template failled building
	ErrFailledOpen = errors.NewError("could not open the config file", 1)
	//ErrFailledParse is for a faillure moving the static folder
	ErrFailledParse = errors.NewError("could not parse the config file", 2)
	//ErrFailledWrite is for a faillure in gathering files.
	ErrFailledWrite = errors.NewError("could not write the config file", 3)
	//ErrNoTemplateFolder is for a faillure in gathering files.
	ErrNoTemplateFolder = errors.NewError("template folder not set", 4)
	//ErrNoOutputFolder is for a faillure in gathering files.
	ErrNoOutputFolder = errors.NewError("output folder not set", 5)
	//ErrFailledCreateLog is for a faillure in gathering files.
	ErrFailledCreateLog = errors.NewError("could not open log file", 6)
)

//GetConfig gets the config file. DOES NOT CHECK FOR MISSING INFORMATION!!
func GetConfig(configLocation string) (cfg *Config, reterr errors.Error) {
	configFile, err := os.Open(configLocation)
	defer configFile.Close()
	if err != nil {
		return nil, ErrFailledOpen.SetRoot(err.Error())
	}
	io.Copy(os.Stdout, configFile)
	configFile.Seek(0, 0)
	dec := json.NewDecoder(configFile)
	err = dec.Decode(&cfg)
	if err != nil {
		return cfg, ErrFailledParse.SetRoot(err.Error())
	}
	return
}

//SaveConfig saves the config file
func SaveConfig(configLocation string, cfg *Config) errors.Error {
	file, err := os.Create(configLocation)
	if err != nil {
		return ErrFailledOpen.SetRoot(err.Error())
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(cfg)
	if err != nil {
		return ErrFailledWrite.SetRoot(err.Error())
	}

	return nil
}

//CleanConfig does everything for you
func CleanConfig(configLocation string, ui uiLoggerSetter) (*Config, errors.Error) {
	cfg, configErr := ParseConfig(configLocation)
	if configErr != nil {
		return nil, ErrFailledParse.SetRoot(configErr.GetRoot())
	}

	file, err := os.OpenFile(cfg.LogConfig.File, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		return nil, ErrFailledCreateLog.SetRoot(err.Error())
	}
	file.Seek(0, 0)
	ui.SetLogfile(file)
	ui.SetPrettyPrint(cfg.LogConfig.PrettyPrint)

	cfg.UILogger = ui
	return cfg, nil
}

//ParseConfig parses the config file and check for any missing information
func ParseConfig(configLocation string) (*Config, errors.Error) {
	cfg, err := GetConfig(configLocation)
	if err != nil {
		return cfg, ErrFailledParse.SetRoot(err.GetRoot())
	}

	if cfg.Folders.Templates == "" {
		return cfg, ErrNoTemplateFolder
	}

	if cfg.Folders.Output == "" {
		return cfg, ErrNoOutputFolder
	}

	return cfg, nil
}

func (l *log) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '{': //if it starts with a { its and object and thus should be parsable as a whole
		cfgl := struct {
			File        string `json:"file"`
			PrettyPrint bool   `json:"prettyprint"`
		}{}

		if err := json.Unmarshal(data, &cfgl); err != nil {
			return err
		}
		*l = cfgl //converts cfg to a propper configLog
	default: //else just parse it ad a string
		if err := json.Unmarshal(data, &l.File); err != nil {
			return err
		}
	}
	return nil
}
