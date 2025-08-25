package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get the database URL from environment or command line
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:nujMat-tykzos-xahqo5@db.hpoauyaopahsdxeryjlb.supabase.co:5432/postgres?sslmode=require"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Fix dirty migration state
	_, err = db.Exec("UPDATE schema_migrations SET dirty = false WHERE version = 4")
	if err != nil {
		log.Fatalf("Failed to fix dirty state: %v", err)
	}

	log.Println("Successfully fixed dirty migration state")
}
