package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

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
	-- Games table - stores all game sessions
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		player1_id INTEGER NOT NULL,
		player1_name TEXT NOT NULL,
		player2_id INTEGER,
		player2_name TEXT,
		mode TEXT NOT NULL DEFAULT 'normal',
		status TEXT NOT NULL DEFAULT 'waiting',
		current_turn INTEGER DEFAULT 1,
		winner_id INTEGER,
		board TEXT DEFAULT '["","","","","","","","",""]',
		move_time_limit INTEGER DEFAULT 0,
		session_timeout INTEGER DEFAULT 60,
		first_to INTEGER DEFAULT 1,
		player1_score INTEGER DEFAULT 0,
		player2_score INTEGER DEFAULT 0,
		current_round INTEGER DEFAULT 1,
		last_move_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP
	);

	-- Moves table - stores all moves made in games
	CREATE TABLE IF NOT EXISTS moves (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		position INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id)
	);

	-- Online users table - tracks who's currently active
	CREATE TABLE IF NOT EXISTS online_users (
		user_id INTEGER PRIMARY KEY,
		user_name TEXT NOT NULL,
		last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		in_game INTEGER DEFAULT 0
	);

	-- Player stats table - aggregated statistics
	CREATE TABLE IF NOT EXISTS player_stats (
		user_id INTEGER PRIMARY KEY,
		user_name TEXT NOT NULL,
		games_played INTEGER DEFAULT 0,
		games_won INTEGER DEFAULT 0,
		games_lost INTEGER DEFAULT 0,
		games_draw INTEGER DEFAULT 0,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Rematch requests table
	CREATE TABLE IF NOT EXISTS rematch_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		requester_id INTEGER NOT NULL,
		opponent_id INTEGER NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id)
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_games_player1 ON games(player1_id);
	CREATE INDEX IF NOT EXISTS idx_games_player2 ON games(player2_id);
	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
	CREATE INDEX IF NOT EXISTS idx_moves_game ON moves(game_id);
	CREATE INDEX IF NOT EXISTS idx_online_users_last_seen ON online_users(last_seen_at);
	CREATE INDEX IF NOT EXISTS idx_rematch_game ON rematch_requests(game_id);
	CREATE INDEX IF NOT EXISTS idx_rematch_status ON rematch_requests(status);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("✅ Database initialized at", DB_PATH)
	
	// Clean up state from previous server run
	cleanupOnServerRestart()
	
	// Clean up old online users (anyone not seen in last 5 minutes is offline)
	cleanupOnlineUsers()
	
	// Clean up expired rematch requests
	cleanupExpiredRematches()
}

// cleanupOnServerRestart cleans up stale state from previous server run
func cleanupOnServerRestart() {
	// Mark all active games as abandoned - can't continue after restart
	_, err := db.Exec(`
		UPDATE games
		SET status = 'abandoned'
		WHERE status = 'active'
	`)
	if err != nil {
		log.Printf("Warning: Failed to mark active games as abandoned on restart: %v", err)
	} else {
		log.Println("✅ Marked active games as abandoned from previous session")
	}
	
	// Clear all pending challenges (waiting games) - these are stale after restart
	_, err = db.Exec(`
		DELETE FROM games
		WHERE status = 'waiting'
	`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup pending challenges on restart: %v", err)
	} else {
		log.Println("✅ Cleared pending challenges from previous session")
	}
	
	// Clear all pending rematch requests - these are stale after restart
	_, err = db.Exec(`
		UPDATE rematch_requests
		SET status = 'expired'
		WHERE status = 'pending'
	`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup pending rematches on restart: %v", err)
	} else {
		log.Println("✅ Cleared pending rematch requests from previous session")
	}
	
	// Clear all online users - everyone is offline after server restart
	_, err = db.Exec(`DELETE FROM online_users`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup online users on restart: %v", err)
	} else {
		log.Println("✅ Cleared online users from previous session")
	}
}

// cleanupOnlineUsers removes stale online user records
func cleanupOnlineUsers() {
	_, err := db.Exec(`
		DELETE FROM online_users 
		WHERE datetime(last_seen_at) < datetime('now', '-5 minutes')
	`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup online users: %v", err)
	}
}

// cleanupExpiredRematches marks expired rematch requests
func cleanupExpiredRematches() {
	_, err := db.Exec(`
		UPDATE rematch_requests
		SET status = 'expired'
		WHERE status = 'pending'
		AND datetime(expires_at) < datetime('now')
	`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup expired rematches: %v", err)
	}
}

// cleanupExpiredChallenges deletes challenge requests older than 30 seconds
func cleanupExpiredChallenges() {
	result, err := db.Exec(`
		DELETE FROM games
		WHERE status = 'waiting'
		AND datetime(created_at) < datetime('now', '-30 seconds')
	`)
	if err != nil {
		log.Printf("Warning: Failed to cleanup expired challenges: %v", err)
		return
	}
	
	rows, _ := result.RowsAffected()
	if rows > 0 {
		log.Printf("🧹 Cleaned up %d expired challenge(s)", rows)
	}
}

// markUserOnline marks a user as online
func markUserOnline(userID int, userName string, inGame bool) error {
	inGameInt := 0
	if inGame {
		inGameInt = 1
	}
	
	if userName == "" {
		// Only update in_game status, don't overwrite username
		_, err := db.Exec(`
			UPDATE online_users 
			SET in_game = ?, last_seen_at = CURRENT_TIMESTAMP
			WHERE user_id = ?
		`, inGameInt, userID)
		return err
	}
	
	_, err := db.Exec(`
		INSERT INTO online_users (user_id, user_name, last_seen_at, in_game)
		VALUES (?, ?, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			user_name = excluded.user_name,
			last_seen_at = CURRENT_TIMESTAMP,
			in_game = excluded.in_game
	`, userID, userName, inGameInt)
	
	return err
}

// updatePlayerStats updates player statistics after a game
func updatePlayerStats(userID int, userName string, won bool, lost bool, draw bool) error {
	// First ensure the player exists in stats
	_, err := db.Exec(`
		INSERT INTO player_stats (user_id, user_name, games_played, games_won, games_lost, games_draw)
		VALUES (?, ?, 0, 0, 0, 0)
		ON CONFLICT(user_id) DO UPDATE SET user_name = excluded.user_name
	`, userID, userName)
	if err != nil {
		return err
	}

	// Update stats
	wonInt, lostInt, drawInt := 0, 0, 0
	if won {
		wonInt = 1
	}
	if lost {
		lostInt = 1
	}
	if draw {
		drawInt = 1
	}

	_, err = db.Exec(`
		UPDATE player_stats 
		SET 
			games_played = games_played + 1,
			games_won = games_won + ?,
			games_lost = games_lost + ?,
			games_draw = games_draw + ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`, wonInt, lostInt, drawInt, userID)

	return err
}

// createRematchRequest creates a new rematch request
func createRematchRequest(gameID, requesterID, opponentID int) (int64, error) {
	// Set expiration to 80 seconds from now (20s + 60s countdown)
	expiresAt := time.Now().Add(80 * time.Second)
	
	result, err := db.Exec(`
		INSERT INTO rematch_requests (game_id, requester_id, opponent_id, expires_at)
		VALUES (?, ?, ?, ?)
	`, gameID, requesterID, opponentID, expiresAt)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

// getRematchRequest gets an active rematch request for a game
func getRematchRequest(gameID int) (*RematchRequest, error) {
	var rm RematchRequest
	var expiresAt sql.NullTime
	
	err := db.QueryRow(`
		SELECT id, game_id, requester_id, opponent_id, status, created_at, expires_at
		FROM rematch_requests
		WHERE game_id = ?
		AND status IN ('pending', 'accepted')
		ORDER BY created_at DESC
		LIMIT 1
	`, gameID).Scan(&rm.ID, &rm.GameID, &rm.RequesterID, &rm.OpponentID, 
		&rm.Status, &rm.CreatedAt, &expiresAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	
	if expiresAt.Valid {
		rm.ExpiresAt = &expiresAt.Time
	}
	
	return &rm, nil
}

// updateRematchStatus updates the status of a rematch request
func updateRematchStatus(rematchID int, status RematchStatus) error {
	_, err := db.Exec(`
		UPDATE rematch_requests
		SET status = ?
		WHERE id = ?
	`, status, rematchID)
	
	return err
}