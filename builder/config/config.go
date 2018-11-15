package config

import (
	"encoding/json"
	"errors"
	"os"

	"gitlab.com/antipy/antibuild/api/host"
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

	log struct {
		File       string `json:"file"`
		PretyPrint bool   `json:"pretyprint"`
	}

	UIlogger interface {
		ui1
		Logger
	}

	ui1 interface {
		ShowCompiling()
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
	}
)

//GetConfig gets the config file. DOES NOT CHECK FOR MISSIN INFORMATION!!
func GetConfig(configLocation string) (cfg *Config, err error) {
	configFile, err := os.Open(configLocation)
	defer configFile.Close()
	if err != nil {
		return nil, errors.New("could not open the config file: " + err.Error())
	}

	dec := json.NewDecoder(configFile)
	err = dec.Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	return
}

//SaveConfig saves the config file
func SaveConfig(configLocation string, cfg *Config) (err error) {
	file, err := os.Create(configLocation)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (l *log) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '{':
		cfgl := struct {
			File       string `json:"file"`
			PretyPrint bool   `json:"pretyprint"`
		}{}
		if err := json.Unmarshal(data, &cfgl); err != nil {
			return err
		}
		*l = cfgl //converts cfg to a propper configLog
	default:
		if err := json.Unmarshal(data, &l.File); err != nil {
			return err
		}
	}
	return nil
}
