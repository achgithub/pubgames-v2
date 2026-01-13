package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// getConfigHandler returns app configuration (public endpoint)
func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := Config{
		AppName:    APP_NAME,
		AppIcon:    APP_ICON,
		BackendURL: "http://localhost:" + BACKEND_PORT,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// getDataHandler returns sample data (protected endpoint)
func getDataHandler(w http.ResponseWriter, r *http.Request) {
	// User is authenticated - user info available in context if needed
	data := map[string]interface{}{
		"message":   "This is protected data",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getItemsHandler returns all items (protected endpoint)
func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, description, created_at 
		FROM items 
		ORDER BY created_at DESC
	`)
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// createItemHandler creates a new item (protected endpoint)
func createItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}

	result, err := db.Exec(`
		INSERT INTO items (name, description) 
		VALUES (?, ?)
	`, item.Name, item.Description)
	if err != nil {
		sendError(w, "Failed to create item", 500)
		return
	}

	id, _ := result.LastInsertId()
	item.ID = int(id)
	item.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// getAdminStatsHandler returns admin statistics (admin only endpoint)
func getAdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	var itemCount int
	db.QueryRow("SELECT COUNT(*) FROM items").Scan(&itemCount)

	stats := map[string]interface{}{
		"total_items": itemCount,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// sendError sends a JSON error response
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}