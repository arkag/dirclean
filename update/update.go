package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/arkag/dirclean/logging"
)

var (
	AppVersion    = "unknown"
	AppOsArch     = "unknown"
	BinaryName    = fmt.Sprintf("dirclean-%s-%s", runtime.GOOS, runtime.GOARCH)
	UpdateURL     = fmt.Sprintf("https://github.com/arkag/dirclean/releases/download/%%s/%s", BinaryName)
	BinaryExt     = ""
	getExecutable = os.Executable
)

func init() {
	if runtime.GOOS == "windows" {
		BinaryExt = ".exe"
	}
}

func UpdateBinary(tag string) error {
	downloadURL := fmt.Sprintf(UpdateURL, tag)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download binary: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "dirclean-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}

	executable, err := getExecutable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %v", err)
	}

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("error setting executable permissions: %v", err)
	}

	if err := os.Rename(tmpFile.Name(), executable); err != nil {
		return fmt.Errorf("error replacing binary: %v", err)
	}

	return nil
}

func RestartBinary() {
	executable, err := os.Executable()
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting executable path: %v", err))
		return
	}

	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error restarting binary: %v", err))
		return
	}

	os.Exit(0)
}
