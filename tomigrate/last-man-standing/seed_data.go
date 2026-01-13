package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const IDENTITY_SERVICE = "http://localhost:3001"

func main() {
	db, err := sql.Open("sqlite3", "./data/lastmanstanding.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rand.Seed(time.Now().UnixNano())

	// Create users in Identity Service
	users := []struct {
		email    string
		name     string
		password string
		isAdmin  bool
	}{
		{"1.andy.c.harris@gmail.com", "1.Andy", "PLAYER01", false},
		{"andr3wharr1s@gmail.com", "andr3w", "PLAYER02", false},
		{"andr3wharr1s@googlemail.com", "andr3wharr1s", "PLAYER03", false},
	}

	userIDs := make(map[string]int)
	
	fmt.Println("Creating users in Identity Service...")
	for _, u := range users {
		// Register user with Identity Service
		regData := map[string]string{
			"email":    u.email,
			"name":     u.name,
			"password": u.password,
		}
		jsonData, _ := json.Marshal(regData)
		
		resp, err := http.Post(
			IDENTITY_SERVICE+"/api/register",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		
		if err != nil {
			log.Printf("Error creating user %s: %v", u.name, err)
			continue
		}
		
		if resp.StatusCode != http.StatusOK {
			body := make([]byte, 1024)
			resp.Body.Read(body)
			log.Printf("Failed to create user %s: %s", u.name, string(body))
			resp.Body.Close()
			continue
		}
		
		var result struct {
			User struct {
				ID    int    `json:"id"`
				Email string `json:"email"`
				Name  string `json:"name"`
			} `json:"user"`
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		
		userIDs[u.name] = result.User.ID
		fmt.Printf("✅ Created user: %s (ID: %d, Password: %s)\n", u.name, result.User.ID, u.password)
		
		// If admin, update role in Identity Service
		if u.isAdmin {
			// Note: You'd need to manually promote to admin in Identity Service
			// Or use the admin backdoor to call admin APIs
			fmt.Printf("   ℹ️  Manually promote %s to admin in Identity Service if needed\n", u.name)
		}
	}

	// Create 4 games
	gameNames := []string{"Game 1", "Game 2", "Game 3", "Game 4"}
	gameIDs := []int{}

	fmt.Println("\nCreating games...")
	for _, gameName := range gameNames {
		result, err := db.Exec("INSERT INTO games (name, status, postponement_rule) VALUES (?, 'active', 'loss')", gameName)
		if err != nil {
			log.Printf("Error creating game %s: %v", gameName, err)
			continue
		}
		gameID, _ := result.LastInsertId()
		gameIDs = append(gameIDs, int(gameID))
		fmt.Printf("✅ Created game: %s (ID: %d)\n", gameName, gameID)
	}

	// Set first game as current
	if len(gameIDs) > 0 {
		db.Exec("UPDATE games SET is_current = 1 WHERE id = ?", gameIDs[0])
		fmt.Printf("✅ Set Game 1 as current game\n")
	}

	// Match data
	matches := []struct {
		matchNum int
		round    int
		date     string
		location string
		home     string
		away     string
	}{
		{67, 12, "11/01/2026 12:00", "Joie Stadium", "Manchester City", "Everton"},
		{68, 12, "11/01/2026 14:00", "Emirates Stadium", "Arsenal", "Manchester United"},
		{69, 12, "11/01/2026 14:00", "Villa Park", "Aston Villa", "Brighton & Hove Albion"},
		{70, 12, "11/01/2026 14:00", "Kingsmeadow", "Chelsea", "West Ham United"},
		{71, 12, "11/01/2026 14:00", "St Helens Stadium", "Liverpool", "London City Lionesses"},
		{72, 12, "11/01/2026 14:00", "Brisbane Road", "Tottenham Hotspur", "Leicester City"},
		{74, 13, "24/01/2026 12:30", "Stamford Bridge", "Chelsea", "Arsenal"},
		{73, 13, "25/01/2026 14:00", "Villa Park", "Aston Villa", "Manchester United"},
		{75, 13, "25/01/2026 14:00", "Goodison Park", "Everton", "Brighton & Hove Albion"},
		{76, 13, "25/01/2026 14:00", "St Helens Stadium", "Liverpool", "Tottenham Hotspur"},
		{77, 13, "25/01/2026 14:00", "The CopperJax Community Stadium", "London City Lionesses", "Manchester City"},
		{78, 13, "25/01/2026 15:00", "King Power Stadium", "Leicester City", "West Ham United"},
		{79, 14, "31/01/2026 12:30", "Emirates Stadium", "Arsenal", "Leicester City"},
		{80, 14, "01/02/2026 12:00", "Chigwell Construction Stadium", "West Ham United", "Tottenham Hotspur"},
		{81, 14, "01/02/2026 13:00", "Leigh Sports Village Stadium", "Manchester United", "Liverpool"},
		{82, 14, "01/02/2026 14:00", "Broadfield Stadium", "Brighton & Hove Albion", "London City Lionesses"},
		{83, 14, "01/02/2026 14:00", "Goodison Park", "Everton", "Aston Villa"},
		{84, 14, "01/02/2026 14:30", "Etihad Stadium", "Manchester City", "Chelsea"},
		{85, 15, "08/02/2026 12:00", "The CopperJax Community Stadium", "London City Lionesses", "Everton"},
		{86, 15, "08/02/2026 12:00", "Chigwell Construction Stadium", "West Ham United", "Brighton & Hove Albion"},
		{87, 15, "08/02/2026 14:00", "Emirates Stadium", "Arsenal", "Manchester City"},
		{88, 15, "08/02/2026 14:00", "St Helens Stadium", "Liverpool", "Aston Villa"},
		{89, 15, "08/02/2026 14:00", "Tottenham Hotspur Stadium", "Tottenham Hotspur", "Chelsea"},
		{90, 15, "08/02/2026 15:00", "King Power Stadium", "Leicester City", "Manchester United"},
		{91, 16, "15/02/2026 12:00", "Joie Stadium", "Manchester City", "Leicester City"},
		{92, 16, "15/02/2026 13:00", "Leigh Sports Village Stadium", "Manchester United", "London City Lionesses"},
		{93, 16, "15/02/2026 14:00", "Villa Park", "Aston Villa", "Tottenham Hotspur"},
		{94, 16, "15/02/2026 14:00", "Broadfield Stadium", "Brighton & Hove Albion", "Arsenal"},
		{95, 16, "15/02/2026 14:00", "Kingsmeadow", "Chelsea", "Liverpool"},
		{96, 16, "15/02/2026 14:00", "Goodison Park", "Everton", "West Ham United"},
	}

	// Create rounds and matches for each game
	fmt.Println("\nCreating rounds and matches...")
	for _, gameID := range gameIDs {
		// Create rounds 12-16
		for round := 12; round <= 16; round++ {
			// Set deadline 30 days from now for testing
			deadline := time.Now().Add(time.Duration(round) * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
			
			_, err := db.Exec("INSERT OR IGNORE INTO rounds (game_id, round_number, submission_deadline, status) VALUES (?, ?, ?, 'draft')",
				gameID, round, deadline)
			if err != nil {
				log.Printf("Error creating round %d for game %d: %v", round, gameID, err)
			}
		}
		fmt.Printf("✅ Created rounds 12-16 for Game ID %d\n", gameID)

		// Insert matches
		for _, m := range matches {
			_, err := db.Exec(`INSERT OR IGNORE INTO matches 
				(game_id, match_number, round_number, date, location, home_team, away_team, result, status) 
				VALUES (?, ?, ?, ?, ?, ?, ?, '', 'upcoming')`,
				gameID, m.matchNum, m.round, m.date, m.location, m.home, m.away)
			if err != nil {
				log.Printf("Error inserting match %d: %v", m.matchNum, err)
			}
		}
		fmt.Printf("✅ Inserted %d matches for Game ID %d\n", len(matches), gameID)
	}

	// Players join all games
	fmt.Println("\nAdding players to games...")
	for _, gameID := range gameIDs {
		for userName, userID := range userIDs {
			if userName == "Andrew" {
				continue // Skip admin
			}
			_, err := db.Exec("INSERT OR IGNORE INTO game_players (user_id, game_id, is_active) VALUES (?, ?, 1)",
				userID, gameID)
			if err != nil {
				log.Printf("Error adding %s to game %d: %v", userName, gameID, err)
			}
		}
		fmt.Printf("✅ Added 3 players to Game ID %d\n", gameID)
	}

	// Create predictions for each game
	fmt.Println("\nCreating predictions...")
	for _, gameID := range gameIDs {
		// Track used teams per player for this game
		usedTeams := make(map[string]map[string]bool) // userName -> team -> used
		for userName := range userIDs {
			usedTeams[userName] = make(map[string]bool)
		}

		for round := 12; round <= 16; round++ {
			// Get matches for this round
			roundMatches := []struct {
				id   int
				home string
				away string
			}{}
			
			rows, _ := db.Query("SELECT id, home_team, away_team FROM matches WHERE game_id = ? AND round_number = ?",
				gameID, round)
			for rows.Next() {
				var m struct {
					id   int
					home string
					away string
				}
				rows.Scan(&m.id, &m.home, &m.away)
				roundMatches = append(roundMatches, m)
			}
			rows.Close()

			if len(roundMatches) == 0 {
				continue
			}

			// Each player makes a prediction
			for userName, userID := range userIDs {
				if userName == "Andrew" {
					continue // Skip admin
				}

				// Try to find a match where the player hasn't used either team
				var selectedMatch *struct {
					id   int
					home string
					away string
				}
				var team string

				// Shuffle matches to get variety
				shuffled := make([]int, len(roundMatches))
				for i := range shuffled {
					shuffled[i] = i
				}
				rand.Shuffle(len(shuffled), func(i, j int) {
					shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
				})

				// Try to find an unused team
				for _, idx := range shuffled {
					match := &roundMatches[idx]
					
					if userName == "1.Andy" {
						// 1.Andy always picks home team if not used
						if !usedTeams[userName][match.home] {
							selectedMatch = match
							team = match.home
							break
						}
					} else {
						// Others try home first, then away
						if !usedTeams[userName][match.home] {
							selectedMatch = match
							team = match.home
							break
						} else if !usedTeams[userName][match.away] {
							selectedMatch = match
							team = match.away
							break
						}
					}
				}

				// If we couldn't find an unused team, player is out of teams
				if selectedMatch == nil {
					fmt.Printf("  ! %s has no available teams for Game %d, Round %d\n", userName, gameID, round)
					continue
				}

				_, err := db.Exec(`INSERT OR IGNORE INTO predictions 
					(user_id, game_id, match_id, round_number, predicted_team) 
					VALUES (?, ?, ?, ?, ?)`,
					userID, gameID, selectedMatch.id, round, team)
				
				if err != nil {
					log.Printf("Error creating prediction for %s, game %d, round %d: %v", userName, gameID, round, err)
				} else {
					// Mark this team as used
					usedTeams[userName][team] = true
				}
			}
		}
		fmt.Printf("✅ Created predictions for Game ID %d (rounds 12-16)\n", gameID)
	}

	fmt.Println("\n✅ Test data injection complete!")
	fmt.Println("\nLogin credentials (use these passwords):")
	fmt.Println("Player: 1.andy.c.harris@gmail.com / PLAYER01")
	fmt.Println("Player: andr3wharr1s@gmail.com / PLAYER02")
	fmt.Println("Player: andr3wharr1s@googlemail.com / PLAYER03")
	fmt.Println("\nNote: Users are created in Identity Service, games/matches/predictions in Last Man Standing database")
}
