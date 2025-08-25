package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// PostgresAssetRepository implements AssetRepository interface for PostgreSQL
type PostgresAssetRepository struct {
	db *sql.DB
}

// NewPostgresAssetRepository creates a new instance of PostgresAssetRepository
func NewPostgresAssetRepository(db *sql.DB) AssetRepository {
	return &PostgresAssetRepository{db: db}
}

// CreateAsset creates a new supported asset
func (r *PostgresAssetRepository) CreateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	query := `
		INSERT INTO supported_assets (
			id, currency_code, currency_name, asset_type, issuer_address, currency_hex,
			decimal_places, minimum_amount, maximum_amount, is_active, trust_line_limit,
			transfer_fee, global_freeze, no_freeze, description, icon_url, documentation_url,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	_, err := r.db.ExecContext(ctx, query,
		asset.ID, asset.CurrencyCode, asset.CurrencyName, asset.AssetType,
		asset.IssuerAddress, asset.CurrencyHex, asset.DecimalPlaces, asset.MinimumAmount,
		asset.MaximumAmount, asset.IsActive, asset.TrustLineLimit, asset.TransferFee,
		asset.GlobalFreeze, asset.NoFreeze, asset.Description, asset.IconURL,
		asset.DocumentationURL, asset.CreatedAt, asset.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	return nil
}

// GetAssetByID retrieves an asset by its ID
func (r *PostgresAssetRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*models.SupportedAsset, error) {
	query := `
		SELECT id, currency_code, currency_name, asset_type, issuer_address, currency_hex,
			   decimal_places, minimum_amount, maximum_amount, is_active, trust_line_limit,
			   transfer_fee, global_freeze, no_freeze, description, icon_url, documentation_url,
			   created_at, updated_at
		FROM supported_assets
		WHERE id = $1`

	asset := &models.SupportedAsset{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&asset.ID, &asset.CurrencyCode, &asset.CurrencyName, &asset.AssetType,
		&asset.IssuerAddress, &asset.CurrencyHex, &asset.DecimalPlaces, &asset.MinimumAmount,
		&asset.MaximumAmount, &asset.IsActive, &asset.TrustLineLimit, &asset.TransferFee,
		&asset.GlobalFreeze, &asset.NoFreeze, &asset.Description, &asset.IconURL,
		&asset.DocumentationURL, &asset.CreatedAt, &asset.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get asset by ID: %w", err)
	}

	return asset, nil
}

// GetAssetByCurrency retrieves an asset by its currency code
func (r *PostgresAssetRepository) GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error) {
	query := `
		SELECT id, currency_code, currency_name, asset_type, issuer_address, currency_hex,
			   decimal_places, minimum_amount, maximum_amount, is_active, trust_line_limit,
			   transfer_fee, global_freeze, no_freeze, description, icon_url, documentation_url,
			   created_at, updated_at
		FROM supported_assets
		WHERE currency_code = $1`

	asset := &models.SupportedAsset{}
	err := r.db.QueryRowContext(ctx, query, currencyCode).Scan(
		&asset.ID, &asset.CurrencyCode, &asset.CurrencyName, &asset.AssetType,
		&asset.IssuerAddress, &asset.CurrencyHex, &asset.DecimalPlaces, &asset.MinimumAmount,
		&asset.MaximumAmount, &asset.IsActive, &asset.TrustLineLimit, &asset.TransferFee,
		&asset.GlobalFreeze, &asset.NoFreeze, &asset.Description, &asset.IconURL,
		&asset.DocumentationURL, &asset.CreatedAt, &asset.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("asset with currency code %s not found", currencyCode)
		}
		return nil, fmt.Errorf("failed to get asset by currency code: %w", err)
	}

	return asset, nil
}

// UpdateAsset updates an existing asset
func (r *PostgresAssetRepository) UpdateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	query := `
		UPDATE supported_assets SET
			currency_name = $2, asset_type = $3, issuer_address = $4, currency_hex = $5,
			decimal_places = $6, minimum_amount = $7, maximum_amount = $8, is_active = $9,
			trust_line_limit = $10, transfer_fee = $11, global_freeze = $12, no_freeze = $13,
			description = $14, icon_url = $15, documentation_url = $16, updated_at = $17
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		asset.ID, asset.CurrencyName, asset.AssetType, asset.IssuerAddress,
		asset.CurrencyHex, asset.DecimalPlaces, asset.MinimumAmount, asset.MaximumAmount,
		asset.IsActive, asset.TrustLineLimit, asset.TransferFee, asset.GlobalFreeze,
		asset.NoFreeze, asset.Description, asset.IconURL, asset.DocumentationURL,
		asset.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset with ID %s not found", asset.ID)
	}

	return nil
}

// DeleteAsset deletes an asset by its ID
func (r *PostgresAssetRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM supported_assets WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset with ID %s not found", id)
	}

	return nil
}

// GetAssets retrieves all assets, optionally filtering for active only
func (r *PostgresAssetRepository) GetAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error) {
	query := `
		SELECT id, currency_code, currency_name, asset_type, issuer_address, currency_hex,
			   decimal_places, minimum_amount, maximum_amount, is_active, trust_line_limit,
			   transfer_fee, global_freeze, no_freeze, description, icon_url, documentation_url,
			   created_at, updated_at
		FROM supported_assets`

	args := []interface{}{}
	if activeOnly {
		query += " WHERE is_active = $1"
		args = append(args, true)
	}

	query += " ORDER BY currency_code"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %w", err)
	}
	defer rows.Close()

	var assets []*models.SupportedAsset
	for rows.Next() {
		asset := &models.SupportedAsset{}
		err := rows.Scan(
			&asset.ID, &asset.CurrencyCode, &asset.CurrencyName, &asset.AssetType,
			&asset.IssuerAddress, &asset.CurrencyHex, &asset.DecimalPlaces, &asset.MinimumAmount,
			&asset.MaximumAmount, &asset.IsActive, &asset.TrustLineLimit, &asset.TransferFee,
			&asset.GlobalFreeze, &asset.NoFreeze, &asset.Description, &asset.IconURL,
			&asset.DocumentationURL, &asset.CreatedAt, &asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assets: %w", err)
	}

	return assets, nil
}

// GetAssetsByType retrieves assets by asset type
func (r *PostgresAssetRepository) GetAssetsByType(ctx context.Context, assetType models.AssetType) ([]*models.SupportedAsset, error) {
	query := `
		SELECT id, currency_code, currency_name, asset_type, issuer_address, currency_hex,
			   decimal_places, minimum_amount, maximum_amount, is_active, trust_line_limit,
			   transfer_fee, global_freeze, no_freeze, description, icon_url, documentation_url,
			   created_at, updated_at
		FROM supported_assets
		WHERE asset_type = $1 AND is_active = true
		ORDER BY currency_code`

	rows, err := r.db.QueryContext(ctx, query, assetType)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets by type: %w", err)
	}
	defer rows.Close()

	var assets []*models.SupportedAsset
	for rows.Next() {
		asset := &models.SupportedAsset{}
		err := rows.Scan(
			&asset.ID, &asset.CurrencyCode, &asset.CurrencyName, &asset.AssetType,
			&asset.IssuerAddress, &asset.CurrencyHex, &asset.DecimalPlaces, &asset.MinimumAmount,
			&asset.MaximumAmount, &asset.IsActive, &asset.TrustLineLimit, &asset.TransferFee,
			&asset.GlobalFreeze, &asset.NoFreeze, &asset.Description, &asset.IconURL,
			&asset.DocumentationURL, &asset.CreatedAt, &asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assets: %w", err)
	}

	return assets, nil
}

// GetAssetCount returns the total number of assets
func (r *PostgresAssetRepository) GetAssetCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM supported_assets`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get asset count: %w", err)
	}

	return count, nil
}

// GetActiveAssetCount returns the number of active assets
func (r *PostgresAssetRepository) GetActiveAssetCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM supported_assets WHERE is_active = true`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active asset count: %w", err)
	}

	return count, nil
}
