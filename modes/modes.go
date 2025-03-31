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
	days := config.OlderThanDays
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
		fmt.Println("\nAnalyzing directories for old files and broken symlinks...")
		fmt.Println("===================================================")
	}

	for _, dir := range matchedDirs {
		// Handle wildcard paths
		if strings.Contains(dir, "*") {
			basePath := dir[:strings.Index(dir, "*")]
			pattern := dir[strings.Index(dir, "*"):]

			// Walk the base path
			err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					logging.LogMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", path, err))
					return filepath.SkipDir
				}

				// Check if the path matches the pattern
				matched, err := filepath.Match(pattern, path[len(basePath):])
				if err != nil {
					logging.LogMessage("ERROR", fmt.Sprintf("Error matching pattern: %v", err))
					return nil
				}

				if matched {
					// Process the matched file/directory
					processPath(path, info, config, tempFile, days, minBytes, maxBytes)
				}
				return nil
			})

			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error walking directory %s: %v", basePath, err))
			}
		} else {
			// Handle non-wildcard paths as before
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					logging.LogMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", path, err))
					return nil
				}
				return processPath(path, info, config, tempFile, days, minBytes, maxBytes)
			})
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error walking directory %s: %v", dir, err))
			}
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
		// If the path contains wildcards, add it directly to matched dirs
		if strings.Contains(dir, "*") {
			// Extract the base path (everything before the first wildcard)
			basePath := dir[:strings.Index(dir, "*")]
			if basePath == "" {
				logging.LogMessage("ERROR", fmt.Sprintf("Invalid wildcard path: %s", dir))
				continue
			}

			// Verify the base path exists
			if info, err := os.Stat(basePath); err == nil && info.IsDir() {
				logging.LogMessage("DEBUG", fmt.Sprintf("Added wildcard path: %s", dir))
				matchedDirs = append(matchedDirs, dir)
			} else {
				logging.LogMessage("ERROR", fmt.Sprintf("Base directory does not exist or is not accessible: %s", basePath))
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

func handleBrokenSymlink(mode string, path string, tempFile *os.File) {
	switch mode {
	case "analyze":
		logging.LogMessage("INFO", fmt.Sprintf("Found broken symlink: %s", path))
	case "dry-run":
		logging.LogMessage("INFO", fmt.Sprintf("Would delete broken symlink: %s", path))
		fmt.Fprintln(tempFile, path)
	case "interactive":
		fmt.Printf("Delete broken symlink %s? (y/n): ", path)
		var response string
		fmt.Scanln(&response)
		if response == "y" || response == "Y" {
			deleteFile(path, tempFile)
		}
	case "scheduled":
		deleteFile(path, tempFile)
	default:
		logging.LogMessage("WARN", fmt.Sprintf("Unknown mode: %s, defaulting to dry-run", mode))
		logging.LogMessage("INFO", fmt.Sprintf("Would delete broken symlink: %s", path))
		fmt.Fprintln(tempFile, path)
	}
}

func handleOldFile(mode string, path string, fileSize int64, modTime time.Time, tempFile *os.File) {
	switch mode {
	case "analyze":
		logging.LogMessage("INFO", fmt.Sprintf("Found candidate: %s (size: %s, modified: %s)",
			path, fileutils.FormatSize(fileSize), modTime.Format("2006-01-02")))
	case "dry-run":
		logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s (size: %s, modified: %s)",
			path, fileutils.FormatSize(fileSize), modTime.Format("2006-01-02")))
		fmt.Fprintln(tempFile, path)
	case "interactive":
		// Clear line and print file info
		fmt.Print("\033[2K\r") // Clear current line
		fmt.Printf("\n%s\n", strings.Repeat("-", 80))
		fmt.Printf("File: %s\n", path)
		fmt.Printf("Size: %s\n", fileutils.FormatSize(fileSize))
		fmt.Printf("Modified: %s (%s ago)\n",
			modTime.Format("2006-01-02 15:04:05"),
			formatTimeAgo(time.Since(modTime)))

		// Add file type info if possible
		if ext := filepath.Ext(path); ext != "" {
			fmt.Printf("Type: %s file\n", strings.TrimPrefix(ext, "."))
		}

		fmt.Printf("%s\n", strings.Repeat("-", 80))
		fmt.Print("Actions: [d]elete, [s]kip, [q]uit: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "d":
			deleteFile(path, tempFile)
			fmt.Printf("✓ Deleted: %s\n", path)
		case "q":
			fmt.Println("\nExiting interactive mode...")
			os.Exit(0)
		case "s":
			fmt.Printf("→ Skipped: %s\n", path)
		default:
			fmt.Printf("→ Skipped: %s\n", path)
		}
	case "scheduled":
		logging.LogMessage("INFO", fmt.Sprintf("Deleting file: %s (size: %s, modified: %s)",
			path, fileutils.FormatSize(fileSize), modTime.Format("2006-01-02")))
		deleteFile(path, tempFile)
	default:
		logging.LogMessage("WARN", fmt.Sprintf("Unknown mode: %s, defaulting to dry-run", mode))
		logging.LogMessage("INFO", fmt.Sprintf("Would delete file: %s (size: %s, modified: %s)",
			path, fileutils.FormatSize(fileSize), modTime.Format("2006-01-02")))
		fmt.Fprintln(tempFile, path)
	}
}

func deleteFile(path string, tempFile *os.File) {
	if err := os.Remove(path); err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
	} else {
		logging.LogMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
		// Write to temp file for summary
		if _, err := fmt.Fprintln(tempFile, path); err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error writing to temp file: %v", err))
		}
	}
}

// formatTimeAgo returns a human-readable string representing how long ago a time was
func formatTimeAgo(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	months := days / 30
	years := months / 12

	switch {
	case years > 0:
		return fmt.Sprintf("%d years", years)
	case months > 0:
		return fmt.Sprintf("%d months", months)
	case days > 0:
		return fmt.Sprintf("%d days", days)
	default:
		hours := int(duration.Hours())
		if hours > 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// Helper function to process a single path
func processPath(path string, info os.FileInfo, config config.Config, tempFile *os.File, days int, minBytes, maxBytes int64) error {
	// Special handling for "**" pattern
	if strings.Contains(path, "**") {
		baseDir := path[:strings.Index(path, "**")]
		return filepath.Walk(baseDir, func(subPath string, subInfo os.FileInfo, err error) error {
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", subPath, err))
				return filepath.SkipDir
			}
			// Process each file in the directory tree
			if !subInfo.IsDir() {
				return processFile(subPath, subInfo, config, tempFile, days, minBytes, maxBytes)
			}
			return nil
		})
	}

	if info.IsDir() {
		return nil
	}

	return processFile(path, info, config, tempFile, days, minBytes, maxBytes)
}

// New helper function to handle individual file processing
func processFile(path string, info os.FileInfo, config config.Config, tempFile *os.File, days int, minBytes, maxBytes int64) error {
	// Check for broken symlinks first if enabled
	if config.CleanBrokenSymlinks {
		linkInfo, err := os.Lstat(path)
		if err == nil && linkInfo.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err == nil {
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(path), target)
				}
				if _, err := os.Stat(target); os.IsNotExist(err) {
					handleBrokenSymlink(config.Mode, path, tempFile)
					return nil
				}
			}
		}
	}

	// Process regular files
	fileSize := info.Size()
	if (minBytes > 0 && fileSize < minBytes) ||
		(maxBytes > 0 && fileSize > maxBytes) {
		return nil
	}

	modTime := info.ModTime()
	cutoff := time.Now().AddDate(0, 0, -days)
	if modTime.Before(cutoff) {
		handleOldFile(config.Mode, path, fileSize, modTime, tempFile)
	}
	return nil
}
