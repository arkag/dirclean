package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		want       GlobalConfig
	}{
		{
			name: "valid config",
			configYAML: `
defaults:
  mode: dry-run
  log_level: INFO
  log_file: dirclean.log
rules:
  - delete_older_than_days: 30
    paths:
      - /tmp/test1
      - /tmp/test2
    mode: analyze`,
			want: GlobalConfig{
				Defaults: Config{
					Mode:     "dry-run",
					LogLevel: "INFO",
					LogFile:  "dirclean.log",
				},
				Rules: []Config{
					{
						DeleteOlderThanDays: 30,
						Paths:               []string{"/tmp/test1", "/tmp/test2"},
						Mode:                "analyze",
					},
				},
			},
		},
		{
			name: "empty config",
			configYAML: `
defaults:
  mode: dry-run
rules: []`,
			want: GlobalConfig{
				Defaults: Config{
					Mode: "dry-run",
				},
				Rules: []Config{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			got := LoadConfig(tmpFile)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Mode:                "dry-run",
				LogLevel:            "INFO",
				DeleteOlderThanDays: 30,
				Paths:               []string{"/tmp/test"},
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: Config{
				Mode: "invalid-mode",
			},
			wantErr: true,
		},
		{
			name: "negative days",
			config: Config{
				DeleteOlderThanDays: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: Config{
				LogLevel: "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeWithFlags(t *testing.T) {
	tests := []struct {
		name  string
		cfg   Config
		flags CLIFlags
		want  Config
	}{
		{
			name: "cli overrides config",
			cfg: Config{
				DeleteOlderThanDays: 30,
				Mode:                "analyze",
				LogLevel:            "INFO",
			},
			flags: CLIFlags{
				Mode:     "dry-run",
				LogLevel: "DEBUG",
			},
			want: Config{
				DeleteOlderThanDays: 30,
				Mode:                "dry-run",
				LogLevel:            "DEBUG",
			},
		},
		{
			name: "empty flags preserve config",
			cfg: Config{
				DeleteOlderThanDays: 30,
				Mode:                "analyze",
				LogLevel:            "INFO",
			},
			flags: CLIFlags{},
			want: Config{
				DeleteOlderThanDays: 30,
				Mode:                "analyze",
				LogLevel:            "INFO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeWithFlags(tt.cfg, tt.flags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeWithFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}
