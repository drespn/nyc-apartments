package main

import (
	"fmt"
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

	// Test: Fetch listings from StreetEasy
	log.Println("Fetching listings from StreetEasy...")
	client := NewStreetEasyClient()
	listings, err := client.FetchListings()
	if err != nil {
		log.Fatalf("Failed to fetch listings: %v", err)
	}

	log.Printf("Found %d listings", len(listings))

	// Print first 5 listings as a sample
	for i, listing := range listings {
		if i >= 5 {
			break
		}
		fmt.Printf("\n--- Listing %d ---\n", i+1)
		fmt.Printf("ID: %s\n", listing.ID)
		fmt.Printf("Address: %s, Unit %s\n", listing.Street, listing.Unit)
		fmt.Printf("Area: %s\n", listing.AreaName)
		fmt.Printf("Price: $%d/mo\n", listing.Price)
		fmt.Printf("Bedrooms: %d\n", listing.BedroomCount)
		fmt.Printf("Bathrooms: %d full, %d half\n", listing.FullBathroomCount, listing.HalfBathroomCount)
		fmt.Printf("URL: https://streeteasy.com%s\n", listing.URLPath)
	}

	log.Println("\nSetup complete!")
}
