package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	discordEmbedColor      = 5814783  // Light blue color
	discordErrorColor      = 15158332 // Red color
	discordStatusColor     = 3066993  // Green color
)

// DiscordClient handles sending webhooks to Discord
type DiscordClient struct {
	webhookURL       string
	errorWebhookURL  string
	statusWebhookURL string
	httpClient       *http.Client
}

// NewDiscordClient creates a new Discord webhook client
func NewDiscordClient(webhookURL, errorWebhookURL, statusWebhookURL string) *DiscordClient {
	return &DiscordClient{
		webhookURL:       webhookURL,
		errorWebhookURL:  errorWebhookURL,
		statusWebhookURL: statusWebhookURL,
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
	// Title is the neighborhood (what you care about most)
	title := listing.AreaName

	// Description is the address
	address := listing.Street
	if listing.Unit != "" {
		address = fmt.Sprintf("%s, Unit %s", listing.Street, listing.Unit)
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
		"description": address,
		"color":       discordEmbedColor,
		"fields":      fields,
	}

	// Add thumbnail if photo available
	if listing.PhotoKey != "" {
		photoURL := fmt.Sprintf("https://photos.zillowstatic.com/fp/%s-se_extra_large_1500_800.webp", listing.PhotoKey)
		embed["thumbnail"] = map[string]interface{}{
			"url": photoURL,
		}
	}

	return embed
}

// SendError sends an error notification to the error webhook
func (d *DiscordClient) SendError(errMsg string) error {
	if d.errorWebhookURL == "" {
		return nil // No error webhook configured
	}

	embed := map[string]interface{}{
		"title":       "Error",
		"description": errMsg,
		"color":       discordErrorColor,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal error payload: %w", err)
	}

	req, err := http.NewRequest("POST", d.errorWebhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create error request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send error webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// SendStatus sends a status update to the status webhook
func (d *DiscordClient) SendStatus(totalListings, newListings int, sampleListings []Listing) error {
	if d.statusWebhookURL == "" {
		return nil // No status webhook configured
	}

	// Build sample listings text
	sampleText := ""
	for i, listing := range sampleListings {
		if i >= 3 { // Show max 3 samples
			break
		}
		sampleText += fmt.Sprintf("• %s - %s, $%d/mo\n", listing.AreaName, listing.Street, listing.Price)
	}
	if sampleText == "" {
		sampleText = "No listings in response"
	}

	embed := map[string]interface{}{
		"title": "Poll Complete",
		"color": discordStatusColor,
		"fields": []map[string]interface{}{
			{
				"name":   "Total Listings",
				"value":  fmt.Sprintf("%d", totalListings),
				"inline": true,
			},
			{
				"name":   "New Listings",
				"value":  fmt.Sprintf("%d", newListings),
				"inline": true,
			},
			{
				"name":   "Sample from Response",
				"value":  sampleText,
				"inline": false,
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal status payload: %w", err)
	}

	req, err := http.NewRequest("POST", d.statusWebhookURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create status request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send status webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
