package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"pubgames/shared/auth"
)

// sendError sends a JSON error response
func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// getConfigHandler returns app configuration (public endpoint)
func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := Config{
		AppName:    APP_NAME,
		AppIcon:    APP_ICON,
		BackendURL: "http://localhost:" + BACKEND_PORT,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// === GAME HANDLERS ===

// getGamesHandler returns all games
func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, status, winner_count, COALESCE(postponement_rule, 'loss'), 
		start_date, end_date, created_at FROM games ORDER BY created_at DESC`)
	if err != nil {
		sendError(w, err.Error(), 500)
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

// getCurrentGameHandler returns the current active game
func getCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := getCurrentGameID()
	if gameID == 0 {
		sendError(w, "No current game", 404)
		return
	}

	var game Competition
	var endDate sql.NullTime
	err := db.QueryRow(`SELECT id, name, status, winner_count, COALESCE(postponement_rule, 'loss'), 
		start_date, end_date, created_at FROM games WHERE id = ?`, gameID).Scan(
		&game.ID, &game.Name, &game.Status, &game.WinnerCount,
		&game.PostponementRule, &game.StartDate, &endDate, &game.CreatedAt)

	if err != nil {
		sendError(w, "Current game not found", 404)
		return
	}

	if endDate.Valid {
		game.EndDate = &endDate.Time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

// createGameHandler creates a new game (admin only)
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
		sendError(w, err.Error(), 500)
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

// setCurrentGameHandler sets the current active game (admin only)
func setCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	_, err := db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('current_game_id', ?)", gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// completeGameHandler marks a game as completed (admin only)
func completeGameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]

	// Check for open rounds
	var openRoundCount int
	db.QueryRow(`SELECT COUNT(*) FROM rounds WHERE game_id = ? AND status = 'open'`, gameID).Scan(&openRoundCount)

	if openRoundCount > 0 {
		sendError(w, "Cannot complete game: there are still open rounds. Close all rounds first.", 400)
		return
	}

	var winnerCount int
	db.QueryRow(`SELECT COUNT(*) FROM game_players WHERE game_id = ? AND is_active = 1`, gameID).Scan(&winnerCount)

	_, err := db.Exec("UPDATE games SET status = 'completed', winner_count = ?, end_date = ? WHERE id = ?",
		winnerCount, time.Now(), gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "winners": winnerCount})
}

// joinGameHandler allows a user to join a game
func joinGameHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	var req struct {
		GameID int `json:"game_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.GameID == 0 {
		req.GameID = getCurrentGameID()
	}

	_, err := db.Exec("INSERT OR REPLACE INTO game_players (user_id, game_id, is_active) VALUES (?, ?, ?)",
		user.ID, req.GameID, true)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getUserGameStatusHandler returns user's status in a game
func getUserGameStatusHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	var isActive bool
	err := db.QueryRow(`SELECT is_active FROM game_players WHERE user_id = ? AND game_id = ?`,
		user.ID, gameID).Scan(&isActive)

	joined := (err == nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"joined":    joined,
		"is_active": isActive,
		"game_id":   gameID,
	})
}

// === ROUND HANDLERS ===

// getRoundsHandler returns all rounds for a game
func getRoundsHandler(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`SELECT id, game_id, round_number, submission_deadline, status, created_at 
		FROM rounds WHERE game_id = ? ORDER BY round_number ASC`, gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
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

// createRoundHandler creates a new round (admin only)
func createRoundHandler(w http.ResponseWriter, r *http.Request) {
	var round Round
	json.NewDecoder(r.Body).Decode(&round)

	if round.GameID == 0 {
		round.GameID = getCurrentGameID()
	}

	result, err := db.Exec(`INSERT INTO rounds (game_id, round_number, submission_deadline, status) 
		VALUES (?, ?, ?, 'draft')`, round.GameID, round.RoundNumber, round.SubmissionDeadline)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	id, _ := result.LastInsertId()
	round.ID = int(id)
	round.Status = "draft"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(round)
}

// updateRoundStatusHandler updates round status (admin only)
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
			sendError(w, "Error checking match results", 500)
			return
		}

		if unmatchedCount > 0 {
			sendError(w, fmt.Sprintf("Cannot close round: %d match(es) still have no results entered", unmatchedCount), 400)
			return
		}

		// Eliminate players with incorrect predictions
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

		// Eliminate players who didn't submit
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

		// Combine and eliminate
		allEliminatedPlayers := append(incorrectPlayers, nonSubmitters...)
		if len(allEliminatedPlayers) > 0 {
			for _, userID := range allEliminatedPlayers {
				db.Exec(`UPDATE game_players SET is_active = 0 
					WHERE game_id = ? AND user_id = ?`, gameID, userID)
				db.Exec(`UPDATE predictions SET voided = 1 
					WHERE game_id = ? AND round_number > ? AND user_id = ?`,
					gameID, roundNumber, userID)
			}
		}
	}

	_, err := db.Exec("UPDATE rounds SET status = ? WHERE game_id = ? AND round_number = ?",
		update.Status, gameID, roundNum)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getOpenRoundsHandler returns rounds user can still predict for
func getOpenRoundsHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	
	gameIDStr := r.URL.Query().Get("game_id")
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
	`, gameID, user.ID, gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
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

// getRoundSummaryHandler returns round statistics (admin only)
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
		sendError(w, err.Error(), 500)
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

// === MATCH HANDLERS ===

// getMatchesHandler returns all matches for a game
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
		sendError(w, err.Error(), 500)
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

// getMatchesByRoundHandler returns matches for a specific round
func getMatchesByRoundHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIDStr := vars["game_id"]
	round := vars["round"]

	gameID, _ := strconv.Atoi(gameIDStr)

	rows, err := db.Query(`SELECT id, game_id, match_number, round_number, date, location, 
		home_team, away_team, result, status, created_at FROM matches 
		WHERE game_id = ? AND round_number = ? ORDER BY date ASC`, gameID, round)
	if err != nil {
		sendError(w, err.Error(), 500)
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

// uploadMatchesHandler uploads matches from CSV (admin only)
func uploadMatchesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	gameIDStr := r.FormValue("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, "Error reading file", 400)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		sendError(w, "Error parsing CSV", 400)
		return
	}

	count := 0
	for i, record := range records {
		if i == 0 {
			continue // Skip header
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

// updateMatchResultHandler updates a match result (admin only)
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
		sendError(w, err.Error(), 500)
		return
	}

	mID, _ := strconv.Atoi(matchID)
	evaluatePredictionsForMatch(gameID, mID, roundNum, homeTeam, awayTeam, update.Result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// parseResult determines winner from match result
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

// evaluatePredictionsForMatch marks predictions as correct/incorrect
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

	if winner == "postponed" {
		if postponementRule == "win" {
			db.Exec("UPDATE predictions SET is_correct = 1 WHERE game_id = ? AND match_id = ?", gameID, matchID)
		} else {
			db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ?", gameID, matchID)
		}
	} else if winner == "draw" {
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ?", gameID, matchID)
	} else {
		db.Exec("UPDATE predictions SET is_correct = 1 WHERE game_id = ? AND match_id = ? AND predicted_team = ?",
			gameID, matchID, winner)
		db.Exec("UPDATE predictions SET is_correct = 0 WHERE game_id = ? AND match_id = ? AND predicted_team != ?",
			gameID, matchID, winner)
	}
}

// === PREDICTION HANDLERS ===

// makePredictionHandler creates a prediction
func makePredictionHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)

	var pred struct {
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
		user.ID, pred.GameID).Scan(&isActive)
	if !isActive {
		sendError(w, "User is eliminated from this game", 403)
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
		sendError(w, "Round is not open for predictions", 400)
		return
	}

	// Check deadline
	if submissionDeadline != "" {
		deadline, err := time.Parse("2006-01-02T15:04:05Z", submissionDeadline)
		if err != nil {
			deadline, err = time.Parse("2006-01-02 15:04:05", submissionDeadline)
		}
		if err == nil && time.Now().After(deadline) {
			sendError(w, "Submission deadline has passed", 400)
			return
		}
	}

	// Check if already predicted for this round
	var count int
	db.QueryRow("SELECT COUNT(*) FROM predictions WHERE user_id = ? AND game_id = ? AND round_number = ?",
		user.ID, pred.GameID, roundNum).Scan(&count)
	if count > 0 {
		sendError(w, "Already predicted for this round", 409)
		return
	}

	// Check if user has already picked this team
	var teamCount int
	db.QueryRow("SELECT COUNT(*) FROM predictions WHERE user_id = ? AND game_id = ? AND predicted_team = ?",
		user.ID, pred.GameID, pred.PredictedTeam).Scan(&teamCount)
	if teamCount > 0 {
		sendError(w, "You have already picked this team in this game", 409)
		return
	}

	_, err := db.Exec(`INSERT INTO predictions (user_id, game_id, match_id, round_number, predicted_team) 
		VALUES (?, ?, ?, ?, ?)`, user.ID, pred.GameID, pred.MatchID, roundNum, pred.PredictedTeam)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getPredictionsHandler returns predictions (own or all if admin)
func getPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	
	gameIDStr := r.URL.Query().Get("game_id")
	viewAllStr := r.URL.Query().Get("view_all")

	gameID, _ := strconv.Atoi(gameIDStr)
	viewAll := viewAllStr == "true"

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	// If trying to view all predictions, must be admin
	if viewAll && !user.IsAdmin {
		sendError(w, "Only admins can view all predictions", 403)
		return
	}

	var rows *sql.Rows
	var err error

	if viewAll && user.IsAdmin {
		rows, err = db.Query(`
			SELECT p.id, p.user_id, p.match_id, p.round_number, p.predicted_team, 
				p.is_correct, p.voided, p.created_at, '', m.home_team, m.away_team, m.result, m.date
			FROM predictions p
			JOIN matches m ON p.match_id = m.id
			WHERE p.game_id = ?
			ORDER BY p.round_number DESC, p.created_at DESC
		`, gameID)
	} else {
		rows, err = db.Query(`
			SELECT p.id, p.user_id, p.match_id, p.round_number, p.predicted_team, 
				p.is_correct, p.voided, p.created_at, '', m.home_team, m.away_team, m.result, m.date
			FROM predictions p
			JOIN matches m ON p.match_id = m.id
			WHERE p.game_id = ? AND p.user_id = ?
			ORDER BY p.round_number DESC, p.created_at DESC
		`, gameID, user.ID)
	}

	if err != nil {
		sendError(w, err.Error(), 500)
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

// getUsedTeamsHandler returns teams user has already picked
func getUsedTeamsHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)

	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`
		SELECT DISTINCT predicted_team 
		FROM predictions 
		WHERE user_id = ? AND game_id = ?
	`, user.ID, gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
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

// === STANDINGS HANDLERS ===

// getStandingsHandler returns current standings
func getStandingsHandler(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, _ := strconv.Atoi(gameIDStr)
	if gameID == 0 {
		gameID = getCurrentGameID()
	}

	rows, err := db.Query(`
		SELECT gp.user_id, gp.is_active, 
			COALESCE(MAX(p.round_number), 0) as last_round
		FROM game_players gp
		LEFT JOIN predictions p ON gp.user_id = p.user_id AND p.game_id = gp.game_id
		WHERE gp.game_id = ?
		GROUP BY gp.user_id
		ORDER BY gp.is_active DESC, last_round DESC
	`, gameID)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	standings := []StandingsEntry{}
	for rows.Next() {
		var s StandingsEntry
		rows.Scan(&s.UserID, &s.IsActive, &s.LastRound)
		// UserName will need to be fetched from Identity Service or left empty
		s.UserName = fmt.Sprintf("User %d", s.UserID)
		standings = append(standings, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}
