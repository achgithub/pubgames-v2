package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"pubgames/shared/auth"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Selection locks - in-memory store
var selectionLocks = make(map[int]*SelectionLock)
var lockMutex sync.Mutex

// Configuration
var (
	BACKEND_PORT = getEnv("BACKEND_PORT", "30021")
	FRONTEND_PORT = getEnv("FRONTEND_PORT", "30020")
	DB_PATH          = getEnv("DB_PATH", "./data/sweepstake.db")
	IDENTITY_SERVICE = getEnv("IDENTITY_SERVICE", "http://localhost:3001")
	ADMIN_PASSWORD   = getEnv("ADMIN_PASSWORD", "backdoor123")
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("🚀 Sweepstake Application Starting...")
	log.Printf("🔗 Identity Service: %s", IDENTITY_SERVICE)
	log.Printf("🔑 Admin backdoor enabled with password: %s", ADMIN_PASSWORD)

	checkPort(BACKEND_PORT, "Backend")

	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Public routes (no auth required)
	api.HandleFunc("/register", registerHandler).Methods("POST")
	api.HandleFunc("/register/admin", createAdminHandler).Methods("POST")
	api.HandleFunc("/login", loginHandler).Methods("POST")
	api.HandleFunc("/config", getConfigHandler).Methods("GET")

	// Create auth middleware
	authMw := auth.AuthMiddleware(auth.Config{
		IdentityServiceURL: IDENTITY_SERVICE,
		AdminPassword:      ADMIN_PASSWORD,
	})

	// Protected routes - require authentication
	api.HandleFunc("/competitions", authMw(getCompetitionsHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/entries", authMw(getEntriesHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/available-count", authMw(getAvailableCountHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/blind-boxes", authMw(getBlindBoxesHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/choose-blind-box", authMw(chooseBlindBoxHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/random-pick", authMw(randomPickHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/lock", authMw(acquireSelectionLockHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/unlock", authMw(releaseSelectionLockHandler)).Methods("POST")
	api.HandleFunc("/competitions/{id}/lock-status", authMw(checkSelectionLockHandler)).Methods("GET")
	api.HandleFunc("/competitions/{id}/all-draws", authMw(getCompetitionDrawsHandler)).Methods("GET")
	api.HandleFunc("/draws", authMw(getUserDrawsHandler)).Methods("GET")

	// Admin routes - require admin privileges
	adminMw := auth.AdminMiddleware
	api.HandleFunc("/competitions", authMw(adminMw(createCompetitionHandler))).Methods("POST")
	api.HandleFunc("/competitions/{id}", authMw(adminMw(updateCompetitionHandler))).Methods("PUT")
	api.HandleFunc("/competitions/{id}/update-position", authMw(adminMw(updateEntryPositionHandler))).Methods("POST")
	api.HandleFunc("/entries/upload", authMw(adminMw(uploadEntriesHandler))).Methods("POST")
	api.HandleFunc("/entries/{id}", authMw(adminMw(updateEntryHandler))).Methods("PUT")
	api.HandleFunc("/entries/{id}", authMw(adminMw(deleteEntryHandler))).Methods("DELETE")

	// CORS configuration
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{
			fmt.Sprintf("http://localhost:%s", FRONTEND_PORT),
			"http://localhost:3000",
			"http://localhost:3002",
			"http://localhost:30000",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Admin-Override"}),
		handlers.AllowCredentials(),
	)

	log.Printf("✅ Server starting on port %s", BACKEND_PORT)
	log.Printf("📡 CORS enabled for frontend on port %s", FRONTEND_PORT)
	log.Printf("🎯 Blind box selection mode enabled!")
	log.Printf("🔐 Using shared authentication library")
	log.Fatal(http.ListenAndServe(":"+BACKEND_PORT, corsHandler(r)))
}

func checkPort(port, name string) {
	conn, err := net.Listen("tcp", ":"+port)
	if err == nil {
		conn.Close()
		return
	}

	log.Printf("⚠️  Port %s (%s) is already in use!", port, name)
	log.Print("Kill the process? (y/n): ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		log.Fatal("Please free up the port and restart")
	}

	killPort(port)
}

func killPort(port string) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", fmt.Sprintf("FOR /F \"tokens=5\" %%P IN ('netstat -ano ^| findstr :%s') DO taskkill /F /PID %%P", port))
	} else {
		pidBytes, err := exec.Command("lsof", "-ti:"+port).Output()
		if err == nil && len(pidBytes) > 0 {
			pid := strings.TrimSpace(string(pidBytes))
			exec.Command("kill", "-9", pid).Run()
		}
	}

	if cmd != nil {
		cmd.Run()
	}

	log.Printf("✅ Killed process on port %s", port)
	time.Sleep(1 * time.Second)
}
