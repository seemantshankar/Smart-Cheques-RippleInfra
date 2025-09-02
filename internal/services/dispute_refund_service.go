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

// DisputeRefundServiceInterface defines the interface for dispute refund operations
type DisputeRefundServiceInterface interface {
	// ProcessRefund processes a refund for a dispute resolution
	ProcessRefund(ctx context.Context, disputeID string, refundRequest *RefundRequest) (*RefundResult, error)

	// ProcessPartialPayment processes a partial payment for a dispute resolution
	ProcessPartialPayment(ctx context.Context, disputeID string, partialPaymentRequest *PartialPaymentRequest) (*PartialPaymentResult, error)

	// GetRefundStatus gets the current status of a refund
	GetRefundStatus(ctx context.Context, refundID string) (*RefundStatus, error)

	// GetRefundHistory gets the refund history for a dispute
	GetRefundHistory(ctx context.Context, disputeID string) ([]*RefundRecord, error)

	// ApproveRefund approves a refund request
	ApproveRefund(ctx context.Context, refundID string, approverID string, approvalNotes string) error

	// RejectRefund rejects a refund request
	RejectRefund(ctx context.Context, refundID string, rejectorID string, rejectionReason string) error

	// ExecuteRefund executes an approved refund
	ExecuteRefund(ctx context.Context, refundID string, executorID string) error
}

// RefundRequest represents a refund request
type RefundRequest struct {
	DisputeID     string                 `json:"dispute_id" validate:"required"`
	SmartChequeID string                 `json:"smart_check_id" validate:"required"`
	RefundAmount  float64                `json:"refund_amount" validate:"required,gt=0"`
	Currency      models.Currency        `json:"currency" validate:"required"`
	RefundReason  string                 `json:"refund_reason" validate:"required"`
	RefundType    RefundType             `json:"refund_type" validate:"required"`
	RequestedBy   string                 `json:"requested_by" validate:"required"`
	Priority      RefundPriority         `json:"priority" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PartialPaymentRequest represents a partial payment request
type PartialPaymentRequest struct {
	DisputeID          string                 `json:"dispute_id" validate:"required"`
	SmartChequeID      string                 `json:"smart_check_id" validate:"required"`
	PartialAmount      float64                `json:"partial_amount" validate:"required,gt=0"`
	OriginalAmount     float64                `json:"original_amount" validate:"required,gt=0"`
	Currency           models.Currency        `json:"currency" validate:"required"`
	PaymentReason      string                 `json:"payment_reason" validate:"required"`
	MilestoneCompleted string                 `json:"milestone_completed,omitempty"`
	RequestedBy        string                 `json:"requested_by" validate:"required"`
	Priority           RefundPriority         `json:"priority" validate:"required"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// RefundType represents the type of refund
type RefundType string

const (
	RefundTypeFull    RefundType = "full"
	RefundTypePartial RefundType = "partial"
	RefundTypePenalty RefundType = "penalty"
)

// RefundPriority represents the priority of a refund
type RefundPriority string

const (
	RefundPriorityLow    RefundPriority = "low"
	RefundPriorityNormal RefundPriority = "normal"
	RefundPriorityHigh   RefundPriority = "high"
	RefundPriorityUrgent RefundPriority = "urgent"
)

// RefundResult represents the result of a refund operation
type RefundResult struct {
	RefundID      string                 `json:"refund_id"`
	DisputeID     string                 `json:"dispute_id"`
	SmartChequeID string                 `json:"smart_check_id"`
	Amount        float64                `json:"amount"`
	Currency      models.Currency        `json:"currency"`
	Status        RefundStatus           `json:"status"`
	RefundType    RefundType             `json:"refund_type"`
	RequestedAt   time.Time              `json:"requested_at"`
	RequestedBy   string                 `json:"requested_by"`
	Priority      RefundPriority         `json:"priority"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PartialPaymentResult represents the result of a partial payment operation
type PartialPaymentResult struct {
	PaymentID          string                 `json:"payment_id"`
	DisputeID          string                 `json:"dispute_id"`
	SmartChequeID      string                 `json:"smart_check_id"`
	Amount             float64                `json:"amount"`
	Currency           models.Currency        `json:"currency"`
	Status             PaymentStatus          `json:"status"`
	MilestoneCompleted string                 `json:"milestone_completed,omitempty"`
	RequestedAt        time.Time              `json:"requested_at"`
	RequestedBy        string                 `json:"requested_by"`
	Priority           RefundPriority         `json:"priority"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusRejected  RefundStatus = "rejected"
	RefundStatusExecuting RefundStatus = "executing"
	RefundStatusCompleted RefundStatus = "completed"
	RefundStatusFailed    RefundStatus = "failed"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusApproved  PaymentStatus = "approved"
	PaymentStatusRejected  PaymentStatus = "rejected"
	PaymentStatusExecuting PaymentStatus = "executing"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

// RefundRecord represents a refund record
type RefundRecord struct {
	ID            string                 `json:"id"`
	DisputeID     string                 `json:"dispute_id"`
	SmartChequeID string                 `json:"smart_check_id"`
	Amount        float64                `json:"amount"`
	Currency      models.Currency        `json:"currency"`
	Status        RefundStatus           `json:"status"`
	RefundType    RefundType             `json:"refund_type"`
	RequestedAt   time.Time              `json:"requested_at"`
	RequestedBy   string                 `json:"requested_by"`
	ApprovedAt    *time.Time             `json:"approved_at,omitempty"`
	ApprovedBy    *string                `json:"approved_by,omitempty"`
	ExecutedAt    *time.Time             `json:"executed_at,omitempty"`
	ExecutedBy    *string                `json:"executed_by,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// DisputeRefundService implements dispute refund operations
type DisputeRefundService struct {
	disputeRepo     repository.DisputeRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
	xrplService     repository.XRPLServiceInterface
	messagingClient *messaging.Service
	auditRepo       repository.AuditRepositoryInterface
}

// NewDisputeRefundService creates a new dispute refund service
func NewDisputeRefundService(
	disputeRepo repository.DisputeRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient *messaging.Service,
	auditRepo repository.AuditRepositoryInterface,
) DisputeRefundServiceInterface {
	return &DisputeRefundService{
		disputeRepo:     disputeRepo,
		smartChequeRepo: smartChequeRepo,
		xrplService:     xrplService,
		messagingClient: messagingClient,
		auditRepo:       auditRepo,
	}
}

// ProcessRefund processes a refund for a dispute resolution
func (s *DisputeRefundService) ProcessRefund(ctx context.Context, disputeID string, refundRequest *RefundRequest) (*RefundResult, error) {
	// Validate the request
	if err := s.validateRefundRequest(refundRequest); err != nil {
		return nil, fmt.Errorf("invalid refund request: %w", err)
	}

	// Get the dispute to validate it exists and is in the right status
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Validate dispute status allows refunds
	if !s.canProcessRefund(dispute.Status) {
		return nil, fmt.Errorf("cannot process refund for dispute in status: %s", dispute.Status)
	}

	// Get the Smart Check to validate the refund amount
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, refundRequest.SmartChequeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart check: %w", err)
	}
	if smartCheque == nil {
		return nil, fmt.Errorf("smart check not found: %s", refundRequest.SmartChequeID)
	}

	// Validate refund amount
	if refundRequest.RefundAmount > smartCheque.Amount {
		return nil, fmt.Errorf("refund amount %f exceeds smart check amount %f", refundRequest.RefundAmount, smartCheque.Amount)
	}

	// Create refund record
	refundID := uuid.New().String()
	refundRecord := &RefundRecord{
		ID:            refundID,
		DisputeID:     disputeID,
		SmartChequeID: refundRequest.SmartChequeID,
		Amount:        refundRequest.RefundAmount,
		Currency:      refundRequest.Currency,
		Status:        RefundStatusPending,
		RefundType:    refundRequest.RefundType,
		RequestedAt:   time.Now(),
		RequestedBy:   refundRequest.RequestedBy,
		Metadata:      refundRequest.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Store refund record (this would be implemented in a repository)
	// For now, we'll simulate the storage
	log.Printf("Processing refund for dispute %s: amount=%f %s", disputeID, refundRequest.RefundAmount, refundRequest.Currency)

	// Create audit log entry
	err = s.createAuditLog(ctx, disputeID, "refund_requested", refundRequest.RequestedBy, map[string]interface{}{
		"refund_id":     refundID,
		"amount":        refundRequest.RefundAmount,
		"currency":      refundRequest.Currency,
		"refund_type":   refundRequest.RefundType,
		"refund_reason": refundRequest.RefundReason,
		"priority":      refundRequest.Priority,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	// Publish refund event
	err = s.publishRefundEvent(ctx, "refund_requested", refundRecord)
	if err != nil {
		log.Printf("Warning: failed to publish refund event: %v", err)
	}

	return &RefundResult{
		RefundID:      refundID,
		DisputeID:     disputeID,
		SmartChequeID: refundRequest.SmartChequeID,
		Amount:        refundRequest.RefundAmount,
		Currency:      refundRequest.Currency,
		Status:        RefundStatusPending,
		RefundType:    refundRequest.RefundType,
		RequestedAt:   refundRecord.RequestedAt,
		RequestedBy:   refundRequest.RequestedBy,
		Priority:      refundRequest.Priority,
		Metadata:      refundRequest.Metadata,
	}, nil
}

// ProcessPartialPayment processes a partial payment for a dispute resolution
func (s *DisputeRefundService) ProcessPartialPayment(ctx context.Context, disputeID string, partialPaymentRequest *PartialPaymentRequest) (*PartialPaymentResult, error) {
	// Validate the request
	if err := s.validatePartialPaymentRequest(partialPaymentRequest); err != nil {
		return nil, fmt.Errorf("invalid partial payment request: %w", err)
	}

	// Get the dispute to validate it exists and is in the right status
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Validate dispute status allows partial payments
	if !s.canProcessPartialPayment(dispute.Status) {
		return nil, fmt.Errorf("cannot process partial payment for dispute in status: %s", dispute.Status)
	}

	// Validate partial payment amount
	if partialPaymentRequest.PartialAmount >= partialPaymentRequest.OriginalAmount {
		return nil, fmt.Errorf("partial amount %f must be less than original amount %f", partialPaymentRequest.PartialAmount, partialPaymentRequest.OriginalAmount)
	}

	// Create partial payment record
	paymentID := uuid.New().String()
	partialPaymentRecord := &PartialPaymentResult{
		PaymentID:          paymentID,
		DisputeID:          disputeID,
		SmartChequeID:      partialPaymentRequest.SmartChequeID,
		Amount:             partialPaymentRequest.PartialAmount,
		Currency:           partialPaymentRequest.Currency,
		Status:             PaymentStatusPending,
		MilestoneCompleted: partialPaymentRequest.MilestoneCompleted,
		RequestedAt:        time.Now(),
		RequestedBy:        partialPaymentRequest.RequestedBy,
		Priority:           partialPaymentRequest.Priority,
		Metadata:           partialPaymentRequest.Metadata,
	}

	// Store partial payment record (this would be implemented in a repository)
	// For now, we'll simulate the storage
	log.Printf("Processing partial payment for dispute %s: amount=%f %s", disputeID, partialPaymentRequest.PartialAmount, partialPaymentRequest.Currency)

	// Create audit log entry
	err = s.createAuditLog(ctx, disputeID, "partial_payment_requested", partialPaymentRequest.RequestedBy, map[string]interface{}{
		"payment_id":          paymentID,
		"amount":              partialPaymentRequest.PartialAmount,
		"currency":            partialPaymentRequest.Currency,
		"payment_reason":      partialPaymentRequest.PaymentReason,
		"milestone_completed": partialPaymentRequest.MilestoneCompleted,
		"priority":            partialPaymentRequest.Priority,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	// Publish partial payment event
	err = s.publishPartialPaymentEvent(ctx, "partial_payment_requested", partialPaymentRecord)
	if err != nil {
		log.Printf("Warning: failed to publish partial payment event: %v", err)
	}

	return partialPaymentRecord, nil
}

// GetRefundStatus gets the current status of a refund
func (s *DisputeRefundService) GetRefundStatus(ctx context.Context, refundID string) (*RefundStatus, error) {
	// In a real implementation, this would query the database
	// For now, return a mock status
	status := RefundStatusPending
	return &status, nil
}

// GetRefundHistory gets the refund history for a dispute
func (s *DisputeRefundService) GetRefundHistory(ctx context.Context, disputeID string) ([]*RefundRecord, error) {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []*RefundRecord{}, nil
}

// ApproveRefund approves a refund request
func (s *DisputeRefundService) ApproveRefund(ctx context.Context, refundID string, approverID string, approvalNotes string) error {
	// In a real implementation, this would update the database
	// For now, just log the action
	log.Printf("Approving refund %s by %s: %s", refundID, approverID, approvalNotes)

	// Create audit log entry
	err := s.createAuditLog(ctx, "unknown", "refund_approved", approverID, map[string]interface{}{
		"refund_id":      refundID,
		"approval_notes": approvalNotes,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// RejectRefund rejects a refund request
func (s *DisputeRefundService) RejectRefund(ctx context.Context, refundID string, rejectorID string, rejectionReason string) error {
	// In a real implementation, this would update the database
	// For now, just log the action
	log.Printf("Rejecting refund %s by %s: %s", refundID, rejectorID, rejectionReason)

	// Create audit log entry
	err := s.createAuditLog(ctx, "unknown", "refund_rejected", rejectorID, map[string]interface{}{
		"refund_id":        refundID,
		"rejection_reason": rejectionReason,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// ExecuteRefund executes an approved refund
func (s *DisputeRefundService) ExecuteRefund(ctx context.Context, refundID string, executorID string) error {
	// In a real implementation, this would execute the refund via XRPL
	// For now, just log the action
	log.Printf("Executing refund %s by %s", refundID, executorID)

	// Create audit log entry
	err := s.createAuditLog(ctx, "unknown", "refund_executed", executorID, map[string]interface{}{
		"refund_id": refundID,
	})
	if err != nil {
		log.Printf("Warning: failed to create audit log: %v", err)
	}

	return nil
}

// Helper methods

// validateRefundRequest validates a refund request
func (s *DisputeRefundService) validateRefundRequest(request *RefundRequest) error {
	if request.RefundAmount <= 0 {
		return fmt.Errorf("refund amount must be greater than 0")
	}
	if request.RefundReason == "" {
		return fmt.Errorf("refund reason is required")
	}
	if request.RequestedBy == "" {
		return fmt.Errorf("requester ID is required")
	}
	return nil
}

// validatePartialPaymentRequest validates a partial payment request
func (s *DisputeRefundService) validatePartialPaymentRequest(request *PartialPaymentRequest) error {
	if request.PartialAmount <= 0 {
		return fmt.Errorf("partial amount must be greater than 0")
	}
	if request.OriginalAmount <= 0 {
		return fmt.Errorf("original amount must be greater than 0")
	}
	if request.PartialAmount >= request.OriginalAmount {
		return fmt.Errorf("partial amount must be less than original amount")
	}
	if request.PaymentReason == "" {
		return fmt.Errorf("payment reason is required")
	}
	if request.RequestedBy == "" {
		return fmt.Errorf("requester ID is required")
	}
	return nil
}

// canProcessRefund checks if a refund can be processed for a dispute status
func (s *DisputeRefundService) canProcessRefund(status models.DisputeStatus) bool {
	// Refunds can be processed when dispute is resolved or closed
	return status == models.DisputeStatusResolved || status == models.DisputeStatusClosed
}

// canProcessPartialPayment checks if a partial payment can be processed for a dispute status
func (s *DisputeRefundService) canProcessPartialPayment(status models.DisputeStatus) bool {
	// Partial payments can be processed when dispute is resolved or closed
	return status == models.DisputeStatusResolved || status == models.DisputeStatusClosed
}

// createAuditLog creates an audit log entry
func (s *DisputeRefundService) createAuditLog(ctx context.Context, disputeID string, action string, userID string, details map[string]interface{}) error {
	// In a real implementation, this would create an audit log entry
	// For now, just log the action
	log.Printf("Creating audit log: %s for dispute %s by user %s", action, disputeID, userID)
	return nil
}

// publishRefundEvent publishes a refund event
func (s *DisputeRefundService) publishRefundEvent(ctx context.Context, eventType string, refundRecord *RefundRecord) error {
	// In a real implementation, this would publish to the messaging system
	// For now, just log the event
	log.Printf("Publishing refund event: %s for refund %s", eventType, refundRecord.ID)
	return nil
}

// publishPartialPaymentEvent publishes a partial payment event
func (s *DisputeRefundService) publishPartialPaymentEvent(ctx context.Context, eventType string, partialPaymentRecord *PartialPaymentResult) error {
	// In a real implementation, this would publish to the messaging system
	// For now, just log the event
	log.Printf("Publishing partial payment event: %s for payment %s", eventType, partialPaymentRecord.PaymentID)
	return nil
}
