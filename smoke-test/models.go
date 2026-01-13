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

// Item represents a sample data item (replace with your app's models)
type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
