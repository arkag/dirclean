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
	versionFlag = flag.Bool("version", false, "Print the version of the application")
	updateFlag  = flag.Bool("update", false, "Update the application to the specified version")
	tagFlag     = flag.String("tag", "latest", "Specify the version tag to update to (default: latest)")
	modeFlag    = flag.String("mode", "dry-run", "Specify the mode: dry-run, interactive, scheduled")
	configFlag  = flag.String("config", "config.yaml", "Path to config file")
	logFlag     = flag.String("log", "dirclean.log", "Path to log file")
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
		logging.LogMessage("INFO", "Update successful. Restarting...")
		update.RestartBinary()
		return
	}

	logging.InitLogging(*logFlag)
	configs := config.LoadConfig(*configFlag)

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

	for _, config := range configs {
		modes.ProcessFiles(config, tempFile, *modeFlag)
	}

	dfAfter, err := fileutils.GetDF("/")
	if err != nil {
		logging.LogMessage("ERROR", fmt.Sprintf("Error getting disk usage after: %v", err))
	}

	fileutils.PrintSummary(tempFile.Name(), dfBefore, dfAfter, logging.GenerateUUID())
}
