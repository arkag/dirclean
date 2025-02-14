package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arkag/dirclean/logging"
)

var (
	AppVersion    = "unknown"
	AppOsArch     = "unknown"
	IsLegacy      = "false"
	BinaryName    = getBinaryName()
	ArchiveName   = fmt.Sprintf("%s.tar.gz", BinaryName)
	UpdateURL     = fmt.Sprintf("https://github.com/arkag/dirclean/releases/download/%%s/%s", ArchiveName)
	ChecksumURL   = "https://github.com/arkag/dirclean/releases/download/%s/checksums.txt"
	BinaryExt     = ""
	getExecutable = os.Executable
)

func getBinaryName() string {
	base := fmt.Sprintf("dirclean-%s-%s", runtime.GOOS, runtime.GOARCH)
	if IsLegacy == "true" {
		return base + "-legacy"
	}
	return base
}

func init() {
	if runtime.GOOS == "windows" {
		BinaryExt = ".exe"
	}
}

func downloadFile(url string) ([]byte, error) {
	logging.LogMessage("DEBUG", fmt.Sprintf("Attempting to download from: %s", url))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s\nURL: %s\nResponse: %s",
			resp.StatusCode, resp.Status, url, string(body))
	}

	return io.ReadAll(resp.Body)
}

func verifyChecksum(data []byte, tag string) error {
	// Download checksums file
	checksumURL := fmt.Sprintf(ChecksumURL, tag)
	resp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums: %s", resp.Status)
	}

	// Calculate checksum of downloaded archive
	calculatedHash := sha256.Sum256(data)
	calculatedHashStr := hex.EncodeToString(calculatedHash[:])

	// Find and verify checksum
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if strings.HasSuffix(parts[1], ArchiveName) {
			if parts[0] != calculatedHashStr {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", parts[0], calculatedHashStr)
			}
			return nil
		}
	}

	return fmt.Errorf("checksum not found for %s", ArchiveName)
}

func extractBinary(archiveData []byte, tmpDir string) (string, error) {
	gzReader, err := gzip.NewReader(strings.NewReader(string(archiveData)))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	binaryPath := ""

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar: %v", err)
		}

		if header.Typeflag == tar.TypeReg {
			// Extract only the binary file
			if strings.HasPrefix(header.Name, "dirclean") {
				binaryPath = filepath.Join(tmpDir, filepath.Base(header.Name))
				outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, 0755)
				if err != nil {
					return "", fmt.Errorf("failed to create binary file: %v", err)
				}
				defer outFile.Close()

				if _, err := io.Copy(outFile, tarReader); err != nil {
					return "", fmt.Errorf("failed to write binary file: %v", err)
				}
				break
			}
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

func UpdateBinary(tag string) error {
	if tag == "latest" {
		logging.LogMessage("DEBUG", "Fetching latest release tag...")
		// Add logic to fetch latest release tag from GitHub API
		resp, err := http.Get("https://api.github.com/repos/arkag/dirclean/releases/latest")
		if err != nil {
			return fmt.Errorf("failed to fetch latest release: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch latest release: HTTP %d", resp.StatusCode)
		}

		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("failed to parse release info: %v", err)
		}
		tag = release.TagName
		logging.LogMessage("DEBUG", fmt.Sprintf("Latest release tag: %s", tag))
	}

	downloadURL := fmt.Sprintf(UpdateURL, tag)
	logging.LogMessage("DEBUG", fmt.Sprintf("Download URL: %s", downloadURL))

	// Download archive
	archiveData, err := downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	// Verify checksum
	logging.LogMessage("DEBUG", "Verifying checksum...")
	if err := verifyChecksum(archiveData, tag); err != nil {
		return fmt.Errorf("checksum verification failed: %v", err)
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "dirclean-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract binary
	logging.LogMessage("DEBUG", "Extracting binary...")
	binaryPath, err := extractBinary(archiveData, tmpDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %v", err)
	}

	// Get current executable path
	executable, err := getExecutable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}

	// Replace current binary
	logging.LogMessage("DEBUG", fmt.Sprintf("Replacing binary at %s", executable))
	if err := os.Rename(binaryPath, executable); err != nil {
		return fmt.Errorf("error replacing binary: %v", err)
	}

	logging.LogMessage("INFO", fmt.Sprintf("Successfully updated to version %s", tag))
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
