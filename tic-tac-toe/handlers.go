package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"pubgames/shared/auth"
)

const (
	DEFAULT_SESSION_TIMEOUT = 60
	DEFAULT_MOVE_TIMEOUT    = 30
)

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	config := Config{
		AppName:              APP_NAME,
		AppIcon:              APP_ICON,
		BackendURL:           "http://localhost:" + BACKEND_PORT,
		DefaultSessionMinutes: DEFAULT_SESSION_TIMEOUT,
		DefaultMoveSeconds:   DEFAULT_MOVE_TIMEOUT,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{
		ID:      authUser.ID,
		Email:   authUser.Email,
		Name:    authUser.Name,
		IsAdmin: authUser.IsAdmin,
	}
	var inGame bool
	err := db.QueryRow(`SELECT COUNT(*) > 0 FROM games WHERE (player1_id = ? OR player2_id = ?) AND status = 'active'`, user.ID, user.ID).Scan(&inGame)
	if err != nil {
		inGame = false
	}
	err = markUserOnline(user.ID, user.Name, inGame)
	if err != nil {
		sendError(w, "Failed to update online status", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		// Already logged out or invalid token
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}
	
	// Remove user from online_users
	_, err := db.Exec(`DELETE FROM online_users WHERE user_id = ?`, authUser.ID)
	if err != nil {
		log.Printf("Warning: Failed to remove user %d from online_users: %v", authUser.ID, err)
	}
	
	log.Printf("👋 User %d (%s) logged out and removed from lobby", authUser.ID, authUser.Name)
	
	// Notify all lobby users that this user went offline
	broadcastUserOffline(authUser.ID, authUser.Name)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func getOnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	cleanupOnlineUsers()
	rows, err := db.Query(`SELECT user_id, user_name, last_seen_at, in_game FROM online_users WHERE user_id != ? AND datetime(last_seen_at) > datetime('now', '-5 minutes') ORDER BY user_name`, user.ID)
	if err != nil {
		sendError(w, "Failed to get online users", 500)
		return
	}
	defer rows.Close()
	onlineUsers := []OnlineUser{}
	for rows.Next() {
		var ou OnlineUser
		var inGameInt int
		err := rows.Scan(&ou.UserID, &ou.UserName, &ou.LastSeenAt, &inGameInt)
		if err != nil {
			continue
		}
		ou.InGame = inGameInt == 1
		onlineUsers = append(onlineUsers, ou)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(onlineUsers)
}

func createChallengeHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	var settings struct {
		OpponentID    int      `json:"opponent_id"`
		Mode          GameMode `json:"mode"`
		MoveTimeLimit int      `json:"move_time_limit"`
		FirstTo       int      `json:"first_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}
	validFirstTo := map[int]bool{1: true, 2: true, 3: true, 5: true, 10: true, 20: true}
	if !validFirstTo[settings.FirstTo] {
		sendError(w, "Invalid first_to value", 400)
		return
	}
	var opponentName string
	err := db.QueryRow(`SELECT user_name FROM online_users WHERE user_id = ? AND datetime(last_seen_at) > datetime('now', '-5 minutes')`, settings.OpponentID).Scan(&opponentName)
	if err == sql.ErrNoRows {
		sendError(w, "Opponent is not online", 400)
		return
	} else if err != nil {
		sendError(w, "Failed to verify opponent", 500)
		return
	}
	var existingGame int
	err = db.QueryRow(`SELECT COUNT(*) FROM games WHERE (player1_id = ? OR player2_id = ? OR player1_id = ? OR player2_id = ?) AND status IN ('waiting', 'active')`, user.ID, user.ID, settings.OpponentID, settings.OpponentID).Scan(&existingGame)
	if err != nil {
		sendError(w, "Failed to check existing games", 500)
		return
	}
	if existingGame > 0 {
		sendError(w, "One of the players is already in a game", 400)
		return
	}
	result, err := db.Exec(`INSERT INTO games (player1_id, player1_name, player2_id, player2_name, mode, status, current_turn, move_time_limit, session_timeout, first_to) VALUES (?, ?, ?, ?, ?, 'waiting', 1, ?, ?, ?)`, user.ID, user.Name, settings.OpponentID, opponentName, settings.Mode, settings.MoveTimeLimit, DEFAULT_SESSION_TIMEOUT, settings.FirstTo)
	if err != nil {
		sendError(w, "Failed to create challenge", 500)
		return
	}
	gameID, _ := result.LastInsertId()
	
	// Notify opponent via lobby WebSocket (if connected)
	var challenge Game
	var player2ID sql.NullInt64
	err = db.QueryRow(`
		SELECT id, player1_id, player1_name, player2_id, player2_name, mode, move_time_limit, first_to, created_at 
		FROM games WHERE id = ?
	`, gameID).Scan(&challenge.ID, &challenge.Player1ID, &challenge.Player1Name, &player2ID, &challenge.Player2Name, &challenge.Mode, &challenge.MoveTimeLimit, &challenge.FirstTo, &challenge.CreatedAt)
	
	if err == nil && player2ID.Valid {
		p2id := int(player2ID.Int64)
		challenge.Player2ID = &p2id
		notifyChallengeReceived(settings.OpponentID, &challenge)
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"game_id": gameID, "message": "Challenge sent"})
}

func getPendingChallengesHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	
	// Cleanup expired challenges before querying
	cleanupExpiredChallenges()
	
	rows, err := db.Query(`SELECT id, player1_id, player1_name, player2_id, player2_name, mode, move_time_limit, first_to, created_at FROM games WHERE player2_id = ? AND status = 'waiting' ORDER BY created_at DESC`, user.ID)
	if err != nil {
		sendError(w, "Failed to get challenges", 500)
		return
	}
	defer rows.Close()
	challenges := []Game{}
	for rows.Next() {
		var g Game
		err := rows.Scan(&g.ID, &g.Player1ID, &g.Player1Name, &g.Player2ID, &g.Player2Name, &g.Mode, &g.MoveTimeLimit, &g.FirstTo, &g.CreatedAt)
		if err != nil {
			continue
		}
		challenges = append(challenges, g)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(challenges)
}

func respondToChallengeHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	vars := mux.Vars(r)
	gameID, _ := strconv.Atoi(vars["id"])
	var response struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}
	var currentStatus string
	var player2ID int
	err := db.QueryRow(`SELECT status, player2_id FROM games WHERE id = ?`, gameID).Scan(&currentStatus, &player2ID)
	if err == sql.ErrNoRows {
		sendError(w, "Game not found", 404)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if player2ID != user.ID {
		sendError(w, "Not your challenge", 403)
		return
	}
	if currentStatus != "waiting" {
		sendError(w, "Challenge already responded to", 400)
		return
	}
	
	// Declare player1ID at function scope so it's available for broadcasts
	var player1ID int
	var player1Name string
	
	newStatus := "declined"
	if response.Accept {
		newStatus = "active"
		markUserOnline(user.ID, user.Name, true)
		db.QueryRow("SELECT player1_id, player1_name FROM games WHERE id = ?", gameID).Scan(&player1ID, &player1Name)
		markUserOnline(player1ID, player1Name, true)
	} else {
		// For decline, we still need player1ID to notify them
		db.QueryRow("SELECT player1_id FROM games WHERE id = ?", gameID).Scan(&player1ID)
	}
	_, err = db.Exec(`UPDATE games SET status = ?, last_move_at = CURRENT_TIMESTAMP WHERE id = ?`, newStatus, gameID)
	if err != nil {
		sendError(w, "Failed to update challenge", 500)
		return
	}
	
	// Notify both players via lobby WebSocket (if connected)
	if response.Accept {
		// Fetch full game to notify
		var game Game
		var gPlayer2ID sql.NullInt64
		err = db.QueryRow(`
			SELECT id, player1_id, player1_name, player2_id, player2_name, mode, status, move_time_limit, first_to, created_at
			FROM games WHERE id = ?
		`, gameID).Scan(&game.ID, &game.Player1ID, &game.Player1Name, &gPlayer2ID, &game.Player2Name, &game.Mode, &game.Status, &game.MoveTimeLimit, &game.FirstTo, &game.CreatedAt)
		
		if err == nil {
			if gPlayer2ID.Valid {
				p2id := int(gPlayer2ID.Int64)
				game.Player2ID = &p2id
			}
			notifyChallengeAccepted(player1ID, user.ID, &game)
		}
	} else {
		// Challenge declined
		notifyChallengeDeclined(player1ID, gameID)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": newStatus, "message": fmt.Sprintf("Challenge %s", newStatus)})
}

func getActiveGameHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	var game Game
	var player2ID sql.NullInt64
	var winnerID sql.NullInt64
	var lastMoveAt sql.NullTime
	var completedAt sql.NullTime
	err := db.QueryRow(`SELECT id, player1_id, player1_name, player2_id, player2_name, mode, status, current_turn, winner_id, board, move_time_limit, session_timeout, first_to, player1_score, player2_score, current_round, last_move_at, created_at, completed_at FROM games WHERE (player1_id = ? OR player2_id = ?) AND status IN ('waiting', 'active') ORDER BY created_at DESC LIMIT 1`, user.ID, user.ID).Scan(&game.ID, &game.Player1ID, &game.Player1Name, &player2ID, &game.Player2Name, &game.Mode, &game.Status, &game.CurrentTurn, &winnerID, &game.Board, &game.MoveTimeLimit, &game.SessionTimeout, &game.FirstTo, &game.Player1Score, &game.Player2Score, &game.CurrentRound, &lastMoveAt, &game.CreatedAt, &completedAt)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nil)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if player2ID.Valid {
		p2id := int(player2ID.Int64)
		game.Player2ID = &p2id
	}
	if winnerID.Valid {
		wid := int(winnerID.Int64)
		game.WinnerID = &wid
	}
	if lastMoveAt.Valid {
		game.LastMoveAt = &lastMoveAt.Time
	}
	if completedAt.Valid {
		game.CompletedAt = &completedAt.Time
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func makeMoveHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	var moveReq struct {
		GameID   int `json:"game_id"`
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&moveReq); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}
	if moveReq.Position < 0 || moveReq.Position > 8 {
		sendError(w, "Invalid position", 400)
		return
	}
	var game Game
	var player2ID sql.NullInt64
	var winnerID sql.NullInt64
	err := db.QueryRow(`SELECT id, player1_id, player2_id, status, current_turn, board, winner_id, first_to, player1_score, player2_score, current_round FROM games WHERE id = ?`, moveReq.GameID).Scan(&game.ID, &game.Player1ID, &player2ID, &game.Status, &game.CurrentTurn, &game.Board, &winnerID, &game.FirstTo, &game.Player1Score, &game.Player2Score, &game.CurrentRound)
	if err == sql.ErrNoRows {
		sendError(w, "Game not found", 404)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if game.Status != "active" {
		sendError(w, "Game is not active", 400)
		return
	}
	var playerNumber int
	var symbol string
	if user.ID == game.Player1ID {
		playerNumber = 1
		symbol = "X"
	} else if player2ID.Valid && user.ID == int(player2ID.Int64) {
		playerNumber = 2
		symbol = "O"
	} else {
		sendError(w, "You are not in this game", 403)
		return
	}
	if game.CurrentTurn != playerNumber {
		sendError(w, "Not your turn", 400)
		return
	}
	var board []string
	if err := json.Unmarshal([]byte(game.Board), &board); err != nil {
		sendError(w, "Invalid board state", 500)
		return
	}
	if board[moveReq.Position] != "" {
		sendError(w, "Position already taken", 400)
		return
	}
	board[moveReq.Position] = symbol
	boardJSON, _ := json.Marshal(board)
	_, err = db.Exec(`INSERT INTO moves (game_id, player_id, position, symbol) VALUES (?, ?, ?, ?)`, moveReq.GameID, user.ID, moveReq.Position, symbol)
	if err != nil {
		sendError(w, "Failed to record move", 500)
		return
	}
	hasWinner, isDraw := checkWinner(board)
	nextTurn := 3 - playerNumber
	roundOver := hasWinner || isDraw
	seriesOver := false
	var finalWinnerID *int
	if roundOver {
		if hasWinner {
			if playerNumber == 1 {
				game.Player1Score++
			} else {
				game.Player2Score++
			}
		}
		if game.Player1Score >= game.FirstTo {
			seriesOver = true
			fwid := game.Player1ID
			finalWinnerID = &fwid
		} else if game.Player2Score >= game.FirstTo {
			seriesOver = true
			fwid := int(player2ID.Int64)
			finalWinnerID = &fwid
		}
		if seriesOver {
			_, err = db.Exec(`UPDATE games SET board = ?, status = 'completed', winner_id = ?, player1_score = ?, player2_score = ?, last_move_at = CURRENT_TIMESTAMP, completed_at = CURRENT_TIMESTAMP WHERE id = ?`, string(boardJSON), finalWinnerID, game.Player1Score, game.Player2Score, moveReq.GameID)
			if player2ID.Valid {
				player2IDInt := int(player2ID.Int64)
				if finalWinnerID != nil {
					updatePlayerStats(*finalWinnerID, "", true, false, false)
					loserID := game.Player1ID
					if *finalWinnerID == game.Player1ID {
						loserID = player2IDInt
					}
					updatePlayerStats(loserID, "", false, true, false)
				}
			}
			markUserOnline(game.Player1ID, "", false)
			if player2ID.Valid {
				markUserOnline(int(player2ID.Int64), "", false)
			}
			
			// Broadcast game ended via WebSocket
			updatedGame, fetchErr := getFullGameState(moveReq.GameID)
			if fetchErr == nil {
				broadcastGameEnded(moveReq.GameID, updatedGame)
			}
		} else {
			emptyBoard := []string{"", "", "", "", "", "", "", "", ""}
			emptyBoardJSON, _ := json.Marshal(emptyBoard)
			newRound := game.CurrentRound + 1
			newStarter := 1
			if newRound%2 == 0 {
				newStarter = 2
			}
			_, err = db.Exec(`UPDATE games SET board = ?, current_turn = ?, current_round = ?, player1_score = ?, player2_score = ?, last_move_at = CURRENT_TIMESTAMP WHERE id = ?`, string(emptyBoardJSON), newStarter, newRound, game.Player1Score, game.Player2Score, moveReq.GameID)
		}
	} else {
		_, err = db.Exec(`UPDATE games SET board = ?, current_turn = ?, last_move_at = CURRENT_TIMESTAMP WHERE id = ?`, string(boardJSON), nextTurn, moveReq.GameID)
	}
	if err != nil {
		sendError(w, "Failed to update game", 500)
		return
	}

	// Fetch the updated game to return in response
	var updatedGame Game
	var updatedPlayer2ID sql.NullInt64
	var updatedWinnerID sql.NullInt64
	var lastMoveAt sql.NullTime
	var completedAt sql.NullTime

	err = db.QueryRow(`
		SELECT id, player1_id, player1_name, player2_id, player2_name,
		       mode, status, current_turn, winner_id, board,
		       move_time_limit, session_timeout, first_to, player1_score, 
		       player2_score, current_round, last_move_at, created_at, completed_at
		FROM games WHERE id = ?
	`, moveReq.GameID).Scan(
		&updatedGame.ID, &updatedGame.Player1ID, &updatedGame.Player1Name, &updatedPlayer2ID, &updatedGame.Player2Name,
		&updatedGame.Mode, &updatedGame.Status, &updatedGame.CurrentTurn, &updatedWinnerID, &updatedGame.Board,
		&updatedGame.MoveTimeLimit, &updatedGame.SessionTimeout, &updatedGame.FirstTo, &updatedGame.Player1Score,
		&updatedGame.Player2Score, &updatedGame.CurrentRound, &lastMoveAt, &updatedGame.CreatedAt, &completedAt,
	)

	if err != nil {
		log.Printf("Warning: Failed to fetch updated game: %v", err)
	} else {
		if updatedPlayer2ID.Valid {
			p2id := int(updatedPlayer2ID.Int64)
			updatedGame.Player2ID = &p2id
		}
		if updatedWinnerID.Valid {
			wid := int(updatedWinnerID.Int64)
			updatedGame.WinnerID = &wid
		}
		if lastMoveAt.Valid {
			updatedGame.LastMoveAt = &lastMoveAt.Time
		}
		if completedAt.Valid {
			updatedGame.CompletedAt = &completedAt.Time
		}
		
		// Broadcast move update via WebSocket (for active games only)
		// Game ended broadcasts are handled separately above
		if !seriesOver {
			broadcastGameUpdate(moveReq.GameID, &updatedGame)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":     true,
		"board":       board,
		"round_over":  roundOver,
		"series_over": seriesOver,
		"is_draw":     isDraw,
		"game":        updatedGame, // Return full game state
	}
	if seriesOver && finalWinnerID != nil {
		response["winner_id"] = *finalWinnerID
	}
	json.NewEncoder(w).Encode(response)
}

func checkWinner(board []string) (bool, bool) {
	wins := [][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}
	for _, combo := range wins {
		if board[combo[0]] != "" && board[combo[0]] == board[combo[1]] && board[combo[1]] == board[combo[2]] {
			return true, false
		}
	}
	full := true
	for _, cell := range board {
		if cell == "" {
			full = false
			break
		}
	}
	return false, full
}

func createRematchHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	var req struct {
		GameID int `json:"game_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}
	var player1ID, player2ID, opponentID int
	var status string
	err := db.QueryRow(`SELECT player1_id, player2_id, status FROM games WHERE id = ?`, req.GameID).Scan(&player1ID, &player2ID, &status)
	if err == sql.ErrNoRows {
		sendError(w, "Game not found", 404)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if status != "completed" {
		sendError(w, "Game is not completed", 400)
		return
	}
	if user.ID == player1ID {
		opponentID = player2ID
	} else if user.ID == player2ID {
		opponentID = player1ID
	} else {
		sendError(w, "You are not in this game", 403)
		return
	}
	existingRematch, err := getRematchRequest(req.GameID)
	if err != nil {
		sendError(w, "Failed to check existing rematch", 500)
		return
	}
	if existingRematch != nil {
		if existingRematch.Status == RematchStatusPending {
			sendError(w, "Rematch request already pending", 400)
			return
		}
	}
	rematchID, err := createRematchRequest(req.GameID, user.ID, opponentID)
	if err != nil {
		sendError(w, "Failed to create rematch request", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rematch_id": rematchID, "message": "Rematch request sent"})
}

func getRematchHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	vars := mux.Vars(r)
	gameID, _ := strconv.Atoi(vars["gameId"])
	cleanupExpiredRematches()
	rematch, err := getRematchRequest(gameID)
	if err != nil {
		sendError(w, "Failed to get rematch request", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rematch)
}

func respondToRematchHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	vars := mux.Vars(r)
	rematchID, _ := strconv.Atoi(vars["id"])
	var response struct {
		Accept bool `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		sendError(w, "Invalid request body", 400)
		return
	}
	var rm RematchRequest
	var gameID int
	err := db.QueryRow(`SELECT id, game_id, requester_id, opponent_id, status FROM rematch_requests WHERE id = ?`, rematchID).Scan(&rm.ID, &gameID, &rm.RequesterID, &rm.OpponentID, &rm.Status)
	if err == sql.ErrNoRows {
		sendError(w, "Rematch request not found", 404)
		return
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if rm.OpponentID != user.ID {
		sendError(w, "Not your rematch request", 403)
		return
	}
	if rm.Status != RematchStatusPending {
		sendError(w, "Rematch already responded to", 400)
		return
	}
	newStatus := RematchStatusDeclined
	if response.Accept {
		newStatus = RematchStatusAccepted
		var mode string
		var moveTimeLimit, firstTo int
		var player1ID, player1Name, player2ID, player2Name string
		db.QueryRow(`SELECT mode, move_time_limit, first_to, player1_id, player1_name, player2_id, player2_name FROM games WHERE id = ?`, gameID).Scan(&mode, &moveTimeLimit, &firstTo, &player1ID, &player1Name, &player2ID, &player2Name)
		_, err = db.Exec(`INSERT INTO games (player1_id, player1_name, player2_id, player2_name, mode, status, current_turn, move_time_limit, session_timeout, first_to) VALUES (?, ?, ?, ?, ?, 'active', 1, ?, ?, ?)`, player1ID, player1Name, player2ID, player2Name, mode, moveTimeLimit, DEFAULT_SESSION_TIMEOUT, firstTo)
		if err != nil {
			sendError(w, "Failed to create new game", 500)
			return
		}
		p1ID, _ := strconv.Atoi(player1ID)
		p2ID, _ := strconv.Atoi(player2ID)
		markUserOnline(p1ID, player1Name, true)
		markUserOnline(p2ID, player2Name, true)
	}
	err = updateRematchStatus(rematchID, newStatus)
	if err != nil {
		sendError(w, "Failed to update rematch status", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": string(newStatus), "message": fmt.Sprintf("Rematch %s", newStatus)})
}

func getLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT user_id, user_name, games_played, games_won, games_lost, games_draw FROM player_stats WHERE games_played > 0 ORDER BY games_won DESC, games_played ASC LIMIT 20`)
	if err != nil {
		sendError(w, "Failed to get leaderboard", 500)
		return
	}
	defer rows.Close()
	leaderboard := []PlayerStats{}
	for rows.Next() {
		var ps PlayerStats
		err := rows.Scan(&ps.UserID, &ps.UserName, &ps.GamesPlayed, &ps.GamesWon, &ps.GamesLost, &ps.GamesDraw)
		if err != nil {
			continue
		}
		if ps.GamesPlayed > 0 {
			ps.WinRate = float64(ps.GamesWon) / float64(ps.GamesPlayed) * 100
		}
		leaderboard = append(leaderboard, ps)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

func getPlayerStatsHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	var stats PlayerStats
	err := db.QueryRow(`SELECT user_id, user_name, games_played, games_won, games_lost, games_draw FROM player_stats WHERE user_id = ?`, user.ID).Scan(&stats.UserID, &stats.UserName, &stats.GamesPlayed, &stats.GamesWon, &stats.GamesLost, &stats.GamesDraw)
	if err == sql.ErrNoRows {
		stats = PlayerStats{UserID: user.ID, UserName: user.Name, GamesPlayed: 0, GamesWon: 0, GamesLost: 0, GamesDraw: 0, WinRate: 0}
	} else if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.GamesWon) / float64(stats.GamesPlayed) * 100
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getGameHistoryHandler(w http.ResponseWriter, r *http.Request) {
	authUser := auth.GetUser(r)
	if authUser == nil {
		sendError(w, "User not found", 401)
		return
	}
	user := &User{ID: authUser.ID, Email: authUser.Email, Name: authUser.Name, IsAdmin: authUser.IsAdmin}
	rows, err := db.Query(`SELECT id, player1_id, player1_name, player2_id, player2_name, mode, status, winner_id, first_to, player1_score, player2_score, created_at, completed_at FROM games WHERE (player1_id = ? OR player2_id = ?) AND status = 'completed' ORDER BY completed_at DESC LIMIT 20`, user.ID, user.ID)
	if err != nil {
		sendError(w, "Failed to get history", 500)
		return
	}
	defer rows.Close()
	history := []Game{}
	for rows.Next() {
		var g Game
		var player2ID sql.NullInt64
		var winnerID sql.NullInt64
		var completedAt sql.NullTime
		err := rows.Scan(&g.ID, &g.Player1ID, &g.Player1Name, &player2ID, &g.Player2Name, &g.Mode, &g.Status, &winnerID, &g.FirstTo, &g.Player1Score, &g.Player2Score, &g.CreatedAt, &completedAt)
		if err != nil {
			continue
		}
		if player2ID.Valid {
			p2id := int(player2ID.Int64)
			g.Player2ID = &p2id
		}
		if winnerID.Valid {
			wid := int(winnerID.Int64)
			g.WinnerID = &wid
		}
		if completedAt.Valid {
			g.CompletedAt = &completedAt.Time
		}
		history = append(history, g)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code})
}
