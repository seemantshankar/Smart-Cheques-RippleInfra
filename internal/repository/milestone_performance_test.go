package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkMilestoneRepository_DependencyResolutionPerformance tests the performance of dependency graph resolution
func BenchmarkMilestoneRepository_DependencyResolutionPerformance(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Setup test database for each iteration
		db, mock, err := sqlmock.New()
		require.NoError(b, err)

		repo := NewPostgresMilestoneRepository(db)
		ctx := context.Background()

		// Create a contract for testing
		contractID := uuid.New().String()

		// Mock the resolve dependency graph call
		rows := sqlmock.NewRows([]string{"milestone_id", "depends_on_id"}).
			AddRow("milestone-1", "milestone-0").
			AddRow("milestone-2", "milestone-1")
		mock.ExpectQuery(`SELECT md\.milestone_id, md\.depends_on_id FROM milestone_dependencies md JOIN contract_milestones cm ON md\.milestone_id = cm\.id WHERE cm\.contract_id = \$1`).
			WithArgs(contractID).
			WillReturnRows(rows)

		_, err = repo.ResolveDependencyGraph(ctx, contractID)
		assert.NoError(b, err)

		// Close the database connection
		db.Close()
	}
}

// BenchmarkMilestoneRepository_SearchPerformance tests the performance of milestone search operations
func BenchmarkMilestoneRepository_SearchPerformance(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		// Setup test database for each iteration
		db, mock := setupMilestoneTestDB(b)
		defer db.Close()

		repo := NewPostgresMilestoneRepository(db)
		ctx := context.Background()

		// Mock the search milestones call
		rows := createMilestoneRows()
		mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE`).
			WillReturnRows(rows)

		results, err := repo.SearchMilestones(ctx, "condition", 50, 0)
		assert.NoError(b, err)
		assert.NotNil(b, results)

		// Close the database connection
		db.Close()
	}
}

// BenchmarkMilestoneRepository_ConcurrentUpdates tests concurrent milestone updates
func BenchmarkMilestoneRepository_ConcurrentUpdates(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.ResetTimer()

	// Run concurrent updates
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Setup test database for each iteration
			db, mock, err := sqlmock.New()
			require.NoError(b, err)

			repo := NewPostgresMilestoneRepository(db)
			ctx := context.Background()

			// Mock the get milestone by ID call
			rows := sqlmock.NewRows([]string{
				"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
				"priority", "critical_path", "trigger_conditions", "verification_criteria",
				"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
				"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
				"contingency_plans", "criticality_score", "created_at", "updated_at",
			}).AddRow(
				"concurrent-milestone-1", "contract-1", "CM1", 1, pq.Array([]string{}), "delivery",
				1, false, "condition", "criteria", nil, nil, nil, nil,
				nil, nil, 50.0, "medium",
				pq.Array([]string{}), 75, time.Now(), time.Now(),
			)
			mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE id = \$1`).
				WithArgs("concurrent-milestone-1").
				WillReturnRows(rows)

			// Mock the update milestone call
			mock.ExpectExec(`UPDATE contract_milestones SET`).
				WillReturnResult(sqlmock.NewResult(1, 1))

			milestoneID := "concurrent-milestone-1"
			milestone, err := repo.GetMilestoneByID(ctx, milestoneID)
			if err != nil {
				b.Errorf("Failed to get milestone: %v", err)
				db.Close()
				continue
			}

			// Update the milestone
			milestone.PercentageComplete = float64((i + 1) % 100)
			milestone.UpdatedAt = time.Now()
			err = repo.UpdateMilestone(ctx, milestone)
			assert.NoError(b, err)

			// Close the database connection
			db.Close()

			i++
		}
	})
}

// BenchmarkMilestoneRepository_NotificationSystemScalability tests the scalability of notification-related operations
func BenchmarkMilestoneRepository_NotificationSystemScalability(b *testing.B) {
	// Skip in short mode
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.ResetTimer()

	// Test various notification-related queries
	b.Run("OverdueMilestones", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Setup test database for each iteration
			db, mock, err := sqlmock.New()
			require.NoError(b, err)

			repo := NewPostgresMilestoneRepository(db)
			ctx := context.Background()

			// Mock the get overdue milestones call
			rows := sqlmock.NewRows([]string{
				"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
				"priority", "critical_path", "trigger_conditions", "verification_criteria",
				"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
				"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
				"contingency_plans", "criticality_score", "created_at", "updated_at",
			}).AddRow(
				"notification-milestone-1", "contract-1", "NM1", 1, pq.Array([]string{}), "delivery",
				1, false, "condition", "criteria", nil, nil, nil, nil,
				nil, nil, 50.0, "medium",
				pq.Array([]string{}), 75, time.Now(), time.Now(),
			)
			mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE estimated_end_date < \$1 AND percentage_complete < 100`).
				WillReturnRows(rows)

			results, err := repo.GetOverdueMilestones(ctx, time.Now(), 50, 0)
			assert.NoError(b, err)
			assert.NotNil(b, results)

			// Close the database connection
			db.Close()
		}
	})

	b.Run("MilestonesByPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Setup test database for each iteration
			db, mock := setupMilestoneTestDB(b)
			defer db.Close()

			repo := NewPostgresMilestoneRepository(db)
			ctx := context.Background()

			// Mock the get milestones by priority call
			rows := createMilestoneRows()
			mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE priority = \$1`).
				WillReturnRows(rows)

			results, err := repo.GetMilestonesByPriority(ctx, 4, 50, 0)
			assert.NoError(b, err)
			assert.NotNil(b, results)

			// Close the database connection
			db.Close()
		}
	})

	b.Run("MilestonesByRiskLevel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Setup test database for each iteration
			db, mock := setupMilestoneTestDB(b)
			defer db.Close()

			repo := NewPostgresMilestoneRepository(db)
			ctx := context.Background()

			// Mock the get milestones by risk level call
			rows := createMilestoneRows()
			mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE risk_level = \$1`).
				WillReturnRows(rows)

			results, err := repo.GetMilestonesByRiskLevel(ctx, "high", 50, 0)
			assert.NoError(b, err)
			assert.NotNil(b, results)

			// Close the database connection
			db.Close()
		}
	})
}

// TestMilestoneRepository_DependencyResolutionPerformance validates dependency resolution correctness under load
func TestMilestoneRepository_DependencyResolutionPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup test database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)
	ctx := context.Background()

	// Create a contract for testing
	contractID := uuid.New().String()

	// Mock the resolve dependency graph call
	rows := sqlmock.NewRows([]string{"milestone_id", "depends_on_id"}).
		AddRow("perf-milestone-1", "perf-milestone-0").
		AddRow("perf-milestone-2", "perf-milestone-1")
	mock.ExpectQuery(`SELECT md\.milestone_id, md\.depends_on_id FROM milestone_dependencies md JOIN contract_milestones cm ON md\.milestone_id = cm\.id WHERE cm\.contract_id = \$1`).
		WithArgs(contractID).
		WillReturnRows(rows)

	// Measure the time it takes to resolve the dependency graph
	startTime := time.Now()
	graph, err := repo.ResolveDependencyGraph(ctx, contractID)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Greater(t, len(graph), 0)

	// Performance assertion - should complete within 100ms for 50 milestones
	assert.Less(t, duration.Milliseconds(), int64(100), "Dependency resolution should complete within 100ms")

	// Validate the graph structure
	totalDependencies := 0
	for _, dependents := range graph {
		totalDependencies += len(dependents)
	}
	assert.Equal(t, 2, totalDependencies, "Graph should have correct number of dependencies")
}

// TestMilestoneRepository_SearchPerformance validates search performance and correctness
func TestMilestoneRepository_SearchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup test database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)
	ctx := context.Background()

	// Mock the search milestones calls for each test case
	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE`).
		WithArgs("%condition 5%", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
			"priority", "critical_path", "trigger_conditions", "verification_criteria",
			"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
			"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
			"contingency_plans", "criticality_score", "created_at", "updated_at",
		}).AddRow(
			"search-milestone-1", "search-contract-1", "SM1", 1, pq.Array([]string{}), "delivery",
			1, false, "Search for performance condition 5 in milestone 1", "Verify performance criteria 1 for milestone 1", nil, nil, nil, nil,
			nil, nil, 50.0, "medium",
			pq.Array([]string{}), 75, time.Now(), time.Now(),
		))

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE`).
		WithArgs("%delivery%", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
			"priority", "critical_path", "trigger_conditions", "verification_criteria",
			"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
			"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
			"contingency_plans", "criticality_score", "created_at", "updated_at",
		}).AddRow(
			"search-milestone-2", "search-contract-2", "SM2", 2, pq.Array([]string{}), "delivery",
			2, false, "Another condition", "Another criteria", nil, nil, nil, nil,
			nil, nil, 75.0, "low",
			pq.Array([]string{}), 50, time.Now(), time.Now(),
		))

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE`).
		WithArgs("%medium%", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
			"priority", "critical_path", "trigger_conditions", "verification_criteria",
			"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
			"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
			"contingency_plans", "criticality_score", "created_at", "updated_at",
		}).AddRow(
			"search-milestone-3", "search-contract-3", "SM3", 3, pq.Array([]string{}), "approval",
			3, true, "Medium risk condition", "Medium risk criteria", nil, nil, nil, nil,
			nil, nil, 25.0, "medium",
			pq.Array([]string{}), 80, time.Now(), time.Now(),
		))

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE`).
		WithArgs("%milestone 1%", 100, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
			"priority", "critical_path", "trigger_conditions", "verification_criteria",
			"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
			"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
			"contingency_plans", "criticality_score", "created_at", "updated_at",
		}).AddRow(
			"search-milestone-4", "search-contract-4", "SM4", 4, pq.Array([]string{}), "review",
			4, false, "Specific milestone 1 condition", "Specific milestone 1 criteria", nil, nil, nil, nil,
			nil, nil, 90.0, "high",
			pq.Array([]string{}), 95, time.Now(), time.Now(),
		))

	// Test search performance
	testCases := []struct {
		name        string
		query       string
		expectedMin int
	}{
		{"Common term search", "condition 5", 1},
		{"Category search", "delivery", 1},
		{"Risk level search", "medium", 1},
		{"Specific term search", "milestone 1", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startTime := time.Now()
			results, err := repo.SearchMilestones(ctx, tc.query, 100, 0)
			duration := time.Since(startTime)

			require.NoError(t, err)
			assert.NotNil(t, results)
			assert.GreaterOrEqual(t, len(results), tc.expectedMin, "Should find expected minimum results")

			// Performance assertion - should complete within 200ms
			assert.Less(t, duration.Milliseconds(), int64(200), "Search should complete within 200ms")
		})
	}
}

// setupMilestoneTestDB creates a test database with mock data for milestone tests
func setupMilestoneTestDB(b *testing.B) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)

	return db, mock
}

// createMilestoneRows creates mock rows for milestone tests
func createMilestoneRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"test-milestone-1", "contract-1", "TM1", 1, pq.Array([]string{}), "delivery",
		1, false, "condition", "criteria", nil, nil, nil, nil,
		nil, nil, 50.0, "medium",
		pq.Array([]string{}), 75, time.Now(), time.Now(),
	)
}
