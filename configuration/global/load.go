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

// OpenGlobalConfig opens and parses a global config file
func OpenGlobalConfig(path string) (*Config, error) {
	c := &Config{}

	err := c.Load(path)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Load loads the global config
func (c *Config) Load(path string) error {
	newc := Config{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = newc.Save(path)
			if err != nil {
				return err
			}
			*c = newc
			return nil
		}
		return err
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(&newc)
	if err != nil {
		return err
	}

	*c = newc
	return nil

}

// LoadDefaultGlobal loads the default global config file
func LoadDefaultGlobal() error {
	configPath := configdir.LocalConfig("antibuild")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")

	DefaultGlobalConfig, err = OpenGlobalConfig(configFile)
	if err != nil {
		return err
	}

	return nil
}
