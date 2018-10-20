package builder

import (
	"encoding/json"
	"fmt"
	"os"
)

var configPath = "config.json"

func GetConfig() (config *Config) {
	config = new(Config)
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

func SaveConfig(config *Config) {
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
