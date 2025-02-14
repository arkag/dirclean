package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile := "test_config.yaml"
	configContent := `
defaults:
  mode: dry-run
  log_level: INFO
  log_file: dirclean.log

rules:
  - delete_older_than_days: 30
    paths:
      - /tmp/test_dir
    mode: analyze
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Error writing test config file: %v", err)
	}
	defer os.Remove(configFile)

	config := LoadConfig(configFile)
	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}

	if config.Rules[0].DeleteOlderThanDays != 30 {
		t.Errorf("Expected DeleteOlderThanDays to be 30, got %d", config.Rules[0].DeleteOlderThanDays)
	}

	if config.Rules[0].Mode != "analyze" {
		t.Errorf("Expected Mode to be 'analyze', got %s", config.Rules[0].Mode)
	}

	if config.Defaults.Mode != "dry-run" {
		t.Errorf("Expected default Mode to be 'dry-run', got %s", config.Defaults.Mode)
	}
}
