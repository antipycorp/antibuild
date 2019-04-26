// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"io"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/pkg/profile"
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
		ui.Fatalf("Could not parse the config file " + err.Error())
		ui.ShowResult()
		return
	}

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if os.Getenv("DEBUG") == "1" { //cant get out of this, itl just loop
			p := profile.Start(profile.MemProfile)

			cache, _ := actualStartParse(cfg)
			net.HostDebug()
			timeout := time.After(10 * time.Second)

			for i := 0; ; i++ {
				select {
				case <-timeout:
					println("did", i, "iterations int one minute")
					p.Stop()
					return
				default:
					cache.configUpdate = true
					cache.rootPage = *cfg.Pages
					startParse2(cfg, cache)
				}
			}
		}

		if isHost {
			go net.HostLocally(cfg.Folders.Output, port) //still continues running, hosting doesn't actually build
		}

		if isRefreshEnabled { // if refresh is enabled run the refresh, if it returns return
			buildOnRefresh(cfg, configLocation, ui)
			return
		}

		_, err = actualStartParse(cfg)
		if err != nil {
			cfg.UILogger.Fatal(err.Error())
			println(err.Error())
			failedToRender(cfg)
		}

		cfg.UILogger.ShowResult()
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
	cfg.UILogger.Debugf("Parsed Config: %v", cfg)

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
		c.config.UILogger.Debug("Initializing module config")
		var moduleConfig = make(map[string]modules.ModuleConfig, len(c.config.Modules.Config))

		for module, mConfig := range c.config.Modules.Config {
			moduleConfig[module] = modules.ModuleConfig{
				Config: mConfig,
			}
		}

		c.config.UILogger.Debug("Loading modules")
		moduleHost, err := modules.LoadModules(c.config.Folders.Modules, c.config.Modules.Dependencies, moduleConfig, c.config.UILogger)
		if err != nil {
			c.config.UILogger.Fatal(err.Error())
			return nil
		}
		if moduleHost != nil {
			c.config.ModuleHost = moduleHost
		}
		c.config.UILogger.Debug("Finished loading modules")

		pages := c.config.Pages

		c.config.Pages = &site.ConfigSite{}
		c.config.Pages.Sites = make([]site.ConfigSite, 1)
		c.config.Pages.Sites[0] = *pages

		site.OutputFolder = c.config.Folders.Output
		site.TemplateFolder = c.config.Folders.Templates
		site.StaticFolder = c.config.Folders.Static
	}

	if c.shouldUnfold || c.configChanged {
		c.config.UILogger.Debug("Unfolding sites")
		var err errors.Error
		//c.cSites, err = site.Unfold(c.config.Pages, c.config.UILogger.(*UI.UI))
		if err != nil {
			return err
		}
		c.config.UILogger.Debug("Finished unfolding sites")
	}

	var sites []*site.Site
	var changed [][16]byte

	if len(c.templatesToRebuild) > 0 || c.shouldUnfold || c.configChanged {
		c.config.UILogger.Debug("Started gathering")
		for hash, cSite := range c.cSites {
			if c.configChanged || c.shouldUnfold {
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
		cSites:        map[[16]byte]*site.ConfigSite{},
		sites:         map[[16]byte]*site.Site{},
	}

	return c, startCachedParse(c)
}

type cach struct {
	rootPage site.ConfigSite
	data     map[string]cachData

	configUpdate bool
	checkData    bool
}

type cachData struct {
	dependencies []string
	site         site.Site
	shouldRemove bool
}

func startParse2(cfg *config.Config, cache *cach) errors.Error {
	start := time.Now()

	if cfg.Folders.Output == "" {
		return ErrNoOutputSpecified
	}

	//if there is a config update reload all modules
	if cache.configUpdate {
		var moduleConfig = make(map[string]modules.ModuleConfig, len(cfg.Modules.Config))

		for module, mConfig := range cfg.Modules.Config {
			moduleConfig[module] = modules.ModuleConfig{
				Config: mConfig,
			}
		}

		moduleHost, err := modules.LoadModules(cfg.Folders.Modules, cfg.Modules.Dependencies, moduleConfig, cfg.UILogger)
		if err != nil {
			cfg.UILogger.Fatal(err.Error())
			return nil
		}
		if moduleHost != nil { // loadModules checks if modules are already loaded
			cfg.ModuleHost = moduleHost
		}

		site.TemplateFolder = cfg.Folders.Templates
		site.OutputFolder = cfg.Folders.Output
	}

	if cache.configUpdate {
		err := os.RemoveAll(cfg.Folders.Output)
		if err != nil {
			return ErrFailedRemoveFile.SetRoot(err.Error())
		}
	}

	pagesC := site.DeepCopy(*cfg.Pages)
	sites, _ := site.Unfold(&pagesC, cfg.UILogger.(*UI.UI))

	updatedSites := make([]*site.Site, 0, len(sites))
	for _, cSite := range sites {
		depChange := false

		var ok bool
		var cd cachData
		if cd, ok = cache.data[cSite.Slug]; ok && !cache.configUpdate {
			if len(cSite.Dependencies) != len(cd.dependencies) {
				depChange = true
			} else {
				for i, d := range cSite.Dependencies {
					if d != cd.dependencies[i] {
						depChange = true
						break
					}
				}
			}
		} else {
			depChange = true
		}

		cd.dependencies = cSite.Dependencies

		os.Remove(path.Join(cfg.Folders.Output, cd.site.Slug))
		var s *site.Site

		datEqual := true
		if cache.checkData {
			s, err := site.Gather(cSite, cfg.UILogger.(*UI.UI))
			if err != nil {
				return err
			}
			for k, v := range s.Data {
				if !reflect.DeepEqual(v, cd.site.Data[k]) {
					datEqual = false
				}
			}
		}

		if depChange || !datEqual || (s != nil && site.GetTemplateTree(s.Template) != site.GetTemplateTree(cd.site.Template)) || cache.configUpdate {
			if s == nil {
				var err errors.Error
				s, err = site.Gather(cSite, cfg.UILogger.(*UI.UI))
				if err != nil {
					return err
				}
			}
			cd.site = *s
			updatedSites = append(updatedSites, s)
		}

		cd.shouldRemove = false
		cache.data[cSite.Slug] = cd
	}

	if len(cfg.Modules.SPPs) > 0 {
		err := site.PostProcess(&updatedSites, cfg.Modules.SPPs, cfg.UILogger.(*UI.UI))
		if err != nil {
			return err
		}
	}

	err := site.Execute(updatedSites, cfg.UILogger.(*UI.UI))
	if err != nil {
		return err
	}

	cfg.UILogger.Infof("Built %d pages", len(updatedSites))
	cfg.UILogger.Infof("Completed in %s", time.Since(start).String())

	for k, v := range cache.data {
		if !v.shouldRemove {
			v.shouldRemove = true
			cache.data[k] = v
		} else {
			os.Remove(path.Join(cfg.Folders.Output, k))
			findTopEmptyDir(cfg.Folders.Output, k)
			delete(cache.data, k)
		}
	}

	//the true state will need to be checked during the process so we leave them true until the end
	cache.configUpdate = false
	cache.checkData = false

	return nil
}

func actualStartParse(cfg *config.Config) (*cach, errors.Error) {
	cache := &cach{
		rootPage:     *cfg.Pages,
		data:         make(map[string]cachData),
		configUpdate: true,
		checkData:    false,
	}
	return cache, startParse2(cfg, cache)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func findTopEmptyDir(base, rel string) error {
	var p = path.Join(base, rel)
	for {
		p, _ = path.Split(p)
		p = p[:len(p)-1]
		f, err := os.Open(p)
		if err != nil {
			return nil //no issue if you cant open the file, it might just be already removed
		}
		defer f.Close()

		// read in ONLY one file
		_, err = f.Readdir(1)

		if err != io.EOF {
			return os.RemoveAll(p)
		}
	}
}
