package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// registerHandler creates a new user account
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}

	// Validate input
	if req.Email == "" || req.Name == "" || req.Code == "" {
		sendError(w, "Email, name, and code are required", 400)
		return
	}

	if len(req.Code) != 6 {
		sendError(w, "Code must be exactly 6 characters", 400)
		return
	}

	// Hash the code
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(req.Code), 12)
	if err != nil {
		sendError(w, "Failed to process code", 500)
		return
	}

	// Insert user
	result, err := db.Exec(`
		INSERT INTO users (email, name, code, is_admin) 
		VALUES (?, ?, ?, ?)
	`, req.Email, req.Name, string(hashedCode), req.IsAdmin)

	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.email" {
			sendError(w, "Email already registered", 409)
		} else {
			sendError(w, "Failed to create user", 500)
		}
		return
	}

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"email":   req.Email,
		"name":    req.Name,
		"message": "User registered successfully",
	})
}

// loginHandler authenticates a user and returns a JWT token
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}

	// Find user
	var user User
	var storedCode string
	err := db.QueryRow(`
		SELECT id, email, name, code, is_admin, created_at 
		FROM users 
		WHERE email = ?
	`, req.Email).Scan(&user.ID, &user.Email, &user.Name, &storedCode, &user.IsAdmin, &user.CreatedAt)

	if err == sql.ErrNoRows {
		sendError(w, "Invalid credentials", 401)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}

	// Verify code
	if err := bcrypt.CompareHashAndPassword([]byte(storedCode), []byte(req.Code)); err != nil {
		sendError(w, "Invalid credentials", 401)
		return
	}

	// Generate JWT token
	token, err := generateToken(&user)
	if err != nil {
		sendError(w, "Failed to generate token", 500)
		return
	}

	// Return user data and token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
		User:  user,
	})
}

// validateTokenHandler validates a JWT token and returns user data
func validateTokenHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := extractToken(r)
	if tokenString == "" {
		sendError(w, "Missing authorization header", 401)
		return
	}

	user, err := validateToken(tokenString)
	if err != nil {
		sendError(w, "Invalid or expired token", 401)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// getUserHandler returns current user info
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// getAppsHandler returns list of available apps
func getAppsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, url, description, icon, is_active 
		FROM apps 
		WHERE is_active = 1 
		ORDER BY name
	`)
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer rows.Close()

	apps := []App{}
	for rows.Next() {
		var app App
		err := rows.Scan(&app.ID, &app.Name, &app.URL, &app.Description, &app.Icon, &app.IsActive)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// getAdminAppsHandler returns all apps (including inactive) for admin
func getAdminAppsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, url, description, icon, is_active, created_at 
		FROM apps 
		ORDER BY created_at DESC
	`)
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer rows.Close()

	apps := []App{}
	for rows.Next() {
		var app App
		err := rows.Scan(&app.ID, &app.Name, &app.URL, &app.Description, &app.Icon, &app.IsActive, &app.CreatedAt)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

// createAppHandler creates a new app entry
func createAppHandler(w http.ResponseWriter, r *http.Request) {
	var app App
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}

	result, err := db.Exec(`
		INSERT INTO apps (name, url, description, icon, is_active) 
		VALUES (?, ?, ?, ?, ?)
	`, app.Name, app.URL, app.Description, app.Icon, app.IsActive)

	if err != nil {
		sendError(w, "Failed to create app", 500)
		return
	}

	id, _ := result.LastInsertId()
	app.ID = int(id)
	app.CreatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// getUsersHandler returns all users (admin only)
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, email, name, is_admin, created_at 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.IsAdmin, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
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

// generateToken creates a JWT token for a user
func generateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWT_SECRET))
}
