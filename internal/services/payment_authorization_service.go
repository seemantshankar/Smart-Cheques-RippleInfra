package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// PaymentAuthorizationServiceInterface defines the interface for SmartCheque payment authorization operations
type PaymentAuthorizationServiceInterface interface {
	// Payment Authorization Request Management
	CreatePaymentAuthorizationRequest(ctx context.Context, req *PaymentAuthorizationRequest) (*PaymentAuthorization, error)
	GetPaymentAuthorizationRequest(ctx context.Context, requestID uuid.UUID) (*PaymentAuthorization, error)
	GetPendingPaymentAuthorizations(ctx context.Context, enterpriseID uuid.UUID) ([]*PaymentAuthorization, error)
	GetPaymentAuthorizationsBySmartCheque(ctx context.Context, smartChequeID string) ([]*PaymentAuthorization, error)

	// Authorization Workflow
	ApprovePayment(ctx context.Context, req *PaymentApprovalRequest) (*PaymentApproval, error)
	RejectPayment(ctx context.Context, req *PaymentRejectionRequest) (*PaymentApproval, error)
	CheckPaymentAuthorizationStatus(ctx context.Context, requestID uuid.UUID) (*PaymentAuthorizationStatus, error)

	// Auto-approval for Low-risk Payments
	ProcessAutoApproval(ctx context.Context, requestID uuid.UUID) error

	// Time-locked Payments for High-risk
	CreateTimeLockPayment(ctx context.Context, req *TimeLockPaymentRequest) (*TimeLockPayment, error)
	ReleaseTimeLockPayment(ctx context.Context, lockID uuid.UUID) error
	GetTimeLockPaymentStatus(ctx context.Context, lockID uuid.UUID) (*TimeLockPaymentStatus, error)

	// Risk Assessment for Payments
	AssessPaymentRisk(ctx context.Context, req *PaymentRiskAssessmentRequest) (*PaymentRiskScore, error)
	GetPaymentRiskProfile(ctx context.Context, enterpriseID uuid.UUID) (*EnterprisePaymentRiskProfile, error)

	// Bulk Operations
	BulkApprovePayments(ctx context.Context, req *BulkPaymentApprovalRequest) (*BulkPaymentApprovalResult, error)
	GetPaymentAuthorizationHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*PaymentAuthorization, error)

	// Milestone-triggered Payment Authorization
	InitiatePaymentFromMilestone(ctx context.Context, milestoneID string) (*PaymentAuthorization, error)
}

// PaymentAuthorizationService implements the payment authorization service interface
type PaymentAuthorizationService struct {
	smartChequeRepo     repository.SmartChequeRepositoryInterface
	milestoneRepo       repository.MilestoneRepositoryInterface
	contractRepo        repository.ContractRepositoryInterface
	userService         UserServiceInterface
	messagingClient     messaging.EventBus
	authorizationConfig *PaymentAuthorizationConfig
}

// NewPaymentAuthorizationService creates a new payment authorization service instance
func NewPaymentAuthorizationService(
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	milestoneRepo repository.MilestoneRepositoryInterface,
	contractRepo repository.ContractRepositoryInterface,
	userService UserServiceInterface,
	messagingClient messaging.EventBus,
	config *PaymentAuthorizationConfig,
) PaymentAuthorizationServiceInterface {
	return &PaymentAuthorizationService{
		smartChequeRepo:     smartChequeRepo,
		milestoneRepo:       milestoneRepo,
		contractRepo:        contractRepo,
		userService:         userService,
		messagingClient:     messagingClient,
		authorizationConfig: config,
	}
}

// PaymentAuthorizationConfig defines configuration for payment authorization
type PaymentAuthorizationConfig struct {
	// Amount thresholds for different approval levels (in USD equivalent)
	LowAmountThreshold    string `json:"low_amount_threshold"`    // e.g., "5000"
	MediumAmountThreshold string `json:"medium_amount_threshold"` // e.g., "25000"
	HighAmountThreshold   string `json:"high_amount_threshold"`   // e.g., "100000"

	// Required approvals per threshold
	LowAmountApprovals    int `json:"low_amount_approvals"`    // e.g., 1
	MediumAmountApprovals int `json:"medium_amount_approvals"` // e.g., 2
	HighAmountApprovals   int `json:"high_amount_approvals"`   // e.g., 3

	// Auto-approval settings
	AutoApprovalEnabled    bool    `json:"auto_approval_enabled"`
	AutoApprovalThreshold  string  `json:"auto_approval_threshold"`   // e.g., "1000"
	RiskScoreAutoThreshold float64 `json:"risk_score_auto_threshold"` // e.g., 0.3

	// Time lock settings
	TimeLockThreshold string        `json:"time_lock_threshold"` // e.g., "50000"
	TimeLockDuration  time.Duration `json:"time_lock_duration"`  // e.g., 24 hours

	// Risk assessment settings
	RiskScoreThreshold   float64       `json:"risk_score_threshold"`    // e.g., 0.7
	VelocityCheckWindow  time.Duration `json:"velocity_check_window"`   // e.g., 24 hours
	MaxDailyPaymentLimit string        `json:"max_daily_payment_limit"` // e.g., "500000"
}

// Request and response types
type PaymentAuthorizationRequest struct {
	SmartChequeID     string    `json:"smart_check_id" validate:"required"`
	MilestoneID       string    `json:"milestone_id" validate:"required"`
	EnterpriseID      uuid.UUID `json:"enterprise_id" validate:"required"`
	InitiatedByUserID uuid.UUID `json:"initiated_by_user_id" validate:"required"`
	Amount            string    `json:"amount" validate:"required"`
	Currency          string    `json:"currency" validate:"required"`
	Purpose           string    `json:"purpose,omitempty"`
	Reference         string    `json:"reference,omitempty"`
	UrgentPayment     bool      `json:"urgent_payment,omitempty"`
}

type PaymentApprovalRequest struct {
	RequestID    uuid.UUID `json:"request_id" validate:"required"`
	ApproverID   uuid.UUID `json:"approver_id" validate:"required"`
	ApprovalType string    `json:"approval_type" validate:"required"` // "approve", "reject"
	Comments     string    `json:"comments,omitempty"`
	AuthToken    string    `json:"auth_token,omitempty"`
}

type PaymentRejectionRequest struct {
	RequestID  uuid.UUID `json:"request_id" validate:"required"`
	RejectorID uuid.UUID `json:"rejector_id" validate:"required"`
	Reason     string    `json:"reason" validate:"required"`
	Comments   string    `json:"comments,omitempty"`
}

type TimeLockPaymentRequest struct {
	PaymentRequestID uuid.UUID     `json:"payment_request_id" validate:"required"`
	LockDuration     time.Duration `json:"lock_duration" validate:"required"`
	Reason           string        `json:"reason,omitempty"`
}

type PaymentRiskAssessmentRequest struct {
	EnterpriseID  uuid.UUID `json:"enterprise_id" validate:"required"`
	SmartChequeID string    `json:"smart_check_id" validate:"required"`
	Amount        string    `json:"amount" validate:"required"`
	Currency      string    `json:"currency" validate:"required"`
	TimeOfDay     time.Time `json:"time_of_day"`
}

type BulkPaymentApprovalRequest struct {
	ApproverID uuid.UUID   `json:"approver_id" validate:"required"`
	RequestIDs []uuid.UUID `json:"request_ids" validate:"required"`
	Comments   string      `json:"comments,omitempty"`
}

// Response types
type PaymentAuthorization struct {
	ID                uuid.UUID          `json:"id"`
	SmartChequeID     string             `json:"smart_check_id"`
	MilestoneID       string             `json:"milestone_id"`
	EnterpriseID      uuid.UUID          `json:"enterprise_id"`
	InitiatedByUserID uuid.UUID          `json:"initiated_by_user_id"`
	Amount            string             `json:"amount"`
	Currency          string             `json:"currency"`
	Purpose           string             `json:"purpose"`
	Reference         string             `json:"reference"`
	Status            PaymentAuthStatus  `json:"status"`
	RequiredApprovals int                `json:"required_approvals"`
	CurrentApprovals  int                `json:"current_approvals"`
	Approvals         []*PaymentApproval `json:"approvals"`
	RiskScore         float64            `json:"risk_score"`
	TimeLocked        bool               `json:"time_locked"`
	TimeLockExpiresAt *time.Time         `json:"time_lock_expires_at,omitempty"`
	AutoApproved      bool               `json:"auto_approved"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	ProcessedAt       *time.Time         `json:"processed_at,omitempty"`
}

type PaymentApproval struct {
	ID           uuid.UUID           `json:"id"`
	RequestID    uuid.UUID           `json:"request_id"`
	ApproverID   uuid.UUID           `json:"approver_id"`
	ApprovalType PaymentApprovalType `json:"approval_type"`
	Comments     string              `json:"comments"`
	CreatedAt    time.Time           `json:"created_at"`
}

type PaymentAuthorizationStatus struct {
	RequestID           uuid.UUID         `json:"request_id"`
	Status              PaymentAuthStatus `json:"status"`
	RequiredApprovals   int               `json:"required_approvals"`
	CurrentApprovals    int               `json:"current_approvals"`
	PendingApprovers    []uuid.UUID       `json:"pending_approvers"`
	NextSteps           []string          `json:"next_steps"`
	EstimatedCompletion *time.Time        `json:"estimated_completion,omitempty"`
	RiskScore           float64           `json:"risk_score"`
	TimeLocked          bool              `json:"time_locked"`
	TimeLockExpiresAt   *time.Time        `json:"time_lock_expires_at,omitempty"`
}

type TimeLockPayment struct {
	ID           uuid.UUID         `json:"id"`
	RequestID    uuid.UUID         `json:"request_id"`
	LockedAt     time.Time         `json:"locked_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Reason       string            `json:"reason"`
	Status       PaymentLockStatus `json:"status"`
	ReleasedAt   *time.Time        `json:"released_at,omitempty"`
	ReleasedByID *uuid.UUID        `json:"released_by_id,omitempty"`
}

type TimeLockPaymentStatus struct {
	LockID        uuid.UUID         `json:"lock_id"`
	RequestID     uuid.UUID         `json:"request_id"`
	Status        PaymentLockStatus `json:"status"`
	TimeRemaining time.Duration     `json:"time_remaining,omitempty"`
	CanRelease    bool              `json:"can_release"`
	ExpiresAt     time.Time         `json:"expires_at"`
}

type PaymentRiskScore struct {
	RequestID     uuid.UUID `json:"request_id"`
	OverallScore  float64   `json:"overall_score"`
	AmountRisk    float64   `json:"amount_risk"`
	VelocityRisk  float64   `json:"velocity_risk"`
	TimeRisk      float64   `json:"time_risk"`
	RecipientRisk float64   `json:"recipient_risk"`
	RiskFactors   []string  `json:"risk_factors"`
	AssessedAt    time.Time `json:"assessed_at"`
}

type EnterprisePaymentRiskProfile struct {
	EnterpriseID       uuid.UUID `json:"enterprise_id"`
	AverageRiskScore   float64   `json:"average_risk_score"`
	TotalPayments      int64     `json:"total_payments"`
	HighRiskPayments   int64     `json:"high_risk_payments"`
	LastAssessmentDate time.Time `json:"last_assessment_date"`
	RiskTrend          string    `json:"risk_trend"` // "improving", "stable", "degrading"
}

type BulkPaymentApprovalResult struct {
	TotalRequests int                      `json:"total_requests"`
	Successful    int                      `json:"successful"`
	Failed        int                      `json:"failed"`
	Results       []*PaymentApprovalResult `json:"results"`
}

type PaymentApprovalResult struct {
	RequestID uuid.UUID `json:"request_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// Enums
type PaymentAuthStatus string

const (
	PaymentAuthStatusPending    PaymentAuthStatus = "pending"
	PaymentAuthStatusApproved   PaymentAuthStatus = "approved"
	PaymentAuthStatusRejected   PaymentAuthStatus = "rejected"
	PaymentAuthStatusTimeLocked PaymentAuthStatus = "time_locked"
	PaymentAuthStatusProcessing PaymentAuthStatus = "processing"
	PaymentAuthStatusCompleted  PaymentAuthStatus = "completed"
	PaymentAuthStatusCancelled  PaymentAuthStatus = "canceled"
	PaymentAuthStatusExpired    PaymentAuthStatus = "expired"
)

type PaymentApprovalType string

const (
	PaymentApprovalTypeApprove PaymentApprovalType = "approve"
	PaymentApprovalTypeReject  PaymentApprovalType = "reject"
)

type PaymentLockStatus string

const (
	PaymentLockStatusActive   PaymentLockStatus = "active"
	PaymentLockStatusExpired  PaymentLockStatus = "expired"
	PaymentLockStatusReleased PaymentLockStatus = "released"
)

// CreatePaymentAuthorizationRequest creates a new payment authorization request
func (s *PaymentAuthorizationService) CreatePaymentAuthorizationRequest(ctx context.Context, req *PaymentAuthorizationRequest) (*PaymentAuthorization, error) {
	// Validate the request
	if err := s.validatePaymentAuthorizationRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("payment authorization request validation failed: %w", err)
	}

	// Assess payment risk
	riskAssessment, err := s.AssessPaymentRisk(ctx, &PaymentRiskAssessmentRequest{
		EnterpriseID:  req.EnterpriseID,
		SmartChequeID: req.SmartChequeID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		TimeOfDay:     time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("payment risk assessment failed: %w", err)
	}

	// Determine required approvals based on amount and risk
	requiredApprovals := s.calculateRequiredApprovals(req.Amount, riskAssessment.OverallScore)

	// Create the authorization request
	authRequest := &PaymentAuthorization{
		ID:                uuid.New(),
		SmartChequeID:     req.SmartChequeID,
		MilestoneID:       req.MilestoneID,
		EnterpriseID:      req.EnterpriseID,
		InitiatedByUserID: req.InitiatedByUserID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Purpose:           req.Purpose,
		Reference:         req.Reference,
		Status:            PaymentAuthStatusPending,
		RequiredApprovals: requiredApprovals,
		CurrentApprovals:  0,
		Approvals:         []*PaymentApproval{},
		RiskScore:         riskAssessment.OverallScore,
		TimeLocked:        false,
		AutoApproved:      false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Check for auto-approval
	if s.authorizationConfig.AutoApprovalEnabled && s.canAutoApprove(authRequest, riskAssessment) {
		authRequest.Status = PaymentAuthStatusApproved
		authRequest.AutoApproved = true
		authRequest.ProcessedAt = &authRequest.CreatedAt
		authRequest.CurrentApprovals = requiredApprovals

		// Publish auto-approval event
		s.publishPaymentAuthorizationEvent(ctx, "payment.auto_approved", authRequest)
	} else {
		// Check for time lock
		if s.shouldTimeLock(authRequest) {
			lockDuration := s.authorizationConfig.TimeLockDuration
			expiresAt := time.Now().Add(lockDuration)
			authRequest.Status = PaymentAuthStatusTimeLocked
			authRequest.TimeLocked = true
			authRequest.TimeLockExpiresAt = &expiresAt
		}

		// Publish authorization request event
		s.publishPaymentAuthorizationEvent(ctx, "payment.authorization_requested", authRequest)
	}

	// TODO: Persist the authorization request to database
	// For now, we'll just return it

	return authRequest, nil
}

// GetPaymentAuthorizationRequest retrieves a payment authorization request by ID
func (s *PaymentAuthorizationService) GetPaymentAuthorizationRequest(ctx context.Context, requestID uuid.UUID) (*PaymentAuthorization, error) {
	// TODO: Retrieve from database
	// For now, return nil to indicate not found
	return nil, fmt.Errorf("payment authorization request not found: %s", requestID)
}

// GetPendingPaymentAuthorizations retrieves pending payment authorizations for an enterprise
func (s *PaymentAuthorizationService) GetPendingPaymentAuthorizations(ctx context.Context, enterpriseID uuid.UUID) ([]*PaymentAuthorization, error) {
	// TODO: Retrieve from database
	// For now, return empty slice
	return []*PaymentAuthorization{}, nil
}

// GetPaymentAuthorizationsBySmartCheque retrieves payment authorizations for a SmartCheque
func (s *PaymentAuthorizationService) GetPaymentAuthorizationsBySmartCheque(ctx context.Context, smartChequeID string) ([]*PaymentAuthorization, error) {
	// TODO: Retrieve from database
	// For now, return empty slice
	return []*PaymentAuthorization{}, nil
}

// ApprovePayment approves a payment authorization request
func (s *PaymentAuthorizationService) ApprovePayment(ctx context.Context, req *PaymentApprovalRequest) (*PaymentApproval, error) {
	// TODO: Retrieve and update the authorization request
	// For now, create a mock approval
	approval := &PaymentApproval{
		ID:           uuid.New(),
		RequestID:    req.RequestID,
		ApproverID:   req.ApproverID,
		ApprovalType: PaymentApprovalTypeApprove,
		Comments:     req.Comments,
		CreatedAt:    time.Now(),
	}

	// TODO: Update the authorization status and check if fully approved
	// Publish approval event
	s.publishPaymentApprovalEvent(ctx, "payment.approved", approval)

	return approval, nil
}

// RejectPayment rejects a payment authorization request
func (s *PaymentAuthorizationService) RejectPayment(ctx context.Context, req *PaymentRejectionRequest) (*PaymentApproval, error) {
	// TODO: Retrieve and update the authorization request
	// For now, create a mock rejection
	approval := &PaymentApproval{
		ID:           uuid.New(),
		RequestID:    req.RequestID,
		ApproverID:   req.RejectorID,
		ApprovalType: PaymentApprovalTypeReject,
		Comments:     req.Comments,
		CreatedAt:    time.Now(),
	}

	// TODO: Update the authorization status to rejected
	// Publish rejection event
	s.publishPaymentRejectionEvent(ctx, "payment.rejected", approval, req.Reason)

	return approval, nil
}

// CheckPaymentAuthorizationStatus checks the status of a payment authorization
func (s *PaymentAuthorizationService) CheckPaymentAuthorizationStatus(ctx context.Context, requestID uuid.UUID) (*PaymentAuthorizationStatus, error) {
	// TODO: Retrieve from database
	// For now, return a mock status
	status := &PaymentAuthorizationStatus{
		RequestID:         requestID,
		Status:            PaymentAuthStatusPending,
		RequiredApprovals: 2,
		CurrentApprovals:  1,
		PendingApprovers:  []uuid.UUID{uuid.New()},
		NextSteps:         []string{"Awaiting approval from finance manager"},
		RiskScore:         0.4,
		TimeLocked:        false,
	}

	return status, nil
}

// ProcessAutoApproval processes auto-approval for low-risk payments
func (s *PaymentAuthorizationService) ProcessAutoApproval(ctx context.Context, requestID uuid.UUID) error {
	// TODO: Implement auto-approval logic
	return fmt.Errorf("auto-approval not yet implemented")
}

// CreateTimeLockPayment creates a time-locked payment
func (s *PaymentAuthorizationService) CreateTimeLockPayment(ctx context.Context, req *TimeLockPaymentRequest) (*TimeLockPayment, error) {
	// TODO: Implement time lock creation
	lock := &TimeLockPayment{
		ID:        uuid.New(),
		RequestID: req.PaymentRequestID,
		LockedAt:  time.Now(),
		ExpiresAt: time.Now().Add(req.LockDuration),
		Reason:    req.Reason,
		Status:    PaymentLockStatusActive,
	}

	return lock, nil
}

// ReleaseTimeLockPayment releases a time-locked payment
func (s *PaymentAuthorizationService) ReleaseTimeLockPayment(ctx context.Context, lockID uuid.UUID) error {
	// TODO: Implement time lock release
	return fmt.Errorf("time lock release not yet implemented")
}

// GetTimeLockPaymentStatus gets the status of a time-locked payment
func (s *PaymentAuthorizationService) GetTimeLockPaymentStatus(ctx context.Context, lockID uuid.UUID) (*TimeLockPaymentStatus, error) {
	// TODO: Retrieve from database
	status := &TimeLockPaymentStatus{
		LockID:     lockID,
		RequestID:  uuid.New(),
		Status:     PaymentLockStatusActive,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CanRelease: false,
	}

	return status, nil
}

// AssessPaymentRisk assesses the risk of a payment
func (s *PaymentAuthorizationService) AssessPaymentRisk(ctx context.Context, req *PaymentRiskAssessmentRequest) (*PaymentRiskScore, error) {
	// TODO: Implement comprehensive risk assessment
	// For now, return a mock risk score
	riskScore := &PaymentRiskScore{
		RequestID:     uuid.New(),
		OverallScore:  0.3, // Low risk
		AmountRisk:    0.2,
		VelocityRisk:  0.1,
		TimeRisk:      0.0,
		RecipientRisk: 0.2,
		RiskFactors:   []string{"Low amount", "Normal business hours"},
		AssessedAt:    time.Now(),
	}

	return riskScore, nil
}

// GetPaymentRiskProfile gets the risk profile for an enterprise
func (s *PaymentAuthorizationService) GetPaymentRiskProfile(ctx context.Context, enterpriseID uuid.UUID) (*EnterprisePaymentRiskProfile, error) {
	// TODO: Calculate from historical data
	profile := &EnterprisePaymentRiskProfile{
		EnterpriseID:       enterpriseID,
		AverageRiskScore:   0.25,
		TotalPayments:      150,
		HighRiskPayments:   5,
		LastAssessmentDate: time.Now(),
		RiskTrend:          "stable",
	}

	return profile, nil
}

// BulkApprovePayments approves multiple payments at once
func (s *PaymentAuthorizationService) BulkApprovePayments(ctx context.Context, req *BulkPaymentApprovalRequest) (*BulkPaymentApprovalResult, error) {
	// TODO: Implement bulk approval
	result := &BulkPaymentApprovalResult{
		TotalRequests: len(req.RequestIDs),
		Successful:    len(req.RequestIDs),
		Failed:        0,
		Results:       make([]*PaymentApprovalResult, len(req.RequestIDs)),
	}

	for i, requestID := range req.RequestIDs {
		result.Results[i] = &PaymentApprovalResult{
			RequestID: requestID,
			Success:   true,
		}
	}

	return result, nil
}

// GetPaymentAuthorizationHistory gets the authorization history for an enterprise
func (s *PaymentAuthorizationService) GetPaymentAuthorizationHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*PaymentAuthorization, error) {
	// TODO: Retrieve from database
	return []*PaymentAuthorization{}, nil
}

// InitiatePaymentFromMilestone initiates payment authorization from a completed milestone
func (s *PaymentAuthorizationService) InitiatePaymentFromMilestone(ctx context.Context, milestoneID string) (*PaymentAuthorization, error) {
	// Get the milestone details
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Get the associated SmartCheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequesByMilestone(ctx, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SmartCheque for milestone %s: %w", milestoneID, err)
	}

	if smartCheque == nil {
		return nil, fmt.Errorf("no SmartCheque found for milestone %s", milestoneID)
	}

	// Get the contract to determine enterprise
	contract, err := s.contractRepo.GetContractByID(ctx, milestone.ContractID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract %s: %w", milestone.ContractID, err)
	}

	// Create payment authorization request
	req := &PaymentAuthorizationRequest{
		SmartChequeID: smartCheque.ID,
		MilestoneID:   milestoneID,
		EnterpriseID:  uuid.New(), // TODO: Extract from contract parties
		Amount:        fmt.Sprintf("%f", smartCheque.Amount),
		Currency:      string(smartCheque.Currency),
		Purpose:       fmt.Sprintf("Milestone completion: %s", milestone.TriggerConditions),
		Reference:     fmt.Sprintf("Milestone %s for contract %s", milestoneID, contract.ID),
	}

	return s.CreatePaymentAuthorizationRequest(ctx, req)
}

// Helper methods

func (s *PaymentAuthorizationService) validatePaymentAuthorizationRequest(ctx context.Context, req *PaymentAuthorizationRequest) error {
	// Check if context is canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if req.SmartChequeID == "" {
		return fmt.Errorf("smart check ID is required")
	}
	if req.MilestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}
	if req.Amount == "" {
		return fmt.Errorf("amount is required")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	// Validate amount is a valid number
	amount, ok := new(big.Float).SetString(req.Amount)
	if !ok {
		return fmt.Errorf("invalid amount format: %s", req.Amount)
	}
	if amount.Sign() <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	return nil
}

func (s *PaymentAuthorizationService) calculateRequiredApprovals(amount string, riskScore float64) int {
	amountFloat, _ := new(big.Float).SetString(amount)

	// Convert thresholds to big.Float for comparison
	mediumThreshold, _ := new(big.Float).SetString(s.authorizationConfig.MediumAmountThreshold)
	highThreshold, _ := new(big.Float).SetString(s.authorizationConfig.HighAmountThreshold)

	// Determine approval level based on amount
	var baseApprovals int
	switch {
	case amountFloat.Cmp(highThreshold) >= 0:
		baseApprovals = s.authorizationConfig.HighAmountApprovals
	case amountFloat.Cmp(mediumThreshold) >= 0:
		baseApprovals = s.authorizationConfig.MediumAmountApprovals
	default:
		baseApprovals = s.authorizationConfig.LowAmountApprovals
	}

	// Increase approvals for high-risk payments
	if riskScore > s.authorizationConfig.RiskScoreThreshold {
		baseApprovals++
	}

	return baseApprovals
}

func (s *PaymentAuthorizationService) canAutoApprove(auth *PaymentAuthorization, riskScore *PaymentRiskScore) bool {
	if !s.authorizationConfig.AutoApprovalEnabled {
		return false
	}

	// Check amount threshold
	amountFloat, _ := new(big.Float).SetString(auth.Amount)
	autoThreshold, _ := new(big.Float).SetString(s.authorizationConfig.AutoApprovalThreshold)

	if amountFloat.Cmp(autoThreshold) > 0 {
		return false
	}

	// Check risk score threshold
	if riskScore.OverallScore > s.authorizationConfig.RiskScoreAutoThreshold {
		return false
	}

	return true
}

func (s *PaymentAuthorizationService) shouldTimeLock(auth *PaymentAuthorization) bool {
	amountFloat, _ := new(big.Float).SetString(auth.Amount)
	timeLockThreshold, _ := new(big.Float).SetString(s.authorizationConfig.TimeLockThreshold)

	return amountFloat.Cmp(timeLockThreshold) >= 0
}

func (s *PaymentAuthorizationService) publishPaymentAuthorizationEvent(ctx context.Context, eventType string, auth *PaymentAuthorization) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "payment-authorization-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"request_id":     auth.ID,
			"smart_check_id": auth.SmartChequeID,
			"milestone_id":   auth.MilestoneID,
			"enterprise_id":  auth.EnterpriseID,
			"amount":         auth.Amount,
			"currency":       auth.Currency,
			"risk_score":     auth.RiskScore,
			"auto_approved":  auth.AutoApproved,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish payment authorization event: %v\n", err)
	}
}

func (s *PaymentAuthorizationService) publishPaymentApprovalEvent(ctx context.Context, eventType string, approval *PaymentApproval) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "payment-authorization-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"request_id":    approval.RequestID,
			"approver_id":   approval.ApproverID,
			"approval_type": approval.ApprovalType,
			"comments":      approval.Comments,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish payment approval event: %v\n", err)
	}
}

func (s *PaymentAuthorizationService) publishPaymentRejectionEvent(ctx context.Context, eventType string, approval *PaymentApproval, reason string) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "payment-authorization-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"request_id":  approval.RequestID,
			"rejector_id": approval.ApproverID,
			"reason":      reason,
			"comments":    approval.Comments,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish payment rejection event: %v\n", err)
	}
}
