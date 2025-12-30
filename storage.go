package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Storage handles SQLite database operations
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new storage instance and initializes the database
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &Storage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the required tables if they don't exist
func (s *Storage) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS seen_listings (
		id TEXT PRIMARY KEY,
		street TEXT,
		unit TEXT,
		area_name TEXT,
		price INTEGER,
		first_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// IsNew checks if a listing ID has not been seen before
func (s *Storage) IsNew(listingID string) (bool, error) {
	var exists int
	query := `SELECT 1 FROM seen_listings WHERE id = ? LIMIT 1`
	err := s.db.QueryRow(query, listingID).Scan(&exists)

	if err == sql.ErrNoRows {
		return true, nil // Not found = new listing
	}
	if err != nil {
		return false, fmt.Errorf("failed to check listing: %w", err)
	}

	return false, nil // Found = not new
}

// MarkSeen inserts a listing into the database
func (s *Storage) MarkSeen(listing Listing) error {
	query := `
	INSERT OR IGNORE INTO seen_listings (id, street, unit, area_name, price)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, listing.ID, listing.Street, listing.Unit, listing.AreaName, listing.Price)
	if err != nil {
		return fmt.Errorf("failed to insert listing: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}
