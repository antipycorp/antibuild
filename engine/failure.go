// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package engine

import (
	"io/ioutil"
	"log"
	"os"

	localConfig "gitlab.com/antipy/antibuild/cli/configuration/local"
)

var (
	failedtoloadconfig = []byte(
		"<html>\n" +
			"failed to load the config file, probably json syntax error :P <br/> more info avaiable in the console output\n" +
			"</html>")
	failedtorender = []byte(
		"<html>\n" +
			"failed to render any part of the file, report bugs and suggestions for better error messages at gitlab/github :P <br/> more info avaiable in the console output\n" +
			"</html>")
)

func failedToLoadConfig(log localConfig.Logger, output string) {
	err := os.MkdirAll(output, 0700)
	if err != nil {
		log.Fatalf("could not place error file in place: %s", err.Error())
	}
	err = ioutil.WriteFile(output+"/index.html", failedtoloadconfig, 0700)
	if err != nil {
		log.Fatalf("could not place error file in place: %s", err.Error())
	}
}

func failedToRender(cfg *localConfig.Config) {
	var err error
	if cfg.Folders.Output == "" {
		cfg.UILogger.Fatal("Output folder is not set.")
	}

	err = os.RemoveAll(cfg.Folders.Output)
	if err != nil {
		cfg.UILogger.Fatalf("Could not remove old files. %s", err.Error())
	}

	err = os.MkdirAll(cfg.Folders.Output, 0700)
	if err != nil {
		log.Fatalf("could not place error file in place: %s", err.Error())
	}

	err = ioutil.WriteFile(cfg.Folders.Output+"/index.html", failedtorender, 0644)
	if err != nil {
		cfg.UILogger.Fatalf("Could not place the 'error' file %s", err.Error())
	}

	cfg.UILogger.ShowResult()
}
