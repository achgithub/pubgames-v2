package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		is_admin INTEGER DEFAULT 0,
		-- Passkey fields (for future WebAuthn support)
		passkey_id TEXT,
		passkey_public_key TEXT,
		passkey_counter INTEGER DEFAULT 0,
		passkey_transports TEXT,
		passkey_created_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Apps table
	CREATE TABLE IF NOT EXISTS apps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		is_active BOOLEAN DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- User activity tracking table
	CREATE TABLE IF NOT EXISTS user_activity (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		app_id INTEGER NOT NULL,
		accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (app_id) REFERENCES apps(id)
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("✅ Database initialized at", DB_PATH)

	// Seed initial data if needed
	seedData()
}

// seedData adds initial admin user and sample apps if database is empty
func seedData() {
	// Check if admin user exists
	var adminCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&adminCount)

	if adminCount == 0 {
		log.Println("   Creating default admin user...")
		
		// Create admin with code "123456"
		hashedCode, _ := bcrypt.GenerateFromPassword([]byte("123456"), 12)
		_, err := db.Exec(`
			INSERT INTO users (email, name, code, is_admin) 
			VALUES (?, ?, ?, ?)
		`, "admin@pubgames.local", "Admin User", string(hashedCode), 1)
		
		if err != nil {
			log.Printf("Warning: Failed to create admin user: %v", err)
		} else {
			log.Println("   ✅ Default admin created: admin@pubgames.local / 123456")
		}
	}

	// Check if apps exist
	var appCount int
	db.QueryRow("SELECT COUNT(*) FROM apps").Scan(&appCount)

	if appCount == 0 {
		log.Println("   Creating sample apps...")
		
		apps := []struct {
			name        string
			url         string
			description string
			icon        string
		}{
			{
				name:        "Last Man Standing",
				url:         "http://localhost:30010",
				description: "Tournament prediction game",
				icon:        "🏆",
			},
			{
				name:        "Sweepstakes",
				url:         "http://localhost:30020",
				description: "Blind box competition draws",
				icon:        "🎰",
			},
		}

		for _, app := range apps {
			_, err := db.Exec(`
				INSERT INTO apps (name, url, description, icon, is_active) 
				VALUES (?, ?, ?, ?, ?)
			`, app.name, app.url, app.description, app.icon, true)
			
			if err != nil {
				log.Printf("Warning: Failed to insert app %s: %v", app.name, err)
			}
		}

		log.Println("   ✅ Sample apps created")
	}
}
