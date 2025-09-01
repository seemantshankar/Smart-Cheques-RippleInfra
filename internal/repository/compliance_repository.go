package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// ComplianceRepository handles database operations for transaction compliance
type ComplianceRepository struct {
	db *sql.DB
}

// NewComplianceRepository creates a new compliance repository
func NewComplianceRepository(db *sql.DB) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

// CreateComplianceStatus creates a new compliance status record
func (r *ComplianceRepository) CreateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	query := `
		INSERT INTO transaction_compliance_status (
			id, transaction_id, status, checks_passed, checks_failed, violations,
			reviewed_by, reviewed_at, comments, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	complianceStatus.ID = uuid.New().String()
	complianceStatus.CreatedAt = time.Now()
	complianceStatus.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		complianceStatus.ID,
		complianceStatus.TransactionID,
		complianceStatus.Status,
		complianceStatus.ChecksPassed,
		complianceStatus.ChecksFailed,
		complianceStatus.Violations,
		complianceStatus.ReviewedBy,
		complianceStatus.ReviewedAt,
		complianceStatus.Comments,
		complianceStatus.CreatedAt,
		complianceStatus.UpdatedAt,
	)

	return err
}

// GetComplianceStatus retrieves compliance status for a transaction
func (r *ComplianceRepository) GetComplianceStatus(transactionID string) (*models.TransactionComplianceStatus, error) {
	query := `
		SELECT id, transaction_id, status, checks_passed, checks_failed, violations,
		       reviewed_by, reviewed_at, comments, created_at, updated_at
		FROM transaction_compliance_status
		WHERE transaction_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var status models.TransactionComplianceStatus
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := r.db.QueryRow(query, transactionID).Scan(
		&status.ID,
		&status.TransactionID,
		&status.Status,
		&status.ChecksPassed,
		&status.ChecksFailed,
		&status.Violations,
		&reviewedBy,
		&reviewedAt,
		&status.Comments,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if reviewedBy.Valid {
		status.ReviewedBy = &reviewedBy.String
	}
	if reviewedAt.Valid {
		status.ReviewedAt = &reviewedAt.Time
	}

	return &status, nil
}

// UpdateComplianceStatus updates an existing compliance status record
func (r *ComplianceRepository) UpdateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	query := `
		UPDATE transaction_compliance_status
		SET status = $1, checks_passed = $2, checks_failed = $3, violations = $4,
		    reviewed_by = $5, reviewed_at = $6, comments = $7, updated_at = $8
		WHERE id = $9
	`

	complianceStatus.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		complianceStatus.Status,
		complianceStatus.ChecksPassed,
		complianceStatus.ChecksFailed,
		complianceStatus.Violations,
		complianceStatus.ReviewedBy,
		complianceStatus.ReviewedAt,
		complianceStatus.Comments,
		complianceStatus.UpdatedAt,
		complianceStatus.ID,
	)

	return err
}

// GetComplianceStatusesByStatus retrieves compliance statuses by status
func (r *ComplianceRepository) GetComplianceStatusesByStatus(status string, limit, offset int) ([]models.TransactionComplianceStatus, error) {
	query := `
		SELECT id, transaction_id, status, checks_passed, checks_failed, violations,
		       reviewed_by, reviewed_at, comments, created_at, updated_at
		FROM transaction_compliance_status
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.getComplianceStatuses(query, status, limit, offset)
}

// GetComplianceStatusesByEnterprise retrieves compliance statuses for an enterprise
func (r *ComplianceRepository) GetComplianceStatusesByEnterprise(enterpriseID string, limit, offset int) ([]models.TransactionComplianceStatus, error) {
	query := `
		SELECT cs.id, cs.transaction_id, cs.status, cs.checks_passed, cs.checks_failed, cs.violations,
		       cs.reviewed_by, cs.reviewed_at, cs.comments, cs.created_at, cs.updated_at
		FROM transaction_compliance_status cs
		INNER JOIN transactions t ON cs.transaction_id = t.id
		WHERE t.enterprise_id = $1
		ORDER BY cs.created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.getComplianceStatuses(query, enterpriseID, limit, offset)
}

// GetFlaggedTransactions retrieves transactions that are flagged for review
func (r *ComplianceRepository) GetFlaggedTransactions(limit, offset int) ([]models.TransactionComplianceStatus, error) {
	query := `
		SELECT id, transaction_id, status, checks_passed, checks_failed, violations,
		       reviewed_by, reviewed_at, comments, created_at, updated_at
		FROM transaction_compliance_status
		WHERE status = 'flagged' AND reviewed_by IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.getComplianceStatuses(query, limit, offset)
}

// ReviewComplianceStatus marks a compliance status as reviewed
func (r *ComplianceRepository) ReviewComplianceStatus(complianceStatusID string, reviewedBy string, comments string) error {
	query := `
		UPDATE transaction_compliance_status
		SET reviewed_by = $1, reviewed_at = $2, comments = $3, updated_at = $4
		WHERE id = $5
	`

	reviewedAt := time.Now()
	updatedAt := time.Now()

	_, err := r.db.Exec(query, reviewedBy, reviewedAt, comments, updatedAt, complianceStatusID)
	return err
}

// GetComplianceStats retrieves compliance statistics
func (r *ComplianceRepository) GetComplianceStats(enterpriseID *string, since *time.Time) (*models.ComplianceStats, error) {
	baseQuery := `
		SELECT
			COUNT(*) as total_transactions,
			COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_transactions,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_transactions,
			COUNT(CASE WHEN status = 'flagged' THEN 1 END) as flagged_transactions,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_transactions,
			COUNT(CASE WHEN reviewed_by IS NOT NULL THEN 1 END) as reviewed_transactions
		FROM transaction_compliance_status cs
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if enterpriseID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM transactions t WHERE t.id = cs.transaction_id AND t.enterprise_id = $%d)", argIndex))
		args = append(args, *enterpriseID)
		argIndex++
	}

	if since != nil {
		conditions = append(conditions, fmt.Sprintf("cs.created_at >= $%d", argIndex))
		args = append(args, *since)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", conditions[0]))
		for i := 1; i < len(conditions); i++ {
			query += fmt.Sprintf(" AND (%s)", conditions[i])
		}
	}

	var stats models.ComplianceStats
	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalTransactions,
		&stats.ApprovedTransactions,
		&stats.RejectedTransactions,
		&stats.FlaggedTransactions,
		&stats.PendingTransactions,
		&stats.ReviewedTransactions,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// Helper method to execute compliance status queries
func (r *ComplianceRepository) getComplianceStatuses(query string, args ...interface{}) ([]models.TransactionComplianceStatus, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []models.TransactionComplianceStatus
	for rows.Next() {
		var status models.TransactionComplianceStatus
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime

		err := rows.Scan(
			&status.ID,
			&status.TransactionID,
			&status.Status,
			&status.ChecksPassed,
			&status.ChecksFailed,
			&status.Violations,
			&reviewedBy,
			&reviewedAt,
			&status.Comments,
			&status.CreatedAt,
			&status.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if reviewedBy.Valid {
			status.ReviewedBy = &reviewedBy.String
		}
		if reviewedAt.Valid {
			status.ReviewedAt = &reviewedAt.Time
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}
