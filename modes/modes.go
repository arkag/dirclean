package modes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arkag/dirclean/config"
	"github.com/arkag/dirclean/logging"
)

func ProcessFiles(config config.Config, tempFile *os.File) {
	days := config.DeleteOlderThanDays
	paths := config.Paths

	if days <= 0 {
		logging.LogMessage("ERROR", fmt.Sprintf("Invalid days value: %d", days))
		return
	}

	// Convert file size limits to bytes
	var minBytes, maxBytes int64
	if config.MinFileSize != nil {
		minBytes = config.MinFileSize.ToBytes()
	}
	if config.MaxFileSize != nil {
		maxBytes = config.MaxFileSize.ToBytes()
	}

	matchedDirs := ValidateDirs(paths)

	for _, dir := range matchedDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", path, err))
				return nil
			}
			if !info.IsDir() {
				fileSize := info.Size()

				// Check file size constraints
				if (minBytes > 0 && fileSize < minBytes) ||
					(maxBytes > 0 && fileSize > maxBytes) {
					return nil
				}

				modTime := info.ModTime()
				cutoff := time.Now().AddDate(0, 0, -days)
				if modTime.Before(cutoff) {
					switch config.Mode {
					case "analyze":
						logging.LogMessage("INFO", fmt.Sprintf("Found candidate: %s (size: %d bytes, modified: %s)",
							path, fileSize, modTime))
					case "dry-run":
						logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s", path))
						fmt.Fprintln(tempFile, path)
					case "interactive":
						fmt.Printf("Delete file %s? (size: %d bytes, modified: %s) (y/n): ",
							path, fileSize, modTime)
						var response string
						fmt.Scanln(&response)
						if response == "y" || response == "Y" {
							if err := os.Remove(path); err != nil {
								logging.LogMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
							} else {
								logging.LogMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
								fmt.Fprintln(tempFile, path)
							}
						}
					case "scheduled":
						if err := os.Remove(path); err != nil {
							logging.LogMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
						} else {
							logging.LogMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
							fmt.Fprintln(tempFile, path)
						}
					default:
						logging.LogMessage("WARN", fmt.Sprintf("Unknown mode: %s, defaulting to dry-run", config.Mode))
						logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s", path))
						fmt.Fprintln(tempFile, path)
					}
				}
			}
			return nil
		})
		if err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error walking directory %s: %v", dir, err))
		}
	}
}

func ValidateDirs(dirs []string) []string {
	var matchedDirs []string
	for _, dir := range dirs {
		if strings.Contains(dir, "*") {
			matches, err := filepath.Glob(dir)
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error with wildcard path: %v", err))
				continue
			}
			for _, match := range matches {
				if info, err := os.Stat(match); err == nil && info.IsDir() {
					logging.LogMessage("DEBUG", fmt.Sprintf("Matched directory: %s", match))
					matchedDirs = append(matchedDirs, match)
				}
			}
		} else if info, err := os.Stat(dir); err == nil && info.IsDir() {
			logging.LogMessage("DEBUG", fmt.Sprintf("Matched directory: %s", dir))
			matchedDirs = append(matchedDirs, dir)
		} else {
			logging.LogMessage("ERROR", fmt.Sprintf("Directory does not exist or is not accessible: %s", dir))
		}
	}
	return matchedDirs
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
