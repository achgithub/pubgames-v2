package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"pubgames/shared/auth"
	"pubgames/shared/config"
)

var db *sql.DB

// Selection locks - in-memory store for blind box selection
var selectionLocks = make(map[int]*SelectionLock)
var lockMutex sync.Mutex

const (
	APP_NAME         = "Sweepstakes"
	APP_ICON         = "⌨️"
	BACKEND_PORT     = "30031"
	FRONTEND_PORT    = "30030"
	DB_PATH          = "./data/sweepstakes.db"
	IDENTITY_SERVICE = "http://localhost:3001"
)

func main() {
	log.Printf("🚀 Starting %s...", APP_NAME)

	// Initialize database
	initDB()
	defer db.Close()

	// Setup router
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	// Setup auth middleware
	authMw := auth.AuthMiddleware(auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
	})
	adminMw := auth.AdminMiddleware

	// ===== PUBLIC ROUTES =====
	api.HandleFunc("/config", getConfigHandler).Methods("GET")

	// ===== PROTECTED ROUTES (authenticated users) =====
	// Competition routes
	api.HandleFunc("/competitions", authMw(getCompetitionsHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/entries", authMw(getEntriesHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/available-count", authMw(getAvailableCountHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/all-draws", authMw(getCompetitionDrawsHandler)).Methods("GET")
	
	// Blind box selection routes
	api.HandleFunc("/competitions/{id}/blind-boxes", authMw(getBlindBoxesHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/choose-blind-box", authMw(chooseBlindBoxHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/random-pick", authMw(randomPickHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/lock", authMw(acquireSelectionLockHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/unlock", authMw(releaseSelectionLockHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/lock-status", authMw(checkSelectionLockHandler)).Methods("GET")
	
	// User's draws
	api.HandleFunc("/draws", authMw(getUserDrawsHandler)).Methods("GET")

	// ===== ADMIN ROUTES =====
	// Competition management
	api.HandleFunc("/competitions", authMw(adminMw(createCompetitionHandler))).Methods("POST")
	api.HandleFunc("/competitions/{id}", authMw(adminMw(updateCompetitionHandler))).Methods("PUT")
	api.HandleFunc("/competitions/{id}/update-position", authMw(adminMw(updateEntryPositionHandler))).Methods("POST")
	
	// Entry management
	api.HandleFunc("/entries/upload", authMw(adminMw(uploadEntriesHandler))).Methods("POST")
	api.HandleFunc("/entries/{id}", authMw(adminMw(updateEntryHandler))).Methods("PUT")
	api.HandleFunc("/entries/{id}", authMw(adminMw(deleteEntryHandler))).Methods("DELETE")

	// Load CORS configuration from shared config
	corsConfig, err := config.LoadCORSConfig()
	if err != nil {
		log.Printf("Warning: CORS config load error: %v", err)
	}
	log.Printf("📋 CORS Mode: %s", corsConfig.CORS.Mode)
	log.Printf("📋 Allowed Origins: %v", corsConfig.GetAllowedOrigins())

	// Enable CORS using shared config
	corsHandler := handlers.CORS(
		handlers.AllowedOriginValidator(func(origin string) bool {
			allowed := corsConfig.IsOriginAllowed(origin)
			if !allowed {
				log.Printf("❌ CORS blocked: %s", origin)
			}
			return allowed
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(r)

	// Start server
	log.Printf("✅ %s backend running on http://localhost:%s", APP_NAME, BACKEND_PORT)
	log.Printf("🎯 Blind box selection mode enabled")
	log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler))
}
