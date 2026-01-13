package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

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

// ===== COMPETITION HANDLERS =====

func getCompetitionsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, type, status, start_date, end_date, description, created_at
		FROM competitions
		ORDER BY created_at DESC
	`)
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer rows.Close()

	competitions := []Competition{}
	for rows.Next() {
		var c Competition
		var startDate, endDate sql.NullTime
		var description sql.NullString

		rows.Scan(&c.ID, &c.Name, &c.Type, &c.Status, &startDate, &endDate,
			&description, &c.CreatedAt)

		if startDate.Valid {
			c.StartDate = &startDate.Time
		}
		if endDate.Valid {
			c.EndDate = &endDate.Time
		}
		if description.Valid {
			c.Description = description.String
		}

		competitions = append(competitions, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(competitions)
}

func createCompetitionHandler(w http.ResponseWriter, r *http.Request) {
	var req Competition
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding competition request: %v", err)
		sendError(w, "Invalid request body: "+err.Error(), 400)
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}

	log.Printf("Creating competition: %+v", req)

	result, err := db.Exec(`
		INSERT INTO competitions (name, type, status, start_date, end_date, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Name, req.Type, req.Status, req.StartDate, req.EndDate, req.Description)

	if err != nil {
		log.Printf("Error inserting competition: %v", err)
		sendError(w, "Failed to create competition: "+err.Error(), 400)
		return
	}

	id, _ := result.LastInsertId()
	req.ID = int(id)

	log.Printf("✅ Competition created: %d - %s", id, req.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func updateCompetitionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req Competition
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding competition update: %v", err)
		sendError(w, "Invalid request body: "+err.Error(), 400)
		return
	}

	log.Printf("Updating competition %s: %+v", id, req)

	// If marking as completed, validate at least one winner exists
	if req.Status == "completed" {
		var first, second, third sql.NullInt64
		db.QueryRow(`
			SELECT 
				COUNT(CASE WHEN position = 1 THEN 1 END) as first,
				COUNT(CASE WHEN position = 2 THEN 1 END) as second,
				COUNT(CASE WHEN position = 3 THEN 1 END) as third
			FROM entries 
			WHERE competition_id = ?
		`, id).Scan(&first, &second, &third)

		if !first.Valid || first.Int64 == 0 {
			sendError(w, "Cannot complete: No 1st place winner set. At least one entry must have position 1.", 400)
			return
		}
	}

	_, err = db.Exec(`
		UPDATE competitions 
		SET name = ?, type = ?, status = ?, start_date = ?, end_date = ?, description = ?
		WHERE id = ?
	`, req.Name, req.Type, req.Status, req.StartDate, req.EndDate, req.Description, id)

	if err != nil {
		log.Printf("Error updating competition: %v", err)
		sendError(w, "Failed to update: "+err.Error(), 500)
		return
	}

	log.Printf("✅ Competition %s updated successfully", id)
	w.WriteHeader(http.StatusOK)
}

// ===== ENTRY HANDLERS =====

func getEntriesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	rows, err := db.Query(`
		SELECT id, competition_id, name, seed, number, status, eliminated_date, position
		FROM entries
		WHERE competition_id = ?
		ORDER BY 
			CASE WHEN status = 'winner' THEN 0
			     WHEN status = 'active' THEN 1 
			     WHEN status = 'eliminated' THEN 2
			     WHEN status = 'taken' THEN 3
			     WHEN status = 'available' THEN 4 END,
			CASE WHEN position IS NOT NULL THEN position ELSE 999 END,
			CASE WHEN seed IS NOT NULL THEN seed ELSE 999 END,
			CASE WHEN number IS NOT NULL THEN number ELSE 999 END,
			name
	`, compID)
	if err != nil {
		log.Printf("Error querying entries: %v", err)
		sendError(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	entries := []map[string]interface{}{}
	for rows.Next() {
		var id, compID int
		var name, status string
		var seed, number, position sql.NullInt64
		var eliminatedDate sql.NullString

		err := rows.Scan(&id, &compID, &name, &seed, &number, &status, &eliminatedDate, &position)
		if err != nil {
			log.Printf("Error scanning entry row: %v", err)
			continue
		}

		entry := map[string]interface{}{
			"id":             id,
			"competition_id": compID,
			"name":           name,
			"status":         status,
		}

		if seed.Valid {
			entry["seed"] = int(seed.Int64)
		}
		if number.Valid {
			entry["number"] = int(number.Int64)
		}
		if eliminatedDate.Valid {
			entry["eliminated_date"] = eliminatedDate.String
		}
		if position.Valid {
			entry["position"] = int(position.Int64)
		}

		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func uploadEntriesHandler(w http.ResponseWriter, r *http.Request) {
	compID := r.FormValue("competition_id")

	var compType string
	err := db.QueryRow("SELECT type FROM competitions WHERE id = ?", compID).Scan(&compType)
	if err != nil {
		sendError(w, "Competition not found", 404)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		sendError(w, "No file uploaded", 400)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		sendError(w, "Invalid CSV file", 400)
		return
	}

	count := 0
	errors := []string{}

	for i, record := range records {
		if i == 0 {
			log.Printf("Header row: %v", record)
			continue
		}

		if len(record) < 1 {
			continue
		}

		name := strings.TrimSpace(record[0])
		if name == "" {
			continue
		}

		var seed, number *int

		if compType == "knockout" && len(record) > 1 && record[1] != "" {
			s, err := strconv.Atoi(strings.TrimSpace(record[1]))
			if err == nil {
				seed = &s
			}
		}

		if compType == "race" && len(record) > 1 && record[1] != "" {
			n, err := strconv.Atoi(strings.TrimSpace(record[1]))
			if err == nil {
				number = &n
			}
		}

		_, err := db.Exec(`
			INSERT INTO entries (competition_id, name, seed, number, status)
			VALUES (?, ?, ?, ?, 'available')
		`, compID, name, seed, number)

		if err != nil {
			errMsg := fmt.Sprintf("Row %d (%s): %v", i, name, err)
			errors = append(errors, errMsg)
		} else {
			count++
		}
	}

	response := fmt.Sprintf("%d entries uploaded successfully", count)
	if len(errors) > 0 {
		response += fmt.Sprintf(", %d errors", len(errors))
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(response))
}

func updateEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req Entry
	json.NewDecoder(r.Body).Decode(&req)

	// Prevent changing taken entries back to available
	if req.Status == "available" {
		var currentStatus string
		db.QueryRow("SELECT status FROM entries WHERE id = ?", id).Scan(&currentStatus)
		if currentStatus == "taken" {
			sendError(w, "Cannot change a picked entry back to available. The entry has been selected by a user.", 400)
			return
		}
	}

	_, err := db.Exec(`
		UPDATE entries 
		SET name = ?, seed = ?, number = ?, status = ?
		WHERE id = ?
	`, req.Name, req.Seed, req.Number, req.Status, id)

	if err != nil {
		sendError(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateEntryPositionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	var req struct {
		EntryID  int  `json:"entry_id"`
		Position *int `json:"position"` // null to clear, or 1-5, 999
	}
	json.NewDecoder(r.Body).Decode(&req)

	_, err := db.Exec(`
		UPDATE entries 
		SET position = ?
		WHERE id = ? AND competition_id = ?
	`, req.Position, req.EntryID, compID)

	if err != nil {
		sendError(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		sendError(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getAvailableCountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	var count int
	db.QueryRow(`
		SELECT COUNT(*) FROM entries 
		WHERE competition_id = ? AND status = 'available'
	`, compID).Scan(&count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// ===== BLIND BOX HANDLERS =====

func getBlindBoxesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]
	
	// Get user email from context (set by auth middleware)
	userEmail := r.Context().Value("user_email").(string)

	// Check if user already has a selection
	var existingCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM draws 
		WHERE user_email = ? AND competition_id = ?
	`, userEmail, compID).Scan(&existingCount)

	if err != nil {
		sendError(w, "Database error", 500)
		return
	}

	if existingCount > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}

	// Get count of available entries
	var totalAvailable int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM entries 
		WHERE competition_id = ? AND status = 'available'
	`, compID).Scan(&totalAvailable)

	if err != nil {
		sendError(w, "Database error", 500)
		return
	}

	// Return anonymous boxes
	boxes := []map[string]interface{}{}
	for i := 1; i <= totalAvailable; i++ {
		boxes = append(boxes, map[string]interface{}{
			"box_number": i,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxes)
}

func chooseBlindBoxHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	// Get user info from context
	userEmail := r.Context().Value("user_email").(string)

	var req struct {
		BoxNumber int `json:"box_number"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	tx, err := db.Begin()
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer tx.Rollback()

	// Check if user already has an entry
	var existingCount int
	tx.QueryRow("SELECT COUNT(*) FROM draws WHERE user_email = ? AND competition_id = ?", userEmail, compID).Scan(&existingCount)
	if existingCount > 0 {
		sendError(w, "You already have an entry", 400)
		return
	}

	// Get available entries in order
	rows, err := tx.Query(`
		SELECT id FROM entries 
		WHERE competition_id = ? AND status = 'available'
		ORDER BY id
	`, compID)
	if err != nil {
		sendError(w, "Failed to fetch entries", 500)
		return
	}
	defer rows.Close()

	var availableIDs []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		availableIDs = append(availableIDs, id)
	}

	if req.BoxNumber < 1 || req.BoxNumber > len(availableIDs) {
		sendError(w, "Invalid box number", 400)
		return
	}

	selectedEntryID := availableIDs[req.BoxNumber-1]

	// Create draw
	_, err = tx.Exec(`
		INSERT INTO draws (user_email, competition_id, entry_id)
		VALUES (?, ?, ?)
	`, userEmail, compID, selectedEntryID)
	if err != nil {
		sendError(w, "Failed to assign entry", 500)
		return
	}

	// Mark entry as taken
	_, err = tx.Exec("UPDATE entries SET status = 'taken' WHERE id = ?", selectedEntryID)
	if err != nil {
		sendError(w, "Failed to update entry", 500)
		return
	}

	if err = tx.Commit(); err != nil {
		sendError(w, "Failed to complete selection", 500)
		return
	}

	// Return the selected entry details
	var entryName string
	var seed, number sql.NullInt64
	db.QueryRow("SELECT name, seed, number FROM entries WHERE id = ?", selectedEntryID).Scan(&entryName, &seed, &number)

	result := map[string]interface{}{
		"entry_id":   selectedEntryID,
		"entry_name": entryName,
	}
	if seed.Valid {
		result["seed"] = int(seed.Int64)
	}
	if number.Valid {
		result["number"] = int(number.Int64)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func randomPickHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	// Get user info from context
	userEmail := r.Context().Value("user_email").(string)

	tx, err := db.Begin()
	if err != nil {
		sendError(w, "Database error", 500)
		return
	}
	defer tx.Rollback()

	// Check if user already has an entry
	var existingCount int
	tx.QueryRow("SELECT COUNT(*) FROM draws WHERE user_email = ? AND competition_id = ?", userEmail, compID).Scan(&existingCount)
	if existingCount > 0 {
		sendError(w, "You already have an entry", 400)
		return
	}

	// Get a random available entry
	var selectedEntryID int
	err = tx.QueryRow(`
		SELECT id FROM entries 
		WHERE competition_id = ? AND status = 'available'
		ORDER BY RANDOM()
		LIMIT 1
	`, compID).Scan(&selectedEntryID)

	if err != nil {
		sendError(w, "No available entries", 400)
		return
	}

	// Create draw
	_, err = tx.Exec(`
		INSERT INTO draws (user_email, competition_id, entry_id)
		VALUES (?, ?, ?)
	`, userEmail, compID, selectedEntryID)
	if err != nil {
		sendError(w, "Failed to assign entry", 500)
		return
	}

	// Mark entry as taken
	_, err = tx.Exec("UPDATE entries SET status = 'taken' WHERE id = ?", selectedEntryID)
	if err != nil {
		sendError(w, "Failed to update entry", 500)
		return
	}

	if err = tx.Commit(); err != nil {
		sendError(w, "Failed to complete selection", 500)
		return
	}

	// Return the selected entry details
	var entryName string
	var seed, number sql.NullInt64
	db.QueryRow("SELECT name, seed, number FROM entries WHERE id = ?", selectedEntryID).Scan(&entryName, &seed, &number)

	result := map[string]interface{}{
		"entry_id":   selectedEntryID,
		"entry_name": entryName,
	}
	if seed.Valid {
		result["seed"] = int(seed.Int64)
	}
	if number.Valid {
		result["number"] = int(number.Int64)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===== DRAW HANDLERS =====

func getCompetitionDrawsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID := vars["id"]

	rows, err := db.Query(`
		SELECT d.id, d.user_email, d.competition_id, d.entry_id, d.drawn_at,
		       e.name, e.status, e.seed, e.number, e.position
		FROM draws d
		JOIN entries e ON d.entry_id = e.id
		WHERE d.competition_id = ?
		ORDER BY 
			CASE e.status 
				WHEN 'winner' THEN 0
				WHEN 'active' THEN 1 
				WHEN 'eliminated' THEN 2 
			END,
			CASE WHEN e.position IS NOT NULL THEN e.position ELSE 999 END,
			d.user_email
	`, compID)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	draws := []map[string]interface{}{}
	for rows.Next() {
		var id, compID, entryID int
		var userEmail, entryName, status string
		var drawnAt time.Time
		var seed, number, position sql.NullInt64

		err := rows.Scan(&id, &userEmail, &compID, &entryID, &drawnAt,
			&entryName, &status, &seed, &number, &position)

		if err != nil {
			continue
		}

		draw := map[string]interface{}{
			"id":             id,
			"user_email":     userEmail,
			"competition_id": compID,
			"entry_id":       entryID,
			"entry_name":     entryName,
			"entry_status":   status,
			"drawn_at":       drawnAt,
		}

		if seed.Valid {
			draw["seed"] = int(seed.Int64)
		}
		if number.Valid {
			draw["number"] = int(number.Int64)
		}
		if position.Valid {
			draw["position"] = int(position.Int64)
		}

		draws = append(draws, draw)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draws)
}

func getUserDrawsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user email from context
	userEmail := r.Context().Value("user_email").(string)
	compID := r.URL.Query().Get("competition_id")

	query := `
		SELECT d.id, d.user_email, d.competition_id, d.entry_id, d.drawn_at,
		       e.name, e.status, c.status as comp_status, e.seed, e.number
		FROM draws d
		JOIN entries e ON d.entry_id = e.id
		JOIN competitions c ON d.competition_id = c.id
		WHERE d.user_email = ?
	`
	args := []interface{}{userEmail}

	if compID != "" {
		query += " AND d.competition_id = ?"
		args = append(args, compID)
	}

	query += " ORDER BY d.drawn_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		sendError(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	draws := []map[string]interface{}{}
	for rows.Next() {
		var id, compID, entryID int
		var userEmail, entryName, status, compStatus string
		var drawnAt time.Time
		var seed, number sql.NullInt64

		err := rows.Scan(&id, &userEmail, &compID, &entryID, &drawnAt,
			&entryName, &status, &compStatus, &seed, &number)

		if err != nil {
			continue
		}

		draw := map[string]interface{}{
			"id":             id,
			"competition_id": compID,
			"entry_id":       entryID,
			"drawn_at":       drawnAt,
			"user_email":     userEmail,
			"entry_name":     entryName,
			"entry_status":   status,
		}

		if seed.Valid {
			draw["seed"] = int(seed.Int64)
		}
		if number.Valid {
			draw["number"] = int(number.Int64)
		}

		draws = append(draws, draw)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draws)
}

// ===== SELECTION LOCK HANDLERS =====

func acquireSelectionLockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID, _ := strconv.Atoi(vars["id"])

	// Get user info from context
	userEmail := r.Context().Value("user_email").(string)
	userName := r.Context().Value("user_name").(string)

	lockMutex.Lock()
	defer lockMutex.Unlock()

	if lock, exists := selectionLocks[compID]; exists {
		if time.Since(lock.LockedAt) < 2*time.Minute {
			if lock.UserEmail == userEmail {
				lock.LockedAt = time.Now()
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]bool{"acquired": true})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"acquired":   false,
				"locked_by":  lock.UserName,
				"locked_at":  lock.LockedAt,
				"locked_for": int(time.Since(lock.LockedAt).Seconds()),
			})
			return
		}
		delete(selectionLocks, compID)
	}

	selectionLocks[compID] = &SelectionLock{
		UserEmail:     userEmail,
		UserName:      userName,
		LockedAt:      time.Now(),
		CompetitionID: compID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"acquired": true})
}

func releaseSelectionLockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID, _ := strconv.Atoi(vars["id"])

	// Get user email from context
	userEmail := r.Context().Value("user_email").(string)

	lockMutex.Lock()
	defer lockMutex.Unlock()

	if lock, exists := selectionLocks[compID]; exists {
		if lock.UserEmail == userEmail {
			delete(selectionLocks, compID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func checkSelectionLockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	compID, _ := strconv.Atoi(vars["id"])
	
	// Get user email from context
	userEmail := r.Context().Value("user_email").(string)

	lockMutex.Lock()
	defer lockMutex.Unlock()

	if lock, exists := selectionLocks[compID]; exists {
		if time.Since(lock.LockedAt) >= 2*time.Minute {
			delete(selectionLocks, compID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"locked": false,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"locked":     true,
			"locked_by":  lock.UserName,
			"locked_at":  lock.LockedAt,
			"is_me":      lock.UserEmail == userEmail,
			"locked_for": int(time.Since(lock.LockedAt).Seconds()),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"locked": false,
	})
}

// ===== UTILITY FUNCTIONS =====

func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// Helper to get user info from context (set by auth middleware)
func getUserFromContext(ctx context.Context) (email string, name string, isAdmin bool) {
	if val := ctx.Value("user_email"); val != nil {
		email = val.(string)
	}
	if val := ctx.Value("user_name"); val != nil {
		name = val.(string)
	}
	if val := ctx.Value("is_admin"); val != nil {
		isAdmin = val.(bool)
	}
	return
}
