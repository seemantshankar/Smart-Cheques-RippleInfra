package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// PaymentConfirmationServiceInterface defines the interface for payment confirmation operations
type PaymentConfirmationServiceInterface interface {
	// Transaction Confirmation Tracking
	StartTransactionConfirmation(ctx context.Context, transactionID string, paymentExecutionID uuid.UUID) error
	MonitorTransactionConfirmations(ctx context.Context) error
	GetTransactionConfirmationStatus(ctx context.Context, transactionID string) (*TransactionConfirmationStatus, error)

	// Payment Status Updates
	UpdatePaymentStatusFromConfirmation(ctx context.Context, transactionID string) error
	ConfirmPaymentExecution(ctx context.Context, executionID uuid.UUID) error
	FailPaymentExecution(ctx context.Context, executionID uuid.UUID, reason string) error

	// Confirmation History
	GetPaymentConfirmationHistory(ctx context.Context, paymentExecutionID uuid.UUID) ([]*PaymentConfirmationRecord, error)

	// Blockchain Interaction
	GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error)
	WaitForConfirmations(ctx context.Context, transactionID string, requiredConfirmations int) error

	// Configuration
	UpdateConfirmationRequirements(ctx context.Context, transactionID string, requiredConfirmations int) error
}

// PaymentConfirmationService implements the payment confirmation service interface
type PaymentConfirmationService struct {
	paymentExecService  PaymentExecutionServiceInterface
	smartChequeRepo     repository.SmartChequeRepositoryInterface
	transactionRepo     repository.TransactionRepositoryInterface
	xrplService         repository.XRPLServiceInterface
	messagingClient     messaging.EventBus
	confirmationConfig  *PaymentConfirmationConfig
	activeConfirmations map[string]*TransactionConfirmation
	confirmationMutex   sync.RWMutex
}

// NewPaymentConfirmationService creates a new payment confirmation service instance
func NewPaymentConfirmationService(
	paymentExecService PaymentExecutionServiceInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient messaging.EventBus,
	config *PaymentConfirmationConfig,
) PaymentConfirmationServiceInterface {
	service := &PaymentConfirmationService{
		paymentExecService:  paymentExecService,
		smartChequeRepo:     smartChequeRepo,
		transactionRepo:     transactionRepo,
		xrplService:         xrplService,
		messagingClient:     messagingClient,
		confirmationConfig:  config,
		activeConfirmations: make(map[string]*TransactionConfirmation),
	}

	// Start background monitoring if enabled
	if config.EnableBackgroundMonitoring {
		go service.monitorConfirmations()
	}

	return service
}

// PaymentConfirmationConfig defines configuration for payment confirmations
type PaymentConfirmationConfig struct {
	// Confirmation settings
	DefaultRequiredConfirmations int           `json:"default_required_confirmations"`
	ConfirmationTimeout          time.Duration `json:"confirmation_timeout"`
	MonitoringInterval           time.Duration `json:"monitoring_interval"`

	// Background monitoring
	EnableBackgroundMonitoring bool `json:"enable_background_monitoring"`

	// Retry settings
	MaxRetryAttempts int           `json:"max_retry_attempts"`
	RetryDelay       time.Duration `json:"retry_delay"`

	// Notification thresholds
	LowConfirmationThreshold    int `json:"low_confirmation_threshold"`
	MediumConfirmationThreshold int `json:"medium_confirmation_threshold"`
	HighConfirmationThreshold   int `json:"high_confirmation_threshold"`
}

// TransactionConfirmation represents an active transaction confirmation tracking
type TransactionConfirmation struct {
	TransactionID         string                            `json:"transaction_id"`
	PaymentExecutionID    uuid.UUID                         `json:"payment_execution_id"`
	RequiredConfirmations int                               `json:"required_confirmations"`
	CurrentConfirmations  int                               `json:"current_confirmations"`
	Status                TransactionConfirmationStatusType `json:"status"`
	StartedAt             time.Time                         `json:"started_at"`
	LastCheckedAt         time.Time                         `json:"last_checked_at"`
	CompletedAt           *time.Time                        `json:"completed_at,omitempty"`
	Error                 string                            `json:"error,omitempty"`
	CheckHistory          []*ConfirmationCheck              `json:"check_history"`
}

// ConfirmationCheck represents a single confirmation check attempt
type ConfirmationCheck struct {
	ID            uuid.UUID `json:"id"`
	CheckedAt     time.Time `json:"checked_at"`
	Confirmations int       `json:"confirmations"`
	LedgerIndex   uint32    `json:"ledger_index"`
	Success       bool      `json:"success"`
	Error         string    `json:"error,omitempty"`
}

// TransactionConfirmationStatus represents the status of transaction confirmation
type TransactionConfirmationStatusType string

const (
	TransactionConfirmationStatusPending    TransactionConfirmationStatusType = "pending"
	TransactionConfirmationStatusConfirming TransactionConfirmationStatusType = "confirming"
	TransactionConfirmationStatusConfirmed  TransactionConfirmationStatusType = "confirmed"
	TransactionConfirmationStatusFailed     TransactionConfirmationStatusType = "failed"
	TransactionConfirmationStatusExpired    TransactionConfirmationStatusType = "expired"
)

// TransactionConfirmationStatus represents the response for confirmation status
type TransactionConfirmationStatus struct {
	TransactionID         string                            `json:"transaction_id"`
	PaymentExecutionID    uuid.UUID                         `json:"payment_execution_id"`
	RequiredConfirmations int                               `json:"required_confirmations"`
	CurrentConfirmations  int                               `json:"current_confirmations"`
	Status                TransactionConfirmationStatusType `json:"status"`
	StartedAt             time.Time                         `json:"started_at"`
	LastCheckedAt         time.Time                         `json:"last_checked_at"`
	CompletedAt           *time.Time                        `json:"completed_at,omitempty"`
	Error                 string                            `json:"error,omitempty"`
}

// TransactionStatus represents the status of a blockchain transaction
type TransactionStatus struct {
	TransactionID      string    `json:"transaction_id"`
	Status             string    `json:"status"`
	Confirmations      int       `json:"confirmations"`
	LedgerIndex        uint32    `json:"ledger_index"`
	CloseTime          time.Time `json:"close_time"`
	Result             string    `json:"result"`
	Fee                string    `json:"fee"`
	LastLedgerSequence *uint32   `json:"last_ledger_sequence,omitempty"`
}

// PaymentConfirmationRecord represents a historical payment confirmation record
type PaymentConfirmationRecord struct {
	ID                 uuid.UUID `json:"id"`
	PaymentExecutionID uuid.UUID `json:"payment_execution_id"`
	TransactionID      string    `json:"transaction_id"`
	Status             string    `json:"status"`
	Confirmations      int       `json:"confirmations"`
	RecordedAt         time.Time `json:"recorded_at"`
	Error              string    `json:"error,omitempty"`
}

// StartTransactionConfirmation starts monitoring confirmations for a transaction
func (s *PaymentConfirmationService) StartTransactionConfirmation(ctx context.Context, transactionID string, paymentExecutionID uuid.UUID) error {
	log.Printf("Starting confirmation monitoring for transaction: %s", transactionID)

	confirmation := &TransactionConfirmation{
		TransactionID:         transactionID,
		PaymentExecutionID:    paymentExecutionID,
		RequiredConfirmations: s.confirmationConfig.DefaultRequiredConfirmations,
		CurrentConfirmations:  0,
		Status:                TransactionConfirmationStatusPending,
		StartedAt:             time.Now(),
		LastCheckedAt:         time.Now(),
		CheckHistory:          make([]*ConfirmationCheck, 0),
	}

	// Store the confirmation tracking
	s.confirmationMutex.Lock()
	s.activeConfirmations[transactionID] = confirmation
	s.confirmationMutex.Unlock()

	// Publish confirmation started event
	s.publishConfirmationEvent(ctx, "payment.confirmation.started", confirmation, nil)

	return nil
}

// MonitorTransactionConfirmations monitors all active transaction confirmations
func (s *PaymentConfirmationService) MonitorTransactionConfirmations(ctx context.Context) error {
	log.Printf("Starting transaction confirmation monitoring")

	ticker := time.NewTicker(s.confirmationConfig.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context canceled, stopping confirmation monitoring")
			return ctx.Err()
		case <-ticker.C:
			if err := s.checkAllConfirmations(ctx); err != nil {
				log.Printf("Error checking confirmations: %v", err)
			}
		}
	}
}

// GetTransactionConfirmationStatus gets the current confirmation status for a transaction
func (s *PaymentConfirmationService) GetTransactionConfirmationStatus(ctx context.Context, transactionID string) (*TransactionConfirmationStatus, error) {
	s.confirmationMutex.RLock()
	confirmation, exists := s.activeConfirmations[transactionID]
	s.confirmationMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("confirmation tracking not found for transaction: %s", transactionID)
	}

	return &TransactionConfirmationStatus{
		TransactionID:         confirmation.TransactionID,
		PaymentExecutionID:    confirmation.PaymentExecutionID,
		RequiredConfirmations: confirmation.RequiredConfirmations,
		CurrentConfirmations:  confirmation.CurrentConfirmations,
		Status:                confirmation.Status,
		StartedAt:             confirmation.StartedAt,
		LastCheckedAt:         confirmation.LastCheckedAt,
		CompletedAt:           confirmation.CompletedAt,
		Error:                 confirmation.Error,
	}, nil
}

// UpdatePaymentStatusFromConfirmation updates payment status based on confirmation
func (s *PaymentConfirmationService) UpdatePaymentStatusFromConfirmation(ctx context.Context, transactionID string) error {
	log.Printf("Updating payment status for confirmed transaction: %s", transactionID)

	s.confirmationMutex.RLock()
	confirmation, exists := s.activeConfirmations[transactionID]
	s.confirmationMutex.RUnlock()

	if !exists {
		return fmt.Errorf("confirmation tracking not found for transaction: %s", transactionID)
	}

	if confirmation.Status != TransactionConfirmationStatusConfirmed {
		return fmt.Errorf("transaction not yet confirmed: %s", transactionID)
	}

	// Update payment execution status to completed
	if err := s.paymentExecService.UpdatePaymentExecutionStatus(
		ctx,
		confirmation.PaymentExecutionID,
		PaymentExecutionStatusCompleted,
		fmt.Sprintf("Transaction confirmed with %d confirmations", confirmation.CurrentConfirmations),
	); err != nil {
		return fmt.Errorf("failed to update payment execution status: %w", err)
	}

	// Update SmartCheque status if applicable
	if err := s.updateSmartChequeFromConfirmation(ctx, confirmation); err != nil {
		log.Printf("Warning: Failed to update SmartCheque status: %v", err)
	}

	// Publish payment completed event
	s.publishConfirmationEvent(ctx, "payment.confirmed", confirmation, nil)

	log.Printf("Successfully updated payment status for transaction: %s", transactionID)
	return nil
}

// ConfirmPaymentExecution manually confirms a payment execution
func (s *PaymentConfirmationService) ConfirmPaymentExecution(ctx context.Context, executionID uuid.UUID) error {
	return s.paymentExecService.UpdatePaymentExecutionStatus(
		ctx,
		executionID,
		PaymentExecutionStatusCompleted,
		"Manually confirmed by administrator",
	)
}

// FailPaymentExecution marks a payment execution as failed
func (s *PaymentConfirmationService) FailPaymentExecution(ctx context.Context, executionID uuid.UUID, reason string) error {
	return s.paymentExecService.UpdatePaymentExecutionStatus(
		ctx,
		executionID,
		PaymentExecutionStatusFailed,
		reason,
	)
}

// GetPaymentConfirmationHistory gets the confirmation history for a payment execution
func (s *PaymentConfirmationService) GetPaymentConfirmationHistory(ctx context.Context, paymentExecutionID uuid.UUID) ([]*PaymentConfirmationRecord, error) {
	// TODO: Implement database query for confirmation history
	// For now, return empty slice
	return []*PaymentConfirmationRecord{}, nil
}

// GetTransactionStatus gets the current status of a transaction from blockchain
func (s *PaymentConfirmationService) GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionStatus, error) {
	// Get transaction details from XRPL
	// This is a simplified implementation - in production, this would query the XRPL API
	status := &TransactionStatus{
		TransactionID: transactionID,
		Status:        "validated",
		Confirmations: 10, // Mock confirmations
		LedgerIndex:   12345678,
		CloseTime:     time.Now().Add(-5 * time.Minute),
		Result:        "tesSUCCESS",
		Fee:           "0.000012",
	}

	return status, nil
}

// WaitForConfirmations waits for a transaction to reach required confirmations
func (s *PaymentConfirmationService) WaitForConfirmations(ctx context.Context, transactionID string, requiredConfirmations int) error {
	log.Printf("Waiting for %d confirmations on transaction: %s", requiredConfirmations, transactionID)

	// Check confirmations in a loop
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(s.confirmationConfig.ConfirmationTimeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("confirmation timeout exceeded for transaction: %s", transactionID)
		case <-ticker.C:
			status, err := s.GetTransactionStatus(ctx, transactionID)
			if err != nil {
				log.Printf("Error getting transaction status: %v", err)
				continue
			}

			if status.Confirmations >= requiredConfirmations {
				log.Printf("Transaction %s reached %d confirmations", transactionID, status.Confirmations)
				return nil
			}

			log.Printf("Transaction %s has %d/%d confirmations", transactionID, status.Confirmations, requiredConfirmations)
		}
	}
}

// UpdateConfirmationRequirements updates the confirmation requirements for a transaction
func (s *PaymentConfirmationService) UpdateConfirmationRequirements(ctx context.Context, transactionID string, requiredConfirmations int) error {
	s.confirmationMutex.Lock()
	defer s.confirmationMutex.Unlock()

	confirmation, exists := s.activeConfirmations[transactionID]
	if !exists {
		return fmt.Errorf("confirmation tracking not found for transaction: %s", transactionID)
	}

	confirmation.RequiredConfirmations = requiredConfirmations
	log.Printf("Updated confirmation requirements for transaction %s to %d", transactionID, requiredConfirmations)

	return nil
}

// Helper methods

func (s *PaymentConfirmationService) monitorConfirmations() {
	log.Printf("Starting background confirmation monitoring")

	ticker := time.NewTicker(s.confirmationConfig.MonitoringInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := s.checkAllConfirmations(ctx); err != nil {
			log.Printf("Error in confirmation monitoring: %v", err)
		}
	}
}

func (s *PaymentConfirmationService) checkAllConfirmations(ctx context.Context) error {
	s.confirmationMutex.Lock()
	defer s.confirmationMutex.Unlock()

	for transactionID, confirmation := range s.activeConfirmations {
		if err := s.checkSingleConfirmation(ctx, confirmation); err != nil {
			log.Printf("Error checking confirmation for %s: %v", transactionID, err)
			continue
		}

		// Clean up completed confirmations after some time
		if confirmation.Status == TransactionConfirmationStatusConfirmed &&
			time.Since(*confirmation.CompletedAt) > time.Hour {
			delete(s.activeConfirmations, transactionID)
			log.Printf("Cleaned up completed confirmation tracking for: %s", transactionID)
		}

		// Mark expired confirmations
		if time.Since(confirmation.StartedAt) > s.confirmationConfig.ConfirmationTimeout &&
			confirmation.Status == TransactionConfirmationStatusPending {
			confirmation.Status = TransactionConfirmationStatusExpired
			confirmation.Error = "Confirmation timeout exceeded"
			log.Printf("Marked confirmation as expired for: %s", transactionID)
		}
	}

	return nil
}

func (s *PaymentConfirmationService) checkSingleConfirmation(ctx context.Context, confirmation *TransactionConfirmation) error {
	// Get current transaction status
	status, err := s.GetTransactionStatus(ctx, confirmation.TransactionID)
	if err != nil {
		// Record failed check
		s.recordConfirmationCheck(confirmation, 0, 0, false, err.Error())
		return err
	}

	// Update confirmation count
	confirmation.CurrentConfirmations = status.Confirmations
	confirmation.LastCheckedAt = time.Now()

	// Record successful check
	s.recordConfirmationCheck(confirmation, status.Confirmations, status.LedgerIndex, true, "")

	// Update status based on confirmations
	if status.Confirmations >= confirmation.RequiredConfirmations {
		if confirmation.Status != TransactionConfirmationStatusConfirmed {
			confirmation.Status = TransactionConfirmationStatusConfirmed
			now := time.Now()
			confirmation.CompletedAt = &now

			log.Printf("Transaction %s confirmed with %d/%d confirmations",
				confirmation.TransactionID, status.Confirmations, confirmation.RequiredConfirmations)

			// Trigger payment status update
			go func() {
				if err := s.UpdatePaymentStatusFromConfirmation(context.Background(), confirmation.TransactionID); err != nil {
					log.Printf("Error updating payment status: %v", err)
				}
			}()
		}
	} else if confirmation.Status == TransactionConfirmationStatusPending {
		confirmation.Status = TransactionConfirmationStatusConfirming
	}

	return nil
}

func (s *PaymentConfirmationService) recordConfirmationCheck(confirmation *TransactionConfirmation, confirmations int, ledgerIndex uint32, success bool, errorMsg string) {
	check := &ConfirmationCheck{
		ID:            uuid.New(),
		CheckedAt:     time.Now(),
		Confirmations: confirmations,
		LedgerIndex:   ledgerIndex,
		Success:       success,
		Error:         errorMsg,
	}

	confirmation.CheckHistory = append(confirmation.CheckHistory, check)

	// Keep only last 50 checks to prevent memory issues
	if len(confirmation.CheckHistory) > 50 {
		confirmation.CheckHistory = confirmation.CheckHistory[1:]
	}
}

func (s *PaymentConfirmationService) updateSmartChequeFromConfirmation(ctx context.Context, confirmation *TransactionConfirmation) error {
	// Get the payment execution to find the SmartCheque
	_, err := s.paymentExecService.GetPaymentExecutionStatus(ctx, confirmation.PaymentExecutionID)
	if err != nil {
		return fmt.Errorf("failed to get payment execution status: %w", err)
	}

	// TODO: Extract SmartCheque ID from execution and update its status
	// This would require extending the payment execution model to store SmartCheque references

	log.Printf("SmartCheque status update not yet implemented for execution: %s", confirmation.PaymentExecutionID)
	return nil
}

func (s *PaymentConfirmationService) publishConfirmationEvent(ctx context.Context, eventType string, confirmation *TransactionConfirmation, additionalData map[string]interface{}) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "payment-confirmation-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"transaction_id":         confirmation.TransactionID,
			"payment_execution_id":   confirmation.PaymentExecutionID,
			"required_confirmations": confirmation.RequiredConfirmations,
			"current_confirmations":  confirmation.CurrentConfirmations,
			"status":                 string(confirmation.Status),
		},
	}

	for k, v := range additionalData {
		event.Data[k] = v
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		log.Printf("Failed to publish confirmation event: %v", err)
	}
}
