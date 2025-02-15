package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Helper function to create a test archive
func createTestArchive(content []byte) ([]byte, string) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	header := &tar.Header{
		Name: fmt.Sprintf("dirclean-%s-%s%s", runtime.GOOS, runtime.GOARCH, BinaryExt),
		Mode: 0755,
		Size: int64(len(content)),
	}

	tarWriter.WriteHeader(header)
	tarWriter.Write(content)
	tarWriter.Close()
	gzWriter.Close()

	archiveData := buf.Bytes()
	hash := sha256.Sum256(archiveData)
	return archiveData, hex.EncodeToString(hash[:])
}

func TestGetBinaryName(t *testing.T) {
	tests := []struct {
		name     string
		isLegacy string
		want     string
	}{
		{
			name:     "normal binary",
			isLegacy: "false",
			want:     fmt.Sprintf("dirclean-%s-%s", runtime.GOOS, runtime.GOARCH),
		},
		{
			name:     "legacy binary",
			isLegacy: "true",
			want:     fmt.Sprintf("dirclean-%s-%s-legacy", runtime.GOOS, runtime.GOARCH),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldIsLegacy := IsLegacy
			IsLegacy = tt.isLegacy
			defer func() { IsLegacy = oldIsLegacy }()

			got := getBinaryName()
			if got != tt.want {
				t.Errorf("getBinaryName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "successful download",
			statusCode: http.StatusOK,
			response:   "test content",
			wantErr:    false,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   "not found",
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   "server error",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			got, err := downloadFile(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("downloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.response {
				t.Errorf("downloadFile() = %v, want %v", string(got), tt.response)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	archiveData, checksum := createTestArchive([]byte("test content"))

	tests := []struct {
		name          string
		checksumData  string
		archiveData   []byte
		wantErr       bool
		errorContains string
	}{
		{
			name:         "valid checksum",
			checksumData: fmt.Sprintf("%s  %s\n", checksum, ArchiveName),
			archiveData:  archiveData,
			wantErr:      false,
		},
		{
			name:          "invalid checksum",
			checksumData:  fmt.Sprintf("invalidchecksum  %s\n", ArchiveName),
			archiveData:   archiveData,
			wantErr:       true,
			errorContains: "checksum mismatch",
		},
		{
			name:          "checksum not found",
			checksumData:  "otherfile  checksumvalue\n",
			archiveData:   archiveData,
			wantErr:       true,
			errorContains: "checksum not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.checksumData))
			}))
			defer server.Close()

			oldChecksumURL := ChecksumURL
			ChecksumURL = server.URL + "/%s"
			defer func() { ChecksumURL = oldChecksumURL }()

			err := verifyChecksum(tt.archiveData, "latest")
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("verifyChecksum() error = %v, want error containing %v", err, tt.errorContains)
			}
		})
	}
}

func TestExtractBinary(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		wantErr     bool
		verifyFiles bool
	}{
		{
			name:        "valid archive",
			content:     []byte("test binary content"),
			wantErr:     false,
			verifyFiles: true,
		},
		{
			name:    "invalid archive",
			content: []byte("invalid content"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var archiveData []byte
			if tt.wantErr {
				archiveData = tt.content
			} else {
				archiveData, _ = createTestArchive(tt.content)
			}

			binaryPath, err := extractBinary(archiveData, tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.verifyFiles {
				if _, err := os.Stat(binaryPath); err != nil {
					t.Errorf("Binary file not created: %v", err)
				}

				content, err := os.ReadFile(binaryPath)
				if err != nil {
					t.Errorf("Failed to read binary: %v", err)
				}
				if string(content) != string(tt.content) {
					t.Errorf("Binary content = %v, want %v", string(content), string(tt.content))
				}
			}
		})
	}
}

func TestUpdateBinary(t *testing.T) {
	// Create mock servers for GitHub API and release download
	latestTag := "v1.0.0"
	binaryContent := []byte("test binary content")
	archiveData, checksum := createTestArchive(binaryContent)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"tag_name": "%s"}`, latestTag)
	}))
	defer apiServer.Close()

	releaseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "checksums.txt") {
			fmt.Fprintf(w, "%s  %s\n", checksum, ArchiveName)
			return
		}
		w.Write(archiveData)
	}))
	defer releaseServer.Close()

	// Create a temporary executable for testing
	tmpDir := t.TempDir()
	tmpExe := filepath.Join(tmpDir, "current"+BinaryExt)
	if err := os.WriteFile(tmpExe, []byte("old content"), 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	// Override global variables for testing
	oldUpdateURL := UpdateURL
	oldChecksumURL := ChecksumURL
	oldGetExecutable := getExecutable
	defer func() {
		UpdateURL = oldUpdateURL
		ChecksumURL = oldChecksumURL
		getExecutable = oldGetExecutable
	}()

	UpdateURL = releaseServer.URL + "/%s/" + ArchiveName
	ChecksumURL = releaseServer.URL + "/%s/checksums.txt"
	getExecutable = func() (string, error) { return tmpExe, nil }

	tests := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{
			name:    "update to latest",
			tag:     "latest",
			wantErr: false,
		},
		{
			name:    "update to specific version",
			tag:     "v1.0.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateBinary(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, err := os.ReadFile(tmpExe)
				if err != nil {
					t.Errorf("Failed to read updated binary: %v", err)
				}
				if string(content) != string(binaryContent) {
					t.Errorf("Updated binary content = %v, want %v", string(content), string(binaryContent))
				}
			}
		})
	}
}
