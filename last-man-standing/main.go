package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"pubgames/shared/auth"
	"pubgames/shared/config"
)

var db *sql.DB

const (
	APP_NAME         = "Last Man Standing"
	APP_ICON         = "⚽"
	BACKEND_PORT     = "30021"
	FRONTEND_PORT    = "30020"
	DB_PATH          = "./data/last-man-standing.db"
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
	api.HandleFunc("/games/current", getCurrentGameHandler).Methods("GET")

	// ===== PROTECTED ROUTES (authenticated users) =====
	
	// Game routes
	api.HandleFunc("/games", authMw(getGamesHandler)).Methods("GET")
	api.HandleFunc("/games/join", authMw(joinGameHandler)).Methods("POST")
	api.HandleFunc("/games/status", authMw(getUserGameStatusHandler)).Methods("GET")

	// Round routes
	api.HandleFunc("/rounds", authMw(getRoundsHandler)).Methods("GET")
	api.HandleFunc("/rounds/open", authMw(getOpenRoundsHandler)).Methods("GET")

	// Match routes
	api.HandleFunc("/matches", authMw(getMatchesHandler)).Methods("GET")
	api.HandleFunc("/matches/{game_id}/round/{round}", authMw(getMatchesByRoundHandler)).Methods("GET")

	// Prediction routes
	api.HandleFunc("/predictions", authMw(makePredictionHandler)).Methods("POST")
	api.HandleFunc("/predictions", authMw(getPredictionsHandler)).Methods("GET")
	api.HandleFunc("/predictions/used-teams", authMw(getUsedTeamsHandler)).Methods("GET")

	// Standings routes
	api.HandleFunc("/standings", authMw(getStandingsHandler)).Methods("GET")

	// ===== ADMIN ROUTES (require admin privilege) =====
	
	// Game admin routes
	api.HandleFunc("/games", authMw(adminMw(createGameHandler))).Methods("POST")
	api.HandleFunc("/games/{id}/set-current", authMw(adminMw(setCurrentGameHandler))).Methods("PUT")
	api.HandleFunc("/games/{id}/complete", authMw(adminMw(completeGameHandler))).Methods("PUT")

	// Round admin routes
	api.HandleFunc("/rounds", authMw(adminMw(createRoundHandler))).Methods("POST")
	api.HandleFunc("/rounds/{game_id}/{round}/status", authMw(adminMw(updateRoundStatusHandler))).Methods("PUT")
	api.HandleFunc("/rounds/{game_id}/{round}/summary", authMw(adminMw(getRoundSummaryHandler))).Methods("GET")

	// Match admin routes
	api.HandleFunc("/matches/upload", authMw(adminMw(uploadMatchesHandler))).Methods("POST")
	api.HandleFunc("/matches/{id}/result", authMw(adminMw(updateMatchResultHandler))).Methods("PUT")

	// Load CORS configuration from shared config
	corsConfig, err := config.LoadCORSConfig()
	if err != nil {
		log.Printf("Warning: CORS config load error: %v", err)
	}
	log.Printf("📋 CORS Mode: %s", corsConfig.CORS.Mode)
	log.Printf("📋 Allowed Origins: %v", corsConfig.GetAllowedOrigins())

	// CORS configuration using shared config
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
	)

	log.Printf("✅ Backend running on :%s", BACKEND_PORT)
	log.Printf("   Frontend should be at :%s", FRONTEND_PORT)
	log.Printf("   Identity Service at %s", IDENTITY_SERVICE)
	log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler(r)))
}
