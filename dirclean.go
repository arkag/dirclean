// This program was generated with the help of Gemini Advanced 2.0 Flash, a large
// language model from Google AI, trained on a massive dataset of text and code.
// Gemini Advanced 2.0 Flash can generate different creative text formats of text
// content, translate languages, write different kinds of code, and answer your
// questions in an informative way.
//
// The following Gems were used in the generation of this program:
// - Code generation
// - Code explanation
// - Code optimization
// - Code debugging

package main

import (
        "bufio"
        "crypto/rand"
        "fmt"
        "io"
        "log"
        "os"
        "path/filepath"
        "regexp"
        "strconv"
        "strings"
        "syscall"
        "time"

        "gopkg.in/yaml.v3"
)

// Config struct defines the structure of the YAML configuration.
// It includes the number of days after which files are considered old
// and a list of paths to be processed.
type Config struct {
        DeleteOlderThanDays int      `yaml:"delete_older_than_days"`
        Paths              string `yaml:"paths"`
}

var (
        // Default log file path.
        logFile = "/var/log/dirclean.log"
        // Default log level.
        logLevel = "INFO"
        // Default YAML configuration file path.
        configFile = "/etc/dirclean/dirclean.yaml"
        // Dry run mode flag - if true, no files will be deleted.
        dryRun = true
        // List of directories to be processed.
        matchedDirsstring
        // Log level mapping for easy comparison.
        logLevels = map[string]int{
                "DEBUG": 0,
                "INFO":  1,
                "OK":    2,
                "WARN":  3,
                "ERROR": 4,
                "FATAL": 5,
                "OFF":   6,
        }
        // Log messages for different error scenarios.
        logMessages = map[string]string{
                "ERROR_CONFIG_NOT_FOUND": "Config file not found:",
                "ERROR_INVALID_YAML":     "Invalid YAML format in config file:",
                "ERROR_INVALID_DAYS":     "Invalid days value:",
                "ERROR_DIR_NOT_EXIST":    "Directory does not exist:",
        }
)

func main() {
        // Check for help flag.
        if contains(os.Args, "-h") || contains(os.Args, "--help") {
                help()
                return
        }

        // Override default values with environment variables if set.
        if val:= os.Getenv("DRY_RUN"); val!= "" {
                dryRun = val == "true" || val == "1"
        }
        if val:= os.Getenv("CONFIG_FILE"); val!= "" {
                configFile = val
        }
        if val:= os.Getenv("LOG_FILE"); val!= "" {
                logFile = val
        }
        if val:= os.Getenv("LOG_LEVEL"); val!= "" {
                logLevel = val
        }

        // Open the log file for writing.
        f, err:= os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
        if err!= nil {
                log.Fatalf("Error opening log file: %v", err)
        }
        defer f.Close()
        log.SetOutput(f)

        // Generate a unique run ID.
        runID:= generateUUID()
        log.Printf("BEGIN: %s: RUN_ID=%s", filepath.Base(os.Args), runID)
        defer log.Printf("END: %s: RUN_ID=%s", filepath.Base(os.Args), runID)

        // Perform pre-run checks.
        preRunCheck()

        // Load configuration from YAML file.
        configs:= loadConfig()

        // Get disk usage before processing.
        dfBefore, err:= getDF()
        if err!= nil {
                log.Printf("Error getting disk usage before: %v", err)
        }

        // Create a temporary file to store the list of files to be deleted.
        tempFile, err:= os.CreateTemp("", "cleanup_")
        if err!= nil {
                log.Fatal(err)
        }
        defer os.Remove(tempFile.Name())
        defer tempFile.Close()

        // Process each configuration.
        for _, config:= range configs {
                processFiles(config, tempFile)
        }

        // Get disk usage after processing.
        dfAfter, err:= getDF()
        if err!= nil {
                log.Printf("Error getting disk usage after: %v", err)
        }

        // Print summary of the cleanup operation.
        printSummary(tempFile.Name(), dfBefore, dfAfter, runID)
}

// help function prints the usage instructions and information about the script.
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

// preRunCheck performs checks before the cleanup process starts.
// Currently, it checks if the YAML configuration file exists.
func preRunCheck() {
        if _, err:= os.Stat(configFile); os.IsNotExist(err) {
                log.Fatalf("%s %s", logMessages["ERROR_CONFIG_NOT_FOUND"], configFile)
        }
}

// loadConfig reads and parses the YAML configuration file.
// It returns a slice of Config structs.
func loadConfig()Config {
        f, err:= os.Open(configFile)
        if err!= nil {
                log.Fatalf("Error opening YAML file: %v", err)
        }
        defer f.Close()

        decoder:= yaml.NewDecoder(f)
        var configsConfig
        for {
                var config Config
                if err:= decoder.Decode(&config); err!= nil {
                        if err == io.EOF {
                                break
                        }
                        log.Fatalf("Error decoding YAML: %v", err)
                }
                configs = append(configs, config)
        }
        return configs
}

// processFiles processes each configuration by traversing the specified paths
// and deleting files older than the specified number of days.
func processFiles(config Config, tempFile *os.File) {
        days:= config.DeleteOlderThanDays
        paths:= config.Paths

        if days <= 0 {
                log.Printf("%s %d", logMessages["ERROR_INVALID_DAYS"], days)
                return
        }

        validateDirs(paths)

        for _, dir:= range matchedDirs {
                err:= filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
                        if err!= nil {
                                log.Printf("Error accessing %s: %v", path, err)
                                return nil // Continue walking
                        }
                        if!info.IsDir() {
                                modTime:= info.ModTime()
                                cutoff:= time.Now().AddDate(0, 0, -days)
                                if modTime.Before(cutoff) {
                                        if!dryRun {
                                                err:= os.Remove(path)
                                                if err!= nil {
                                                        log.Printf("Error deleting file %s: %v", path, err)
                                                }
                                        } else {
                                                fmt.Fprintln(tempFile, path) // Write to temp file for summary
                                        }
                                }
                        }
                        return nil
                })
                if err!= nil {
                        log.Printf("Error walking directory %s: %v", dir, err)
                }
        }
}

// validateDirs validates the provided paths and populates the matchedDirs slice.
func validateDirs(dirsstring) {
        matchedDirs =string{}
        for _, dir:= range dirs {
                if strings.Contains(dir, "*") {
                        matches, err:= filepath.Glob(dir)
                        if err!= nil {
                                log.Printf("Error with wildcard path: %v", err)
                                continue
                        }
                        matchedDirs = append(matchedDirs, matches...)
                } else if _, err:= os.Stat(dir); err == nil {
                        matchedDirs = append(matchedDirs, dir)
                }
        }
}

// getDF retrieves disk usage information.
func getDF() (map[string]uint64, error) {
        var stat syscall.Statfs_t
        err:= syscall.Statfs("/", &stat)
        if err!= nil {
                return nil, err
        }

        diskUsage:= make(map[string]uint64)
        diskUsage["Available"] = (uint64(stat.Bavail) * uint64(stat.Bsize)) / (1024 * 1024 * 1024) // Available in GB
        diskUsage["Total"] = (uint64(stat.Blocks) * uint64(stat.Bsize)) / (1024 * 1024 * 1024)     // Total in GB

        return diskUsage, nil
}

// getDFDiff compares disk usage before and after the cleanup operation.
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

// countLines counts the number of lines in a file.
func countLines(filename string) int {
        f, err:= os.Open(filename)
        if err!= nil {
                log.Printf("Error opening file: %v", err)
                return 0
        }
        defer f.Close()

        scanner:= bufio.NewScanner(f)
        lineCount:= 0
        for scanner.Scan() {
                lineCount++
        }
        if err:= scanner.Err(); err!= nil {
                log.Printf("Error reading file: %v", err)
        }
        return lineCount
}

// printSummary prints a summary of the cleanup operation.
func printSummary(tempFile string, dfBefore, dfAfter map[string]uint64, runID string) {
        fileCount:= countLines(tempFile)
        fileSize:= getTotalSize(tempFile)

        fmt.Println("\n-------------------------------------------------------------------------------")
        fmt.Println("SUMMARY")
        fmt.Println("-------------------------------------------------------------------------------")
        fmt.Printf("Script:\t\t\t%s/%s\n", filepath.Dir(os.Args), filepath.Base(os.Args))
        fmt.Printf("Run ID:\t\t\t%s\n", runID)
        fmt.Printf("Time:\t\t\t%s\n", time.Now().Format("2006-01-02 15:04"))
        fmt.Printf("Total files processed:\t%d\n", fileCount)
        fmt.Printf("Total size of files:\t%.2f GB\n", fileSize)

        fmt.Println(getDFDiff(dfBefore, dfAfter))
        fmt.Println("-------------------------------------------------------------------------------\n")
}

// getTotalSize calculates the total size of files listed in a file.
func getTotalSize(filename string) float64 {
        var totalSize float64
        file, err:= os.Open(filename)
        if err!= nil {
                log.Printf("Error opening file: %v", err)
                return 0
        }
        defer file.Close()

        scanner:= bufio.NewScanner(file)
        re:= regexp.MustCompile(`\s+`) // Match one or more whitespace characters
        for scanner.Scan() {
                fields:= re.Split(scanner.Text(), -1)
                if len(fields) >= 7 {
                        size, err:= strconv.ParseFloat(fields, 64)
                        if err!= nil {
                                log.Printf("Error parsing size: %v", err)
                                continue
                        }
                        totalSize += size
                }
        }

        if err:= scanner.Err(); err!= nil {
                log.Printf("Error scanning file: %v", err)
        }

        return totalSize / 1024 / 1024 / 1024 // Convert to GB
}

// generateUUID generates a UUID string.
func generateUUID() string {
        // Generate a UUID using crypto/rand
        b:= make(byte, 16)
        _, err:= rand.Read(b)
        if err!= nil {
                log.Fatalf("Error generating UUID: %v", err)
        }
        uuid:= fmt.Sprintf("%x-%x-%x-%x-%x",
                b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
        return uuid
}

// contains checks if a string slice contains a specific string.
func contains(sstring, e string) bool {
        for _, a:= range s {
                if a == e {
                        return true
                }
        }
        return false
}