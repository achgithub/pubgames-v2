package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", DB_PATH)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🔍 Checking database schema...")

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		is_admin INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

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

	CREATE TABLE IF NOT EXISTS draws (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		competition_id INTEGER NOT NULL,
		entry_id INTEGER NOT NULL,
		drawn_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (competition_id) REFERENCES competitions(id),
		FOREIGN KEY (entry_id) REFERENCES entries(id)
	);

	CREATE TABLE IF NOT EXISTS admin_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		admin_password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	ensureTableIntegrity()

	// Create default admin password
	var count int
	db.QueryRow("SELECT COUNT(*) FROM admin_settings").Scan(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Exec("INSERT INTO admin_settings (admin_password) VALUES (?)", string(hashedPassword))
		log.Println("✅ Default admin password created: admin123")
	}

	log.Println("✅ Database initialized successfully")
}

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

	// Remove old placement column if exists
	var hasPlacement int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('entries') WHERE name='placement'").Scan(&hasPlacement)
	if hasPlacement > 0 {
		log.Printf("ℹ️  Note: 'placement' column exists but 'position' is now used")
	}
}
