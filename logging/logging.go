package logging

import (
	"log"
	"os"

	"github.com/google/uuid"
)

var (
	logFile  = "dirclean.log"
	logLevel = "INFO"
)

func InitLogging(logFilePath string) {
	logFile = logFilePath
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(f)
}

func LogMessage(level, message string) {
	log.Printf("[%s] %s", level, message)
}

func GenerateUUID() string {
	return uuid.New().String()
}
