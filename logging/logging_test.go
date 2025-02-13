package logging

import (
	"os"
	"strings"
	"testing"
)

func TestLogMessage(t *testing.T) {
	logFile := "test_log.log"
	InitLogging(logFile)
	defer os.Remove(logFile)

	LogMessage("INFO", "Test log message")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Error reading log file: %v", err)
	}

	if !strings.Contains(string(content), "Test log message") {
		t.Errorf("Log message not found in log file")
	}
}
