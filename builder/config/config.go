package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
		UILogger   uilogger
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
		Dependencies       map[string]string               `json:"dependencies"`
		Config             map[string]modules.ModuleConfig `json:"config"`
		SitePostProcessors []string                        `json:"site_post_processors"`
	}

	log struct {
		File       string `json:"file"`
		PretyPrint bool   `json:"pretyprint"`
	}

	uilogger interface {
		ui1
		logger
	}

	ui1 interface {
		ShowCompiling()
		ShowResult()
	}

	logger interface {
		Info(err string)
		Error(err string)
		Fatal(err string)
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
		fmt.Println(err)
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(cfg)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (l *log) UnmarshalJSON(data []byte) error {
	defer fmt.Println(l.File)
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
