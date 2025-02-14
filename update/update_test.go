package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

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

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test binary content"))
	}))
	defer ts.Close()

	// Store original values
	originalURL := UpdateURL
	originalGetExecutable := getExecutable

	// Override the URL and executable path for testing
	UpdateURL = ts.URL + "/%s"
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
	UpdateURL = originalURL
	getExecutable = originalGetExecutable
}
