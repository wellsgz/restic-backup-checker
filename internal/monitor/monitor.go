package monitor

import (
	"fmt"
	"sync"
	"time"

	"restic-backup-checker/internal/config"
	"restic-backup-checker/internal/logger"
	"restic-backup-checker/internal/onedrive"
	"restic-backup-checker/internal/telegram"

	"golang.org/x/oauth2"
)

// Monitor represents the backup monitoring service
type Monitor struct {
	config       *config.Config
	onedriveAuth *onedrive.Authenticator
	telegram     *telegram.Client
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// BackupStatus represents the status of a backup check
type BackupStatus struct {
	ClientName string
	FolderPath string
	HasBackup  bool
	FileCount  int
	LastBackup time.Time
	Error      error
}

// New creates a new Monitor instance
func New(cfg *config.Config) *Monitor {
	auth := onedrive.NewAuthenticator()
	tg := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.ChatID)

	return &Monitor{
		config:       cfg,
		onedriveAuth: auth,
		telegram:     tg,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the monitoring service
func (m *Monitor) Start() error {
	if !m.config.Monitoring.Enabled {
		logger.Info("Monitoring is disabled")
		return nil
	}

	logger.Info("Starting backup monitoring service...")

	// Run initial check
	if err := m.CheckOnce(); err != nil {
		logger.Error("Initial backup check failed: %v", err)
	}

	// Start periodic monitoring
	m.wg.Add(1)
	go m.monitoringLoop()

	logger.Info("Backup monitoring service started")

	// Wait for stop signal
	<-m.stopChan
	m.wg.Wait()

	return nil
}

// Stop stops the monitoring service
func (m *Monitor) Stop() {
	logger.Info("Stopping backup monitoring service...")
	close(m.stopChan)
	m.wg.Wait()
	logger.Info("Backup monitoring service stopped")
}

// CheckOnce performs a single backup check
func (m *Monitor) CheckOnce() error {
	logger.Info("Starting backup check...")

	// Refresh token if needed
	if err := m.refreshTokenIfNeeded(); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	client := onedrive.NewClient(m.config.OneDrive.AccessToken)

	var statuses []BackupStatus
	var successCount, failedCount int
	var failedClients []string

	// Check each monitored path
	for i, folderID := range m.config.OneDrive.MonitorPaths {
		logger.Debug("Checking monitored path %d/%d: %s", i+1, len(m.config.OneDrive.MonitorPaths), folderID)

		// Get folder info for client names
		subfolders, err := client.GetSubfolders(folderID)
		if err != nil {
			logger.Error("Failed to get subfolders for %s: %v", folderID, err)
			continue
		}

		logger.Debug("Found %d client folders in monitored path: %s", len(subfolders), folderID)

		// Check each client folder
		for _, subfolder := range subfolders {
			logger.Debug("Checking client: %s (ID: %s)", subfolder.Name, subfolder.ID)

			status := m.checkClientBackup(client, subfolder.ID, subfolder.Name)
			statuses = append(statuses, status)

			if status.Error != nil {
				logger.Error("Error checking client %s: %v", status.ClientName, status.Error)
				failedCount++
				failedClients = append(failedClients, status.ClientName)
			} else if status.HasBackup {
				successCount++
				logger.Info("✅ Client %s: Backup found in last 24 hours (%d files)",
					status.ClientName, status.FileCount)
			} else {
				failedCount++
				failedClients = append(failedClients, status.ClientName)
				lastBackupStr := "Unknown"
				if !status.LastBackup.IsZero() {
					lastBackupStr = status.LastBackup.Format("2006-01-02 15:04:05")
				}
				logger.Error("❌ Client %s: No backup in last 24 hours, last backup: %s",
					status.ClientName, lastBackupStr)
			}
		}
	}

	// Send notifications
	if err := m.sendNotifications(statuses, successCount, failedCount, failedClients); err != nil {
		logger.Error("Failed to send notifications: %v", err)
	}

	logger.Info("Backup check completed. Success: %d, Failed: %d", successCount, failedCount)
	return nil
}

// monitoringLoop runs the periodic monitoring
func (m *Monitor) monitoringLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Duration(m.config.Monitoring.CheckInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.CheckOnce(); err != nil {
				logger.Error("Periodic backup check failed: %v", err)
			}
		case <-m.stopChan:
			return
		}
	}
}

// checkClientBackup checks backup status for a single client
func (m *Monitor) checkClientBackup(client *onedrive.Client, folderID, clientName string) BackupStatus {
	status := BackupStatus{
		ClientName: clientName,
		FolderPath: folderID,
	}

	// Check if there are backups in the last 24 hours
	hasBackup, recentFiles, err := client.CheckTodayBackups(folderID)
	if err != nil {
		status.Error = err
		logger.Error("Failed to check backup for client %s: %v", clientName, err)
		return status
	}

	status.HasBackup = hasBackup
	status.FileCount = len(recentFiles)

	// Get all backup files to find the most recent one
	allFiles, err := client.GetAllSnapshots(folderID)
	if err != nil {
		logger.Error("Failed to get all snapshots for client %s: %v", clientName, err)
		// Continue with partial data - we still have today's backup status
	} else {
		// Find the most recent backup from all files
		var latestBackup time.Time
		for _, file := range allFiles {
			if file.CreatedTime.After(latestBackup) {
				latestBackup = file.CreatedTime
			}
		}
		status.LastBackup = latestBackup

		// Log backup information for debugging
		if !latestBackup.IsZero() {
			logger.Debug("Client %s: Last backup was %s, Recent backup (24h): %v",
				clientName, latestBackup.Format("2006-01-02 15:04:05"), hasBackup)
		} else {
			logger.Debug("Client %s: No backups found, Recent backup (24h): %v",
				clientName, hasBackup)
		}
	}

	return status
}

// sendNotifications sends appropriate notifications based on backup status
func (m *Monitor) sendNotifications(statuses []BackupStatus, successCount, failedCount int, failedClients []string) error {
	if m.telegram == nil {
		return fmt.Errorf("telegram client not initialized")
	}

	// Send individual alerts for failed backups
	for _, status := range statuses {
		if !status.HasBackup {
			lastBackupStr := "Unknown"
			if !status.LastBackup.IsZero() {
				lastBackupStr = status.LastBackup.Format("2006-01-02 15:04:05")
			}

			if err := m.telegram.SendBackupAlert(
				status.ClientName,
				status.FolderPath,
				lastBackupStr,
			); err != nil {
				logger.Error("Failed to send backup alert for %s: %v", status.ClientName, err)
			}
		}
	}

	// Send summary report
	totalClients := len(statuses)
	if err := m.telegram.SendSummaryReport(totalClients, successCount, failedCount, failedClients); err != nil {
		logger.Error("Failed to send summary report: %v", err)
		return err
	}

	return nil
}

// refreshTokenIfNeeded refreshes the OAuth token if it's expired
func (m *Monitor) refreshTokenIfNeeded() error {
	if m.config.OneDrive.TokenExpiry == 0 {
		return fmt.Errorf("no token expiry set")
	}

	expiry := time.Unix(m.config.OneDrive.TokenExpiry, 0)
	if time.Now().Before(expiry.Add(-10 * time.Minute)) {
		// Token is still valid (with 10 minute buffer)
		return nil
	}

	logger.Info("Refreshing OneDrive token...")

	// Create token from stored values
	token := &oauth2.Token{
		AccessToken:  m.config.OneDrive.AccessToken,
		RefreshToken: m.config.OneDrive.RefreshToken,
		Expiry:       expiry,
	}

	// Refresh the token
	newToken, err := m.onedriveAuth.RefreshToken(token)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update configuration
	m.config.OneDrive.AccessToken = newToken.AccessToken
	m.config.OneDrive.RefreshToken = newToken.RefreshToken
	m.config.OneDrive.TokenExpiry = newToken.Expiry.Unix()

	// Save updated configuration
	if err := m.config.Save(); err != nil {
		logger.Error("Failed to save updated token: %v", err)
	}

	logger.Info("OneDrive token refreshed successfully")
	return nil
}
