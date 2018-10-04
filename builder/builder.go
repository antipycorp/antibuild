// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"gitlab.com/antipy/antibuild/cli/module/host"
	"gitlab.com/antipy/antibuild/cli/site"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

type (
	dataFile struct {
		Data map[string]interface{} `json:"Data"`
	}

	config struct {
		Folders    configFolder  `json:"folders"`
		Modules    configModules `json:"modules"`
		Pages      site.Site     `json:"pages"`
		moduleHost map[string]*host.ModuleHost
	}

	configFolder struct {
		Templates string `json:"templates"`
		Data      string `json:"data"`
		Static    string `json:"static"`
		Output    string `json:"output"`
		Modules   string `json:"modules"`
	}

	configModules struct {
		Dependencies map[string]string                 `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config"`
	}
)

var ui = &UI.UI{}

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		config, parseErr := startParse(configLocation)
		if parseErr != nil {
			ui.LogImportant(UI.DataFolder, "/config.json", "?", nil)
		}

		if isHost {
			hostLocally(config, port)
		}

		if isRefreshEnabled {
			buildOnRefresh(config, configLocation)
		}
	}
}

func startParse(configLocation string) (*config, error) {
	//record start time
	start := time.Now()

	//show compiling on ui
	ui.ShowCompiling()

	//reparse the config
	config, configErr := parseConfig(configLocation)
	if configErr != nil {
		return config, configErr
	}

	//check if modules have already been loaded
	if loadedModules == false {
		//load modules and make sure they dont get loaded again
		loadModules(config)
		loadedModules = true
	}

	//actually run the template
	templateErr := executeTemplate(config)
	if templateErr != nil {
		fmt.Println("failed building templates: ", templateErr.Error())
		return config, templateErr
	}

	//print finish time
	ui.ShowBuiltSuccess(time.Since(start).String())
	return config, nil
}

//parses the config file
func parseConfig(configLocation string) (*config, error) {
	var config config

	//open the file
	configFile, err := os.Open(configLocation)
	defer configFile.Close()
	if err != nil {
		return nil, errors.New("could not open the config file: " + err.Error())
	}

	dec := json.NewDecoder(configFile)
	err = dec.Decode(&config)
	if err != nil {
		return &config, err
	}

	if config.Folders.Templates == "" {
		return &config, errors.New("template folder not set")
	}
	if config.Folders.Data == "" {
		return &config, errors.New("data folder not set")
	}
	if config.Folders.Output == "" {
		return &config, errors.New("output folder not set")
	}

	//return config
	return &config, nil
}
