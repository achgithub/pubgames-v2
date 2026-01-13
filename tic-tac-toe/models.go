package main

import "time"

// User represents a user in the system (from Identity Service)
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// Config represents app configuration
type Config struct {
	AppName              string `json:"app_name"`
	AppIcon              string `json:"app_icon"`
	BackendURL           string `json:"backend_url"`
	DefaultSessionMinutes int    `json:"default_session_minutes"`
	DefaultMoveSeconds   int    `json:"default_move_seconds"`
}

// GameMode represents the type of game
type GameMode string

const (
	GameModeNormal GameMode = "normal" // No time limit per move
	GameModeTimed  GameMode = "timed"  // Time limit per move
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"    // Waiting for opponent to accept
	GameStatusActive     GameStatus = "active"     // Game in progress
	GameStatusCompleted  GameStatus = "completed"  // Game finished
	GameStatusAbandoned  GameStatus = "abandoned"  // Player disconnected/timeout
	GameStatusDeclined   GameStatus = "declined"   // Challenge was declined
)

// RematchStatus represents the status of a rematch request
type RematchStatus string

const (
	RematchStatusPending  RematchStatus = "pending"
	RematchStatusAccepted RematchStatus = "accepted"
	RematchStatusDeclined RematchStatus = "declined"
	RematchStatusExpired  RematchStatus = "expired"
)

// Game represents a tic-tac-toe game session
type Game struct {
	ID              int        `json:"id"`
	Player1ID       int        `json:"player1_id"`
	Player1Name     string     `json:"player1_name"`
	Player2ID       *int       `json:"player2_id"`
	Player2Name     string     `json:"player2_name,omitempty"`
	Mode            GameMode   `json:"mode"`
	Status          GameStatus `json:"status"`
	CurrentTurn     int        `json:"current_turn"`       // 1 or 2
	WinnerID        *int       `json:"winner_id"`
	Board           string     `json:"board"`              // JSON array: ["","","","","","","","",""]
	MoveTimeLimit   int        `json:"move_time_limit"`    // Seconds (0 = no limit)
	SessionTimeout  int        `json:"session_timeout"`    // Minutes
	FirstTo         int        `json:"first_to"`           // First to X wins (1,2,3,5,10,20)
	Player1Score    int        `json:"player1_score"`      // Wins in this series
	Player2Score    int        `json:"player2_score"`      // Wins in this series
	CurrentRound    int        `json:"current_round"`      // Which round in the series
	LastMoveAt      *time.Time `json:"last_move_at"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

// Move represents a single move in a game
type Move struct {
	ID        int       `json:"id"`
	GameID    int       `json:"game_id"`
	PlayerID  int       `json:"player_id"`
	Position  int       `json:"position"`  // 0-8
	Symbol    string    `json:"symbol"`    // "X" or "O"
	CreatedAt time.Time `json:"created_at"`
}

// PlayerStats represents player statistics
type PlayerStats struct {
	UserID       int     `json:"user_id"`
	UserName     string  `json:"user_name"`
	GamesPlayed  int     `json:"games_played"`
	GamesWon     int     `json:"games_won"`
	GamesLost    int     `json:"games_lost"`
	GamesDraw    int     `json:"games_draw"`
	WinRate      float64 `json:"win_rate"`
}

// OnlineUser represents a user currently online
type OnlineUser struct {
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	LastSeenAt time.Time `json:"last_seen_at"`
	InGame     bool      `json:"in_game"`
}

// Challenge represents a game challenge request
type Challenge struct {
	ChallengerID   int      `json:"challenger_id"`
	ChallengerName string   `json:"challenger_name"`
	OpponentID     int      `json:"opponent_id"`
	OpponentName   string   `json:"opponent_name"`
	GameMode       GameMode `json:"game_mode"`
	MoveTimeLimit  int      `json:"move_time_limit"`
	FirstTo        int      `json:"first_to"`
}

// GameSettings represents game configuration
type GameSettings struct {
	Mode          GameMode `json:"mode"`
	MoveTimeLimit int      `json:"move_time_limit"` // 0 = unlimited
	FirstTo       int      `json:"first_to"`        // 1, 2, 3, 5, 10, 20
}

// RematchRequest represents a request to play again
type RematchRequest struct {
	ID          int           `json:"id"`
	GameID      int           `json:"game_id"`
	RequesterID int           `json:"requester_id"`
	OpponentID  int           `json:"opponent_id"`
	Status      RematchStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   *time.Time    `json:"expires_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}