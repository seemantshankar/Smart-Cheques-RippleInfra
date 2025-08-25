package main

import (
	"flag"
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
		action = flag.String("action", "up", "Migration action: up, down, version, seed, clear")
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

	// Connect to MongoDB
	mongo, err := database.NewMongoConnection(cfg.Database.MongoURL, "smart_payment")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongo.Close()

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
		seeder := database.NewSeeder(postgres, mongo)
		if err := seeder.SeedDevelopmentData(); err != nil {
			log.Fatalf("Failed to seed development data: %v", err)
		}

	case "clear":
		seeder := database.NewSeeder(postgres, mongo)
		if err := seeder.ClearDevelopmentData(); err != nil {
			log.Fatalf("Failed to clear development data: %v", err)
		}

	default:
		log.Fatalf("Unknown action: %s. Available actions: up, down, version, seed, clear", *action)
	}

	log.Println("Database operation completed successfully")
}