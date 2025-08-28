package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smart-payment-infrastructure/internal/models"
)

// ContractDocumentRepository provides focused persistence for contract document fields.
// It does not change existing ContractRepository SQL to keep backward compatibility with tests.
type ContractDocumentRepository struct {
	db *sql.DB
}

func NewContractDocumentRepository(db *sql.DB) *ContractDocumentRepository {
	return &ContractDocumentRepository{db: db}
}

// UpdateDocumentFields upserts document-related columns for a contract.
// Pass nil pointers to set NULL in the database.
func (r *ContractDocumentRepository) UpdateDocumentFields(
	ctx context.Context,
	contractID string,
	originalFilename *string,
	fileSize *int64,
	mimeType *string,
) error {
	query := `
		UPDATE contracts SET
			original_filename = $2,
			file_size = $3,
			mime_type = $4
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, contractID, originalFilename, fileSize, mimeType)
	if err != nil {
		return fmt.Errorf("failed to update document fields: %w", err)
	}
	return nil
}

// GetDocumentFields fetches document metadata columns for a contract.
func (r *ContractDocumentRepository) GetDocumentFields(ctx context.Context, contractID string) (*models.DocumentMetadata, error) {
	query := `
		SELECT original_filename, file_size, mime_type
		FROM contracts
		WHERE id = $1`

	var (
		fn sql.NullString
		sz sql.NullInt64
		mt sql.NullString
	)
	if err := r.db.QueryRowContext(ctx, query, contractID).Scan(&fn, &sz, &mt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract with ID %s not found", contractID)
		}
		return nil, fmt.Errorf("failed to get document fields: %w", err)
	}
	meta := &models.DocumentMetadata{}
	if fn.Valid {
		meta.OriginalFilename = fn.String
	}
	if sz.Valid {
		meta.FileSize = sz.Int64
	}
	if mt.Valid {
		meta.MimeType = mt.String
	}
	return meta, nil
}
