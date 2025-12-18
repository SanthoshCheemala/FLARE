package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"

	"github.com/SanthoshCheemala/FLARE/backend/internal/auth"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Ensure data directory exists
	os.MkdirAll("./data", 0755)
	// Use absolute path or path relative to where server runs from
	dbPath := "../../data/flare_server.db" // From cmd/seed_server → backend/data
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// 1. Clear existing sanction data
	log.Println("Clearing existing sanction data...")
	if _, err := db.Exec("DELETE FROM sanctions"); err != nil {
		log.Fatalf("Failed to clear sanctions: %v", err)
	}
	if _, err := db.Exec("DELETE FROM sanction_lists"); err != nil {
		log.Fatalf("Failed to clear sanction lists: %v", err)
	}

	// 2. Seed Default List
	csvPath := "../data/server_data_small.csv" // Relative to cmd/seed_server
	// Check if file exists, if not try absolute path or other relative path
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		csvPath = "./data/server_data_small.csv" // Try from root
		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			// Try absolute path based on user context
			csvPath = "/Users/santhoshcheemala/ALL_IN_ONE/Research_Implimentation/FLARE/backend/data/server_data_small.csv"
		}
	}

	log.Printf("Seeding default sanction list from: %s", csvPath)
	
	// Create List Entry
	res, err := db.Exec(`
		INSERT INTO sanction_lists (name, source, description, file_path, record_count, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, "server_data_small.csv", "System", "Default pre-loaded sanctions list", csvPath, 0)
	if err != nil {
		log.Fatalf("Failed to create sanction list: %v", err)
	}
	listID, _ := res.LastInsertId()

	// Read and Insert Records
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read() // Skip header
	if err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}

	// Map headers
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		name := record[headerMap["name"]]
		dob := record[headerMap["dob"]]
		country := record[headerMap["country"]]
		program := record[headerMap["sanction_program"]]
		
		// Get the psi_key column value (pre-computed normalized serialization)
		// The psi_key column already contains the format: "name|dob|country"
		// But country codes might be uppercase, so we normalize to lowercase
		psiKey := strings.ToLower(record[headerMap["psi_key"]])
		
		// Hash the psi_key (now fully normalized)
		hashBytes := sha256.Sum256([]byte(psiKey))
		hash := binary.BigEndian.Uint64(hashBytes[:8])

		_, err = db.Exec(`
			INSERT INTO sanctions (name, dob, country, program, source, list_id, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, name, dob, country, program, "System", listID, int64(hash))
		
		if err != nil {
			log.Printf("Failed to insert sanction: %v", err)
		} else {
			count++
		}
	}

	// Update count
	_, err = db.Exec("UPDATE sanction_lists SET record_count = ? WHERE id = ?", count, listID)
	if err != nil {
		log.Fatalf("Failed to update count: %v", err)
	}

	log.Printf("Seeded %d sanctions.", count)

	// 3. Create/Update Admin User
	adminEmail := "authority_admin@flare.local"
	password := "authority123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if user exists
	var existingID int64
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", adminEmail).Scan(&existingID)
	
	if err == nil {
		log.Printf("Updating existing user %s to AUTHORITY_ADMIN", adminEmail)
		_, err = db.Exec(`
			UPDATE users SET role = ?, password_hash = ?, active = ? WHERE email = ?
		`, "AUTHORITY_ADMIN", hash, true, adminEmail)
	} else {
		log.Printf("Creating new user %s", adminEmail)
		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, role, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, adminEmail, hash, "AUTHORITY_ADMIN", true)
	}

	log.Printf("Successfully initialized server DB.")
}
