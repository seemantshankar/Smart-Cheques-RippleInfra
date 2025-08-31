//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

// TestMilestoneRepositoryIntegration demonstrates the milestone repository
// working with actual database operations (requires test database setup)
func TestMilestoneRepositoryIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would require actual database setup in CI/CD
	t.Skip("Integration test requires database setup - for demonstration only")

	// Example of how integration test would work:
	ctx := context.Background()

	// Setup test database connection
	db := setupTestDatabase(t)
	defer db.Close()

	// Initialize repositories
	milestoneRepo := NewPostgresMilestoneRepository(db)
	templateRepo := NewPostgresMilestoneTemplateRepository(db)

	// Test data
	contractID := uuid.New().String()

	// Test 1: Create milestone template
	template := &models.MilestoneTemplate{
		ID:              uuid.New().String(),
		Name:            "Integration Test Template",
		Description:     "Template for testing milestone repository integration",
		DefaultCategory: "delivery",
		DefaultPriority: 3,
		Variables:       []string{"delivery_address", "expected_date"},
		Version:         "1.0",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Test 2: Retrieve template
	retrievedTemplate, err := templateRepo.GetTemplateByID(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.Name, retrievedTemplate.Name)
	assert.Equal(t, template.DefaultCategory, retrievedTemplate.DefaultCategory)

	// Test 3: Create milestone from template
	variables := map[string]interface{}{
		"description":           "Package delivery milestone",
		"verification_criteria": "Delivery confirmation received",
		"estimated_duration":    int64(24 * time.Hour),
		"critical_path":         true,
	}

	milestone, err := templateRepo.InstantiateTemplate(ctx, template.ID, variables)
	require.NoError(t, err)
	assert.NotNil(t, milestone)

	// Set required fields for milestone creation
	milestone.ID = uuid.New().String()
	milestone.ContractID = contractID
	milestone.MilestoneID = "m1"
	milestone.SequenceNumber = 1

	// Test 4: Create milestone
	err = milestoneRepo.CreateMilestone(ctx, milestone)
	require.NoError(t, err)

	// Test 5: Retrieve milestone
	retrievedMilestone, err := milestoneRepo.GetMilestoneByID(ctx, milestone.ID)
	require.NoError(t, err)
	assert.Equal(t, milestone.ContractID, retrievedMilestone.ContractID)
	assert.Equal(t, milestone.Category, retrievedMilestone.Category)
	assert.True(t, retrievedMilestone.CriticalPath)

	// Test 6: Update milestone progress
	retrievedMilestone.PercentageComplete = 50.0
	retrievedMilestone.UpdatedAt = time.Now()
	err = milestoneRepo.UpdateMilestone(ctx, retrievedMilestone)
	require.NoError(t, err)

	// Test 7: Create milestone dependency
	milestone2 := &models.ContractMilestone{
		ID:                   uuid.New().String(),
		ContractID:           contractID,
		MilestoneID:          "m2",
		SequenceNumber:       2,
		Category:             "approval",
		Priority:             4,
		TriggerConditions:    "Approval milestone",
		VerificationCriteria: "Manager approval received",
		PercentageComplete:   0,
		RiskLevel:            "low",
		CriticalityScore:     30,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = milestoneRepo.CreateMilestone(ctx, milestone2)
	require.NoError(t, err)

	dependency := &models.MilestoneDependency{
		ID:             uuid.New().String(),
		MilestoneID:    milestone2.ID,
		DependsOnID:    milestone.ID,
		DependencyType: "prerequisite",
	}

	err = milestoneRepo.CreateMilestoneDependency(ctx, dependency)
	require.NoError(t, err)

	// Test 8: Get milestones by contract
	contractMilestones, err := milestoneRepo.GetMilestonesByContract(ctx, contractID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, contractMilestones, 2)

	// Test 9: Get dependency graph
	dependencies, err := milestoneRepo.GetMilestoneDependencies(ctx, milestone2.ID)
	require.NoError(t, err)
	assert.Len(t, dependencies, 1)
	assert.Equal(t, milestone.ID, dependencies[0].DependsOnID)

	// Test 10: Validate dependency graph (no cycles)
	isValid, err := milestoneRepo.ValidateDependencyGraph(ctx, contractID)
	require.NoError(t, err)
	assert.True(t, isValid)

	// Test 11: Get topological order
	order, err := milestoneRepo.GetTopologicalOrder(ctx, contractID)
	require.NoError(t, err)
	assert.Len(t, order, 2)
	assert.Equal(t, milestone.ID, order[0])  // Should come first (no dependencies)
	assert.Equal(t, milestone2.ID, order[1]) // Should come second (depends on first)

	// Test 12: Get completion stats
	stats, err := milestoneRepo.GetMilestoneCompletionStats(ctx, &contractID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalMilestones)
	assert.Equal(t, 0, stats.CompletedMilestones)
	assert.Equal(t, 25.0, stats.AverageCompletion) // (50 + 0) / 2

	// Test 13: Batch update milestone status
	err = milestoneRepo.BatchUpdateMilestoneStatus(ctx, []string{milestone.ID}, "completed")
	require.NoError(t, err)

	// Verify the update
	updatedMilestone, err := milestoneRepo.GetMilestoneByID(ctx, milestone.ID)
	require.NoError(t, err)
	assert.Equal(t, 100.0, updatedMilestone.PercentageComplete)

	// Test 14: Search milestones
	searchResults, err := milestoneRepo.SearchMilestones(ctx, "delivery", 10, 0)
	require.NoError(t, err)
	assert.Len(t, searchResults, 1)
	assert.Equal(t, milestone.ID, searchResults[0].ID)

	// Test 15: Get critical path milestones
	criticalMilestones, err := milestoneRepo.GetCriticalPathMilestones(ctx, contractID)
	require.NoError(t, err)
	assert.Len(t, criticalMilestones, 1)
	assert.Equal(t, milestone.ID, criticalMilestones[0].ID)

	// Test 16: Template versioning
	templateV2 := &models.MilestoneTemplate{
		ID:              uuid.New().String(),
		Name:            "Integration Test Template v2.0",
		Description:     "Updated template with new features",
		DefaultCategory: "delivery",
		DefaultPriority: 4,
		Variables:       []string{"delivery_address", "expected_date", "tracking_number"},
		Version:         "2.0",
		CreatedAt:       time.Now(),
	}

	err = templateRepo.CreateTemplateVersion(ctx, template.ID, templateV2)
	require.NoError(t, err)

	// Get template versions
	versions, err := templateRepo.GetTemplateVersions(ctx, template.ID)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, "2.0", versions[0].Version)

	// Test 17: Template sharing
	sharedWithUserID := "test-user-123"
	permissions := []string{"read", "instantiate"}
	err = templateRepo.ShareTemplate(ctx, template.ID, sharedWithUserID, permissions)
	require.NoError(t, err)

	// Get shared templates
	sharedTemplates, err := templateRepo.GetSharedTemplates(ctx, sharedWithUserID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, sharedTemplates, 1)
	assert.Equal(t, template.ID, sharedTemplates[0].ID)

	// Test 18: Template permissions
	userPermissions, err := templateRepo.GetTemplatePermissions(ctx, template.ID, sharedWithUserID)
	require.NoError(t, err)
	assert.Equal(t, permissions, userPermissions)

	// Cleanup: Delete created data
	err = milestoneRepo.DeleteMilestone(ctx, milestone.ID)
	require.NoError(t, err)

	err = milestoneRepo.DeleteMilestone(ctx, milestone2.ID)
	require.NoError(t, err)

	err = templateRepo.DeleteTemplate(ctx, template.ID)
	require.NoError(t, err)
}

// setupTestDatabase would set up a test database connection
// This is a placeholder function for actual database setup
func setupTestDatabase(t *testing.T) *sql.DB {
	// In a real implementation, this would:
	// 1. Create a test database connection
	// 2. Run migrations to set up schema
	// 3. Return the database connection
	//
	// Example:
	// db, err := sql.Open("postgres", "postgresql://test:test@localhost/test_db?sslmode=disable")
	// require.NoError(t, err)
	//
	// // Run migrations
	// migrate := migrate.NewWithDatabaseInstance("file://../../migrations", "postgres", driver)
	// migrate.Up()
	//
	// return db

	panic("setupTestDatabase not implemented - for demonstration only")
}

// Example of how to run performance tests
func BenchmarkMilestoneRepository_CreateMilestone(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Setup
	ctx := context.Background()
	db := setupTestDatabase(nil) // Would need proper setup
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)
	contractID := uuid.New().String()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		milestone := &models.ContractMilestone{
			ID:                   uuid.New().String(),
			ContractID:           contractID,
			MilestoneID:          fmt.Sprintf("m%d", i),
			SequenceNumber:       i + 1,
			Category:             "delivery",
			Priority:             3,
			TriggerConditions:    fmt.Sprintf("Milestone %d", i),
			VerificationCriteria: fmt.Sprintf("Criteria %d", i),
			PercentageComplete:   0,
			RiskLevel:            "medium",
			CriticalityScore:     50,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := repo.CreateMilestone(ctx, milestone); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMilestoneRepository_BatchCreate(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Setup
	ctx := context.Background()
	db := setupTestDatabase(nil) // Would need proper setup
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)
	contractID := uuid.New().String()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var milestones []*models.ContractMilestone

		// Create batch of 100 milestones
		for j := 0; j < 100; j++ {
			milestone := &models.ContractMilestone{
				ID:                   uuid.New().String(),
				ContractID:           contractID,
				MilestoneID:          fmt.Sprintf("m%d_%d", i, j),
				SequenceNumber:       j + 1,
				Category:             "delivery",
				Priority:             3,
				TriggerConditions:    fmt.Sprintf("Milestone %d_%d", i, j),
				VerificationCriteria: fmt.Sprintf("Criteria %d_%d", i, j),
				PercentageComplete:   0,
				RiskLevel:            "medium",
				CriticalityScore:     50,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}
			milestones = append(milestones, milestone)
		}

		if err := repo.BatchCreateMilestones(ctx, milestones); err != nil {
			b.Fatal(err)
		}
	}
}
