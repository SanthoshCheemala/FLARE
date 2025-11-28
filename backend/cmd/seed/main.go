package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/SanthoshCheemala/FLARE/backend/internal/auth"
	"github.com/SanthoshCheemala/FLARE/backend/internal/config"
	"github.com/SanthoshCheemala/FLARE/backend/internal/repository"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure data directory exists
	if cfg.DatabaseDriver() == "sqlite3" {
		os.MkdirAll("./data", 0755)
	}

	db, err := sql.Open(cfg.DatabaseDriver(), cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	if err := repo.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create or Update admin user
	adminEmail := "bank_admin@flare.local"
	password := "bank123"
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
		log.Printf("Updating existing user %s to BANK_ADMIN", adminEmail)
		_, err = db.Exec(`
			UPDATE users SET role = ?, password_hash = ?, active = ? WHERE email = ?
		`, "BANK_ADMIN", hash, true, adminEmail)
	} else {
		log.Printf("Creating new user %s", adminEmail)
		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, role, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, adminEmail, hash, "BANK_ADMIN", true)
	}

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Successfully created admin user: %s / %s", adminEmail, password)
}
