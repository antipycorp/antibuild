// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/module/host"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

type (
	Config struct {
		Folders    ConfigFolder     `json:"folders"`
		Modules    ConfigModules    `json:"modules"`
		Pages      *site.ConfigSite `json:"pages"`
		moduleHost map[string]*host.ModuleHost
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
		Dependencies       map[string]string                 `json:"dependencies"`
		Config             map[string]map[string]interface{} `json:"config"`
		SitePostProcessors []string                          `json:"site_post_processors"`
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
			log.Fatal(parseErr.Error())
		}

		if isHost {
			hostLocally(config, port)
		}

		if isRefreshEnabled {
			buildOnRefresh(config, configLocation)
		}
	}
}

func startParse(configLocation string) (*Config, error) {
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
	ui.ShowBuiltSuccess()
	return config, nil
}

//parses the config file
func parseConfig(configLocation string) (*Config, error) {
	var config Config

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

	if config.Folders.Output == "" {
		return &config, errors.New("output folder not set")
	}

	return &config, nil
}

//start the template execution
func executeTemplate(config *Config) (err error) {
	//check if the output folder is there and delete its contents
	if config.Folders.Output == "" {
		err = os.RemoveAll(config.Folders.Output)
	}

	if err != nil {
		fmt.Println("output not specified")
	}

	sites := config.Pages

	config.Pages = &site.ConfigSite{}
	config.Pages.Sites = make([]*site.ConfigSite, 1)
	config.Pages.Sites[0] = sites

	site.OutputFolder = config.Folders.Output
	site.TemplateFolder = config.Folders.Templates
	site.StaticFolder = config.Folders.Static

	configPages, err := site.Unfold(config.Pages)
	if err != nil {
		fmt.Println("failed to unfold:", err)
	}

	pages, err := site.Convert(configPages)
	if err != nil {
		fmt.Println("failed to convert:", err)
	}

	for _, spp := range config.Modules.SitePostProcessors {
		pages = sitePostProcessors[spp].Process(pages, "")
	}

	err = site.Execute(pages)
	if err != nil {
		fmt.Println("failed to execute function:", err)
	}

	return
}
