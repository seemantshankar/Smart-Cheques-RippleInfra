package testutils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultTestDBConfig returns the default test database configuration
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5432"),
		User:     getEnv("TEST_DB_USER", "user"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		DBName:   getEnv("TEST_DB_NAME", "smart_payment"),
		SSLMode:  getEnv("TEST_DB_SSLMODE", "disable"),
	}
}

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	config := DefaultTestDBConfig()

	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Run migrations if needed
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close test database connection: %v", err)
		}
	})

	return db
}

// SetupTestDBWithCleanup creates a test database and provides cleanup functions
func SetupTestDBWithCleanup(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db := SetupTestDB(t)

	cleanup := func() {
		// Clean up test data
		if err := cleanupTestData(db); err != nil {
			log.Printf("Failed to cleanup test data: %v", err)
		}

		if err := db.Close(); err != nil {
			log.Printf("Failed to close test database connection: %v", err)
		}
	}

	t.Cleanup(cleanup)
	return db, cleanup
}

// runTestMigrations runs migrations on the test database
func runTestMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get migrations directory
	migrationsPath := "migrations"
	if os.Getenv("TEST_MIGRATIONS_PATH") != "" {
		migrationsPath = os.Getenv("TEST_MIGRATIONS_PATH")
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Check current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	// If database is dirty, force it to current version
	if dirty {
		log.Println("Test database is dirty, forcing to current version...")
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
	}

	// Run migrations up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Test database migrations completed")
	return nil
}

// cleanupTestData removes test data from the database
func cleanupTestData(db *sql.DB) error {
	// List of tables to clean (in reverse dependency order)
	tables := []string{
		"categorization_predictions",
		"categorization_training_data",
		"ml_model_metrics",
		"categorization_ml_models",
		"categorization_rule_performance",
		"categorization_rules",
		"categorization_rule_templates",
		"categorization_rule_groups",
		"dispute_evidence",
		"dispute_comments",
		"dispute_notifications",
		"dispute_audit_logs",
		"disputes",
		"fraud_alerts",
		"fraud_patterns",
		"fraud_rules",
		"fraud_incidents",
		"asset_transactions",
		"assets",
		"contract_milestones",
		"contracts",
		"milestone_templates",
		"milestone_repositories",
		"milestones",
		"oracle_data",
		"oracle_sources",
		"smart_cheques",
		"transactions",
		"wallets",
		"audit_logs",
		"enterprises",
		"users",
	}

	// Disable foreign key checks temporarily
	if _, err := db.Exec("SET session_replication_role = replica;"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Clean each table
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s;", table)); err != nil {
			log.Printf("Warning: failed to clean table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := db.Exec("SET session_replication_role = DEFAULT;"); err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *sql.DB, email, password, firstName, lastName, role string) string {
	t.Helper()

	userID := "11111111-1111-1111-1111-111111111111"

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			role = EXCLUDED.role,
			updated_at = NOW()
	`

	// In a real implementation, you'd hash the password
	passwordHash := password // Simplified for testing

	_, err := db.Exec(query, userID, email, passwordHash, firstName, lastName, role)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// CreateTestEnterprise creates a test enterprise in the database
func CreateTestEnterprise(t *testing.T, db *sql.DB, name, industry string) string {
	t.Helper()

	enterpriseID := "22222222-2222-2222-2222-222222222222"

	query := `
		INSERT INTO enterprises (id, name, industry, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			industry = EXCLUDED.industry,
			updated_at = NOW()
	`

	_, err := db.Exec(query, enterpriseID, name, industry)
	if err != nil {
		t.Fatalf("Failed to create test enterprise: %v", err)
	}

	return enterpriseID
}

// CreateTestWallet creates a test wallet in the database
func CreateTestWallet(t *testing.T, db *sql.DB, enterpriseID, currencyCode string) string {
	t.Helper()

	walletID := "33333333-3333-3333-3333-333333333333"

	query := `
		INSERT INTO wallets (id, enterprise_id, currency_code, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, 0, 'active', NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			enterprise_id = EXCLUDED.enterprise_id,
			currency_code = EXCLUDED.currency_code,
			updated_at = NOW()
	`

	_, err := db.Exec(query, walletID, enterpriseID, currencyCode)
	if err != nil {
		t.Fatalf("Failed to create test wallet: %v", err)
	}

	return walletID
}
