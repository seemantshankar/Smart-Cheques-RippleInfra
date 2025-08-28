package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

// PostgresContractRepository implements ContractRepositoryInterface using PostgreSQL
type PostgresContractRepository struct {
	db *sql.DB
}

func NewPostgresContractRepository(db *sql.DB) ContractRepositoryInterface {
	return &PostgresContractRepository{db: db}
}

// CreateContract inserts a new contract
func (r *PostgresContractRepository) CreateContract(ctx context.Context, c *models.Contract) error {
	query := `
		INSERT INTO contracts (
			id, parties, status, contract_type, version, parent_contract_id,
			expiration_date, renewal_terms, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10
		)`

	var parent interface{}
	if c.ParentContractID != nil {
		parent = *c.ParentContractID
	} else {
		parent = nil
	}

	_, err := r.db.ExecContext(
		ctx, query,
		c.ID,
		pq.Array(c.Parties),
		c.Status,
		c.ContractType,
		c.Version,
		parent,
		c.ExpirationDate,
		c.RenewalTerms,
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}
	return nil
}

// GetContractByID retrieves a contract by ID
func (r *PostgresContractRepository) GetContractByID(ctx context.Context, id string) (*models.Contract, error) {
	query := `
		SELECT id, parties, status, contract_type, version, parent_contract_id,
		       expiration_date, renewal_terms, created_at, updated_at
		FROM contracts
		WHERE id = $1`

	var (
		parent     sql.NullString
		expiration sql.NullTime
	)

	c := &models.Contract{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		pq.Array(&c.Parties),
		&c.Status,
		&c.ContractType,
		&c.Version,
		&parent,
		&expiration,
		&c.RenewalTerms,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get contract by ID: %w", err)
	}

	if parent.Valid {
		val := parent.String
		c.ParentContractID = &val
	}
	if expiration.Valid {
		c.ExpirationDate = &expiration.Time
	}

	return c, nil
}

// UpdateContract updates a contract
func (r *PostgresContractRepository) UpdateContract(ctx context.Context, c *models.Contract) error {
	query := `
		UPDATE contracts SET
			parties = $2, status = $3, contract_type = $4, version = $5,
			parent_contract_id = $6, expiration_date = $7, renewal_terms = $8, updated_at = $9
		WHERE id = $1`

	var parent interface{}
	if c.ParentContractID != nil {
		parent = *c.ParentContractID
	} else {
		parent = nil
	}

	res, err := r.db.ExecContext(
		ctx, query,
		c.ID,
		pq.Array(c.Parties),
		c.Status,
		c.ContractType,
		c.Version,
		parent,
		c.ExpirationDate,
		c.RenewalTerms,
		c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("contract with ID %s not found", c.ID)
	}
	return nil
}

// DeleteContract deletes a contract by ID
func (r *PostgresContractRepository) DeleteContract(ctx context.Context, id string) error {
	query := `DELETE FROM contracts WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("contract with ID %s not found", id)
	}
	return nil
}

// scanContracts is a helper method to scan contract rows from the database
func (r *PostgresContractRepository) scanContracts(rows *sql.Rows) ([]*models.Contract, error) {
	var list []*models.Contract
	for rows.Next() {
		var (
			c          models.Contract
			parent     sql.NullString
			expiration sql.NullTime
		)
		if err := rows.Scan(
			&c.ID,
			pq.Array(&c.Parties),
			&c.Status,
			&c.ContractType,
			&c.Version,
			&parent,
			&expiration,
			&c.RenewalTerms,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}
		if parent.Valid {
			val := parent.String
			c.ParentContractID = &val
		}
		if expiration.Valid {
			c.ExpirationDate = &expiration.Time
		}
		list = append(list, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contracts: %w", err)
	}
	return list, nil
}

// GetContractsByStatus lists contracts by status with pagination
func (r *PostgresContractRepository) GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error) {
	query := `
		SELECT id, parties, status, contract_type, version, parent_contract_id,
		       expiration_date, renewal_terms, created_at, updated_at
		FROM contracts
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contracts by status: %w", err)
	}
	defer rows.Close()

	return r.scanContracts(rows)
}

// GetContractsByType lists contracts by type with pagination
func (r *PostgresContractRepository) GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error) {
	query := `
		SELECT id, parties, status, contract_type, version, parent_contract_id,
		       expiration_date, renewal_terms, created_at, updated_at
		FROM contracts
		WHERE contract_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, contractType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contracts by type: %w", err)
	}
	defer rows.Close()

	return r.scanContracts(rows)
}

// GetContractsByParty lists contracts where party is present
func (r *PostgresContractRepository) GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error) {
	query := `
		SELECT id, parties, status, contract_type, version, parent_contract_id,
		       expiration_date, renewal_terms, created_at, updated_at
		FROM contracts
		WHERE $1 = ANY(parties)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, party, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contracts by party: %w", err)
	}
	defer rows.Close()

	return r.scanContracts(rows)
}

// PostgresContractMilestoneRepository implements ContractMilestoneRepositoryInterface
type PostgresContractMilestoneRepository struct {
	db *sql.DB
}

func NewPostgresContractMilestoneRepository(db *sql.DB) ContractMilestoneRepositoryInterface {
	return &PostgresContractMilestoneRepository{db: db}
}

// CreateMilestone inserts a new contract milestone
func (r *PostgresContractMilestoneRepository) CreateMilestone(ctx context.Context, m *models.ContractMilestone) error {
	query := `
		INSERT INTO contract_milestones (
			id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
			estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)`

	var actual interface{}
	if m.ActualDuration != nil {
		actual = int64(*m.ActualDuration)
	} else {
		actual = nil
	}

	_, err := r.db.ExecContext(ctx, query,
		m.ID,
		m.ContractID,
		m.MilestoneID,
		m.SequenceOrder,
		m.TriggerConditions,
		m.VerificationCriteria,
		int64(m.EstimatedDuration),
		actual,
		m.RiskLevel,
		m.CriticalityScore,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create contract milestone: %w", err)
	}
	return nil
}

// GetMilestoneByID retrieves a milestone by ID
func (r *PostgresContractMilestoneRepository) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
		       estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE id = $1`

	m := &models.ContractMilestone{}
	var est sql.NullInt64
	var act sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID,
		&m.ContractID,
		&m.MilestoneID,
		&m.SequenceOrder,
		&m.TriggerConditions,
		&m.VerificationCriteria,
		&est,
		&act,
		&m.RiskLevel,
		&m.CriticalityScore,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract milestone with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get contract milestone by ID: %w", err)
	}
	if est.Valid {
		m.EstimatedDuration = time.Duration(est.Int64)
	} else {
		m.EstimatedDuration = 0
	}
	if act.Valid {
		d := time.Duration(act.Int64)
		m.ActualDuration = &d
	} else {
		m.ActualDuration = nil
	}
	return m, nil
}

// UpdateMilestone updates an existing milestone
func (r *PostgresContractMilestoneRepository) UpdateMilestone(ctx context.Context, m *models.ContractMilestone) error {
	query := `
		UPDATE contract_milestones SET
			sequence_order = $2, trigger_conditions = $3, verification_criteria = $4,
			estimated_duration = $5, actual_duration = $6, risk_level = $7, criticality_score = $8,
			updated_at = $9
		WHERE id = $1`

	var actual interface{}
	if m.ActualDuration != nil {
		actual = int64(*m.ActualDuration)
	} else {
		actual = nil
	}

	res, err := r.db.ExecContext(ctx, query,
		m.ID,
		m.SequenceOrder,
		m.TriggerConditions,
		m.VerificationCriteria,
		int64(m.EstimatedDuration),
		actual,
		m.RiskLevel,
		m.CriticalityScore,
		m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update contract milestone: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("contract milestone with ID %s not found", m.ID)
	}
	return nil
}

// DeleteMilestone deletes a milestone by ID
func (r *PostgresContractMilestoneRepository) DeleteMilestone(ctx context.Context, id string) error {
	query := `DELETE FROM contract_milestones WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contract milestone: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("contract milestone with ID %s not found", id)
	}
	return nil
}

// GetMilestonesByContractID lists milestones for a contract
func (r *PostgresContractMilestoneRepository) GetMilestonesByContractID(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	query := `
		SELECT id, contract_id, milestone_id, sequence_order, trigger_conditions, verification_criteria,
		       estimated_duration, actual_duration, risk_level, criticality_score, created_at, updated_at
		FROM contract_milestones
		WHERE contract_id = $1
		ORDER BY sequence_order`

	rows, err := r.db.QueryContext(ctx, query, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestones by contract id: %w", err)
	}
	defer rows.Close()

	var list []*models.ContractMilestone
	for rows.Next() {
		var (
			m   models.ContractMilestone
			est sql.NullInt64
			act sql.NullInt64
		)
		if err := rows.Scan(
			&m.ID,
			&m.ContractID,
			&m.MilestoneID,
			&m.SequenceOrder,
			&m.TriggerConditions,
			&m.VerificationCriteria,
			&est,
			&act,
			&m.RiskLevel,
			&m.CriticalityScore,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contract milestone: %w", err)
		}
		if est.Valid {
			m.EstimatedDuration = time.Duration(est.Int64)
		} else {
			m.EstimatedDuration = 0
		}
		if act.Valid {
			d := time.Duration(act.Int64)
			m.ActualDuration = &d
		} else {
			m.ActualDuration = nil
		}
		list = append(list, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contract milestones: %w", err)
	}
	return list, nil
}
