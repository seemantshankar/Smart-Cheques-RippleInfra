package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/smart-payment-infrastructure/internal/models"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(wallet *models.Wallet) error {
	query := `
		INSERT INTO wallets (
			id, enterprise_id, address, public_key, encrypted_private_key, 
			encrypted_seed, status, is_whitelisted, network_type, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(
		query,
		wallet.ID,
		wallet.EnterpriseID,
		wallet.Address,
		wallet.PublicKey,
		wallet.EncryptedPrivateKey,
		wallet.EncryptedSeed,
		wallet.Status,
		wallet.IsWhitelisted,
		wallet.NetworkType,
		wallet.Metadata,
	).Scan(&wallet.CreatedAt, &wallet.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "wallets_address_key" {
					return fmt.Errorf("wallet address already exists: %s", wallet.Address)
				}
				if pqErr.Constraint == "idx_wallets_enterprise_network_active" {
					return fmt.Errorf("enterprise already has an active wallet for network: %s", wallet.NetworkType)
				}
			case "23503": // foreign_key_violation
				return fmt.Errorf("enterprise not found: %s", wallet.EnterpriseID)
			}
		}
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

func (r *WalletRepository) GetByID(id uuid.UUID) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT id, enterprise_id, address, public_key, encrypted_private_key,
			   encrypted_seed, status, is_whitelisted, network_type, metadata,
			   last_activity, created_at, updated_at
		FROM wallets WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&wallet.ID,
		&wallet.EnterpriseID,
		&wallet.Address,
		&wallet.PublicKey,
		&wallet.EncryptedPrivateKey,
		&wallet.EncryptedSeed,
		&wallet.Status,
		&wallet.IsWhitelisted,
		&wallet.NetworkType,
		&wallet.Metadata,
		&wallet.LastActivity,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetByAddress(address string) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT id, enterprise_id, address, public_key, encrypted_private_key,
			   encrypted_seed, status, is_whitelisted, network_type, metadata,
			   last_activity, created_at, updated_at
		FROM wallets WHERE address = $1`

	err := r.db.QueryRow(query, address).Scan(
		&wallet.ID,
		&wallet.EnterpriseID,
		&wallet.Address,
		&wallet.PublicKey,
		&wallet.EncryptedPrivateKey,
		&wallet.EncryptedSeed,
		&wallet.Status,
		&wallet.IsWhitelisted,
		&wallet.NetworkType,
		&wallet.Metadata,
		&wallet.LastActivity,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found: %s", address)
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) GetByEnterpriseID(enterpriseID uuid.UUID) ([]*models.Wallet, error) {
	query := `
		SELECT id, enterprise_id, address, public_key, encrypted_private_key,
			   encrypted_seed, status, is_whitelisted, network_type, metadata,
			   last_activity, created_at, updated_at
		FROM wallets WHERE enterprise_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets for enterprise: %w", err)
	}
	defer rows.Close()

	var wallets []*models.Wallet
	for rows.Next() {
		wallet := &models.Wallet{}
		err := rows.Scan(
			&wallet.ID,
			&wallet.EnterpriseID,
			&wallet.Address,
			&wallet.PublicKey,
			&wallet.EncryptedPrivateKey,
			&wallet.EncryptedSeed,
			&wallet.Status,
			&wallet.IsWhitelisted,
			&wallet.NetworkType,
			&wallet.Metadata,
			&wallet.LastActivity,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

func (r *WalletRepository) GetActiveByEnterpriseAndNetwork(enterpriseID uuid.UUID, networkType string) (*models.Wallet, error) {
	wallet := &models.Wallet{}
	query := `
		SELECT id, enterprise_id, address, public_key, encrypted_private_key,
			   encrypted_seed, status, is_whitelisted, network_type, metadata,
			   last_activity, created_at, updated_at
		FROM wallets 
		WHERE enterprise_id = $1 AND network_type = $2 AND status = 'active'`

	err := r.db.QueryRow(query, enterpriseID, networkType).Scan(
		&wallet.ID,
		&wallet.EnterpriseID,
		&wallet.Address,
		&wallet.PublicKey,
		&wallet.EncryptedPrivateKey,
		&wallet.EncryptedSeed,
		&wallet.Status,
		&wallet.IsWhitelisted,
		&wallet.NetworkType,
		&wallet.Metadata,
		&wallet.LastActivity,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active wallet found for enterprise %s on network %s", enterpriseID, networkType)
		}
		return nil, fmt.Errorf("failed to get active wallet: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) Update(wallet *models.Wallet) error {
	query := `
		UPDATE wallets 
		SET status = $2, is_whitelisted = $3, metadata = $4, last_activity = $5
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRow(
		query,
		wallet.ID,
		wallet.Status,
		wallet.IsWhitelisted,
		wallet.Metadata,
		wallet.LastActivity,
	).Scan(&wallet.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("wallet not found: %s", wallet.ID)
		}
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

func (r *WalletRepository) UpdateLastActivity(walletID uuid.UUID) error {
	query := `UPDATE wallets SET last_activity = NOW() WHERE id = $1`
	
	result, err := r.db.Exec(query, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %s", walletID)
	}

	return nil
}

func (r *WalletRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM wallets WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %s", id)
	}

	return nil
}

func (r *WalletRepository) GetWhitelistedWallets() ([]*models.Wallet, error) {
	query := `
		SELECT id, enterprise_id, address, public_key, encrypted_private_key,
			   encrypted_seed, status, is_whitelisted, network_type, metadata,
			   last_activity, created_at, updated_at
		FROM wallets 
		WHERE is_whitelisted = true AND status = 'active'
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelisted wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*models.Wallet
	for rows.Next() {
		wallet := &models.Wallet{}
		err := rows.Scan(
			&wallet.ID,
			&wallet.EnterpriseID,
			&wallet.Address,
			&wallet.PublicKey,
			&wallet.EncryptedPrivateKey,
			&wallet.EncryptedSeed,
			&wallet.Status,
			&wallet.IsWhitelisted,
			&wallet.NetworkType,
			&wallet.Metadata,
			&wallet.LastActivity,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}