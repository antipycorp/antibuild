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

	UI "gitlab.com/antipy/antibuild/cli/ui"
)

var (
	//ErrFailedUnfold is when the template failed building
	ErrFailedUnfold = errors.NewError("failed to unfold", 1)
	//ErrFailedExport is for a failure moving the static folder
	ErrFailedExport = errors.NewError("failed to export the template files", 2)
	//ErrNoOutpuSpecified is for a failure in gathering files.
	ErrNoOutpuSpecified = errors.NewError("no output folder specified", 3)
	//ErrFailedRemoveFile is for a failure in gathering files.
	ErrFailedRemoveFile = errors.NewError("failed removing filles", 4)
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

		err := startParse(cfg)
		if err != nil {
			failledToRender(cfg)
		} else {
			cfg.UILogger.ShowResult()
		}
	}
}

func startParse(cfg *config.Config) errors.Error {

	cfg.UILogger.Info("started compiling...")

	mhost := modules.LoadModules(cfg.Folders.Modules, cfg.Modules.Dependencies, cfg.Modules.Config, cfg.UILogger)
	if mhost != nil { // loadModules checks if modules are already loaded
		cfg.ModuleHost = mhost
	}
	cfg.UILogger.Debug("loaded modules")
	//actually run the template
	templateErr := executeTemplate(cfg)
	cfg.UILogger.Debug("ran the templates")
	if templateErr != nil {
		cfg.UILogger.Fatal("failed building templates:" + templateErr.Error())
	}
	cfg.UILogger.Info("succesfully exported the templates")
	return nil
}

//start the template execution
func executeTemplate(cfg *config.Config) errors.Error {
	//check if the output folder is there and delete its contents
	if cfg.Folders.Output != "" {
		err := os.RemoveAll(cfg.Folders.Output)
		if err != nil {
			return ErrFailedRemoveFile.SetRoot(err.Error())
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
		return ErrFailedUnfold.SetRoot(err.GetRoot())
	}

	err = site.Execute(pages)
	if err != nil {
		return ErrFailedExport.SetRoot(err.GetRoot())
	}
	return nil
}
