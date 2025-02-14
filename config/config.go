package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/arkag/dirclean/logging"
	"gopkg.in/yaml.v3"
)

type FileSize struct {
	Value float64
	Unit  string
}

type Config struct {
	DeleteOlderThanDays int       `yaml:"delete_older_than_days"`
	Paths               []string  `yaml:"paths"`
	MinFileSize         *FileSize `yaml:"min_file_size,omitempty"`
	MaxFileSize         *FileSize `yaml:"max_file_size,omitempty"`
	Mode                string    `yaml:"mode,omitempty"`
	LogLevel            string    `yaml:"log_level,omitempty"`
	LogFile             string    `yaml:"log_file,omitempty"`
}

type GlobalConfig struct {
	Defaults Config   `yaml:"defaults"`
	Rules    []Config `yaml:"rules"`
}

// UnmarshalYAML implements custom unmarshaling for FileSize
func (f *FileSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sizeStr string
	if err := unmarshal(&sizeStr); err != nil {
		return err
	}

	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	var value float64
	var unit string

	_, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit)
	if err != nil {
		return fmt.Errorf("invalid file size format: %s", sizeStr)
	}

	f.Value = value
	f.Unit = unit
	return nil
}

// ToBytes converts FileSize to bytes
func (f *FileSize) ToBytes() int64 {
	multiplier := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	if m, ok := multiplier[f.Unit]; ok {
		return int64(f.Value * float64(m))
	}
	return 0
}

func LoadConfig(configFile string) GlobalConfig {
	f, err := os.Open(configFile)
	if err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error opening YAML file: %v", err))
		os.Exit(1)
	}
	defer f.Close()

	var globalConfig GlobalConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&globalConfig); err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error decoding YAML: %v", err))
		os.Exit(1)
	}

	// Set default values if not specified
	if globalConfig.Defaults.Mode == "" {
		globalConfig.Defaults.Mode = "dry-run"
	}
	if globalConfig.Defaults.LogLevel == "" {
		globalConfig.Defaults.LogLevel = "INFO"
	}
	if globalConfig.Defaults.LogFile == "" {
		globalConfig.Defaults.LogFile = "dirclean.log"
	}

	logging.LogMessage("DEBUG", fmt.Sprintf("Loaded config: %+v", globalConfig))
	return globalConfig
}

// MergeWithFlags merges CLI flags with config values
func MergeWithFlags(config Config, flags CLIFlags) Config {
	// CLI flags take precedence over config file values
	if flags.Mode != "" {
		config.Mode = flags.Mode
	}
	if flags.LogFile != "" {
		config.LogFile = flags.LogFile
	}
	if flags.LogLevel != "" {
		config.LogLevel = flags.LogLevel
	}
	if flags.MinFileSize != nil {
		config.MinFileSize = flags.MinFileSize
	}
	if flags.MaxFileSize != nil {
		config.MaxFileSize = flags.MaxFileSize
	}
	return config
}

type CLIFlags struct {
	Mode        string
	LogFile     string
	LogLevel    string
	MinFileSize *FileSize
	MaxFileSize *FileSize
}
