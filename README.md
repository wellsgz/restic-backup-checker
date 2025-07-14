# Restic Backup Checker

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/wellsgz/restic-backup-checker/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-brightgreen.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A cross-platform restic backup monitoring tool that checks OneDrive for daily restic backup snapshots and sends notifications via Telegram.

## Version

**Current Version**: 1.0.0

To check the version of your installation:

```bash
./restic-backup-checker version
```

## Features

- **OneDrive Integration**: Monitors OneDrive folders for backup files
- **Automated Authentication**: Handles OAuth2 authentication with token refresh
- **Telegram Notifications**: Sends alerts and daily reports via Telegram
- **Encrypted Configuration**: Stores sensitive data securely with AES-GCM encryption
- **Cross-Platform**: Built with Go for Linux, macOS, and Windows compatibility
- **Flexible Monitoring**: Configurable check intervals and folder monitoring
- **CLI Interface**: Interactive setup and management commands

## Architecture

The application consists of several key components:

- **CLI Interface**: Interactive command-line interface for setup and management
- **OneDrive Client**: Handles authentication and API operations
- **Telegram Client**: Sends notifications and reports
- **Monitor Service**: Performs periodic backup checks
- **Config Manager**: Handles encrypted configuration storage

## Prerequisites

### Telegram Setup

1. **Create Bot**:
   - Message [@BotFather](https://t.me/BotFather) on Telegram
   - Send `/newbot` and follow instructions
   - Note down the **Bot Token**

2. **Get Chat ID**:
   - Send a message to your bot
   - Visit `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Find your **Chat ID** in the response

## Installation

### Option 1: Pre-built Binaries

Download the latest release from the [Releases](https://github.com/wellsgz/restic-backup-checker/releases) page.

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/wellsgz/restic-backup-checker.git
cd restic-backup-checker

# Build the application
go build -o restic-backup-checker ./cmd/main.go

# Or build for specific platforms
# Linux
GOOS=linux GOARCH=amd64 go build -o restic-backup-checker-linux ./cmd/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o restic-backup-checker-macos ./cmd/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o restic-backup-checker-windows.exe ./cmd/main.go
```

## Quick Start

### 1. OneDrive Authentication

First, authenticate with OneDrive using the simple device code flow:

```bash
./restic-backup-checker login
```

This will:
- Display a URL and code to visit
- Open your browser to complete authentication
- Store authentication tokens securely

### 2. Setup Configuration

Run the setup command to configure monitoring and Telegram:

```bash
./restic-backup-checker setup
```

The setup wizard will guide you through:
- Folder selection for monitoring
- Telegram bot configuration
- Monitoring interval configuration

### 3. Manual Check

Test the configuration with a manual backup check:

```bash
./restic-backup-checker check
```

### 4. Start Monitoring

Start the continuous monitoring service:

```bash
./restic-backup-checker
```

The service will:
- Check backups at configured intervals
- Send Telegram notifications for failed backups
- Send daily summary reports

## Usage

### Commands

```bash
# Show version information
./restic-backup-checker version

# Authenticate with OneDrive
./restic-backup-checker login

# Clear OneDrive authentication
./restic-backup-checker logout

# Setup monitoring and Telegram
./restic-backup-checker setup

# Manual backup check
./restic-backup-checker check

# Start monitoring service
./restic-backup-checker

# View current configuration
./restic-backup-checker config show

# Reset configuration
./restic-backup-checker config reset
```

### Configuration Options

The application stores configuration in `~/.config/restic-backup-checker/config.enc` (encrypted).

Configuration includes:
- OneDrive authentication tokens
- Telegram bot credentials
- Monitored folder paths
- Check interval (in minutes)

### Folder Structure

The application expects the following OneDrive folder structure:

```
OneDrive/
├── BackupFolder1/
│   ├── Client1/
│   │   └── snapshots/
│   │       ├── backup-2024-01-01.zip
│   │       └── backup-2024-01-02.zip
│   └── Client2/
│       └── snapshots/
│           └── backup-2024-01-02.zip
└── BackupFolder2/
    └── Client3/
        └── snapshots/
            └── backup-2024-01-02.zip
```

**Key Points**:
- Top-level folders are selected during setup
- Each client has its own subfolder
- Backup files must be in a `snapshots` subfolder
- Files created within the last 24 hours indicate successful backups

## Monitoring Logic

The application checks for backup files created within the last 24 hours within each client's `snapshots` folder. If no files are found within this timeframe, it triggers an alert.

### Backup Validation

1. **24-Hour Check**: Looks for files created within the last 24 hours in each `snapshots` folder
2. **Client Status**: Each client folder is checked independently
3. **Notifications**: Alerts sent for failed backups, summary reports for all clients

### Notification Types

1. **Backup Alerts**: Sent immediately when a backup is missing
2. **Success Notifications**: Sent when backups are found (optional)
3. **Daily Summary**: Overall status report with success/failure counts

## Examples

### Login Example

```bash
$ ./restic-backup-checker login

🔐 OneDrive Authentication Required
Please visit: https://microsoft.com/devicelogin
Enter this code: ABC123

Waiting for authorization...........
✅ Successfully authenticated!

Successfully logged in to OneDrive!
```

### Setup Example

```bash
$ ./restic-backup-checker setup
=== OneDrive Setup ===

Available top-level folders:
1. Backups
2. Documents
3. Pictures

Enter folder numbers to monitor (comma-separated): 1

=== Telegram Setup ===
Create a bot with @BotFather on Telegram and get the bot token.

Enter Telegram Bot Token: 123456789:ABCdefGHIjklMNOpqrsTUVwxyz
Enter Telegram Chat ID: 123456789

✓ Telegram test message sent successfully!

=== Monitoring Setup ===
Enter check interval in minutes (default: 60): 30

Setup completed successfully!
```

### Sample Telegram Notifications

**Backup Alert:**
```
🚨 Backup Alert

Client: DatabaseServer
Folder: /drive/items/ABC123
Issue: No backup found for today
Last Backup: 2024-01-01 14:30:00

Please check the backup client immediately.
```

**Daily Summary:**
```
📊 Daily Backup Report

Status: 🚨 Issues Found
Total Clients: 5
Successful: 4
Failed: 1

Failed Clients:
• DatabaseServer
```

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Ensure you have internet connectivity
   - Try running `restic-backup-checker logout` then `restic-backup-checker login`
   - Check that you're entering the correct device code

2. **Token Expired**
   - The application automatically refreshes tokens
   - If refresh fails, run `restic-backup-checker login` again

3. **Telegram Not Working**
   - Verify bot token and chat ID
   - Send a message to the bot first
   - Check bot has permission to send messages

4. **No Backups Detected**
   - Verify folder structure (`snapshots` subfolder required)
   - Check file creation dates (must be today in UTC)
   - Ensure OneDrive sync is complete

### Debug Mode

Run with verbose logging:

```bash
./restic-backup-checker check --verbose
```

### Configuration Issues

Reset configuration and start over:

```bash
./restic-backup-checker config reset
./restic-backup-checker login
./restic-backup-checker setup
```

## Development

### Building

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o restic-backup-checker ./cmd/main.go

# Cross-compile for different platforms
make build-all
```

### Project Structure

```
restic-backup-checker/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── cli/                 # Command-line interface
│   │   └── cli.go
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── logger/              # Logging utilities
│   │   └── logger.go
│   ├── monitor/             # Backup monitoring service
│   │   └── monitor.go
│   ├── onedrive/            # OneDrive API client
│   │   ├── auth.go
│   │   └── client.go
│   └── telegram/            # Telegram notifications
│       └── telegram.go
├── go.mod                   # Go module dependencies
├── go.sum                   # Dependency checksums
└── README.md               # This file
```

### Adding Features

The application is designed with modularity in mind:

1. **New Storage Backends**: Implement new clients in separate packages
2. **Additional Notifications**: Add new notification providers
3. **Enhanced Monitoring**: Extend the monitor package
4. **Custom Backup Logic**: Modify validation rules in the monitor

## Security

### Data Protection

- **Encrypted Storage**: All sensitive data is encrypted using AES-GCM
- **Key Derivation**: Encryption keys are derived from machine-specific data
- **Token Management**: OAuth tokens are securely stored and auto-refreshed
- **No Plain Text**: Credentials are never stored in plain text

### Best Practices

1. **File Permissions**: Configuration files are created with 0600 permissions
2. **Token Rotation**: Refresh tokens are used for automatic renewal
3. **Input Validation**: All user inputs are validated and sanitized
4. **Error Handling**: Sensitive data is not included in error messages

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Changelog

### Version 1.0.0 (2025-07-12)

**Initial Release**

#### Features
- ✅ OneDrive Integration with OAuth2 authentication
- ✅ Automated token refresh handling
- ✅ Telegram bot notifications and alerts
- ✅ Encrypted configuration storage using AES-GCM
- ✅ Cross-platform support (Linux, macOS, Windows)
- ✅ Flexible monitoring intervals and folder selection
- ✅ Interactive CLI setup wizard
- ✅ Manual backup verification
- ✅ Daily backup status reports
- ✅ Comprehensive logging and error handling

#### Commands
- `login` - OneDrive authentication via device code flow
- `logout` - Clear stored authentication tokens
- `setup` - Interactive configuration wizard
- `check` - Manual backup status verification
- `config show` - Display current configuration
- `config reset` - Reset all configuration
- `version` - Show version information

#### Architecture
- **CLI Interface**: Command-line interface with Cobra
- **OneDrive Client**: Microsoft Graph API integration
- **Telegram Client**: Bot API for notifications
- **Monitor Service**: Automated backup checking
- **Config Manager**: Encrypted configuration storage

#### Security
- All sensitive data encrypted at rest
- Machine-specific key derivation
- Secure token management
- No plain text credential storage

#### Compatibility
- Go 1.21+
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

---

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the logs for error details

---

**Note**: This application requires internet connectivity for OneDrive API calls and Telegram notifications. Ensure your firewall allows outbound HTTPS connections. 