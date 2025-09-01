package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// SmartChequeXRPLServiceInterface defines the interface for Smart Cheque XRPL integration operations
type SmartChequeXRPLServiceInterface interface {
	// CreateEscrowForSmartCheque creates an XRPL escrow for a Smart Cheque
	CreateEscrowForSmartCheque(ctx context.Context, smartChequeID string, payerWalletAddress, payeeWalletAddress string) error

	// CompleteMilestonePayment completes a milestone payment by finishing the XRPL escrow
	CompleteMilestonePayment(ctx context.Context, smartChequeID, milestoneID string) error

	// CancelSmartChequeEscrow cancels the XRPL escrow for a Smart Cheque
	CancelSmartChequeEscrow(ctx context.Context, smartChequeID string) error

	// CancelSmartChequeEscrowWithReason cancels escrow with specific reason and optional notes
	CancelSmartChequeEscrowWithReason(ctx context.Context, smartChequeID, reason, notes string) error

	// PartialRefundEscrow performs a partial refund based on completed milestones
	PartialRefundEscrow(ctx context.Context, smartChequeID string, refundPercentage float64) error

	// SyncEscrowStatus syncs the XRPL escrow status with the Smart Cheque status
	SyncEscrowStatus(ctx context.Context, smartChequeID string) error

	// MonitorEscrowStatus continuously monitors escrow status and updates Smart Cheque accordingly
	MonitorEscrowStatus(ctx context.Context, smartChequeID string, interval time.Duration) error

	// GetEscrowHealthStatus provides comprehensive health status of the escrow
	GetEscrowHealthStatus(ctx context.Context, smartChequeID string) (*EscrowHealthStatus, error)

	// GetXRPLTransactionHistory retrieves XRPL transaction history for a Smart Cheque
	GetXRPLTransactionHistory(ctx context.Context, smartChequeID string) ([]*models.Transaction, error)
}

// smartChequeXRPLService implements SmartChequeXRPLServiceInterface
type smartChequeXRPLService struct {
	smartChequeRepo repository.SmartChequeRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	xrplService     repository.XRPLServiceInterface
	milestoneRepo   repository.MilestoneRepositoryInterface
}

// NewSmartChequeXRPLService creates a new Smart Cheque XRPL service
func NewSmartChequeXRPLService(
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	milestoneRepo repository.MilestoneRepositoryInterface,
) SmartChequeXRPLServiceInterface {
	return &smartChequeXRPLService{
		smartChequeRepo: smartChequeRepo,
		transactionRepo: transactionRepo,
		xrplService:     xrplService,
		milestoneRepo:   milestoneRepo,
	}
}

// CreateEscrowForSmartCheque creates an XRPL escrow for a Smart Cheque
func (s *smartChequeXRPLService) CreateEscrowForSmartCheque(ctx context.Context, smartChequeID string, payerWalletAddress, payeeWalletAddress string) error {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	// Validate wallet addresses
	if !s.xrplService.ValidateAddress(payerWalletAddress) {
		return fmt.Errorf("invalid payer wallet address: %s", payerWalletAddress)
	}
	if !s.xrplService.ValidateAddress(payeeWalletAddress) {
		return fmt.Errorf("invalid payee wallet address: %s", payeeWalletAddress)
	}

	// Create the XRPL escrow with milestone-based conditions
	result, fulfillment, err := s.xrplService.CreateSmartChequeEscrowWithMilestones(
		payerWalletAddress,
		payeeWalletAddress,
		smartCheque.Amount,
		string(smartCheque.Currency),
		smartCheque.Milestones,
	)
	if err != nil {
		return fmt.Errorf("failed to create XRPL escrow: %w", err)
	}

	// Update the Smart Cheque with escrow information
	smartCheque.EscrowAddress = result.TransactionID
	smartCheque.Status = models.SmartChequeStatusLocked
	smartCheque.UpdatedAt = time.Now()

	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque with escrow info: %w", err)
	}

	// Create a transaction record for tracking
	transaction := models.NewTransaction(
		models.TransactionTypeEscrowCreate,
		payerWalletAddress,
		payeeWalletAddress,
		fmt.Sprintf("%f", smartCheque.Amount),
		string(smartCheque.Currency),
		smartCheque.PayerID,
		smartCheque.PayerID, // Using payer ID as user ID for now
	)

	// Set XRPL-specific fields
	transaction.SmartChequeID = &smartChequeID
	transaction.TransactionHash = result.TransactionID
	transaction.Fulfillment = fulfillment
	transaction.Status = models.TransactionStatusConfirmed
	now := time.Now()
	transaction.ConfirmedAt = &now

	// Save the transaction
	if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
		log.Printf("Warning: Failed to save transaction record: %v", err)
	}

	log.Printf("Created XRPL escrow for Smart Cheque %s with transaction ID %s", smartChequeID, result.TransactionID)
	return nil
}

// CompleteMilestonePayment completes a milestone payment by finishing the XRPL escrow
func (s *smartChequeXRPLService) CompleteMilestonePayment(ctx context.Context, smartChequeID, milestoneID string) error {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	// Check if escrow address exists
	if smartCheque.EscrowAddress == "" {
		return fmt.Errorf("smart cheque has no escrow address")
	}

	// Find the milestone in the smart cheque
	var milestone *models.Milestone
	for _, m := range smartCheque.Milestones {
		if m.ID == milestoneID {
			milestone = &m
			break
		}
	}

	if milestone == nil {
		return fmt.Errorf("milestone not found in smart cheque: %s", milestoneID)
	}

	// For now, we'll use a simple approach to complete the milestone
	// In a real implementation, this would involve more complex logic
	// including retrieving the fulfillment from the original escrow creation

	// Generate condition and fulfillment (in a real implementation, these would be retrieved)
	milestoneSecret := fmt.Sprintf("smartcheque_%s_secret_%d", smartChequeID, time.Now().Unix())
	condition, fulfillment, err := s.xrplService.GenerateCondition(milestoneSecret)
	if err != nil {
		return fmt.Errorf("failed to generate condition: %w", err)
	}

	// Complete the XRPL escrow
	// Note: This is a simplified implementation. In reality, we would need the actual
	// sequence number and other details from the original escrow creation
	result, err := s.xrplService.CompleteSmartChequeMilestone(
		smartCheque.EscrowAddress, // Using escrow address as payee for this example
		smartCheque.EscrowAddress, // Using escrow address as owner for this example
		1,                         // Sequence number - would need to be retrieved from original transaction
		condition,
		fulfillment,
	)
	if err != nil {
		return fmt.Errorf("failed to complete XRPL escrow: %w", err)
	}

	// Update milestone status
	milestone.Status = models.MilestoneStatusVerified
	now := time.Now()
	milestone.CompletedAt = &now
	milestone.UpdatedAt = now

	// Update Smart Cheque status if all milestones are completed
	allCompleted := true
	for i, m := range smartCheque.Milestones {
		if m.ID == milestoneID {
			smartCheque.Milestones[i].Status = models.MilestoneStatusVerified
			smartCheque.Milestones[i].CompletedAt = &now
		}
		if smartCheque.Milestones[i].Status != models.MilestoneStatusVerified {
			allCompleted = false
		}
	}

	if allCompleted {
		smartCheque.Status = models.SmartChequeStatusCompleted
	}

	smartCheque.UpdatedAt = time.Now()
	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	// Create a transaction record for tracking
	transaction := models.NewTransaction(
		models.TransactionTypeEscrowFinish,
		smartCheque.EscrowAddress,
		smartCheque.EscrowAddress,
		fmt.Sprintf("%f", milestone.Amount),
		string(smartCheque.Currency),
		smartCheque.PayerID,
		smartCheque.PayerID, // Using payer ID as user ID for now
	)

	// Set XRPL-specific fields
	transaction.SmartChequeID = &smartChequeID
	transaction.MilestoneID = &milestoneID
	transaction.TransactionHash = result.TransactionID
	transaction.Status = models.TransactionStatusConfirmed
	now = time.Now()
	transaction.ConfirmedAt = &now

	// Save the transaction
	if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
		log.Printf("Warning: Failed to save transaction record: %v", err)
	}

	log.Printf("Completed milestone payment for Smart Cheque %s, milestone %s with transaction ID %s",
		smartChequeID, milestoneID, result.TransactionID)
	return nil
}

// CancelSmartChequeEscrow cancels the XRPL escrow for a Smart Cheque with refund calculation
func (s *smartChequeXRPLService) CancelSmartChequeEscrow(ctx context.Context, smartChequeID string) error {
	return s.CancelSmartChequeEscrowWithReason(ctx, smartChequeID, CancellationReasonMutualAgreement, "")
}

// CancelSmartChequeEscrowWithReason cancels escrow with specific reason and optional notes
func (s *smartChequeXRPLService) CancelSmartChequeEscrowWithReason(ctx context.Context, smartChequeID, reason, notes string) error {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	// Check if escrow address exists
	if smartCheque.EscrowAddress == "" {
		return fmt.Errorf("smart cheque has no escrow address")
	}

	// Validate escrow can be cancelled
	if err := s.validateEscrowCancellation(smartCheque); err != nil {
		return fmt.Errorf("escrow cancellation validation failed: %w", err)
	}

	// Calculate refund amount based on completed milestones
	refundAmount := s.calculateRefundAmount(smartCheque)

	// Cancel the XRPL escrow
	result, err := s.xrplService.CancelSmartCheque(
		smartCheque.EscrowAddress,
		smartCheque.EscrowAddress,
		1, // Sequence number - would need to be retrieved from original transaction
	)
	if err != nil {
		return fmt.Errorf("failed to cancel XRPL escrow: %w", err)
	}

	// Update Smart Cheque status based on cancellation reason
	newStatus := s.determineStatusAfterCancellation(smartCheque, reason)
	smartCheque.Status = newStatus
	smartCheque.UpdatedAt = time.Now()

	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	// Create transaction records for cancellation and refund
	if err := s.createCancellationTransactions(ctx, smartCheque, result.TransactionID, refundAmount, reason, notes); err != nil {
		log.Printf("Warning: Failed to create cancellation transaction records: %v", err)
	}

	log.Printf("Cancelled XRPL escrow for Smart Cheque %s with reason '%s', refund amount: %f, transaction ID: %s",
		smartChequeID, reason, refundAmount, result.TransactionID)
	return nil
}

// PartialRefundEscrow performs a partial refund based on completed milestones
func (s *smartChequeXRPLService) PartialRefundEscrow(ctx context.Context, smartChequeID string, refundPercentage float64) error {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	// Validate partial refund is possible
	if err := s.validatePartialRefund(smartCheque, refundPercentage); err != nil {
		return fmt.Errorf("partial refund validation failed: %w", err)
	}

	// Calculate refund amount
	refundAmount := smartCheque.Amount * (refundPercentage / 100)

	// In a real implementation, this would involve:
	// 1. Finishing the escrow with partial fulfillment
	// 2. Creating a new escrow for the remaining amount
	// For now, we'll simulate this with a full cancellation and note the partial nature

	result, err := s.xrplService.CancelSmartCheque(
		smartCheque.EscrowAddress,
		smartCheque.EscrowAddress,
		1,
	)
	if err != nil {
		return fmt.Errorf("failed to perform partial refund: %w", err)
	}

	// Update Smart Cheque status
	smartCheque.Status = models.SmartChequeStatusDisputed
	smartCheque.UpdatedAt = time.Now()

	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	// Create partial refund transaction record
	if err := s.createPartialRefundTransaction(ctx, smartCheque, result.TransactionID, refundAmount, refundPercentage); err != nil {
		log.Printf("Warning: Failed to create partial refund transaction record: %v", err)
	}

	log.Printf("Performed partial refund for Smart Cheque %s: %f%% (%f %s), transaction ID: %s",
		smartChequeID, refundPercentage, refundAmount, smartCheque.Currency, result.TransactionID)
	return nil
}

// validateEscrowCancellation validates if escrow can be cancelled
func (s *smartChequeXRPLService) validateEscrowCancellation(smartCheque *models.SmartCheque) error {
	// Check current status
	if smartCheque.Status == models.SmartChequeStatusCompleted {
		return fmt.Errorf("cannot cancel completed smart cheque")
	}

	// Check if escrow has expired
	// In a real implementation, this would check the escrow's cancel_after time
	// For now, we'll allow cancellation if the Smart Cheque is in appropriate states

	validStatuses := []models.SmartChequeStatus{
		models.SmartChequeStatusLocked,
		models.SmartChequeStatusInProgress,
		models.SmartChequeStatusDisputed,
	}

	for _, status := range validStatuses {
		if smartCheque.Status == status {
			return nil
		}
	}

	return fmt.Errorf("smart cheque status %s does not allow cancellation", smartCheque.Status)
}

// validatePartialRefund validates if partial refund is possible
func (s *smartChequeXRPLService) validatePartialRefund(smartCheque *models.SmartCheque, refundPercentage float64) error {
	if refundPercentage <= 0 || refundPercentage > 100 {
		return fmt.Errorf("invalid refund percentage: %f", refundPercentage)
	}

	if smartCheque.Status == models.SmartChequeStatusCompleted {
		return fmt.Errorf("cannot perform partial refund on completed smart cheque")
	}

	// Check if some milestones are completed to justify partial refund
	completedMilestones := 0
	for _, milestone := range smartCheque.Milestones {
		if milestone.Status == models.MilestoneStatusVerified {
			completedMilestones++
		}
	}

	if completedMilestones == 0 {
		return fmt.Errorf("no completed milestones found for partial refund")
	}

	if completedMilestones == len(smartCheque.Milestones) {
		return fmt.Errorf("all milestones completed - use full completion instead")
	}

	return nil
}

// calculateRefundAmount calculates the refund amount based on completed milestones
func (s *smartChequeXRPLService) calculateRefundAmount(smartCheque *models.SmartCheque) float64 {
	completedAmount := 0.0
	totalAmount := 0.0

	for _, milestone := range smartCheque.Milestones {
		totalAmount += milestone.Amount
		if milestone.Status == models.MilestoneStatusVerified {
			completedAmount += milestone.Amount
		}
	}

	// If no milestones are defined or amounts don't match, return full amount
	if totalAmount == 0 || totalAmount != smartCheque.Amount {
		return smartCheque.Amount
	}

	// Return amount for completed milestones
	return completedAmount
}

// determineStatusAfterCancellation determines the appropriate status after cancellation
func (s *smartChequeXRPLService) determineStatusAfterCancellation(smartCheque *models.SmartCheque, reason string) models.SmartChequeStatus {
	switch reason {
	case CancellationReasonTimeout:
		return models.SmartChequeStatusDisputed
	case CancellationReasonDispute:
		return models.SmartChequeStatusDisputed
	case CancellationReasonMutualAgreement:
		// Check if any progress was made
		hasCompletedMilestones := false
		for _, milestone := range smartCheque.Milestones {
			if milestone.Status == models.MilestoneStatusVerified {
				hasCompletedMilestones = true
				break
			}
		}
		if hasCompletedMilestones {
			return models.SmartChequeStatusDisputed
		}
		return models.SmartChequeStatusDisputed
	default:
		return models.SmartChequeStatusDisputed
	}
}

// Cancellation reason constants
const (
	CancellationReasonTimeout         = "timeout"
	CancellationReasonDispute         = "dispute"
	CancellationReasonMutualAgreement = "mutual_agreement"
	CancellationReasonExpired         = "expired"
)

// createCancellationTransactions creates transaction records for cancellation
func (s *smartChequeXRPLService) createCancellationTransactions(ctx context.Context, smartCheque *models.SmartCheque, txHash string, refundAmount float64, reason, notes string) error {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create cancellation transaction
	cancelTx := models.NewTransaction(
		models.TransactionTypeEscrowCancel,
		smartCheque.EscrowAddress,
		smartCheque.PayerID, // Refund goes back to payer
		fmt.Sprintf("%f", refundAmount),
		string(smartCheque.Currency),
		smartCheque.PayerID,
		smartCheque.PayerID,
	)

	cancelTx.SmartChequeID = &smartCheque.ID
	cancelTx.TransactionHash = txHash
	cancelTx.Status = models.TransactionStatusConfirmed
	cancelTx.Metadata = map[string]interface{}{
		"cancellation_reason": reason,
		"cancellation_notes":  notes,
		"refund_amount":       refundAmount,
	}
	now := time.Now()
	cancelTx.ConfirmedAt = &now

	return s.transactionRepo.CreateTransaction(cancelTx)
}

// createPartialRefundTransaction creates transaction record for partial refund
func (s *smartChequeXRPLService) createPartialRefundTransaction(ctx context.Context, smartCheque *models.SmartCheque, txHash string, refundAmount, refundPercentage float64) error {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create partial refund transaction
	refundTx := models.NewTransaction(
		models.TransactionTypeEscrowCancel,
		smartCheque.EscrowAddress,
		smartCheque.PayerID,
		fmt.Sprintf("%f", refundAmount),
		string(smartCheque.Currency),
		smartCheque.PayerID,
		smartCheque.PayerID,
	)

	refundTx.SmartChequeID = &smartCheque.ID
	refundTx.TransactionHash = txHash
	refundTx.Status = models.TransactionStatusConfirmed
	refundTx.Metadata = map[string]interface{}{
		"refund_type":       "partial",
		"refund_percentage": refundPercentage,
		"refund_amount":     refundAmount,
		"remaining_amount":  smartCheque.Amount - refundAmount,
	}
	now := time.Now()
	refundTx.ConfirmedAt = &now

	return s.transactionRepo.CreateTransaction(refundTx)
}

// SyncEscrowStatus syncs the XRPL escrow status with the Smart Cheque status
func (s *smartChequeXRPLService) SyncEscrowStatus(ctx context.Context, smartChequeID string) error {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	// Check if escrow address exists
	if smartCheque.EscrowAddress == "" {
		return fmt.Errorf("smart cheque has no escrow address")
	}

	// Get escrow status from XRPL using the transaction hash as sequence
	escrowInfo, err := s.xrplService.GetEscrowStatus(smartCheque.EscrowAddress, smartCheque.EscrowAddress)
	if err != nil {
		log.Printf("Warning: Failed to get XRPL escrow status for Smart Cheque %s: %v", smartChequeID, err)
		// Don't return error here, just log and continue with available data
		return nil
	}

	// Update Smart Cheque status based on XRPL escrow status
	if err := s.updateSmartChequeFromEscrowStatus(ctx, smartCheque, escrowInfo); err != nil {
		return fmt.Errorf("failed to update smart cheque from escrow status: %w", err)
	}

	log.Printf("Successfully synced escrow status for Smart Cheque %s", smartChequeID)
	return nil
}

// updateSmartChequeFromEscrowStatus updates the Smart Cheque status based on XRPL escrow status
func (s *smartChequeXRPLService) updateSmartChequeFromEscrowStatus(ctx context.Context, smartCheque *models.SmartCheque, escrowInfo *xrpl.EscrowInfo) error {
	// Determine the current escrow state based on available information
	// In a real implementation, this would involve checking:
	// - If escrow exists in ledger
	// - If condition is fulfilled
	// - If escrow has been finished or cancelled
	// - Current balance and flags

	// For now, we'll use a simplified state determination
	// This would be enhanced with actual XRPL ledger queries in production

	// Check if escrow is still active (not finished or cancelled)
	isActive := escrowInfo.Flags == 0 // Simplified check

	if !isActive {
		// Escrow is no longer active - check if it was completed or cancelled
		if smartCheque.Status == models.SmartChequeStatusLocked {
			// Assume completion if we have verified milestones
			allMilestonesVerified := true
			for _, milestone := range smartCheque.Milestones {
				if milestone.Status != models.MilestoneStatusVerified {
					allMilestonesVerified = false
					break
				}
			}

			if allMilestonesVerified {
				smartCheque.Status = models.SmartChequeStatusCompleted
				log.Printf("Updated Smart Cheque %s status to completed based on escrow status", smartCheque.ID)
			} else {
				smartCheque.Status = models.SmartChequeStatusDisputed
				log.Printf("Updated Smart Cheque %s status to disputed - escrow finished but milestones not verified", smartCheque.ID)
			}
		}
	}

	smartCheque.UpdatedAt = time.Now()

	// Save the updated Smart Cheque
	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	return nil
}

// MonitorEscrowStatus continuously monitors escrow status and updates Smart Cheque accordingly
func (s *smartChequeXRPLService) MonitorEscrowStatus(ctx context.Context, smartChequeID string, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting escrow monitoring for Smart Cheque %s with interval %v", smartChequeID, interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping escrow monitoring for Smart Cheque %s", smartChequeID)
			return ctx.Err()
		case <-ticker.C:
			if err := s.SyncEscrowStatus(ctx, smartChequeID); err != nil {
				log.Printf("Error syncing escrow status for Smart Cheque %s: %v", smartChequeID, err)
				// Continue monitoring despite errors
			}
		}
	}
}

// GetEscrowHealthStatus provides comprehensive health status of the escrow
func (s *smartChequeXRPLService) GetEscrowHealthStatus(ctx context.Context, smartChequeID string) (*EscrowHealthStatus, error) {
	// Get the Smart Cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return nil, fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}

	status := &EscrowHealthStatus{
		SmartChequeID: smartChequeID,
		Status:        string(smartCheque.Status),
		LastSync:      time.Now(),
	}

	// Check if escrow address exists
	if smartCheque.EscrowAddress == "" {
		status.Health = "no_escrow"
		status.Message = "Smart Cheque has no escrow address"
		return status, nil
	}

	// Get escrow status from XRPL
	escrowInfo, err := s.xrplService.GetEscrowStatus(smartCheque.EscrowAddress, smartCheque.EscrowAddress)
	if err != nil {
		status.Health = "sync_error"
		status.Message = fmt.Sprintf("Failed to get XRPL escrow status: %v", err)
		return status, nil
	}

	// Analyze escrow health
	status.Health = s.analyzeEscrowHealth(smartCheque, escrowInfo)
	status.Message = s.generateHealthMessage(status.Health, smartCheque, escrowInfo)
	status.EscrowInfo = escrowInfo

	return status, nil
}

// EscrowHealthStatus represents the health status of an escrow
type EscrowHealthStatus struct {
	SmartChequeID string           `json:"smart_cheque_id"`
	Status        string           `json:"status"`
	Health        string           `json:"health"`
	Message       string           `json:"message"`
	LastSync      time.Time        `json:"last_sync"`
	EscrowInfo    *xrpl.EscrowInfo `json:"escrow_info,omitempty"`
}

// analyzeEscrowHealth analyzes the health of an escrow based on various factors
func (s *smartChequeXRPLService) analyzeEscrowHealth(smartCheque *models.SmartCheque, escrowInfo *xrpl.EscrowInfo) string {
	// Check escrow flags and status
	if escrowInfo.Flags != 0 {
		return "inactive"
	}

	// Check if escrow is expired
	currentTime := time.Now().Unix()
	if escrowInfo.CancelAfter > 0 && uint32(currentTime) > escrowInfo.CancelAfter {
		return "expired"
	}

	// Check milestone completion status
	completedMilestones := 0
	for _, milestone := range smartCheque.Milestones {
		if milestone.Status == models.MilestoneStatusVerified {
			completedMilestones++
		}
	}

	if completedMilestones == len(smartCheque.Milestones) {
		return "ready_for_release"
	}

	if completedMilestones > 0 {
		return "partially_complete"
	}

	return "active"
}

// generateHealthMessage generates a human-readable health message
func (s *smartChequeXRPLService) generateHealthMessage(health string, smartCheque *models.SmartCheque, escrowInfo *xrpl.EscrowInfo) string {
	// escrowInfo parameter is available for future use in generating more detailed health messages
	_ = escrowInfo // Explicitly ignore to satisfy linter

	switch health {
	case "active":
		return "Escrow is active and monitoring milestones"
	case "partially_complete":
		completed := 0
		for _, m := range smartCheque.Milestones {
			if m.Status == models.MilestoneStatusVerified {
				completed++
			}
		}
		return fmt.Sprintf("%d of %d milestones completed", completed, len(smartCheque.Milestones))
	case "ready_for_release":
		return "All milestones completed, escrow ready for release"
	case "inactive":
		return "Escrow is no longer active"
	case "expired":
		return "Escrow has expired and can be cancelled"
	case "sync_error":
		return "Unable to sync with XRPL ledger"
	default:
		return "Unknown escrow health status"
	}
}

// GetXRPLTransactionHistory retrieves XRPL transaction history for a Smart Cheque
func (s *smartChequeXRPLService) GetXRPLTransactionHistory(ctx context.Context, smartChequeID string) ([]*models.Transaction, error) {
	// Query transactions associated with this Smart Cheque
	transactions, err := s.transactionRepo.GetTransactionsBySmartChequeID(smartChequeID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	return transactions, nil
}

// EscrowMonitoringService provides comprehensive escrow monitoring capabilities
type EscrowMonitoringService struct {
	smartChequeXRPLService SmartChequeXRPLServiceInterface
	smartChequeRepo        repository.SmartChequeRepositoryInterface
	monitoringInterval     time.Duration
	activeMonitors         map[string]context.CancelFunc
	mu                     sync.RWMutex
}

// NewEscrowMonitoringService creates a new escrow monitoring service
func NewEscrowMonitoringService(
	smartChequeXRPLService SmartChequeXRPLServiceInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	monitoringInterval time.Duration,
) *EscrowMonitoringService {
	return &EscrowMonitoringService{
		smartChequeXRPLService: smartChequeXRPLService,
		smartChequeRepo:        smartChequeRepo,
		monitoringInterval:     monitoringInterval,
		activeMonitors:         make(map[string]context.CancelFunc),
	}
}

// StartMonitoringForSmartCheque starts monitoring for a specific Smart Cheque
func (m *EscrowMonitoringService) StartMonitoringForSmartCheque(ctx context.Context, smartChequeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already monitoring
	if _, exists := m.activeMonitors[smartChequeID]; exists {
		log.Printf("Already monitoring Smart Cheque %s", smartChequeID)
		return nil
	}

	// Verify Smart Cheque exists and has escrow
	smartCheque, err := m.smartChequeRepo.GetSmartChequeByID(ctx, smartChequeID)
	if err != nil {
		return fmt.Errorf("failed to get smart cheque: %w", err)
	}
	if smartCheque == nil {
		return fmt.Errorf("smart cheque not found: %s", smartChequeID)
	}
	if smartCheque.EscrowAddress == "" {
		return fmt.Errorf("smart cheque has no escrow address: %s", smartChequeID)
	}

	// Create monitoring context
	monitorCtx, cancel := context.WithCancel(ctx)

	// Store cancel function
	m.activeMonitors[smartChequeID] = cancel

	// Start monitoring in background
	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.activeMonitors, smartChequeID)
			m.mu.Unlock()
		}()

		log.Printf("Started monitoring escrow for Smart Cheque %s", smartChequeID)
		err := m.smartChequeXRPLService.MonitorEscrowStatus(monitorCtx, smartChequeID, m.monitoringInterval)
		if err != nil && err != context.Canceled {
			log.Printf("Error monitoring escrow for Smart Cheque %s: %v", smartChequeID, err)
		}
	}()

	return nil
}

// StopMonitoringForSmartCheque stops monitoring for a specific Smart Cheque
func (m *EscrowMonitoringService) StopMonitoringForSmartCheque(smartChequeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, exists := m.activeMonitors[smartChequeID]
	if !exists {
		return fmt.Errorf("no active monitoring found for Smart Cheque %s", smartChequeID)
	}

	cancel()
	delete(m.activeMonitors, smartChequeID)
	log.Printf("Stopped monitoring escrow for Smart Cheque %s", smartChequeID)
	return nil
}

// GetMonitoredSmartCheques returns list of currently monitored Smart Cheques
func (m *EscrowMonitoringService) GetMonitoredSmartCheques() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	monitored := make([]string, 0, len(m.activeMonitors))
	for smartChequeID := range m.activeMonitors {
		monitored = append(monitored, smartChequeID)
	}

	return monitored
}

// StartMonitoringAllActiveEscrows starts monitoring for all active Smart Cheques with escrows
func (m *EscrowMonitoringService) StartMonitoringAllActiveEscrows(ctx context.Context) error {
	// Get all Smart Cheques with escrow addresses
	smartCheques, err := m.smartChequeRepo.GetSmartChequesByStatus(ctx, models.SmartChequeStatusLocked, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get active smart cheques: %w", err)
	}

	startedCount := 0
	for _, smartCheque := range smartCheques {
		if smartCheque.EscrowAddress != "" {
			if err := m.StartMonitoringForSmartCheque(ctx, smartCheque.ID); err != nil {
				log.Printf("Failed to start monitoring for Smart Cheque %s: %v", smartCheque.ID, err)
			} else {
				startedCount++
			}
		}
	}

	log.Printf("Started monitoring %d active escrows", startedCount)
	return nil
}

// StopAllMonitoring stops all active monitoring
func (m *EscrowMonitoringService) StopAllMonitoring() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stoppedCount := 0
	for smartChequeID, cancel := range m.activeMonitors {
		cancel()
		delete(m.activeMonitors, smartChequeID)
		stoppedCount++
	}

	log.Printf("Stopped monitoring %d escrows", stoppedCount)
	return nil
}

// GetMonitoringStats returns statistics about escrow monitoring
func (m *EscrowMonitoringService) GetMonitoringStats(ctx context.Context) (*MonitoringStats, error) {
	m.mu.RLock()
	activeCount := len(m.activeMonitors)
	m.mu.RUnlock()

	// Get total Smart Cheques with escrows
	totalEscrows := 0
	statuses := []models.SmartChequeStatus{
		models.SmartChequeStatusLocked,
		models.SmartChequeStatusInProgress,
		models.SmartChequeStatusCompleted,
		models.SmartChequeStatusDisputed,
	}

	for _, status := range statuses {
		smartCheques, err := m.smartChequeRepo.GetSmartChequesByStatus(ctx, status, 1000, 0)
		if err != nil {
			continue
		}
		for _, sc := range smartCheques {
			if sc.EscrowAddress != "" {
				totalEscrows++
			}
		}
	}

	return &MonitoringStats{
		ActiveMonitors: activeCount,
		TotalEscrows:   totalEscrows,
		MonitoringRate: float64(activeCount) / float64(totalEscrows) * 100,
		LastUpdated:    time.Now(),
	}, nil
}

// MonitoringStats represents monitoring statistics
type MonitoringStats struct {
	ActiveMonitors int       `json:"active_monitors"`
	TotalEscrows   int       `json:"total_escrows"`
	MonitoringRate float64   `json:"monitoring_rate_percent"`
	LastUpdated    time.Time `json:"last_updated"`
}
