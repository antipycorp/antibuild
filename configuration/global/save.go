// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package global

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
)

// SaveDefaultGlobal config file
func SaveDefaultGlobal() error {
	configPath := configdir.LocalConfig("antibuild")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")

	err = DefaultGlobalConfig.Save(configFile)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the global config
func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(c)
	if err != nil {
		return err
	}

	return nil
}
