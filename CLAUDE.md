# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application that monitors OneDrive folders for daily restic backup snapshots and sends notifications via Telegram. It's a security/backup monitoring tool with encrypted configuration storage.

## Development Commands

### Building
- `make build` - Build for current platform
- `make build-all` - Build for all platforms (Linux, macOS, Windows)
- `go build -o build/restic-backup-checker ./cmd/main.go` - Direct build

### Testing
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `go test ./...` - Direct test command

### Code Quality
- `make fmt` - Format code with go fmt
- `make vet` - Run go vet
- `make lint` - Run golangci-lint (requires golangci-lint installed)
- `make dev` - Run fmt, vet, test, and build in sequence

### Running
- `make run` - Build and run the main application
- `make setup` - Build and run setup wizard
- `make check` - Build and run manual backup check

## Architecture

### Core Components
- **CLI Interface** (`internal/cli/`): Cobra-based command interface with subcommands
- **Configuration** (`internal/config/`): AES-GCM encrypted config management
- **OneDrive Client** (`internal/onedrive/`): OAuth2 authentication and Microsoft Graph API integration
- **Telegram Client** (`internal/telegram/`): Bot API for notifications
- **Monitor Service** (`internal/monitor/`): Backup checking logic and scheduling
- **Logger** (`internal/logger/`): Centralized logging

### Key Design Patterns
- Configuration is encrypted at rest using AES-GCM with machine-specific key derivation
- OAuth2 tokens are automatically refreshed via refresh tokens
- CLI uses Cobra framework with subcommands (login, logout, setup, check, config, version)
- Main entry point is `cmd/main.go` which initializes logger, loads config, and executes CLI

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/oauth2` - OAuth2 authentication
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram bot API
- `golang.org/x/crypto` - Encryption utilities

## Configuration

- Config stored encrypted in `~/.config/restic-backup-checker/config.enc`
- Contains OneDrive tokens, Telegram credentials, and monitoring settings
- Use `./restic-backup-checker config show` to view current config
- Use `./restic-backup-checker config reset` to reset configuration

## Testing Strategy

Run tests before any major changes:
```bash
make test
```

For development workflow including all checks:
```bash
make dev
```

## Go Version

Requires Go 1.21+