package main

import (
	"log"
)

func main() {
	log.Println("NYC Apartment Notifier starting...")

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded. Database path: %s", cfg.DatabasePath)
	log.Println("Setup complete!")
}
