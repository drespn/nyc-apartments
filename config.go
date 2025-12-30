package main

import (
	"errors"
	"os"
)

// Config holds the application configuration
type Config struct {
	DiscordWebhookURL string
	DatabasePath      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return nil, errors.New("DISCORD_WEBHOOK_URL environment variable is required")
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./apartments.db"
	}

	return &Config{
		DiscordWebhookURL: webhookURL,
		DatabasePath:      dbPath,
	}, nil
}
