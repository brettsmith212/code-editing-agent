package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile *os.File
	logger  *log.Logger
	mu      sync.Mutex
)

// Initialize sets up the logger with the specified log file path
func Initialize(logDir string) error {
	mu.Lock()
	defer mu.Unlock()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Create a new log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("chat_%s.log", timestamp))
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	logFile = file
	logger = log.New(file, "", log.LstdFlags)
	return nil
}

// LogMessage logs a chat message with its source (User/AI)
func LogMessage(source, content string) {
	mu.Lock()
	defer mu.Unlock()

	if logger != nil {
		logger.Printf("[%s] %s", source, content)
	}
}

// Close closes the log file
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		return logFile.Close()
	}
	return nil
}
