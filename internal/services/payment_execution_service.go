package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// PaymentExecutionServiceInterface defines the interface for payment execution operations
type PaymentExecutionServiceInterface interface {
	// Payment Execution
	ExecutePayment(ctx context.Context, paymentRequestID uuid.UUID) (*PaymentExecutionResult, error)
	ExecutePaymentFromAuthorization(ctx context.Context, authorizationID uuid.UUID) (*PaymentExecutionResult, error)

	// Batch Payment Execution
	ExecuteBulkPayments(ctx context.Context, paymentRequestIDs []uuid.UUID) (*BulkPaymentExecutionResult, error)

	// Transaction Monitoring
	MonitorPaymentExecution(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionStatus, error)
	GetPaymentExecutionHistory(ctx context.Context, paymentRequestID uuid.UUID) ([]*PaymentExecutionRecord, error)

	// Condition Management
	GeneratePaymentFulfillment(ctx context.Context, smartChequeID, milestoneID string) (*PaymentFulfillment, error)
	ValidatePaymentCondition(ctx context.Context, smartChequeID, milestoneID string, condition, fulfillment string) error

	// Retry and Recovery
	RetryFailedPayment(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionResult, error)
	CancelPaymentExecution(ctx context.Context, executionID uuid.UUID) error

	// Status Management
	GetPaymentExecutionStatus(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionStatus, error)
	UpdatePaymentExecutionStatus(ctx context.Context, executionID uuid.UUID, status PaymentExecutionStatusType, details string) error
}

// PaymentExecutionService implements the payment execution service interface
type PaymentExecutionService struct {
	paymentAuthService PaymentAuthorizationServiceInterface
	smartChequeRepo    repository.SmartChequeRepositoryInterface
	transactionRepo    repository.TransactionRepositoryInterface
	xrplService        repository.XRPLServiceInterface
	messagingClient    messaging.EventBus
	executionConfig    *PaymentExecutionConfig
	activeExecutions   map[uuid.UUID]*PaymentExecution
	executionMutex     sync.RWMutex
}

// NewPaymentExecutionService creates a new payment execution service instance
func NewPaymentExecutionService(
	paymentAuthService PaymentAuthorizationServiceInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient messaging.EventBus,
	config *PaymentExecutionConfig,
) PaymentExecutionServiceInterface {
	service := &PaymentExecutionService{
		paymentAuthService: paymentAuthService,
		smartChequeRepo:    smartChequeRepo,
		transactionRepo:    transactionRepo,
		xrplService:        xrplService,
		messagingClient:    messagingClient,
		executionConfig:    config,
		activeExecutions:   make(map[uuid.UUID]*PaymentExecution),
	}

	// Start background monitoring if enabled
	if config.EnableBackgroundMonitoring {
		go service.monitorActiveExecutions()
	}

	return service
}

// PaymentExecutionConfig defines configuration for payment execution
type PaymentExecutionConfig struct {
	// Execution settings
	MaxConcurrentExecutions int           `json:"max_concurrent_executions"`
	ExecutionTimeout        time.Duration `json:"execution_timeout"`
	RetryAttempts           int           `json:"retry_attempts"`
	RetryDelay              time.Duration `json:"retry_delay"`

	// Monitoring settings
	EnableBackgroundMonitoring bool          `json:"enable_background_monitoring"`
	MonitoringInterval         time.Duration `json:"monitoring_interval"`
	ConfirmationTimeout        time.Duration `json:"confirmation_timeout"`

	// Fee settings
	MaxFeePercentage float64 `json:"max_fee_percentage"`
	PriorityFeeBump  int64   `json:"priority_fee_bump"`

	// Security settings
	RequireDoubleCheck bool   `json:"require_double_check"`
	MaxAmountThreshold string `json:"max_amount_threshold"`
}

// PaymentExecution represents an active payment execution
type PaymentExecution struct {
	ID               uuid.UUID                  `json:"id"`
	PaymentRequestID uuid.UUID                  `json:"payment_request_id"`
	Status           PaymentExecutionStatusType `json:"status"`
	StartedAt        time.Time                  `json:"started_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	Attempts         int                        `json:"attempts"`
	LastError        string                     `json:"last_error,omitempty"`
	TransactionID    string                     `json:"transaction_id,omitempty"`
	Fulfillment      *PaymentFulfillment        `json:"fulfillment,omitempty"`
	Steps            []*PaymentExecutionStep    `json:"steps"`
}

// PaymentExecutionStep represents a step in the payment execution process
type PaymentExecutionStep struct {
	ID          uuid.UUID              `json:"id"`
	StepType    string                 `json:"step_type"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentFulfillment contains the condition and fulfillment for payment release
type PaymentFulfillment struct {
	SmartChequeID string    `json:"smart_cheque_id"`
	MilestoneID   string    `json:"milestone_id"`
	Condition     string    `json:"condition"`
	Fulfillment   string    `json:"fulfillment"`
	Sequence      uint32    `json:"sequence"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// PaymentExecutionResult contains the result of a payment execution
type PaymentExecutionResult struct {
	ExecutionID      uuid.UUID                  `json:"execution_id"`
	PaymentRequestID uuid.UUID                  `json:"payment_request_id"`
	Status           PaymentExecutionStatusType `json:"status"`
	TransactionID    string                     `json:"transaction_id,omitempty"`
	Amount           string                     `json:"amount"`
	Currency         string                     `json:"currency"`
	ExecutedAt       time.Time                  `json:"executed_at"`
	Confirmations    int                        `json:"confirmations"`
	Fee              string                     `json:"fee"`
	Error            string                     `json:"error,omitempty"`
	Steps            []*PaymentExecutionStep    `json:"steps"`
}

// BulkPaymentExecutionResult contains the result of bulk payment execution
type BulkPaymentExecutionResult struct {
	TotalRequested int                       `json:"total_requested"`
	Successful     int                       `json:"successful"`
	Failed         int                       `json:"failed"`
	Pending        int                       `json:"pending"`
	Results        []*PaymentExecutionResult `json:"results"`
}

// PaymentExecutionStatus contains the current status of a payment execution
type PaymentExecutionStatus struct {
	ExecutionID         uuid.UUID                  `json:"execution_id"`
	Status              PaymentExecutionStatusType `json:"status"`
	Progress            float64                    `json:"progress"`
	CurrentStep         string                     `json:"current_step"`
	EstimatedCompletion *time.Time                 `json:"estimated_completion,omitempty"`
	LastUpdated         time.Time                  `json:"last_updated"`
	Error               string                     `json:"error,omitempty"`
}

// PaymentExecutionRecord represents a historical payment execution record
type PaymentExecutionRecord struct {
	ID               uuid.UUID                  `json:"id"`
	PaymentRequestID uuid.UUID                  `json:"payment_request_id"`
	Status           PaymentExecutionStatusType `json:"status"`
	TransactionID    string                     `json:"transaction_id,omitempty"`
	ExecutedAt       time.Time                  `json:"executed_at"`
	Error            string                     `json:"error,omitempty"`
}

// PaymentExecutionStatusType represents the status of a payment execution
type PaymentExecutionStatusType string

const (
	PaymentExecutionStatusPending    PaymentExecutionStatusType = "pending"
	PaymentExecutionStatusProcessing PaymentExecutionStatusType = "processing"
	PaymentExecutionStatusConfirming PaymentExecutionStatusType = "confirming"
	PaymentExecutionStatusCompleted  PaymentExecutionStatusType = "completed"
	PaymentExecutionStatusFailed     PaymentExecutionStatusType = "failed"
	PaymentExecutionStatusCancelled  PaymentExecutionStatusType = "cancelled"
	PaymentExecutionStatusRetry      PaymentExecutionStatusType = "retry"
)

// ExecutePayment executes a payment based on a payment request ID
func (s *PaymentExecutionService) ExecutePayment(ctx context.Context, paymentRequestID uuid.UUID) (*PaymentExecutionResult, error) {
	log.Printf("Starting payment execution for payment request: %s", paymentRequestID)

	// Create execution record
	executionID := uuid.New()
	execution := &PaymentExecution{
		ID:               executionID,
		PaymentRequestID: paymentRequestID,
		Status:           PaymentExecutionStatusPending,
		StartedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Attempts:         0,
		Steps:            make([]*PaymentExecutionStep, 0),
	}

	// Store execution
	s.executionMutex.Lock()
	s.activeExecutions[executionID] = execution
	s.executionMutex.Unlock()

	// Execute the payment
	return s.executePaymentInternal(ctx, execution)
}

// ExecutePaymentFromAuthorization executes a payment from an approved authorization
func (s *PaymentExecutionService) ExecutePaymentFromAuthorization(ctx context.Context, authorizationID uuid.UUID) (*PaymentExecutionResult, error) {
	log.Printf("Starting payment execution from authorization: %s", authorizationID)

	// Get the payment authorization
	auth, err := s.paymentAuthService.GetPaymentAuthorizationRequest(ctx, authorizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment authorization: %w", err)
	}

	if auth.Status != PaymentAuthStatusApproved {
		return nil, fmt.Errorf("payment authorization is not approved (status: %s)", auth.Status)
	}

	// Execute the payment
	return s.ExecutePayment(ctx, authorizationID)
}

// executePaymentInternal performs the actual payment execution
func (s *PaymentExecutionService) executePaymentInternal(ctx context.Context, execution *PaymentExecution) (*PaymentExecutionResult, error) {
	execution.Status = PaymentExecutionStatusProcessing
	execution.UpdatedAt = time.Now()

	// Add initial step
	s.addExecutionStep(execution, "validation", "Validating payment request", "in_progress")

	// Get payment authorization
	auth, err := s.paymentAuthService.GetPaymentAuthorizationRequest(ctx, execution.PaymentRequestID)
	if err != nil {
		s.updateExecutionStep(execution, "validation", "failed", err.Error())
		return s.createFailedResult(execution, fmt.Errorf("failed to get payment authorization: %w", err))
	}

	// Validate authorization status
	if auth.Status != PaymentAuthStatusApproved {
		s.updateExecutionStep(execution, "validation", "failed", "Payment not approved")
		return s.createFailedResult(execution, fmt.Errorf("payment authorization not approved"))
	}

	s.updateExecutionStep(execution, "validation", "completed", "Payment authorization validated")

	// Add fulfillment generation step
	s.addExecutionStep(execution, "fulfillment_generation", "Generating payment fulfillment", "in_progress")

	// Generate fulfillment
	fulfillment, err := s.GeneratePaymentFulfillment(ctx, auth.SmartChequeID, auth.MilestoneID)
	if err != nil {
		s.updateExecutionStep(execution, "fulfillment_generation", "failed", err.Error())
		return s.createFailedResult(execution, fmt.Errorf("failed to generate fulfillment: %w", err))
	}

	execution.Fulfillment = fulfillment
	s.updateExecutionStep(execution, "fulfillment_generation", "completed", "Fulfillment generated successfully")

	// Add XRPL transaction step
	s.addExecutionStep(execution, "xrpl_transaction", "Executing XRPL escrow finish", "in_progress")

	// Execute XRPL escrow finish
	transactionResult, err := s.executeXRPLEscrowFinish(ctx, auth, fulfillment)
	if err != nil {
		s.updateExecutionStep(execution, "xrpl_transaction", "failed", err.Error())
		return s.createFailedResult(execution, fmt.Errorf("failed to execute XRPL transaction: %w", err))
	}

	execution.TransactionID = transactionResult.TransactionID
	s.updateExecutionStep(execution, "xrpl_transaction", "completed", fmt.Sprintf("Transaction submitted: %s", transactionResult.TransactionID))

	// Add confirmation step
	s.addExecutionStep(execution, "confirmation", "Waiting for blockchain confirmation", "in_progress")

	// Update execution status
	execution.Status = PaymentExecutionStatusConfirming
	execution.UpdatedAt = time.Now()

	// Create result
	result := &PaymentExecutionResult{
		ExecutionID:      execution.ID,
		PaymentRequestID: execution.PaymentRequestID,
		Status:           PaymentExecutionStatusConfirming,
		TransactionID:    transactionResult.TransactionID,
		Amount:           auth.Amount,
		Currency:         auth.Currency,
		ExecutedAt:       time.Now(),
		Confirmations:    0,
		Fee:              "0.00001", // Placeholder fee
		Steps:            execution.Steps,
	}

	// Publish payment execution started event
	s.publishPaymentExecutionEvent(ctx, "payment.execution.started", execution, result)

	return result, nil
}

// executeXRPLEscrowFinish executes the XRPL escrow finish transaction
func (s *PaymentExecutionService) executeXRPLEscrowFinish(ctx context.Context, auth *PaymentAuthorization, fulfillment *PaymentFulfillment) (*xrpl.TransactionResult, error) {
	// Get the SmartCheque to find escrow details
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, auth.SmartChequeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
	}

	if smartCheque.EscrowAddress == "" {
		return nil, fmt.Errorf("smart cheque has no escrow address")
	}

	// Use the escrow address as both payee and owner for the escrow finish
	// In a real implementation, these would be determined from the escrow creation
	payeeAddress := smartCheque.EscrowAddress
	ownerAddress := smartCheque.EscrowAddress

	// Execute the escrow finish
	result, err := s.xrplService.CompleteSmartChequeMilestone(
		payeeAddress,
		ownerAddress,
		fulfillment.Sequence,
		fulfillment.Condition,
		fulfillment.Fulfillment,
	)
	if err != nil {
		return nil, fmt.Errorf("XRPL escrow finish failed: %w", err)
	}

	return result, nil
}

// GeneratePaymentFulfillment generates the condition and fulfillment for payment release
func (s *PaymentExecutionService) GeneratePaymentFulfillment(ctx context.Context, smartChequeID, milestoneID string) (*PaymentFulfillment, error) {
	// Validate that the SmartCheque exists
	_, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
	}

	// Generate condition and fulfillment based on milestone verification
	// This is a simplified implementation - in production, this would use
	// cryptographic proofs and milestone verification data
	secret := fmt.Sprintf("payment_%s_%s_%d", smartChequeID, milestoneID, time.Now().Unix())
	condition, fulfillment, err := s.xrplService.GenerateCondition(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate condition: %w", err)
	}

	// Use sequence number 1 as default - in production this would be retrieved
	// from the original escrow creation transaction
	sequence := uint32(1)

	return &PaymentFulfillment{
		SmartChequeID: smartChequeID,
		MilestoneID:   milestoneID,
		Condition:     condition,
		Fulfillment:   fulfillment,
		Sequence:      sequence,
		GeneratedAt:   time.Now(),
	}, nil
}

// ValidatePaymentCondition validates that a payment condition can be fulfilled
func (s *PaymentExecutionService) ValidatePaymentCondition(ctx context.Context, smartChequeID, milestoneID string, condition, fulfillment string) error {
	// Get the SmartCheque and milestone
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}

	// Find the milestone
	var milestone *models.Milestone
	for _, m := range smartCheque.Milestones {
		if m.ID == milestoneID {
			milestone = &m
			break
		}
	}

	if milestone == nil {
		return fmt.Errorf("milestone not found in smart cheque")
	}

	// Validate milestone is completed
	if milestone.Status != models.MilestoneStatusVerified {
		return fmt.Errorf("milestone is not verified (status: %s)", milestone.Status)
	}

	// Additional validation could include:
	// - Checking condition format
	// - Verifying fulfillment matches condition
	// - Validating cryptographic proofs
	// - Checking authorization status

	return nil
}

// MonitorPaymentExecution monitors the status of a payment execution
func (s *PaymentExecutionService) MonitorPaymentExecution(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionStatus, error) {
	s.executionMutex.RLock()
	execution, exists := s.activeExecutions[executionID]
	s.executionMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// Calculate progress based on completed steps
	totalSteps := len(execution.Steps)
	completedSteps := 0
	var currentStep string

	for _, step := range execution.Steps {
		switch step.Status {
		case "completed":
			completedSteps++
		case "in_progress":
			currentStep = step.StepType
		}
	}

	progress := float64(completedSteps) / float64(totalSteps) * 100

	// Estimate completion time
	var estimatedCompletion *time.Time
	if execution.Status == PaymentExecutionStatusConfirming {
		completion := time.Now().Add(s.executionConfig.ConfirmationTimeout)
		estimatedCompletion = &completion
	}

	return &PaymentExecutionStatus{
		ExecutionID:         executionID,
		Status:              execution.Status,
		Progress:            progress,
		CurrentStep:         currentStep,
		EstimatedCompletion: estimatedCompletion,
		LastUpdated:         execution.UpdatedAt,
		Error:               execution.LastError,
	}, nil
}

// GetPaymentExecutionHistory gets the execution history for a payment request
func (s *PaymentExecutionService) GetPaymentExecutionHistory(ctx context.Context, paymentRequestID uuid.UUID) ([]*PaymentExecutionRecord, error) {
	// TODO: Implement database query for execution history
	// For now, return empty slice
	return []*PaymentExecutionRecord{}, nil
}

// ExecuteBulkPayments executes multiple payments in batch
func (s *PaymentExecutionService) ExecuteBulkPayments(ctx context.Context, paymentRequestIDs []uuid.UUID) (*BulkPaymentExecutionResult, error) {
	result := &BulkPaymentExecutionResult{
		TotalRequested: len(paymentRequestIDs),
		Successful:     0,
		Failed:         0,
		Pending:        0,
		Results:        make([]*PaymentExecutionResult, 0, len(paymentRequestIDs)),
	}

	// Execute payments with concurrency control
	semaphore := make(chan struct{}, s.executionConfig.MaxConcurrentExecutions)

	for _, requestID := range paymentRequestIDs {
		go func(id uuid.UUID) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			execResult, err := s.ExecutePayment(ctx, id)
			if err != nil {
				result.Failed++
				result.Results = append(result.Results, &PaymentExecutionResult{
					PaymentRequestID: id,
					Status:           PaymentExecutionStatusFailed,
					Error:            err.Error(),
				})
			} else {
				result.Successful++
				result.Results = append(result.Results, execResult)
			}
		}(requestID)
	}

	// Wait for all executions to complete
	for i := 0; i < s.executionConfig.MaxConcurrentExecutions; i++ {
		semaphore <- struct{}{}
	}

	result.Pending = result.TotalRequested - result.Successful - result.Failed

	return result, nil
}

// RetryFailedPayment retries a failed payment execution
func (s *PaymentExecutionService) RetryFailedPayment(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionResult, error) {
	s.executionMutex.RLock()
	execution, exists := s.activeExecutions[executionID]
	s.executionMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	if execution.Status != PaymentExecutionStatusFailed {
		return nil, fmt.Errorf("execution is not in failed status: %s", execution.Status)
	}

	if execution.Attempts >= s.executionConfig.RetryAttempts {
		return nil, fmt.Errorf("maximum retry attempts exceeded (%d)", s.executionConfig.RetryAttempts)
	}

	// Reset execution for retry
	execution.Status = PaymentExecutionStatusPending
	execution.Attempts++
	execution.LastError = ""
	execution.UpdatedAt = time.Now()

	// Retry after delay
	time.Sleep(s.executionConfig.RetryDelay)

	return s.executePaymentInternal(ctx, execution)
}

// CancelPaymentExecution cancels a payment execution
func (s *PaymentExecutionService) CancelPaymentExecution(ctx context.Context, executionID uuid.UUID) error {
	s.executionMutex.Lock()
	defer s.executionMutex.Unlock()

	execution, exists := s.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	if execution.Status == PaymentExecutionStatusCompleted || execution.Status == PaymentExecutionStatusFailed {
		return fmt.Errorf("cannot cancel execution in status: %s", execution.Status)
	}

	execution.Status = PaymentExecutionStatusCancelled
	execution.UpdatedAt = time.Now()
	execution.LastError = "Execution cancelled by user"

	// Publish cancellation event
	s.publishPaymentExecutionEvent(ctx, "payment.execution.cancelled", execution, nil)

	return nil
}

// GetPaymentExecutionStatus gets the current status of a payment execution
func (s *PaymentExecutionService) GetPaymentExecutionStatus(ctx context.Context, executionID uuid.UUID) (*PaymentExecutionStatus, error) {
	return s.MonitorPaymentExecution(ctx, executionID)
}

// UpdatePaymentExecutionStatus updates the status of a payment execution
func (s *PaymentExecutionService) UpdatePaymentExecutionStatus(ctx context.Context, executionID uuid.UUID, status PaymentExecutionStatusType, details string) error {
	s.executionMutex.Lock()
	defer s.executionMutex.Unlock()

	execution, exists := s.activeExecutions[executionID]
	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	execution.Status = status
	execution.UpdatedAt = time.Now()

	if details != "" {
		execution.LastError = details
	}

	return nil
}

// Helper methods

func (s *PaymentExecutionService) addExecutionStep(execution *PaymentExecution, stepType, description, status string) {
	step := &PaymentExecutionStep{
		ID:          uuid.New(),
		StepType:    stepType,
		Description: description,
		Status:      status,
		StartedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	execution.Steps = append(execution.Steps, step)
}

func (s *PaymentExecutionService) updateExecutionStep(execution *PaymentExecution, stepType, status, errorMsg string) {
	for _, step := range execution.Steps {
		if step.StepType == stepType {
			step.Status = status
			if status == "completed" || status == "failed" {
				now := time.Now()
				step.CompletedAt = &now
			}
			if errorMsg != "" {
				step.Error = errorMsg
			}
			break
		}
	}
}

func (s *PaymentExecutionService) createFailedResult(execution *PaymentExecution, err error) (*PaymentExecutionResult, error) {
	execution.Status = PaymentExecutionStatusFailed
	execution.LastError = err.Error()
	execution.UpdatedAt = time.Now()

	return &PaymentExecutionResult{
		ExecutionID:      execution.ID,
		PaymentRequestID: execution.PaymentRequestID,
		Status:           PaymentExecutionStatusFailed,
		Error:            err.Error(),
		ExecutedAt:       time.Now(),
		Steps:            execution.Steps,
	}, err
}

func (s *PaymentExecutionService) monitorActiveExecutions() {
	ticker := time.NewTicker(s.executionConfig.MonitoringInterval)
	defer ticker.Stop()

	log.Printf("Starting payment execution monitoring")

	for range ticker.C {
		s.checkExecutionStatuses()
	}
}

func (s *PaymentExecutionService) checkExecutionStatuses() {
	s.executionMutex.Lock()
	defer s.executionMutex.Unlock()

	for executionID, execution := range s.activeExecutions {
		if execution.Status == PaymentExecutionStatusConfirming {
			// Check if transaction is confirmed
			// TODO: Implement blockchain confirmation checking
			if time.Since(execution.UpdatedAt) > s.executionConfig.ConfirmationTimeout {
				execution.Status = PaymentExecutionStatusCompleted
				execution.UpdatedAt = time.Now()
				log.Printf("Payment execution completed: %s", executionID)
			}
		}

		// Clean up old executions
		if time.Since(execution.UpdatedAt) > 24*time.Hour {
			delete(s.activeExecutions, executionID)
			log.Printf("Cleaned up old execution: %s", executionID)
		}
	}
}

func (s *PaymentExecutionService) publishPaymentExecutionEvent(ctx context.Context, eventType string, execution *PaymentExecution, result *PaymentExecutionResult) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "payment-execution-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"execution_id":       execution.ID,
			"payment_request_id": execution.PaymentRequestID,
			"status":             string(execution.Status),
			"attempts":           execution.Attempts,
		},
	}

	if result != nil {
		event.Data["transaction_id"] = result.TransactionID
		event.Data["amount"] = result.Amount
		event.Data["currency"] = result.Currency
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		log.Printf("Failed to publish payment execution event: %v", err)
	}
}
