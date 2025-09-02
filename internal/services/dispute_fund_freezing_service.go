package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// DisputeFundFreezingServiceInterface defines the interface for dispute fund freezing operations
type DisputeFundFreezingServiceInterface interface {
	// FreezeFundsForDispute freezes funds when a dispute is initiated
	FreezeFundsForDispute(ctx context.Context, disputeID string, smartChequeID string, freezeReason string, userID string) error

	// UnfreezeFundsForDispute unfreezes funds when a dispute is resolved
	UnfreezeFundsForDispute(ctx context.Context, disputeID string, smartChequeIDID string, unfreezeReason string, userID string) error

	// GetFundFreezingStatus gets the current fund freezing status for a dispute
	GetFundFreezingStatus(ctx context.Context, disputeID string) (*FundFreezingStatus, error)

	// GetFrozenFundsByDispute gets all frozen funds for a specific dispute
	GetFrozenFundsByDispute(ctx context.Context, disputeID string) ([]*FrozenFund, error)

	// GetFrozenFundsByEnterprise gets all frozen funds for an enterprise
	GetFrozenFundsByEnterprise(ctx context.Context, enterpriseID string) ([]*FrozenFund, error)

	// UpdateFreezingStatus updates the freezing status for a dispute
	UpdateFreezingStatus(ctx context.Context, disputeID string, status FundFreezingStatusType, reason string, userID string) error

	// GetFreezingHistory gets the freezing history for a dispute
	GetFreezingHistory(ctx context.Context, disputeID string) ([]*FundFreezingEvent, error)
}

// FundFreezingStatusType represents the status of fund freezing
type FundFreezingStatusType string

const (
	FundFreezingStatusNotFrozen FundFreezingStatusType = "not_frozen"
	FundFreezingStatusFrozen    FundFreezingStatusType = "frozen"
	FundFreezingStatusUnfrozen  FundFreezingStatusType = "unfrozen"
	FundFreezingStatusPending   FundFreezingStatusType = "pending"
)

// FundFreezingStatus represents the current status of fund freezing for a dispute
type FundFreezingStatus struct {
	DisputeID      string                 `json:"dispute_id"`
	SmartChequeID  string                 `json:"smart_check_id"`
	Status         FundFreezingStatusType `json:"status"`
	FrozenAmount   float64                `json:"frozen_amount"`
	Currency       models.Currency        `json:"currency"`
	FrozenAt       *time.Time             `json:"frozen_at,omitempty"`
	UnfrozenAt     *time.Time             `json:"unfrozen_at,omitempty"`
	FreezeReason   string                 `json:"freeze_reason"`
	UnfreezeReason *string                `json:"unfreeze_reason,omitempty"`
	FrozenBy       string                 `json:"frozen_by"`
	UnfrozenBy     *string                `json:"unfrozen_by,omitempty"`
	LastUpdatedAt  time.Time              `json:"last_updated_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// FrozenFund represents a frozen fund entry
type FrozenFund struct {
	ID             string                 `json:"id"`
	DisputeID      string                 `json:"dispute_id"`
	SmartChequeID  string                 `json:"smart_check_id"`
	EnterpriseID   string                 `json:"enterprise_id"`
	Amount         float64                `json:"amount"`
	Currency       models.Currency        `json:"currency"`
	Status         FundFreezingStatusType `json:"status"`
	FrozenAt       time.Time              `json:"frozen_at"`
	FreezeReason   string                 `json:"freeze_reason"`
	FrozenBy       string                 `json:"frozen_by"`
	UnfrozenAt     *time.Time             `json:"unfrozen_at,omitempty"`
	UnfreezeReason *string                `json:"unfreeze_reason,omitempty"`
	UnfrozenBy     *string                `json:"unfrozen_by,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// FundFreezingEvent represents a fund freezing event
type FundFreezingEvent struct {
	ID        string                 `json:"id"`
	DisputeID string                 `json:"dispute_id"`
	EventType string                 `json:"event_type"` // freeze, unfreeze, status_update
	Status    FundFreezingStatusType `json:"status"`
	Amount    float64                `json:"amount"`
	Currency  models.Currency        `json:"currency"`
	Reason    string                 `json:"reason"`
	UserID    string                 `json:"user_id"`
	UserType  string                 `json:"user_type"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// DisputeFundFreezingService implements dispute fund freezing operations
type DisputeFundFreezingService struct {
	disputeRepo     repository.DisputeRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
	fraudRepo       repository.FraudRepositoryInterface
	xrplService     repository.XRPLServiceInterface
	messagingClient *messaging.Service
	auditRepo       repository.AuditRepositoryInterface
}

// NewDisputeFundFreezingService creates a new dispute fund freezing service
func NewDisputeFundFreezingService(
	disputeRepo repository.DisputeRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	fraudRepo repository.FraudRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient *messaging.Service,
	auditRepo repository.AuditRepositoryInterface,
) DisputeFundFreezingServiceInterface {
	return &DisputeFundFreezingService{
		disputeRepo:     disputeRepo,
		smartChequeRepo: smartChequeRepo,
		fraudRepo:       fraudRepo,
		xrplService:     xrplService,
		messagingClient: messagingClient,
		auditRepo:       auditRepo,
	}
}

// FreezeFundsForDispute freezes funds when a dispute is initiated
func (s *DisputeFundFreezingService) FreezeFundsForDispute(ctx context.Context, disputeID string, smartChequeID string, freezeReason string, userID string) error {
	// Get the dispute to validate it exists and is in the right status
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Validate dispute status allows freezing
	if !s.canFreezeFunds(dispute.Status) {
		return fmt.Errorf("cannot freeze funds for dispute in status: %s", dispute.Status)
	}

	// Get the Smart Check to get amount and currency
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart check: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart check not found: %s", smartChequeID)
	}

	// Check if funds are already frozen
	existingStatus, err := s.GetFundFreezingStatus(ctx, disputeID)
	if err == nil && existingStatus.Status == FundFreezingStatusFrozen {
		return fmt.Errorf("funds are already frozen for dispute: %s", disputeID)
	}

	// Create frozen fund entry
	frozenFund := &FrozenFund{
		ID:            uuid.New().String(),
		DisputeID:     disputeID,
		SmartChequeID: smartChequeID,
		EnterpriseID:  smartCheque.PayerID,
		Amount:        smartCheque.Amount,
		Currency:      smartCheque.Currency,
		Status:        FundFreezingStatusFrozen,
		FrozenAt:      time.Now(),
		FreezeReason:  freezeReason,
		FrozenBy:      userID,
		Metadata: map[string]interface{}{
			"dispute_category": dispute.Category,
			"dispute_priority": dispute.Priority,
			"disputed_amount":  dispute.DisputedAmount,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store frozen fund (this would be implemented in a repository)
	// For now, we'll simulate the storage
	log.Printf("Freezing funds for dispute %s: amount=%f %s", disputeID, frozenFund.Amount, frozenFund.Currency)

	// Update enterprise fraud status to indicate frozen funds
	err = s.updateEnterpriseFraudStatus(ctx, smartCheque.PayerID, "frozen", freezeReason, userID)
	if err != nil {
		log.Printf("Warning: failed to update enterprise fraud status: %v", err)
	}

	// Publish fund freezing event
	err = s.publishFundFreezingEvent(ctx, &FundFreezingEvent{
		ID:        uuid.New().String(),
		DisputeID: disputeID,
		EventType: "freeze",
		Status:    FundFreezingStatusFrozen,
		Amount:    frozenFund.Amount,
		Currency:  frozenFund.Currency,
		Reason:    freezeReason,
		UserID:    userID,
		UserType:  "admin", // This should come from the user context
		Metadata: map[string]interface{}{
			"smart_check_id": smartChequeID,
			"enterprise_id":  smartCheque.PayerID,
		},
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Printf("Warning: failed to publish fund freezing event: %v", err)
	}

	// Create audit log entry
	err = s.createAuditLog(ctx, disputeID, "funds_frozen", userID, map[string]interface{}{
		"amount":      frozenFund.Amount,
		"currency":    frozenFund.Currency,
		"reason":      freezeReason,
		"smart_check": smartChequeID,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// UnfreezeFundsForDispute unfreezes funds when a dispute is resolved
func (s *DisputeFundFreezingService) UnfreezeFundsForDispute(ctx context.Context, disputeID string, smartChequeID string, unfreezeReason string, userID string) error {
	// Get the dispute to validate it exists and is resolved
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Validate dispute status allows unfreezing
	if !s.canUnfreezeFunds(dispute.Status) {
		return fmt.Errorf("cannot unfreeze funds for dispute in status: %s", dispute.Status)
	}

	// Get the current freezing status
	freezingStatus, err := s.GetFundFreezingStatus(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get fund freezing status: %w", err)
	}

	if freezingStatus.Status != FundFreezingStatusFrozen {
		return fmt.Errorf("funds are not currently frozen for dispute: %s", disputeID)
	}

	// Update frozen fund status
	// In a real implementation, this would update the database
	log.Printf("Unfreezing funds for dispute %s: amount=%f %s", disputeID, freezingStatus.FrozenAmount, freezingStatus.Currency)

	// Update enterprise fraud status to remove frozen status
	err = s.updateEnterpriseFraudStatus(ctx, dispute.InitiatorID, "normal", unfreezeReason, userID)
	if err != nil {
		log.Printf("Warning: failed to update enterprise fraud status: %v", err)
	}

	// Publish fund unfreezing event
	err = s.publishFundFreezingEvent(ctx, &FundFreezingEvent{
		ID:        uuid.New().String(),
		DisputeID: disputeID,
		EventType: "unfreeze",
		Status:    FundFreezingStatusUnfrozen,
		Amount:    freezingStatus.FrozenAmount,
		Currency:  freezingStatus.Currency,
		Reason:    unfreezeReason,
		UserID:    userID,
		UserType:  "admin", // This should come from the user context
		Metadata: map[string]interface{}{
			"smart_check_id": smartChequeID,
			"enterprise_id":  dispute.InitiatorID,
		},
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Printf("Warning: failed to publish fund unfreezing event: %v", err)
	}

	// Create audit log entry
	err = s.createAuditLog(ctx, disputeID, "funds_unfrozen", userID, map[string]interface{}{
		"amount":      freezingStatus.FrozenAmount,
		"currency":    freezingStatus.Currency,
		"reason":      unfreezeReason,
		"smart_check": smartChequeID,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// GetFundFreezingStatus gets the current fund freezing status for a dispute
func (s *DisputeFundFreezingService) GetFundFreezingStatus(ctx context.Context, disputeID string) (*FundFreezingStatus, error) {
	// In a real implementation, this would query the database
	// For now, return a mock status
	return &FundFreezingStatus{
		DisputeID:     disputeID,
		SmartChequeID: "mock-smart-check-id",
		Status:        FundFreezingStatusNotFrozen,
		FrozenAmount:  0,
		Currency:      models.CurrencyUSDT,
		LastUpdatedAt: time.Now(),
	}, nil
}

// GetFrozenFundsByDispute gets all frozen funds for a specific dispute
func (s *DisputeFundFreezingService) GetFrozenFundsByDispute(ctx context.Context, disputeID string) ([]*FrozenFund, error) {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []*FrozenFund{}, nil
}

// GetFrozenFundsByEnterprise gets all frozen funds for an enterprise
func (s *DisputeFundFreezingService) GetFrozenFundsByEnterprise(ctx context.Context, enterpriseID string) ([]*FrozenFund, error) {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []*FrozenFund{}, nil
}

// UpdateFreezingStatus updates the freezing status for a dispute
func (s *DisputeFundFreezingService) UpdateFreezingStatus(ctx context.Context, disputeID string, status FundFreezingStatusType, reason string, userID string) error {
	// Validate status transition
	if !s.isValidStatusTransition(status) {
		return fmt.Errorf("invalid status transition to: %s", status)
	}

	// In a real implementation, this would update the database
	log.Printf("Updating freezing status for dispute %s to %s: %s", disputeID, status, reason)

	// Create audit log entry
	err := s.createAuditLog(ctx, disputeID, "freezing_status_updated", userID, map[string]interface{}{
		"new_status": status,
		"reason":     reason,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// GetFreezingHistory gets the freezing history for a dispute
func (s *DisputeFundFreezingService) GetFreezingHistory(ctx context.Context, disputeID string) ([]*FundFreezingEvent, error) {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []*FundFreezingEvent{}, nil
}

// Helper methods

// canFreezeFunds checks if funds can be frozen for a dispute status
func (s *DisputeFundFreezingService) canFreezeFunds(status models.DisputeStatus) bool {
	// Funds can be frozen when dispute is initiated or under review
	return status == models.DisputeStatusInitiated || status == models.DisputeStatusUnderReview
}

// canUnfreezeFunds checks if funds can be unfrozen for a dispute status
func (s *DisputeFundFreezingService) canUnfreezeFunds(status models.DisputeStatus) bool {
	// Funds can be unfrozen when dispute is resolved or closed
	return status == models.DisputeStatusResolved || status == models.DisputeStatusClosed
}

// isValidStatusTransition checks if a status transition is valid
func (s *DisputeFundFreezingService) isValidStatusTransition(status FundFreezingStatusType) bool {
	validStatuses := []FundFreezingStatusType{
		FundFreezingStatusNotFrozen,
		FundFreezingStatusFrozen,
		FundFreezingStatusUnfrozen,
		FundFreezingStatusPending,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// updateEnterpriseFraudStatus updates the fraud status for an enterprise
func (s *DisputeFundFreezingService) updateEnterpriseFraudStatus(_ context.Context, enterpriseID string, status string, reason string, _ string) error {
	// In a real implementation, this would call the fraud prevention service
	// For now, just log the action
	log.Printf("Updating enterprise %s fraud status to %s: %s", enterpriseID, status, reason)
	return nil
}

// publishFundFreezingEvent publishes a fund freezing event
func (s *DisputeFundFreezingService) publishFundFreezingEvent(_ context.Context, event *FundFreezingEvent) error {
	// In a real implementation, this would publish to the messaging system
	// For now, just log the event
	log.Printf("Publishing fund freezing event: %s for dispute %s", event.EventType, event.DisputeID)
	return nil
}

// createAuditLog creates an audit log entry
func (s *DisputeFundFreezingService) createAuditLog(ctx context.Context, disputeID string, action string, userID string, details map[string]interface{}) error {
	// In a real implementation, this would create an audit log entry
	// For now, just log the action
	log.Printf("Creating audit log: %s for dispute %s by user %s", action, disputeID, userID)
	return nil
}
