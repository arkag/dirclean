package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateBinary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test binary content"))
	}))
	defer ts.Close()

	AppVersion = "test"
	AppOsArch = "test"

	err := UpdateBinary("test")
	if err != nil {
		t.Errorf("Error updating binary: %v", err)
	}
}
