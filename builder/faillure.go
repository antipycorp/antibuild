package builder

import (
	"io/ioutil"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder/config"
)

var (
	failledtoloadconfig = []byte(
		"<html>\n" +
			"failled to load the config file, probably json syntax error :P <br/> more info avaiable in the console output\n" +
			"</html>")
	failledtorender = []byte(
		"<html>\n" +
			"failled to render any part of the file, report bugs and suggestions for better error messages at gitlab/github :P <br/> more info avaiable in the console output\n" +
			"</html>")
)

func failledToLoadConfig(log config.Logger, output string) {
	var err error
	err = os.MkdirAll(output, 0700)
	if err != nil {
		log.Fatalf("could not place error file in place: %s", err.Error())
	}
	err = ioutil.WriteFile(output+"/index.html", failledtoloadconfig, 0700)
	if err != nil {
		log.Fatalf("could not place error file in place: %s", err.Error())
	}
}

func failledToRender(cfg *config.Config) {
	var err error
	if cfg.Folders.Output == "" {
		err = os.RemoveAll(cfg.Folders.Output)
		if err != nil {
			cfg.UILogger.Fatalf("could not remove old files: %s", err.Error())
		}
	}
	err = ioutil.WriteFile(cfg.Folders.Output+"/index.html", failledtorender, 0644)
	if err != nil {
		cfg.UILogger.Fatalf("could not place error file in place: %s", err.Error())
	}
}
