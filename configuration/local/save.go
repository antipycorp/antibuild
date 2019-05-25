// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package local

import (
	"encoding/json"
	"os"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

//SaveConfig saves the config file
func SaveConfig(configLocation string, cfg *Config) errors.Error {
	file, err := os.Create(configLocation + ".new")
	if err != nil {
		return ErrFailedOpen.SetRoot(err.Error())
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cfg)
	if err != nil {
		return ErrFailedWrite.SetRoot(err.Error())
	}

	os.Rename(configLocation+".new", configLocation)

	return nil
}
