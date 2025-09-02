package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

// Test constants
const (
	testContractID   = "testContractID"
	testMilestoneID  = "milestone-123"
	testTemplateID   = "template-123"
	testSharedUserID = "user-456"
	testCategory     = "delivery"
	testTemplateName = "Standard Delivery Template"
)

func TestPostgresMilestoneRepository_CreateMilestone(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestone := &models.ContractMilestone{
		ID:                   testMilestoneID,
		ContractID:           testContractID,
		MilestoneID:          "m1",
		SequenceNumber:       1,
		Dependencies:         []string{"dep1", "dep2"},
		Category:             "delivery",
		Priority:             3,
		CriticalPath:         true,
		TriggerConditions:    "Package delivered",
		VerificationCriteria: "Delivery confirmation received",
		PercentageComplete:   0,
		RiskLevel:            "medium",
		ContingencyPlans:     []string{"plan1", "plan2"},
		CriticalityScore:     75,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	mock.ExpectExec(`INSERT INTO contract_milestones`).
		WithArgs(
			milestone.ID,
			milestone.ContractID,
			milestone.MilestoneID,
			milestone.SequenceNumber,
			pq.Array(milestone.Dependencies),
			milestone.Category,
			milestone.Priority,
			milestone.CriticalPath,
			milestone.TriggerConditions,
			milestone.VerificationCriteria,
			milestone.EstimatedStartDate,
			milestone.EstimatedEndDate,
			milestone.ActualStartDate,
			milestone.ActualEndDate,
			int64(milestone.EstimatedDuration),
			nil, // actualDuration is nil
			milestone.PercentageComplete,
			milestone.RiskLevel,
			pq.Array(milestone.ContingencyPlans),
			milestone.CriticalityScore,
			milestone.CreatedAt,
			milestone.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateMilestone(context.Background(), milestone)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestoneByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestoneID := testMilestoneID
	expectedMilestone := &models.ContractMilestone{
		ID:                   milestoneID,
		ContractID:           testContractID,
		MilestoneID:          "m1",
		SequenceNumber:       1,
		Dependencies:         []string{"dep1", "dep2"},
		Category:             "delivery",
		Priority:             3,
		CriticalPath:         true,
		TriggerConditions:    "Package delivered",
		VerificationCriteria: "Delivery confirmation received",
		PercentageComplete:   50,
		RiskLevel:            "medium",
		ContingencyPlans:     []string{"plan1", "plan2"},
		CriticalityScore:     75,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		expectedMilestone.ID,
		expectedMilestone.ContractID,
		expectedMilestone.MilestoneID,
		expectedMilestone.SequenceNumber,
		pq.Array(expectedMilestone.Dependencies),
		expectedMilestone.Category,
		expectedMilestone.Priority,
		expectedMilestone.CriticalPath,
		expectedMilestone.TriggerConditions,
		expectedMilestone.VerificationCriteria,
		nil, // estimated_start_date
		nil, // estimated_end_date
		nil, // actual_start_date
		nil, // actual_end_date
		nil, // estimated_duration
		nil, // actual_duration
		expectedMilestone.PercentageComplete,
		expectedMilestone.RiskLevel,
		pq.Array(expectedMilestone.ContingencyPlans),
		expectedMilestone.CriticalityScore,
		expectedMilestone.CreatedAt,
		expectedMilestone.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE id = \$1`).
		WithArgs(milestoneID).
		WillReturnRows(rows)

	result, err := repo.GetMilestoneByID(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMilestone.ID, result.ID)
	assert.Equal(t, expectedMilestone.ContractID, result.ContractID)
	assert.Equal(t, expectedMilestone.Category, result.Category)
	assert.Equal(t, expectedMilestone.Priority, result.Priority)
	assert.Equal(t, expectedMilestone.CriticalPath, result.CriticalPath)
	assert.Equal(t, expectedMilestone.PercentageComplete, result.PercentageComplete)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestoneByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestoneID := "nonexistent-milestone"

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE id = \$1`).
		WithArgs(milestoneID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetMilestoneByID(context.Background(), milestoneID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "milestone not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_UpdateMilestone(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestone := &models.ContractMilestone{
		ID:                   testMilestoneID,
		SequenceNumber:       2,
		Dependencies:         []string{"dep1", "dep2", "dep3"},
		Category:             "approval",
		Priority:             5,
		CriticalPath:         false,
		TriggerConditions:    "Updated conditions",
		VerificationCriteria: "Updated criteria",
		PercentageComplete:   75,
		RiskLevel:            "high",
		ContingencyPlans:     []string{"plan1", "plan2", "plan3"},
		CriticalityScore:     90,
		UpdatedAt:            time.Now(),
	}

	mock.ExpectExec(`UPDATE contract_milestones SET`).
		WithArgs(
			milestone.ID,
			milestone.SequenceNumber,
			pq.Array(milestone.Dependencies),
			milestone.Category,
			milestone.Priority,
			milestone.CriticalPath,
			milestone.TriggerConditions,
			milestone.VerificationCriteria,
			milestone.EstimatedStartDate,
			milestone.EstimatedEndDate,
			milestone.ActualStartDate,
			milestone.ActualEndDate,
			int64(milestone.EstimatedDuration),
			nil, // actualDuration is nil
			milestone.PercentageComplete,
			milestone.RiskLevel,
			pq.Array(milestone.ContingencyPlans),
			milestone.CriticalityScore,
			milestone.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateMilestone(context.Background(), milestone)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_UpdateMilestone_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestone := &models.ContractMilestone{
		ID:        "nonexistent-milestone",
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec(`UPDATE contract_milestones SET`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	err = repo.UpdateMilestone(context.Background(), milestone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_DeleteMilestone(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestoneID := testMilestoneID

	mock.ExpectExec(`DELETE FROM contract_milestones WHERE id = \$1`).
		WithArgs(milestoneID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteMilestone(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestonesByContract(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID
	limit, offset := 10, 0

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"milestone-1", contractID, "m1", 1, pq.Array([]string{}), "delivery",
		3, true, "First milestone", "First criteria", nil, nil, nil, nil,
		nil, nil, 0, "medium", pq.Array([]string{}), 50, time.Now(), time.Now(),
	).AddRow(
		"milestone-2", contractID, "m2", 2, pq.Array([]string{"milestone-1"}), "approval",
		5, false, "Second milestone", "Second criteria", nil, nil, nil, nil,
		nil, nil, 25, "low", pq.Array([]string{}), 25, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE contract_id = \$1 ORDER BY sequence_number, created_at LIMIT \$2 OFFSET \$3`).
		WithArgs(contractID, limit, offset).
		WillReturnRows(rows)

	results, err := repo.GetMilestonesByContract(context.Background(), contractID, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "milestone-1", results[0].ID)
	assert.Equal(t, "milestone-2", results[1].ID)
	assert.Equal(t, contractID, results[0].ContractID)
	assert.Equal(t, contractID, results[1].ContractID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestonesByStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	status := "completed"
	limit, offset := 5, 0

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"milestone-completed", "testContractID", "m1", 1, pq.Array([]string{}), "delivery",
		3, true, "Completed milestone", "Completed criteria", nil, nil, nil, nil,
		nil, nil, 100, "medium", pq.Array([]string{}), 50, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE percentage_complete = 100 ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(rows)

	results, err := repo.GetMilestonesByStatus(context.Background(), status, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "milestone-completed", results[0].ID)
	assert.Equal(t, float64(100), results[0].PercentageComplete)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetOverdueMilestones(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	asOfDate := time.Now()
	limit, offset := 10, 0

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"overdue-milestone", "testContractID", "m1", 1, pq.Array([]string{}), "delivery",
		3, true, "Overdue milestone", "Overdue criteria", nil, time.Now().Add(-24*time.Hour),
		nil, nil, nil, nil, 50, "high", pq.Array([]string{}), 90, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE estimated_end_date < \$1 AND percentage_complete < 100 ORDER BY estimated_end_date LIMIT \$2 OFFSET \$3`).
		WithArgs(asOfDate, limit, offset).
		WillReturnRows(rows)

	results, err := repo.GetOverdueMilestones(context.Background(), asOfDate, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "overdue-milestone", results[0].ID)
	assert.Equal(t, "high", results[0].RiskLevel)
	assert.Equal(t, float64(50), results[0].PercentageComplete)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetCriticalPathMilestones(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"critical-milestone-1", contractID, "m1", 1, pq.Array([]string{}), "delivery",
		5, true, "Critical milestone 1", "Critical criteria 1", nil, nil, nil, nil,
		nil, nil, 0, "high", pq.Array([]string{}), 95, time.Now(), time.Now(),
	).AddRow(
		"critical-milestone-2", contractID, "m3", 3, pq.Array([]string{"critical-milestone-1"}), "approval",
		5, true, "Critical milestone 2", "Critical criteria 2", nil, nil, nil, nil,
		nil, nil, 0, "high", pq.Array([]string{}), 90, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE contract_id = \$1 AND critical_path = true ORDER BY sequence_number`).
		WithArgs(contractID).
		WillReturnRows(rows)

	results, err := repo.GetCriticalPathMilestones(context.Background(), contractID)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "critical-milestone-1", results[0].ID)
	assert.Equal(t, "critical-milestone-2", results[1].ID)
	assert.True(t, results[0].CriticalPath)
	assert.True(t, results[1].CriticalPath)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_BatchUpdateMilestoneStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestoneIDs := []string{"milestone-1", "milestone-2", "milestone-3"}
	status := "completed"

	mock.ExpectExec(`UPDATE contract_milestones SET percentage_complete = \$1, updated_at = \$2 WHERE id = ANY\(\$3\)`).
		WithArgs(float64(100), sqlmock.AnyArg(), pq.Array(milestoneIDs)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err = repo.BatchUpdateMilestoneStatus(context.Background(), milestoneIDs, status)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_BatchUpdateMilestoneProgress(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	updates := []MilestoneProgressUpdate{
		{MilestoneID: "milestone-1", PercentageComplete: 25, Notes: "Progress update 1"},
		{MilestoneID: "milestone-2", PercentageComplete: 50, Notes: "Progress update 2"},
		{MilestoneID: "milestone-3", PercentageComplete: 75, Notes: "Progress update 3"},
	}

	mock.ExpectBegin()
	for _, update := range updates {
		mock.ExpectExec(`UPDATE contract_milestones SET percentage_complete = \$1, updated_at = \$2 WHERE id = \$3`).
			WithArgs(update.PercentageComplete, sqlmock.AnyArg(), update.MilestoneID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	err = repo.BatchUpdateMilestoneProgress(context.Background(), updates)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestoneCompletionStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID

	rows := sqlmock.NewRows([]string{
		"total_milestones", "completed_milestones", "pending_milestones", "overdue_milestones", "average_completion",
	}).AddRow(10, 3, 5, 2, 35.5)

	mock.ExpectQuery(`SELECT COUNT\(\*\) as total_milestones`).
		WithArgs(contractID).
		WillReturnRows(rows)

	stats, err := repo.GetMilestoneCompletionStats(context.Background(), &contractID, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 10, stats.TotalMilestones)
	assert.Equal(t, 3, stats.CompletedMilestones)
	assert.Equal(t, 5, stats.PendingMilestones)
	assert.Equal(t, 2, stats.OverdueMilestones)
	assert.Equal(t, 35.5, stats.AverageCompletion)
	assert.Equal(t, 30.0, stats.CompletionRate) // 3/10 * 100
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_CreateMilestoneDependency(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	dependency := &models.MilestoneDependency{
		ID:             "dep-123",
		MilestoneID:    "milestone-2",
		DependsOnID:    "milestone-1",
		DependencyType: "prerequisite",
	}

	mock.ExpectExec(`INSERT INTO milestone_dependencies`).
		WithArgs(dependency.ID, dependency.MilestoneID, dependency.DependsOnID, dependency.DependencyType).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateMilestoneDependency(context.Background(), dependency)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetMilestoneDependencies(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	milestoneID := "milestone-2"

	rows := sqlmock.NewRows([]string{"id", "milestone_id", "depends_on_id", "dependency_type"}).
		AddRow("dep-1", milestoneID, "milestone-1", "prerequisite").
		AddRow("dep-2", milestoneID, "milestone-0", "parallel")

	mock.ExpectQuery(`SELECT id, milestone_id, depends_on_id, dependency_type FROM milestone_dependencies WHERE milestone_id = \$1`).
		WithArgs(milestoneID).
		WillReturnRows(rows)

	dependencies, err := repo.GetMilestoneDependencies(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.Len(t, dependencies, 2)
	assert.Equal(t, "dep-1", dependencies[0].ID)
	assert.Equal(t, "milestone-1", dependencies[0].DependsOnID)
	assert.Equal(t, "prerequisite", dependencies[0].DependencyType)
	assert.Equal(t, "dep-2", dependencies[1].ID)
	assert.Equal(t, "milestone-0", dependencies[1].DependsOnID)
	assert.Equal(t, "parallel", dependencies[1].DependencyType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_SearchMilestones(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	query := "delivery"
	limit, offset := 5, 0

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"milestone-search", "testContractID", "m1", 1, pq.Array([]string{}), "delivery",
		3, true, "Package delivery milestone", "Delivery confirmation", nil, nil, nil, nil,
		nil, nil, 0, "medium", pq.Array([]string{}), 50, time.Now(), time.Now(),
	)

	searchPattern := "%" + query + "%"
	mock.ExpectQuery(`SELECT .+ FROM contract_milestones WHERE trigger_conditions ILIKE \$1 OR verification_criteria ILIKE \$1 OR category ILIKE \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(searchPattern, limit, offset).
		WillReturnRows(rows)

	results, err := repo.SearchMilestones(context.Background(), query, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "milestone-search", results[0].ID)
	assert.Equal(t, "delivery", results[0].Category)
	assert.Contains(t, results[0].TriggerConditions, "delivery")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_ValidateDependencyGraph(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID

	// Mock a simple acyclic graph: A -> B -> C
	rows := sqlmock.NewRows([]string{"milestone_id", "depends_on_id"}).
		AddRow("B", "A").
		AddRow("C", "B")

	mock.ExpectQuery(`SELECT md.milestone_id, md.depends_on_id FROM milestone_dependencies md JOIN contract_milestones cm ON md.milestone_id = cm.id WHERE cm.contract_id = \$1`).
		WithArgs(contractID).
		WillReturnRows(rows)

	isValid, err := repo.ValidateDependencyGraph(context.Background(), contractID)
	assert.NoError(t, err)
	assert.True(t, isValid)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_GetTopologicalOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID

	// Mock a simple acyclic graph: A -> B -> C
	rows := sqlmock.NewRows([]string{"milestone_id", "depends_on_id"}).
		AddRow("B", "A").
		AddRow("C", "B")

	mock.ExpectQuery(`SELECT md.milestone_id, md.depends_on_id FROM milestone_dependencies md JOIN contract_milestones cm ON md.milestone_id = cm.id WHERE cm.contract_id = \$1`).
		WithArgs(contractID).
		WillReturnRows(rows)

	order, err := repo.GetTopologicalOrder(context.Background(), contractID)
	assert.NoError(t, err)
	assert.Len(t, order, 3)

	// A should come first (no dependencies)
	assert.Equal(t, "A", order[0])
	// B should come after A
	assert.Equal(t, "B", order[1])
	// C should come after B
	assert.Equal(t, "C", order[2])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneRepository_FilterMilestones(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneRepository(db)

	contractID := testContractID
	category := "delivery"
	priority := 3
	criticalPath := true

	filter := &MilestoneFilter{
		ContractID:   &contractID,
		Category:     &category,
		Priority:     &priority,
		CriticalPath: &criticalPath,
	}

	rows := sqlmock.NewRows([]string{
		"id", "contract_id", "milestone_id", "sequence_number", "dependencies", "category",
		"priority", "critical_path", "trigger_conditions", "verification_criteria",
		"estimated_start_date", "estimated_end_date", "actual_start_date", "actual_end_date",
		"estimated_duration", "actual_duration", "percentage_complete", "risk_level",
		"contingency_plans", "criticality_score", "created_at", "updated_at",
	}).AddRow(
		"milestone-filtered", contractID, "m1", 1, pq.Array([]string{}), category,
		priority, criticalPath, "Filtered milestone", "Filtered criteria", nil, nil, nil, nil,
		nil, nil, 25, "medium", pq.Array([]string{}), 50, time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM contract_milestones .+ ORDER BY created_at DESC`).
		WithArgs(contractID, category, priority, criticalPath).
		WillReturnRows(rows)

	results, err := repo.FilterMilestones(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "milestone-filtered", results[0].ID)
	assert.Equal(t, contractID, results[0].ContractID)
	assert.Equal(t, category, results[0].Category)
	assert.Equal(t, priority, results[0].Priority)
	assert.Equal(t, criticalPath, results[0].CriticalPath)
	assert.NoError(t, mock.ExpectationsWereMet())
}
