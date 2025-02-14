package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arkag/dirclean/config"
	"github.com/arkag/dirclean/fileutils"
	"github.com/arkag/dirclean/logging"
	"github.com/arkag/dirclean/modes"
	"github.com/arkag/dirclean/update"
)

var (
	versionFlag  = flag.Bool("version", false, "Print the version of the application")
	updateFlag   = flag.Bool("update", false, "Update the application to the specified version")
	tagFlag      = flag.String("tag", "latest", "Specify the version tag to update to (default: latest)")
	modeFlag     = flag.String("mode", "", "Specify the mode: analyze, dry-run, interactive, scheduled")
	configFlag   = flag.String("config", "config.yaml", "Path to config file")
	logFlag      = flag.String("log", "", "Path to log file")
	logLevelFlag = flag.String("log-level", "", "Log level (DEBUG, INFO, WARN, ERROR, FATAL)")
	minSizeFlag  = flag.String("min-size", "", "Minimum file size (e.g., 100MB)")
	maxSizeFlag  = flag.String("max-size", "", "Maximum file size (e.g., 1GB)")
)

func parseFileSize(sizeStr string) (*config.FileSize, error) {
	if sizeStr == "" {
		return nil, nil
	}

	size := &config.FileSize{}
	var value float64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit)
	if err != nil {
		return nil, fmt.Errorf("invalid file size format: %s", sizeStr)
	}
	size.Value = value
	size.Unit = unit
	return size, nil
}

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("dirclean version: %s\n", update.AppVersion)
		fmt.Printf("dirclean osarch: %s\n", update.AppOsArch)
		return
	}

	if *updateFlag {
		if err := update.UpdateBinary(*tagFlag); err != nil {
			logging.LogMessage("ERROR", fmt.Sprintf("Error updating binary: %v", err))
			return
		}
		logging.LogMessage("INFO", "Update successful. Restarting...")
		update.RestartBinary()
		return
	}

	// Parse file sizes from flags
	minSize, err := parseFileSize(*minSizeFlag)
	if err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error parsing min file size: %v", err))
		os.Exit(1)
	}

	maxSize, err := parseFileSize(*maxSizeFlag)
	if err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error parsing max file size: %v", err))
		os.Exit(1)
	}

	// Load config and merge with CLI flags
	globalConfig := config.LoadConfig(*configFlag)
	cliFlags := config.CLIFlags{
		Mode:        *modeFlag,
		LogFile:     *logFlag,
		LogLevel:    *logLevelFlag,
		MinFileSize: minSize,
		MaxFileSize: maxSize,
	}

	// Initialize logging with merged config
	logging.InitLogging(globalConfig.Defaults.LogFile)
	if cliFlags.LogFile != "" {
		logging.InitLogging(cliFlags.LogFile)
	}

	dfBefore, err := fileutils.GetDF("/")
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting disk usage before: %v", err))
	}

	tempFile, err := os.CreateTemp("", "cleanup_")
	if err != nil {
		logging.LogMessage("FATAL", fmt.Sprintf("Error creating temp file: %v", err))
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Process each rule with merged config
	for _, rule := range globalConfig.Rules {
		// Merge defaults with rule
		mergedConfig := config.MergeWithFlags(rule, cliFlags)
		modes.ProcessFiles(mergedConfig, tempFile)
	}

	dfAfter, err := fileutils.GetDF("/")
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting disk usage after: %v", err))
	}

	fileutils.PrintSummary(tempFile.Name(), dfBefore, dfAfter, logging.GenerateUUID())
}
