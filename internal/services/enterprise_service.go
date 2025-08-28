package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

var (
	ErrEnterpriseAlreadyExists = errors.New("enterprise already exists")
	ErrEnterpriseNotFound      = errors.New("enterprise not found")
	ErrInvalidDocumentType     = errors.New("invalid document type")
	ErrFileTooLarge            = errors.New("file too large")
	ErrInvalidFileType         = errors.New("invalid file type")
)

const (
	MaxFileSize = 10 * 1024 * 1024 // 10MB
	UploadDir   = "uploads/documents"
)

// EnterpriseService handles enterprise business logic
type EnterpriseService struct {
	enterpriseRepo repository.EnterpriseRepositoryInterface
	walletService  *WalletService
}

// NewEnterpriseService creates a new enterprise service
func NewEnterpriseService(enterpriseRepo repository.EnterpriseRepositoryInterface, walletService *WalletService) *EnterpriseService {
	return &EnterpriseService{
		enterpriseRepo: enterpriseRepo,
		walletService:  walletService,
	}
}

// RegisterEnterprise registers a new enterprise
func (s *EnterpriseService) RegisterEnterprise(req *models.EnterpriseRegistrationRequest) (*models.Enterprise, error) {
	// Check if enterprise already exists
	exists, err := s.enterpriseRepo.RegistrationNumberExists(req.RegistrationNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEnterpriseAlreadyExists
	}

	// Create enterprise
	enterprise := &models.Enterprise{
		LegalName:          req.LegalName,
		TradeName:          req.TradeName,
		RegistrationNumber: req.RegistrationNumber,
		TaxID:              req.TaxID,
		Jurisdiction:       req.Jurisdiction,
		BusinessType:       req.BusinessType,
		Industry:           req.Industry,
		Website:            req.Website,
		Email:              req.Email,
		Phone:              req.Phone,
		Address:            req.Address,
		KYBStatus:          models.KYBStatusPending,
		ComplianceStatus:   models.ComplianceStatusPending,
		IsActive:           true,
		AuthorizedRepresentatives: []models.AuthorizedRepresentative{
			{
				FirstName: req.AuthorizedRepresentative.FirstName,
				LastName:  req.AuthorizedRepresentative.LastName,
				Email:     req.AuthorizedRepresentative.Email,
				Phone:     req.AuthorizedRepresentative.Phone,
				Position:  req.AuthorizedRepresentative.Position,
			},
		},
		ComplianceProfile: models.ComplianceProfile{
			AMLRiskScore: 0, // Initial score
		},
	}

	// Save enterprise to database first
	if err := s.enterpriseRepo.CreateEnterprise(enterprise); err != nil {
		return nil, err
	}

	// Create XRPL wallet for the enterprise (testnet by default)
	if s.walletService != nil {
		walletResponse, err := s.walletService.CreateWalletForEnterprise(enterprise.ID, "testnet")
		if err != nil {
			// Log the error but don't fail the enterprise creation
			// The wallet can be created later
			fmt.Printf("Warning: Failed to create wallet for enterprise %s: %v\n", enterprise.ID, err)
		} else {
			// Update enterprise with wallet address
			enterprise.XRPLWallet = walletResponse.Address
			if err := s.enterpriseRepo.UpdateEnterpriseXRPLWallet(enterprise.ID, walletResponse.Address); err != nil {
				fmt.Printf("Warning: Failed to update enterprise with wallet address: %v\n", err)
			}
		}
	}

	return enterprise, nil
}

// GetEnterpriseByID retrieves an enterprise by ID
func (s *EnterpriseService) GetEnterpriseByID(id uuid.UUID) (*models.Enterprise, error) {
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(id)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
		return nil, ErrEnterpriseNotFound
	}

	return enterprise, nil
}

// UpdateKYBStatus updates the KYB status of an enterprise
func (s *EnterpriseService) UpdateKYBStatus(id uuid.UUID, status models.KYBStatus, comments string) error {
	// Verify enterprise exists
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(id)
	if err != nil {
		return err
	}
	if enterprise == nil {
		return ErrEnterpriseNotFound
	}

	// Update KYB status
	if err := s.enterpriseRepo.UpdateEnterpriseKYBStatus(id, status); err != nil {
		return err
	}

	// TODO: Send notification to enterprise about status change
	// TODO: Log audit event for KYB status change

	return nil
}

// UpdateComplianceStatus updates the compliance status of an enterprise
func (s *EnterpriseService) UpdateComplianceStatus(id uuid.UUID, status models.ComplianceStatus) error {
	// Verify enterprise exists
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(id)
	if err != nil {
		return err
	}
	if enterprise == nil {
		return ErrEnterpriseNotFound
	}

	// Update compliance status
	if err := s.enterpriseRepo.UpdateEnterpriseComplianceStatus(id, status); err != nil {
		return err
	}

	// TODO: Send notification to enterprise about status change
	// TODO: Log audit event for compliance status change

	return nil
}

// UploadDocument handles document upload for KYB
func (s *EnterpriseService) UploadDocument(enterpriseID uuid.UUID, docType models.DocumentType, file multipart.File, header *multipart.FileHeader) (*models.EnterpriseDocument, error) {
	// Verify enterprise exists
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(enterpriseID)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
		return nil, ErrEnterpriseNotFound
	}

	// Validate file size
	if header.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Validate file type
	if !isValidFileType(header.Header.Get("Content-Type")) {
		return nil, ErrInvalidFileType
	}

	// Create upload directory if it doesn't exist
	uploadPath := filepath.Join(UploadDir, enterpriseID.String())
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	fileExt := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s_%s%s", docType, uuid.New().String(), fileExt)
	filePath := filepath.Join(uploadPath, fileName)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create document record
	doc := &models.EnterpriseDocument{
		EnterpriseID: enterpriseID,
		DocumentType: docType,
		FileName:     header.Filename,
		FilePath:     filePath,
		FileSize:     header.Size,
		MimeType:     header.Header.Get("Content-Type"),
	}

	if err := s.enterpriseRepo.CreateDocument(doc); err != nil {
		// Clean up file if database operation fails
		os.Remove(filePath)
		return nil, err
	}

	return doc, nil
}

// VerifyDocument updates the verification status of a document
func (s *EnterpriseService) VerifyDocument(docID uuid.UUID, status models.DocumentStatus) error {
	return s.enterpriseRepo.UpdateDocumentStatus(docID, status)
}

// GetEnterpriseDocuments retrieves all documents for an enterprise
func (s *EnterpriseService) GetEnterpriseDocuments(enterpriseID uuid.UUID) ([]models.EnterpriseDocument, error) {
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(enterpriseID)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
		return nil, ErrEnterpriseNotFound
	}

	return enterprise.Documents, nil
}

// PerformKYBCheck performs automated KYB checks
func (s *EnterpriseService) PerformKYBCheck(enterpriseID uuid.UUID) error {
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(enterpriseID)
	if err != nil {
		return err
	}
	if enterprise == nil {
		return ErrEnterpriseNotFound
	}

	// TODO: Implement actual KYB checks
	// - Verify business registration with government databases
	// - Check sanctions lists
	// - Verify tax ID
	// - Check PEP lists
	// - Validate business address

	// For now, we'll just update the status to in_review
	return s.enterpriseRepo.UpdateEnterpriseKYBStatus(enterpriseID, models.KYBStatusInReview)
}

// PerformComplianceCheck performs automated compliance checks
func (s *EnterpriseService) PerformComplianceCheck(enterpriseID uuid.UUID) error {
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(enterpriseID)
	if err != nil {
		return err
	}
	if enterprise == nil {
		return ErrEnterpriseNotFound
	}

	// TODO: Implement actual compliance checks
	// - AML risk assessment
	// - Regulatory compliance verification
	// - Industry-specific compliance checks
	// - Jurisdiction-specific requirements

	// For now, we'll just update the status to under_review
	return s.enterpriseRepo.UpdateEnterpriseComplianceStatus(enterpriseID, models.ComplianceStatusUnderReview)
}

// isValidFileType checks if the file type is allowed for document upload
func isValidFileType(mimeType string) bool {
	allowedTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg",
		"image/png",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	for _, allowed := range allowedTypes {
		if strings.EqualFold(mimeType, allowed) {
			return true
		}
	}

	return false
}
