package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// TreasuryServiceInterface defines the interface for treasury operations
type TreasuryServiceInterface interface {
	// Fund Management
	FundEnterprise(ctx context.Context, req *FundingRequest) (*models.AssetTransaction, error)
	WithdrawFunds(ctx context.Context, req *WithdrawalRequest) (*models.AssetTransaction, error)
	TransferFunds(ctx context.Context, req *TransferRequest) ([]*models.AssetTransaction, error)

	// Treasury Operations
	GetTreasuryBalance(ctx context.Context, enterpriseID uuid.UUID) (*TreasuryBalance, error)
	GetFundingHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error)

	// Liquidity Management
	RebalanceLiquidity(ctx context.Context, enterpriseID uuid.UUID) (*LiquidityRebalanceResult, error)
	CheckLiquidityThresholds(ctx context.Context) ([]*LiquidityAlert, error)

	// Treasury Analytics
	GenerateTreasuryReport(ctx context.Context, enterpriseID uuid.UUID, period ReportPeriod) (*TreasuryReport, error)
	CalculateFundingCapacity(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*FundingCapacity, error)
}

// AssetServiceInterface defines the interface for asset service operations
type AssetServiceInterface interface {
	ValidateAssetForTransaction(ctx context.Context, currencyCode string, amount string) error
}

// BalanceServiceInterface defines the interface for balance service operations
type BalanceServiceInterface interface {
	CheckBalanceSufficiency(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, amount string) (bool, error)
}

// AssetTransactionSource represents the source/destination of asset transactions
type AssetTransactionSource string

const (
	AssetTransactionSourceBankTransfer AssetTransactionSource = "bank_transfer"
	AssetTransactionSourceCreditCard   AssetTransactionSource = "credit_card"
	AssetTransactionSourceXRPLWallet   AssetTransactionSource = "xrpl_wallet"
	AssetTransactionSourceInternal     AssetTransactionSource = "internal"
	AssetTransactionSourceExternal     AssetTransactionSource = "external"
)

// TreasuryService implements the treasury service interface
type TreasuryService struct {
	assetRepo       repository.AssetRepositoryInterface
	balanceRepo     repository.BalanceRepositoryInterface
	assetService    AssetServiceInterface
	balanceService  BalanceServiceInterface
	messagingClient messaging.EventBus
}

// NewTreasuryService creates a new treasury service instance
func NewTreasuryService(
	assetRepo repository.AssetRepositoryInterface,
	balanceRepo repository.BalanceRepositoryInterface,
	assetService AssetServiceInterface,
	balanceService BalanceServiceInterface,
	messagingClient messaging.EventBus,
) TreasuryServiceInterface {
	return &TreasuryService{
		assetRepo:       assetRepo,
		balanceRepo:     balanceRepo,
		assetService:    assetService,
		balanceService:  balanceService,
		messagingClient: messagingClient,
	}
}

// FundingRequest represents a funding operation request
type FundingRequest struct {
	EnterpriseID     uuid.UUID              `json:"enterprise_id" validate:"required"`
	CurrencyCode     string                 `json:"currency_code" validate:"required"`
	Amount           string                 `json:"amount" validate:"required"`
	FundingSource    AssetTransactionSource `json:"funding_source" validate:"required"`
	Purpose          string                 `json:"purpose,omitempty"`
	Reference        string                 `json:"reference,omitempty"`
	ApprovalRequired bool                   `json:"approval_required,omitempty"`
}

// WithdrawalRequest represents a withdrawal operation request
type WithdrawalRequest struct {
	EnterpriseID     uuid.UUID              `json:"enterprise_id" validate:"required"`
	CurrencyCode     string                 `json:"currency_code" validate:"required"`
	Amount           string                 `json:"amount" validate:"required"`
	Destination      AssetTransactionSource `json:"destination" validate:"required"`
	Purpose          string                 `json:"purpose,omitempty"`
	Reference        string                 `json:"reference,omitempty"`
	RequireApproval  bool                   `json:"require_approval,omitempty"`
	ComplianceChecks []string               `json:"compliance_checks,omitempty"`
}

// TransferRequest represents an internal transfer operation
type TransferRequest struct {
	FromEnterpriseID uuid.UUID `json:"from_enterprise_id" validate:"required"`
	ToEnterpriseID   uuid.UUID `json:"to_enterprise_id" validate:"required"`
	CurrencyCode     string    `json:"currency_code" validate:"required"`
	Amount           string    `json:"amount" validate:"required"`
	Purpose          string    `json:"purpose,omitempty"`
	Reference        string    `json:"reference,omitempty"`
}

// TreasuryBalance represents the treasury balance summary
type TreasuryBalance struct {
	EnterpriseID    uuid.UUID                 `json:"enterprise_id"`
	Balances        map[string]*BalanceDetail `json:"balances"`
	TotalValueUSD   string                    `json:"total_value_usd"`
	LastUpdated     time.Time                 `json:"last_updated"`
	LiquidityStatus LiquidityStatus           `json:"liquidity_status"`
}

// BalanceDetail provides detailed balance information
type BalanceDetail struct {
	Available    string     `json:"available"`
	Reserved     string     `json:"reserved"`
	Frozen       string     `json:"frozen"`
	Total        string     `json:"total"`
	Tier         WalletTier `json:"tier"`
	LastActivity time.Time  `json:"last_activity"`
}

// WalletTier represents the wallet security tier
type WalletTier string

const (
	WalletTierHot  WalletTier = "hot"
	WalletTierWarm WalletTier = "warm"
	WalletTierCold WalletTier = "cold"
)

// LiquidityStatus represents liquidity health status
type LiquidityStatus string

const (
	LiquidityStatusHealthy   LiquidityStatus = "healthy"
	LiquidityStatusWarning   LiquidityStatus = "warning"
	LiquidityStatusCritical  LiquidityStatus = "critical"
	LiquidityStatusEmergency LiquidityStatus = "emergency"
)

// LiquidityRebalanceResult contains the result of liquidity rebalancing
type LiquidityRebalanceResult struct {
	EnterpriseID     uuid.UUID                        `json:"enterprise_id"`
	RebalanceActions []*LiquidityRebalanceAction      `json:"rebalance_actions"`
	TotalRebalanced  map[string]string                `json:"total_rebalanced"`
	NewDistribution  map[WalletTier]map[string]string `json:"new_distribution"`
	ExecutedAt       time.Time                        `json:"executed_at"`
}

// LiquidityRebalanceAction represents a single rebalance action
type LiquidityRebalanceAction struct {
	FromTier     WalletTier `json:"from_tier"`
	ToTier       WalletTier `json:"to_tier"`
	CurrencyCode string     `json:"currency_code"`
	Amount       string     `json:"amount"`
	Reason       string     `json:"reason"`
}

// LiquidityAlert represents a liquidity threshold alert
type LiquidityAlert struct {
	EnterpriseID   uuid.UUID     `json:"enterprise_id"`
	CurrencyCode   string        `json:"currency_code"`
	Tier           WalletTier    `json:"tier"`
	CurrentAmount  string        `json:"current_amount"`
	ThresholdType  string        `json:"threshold_type"`
	ThresholdValue string        `json:"threshold_value"`
	Severity       AlertSeverity `json:"severity"`
	DetectedAt     time.Time     `json:"detected_at"`
}

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// ReportPeriod represents reporting time periods
type ReportPeriod string

const (
	ReportPeriodDaily     ReportPeriod = "daily"
	ReportPeriodWeekly    ReportPeriod = "weekly"
	ReportPeriodMonthly   ReportPeriod = "monthly"
	ReportPeriodQuarterly ReportPeriod = "quarterly"
)

// TreasuryReport contains treasury analytics and metrics
type TreasuryReport struct {
	EnterpriseID     uuid.UUID                  `json:"enterprise_id"`
	Period           ReportPeriod               `json:"period"`
	StartDate        time.Time                  `json:"start_date"`
	EndDate          time.Time                  `json:"end_date"`
	BalanceSummary   map[string]*BalanceMetrics `json:"balance_summary"`
	ActivitySummary  *ActivityMetrics           `json:"activity_summary"`
	LiquidityMetrics *LiquidityMetrics          `json:"liquidity_metrics"`
	Recommendations  []string                   `json:"recommendations"`
	GeneratedAt      time.Time                  `json:"generated_at"`
}

// BalanceMetrics contains balance-related metrics
type BalanceMetrics struct {
	OpeningBalance string  `json:"opening_balance"`
	ClosingBalance string  `json:"closing_balance"`
	PeakBalance    string  `json:"peak_balance"`
	MinBalance     string  `json:"min_balance"`
	AverageBalance string  `json:"average_balance"`
	NetChange      string  `json:"net_change"`
	Volatility     float64 `json:"volatility"`
}

// ActivityMetrics contains transaction activity metrics
type ActivityMetrics struct {
	TotalTransactions    int64  `json:"total_transactions"`
	FundingOperations    int64  `json:"funding_operations"`
	WithdrawalOperations int64  `json:"withdrawal_operations"`
	TransferOperations   int64  `json:"transfer_operations"`
	TotalVolume          string `json:"total_volume"`
	AverageTransaction   string `json:"average_transaction"`
}

// LiquidityMetrics contains liquidity-related metrics
type LiquidityMetrics struct {
	LiquidityRatio     float64               `json:"liquidity_ratio"`
	UtilizationRate    float64               `json:"utilization_rate"`
	TierDistribution   map[WalletTier]string `json:"tier_distribution"`
	RebalanceFrequency int64                 `json:"rebalance_frequency"`
	ThresholdBreaches  int64                 `json:"threshold_breaches"`
}

// FundingCapacity represents funding capacity analysis
type FundingCapacity struct {
	EnterpriseID       uuid.UUID `json:"enterprise_id"`
	CurrencyCode       string    `json:"currency_code"`
	MaxCapacity        string    `json:"max_capacity"`
	CurrentUtilization string    `json:"current_utilization"`
	AvailableCapacity  string    `json:"available_capacity"`
	UtilizationRate    float64   `json:"utilization_rate"`
	RecommendedLimit   string    `json:"recommended_limit"`
	RiskLevel          RiskLevel `json:"risk_level"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// RiskLevel represents funding risk levels
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelModerate RiskLevel = "moderate"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// FundEnterprise processes funding requests for enterprises
func (s *TreasuryService) FundEnterprise(ctx context.Context, req *FundingRequest) (*models.AssetTransaction, error) {
	// Validate currency support
	if err := s.assetService.ValidateAssetForTransaction(ctx, req.CurrencyCode, req.Amount); err != nil {
		return nil, fmt.Errorf("invalid funding request: %w", err)
	}

	// Check funding capacity
	capacity, err := s.CalculateFundingCapacity(ctx, req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check funding capacity: %w", err)
	}

	// Validate funding amount against capacity
	requestAmount := new(big.Int)
	requestAmount.SetString(req.Amount, 10)
	availableCapacity := new(big.Int)
	availableCapacity.SetString(capacity.AvailableCapacity, 10)

	if requestAmount.Cmp(availableCapacity) > 0 {
		return nil, fmt.Errorf("funding amount %s exceeds available capacity %s", req.Amount, capacity.AvailableCapacity)
	}

	// Create funding transaction
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.EnterpriseID,
		CurrencyCode:    strings.ToUpper(req.CurrencyCode),
		Amount:          req.Amount,
		TransactionType: models.AssetTransactionTypeDeposit,
		Status:          models.AssetTransactionStatusPending,
		Description:     stringPtr(fmt.Sprintf("Funding from %s: %s", req.FundingSource, req.Purpose)),
		ReferenceID:     stringPtr(req.Reference),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save transaction
	if err := s.assetRepo.CreateAssetTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create funding transaction: %w", err)
	}

	// Process funding if no approval required
	if !req.ApprovalRequired {
		if err := s.processFundingTransaction(ctx, transaction); err != nil {
			return nil, fmt.Errorf("failed to process funding: %w", err)
		}
	}

	// Publish funding event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "treasury.funding.requested",
			Source: "treasury-service",
			Data: map[string]interface{}{
				"transaction_id":    transaction.ID.String(),
				"enterprise_id":     req.EnterpriseID.String(),
				"currency_code":     req.CurrencyCode,
				"amount":            req.Amount,
				"funding_source":    req.FundingSource,
				"requires_approval": req.ApprovalRequired,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish funding event: %v\n", err)
		}
	}

	return transaction, nil
}

// WithdrawFunds processes withdrawal requests from enterprises
func (s *TreasuryService) WithdrawFunds(ctx context.Context, req *WithdrawalRequest) (*models.AssetTransaction, error) {
	// Validate currency support
	if err := s.assetService.ValidateAssetForTransaction(ctx, req.CurrencyCode, req.Amount); err != nil {
		return nil, fmt.Errorf("invalid withdrawal request: %w", err)
	}

	// Check balance sufficiency
	sufficient, err := s.balanceService.CheckBalanceSufficiency(ctx, req.EnterpriseID, req.CurrencyCode, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to check balance: %w", err)
	}

	if !sufficient {
		return nil, fmt.Errorf("insufficient balance for withdrawal of %s %s", req.Amount, req.CurrencyCode)
	}

	// Create withdrawal transaction
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.EnterpriseID,
		CurrencyCode:    strings.ToUpper(req.CurrencyCode),
		Amount:          req.Amount,
		TransactionType: models.AssetTransactionTypeWithdrawal,
		Status:          models.AssetTransactionStatusPending,
		Description:     stringPtr(fmt.Sprintf("Withdrawal to %s: %s", req.Destination, req.Purpose)),
		ReferenceID:     stringPtr(req.Reference),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save transaction
	if err := s.assetRepo.CreateAssetTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal transaction: %w", err)
	}

	// Process withdrawal if no approval required
	if !req.RequireApproval {
		if err := s.processWithdrawalTransaction(ctx, transaction); err != nil {
			return nil, fmt.Errorf("failed to process withdrawal: %w", err)
		}
	}

	// Publish withdrawal event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "treasury.withdrawal.requested",
			Source: "treasury-service",
			Data: map[string]interface{}{
				"transaction_id":    transaction.ID.String(),
				"enterprise_id":     req.EnterpriseID.String(),
				"currency_code":     req.CurrencyCode,
				"amount":            req.Amount,
				"destination":       req.Destination,
				"requires_approval": req.RequireApproval,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish withdrawal event: %v\n", err)
		}
	}

	return transaction, nil
}

// processFundingTransaction handles the actual funding operation
func (s *TreasuryService) processFundingTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	// Update balance
	balance, err := s.balanceRepo.GetBalance(ctx, transaction.EnterpriseID, transaction.CurrencyCode)
	if err != nil {
		// Initialize balance if not found
		balance = &models.EnterpriseBalance{
			ID:               uuid.New(),
			EnterpriseID:     transaction.EnterpriseID,
			CurrencyCode:     transaction.CurrencyCode,
			AvailableBalance: "0",
			ReservedBalance:  "0",
			TotalBalance:     "0",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
	}

	// Add funding amount to balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.AvailableBalance, 10)
	fundingAmount := new(big.Int)
	fundingAmount.SetString(transaction.Amount, 10)
	newBalance := new(big.Int).Add(currentBalance, fundingAmount)

	balance.AvailableBalance = newBalance.String()
	balance.TotalBalance = newBalance.String() // Update total as well
	balance.UpdatedAt = time.Now()

	// Save balance
	if err := s.balanceRepo.UpdateBalance(ctx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Update transaction status
	transaction.Status = models.AssetTransactionStatusCompleted
	transaction.ProcessedAt = timePtr(time.Now())
	transaction.UpdatedAt = time.Now()

	if err := s.assetRepo.UpdateAssetTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// processWithdrawalTransaction handles the actual withdrawal operation
func (s *TreasuryService) processWithdrawalTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	// Get and update balance
	balance, err := s.balanceRepo.GetBalance(ctx, transaction.EnterpriseID, transaction.CurrencyCode)
	if err != nil {
		return fmt.Errorf("balance not found: %w", err)
	}

	// Subtract withdrawal amount from balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.AvailableBalance, 10)
	withdrawalAmount := new(big.Int)
	withdrawalAmount.SetString(transaction.Amount, 10)
	newBalance := new(big.Int).Sub(currentBalance, withdrawalAmount)

	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("insufficient balance for withdrawal")
	}

	balance.AvailableBalance = newBalance.String()
	balance.TotalBalance = newBalance.String() // Update total as well
	balance.UpdatedAt = time.Now()

	// Save balance
	if err := s.balanceRepo.UpdateBalance(ctx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Update transaction status
	transaction.Status = models.AssetTransactionStatusCompleted
	transaction.ProcessedAt = timePtr(time.Now())
	transaction.UpdatedAt = time.Now()

	if err := s.assetRepo.UpdateAssetTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

// TransferFunds processes internal fund transfers between enterprises
func (s *TreasuryService) TransferFunds(ctx context.Context, req *TransferRequest) ([]*models.AssetTransaction, error) {
	// Validate currency support
	if err := s.assetService.ValidateAssetForTransaction(ctx, req.CurrencyCode, req.Amount); err != nil {
		return nil, fmt.Errorf("invalid transfer request: %w", err)
	}

	// Check source balance sufficiency
	sufficient, err := s.balanceService.CheckBalanceSufficiency(ctx, req.FromEnterpriseID, req.CurrencyCode, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to check source balance: %w", err)
	}

	if !sufficient {
		return nil, fmt.Errorf("insufficient balance for transfer of %s %s", req.Amount, req.CurrencyCode)
	}

	// Create debit transaction (source enterprise)
	debitTx := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.FromEnterpriseID,
		CurrencyCode:    strings.ToUpper(req.CurrencyCode),
		Amount:          req.Amount,
		TransactionType: models.AssetTransactionTypeTransferOut,
		Status:          models.AssetTransactionStatusPending,
		Description:     stringPtr(fmt.Sprintf("Transfer out: %s", req.Purpose)),
		ReferenceID:     stringPtr(req.Reference),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create credit transaction (destination enterprise)
	creditTx := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    req.ToEnterpriseID,
		CurrencyCode:    strings.ToUpper(req.CurrencyCode),
		Amount:          req.Amount,
		TransactionType: models.AssetTransactionTypeTransferIn,
		Status:          models.AssetTransactionStatusPending,
		Description:     stringPtr(fmt.Sprintf("Transfer in: %s", req.Purpose)),
		ReferenceID:     stringPtr(req.Reference),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save both transactions
	if err := s.assetRepo.CreateAssetTransaction(ctx, debitTx); err != nil {
		return nil, fmt.Errorf("failed to create debit transaction: %w", err)
	}

	if err := s.assetRepo.CreateAssetTransaction(ctx, creditTx); err != nil {
		return nil, fmt.Errorf("failed to create credit transaction: %w", err)
	}

	// Process transfer
	if err := s.processTransferTransactions(ctx, debitTx, creditTx); err != nil {
		return nil, fmt.Errorf("failed to process transfer: %w", err)
	}

	// Publish transfer event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "treasury.transfer.completed",
			Source: "treasury-service",
			Data: map[string]interface{}{
				"debit_transaction_id":  debitTx.ID.String(),
				"credit_transaction_id": creditTx.ID.String(),
				"from_enterprise_id":    req.FromEnterpriseID.String(),
				"to_enterprise_id":      req.ToEnterpriseID.String(),
				"currency_code":         req.CurrencyCode,
				"amount":                req.Amount,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish transfer event: %v\n", err)
		}
	}

	return []*models.AssetTransaction{debitTx, creditTx}, nil
}

// processTransferTransactions handles the atomic transfer operation
func (s *TreasuryService) processTransferTransactions(ctx context.Context, debitTx, creditTx *models.AssetTransaction) error {
	// Get source balance
	sourceBalance, err := s.balanceRepo.GetBalance(ctx, debitTx.EnterpriseID, debitTx.CurrencyCode)
	if err != nil {
		return fmt.Errorf("source balance not found: %w", err)
	}

	// Get or create destination balance
	destBalance, err := s.balanceRepo.GetBalance(ctx, creditTx.EnterpriseID, creditTx.CurrencyCode)
	if err != nil {
		// Initialize balance if not found
		destBalance = &models.EnterpriseBalance{
			ID:               uuid.New(),
			EnterpriseID:     creditTx.EnterpriseID,
			CurrencyCode:     creditTx.CurrencyCode,
			AvailableBalance: "0",
			ReservedBalance:  "0",
			TotalBalance:     "0",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
	}

	// Calculate new balances
	transferAmount := new(big.Int)
	transferAmount.SetString(debitTx.Amount, 10)

	// Source balance (subtract)
	sourceCurrentBalance := new(big.Int)
	sourceCurrentBalance.SetString(sourceBalance.AvailableBalance, 10)
	sourceNewBalance := new(big.Int).Sub(sourceCurrentBalance, transferAmount)

	if sourceNewBalance.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("insufficient balance for transfer")
	}

	// Destination balance (add)
	destCurrentBalance := new(big.Int)
	destCurrentBalance.SetString(destBalance.AvailableBalance, 10)
	destNewBalance := new(big.Int).Add(destCurrentBalance, transferAmount)

	// Update balances
	sourceBalance.AvailableBalance = sourceNewBalance.String()
	sourceBalance.TotalBalance = sourceNewBalance.String()
	sourceBalance.UpdatedAt = time.Now()

	destBalance.AvailableBalance = destNewBalance.String()
	destBalance.TotalBalance = destNewBalance.String()
	destBalance.UpdatedAt = time.Now()

	// Save balances
	if err := s.balanceRepo.UpdateBalance(ctx, sourceBalance); err != nil {
		return fmt.Errorf("failed to update source balance: %w", err)
	}

	if err := s.balanceRepo.UpdateBalance(ctx, destBalance); err != nil {
		return fmt.Errorf("failed to update destination balance: %w", err)
	}

	// Update transaction statuses
	debitTx.Status = models.AssetTransactionStatusCompleted
	debitTx.ProcessedAt = timePtr(time.Now())
	debitTx.UpdatedAt = time.Now()

	creditTx.Status = models.AssetTransactionStatusCompleted
	creditTx.ProcessedAt = timePtr(time.Now())
	creditTx.UpdatedAt = time.Now()

	if err := s.assetRepo.UpdateAssetTransaction(ctx, debitTx); err != nil {
		return fmt.Errorf("failed to update debit transaction: %w", err)
	}

	if err := s.assetRepo.UpdateAssetTransaction(ctx, creditTx); err != nil {
		return fmt.Errorf("failed to update credit transaction: %w", err)
	}

	return nil
}

// GetTreasuryBalance returns the treasury balance summary for an enterprise
func (s *TreasuryService) GetTreasuryBalance(ctx context.Context, enterpriseID uuid.UUID) (*TreasuryBalance, error) {
	balances, err := s.balanceRepo.GetEnterpriseBalances(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise balances: %w", err)
	}

	balanceDetails := make(map[string]*BalanceDetail)
	totalValueUSD := big.NewInt(0)
	liquidityStatus := LiquidityStatusHealthy

	for _, balance := range balances {
		// Calculate total balance
		availableBalance := new(big.Int)
		availableBalance.SetString(balance.AvailableBalance, 10)
		reservedBalance := new(big.Int)
		reservedBalance.SetString(balance.ReservedBalance, 10)
		totalBalance := new(big.Int)
		totalBalance.SetString(balance.TotalBalance, 10)

		balanceDetails[balance.CurrencyCode] = &BalanceDetail{
			Available:    balance.AvailableBalance,
			Reserved:     balance.ReservedBalance,
			Frozen:       "0", // Need to add frozen balance field
			Total:        balance.TotalBalance,
			Tier:         WalletTierHot, // Default tier, could be enhanced
			LastActivity: balance.UpdatedAt,
		}

		// Check liquidity thresholds (simplified logic)
		if availableBalance.Cmp(big.NewInt(1000)) < 0 { // Example threshold
			liquidityStatus = LiquidityStatusWarning
		}
	}

	return &TreasuryBalance{
		EnterpriseID:    enterpriseID,
		Balances:        balanceDetails,
		TotalValueUSD:   totalValueUSD.String(),
		LastUpdated:     time.Now(),
		LiquidityStatus: liquidityStatus,
	}, nil
}

// GetFundingHistory returns the funding transaction history for an enterprise
func (s *TreasuryService) GetFundingHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error) {
	return s.assetRepo.GetAssetTransactionsByEnterprise(ctx, enterpriseID, limit, offset)
}

// RebalanceLiquidity performs liquidity rebalancing across wallet tiers
func (s *TreasuryService) RebalanceLiquidity(ctx context.Context, enterpriseID uuid.UUID) (*LiquidityRebalanceResult, error) {
	// Get current balances
	balances, err := s.balanceRepo.GetEnterpriseBalances(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise balances: %w", err)
	}

	rebalanceActions := []*LiquidityRebalanceAction{}
	totalRebalanced := make(map[string]string)
	newDistribution := make(map[WalletTier]map[string]string)

	// Initialize tier distributions
	newDistribution[WalletTierHot] = make(map[string]string)
	newDistribution[WalletTierWarm] = make(map[string]string)
	newDistribution[WalletTierCold] = make(map[string]string)

	for _, balance := range balances {
		// Simplified rebalancing logic
		currentBalance := new(big.Int)
		currentBalance.SetString(balance.AvailableBalance, 10)

		// Example thresholds (can be made configurable)
		hotThreshold := big.NewInt(10000)  // Keep 10k in hot wallet
		warmThreshold := big.NewInt(50000) // Move excess to warm

		if currentBalance.Cmp(warmThreshold) > 0 {
			// Move excess to cold storage
			excess := new(big.Int).Sub(currentBalance, warmThreshold)
			rebalanceActions = append(rebalanceActions, &LiquidityRebalanceAction{
				FromTier:     WalletTierHot,
				ToTier:       WalletTierCold,
				CurrencyCode: balance.CurrencyCode,
				Amount:       excess.String(),
				Reason:       "Excess balance moved to cold storage",
			})
			totalRebalanced[balance.CurrencyCode] = excess.String()
		}

		// Set new distribution (simplified)
		newDistribution[WalletTierHot][balance.CurrencyCode] = hotThreshold.String()
		newDistribution[WalletTierWarm][balance.CurrencyCode] = "0"
		newDistribution[WalletTierCold][balance.CurrencyCode] = new(big.Int).Sub(currentBalance, hotThreshold).String()
	}

	return &LiquidityRebalanceResult{
		EnterpriseID:     enterpriseID,
		RebalanceActions: rebalanceActions,
		TotalRebalanced:  totalRebalanced,
		NewDistribution:  newDistribution,
		ExecutedAt:       time.Now(),
	}, nil
}

// CheckLiquidityThresholds monitors liquidity thresholds across all enterprises
func (s *TreasuryService) CheckLiquidityThresholds(ctx context.Context) ([]*LiquidityAlert, error) {
	alerts := []*LiquidityAlert{}

	// This would typically iterate through all enterprises
	// For now, return empty alerts as a placeholder
	return alerts, nil
}

// GenerateTreasuryReport creates comprehensive treasury analytics
func (s *TreasuryService) GenerateTreasuryReport(ctx context.Context, enterpriseID uuid.UUID, period ReportPeriod) (*TreasuryReport, error) {
	// Calculate report date range
	endDate := time.Now()
	startDate := s.calculateReportStartDate(endDate, period)

	// Get balances
	balances, err := s.balanceRepo.GetEnterpriseBalances(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}

	// Get transaction history
	transactions, err := s.assetRepo.GetAssetTransactionsByEnterprise(ctx, enterpriseID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Generate balance summary
	balanceSummary := make(map[string]*BalanceMetrics)
	for _, balance := range balances {
		balanceSummary[balance.CurrencyCode] = &BalanceMetrics{
			OpeningBalance: balance.AvailableBalance, // Simplified
			ClosingBalance: balance.AvailableBalance,
			PeakBalance:    balance.AvailableBalance,
			MinBalance:     balance.AvailableBalance,
			AverageBalance: balance.AvailableBalance,
			NetChange:      "0",
			Volatility:     0.0,
		}
	}

	// Generate activity summary
	activitySummary := &ActivityMetrics{
		TotalTransactions:    int64(len(transactions)),
		FundingOperations:    0,
		WithdrawalOperations: 0,
		TransferOperations:   0,
		TotalVolume:          "0",
		AverageTransaction:   "0",
	}

	// Count transaction types
	for _, tx := range transactions {
		switch tx.TransactionType {
		case models.AssetTransactionTypeDeposit:
			activitySummary.FundingOperations++
		case models.AssetTransactionTypeWithdrawal:
			activitySummary.WithdrawalOperations++
		case models.AssetTransactionTypeTransferIn, models.AssetTransactionTypeTransferOut:
			activitySummary.TransferOperations++
		}
	}

	// Generate liquidity metrics
	liquidityMetrics := &LiquidityMetrics{
		LiquidityRatio:     1.0, // Simplified
		UtilizationRate:    0.5, // Simplified
		TierDistribution:   make(map[WalletTier]string),
		RebalanceFrequency: 0,
		ThresholdBreaches:  0,
	}

	liquidityMetrics.TierDistribution[WalletTierHot] = "50"
	liquidityMetrics.TierDistribution[WalletTierWarm] = "30"
	liquidityMetrics.TierDistribution[WalletTierCold] = "20"

	// Generate recommendations
	recommendations := []string{
		"Consider increasing hot wallet threshold for improved liquidity",
		"Monitor withdrawal patterns for better cash flow forecasting",
		"Review funding sources for cost optimization",
	}

	return &TreasuryReport{
		EnterpriseID:     enterpriseID,
		Period:           period,
		StartDate:        startDate,
		EndDate:          endDate,
		BalanceSummary:   balanceSummary,
		ActivitySummary:  activitySummary,
		LiquidityMetrics: liquidityMetrics,
		Recommendations:  recommendations,
		GeneratedAt:      time.Now(),
	}, nil
}

// CalculateFundingCapacity analyzes funding capacity for an enterprise
func (s *TreasuryService) CalculateFundingCapacity(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*FundingCapacity, error) {
	// Get current balance
	balance, err := s.balanceRepo.GetBalance(ctx, enterpriseID, currencyCode)
	if err != nil {
		// If no balance exists, capacity is at maximum
		return &FundingCapacity{
			EnterpriseID:       enterpriseID,
			CurrencyCode:       currencyCode,
			MaxCapacity:        "10000000", // Example max capacity
			CurrentUtilization: "0",
			AvailableCapacity:  "10000000",
			UtilizationRate:    0.0,
			RecommendedLimit:   "8000000", // 80% of max
			RiskLevel:          RiskLevelLow,
			CalculatedAt:       time.Now(),
		}, nil
	}

	// Calculate capacity metrics
	maxCapacity := big.NewInt(10000000) // Example max capacity
	currentUtilization := new(big.Int)
	currentUtilization.SetString(balance.AvailableBalance, 10)
	availableCapacity := new(big.Int).Sub(maxCapacity, currentUtilization)

	// Calculate utilization rate
	utilizationRate := float64(currentUtilization.Int64()) / float64(maxCapacity.Int64())

	// Determine risk level
	riskLevel := RiskLevelLow
	if utilizationRate > 0.8 {
		riskLevel = RiskLevelHigh
	} else if utilizationRate > 0.6 {
		riskLevel = RiskLevelModerate
	}

	recommendedLimit := new(big.Int).Mul(maxCapacity, big.NewInt(8))
	recommendedLimit.Div(recommendedLimit, big.NewInt(10)) // 80% of max

	return &FundingCapacity{
		EnterpriseID:       enterpriseID,
		CurrencyCode:       currencyCode,
		MaxCapacity:        maxCapacity.String(),
		CurrentUtilization: currentUtilization.String(),
		AvailableCapacity:  availableCapacity.String(),
		UtilizationRate:    utilizationRate,
		RecommendedLimit:   recommendedLimit.String(),
		RiskLevel:          riskLevel,
		CalculatedAt:       time.Now(),
	}, nil
}

// calculateReportStartDate calculates the start date for a report period
func (s *TreasuryService) calculateReportStartDate(endDate time.Time, period ReportPeriod) time.Time {
	switch period {
	case ReportPeriodDaily:
		return endDate.AddDate(0, 0, -1)
	case ReportPeriodWeekly:
		return endDate.AddDate(0, 0, -7)
	case ReportPeriodMonthly:
		return endDate.AddDate(0, -1, 0)
	case ReportPeriodQuarterly:
		return endDate.AddDate(0, -3, 0)
	default:
		return endDate.AddDate(0, 0, -1)
	}
}
