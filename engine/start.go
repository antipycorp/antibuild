// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package engine

import (
	"io"
	"os"
	"time"

	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"
	"gitlab.com/antipy/antibuild/cli/internal/net"
)

//Start the build process
func Start(isRefreshEnabled bool, isHost bool, configLocation string, isConfigSet bool, port string) {
	UI := &ui.UI{}
	cfg, err := localConfig.CleanConfig(configLocation, UI, true)
	if err != nil {
		failedToLoadConfig(UI, "public/")
		return
	}

	if isConfigSet {
		UI.HostingEnabled = isHost
		UI.Port = port

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
			buildOnRefresh(cfg, configLocation, UI)
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
	UI := &ui.UI{}
	UI.SetLogfile(output)
	UI.SetPrettyPrint(false)

	cfg, err := localConfig.CleanConfig(configLocation, UI, false)
	UI.SetLogfile(output)
	UI.SetPrettyPrint(false)

	if err != nil {
		UI.Fatalf("could not parse the config file: %s", err.Error())
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

func startParse(cfg *localConfig.Config) (*cache, errors.Error) {
	cache := &cache{
		rootPage:     *cfg.Pages,
		data:         make(map[string]cacheData),
		configUpdate: true,
		checkData:    false,
	}
	return cache, startCachedParse(cfg, cache)
}
