package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory
	dirPath := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a valid symlink
	validSymlink := filepath.Join(tmpDir, "valid.symlink")
	if err := os.Symlink(regularFile, validSymlink); err != nil {
		t.Fatalf("Failed to create valid symlink: %v", err)
	}

	// Create a broken symlink
	brokenSymlink := filepath.Join(tmpDir, "broken.symlink")
	if err := os.Symlink(filepath.Join(tmpDir, "nonexistent"), brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Create a relative symlink
	relativeSymlink := filepath.Join(tmpDir, "relative.symlink")
	if err := os.Symlink("regular.txt", relativeSymlink); err != nil {
		t.Fatalf("Failed to create relative symlink: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "regular file exists",
			path: regularFile,
			want: true,
		},
		{
			name: "directory exists",
			path: dirPath,
			want: true,
		},
		{
			name: "valid symlink exists",
			path: validSymlink,
			want: true,
		},
		{
			name: "broken symlink does not exist",
			path: brokenSymlink,
			want: false,
		},
		{
			name: "relative symlink exists",
			path: relativeSymlink,
			want: true,
		},
		{
			name: "nonexistent file",
			path: filepath.Join(tmpDir, "nonexistent"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExists(tt.path)
			if got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a directory
	dirPath := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "regular file is not dir",
			path: regularFile,
			want: false,
		},
		{
			name: "directory is dir",
			path: dirPath,
			want: true,
		},
		{
			name: "nonexistent path is not dir",
			path: filepath.Join(tmpDir, "nonexistent"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDir(tt.path)
			if got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "already absolute path",
			path: filepath.Join(tmpDir, "file.txt"),
			want: filepath.Join(tmpDir, "file.txt"),
		},
		{
			name: "relative path",
			path: "file.txt",
			want: filepath.Join(currentDir, "file.txt"),
		},
		{
			name: "dot path",
			path: ".",
			want: currentDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAbsPath(tt.path)
			if got != tt.want {
				t.Errorf("GetAbsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a valid symlink
	validSymlink := filepath.Join(tmpDir, "valid.symlink")
	if err := os.Symlink(regularFile, validSymlink); err != nil {
		t.Fatalf("Failed to create valid symlink: %v", err)
	}

	// Create a broken symlink
	brokenSymlink := filepath.Join(tmpDir, "broken.symlink")
	if err := os.Symlink(filepath.Join(tmpDir, "nonexistent"), brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:    "regular file is not broken symlink",
			path:    regularFile,
			want:    false,
			wantErr: false,
		},
		{
			name:    "valid symlink is not broken",
			path:    validSymlink,
			want:    false,
			wantErr: false,
		},
		{
			name:    "broken symlink is broken",
			path:    brokenSymlink,
			want:    true,
			wantErr: false,
		},
		{
			name:    "nonexistent path",
			path:    filepath.Join(tmpDir, "nonexistent"),
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBrokenSymlink(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsBrokenSymlink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsBrokenSymlink() = %v, want %v", got, tt.want)
			}
		})
	}
}
