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
	"testing"
	"time"
)

func createTestArchive(content []byte) ([]byte, string) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	// Add binary to archive
	header := &tar.Header{
		Name:    BinaryName,
		Size:    int64(len(content)),
		Mode:    0755,
		ModTime: time.Now(),
	}

	tarWriter.WriteHeader(header)
	tarWriter.Write(content)
	tarWriter.Close()
	gzWriter.Close()

	archiveData := buf.Bytes()
	hash := sha256.Sum256(archiveData)
	return archiveData, hex.EncodeToString(hash[:])
}

func TestUpdateBinary(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "dirclean-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy executable in the temp directory
	testBinary := filepath.Join(tmpDir, "dirclean-test"+BinaryExt)
	if err := os.WriteFile(testBinary, []byte("dummy content"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Create test archive
	archiveData, checksum := createTestArchive([]byte("test binary content"))

	// Create test servers
	binaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(archiveData)
	}))
	defer binaryServer.Close()

	checksumServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", checksum, ArchiveName)
	}))
	defer checksumServer.Close()

	// Store original values
	originalUpdateURL := UpdateURL
	originalChecksumURL := ChecksumURL
	originalGetExecutable := getExecutable

	// Override the URLs and executable path for testing
	UpdateURL = binaryServer.URL + "/%s"
	ChecksumURL = checksumServer.URL + "/%s"
	getExecutable = func() (string, error) {
		return testBinary, nil
	}

	// Run test
	err = UpdateBinary("test")
	if err != nil {
		t.Errorf("Error updating binary: %v", err)
	}

	// Verify the content was updated
	content, err := os.ReadFile(testBinary)
	if err != nil {
		t.Errorf("Failed to read updated binary: %v", err)
	}
	if string(content) != "test binary content" {
		t.Errorf("Binary content not updated correctly")
	}

	// Restore original values
	UpdateURL = originalUpdateURL
	ChecksumURL = originalChecksumURL
	getExecutable = originalGetExecutable
}
