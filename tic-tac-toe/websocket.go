package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"pubgames/shared/auth"
)

// Connection manager - tracks all active WebSocket connections
type ConnectionManager struct {
	connections map[int]map[int]*websocket.Conn // gameID -> userID -> connection
	mu          sync.RWMutex
}

var connManager = &ConnectionManager{
	connections: make(map[int]map[int]*websocket.Conn),
}

// WebSocket message structure
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Message types:
// Client -> Server: "ping", "ack", "reconnecting"
// Server -> Client: "pong", "ready", "move_update", "game_ended", "opponent_disconnected"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from frontend
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:30040" || 
		       origin == "http://192.168.1.45:30040"
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// gameWebSocketHandler handles WebSocket connections for a specific game
// Endpoint: /api/ws/game/{gameId}
func gameWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Manually authenticate via token query parameter (WebSockets can't send custom headers)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", 401)
		return
	}

	// Validate token with Identity Service
	authUser, err := validateTokenWithIdentity(token)
	if err != nil {
		log.Printf("WebSocket auth failed: %v", err)
		http.Error(w, "Unauthorized", 401)
		return
	}
	
	user := &User{
		ID:      authUser.ID,
		Email:   authUser.Email,
		Name:    authUser.Name,
		IsAdmin: authUser.IsAdmin,
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameID, err := strconv.Atoi(vars["gameId"])
	if err != nil {
		http.Error(w, "Invalid game ID", 400)
		return
	}

	// Verify user is in this game
	if !isUserInGame(user.ID, gameID) {
		http.Error(w, "Not authorized for this game", 403)
		return
	}

	// Check if user already has connection for this game (prevent multiple tabs)
	if hasExistingConnection(user.ID, gameID) {
		http.Error(w, "Game already open in another tab", 409)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register connection
	registerConnection(gameID, user.ID, conn)
	defer unregisterConnection(gameID, user.ID)

	log.Printf("🔌 WebSocket connection attempt: User %d (%s) for Game %d", user.ID, user.Name, gameID)

	// Perform bidirectional handshake
	if !performHandshake(conn, gameID, user.ID) {
		log.Printf("❌ Handshake failed: User %d, Game %d", user.ID, gameID)
		return
	}

	log.Printf("✅ WebSocket ready: User %d (%s), Game %d", user.ID, user.Name, gameID)

	// Start listening for messages and maintain connection
	handleGameConnection(conn, gameID, user.ID)
}

// performHandshake conducts the bidirectional handshake
// Flow: Client PING -> Server PONG -> Client ACK -> Server READY
func performHandshake(conn *websocket.Conn, gameID, userID int) bool {
	// 1. Wait for client PING
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg WSMessage
	if err := conn.ReadJSON(&msg); err != nil {
		log.Printf("Handshake error: Failed to read PING: %v", err)
		return false
	}
	if msg.Type != "ping" {
		log.Printf("Handshake error: Expected PING, got %s", msg.Type)
		return false
	}
	log.Printf("📨 Received PING from User %d", userID)

	// 2. Send PONG
	if err := conn.WriteJSON(WSMessage{Type: "pong"}); err != nil {
		log.Printf("Handshake error: Failed to send PONG: %v", err)
		return false
	}
	log.Printf("📤 Sent PONG to User %d", userID)

	// 3. Wait for client ACK
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err := conn.ReadJSON(&msg); err != nil {
		log.Printf("Handshake error: Failed to read ACK: %v", err)
		return false
	}
	if msg.Type != "ack" {
		log.Printf("Handshake error: Expected ACK, got %s", msg.Type)
		return false
	}
	log.Printf("📨 Received ACK from User %d", userID)

	// 4. Fetch current game state
	game, err := getFullGameState(gameID)
	if err != nil {
		log.Printf("Handshake error: Failed to fetch game state: %v", err)
		return false
	}

	// 5. Send READY with game state
	if err := conn.WriteJSON(WSMessage{
		Type:    "ready",
		Payload: game,
	}); err != nil {
		log.Printf("Handshake error: Failed to send READY: %v", err)
		return false
	}
	log.Printf("📤 Sent READY to User %d", userID)

	// Reset read deadline for normal operation
	conn.SetReadDeadline(time.Time{})

	log.Printf("✅ Handshake complete: User %d, Game %d", userID, gameID)
	return true
}

// handleGameConnection maintains the WebSocket connection
func handleGameConnection(conn *websocket.Conn, gameID, userID int) {
	// Set up ping/pong for connection health monitoring
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	// Read messages (for reconnection notifications and disconnect detection)
	go func() {
		defer close(done)
		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			// Handle reconnection attempts
			if msg.Type == "reconnecting" {
				log.Printf("🔄 User %d attempting reconnection to Game %d", userID, gameID)
			}
		}
	}()

	// Keep connection alive
	for {
		select {
		case <-done:
			// Connection closed
			log.Printf("🔌 Connection closed: User %d, Game %d", userID, gameID)
			notifyOpponentDisconnected(gameID, userID)
			return

		case <-ticker.C:
			// Send ping to check if connection is alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping failed: User %d, Game %d: %v", userID, gameID, err)
				return
			}
		}
	}
}

// broadcastGameUpdate sends game state update to all connected players
func broadcastGameUpdate(gameID int, game *Game) {
	connManager.mu.RLock()
	defer connManager.mu.RUnlock()

	gameConns, exists := connManager.connections[gameID]
	if !exists {
		log.Printf("No WebSocket connections for Game %d", gameID)
		return
	}

	msg := WSMessage{
		Type:    "move_update",
		Payload: game,
	}

	for userID, conn := range gameConns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to send update to User %d: %v", userID, err)
		} else {
			log.Printf("📤 Sent move_update to User %d for Game %d", userID, gameID)
		}
	}
}

// broadcastGameEnded notifies all players that game has ended
func broadcastGameEnded(gameID int, game *Game) {
	connManager.mu.RLock()
	defer connManager.mu.RUnlock()

	gameConns, exists := connManager.connections[gameID]
	if !exists {
		return
	}

	msg := WSMessage{
		Type:    "game_ended",
		Payload: game,
	}

	for userID, conn := range gameConns {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to send game_ended to User %d: %v", userID, err)
		} else {
			log.Printf("📤 Sent game_ended to User %d for Game %d", userID, gameID)
		}
	}
}

// notifyOpponentDisconnected alerts opponent when player disconnects
func notifyOpponentDisconnected(gameID, disconnectedUserID int) {
	connManager.mu.RLock()
	defer connManager.mu.RUnlock()

	gameConns, exists := connManager.connections[gameID]
	if !exists {
		return
	}

	msg := WSMessage{
		Type: "opponent_disconnected",
		Payload: map[string]int{
			"user_id": disconnectedUserID,
		},
	}

	for userID, conn := range gameConns {
		if userID != disconnectedUserID {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send disconnect notification to User %d: %v", userID, err)
			} else {
				log.Printf("📤 Notified User %d that User %d disconnected from Game %d", userID, disconnectedUserID, gameID)
			}
		}
	}
}

// Connection management functions

func registerConnection(gameID, userID int, conn *websocket.Conn) {
	connManager.mu.Lock()
	defer connManager.mu.Unlock()

	if connManager.connections[gameID] == nil {
		connManager.connections[gameID] = make(map[int]*websocket.Conn)
	}
	connManager.connections[gameID][userID] = conn

	log.Printf("✅ Registered WebSocket: User %d in Game %d", userID, gameID)
}

func unregisterConnection(gameID, userID int) {
	connManager.mu.Lock()
	defer connManager.mu.Unlock()

	if connManager.connections[gameID] != nil {
		delete(connManager.connections[gameID], userID)
		if len(connManager.connections[gameID]) == 0 {
			delete(connManager.connections, gameID)
			log.Printf("🗑️  Cleaned up empty game connection map for Game %d", gameID)
		}
	}

	log.Printf("🔌 Unregistered WebSocket: User %d from Game %d", userID, gameID)
}

func hasExistingConnection(userID, gameID int) bool {
	connManager.mu.RLock()
	defer connManager.mu.RUnlock()

	if gameConns, exists := connManager.connections[gameID]; exists {
		_, hasConn := gameConns[userID]
		return hasConn
	}
	return false
}

// Helper functions

func isUserInGame(userID, gameID int) bool {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM games 
		WHERE id = ? AND (player1_id = ? OR player2_id = ?)
		AND status IN ('active', 'waiting')
	`, gameID, userID, userID).Scan(&count)
	
	if err != nil {
		log.Printf("Error checking user in game: %v", err)
		return false
	}
	
	return count > 0
}

func getFullGameState(gameID int) (*Game, error) {
	var game Game
	var player2ID sql.NullInt64
	var winnerID sql.NullInt64
	var lastMoveAt sql.NullTime
	var completedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, player1_id, player1_name, player2_id, player2_name,
		       mode, status, current_turn, winner_id, board,
		       move_time_limit, session_timeout, first_to, player1_score, 
		       player2_score, current_round, last_move_at, created_at, completed_at
		FROM games WHERE id = ?
	`, gameID).Scan(
		&game.ID, &game.Player1ID, &game.Player1Name, &player2ID, &game.Player2Name,
		&game.Mode, &game.Status, &game.CurrentTurn, &winnerID, &game.Board,
		&game.MoveTimeLimit, &game.SessionTimeout, &game.FirstTo, &game.Player1Score,
		&game.Player2Score, &game.CurrentRound, &lastMoveAt, &game.CreatedAt, &completedAt,
	)

	if err != nil {
		return nil, err
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

	return &game, nil
}

// ============================================================================
// SMART LOBBY WEBSOCKET - Event-driven, auto-disconnects after 30s
// ============================================================================

// LobbyConnection tracks a user's lobby WebSocket with auto-disconnect
type LobbyConnection struct {
	conn      *websocket.Conn
	timer     *time.Timer
	userID    int
	connected time.Time
}

// LobbyConnectionManager tracks temporary lobby WebSocket connections
type LobbyConnectionManager struct {
	connections map[int]*LobbyConnection // userID -> connection
	mu          sync.RWMutex
}

var lobbyConnManager = &LobbyConnectionManager{
	connections: make(map[int]*LobbyConnection),
}

// lobbyWebSocketHandler handles temporary WebSocket for challenge notifications
// Automatically disconnects after 30 seconds to prevent mobile battery drain
func lobbyWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate via token query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", 401)
		return
	}

	authUser, err := validateTokenWithIdentity(token)
	if err != nil {
		log.Printf("Lobby WebSocket auth failed: %v", err)
		http.Error(w, "Unauthorized", 401)
		return
	}
	
	user := &User{
		ID:      authUser.ID,
		Email:   authUser.Email,
		Name:    authUser.Name,
		IsAdmin: authUser.IsAdmin,
	}

	// Close any existing connection (prevents duplicates)
	if existingConn := getLobbyConnection(user.ID); existingConn != nil {
		log.Printf("⚠️ Closing existing lobby connection for User %d", user.ID)
		existingConn.conn.Close()
		unregisterLobbyConnection(user.ID)
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Lobby WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register with 30-second auto-disconnect timer
	registerLobbyConnection(user.ID, conn)
	defer unregisterLobbyConnection(user.ID)

	log.Printf("🏛️ Lobby WS connected: User %d (auto-disconnect in 30s)", user.ID)

	// Send connected confirmation
	if err := conn.WriteJSON(WSMessage{Type: "lobby_connected"}); err != nil {
		log.Printf("Failed to send lobby_connected: %v", err)
		return
	}

	// Keep alive with pings (for the 30s duration)
	conn.SetReadDeadline(time.Now().Add(35 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(35 * time.Second))
		return nil
	})

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})

	// Read messages (mostly for disconnect detection)
	go func() {
		defer close(done)
		for {
			var msg WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			// Could handle client messages here if needed
		}
	}()

	// Keep connection alive until auto-disconnect or manual close
	for {
		select {
		case <-done:
			log.Printf("🏛️ Lobby WS closed by client: User %d", user.ID)
			return

		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Lobby connection management with auto-disconnect

func registerLobbyConnection(userID int, conn *websocket.Conn) {
	lobbyConnManager.mu.Lock()
	defer lobbyConnManager.mu.Unlock()
	
	// Create auto-disconnect timer (30 seconds)
	timer := time.AfterFunc(30*time.Second, func() {
		log.Printf("⏰ Auto-disconnecting lobby WS: User %d (30s timeout)", userID)
		lobbyConnManager.mu.Lock()
		if lc, exists := lobbyConnManager.connections[userID]; exists {
			lc.conn.Close()
			delete(lobbyConnManager.connections, userID)
		}
		lobbyConnManager.mu.Unlock()
	})
	
	lobbyConnManager.connections[userID] = &LobbyConnection{
		conn:      conn,
		timer:     timer,
		userID:    userID,
		connected: time.Now(),
	}
	
	log.Printf("✅ Registered lobby WS: User %d (expires in 30s)", userID)
}

func unregisterLobbyConnection(userID int) {
	lobbyConnManager.mu.Lock()
	defer lobbyConnManager.mu.Unlock()
	
	if lc, exists := lobbyConnManager.connections[userID]; exists {
		// Cancel auto-disconnect timer
		if lc.timer != nil {
			lc.timer.Stop()
		}
		delete(lobbyConnManager.connections, userID)
		log.Printf("🔌 Unregistered lobby WS: User %d", userID)
	}
}

func getLobbyConnection(userID int) *LobbyConnection {
	lobbyConnManager.mu.RLock()
	defer lobbyConnManager.mu.RUnlock()
	return lobbyConnManager.connections[userID]
}

// Targeted notification functions - only notify specific users

func notifyChallengeReceived(opponentID int, challenge *Game) {
	lc := getLobbyConnection(opponentID)
	if lc == nil {
		log.Printf("User %d not connected to lobby WS (OK - they'll poll)", opponentID)
		return
	}

	msg := WSMessage{
		Type:    "challenge_received",
		Payload: challenge,
	}

	if err := lc.conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send challenge_received to User %d: %v", opponentID, err)
	} else {
		log.Printf("📨 Sent challenge_received to User %d", opponentID)
	}
}

func notifyChallengeAccepted(player1ID, player2ID int, game *Game) {
	msg := WSMessage{
		Type:    "challenge_accepted",
		Payload: game,
	}

	// Notify challenger (player1)
	if lc := getLobbyConnection(player1ID); lc != nil {
		if err := lc.conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to send challenge_accepted to User %d: %v", player1ID, err)
		} else {
			log.Printf("📨 Sent challenge_accepted to User %d (challenger)", player1ID)
		}
	}

	// Notify accepter (player2)
	if lc := getLobbyConnection(player2ID); lc != nil {
		if err := lc.conn.WriteJSON(msg); err != nil {
			log.Printf("Failed to send challenge_accepted to User %d: %v", player2ID, err)
		} else {
			log.Printf("📨 Sent challenge_accepted to User %d (accepter)", player2ID)
		}
	}
}

func notifyChallengeDeclined(challengerID int, gameID int) {
	lc := getLobbyConnection(challengerID)
	if lc == nil {
		return
	}

	msg := WSMessage{
		Type: "challenge_declined",
		Payload: map[string]int{
			"game_id": gameID,
		},
	}

	if err := lc.conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send challenge_declined to User %d: %v", challengerID, err)
	} else {
		log.Printf("📨 Sent challenge_declined to User %d", challengerID)
	}
}

// Broadcast user offline to all connected lobby users
func broadcastUserOffline(userID int, userName string) {
	lobbyConnManager.mu.RLock()
	defer lobbyConnManager.mu.RUnlock()
	
	msg := WSMessage{
		Type: "user_offline",
		Payload: map[string]interface{}{
			"user_id":   userID,
			"user_name": userName,
		},
	}
	
	// Send to all connected lobby users
	for _, lc := range lobbyConnManager.connections {
		if lc.userID != userID { // Don't send to the user who logged out
			if err := lc.conn.WriteJSON(msg); err != nil {
				log.Printf("Failed to send user_offline to User %d: %v", lc.userID, err)
			} else {
				log.Printf("📨 Sent user_offline (User %d) to User %d", userID, lc.userID)
			}
		}
	}
}

// validateTokenWithIdentity validates a JWT token with the Identity Service
func validateTokenWithIdentity(token string) (*auth.User, error) {
	req, err := http.NewRequest("GET", IDENTITY_SERVICE+"/api/validate-token", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, http.ErrAbortHandler
	}

	var user auth.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
