package fileutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arkag/dirclean/logging"
)

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
		info, err := os.Stat(filePath)
		if err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error getting file info for %s: %v", filePath, err))
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
