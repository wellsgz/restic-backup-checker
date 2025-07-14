package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Client represents a Telegram bot client
type Client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

// New creates a new Telegram client
func New(botToken string, chatID int64) *Client {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("Failed to create Telegram bot: %v", err)
		return nil
	}

	return &Client{
		bot:    bot,
		chatID: chatID,
	}
}

// SendMessage sends a message to the configured chat
func (c *Client) SendMessage(message string) error {
	if c.bot == nil {
		return fmt.Errorf("telegram bot not initialized")
	}

	msg := tgbotapi.NewMessage(c.chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

// SendBackupAlert sends a backup failure alert
func (c *Client) SendBackupAlert(clientName string, folderPath string, lastBackupTime string) error {
	message := fmt.Sprintf(
		"🚨 *Backup Alert*\n\n"+
			"*Client:* %s\n"+
			"*Folder:* %s\n"+
			"*Issue:* No backup found for today\n"+
			"*Last Backup:* %s\n\n"+
			"Please check the backup client immediately.",
		clientName, folderPath, lastBackupTime,
	)

	return c.SendMessage(message)
}

// SendBackupSuccess sends a backup success notification
func (c *Client) SendBackupSuccess(clientName string, folderPath string, fileCount int) error {
	message := fmt.Sprintf(
		"✅ *Backup Success*\n\n"+
			"*Client:* %s\n"+
			"*Folder:* %s\n"+
			"*Files:* %d backup files found for today\n\n"+
			"All backups are up to date.",
		clientName, folderPath, fileCount,
	)

	return c.SendMessage(message)
}

// SendSummaryReport sends a daily summary report
func (c *Client) SendSummaryReport(totalClients int, successCount int, failedCount int, failedClients []string) error {
	status := "✅ All Good"
	if failedCount > 0 {
		status = "🚨 Issues Found"
	}

	message := fmt.Sprintf(
		"📊 *Daily Backup Report*\n\n"+
			"*Status:* %s\n"+
			"*Total Clients:* %d\n"+
			"*Successful:* %d\n"+
			"*Failed:* %d\n",
		status, totalClients, successCount, failedCount,
	)

	if len(failedClients) > 0 {
		message += "\n*Failed Clients:*\n"
		for _, client := range failedClients {
			message += fmt.Sprintf("• %s\n", client)
		}
	}

	return c.SendMessage(message)
}
