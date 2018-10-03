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
)

type (
	dataFile struct {
		Data map[string]interface{} `json:"Data"`
	}

	config struct {
		Folders    configFolder  `json:"folders"`
		Modules    configModules `json:"modules"`
		Pages      site          `json:"pages"`
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

	site struct {
		Slug      string   `json:"slug"`
		Templates []string `json:"templates"`
		Data      []string `json:"data"`

		Sites []site `json:"sites"`
	}
)

var (
	errNoTemplate = errors.New("the template folder is not set")
	errNoData     = errors.New("the data folder is not set")
	errNoOutput   = errors.New("the output folder is not set")
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	if isConfigSet {
		config, parseErr := startParse(configLocation)

		if parseErr != nil {
			panic(fmt.Sprintf("could not get the output folder from config.json: %s", parseErr))
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

	//reparse the config
	config, configErr := parseConfig(configLocation)
	if configErr != nil {
		fmt.Println(configErr.Error())
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
	fmt.Printf("Completed parse in %s\n", time.Since(start).String())
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

	//decode the file
	dec := json.NewDecoder(configFile)
	err = dec.Decode(&config)
	if err != nil {
		return &config, err
	}

	//check if all necessary fields are present
	if config.Folders.Templates == "" {
		return &config, errNoTemplate
	}
	if config.Folders.Data == "" {
		return &config, errNoData
	}
	if config.Folders.Output == "" {
		return &config, errNoOutput
	}

	//return config
	return &config, nil
}
