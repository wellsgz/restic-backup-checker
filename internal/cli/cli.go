package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"restic-backup-checker/internal/config"
	"restic-backup-checker/internal/logger"
	"restic-backup-checker/internal/monitor"
	"restic-backup-checker/internal/onedrive"
	"restic-backup-checker/internal/telegram"

	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command
func NewRootCommand(cfg *config.Config, version string) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "restic-backup-checker",
		Short: "A tool to check restic backup status on OneDrive",
		Long:  `Restic Backup Checker monitors OneDrive folders for daily restic backup snapshots and sends notifications via Telegram.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !cfg.IsConfigured() {
				logger.Info("Configuration not found. Please run 'restic-backup-checker setup' first.")
				return
			}

			// Start monitoring
			monitor := monitor.New(cfg)
			if err := monitor.Start(); err != nil {
				logger.Error("Failed to start monitoring: %v", err)
				return
			}
		},
	}

	// Add subcommands
	rootCmd.AddCommand(newLoginCommand(cfg))
	rootCmd.AddCommand(newLogoutCommand(cfg))
	rootCmd.AddCommand(newSetupCommand(cfg))
	rootCmd.AddCommand(newCheckCommand(cfg))
	rootCmd.AddCommand(newConfigCommand(cfg))
	rootCmd.AddCommand(newVersionCommand(version))

	return rootCmd
}

// newVersionCommand creates the version command
func newVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the current version of restic-backup-checker.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("restic-backup-checker version %s\n", version)
		},
	}
}

// newSetupCommand creates the setup command
func newSetupCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Set up folder monitoring and Telegram notifications",
		Long:  `Interactive setup for folder monitoring and Telegram notifications. Run 'restic-backup-checker login' first to authenticate with OneDrive.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := setupOneDrive(cfg); err != nil {
				logger.Error("Failed to setup OneDrive: %v", err)
				return
			}

			if err := setupTelegram(cfg); err != nil {
				logger.Error("Failed to setup Telegram: %v", err)
				return
			}

			if err := setupMonitoring(cfg); err != nil {
				logger.Error("Failed to setup monitoring: %v", err)
				return
			}

			if err := cfg.Save(); err != nil {
				logger.Error("Failed to save configuration: %v", err)
				return
			}

			logger.Info("Setup completed successfully!")
		},
	}
}

// newCheckCommand creates the check command for manual backup verification
func newCheckCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Manually check backup status",
		Long:  `Manually check if backups are up to date and send notifications if needed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !cfg.IsConfigured() {
				logger.Error("Configuration not found. Please run 'restic-backup-checker setup' first.")
				return
			}

			monitor := monitor.New(cfg)
			if err := monitor.CheckOnce(); err != nil {
				logger.Error("Failed to check backups: %v", err)
				return
			}

			logger.Info("Backup check completed.")
		},
	}
}

// newConfigCommand creates the config command
func newConfigCommand(cfg *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View and modify application configuration.`,
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			showConfig(cfg)
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if confirmReset() {
				*cfg = config.Config{}
				if err := cfg.Save(); err != nil {
					logger.Error("Failed to reset configuration: %v", err)
					return
				}
				logger.Info("Configuration reset successfully.")
			}
		},
	})

	return configCmd
}

// newLoginCommand creates the login command
func newLoginCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with OneDrive",
		Long:  `Authenticate with OneDrive using device code flow.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := loginToOneDrive(cfg); err != nil {
				logger.Error("Failed to login to OneDrive: %v", err)
				return
			}
			logger.Info("Successfully logged in to OneDrive!")
		},
	}
}

// newLogoutCommand creates the logout command
func newLogoutCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear OneDrive authentication",
		Long:  `Clear stored OneDrive authentication tokens.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := logoutFromOneDrive(cfg); err != nil {
				logger.Error("Failed to logout from OneDrive: %v", err)
				return
			}
			logger.Info("Successfully logged out from OneDrive!")
		},
	}
}

// loginToOneDrive performs device code flow authentication
func loginToOneDrive(cfg *config.Config) error {
	auth := onedrive.NewAuthenticator()
	token, err := auth.Authenticate()
	if err != nil {
		return fmt.Errorf("failed to authenticate with OneDrive: %w", err)
	}

	cfg.OneDrive.AccessToken = token.AccessToken
	cfg.OneDrive.RefreshToken = token.RefreshToken
	cfg.OneDrive.TokenExpiry = token.Expiry.Unix()

	// Save the updated configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// logoutFromOneDrive clears OneDrive authentication
func logoutFromOneDrive(cfg *config.Config) error {
	cfg.OneDrive.AccessToken = ""
	cfg.OneDrive.RefreshToken = ""
	cfg.OneDrive.TokenExpiry = 0
	cfg.OneDrive.MonitorPaths = []string{}

	// Save the updated configuration
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// setupOneDrive sets up OneDrive folder monitoring
func setupOneDrive(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== OneDrive Setup ===")

	// Check if already logged in
	if cfg.OneDrive.AccessToken == "" {
		fmt.Println("Please login to OneDrive first using: restic-backup-checker login")
		fmt.Print("Would you like to login now? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response == "y" || response == "yes" {
			if err := loginToOneDrive(cfg); err != nil {
				return fmt.Errorf("failed to login to OneDrive: %w", err)
			}
		} else {
			return fmt.Errorf("OneDrive login required")
		}
	}

	// Get available folders
	client := onedrive.NewClient(cfg.OneDrive.AccessToken)
	folders, err := client.GetTopLevelFolders()
	if err != nil {
		return fmt.Errorf("failed to get OneDrive folders: %w", err)
	}

	// Prompt user to select folders to monitor
	fmt.Println("\nAvailable top-level folders:")
	for i, folder := range folders {
		fmt.Printf("%d. %s\n", i+1, folder.Name)
	}

	fmt.Print("\nEnter folder numbers to monitor (comma-separated): ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if selection != "" {
		indices := strings.Split(selection, ",")
		for _, idx := range indices {
			if i, err := strconv.Atoi(strings.TrimSpace(idx)); err == nil && i > 0 && i <= len(folders) {
				cfg.OneDrive.MonitorPaths = append(cfg.OneDrive.MonitorPaths, folders[i-1].ID)
			}
		}
	}

	return nil
}

// setupTelegram sets up Telegram configuration
func setupTelegram(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Telegram Setup ===")
	fmt.Println("Create a bot with @BotFather on Telegram and get the bot token.")
	fmt.Println()

	fmt.Print("Enter Telegram Bot Token: ")
	botToken, _ := reader.ReadString('\n')
	cfg.Telegram.BotToken = strings.TrimSpace(botToken)

	fmt.Print("Enter Telegram Chat ID: ")
	chatIDStr, _ := reader.ReadString('\n')
	chatIDStr = strings.TrimSpace(chatIDStr)

	if chatID, err := strconv.ParseInt(chatIDStr, 10, 64); err == nil {
		cfg.Telegram.ChatID = chatID
	} else {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	// Test Telegram connection
	tg := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	if err := tg.SendMessage("Backup checker setup completed successfully!"); err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}

	fmt.Println("✓ Telegram test message sent successfully!")
	return nil
}

// setupMonitoring sets up monitoring configuration
func setupMonitoring(cfg *config.Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Monitoring Setup ===")
	fmt.Print("Enter check interval in minutes (default: 60): ")
	intervalStr, _ := reader.ReadString('\n')
	intervalStr = strings.TrimSpace(intervalStr)

	if intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			cfg.Monitoring.CheckInterval = interval
		}
	}

	cfg.Monitoring.Enabled = true
	return nil
}

// showConfig displays the current configuration
func showConfig(cfg *config.Config) {
	fmt.Println("=== Current Configuration ===")
	fmt.Printf("OneDrive Authenticated: %v\n", cfg.OneDrive.AccessToken != "")
	fmt.Printf("OneDrive Monitoring Paths: %v\n", cfg.OneDrive.MonitorPaths)
	fmt.Printf("Telegram Bot Token: %s\n", maskToken(cfg.Telegram.BotToken))
	fmt.Printf("Telegram Chat ID: %d\n", cfg.Telegram.ChatID)
	fmt.Printf("Check Interval: %d minutes\n", cfg.Monitoring.CheckInterval)
	fmt.Printf("Monitoring Enabled: %v\n", cfg.Monitoring.Enabled)
}

// maskToken masks sensitive token information
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// confirmReset asks for confirmation before resetting configuration
func confirmReset() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to reset the configuration? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
} 