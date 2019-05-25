// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package local

import (
	"encoding/json"
	"os"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

//GetConfig gets the config file. DOES NOT CHECK FOR MISSING INFORMATION!!
func GetConfig(configLocation string) (cfg *Config, reterr errors.Error) {
	file, err := os.Open(configLocation)
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

//CleanConfig does everything for you
func CleanConfig(configLocation string, ui uiLoggerSetter, useLog bool) (*Config, errors.Error) {
	cfg, configErr := ParseConfig(configLocation)
	if configErr != nil {
		return nil, ErrFailedParse.SetRoot(configErr.GetRoot())
	}

	if !useLog {
		return cfg, nil
	}
	file, err := os.OpenFile(cfg.LogConfig.File, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		return nil, ErrFailedCreateLog.SetRoot(err.Error())
	}
	file.Seek(0, 0)
	ui.SetLogfile(file)
	ui.SetPrettyPrint(cfg.LogConfig.PrettyPrint)
	ui.ShouldEnableDebug(cfg.LogConfig.EnableDebug)

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
