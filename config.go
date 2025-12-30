package main

import (
	"errors"
	"os"
)

// Config holds the application configuration
type Config struct {
	DiscordWebhookURL       string
	DiscordErrorWebhookURL  string
	DiscordStatusWebhookURL string
	DatabasePath            string
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

	errorWebhookURL := os.Getenv("DISCORD_ERROR_WEBHOOK_URL")
	statusWebhookURL := os.Getenv("DISCORD_STATUS_WEBHOOK_URL")

	return &Config{
		DiscordWebhookURL:       webhookURL,
		DiscordErrorWebhookURL:  errorWebhookURL,
		DiscordStatusWebhookURL: statusWebhookURL,
		DatabasePath:            dbPath,
	}, nil
}
