// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/modules"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}
	cfg, configErr := parseConfig(configLocation)
	if configErr != nil {
		ui.Fatal("Could not parse the config file")
		return
	}

	file, err := os.OpenFile(cfg.LogConfig.File, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		ui.Fatal(err.Error())
	}
	file.Seek(0, 0)
	ui.LogFile = file
	ui.PrettyLog = cfg.LogConfig.PretyPrint

	cfg.UILogger = ui

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if isHost {
			go hostLocally(cfg, port)
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(cfg, configLocation)
			return
		}

		parseErr := startParse(cfg)
		if parseErr != nil {
			log.Fatal(parseErr.Error())
			return
		}
	}
}

func startParse(cfg *config.Config) error {

	cfg.UILogger.ShowCompiling()
	mhost := modules.LoadModules(cfg.Folders.Modules, cfg.Modules.Dependencies, cfg.Modules.Config)
	if mhost != nil { // loadModules checks if modules are already loaded
		cfg.ModuleHost = mhost
	}

	//actually run the template
	templateErr := executeTemplate(cfg)
	if templateErr != nil {
		fmt.Println("failed building templates: ", templateErr.Error())
		return templateErr
	}

	//print finish time
	cfg.UILogger.ShowResult()
	return nil
}

//parses the config file and check for any missing information
func parseConfig(configLocation string) (*config.Config, error) {
	cfg, err := config.GetConfig(configLocation)
	if err != nil {
		return cfg, err
	}

	if cfg.Folders.Templates == "" {
		return cfg, errors.New("template folder not set")
	}

	if cfg.Folders.Output == "" {
		return cfg, errors.New("output folder not set")
	}

	return cfg, nil
}

//start the template execution
func executeTemplate(cfg *config.Config) (err error) {
	//check if the output folder is there and delete its contents
	if cfg.Folders.Output == "" {
		err = os.RemoveAll(cfg.Folders.Output)
	}

	if err != nil {
		fmt.Println("output not specified")
	}

	sites := cfg.Pages

	cfg.Pages = &site.ConfigSite{}
	cfg.Pages.Sites = make([]*site.ConfigSite, 1)
	cfg.Pages.Sites[0] = sites

	site.OutputFolder = cfg.Folders.Output
	site.TemplateFolder = cfg.Folders.Templates
	site.StaticFolder = cfg.Folders.Static

	pages, err := site.Unfold(cfg.Pages, cfg.Modules.SPPs)
	if err != nil {
		fmt.Println("failed to unfold:", err)
	}

	err = site.Execute(pages)
	if err != nil {
		fmt.Println("failed to execute function:", err)
	}

	return
}
