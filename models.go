package main

// Listing represents an apartment listing from StreetEasy
type Listing struct {
	ID                string
	AreaName          string
	BedroomCount      int
	BuildingType      string
	FullBathroomCount int
	HalfBathroomCount int
	PhotoKey          string
	Price             int
	SourceGroupLabel  string
	Status            string
	Street            string
	Unit              string
	URLPath           string
}

// GraphQL response structures

type GraphQLResponse struct {
	Data   *ResponseData   `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type ResponseData struct {
	SearchRentals SearchRentalsResult `json:"searchRentals"`
}

type SearchRentalsResult struct {
	TotalCount int    `json:"totalCount"`
	Edges      []Edge `json:"edges"`
}

type Edge struct {
	Node *ListingNode `json:"node"`
}

type ListingNode struct {
	ID                string     `json:"id"`
	AreaName          string     `json:"areaName"`
	BedroomCount      int        `json:"bedroomCount"`
	BuildingType      string     `json:"buildingType"`
	FullBathroomCount int        `json:"fullBathroomCount"`
	HalfBathroomCount int        `json:"halfBathroomCount"`
	LeadMedia         *LeadMedia `json:"leadMedia"`
	Price             int        `json:"price"`
	SourceGroupLabel  string     `json:"sourceGroupLabel"`
	Status            string     `json:"status"`
	Street            string     `json:"street"`
	Unit              string     `json:"unit"`
	URLPath           string     `json:"urlPath"`
}

type LeadMedia struct {
	Photo *Photo `json:"photo"`
}

type Photo struct {
	Key string `json:"key"`
}

// ToListing converts a ListingNode from the API response to our Listing model
func (n *ListingNode) ToListing() Listing {
	photoKey := ""
	if n.LeadMedia != nil && n.LeadMedia.Photo != nil {
		photoKey = n.LeadMedia.Photo.Key
	}

	return Listing{
		ID:                n.ID,
		AreaName:          n.AreaName,
		BedroomCount:      n.BedroomCount,
		BuildingType:      n.BuildingType,
		FullBathroomCount: n.FullBathroomCount,
		HalfBathroomCount: n.HalfBathroomCount,
		PhotoKey:          photoKey,
		Price:             n.Price,
		SourceGroupLabel:  n.SourceGroupLabel,
		Status:            n.Status,
		Street:            n.Street,
		Unit:              n.Unit,
		URLPath:           n.URLPath,
	}
}
