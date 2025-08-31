package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

// Milestone status constants
const (
	MilestoneStatusPending    = "pending"
	MilestoneStatusInProgress = "in_progress"
	MilestoneStatusCompleted  = "completed"
	MilestoneStatusFailed     = "failed"
	MilestoneStatusOnHold     = "on_hold"
)

// PostgresMilestoneRepository implements MilestoneRepositoryInterface using PostgreSQL
type PostgresMilestoneRepository struct {
	db *sql.DB
}

func NewPostgresMilestoneRepository(db *sql.DB) MilestoneRepositoryInterface {
	return &PostgresMilestoneRepository{db: db}
}

// CRUD operations for milestones

// CreateMilestone inserts a new milestone
func (r *PostgresMilestoneRepository) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	query := `
		INSERT INTO contract_milestones (
			id, contract_id, milestone_id, sequence_number, dependencies, category,
			priority, critical_path, trigger_conditions, verification_criteria,
			estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
			estimated_duration, actual_duration, percentage_complete, risk_level,
			contingency_plans, criticality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)`

	var actualDuration interface{}
	if milestone.ActualDuration != nil {
		actualDuration = int64(*milestone.ActualDuration)
	}

	_, err := r.db.ExecContext(ctx, query,
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
		actualDuration,
		milestone.PercentageComplete,
		milestone.RiskLevel,
		pq.Array(milestone.ContingencyPlans),
		milestone.CriticalityScore,
		milestone.CreatedAt,
		milestone.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create milestone: %w", err)
	}
	return nil
}

// GetMilestoneByID retrieves a milestone by ID
func (r *PostgresMilestoneRepository) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE id = $1`

	milestone, err := r.scanSingleMilestone(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone by ID: %w", err)
	}
	return milestone, nil
}

// UpdateMilestone updates an existing milestone
func (r *PostgresMilestoneRepository) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	query := `
		UPDATE contract_milestones SET
			sequence_number = $2, dependencies = $3, category = $4, priority = $5,
			critical_path = $6, trigger_conditions = $7, verification_criteria = $8,
			estimated_start_date = $9, estimated_end_date = $10, actual_start_date = $11,
			actual_end_date = $12, estimated_duration = $13, actual_duration = $14,
			percentage_complete = $15, risk_level = $16, contingency_plans = $17,
			criticality_score = $18, updated_at = $19
		WHERE id = $1`

	var actualDuration interface{}
	if milestone.ActualDuration != nil {
		actualDuration = int64(*milestone.ActualDuration)
	}

	res, err := r.db.ExecContext(ctx, query,
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
		actualDuration,
		milestone.PercentageComplete,
		milestone.RiskLevel,
		pq.Array(milestone.ContingencyPlans),
		milestone.CriticalityScore,
		milestone.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update milestone: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("milestone with ID %s not found", milestone.ID)
	}

	return nil
}

// DeleteMilestone deletes a milestone by ID
func (r *PostgresMilestoneRepository) DeleteMilestone(ctx context.Context, id string) error {
	query := `DELETE FROM contract_milestones WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete milestone: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("milestone with ID %s not found", id)
	}

	return nil
}

// Query methods

// GetMilestonesByContract retrieves milestones by contract ID
func (r *PostgresMilestoneRepository) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE contract_id = $1
		ORDER BY sequence_number, created_at
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, contractID, limit, offset)
}

// GetMilestonesByStatus retrieves milestones by status
func (r *PostgresMilestoneRepository) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	// Status is derived from percentage_complete
	var condition string
	switch status {
	case MilestoneStatusPending:
		condition = "percentage_complete = 0"
	case MilestoneStatusInProgress:
		condition = "percentage_complete > 0 AND percentage_complete < 100"
	case MilestoneStatusCompleted:
		condition = "percentage_complete = 100"
	default:
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	query := fmt.Sprintf(`
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, condition)

	return r.scanMilestones(ctx, query, limit, offset)
}

// GetOverdueMilestones retrieves overdue milestones
func (r *PostgresMilestoneRepository) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE estimated_end_date < $1 AND percentage_complete < 100
		ORDER BY estimated_end_date
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, asOfDate, limit, offset)
}

// GetMilestonesByPriority retrieves milestones by priority
func (r *PostgresMilestoneRepository) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE priority = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, priority, limit, offset)
}

// GetMilestonesByCategory retrieves milestones by category
func (r *PostgresMilestoneRepository) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE category = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, category, limit, offset)
}

// GetMilestonesByRiskLevel retrieves milestones by risk level
func (r *PostgresMilestoneRepository) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE risk_level = $1
		ORDER BY criticality_score DESC
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, riskLevel, limit, offset)
}

// GetCriticalPathMilestones retrieves critical path milestones for a contract
func (r *PostgresMilestoneRepository) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE contract_id = $1 AND critical_path = true
		ORDER BY sequence_number`

	return r.scanMilestones(ctx, query, contractID)
}

// SearchMilestones searches milestones by text query
func (r *PostgresMilestoneRepository) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	searchQuery := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE trigger_conditions ILIKE $1
		   OR verification_criteria ILIKE $1
		   OR category ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	searchPattern := "%" + query + "%"
	return r.scanMilestones(ctx, searchQuery, searchPattern, limit, offset)
}

// Dependency resolution methods

// GetMilestoneDependencies retrieves dependencies for a milestone
func (r *PostgresMilestoneRepository) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	query := `
		SELECT id, milestone_id, depends_on_id, dependency_type
		FROM milestone_dependencies
		WHERE milestone_id = $1`

	rows, err := r.db.QueryContext(ctx, query, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []*models.MilestoneDependency
	for rows.Next() {
		dep := &models.MilestoneDependency{}
		if err := rows.Scan(&dep.ID, &dep.MilestoneID, &dep.DependsOnID, &dep.DependencyType); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, rows.Err()
}

// GetMilestoneDependents retrieves dependents for a milestone
func (r *PostgresMilestoneRepository) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	query := `
		SELECT id, milestone_id, depends_on_id, dependency_type
		FROM milestone_dependencies
		WHERE depends_on_id = $1`

	rows, err := r.db.QueryContext(ctx, query, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone dependents: %w", err)
	}
	defer rows.Close()

	var dependents []*models.MilestoneDependency
	for rows.Next() {
		dep := &models.MilestoneDependency{}
		if err := rows.Scan(&dep.ID, &dep.MilestoneID, &dep.DependsOnID, &dep.DependencyType); err != nil {
			return nil, fmt.Errorf("failed to scan dependent: %w", err)
		}
		dependents = append(dependents, dep)
	}

	return dependents, rows.Err()
}

// ResolveDependencyGraph resolves dependency graph for a contract
func (r *PostgresMilestoneRepository) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	query := `
		SELECT md.milestone_id, md.depends_on_id
		FROM milestone_dependencies md
		JOIN contract_milestones cm ON md.milestone_id = cm.id
		WHERE cm.contract_id = $1`

	rows, err := r.db.QueryContext(ctx, query, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependency graph: %w", err)
	}
	defer rows.Close()

	graph := make(map[string][]string)
	for rows.Next() {
		var milestoneID, dependsOnID string
		if err := rows.Scan(&milestoneID, &dependsOnID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		// Build reverse graph: dependsOnID -> milestoneID (what depends on this node)
		graph[dependsOnID] = append(graph[dependsOnID], milestoneID)
	}

	return graph, rows.Err()
}

// ValidateDependencyGraph validates dependency graph for cycles
func (r *PostgresMilestoneRepository) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	graph, err := r.ResolveDependencyGraph(ctx, contractID)
	if err != nil {
		return false, err
	}

	// Simple cycle detection using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if !visited[neighbor] && hasCycle(neighbor) {
				return true
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] && hasCycle(node) {
			return false, nil
		}
	}

	return true, nil
}

// GetTopologicalOrder returns topological order of milestones
func (r *PostgresMilestoneRepository) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	graph, err := r.ResolveDependencyGraph(ctx, contractID)
	if err != nil {
		return nil, err
	}

	// Get all nodes
	nodes := make(map[string]bool)
	for node := range graph {
		nodes[node] = true
		for _, dep := range graph[node] {
			nodes[dep] = true
		}
	}

	// Calculate in-degrees
	inDegree := make(map[string]int)
	for node := range nodes {
		inDegree[node] = 0
	}
	for _, deps := range graph {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Kahn's algorithm
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var result []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(nodes) {
		return nil, fmt.Errorf("dependency cycle detected")
	}

	return result, nil
}

// CreateMilestoneDependency creates a new milestone dependency
func (r *PostgresMilestoneRepository) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	query := `
		INSERT INTO milestone_dependencies (id, milestone_id, depends_on_id, dependency_type)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query,
		dependency.ID,
		dependency.MilestoneID,
		dependency.DependsOnID,
		dependency.DependencyType,
	)
	if err != nil {
		return fmt.Errorf("failed to create milestone dependency: %w", err)
	}
	return nil
}

// DeleteMilestoneDependency deletes a milestone dependency
func (r *PostgresMilestoneRepository) DeleteMilestoneDependency(ctx context.Context, id string) error {
	query := `DELETE FROM milestone_dependencies WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete milestone dependency: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("milestone dependency with ID %s not found", id)
	}

	return nil
}

// Batch operations for milestone updates

// BatchUpdateMilestoneStatus updates status for multiple milestones
func (r *PostgresMilestoneRepository) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	var percentage float64
	switch status {
	case MilestoneStatusPending:
		percentage = 0
	case MilestoneStatusInProgress:
		percentage = 50 // Default to 50% for in-progress
	case MilestoneStatusCompleted:
		percentage = 100
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE contract_milestones
		SET percentage_complete = $1, updated_at = $2
		WHERE id = ANY($3)`

	_, err := r.db.ExecContext(ctx, query, percentage, time.Now(), pq.Array(milestoneIDs))
	if err != nil {
		return fmt.Errorf("failed to batch update milestone status: %w", err)
	}
	return nil
}

// BatchUpdateMilestoneProgress updates progress for multiple milestones
func (r *PostgresMilestoneRepository) BatchUpdateMilestoneProgress(ctx context.Context, updates []MilestoneProgressUpdate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil {
			fmt.Printf("Warning: failed to rollback transaction: %v\n", rbErr)
		}
	}()

	query := `
		UPDATE contract_milestones
		SET percentage_complete = $1, updated_at = $2
		WHERE id = $3`

	for _, update := range updates {
		_, err := tx.ExecContext(ctx, query,
			update.PercentageComplete,
			time.Now(),
			update.MilestoneID,
		)
		if err != nil {
			return fmt.Errorf("failed to update milestone %s: %w", update.MilestoneID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// BatchCreateMilestones creates multiple milestones
func (r *PostgresMilestoneRepository) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil {
			fmt.Printf("Warning: failed to rollback transaction: %v\n", rbErr)
		}
	}()

	query := `
		INSERT INTO contract_milestones (
			id, contract_id, milestone_id, sequence_number, dependencies, category,
			priority, critical_path, trigger_conditions, verification_criteria,
			estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
			estimated_duration, actual_duration, percentage_complete, risk_level,
			contingency_plans, criticality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)`

	for _, milestone := range milestones {
		var actualDuration interface{}
		if milestone.ActualDuration != nil {
			actualDuration = int64(*milestone.ActualDuration)
		}

		_, err := tx.ExecContext(ctx, query,
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
			actualDuration,
			milestone.PercentageComplete,
			milestone.RiskLevel,
			pq.Array(milestone.ContingencyPlans),
			milestone.CriticalityScore,
			milestone.CreatedAt,
			milestone.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create milestone %s: %w", milestone.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// BatchDeleteMilestones deletes multiple milestones
func (r *PostgresMilestoneRepository) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	query := `DELETE FROM contract_milestones WHERE id = ANY($1)`

	_, err := r.db.ExecContext(ctx, query, pq.Array(milestoneIDs))
	if err != nil {
		return fmt.Errorf("failed to batch delete milestones: %w", err)
	}
	return nil
}

// Milestone analytics and reporting queries

// GetMilestoneCompletionStats retrieves milestone completion statistics
func (r *PostgresMilestoneRepository) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*MilestoneStats, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	whereClause.WriteString("WHERE 1=1")

	if contractID != nil {
		whereClause.WriteString(fmt.Sprintf(" AND contract_id = $%d", argIndex))
		args = append(args, *contractID)
		argIndex++
	}

	if startDate != nil {
		whereClause.WriteString(fmt.Sprintf(" AND created_at >= $%d", argIndex))
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		whereClause.WriteString(fmt.Sprintf(" AND created_at <= $%d", argIndex))
		args = append(args, *endDate)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_milestones,
			COUNT(CASE WHEN percentage_complete = 100 THEN 1 END) as completed_milestones,
			COUNT(CASE WHEN percentage_complete > 0 AND percentage_complete < 100 THEN 1 END) as pending_milestones,
			COUNT(CASE WHEN estimated_end_date < NOW() AND percentage_complete < 100 THEN 1 END) as overdue_milestones,
			AVG(percentage_complete) as average_completion
		FROM contract_milestones
		%s`, whereClause.String())

	var stats MilestoneStats
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalMilestones,
		&stats.CompletedMilestones,
		&stats.PendingMilestones,
		&stats.OverdueMilestones,
		&stats.AverageCompletion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone completion stats: %w", err)
	}

	if stats.TotalMilestones > 0 {
		stats.CompletionRate = float64(stats.CompletedMilestones) / float64(stats.TotalMilestones) * 100
	}

	return &stats, nil
}

// GetMilestonePerformanceMetrics retrieves milestone performance metrics
func (r *PostgresMilestoneRepository) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*MilestonePerformanceMetrics, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	whereClause.WriteString("WHERE percentage_complete = 100 AND actual_end_date IS NOT NULL AND estimated_end_date IS NOT NULL")

	if contractID != nil {
		whereClause.WriteString(fmt.Sprintf(" AND contract_id = $%d", argIndex))
		args = append(args, *contractID)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			AVG(EXTRACT(EPOCH FROM (actual_end_date - actual_start_date))) as avg_completion_time,
			COUNT(CASE WHEN actual_end_date <= estimated_end_date THEN 1 END) * 100.0 / COUNT(*) as on_time_rate,
			COUNT(CASE WHEN actual_end_date < estimated_end_date THEN 1 END) * 100.0 / COUNT(*) as early_rate,
			COUNT(CASE WHEN actual_end_date > estimated_end_date THEN 1 END) * 100.0 / COUNT(*) as delayed_rate,
			AVG(CASE WHEN actual_end_date > estimated_end_date THEN EXTRACT(EPOCH FROM (actual_end_date - estimated_end_date)) ELSE 0 END) as avg_delay
		FROM contract_milestones
		%s`, whereClause.String())

	var metrics MilestonePerformanceMetrics
	var avgCompletionSeconds, avgDelaySeconds sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&avgCompletionSeconds,
		&metrics.OnTimeCompletionRate,
		&metrics.EarlyCompletionRate,
		&metrics.DelayedCompletionRate,
		&avgDelaySeconds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone performance metrics: %w", err)
	}

	if avgCompletionSeconds.Valid {
		metrics.AverageCompletionTime = time.Duration(avgCompletionSeconds.Float64) * time.Second
	}
	if avgDelaySeconds.Valid {
		metrics.AverageDelay = time.Duration(avgDelaySeconds.Float64) * time.Second
	}

	// Calculate efficiency score (100 - average delay percentage)
	metrics.EfficiencyScore = 100 - metrics.DelayedCompletionRate

	return &metrics, nil
}

// GetMilestoneTimelineAnalysis performs timeline analysis for a contract
func (r *PostgresMilestoneRepository) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*MilestoneTimelineAnalysis, error) {
	// This is a simplified implementation - in a real system, you'd implement
	// a proper Critical Path Method (CPM) algorithm
	query := `
		SELECT id, estimated_start_date, estimated_end_date, estimated_duration, critical_path
		FROM contract_milestones
		WHERE contract_id = $1
		ORDER BY sequence_number`

	rows, err := r.db.QueryContext(ctx, query, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone timeline data: %w", err)
	}
	defer rows.Close()

	analysis := &MilestoneTimelineAnalysis{
		ContractID: contractID,
		Milestones: make([]MilestoneTimelineEntry, 0),
	}

	var totalDuration time.Duration
	var criticalPathDuration time.Duration

	for rows.Next() {
		var (
			id                           string
			estimatedStart, estimatedEnd sql.NullTime
			estimatedDuration            sql.NullInt64
			criticalPath                 bool
		)

		err := rows.Scan(&id, &estimatedStart, &estimatedEnd, &estimatedDuration, &criticalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline data: %w", err)
		}

		entry := MilestoneTimelineEntry{
			MilestoneID: id,
			IsCritical:  criticalPath,
		}

		if estimatedStart.Valid {
			entry.EarliestStart = estimatedStart.Time
			entry.LatestStart = estimatedStart.Time
		}
		if estimatedEnd.Valid {
			entry.EarliestFinish = estimatedEnd.Time
			entry.LatestFinish = estimatedEnd.Time
		}

		if estimatedDuration.Valid {
			duration := time.Duration(estimatedDuration.Int64)
			totalDuration += duration
			if criticalPath {
				criticalPathDuration += duration
			}
		}

		analysis.Milestones = append(analysis.Milestones, entry)
	}

	analysis.TotalDuration = totalDuration
	analysis.CriticalPathDuration = criticalPathDuration
	analysis.SlackTime = totalDuration - criticalPathDuration

	return analysis, rows.Err()
}

// GetMilestoneRiskAnalysis performs risk analysis on milestones
func (r *PostgresMilestoneRepository) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*MilestoneRiskAnalysis, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	whereClause.WriteString("WHERE 1=1")

	if contractID != nil {
		whereClause.WriteString(fmt.Sprintf(" AND contract_id = $%d", argIndex))
		args = append(args, *contractID)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_count,
			COUNT(CASE WHEN risk_level = 'high' THEN 1 END) as high_risk_count,
			COUNT(CASE WHEN risk_level = 'medium' THEN 1 END) as medium_risk_count,
			COUNT(CASE WHEN risk_level = 'low' THEN 1 END) as low_risk_count,
			AVG(criticality_score) as avg_risk_score
		FROM contract_milestones
		%s`, whereClause.String())

	var totalCount, highCount, mediumCount, lowCount int
	var avgRiskScore sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&totalCount, &highCount, &mediumCount, &lowCount, &avgRiskScore,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone risk analysis: %w", err)
	}

	analysis := &MilestoneRiskAnalysis{
		HighRiskCount:   highCount,
		MediumRiskCount: mediumCount,
		LowRiskCount:    lowCount,
		RiskDistribution: map[string]int{
			"high":   highCount,
			"medium": mediumCount,
			"low":    lowCount,
		},
	}

	if avgRiskScore.Valid {
		analysis.OverallRiskScore = avgRiskScore.Float64
	}

	// Get detailed risk milestones
	riskQuery := fmt.Sprintf(`
		SELECT id, risk_level, criticality_score, contingency_plans
		FROM contract_milestones
		%s
		ORDER BY criticality_score DESC`, whereClause.String())

	rows, err := r.db.QueryContext(ctx, riskQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk milestones: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, riskLevel string
		var criticalityScore int
		var contingencyPlans pq.StringArray

		err := rows.Scan(&id, &riskLevel, &criticalityScore, &contingencyPlans)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk milestone: %w", err)
		}

		riskEntry := MilestoneRiskEntry{
			MilestoneID:      id,
			RiskLevel:        riskLevel,
			RiskScore:        float64(criticalityScore),
			ContingencyPlans: []string(contingencyPlans),
		}

		analysis.RiskMilestones = append(analysis.RiskMilestones, riskEntry)
	}

	return analysis, rows.Err()
}

// GetMilestoneProgressTrends retrieves milestone progress trends
func (r *PostgresMilestoneRepository) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*MilestoneProgressTrend, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	whereClause.WriteString("WHERE 1=1")

	if contractID != nil {
		whereClause.WriteString(fmt.Sprintf(" AND contract_id = $%d", argIndex))
		args = append(args, *contractID)
		argIndex++
	}

	query := fmt.Sprintf(`
		WITH date_series AS (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '%d days',
				CURRENT_DATE,
				INTERVAL '1 day'
			)::date AS date
		),
		milestone_stats AS (
			SELECT
				DATE(updated_at) as update_date,
				COUNT(*) as total_milestones,
				COUNT(CASE WHEN percentage_complete = 100 THEN 1 END) as completed_milestones,
				AVG(percentage_complete) as avg_completion
			FROM contract_milestones
			%s
			GROUP BY DATE(updated_at)
		)
		SELECT
			ds.date,
			COALESCE(ms.avg_completion, 0) as completion_rate,
			COALESCE(ms.completed_milestones, 0) as milestones_completed,
			COALESCE(ms.total_milestones, 0) as total_milestones
		FROM date_series ds
		LEFT JOIN milestone_stats ms ON ds.date = ms.update_date
		ORDER BY ds.date`, days, whereClause.String())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone progress trends: %w", err)
	}
	defer rows.Close()

	var trends []*MilestoneProgressTrend
	for rows.Next() {
		trend := &MilestoneProgressTrend{}
		err := rows.Scan(
			&trend.Date,
			&trend.CompletionRate,
			&trend.MilestonesCompleted,
			&trend.TotalMilestones,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress trend: %w", err)
		}
		trends = append(trends, trend)
	}

	return trends, rows.Err()
}

// GetDelayedMilestonesReport generates a report of delayed milestones
func (r *PostgresMilestoneRepository) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*DelayedMilestoneReport, error) {
	thresholdDate := time.Now().Add(-threshold)

	query := `
		SELECT
			id,
			contract_id,
			trigger_conditions as description,
			estimated_end_date,
			actual_end_date,
			EXTRACT(EPOCH FROM (COALESCE(actual_end_date, NOW()) - estimated_end_date)) as delay_seconds
		FROM contract_milestones
		WHERE estimated_end_date < $1
		  AND percentage_complete < 100
		ORDER BY estimated_end_date`

	rows, err := r.db.QueryContext(ctx, query, thresholdDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed milestones report: %w", err)
	}
	defer rows.Close()

	var reports []*DelayedMilestoneReport
	for rows.Next() {
		var (
			id, contractID, description string
			estimatedEnd                time.Time
			actualEnd                   sql.NullTime
			delaySeconds                float64
		)

		err := rows.Scan(&id, &contractID, &description, &estimatedEnd, &actualEnd, &delaySeconds)
		if err != nil {
			return nil, fmt.Errorf("failed to scan delayed milestone: %w", err)
		}

		report := &DelayedMilestoneReport{
			MilestoneID:     id,
			ContractID:      contractID,
			Description:     description,
			OriginalDueDate: estimatedEnd,
			DelayDuration:   time.Duration(delaySeconds) * time.Second,
			DelayReason:     "Unknown", // Would be populated from additional data
			ImpactLevel:     "Medium",  // Would be calculated based on criticality
		}

		if actualEnd.Valid {
			report.CurrentDueDate = &actualEnd.Time
		}

		reports = append(reports, report)
	}

	return reports, rows.Err()
}

// Progress tracking and history

// CreateMilestoneProgressEntry creates a new progress entry
func (r *PostgresMilestoneRepository) CreateMilestoneProgressEntry(ctx context.Context, entry *MilestoneProgressEntry) error {
	query := `
		INSERT INTO milestone_progress_history (
			id, milestone_id, percentage_complete, status, notes, recorded_by, recorded_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.MilestoneID,
		entry.PercentageComplete,
		entry.Status,
		entry.Notes,
		entry.RecordedBy,
		entry.RecordedAt,
		entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create milestone progress entry: %w", err)
	}
	return nil
}

// GetMilestoneProgressHistory retrieves progress history for a milestone
func (r *PostgresMilestoneRepository) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*MilestoneProgressEntry, error) {
	query := `
		SELECT id, milestone_id, percentage_complete, status, notes, recorded_by, recorded_at, created_at
		FROM milestone_progress_history
		WHERE milestone_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, milestoneID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone progress history: %w", err)
	}
	defer rows.Close()

	var entries []*MilestoneProgressEntry
	for rows.Next() {
		entry := &MilestoneProgressEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.MilestoneID,
			&entry.PercentageComplete,
			&entry.Status,
			&entry.Notes,
			&entry.RecordedBy,
			&entry.RecordedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// GetLatestProgressUpdate retrieves the latest progress update for a milestone
func (r *PostgresMilestoneRepository) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*MilestoneProgressEntry, error) {
	query := `
		SELECT id, milestone_id, percentage_complete, status, notes, recorded_by, recorded_at, created_at
		FROM milestone_progress_history
		WHERE milestone_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1`

	entry := &MilestoneProgressEntry{}
	err := r.db.QueryRowContext(ctx, query, milestoneID).Scan(
		&entry.ID,
		&entry.MilestoneID,
		&entry.PercentageComplete,
		&entry.Status,
		&entry.Notes,
		&entry.RecordedBy,
		&entry.RecordedAt,
		&entry.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no progress updates found for milestone %s", milestoneID)
		}
		return nil, fmt.Errorf("failed to get latest progress update: %w", err)
	}

	return entry, nil
}

// Search and filtering capabilities

// FilterMilestones filters milestones based on provided criteria
func (r *PostgresMilestoneRepository) FilterMilestones(ctx context.Context, filter *MilestoneFilter) ([]*models.ContractMilestone, error) {
	var whereClause strings.Builder
	var args []interface{}
	argIndex := 1

	whereClause.WriteString("WHERE 1=1")

	if filter.ContractID != nil {
		whereClause.WriteString(fmt.Sprintf(" AND contract_id = $%d", argIndex))
		args = append(args, *filter.ContractID)
		argIndex++
	}

	if filter.Category != nil {
		whereClause.WriteString(fmt.Sprintf(" AND category = $%d", argIndex))
		args = append(args, *filter.Category)
		argIndex++
	}

	if filter.Priority != nil {
		whereClause.WriteString(fmt.Sprintf(" AND priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	if filter.RiskLevel != nil {
		whereClause.WriteString(fmt.Sprintf(" AND risk_level = $%d", argIndex))
		args = append(args, *filter.RiskLevel)
		argIndex++
	}

	if filter.CriticalPath != nil {
		whereClause.WriteString(fmt.Sprintf(" AND critical_path = $%d", argIndex))
		args = append(args, *filter.CriticalPath)
		argIndex++
	}

	if filter.StartDateFrom != nil {
		whereClause.WriteString(fmt.Sprintf(" AND estimated_start_date >= $%d", argIndex))
		args = append(args, *filter.StartDateFrom)
		argIndex++
	}

	if filter.StartDateTo != nil {
		whereClause.WriteString(fmt.Sprintf(" AND estimated_start_date <= $%d", argIndex))
		args = append(args, *filter.StartDateTo)
		argIndex++
	}

	if filter.DueDateFrom != nil {
		whereClause.WriteString(fmt.Sprintf(" AND estimated_end_date >= $%d", argIndex))
		args = append(args, *filter.DueDateFrom)
		argIndex++
	}

	if filter.DueDateTo != nil {
		whereClause.WriteString(fmt.Sprintf(" AND estimated_end_date <= $%d", argIndex))
		args = append(args, *filter.DueDateTo)
		argIndex++
	}

	if filter.MinCompletion != nil {
		whereClause.WriteString(fmt.Sprintf(" AND percentage_complete >= $%d", argIndex))
		args = append(args, *filter.MinCompletion)
		argIndex++
	}

	if filter.MaxCompletion != nil {
		whereClause.WriteString(fmt.Sprintf(" AND percentage_complete <= $%d", argIndex))
		args = append(args, *filter.MaxCompletion)
		argIndex++
	}

	if filter.SearchText != nil {
		whereClause.WriteString(fmt.Sprintf(" AND (trigger_conditions ILIKE $%d OR verification_criteria ILIKE $%d)", argIndex, argIndex))
		searchPattern := "%" + *filter.SearchText + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	if len(filter.ExcludeStatuses) > 0 {
		// Convert status to percentage completion conditions
		for _, status := range filter.ExcludeStatuses {
			switch status {
			case MilestoneStatusPending:
				whereClause.WriteString(" AND percentage_complete != 0")
			case MilestoneStatusCompleted:
				whereClause.WriteString(" AND percentage_complete != 100")
			case MilestoneStatusInProgress:
				whereClause.WriteString(" AND NOT (percentage_complete > 0 AND percentage_complete < 100)")
			}
		}
	}

	query := fmt.Sprintf(`
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		%s
		ORDER BY created_at DESC`, whereClause.String())

	return r.scanMilestones(ctx, query, args...)
}

// GetMilestonesByDateRange retrieves milestones within a date range
func (r *PostgresMilestoneRepository) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	return r.scanMilestones(ctx, query, startDate, endDate, limit, offset)
}

// GetUpcomingMilestones retrieves milestones due within specified days
func (r *PostgresMilestoneRepository) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	futureDate := time.Now().AddDate(0, 0, daysAhead)

	query := `
		SELECT id, contract_id, milestone_id, sequence_number, dependencies, category,
		       priority, critical_path, trigger_conditions, verification_criteria,
		       estimated_start_date, estimated_end_date, actual_start_date, actual_end_date,
		       estimated_duration, actual_duration, percentage_complete, risk_level,
		       contingency_plans, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE estimated_end_date BETWEEN NOW() AND $1
		  AND percentage_complete < 100
		ORDER BY estimated_end_date
		LIMIT $2 OFFSET $3`

	return r.scanMilestones(ctx, query, futureDate, limit, offset)
}

// Helper methods

// scanSingleMilestone scans a single milestone from a query
func (r *PostgresMilestoneRepository) scanSingleMilestone(ctx context.Context, query string, args ...interface{}) (*models.ContractMilestone, error) {
	milestone := &models.ContractMilestone{}
	var (
		estimatedDuration, actualDuration sql.NullInt64
		estimatedStart, estimatedEnd      sql.NullTime
		actualStart, actualEnd            sql.NullTime
		dependencies, contingencyPlans    pq.StringArray
	)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&milestone.ID,
		&milestone.ContractID,
		&milestone.MilestoneID,
		&milestone.SequenceNumber,
		&dependencies,
		&milestone.Category,
		&milestone.Priority,
		&milestone.CriticalPath,
		&milestone.TriggerConditions,
		&milestone.VerificationCriteria,
		&estimatedStart,
		&estimatedEnd,
		&actualStart,
		&actualEnd,
		&estimatedDuration,
		&actualDuration,
		&milestone.PercentageComplete,
		&milestone.RiskLevel,
		&contingencyPlans,
		&milestone.CriticalityScore,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone not found")
		}
		return nil, err
	}

	milestone.Dependencies = []string(dependencies)
	milestone.ContingencyPlans = []string(contingencyPlans)

	if estimatedStart.Valid {
		milestone.EstimatedStartDate = &estimatedStart.Time
	}
	if estimatedEnd.Valid {
		milestone.EstimatedEndDate = &estimatedEnd.Time
	}
	if actualStart.Valid {
		milestone.ActualStartDate = &actualStart.Time
	}
	if actualEnd.Valid {
		milestone.ActualEndDate = &actualEnd.Time
	}
	if estimatedDuration.Valid {
		milestone.EstimatedDuration = time.Duration(estimatedDuration.Int64)
	}
	if actualDuration.Valid {
		duration := time.Duration(actualDuration.Int64)
		milestone.ActualDuration = &duration
	}

	return milestone, nil
}

// scanMilestones scans multiple milestones from a query
func (r *PostgresMilestoneRepository) scanMilestones(ctx context.Context, query string, args ...interface{}) ([]*models.ContractMilestone, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []*models.ContractMilestone
	for rows.Next() {
		milestone := &models.ContractMilestone{}
		var (
			estimatedDuration, actualDuration sql.NullInt64
			estimatedStart, estimatedEnd      sql.NullTime
			actualStart, actualEnd            sql.NullTime
			dependencies, contingencyPlans    pq.StringArray
		)

		err := rows.Scan(
			&milestone.ID,
			&milestone.ContractID,
			&milestone.MilestoneID,
			&milestone.SequenceNumber,
			&dependencies,
			&milestone.Category,
			&milestone.Priority,
			&milestone.CriticalPath,
			&milestone.TriggerConditions,
			&milestone.VerificationCriteria,
			&estimatedStart,
			&estimatedEnd,
			&actualStart,
			&actualEnd,
			&estimatedDuration,
			&actualDuration,
			&milestone.PercentageComplete,
			&milestone.RiskLevel,
			&contingencyPlans,
			&milestone.CriticalityScore,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}

		milestone.Dependencies = []string(dependencies)
		milestone.ContingencyPlans = []string(contingencyPlans)

		if estimatedStart.Valid {
			milestone.EstimatedStartDate = &estimatedStart.Time
		}
		if estimatedEnd.Valid {
			milestone.EstimatedEndDate = &estimatedEnd.Time
		}
		if actualStart.Valid {
			milestone.ActualStartDate = &actualStart.Time
		}
		if actualEnd.Valid {
			milestone.ActualEndDate = &actualEnd.Time
		}
		if estimatedDuration.Valid {
			milestone.EstimatedDuration = time.Duration(estimatedDuration.Int64)
		}
		if actualDuration.Valid {
			duration := time.Duration(actualDuration.Int64)
			milestone.ActualDuration = &duration
		}

		milestones = append(milestones, milestone)
	}

	return milestones, rows.Err()
}
