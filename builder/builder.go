// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
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

	apiSite "gitlab.com/antipy/antibuild/api/site"
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

		_, err = startParse(cfg)
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

	_, err = startParse(cfg)
	if err != nil {
		cfg.UILogger.Fatalf(err.Error())
		return
	}
}

func startCachedParse(c *cache) errors.Error {
	start := time.Now()

	//check if the output folder is there and delete its contents
	if c.config.Folders.Output != "" {
		if c.configChanged {
			os.RemoveAll(c.config.Folders.Output)
		}
	} else {
		return ErrNoOutputSpecified
	}

	if c.configChanged {
		c.config.UILogger.Debug("Initalizing module config")
		var moduleConfig = make(map[string]modules.ModuleConfig, len(c.config.Modules.Config))

		for module, mConfig := range c.config.Modules.Config {
			moduleConfig[module] = modules.ModuleConfig{
				Config: mConfig,
			}
		}

		c.config.UILogger.Debug("Loading modules")
		mhost, err := modules.LoadModules(c.config.Folders.Modules, c.config.Modules.Dependencies, moduleConfig, c.config.UILogger)
		if err != nil {
			c.config.UILogger.Fatal(err.Error())
			return nil
		}
		if mhost != nil {
			c.config.ModuleHost = mhost
		}
		c.config.UILogger.Debug("Finished loading modules")

		pages := c.config.Pages

		c.config.Pages = &site.ConfigSite{}
		c.config.Pages.Sites = make([]*site.ConfigSite, 1)
		c.config.Pages.Sites[0] = pages

		site.OutputFolder = c.config.Folders.Output
		site.TemplateFolder = c.config.Folders.Templates
		site.StaticFolder = c.config.Folders.Static
	}

	if c.shouldUnfold || c.configChanged {
		c.config.UILogger.Debug("Unfolding sites")
		var err errors.Error
		c.cSites, err = site.Unfold(c.config.Pages, c.config.UILogger.(*UI.UI))
		if err != nil {
			return err
		}
		c.config.UILogger.Debug("Finished unfolding sites")
	}

	var sites []*apiSite.Site
	var changed []string

	if len(c.templatesToRebuild) > 0 || c.configChanged {
		c.config.UILogger.Debug("Started gathering")
		for hash, cSite := range c.cSites {
			if c.configChanged {
				s, err := site.Gather(cSite, c.config.UILogger.(*UI.UI))
				if err != nil {
					return err
				}
				c.sites[hash] = s
				changed = append(changed, hash)
			} else {
				for _, t := range c.templatesToRebuild {
					if contains(cSite.Templates, t) {
						s, err := site.Gather(cSite, c.config.UILogger.(*UI.UI))
						if err != nil {
							return err
						}
						c.sites[hash] = s
						changed = append(changed, hash)
					}
				}
			}

		}
		c.config.UILogger.Debug("Finished gathering")

		if len(c.config.Modules.SPPs) > 0 {
			for _, cSite := range c.sites {
				sites = append(sites, cSite)
			}

			c.config.UILogger.Debug("Started post-processing")
			err := site.PostProcess(&sites, c.config.Modules.SPPs, c.config.UILogger.(*UI.UI))
			if err != nil {
				return err
			}
			c.config.UILogger.Debug("Finished post-processing")

		}
	}

	c.config.UILogger.Debug("Started building")
	if len(sites) == 0 && len(c.config.Modules.SPPs) == 0 {
		for _, hash := range changed {
			sites = append(sites, c.sites[hash])
		}
	}
	err := site.Execute(sites, c.config.UILogger.(*UI.UI))
	if err != nil {
		return err
	}
	c.config.UILogger.Infof("Built %d pages", len(sites))

	c.config.UILogger.Infof("Completed in %s", time.Since(start).String())

	c.configChanged = false
	c.shouldUnfold = false
	c.templatesToRebuild = []string{}

	return nil
}

func startParse(cfg *config.Config) (*cache, errors.Error) {
	c := &cache{
		config:        cfg,
		configChanged: true,
		moduleConfig:  map[string]modules.ModuleConfig{},
		cSites:        map[string]*site.ConfigSite{},
		sites:         map[string]*apiSite.Site{},
	}

	return c, startCachedParse(c)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
