package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// initDB initializes the database connection and creates tables
func initDB() {
	// Ensure data directory exists
	dataDir := filepath.Dir(DB_PATH)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Open database connection
	var err error
	db, err = sql.Open("sqlite3", DB_PATH)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create schema
	schema := `
	-- Users table (for local reference, actual auth via Identity Service)
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		is_admin INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Sample items table (replace with your app's tables)
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Add more tables for your specific app here
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("✅ Database initialized at", DB_PATH)
}

// seedData adds sample data (optional, for testing)
func seedData() {
	// Check if data already exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)
	if count > 0 {
		log.Println("   Database already contains data, skipping seed")
		return
	}

	// Insert sample data
	items := []struct {
		name        string
		description string
	}{
		{"Sample Item 1", "This is a sample item"},
		{"Sample Item 2", "Another sample item"},
		{"Sample Item 3", "Yet another sample item"},
	}

	for _, item := range items {
		_, err := db.Exec(`
			INSERT INTO items (name, description) 
			VALUES (?, ?)
		`, item.name, item.description)
		if err != nil {
			log.Printf("Warning: Failed to insert sample item: %v", err)
		}
	}

	log.Println("   Seeded sample data")
}
