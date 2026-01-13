package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"pubgames/shared/config"
)

var db *sql.DB

const (
	BACKEND_PORT  = "3001"
	FRONTEND_PORT = "30000"
	DB_PATH       = "./data/identity.db"
	JWT_SECRET    = "your-secret-key-change-in-production" // TODO: Use environment variable
)

func main() {
	log.Println("🚀 Starting PubGames Identity Service...")

	// Initialize database
	initDB()
	defer db.Close()

	// Setup router
	r := mux.NewRouter()

	// Serve static files (shared CSS)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Public routes
	api.HandleFunc("/register", registerHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/login", loginHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/apps", getAppsHandler).Methods("GET")
	api.HandleFunc("/server-info", getServerInfoHandler).Methods("GET")

	// Protected routes
	api.HandleFunc("/validate-token", validateTokenHandler).Methods("GET")
	api.HandleFunc("/user", authMiddleware(getUserHandler)).Methods("GET")

	// Admin routes
	api.HandleFunc("/admin/apps", authMiddleware(adminMiddleware(getAdminAppsHandler))).Methods("GET")
	api.HandleFunc("/admin/apps", authMiddleware(adminMiddleware(createAppHandler))).Methods("POST")
	api.HandleFunc("/admin/users", authMiddleware(adminMiddleware(getUsersHandler))).Methods("GET")

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

	log.Printf("✅ Identity Service backend running on :%s", BACKEND_PORT)
	log.Printf("   Frontend should be at :%s", FRONTEND_PORT)
	log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler(r)))
}
