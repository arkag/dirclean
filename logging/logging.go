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

// LogLevel represents logging severity levels
type LogLevel int

const (
	DEBUG LogLevel = iota // 0
	INFO                  // 1
	WARN                  // 2
	ERROR                 // 3
	FATAL                 // 4
)

// Convert string level to LogLevel type
func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

func InitLogging(logFilePath string) {
	logFile = logFilePath
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(f)
}

func SetLogLevel(level string) {
	logLevel = level
}

func LogMessage(level, message string) {
	messageLevel := parseLogLevel(level)
	configuredLevel := parseLogLevel(logLevel)

	// Only log if message severity is equal to or higher than configured severity
	if messageLevel >= configuredLevel {
		log.Printf("[%s] %s", level, message)
	}
}

func GenerateUUID() string {
	return uuid.New().String()
}
