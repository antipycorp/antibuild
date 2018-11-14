// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"errors"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/builder/websocket"
	"gitlab.com/antipy/antibuild/cli/modules"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}

	cfg, configErr := parseConfig(configLocation)
	if configErr != nil {
		ui.Fatalf("could not parse the config file: %s", configErr)

		failledToLoadConfig(ui, os.TempDir()+"/abm/public")

		ui.ShowResult()
		if isHost {
			hostLocally(os.TempDir()+"/abm/public", "8080")
		}
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
			go hostLocally(cfg.Folders.Output, port) //still continues running, hosting doesnt actually build
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(cfg, configLocation)
			return
		}

		parseErr := startParse(cfg)
		if parseErr != nil {
			failledToRender(cfg)

			ui.Fatal(parseErr.Error())
			ui.ShowResult()

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
		return errors.New("failed building templates:" + templateErr.Error())

	}

	//print finish time
	cfg.UILogger.ShowResult()
	websocket.SendUpdate()
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
		cfg.UILogger.Fatalf("output not specified: %s", err)
		failledToRender(cfg)
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
		failledToRender(cfg)
		return errors.New("failed to unfold: 1" + err.Error())
	}

	err = site.Execute(pages)
	if err != nil {
		failledToRender(cfg)
		return errors.New("failed to execute function:" + err.Error())

	}
	return
}
