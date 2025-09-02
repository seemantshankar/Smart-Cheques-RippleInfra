package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
)

// smartChequeRepository implements SmartChequeRepositoryInterface
type smartChequeRepository struct {
	db *sql.DB
}

// NewSmartChequeRepository creates a new smart check repository
func NewSmartChequeRepository(db *sql.DB) SmartChequeRepositoryInterface {
	return &smartChequeRepository{db: db}
}

// CreateSmartCheque creates a new smart check in the database
func (r *smartChequeRepository) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	query := `
		INSERT INTO smart_checks (
			id, payer_id, payee_id, amount, currency, 
			milestones, escrow_address, status, contract_hash, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Convert milestones to JSON
	milestonesJSON, err := json.Marshal(smartCheque.Milestones)
	if err != nil {
		return fmt.Errorf("failed to marshal milestones: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx, query,
		smartCheque.ID,
		smartCheque.PayerID,
		smartCheque.PayeeID,
		smartCheque.Amount,
		string(smartCheque.Currency),
		milestonesJSON,
		smartCheque.EscrowAddress,
		string(smartCheque.Status),
		smartCheque.ContractHash,
		smartCheque.CreatedAt,
		smartCheque.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create smart check: %w", err)
	}

	return nil
}

// BatchCreateSmartCheques creates multiple smart checks in a single transaction
func (r *smartChequeRepository) BatchCreateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	if len(smartCheques) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Prepare statement for batch insert
	query := `
		INSERT INTO smart_checks (
			id, payer_id, payee_id, amount, currency, 
			milestones, escrow_address, status, contract_hash, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch insert
	for _, smartCheque := range smartCheques {
		// Convert milestones to JSON
		milestonesJSON, err := json.Marshal(smartCheque.Milestones)
		if err != nil {
			return fmt.Errorf("failed to marshal milestones for smart check %s: %w", smartCheque.ID, err)
		}

		_, err = stmt.ExecContext(
			ctx,
			smartCheque.ID,
			smartCheque.PayerID,
			smartCheque.PayeeID,
			smartCheque.Amount,
			string(smartCheque.Currency),
			milestonesJSON,
			smartCheque.EscrowAddress,
			string(smartCheque.Status),
			smartCheque.ContractHash,
			smartCheque.CreatedAt,
			smartCheque.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create smart check %s: %w", smartCheque.ID, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSmartChequeByID retrieves a smart check by its ID
func (r *smartChequeRepository) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE id = $1
	`

	var smartCheque models.SmartCheque
	var currencyStr string
	var statusStr string
	var milestonesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&smartCheque.ID,
		&smartCheque.PayerID,
		&smartCheque.PayeeID,
		&smartCheque.Amount,
		&currencyStr,
		&milestonesJSON,
		&smartCheque.EscrowAddress,
		&statusStr,
		&smartCheque.ContractHash,
		&smartCheque.CreatedAt,
		&smartCheque.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get smart check: %w", err)
	}

	// Convert string values to typed values
	smartCheque.Currency = models.Currency(currencyStr)
	smartCheque.Status = models.SmartChequeStatus(statusStr)

	// Unmarshal milestones
	if len(milestonesJSON) > 0 {
		if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
			return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
		}
	}

	return &smartCheque, nil
}

// UpdateSmartCheque updates an existing smart check
func (r *smartChequeRepository) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	query := `
		UPDATE smart_checks 
		SET payer_id = $1, payee_id = $2, amount = $3, currency = $4, 
		    milestones = $5, escrow_address = $6, status = $7, contract_hash = $8, 
		    updated_at = $9
		WHERE id = $10
	`

	// Convert milestones to JSON
	milestonesJSON, err := json.Marshal(smartCheque.Milestones)
	if err != nil {
		return fmt.Errorf("failed to marshal milestones: %w", err)
	}

	result, err := r.db.ExecContext(
		ctx, query,
		smartCheque.PayerID,
		smartCheque.PayeeID,
		smartCheque.Amount,
		string(smartCheque.Currency),
		milestonesJSON,
		smartCheque.EscrowAddress,
		string(smartCheque.Status),
		smartCheque.ContractHash,
		smartCheque.UpdatedAt,
		smartCheque.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update smart check: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("smart check not found: %s", smartCheque.ID)
	}

	return nil
}

// DeleteSmartCheque deletes a smart check by its ID
func (r *smartChequeRepository) DeleteSmartCheque(ctx context.Context, id string) error {
	query := `DELETE FROM smart_checks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete smart check: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("smart check not found: %s", id)
	}

	return nil
}

// GetSmartChequesByPayer retrieves smart checks by payer ID
func (r *smartChequeRepository) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	return r.getSmartChequesByEntity(ctx, payerID, "payer_id", limit, offset)
}

// getSmartChequesByEntity retrieves smart checks by entity type (payer/payee/contract)
func (r *smartChequeRepository) getSmartChequesByEntity(ctx context.Context, entityID string, entityColumn string, limit, offset int) ([]*models.SmartCheque, error) {
	query := fmt.Sprintf(`
		SELECT id, payer_id, payee_id, amount, currency,
		       milestones, escrow_address, status, contract_hash,
		       created_at, updated_at
		FROM smart_checks
		WHERE %s = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, entityColumn)

	rows, err := r.db.QueryContext(ctx, query, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart checks: %w", err)
	}
	defer rows.Close()

	smartCheques := make([]*models.SmartCheque, 0, 50) // Pre-allocate for better performance
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var statusStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&statusStr,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Convert string values to typed values
		smartCheque.Currency = models.Currency(currencyStr)
		smartCheque.Status = models.SmartChequeStatus(statusStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}

// GetSmartChequesByPayee retrieves smart checks by payee ID
func (r *smartChequeRepository) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE payee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, payeeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart checks: %w", err)
	}
	defer rows.Close()

	var smartCheques []*models.SmartCheque
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var statusStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&statusStr,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Convert string values to typed values
		smartCheque.Currency = models.Currency(currencyStr)
		smartCheque.Status = models.SmartChequeStatus(statusStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}

// GetSmartChequesByStatus retrieves smart checks by status
func (r *smartChequeRepository) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart checks: %w", err)
	}
	defer rows.Close()

	var smartCheques []*models.SmartCheque
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Set status and convert currency
		smartCheque.Status = status
		smartCheque.Currency = models.Currency(currencyStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}

// GetSmartChequesByContract retrieves smart checks by contract ID
func (r *smartChequeRepository) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	return r.getSmartChequesByEntity(ctx, contractID, "contract_hash", limit, offset)
}

// GetSmartChequesByMilestone retrieves a smart check by milestone ID
func (r *smartChequeRepository) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE milestones @> $1
	`

	// Create JSON search pattern for milestone ID
	milestonePattern := fmt.Sprintf(`[{"id": "%s"}]`, milestoneID)

	var smartCheque models.SmartCheque
	var currencyStr string
	var statusStr string
	var milestonesJSON []byte

	err := r.db.QueryRowContext(ctx, query, milestonePattern).Scan(
		&smartCheque.ID,
		&smartCheque.PayerID,
		&smartCheque.PayeeID,
		&smartCheque.Amount,
		&currencyStr,
		&milestonesJSON,
		&smartCheque.EscrowAddress,
		&statusStr,
		&smartCheque.ContractHash,
		&smartCheque.CreatedAt,
		&smartCheque.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get smart check by milestone: %w", err)
	}

	// Convert string values to typed values
	smartCheque.Currency = models.Currency(currencyStr)
	smartCheque.Status = models.SmartChequeStatus(statusStr)

	// Unmarshal milestones
	if len(milestonesJSON) > 0 {
		if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
			return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
		}
	}

	return &smartCheque, nil
}

// GetSmartChequeCount returns the total count of smart checks
func (r *smartChequeRepository) GetSmartChequeCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM smart_checks`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get smart check count: %w", err)
	}

	return count, nil
}

// GetSmartChequeCountByStatus returns the count of smart checks grouped by status
func (r *smartChequeRepository) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	query := `SELECT status, COUNT(*) FROM smart_checks GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart check counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[models.SmartChequeStatus]int64)
	for rows.Next() {
		var statusStr string
		var count int64

		err := rows.Scan(&statusStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check count: %w", err)
		}

		counts[models.SmartChequeStatus(statusStr)] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return counts, nil
}

// GetSmartChequeCountByCurrency retrieves the count of smart checks grouped by currency
func (r *smartChequeRepository) GetSmartChequeCountByCurrency(ctx context.Context) (map[models.Currency]int64, error) {
	query := `SELECT currency, COUNT(*) FROM smart_checks GROUP BY currency`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart check counts by currency: %w", err)
	}
	defer rows.Close()

	countByCurrency := make(map[models.Currency]int64)
	for rows.Next() {
		var currencyStr string
		var count int64

		err := rows.Scan(&currencyStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check count by currency: %w", err)
		}

		countByCurrency[models.Currency(currencyStr)] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating count by currency rows: %w", err)
	}

	return countByCurrency, nil
}

// GetRecentSmartCheques retrieves the most recent smart checks
func (r *smartChequeRepository) GetRecentSmartCheques(ctx context.Context, limit int) ([]*models.SmartCheque, error) {
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent smart checks: %w", err)
	}
	defer rows.Close()

	var smartCheques []*models.SmartCheque
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var statusStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&statusStr,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Convert string values to typed values
		smartCheque.Currency = models.Currency(currencyStr)
		smartCheque.Status = models.SmartChequeStatus(statusStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}

// BatchDeleteSmartCheques deletes multiple smart checks by their IDs
func (r *smartChequeRepository) BatchDeleteSmartCheques(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Create placeholders for the query
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM smart_checks WHERE id IN (%s)", strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete smart checks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no smart checks found with provided IDs")
	}

	return nil
}

// BatchUpdateSmartCheques updates multiple smart checks
func (r *smartChequeRepository) BatchUpdateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	if len(smartCheques) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Prepare statement for batch update
	query := `
		UPDATE smart_checks 
		SET payer_id = $1, payee_id = $2, amount = $3, currency = $4, 
		    milestones = $5, escrow_address = $6, status = $7, contract_hash = $8, 
		    updated_at = $9
		WHERE id = $10
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch update
	for _, smartCheque := range smartCheques {
		// Convert milestones to JSON
		milestonesJSON, err := json.Marshal(smartCheque.Milestones)
		if err != nil {
			return fmt.Errorf("failed to marshal milestones for smart check %s: %w", smartCheque.ID, err)
		}

		_, err = stmt.ExecContext(
			ctx,
			smartCheque.PayerID,
			smartCheque.PayeeID,
			smartCheque.Amount,
			string(smartCheque.Currency),
			milestonesJSON,
			smartCheque.EscrowAddress,
			string(smartCheque.Status),
			smartCheque.ContractHash,
			smartCheque.UpdatedAt,
			smartCheque.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update smart check %s: %w", smartCheque.ID, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchUpdateSmartChequeStatus updates the status of multiple smart checks
func (r *smartChequeRepository) BatchUpdateSmartChequeStatus(ctx context.Context, ids []string, status models.SmartChequeStatus) error {
	if len(ids) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Prepare statement for batch update
	query := `UPDATE smart_checks SET status = $1, updated_at = $2 WHERE id = $3`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch update
	now := time.Now()
	for _, id := range ids {
		_, err := stmt.ExecContext(ctx, string(status), now, id)
		if err != nil {
			return fmt.Errorf("failed to update smart check %s: %w", id, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchGetSmartCheques retrieves multiple smart checks by their IDs
func (r *smartChequeRepository) BatchGetSmartCheques(ctx context.Context, ids []string) ([]*models.SmartCheque, error) {
	if len(ids) == 0 {
		return []*models.SmartCheque{}, nil
	}

	// Create placeholders for the query
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE id IN (%s)
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart checks: %w", err)
	}
	defer rows.Close()

	var smartCheques []*models.SmartCheque
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var statusStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&statusStr,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Convert string values to typed values
		smartCheque.Currency = models.Currency(currencyStr)
		smartCheque.Status = models.SmartChequeStatus(statusStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}

// BatchUpdateSmartChequeStatuses updates the status of multiple smart checks with different statuses
func (r *smartChequeRepository) BatchUpdateSmartChequeStatuses(ctx context.Context, updates map[string]models.SmartChequeStatus) error {
	if len(updates) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Prepare statement for batch update
	query := `UPDATE smart_checks SET status = $1, updated_at = $2 WHERE id = $3`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute batch update
	now := time.Now()
	for id, status := range updates {
		_, err := stmt.ExecContext(ctx, string(status), now, id)
		if err != nil {
			return fmt.Errorf("failed to update smart check %s: %w", id, err)
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSmartChequeAmountStatistics retrieves statistics about smart check amounts
func (r *smartChequeRepository) GetSmartChequeAmountStatistics(ctx context.Context) (totalAmount, averageAmount, largestAmount, smallestAmount float64, err error) {
	query := `
		SELECT 
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(AVG(amount), 0) as average_amount,
			COALESCE(MAX(amount), 0) as largest_amount,
			COALESCE(MIN(amount), 0) as smallest_amount
		FROM smart_checks
	`

	err = r.db.QueryRowContext(ctx, query).Scan(&totalAmount, &averageAmount, &largestAmount, &smallestAmount)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get smart check amount statistics: %w", err)
	}

	return totalAmount, averageAmount, largestAmount, smallestAmount, nil
}

// GetSmartChequeAnalyticsByPayer retrieves analytics for smart checks by payer
func (r *smartChequeRepository) GetSmartChequeAnalyticsByPayer(ctx context.Context, payerID string) (*SmartChequeAnalytics, error) {
	return r.getSmartChequeAnalytics(ctx, payerID, "payer", r.GetSmartChequesByPayer)
}

// getSmartChequeAnalytics retrieves analytics for smart checks by entity (payer/payee)
func (r *smartChequeRepository) getSmartChequeAnalytics(ctx context.Context, entityID string, entityType string, getRecentActivity func(context.Context, string, int, int) ([]*models.SmartCheque, error)) (*SmartChequeAnalytics, error) {
	if entityID == "" {
		return nil, fmt.Errorf("%s ID is required", entityType)
	}

	entityColumn := entityType + "_id"

	// Get count by status for this entity
	countByStatusQuery := fmt.Sprintf(`SELECT status, COUNT(*) FROM smart_checks WHERE %s = $1 GROUP BY status`, entityColumn)
	countByStatusRows, err := r.db.QueryContext(ctx, countByStatusQuery, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart check counts by status: %w", err)
	}
	defer countByStatusRows.Close()

	countByStatus := make(map[models.SmartChequeStatus]int64)
	for countByStatusRows.Next() {
		var statusStr string
		var count int64

		err := countByStatusRows.Scan(&statusStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check count by status: %w", err)
		}

		countByStatus[models.SmartChequeStatus(statusStr)] = count
	}

	if err = countByStatusRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating count by status rows: %w", err)
	}

	// Get count by currency for this entity
	countByCurrencyQuery := fmt.Sprintf(`SELECT currency, COUNT(*) FROM smart_checks WHERE %s = $1 GROUP BY currency`, entityColumn)
	countByCurrencyRows, err := r.db.QueryContext(ctx, countByCurrencyQuery, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart check counts by currency: %w", err)
	}
	defer countByCurrencyRows.Close()

	countByCurrency := make(map[models.Currency]int64)
	for countByCurrencyRows.Next() {
		var currencyStr string
		var count int64

		err := countByCurrencyRows.Scan(&currencyStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check count by currency: %w", err)
		}

		countByCurrency[models.Currency(currencyStr)] = count
	}

	if err = countByCurrencyRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating count by currency rows: %w", err)
	}

	// Get amount statistics for this entity
	amountStatsQuery := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(AVG(amount), 0) as average_amount,
			COALESCE(MAX(amount), 0) as largest_amount,
			COALESCE(MIN(amount), 0) as smallest_amount
		FROM smart_checks
		WHERE %s = $1
	`, entityColumn)

	var totalAmount, averageAmount, largestAmount, smallestAmount float64
	err = r.db.QueryRowContext(ctx, amountStatsQuery, entityID).Scan(&totalAmount, &averageAmount, &largestAmount, &smallestAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check amount statistics: %w", err)
	}

	// Get recent activity for this entity (last 10 smart checks)
	recentActivity, err := getRecentActivity(ctx, entityID, 10, 0)
	if err != nil {
		// If we can't get recent activity, continue with empty list
		recentActivity = []*models.SmartCheque{}
	}

	// Get status trends for this entity (last 30 days)
	trendsQuery := fmt.Sprintf(`
		SELECT
			DATE(created_at) as creation_date,
			COUNT(*) as count
		FROM smart_checks
		WHERE %s = $1 AND created_at >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY creation_date
	`, entityColumn)

	trendsRows, err := r.db.QueryContext(ctx, trendsQuery, entityID)
	if err != nil {
		// If we can't get trends, continue with empty map
		statusTrends := make(map[string]int64)
		analytics := &SmartChequeAnalytics{
			TotalCount:      0,
			CountByStatus:   countByStatus,
			CountByCurrency: countByCurrency,
			AverageAmount:   averageAmount,
			TotalAmount:     totalAmount,
			LargestAmount:   largestAmount,
			SmallestAmount:  smallestAmount,
			RecentActivity:  recentActivity,
			StatusTrends:    statusTrends,
		}

		// Calculate total count
		for _, count := range countByStatus {
			analytics.TotalCount += count
		}

		return analytics, nil
	}
	defer trendsRows.Close()

	statusTrends := make(map[string]int64)
	for trendsRows.Next() {
		var dateStr string
		var count int64

		err := trendsRows.Scan(&dateStr, &count)
		if err != nil {
			// If we can't scan trends, continue with empty map
			statusTrends := make(map[string]int64)
			analytics := &SmartChequeAnalytics{
				TotalCount:      0,
				CountByStatus:   countByStatus,
				CountByCurrency: countByCurrency,
				AverageAmount:   averageAmount,
				TotalAmount:     totalAmount,
				LargestAmount:   largestAmount,
				SmallestAmount:  smallestAmount,
				RecentActivity:  recentActivity,
				StatusTrends:    statusTrends,
			}

			// Calculate total count
			for _, count := range countByStatus {
				analytics.TotalCount += count
			}

			return analytics, nil
		}

		statusTrends[dateStr] = count
	}

	if err = trendsRows.Err(); err != nil {
		// If we have an error iterating trends, continue with empty map
		statusTrends := make(map[string]int64)
		analytics := &SmartChequeAnalytics{
			TotalCount:      0,
			CountByStatus:   countByStatus,
			CountByCurrency: countByCurrency,
			AverageAmount:   averageAmount,
			TotalAmount:     totalAmount,
			LargestAmount:   largestAmount,
			SmallestAmount:  smallestAmount,
			RecentActivity:  recentActivity,
			StatusTrends:    statusTrends,
		}

		// Calculate total count
		for _, count := range countByStatus {
			analytics.TotalCount += count
		}

		return analytics, nil
	}

	analytics := &SmartChequeAnalytics{
		TotalCount:      0,
		CountByStatus:   countByStatus,
		CountByCurrency: countByCurrency,
		AverageAmount:   averageAmount,
		TotalAmount:     totalAmount,
		LargestAmount:   largestAmount,
		SmallestAmount:  smallestAmount,
		RecentActivity:  recentActivity,
		StatusTrends:    statusTrends,
	}

	// Calculate total count
	for _, count := range countByStatus {
		analytics.TotalCount += count
	}

	return analytics, nil
}

// GetSmartChequeAnalyticsByPayee retrieves analytics for smart checks by payee
func (r *smartChequeRepository) GetSmartChequeAnalyticsByPayee(ctx context.Context, payeeID string) (*SmartChequeAnalytics, error) {
	return r.getSmartChequeAnalytics(ctx, payeeID, "payee", r.GetSmartChequesByPayee)
}

// GetSmartChequeAuditTrail retrieves the audit trail for a specific smart check
func (r *smartChequeRepository) GetSmartChequeAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]models.AuditLog, error) {
	if smartChequeID == "" {
		return nil, fmt.Errorf("smart check ID is required")
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, user_id, enterprise_id, action, resource, resource_id, details, ip_address, user_agent, success, created_at
		FROM audit_logs 
		WHERE resource = 'smart_cheque' AND resource_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, smartChequeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []models.AuditLog
	for rows.Next() {
		var auditLog models.AuditLog
		var userID *uuid.UUID
		var enterpriseID *uuid.UUID
		var resourceID *string

		err := rows.Scan(
			&auditLog.ID,
			&userID,
			&enterpriseID,
			&auditLog.Action,
			&auditLog.Resource,
			&resourceID,
			&auditLog.Details,
			&auditLog.IPAddress,
			&auditLog.UserAgent,
			&auditLog.Success,
			&auditLog.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Set pointers if not null
		if userID != nil && *userID != uuid.Nil {
			auditLog.UserID = *userID
		}

		if enterpriseID != nil && *enterpriseID != uuid.Nil {
			auditLog.EnterpriseID = enterpriseID
		}

		if resourceID != nil {
			auditLog.ResourceID = resourceID
		}

		auditLogs = append(auditLogs, auditLog)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return auditLogs, nil
}

// GetSmartChequeComplianceReport generates a compliance report for a smart check
func (r *smartChequeRepository) GetSmartChequeComplianceReport(ctx context.Context, smartChequeID string) (*SmartChequeComplianceReport, error) {
	if smartChequeID == "" {
		return nil, fmt.Errorf("smart check ID is required")
	}

	// Get transaction count for this smart check
	txQuery := `
		SELECT COUNT(*) as total_transactions,
		       COUNT(CASE WHEN status = 'completed' THEN 1 END) as compliant_tx_count,
		       COUNT(CASE WHEN status IN ('failed', 'cancelled', 'rejected') THEN 1 END) as non_compliant_tx_count
		FROM transactions 
		WHERE smart_check_id = $1
	`

	var totalTx, compliantTx, nonCompliantTx int64
	err := r.db.QueryRowContext(ctx, txQuery, smartChequeID).Scan(&totalTx, &compliantTx, &nonCompliantTx)
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction counts: %w", err)
	}

	// Calculate compliance rate
	complianceRate := 1.0
	if totalTx > 0 {
		complianceRate = float64(compliantTx) / float64(totalTx)
	}

	// Get the latest audit log date for this smart check
	auditQuery := `
		SELECT MAX(created_at) as last_audit_date
		FROM audit_logs 
		WHERE resource = 'smart_cheque' AND resource_id = $1
	`

	var lastAuditDate *time.Time
	err = r.db.QueryRowContext(ctx, auditQuery, smartChequeID).Scan(&lastAuditDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query last audit date: %w", err)
	}

	// If no audit date found, use current time
	if lastAuditDate == nil {
		now := time.Now()
		lastAuditDate = &now
	}

	// Get any audit findings (this would typically come from a separate findings table)
	// For now, we'll create a simple placeholder
	auditFindings := []SmartChequeAuditFinding{}

	// Determine regulatory status based on compliance rate
	regulatoryStatus := "compliant"
	if complianceRate < 0.8 {
		regulatoryStatus = "non_compliant"
	} else if complianceRate < 0.95 {
		regulatoryStatus = "partially_compliant"
	}

	report := &SmartChequeComplianceReport{
		SmartChequeID:       smartChequeID,
		TotalTransactions:   totalTx,
		CompliantTxCount:    compliantTx,
		NonCompliantTxCount: nonCompliantTx,
		ComplianceRate:      complianceRate,
		LastAuditDate:       *lastAuditDate,
		AuditFindings:       auditFindings,
		RegulatoryStatus:    regulatoryStatus,
	}

	return report, nil
}

// GetSmartChequePerformanceMetrics retrieves performance metrics for smart checks
func (r *smartChequeRepository) GetSmartChequePerformanceMetrics(ctx context.Context, filters *SmartChequeFilter) (*SmartChequePerformanceMetrics, error) {
	// Build query with optional filters
	query := `
		SELECT 
			COALESCE(AVG(EXTRACT(EPOCH FROM (t.updated_at - t.created_at))), 0) as avg_processing_time,
			COALESCE(SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) * 1.0 / COUNT(*), 1) as success_rate,
			COALESCE(SUM(CASE WHEN t.status IN ('failed', 'cancelled', 'rejected') THEN 1 ELSE 0 END) * 1.0 / COUNT(*), 0) as failure_rate,
			COALESCE(AVG(s.amount), 0) as average_amount,
			COALESCE(SUM(s.amount), 0) as total_volume,
			COALESCE(MAX(s.amount), 0) as peak_hour_volume
		FROM smart_checks s
		LEFT JOIN transactions t ON s.id = t.smart_check_id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Add filters if provided
	if filters != nil {
		if filters.PayerID != nil {
			query += fmt.Sprintf(" AND s.payer_id = $%d", argIndex)
			args = append(args, *filters.PayerID)
			argIndex++
		}

		if filters.PayeeID != nil {
			query += fmt.Sprintf(" AND s.payee_id = $%d", argIndex)
			args = append(args, *filters.PayeeID)
			argIndex++
		}

		if filters.Status != nil {
			query += fmt.Sprintf(" AND s.status = $%d", argIndex)
			args = append(args, string(*filters.Status))
			argIndex++
		}

		if filters.Currency != nil {
			query += fmt.Sprintf(" AND s.currency = $%d", argIndex)
			args = append(args, string(*filters.Currency))
			argIndex++
		}

		if filters.DateFrom != nil {
			query += fmt.Sprintf(" AND s.created_at >= $%d", argIndex)
			args = append(args, *filters.DateFrom)
			argIndex++
		}

		if filters.DateTo != nil {
			query += fmt.Sprintf(" AND s.created_at <= $%d", argIndex)
			args = append(args, *filters.DateTo)
			argIndex++
		}

		if filters.MinAmount != nil {
			query += fmt.Sprintf(" AND s.amount >= $%d", argIndex)
			args = append(args, *filters.MinAmount)
			argIndex++
		}

		if filters.MaxAmount != nil {
			query += fmt.Sprintf(" AND s.amount <= $%d", argIndex)
			args = append(args, *filters.MaxAmount)
			argIndex++
		}

		if filters.ContractHash != nil {
			query += fmt.Sprintf(" AND s.contract_hash = $%d", argIndex)
			args = append(args, *filters.ContractHash)
			argIndex++
		}
	}

	var avgProcessingTime, successRate, failureRate, averageAmount, totalVolume, peakHourVolume float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&avgProcessingTime,
		&successRate,
		&failureRate,
		&averageAmount,
		&totalVolume,
		&peakHourVolume,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check performance metrics: %w", err)
	}

	metrics := &SmartChequePerformanceMetrics{
		AverageProcessingTime: time.Duration(avgProcessingTime * float64(time.Second)),
		SuccessRate:           successRate,
		FailureRate:           failureRate,
		AverageAmount:         averageAmount,
		TotalVolume:           totalVolume,
		PeakHourVolume:        peakHourVolume,
	}

	return metrics, nil
}

// GetSmartChequeTrends retrieves trends for smart checks over a number of days
func (r *smartChequeRepository) GetSmartChequeTrends(ctx context.Context, days int) (map[string]int64, error) {
	if days <= 0 {
		days = 30
	}

	if days > 365 {
		days = 365
	}

	query := `
		SELECT 
			DATE(created_at) as creation_date,
			COUNT(*) as count
		FROM smart_checks
		WHERE created_at >= CURRENT_DATE - INTERVAL '` + fmt.Sprintf("%d", days) + ` days'
		GROUP BY DATE(created_at)
		ORDER BY creation_date
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart check trends: %w", err)
	}
	defer rows.Close()

	trends := make(map[string]int64)
	for rows.Next() {
		var dateStr string
		var count int64

		err := rows.Scan(&dateStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check trend: %w", err)
		}

		trends[dateStr] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trend rows: %w", err)
	}

	return trends, nil
}

// SearchSmartCheques searches for smart checks based on a query string
func (r *smartChequeRepository) SearchSmartCheques(ctx context.Context, query string, limit, offset int) ([]*models.SmartCheque, error) {
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// Build search query - search in payer_id, payee_id, contract_hash, and id fields
	searchQuery := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_checks 
		WHERE id ILIKE $1 OR payer_id ILIKE $1 OR payee_id ILIKE $1 OR contract_hash ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	// Prepare search term with wildcards
	searchTerm := "%" + query + "%"

	rows, err := r.db.QueryContext(ctx, searchQuery, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search smart checks: %w", err)
	}
	defer rows.Close()

	var smartCheques []*models.SmartCheque
	for rows.Next() {
		var smartCheque models.SmartCheque
		var currencyStr string
		var statusStr string
		var milestonesJSON []byte

		err := rows.Scan(
			&smartCheque.ID,
			&smartCheque.PayerID,
			&smartCheque.PayeeID,
			&smartCheque.Amount,
			&currencyStr,
			&milestonesJSON,
			&smartCheque.EscrowAddress,
			&statusStr,
			&smartCheque.ContractHash,
			&smartCheque.CreatedAt,
			&smartCheque.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart check: %w", err)
		}

		// Convert string values to typed values
		smartCheque.Currency = models.Currency(currencyStr)
		smartCheque.Status = models.SmartChequeStatus(statusStr)

		// Unmarshal milestones
		if len(milestonesJSON) > 0 {
			if err := json.Unmarshal(milestonesJSON, &smartCheque.Milestones); err != nil {
				return nil, fmt.Errorf("failed to unmarshal milestones: %w", err)
			}
		}

		smartCheques = append(smartCheques, &smartCheque)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return smartCheques, nil
}
