package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestCreateContract(t *testing.T) {
	// Use real database for integration testing
	// In container environment, connect to postgres service
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error connecting to test database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Logf("Database ping failed: %v", err)
		t.Skip("Test database not available, skipping integration test")
		return
	}

	repo := NewPostgresContractRepository(db)

	now := time.Now()
	c := &models.Contract{
		ID:           "11111111-1111-1111-1111-111111111111",
		Parties:      []string{"A", "B"},
		Status:       "draft",
		ContractType: "service_agreement",
		Version:      "v1",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Clean up any existing test data
	db.Exec("DELETE FROM contracts WHERE id = $1", c.ID)

	if err := repo.CreateContract(context.Background(), c); err != nil {
		t.Fatalf("CreateContract error: %v", err)
	}

	// Verify the contract was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contracts WHERE id = $1", c.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify contract creation: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 contract, got %d", count)
	}

	// Clean up
	db.Exec("DELETE FROM contracts WHERE id = $1", c.ID)
}

func TestGetContractByID(t *testing.T) {
	// Use real database for integration testing
	// In container environment, connect to postgres service
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error connecting to test database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Logf("Database ping failed: %v", err)
		t.Skip("Test database not available, skipping integration test")
		return
	}

	repo := NewPostgresContractRepository(db)

	// Create a test contract first
	now := time.Now()
	testContract := &models.Contract{
		ID:           "11111111-1111-1111-1111-111111111111",
		Parties:      []string{"A", "B"},
		Status:       "draft",
		ContractType: "service_agreement",
		Version:      "v1",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Clean up any existing test data
	db.Exec("DELETE FROM contracts WHERE id = $1", testContract.ID)

	// Create the contract
	if err := repo.CreateContract(context.Background(), testContract); err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}

	// Now test getting it by ID
	c, err := repo.GetContractByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GetContractByID error: %v", err)
	}
	if c == nil || c.ID == "" || len(c.Parties) != 2 {
		t.Fatalf("unexpected contract result: %+v", c)
	}

	// Clean up
	db.Exec("DELETE FROM contracts WHERE id = $1", testContract.ID)
}

func TestUpdateAndDeleteContract(t *testing.T) {
	// Use real database for integration testing
	// In container environment, connect to postgres service
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error connecting to test database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		t.Logf("Database ping failed: %v", err)
		t.Skip("Test database not available, skipping integration test")
		return
	}

	repo := NewPostgresContractRepository(db)
	now := time.Now()
	c := &models.Contract{
		ID:           "11111111-1111-1111-1111-111111111111",
		Parties:      []string{"A", "B"},
		Status:       "active",
		ContractType: "service_agreement",
		Version:      "v2",
		UpdatedAt:    now,
	}

	// Clean up any existing test data
	db.Exec("DELETE FROM contracts WHERE id = $1", c.ID)

	// Create the contract first
	if err := repo.CreateContract(context.Background(), c); err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}

	// Test update
	if err := repo.UpdateContract(context.Background(), c); err != nil {
		t.Fatalf("UpdateContract error: %v", err)
	}

	// Verify update
	var updatedStatus string
	err = db.QueryRow("SELECT status FROM contracts WHERE id = $1", c.ID).Scan(&updatedStatus)
	if err != nil {
		t.Fatalf("Failed to verify contract update: %v", err)
	}
	if updatedStatus != "active" {
		t.Fatalf("Expected status 'active', got '%s'", updatedStatus)
	}

	// Test delete
	if err := repo.DeleteContract(context.Background(), c.ID); err != nil {
		t.Fatalf("DeleteContract error: %v", err)
	}

	// Verify deletion
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contracts WHERE id = $1", c.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to verify contract deletion: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 contracts after deletion, got %d", count)
	}
}
