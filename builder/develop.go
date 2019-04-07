// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package builder

import (
	"os"
	"path/filepath"

	"github.com/eiannone/keyboard"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/modules"

	"github.com/fsnotify/fsnotify"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/net"
	"gitlab.com/antipy/antibuild/cli/net/websocket"
	UI "gitlab.com/antipy/antibuild/cli/ui"
)

type cache struct {
	config        *config.Config
	moduleConfig  map[string]modules.ModuleConfig
	configChanged bool

	cSites       map[[16]byte]*site.ConfigSite
	shouldUnfold bool

	sites              map[[16]byte]*site.Site
	templatesToRebuild []string
}

//watches files and folders and rebuilds when things change
func buildOnRefresh(cfg *config.Config, configLocation string, ui *UI.UI) {
	cache, err := actualStartParse(cfg)
	if err != nil {
		cfg.UILogger.Fatal(err.Error())
		println(err.Error())
		failedToRender(cfg)
	}

	cfg.UILogger.ShowResult()

	shutdown := make(chan int, 10) // 10 is enough for some watcher to not get stuck on the chan
	if cfg.Folders.Static != "" {
		go staticWatch(cfg.Folders.Static, cfg.Folders.Output, shutdown, ui)
	}

	watchBuild(cfg, cache, configLocation, shutdown, ui)
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
		case <-shutdown:
			shutdown <- 0
			return
		}
	}
}

//! modules will not be able to call a refresh and thus we can only use the (local) templates as a source
func watchBuild(cfg *config.Config, c *cach, configloc string, shutdown chan int, ui *UI.UI) {
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

	err = keyboard.Open()
	if err != nil {
		ui.Errorf("could not start keyboard listener: %s", err.Error())
	}
	defer keyboard.Close()

	keyChannel := make(chan rune)
	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				ui.Errorf("getting key failed: %s", err.Error())
			} else if key == keyboard.KeyCtrlC || key == keyboard.KeyEsc {
				shutdown <- 1
			} else {
				keyChannel <- char
			}

		}
	}()

	keys := []rune("Rr")

	for {
		//listen for watcher events
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				shutdown <- 0
				return
			}

			if e.Op != fsnotify.Create && e.Op != fsnotify.Remove && e.Op != fsnotify.Rename && e.Op != fsnotify.Write {
				break
			}

			ui.Infof("Refreshing because of page %s", e.Name)

			if e.Name == configloc {
				ui.Info("Changed file is config. Reloading...")
				ncfg, err := config.CleanConfig(configloc, ui)
				if err != nil {
					ui.Fatalf("Failed to load config: %s", err.Error())
					ui.ShowResult()

					failedToLoadConfig(ui, os.TempDir()+"/abm/public")
					go net.HostLocally(os.TempDir()+"/abm/public", "8080")
					continue
				} else {
					cfg = ncfg
					c.configUpdate = true
				}
			} 

			err = startParse2(cfg, c)
			if err != nil {
				failedToRender(cfg)
			} else {
				ui.ShowResult()
				websocket.SendUpdate()
			}

		case key := <-keyChannel:
			switch key {
			case keys[0]:
				ui.Info("Reloading config...")
				cfg, err = config.CleanConfig(configloc, ui)
				if err != nil {
					failedToRender(cfg)
					continue
				}

				c.fullRebuild = true
				err = startParse2(cfg, c)
			case keys[1]:
				c.configUpdate = true
				err = startParse2(cfg, c)
			}

			if err != nil {
				failedToRender(cfg)
			} else {
				ui.ShowResult()
				websocket.SendUpdate()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				shutdown <- 0
				return
			}
			ui.Errorf("Error during watch... %s", err.Error())

		case <-shutdown:
			shutdown <- 0
			return
		}
	}
}
