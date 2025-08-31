package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
)

// PostgresBalanceRepository implements BalanceRepository interface for PostgreSQL
type PostgresBalanceRepository struct {
	db *sql.DB
}

// NewPostgresBalanceRepository creates a new instance of PostgresBalanceRepository
func NewPostgresBalanceRepository(db *sql.DB) BalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

// CreateEnterpriseBalance creates a new enterprise balance record
func (r *PostgresBalanceRepository) CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	query := `
		INSERT INTO enterprise_balances (
			id, enterprise_id, currency_code, available_balance, reserved_balance, total_balance,
			xrpl_balance, last_xrpl_sync, daily_limit, monthly_limit, max_transaction_amount,
			is_frozen, freeze_reason, last_transaction_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err := r.db.ExecContext(ctx, query,
		balance.ID, balance.EnterpriseID, balance.CurrencyCode, balance.AvailableBalance,
		balance.ReservedBalance, balance.TotalBalance, balance.XRPLBalance, balance.LastXRPLSync,
		balance.DailyLimit, balance.MonthlyLimit, balance.MaxTransactionAmount, balance.IsFrozen,
		balance.FreezeReason, balance.LastTransactionAt, balance.CreatedAt, balance.UpdatedAt,
	)

	return err
}

// GetEnterpriseBalance retrieves a specific enterprise balance
func (r *PostgresBalanceRepository) GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	query := `
		SELECT id, enterprise_id, currency_code, available_balance, reserved_balance, total_balance,
			   xrpl_balance, last_xrpl_sync, daily_limit, monthly_limit, max_transaction_amount,
			   is_frozen, freeze_reason, last_transaction_at, created_at, updated_at
		FROM enterprise_balances
		WHERE enterprise_id = $1 AND currency_code = $2`

	balance := &models.EnterpriseBalance{}
	err := r.db.QueryRowContext(ctx, query, enterpriseID, currencyCode).Scan(
		&balance.ID, &balance.EnterpriseID, &balance.CurrencyCode, &balance.AvailableBalance,
		&balance.ReservedBalance, &balance.TotalBalance, &balance.XRPLBalance, &balance.LastXRPLSync,
		&balance.DailyLimit, &balance.MonthlyLimit, &balance.MaxTransactionAmount, &balance.IsFrozen,
		&balance.FreezeReason, &balance.LastTransactionAt, &balance.CreatedAt, &balance.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("balance not found")
		}
		return nil, err
	}

	return balance, nil
}

// GetEnterpriseBalances retrieves all balances for an enterprise
func (r *PostgresBalanceRepository) GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error) {
	query := `
		SELECT id, enterprise_id, currency_code, available_balance, reserved_balance, total_balance,
			   xrpl_balance, last_xrpl_sync, daily_limit, monthly_limit, max_transaction_amount,
			   is_frozen, freeze_reason, last_transaction_at, created_at, updated_at
		FROM enterprise_balances
		WHERE enterprise_id = $1
		ORDER BY currency_code`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []*models.EnterpriseBalance
	for rows.Next() {
		balance := &models.EnterpriseBalance{}
		err := rows.Scan(
			&balance.ID, &balance.EnterpriseID, &balance.CurrencyCode, &balance.AvailableBalance,
			&balance.ReservedBalance, &balance.TotalBalance, &balance.XRPLBalance, &balance.LastXRPLSync,
			&balance.DailyLimit, &balance.MonthlyLimit, &balance.MaxTransactionAmount, &balance.IsFrozen,
			&balance.FreezeReason, &balance.LastTransactionAt, &balance.CreatedAt, &balance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}

	return balances, rows.Err()
}

// UpdateEnterpriseBalance updates an existing enterprise balance
func (r *PostgresBalanceRepository) UpdateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	query := `
		UPDATE enterprise_balances SET
			available_balance = $3, reserved_balance = $4, total_balance = $5, xrpl_balance = $6,
			last_xrpl_sync = $7, daily_limit = $8, monthly_limit = $9, max_transaction_amount = $10,
			is_frozen = $11, freeze_reason = $12, last_transaction_at = $13, updated_at = $14
		WHERE enterprise_id = $1 AND currency_code = $2`

	result, err := r.db.ExecContext(ctx, query,
		balance.EnterpriseID, balance.CurrencyCode, balance.AvailableBalance, balance.ReservedBalance,
		balance.TotalBalance, balance.XRPLBalance, balance.LastXRPLSync, balance.DailyLimit,
		balance.MonthlyLimit, balance.MaxTransactionAmount, balance.IsFrozen, balance.FreezeReason,
		balance.LastTransactionAt, balance.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("balance not found")
	}

	return nil
}

// GetEnterpriseBalanceSummary retrieves balance summary for an enterprise
func (r *PostgresBalanceRepository) GetEnterpriseBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error) {
	query := `
		SELECT 
			enterprise_id, enterprise_name, currency_code, currency_name,
			available_balance, reserved_balance, total_balance, xrpl_balance,
			is_frozen, last_transaction_at, last_xrpl_sync
		FROM enterprise_balance_summary
		WHERE enterprise_id = $1
		ORDER BY currency_code`

	rows, err := r.db.QueryContext(ctx, query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.EnterpriseBalanceSummary
	for rows.Next() {
		summary := &models.EnterpriseBalanceSummary{}
		err := rows.Scan(
			&summary.EnterpriseID, &summary.EnterpriseName, &summary.CurrencyCode, &summary.CurrencyName,
			&summary.AvailableBalance, &summary.ReservedBalance, &summary.TotalBalance, &summary.XRPLBalance,
			&summary.IsFrozen, &summary.LastTransactionAt, &summary.LastXRPLSync,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// GetAllBalanceSummaries retrieves balance summaries for all enterprises
func (r *PostgresBalanceRepository) GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error) {
	query := `
		SELECT 
			enterprise_id, enterprise_name, currency_code, currency_name,
			available_balance, reserved_balance, total_balance, xrpl_balance,
			is_frozen, last_transaction_at, last_xrpl_sync
		FROM enterprise_balance_summary
		ORDER BY enterprise_name, currency_code`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.EnterpriseBalanceSummary
	for rows.Next() {
		summary := &models.EnterpriseBalanceSummary{}
		err := rows.Scan(
			&summary.EnterpriseID, &summary.EnterpriseName, &summary.CurrencyCode, &summary.CurrencyName,
			&summary.AvailableBalance, &summary.ReservedBalance, &summary.TotalBalance, &summary.XRPLBalance,
			&summary.IsFrozen, &summary.LastTransactionAt, &summary.LastXRPLSync,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// IsAssetInUse checks if an asset is currently being used
func (r *PostgresBalanceRepository) IsAssetInUse(ctx context.Context, currencyCode string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM enterprise_balances 
			WHERE currency_code = $1 AND (
				available_balance != '0' OR 
				reserved_balance != '0' OR 
				total_balance != '0'
			)
		)`

	var inUse bool
	err := r.db.QueryRowContext(ctx, query, currencyCode).Scan(&inUse)
	return inUse, err
}

// CreateAssetTransaction creates a new asset transaction record
func (r *PostgresBalanceRepository) CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	metadataJSON, _ := json.Marshal(transaction.Metadata)

	query := `
		INSERT INTO asset_transactions (
			id, enterprise_id, currency_code, transaction_type, amount, fee, reference_id,
			external_tx_hash, balance_before, balance_after, status, description, metadata,
			approved_by, approved_at, created_at, updated_at, processed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)`

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.EnterpriseID, transaction.CurrencyCode, transaction.TransactionType,
		transaction.Amount, transaction.Fee, transaction.ReferenceID, transaction.ExternalTxHash,
		transaction.BalanceBefore, transaction.BalanceAfter, transaction.Status, transaction.Description,
		metadataJSON, transaction.ApprovedBy, transaction.ApprovedAt, transaction.CreatedAt,
		transaction.UpdatedAt, transaction.ProcessedAt,
	)

	return err
}

// GetAssetTransaction retrieves an asset transaction by ID
func (r *PostgresBalanceRepository) GetAssetTransaction(ctx context.Context, id uuid.UUID) (*models.AssetTransaction, error) {
	query := `
		SELECT id, enterprise_id, currency_code, transaction_type, amount, fee, reference_id,
			   external_tx_hash, balance_before, balance_after, status, description, metadata,
			   approved_by, approved_at, created_at, updated_at, processed_at
		FROM asset_transactions
		WHERE id = $1`

	transaction := &models.AssetTransaction{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.EnterpriseID, &transaction.CurrencyCode, &transaction.TransactionType,
		&transaction.Amount, &transaction.Fee, &transaction.ReferenceID, &transaction.ExternalTxHash,
		&transaction.BalanceBefore, &transaction.BalanceAfter, &transaction.Status, &transaction.Description,
		&metadataJSON, &transaction.ApprovedBy, &transaction.ApprovedAt, &transaction.CreatedAt,
		&transaction.UpdatedAt, &transaction.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset transaction not found")
		}
		return nil, err
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &transaction.Metadata); err != nil {
			log.Printf("Error unmarshaling metadata: %v", err)
		}
	}

	return transaction, nil
}

// GetAssetTransactionsByEnterprise retrieves asset transactions for an enterprise
func (r *PostgresBalanceRepository) GetAssetTransactionsByEnterprise(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*models.AssetTransaction, error) {
	// Implementation similar to GetAssetTransaction but with WHERE clause for enterprise_id
	// and LIMIT/OFFSET for pagination - simplified for brevity
	return nil, fmt.Errorf("not implemented")
}

// GetAssetTransactionsByCurrency retrieves asset transactions for a currency
func (r *PostgresBalanceRepository) GetAssetTransactionsByCurrency(_ context.Context, _ string, _ int, _ int) ([]*models.AssetTransaction, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetAssetTransactionsByType retrieves asset transactions by type
func (r *PostgresBalanceRepository) GetAssetTransactionsByType(_ context.Context, _ models.AssetTransactionType, _ int, _ int) ([]*models.AssetTransaction, error) {
	return nil, fmt.Errorf("not implemented")
}

// UpdateAssetTransaction updates an asset transaction
func (r *PostgresBalanceRepository) UpdateAssetTransaction(_ context.Context, _ *models.AssetTransaction) error {
	return fmt.Errorf("not implemented")
}

// UpdateBalance updates enterprise balance and creates a transaction record
func (r *PostgresBalanceRepository) UpdateBalance(_ context.Context, _ uuid.UUID, _ string, _ string, _ models.AssetTransactionType, _ *string) error {
	// This would implement the complex balance update logic with transactions
	// Simplified for now
	return fmt.Errorf("not implemented")
}

// FreezeBalance freezes a balance for an enterprise
func (r *PostgresBalanceRepository) FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error {
	query := `
		UPDATE enterprise_balances SET
			is_frozen = true, freeze_reason = $3, updated_at = $4
		WHERE enterprise_id = $1 AND currency_code = $2`

	_, err := r.db.ExecContext(ctx, query, enterpriseID, currencyCode, reason, time.Now())
	return err
}

// UnfreezeBalance unfreezes a balance for an enterprise
func (r *PostgresBalanceRepository) UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error {
	query := `
		UPDATE enterprise_balances SET
			is_frozen = false, freeze_reason = NULL, updated_at = $3
		WHERE enterprise_id = $1 AND currency_code = $2`

	_, err := r.db.ExecContext(ctx, query, enterpriseID, currencyCode, time.Now())
	return err
}
