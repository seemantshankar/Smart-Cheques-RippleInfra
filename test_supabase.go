package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Fatal("POSTGRES_URL environment variable not set")
	}

	fmt.Printf("Attempting to connect to: %s\n", dbURL)

	// Try to connect
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Set connection timeout
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test the connection with timeout
	fmt.Println("Testing connection with 10 second timeout...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Successfully connected to Supabase database!")

	// Test a simple query
	rows, err := db.Query("SELECT version()")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	var version string
	for rows.Next() {
		if err := rows.Scan(&version); err != nil {
			log.Fatalf("Failed to scan version: %v", err)
		}
	}

	fmt.Printf("Database version: %s\n", version)
	fmt.Println("✅ Database connection test completed successfully!")
}
