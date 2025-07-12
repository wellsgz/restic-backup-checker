package main

import (
	"fmt"
	"log"
	"os"

	"restic-backup-checker/internal/cli"
	"restic-backup-checker/internal/config"
	"restic-backup-checker/internal/logger"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// Initialize logger
	logger.Init()

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and execute CLI
	rootCmd := cli.NewRootCommand(cfg, version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 