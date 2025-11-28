package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/SanthoshCheemala/FLARE/backend/internal/auth"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Ensure data directory exists
	os.MkdirAll("./data", 0755)

	dbPath := "./data/flare_server.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create or Update admin user
	adminEmail := "authority_admin@flare.local"
	password := "authority123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if user exists (using the new email)
	existingUser, err := repo.GetUserByEmail(context.Background(), adminEmail)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("Failed to check existing user: %v", err)
	}



	if existingUser != nil {
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

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Successfully created admin user in %s: %s / %s", dbPath, adminEmail, password)
}
