package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/smart-payment-infrastructure/internal/config"
	"github.com/smart-payment-infrastructure/pkg/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse command line flags
	var (
		action         = flag.String("action", "up", "Migration action: up, down, version, seed, clear, reset, fix")
		migrationsPath = flag.String("migrations", "migrations", "Path to migrations directory")
	)
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	postgres, err := database.NewPostgresConnection(cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgres.Close()

	// Connect to MongoDB (optional for Supabase-only setup)
	var mongo interface{}
	if cfg.Database.MongoURL != "" {
		mongoConn, err := database.NewMongoConnection(cfg.Database.MongoURL, "smart_payment")
		if err != nil {
			log.Printf("Warning: Failed to connect to MongoDB: %v", err)
			log.Println("Continuing with PostgreSQL-only operations...")
		} else {
			mongo = mongoConn
			defer mongoConn.Close()
		}
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(*migrationsPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for migrations: %v", err)
	}

	switch *action {
	case "up":
		migrator, err := database.NewMigrator(postgres.DB, absPath)
		if err != nil {
			log.Fatalf("Failed to create migrator: %v", err)
		}
		defer migrator.Close()

		if err := migrator.Up(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

	case "down":
		migrator, err := database.NewMigrator(postgres.DB, absPath)
		if err != nil {
			log.Fatalf("Failed to create migrator: %v", err)
		}
		defer migrator.Close()

		if err := migrator.Down(); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}

	case "version":
		migrator, err := database.NewMigrator(postgres.DB, absPath)
		if err != nil {
			log.Fatalf("Failed to create migrator: %v", err)
		}
		defer migrator.Close()

		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		log.Printf("Current migration version: %d (dirty: %v)", version, dirty)

	case "seed":
		if mongo == nil {
			log.Println("Warning: MongoDB not available, skipping seed operation")
			log.Println("For PostgreSQL-only seeding, you can run SQL scripts manually")
			return
		}
		mongoConn := mongo.(*database.MongoDB)
		seeder := database.NewSeeder(postgres, mongoConn)
		if err := seeder.SeedDevelopmentData(); err != nil {
			log.Fatalf("Failed to seed development data: %v", err)
		}

	case "clear":
		if mongo == nil {
			log.Println("Warning: MongoDB not available, skipping clear operation")
			return
		}
		mongoConn := mongo.(*database.MongoDB)
		seeder := database.NewSeeder(postgres, mongoConn)
		if err := seeder.ClearDevelopmentData(); err != nil {
			log.Fatalf("Failed to clear development data: %v", err)
		}

	case "reset":
		log.Println("🧹 Resetting database for fresh setup...")
		// Drop all tables and recreate
		if err := resetDatabase(postgres.DB); err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}
		log.Println("🎉 Database reset successfully!")

	case "fix":
		log.Println("🔧 Fixing dirty migration state...")
		// Fix dirty migration state
		if err := fixMigrationState(postgres.DB); err != nil {
			log.Fatalf("Failed to fix migration state: %v", err)
		}
		log.Println("✅ Migration state fixed successfully!")

	default:
		log.Fatalf("Unknown action: %s. Available actions: up, down, version, seed, clear, reset, fix", *action)
	}

	log.Println("Database operation completed successfully")
}

// resetDatabase drops all tables and recreates the database
func resetDatabase(db *sql.DB) error {
	// Get all existing tables dynamically
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		return fmt.Errorf("failed to query existing tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("Error scanning table name: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	// Drop all existing tables
	for _, table := range tables {
		_, err = db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("✅ Dropped table: %s", table)
		}
	}

	return nil
}

// fixMigrationState fixes dirty migration state
func fixMigrationState(db *sql.DB) error {
	// Get current migration version
	var currentVersion uint
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}

	// Fix dirty migration state for current version
	_, err = db.Exec("UPDATE schema_migrations SET dirty = false WHERE version = $1", currentVersion)
	if err != nil {
		return fmt.Errorf("failed to fix dirty state for version %d: %w", currentVersion, err)
	}

	log.Printf("Fixed dirty state for migration version %d", currentVersion)
	return nil
}
