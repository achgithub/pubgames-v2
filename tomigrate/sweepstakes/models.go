package main

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Code      string    `json:"code,omitempty"`
	Role      string    `json:"role,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

type Competition struct {
	ID               int        `json:"id"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Status           string     `json:"status"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Description      string     `json:"description"`
	SelectionMode    string     `json:"selection_mode"`
	BlindBoxInterval int        `json:"blind_box_interval"`
	CreatedAt        time.Time  `json:"created_at"`
}

type SelectionLock struct {
	CompetitionID int       `json:"competition_id"`
	UserID        int       `json:"user_id"`
	UserName      string    `json:"user_name"`
	LockedAt      time.Time `json:"locked_at"`
}

type Entry struct {
	ID            int       `json:"id"`
	CompetitionID int       `json:"competition_id"`
	Name          string    `json:"name"`
	Number        *int      `json:"number"`
	Seed          *int      `json:"seed"`
	Status        string    `json:"status"`
	Position      int       `json:"position"`
	CreatedAt     time.Time `json:"created_at"`
}

type Draw struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	CompetitionID int       `json:"competition_id"`
	EntryID       int       `json:"entry_id"`
	DrawnAt       time.Time `json:"drawn_at"`
	EntryName     string    `json:"entry_name"`
	UserName      string    `json:"user_name"`
}

type Config struct {
	VenueName string `json:"venue_name"`
	LogoURL   string `json:"logo_url"`
}
