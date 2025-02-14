package modes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arkag/dirclean/config"
	"github.com/arkag/dirclean/fileutils"
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

	if config.Mode == "analyze" {
		fmt.Println("\nAnalyzing directories for old files...")
		fmt.Println("=====================================")
	}

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
					if config.Mode == "analyze" {
						logging.LogMessage("INFO", fmt.Sprintf("Found candidate: %s (size: %s, modified: %s)",
							path, fileutils.FormatSize(fileSize), modTime.Format("2006-01-02")))
					}
					switch config.Mode {
					case "analyze":
						logging.LogMessage("INFO", fmt.Sprintf("Found candidate: %s (size: %s, modified: %s)",
							path, fileutils.FormatSize(fileSize), modTime))
					case "dry-run":
						logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s", path))
						fmt.Fprintln(tempFile, path)
					case "interactive":
						fmt.Printf("Delete file %s? (size: %s, modified: %s) (y/n): ",
							path, fileutils.FormatSize(fileSize), modTime)
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

	if config.Mode == "analyze" {
		suggestions := fileutils.GetSuggestedDirs(paths, 100) // 100MB minimum size
		if len(suggestions) > 0 {
			fmt.Println("\nLarge directories that may need attention:")
			fmt.Println("=========================================")
			for i, dir := range suggestions {
				fmt.Printf("\n%d. Directory: %s\n", i+1, dir.Path)
				fmt.Printf("   Total size: %s\n", fileutils.FormatSize(dir.Size))
				fmt.Printf("   Last accessed: %s\n", dir.LastUsed.Format("2006-01-02"))
				fmt.Printf("   Files: %d\n", dir.FileCount)

				// Calculate and show percentage of old files
				oldFilesCount := 0
				oldFilesSize := int64(0)
				err := filepath.Walk(dir.Path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					if !info.IsDir() {
						if info.ModTime().Before(time.Now().AddDate(0, 0, -days)) {
							oldFilesCount++
							oldFilesSize += info.Size()
						}
					}
					return nil
				})
				if err == nil && dir.FileCount > 0 {
					percentOld := float64(oldFilesCount) / float64(dir.FileCount) * 100
					percentSize := float64(oldFilesSize) / float64(dir.Size) * 100
					fmt.Printf("   Old files: %d (%.1f%% of files, %.1f%% of size)\n",
						oldFilesCount, percentOld, percentSize)
				}
			}

			fmt.Println("\nTo clean these directories:")
			fmt.Println("1. Add them to your config file, or")
			fmt.Printf("2. Run: dirclean --mode=interactive --path=<directory_path> --days=%d\n", days)
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
