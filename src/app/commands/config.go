package commands

import (
	"autossh/src/utils"
)

type Config struct {
	Value string
}

func (config *Config) Process() bool {
	if config.Value == "" {
		config.Value = "./config.json"
	} else {
		path, err := utils.ParsePath(config.Value)
		if err == nil {
			config.Value = path
		} else {
			config.Value = ""
		}
	}

	return false
}
