package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

// Config holds configuration for auth middleware
type Config struct {
	IdentityServiceURL string
}

// User represents authenticated user information
type User struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

// AuthMiddleware validates JWT tokens with Identity Service
func AuthMiddleware(config Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Parse Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendError(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// Validate token with Identity Service
			user, err := validateToken(config.IdentityServiceURL, token)
			if err != nil {
				sendError(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// AdminMiddleware ensures the user has admin privileges
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*User)
		if !ok {
			sendError(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin {
			sendError(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// GetUser retrieves the authenticated user from request context
func GetUser(r *http.Request) *User {
	user, ok := r.Context().Value(UserContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// validateToken validates a JWT token with the Identity Service
func validateToken(identityURL string, token string) (*User, error) {
	req, err := http.NewRequest("GET", identityURL+"/api/validate-token", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, http.ErrAbortHandler
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// sendError sends a JSON error response
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  code,
	})
}
