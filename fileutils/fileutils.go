package fileutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/arkag/dirclean/logging"
	"github.com/arkag/dirclean/utils"
)

type DirInfo struct {
	Path      string
	Size      int64
	LastUsed  time.Time
	FileCount int
}

// GetDF returns disk usage information for the given path
func GetDF(path string) (map[string]uint64, error) {
	diskUsage := make(map[string]uint64)

	// Get filesystem statistics
	var available, total uint64
	var err error

	available, total, err = getDiskSpace(path)
	if err != nil {
		return nil, err
	}

	diskUsage["Available"] = available
	diskUsage["Total"] = total

	return diskUsage, nil
}

func PrintSummary(tempFile string, dfBefore, dfAfter map[string]uint64, runID string) {
	fileCount := CountLines(tempFile)
	fileSize := GetTotalSize(tempFile)

	fmt.Println("\n-------------------------------------------------------------------------------")
	fmt.Println("SUMMARY")
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Printf("Script:\t\t\t%s/%s\n", filepath.Dir(os.Args[0]), filepath.Base(os.Args[0]))
	fmt.Printf("Run ID:\t\t\t%s\n", runID)
	fmt.Printf("Time:\t\t\t%s\n", time.Now().Format("2006-01-02 15:04"))
	fmt.Printf("Total files processed:\t%d\n", fileCount)
	fmt.Printf("Total size of files:\t%.2f GB\n", fileSize)

	fmt.Println(GetDFDiff(dfBefore, dfAfter))
	fmt.Println("-------------------------------------------------------------------------------")
}

func CountLines(filename string) int {
	f, err := os.Open(filename)
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error opening file: %v", err))
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error reading file: %v", err))
	}
	return lineCount
}

func GetTotalSize(filename string) float64 {
	var totalSize int64
	file, err := os.Open(filename)
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error opening file: %v", err))
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		filePath := scanner.Text()

		// Use the existing utils.GetAbsPath function
		absPath := utils.GetAbsPath(filePath)

		// Verify file exists before attempting to stat
		if !utils.FileExists(absPath) {
			logging.LogMessage("ERROR", fmt.Sprintf("File does not exist at path: %s", absPath))
			continue
		}

		logging.LogMessage("DEBUG", fmt.Sprintf("Processing file: %s\nAbsolute path: %s", filePath, absPath))

		info, err := os.Stat(absPath)
		if err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error getting file info for %s\nAbsolute path: %s\nError: %v",
				filePath, absPath, err))
			continue
		}
		totalSize += info.Size()
	}

	if err := scanner.Err(); err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error scanning file: %v", err))
	}

	return float64(totalSize) / (1024 * 1024 * 1024)
}

func GetDFDiff(before, after map[string]uint64) string {
	if before["Available"] == after["Available"] {
		return "No changes to file system"
	}

	return fmt.Sprintf(`
File system differences before and after old files were removed:
Available space before: %d GB
Available space after: %d GB
`, before["Available"], after["Available"])
}

// GetLargestDirs returns a sorted list of directories consuming the most space
func GetLargestDirs(rootPaths []string, minSize int64) ([]DirInfo, error) {
	var dirs []DirInfo
	seen := make(map[string]bool)

	for _, root := range rootPaths {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.LogMessage("ERROR", fmt.Sprintf("Error accessing path %s: %v", path, err))
				return filepath.SkipDir
			}

			if info.IsDir() {
				// Skip if we've already processed this directory
				if seen[path] {
					return filepath.SkipDir
				}
				seen[path] = true

				dirInfo, err := analyzeDirUsage(path)
				if err != nil {
					logging.LogMessage("ERROR", fmt.Sprintf("Error analyzing directory %s: %v", path, err))
					return nil
				}

				if dirInfo.Size >= minSize {
					dirs = append(dirs, dirInfo)
				}
			}
			return nil
		})
		if err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error walking path %s: %v", root, err))
		}
	}

	// Sort directories by size in descending order
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Size > dirs[j].Size
	})

	return dirs, nil
}

// analyzeDirUsage calculates directory size and last access time
func analyzeDirUsage(dirPath string) (DirInfo, error) {
	var totalSize int64
	var lastUsed time.Time
	var fileCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
			if info.ModTime().After(lastUsed) {
				lastUsed = info.ModTime()
			}
		}
		return nil
	})

	return DirInfo{
		Path:      dirPath,
		Size:      totalSize,
		LastUsed:  lastUsed,
		FileCount: fileCount,
	}, err
}

// FormatSize converts bytes to human-readable format
func FormatSize(bytes int64) string {
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

// GetSuggestedDirs returns directories that are good candidates for cleanup
func GetSuggestedDirs(rootPaths []string, minSizeMB int64) []DirInfo {
	minSizeBytes := minSizeMB * 1024 * 1024
	dirs, err := GetLargestDirs(rootPaths, minSizeBytes)
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting largest directories: %v", err))
		return nil
	}

	// Filter and sort directories based on size and last use
	var suggestions []DirInfo
	for _, dir := range dirs {
		// Consider directories that haven't been used in the last 30 days
		if time.Since(dir.LastUsed) > 30*24*time.Hour {
			suggestions = append(suggestions, dir)
		}
	}

	// Limit to top 10 suggestions
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions
}
