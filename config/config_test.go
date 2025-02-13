package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile := "test_config.yaml"
	configContent := `
- delete_older_than_days: 30
  paths:
    - /tmp/test_dir
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Error writing test config file: %v", err)
	}
	defer os.Remove(configFile)

	configs := LoadConfig(configFile)
	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}
}
