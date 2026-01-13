package main

import (
	"encoding/json"
	"net/http"

	"pubgames/shared/auth"
)

// Auth Handlers using shared library
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	config := auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
		AdminPassword:      ADMIN_PASSWORD,
	}

	result, err := auth.RegisterUser(config, req.Email, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createAdminHandler(w http.ResponseWriter, r *http.Request) {
	// Admins register same way, promoted in Identity Service
	registerHandler(w, r)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	config := auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
		AdminPassword:      ADMIN_PASSWORD,
	}

	result, err := auth.Login(config, req.Email, req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
