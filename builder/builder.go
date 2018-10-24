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
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}
	config, configErr := parseConfig(configLocation)
	if configErr != nil {
		ui.Fatal("Could not parse the config file")
		return
	}
	config.uilogger = ui

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if isHost {
			go hostLocally(config, port)
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(config, configLocation)
			return
		}

		parseErr := startParse(config)
		if parseErr != nil {
			log.Fatal(parseErr.Error())
			return
		}
	}
}

func startParse(config *Config) error {

	config.uilogger.ShowCompiling()

	loadModules(config) // loadModules checks if modules are already loaded

	//actually run the template
	templateErr := executeTemplate(config)
	if templateErr != nil {
		fmt.Println("failed building templates: ", templateErr.Error())
		return templateErr
	}

	//print finish time
	config.uilogger.ShowResult()
	return nil
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

	pages, err := site.Unfold(config.Pages)
	if err != nil {
		fmt.Println("failed to unfold:", err)
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
