package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gitlab.com/antipy/antibuild/cli/builder"
)

var configPath = "config.json"

func getConfig() (config *builder.Config) {
	config = new(builder.Config)
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func saveConfig(config *builder.Config) {
	file, err := os.Create(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}
