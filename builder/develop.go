// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/net"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

//watches files and folders and rebuilds when things change
func buildOnRefresh(cfg *config.Config, configLocation string, ui *UI.UI) {
	ui.Debug("making a refresh")

	startParse(cfg)

	shutdown := make(chan int, 10) // 10 is enough for some watcher to not get stuck on the chan
	if cfg.Folders.Static != "" {
		go staticWatch(cfg.Folders.Static, cfg.Folders.Output, shutdown, ui)
	}
	watchBuild(cfg, configLocation, shutdown, ui)
}

func staticWatch(src, dst string, shutdown chan int, log config.UIlogger) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("could not open a file watcher: %s", err.Error())
		shutdown <- 0
		return
	}
	//add static folder to watcher
	err = filepath.Walk(src, func(path string, file os.FileInfo, err error) error {
		return watcher.Add(path)
	})
	if err != nil {
		log.Errorf("could not watch all files %s", err.Error())
		shutdown <- 0
		return
	}
	for {
		//listen for watcher events
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				shutdown <- 0
				return
			}

			info, err := os.Lstat(src)
			if err != nil {
				log.Errorf("couldn't move files form static to out: %s", err.Error())
				continue
			}

			internal.GenCopy(src, dst, info)
		case err, ok := <-watcher.Errors:
			if !ok {
				shutdown <- 0
				return
			}
			log.Errorf("error: %s", err.Error())
		case _ = <-shutdown:
			return
		}
	}
}

//! modules will not be able to call a refresh and thus we can only use the (local) templates as a source
func watchBuild(cfg *config.Config, configloc string, shutdown chan int, ui *UI.UI) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ui.Errorf("could not open a file watcher: %s", err.Error())
		shutdown <- 0
		return
	}

	//add static folder to watcher
	err = filepath.Walk(cfg.Folders.Templates, func(path string, file os.FileInfo, err error) error {
		return watcher.Add(path)
	})

	if err != nil {
		ui.Errorf("could not watch all files %s", err.Error())
		shutdown <- 0
		return
	}

	err = watcher.Add(configloc)
	if err != nil {
		ui.Errorf("could not watch config file %s", err.Error())
	}
	for {
		//listen for watcher events
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				shutdown <- 0
				return
			}
			ui.Debug("making a refresh")
			if e.Name == configloc {
				cfg, err = config.CleanConfig(configloc, ui)
				if err != nil {
					failledToLoadConfig(ui, os.TempDir()+"/abm/public")
					net.HostLocally(os.TempDir()+"/abm/public", "8080")
				}
			}

			startParse(cfg)

		case err, ok := <-watcher.Errors:
			if !ok {
				shutdown <- 0
				return
			}
			ui.Errorf("error: %s", err.Error())
		case _ = <-shutdown:
			return
		}
	}
}
