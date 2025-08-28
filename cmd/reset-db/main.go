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

	log.Println("🧹 Cleaning database for fresh PMF setup...")

	// First, get all existing tables dynamically
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		log.Printf("Warning: Failed to query existing tables: %v", err)
	} else {
		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Printf("Error scanning table name: %v", err)
				continue
			}
			tables = append(tables, tableName)
		}
		rows.Close()

		// Drop all existing tables
		for _, table := range tables {
			_, err = db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
			if err != nil {
				log.Printf("Warning: Failed to drop table %s: %v", table, err)
			} else {
				log.Printf("✅ Dropped table: %s", table)
			}
		}
	}

	// Also drop our known application tables (in case the query missed any)
	tables := []string{
		"refresh_tokens", // Supabase auth table
		"milestones",
		"audit_logs",
		"contracts",
		"smart_checks",
		"enterprise_documents",
		"authorized_representatives",
		"wallets",
		"enterprises",
		"users",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err = db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("✅ Dropped table: %s", table)
		}
	}

	log.Println("🎉 Database cleaned successfully!")
	log.Println("Ready for fresh migrations!")
}
