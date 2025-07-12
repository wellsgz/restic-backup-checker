package logger

import (
	"log"
	"os"
)

// Init initializes the logger
func Init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[restic-backup-checker] ")
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	log.Printf("INFO: "+format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	log.Printf("ERROR: "+format, args...)
}

// Fatal logs a fatal error and exits
func Fatal(format string, args ...interface{}) {
	log.Printf("FATAL: "+format, args...)
	os.Exit(1)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	log.Printf("DEBUG: "+format, args...)
} 