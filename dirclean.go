package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DeleteOlderThanDays int      `yaml:"delete_older_than_days"`
	Paths               []string `yaml:"paths"`
}

var (
	logFile    = "dirclean.log"
	logLevel   = "INFO"
	configFile = "config.yaml"
	dryRun     = true
	logLevels  = map[string]int{
		"DEBUG": 0,
		"INFO":  1,
		"OK":    2,
		"WARN":  3,
		"ERROR": 4,
		"FATAL": 5,
		"OFF":   6,
	}
	logMessages = map[string]string{
		"ERROR_CONFIG_NOT_FOUND": "Config file not found:",
		"ERROR_INVALID_YAML":     "Invalid YAML format in config file:",
		"ERROR_INVALID_DAYS":     "Invalid days value:",
		"ERROR_DIR_NOT_EXIST":    "Directory does not exist:",
	}
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		help()
		return
	}

	if val := os.Getenv("DRY_RUN"); val != "" {
		dryRun = val == "true" || val == "1"
	}
	if val := os.Getenv("CONFIG_FILE"); val != "" {
		configFile = val
	}
	if val := os.Getenv("LOG_FILE"); val != "" {
		logFile = val
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		logLevel = val
	}

	logMessage("DEBUG", fmt.Sprintf("DRY_RUN: %v", dryRun))
	logMessage("DEBUG", fmt.Sprintf("CONFIG_FILE: %s", configFile))
	logMessage("DEBUG", fmt.Sprintf("LOG_FILE: %s", logFile))
	logMessage("DEBUG", fmt.Sprintf("LOG_LEVEL: %s", logLevel))

	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	runID := generateUUID()
	logMessage("INFO", fmt.Sprintf("BEGIN: %s: RUN_ID=%s", filepath.Base(os.Args[0]), runID))
	defer logMessage("INFO", fmt.Sprintf("END: %s: RUN_ID=%s", filepath.Base(os.Args[0]), runID))

	if dryRun {
		logMessage("INFO", "Running in DRY_RUN mode. No files will be deleted.")
	}

	preRunCheck()
	configs := loadConfig()

	dfBefore, err := getDF("/")
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error getting disk usage before: %v", err))
	}

	tempFile, err := os.CreateTemp("", "cleanup_")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	for _, config := range configs {
		processFiles(config, tempFile)
	}

	dfAfter, err := getDF("/")
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error getting disk usage after: %v", err))
	}

	printSummary(tempFile.Name(), dfBefore, dfAfter, runID)
}

func help() {
	fmt.Println(`
dirclean - Clean up old files from directories

The script removes files older than specified days from directories defined in
` + configFile + `

Config file format:
   - delete_older_than_days: 30
     paths:
       - /foo_dir/foo_sub_dir
       - /foo_dir/foo_sub_dir/foo_wildcard_dir*

Usage:
  dirclean [--help|-h]

Environment Variables:
   DRY_RUN        When true, only show what would be deleted (default: true)
   CONFIG_FILE    Path to config file (default: ` + configFile + `)
   LOG_FILE       Path to log file (default: /var/log/dirclean.log)
   LOG_LEVEL      Logging level: DEBUG|INFO|WARN|ERROR|FATAL (default: INFO)
`)
}

func preRunCheck() {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("%s %s", logMessages["ERROR_CONFIG_NOT_FOUND"], configFile)
	}
}

func loadConfig() []Config {
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Error opening YAML file: %v", err)
	}
	defer f.Close()

	var configs []Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&configs); err != nil {
		log.Fatalf("Error decoding YAML: %v", err)
	}

	logMessage("DEBUG", fmt.Sprintf("Loaded configs: %+v", configs))
	return configs
}

func processFiles(config Config, tempFile *os.File) {
	days := config.DeleteOlderThanDays
	paths := config.Paths

	if days <= 0 {
		logMessage("ERROR", fmt.Sprintf("%s %d", logMessages["ERROR_INVALID_DAYS"], days))
		return
	}

	matchedDirs := validateDirs(paths)

	for _, dir := range matchedDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logMessage("ERROR", fmt.Sprintf("Error accessing %s: %v", path, err))
				return nil
			}
			if !info.IsDir() {
				modTime := info.ModTime()
				cutoff := time.Now().AddDate(0, 0, -days)
				if modTime.Before(cutoff) {
					if !dryRun {
						err := os.Remove(path)
						if err != nil {
							logMessage("ERROR", fmt.Sprintf("Error deleting file %s: %v", path, err))
						} else {
							logMessage("INFO", fmt.Sprintf("Deleted file: %s", path))
						}
					} else {
						logMessage("INFO", fmt.Sprintf("Would delete file: %s", path))
						fmt.Fprintln(tempFile, path)
					}
				}
			}
			return nil
		})
		if err != nil {
			logMessage("ERROR", fmt.Sprintf("Error walking directory %s: %v", dir, err))
		}
	}
}

func validateDirs(dirs []string) []string {
	var matchedDirs []string
	for _, dir := range dirs {
		if strings.Contains(dir, "*") {
			matches, err := filepath.Glob(dir)
			if err != nil {
				logMessage("ERROR", fmt.Sprintf("Error with wildcard path: %v", err))
				continue
			}
			for _, match := range matches {
				if info, err := os.Stat(match); err == nil && info.IsDir() {
					logMessage("DEBUG", fmt.Sprintf("Matched directory: %s", match))
					matchedDirs = append(matchedDirs, match)
				}
			}
		} else if info, err := os.Stat(dir); err == nil && info.IsDir() {
			logMessage("DEBUG", fmt.Sprintf("Matched directory: %s", dir))
			matchedDirs = append(matchedDirs, dir)
		} else {
			logMessage("ERROR", fmt.Sprintf("Directory does not exist or is not accessible: %s", dir))
		}
	}
	return matchedDirs
}

func getDF(path string) (map[string]uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	diskUsage := make(map[string]uint64)
	diskUsage["Available"] = (uint64(stat.Bavail) * uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	diskUsage["Total"] = (uint64(stat.Blocks) * uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	return diskUsage, nil
}

func getDFDiff(before, after map[string]uint64) string {
	if before["Available"] == after["Available"] {
		if dryRun {
			return "Script ran in DRY_RUN mode\nNo changes to file system"
		}
		return "No changes to file system"
	}

	return fmt.Sprintf(`
File system differences before and after old files were removed:
Available space before: %d GB
Available space after: %d GB
`, before["Available"], after["Available"])
}

func countLines(filename string) int {
	f, err := os.Open(filename)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error opening file: %v", err))
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		logMessage("ERROR", fmt.Sprintf("Error reading file: %v", err))
	}
	return lineCount
}

func printSummary(tempFile string, dfBefore, dfAfter map[string]uint64, runID string) {
	fileCount := countLines(tempFile)
	fileSize := getTotalSize(tempFile)

	fmt.Println("\n-------------------------------------------------------------------------------")
	fmt.Println("SUMMARY")
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Printf("Script:\t\t\t%s/%s\n", filepath.Dir(os.Args[0]), filepath.Base(os.Args[0]))
	fmt.Printf("Run ID:\t\t\t%s\n", runID)
	fmt.Printf("Time:\t\t\t%s\n", time.Now().Format("2006-01-02 15:04"))
	fmt.Printf("Total files processed:\t%d\n", fileCount)
	fmt.Printf("Total size of files:\t%.2f GB\n", fileSize)

	fmt.Println(getDFDiff(dfBefore, dfAfter))
	fmt.Println("-------------------------------------------------------------------------------\n")
}

func getTotalSize(filename string) float64 {
	var totalSize int64
	file, err := os.Open(filename)
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("Error opening file: %v", err))
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		filePath := scanner.Text()
		info, err := os.Stat(filePath)
		if err != nil {
			logMessage("ERROR", fmt.Sprintf("Error getting file info for %s: %v", filePath, err))
			continue
		}
		totalSize += info.Size()
	}

	if err := scanner.Err(); err != nil {
		logMessage("ERROR", fmt.Sprintf("Error scanning file: %v", err))
	}

	return float64(totalSize) / (1024 * 1024 * 1024)
}

func generateUUID() string {
	return uuid.New().String()
}

func logMessage(level, message string) {
	if logLevels[level] >= logLevels[logLevel] {
		log.Printf("[%s] %s", level, message)
	}
}
