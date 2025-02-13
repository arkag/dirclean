package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateBinary(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test binary content"))
	}))
	defer ts.Close()

	// Store original values
	originalBinaryName := BinaryName
	originalURL := UpdateURL

	// Override the URL with test server URL
	UpdateURL = ts.URL + "/dirclean"

	// Run test
	err := UpdateBinary("test")
	if err != nil {
		t.Errorf("Error updating binary: %v", err)
	}

	// Restore original values
	BinaryName = originalBinaryName
	UpdateURL = originalURL
}
