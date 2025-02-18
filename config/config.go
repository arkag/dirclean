package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arkag/dirclean/logging"
	"gopkg.in/yaml.v3"
)

type FileSize struct {
	Value float64
	Unit  string
}

type Config struct {
	OlderThanDays       int       `yaml:"older_than_days"`
	Paths               []string  `yaml:"paths"`
	MinFileSize         *FileSize `yaml:"min_file_size,omitempty"`
	MaxFileSize         *FileSize `yaml:"max_file_size,omitempty"`
	Mode                string    `yaml:"mode,omitempty"`
	LogLevel            string    `yaml:"log_level,omitempty"`
	LogFile             string    `yaml:"log_file,omitempty"`
	CleanBrokenSymlinks bool      `yaml:"clean_broken_symlinks,omitempty"`
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
func (fs *FileSize) ToBytes() int64 {
	multiplier := int64(1)
	switch strings.ToUpper(fs.Unit) {
	case "B":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	return int64(fs.Value * float64(multiplier))
}

// GetExampleConfigPath returns the OS-specific path for the example config
func GetExampleConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		// Check for Apple Silicon first (opt/homebrew)
		if _, err := os.Stat("/opt/homebrew/share/dirclean/example.config.yaml"); err == nil {
			return "/opt/homebrew/share/dirclean/example.config.yaml"
		}
		// Fallback to Intel Mac path (usr/local)
		return "/usr/local/share/dirclean/example.config.yaml"
	case "linux":
		return "/usr/share/dirclean/example.config.yaml"
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "dirclean", "example.config.yaml")
	default:
		return "example.config.yaml"
	}
}

// getDefaultConfigPath returns the OS-specific system-level path for the default config
func getDefaultConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		// Check for Apple Silicon first (opt/homebrew)
		if _, err := os.Stat("/opt/homebrew/etc/dirclean"); err == nil {
			return "/opt/homebrew/etc/dirclean/config.yaml"
		}
		// Fallback to Intel Mac path
		return "/usr/local/etc/dirclean/config.yaml"
	case "linux":
		return "/etc/dirclean/config.yaml"
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "dirclean", "config.yaml")
	default:
		return "config.yaml"
	}
}

// LoadConfig attempts to load the config file from the specified path or default location
func LoadConfig(configFile string) GlobalConfig {
	var err error
	var f *os.File

	// If no config file is specified, use the default path
	if configFile == "config.yaml" {
		configFile = getDefaultConfigPath()
	}

	// Try to open the config file
	f, err = os.Open(configFile)
	if err != nil {
		examplePath := GetExampleConfigPath()
		logging.LogMessage("FATAL", fmt.Sprintf(
			"Could not find config file at %s\n"+
				"To get started, copy the example config:\n"+
				"Example config location: %s\n"+
				"Run: sudo mkdir -p $(dirname %s) && sudo cp %s %s",
			configFile, examplePath, configFile, examplePath, configFile))
		os.Exit(1)
	}
	defer f.Close()

	var globalConfig GlobalConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&globalConfig); err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error decoding YAML: %v", err))
		os.Exit(1)
	}

	logging.LogMessage("DEBUG", fmt.Sprintf("Loaded defaults: %+v", globalConfig.Defaults))

	// Merge defaults with each rule
	for i := range globalConfig.Rules {
		logging.LogMessage("DEBUG", fmt.Sprintf("Before merge - Rule %d: %+v", i, globalConfig.Rules[i]))

		if globalConfig.Rules[i].OlderThanDays == 0 {
			globalConfig.Rules[i].OlderThanDays = globalConfig.Defaults.OlderThanDays
		}
		if globalConfig.Rules[i].Mode == "" {
			globalConfig.Rules[i].Mode = globalConfig.Defaults.Mode
		}
		if globalConfig.Rules[i].LogLevel == "" {
			globalConfig.Rules[i].LogLevel = globalConfig.Defaults.LogLevel
		}
		if globalConfig.Rules[i].LogFile == "" {
			globalConfig.Rules[i].LogFile = globalConfig.Defaults.LogFile
		}
		// Change this to check explicitly for false
		if !globalConfig.Rules[i].CleanBrokenSymlinks {
			globalConfig.Rules[i].CleanBrokenSymlinks = globalConfig.Defaults.CleanBrokenSymlinks
		}

		logging.LogMessage("DEBUG", fmt.Sprintf("After merge - Rule %d: %+v", i, globalConfig.Rules[i]))
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

// ValidateConfig validates the configuration values
func ValidateConfig(config Config) error {
	// Validate mode if specified
	if config.Mode != "" && config.Mode != "analyze" && config.Mode != "dry-run" &&
		config.Mode != "interactive" && config.Mode != "scheduled" {
		return fmt.Errorf("invalid mode: %s", config.Mode)
	}

	// Validate log level if specified
	if config.LogLevel != "" && config.LogLevel != "DEBUG" && config.LogLevel != "INFO" &&
		config.LogLevel != "WARN" && config.LogLevel != "ERROR" && config.LogLevel != "FATAL" {
		return fmt.Errorf("invalid log level: %s", config.LogLevel)
	}

	// Validate older_than_days
	if config.OlderThanDays < 0 {
		return fmt.Errorf("older_than_days must be non-negative, got: %d", config.OlderThanDays)
	}

	return nil
}
