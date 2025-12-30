package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	streetEasyAPI     = "https://api-v6.streeteasy.com/"
	apolloClientName  = "srp-frontend-service"
	apolloVersion     = "version 28acce3818ba1c642a4e7f28710199fdbc967f37"
	userAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
)

// StreetEasyClient handles API requests to StreetEasy
type StreetEasyClient struct {
	httpClient *http.Client
}

// NewStreetEasyClient creates a new StreetEasy API client
func NewStreetEasyClient() *StreetEasyClient {
	return &StreetEasyClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchListings fetches apartment listings from StreetEasy
func (c *StreetEasyClient) FetchListings() ([]Listing, error) {
	// Build the request body
	requestBody := c.buildRequestBody()

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", streetEasyAPI, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers (matching browser request)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://streeteasy.com")
	req.Header.Set("Referer", "https://streeteasy.com/")
	req.Header.Set("apollographql-client-name", apolloClientName)
	req.Header.Set("apollographql-client-version", apolloVersion)
	req.Header.Set("app-version", "1.0.0")
	req.Header.Set("os", "web")
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var graphQLResponse GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphQLResponse.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResponse.Errors[0].Message)
	}

	// Check for nil data
	if graphQLResponse.Data == nil {
		return nil, fmt.Errorf("no data in response")
	}

	// Convert to Listing slice
	var listings []Listing
	for _, edge := range graphQLResponse.Data.SearchRentals.Edges {
		if edge.Node != nil {
			listings = append(listings, edge.Node.ToListing())
		}
	}

	return listings, nil
}

// buildRequestBody constructs the GraphQL request body
func (c *StreetEasyClient) buildRequestBody() map[string]interface{} {
	query := `
  query GetListingRental($input: SearchRentalsInput!) {
    searchRentals(input: $input) {
      search {
        criteria
      }
      totalCount
      edges {
        ... on OrganicRentalEdge {
          node {
            id
            areaName
            bedroomCount
            buildingType
            fullBathroomCount
            geoPoint {
              latitude
              longitude
            }
            halfBathroomCount
            leadMedia {
              photo {
                  key
              }
            }
            price
            relloExpress {
              ctaEnabled
              link
              rentalId
            }
            sourceGroupLabel
            status
            street
            unit
            urlPath
            tier
          }
        }
      }
    }
  }
`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"filters": map[string]interface{}{
				"rentalStatus": "ACTIVE",
				"areas": []int{
					101, 103, 104, 105, 106, 107, 108, 109, 110, 112, 113, 115, 116, 117,
					120, 122, 130, 131, 132, 133, 136, 141, 146, 152, 157, 158, 162, 478,
				},
				"price": map[string]interface{}{
					"lowerBound": 2000,
					"upperBound": 2750,
				},
				"boundingBox": map[string]interface{}{
					"topLeft": map[string]interface{}{
						"latitude":  40.774,
						"longitude": -74.036,
					},
					"bottomRight": map[string]interface{}{
						"latitude":  40.698,
						"longitude": -73.926,
					},
				},
			},
			"page":            1,
			"perPage":         500,
			"sorting": map[string]interface{}{
				"attribute": "RECOMMENDED",
				"direction": "DESCENDING",
			},
			"userSearchToken": uuid.New().String(),
			"adStrategy":      "NONE",
		},
	}

	return map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
}
