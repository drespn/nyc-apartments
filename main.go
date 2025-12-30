package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	log.Println("NYC Apartment Notifier starting...")

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config loaded. Database path: %s", cfg.DatabasePath)

	// Initialize storage
	storage, err := NewStorage(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()
	log.Println("Database initialized")

	// Initialize clients
	streetEasyClient := NewStreetEasyClient()
	discordClient := NewDiscordClient(cfg.DiscordWebhookURL, cfg.DiscordErrorWebhookURL, cfg.DiscordStatusWebhookURL)

	// Create poll function
	poll := func() {
		log.Println("Starting poll...")

		listings, err := streetEasyClient.FetchListings()
		if err != nil {
			log.Printf("Error fetching listings: %v", err)
			discordClient.SendError(fmt.Sprintf("Failed to fetch listings: %v", err))
			return
		}
		log.Printf("Fetched %d total listings", len(listings))

		newCount := 0
		for _, listing := range listings {
			isNew, err := storage.IsNew(listing.ID)
			if err != nil {
				log.Printf("Error checking listing %s: %v", listing.ID, err)
				discordClient.SendError(fmt.Sprintf("Error checking listing %s: %v", listing.ID, err))
				continue
			}

			if isNew {
				// Send Discord notification
				if err := discordClient.SendListing(listing); err != nil {
					log.Printf("Error sending Discord notification for %s: %v", listing.ID, err)
					discordClient.SendError(fmt.Sprintf("Error sending notification for %s: %v", listing.ID, err))
					continue
				}

				// Mark as seen
				if err := storage.MarkSeen(listing); err != nil {
					log.Printf("Error marking listing %s as seen: %v", listing.ID, err)
					discordClient.SendError(fmt.Sprintf("Error marking listing %s as seen: %v", listing.ID, err))
					continue
				}

				log.Printf("New listing: %s, %s - $%d/mo (%s)",
					listing.Street, listing.Unit, listing.Price, listing.AreaName)
				newCount++

				// Rate limit: wait 500ms between Discord messages
				time.Sleep(500 * time.Millisecond)
			}
		}

		log.Printf("Poll complete. Found %d new listings.", newCount)

		// Send status update
		if err := discordClient.SendStatus(len(listings), newCount, listings); err != nil {
			log.Printf("Error sending status update: %v", err)
			discordClient.SendError(fmt.Sprintf("Error sending status update: %v", err))
		}
	}

	// Run poll immediately on startup
	log.Println("Running initial poll...")
	poll()

	// Set up cron scheduler for every 30 minutes
	c := cron.New()
	_, err = c.AddFunc("*/30 * * * *", poll)
	if err != nil {
		discordClient.SendError(fmt.Sprintf("Failed to add cron job: %v", err))
		log.Fatalf("Failed to add cron job: %v", err)
	}
	c.Start()
	log.Println("Scheduler started. Polling every 30 minutes.")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	c.Stop()
	log.Println("Scheduler stopped. Goodbye!")
}
