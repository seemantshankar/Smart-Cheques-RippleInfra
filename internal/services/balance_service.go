package services

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// BalanceService handles balance tracking and validation operations
type BalanceService struct {
	balanceRepo     repository.BalanceRepository
	assetRepo       repository.AssetRepository
	messagingClient *messaging.MessagingService
}

// NewBalanceService creates a new balance service instance
func NewBalanceService(balanceRepo repository.BalanceRepository, assetRepo repository.AssetRepository, messagingClient *messaging.MessagingService) *BalanceService {
	return &BalanceService{
		balanceRepo:     balanceRepo,
		assetRepo:       assetRepo,
		messagingClient: messagingClient,
	}
}

// BalanceOperationRequest represents a request to modify an enterprise balance
type BalanceOperationRequest struct {
	EnterpriseID  uuid.UUID                   `json:"enterprise_id" validate:"required"`
	CurrencyCode  string                      `json:"currency_code" validate:"required"`
	Amount        string                      `json:"amount" validate:"required"`
	OperationType models.AssetTransactionType `json:"operation_type" validate:"required"`
	ReferenceID   *string                     `json:"reference_id,omitempty"`
	Description   *string                     `json:"description,omitempty"`
	Metadata      map[string]interface{}      `json:"metadata,omitempty"`
}

// InitializeEnterpriseBalance creates initial balance records for an enterprise
func (s *BalanceService) InitializeEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	// Validate currency is supported
	asset, err := s.assetRepo.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("unsupported currency %s: %w", currencyCode, err)
	}

	if !asset.IsActive {
		return nil, fmt.Errorf("currency %s is currently deactivated", currencyCode)
	}

	// Check if balance already exists
	existingBalance, err := s.balanceRepo.GetEnterpriseBalance(ctx, enterpriseID, currencyCode)
	if err == nil && existingBalance != nil {
		return existingBalance, nil // Already exists
	}

	// Create new balance record
	balance := &models.EnterpriseBalance{
		ID:               uuid.New(),
		EnterpriseID:     enterpriseID,
		CurrencyCode:     currencyCode,
		AvailableBalance: "0",
		ReservedBalance:  "0",
		TotalBalance:     "0",
		XRPLBalance:      "0",
		IsFrozen:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.balanceRepo.CreateEnterpriseBalance(ctx, balance); err != nil {
		return nil, fmt.Errorf("failed to create balance record: %w", err)
	}

	// Publish balance initialization event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "balance.initialized",
			Source: "balance-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID,
				"currency_code": currencyCode,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		s.messagingClient.PublishEvent(event)
	}

	return balance, nil
}

// GetEnterpriseBalance retrieves a specific balance for validation or display
func (s *BalanceService) GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	return s.balanceRepo.GetEnterpriseBalance(ctx, enterpriseID, currencyCode)
}

// GetEnterpriseBalances retrieves all balances for an enterprise
func (s *BalanceService) GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error) {
	return s.balanceRepo.GetEnterpriseBalances(ctx, enterpriseID)
}

// ValidateBalanceOperation validates if a balance operation can be performed
func (s *BalanceService) ValidateBalanceOperation(ctx context.Context, req *BalanceOperationRequest) error {
	// Validate currency is supported and active
	asset, err := s.assetRepo.GetAssetByCurrency(ctx, req.CurrencyCode)
	if err != nil {
		return fmt.Errorf("unsupported currency %s: %w", req.CurrencyCode, err)
	}

	if !asset.IsActive {
		return fmt.Errorf("currency %s is currently deactivated", req.CurrencyCode)
	}

	// Validate amount format and range
	amountBig := new(big.Int)
	_, ok := amountBig.SetString(req.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount format: %s", req.Amount)
	}

	if amountBig.Sign() <= 0 {
		return fmt.Errorf("amount must be positive: %s", req.Amount)
	}

	// Check minimum amount constraint
	minAmount, err := asset.GetMinimumAmountBigInt()
	if err != nil {
		return fmt.Errorf("invalid minimum amount configuration: %w", err)
	}

	if amountBig.Cmp(minAmount) < 0 {
		return fmt.Errorf("amount %s is below minimum %s for currency %s", req.Amount, asset.MinimumAmount, req.CurrencyCode)
	}

	// Check maximum amount constraint if configured
	if asset.MaximumAmount != nil {
		maxAmount := new(big.Int)
		_, ok := maxAmount.SetString(*asset.MaximumAmount, 10)
		if !ok {
			return fmt.Errorf("invalid maximum amount configuration for currency %s", req.CurrencyCode)
		}

		if amountBig.Cmp(maxAmount) > 0 {
			return fmt.Errorf("amount %s exceeds maximum %s for currency %s", req.Amount, *asset.MaximumAmount, req.CurrencyCode)
		}
	}

	// Get current balance to validate operation-specific constraints
	balance, err := s.balanceRepo.GetEnterpriseBalance(ctx, req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		// If balance doesn't exist and this is a debit operation, that's an error
		tx := &models.AssetTransaction{TransactionType: req.OperationType}
		if tx.IsDebit() {
			return fmt.Errorf("insufficient balance: balance record not found")
		}
		// For credit operations, balance will be created if needed
		return nil
	}

	// Check if balance is frozen
	if balance.IsFrozen {
		return fmt.Errorf("balance is frozen: %s", *balance.FreezeReason)
	}

	// For debit operations, verify sufficient available balance
	tx := &models.AssetTransaction{TransactionType: req.OperationType}
	if tx.IsDebit() {
		sufficient, err := balance.HasSufficientBalance(req.Amount)
		if err != nil {
			return fmt.Errorf("failed to check balance sufficiency: %w", err)
		}

		if !sufficient {
			return fmt.Errorf("insufficient available balance: have %s, need %s", balance.AvailableBalance, req.Amount)
		}
	}

	return nil
}

// ProcessBalanceOperation processes a balance operation after validation
func (s *BalanceService) ProcessBalanceOperation(ctx context.Context, req *BalanceOperationRequest) (*models.AssetTransaction, error) {
	// Validate the operation first
	if err := s.ValidateBalanceOperation(ctx, req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Ensure balance record exists
	_, err := s.balanceRepo.GetEnterpriseBalance(ctx, req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		// Create balance record if it doesn't exist and this is a credit operation
		tx := &models.AssetTransaction{TransactionType: req.OperationType}
		if !tx.IsDebit() {
			_, err = s.InitializeEnterpriseBalance(ctx, req.EnterpriseID, req.CurrencyCode)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize balance: %w", err)
			}
		} else {
			return nil, fmt.Errorf("balance record not found for debit operation")
		}
	}

	// Create transaction record with metadata
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.EnterpriseID,
		CurrencyCode:    req.CurrencyCode,
		TransactionType: req.OperationType,
		Amount:          req.Amount,
		Fee:             "0",
		ReferenceID:     req.ReferenceID,
		Status:          models.AssetTransactionStatusPending,
		Description:     req.Description,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Process the balance update
	if err := s.balanceRepo.UpdateBalance(ctx, req.EnterpriseID, req.CurrencyCode, req.Amount, req.OperationType, req.ReferenceID); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Publish balance operation event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "balance.operation",
			Source: "balance-service",
			Data: map[string]interface{}{
				"enterprise_id":  req.EnterpriseID,
				"currency_code":  req.CurrencyCode,
				"operation_type": req.OperationType,
				"amount":         req.Amount,
				"transaction_id": transaction.ID,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			// Log the error but don't fail the operation
			log.Printf("Failed to publish balance operation event: %v", err)
		}
	}

	return transaction, nil
}

// CheckBalanceSufficiency checks if an enterprise has sufficient balance for an operation
func (s *BalanceService) CheckBalanceSufficiency(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, requiredAmount string) (bool, error) {
	balance, err := s.balanceRepo.GetEnterpriseBalance(ctx, enterpriseID, currencyCode)
	if err != nil {
		return false, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.IsFrozen {
		return false, fmt.Errorf("balance is frozen")
	}

	return balance.HasSufficientBalance(requiredAmount)
}

// ReserveBalance reserves amount from available balance (moves to reserved)
func (s *BalanceService) ReserveBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, amount string, referenceID *string) error {
	req := &BalanceOperationRequest{
		EnterpriseID:  enterpriseID,
		CurrencyCode:  currencyCode,
		Amount:        amount,
		OperationType: models.AssetTransactionTypeEscrowLock,
		ReferenceID:   referenceID,
		Description:   &[]string{"Balance reservation"}[0],
	}

	_, err := s.ProcessBalanceOperation(ctx, req)
	return err
}

// ReleaseReservedBalance releases reserved balance back to available
func (s *BalanceService) ReleaseReservedBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, amount string, referenceID *string) error {
	req := &BalanceOperationRequest{
		EnterpriseID:  enterpriseID,
		CurrencyCode:  currencyCode,
		Amount:        amount,
		OperationType: models.AssetTransactionTypeEscrowRelease,
		ReferenceID:   referenceID,
		Description:   &[]string{"Reserved balance release"}[0],
	}

	_, err := s.ProcessBalanceOperation(ctx, req)
	return err
}

// FreezeBalance freezes a balance to prevent operations
func (s *BalanceService) FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error {
	if err := s.balanceRepo.FreezeBalance(ctx, enterpriseID, currencyCode, reason); err != nil {
		return fmt.Errorf("failed to freeze balance: %w", err)
	}

	// Publish freeze event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "balance.frozen",
			Source: "balance-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID,
				"currency_code": currencyCode,
				"reason":        reason,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			// Log the error but don't fail the operation
			log.Printf("Failed to publish balance frozen event: %v", err)
		}
	}

	return nil
}

// UnfreezeBalance unfreezes a balance to allow operations
func (s *BalanceService) UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error {
	if err := s.balanceRepo.UnfreezeBalance(ctx, enterpriseID, currencyCode); err != nil {
		return fmt.Errorf("failed to unfreeze balance: %w", err)
	}

	// Publish unfreeze event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "balance.unfrozen",
			Source: "balance-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID,
				"currency_code": currencyCode,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			// Log the error but don't fail the operation
			log.Printf("Failed to publish balance unfrozen event: %v", err)
		}
	}

	return nil
}

// GetBalanceSummary returns balance summary for an enterprise
func (s *BalanceService) GetBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error) {
	return s.balanceRepo.GetEnterpriseBalanceSummary(ctx, enterpriseID)
}

// GetAllBalanceSummaries returns balance summaries for all enterprises
func (s *BalanceService) GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error) {
	return s.balanceRepo.GetAllBalanceSummaries(ctx)
}

// SyncXRPLBalance synchronizes balance with XRPL network
func (s *BalanceService) SyncXRPLBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, xrplBalance string) error {
	balance, err := s.balanceRepo.GetEnterpriseBalance(ctx, enterpriseID, currencyCode)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Update XRPL balance and sync timestamp
	balance.XRPLBalance = xrplBalance
	syncTime := time.Now()
	balance.LastXRPLSync = &syncTime
	balance.UpdatedAt = time.Now()

	if err := s.balanceRepo.UpdateEnterpriseBalance(ctx, balance); err != nil {
		return fmt.Errorf("failed to update XRPL balance: %w", err)
	}

	// Publish sync event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "balance.xrpl_synced",
			Source: "balance-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID,
				"currency_code": currencyCode,
				"xrpl_balance":  xrplBalance,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			// Log the error but don't fail the operation
			log.Printf("Failed to publish XRPL balance sync event: %v", err)
		}
	}

	return nil
}

// ValidateBalanceConsistency checks if internal balance matches XRPL balance
func (s *BalanceService) ValidateBalanceConsistency(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*BalanceConsistencyReport, error) {
	balance, err := s.balanceRepo.GetEnterpriseBalance(ctx, enterpriseID, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	report := &BalanceConsistencyReport{
		EnterpriseID:    enterpriseID,
		CurrencyCode:    currencyCode,
		InternalBalance: balance.TotalBalance,
		XRPLBalance:     balance.XRPLBalance,
		LastSyncTime:    balance.LastXRPLSync,
		IsConsistent:    balance.TotalBalance == balance.XRPLBalance,
		CheckedAt:       time.Now(),
	}

	if !report.IsConsistent {
		internalBig := new(big.Int)
		xrplBig := new(big.Int)

		internalBig.SetString(balance.TotalBalance, 10)
		xrplBig.SetString(balance.XRPLBalance, 10)

		diff := new(big.Int).Sub(internalBig, xrplBig)
		report.Discrepancy = diff.String()
	}

	return report, nil
}

// BalanceConsistencyReport represents a balance consistency check result
type BalanceConsistencyReport struct {
	EnterpriseID    uuid.UUID  `json:"enterprise_id"`
	CurrencyCode    string     `json:"currency_code"`
	InternalBalance string     `json:"internal_balance"`
	XRPLBalance     string     `json:"xrpl_balance"`
	LastSyncTime    *time.Time `json:"last_sync_time,omitempty"`
	IsConsistent    bool       `json:"is_consistent"`
	Discrepancy     string     `json:"discrepancy,omitempty"`
	CheckedAt       time.Time  `json:"checked_at"`
}
