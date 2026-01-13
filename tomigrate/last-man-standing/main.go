package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Code      string    `json:"code,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

type Config struct {
	VenueName string `json:"venue_name"`
	LogoURL   string `json:"logo_url"`
}

type Competition struct {
	ID               int        `json:"id"`
	Name             string     `json:"name"`
	Status           string     `json:"status"`
	WinnerCount      int        `json:"winner_count"`
	PostponementRule string     `json:"postponement_rule"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Round struct {
	ID                 int       `json:"id"`
	GameID             int       `json:"game_id"`
	RoundNumber        int       `json:"round_number"`
	SubmissionDeadline string    `json:"submission_deadline"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}

type Match struct {
	ID          int       `json:"id"`
	GameID      int       `json:"game_id"`
	MatchNumber int       `json:"match_number"`
	RoundNumber int       `json:"round_number"`
	Date        string    `json:"date"`
	Location    string    `json:"location"`
	HomeTeam    string    `json:"home_team"`
	AwayTeam    string    `json:"away_team"`
	Result      string    `json:"result"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Prediction struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	GameID        int       `json:"game_id"`
	MatchID       int       `json:"match_id"`
	RoundNumber   int       `json:"round_number"`
	PredictedTeam string    `json:"predicted_team"`
	IsCorrect     *bool     `json:"is_correct"`
	Voided        bool      `json:"voided"`
	CreatedAt     time.Time `json:"created_at"`
}

type PredictionResponse struct {
	Prediction
	UserName  string `json:"user_name"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	Result    string `json:"result"`
	MatchDate string `json:"match_date"`
}

type StandingsEntry struct {
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
	IsActive  bool   `json:"is_active"`
	LastRound int    `json:"last_round"`
}

type RoundSummary struct {
	RoundNumber       int        `json:"round_number"`
	TotalPlayers      int        `json:"total_players"`
	PlayersEliminated int        `json:"players_eliminated"`
	TeamStats         []TeamStat `json:"team_stats"`
}

type TeamStat struct {
	TeamName          string `json:"team_name"`
	PlayerCount       int    `json:"player_count"`
	PlayersEliminated int    `json:"players_eliminated"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data/lastmanstanding.db")
	if err != nil {
		log.Fatal(err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		status TEXT DEFAULT 'active',
		winner_count INTEGER DEFAULT 0,
		postponement_rule TEXT DEFAULT 'loss',
		start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_date TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS game_players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		game_id INTEGER NOT NULL,
		is_active BOOLEAN DEFAULT 1,
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (game_id) REFERENCES games (id),
		UNIQUE(user_id, game_id)
	);

	CREATE TABLE IF NOT EXISTS rounds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		submission_deadline TEXT NOT NULL,
		status TEXT DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id),
		UNIQUE(game_id, round_number)
	);

	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		match_number INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		date TEXT NOT NULL,
		location TEXT NOT NULL,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		result TEXT DEFAULT '',
		status TEXT DEFAULT 'upcoming',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games (id)
	);

	CREATE TABLE IF NOT EXISTS predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		game_id INTEGER NOT NULL,
		match_id INTEGER NOT NULL,
		round_number INTEGER NOT NULL,
		predicted_team TEXT NOT NULL,
		is_correct BOOLEAN DEFAULT NULL,
		voided BOOLEAN DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (game_id) REFERENCES games (id),
		FOREIGN KEY (match_id) REFERENCES matches (id),
		UNIQUE(user_id, game_id, round_number)
	);

	CREATE TABLE IF NOT EXISTS admin_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		admin_password TEXT NOT NULL,
		current_game_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (current_game_id) REFERENCES games (id)
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	// Add postponement_rule column if it doesn't exist
	db.Exec("ALTER TABLE games ADD COLUMN postponement_rule TEXT DEFAULT 'loss'")
	
	// Add voided column if it doesn't exist
	db.Exec("ALTER TABLE predictions ADD COLUMN voided BOOLEAN DEFAULT 0")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM admin_settings").Scan(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Exec("INSERT INTO admin_settings (admin_password) VALUES (?)", string(hash))
		log.Println("Default admin password: admin123")
	}

	var gameCount int
	db.QueryRow("SELECT COUNT(*) FROM games").Scan(&gameCount)
	if gameCount == 0 {
		result, _ := db.Exec("INSERT INTO games (name, status, postponement_rule) VALUES (?, ?, ?)", "Game 1", "active", "loss")
		gameID, _ := result.LastInsertId()
		db.Exec("UPDATE admin_settings SET current_game_id = ? WHERE id = 1", gameID)
		log.Printf("Created Game 1 (ID: %d)\n", gameID)
	}
}

func generateCode(email string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(email+time.Now().String()), bcrypt.DefaultCost)
	if err != nil {
		hash = []byte(fmt.Sprintf("%x%d", email, time.Now().Unix()))
	}
	hexStr := fmt.Sprintf("%x", hash)
	if len(hexStr) < 8 {
		hexStr = hexStr + "00000000"
	}
	return strings.ToUpper(hexStr[:8])
}

func getCurrentGameID() int {
	var gameID sql.NullInt64
	db.QueryRow("SELECT current_game_id FROM admin_settings WHERE id = 1").Scan(&gameID)
	if gameID.Valid {
		return int(gameID.Int64)
	}
	return 0
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" || user.Name == "" {
		http.Error(w, "Email and name required", http.StatusBadRequest)
		return
	}

	code := generateCode(user.Email)
	result, err := db.Exec("INSERT INTO users (email, name, code, is_admin) VALUES (?, ?, ?, ?)",
		user.Email, user.Name, code, false)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	user.Code = code
	user.IsAdmin = false

	gameID := getCurrentGameID()
	if gameID > 0 {
		db.Exec("INSERT OR IGNORE INTO game_players (user_id, game_id) VALUES (?, ?)", user.ID, gameID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	json.NewDecoder(r.Body).Decode(&credentials)

	var user User
	err := db.QueryRow("SELECT id, email, name, code, is_admin FROM users WHERE email = ? AND code = ?",
		credentials.Email, credentials.Code).Scan(&user.ID, &user.Email, &user.Name, &user.Code, &user.IsAdmin)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func createAdminHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var storedHash string
	db.QueryRow("SELECT admin_password FROM admin_settings WHERE id = 1").Scan(&storedHash)
	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
		http.Error(w, "Invalid admin password", http.StatusUnauthorized)
		return
	}

	code := generateCode(req.Email)
	result, err := db.Exec("INSERT INTO users (email, name, code, is_admin) VALUES (?, ?, ?, ?)",
		req.Email, req.Name, code, true)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	user := User{ID: int(id), Email: req.Email, Name: req.Name, Code: code, IsAdmin: true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, status, winner_count, COALESCE(postponement_rule, 'loss'), start_date, end_date, created_at 
		FROM games ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	games := []Competition{}
	for rows.Next() {
		var g Competition
		var endDate sql.NullTime
		rows.Scan(&g.ID, &g.Name, &g.Status, &g.WinnerCount, &g.PostponementRule, &g.StartDate, &endDate, &g.CreatedAt)
		if endDate.Valid {
			g.EndDate = &endDate.Time
		}
		games = append(games, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

func getCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := getCurrentGameID()
	if gameID == 0 {
		http.Error(w, "No current game", http.StatusNotFound)
		return
	}

	var game Competition
	var endDate sql.NullTime
	err := db.QueryRow(`SELECT id, name, status, winner_count, COALESCE(postponement_rule, 'loss'), start_date, end_date, created_at 
		FROM games WHERE id = ?`, gameID).Scan(&game.ID, &game.Name, &game.Status, &game.WinnerCount,
		&game.PostponementRule, &game.StartDate, &endDate, &game.CreatedAt)

	if err != nil {
		http.Error(w, "Current game not found", http.StatusNotFound)
		return
	}

	if endDate.Valid {
		game.EndDate = &endDate.Time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func createGameHandler(w http.ResponseWriter, r *http.Request) {
	var game struct {
		Name             string `json:"name"`
		PostponementRule string `json:"postponement_rule"`
	}
	json.NewDecoder(r.Body).Decode(&game)

	if game.PostponementRule == "" {
		game.PostponementRule = "loss"
	}

	result, err := db.Exec("INSERT INTO games (name, status, postponement_rule) VALUES (?, ?, ?)",
		game.Name, "active", game.PostponementRule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	resp := Competition{
		ID:               int(id),
		Name:             game.Name,
		Status:           "active",
		PostponementRule: game.PostponementRule,
		StartDate:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func setCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	_, err := db.Exec("UPDATE admin_settings SET current_game_id = ? WHERE id = 1", gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func completeGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	// Check for open rounds
	var openRoundCount int
	db.QueryRow(`SELECT COUNT(*) FROM rounds WHERE game_id = ? AND status = 'open'`, gameID).Scan(&openRoundCount)

	if openRoundCount > 0 {
		http.Error(w, "Cannot complete game: there are still open rounds. Close all rounds first.", http.StatusBadRequest)
		return
	}

	var winnerCount int
	db.QueryRow(`SELECT COUNT(*) FROM game_players WHERE game_id = ? AND is_active = 1`, gameID).Scan(&winnerCount)

	_, err := db.Exec("UPDATE games SET status = 'completed', winner_count = ?, end_date = ? WHERE id = ?",
		winnerCount, time.Now(), gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "winners": winnerCount})
}

func joinGameHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int `json:"user_id"`
		GameID int `json:"game_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	_, err := db.Exec("INSERT OR REPLACE INTO game_players (user_id, game_id, is_active) VALUES (?, ?, ?)",
		req.UserID, req.GameID, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getUserGameStatusHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	gameIDStr := r.URL.Query().Get("game_id")

	userID, _ := strconv.Atoi(userIDStr)
	gameID, _ := strconv.Atoi(gameIDStr)

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	var isActive bool
	err := db.QueryRow(`SELECT is_active FROM game_players WHERE user_id = ? AND game_id = ?`,
		userID, gameID).Scan(&isActive)

	joined := (err == nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"joined":    joined,
		"is_active": isActive,
		"game_id":   gameID,
	})
}

func uploadMatchesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	gameIDStr := r.FormValue("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Error parsing CSV", http.StatusBadRequest)
		return
	}

	count := 0
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 7 {
			continue
		}

		matchNum, _ := strconv.Atoi(record[0])
		roundNum, _ := strconv.Atoi(record[1])
		result := strings.TrimSpace(record[6])

		status := "upcoming"
		if result != "" {
			status = "completed"
		}

		_, err := db.Exec(`INSERT OR REPLACE INTO matches 
			(game_id, match_number, round_number, date, location, home_team, away_team, result, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			gameID, matchNum, roundNum, record[2], record[3], record[4], record[5], result, status)

		if err == nil {
			count++
			if result != "" {
				evaluatePredictionsForMatch(gameID, matchNum, roundNum, record[4], record[5], result)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "uploaded": count, "game_id": gameID})
}

func parseResult(result, homeTeam, awayTeam string) string {
	result = strings.TrimSpace(result)

	// Check for postponement
	if strings.ToUpper(result) == "P - P" || strings.ToUpper(result) == "P-P" {
		return "postponed"
	}

	parts := strings.Split(result, "-")
	if len(parts) != 2 {
		return ""
	}
	home := strings.TrimSpace(parts[0])
	away := strings.TrimSpace(parts[1])

	if home == away {
		return "draw"
	} else if home > away {
		return homeTeam
	}
	return awayTeam
}

func evaluatePredictionsForMatch(gameID, matchNum, roundNum int, homeTeam, awayTeam, result string) {
	var matchID int
	db.QueryRow("SELECT id FROM matches WHERE game_id = ? AND match_number = ? AND round_number = ?",
		gameID, matchNum, roundNum).Scan(&matchID)

	if matchID == 0 {
		return
	}

	winner := parseResult(result, homeTeam, awayTeam)

	// Get game's postponement rule
	var postponementRule string
	db.QueryRow("SELECT COALESCE(postponement_rule, 'loss') FROM games WHERE id = ?", gameID).Scan(&postponementRule)

	// Only mark predictions as correct/incorrect - do NOT eliminate players
	// Elimination happens when round is closed
	if winner == "postponed" {
		// Apply postponement rule
		if postponementRule == "win" {
			// Mark everyone as correct
			db.Exec("UPDATE predictions SET is_correct = 1 WHERE game_id = ? AND match_id = ?", gameID, matchID)
		} else {
			// Mark everyone as incorrect (default)
			db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ?", gameID, matchID)
		}
	} else if winner == "draw" {
		// All predictions are wrong on a draw
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ?", gameID, matchID)
	} else {
		// Mark correct/incorrect based on predicted team
		db.Exec("UPDATE predictions SET is_correct = 1 WHERE game_id = ? AND match_id = ? AND predicted_team = ?",
			gameID, matchID, winner)
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ? AND predicted_team != ?",
			gameID, matchID, winner)
	}
}

func getRoundsHandler(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`SELECT id, game_id, round_number, submission_deadline, status, created_at 
		FROM rounds WHERE game_id = ? ORDER BY round_number ASC`, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rounds := []Round{}
	for rows.Next() {
		var r Round
		rows.Scan(&r.ID, &r.GameID, &r.RoundNumber, &r.SubmissionDeadline, &r.Status, &r.CreatedAt)
		rounds = append(rounds, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rounds)
}

func createRoundHandler(w http.ResponseWriter, r *http.Request) {
	var round Round
	json.NewDecoder(r.Body).Decode(&round)

	if round.GameID == 0 {
		round.GameID = getCurrentGameID()
	}

	result, err := db.Exec(`INSERT INTO rounds (game_id, round_number, submission_deadline, status) 
		VALUES (?, ?, ?, 'draft')`, round.GameID, round.RoundNumber, round.SubmissionDeadline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	round.ID = int(id)
	round.Status = "draft"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(round)
}

func updateRoundStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIDStr := vars["game_id"]
	roundNum := vars["round"]

	gameID, _ := strconv.Atoi(gameIDStr)
	roundNumber, _ := strconv.Atoi(roundNum)

	var update struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&update)

	// If closing a round, check that all matches have results
	if update.Status == "closed" {
		var unmatchedCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM matches 
			WHERE game_id = ? AND round_number = ? AND (result IS NULL OR result = '' OR TRIM(result) = '')`,
			gameID, roundNum).Scan(&unmatchedCount)

		if err != nil {
			http.Error(w, "Error checking match results", http.StatusInternalServerError)
			return
		}

		if unmatchedCount > 0 {
			http.Error(w, fmt.Sprintf("Cannot close round: %d match(es) still have no results entered", unmatchedCount), http.StatusBadRequest)
			return
		}

		// Get players who made incorrect predictions this round
		incorrectPlayers := []int{}
		rows, _ := db.Query(`
			SELECT DISTINCT user_id 
			FROM predictions 
			WHERE game_id = ? AND round_number = ? AND is_correct = 0`,
			gameID, roundNumber)
		for rows.Next() {
			var userID int
			rows.Scan(&userID)
			incorrectPlayers = append(incorrectPlayers, userID)
		}
		rows.Close()

		// Get players who didn't submit at all
		nonSubmitters := []int{}
		rows2, _ := db.Query(`
			SELECT user_id 
			FROM game_players 
			WHERE game_id = ? 
			AND is_active = 1
			AND user_id NOT IN (
				SELECT user_id 
				FROM predictions 
				WHERE game_id = ? AND round_number = ?
			)`, gameID, gameID, roundNumber)
		for rows2.Next() {
			var userID int
			rows2.Scan(&userID)
			nonSubmitters = append(nonSubmitters, userID)
		}
		rows2.Close()

		// Combine both lists
		allEliminatedPlayers := append(incorrectPlayers, nonSubmitters...)

		// Eliminate all these players
		if len(allEliminatedPlayers) > 0 {
			for _, userID := range allEliminatedPlayers {
				// Mark player as inactive
				db.Exec(`UPDATE game_players SET is_active = 0 
					WHERE game_id = ? AND user_id = ?`, gameID, userID)

				// Void all future predictions for this player
				db.Exec(`UPDATE predictions SET voided = 1 
					WHERE game_id = ? AND round_number > ? AND user_id = ?`,
					gameID, roundNumber, userID)
			}
		}
	}

	_, err := db.Exec("UPDATE rounds SET status = ? WHERE game_id = ? AND round_number = ?",
		update.Status, gameID, roundNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getMatchesHandler(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`SELECT id, game_id, match_number, round_number, date, location, 
		home_team, away_team, result, status, created_at FROM matches 
		WHERE game_id = ? ORDER BY round_number, date ASC`, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		rows.Scan(&m.ID, &m.GameID, &m.MatchNumber, &m.RoundNumber, &m.Date, &m.Location,
			&m.HomeTeam, &m.AwayTeam, &m.Result, &m.Status, &m.CreatedAt)
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func getOpenRoundsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	gameIDStr := r.URL.Query().Get("game_id")

	userID, _ := strconv.Atoi(userIDStr)
	gameID, _ := strconv.Atoi(gameIDStr)

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`
		SELECT DISTINCT r.round_number 
		FROM rounds r
		WHERE r.game_id = ? AND r.status = 'open'
		AND r.round_number NOT IN (
			SELECT round_number FROM predictions WHERE user_id = ? AND game_id = ?
		)
		ORDER BY r.round_number ASC
	`, gameID, userID, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	rounds := []int{}
	for rows.Next() {
		var round int
		rows.Scan(&round)
		rounds = append(rounds, round)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rounds)
}

func getMatchesByRoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIDStr := vars["game_id"]
	round := vars["round"]

	gameID, _ := strconv.Atoi(gameIDStr)

	rows, err := db.Query(`SELECT id, game_id, match_number, round_number, date, location, 
		home_team, away_team, result, status, created_at FROM matches 
		WHERE game_id = ? AND round_number = ? ORDER BY date ASC`, gameID, round)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := []Match{}
	for rows.Next() {
		var m Match
		rows.Scan(&m.ID, &m.GameID, &m.MatchNumber, &m.RoundNumber, &m.Date, &m.Location,
			&m.HomeTeam, &m.AwayTeam, &m.Result, &m.Status, &m.CreatedAt)
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func makePredictionHandler(w http.ResponseWriter, r *http.Request) {
	var pred struct {
		UserID        int    `json:"user_id"`
		GameID        int    `json:"game_id"`
		MatchID       int    `json:"match_id"`
		PredictedTeam string `json:"predicted_team"`
	}
	json.NewDecoder(r.Body).Decode(&pred)

	if pred.GameID == 0 {
		pred.GameID = getCurrentGameID()
	}

	var isActive bool
	db.QueryRow("SELECT is_active FROM game_players WHERE user_id = ? AND game_id = ?",
		pred.UserID, pred.GameID).Scan(&isActive)
	if !isActive {
		http.Error(w, "User is eliminated from this game", http.StatusForbidden)
		return
	}

	var roundNum int
	var roundStatus string
	var submissionDeadline string
	db.QueryRow(`SELECT m.round_number, COALESCE(r.status, 'draft'), COALESCE(r.submission_deadline, '')
		FROM matches m 
		LEFT JOIN rounds r ON m.game_id = r.game_id AND m.round_number = r.round_number 
		WHERE m.id = ? AND m.game_id = ?`, pred.MatchID, pred.GameID).Scan(&roundNum, &roundStatus, &submissionDeadline)

	if roundStatus != "open" {
		http.Error(w, "Round is not open for predictions", http.StatusBadRequest)
		return
	}

	// Check if deadline has passed
	if submissionDeadline != "" {
		deadline, err := time.Parse("2006-01-02T15:04:05Z", submissionDeadline)
		if err != nil {
			// Try alternative format
			deadline, err = time.Parse("2006-01-02 15:04:05", submissionDeadline)
		}
		if err == nil && time.Now().After(deadline) {
			http.Error(w, "Submission deadline has passed", http.StatusBadRequest)
			return
		}
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM predictions WHERE user_id = ? AND game_id = ? AND round_number = ?",
		pred.UserID, pred.GameID, roundNum).Scan(&count)
	if count > 0 {
		http.Error(w, "Already predicted for this round", http.StatusConflict)
		return
	}

	// Check if user has already picked this team in this game
	var teamCount int
	db.QueryRow("SELECT COUNT(*) FROM predictions WHERE user_id = ? AND game_id = ? AND predicted_team = ?",
		pred.UserID, pred.GameID, pred.PredictedTeam).Scan(&teamCount)
	if teamCount > 0 {
		http.Error(w, "You have already picked this team in this game", http.StatusConflict)
		return
	}

	_, err := db.Exec(`INSERT INTO predictions (user_id, game_id, match_id, round_number, predicted_team) 
		VALUES (?, ?, ?, ?, ?)`, pred.UserID, pred.GameID, pred.MatchID, roundNum, pred.PredictedTeam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func updateMatchResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID := vars["id"]

	var update struct {
		Result string `json:"result"`
	}
	json.NewDecoder(r.Body).Decode(&update)

	var homeTeam, awayTeam string
	var roundNum, gameID int
	db.QueryRow("SELECT home_team, away_team, round_number, game_id FROM matches WHERE id = ?", matchID).
		Scan(&homeTeam, &awayTeam, &roundNum, &gameID)

	_, err := db.Exec("UPDATE matches SET result = ?, status = 'completed' WHERE id = ?",
		update.Result, matchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	winner := parseResult(update.Result, homeTeam, awayTeam)

	// Get game's postponement rule
	var postponementRule string
	db.QueryRow("SELECT COALESCE(postponement_rule, 'loss') FROM games WHERE id = ?", gameID).Scan(&postponementRule)

	mID, _ := strconv.Atoi(matchID)

	// Only mark predictions as correct/incorrect - do NOT eliminate players
	// Elimination happens when round is closed
	if winner == "postponed" {
		// Apply postponement rule
		if postponementRule == "win" {
			// Mark everyone as correct
			db.Exec("UPDATE predictions SET is_correct = 1 WHERE match_id = ?", mID)
		} else {
			// Mark everyone as incorrect (default)
			db.Exec("UPDATE predictions SET is_correct = 0 WHERE match_id = ?", mID)
		}
	} else if winner == "draw" {
		// All predictions are wrong on a draw
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE match_id = ?", mID)
	} else {
		// Mark correct/incorrect based on predicted team
		db.Exec("UPDATE predictions SET is_correct = 1 WHERE match_id = ? AND predicted_team = ?",
			mID, winner)
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE match_id = ? AND predicted_team != ?",
			mID, winner)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getStandingsHandler(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`
		SELECT u.id, u.name, gp.is_active, 
			COALESCE(MAX(p.round_number), 0) as last_round
		FROM users u
		JOIN game_players gp ON u.id = gp.user_id
		LEFT JOIN predictions p ON u.id = p.user_id AND p.game_id = gp.game_id
		WHERE u.is_admin = 0 AND gp.game_id = ?
		GROUP BY u.id
		ORDER BY gp.is_active DESC, last_round DESC
	`, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	standings := []StandingsEntry{}
	for rows.Next() {
		var s StandingsEntry
		rows.Scan(&s.UserID, &s.UserName, &s.IsActive, &s.LastRound)
		standings = append(standings, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}

func getRoundSummaryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIDStr := vars["game_id"]
	roundNum := vars["round"]

	gameID, _ := strconv.Atoi(gameIDStr)

	var totalPlayers int
	db.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM predictions 
		WHERE game_id = ? AND round_number = ? AND voided = 0`,
		gameID, roundNum).Scan(&totalPlayers)

	var playersEliminated int
	db.QueryRow(`SELECT COUNT(DISTINCT p.user_id) 
		FROM predictions p 
		WHERE p.game_id = ? AND p.round_number = ? AND p.is_correct = 0 AND p.voided = 0`, 
		gameID, roundNum).Scan(&playersEliminated)

	rows, err := db.Query(`
		SELECT 
			p.predicted_team,
			COUNT(p.user_id) as player_count,
			SUM(CASE WHEN p.is_correct = 0 THEN 1 ELSE 0 END) as eliminated
		FROM predictions p
		WHERE p.game_id = ? AND p.round_number = ? AND p.voided = 0
		GROUP BY p.predicted_team
		ORDER BY player_count DESC
	`, gameID, roundNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teamStats := []TeamStat{}
	for rows.Next() {
		var ts TeamStat
		rows.Scan(&ts.TeamName, &ts.PlayerCount, &ts.PlayersEliminated)
		teamStats = append(teamStats, ts)
	}

	rNum, _ := strconv.Atoi(roundNum)
	summary := RoundSummary{
		RoundNumber:       rNum,
		TotalPlayers:      totalPlayers,
		PlayersEliminated: playersEliminated,
		TeamStats:         teamStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func getPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	gameIDStr := r.URL.Query().Get("game_id")
	viewAllStr := r.URL.Query().Get("view_all")

	userID, _ := strconv.Atoi(userIDStr)
	gameID, _ := strconv.Atoi(gameIDStr)
	viewAll := viewAllStr == "true"

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	var isAdmin bool
	db.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)

	// If trying to view all predictions, must be admin
	if viewAll && !isAdmin {
		http.Error(w, "Only admins can view all predictions", http.StatusForbidden)
		return
	}

	var rows *sql.Rows
	var err error

	if viewAll && isAdmin {
		// Admin viewing all predictions
		rows, err = db.Query(`
			SELECT p.id, p.user_id, p.match_id, p.round_number, p.predicted_team, 
				p.is_correct, p.voided, p.created_at, u.name, m.home_team, m.away_team, m.result, m.date
			FROM predictions p
			JOIN users u ON p.user_id = u.id
			JOIN matches m ON p.match_id = m.id
			WHERE p.game_id = ?
			ORDER BY p.round_number DESC, p.created_at DESC
		`, gameID)
	} else {
		// User viewing their own predictions
		rows, err = db.Query(`
			SELECT p.id, p.user_id, p.match_id, p.round_number, p.predicted_team, 
				p.is_correct, p.voided, p.created_at, u.name, m.home_team, m.away_team, m.result, m.date
			FROM predictions p
			JOIN users u ON p.user_id = u.id
			JOIN matches m ON p.match_id = m.id
			WHERE p.game_id = ? AND p.user_id = ?
			ORDER BY p.round_number DESC, p.created_at DESC
		`, gameID, userID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	predictions := []PredictionResponse{}
	for rows.Next() {
		var p PredictionResponse
		var isCorrect sql.NullBool
		rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.RoundNumber, &p.PredictedTeam,
			&isCorrect, &p.Voided, &p.CreatedAt, &p.UserName, &p.HomeTeam, &p.AwayTeam, &p.Result, &p.MatchDate)
		if isCorrect.Valid {
			val := isCorrect.Bool
			p.IsCorrect = &val
		}
		predictions = append(predictions, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(predictions)
}

func getUsedTeamsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	gameIDStr := r.URL.Query().Get("game_id")

	userID, _ := strconv.Atoi(userIDStr)
	gameID, _ := strconv.Atoi(gameIDStr)

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`
		SELECT DISTINCT predicted_team 
		FROM predictions 
		WHERE user_id = ? AND game_id = ?
	`, userID, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	teams := []string{}
	for rows.Next() {
		var team string
		rows.Scan(&team)
		teams = append(teams, team)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Try to read config file, fallback to defaults
	config := Config{
		VenueName: "Football Prediction - Last Man Standing",
		LogoURL:   "",
	}

	// Try to read from ../data/config.json
	file, err := os.ReadFile("../data/config.json")
	if err == nil {
		json.Unmarshal(file, &config)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:30010"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	r.HandleFunc("/api/register", registerHandler).Methods("POST")
	r.HandleFunc("/api/register/admin", createAdminHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/config", getConfigHandler).Methods("GET")

	r.HandleFunc("/api/games/current", getCurrentGameHandler).Methods("GET")
	r.HandleFunc("/api/games/join", joinGameHandler).Methods("POST")
	r.HandleFunc("/api/games/status", getUserGameStatusHandler).Methods("GET")
	r.HandleFunc("/api/games/{id}/set-current", setCurrentGameHandler).Methods("PUT")
	r.HandleFunc("/api/games/{id}/complete", completeGameHandler).Methods("PUT")
	r.HandleFunc("/api/games", getGamesHandler).Methods("GET")
	r.HandleFunc("/api/games", createGameHandler).Methods("POST")

	r.HandleFunc("/api/rounds/open", getOpenRoundsHandler).Methods("GET")
	r.HandleFunc("/api/rounds/{game_id}/{round}/status", updateRoundStatusHandler).Methods("PUT")
	r.HandleFunc("/api/rounds/{game_id}/{round}/summary", getRoundSummaryHandler).Methods("GET")
	r.HandleFunc("/api/rounds", getRoundsHandler).Methods("GET")
	r.HandleFunc("/api/rounds", createRoundHandler).Methods("POST")

	r.HandleFunc("/api/matches/upload", uploadMatchesHandler).Methods("POST")
	r.HandleFunc("/api/matches/{game_id}/round/{round}", getMatchesByRoundHandler).Methods("GET")
	r.HandleFunc("/api/matches/{id}/result", updateMatchResultHandler).Methods("PUT")
	r.HandleFunc("/api/matches", getMatchesHandler).Methods("GET")

	r.HandleFunc("/api/predictions/used-teams", getUsedTeamsHandler).Methods("GET")
	r.HandleFunc("/api/predictions", makePredictionHandler).Methods("POST")
	r.HandleFunc("/api/predictions", getPredictionsHandler).Methods("GET")

	r.HandleFunc("/api/standings", getStandingsHandler).Methods("GET")

	port := ":30011"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, corsHandler(r)))
}