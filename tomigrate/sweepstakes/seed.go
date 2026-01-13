package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("sqlite3", "../data/sweepstake.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🌱 Sweepstake Seed Data Script (Simplified)")
	fmt.Println("==========================================")
	fmt.Print("\nClear existing data? (y/n): ")

	var clearData string
	fmt.Scanln(&clearData)

	if clearData == "y" || clearData == "Y" {
		fmt.Println("\n🗑️  Clearing existing data...")
		db.Exec("DELETE FROM draws")
		db.Exec("DELETE FROM entries")
		db.Exec("DELETE FROM competitions")
		db.Exec("DELETE FROM users WHERE is_admin = 0")
		fmt.Println("✅ Data cleared!")
	}

	fmt.Println("\n👥 Creating users...")
	createUsers(db)

	fmt.Println("\n🏆 Creating competitions...")
	createCompetitions(db)

	fmt.Println("\n✅ Seed data complete!")
	fmt.Println("\n📊 Summary:")
	showSummary(db)
}

func createUsers(db *sql.DB) {
	users := []struct {
		email   string
		name    string
		code    string
		isAdmin bool
	}{
		{"andrew_c_harris@outlook.com", "Andy", "36313665", true},
		{"alice@example.com", "Alice", "password123", false},
		{"bob@example.com", "Bob", "password123", false},
		{"charlie@example.com", "Charlie", "password123", false},
		{"diana@example.com", "Diana", "password123", false},
		{"eve@example.com", "Eve", "password123", false},
	}

	for _, u := range users {
		hashedCode, _ := bcrypt.GenerateFromPassword([]byte(u.code), bcrypt.DefaultCost)
		adminFlag := 0
		if u.isAdmin {
			adminFlag = 1
		}

		_, err := db.Exec(`
			INSERT INTO users (email, name, code, is_admin)
			VALUES (?, ?, ?, ?)
		`, u.email, u.name, string(hashedCode), adminFlag)

		if err != nil {
			fmt.Printf("⚠️  User %s already exists\n", u.email)
		} else {
			adminStr := ""
			if u.isAdmin {
				adminStr = " (ADMIN)"
			}
			fmt.Printf("✅ Created: %s - %s%s\n", u.email, u.name, adminStr)
		}
	}
}

func createCompetitions(db *sql.DB) {
	now := time.Now()

	// World Cup 2026 Teams
	worldCupTeams := []struct {
		name string
		seed int
	}{
		{"Brazil", 1}, {"Germany", 2}, {"France", 3}, {"Argentina", 4},
		{"Spain", 5}, {"England", 6}, {"Portugal", 7}, {"Italy", 8},
		{"Netherlands", 9}, {"Belgium", 10}, {"Uruguay", 11}, {"Croatia", 12},
		{"Denmark", 13}, {"Switzerland", 14}, {"Mexico", 15}, {"USA", 16},
		{"Colombia", 17}, {"Chile", 18}, {"Sweden", 19}, {"Poland", 20},
		{"Austria", 21}, {"Wales", 22}, {"Serbia", 23}, {"Ukraine", 24},
		{"Turkey", 25}, {"Czech Republic", 26}, {"Greece", 27}, {"Nigeria", 28},
		{"Egypt", 29}, {"South Korea", 30}, {"Japan", 31}, {"Australia", 32},
	}

	// Champions League Teams
	championsLeagueTeams := []struct {
		name string
		seed int
	}{
		{"Manchester City", 1}, {"Real Madrid", 2}, {"Bayern Munich", 3}, {"PSG", 4},
		{"Barcelona", 5}, {"Liverpool", 6}, {"Inter Milan", 7}, {"Atletico Madrid", 8},
		{"Borussia Dortmund", 9}, {"Napoli", 10}, {"Chelsea", 11}, {"AC Milan", 12},
		{"Benfica", 13}, {"Porto", 14}, {"Ajax", 15}, {"RB Leipzig", 16},
	}

	// Grand National Horses
	grandNationalHorses := []string{
		"Thunder Bolt", "Lightning Strike", "Storm Chaser", "Wind Runner",
		"Fire Dancer", "Ocean Wave", "Mountain King", "Valley Queen",
		"Desert Fox", "Arctic Wolf", "Jungle Cat", "Prairie Star",
		"River Spirit", "Forest Ranger", "Sunset Glory", "Dawn Rider",
		"Midnight Express", "Golden Arrow", "Silver Bullet", "Bronze Medal",
		"Diamond Dust", "Ruby Runner", "Emerald Flash", "Sapphire Sky",
		"Crystal Clear", "Amber Alert", "Topaz Thunder", "Opal Dream",
		"Pearl Harbor", "Jade Warrior", "Onyx Shadow", "Marble Marvel",
		"Comet Tail", "Star Gazer", "Moon Shadow", "Sun Dancer",
		"Cloud Walker", "Sky Pilot", "Earth Shaker", "Fire Storm",
	}

	// Kentucky Derby Horses
	kentuckyDerbyHorses := []string{
		"American Pharoah", "Justify", "Secretariat", "Citation",
		"Seattle Slew", "Affirmed", "Count Fleet", "Triple Crown",
		"War Admiral", "Whirlaway", "Northern Dancer", "Spectacular Bid",
		"Sunday Silence", "Alydar", "Ruffian", "Man o' War",
		"Seabiscuit", "Zenyatta", "California Chrome", "Big Brown",
	}

	// Royal Ascot Horses
	royalAscotHorses := []string{
		"Frankel", "Enable", "Stradivarius", "Golden Horn",
		"Treve", "Sea The Stars", "Duke of Marmalade", "Rip Van Winkle",
		"Ouija Board", "Black Caviar", "Winx", "Kingman",
		"Goldikova", "Paco Boy", "Canford Cliffs", "Timeform",
		"Harbinger", "Workforce", "Nathaniel", "Taghrooda",
		"Jack Hobbs", "Highland Reel", "Found", "Order of St George",
	}

	competitions := []struct {
		name        string
		compType    string
		description string
		dataSource  string
		data        interface{}
	}{
		{
			name:        "World Cup 2026",
			compType:    "knockout",
			description: "FIFA World Cup 2026 - Pick your country and compete for glory! Tournament runs June-July 2026.",
			dataSource:  "football",
			data:        worldCupTeams,
		},
		{
			name:        "UEFA Champions League 2025/26",
			compType:    "knockout",
			description: "Europe's premier club competition. Pick your team and follow them through the knockout stages.",
			dataSource:  "football",
			data:        championsLeagueTeams,
		},
		{
			name:        "Grand National 2026",
			compType:    "race",
			description: "The world's most famous steeplechase at Aintree. Pick your horse and hope it makes it over all 30 fences!",
			dataSource:  "horses",
			data:        grandNationalHorses,
		},
		{
			name:        "Kentucky Derby 2026",
			compType:    "race",
			description: "The most exciting two minutes in sports! Pick your thoroughbred for the Run for the Roses.",
			dataSource:  "horses",
			data:        kentuckyDerbyHorses,
		},
		{
			name:        "Royal Ascot Gold Cup 2026",
			compType:    "race",
			description: "British horse racing's most prestigious event. Pick your horse for this royal racing spectacle.",
			dataSource:  "horses",
			data:        royalAscotHorses,
		},
	}

	for _, comp := range competitions {
		var startDate, endDate *time.Time
		
		if comp.compType == "knockout" {
			start := now.AddDate(0, 0, 7)
			end := now.AddDate(0, 1, 0)
			startDate = &start
			endDate = &end
		} else {
			end := now.AddDate(0, 0, 14)
			endDate = &end
		}

		result, err := db.Exec(`
			INSERT INTO competitions (name, type, status, start_date, end_date, description)
			VALUES (?, ?, 'draft', ?, ?, ?)
		`, comp.name, comp.compType, startDate, endDate, comp.description)

		if err != nil {
			fmt.Printf("⚠️  Failed to create %s: %v\n", comp.name, err)
			continue
		}

		compID, _ := result.LastInsertId()

		// Add entries
		if comp.dataSource == "football" {
			teams := comp.data.([]struct {
				name string
				seed int
			})
			for _, team := range teams {
				db.Exec(`
					INSERT INTO entries (competition_id, name, seed, status)
					VALUES (?, ?, ?, 'available')
				`, compID, team.name, team.seed)
			}
			fmt.Printf("✅ Created: %s (%d teams)\n", comp.name, len(teams))
		} else {
			horses := comp.data.([]string)
			for idx, horse := range horses {
				db.Exec(`
					INSERT INTO entries (competition_id, name, number, status)
					VALUES (?, ?, ?, 'available')
				`, compID, horse, idx+1)
			}
			fmt.Printf("✅ Created: %s (%d horses)\n", comp.name, len(horses))
		}
	}
}

func showSummary(db *sql.DB) {
	var userCount, compCount, entryCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM competitions").Scan(&compCount)
	db.QueryRow("SELECT COUNT(*) FROM entries").Scan(&entryCount)

	fmt.Printf("\n👥 Users: %d\n", userCount)
	fmt.Printf("🏆 Competitions: %d\n", compCount)
	fmt.Printf("🎯 Total Entries: %d\n", entryCount)

	fmt.Println("\n📋 Competition Details:")
	rows, _ := db.Query(`
		SELECT c.name, c.type, c.status, COUNT(e.id) as entry_count
		FROM competitions c
		LEFT JOIN entries e ON c.id = e.competition_id
		GROUP BY c.id
		ORDER BY c.id
	`)
	defer rows.Close()

	for rows.Next() {
		var name, compType, status string
		var entryCount int
		rows.Scan(&name, &compType, &status, &entryCount)
		
		fmt.Printf("  • %s\n", name)
		fmt.Printf("    Type: %s | Status: %s | Entries: %d\n", compType, status, entryCount)
	}

	fmt.Println("\n🔑 Login Credentials:")
	fmt.Println("  Admin:")
	fmt.Println("    Email: andrew_c_harris@outlook.com")
	fmt.Println("    Code: 36313665")
	fmt.Println("\n  Regular Users (all use code: password123):")
	fmt.Println("    • alice@example.com - Alice")
	fmt.Println("    • bob@example.com - Bob")
	fmt.Println("    • charlie@example.com - Charlie")
	fmt.Println("    • diana@example.com - Diana")
	fmt.Println("    • eve@example.com - Eve")
	
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("  1. Login as admin")
	fmt.Println("  2. Go to 'Manage Competitions'")
	fmt.Println("  3. Click 'Open for Users' on competitions you want to release")
	fmt.Println("  4. Users can then pick their mystery boxes!")
}