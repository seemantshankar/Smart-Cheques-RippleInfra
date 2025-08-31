package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// SmartChequeServiceInterface defines the interface for smart cheque service operations
type SmartChequeServiceInterface interface {
	// CreateSmartCheque creates a new smart cheque
	CreateSmartCheque(ctx context.Context, request *CreateSmartChequeRequest) (*models.SmartCheque, error)

	// GetSmartCheque retrieves a smart cheque by ID
	GetSmartCheque(ctx context.Context, id string) (*models.SmartCheque, error)

	// UpdateSmartCheque updates an existing smart cheque
	UpdateSmartCheque(ctx context.Context, id string, request *UpdateSmartChequeRequest) (*models.SmartCheque, error)

	// DeleteSmartCheque deletes a smart cheque
	DeleteSmartCheque(ctx context.Context, id string) error

	// ListSmartChequesByPayer lists smart cheques by payer ID
	ListSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error)

	// ListSmartChequesByPayee lists smart cheques by payee ID
	ListSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error)

	// ListSmartChequesByStatus lists smart cheques by status
	ListSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error)

	// UpdateSmartChequeStatus updates the status of a smart cheque
	UpdateSmartChequeStatus(ctx context.Context, id string, status models.SmartChequeStatus) error

	// GetSmartChequeStatistics returns statistics about smart cheques
	GetSmartChequeStatistics(ctx context.Context) (*SmartChequeStatistics, error)
}

// CreateSmartChequeRequest represents the request to create a smart cheque
type CreateSmartChequeRequest struct {
	PayerID      string              `json:"payer_id" binding:"required"`
	PayeeID      string              `json:"payee_id" binding:"required"`
	Amount       float64             `json:"amount" binding:"required,gt=0"`
	Currency     models.Currency     `json:"currency" binding:"required"`
	Milestones   []models.Milestone  `json:"milestones"`
	ContractHash string              `json:"contract_hash"`
}

// UpdateSmartChequeRequest represents the request to update a smart cheque
type UpdateSmartChequeRequest struct {
	PayerID      *string             `json:"payer_id,omitempty"`
	PayeeID      *string             `json:"payee_id,omitempty"`
	Amount       *float64            `json:"amount,omitempty"`
	Currency     *models.Currency    `json:"currency,omitempty"`
	Milestones   *[]models.Milestone `json:"milestones,omitempty"`
	EscrowAddress *string            `json:"escrow_address,omitempty"`
	Status       *models.SmartChequeStatus `json:"status,omitempty"`
	ContractHash *string             `json:"contract_hash,omitempty"`
}

// SmartChequeStatistics represents statistics about smart cheques
type SmartChequeStatistics struct {
	TotalCount     int64                                `json:"total_count"`
	CountByStatus  map[models.SmartChequeStatus]int64   `json:"count_by_status"`
}

// smartChequeService implements SmartChequeServiceInterface
type smartChequeService struct {
	smartChequeRepo repository.SmartChequeRepositoryInterface
}

// NewSmartChequeService creates a new smart cheque service
func NewSmartChequeService(smartChequeRepo repository.SmartChequeRepositoryInterface) SmartChequeServiceInterface {
	return &smartChequeService{
		smartChequeRepo: smartChequeRepo,
	}
}

// CreateSmartCheque creates a new smart cheque
func (s *smartChequeService) CreateSmartCheque(ctx context.Context, request *CreateSmartChequeRequest) (*models.SmartCheque, error) {
	// Validate request
	if err := s.validateCreateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create smart cheque model
	smartCheque := &models.SmartCheque{
		ID:            uuid.New().String(),
		PayerID:       request.PayerID,
		PayeeID:       request.PayeeID,
		Amount:        request.Amount,
		Currency:      request.Currency,
		Milestones:    request.Milestones,
		EscrowAddress: "", // Will be set when escrow is created
		Status:        models.SmartChequeStatusCreated,
		ContractHash:  request.ContractHash,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to repository
	if err := s.smartChequeRepo.CreateSmartCheque(ctx, smartCheque); err != nil {
		return nil, fmt.Errorf("failed to create smart cheque: %w", err)
	}

	return smartCheque, nil
}

// validateCreateRequest validates the create smart cheque request
func (s *smartChequeService) validateCreateRequest(request *CreateSmartChequeRequest) error {
	if request.PayerID == "" {
		return fmt.Errorf("payer_id is required")
	}

	if request.PayeeID == "" {
		return fmt.Errorf("payee_id is required")
	}

	if request.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Validate currency
	switch request.Currency {
	case models.CurrencyUSDT, models.CurrencyUSDC, models.CurrencyERupee:
		// Valid currency
	default:
		return fmt.Errorf("invalid currency: %s", request.Currency)
	}

	// Validate milestones
	for i, milestone := range request.Milestones {
		if milestone.ID == "" {
			return fmt.Errorf("milestone %d: id is required", i)
		}

		if milestone.Description == "" {
			return fmt.Errorf("milestone %d: description is required", i)
		}

		if milestone.Amount <= 0 {
			return fmt.Errorf("milestone %d: amount must be greater than 0", i)
		}

		// Validate verification method
		switch milestone.VerificationMethod {
		case models.VerificationMethodOracle, models.VerificationMethodManual, models.VerificationMethodHybrid:
			// Valid verification method
		default:
			return fmt.Errorf("milestone %d: invalid verification method: %s", i, milestone.VerificationMethod)
		}

		// Validate status
		switch milestone.Status {
		case models.MilestoneStatusPending, models.MilestoneStatusVerified, models.MilestoneStatusFailed:
			// Valid status
		default:
			return fmt.Errorf("milestone %d: invalid status: %s", i, milestone.Status)
		}
	}

	return nil
}

// GetSmartCheque retrieves a smart cheque by ID
func (s *smartChequeService) GetSmartCheque(ctx context.Context, id string) (*models.SmartCheque, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
	}

	if smartCheque == nil {
		return nil, fmt.Errorf("smart cheque not found: %s", id)
	}

	return smartCheque, nil
}

// UpdateSmartCheque updates an existing smart cheque
func (s *smartChequeService) UpdateSmartCheque(ctx context.Context, id string, request *UpdateSmartChequeRequest) (*models.SmartCheque, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	// Get existing smart cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
	}

	if smartCheque == nil {
		return nil, fmt.Errorf("smart cheque not found: %s", id)
	}

	// Update fields if provided
	if request.PayerID != nil {
		smartCheque.PayerID = *request.PayerID
	}

	if request.PayeeID != nil {
		smartCheque.PayeeID = *request.PayeeID
	}

	if request.Amount != nil {
		smartCheque.Amount = *request.Amount
	}

	if request.Currency != nil {
		smartCheque.Currency = *request.Currency
	}

	if request.Milestones != nil {
		smartCheque.Milestones = *request.Milestones
	}

	if request.EscrowAddress != nil {
		smartCheque.EscrowAddress = *request.EscrowAddress
	}

	if request.Status != nil {
		smartCheque.Status = *request.Status
	}

	if request.ContractHash != nil {
		smartCheque.ContractHash = *request.ContractHash
	}

	// Update timestamp
	smartCheque.UpdatedAt = time.Now()

	// Save to repository
	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return nil, fmt.Errorf("failed to update smart cheque: %w", err)
	}

	return smartCheque, nil
}

// DeleteSmartCheque deletes a smart cheque
func (s *smartChequeService) DeleteSmartCheque(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// Check if smart cheque exists
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}

	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", id)
	}

	// Delete from repository
	if err := s.smartChequeRepo.DeleteSmartCheque(ctx, id); err != nil {
		return fmt.Errorf("failed to delete smart cheque: %w", err)
	}

	return nil
}

// ListSmartChequesByPayer lists smart cheques by payer ID
func (s *smartChequeService) ListSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	if payerID == "" {
		return nil, fmt.Errorf("payer_id is required")
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	smartCheques, err := s.smartChequeRepo.GetSmartChequesByPayer(ctx, payerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list smart cheques by payer: %w", err)
	}

	return smartCheques, nil
}

// ListSmartChequesByPayee lists smart cheques by payee ID
func (s *smartChequeService) ListSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	if payeeID == "" {
		return nil, fmt.Errorf("payee_id is required")
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	smartCheques, err := s.smartChequeRepo.GetSmartChequesByPayee(ctx, payeeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list smart cheques by payee: %w", err)
	}

	return smartCheques, nil
}

// ListSmartChequesByStatus lists smart cheques by status
func (s *smartChequeService) ListSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	// Validate status
	switch status {
	case models.SmartChequeStatusCreated, models.SmartChequeStatusLocked, 
	     models.SmartChequeStatusInProgress, models.SmartChequeStatusCompleted, 
	     models.SmartChequeStatusDisputed:
		// Valid status
	default:
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	smartCheques, err := s.smartChequeRepo.GetSmartChequesByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list smart cheques by status: %w", err)
	}

	return smartCheques, nil
}

// UpdateSmartChequeStatus updates the status of a smart cheque
func (s *smartChequeService) UpdateSmartChequeStatus(ctx context.Context, id string, status models.SmartChequeStatus) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// Validate status
	switch status {
	case models.SmartChequeStatusCreated, models.SmartChequeStatusLocked, 
	     models.SmartChequeStatusInProgress, models.SmartChequeStatusCompleted, 
	     models.SmartChequeStatusDisputed:
		// Valid status
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// Get existing smart cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}

	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", id)
	}

	// Update status
	smartCheque.Status = status
	smartCheque.UpdatedAt = time.Now()

	// Save to repository
	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque status: %w", err)
	}

	return nil
}

// GetSmartChequeStatistics returns statistics about smart cheques
func (s *smartChequeService) GetSmartChequeStatistics(ctx context.Context) (*SmartChequeStatistics, error) {
	// Get total count
	totalCount, err := s.smartChequeRepo.GetSmartChequeCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque count: %w", err)
	}

	// Get count by status
	countByStatus, err := s.smartChequeRepo.GetSmartChequeCountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque count by status: %w", err)
	}

	statistics := &SmartChequeStatistics{
		TotalCount:    totalCount,
		CountByStatus: countByStatus,
	}

	return statistics, nil
}