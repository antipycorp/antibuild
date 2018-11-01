package builder

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
		LogConfig  configLog        `json:"logfile"`
		Folders    ConfigFolder     `json:"folders"`
		Modules    ConfigModules    `json:"modules"`
		Pages      *site.ConfigSite `json:"pages"`
		moduleHost map[string]*host.ModuleHost
		uilogger   uilogger
	}
	//ConfigFolder is the part of the config file that handles folders
	ConfigFolder struct {
		Templates string `json:"templates"`
		Static    string `json:"static"`
		Output    string `json:"output"`
		Modules   string `json:"modules"`
	}

	//ConfigModules is the part of the config file that handles modules
	ConfigModules struct {
		Dependencies       map[string]string               `json:"dependencies"`
		Config             map[string]modules.ModuleConfig `json:"config"`
		SitePostProcessors []string                        `json:"site_post_processors"`
	}

	configLog struct {
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
func GetConfig(configLocation string) (config *Config, err error) {
	configFile, err := os.Open(configLocation)
	defer configFile.Close()
	if err != nil {
		return nil, errors.New("could not open the config file: " + err.Error())
	}

	dec := json.NewDecoder(configFile)
	err = dec.Decode(&config)
	if err != nil {
		return config, err
	}
	return
}

//SaveConfig saves the config file
func SaveConfig(configLocation string, config *Config) (err error) {
	file, err := os.Create(configLocation)
	if err != nil {
		fmt.Println(err)
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (l *configLog) UnmarshalJSON(data []byte) error {
	defer fmt.Println(l.File)
	switch data[0] {
	case '{':
		cfg := struct {
			File       string `json:"file"`
			PretyPrint bool   `json:"pretyprint"`
		}{}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return err
		}
		*l = cfg //converts cfg to a propper configLog
	default:
		if err := json.Unmarshal(data, &l.File); err != nil {
			return err
		}
	}
	return nil
}
