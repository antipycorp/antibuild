// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/kirsle/configdir"
	"gitlab.com/antipy/antibuild/api/host"
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

	// Modules is the part of the config file that handles modules
	Modules struct {
		Dependencies map[string]*Module                `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config,omitempty"`
		SPPs         []string                          `json:"spps,omitempty"`
	}

	// Module with info about the path and version
	Module struct {
		Repository string
		Version    string
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

		*l = cfgl //converts cfg to a propper configLog
	default: //else just parse it add a string
		if err := json.Unmarshal(data, &l.File); err != nil {
			return err
		}
	}
	return nil
}

// LoadDefaultGlobal config file
func LoadDefaultGlobal() error {
	configPath := configdir.LocalConfig("antibuild")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")

	DefaultGlobalConfig, err = OpenGlobalConfig(configFile)
	if err != nil {
		return err
	}

	return nil
}

// SaveDefaultGlobal config file
func SaveDefaultGlobal() error {
	configPath := configdir.LocalConfig("antibuild")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")

	err = DefaultGlobalConfig.Save(configFile)
	if err != nil {
		return err
	}

	return nil
}

// DefaultGlobalConfig that gets auto opened
var DefaultGlobalConfig *GlobalConfig

// GlobalConfig is a global antibuild configuration
type GlobalConfig struct {
	Repositories []string `json:"repositories"`
}

// OpenGlobalConfig opens and parses a global config file
func OpenGlobalConfig(path string) (*GlobalConfig, error) {
	c := new(GlobalConfig)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			c = &GlobalConfig{Repositories: []string{}}
			err = c.Save(path)
			if err != nil {
				return nil, err
			}
			return OpenGlobalConfig(path)
		}

		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Save the global config
func (c *GlobalConfig) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(c)
	if err != nil {
		return err
	}

	return nil
}

// UnmarshalJSON on a module
func (m *Module) UnmarshalJSON(data []byte) error {
	d := string(data)
	d = strings.Trim(d, "\"")
	split := strings.Split(d, "@")

	if len(split) < 1 || len(split) > 2 {
		return ErrDependencyWrongFormat
	}

	if split[0] == "" {
		return ErrDependencyWrongFormat
	}

	m.Repository = split[0]

	if len(split) == 2 {
		if split[1] == "" {
			return ErrDependencyWrongFormat
		}
		m.Version = split[1]
	} else {
		m.Version = "latest"
	}

	return nil
}

// ParseModuleString for config and cli
func ParseModuleString(moduleString string) (m *Module, err errors.Error) {
	m = new(Module)

	d := moduleString
	d = strings.Trim(d, "\"")
	split := strings.SplitN(d, "@", -1)

	if len(split) < 1 || len(split) > 2 {
		err = ErrDependencyWrongFormat
		return
	}

	if split[0] == "" {
		err = ErrDependencyWrongFormat
		return
	}

	m.Repository = split[0]

	if len(split) == 2 {
		if split[1] == "" {
			err = ErrDependencyWrongFormat
			return
		}

		m.Version = split[1]
	} else {
		m.Version = "latest"
	}

	return
}

// MarshalJSON on a module
func (m *Module) MarshalJSON() ([]byte, error) {
	v := ""
	if m.Version != "" {
		v = "@" + m.Version
	}

	return []byte("\"" + m.Repository + v + "\""), nil
}
