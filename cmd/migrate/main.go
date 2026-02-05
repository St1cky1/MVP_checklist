package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Fatalf("Failed to glob migrations: %v", err)
	}
	sort.Strings(files)

	for _, file := range files {
		log.Printf("Applying migration: %s", file)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration %s: %v", file, err)
		}

		// Split by semicolon for basic execution, though pgx can handle multiple statements in some cases
		_, err = conn.Exec(ctx, string(content))
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				log.Printf("Migration %s already applied (or partial), skipping error: %v", file, err)
				continue
			}
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}
	}

	log.Println("Migrations applied successfully!")
}
