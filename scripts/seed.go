//go:build ignore

// seed.go creates the initial admin user in the blog database.
// Run with: go run ./scripts/seed.go -username admin -password "YourPassword" -email "you@example.com"
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	username := flag.String("username", "", "Admin username (required)")
	password := flag.String("password", "", "Admin password (required)")
	email    := flag.String("email", "", "Admin email (optional)")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/seed.go -username <u> -password <p> [-email <e>]")
		os.Exit(1)
	}

	// Load .env if present
	_ = godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/blog.db"
	}

	if err := os.MkdirAll("./data", 0o755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	// Ensure table exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS admin_users (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		username      TEXT     NOT NULL UNIQUE,
		password_hash TEXT     NOT NULL,
		email         TEXT     NOT NULL DEFAULT '',
		created_at    DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
		updated_at    DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO admin_users (username, password_hash, email) VALUES (?, ?, ?)
		 ON CONFLICT(username) DO UPDATE SET password_hash=excluded.password_hash, email=excluded.email`,
		*username, string(hash), *email)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	fmt.Printf("✓ Admin user '%s' created/updated in %s\n", *username, dbPath)
}
