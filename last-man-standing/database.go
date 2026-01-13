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
	-- Games/Competitions table
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		status TEXT DEFAULT 'active',
		winner_count INTEGER DEFAULT 0,
		postponement_rule TEXT DEFAULT 'loss',
		start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_date TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Game players junction table (links users to games)
	CREATE TABLE IF NOT EXISTS game_players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		game_id INTEGER NOT NULL,
		is_active BOOLEAN DEFAULT 1,
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id),
		UNIQUE(user_id, game_id)
	);

	-- Rounds table
	CREATE TABLE IF NOT EXISTS rounds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		submission_deadline TEXT NOT NULL,
		status TEXT DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id),
		UNIQUE(game_id, round_number)
	);

	-- Matches table
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		match_number INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		date TEXT NOT NULL,
		location TEXT NOT NULL,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		result TEXT DEFAULT '',
		status TEXT DEFAULT 'upcoming',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id)
	);

	-- Predictions table
	CREATE TABLE IF NOT EXISTS predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		game_id INTEGER NOT NULL,
		match_id INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		predicted_team TEXT NOT NULL,
		is_correct BOOLEAN DEFAULT NULL,
		voided BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id),
		FOREIGN KEY (match_id) REFERENCES matches (id),
		UNIQUE(user_id, game_id, round_number)
	);

	-- Current game tracking table (simple key-value store)
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("✅ Database initialized at", DB_PATH)

	// Initialize default game if none exists
	initializeDefaultGame()
}

// initializeDefaultGame creates a default game if database is empty
func initializeDefaultGame() {
	var gameCount int
	db.QueryRow("SELECT COUNT(*) FROM games").Scan(&gameCount)
	
	if gameCount == 0 {
		result, err := db.Exec("INSERT INTO games (name, status, postponement_rule) VALUES (?, ?, ?)", 
			"Game 1", "active", "loss")
		if err != nil {
			log.Printf("Warning: Failed to create default game: %v", err)
			return
		}
		
		gameID, _ := result.LastInsertId()
		
		// Set as current game
		db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('current_game_id', ?)", gameID)
		
		log.Printf("   Created default game (ID: %d)", gameID)
	}
}

// getCurrentGameID retrieves the current active game ID
func getCurrentGameID() int {
	var gameID int
	err := db.QueryRow("SELECT CAST(value AS INTEGER) FROM settings WHERE key = 'current_game_id'").Scan(&gameID)
	if err != nil {
		// If no current game set, return 0
		return 0
	}
	return gameID
}
