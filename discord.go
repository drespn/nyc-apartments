package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	discordEmbedColor = 5814783 // Light blue color
)

// DiscordClient handles sending webhooks to Discord
type DiscordClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewDiscordClient creates a new Discord webhook client
func NewDiscordClient(webhookURL string) *DiscordClient {
	return &DiscordClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendListing sends a formatted listing embed to Discord
func (d *DiscordClient) SendListing(listing Listing) error {
	embed := d.buildEmbed(listing)

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", d.webhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}

	return nil
}

// buildEmbed constructs the Discord embed for a listing
func (d *DiscordClient) buildEmbed(listing Listing) map[string]interface{} {
	// Build title: "123 Main Street, Unit 4B"
	title := listing.Street
	if listing.Unit != "" {
		title = fmt.Sprintf("%s, Unit %s", listing.Street, listing.Unit)
	}

	// Build listing URL
	listingURL := fmt.Sprintf("https://streeteasy.com%s", listing.URLPath)

	// Build bedroom display
	bedroomDisplay := "Studio"
	if listing.BedroomCount == 1 {
		bedroomDisplay = "1 Bed"
	} else if listing.BedroomCount > 1 {
		bedroomDisplay = fmt.Sprintf("%d Beds", listing.BedroomCount)
	}

	// Build bathroom display
	totalBaths := float64(listing.FullBathroomCount) + (float64(listing.HalfBathroomCount) * 0.5)
	bathroomDisplay := fmt.Sprintf("%.0f Bath", totalBaths)
	if totalBaths != 1 {
		if totalBaths == float64(int(totalBaths)) {
			bathroomDisplay = fmt.Sprintf("%.0f Baths", totalBaths)
		} else {
			bathroomDisplay = fmt.Sprintf("%.1f Baths", totalBaths)
		}
	}

	// Build fields
	fields := []map[string]interface{}{
		{
			"name":   "Price",
			"value":  fmt.Sprintf("$%d/mo", listing.Price),
			"inline": true,
		},
		{
			"name":   "Type",
			"value":  bedroomDisplay,
			"inline": true,
		},
		{
			"name":   "Bath",
			"value":  bathroomDisplay,
			"inline": true,
		},
	}

	// Add broker if available
	if listing.SourceGroupLabel != "" {
		fields = append(fields, map[string]interface{}{
			"name":   "Broker",
			"value":  listing.SourceGroupLabel,
			"inline": false,
		})
	}

	embed := map[string]interface{}{
		"title":       title,
		"url":         listingURL,
		"description": listing.AreaName,
		"color":       discordEmbedColor,
		"fields":      fields,
	}

	// Add thumbnail if photo available
	if listing.PhotoKey != "" {
		photoURL := fmt.Sprintf("https://photos.streeteasy.com/%s/webp/large", listing.PhotoKey)
		embed["thumbnail"] = map[string]interface{}{
			"url": photoURL,
		}
	}

	return embed
}
