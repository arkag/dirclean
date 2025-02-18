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
	modeFlag     = flag.String("mode", "", "Only process paths configured with this mode (analyze, dry-run, interactive, scheduled)")
	configFlag   = flag.String("config", "config.yaml", "Path to config file (default: /etc/dirclean/config.yaml on Linux)")
	logFlag      = flag.String("log", "", "Path to log file")
	logLevelFlag = flag.String("log-level", "", "Log level (DEBUG, INFO, WARN, ERROR, FATAL)")
)

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
		logging.LogMessage("INFO", "Update successful.")
		return
	}

	var globalConfig config.GlobalConfig
	var cliFlags config.CLIFlags

	if *configFlag == "" {
		logging.LogMessage("FATAL", "Config file must be specified")
		os.Exit(1)
	}
	globalConfig = config.LoadConfig(*configFlag)

	// Populate cliFlags with command line overrides
	cliFlags = config.CLIFlags{
		Mode:     *modeFlag,
		LogFile:  *logFlag,
		LogLevel: *logLevelFlag,
	}

	// Initialize logging with merged config
	logging.InitLogging(globalConfig.Defaults.LogFile)
	if cliFlags.LogFile != "" {
		logging.InitLogging(cliFlags.LogFile)
	}

	// Set log level - first from config defaults, then override with CLI flag if present
	logging.SetLogLevel(globalConfig.Defaults.LogLevel)
	if cliFlags.LogLevel != "" {
		logging.SetLogLevel(cliFlags.LogLevel)
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
		// Skip rules that don't match the specified mode
		if *modeFlag != "" && rule.Mode != *modeFlag {
			continue
		}
		modes.ProcessFiles(rule, tempFile)
	}

	dfAfter, err := fileutils.GetDF("/")
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting disk usage after: %v", err))
	}

	// Collect all paths from all rules
	var allPaths []string
	for _, rule := range globalConfig.Rules {
		allPaths = append(allPaths, rule.Paths...)
	}

	// Remove duplicates from allPaths
	pathMap := make(map[string]bool)
	var uniquePaths []string
	for _, path := range allPaths {
		if !pathMap[path] {
			pathMap[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	fileutils.PrintSummary(tempFile.Name(), dfBefore, dfAfter, logging.GenerateUUID(), uniquePaths)
}
