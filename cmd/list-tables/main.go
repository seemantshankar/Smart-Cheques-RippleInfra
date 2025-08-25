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

	log.Println("📋 Current tables in database:")
	log.Println("==================================================")

	// Get all tables from public schema
	rows, err := db.Query(`
		SELECT table_name, table_type
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	tableCount := 0
	for rows.Next() {
		var tableName, tableType string
		err := rows.Scan(&tableName, &tableType)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		tableCount++
		log.Printf("📊 %s (%s)", tableName, tableType)
	}

	log.Printf("\n✅ Total tables found: %d", tableCount)

	if tableCount == 0 {
		log.Println("🎉 Database is clean!")
	}
}
