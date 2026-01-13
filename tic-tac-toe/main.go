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
	APP_NAME         = "Tic Tac Toe"
	APP_ICON         = "📤"
	BACKEND_PORT     = "30041"
	FRONTEND_PORT    = "30040"
	DB_PATH          = "./data/tic-tac-toe.db"
	IDENTITY_SERVICE = "http://localhost:3001"
)

func main() {
	log.Printf("🚀 Starting %s...", APP_NAME)
	initDB()
	defer db.Close()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	authMw := auth.AuthMiddleware(auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
	})

	api.HandleFunc("/config", getConfigHandler).Methods("GET")
	api.HandleFunc("/heartbeat", authMw(heartbeatHandler)).Methods("POST")
	api.HandleFunc("/logout", authMw(logoutHandler)).Methods("POST")
	api.HandleFunc("/online-users", authMw(getOnlineUsersHandler)).Methods("GET")
	api.HandleFunc("/game/active", authMw(getActiveGameHandler)).Methods("GET")
	api.HandleFunc("/game/create-challenge", authMw(createChallengeHandler)).Methods("POST")
	api.HandleFunc("/game/pending-challenges", authMw(getPendingChallengesHandler)).Methods("GET")
	api.HandleFunc("/game/{id}/respond", authMw(respondToChallengeHandler)).Methods("POST")
	api.HandleFunc("/game/move", authMw(makeMoveHandler)).Methods("POST")
	
	// WebSocket endpoints (handle auth internally via query param)
	api.HandleFunc("/ws/lobby", lobbyWebSocketHandler).Methods("GET")
	api.HandleFunc("/ws/game/{gameId}", gameWebSocketHandler).Methods("GET")
	
	api.HandleFunc("/game/rematch", authMw(createRematchHandler)).Methods("POST")
	api.HandleFunc("/game/rematch/{gameId}", authMw(getRematchHandler)).Methods("GET")
	api.HandleFunc("/game/rematch/{id}/respond", authMw(respondToRematchHandler)).Methods("POST")
	api.HandleFunc("/stats/player", authMw(getPlayerStatsHandler)).Methods("GET")
	api.HandleFunc("/stats/leaderboard", authMw(getLeaderboardHandler)).Methods("GET")
	api.HandleFunc("/history", authMw(getGameHistoryHandler)).Methods("GET")

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
