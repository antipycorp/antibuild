// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package config

import (
	"encoding/json"
	"os"

	"gitlab.com/antipy/antibuild/cli/api/host"
	"gitlab.com/antipy/antibuild/cli/internal/errors"

	"gitlab.com/antipy/antibuild/cli/builder/site"
)

type (
	//Config is the config struct
	Config struct {
		LogConfig  log                         `json:"logging"`
		Folders    Folder                      `json:"folders"`
		Modules    Modules                     `json:"modules"`
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

	//Modules is the part of the config file that handles modules
	Modules struct {
		Dependencies map[string]string                 `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config,omitempty"`
		SPPs         []string                          `json:"spps,omitempty"`
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
)

//GetConfig gets the config file. DOES NOT CHECK FOR MISSING INFORMATION!!
func GetConfig(configLocation string) (cfg *Config, reterr errors.Error) {
	file, err := os.Open(configLocation)
	defer file.Close()
	if err != nil {
		return nil, ErrFailedOpen.SetRoot(err.Error())
	}
	defer file.Close()

	file.Seek(0, 0)
	dec := json.NewDecoder(file)
	err = dec.Decode(&cfg)
	if err != nil {
		return cfg, ErrFailedParse.SetRoot(err.Error())
	}
	return
}

//SaveConfig saves the config file
func SaveConfig(configLocation string, cfg *Config) errors.Error {
	file, err := os.Create(configLocation + ".new")
	if err != nil {
		return ErrFailedOpen.SetRoot(err.Error())
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cfg)
	if err != nil {
		return ErrFailedWrite.SetRoot(err.Error())
	}

	os.Rename(configLocation+".new", configLocation)

	return nil
}

//CleanConfig does everything for you
func CleanConfig(configLocation string, ui uiLoggerSetter) (*Config, errors.Error) {
	cfg, configErr := ParseConfig(configLocation)
	if configErr != nil {
		return nil, ErrFailedParse.SetRoot(configErr.GetRoot())
	}

	file, err := os.OpenFile(cfg.LogConfig.File, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		return nil, ErrFailedCreateLog.SetRoot(err.Error())
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
		return cfg, ErrFailedParse.SetRoot(err.GetRoot())
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
			PrettyPrint bool   `json:"pretty_print"`
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
