package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/pbkdf2"
)

// Config represents the application configuration
type Config struct {
	OneDrive     OneDriveConfig     `json:"onedrive"`
	Telegram     TelegramConfig     `json:"telegram"`
	Monitoring   MonitoringConfig   `json:"monitoring"`
	configPath   string
	encryptionKey []byte
}

// OneDriveConfig holds OneDrive authentication and configuration
type OneDriveConfig struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenExpiry  int64    `json:"token_expiry"`
	MonitorPaths []string `json:"monitor_paths"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

// MonitoringConfig holds monitoring settings
type MonitoringConfig struct {
	CheckInterval int  `json:"check_interval"` // in minutes
	Enabled       bool `json:"enabled"`
}

// Load loads the configuration from encrypted file
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	cfg := &Config{
		configPath: configPath,
		Monitoring: MonitoringConfig{
			CheckInterval: 60, // default to 1 hour
			Enabled:       true,
		},
	}

	// Generate encryption key from machine-specific data
	cfg.encryptionKey = generateEncryptionKey()

	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		if err := cfg.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return cfg, nil
}

// Save saves the configuration to encrypted file
func (c *Config) Save() error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	encrypted, err := c.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(c.configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(c.configPath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadFromFile loads configuration from encrypted file
func (c *Config) loadFromFile() error {
	encrypted, err := os.ReadFile(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	decrypted, err := c.decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt config: %w", err)
	}

	if err := json.Unmarshal(decrypted, c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// encrypt encrypts data using AES-GCM
func (c *Config) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func (c *Config) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "restic-backup-checker", "config.enc"), nil
}

// generateEncryptionKey generates a machine-specific encryption key
func generateEncryptionKey() []byte {
	// Use hostname and user as base for key derivation
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}

	salt := fmt.Sprintf("%s:%s", hostname, user)
	return pbkdf2.Key([]byte(salt), []byte("restic-backup-checker-salt"), 100000, 32, sha256.New)
}

// IsConfigured returns true if the configuration is properly set up
func (c *Config) IsConfigured() bool {
	return c.OneDrive.AccessToken != "" && c.Telegram.BotToken != ""
} 