package fileutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCountLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name:    "empty file",
			content: "",
			want:    0,
		},
		{
			name:    "single line",
			content: "line1",
			want:    1,
		},
		{
			name:    "multiple lines",
			content: "line1\nline2\nline3",
			want:    3,
		},
		{
			name:    "blank lines",
			content: "\n\n\n",
			want:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "test.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got := CountLines(tmpFile)
			if got != tt.want {
				t.Errorf("CountLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDF(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "root path",
			path:    "/",
			wantErr: false,
		},
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "non-existent path",
			path:    "/nonexistent/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDF(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got["Total"] <= 0 || got["Available"] <= 0) {
				t.Errorf("GetDF() = %v, want > 0", got)
			}
		})
	}
}

func TestIsOlderThan(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		fileAge   time.Duration
		checkDays int
		want      bool
	}{
		{
			name:      "file older than check",
			fileAge:   time.Hour * 24 * 10,
			checkDays: 5,
			want:      true,
		},
		{
			name:      "file newer than check",
			fileAge:   time.Hour * 24 * 2,
			checkDays: 5,
			want:      false,
		},
		{
			name:      "file same age as check",
			fileAge:   time.Hour * 24 * 5,
			checkDays: 5,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name)
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			modTime := time.Now().Add(-tt.fileAge)
			if err := os.Chtimes(testFile, modTime, modTime); err != nil {
				t.Fatalf("Failed to set file time: %v", err)
			}

			got := IsOlderThan(testFile, tt.checkDays)
			if got != tt.want {
				t.Errorf("IsOlderThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileSize(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantSize int64
		wantErr  bool
	}{
		{
			name:     "empty file",
			content:  "",
			wantSize: 0,
		},
		{
			name:     "small file",
			content:  "test content",
			wantSize: 12,
		},
		{
			name:    "non-existent file",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if !tt.wantErr {
				tmpFile := filepath.Join(t.TempDir(), "test.txt")
				if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				path = tmpFile
			} else {
				path = "/nonexistent/file.txt"
			}

			got, err := GetFileSize(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantSize {
				t.Errorf("GetFileSize() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name    string
		size    string
		want    int64
		wantErr bool
	}{
		{"bytes", "1024", 1024, false},
		{"kilobytes", "1KB", 1024, false},
		{"megabytes", "1MB", 1024 * 1024, false},
		{"gigabytes", "1GB", 1024 * 1024 * 1024, false},
		{"invalid format", "1XB", 0, true},
		{"negative", "-1KB", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
