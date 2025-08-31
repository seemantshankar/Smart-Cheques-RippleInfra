package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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
		INSERT INTO smart_cheques (
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
		return fmt.Errorf("failed to create smart cheque: %w", err)
	}

	return nil
}

// GetSmartChequeByID retrieves a smart check by its ID
func (r *smartChequeRepository) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_cheques 
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
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
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
		UPDATE smart_cheques 
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
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("smart cheque not found: %s", smartCheque.ID)
	}

	return nil
}

// DeleteSmartCheque deletes a smart check by its ID
func (r *smartChequeRepository) DeleteSmartCheque(ctx context.Context, id string) error {
	query := `DELETE FROM smart_cheques WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete smart cheque: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("smart cheque not found: %s", id)
	}

	return nil
}

// GetSmartChequesByPayer retrieves smart checks by payer ID
func (r *smartChequeRepository) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_cheques 
		WHERE payer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, payerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart cheques: %w", err)
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
			return nil, fmt.Errorf("failed to scan smart cheque: %w", err)
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
		FROM smart_cheques 
		WHERE payee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, payeeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart cheques: %w", err)
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
			return nil, fmt.Errorf("failed to scan smart cheque: %w", err)
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
		FROM smart_cheques 
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart cheques: %w", err)
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
			return nil, fmt.Errorf("failed to scan smart cheque: %w", err)
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

// GetSmartChequesByContract retrieves smart cheques by contract ID
func (r *smartChequeRepository) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_cheques 
		WHERE contract_hash = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, contractID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart cheques: %w", err)
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
			return nil, fmt.Errorf("failed to scan smart cheque: %w", err)
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

// GetSmartChequesByMilestone retrieves a smart cheque by milestone ID
func (r *smartChequeRepository) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	query := `
		SELECT id, payer_id, payee_id, amount, currency, 
		       milestones, escrow_address, status, contract_hash, 
		       created_at, updated_at
		FROM smart_cheques 
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
		return nil, fmt.Errorf("failed to get smart cheque by milestone: %w", err)
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

// GetSmartChequeCount returns the total count of smart cheques
func (r *smartChequeRepository) GetSmartChequeCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM smart_cheques`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get smart cheque count: %w", err)
	}

	return count, nil
}

// GetSmartChequeCountByStatus returns the count of smart cheques grouped by status
func (r *smartChequeRepository) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	query := `SELECT status, COUNT(*) FROM smart_cheques GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query smart cheque counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[models.SmartChequeStatus]int64)
	for rows.Next() {
		var statusStr string
		var count int64

		err := rows.Scan(&statusStr, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan smart cheque count: %w", err)
		}

		counts[models.SmartChequeStatus(statusStr)] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return counts, nil
}
