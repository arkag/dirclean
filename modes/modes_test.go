package modes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arkag/dirclean/config"
)

func TestProcessFiles(t *testing.T) {
	testDir := t.TempDir() + "/test_dir"

	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Error creating test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	oldFile := filepath.Join(testDir, "oldfile.txt")
	err = os.WriteFile(oldFile, []byte("old content"), 0644)
	if err != nil {
		t.Fatalf("Error creating old file: %v", err)
	}

	oldTime := time.Now().AddDate(0, 0, -10)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Error setting file time: %v", err)
	}

	config := config.Config{
		DeleteOlderThanDays: 5,
		Paths:               []string{testDir},
		Mode:                "dry-run",
	}

	tempFile, err := os.CreateTemp("", "cleanup_")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	ProcessFiles(config, tempFile)
	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Error reading temp file: %v", err)
	}

	if !strings.Contains(string(content), oldFile) {
		t.Errorf("Old file not found in temp file")
	}
}
