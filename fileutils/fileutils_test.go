package fileutils

import (
	"os"
	"testing"
)

func TestCountLines(t *testing.T) {
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	lines := []string{"line1", "line2", "line3"}
	for _, line := range lines {
		_, err := tempFile.WriteString(line + "\n")
		if err != nil {
			t.Fatalf("Error writing to temp file: %v", err)
		}
	}

	count := CountLines(tempFile.Name())
	if count != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), count)
	}
}
