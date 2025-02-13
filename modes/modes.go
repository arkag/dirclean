package modes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arkag/dirclean/logging"
)

func ProcessFiles(config Config, tempFile *os.File, mode string) {
	days := config.DeleteOlderThanDays
	paths := config.Paths

	if days <= 0 {
		logging.LogMessage("ERROR", fmt.Sprintf("Invalid days value: %d", days))
		return
	}

	matchedDirs := ValidateDirs(paths)

	for _, dir := range matchedDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", path, err))
				return nil
			}
			if !info.IsDir() {
				modTime := info.ModTime()
				cutoff := time.Now().AddDate(0, 0, -days)
				if modTime.Before(cutoff) {
					switch mode {
					case "dry-run":
						logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s", path))
						fmt.Fprintln(tempFile, path)
					case "interactive":
						fmt.Printf("Delete file %s? (y/n): ", path)
						var response string
						fmt.Scanln(&response)
						if response == "y" {
							if err := os.Remove(path); err != nil {
								logging.LogMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
							} else {
								logging.LogMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
							}
						}
					case "scheduled":
						if err := os.Remove(path); err != nil {
							logging.LogMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
						} else {
							logging.LogMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
						}
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
