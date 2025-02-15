package logging

import (
	"os"
	"strings"
	"testing"
)

func TestInitLogging(t *testing.T) {
	tests := []struct {
		name     string
		logFile  string
		wantErr  bool
		setup    func() error
		cleanup  func()
		errCheck func(error) bool
	}{
		{
			name:    "valid log file",
			logFile: "test.log",
			cleanup: func() { os.Remove("test.log") },
		},
		{
			name:    "invalid directory",
			logFile: "/nonexistent/test.log",
			wantErr: true,
			errCheck: func(err error) bool {
				return strings.Contains(err.Error(), "no such file or directory")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := InitLogging(tt.logFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogging() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("InitLogging() error = %v, did not match expected error condition", err)
				}
			}
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel string
	}{
		{"debug level", "DEBUG", "DEBUG"},
		{"info level", "INFO", "INFO"},
		{"warn level", "WARN", "WARN"},
		{"error level", "ERROR", "ERROR"},
		{"fatal level", "FATAL", "FATAL"},
		{"invalid level", "INVALID", "INFO"}, // should default to INFO
		{"empty level", "", "INFO"},          // should default to INFO
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.level)
			if GetLogLevel() != tt.wantLevel {
				t.Errorf("SetLogLevel() = %v, want %v", GetLogLevel(), tt.wantLevel)
			}
		})
	}
}

func TestLogMessage(t *testing.T) {
	logFile := "test_log.log"

	tests := []struct {
		name     string
		level    string
		message  string
		wantLog  bool
		setup    func() error
		validate func(string) bool
	}{
		{
			name:    "info message",
			level:   "INFO",
			message: "test info message",
			wantLog: true,
			validate: func(content string) bool {
				return strings.Contains(content, "INFO") &&
					strings.Contains(content, "test info message")
			},
		},
		{
			name:    "debug message",
			level:   "DEBUG",
			message: "test debug message",
			wantLog: true,
			validate: func(content string) bool {
				return strings.Contains(content, "DEBUG") &&
					strings.Contains(content, "test debug message")
			},
		},
		{
			name:    "error message",
			level:   "ERROR",
			message: "test error message",
			wantLog: true,
			validate: func(content string) bool {
				return strings.Contains(content, "ERROR") &&
					strings.Contains(content, "test error message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitLogging(logFile); err != nil {
				t.Fatalf("Failed to initialize logging: %v", err)
			}
			defer os.Remove(logFile)

			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			LogMessage(tt.level, tt.message)

			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			if tt.wantLog {
				if !tt.validate(string(content)) {
					t.Errorf("Log content did not match expected format: %s", content)
				}
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		setLevel      string
		expectedLevel string
	}{
		{"debug level", "DEBUG", "DEBUG"},
		{"info level", "INFO", "INFO"},
		{"warn level", "WARN", "WARN"},
		{"error level", "ERROR", "ERROR"},
		{"fatal level", "FATAL", "FATAL"},
		{"default level", "", "INFO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.setLevel)
			level := GetLogLevel()
			if level != tt.expectedLevel {
				t.Errorf("GetLogLevel() = %v, want %v", level, tt.expectedLevel)
			}
		})
	}
}

func TestShouldLog(t *testing.T) {
	tests := []struct {
		name           string
		currentLevel   string
		messageLevel   string
		shouldLogValue bool
	}{
		{"debug shows debug", "DEBUG", "DEBUG", true},
		{"debug shows info", "DEBUG", "INFO", true},
		{"info hides debug", "INFO", "DEBUG", false},
		{"info shows info", "INFO", "INFO", true},
		{"error hides warn", "ERROR", "WARN", false},
		{"error shows fatal", "ERROR", "FATAL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.currentLevel)
			result := shouldLog(tt.messageLevel)
			if result != tt.shouldLogValue {
				t.Errorf("shouldLog() with currentLevel=%v, messageLevel=%v = %v, want %v",
					tt.currentLevel, tt.messageLevel, result, tt.shouldLogValue)
			}
		})
	}
}
