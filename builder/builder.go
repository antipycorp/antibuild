// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"io"
	"os"
	"time"

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
	ErrFailedExport = errors.NewError("failed to export the output files", 2)
	//ErrNoOutputSpecified is for a failure in gathering files.
	ErrNoOutputSpecified = errors.NewError("no output folder specified", 3)
	//ErrFailedRemoveFile is for a failure in gathering files.
	ErrFailedRemoveFile = errors.NewError("failed removing output folder", 4)
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}
	cfg, err := config.CleanConfig(configLocation, ui)
	if err != nil {
		if isHost {
			failedToLoadConfig(ui, os.TempDir()+"/abm/public")
			net.HostLocally(os.TempDir()+"/abm/public", "8080")
		}
		ui.Fatalf("could not parse the config file " + err.Error())
		return
	}

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if isHost {
			//cfg.Folders.Output, _ = ioutil.TempDir("", "antibuild_hosting")
			go net.HostLocally(cfg.Folders.Output, port) //still continues running, hosting doesnt actually build
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(cfg, configLocation, ui)
			return
		}

		err = startParse(cfg)
		if err != nil {
			failedToRender(cfg)
		} else {
			cfg.UILogger.ShowResult()
		}
	}
}

//HeadlesStart starts a headless parser which just parses one thing
func HeadlesStart(configLocation string, output io.Writer) {
	ui := &UI.UI{}
	ui.SetLogfile(output)
	ui.SetPrettyPrint(false)

	cfg, err := config.CleanConfig(configLocation, ui)
	if err != nil {
		ui.Fatalf("could not parse the config file: %s", err.Error())
		return
	}

	cfg.UILogger.Info("Config is parsed and valid")
	cfg.UILogger.Debugf("Parsed Config: %s", cfg)

	err = startParse(cfg)
	if err != nil {
		cfg.UILogger.Fatalf(err.Error())
		return
	}
}

func startParse(cfg *config.Config) errors.Error {
	start := time.Now()

	cfg.UILogger.Debug("Initalizing module config")
	var moduleConfig = make(map[string]modules.ModuleConfig, len(cfg.Modules.Config))

	for module, mConfig := range cfg.Modules.Config {
		moduleConfig[module] = modules.ModuleConfig{
			Config: mConfig,
		}
	}

	cfg.UILogger.Debug("Loading modules")
	mhost, err := modules.LoadModules(cfg.Folders.Modules, cfg.Modules.Dependencies, moduleConfig, cfg.UILogger)
	if err != nil {
		cfg.UILogger.Fatal(err.Error())
		return nil
	}

	if mhost != nil {
		cfg.ModuleHost = mhost
	}

	cfg.UILogger.Debug("Finished loading modules")

	templateErr := executeTemplate(cfg)
	if templateErr != nil {
		cfg.UILogger.Fatal(templateErr.Error())
		return nil
	}

	cfg.UILogger.Infof("Completed in %s", time.Since(start).String())

	return nil
}

//start the template execution
func executeTemplate(cfg *config.Config) errors.Error {
	//check if the output folder is there and delete its contents
	if cfg.Folders.Output != "" {
		os.RemoveAll(cfg.Folders.Output)
	} else {
		return ErrNoOutputSpecified
	}
	sites := cfg.Pages

	cfg.Pages = &site.ConfigSite{}
	cfg.Pages.Sites = make([]*site.ConfigSite, 1)
	cfg.Pages.Sites[0] = sites

	site.OutputFolder = cfg.Folders.Output
	site.TemplateFolder = cfg.Folders.Templates
	site.StaticFolder = cfg.Folders.Static

	cfg.UILogger.Debug("Unfolding sites")

	startUnfold := time.Now()

	pages, err := site.Unfold(cfg.Pages, cfg.Modules.SPPs, cfg.UILogger.(*UI.UI))
	if err != nil {
		return err
	}

	cfg.UILogger.Debugf("Unfolding took %s", time.Since(startUnfold).String())

	cfg.UILogger.Debug("Finished unfolding sites")

	cfg.UILogger.Debug("Started building")
	startBuild := time.Now()

	err = site.Execute(pages, cfg.UILogger.(*UI.UI))
	if err != nil {
		return err
	}

	cfg.UILogger.Debugf("Building took %s", time.Since(startBuild).String())

	cfg.UILogger.Infof("Built %d pages", len(pages))

	return nil
}
