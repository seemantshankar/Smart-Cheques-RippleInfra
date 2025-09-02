package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

// DisputeRepository implements the DisputeRepositoryInterface for PostgreSQL
type DisputeRepository struct {
	db *sql.DB
}

// NewDisputeRepository creates a new dispute repository instance
func NewDisputeRepository(db *sql.DB) *DisputeRepository {
	return &DisputeRepository{db: db}
}

// CreateDispute creates a new dispute in the database
func (r *DisputeRepository) CreateDispute(ctx context.Context, dispute *models.Dispute) error {
	if dispute.ID == "" {
		dispute.ID = uuid.New().String()
	}

	now := time.Now()
	dispute.CreatedAt = now
	dispute.UpdatedAt = now
	dispute.InitiatedAt = now
	dispute.LastActivityAt = now

	query := `
		INSERT INTO disputes (
			id, title, description, category, priority, status,
			smart_check_id, milestone_id, contract_id, transaction_id,
			initiator_id, initiator_type, respondent_id, respondent_type,
			disputed_amount, currency, initiated_at, last_activity_at,
			tags, metadata, created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24)`

	_, err := r.db.ExecContext(ctx, query,
		dispute.ID, dispute.Title, dispute.Description, dispute.Category,
		dispute.Priority, dispute.Status, dispute.SmartChequeID, dispute.MilestoneID,
		dispute.ContractID, dispute.TransactionID, dispute.InitiatorID, dispute.InitiatorType,
		dispute.RespondentID, dispute.RespondentType, dispute.DisputedAmount, dispute.Currency,
		dispute.InitiatedAt, dispute.LastActivityAt, pq.Array(dispute.Tags),
		dispute.Metadata, dispute.CreatedBy, dispute.UpdatedBy, dispute.CreatedAt, dispute.UpdatedAt)

	return err
}

// GetDisputeByID retrieves a dispute by its ID
func (r *DisputeRepository) GetDisputeByID(ctx context.Context, id string) (*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE id = $1`

	dispute := &models.Dispute{}
	var tags pq.StringArray
	var metadata map[string]interface{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dispute.ID, &dispute.Title, &dispute.Description, &dispute.Category,
		&dispute.Priority, &dispute.Status, &dispute.SmartChequeID, &dispute.MilestoneID,
		&dispute.ContractID, &dispute.TransactionID, &dispute.InitiatorID, &dispute.InitiatorType,
		&dispute.RespondentID, &dispute.RespondentType, &dispute.DisputedAmount, &dispute.Currency,
		&dispute.InitiatedAt, &dispute.LastActivityAt, &dispute.ResolvedAt, &dispute.ClosedAt,
		&tags, &metadata, &dispute.CreatedBy, &dispute.UpdatedBy,
		&dispute.CreatedAt, &dispute.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	dispute.Tags = []string(tags)
	dispute.Metadata = metadata

	return dispute, nil
}

// UpdateDispute updates an existing dispute
func (r *DisputeRepository) UpdateDispute(ctx context.Context, dispute *models.Dispute) error {
	dispute.UpdatedAt = time.Now()
	dispute.LastActivityAt = time.Now()

	query := `
		UPDATE disputes SET
			title = $2, description = $3, category = $4, priority = $5, status = $6,
			smart_check_id = $7, milestone_id = $8, contract_id = $9, transaction_id = $10,
			initiator_id = $11, initiator_type = $12, respondent_id = $13, respondent_type = $14,
			disputed_amount = $15, currency = $16, last_activity_at = $17,
			resolved_at = $18, closed_at = $19, tags = $20, metadata = $21,
			updated_by = $22, updated_at = $23
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		dispute.ID, dispute.Title, dispute.Description, dispute.Category,
		dispute.Priority, dispute.Status, dispute.SmartChequeID, dispute.MilestoneID,
		dispute.ContractID, dispute.TransactionID, dispute.InitiatorID, dispute.InitiatorType,
		dispute.RespondentID, dispute.RespondentType, dispute.DisputedAmount, dispute.Currency,
		dispute.LastActivityAt, dispute.ResolvedAt, dispute.ClosedAt, pq.Array(dispute.Tags),
		dispute.Metadata, dispute.UpdatedBy, dispute.UpdatedAt)

	return err
}

// DeleteDispute deletes a dispute by ID
func (r *DisputeRepository) DeleteDispute(ctx context.Context, id string) error {
	query := `DELETE FROM disputes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetDisputes retrieves disputes with filtering
func (r *DisputeRepository) GetDisputes(ctx context.Context, filter *models.DisputeFilter, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes`

	var conditions []string
	var args []interface{}
	argCount := 0

	if filter != nil {
		if filter.InitiatorID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("initiator_id = $%d", argCount))
			args = append(args, *filter.InitiatorID)
		}
		if filter.RespondentID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("respondent_id = $%d", argCount))
			args = append(args, *filter.RespondentID)
		}
		if filter.Category != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("category = $%d", argCount))
			args = append(args, *filter.Category)
		}
		if filter.Priority != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("priority = $%d", argCount))
			args = append(args, *filter.Priority)
		}
		if filter.Status != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
			args = append(args, *filter.Status)
		}
		if filter.SmartChequeID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("smart_check_id = $%d", argCount))
			args = append(args, *filter.SmartChequeID)
		}
		if filter.MilestoneID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("milestone_id = $%d", argCount))
			args = append(args, *filter.MilestoneID)
		}
		if filter.ContractID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("contract_id = $%d", argCount))
			args = append(args, *filter.ContractID)
		}
		if filter.DateFrom != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("initiated_at >= $%d", argCount))
			args = append(args, *filter.DateFrom)
		}
		if filter.DateTo != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("initiated_at <= $%d", argCount))
			args = append(args, *filter.DateTo)
		}
		if filter.MinAmount != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("disputed_amount >= $%d", argCount))
			args = append(args, *filter.MinAmount)
		}
		if filter.MaxAmount != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("disputed_amount <= $%d", argCount))
			args = append(args, *filter.MaxAmount)
		}
		if filter.SearchText != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("title || ' ' || description @@ plainto_tsquery('english', $%d)", argCount))
			args = append(args, *filter.SearchText)
		}
		if len(filter.Tags) > 0 {
			argCount++
			conditions = append(conditions, fmt.Sprintf("tags && $%d", argCount))
			args = append(args, pq.Array(filter.Tags))
		}
		if len(filter.ExcludeStatuses) > 0 {
			argCount++
			conditions = append(conditions, fmt.Sprintf("status != ALL($%d)", argCount))
			args = append(args, pq.Array(filter.ExcludeStatuses))
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	return r.executeDisputeQuery(ctx, query, args...)
}

// GetDisputesByInitiator retrieves disputes by initiator ID
func (r *DisputeRepository) GetDisputesByInitiator(ctx context.Context, initiatorID string, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE initiator_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, initiatorID)
}

// GetDisputesByRespondent retrieves disputes by respondent ID
func (r *DisputeRepository) GetDisputesByRespondent(ctx context.Context, respondentID string, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE respondent_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, respondentID)
}

// GetDisputesByStatus retrieves disputes by status
func (r *DisputeRepository) GetDisputesByStatus(ctx context.Context, status models.DisputeStatus, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE status = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, status)
}

// GetDisputesByCategory retrieves disputes by category
func (r *DisputeRepository) GetDisputesByCategory(ctx context.Context, category models.DisputeCategory, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE category = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, category)
}

// GetDisputesByPriority retrieves disputes by priority
func (r *DisputeRepository) GetDisputesByPriority(ctx context.Context, priority models.DisputePriority, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE priority = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, priority)
}

// GetDisputesBySmartCheque retrieves disputes by smart check ID
func (r *DisputeRepository) GetDisputesBySmartCheque(ctx context.Context, smartChequeID string, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE smart_check_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, smartChequeID)
}

// GetDisputesByMilestone retrieves disputes by milestone ID
func (r *DisputeRepository) GetDisputesByMilestone(ctx context.Context, milestoneID string, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE milestone_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, milestoneID)
}

// GetDisputesByContract retrieves disputes by contract ID
func (r *DisputeRepository) GetDisputesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE contract_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, contractID)
}

// GetActiveDisputes retrieves active (non-resolved, non-closed) disputes
func (r *DisputeRepository) GetActiveDisputes(ctx context.Context, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE status NOT IN ('resolved', 'closed', 'cancelled')
		ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query)
}

// GetOverdueDisputes retrieves disputes that are overdue (no activity for extended period)
func (r *DisputeRepository) GetOverdueDisputes(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.Dispute, error) {
	query := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE last_activity_at < $1 AND status NOT IN ('resolved', 'closed', 'cancelled')
		ORDER BY last_activity_at ASC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, query, asOfDate)
}

// SearchDisputes performs full-text search on disputes
func (r *DisputeRepository) SearchDisputes(ctx context.Context, query string, limit, offset int) ([]*models.Dispute, error) {
	sqlQuery := `
		SELECT id, title, description, category, priority, status,
			   smart_check_id, milestone_id, contract_id, transaction_id,
			   initiator_id, initiator_type, respondent_id, respondent_type,
			   disputed_amount, currency, initiated_at, last_activity_at,
			   resolved_at, closed_at, tags, metadata, created_by, updated_by,
			   created_at, updated_at
		FROM disputes WHERE title || ' ' || description @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank_cd(to_tsvector('english', title || ' ' || description), plainto_tsquery('english', $1)) DESC`

	if limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET %d", offset)
	}

	return r.executeDisputeQuery(ctx, sqlQuery, query)
}

// Helper method to execute dispute queries
func (r *DisputeRepository) executeDisputeQuery(ctx context.Context, query string, args ...interface{}) ([]*models.Dispute, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*models.Dispute
	for rows.Next() {
		dispute := &models.Dispute{}
		var tags pq.StringArray
		var metadata map[string]interface{}

		err := rows.Scan(
			&dispute.ID, &dispute.Title, &dispute.Description, &dispute.Category,
			&dispute.Priority, &dispute.Status, &dispute.SmartChequeID, &dispute.MilestoneID,
			&dispute.ContractID, &dispute.TransactionID, &dispute.InitiatorID, &dispute.InitiatorType,
			&dispute.RespondentID, &dispute.RespondentType, &dispute.DisputedAmount, &dispute.Currency,
			&dispute.InitiatedAt, &dispute.LastActivityAt, &dispute.ResolvedAt, &dispute.ClosedAt,
			&tags, &metadata, &dispute.CreatedBy, &dispute.UpdatedBy,
			&dispute.CreatedAt, &dispute.UpdatedAt)

		if err != nil {
			return nil, err
		}

		dispute.Tags = []string(tags)
		dispute.Metadata = metadata
		disputes = append(disputes, dispute)
	}

	return disputes, rows.Err()
}

// GetDisputeCount returns the total count of disputes
func (r *DisputeRepository) GetDisputeCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM disputes`
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetDisputeCountByStatus returns dispute count grouped by status
func (r *DisputeRepository) GetDisputeCountByStatus(ctx context.Context) (map[models.DisputeStatus]int64, error) {
	query := `SELECT status, COUNT(*) FROM disputes GROUP BY status`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[models.DisputeStatus]int64)
	for rows.Next() {
		var status models.DisputeStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}

	return counts, rows.Err()
}

// GetDisputeCountByCategory returns dispute count grouped by category
func (r *DisputeRepository) GetDisputeCountByCategory(ctx context.Context) (map[models.DisputeCategory]int64, error) {
	query := `SELECT category, COUNT(*) FROM disputes GROUP BY category`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[models.DisputeCategory]int64)
	for rows.Next() {
		var category models.DisputeCategory
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		counts[category] = count
	}

	return counts, rows.Err()
}

// GetDisputeCountByPriority returns dispute count grouped by priority
func (r *DisputeRepository) GetDisputeCountByPriority(ctx context.Context) (map[models.DisputePriority]int64, error) {
	query := `SELECT priority, COUNT(*) FROM disputes GROUP BY priority`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[models.DisputePriority]int64)
	for rows.Next() {
		var priority models.DisputePriority
		var count int64
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		counts[priority] = count
	}

	return counts, rows.Err()
}

// GetDisputeStats returns comprehensive dispute statistics
func (r *DisputeRepository) GetDisputeStats(ctx context.Context) (*models.DisputeStats, error) {
	query := `
		SELECT
			COUNT(*) as total_disputes,
			COUNT(*) FILTER (WHERE status IN ('resolved', 'closed')) as resolved_disputes,
			AVG(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as avg_resolution_time,
			MIN(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as min_resolution_time,
			MAX(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as max_resolution_time
		FROM disputes`

	stats := &models.DisputeStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalDisputes,
		&stats.ResolvedDisputes,
		&stats.AverageResolutionTime,
		&stats.DisputesByStatus, // This would need separate queries
		&stats.DisputesByCategory,
		&stats.DisputesByPriority,
	)

	if err != nil {
		return nil, err
	}

	// Calculate active disputes
	stats.ActiveDisputes = stats.TotalDisputes - stats.ResolvedDisputes

	// Get breakdowns by status, category, and priority
	statusCounts, err := r.GetDisputeCountByStatus(ctx)
	if err != nil {
		return nil, err
	}
	stats.DisputesByStatus = make(map[models.DisputeStatus]int64)
	for status, count := range statusCounts {
		stats.DisputesByStatus[status] = count
	}

	categoryCounts, err := r.GetDisputeCountByCategory(ctx)
	if err != nil {
		return nil, err
	}
	stats.DisputesByCategory = make(map[models.DisputeCategory]int64)
	for category, count := range categoryCounts {
		stats.DisputesByCategory[category] = count
	}

	priorityCounts, err := r.GetDisputeCountByPriority(ctx)
	if err != nil {
		return nil, err
	}
	stats.DisputesByPriority = make(map[models.DisputePriority]int64)
	for priority, count := range priorityCounts {
		stats.DisputesByPriority[priority] = count
	}

	return stats, nil
}

// Evidence operations
func (r *DisputeRepository) CreateEvidence(ctx context.Context, evidence *models.DisputeEvidence) error {
	if evidence.ID == "" {
		evidence.ID = uuid.New().String()
	}

	now := time.Now()
	evidence.CreatedAt = now
	evidence.UpdatedAt = now

	query := `
		INSERT INTO dispute_evidence (
			id, dispute_id, file_name, file_type, file_size, file_path,
			description, uploaded_by, is_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		evidence.ID, evidence.DisputeID, evidence.FileName, evidence.FileType,
		evidence.FileSize, evidence.FilePath, evidence.Description, evidence.UploadedBy,
		evidence.IsPublic, evidence.CreatedAt, evidence.UpdatedAt)

	return err
}

func (r *DisputeRepository) GetEvidenceByID(ctx context.Context, id string) (*models.DisputeEvidence, error) {
	query := `
		SELECT id, dispute_id, file_name, file_type, file_size, file_path,
			   description, uploaded_by, is_public, created_at, updated_at
		FROM dispute_evidence WHERE id = $1`

	evidence := &models.DisputeEvidence{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evidence.ID, &evidence.DisputeID, &evidence.FileName, &evidence.FileType,
		&evidence.FileSize, &evidence.FilePath, &evidence.Description, &evidence.UploadedBy,
		&evidence.IsPublic, &evidence.CreatedAt, &evidence.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return evidence, nil
}

func (r *DisputeRepository) GetEvidenceByDisputeID(ctx context.Context, disputeID string) ([]*models.DisputeEvidence, error) {
	query := `
		SELECT id, dispute_id, file_name, file_type, file_size, file_path,
			   description, uploaded_by, is_public, created_at, updated_at
		FROM dispute_evidence WHERE dispute_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidence []*models.DisputeEvidence
	for rows.Next() {
		e := &models.DisputeEvidence{}
		err := rows.Scan(
			&e.ID, &e.DisputeID, &e.FileName, &e.FileType, &e.FileSize, &e.FilePath,
			&e.Description, &e.UploadedBy, &e.IsPublic, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, e)
	}

	return evidence, rows.Err()
}

func (r *DisputeRepository) UpdateEvidence(ctx context.Context, evidence *models.DisputeEvidence) error {
	evidence.UpdatedAt = time.Now()

	query := `
		UPDATE dispute_evidence SET
			file_name = $2, file_type = $3, file_size = $4, file_path = $5,
			description = $6, uploaded_by = $7, is_public = $8, updated_at = $9
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		evidence.ID, evidence.FileName, evidence.FileType, evidence.FileSize,
		evidence.FilePath, evidence.Description, evidence.UploadedBy, evidence.IsPublic,
		evidence.UpdatedAt)

	return err
}

func (r *DisputeRepository) DeleteEvidence(ctx context.Context, id string) error {
	query := `DELETE FROM dispute_evidence WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Resolution operations
func (r *DisputeRepository) CreateResolution(ctx context.Context, resolution *models.DisputeResolution) error {
	if resolution.ID == "" {
		resolution.ID = uuid.New().String()
	}

	now := time.Now()
	resolution.CreatedAt = now
	resolution.UpdatedAt = now

	query := `
		INSERT INTO dispute_resolutions (
			id, dispute_id, method, resolution_details, outcome_amount, outcome_description,
			mediator_id, arbitrator_id, court_case_number, initiator_accepted, respondent_accepted,
			acceptance_deadline, is_executed, executed_at, executed_by, created_by, updated_by,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err := r.db.ExecContext(ctx, query,
		resolution.ID, resolution.DisputeID, resolution.Method, resolution.ResolutionDetails,
		resolution.OutcomeAmount, resolution.OutcomeDescription, resolution.MediatorID,
		resolution.ArbitratorID, resolution.CourtCaseNumber, resolution.InitiatorAccepted,
		resolution.RespondentAccepted, resolution.AcceptanceDeadline, resolution.IsExecuted,
		resolution.ExecutedAt, resolution.ExecutedBy, resolution.CreatedBy, resolution.UpdatedBy,
		resolution.CreatedAt, resolution.UpdatedAt)

	return err
}

func (r *DisputeRepository) GetResolutionByID(ctx context.Context, id string) (*models.DisputeResolution, error) {
	query := `
		SELECT id, dispute_id, method, resolution_details, outcome_amount, outcome_description,
			   mediator_id, arbitrator_id, court_case_number, initiator_accepted, respondent_accepted,
			   acceptance_deadline, is_executed, executed_at, executed_by, created_by, updated_by,
			   created_at, updated_at
		FROM dispute_resolutions WHERE id = $1`

	resolution := &models.DisputeResolution{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&resolution.ID, &resolution.DisputeID, &resolution.Method, &resolution.ResolutionDetails,
		&resolution.OutcomeAmount, &resolution.OutcomeDescription, &resolution.MediatorID,
		&resolution.ArbitratorID, &resolution.CourtCaseNumber, &resolution.InitiatorAccepted,
		&resolution.RespondentAccepted, &resolution.AcceptanceDeadline, &resolution.IsExecuted,
		&resolution.ExecutedAt, &resolution.ExecutedBy, &resolution.CreatedBy, &resolution.UpdatedBy,
		&resolution.CreatedAt, &resolution.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return resolution, nil
}

func (r *DisputeRepository) GetResolutionByDisputeID(ctx context.Context, disputeID string) (*models.DisputeResolution, error) {
	query := `
		SELECT id, dispute_id, method, resolution_details, outcome_amount, outcome_description,
			   mediator_id, arbitrator_id, court_case_number, initiator_accepted, respondent_accepted,
			   acceptance_deadline, is_executed, executed_at, executed_by, created_by, updated_by,
			   created_at, updated_at
		FROM dispute_resolutions WHERE dispute_id = $1 ORDER BY created_at DESC LIMIT 1`

	resolution := &models.DisputeResolution{}
	err := r.db.QueryRowContext(ctx, query, disputeID).Scan(
		&resolution.ID, &resolution.DisputeID, &resolution.Method, &resolution.ResolutionDetails,
		&resolution.OutcomeAmount, &resolution.OutcomeDescription, &resolution.MediatorID,
		&resolution.ArbitratorID, &resolution.CourtCaseNumber, &resolution.InitiatorAccepted,
		&resolution.RespondentAccepted, &resolution.AcceptanceDeadline, &resolution.IsExecuted,
		&resolution.ExecutedAt, &resolution.ExecutedBy, &resolution.CreatedBy, &resolution.UpdatedBy,
		&resolution.CreatedAt, &resolution.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return resolution, nil
}

func (r *DisputeRepository) UpdateResolution(ctx context.Context, resolution *models.DisputeResolution) error {
	resolution.UpdatedAt = time.Now()

	query := `
		UPDATE dispute_resolutions SET
			method = $2, resolution_details = $3, outcome_amount = $4, outcome_description = $5,
			mediator_id = $6, arbitrator_id = $7, court_case_number = $8, initiator_accepted = $9,
			respondent_accepted = $10, acceptance_deadline = $11, is_executed = $12, executed_at = $13,
			executed_by = $14, updated_by = $15, updated_at = $16
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		resolution.ID, resolution.Method, resolution.ResolutionDetails, resolution.OutcomeAmount,
		resolution.OutcomeDescription, resolution.MediatorID, resolution.ArbitratorID,
		resolution.CourtCaseNumber, resolution.InitiatorAccepted, resolution.RespondentAccepted,
		resolution.AcceptanceDeadline, resolution.IsExecuted, resolution.ExecutedAt,
		resolution.ExecutedBy, resolution.UpdatedBy, resolution.UpdatedAt)

	return err
}

func (r *DisputeRepository) DeleteResolution(ctx context.Context, id string) error {
	query := `DELETE FROM dispute_resolutions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Comment operations
func (r *DisputeRepository) CreateComment(ctx context.Context, comment *models.DisputeComment) error {
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	query := `
		INSERT INTO dispute_comments (
			id, dispute_id, author_id, author_type, content, is_internal, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		comment.ID, comment.DisputeID, comment.AuthorID, comment.AuthorType,
		comment.Content, comment.IsInternal, comment.CreatedAt, comment.UpdatedAt)

	return err
}

func (r *DisputeRepository) GetCommentByID(ctx context.Context, id string) (*models.DisputeComment, error) {
	query := `
		SELECT id, dispute_id, author_id, author_type, content, is_internal, created_at, updated_at
		FROM dispute_comments WHERE id = $1`

	comment := &models.DisputeComment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID, &comment.DisputeID, &comment.AuthorID, &comment.AuthorType,
		&comment.Content, &comment.IsInternal, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return comment, nil
}

func (r *DisputeRepository) GetCommentsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeComment, error) {
	query := `
		SELECT id, dispute_id, author_id, author_type, content, is_internal, created_at, updated_at
		FROM dispute_comments WHERE dispute_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.db.QueryContext(ctx, query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.DisputeComment
	for rows.Next() {
		comment := &models.DisputeComment{}
		err := rows.Scan(
			&comment.ID, &comment.DisputeID, &comment.AuthorID, &comment.AuthorType,
			&comment.Content, &comment.IsInternal, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func (r *DisputeRepository) UpdateComment(ctx context.Context, comment *models.DisputeComment) error {
	comment.UpdatedAt = time.Now()

	query := `
		UPDATE dispute_comments SET
			content = $2, is_internal = $3, updated_at = $4
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		comment.ID, comment.Content, comment.IsInternal, comment.UpdatedAt)

	return err
}

func (r *DisputeRepository) DeleteComment(ctx context.Context, id string) error {
	query := `DELETE FROM dispute_comments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Audit operations
func (r *DisputeRepository) CreateAuditLog(ctx context.Context, auditLog *models.DisputeAuditLog) error {
	if auditLog.ID == "" {
		auditLog.ID = uuid.New().String()
	}

	auditLog.CreatedAt = time.Now()

	query := `
		INSERT INTO dispute_audit_logs (
			id, dispute_id, action, user_id, user_type, details, old_value, new_value,
			ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID, auditLog.DisputeID, auditLog.Action, auditLog.UserID, auditLog.UserType,
		auditLog.Details, auditLog.OldValue, auditLog.NewValue, auditLog.IPAddress,
		auditLog.UserAgent, auditLog.CreatedAt)

	return err
}

func (r *DisputeRepository) GetAuditLogsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeAuditLog, error) {
	query := `
		SELECT id, dispute_id, action, user_id, user_type, details, old_value, new_value,
			   ip_address, user_agent, created_at
		FROM dispute_audit_logs WHERE dispute_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.db.QueryContext(ctx, query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auditLogs []*models.DisputeAuditLog
	for rows.Next() {
		log := &models.DisputeAuditLog{}
		err := rows.Scan(
			&log.ID, &log.DisputeID, &log.Action, &log.UserID, &log.UserType,
			&log.Details, &log.OldValue, &log.NewValue, &log.IPAddress,
			&log.UserAgent, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		auditLogs = append(auditLogs, log)
	}

	return auditLogs, rows.Err()
}

// Notification operations
func (r *DisputeRepository) CreateNotification(ctx context.Context, notification *models.DisputeNotification) error {
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	notification.CreatedAt = time.Now()

	query := `
		INSERT INTO dispute_notifications (
			id, dispute_id, recipient, type, channel, subject, message, status,
			sent_at, error_msg, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.DisputeID, notification.Recipient, notification.Type,
		notification.Channel, notification.Subject, notification.Message, notification.Status,
		notification.SentAt, notification.ErrorMsg, notification.Metadata, notification.CreatedAt)

	return err
}

func (r *DisputeRepository) GetNotificationByID(ctx context.Context, id string) (*models.DisputeNotification, error) {
	query := `
		SELECT id, dispute_id, recipient, type, channel, subject, message, status,
			   sent_at, error_msg, metadata, created_at
		FROM dispute_notifications WHERE id = $1`

	notification := &models.DisputeNotification{}
	var metadata map[string]interface{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID, &notification.DisputeID, &notification.Recipient, &notification.Type,
		&notification.Channel, &notification.Subject, &notification.Message, &notification.Status,
		&notification.SentAt, &notification.ErrorMsg, &metadata, &notification.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	notification.Metadata = metadata
	return notification, nil
}

func (r *DisputeRepository) GetNotificationsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeNotification, error) {
	query := `
		SELECT id, dispute_id, recipient, type, channel, subject, message, status,
			   sent_at, error_msg, metadata, created_at
		FROM dispute_notifications WHERE dispute_id = $1 ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := r.db.QueryContext(ctx, query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.DisputeNotification
	for rows.Next() {
		notification := &models.DisputeNotification{}
		var metadata map[string]interface{}

		err := rows.Scan(
			&notification.ID, &notification.DisputeID, &notification.Recipient, &notification.Type,
			&notification.Channel, &notification.Subject, &notification.Message, &notification.Status,
			&notification.SentAt, &notification.ErrorMsg, &metadata, &notification.CreatedAt)
		if err != nil {
			return nil, err
		}

		notification.Metadata = metadata
		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

func (r *DisputeRepository) UpdateNotification(ctx context.Context, notification *models.DisputeNotification) error {
	query := `
		UPDATE dispute_notifications SET
			status = $2, sent_at = $3, error_msg = $4, metadata = $5
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.Status, notification.SentAt,
		notification.ErrorMsg, notification.Metadata)

	return err
}

func (r *DisputeRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.DisputeNotification, error) {
	query := `
		SELECT id, dispute_id, recipient, type, channel, subject, message, status,
			   sent_at, error_msg, metadata, created_at
		FROM dispute_notifications WHERE status = 'pending' ORDER BY created_at ASC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.DisputeNotification
	for rows.Next() {
		notification := &models.DisputeNotification{}
		var metadata map[string]interface{}

		err := rows.Scan(
			&notification.ID, &notification.DisputeID, &notification.Recipient, &notification.Type,
			&notification.Channel, &notification.Subject, &notification.Message, &notification.Status,
			&notification.SentAt, &notification.ErrorMsg, &metadata, &notification.CreatedAt)
		if err != nil {
			return nil, err
		}

		notification.Metadata = metadata
		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}
