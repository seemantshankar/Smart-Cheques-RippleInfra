package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// SmartChequeServiceInterface defines the interface for smart check service operations
type SmartChequeServiceInterface interface {
	// CreateSmartCheque creates a new smart check
	CreateSmartCheque(ctx context.Context, request *CreateSmartChequeRequest) (*models.SmartCheque, error)

	// GetSmartCheque retrieves a smart check by ID
	GetSmartCheque(ctx context.Context, id string) (*models.SmartCheque, error)

	// UpdateSmartCheque updates an existing smart check
	UpdateSmartCheque(ctx context.Context, id string, request *UpdateSmartChequeRequest) (*models.SmartCheque, error)

	// DeleteSmartCheque deletes a smart check
	DeleteSmartCheque(ctx context.Context, id string) error

	// ListSmartChequesByPayer lists smart checks by payer ID
	ListSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error)

	// ListSmartChequesByPayee lists smart checks by payee ID
	ListSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error)

	// ListSmartChequesByStatus lists smart checks by status
	ListSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error)

	// UpdateSmartChequeStatus updates the status of a smart check
	UpdateSmartChequeStatus(ctx context.Context, id string, status models.SmartChequeStatus) error

	// GetSmartChequeStatistics returns statistics about smart checks
	GetSmartChequeStatistics(ctx context.Context) (*SmartChequeStatistics, error)

	// GetSmartChequeAnalytics returns detailed analytics about smart checks
	GetSmartChequeAnalytics(ctx context.Context) (*SmartChequeAnalytics, error)

	// CreateSmartChequeBatch creates multiple smart checks in a batch operation
	CreateSmartChequeBatch(ctx context.Context, requests []*CreateSmartChequeRequest) (*SmartChequeBatchOperationResult, error)

	// UpdateSmartChequeBatch updates multiple smart checks in a batch operation
	UpdateSmartChequeBatch(ctx context.Context, updates map[string]*UpdateSmartChequeRequest) (*SmartChequeBatchOperationResult, error)

	// CreateAuditLog creates an audit log entry for smart check operations
	CreateAuditLog(ctx context.Context, entry *AuditLogEntry) error

	// GetAuditTrail retrieves the audit trail for a specific smart check
	GetAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]AuditLogEntry, error)
}

// CreateSmartChequeRequest represents the request to create a smart check
type CreateSmartChequeRequest struct {
	PayerID      string             `json:"payer_id" binding:"required"`
	PayeeID      string             `json:"payee_id" binding:"required"`
	Amount       float64            `json:"amount" binding:"required,gt=0"`
	Currency     models.Currency    `json:"currency" binding:"required"`
	Milestones   []models.Milestone `json:"milestones"`
	ContractHash string             `json:"contract_hash"`
}

// UpdateSmartChequeRequest represents the request to update a smart check
type UpdateSmartChequeRequest struct {
	PayerID       *string                   `json:"payer_id,omitempty"`
	PayeeID       *string                   `json:"payee_id,omitempty"`
	Amount        *float64                  `json:"amount,omitempty"`
	Currency      *models.Currency          `json:"currency,omitempty"`
	Milestones    *[]models.Milestone       `json:"milestones,omitempty"`
	EscrowAddress *string                   `json:"escrow_address,omitempty"`
	Status        *models.SmartChequeStatus `json:"status,omitempty"`
	ContractHash  *string                   `json:"contract_hash,omitempty"`
}

// SmartChequeStatistics represents statistics about smart checks
type SmartChequeStatistics struct {
	TotalCount    int64                              `json:"total_count"`
	CountByStatus map[models.SmartChequeStatus]int64 `json:"count_by_status"`
}

// SmartChequeAnalytics represents detailed analytics for smart checks
type SmartChequeAnalytics struct {
	TotalCount           int64                              `json:"total_count"`
	CountByStatus        map[models.SmartChequeStatus]int64 `json:"count_by_status"`
	CountByCurrency      map[models.Currency]int64          `json:"count_by_currency"`
	AverageAmount        float64                            `json:"average_amount"`
	TotalAmount          float64                            `json:"total_amount"`
	LargestAmount        float64                            `json:"largest_amount"`
	SmallestAmount       float64                            `json:"smallest_amount"`
	RecentActivity       []*models.SmartCheque              `json:"recent_activity"`
	StatusTrends         map[string]int64                   `json:"status_trends"`
	CurrencyDistribution map[models.Currency]float64        `json:"currency_distribution"`
}

// SmartChequeBatchOperationResult represents the result of a batch operation
type SmartChequeBatchOperationResult struct {
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Results      []BatchOperationResult `json:"results"`
}

// BatchOperationResult represents the result of a single operation in a batch
type BatchOperationResult struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AuditLogEntry represents an audit log entry for smart check operations
type AuditLogEntry struct {
	ID            string    `json:"id"`
	SmartChequeID string    `json:"smart_check_id"`
	Action        string    `json:"action"`
	UserID        string    `json:"user_id,omitempty"`
	Details       string    `json:"details,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	IPAddress     string    `json:"ip_address,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
}

// smartChequeService implements SmartChequeServiceInterface
type smartChequeService struct {
	smartChequeRepo repository.SmartChequeRepositoryInterface
	auditRepo       repository.AuditRepositoryInterface // Add audit repository
}

// NewSmartChequeService creates a new smart check service
func NewSmartChequeService(smartChequeRepo repository.SmartChequeRepositoryInterface, auditRepo repository.AuditRepositoryInterface) SmartChequeServiceInterface {
	return &smartChequeService{
		smartChequeRepo: smartChequeRepo,
		auditRepo:       auditRepo,
	}
}

// CreateSmartCheque creates a new smart check
func (s *smartChequeService) CreateSmartCheque(ctx context.Context, request *CreateSmartChequeRequest) (*models.SmartCheque, error) {
	// Validate request
	if err := s.validateCreateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create smart check model
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
		return nil, fmt.Errorf("failed to create smart check: %w", err)
	}

	return smartCheque, nil
}

// validateCreateRequest validates the create smart check request
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

	// Validate currency with asset service
	switch request.Currency {
	case models.CurrencyUSDT, models.CurrencyUSDC, models.CurrencyERupee:
		// Valid currency
	default:
		return fmt.Errorf("invalid currency: %s", request.Currency)
	}

	// Validate contract hash if provided (basic validation)
	if request.ContractHash != "" {
		// Allow both UUID format and other formats for backward compatibility
		// Just ensure it's not too long
		if len(request.ContractHash) > 100 {
			return fmt.Errorf("contract hash is too long")
		}
	}

	// Validate milestones
	var totalMilestoneAmount float64
	for i, milestone := range request.Milestones {
		if err := s.validateMilestone(milestone, i); err != nil {
			return err
		}
		totalMilestoneAmount += milestone.Amount
	}

	// Validate that milestone amounts sum up to total smart check amount
	// Only validate if we have milestones
	if len(request.Milestones) > 0 && totalMilestoneAmount != request.Amount {
		return fmt.Errorf("sum of milestone amounts (%f) must equal smart check amount (%f)", totalMilestoneAmount, request.Amount)
	}

	return nil
}

// validateMilestone validates a single milestone
func (s *smartChequeService) validateMilestone(milestone models.Milestone, index int) error {
	if milestone.ID == "" {
		return fmt.Errorf("milestone %d: id is required", index)
	}

	if milestone.Description == "" {
		return fmt.Errorf("milestone %d: description is required", index)
	}

	if milestone.Amount <= 0 {
		return fmt.Errorf("milestone %d: amount must be greater than 0", index)
	}

	// Validate verification method
	switch milestone.VerificationMethod {
	case models.VerificationMethodOracle, models.VerificationMethodManual, models.VerificationMethodHybrid:
		// Valid verification method
	default:
		return fmt.Errorf("milestone %d: invalid verification method: %s", index, milestone.VerificationMethod)
	}

	// Validate status
	switch milestone.Status {
	case models.MilestoneStatusPending, models.MilestoneStatusVerified, models.MilestoneStatusFailed:
		// Valid status
	default:
		return fmt.Errorf("milestone %d: invalid status: %s", index, milestone.Status)
	}

	// Validate sequence numbers if provided
	if milestone.SequenceNumber < 0 {
		return fmt.Errorf("milestone %d: sequence number cannot be negative", index)
	}

	// Validate priority if provided
	if milestone.Priority < 0 {
		return fmt.Errorf("milestone %d: priority cannot be negative", index)
	}

	// Validate percentage complete if provided
	if milestone.PercentageComplete < 0 || milestone.PercentageComplete > 100 {
		return fmt.Errorf("milestone %d: percentage complete must be between 0 and 100", index)
	}

	// Validate criticality score if provided
	if milestone.CriticalityScore < 0 || milestone.CriticalityScore > 100 {
		return fmt.Errorf("milestone %d: criticality score must be between 0 and 100", index)
	}

	// Validate dates if provided
	if milestone.EstimatedStartDate != nil && milestone.EstimatedEndDate != nil {
		if milestone.EstimatedStartDate.After(*milestone.EstimatedEndDate) {
			return fmt.Errorf("milestone %d: estimated start date cannot be after estimated end date", index)
		}
	}

	if milestone.ActualStartDate != nil && milestone.ActualEndDate != nil {
		if milestone.ActualStartDate.After(*milestone.ActualEndDate) {
			return fmt.Errorf("milestone %d: actual start date cannot be after actual end date", index)
		}
	}

	// Validate risk level if provided
	if milestone.RiskLevel != "" {
		validRiskLevels := []string{"low", "medium", "high", "critical"}
		isValid := false
		for _, level := range validRiskLevels {
			if milestone.RiskLevel == level {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("milestone %d: invalid risk level: %s", index, milestone.RiskLevel)
		}
	}

	// Validate category if provided
	if milestone.Category != "" {
		validCategories := []string{"delivery", "payment", "approval", "compliance", "other"}
		isValid := false
		for _, category := range validCategories {
			if milestone.Category == category {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("milestone %d: invalid category: %s", index, milestone.Category)
		}
	}

	return nil
}

// validateUpdateRequest validates the update smart check request
func (s *smartChequeService) validateUpdateRequest(id string, request *UpdateSmartChequeRequest) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// Validate currency if provided
	if request.Currency != nil {
		switch *request.Currency {
		case models.CurrencyUSDT, models.CurrencyUSDC, models.CurrencyERupee:
			// Valid currency
		default:
			return fmt.Errorf("invalid currency: %s", *request.Currency)
		}
	}

	// Validate status transition if provided
	if request.Status != nil {
		// Get current smart check to validate status transition
		current, err := s.smartChequeRepo.GetSmartChequeByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf("failed to get current smart check for validation: %w", err)
		}
		if current == nil {
			return fmt.Errorf("smart check not found: %s", id)
		}

		if err := s.validateStatusTransition(current.Status, *request.Status); err != nil {
			return err
		}
	}

	// Validate milestones if provided
	if request.Milestones != nil {
		for i, milestone := range *request.Milestones {
			if err := s.validateMilestone(milestone, i); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateStatusTransition validates that a status transition is allowed
func (s *smartChequeService) validateStatusTransition(from, to models.SmartChequeStatus) error {
	// Define valid status transitions
	validTransitions := map[models.SmartChequeStatus][]models.SmartChequeStatus{
		models.SmartChequeStatusCreated: {
			models.SmartChequeStatusLocked,
			models.SmartChequeStatusInProgress,
			models.SmartChequeStatusDisputed,
		},
		models.SmartChequeStatusLocked: {
			models.SmartChequeStatusInProgress,
			models.SmartChequeStatusDisputed,
		},
		models.SmartChequeStatusInProgress: {
			models.SmartChequeStatusCompleted,
			models.SmartChequeStatusDisputed,
		},
		models.SmartChequeStatusCompleted: {
			models.SmartChequeStatusDisputed,
		},
		models.SmartChequeStatusDisputed: {
			models.SmartChequeStatusInProgress,
			models.SmartChequeStatusCompleted,
		},
	}

	// Check if transition is valid
	if allowedTransitions, exists := validTransitions[from]; exists {
		for _, allowed := range allowedTransitions {
			if to == allowed {
				return nil
			}
		}
		return fmt.Errorf("invalid status transition from %s to %s", from, to)
	}

	// If from status is not in the map, allow any transition (for new statuses)
	return nil
}

// GetSmartCheque retrieves a smart check by ID
func (s *smartChequeService) GetSmartCheque(ctx context.Context, id string) (*models.SmartCheque, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check: %w", err)
	}

	if smartCheque == nil {
		return nil, fmt.Errorf("smart check not found: %s", id)
	}

	return smartCheque, nil
}

// UpdateSmartCheque updates an existing smart check
func (s *smartChequeService) UpdateSmartCheque(ctx context.Context, id string, request *UpdateSmartChequeRequest) (*models.SmartCheque, error) {
	// Validate request
	if err := s.validateUpdateRequest(id, request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing smart check
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check: %w", err)
	}

	if smartCheque == nil {
		return nil, fmt.Errorf("smart check not found: %s", id)
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
		return nil, fmt.Errorf("failed to update smart check: %w", err)
	}

	return smartCheque, nil
}

// DeleteSmartCheque deletes a smart check
func (s *smartChequeService) DeleteSmartCheque(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	// Check if smart check exists
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get smart check: %w", err)
	}

	if smartCheque == nil {
		return fmt.Errorf("smart check not found: %s", id)
	}

	// Delete from repository
	if err := s.smartChequeRepo.DeleteSmartCheque(ctx, id); err != nil {
		return fmt.Errorf("failed to delete smart check: %w", err)
	}

	return nil
}

// ListSmartChequesByPayer lists smart checks by payer ID
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
		return nil, fmt.Errorf("failed to list smart checks by payer: %w", err)
	}

	return smartCheques, nil
}

// ListSmartChequesByPayee lists smart checks by payee ID
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
		return nil, fmt.Errorf("failed to list smart checks by payee: %w", err)
	}

	return smartCheques, nil
}

// ListSmartChequesByStatus lists smart checks by status
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
		return nil, fmt.Errorf("failed to list smart checks by status: %w", err)
	}

	return smartCheques, nil
}

// UpdateSmartChequeStatus updates the status of a smart check
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

	// Get existing smart check
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get smart check: %w", err)
	}

	if smartCheque == nil {
		return fmt.Errorf("smart check not found: %s", id)
	}

	// Update status
	smartCheque.Status = status
	smartCheque.UpdatedAt = time.Now()

	// Save to repository
	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart check status: %w", err)
	}

	return nil
}

// GetSmartChequeStatistics returns statistics about smart checks
func (s *smartChequeService) GetSmartChequeStatistics(ctx context.Context) (*SmartChequeStatistics, error) {
	// Get total count
	totalCount, err := s.smartChequeRepo.GetSmartChequeCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check count: %w", err)
	}

	// Get count by status
	countByStatus, err := s.smartChequeRepo.GetSmartChequeCountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check count by status: %w", err)
	}

	statistics := &SmartChequeStatistics{
		TotalCount:    totalCount,
		CountByStatus: countByStatus,
	}

	return statistics, nil
}

// GetSmartChequeAnalytics returns detailed analytics about smart checks
func (s *smartChequeService) GetSmartChequeAnalytics(ctx context.Context) (*SmartChequeAnalytics, error) {
	// Get count by status
	countByStatus, err := s.smartChequeRepo.GetSmartChequeCountByStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check count by status: %w", err)
	}

	// Get count by currency
	countByCurrency, err := s.smartChequeRepo.GetSmartChequeCountByCurrency(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check count by currency: %w", err)
	}

	// Get amount statistics
	totalAmount, averageAmount, largestAmount, smallestAmount, err := s.smartChequeRepo.GetSmartChequeAmountStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check amount statistics: %w", err)
	}

	// Get recent activity (last 10 smart checks)
	recentActivity, err := s.smartChequeRepo.GetRecentSmartCheques(ctx, 10)
	if err != nil {
		// If we can't get recent activity, continue with empty list
		recentActivity = []*models.SmartCheque{}
	}

	// Get status trends (last 30 days)
	statusTrends, err := s.smartChequeRepo.GetSmartChequeTrends(ctx, 30)
	if err != nil {
		// If we can't get trends, continue with empty map
		statusTrends = make(map[string]int64)
	}

	// Calculate currency distribution
	currencyDistribution := make(map[models.Currency]float64)
	if totalAmount > 0 {
		for currency, count := range countByCurrency {
			// This is a simplified calculation - in a real implementation, we would
			// calculate the actual distribution by amount
			currencyDistribution[currency] = float64(count) / float64(len(countByCurrency))
		}
	}

	analytics := &SmartChequeAnalytics{
		TotalCount:           0, // Will be calculated below
		CountByStatus:        countByStatus,
		CountByCurrency:      countByCurrency,
		AverageAmount:        averageAmount,
		TotalAmount:          totalAmount,
		LargestAmount:        largestAmount,
		SmallestAmount:       smallestAmount,
		RecentActivity:       recentActivity,
		StatusTrends:         statusTrends,
		CurrencyDistribution: currencyDistribution,
	}

	// Calculate total count
	for _, count := range countByStatus {
		analytics.TotalCount += count
	}

	return analytics, nil
}

// CreateSmartChequeBatch creates multiple smart checks in a batch operation
func (s *smartChequeService) CreateSmartChequeBatch(ctx context.Context, requests []*CreateSmartChequeRequest) (*SmartChequeBatchOperationResult, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}

	// Limit batch size to prevent excessive load
	if len(requests) > 100 {
		return nil, fmt.Errorf("batch size limit exceeded: maximum 100 smart checks per batch")
	}

	// Validate all requests first
	smartCheques := make([]*models.SmartCheque, len(requests))
	for i, request := range requests {
		// Validate request
		if err := s.validateCreateRequest(request); err != nil {
			return nil, fmt.Errorf("validation failed for request %d: %w", i, err)
		}

		// Create smart check model
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

		smartCheques[i] = smartCheque
	}

	// Use repository batch create method
	err := s.smartChequeRepo.BatchCreateSmartCheques(ctx, smartCheques)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart checks in batch: %w", err)
	}

	// Create result
	result := &SmartChequeBatchOperationResult{
		SuccessCount: len(smartCheques),
		FailureCount: 0,
		Results:      make([]BatchOperationResult, len(smartCheques)),
	}

	for i, smartCheque := range smartCheques {
		result.Results[i] = BatchOperationResult{
			ID:      smartCheque.ID,
			Success: true,
		}
	}

	// Create audit log entries for each created smart check
	for _, smartCheque := range smartCheques {
		auditEntry := &AuditLogEntry{
			SmartChequeID: smartCheque.ID,
			Action:        "smart_check_created",
			Details:       fmt.Sprintf("Smart check created with amount %f %s", smartCheque.Amount, smartCheque.Currency),
			Timestamp:     time.Now(),
		}
		_ = s.CreateAuditLog(ctx, auditEntry) // Log error but don't fail the operation
	}

	return result, nil
}

// UpdateSmartChequeBatch updates multiple smart checks in a batch operation
func (s *smartChequeService) UpdateSmartChequeBatch(ctx context.Context, updates map[string]*UpdateSmartChequeRequest) (*SmartChequeBatchOperationResult, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	// Limit batch size to prevent excessive load
	if len(updates) > 100 {
		return nil, fmt.Errorf("batch size limit exceeded: maximum 100 smart checks per batch")
	}

	// Validate and prepare smart checks for update
	smartCheques := make([]*models.SmartCheque, 0, len(updates))
	result := &SmartChequeBatchOperationResult{
		SuccessCount: 0,
		FailureCount: 0,
		Results:      make([]BatchOperationResult, 0, len(updates)),
	}

	// Process each update
	for id, request := range updates {
		batchResult := BatchOperationResult{
			ID: id,
		}

		// Validate request
		if err := s.validateUpdateRequest(id, request); err != nil {
			batchResult.Success = false
			batchResult.Error = err.Error()
			result.FailureCount++
			result.Results = append(result.Results, batchResult)
			continue
		}

		// Get existing smart check
		smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, id)
		if err != nil {
			batchResult.Success = false
			batchResult.Error = fmt.Sprintf("failed to get smart check: %v", err)
			result.FailureCount++
			result.Results = append(result.Results, batchResult)
			continue
		}

		if smartCheque == nil {
			batchResult.Success = false
			batchResult.Error = "smart check not found"
			result.FailureCount++
			result.Results = append(result.Results, batchResult)
			continue
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

		smartCheques = append(smartCheques, smartCheque)
		batchResult.Success = true
		result.SuccessCount++
		result.Results = append(result.Results, batchResult)
	}

	// Use repository batch update method
	if len(smartCheques) > 0 {
		err := s.smartChequeRepo.BatchUpdateSmartCheques(ctx, smartCheques)
		if err != nil {
			return nil, fmt.Errorf("failed to update smart checks in batch: %w", err)
		}

		// Create audit log entries for each updated smart check
		for _, smartCheque := range smartCheques {
			auditEntry := &AuditLogEntry{
				SmartChequeID: smartCheque.ID,
				Action:        "smart_check_updated",
				Details:       fmt.Sprintf("Smart check updated with status %s", smartCheque.Status),
				Timestamp:     time.Now(),
			}
			_ = s.CreateAuditLog(ctx, auditEntry) // Log error but don't fail the operation
		}
	}

	return result, nil
}

// CreateAuditLog creates an audit log entry for smart check operations
func (s *smartChequeService) CreateAuditLog(ctx context.Context, entry *AuditLogEntry) error {
	// Convert our audit log entry to the model used by the audit repository
	auditLog := &models.AuditLog{
		Action:     entry.Action,
		Resource:   "smart_check",
		ResourceID: &entry.SmartChequeID,
		Details:    entry.Details,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		Success:    true, // Assume success for audit logs of operations
		CreatedAt:  entry.Timestamp,
	}

	// If we have a user ID, convert it to UUID
	if entry.UserID != "" {
		if userID, err := uuid.Parse(entry.UserID); err == nil {
			auditLog.UserID = userID
		}
	}

	return s.auditRepo.CreateAuditLog(auditLog)
}

// GetAuditTrail retrieves the audit trail for a specific smart check
func (s *smartChequeService) GetAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]AuditLogEntry, error) {
	if smartChequeID == "" {
		return nil, fmt.Errorf("smart check ID is required")
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// Get audit logs from repository
	auditLogs, err := s.auditRepo.GetAuditLogs(nil, nil, "", "smart_check", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// Filter logs for this specific smart check
	var result []AuditLogEntry
	for _, log := range auditLogs {
		// Check if this log is for our smart check
		if log.ResourceID != nil && *log.ResourceID == smartChequeID {
			entry := AuditLogEntry{
				ID:            log.ID.String(),
				SmartChequeID: *log.ResourceID,
				Action:        log.Action,
				Details:       log.Details,
				Timestamp:     log.CreatedAt,
				IPAddress:     log.IPAddress,
				UserAgent:     log.UserAgent,
			}

			// Add user ID if available
			if log.UserID != uuid.Nil {
				entry.UserID = log.UserID.String()
			}

			result = append(result, entry)
		}
	}

	return result, nil
}
