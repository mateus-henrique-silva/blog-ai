package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(dbPath string) *sql.DB {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("database: failed to create data directory: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("database: failed to open: %v", err)
	}

	db.SetMaxOpenConns(1) // SQLite supports one writer at a time
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		log.Fatalf("database: failed to ping: %v", err)
	}

	return db
}

func RunMigrations(db *sql.DB) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		log.Fatalf("database: failed to read migrations dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := fmt.Sprintf("migrations/%s", entry.Name())
		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			log.Fatalf("database: failed to read migration %s: %v", entry.Name(), err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("database: failed to run migration %s: %v", entry.Name(), err)
		}
		log.Printf("database: applied migration %s", entry.Name())
	}
}
