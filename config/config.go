package config

import (
	"fmt"
	"os"

	"github.com/arkag/dirclean/logging"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DeleteOlderThanDays int      `yaml:"delete_older_than_days"`
	Paths               []string `yaml:"paths"`
}

func LoadConfig(configFile string) []Config {
	f, err := os.Open(configFile)
	if err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error opening YAML file: %v", err))
	}
	defer f.Close()

	var configs []Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&configs); err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error decoding YAML: %v", err))
	}

	logging.LogMessage("DEBUG", fmt.Sprintf("Loaded configs: %+v", configs))
	return configs
}
