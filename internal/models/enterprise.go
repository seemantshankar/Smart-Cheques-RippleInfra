package models

import (
	"time"

	"github.com/google/uuid"
)

// Enterprise represents a business entity in the system
type Enterprise struct {
	ID                         uuid.UUID                   `json:"id" db:"id"`
	LegalName                  string                      `json:"legal_name" db:"legal_name"`
	TradeName                  string                      `json:"trade_name,omitempty" db:"trade_name"`
	RegistrationNumber         string                      `json:"registration_number" db:"registration_number"`
	TaxID                      string                      `json:"tax_id" db:"tax_id"`
	Jurisdiction               string                      `json:"jurisdiction" db:"jurisdiction"`
	BusinessType               string                      `json:"business_type" db:"business_type"`
	Industry                   string                      `json:"industry" db:"industry"`
	Website                    string                      `json:"website,omitempty" db:"website"`
	Email                      string                      `json:"email" db:"email"`
	Phone                      string                      `json:"phone" db:"phone"`
	Address                    Address                     `json:"address"`
	KYBStatus                  KYBStatus                   `json:"kyb_status" db:"kyb_status"`
	ComplianceStatus           ComplianceStatus            `json:"compliance_status" db:"compliance_status"`
	XRPLWallet                 string                      `json:"-" db:"xrpl_wallet"` // Hidden from API responses
	AuthorizedRepresentatives  []AuthorizedRepresentative  `json:"authorized_representatives"`
	Documents                  []EnterpriseDocument        `json:"documents"`
	ComplianceProfile          ComplianceProfile           `json:"compliance_profile"`
	IsActive                   bool                        `json:"is_active" db:"is_active"`
	CreatedAt                  time.Time                   `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time                   `json:"updated_at" db:"updated_at"`
	VerifiedAt                 *time.Time                  `json:"verified_at,omitempty" db:"verified_at"`
}

// Address represents a business address
type Address struct {
	Street1    string `json:"street1" db:"street1"`
	Street2    string `json:"street2,omitempty" db:"street2"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	Country    string `json:"country" db:"country"`
}

// KYBStatus represents the KYB verification status
type KYBStatus string

const (
	KYBStatusPending   KYBStatus = "pending"
	KYBStatusInReview  KYBStatus = "in_review"
	KYBStatusVerified  KYBStatus = "verified"
	KYBStatusRejected  KYBStatus = "rejected"
	KYBStatusSuspended KYBStatus = "suspended"
)

// ComplianceStatus represents the compliance verification status
type ComplianceStatus string

const (
	ComplianceStatusPending    ComplianceStatus = "pending"
	ComplianceStatusCompliant  ComplianceStatus = "compliant"
	ComplianceStatusNonCompliant ComplianceStatus = "non_compliant"
	ComplianceStatusUnderReview ComplianceStatus = "under_review"
)

// AuthorizedRepresentative represents an authorized person for the enterprise
type AuthorizedRepresentative struct {
	ID          uuid.UUID `json:"id" db:"id"`
	EnterpriseID uuid.UUID `json:"enterprise_id" db:"enterprise_id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Email       string    `json:"email" db:"email"`
	Phone       string    `json:"phone" db:"phone"`
	Position    string    `json:"position" db:"position"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// EnterpriseDocument represents uploaded documents for KYB
type EnterpriseDocument struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	EnterpriseID uuid.UUID    `json:"enterprise_id" db:"enterprise_id"`
	DocumentType DocumentType `json:"document_type" db:"document_type"`
	FileName     string       `json:"file_name" db:"file_name"`
	FilePath     string       `json:"-" db:"file_path"` // Hidden from API responses
	FileSize     int64        `json:"file_size" db:"file_size"`
	MimeType     string       `json:"mime_type" db:"mime_type"`
	Status       DocumentStatus `json:"status" db:"status"`
	UploadedAt   time.Time    `json:"uploaded_at" db:"uploaded_at"`
	VerifiedAt   *time.Time   `json:"verified_at,omitempty" db:"verified_at"`
}

// DocumentType represents the type of document uploaded
type DocumentType string

const (
	DocumentTypeCertificateOfIncorporation DocumentType = "certificate_of_incorporation"
	DocumentTypeBusinessLicense           DocumentType = "business_license"
	DocumentTypeTaxCertificate           DocumentType = "tax_certificate"
	DocumentTypeBankStatement            DocumentType = "bank_statement"
	DocumentTypeDirectorID              DocumentType = "director_id"
	DocumentTypeProofOfAddress           DocumentType = "proof_of_address"
	DocumentTypeArticlesOfAssociation    DocumentType = "articles_of_association"
	DocumentTypeMemorandumOfAssociation  DocumentType = "memorandum_of_association"
)

// DocumentStatus represents the verification status of a document
type DocumentStatus string

const (
	DocumentStatusPending  DocumentStatus = "pending"
	DocumentStatusVerified DocumentStatus = "verified"
	DocumentStatusRejected DocumentStatus = "rejected"
)

// ComplianceProfile represents compliance-related information
type ComplianceProfile struct {
	AMLRiskScore        int       `json:"aml_risk_score" db:"aml_risk_score"`
	SanctionsCheckDate  *time.Time `json:"sanctions_check_date,omitempty" db:"sanctions_check_date"`
	PEPCheckDate        *time.Time `json:"pep_check_date,omitempty" db:"pep_check_date"`
	ComplianceOfficer   string    `json:"compliance_officer,omitempty" db:"compliance_officer"`
	LastReviewDate      *time.Time `json:"last_review_date,omitempty" db:"last_review_date"`
	NextReviewDate      *time.Time `json:"next_review_date,omitempty" db:"next_review_date"`
	RegulatoryFlags     []string  `json:"regulatory_flags,omitempty"`
}

// EnterpriseRegistrationRequest represents the request payload for enterprise registration
type EnterpriseRegistrationRequest struct {
	LegalName            string                      `json:"legal_name" binding:"required"`
	TradeName            string                      `json:"trade_name,omitempty"`
	RegistrationNumber   string                      `json:"registration_number" binding:"required"`
	TaxID                string                      `json:"tax_id" binding:"required"`
	Jurisdiction         string                      `json:"jurisdiction" binding:"required"`
	BusinessType         string                      `json:"business_type" binding:"required"`
	Industry             string                      `json:"industry" binding:"required"`
	Website              string                      `json:"website,omitempty"`
	Email                string                      `json:"email" binding:"required,email"`
	Phone                string                      `json:"phone" binding:"required"`
	Address              Address                     `json:"address" binding:"required"`
	AuthorizedRepresentative AuthorizedRepresentativeRequest `json:"authorized_representative" binding:"required"`
}

// AuthorizedRepresentativeRequest represents the request for adding an authorized representative
type AuthorizedRepresentativeRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required"`
	Position  string `json:"position" binding:"required"`
}

// EnterpriseUpdateRequest represents the request payload for updating enterprise information
type EnterpriseUpdateRequest struct {
	TradeName   string  `json:"trade_name,omitempty"`
	Website     string  `json:"website,omitempty"`
	Email       string  `json:"email,omitempty" binding:"omitempty,email"`
	Phone       string  `json:"phone,omitempty"`
	Address     *Address `json:"address,omitempty"`
}

// DocumentUploadRequest represents the request for uploading a document
type DocumentUploadRequest struct {
	DocumentType DocumentType `json:"document_type" binding:"required"`
	FileName     string       `json:"file_name" binding:"required"`
	FileSize     int64        `json:"file_size" binding:"required"`
	MimeType     string       `json:"mime_type" binding:"required"`
}

// KYBVerificationRequest represents the request for KYB verification
type KYBVerificationRequest struct {
	EnterpriseID uuid.UUID `json:"enterprise_id" binding:"required"`
	Action       string    `json:"action" binding:"required,oneof=approve reject"`
	Comments     string    `json:"comments,omitempty"`
}