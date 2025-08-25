package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// EnterpriseRepository handles database operations for enterprises
type EnterpriseRepository struct {
	db *sql.DB
}

// NewEnterpriseRepository creates a new enterprise repository
func NewEnterpriseRepository(db *sql.DB) *EnterpriseRepository {
	return &EnterpriseRepository{db: db}
}

// CreateEnterprise creates a new enterprise in the database
func (r *EnterpriseRepository) CreateEnterprise(enterprise *models.Enterprise) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert enterprise
	query := `
		INSERT INTO enterprises (
			id, legal_name, trade_name, registration_number, tax_id, jurisdiction,
			business_type, industry, website, email, phone, street1, street2,
			city, state, postal_code, country, kyb_status, compliance_status,
			xrpl_wallet, aml_risk_score, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)
	`

	enterprise.ID = uuid.New()
	enterprise.KYBStatus = models.KYBStatusPending
	enterprise.ComplianceStatus = models.ComplianceStatusPending
	enterprise.IsActive = true
	enterprise.CreatedAt = time.Now()
	enterprise.UpdatedAt = time.Now()

	_, err = tx.Exec(query,
		enterprise.ID,
		enterprise.LegalName,
		enterprise.TradeName,
		enterprise.RegistrationNumber,
		enterprise.TaxID,
		enterprise.Jurisdiction,
		enterprise.BusinessType,
		enterprise.Industry,
		enterprise.Website,
		enterprise.Email,
		enterprise.Phone,
		enterprise.Address.Street1,
		enterprise.Address.Street2,
		enterprise.Address.City,
		enterprise.Address.State,
		enterprise.Address.PostalCode,
		enterprise.Address.Country,
		enterprise.KYBStatus,
		enterprise.ComplianceStatus,
		enterprise.XRPLWallet,
		enterprise.ComplianceProfile.AMLRiskScore,
		enterprise.IsActive,
		enterprise.CreatedAt,
		enterprise.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// Insert authorized representatives
	for _, rep := range enterprise.AuthorizedRepresentatives {
		repQuery := `
			INSERT INTO authorized_representatives (
				id, enterprise_id, first_name, last_name, email, phone, position, is_active, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		repID := uuid.New()
		_, err = tx.Exec(repQuery,
			repID,
			enterprise.ID,
			rep.FirstName,
			rep.LastName,
			rep.Email,
			rep.Phone,
			rep.Position,
			true,
			time.Now(),
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetEnterpriseByID retrieves an enterprise by ID
func (r *EnterpriseRepository) GetEnterpriseByID(id uuid.UUID) (*models.Enterprise, error) {
	query := `
		SELECT 
			id, legal_name, trade_name, registration_number, tax_id, jurisdiction,
			business_type, industry, website, email, phone, street1, street2,
			city, state, postal_code, country, kyb_status, compliance_status,
			xrpl_wallet, aml_risk_score, is_active, created_at, updated_at, verified_at
		FROM enterprises
		WHERE id = $1 AND is_active = true
	`

	enterprise := &models.Enterprise{}
	var verifiedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&enterprise.ID,
		&enterprise.LegalName,
		&enterprise.TradeName,
		&enterprise.RegistrationNumber,
		&enterprise.TaxID,
		&enterprise.Jurisdiction,
		&enterprise.BusinessType,
		&enterprise.Industry,
		&enterprise.Website,
		&enterprise.Email,
		&enterprise.Phone,
		&enterprise.Address.Street1,
		&enterprise.Address.Street2,
		&enterprise.Address.City,
		&enterprise.Address.State,
		&enterprise.Address.PostalCode,
		&enterprise.Address.Country,
		&enterprise.KYBStatus,
		&enterprise.ComplianceStatus,
		&enterprise.XRPLWallet,
		&enterprise.ComplianceProfile.AMLRiskScore,
		&enterprise.IsActive,
		&enterprise.CreatedAt,
		&enterprise.UpdatedAt,
		&verifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if verifiedAt.Valid {
		enterprise.VerifiedAt = &verifiedAt.Time
	}

	// Load authorized representatives
	reps, err := r.getAuthorizedRepresentatives(enterprise.ID)
	if err != nil {
		return nil, err
	}
	enterprise.AuthorizedRepresentatives = reps

	// Load documents
	docs, err := r.getEnterpriseDocuments(enterprise.ID)
	if err != nil {
		return nil, err
	}
	enterprise.Documents = docs

	return enterprise, nil
}

// GetEnterpriseByRegistrationNumber retrieves an enterprise by registration number
func (r *EnterpriseRepository) GetEnterpriseByRegistrationNumber(regNumber string) (*models.Enterprise, error) {
	query := `SELECT id FROM enterprises WHERE registration_number = $1 AND is_active = true`
	
	var id uuid.UUID
	err := r.db.QueryRow(query, regNumber).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.GetEnterpriseByID(id)
}

// UpdateEnterpriseKYBStatus updates the KYB status of an enterprise
func (r *EnterpriseRepository) UpdateEnterpriseKYBStatus(id uuid.UUID, status models.KYBStatus) error {
	query := `
		UPDATE enterprises 
		SET kyb_status = $1, updated_at = $2, verified_at = CASE WHEN $1 = 'verified' THEN $2 ELSE verified_at END
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// UpdateEnterpriseComplianceStatus updates the compliance status of an enterprise
func (r *EnterpriseRepository) UpdateEnterpriseComplianceStatus(id uuid.UUID, status models.ComplianceStatus) error {
	query := `UPDATE enterprises SET compliance_status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// UpdateEnterpriseXRPLWallet updates the XRPL wallet address for an enterprise
func (r *EnterpriseRepository) UpdateEnterpriseXRPLWallet(id uuid.UUID, walletAddress string) error {
	query := `UPDATE enterprises SET xrpl_wallet = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, walletAddress, time.Now(), id)
	return err
}

// RegistrationNumberExists checks if a registration number already exists
func (r *EnterpriseRepository) RegistrationNumberExists(regNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM enterprises WHERE registration_number = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, regNumber).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CreateDocument creates a new document record
func (r *EnterpriseRepository) CreateDocument(doc *models.EnterpriseDocument) error {
	query := `
		INSERT INTO enterprise_documents (
			id, enterprise_id, document_type, file_name, file_path, file_size, mime_type, status, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	doc.ID = uuid.New()
	doc.Status = models.DocumentStatusPending
	doc.UploadedAt = time.Now()

	_, err := r.db.Exec(query,
		doc.ID,
		doc.EnterpriseID,
		doc.DocumentType,
		doc.FileName,
		doc.FilePath,
		doc.FileSize,
		doc.MimeType,
		doc.Status,
		doc.UploadedAt,
	)

	return err
}

// UpdateDocumentStatus updates the status of a document
func (r *EnterpriseRepository) UpdateDocumentStatus(docID uuid.UUID, status models.DocumentStatus) error {
	query := `
		UPDATE enterprise_documents 
		SET status = $1, verified_at = CASE WHEN $1 = 'verified' THEN $2 ELSE verified_at END
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), docID)
	return err
}

// getAuthorizedRepresentatives retrieves authorized representatives for an enterprise
func (r *EnterpriseRepository) getAuthorizedRepresentatives(enterpriseID uuid.UUID) ([]models.AuthorizedRepresentative, error) {
	query := `
		SELECT id, enterprise_id, first_name, last_name, email, phone, position, is_active, created_at
		FROM authorized_representatives
		WHERE enterprise_id = $1 AND is_active = true
		ORDER BY created_at
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var representatives []models.AuthorizedRepresentative
	for rows.Next() {
		var rep models.AuthorizedRepresentative
		err := rows.Scan(
			&rep.ID,
			&rep.EnterpriseID,
			&rep.FirstName,
			&rep.LastName,
			&rep.Email,
			&rep.Phone,
			&rep.Position,
			&rep.IsActive,
			&rep.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		representatives = append(representatives, rep)
	}

	return representatives, nil
}

// getEnterpriseDocuments retrieves documents for an enterprise
func (r *EnterpriseRepository) getEnterpriseDocuments(enterpriseID uuid.UUID) ([]models.EnterpriseDocument, error) {
	query := `
		SELECT id, enterprise_id, document_type, file_name, file_path, file_size, mime_type, status, uploaded_at, verified_at
		FROM enterprise_documents
		WHERE enterprise_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(query, enterpriseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []models.EnterpriseDocument
	for rows.Next() {
		var doc models.EnterpriseDocument
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&doc.ID,
			&doc.EnterpriseID,
			&doc.DocumentType,
			&doc.FileName,
			&doc.FilePath,
			&doc.FileSize,
			&doc.MimeType,
			&doc.Status,
			&doc.UploadedAt,
			&verifiedAt,
		)
		if err != nil {
			return nil, err
		}

		if verifiedAt.Valid {
			doc.VerifiedAt = &verifiedAt.Time
		}

		documents = append(documents, doc)
	}

	return documents, nil
}