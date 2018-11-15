package builder

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	"gitlab.com/antipy/antibuild/cli/internal"
)

//watches files and folders and rebuilds when things change
func buildOnRefresh(cfg *config.Config, configLocation string) {
	ui := cfg.UILogger // the UI logger cannot change
	ui.Debug("making a refresh")

	startParse(cfg)

	shutdown := make(chan int, 10) // 10 is enough for some watcher to not get stuck on the chan
	if cfg.Folders.Static != "" {
		go staticWatch(cfg.Folders.Static, cfg.Folders.Output, shutdown, ui)
	}
	watchBuild(cfg.Folders.Templates, configLocation, shutdown, ui)

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
func watchBuild(src, configloc string, shutdown chan int, log config.UIlogger) {
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

	err = watcher.Add(configloc)
	if err != nil {
		log.Infof("could not watch config file %s", err.Error())
	}
	for {
		//listen for watcher events
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				shutdown <- 0
				return
			}
			log.Debug("making a refresh")
			cfg, configErr := parseConfig(configloc)
			if configErr != nil {
				log.Errorf("Could not parse the config file: %s", configErr)
				continue
			}

			cfg.UILogger = log

			startParse(cfg)

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
