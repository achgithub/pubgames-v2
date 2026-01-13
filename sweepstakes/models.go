package main

import "time"

// Config represents app configuration
type Config struct {
	AppName    string `json:"app_name"`
	AppIcon    string `json:"app_icon"`
	BackendURL string `json:"backend_url"`
}

// Competition represents a sweepstake competition
type Competition struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"` // "knockout" or "race"
	Status      string     `json:"status"` // "draft", "open", "locked", "completed", "archived"
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Entry represents an item in a competition (horse, team, etc.)
type Entry struct {
	ID            int       `json:"id"`
	CompetitionID int       `json:"competition_id"`
	Name          string    `json:"name"`
	Number        *int      `json:"number"` // For race-type competitions
	Seed          *int      `json:"seed"` // For knockout-type competitions
	Status        string    `json:"status"` // "available", "taken", "active", "eliminated", "winner"
	Position      *int      `json:"position"` // Final position (1st, 2nd, 3rd, etc.)
	CreatedAt     time.Time `json:"created_at"`
}

// Draw represents a user's selection/assignment of an entry
type Draw struct {
	ID            int       `json:"id"`
	UserEmail     string    `json:"user_email"` // From JWT token
	CompetitionID int       `json:"competition_id"`
	EntryID       int       `json:"entry_id"`
	DrawnAt       time.Time `json:"drawn_at"`
	// Joined fields for responses
	EntryName  string `json:"entry_name,omitempty"`
	UserName   string `json:"user_name,omitempty"`
}

// SelectionLock represents an in-memory lock for blind box selection
type SelectionLock struct {
	CompetitionID int       `json:"competition_id"`
	UserEmail     string    `json:"user_email"`
	UserName      string    `json:"user_name"`
	LockedAt      time.Time `json:"locked_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
