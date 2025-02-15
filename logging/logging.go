package logging

import (
	"fmt"
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

func InitLogging(logFilePath string) error {
	logFile = logFilePath
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	log.SetOutput(f)
	return nil
}

func SetLogLevel(level string) {
	logLevel = level
}

func LogMessage(level, message string) {
	if shouldLog(level) {
		log.Printf("[%s] %s", level, message)
	}
}

func GenerateUUID() string {
	return uuid.New().String()
}

// GetLogLevel returns the current log level as a string
func GetLogLevel() string {
	return logLevel
}

// shouldLog determines if a message should be logged based on the configured log level
func shouldLog(messageLevel string) bool {
	messageLogLevel := parseLogLevel(messageLevel)
	configuredLogLevel := parseLogLevel(logLevel)
	return messageLogLevel >= configuredLogLevel
}
