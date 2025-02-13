package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestUpdateBinary(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test binary content"))
	}))
	defer ts.Close()

	// Store original URL format
	originalBinaryName := fmt.Sprintf("dirclean-%s-%s", runtime.GOOS, runtime.GOARCH)

	// Override the binary name and URL for testing
	binaryName = originalBinaryName
	oldURL := fmt.Sprintf("https://github.com/arkag/dirclean/releases/download/%%s/%s", binaryName)

	// Replace the URL with test server URL
	url = ts.URL + "/" + binaryName

	// Run test
	err := UpdateBinary("test")
	if err != nil {
		t.Errorf("Error updating binary: %v", err)
	}

	// Restore original URL format
	url = oldURL
}
