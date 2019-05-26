// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package engine

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/engine/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"
)

type (
	cache struct {
		rootPage site.ConfigSite
		data     map[string]cacheData

		configUpdate   bool
		checkData      bool
		templateUpdate string //tha absolute path to the updated template
	}

	cacheData struct {
		dependencies []string
		site         site.Site
		shouldRemove bool
	}
)

var (
	//ErrFailedUnfold is when the template failed building
	ErrFailedUnfold = errors.NewError("failed to unfold", 1)
	//ErrFailedExport is for a failure moving the static folder
	ErrFailedExport = errors.NewError("failed to export the output files", 2)
	//ErrNoOutputSpecified is for a failure in gathering files.
	ErrNoOutputSpecified = errors.NewError("no output folder specified", 3)
	//ErrFailedRemoveFile is for failling to remove the output folder.
	ErrFailedRemoveFile = errors.NewError("failed removing output folder", 4)
	//ErrFailedCache is for failing to do something in cache operations.
	ErrFailedCache = errors.NewError("failed in determaning if the cache should be invalidated:", 4)
)

func startCachedParse(cfg *localConfig.Config, cache *cache) errors.Error {
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

		er := os.RemoveAll(cfg.Folders.Output)
		if er != nil {
			return ErrFailedRemoveFile.SetRoot(err.Error())
		}

	}

	pagesC := site.DeepCopy(*cfg.Pages)
	sites, err := site.Unfold(&pagesC, cfg.UILogger.(*ui.UI))
	if err != nil {
		return ErrFailedUnfold.SetRoot(err.Error())
	}
	updatedSites := make([]*site.Site, 0, len(sites))
	for i := range sites {

		cd, foundCache := cache.data[sites[i].Slug]

		shouldChange := !foundCache || depChange(cd, &sites[i]) || cache.configUpdate

		cd.dependencies = sites[i].Dependencies

		var s *site.Site

		if cache.checkData {
			var err errors.Error
			s, err = site.Gather(sites[i], cfg.UILogger.(*ui.UI))
			if err != nil {
				return err
			}
			for k, v := range s.Data {
				if !reflect.DeepEqual(v, cd.site.Data[k]) {
					// shouldChange = shouldChange || true
					shouldChange = true
				}
			}
		}
		for _, v := range sites[i].Templates {
			abs, err := filepath.Abs(path.Join(cfg.Folders.Templates, v))
			if err != nil {
				return ErrFailedCache.SetRoot("the template path is not inside the templates folder: " + v)
			}
			if abs == cache.templateUpdate {
				path := filepath.Join(site.TemplateFolder, v)
				site.RemoveTemplate(path)
				//shouldChange = shouldChange || true
				shouldChange = true
			}
		}

		if shouldChange {
			if s == nil {
				var err errors.Error
				s, err = site.Gather(sites[i], cfg.UILogger.(*ui.UI))
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
		err := site.PostProcess(&updatedSites, cfg.Modules.SPPs, cfg.UILogger.(*ui.UI))
		if err != nil {
			return err
		}
	}

	err = site.Execute(updatedSites, cfg.UILogger.(*ui.UI))
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
			removeTopEmptyDir(cfg.Folders.Output, k)
			delete(cache.data, k)
		}
	}

	//the true state will need to be checked during the process so we leave them true until the end
	cache.configUpdate = false
	cache.checkData = false
	cache.templateUpdate = ""

	return nil
}

func depChange(cd cacheData, s *site.ConfigSite) bool {
	if len(s.Dependencies) != len(cd.dependencies) {
		return true
	}
	for i, d := range s.Dependencies {
		if d != cd.dependencies[i] {
			return true
		}
	}
	return false
}

func removeTopEmptyDir(base, rel string) error {
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
