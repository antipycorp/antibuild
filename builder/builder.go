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

	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules"
	"gitlab.com/antipy/antibuild/cli/net"

	UI "gitlab.com/antipy/antibuild/cli/ui"
)

type cache struct {
	rootPage site.ConfigSite
	data     map[string]cacheData

	configUpdate bool
	checkData    bool
}

type cacheData struct {
	dependencies []string
	site         site.Site
	shouldRemove bool
}

var (
	//ErrFailedUnfold is when the template failed building
	ErrFailedUnfold = errors.NewError("failed to unfold", 1)
	//ErrFailedExport is for a failure moving the static folder
	ErrFailedExport = errors.NewError("failed to export the output files", 2)
	//ErrNoOutputSpecified is for a failure in gathering files.
	ErrNoOutputSpecified = errors.NewError("no output folder specified", 3)
	//ErrFailedRemoveFile is for failling to remove the output folder.
	ErrFailedRemoveFile = errors.NewError("failed removing output folder", 4)
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	ui := &UI.UI{}
	cfg, err := config.CleanConfig(configLocation, ui, true)
	if err != nil {
		ui.Fatalf("Could not parse the config file " + err.Error())
		ui.ShowResult()
		return
	}

	if isConfigSet {
		ui.HostingEnabled = isHost
		ui.Port = port

		if os.Getenv("DEBUG") == "1" { //cant get out of this, itl just loop
			cache, _ := startParse(cfg)
			net.HostDebug()
			timeout := time.After(1 * time.Minute)

			for i := 0; ; i++ {
				select {
				case <-timeout:
					println("did", i, "iterations int one minute")
					return
				default:
					cache.configUpdate = true
					cache.rootPage = *cfg.Pages
					startCachedParse(cfg, cache)
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

		_, err = startParse(cfg)
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

	cfg, err := config.CleanConfig(configLocation, ui, false)
	ui.SetLogfile(output)
	ui.SetPrettyPrint(false)

	if err != nil {
		ui.Fatalf("could not parse the config file: %s", err.Error())
		return
	}

	cfg.UILogger.Info("Config is parsed and valid")
	cfg.UILogger.Debugf("Parsed Config: %v", cfg)

	cache := &cache{
		rootPage:     *cfg.Pages,
		data:         make(map[string]cacheData),
		configUpdate: true,
		checkData:    false,
	}

	err = startCachedParse(cfg, cache)
	if err != nil {
		cfg.UILogger.Fatalf(err.Error())
		return
	}
}

func startCachedParse(cfg *config.Config, cache *cache) errors.Error {
	start := time.Now()

	if cfg.Folders.Output == "" {
		return ErrNoOutputSpecified
	}

	//if there is a config update reload all modules
	if cache.configUpdate {

		moduleHost, err := modules.LoadModules(cfg.Folders.Modules, cfg.Modules, cfg.UILogger)
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
	sites, err := site.Unfold(&pagesC, cfg.UILogger.(*UI.UI))
	if err != nil {
		return ErrFailedUnfold.SetRoot(err.Error())
	}
	updatedSites := make([]*site.Site, 0, len(sites))
	for i := range sites {
		depChange := false

		var ok bool
		var cd cacheData
		if cd, ok = cache.data[sites[i].Slug]; ok && !cache.configUpdate {
			if len(sites[i].Dependencies) != len(cd.dependencies) {
				depChange = true
			} else {
				for i, d := range sites[i].Dependencies {
					if d != cd.dependencies[i] {
						depChange = true
						break
					}
				}
			}
		} else {
			depChange = true
		}

		cd.dependencies = sites[i].Dependencies

		os.Remove(path.Join(cfg.Folders.Output, cd.site.Slug))
		var s *site.Site

		datEqual := true
		if cache.checkData {
			s, err := site.Gather(sites[i], cfg.UILogger.(*UI.UI))
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
				s, err = site.Gather(sites[i], cfg.UILogger.(*UI.UI))
				if err != nil {
					return err
				}
			}
			cd.site = *s
			updatedSites = append(updatedSites, s)
		}

		cd.shouldRemove = false
		cache.data[sites[i].Slug] = cd
	}

	if len(cfg.Modules.SPPs) > 0 {
		err := site.PostProcess(&updatedSites, cfg.Modules.SPPs, cfg.UILogger.(*UI.UI))
		if err != nil {
			return err
		}
	}

	err = site.Execute(updatedSites, cfg.UILogger.(*UI.UI))
	if err != nil {
		return ErrFailedExport.SetRoot(err.Error())
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

func startParse(cfg *config.Config) (*cache, errors.Error) {
	cache := &cache{
		rootPage:     *cfg.Pages,
		data:         make(map[string]cacheData),
		configUpdate: true,
		checkData:    false,
	}
	return cache, startCachedParse(cfg, cache)
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
