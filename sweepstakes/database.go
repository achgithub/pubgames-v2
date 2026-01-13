package main

import (
	"database/sql"
	"fmt"
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

	log.Println("🔍 Checking database schema...")

	// Create schema
	schema := `
	-- Competitions table
	CREATE TABLE IF NOT EXISTS competitions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL CHECK(type IN ('knockout', 'race')),
		status TEXT DEFAULT 'draft' CHECK(status IN ('draft', 'open', 'locked', 'completed', 'archived')),
		start_date DATETIME,
		end_date DATETIME,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Entries table (horses, teams, participants in competitions)
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		competition_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		seed INTEGER,
		number INTEGER,
		status TEXT DEFAULT 'available' CHECK(status IN ('available', 'taken', 'active', 'eliminated', 'winner')),
		stage TEXT,
		eliminated_date DATETIME,
		position INTEGER,
		FOREIGN KEY (competition_id) REFERENCES competitions(id),
		UNIQUE(competition_id, name)
	);

	-- Draws table (user selections/assignments)
	-- Note: user_email is the JWT sub (email) from Identity Service, not a local user_id
	CREATE TABLE IF NOT EXISTS draws (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_email TEXT NOT NULL,
		competition_id INTEGER NOT NULL,
		entry_id INTEGER NOT NULL,
		drawn_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (competition_id) REFERENCES competitions(id),
		FOREIGN KEY (entry_id) REFERENCES entries(id)
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	ensureTableIntegrity()

	log.Println("✅ Database initialized successfully")
}

// ensureTableIntegrity adds missing columns for older databases
func ensureTableIntegrity() {
	columns := []struct {
		table  string
		column string
		def    string
	}{
		{"entries", "eliminated_date", "DATETIME"},
		{"entries", "position", "INTEGER"},
	}

	for _, col := range columns {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'", col.table, col.column)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			continue
		}

		if count == 0 {
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", col.table, col.column, col.def)
			_, err := db.Exec(alterSQL)
			if err == nil {
				log.Printf("✅ Added column: %s.%s", col.table, col.column)
			}
		}
	}
}
