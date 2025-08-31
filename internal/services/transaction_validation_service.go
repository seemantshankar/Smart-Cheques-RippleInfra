package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// TransactionValidationService handles transaction validation and processing
type TransactionValidationService struct {
	assetRepo       repository.AssetRepository
	balanceRepo     repository.BalanceRepository
	transactionRepo repository.TransactionRepositoryInterface
	messagingClient *messaging.Service
}

// NewTransactionValidationService creates a new transaction validation service
func NewTransactionValidationService(
	assetRepo repository.AssetRepository,
	balanceRepo repository.BalanceRepository,
	transactionRepo repository.TransactionRepositoryInterface,
	messagingClient *messaging.Service,
) *TransactionValidationService {
	return &TransactionValidationService{
		assetRepo:       assetRepo,
		balanceRepo:     balanceRepo,
		transactionRepo: transactionRepo,
		messagingClient: messagingClient,
	}
}

// TransactionValidationRequest represents a transaction to be validated
type TransactionValidationRequest struct {
	EnterpriseID    uuid.UUID                   `json:"enterprise_id" validate:"required"`
	CurrencyCode    string                      `json:"currency_code" validate:"required"`
	Amount          string                      `json:"amount" validate:"required"`
	TransactionType models.AssetTransactionType `json:"transaction_type" validate:"required"`
	FromAddress     *string                     `json:"from_address,omitempty"`
	ToAddress       *string                     `json:"to_address,omitempty"`
	ReferenceID     *string                     `json:"reference_id,omitempty"`
	Description     *string                     `json:"description,omitempty"`
	Metadata        map[string]interface{}      `json:"metadata,omitempty"`
	RequireApproval bool                        `json:"require_approval"`
}

// TransactionValidationResult represents the result of transaction validation
type TransactionValidationResult struct {
	IsValid          bool                   `json:"is_valid"`
	Errors           []string               `json:"errors,omitempty"`
	Warnings         []string               `json:"warnings,omitempty"`
	EstimatedFee     string                 `json:"estimated_fee"`
	RequiredApproval bool                   `json:"required_approval"`
	RiskScore        int                    `json:"risk_score"` // 1-100, 100 being highest risk
	ProcessingTime   time.Duration          `json:"processing_time"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ValidateTransaction performs comprehensive transaction validation
func (s *TransactionValidationService) ValidateTransaction(ctx context.Context, req *TransactionValidationRequest) (*TransactionValidationResult, error) {
	startTime := time.Now()
	result := &TransactionValidationResult{
		IsValid:      true,
		Errors:       []string{},
		Warnings:     []string{},
		EstimatedFee: "0",
		RiskScore:    0,
		Metadata:     make(map[string]interface{}),
	}

	// 1. Validate asset/currency
	if err := s.validateAsset(ctx, req.CurrencyCode, result); err != nil {
		return result, err
	}

	// 2. Validate amount format and constraints
	if err := s.validateAmount(ctx, req.CurrencyCode, req.Amount, result); err != nil {
		return result, err
	}

	// 3. Validate enterprise and balance
	if err := s.validateEnterpriseBalance(ctx, req, result); err != nil {
		return result, err
	}

	// 4. Validate transaction type specific rules
	if err := s.validateTransactionType(ctx, req, result); err != nil {
		return result, err
	}

	// 5. Calculate risk score
	s.calculateRiskScore(req, result)

	// 6. Determine approval requirements
	s.determineApprovalRequirements(req, result)

	// 7. Estimate transaction fee
	s.estimateTransactionFee(req, result)

	// 8. Final validation
	result.IsValid = len(result.Errors) == 0
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// validateAsset validates the currency/asset
// Returns error which is always nil in this implementation
func (s *TransactionValidationService) validateAsset(ctx context.Context, currencyCode string, result *TransactionValidationResult) error {
	asset, err := s.assetRepo.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported currency: %s", currencyCode))
		return nil
	}

	if !asset.IsActive {
		result.Errors = append(result.Errors, fmt.Sprintf("Currency %s is currently deactivated", currencyCode))
		return nil
	}

	if asset.GlobalFreeze {
		result.Errors = append(result.Errors, fmt.Sprintf("Currency %s is globally frozen", currencyCode))
		return nil
	}

	result.Metadata["asset_type"] = asset.AssetType
	result.Metadata["requires_trust_line"] = asset.RequiresTrustLine()

	return nil
}

// validateAmount validates transaction amount
// Returns error which is always nil in this implementation
func (s *TransactionValidationService) validateAmount(ctx context.Context, currencyCode, amount string, result *TransactionValidationResult) error {
	// Parse amount
	amountBig := new(big.Int)
	_, ok := amountBig.SetString(amount, 10)
	if !ok {
		result.Errors = append(result.Errors, "Invalid amount format")
		return nil
	}

	if amountBig.Sign() <= 0 {
		result.Errors = append(result.Errors, "Amount must be positive")
		return nil
	}

	// Get asset for constraints
	asset, err := s.assetRepo.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		return nil // Already handled in validateAsset
	}

	// Check minimum amount
	minAmount, err := asset.GetMinimumAmountBigInt()
	if err != nil {
		result.Warnings = append(result.Warnings, "Invalid minimum amount configuration")
	} else if amountBig.Cmp(minAmount) < 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("Amount below minimum %s", asset.MinimumAmount))
	}

	// Check maximum amount if configured
	if asset.MaximumAmount != nil {
		maxAmount := new(big.Int)
		if _, ok := maxAmount.SetString(*asset.MaximumAmount, 10); ok {
			if amountBig.Cmp(maxAmount) > 0 {
				result.Errors = append(result.Errors, fmt.Sprintf("Amount exceeds maximum %s", *asset.MaximumAmount))
			}
		}
	}

	return nil
}

// validateEnterpriseBalance validates enterprise balance requirements
// Returns error which is always nil in this implementation
func (s *TransactionValidationService) validateEnterpriseBalance(ctx context.Context, req *TransactionValidationRequest, result *TransactionValidationResult) error {
	balance, err := s.balanceRepo.GetEnterpriseBalance(ctx, req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		// Check if this is a debit operation
		tx := &models.AssetTransaction{TransactionType: req.TransactionType}
		if tx.IsDebit() {
			result.Errors = append(result.Errors, "No balance record found for debit operation")
		} else {
			result.Warnings = append(result.Warnings, "Balance record will be created")
		}
		return nil
	}

	// Check if balance is frozen
	if balance.IsFrozen {
		result.Errors = append(result.Errors, fmt.Sprintf("Balance is frozen: %s", *balance.FreezeReason))
		return nil
	}

	// For debit operations, check sufficient balance
	tx := &models.AssetTransaction{TransactionType: req.TransactionType}
	if tx.IsDebit() {
		sufficient, err := balance.HasSufficientBalance(req.Amount)
		if err != nil {
			result.Warnings = append(result.Warnings, "Failed to check balance sufficiency")
		} else if !sufficient {
			result.Errors = append(result.Errors, fmt.Sprintf("Insufficient balance: available %s, required %s", balance.AvailableBalance, req.Amount))
		}
	}

	// Check daily/monthly limits if configured
	if balance.DailyLimit != nil {
		result.Warnings = append(result.Warnings, "Daily limit checking not implemented")
		// TODO: Implement daily limit checking
	}

	if balance.MonthlyLimit != nil {
		result.Warnings = append(result.Warnings, "Monthly limit checking not implemented")
		// TODO: Implement monthly limit checking
	}

	result.Metadata["current_balance"] = balance.AvailableBalance
	result.Metadata["reserved_balance"] = balance.ReservedBalance

	return nil
}

// validateTransactionType validates transaction type specific rules
func (s *TransactionValidationService) validateTransactionType(_ context.Context, req *TransactionValidationRequest, result *TransactionValidationResult) error {
	switch req.TransactionType {
	case models.AssetTransactionTypeDeposit:
		return s.validateDepositTransaction(req, result)
	case models.AssetTransactionTypeWithdrawal:
		return s.validateWithdrawalTransaction(req, result)
	case models.AssetTransactionTypeTransferIn, models.AssetTransactionTypeTransferOut:
		return s.validateTransferTransaction(req, result)
	case models.AssetTransactionTypeEscrowLock, models.AssetTransactionTypeEscrowRelease:
		return s.validateEscrowTransaction(req, result)
	case models.AssetTransactionTypeFee:
		return s.validateFeeTransaction(req, result)
	case models.AssetTransactionTypeAdjustment:
		return s.validateAdjustmentTransaction(req, result)
	default:
		result.Errors = append(result.Errors, fmt.Sprintf("Unsupported transaction type: %s", req.TransactionType))
	}

	return nil
}

// validateDepositTransaction validates deposit-specific rules
func (s *TransactionValidationService) validateDepositTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Deposits should have source information
	if req.Metadata != nil {
		if source, exists := req.Metadata["source"]; !exists || source == "" {
			result.Warnings = append(result.Warnings, "Deposit source not specified")
		}
	}

	// Check for external transaction hash for verification
	if req.Metadata != nil {
		if _, exists := req.Metadata["external_tx_hash"]; !exists {
			result.Warnings = append(result.Warnings, "No external transaction hash provided for verification")
			result.RiskScore += 10
		}
	}

	return nil
}

// validateWithdrawalTransaction validates withdrawal-specific rules
func (s *TransactionValidationService) validateWithdrawalTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Withdrawals require destination address
	if req.ToAddress == nil || *req.ToAddress == "" {
		if req.Metadata != nil {
			if dest, exists := req.Metadata["destination_address"]; !exists || dest == "" {
				result.Errors = append(result.Errors, "Destination address required for withdrawal")
			}
		} else {
			result.Errors = append(result.Errors, "Destination address required for withdrawal")
		}
	}

	// Withdrawals require higher scrutiny
	result.RequiredApproval = true
	result.RiskScore += 20

	return nil
}

// validateTransferTransaction validates transfer-specific rules
func (s *TransactionValidationService) validateTransferTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Transfers should have both from and to addresses
	if req.FromAddress == nil || *req.FromAddress == "" {
		result.Warnings = append(result.Warnings, "From address not specified for transfer")
	}

	if req.ToAddress == nil || *req.ToAddress == "" {
		result.Warnings = append(result.Warnings, "To address not specified for transfer")
	}

	return nil
}

// validateEscrowTransaction validates escrow-specific rules
func (s *TransactionValidationService) validateEscrowTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Escrow operations should have reference ID
	if req.ReferenceID == nil || *req.ReferenceID == "" {
		result.Warnings = append(result.Warnings, "Escrow operation without reference ID")
		result.RiskScore += 5
	}

	return nil
}

// validateFeeTransaction validates fee-specific rules
func (s *TransactionValidationService) validateFeeTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Fee transactions should have reference to the original transaction
	if req.ReferenceID == nil {
		result.Warnings = append(result.Warnings, "Fee transaction without reference ID")
	}

	return nil
}

// validateAdjustmentTransaction validates adjustment-specific rules
func (s *TransactionValidationService) validateAdjustmentTransaction(req *TransactionValidationRequest, result *TransactionValidationResult) error {
	// Adjustments require approval and detailed description
	result.RequiredApproval = true
	result.RiskScore += 30

	if req.Description == nil || *req.Description == "" {
		result.Errors = append(result.Errors, "Adjustment transactions require detailed description")
	}

	return nil
}

// calculateRiskScore calculates transaction risk score
func (s *TransactionValidationService) calculateRiskScore(req *TransactionValidationRequest, result *TransactionValidationResult) {
	// Parse amount for risk calculation
	amountBig := new(big.Int)
	if _, ok := amountBig.SetString(req.Amount, 10); !ok {
		result.RiskScore += 50 // Invalid amount is high risk
		return
	}

	// High amount transactions are riskier
	// This is a simplified example - in practice, this would be more sophisticated
	if amountBig.Cmp(big.NewInt(100000000)) > 0 { // > 100 units (assuming 6 decimal places)
		result.RiskScore += 20
	}

	// Weekend/off-hours transactions might be riskier
	now := time.Now()
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		result.RiskScore += 5
	}

	// Late night transactions
	if now.Hour() < 6 || now.Hour() > 22 {
		result.RiskScore += 10
	}

	// Ensure risk score doesn't exceed 100
	if result.RiskScore > 100 {
		result.RiskScore = 100
	}

	result.Metadata["risk_factors"] = map[string]interface{}{
		"high_amount": amountBig.Cmp(big.NewInt(100000000)) > 0,
		"weekend":     now.Weekday() == time.Saturday || now.Weekday() == time.Sunday,
		"off_hours":   now.Hour() < 6 || now.Hour() > 22,
	}
}

// determineApprovalRequirements determines if transaction requires approval
func (s *TransactionValidationService) determineApprovalRequirements(req *TransactionValidationRequest, result *TransactionValidationResult) {
	// High risk transactions require approval
	if result.RiskScore > 50 {
		result.RequiredApproval = true
	}

	// Manual approval flag
	if req.RequireApproval {
		result.RequiredApproval = true
	}

	// Certain transaction types always require approval
	switch req.TransactionType {
	case models.AssetTransactionTypeAdjustment:
		result.RequiredApproval = true
	case models.AssetTransactionTypeWithdrawal:
		// Parse amount for withdrawal approval threshold
		amountBig := new(big.Int)
		if _, ok := amountBig.SetString(req.Amount, 10); ok {
			// Withdrawals over 50 units require approval
			if amountBig.Cmp(big.NewInt(50000000)) > 0 {
				result.RequiredApproval = true
			}
		}
	}
}

// estimateTransactionFee estimates the transaction fee
func (s *TransactionValidationService) estimateTransactionFee(req *TransactionValidationRequest, result *TransactionValidationResult) {
	// Simplified fee calculation - in practice this would be more sophisticated
	baseFee := big.NewInt(1000) // Base fee of 0.001 units

	// Higher fees for certain transaction types
	switch req.TransactionType {
	case models.AssetTransactionTypeWithdrawal:
		baseFee.Mul(baseFee, big.NewInt(10)) // 10x fee for withdrawals
	case models.AssetTransactionTypeTransferOut:
		baseFee.Mul(baseFee, big.NewInt(5)) // 5x fee for transfers
	}

	// Higher fees for high-risk transactions
	if result.RiskScore > 70 {
		baseFee.Mul(baseFee, big.NewInt(2)) // 2x fee for high-risk
	}

	result.EstimatedFee = baseFee.String()
}

// ProcessValidatedTransaction processes a transaction that has been validated
func (s *TransactionValidationService) ProcessValidatedTransaction(ctx context.Context, req *TransactionValidationRequest, validationResult *TransactionValidationResult) (*models.AssetTransaction, error) {
	if !validationResult.IsValid {
		return nil, fmt.Errorf("cannot process invalid transaction: %v", validationResult.Errors)
	}

	// Create asset transaction record
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.EnterpriseID,
		CurrencyCode:    req.CurrencyCode,
		TransactionType: req.TransactionType,
		Amount:          req.Amount,
		Fee:             validationResult.EstimatedFee,
		ReferenceID:     req.ReferenceID,
		Status:          models.AssetTransactionStatusPending,
		Description:     req.Description,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// If requires approval, keep as pending
	if validationResult.RequiredApproval {
		transaction.Status = models.AssetTransactionStatusPending
	} else {
		transaction.Status = models.AssetTransactionStatusProcessing
	}

	// Create the transaction record
	if err := s.balanceRepo.CreateAssetTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Publish transaction event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "transaction.validated",
			Source: "transaction-validation-service",
			Data: map[string]interface{}{
				"transaction_id":    transaction.ID,
				"enterprise_id":     req.EnterpriseID,
				"currency_code":     req.CurrencyCode,
				"transaction_type":  req.TransactionType,
				"amount":            req.Amount,
				"requires_approval": validationResult.RequiredApproval,
				"risk_score":        validationResult.RiskScore,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(event)
	}

	return transaction, nil
}

// GetTransactionValidationHistory gets validation history for analysis
func (s *TransactionValidationService) GetTransactionValidationHistory(_ context.Context, _ uuid.UUID, _ int) ([]TransactionValidationSummary, error) {
	// This would typically query a validation history table
	// For now, return empty slice as placeholder
	return []TransactionValidationSummary{}, nil
}

// TransactionValidationSummary represents a summary of past validations
type TransactionValidationSummary struct {
	TransactionID    uuid.UUID `json:"transaction_id"`
	EnterpriseID     uuid.UUID `json:"enterprise_id"`
	CurrencyCode     string    `json:"currency_code"`
	Amount           string    `json:"amount"`
	ValidationTime   time.Time `json:"validation_time"`
	RiskScore        int       `json:"risk_score"`
	RequiredApproval bool      `json:"required_approval"`
	WasValid         bool      `json:"was_valid"`
	ErrorCount       int       `json:"error_count"`
	WarningCount     int       `json:"warning_count"`
}
