// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"fmt"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules"
	"gitlab.com/antipy/antibuild/cli/net"
	"gitlab.com/antipy/antibuild/cli/net/websocket"

	UI "gitlab.com/antipy/antibuild/cli/ui"
)

var (
	//ErrFailledUnfold is when the template failled building
	ErrFailledUnfold = errors.NewError("failed to unfold", 1)
	//ErrFailledExport is for a faillure moving the static folder
	ErrFailledExport = errors.NewError("failled to export the template files", 2)
	//ErrNoOutpuSpecified is for a faillure in gathering files.
	ErrNoOutpuSpecified = errors.NewError("no output folder specified", 3)
	//ErrFailledRemoveFile is for a faillure in gathering files.
	ErrFailledRemoveFile = errors.NewError("failled removing filles", 4)
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}
	cfg, err := config.CleanConfig(configLocation, ui)
	if err != nil {
		if isHost {
			failledToLoadConfig(ui, os.TempDir()+"/abm/public")
			net.HostLocally(os.TempDir()+"/abm/public", "8080")
		}
	}

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if isHost {
			go net.HostLocally(cfg.Folders.Output, port) //still continues running, hosting doesnt actually build
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(cfg, configLocation, ui)
			return
		}
		fmt.Println("started parsing")
		startParse(cfg)
	}
}

func startParse(cfg *config.Config) {
	defer func() {
		websocket.SendUpdate()
		cfg.UILogger.ShowResult()
	}()

	cfg.UILogger.ShowCompiling()

	mhost := modules.LoadModules(cfg.Folders.Modules, cfg.Modules.Dependencies, cfg.Modules.Config, cfg.UILogger)
	if mhost != nil { // loadModules checks if modules are already loaded
		cfg.ModuleHost = mhost
	}
	fmt.Println("loaded modules")
	//actually run the template
	templateErr := executeTemplate(cfg)
	fmt.Println("ran the templates")
	if templateErr != nil {
		cfg.UILogger.Fatal("failed building templates:" + templateErr.Error())

		failledToRender(cfg)
	}
}

//start the template execution
func executeTemplate(cfg *config.Config) errors.Error {
	//check if the output folder is there and delete its contents
	if cfg.Folders.Output != "" {
		err := os.RemoveAll(cfg.Folders.Output)
		if err != nil {
			return ErrFailledRemoveFile.SetRoot(err.Error())
		}
	} else {
		return ErrNoOutpuSpecified
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
		return ErrFailledUnfold.SetRoot(err.GetRoot())
	}

	err = site.Execute(pages)
	if err != nil {
		return ErrFailledExport.SetRoot(err.GetRoot())
	}
	return nil
}
