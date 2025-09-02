package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use environment variables for database connection
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	connStr := fmt.Sprintf("postgres://user:password@%s:5432/smart_payment?sslmode=disable", dbHost)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

func TestOracleRepository_CreateAndGetOracleProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewOracleRepository(db)

	// Test data
	provider := &models.OracleProvider{
		ID:          uuid.New(),
		Name:        "Test API Oracle",
		Description: "Test API Oracle for testing",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.example.com/verify",
		AuthConfig: models.OracleAuthConfig{
			Type: "bearer",
			ConfigData: map[string]string{
				"token": "test-token",
			},
		},
		RateLimitConfig: models.OracleRateLimitConfig{
			RequestsPerSecond: 10,
			BurstLimit:        20,
		},
		IsActive:     true,
		Reliability:  1.0,
		Capabilities: []string{"delivery_verification", "quality_check"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Execute
	ctx := context.Background()
	err := repo.CreateOracleProvider(ctx, provider)
	require.NoError(t, err)

	// Retrieve and verify
	retrievedProvider, err := repo.GetOracleProviderByID(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, provider.ID, retrievedProvider.ID)
	assert.Equal(t, provider.Name, retrievedProvider.Name)
	assert.Equal(t, provider.Type, retrievedProvider.Type)
	assert.Equal(t, provider.Endpoint, retrievedProvider.Endpoint)
	assert.Equal(t, provider.IsActive, retrievedProvider.IsActive)
	assert.Equal(t, provider.Reliability, retrievedProvider.Reliability)
	assert.Equal(t, provider.Capabilities, retrievedProvider.Capabilities)

	// Cleanup
	if _, err := db.Exec("DELETE FROM oracle_providers WHERE id = $1", provider.ID); err != nil {
		t.Logf("Warning: failed to cleanup oracle provider: %v", err)
	}
}

func TestOracleRepository_UpdateOracleProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewOracleRepository(db)

	// Create initial provider
	provider := &models.OracleProvider{
		ID:          uuid.New(),
		Name:        "Test API Oracle",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.example.com/verify",
		IsActive:    true,
		Reliability: 1.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	err := repo.CreateOracleProvider(ctx, provider)
	require.NoError(t, err)

	// Update provider
	provider.Name = "Updated Test API Oracle"
	provider.IsActive = false
	provider.UpdatedAt = time.Now()

	err = repo.UpdateOracleProvider(ctx, provider)
	require.NoError(t, err)

	// Retrieve and verify
	retrievedProvider, err := repo.GetOracleProviderByID(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, provider.Name, retrievedProvider.Name)
	assert.Equal(t, provider.IsActive, retrievedProvider.IsActive)

	// Cleanup
	if _, err := db.Exec("DELETE FROM oracle_providers WHERE id = $1", provider.ID); err != nil {
		t.Logf("Warning: failed to cleanup oracle provider: %v", err)
	}
}

func TestOracleRepository_DeleteOracleProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewOracleRepository(db)

	// Create initial provider
	provider := &models.OracleProvider{
		ID:          uuid.New(),
		Name:        "Test API Oracle",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.example.com/verify",
		IsActive:    true,
		Reliability: 1.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()
	err := repo.CreateOracleProvider(ctx, provider)
	require.NoError(t, err)

	// Delete provider
	err = repo.DeleteOracleProvider(ctx, provider.ID)
	require.NoError(t, err)

	// Verify deletion
	deletedProvider, err := repo.GetOracleProviderByID(ctx, provider.ID)
	require.Error(t, err)
	assert.Nil(t, deletedProvider)
	assert.Contains(t, err.Error(), "not found")
}
