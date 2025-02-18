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

// LoadConfig attempts to load the config file, falling back to the example config if specified file doesn't exist
func LoadConfig(configFile string) GlobalConfig {
	var err error
	var f *os.File

	// Try to open the specified config file
	f, err = os.Open(configFile)
	if err != nil {
		// If the specified file doesn't exist, try the example config
		examplePath := GetExampleConfigPath()
		f, err = os.Open(examplePath)
		if err != nil {
			logging.LogMessage("FATAL", fmt.Sprintf("Error opening config file %s or example config %s: %v", configFile, examplePath, err))
			os.Exit(1)
		}
		logging.LogMessage("INFO", fmt.Sprintf("Using example config from %s", examplePath))
	}
	defer f.Close()

	var globalConfig GlobalConfig
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&globalConfig); err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error decoding YAML: %v", err))
		os.Exit(1)
	}

	// Merge defaults with each rule
	for i := range globalConfig.Rules {
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
		if !globalConfig.Rules[i].CleanBrokenSymlinks {
			globalConfig.Rules[i].CleanBrokenSymlinks = globalConfig.Defaults.CleanBrokenSymlinks
		}
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
