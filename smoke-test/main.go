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
	APP_NAME         = "Smoke test"
	APP_ICON         = "🃏"
	BACKEND_PORT     = "30011" // Replace X with app number
	FRONTEND_PORT    = "30010"
	DB_PATH          = "./data/smoke-test.db"
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

	// Public routes
	api.HandleFunc("/config", getConfigHandler).Methods("GET")

	// Protected routes (require authentication)
	authMw := auth.AuthMiddleware(auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
	})
	api.HandleFunc("/data", authMw(getDataHandler)).Methods("GET")
	api.HandleFunc("/items", authMw(getItemsHandler)).Methods("GET")
	api.HandleFunc("/items", authMw(createItemHandler)).Methods("POST")

	// Admin routes (require admin privilege)
	adminMw := auth.AdminMiddleware
	api.HandleFunc("/admin/stats", authMw(adminMw(getAdminStatsHandler))).Methods("GET")

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
	log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler(r)))
}