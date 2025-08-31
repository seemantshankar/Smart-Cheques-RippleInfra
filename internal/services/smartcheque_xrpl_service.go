package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// SmartChequeXRPLServiceInterface defines the interface for Smart Cheque XRPL integration operations
type SmartChequeXRPLServiceInterface interface {
	// CreateEscrowForSmartCheque creates an XRPL escrow for a Smart Cheque
	CreateEscrowForSmartCheque(ctx context.Context, smartChequeID string, payerWalletAddress, payeeWalletAddress string) error

	// CompleteMilestonePayment completes a milestone payment by finishing the XRPL escrow
	CompleteMilestonePayment(ctx context.Context, smartChequeID, milestoneID string) error

	// CancelSmartChequeEscrow cancels the XRPL escrow for a Smart Cheque
	CancelSmartChequeEscrow(ctx context.Context, smartChequeID string) error

	// SyncEscrowStatus syncs the XRPL escrow status with the Smart Cheque status
	SyncEscrowStatus(ctx context.Context, smartChequeID string) error

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

	// Generate a secret for the escrow condition
	milestoneSecret := fmt.Sprintf("smartcheque_%s_secret_%d", smartChequeID, time.Now().Unix())

	// Create the XRPL escrow
	result, fulfillment, err := s.xrplService.CreateSmartChequeEscrow(
		payerWalletAddress,
		payeeWalletAddress,
		smartCheque.Amount,
		string(smartCheque.Currency),
		milestoneSecret,
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

// CancelSmartChequeEscrow cancels the XRPL escrow for a Smart Cheque
func (s *smartChequeXRPLService) CancelSmartChequeEscrow(ctx context.Context, smartChequeID string) error {
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

	// Cancel the XRPL escrow
	// Note: This is a simplified implementation. In reality, we would need the actual
	// sequence number and other details from the original escrow creation
	result, err := s.xrplService.CancelSmartCheque(
		smartCheque.EscrowAddress, // Using escrow address as account for this example
		smartCheque.EscrowAddress, // Using escrow address as owner for this example
		1,                         // Sequence number - would need to be retrieved from original transaction
	)
	if err != nil {
		return fmt.Errorf("failed to cancel XRPL escrow: %w", err)
	}

	// Update Smart Cheque status
	smartCheque.Status = models.SmartChequeStatusDisputed
	smartCheque.UpdatedAt = time.Now()

	if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
		return fmt.Errorf("failed to update smart cheque: %w", err)
	}

	// Create a transaction record for tracking
	transaction := models.NewTransaction(
		models.TransactionTypeEscrowCancel,
		smartCheque.EscrowAddress,
		smartCheque.EscrowAddress,
		fmt.Sprintf("%f", smartCheque.Amount),
		string(smartCheque.Currency),
		smartCheque.PayerID,
		smartCheque.PayerID, // Using payer ID as user ID for now
	)

	// Set XRPL-specific fields
	transaction.SmartChequeID = &smartChequeID
	transaction.TransactionHash = result.TransactionID
	transaction.Status = models.TransactionStatusConfirmed
	now := time.Now()
	transaction.ConfirmedAt = &now

	// Save the transaction
	if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
		log.Printf("Warning: Failed to save transaction record: %v", err)
	}

	log.Printf("Cancelled XRPL escrow for Smart Cheque %s with transaction ID %s",
		smartChequeID, result.TransactionID)
	return nil
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

	// Get escrow status from XRPL
	// Note: This is a simplified implementation. In reality, we would need the actual
	// sequence number from the original escrow creation
	escrowInfo, err := s.xrplService.GetEscrowStatus(smartCheque.EscrowAddress, "1")
	if err != nil {
		return fmt.Errorf("failed to get XRPL escrow status: %w", err)
	}

	// Log the escrow info for debugging
	log.Printf("Escrow info for Smart Cheque %s: %+v", smartChequeID, escrowInfo)

	// In a real implementation, we would update the Smart Cheque status based on
	// the actual escrow status from XRPL

	return nil
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
