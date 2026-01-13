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
	AppName    string `json:"app_name"`
	AppIcon    string `json:"app_icon"`
	BackendURL string `json:"backend_url"`
}

// Competition represents a game/tournament
type Competition struct {
	ID               int        `json:"id"`
	Name             string     `json:"name"`
	Status           string     `json:"status"`
	WinnerCount      int        `json:"winner_count"`
	PostponementRule string     `json:"postponement_rule"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Round represents a prediction round in a competition
type Round struct {
	ID                 int       `json:"id"`
	GameID             int       `json:"game_id"`
	RoundNumber        int       `json:"round_number"`
	SubmissionDeadline string    `json:"submission_deadline"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

// Match represents a football match
type Match struct {
	ID          int       `json:"id"`
	GameID      int       `json:"game_id"`
	MatchNumber int       `json:"match_number"`
	RoundNumber int       `json:"round_number"`
	Date        string    `json:"date"`
	Location    string    `json:"location"`
	HomeTeam    string    `json:"home_team"`
	AwayTeam    string    `json:"away_team"`
	Result      string    `json:"result"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Prediction represents a user's prediction for a match
type Prediction struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	GameID        int       `json:"game_id"`
	MatchID       int       `json:"match_id"`
	RoundNumber   int       `json:"round_number"`
	PredictedTeam string    `json:"predicted_team"`
	IsCorrect     *bool     `json:"is_correct"`
	Voided        bool      `json:"voided"`
	CreatedAt     time.Time `json:"created_at"`
}

// PredictionResponse represents a prediction with joined data
type PredictionResponse struct {
	Prediction
	UserName  string `json:"user_name"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	Result    string `json:"result"`
	MatchDate string `json:"match_date"`
}

// StandingsEntry represents a player's standing in the competition
type StandingsEntry struct {
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	IsActive  bool   `json:"is_active"`
	LastRound int    `json:"last_round"`
}

// RoundSummary provides statistics for a round
type RoundSummary struct {
	RoundNumber       int        `json:"round_number"`
	TotalPlayers      int        `json:"total_players"`
	PlayersEliminated int        `json:"players_eliminated"`
	TeamStats         []TeamStat `json:"team_stats"`
}

// TeamStat represents statistics for a team in a round
type TeamStat struct {
	TeamName          string `json:"team_name"`
	PlayerCount       int    `json:"player_count"`
	PlayersEliminated int    `json:"players_eliminated"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
