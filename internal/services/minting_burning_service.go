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

// MintingBurningServiceInterface defines the interface for minting and burning operations
type MintingBurningServiceInterface interface {
	// Minting Operations
	MintWrappedAsset(ctx context.Context, req *MintingRequest) (*MintingResult, error)
	ValidateCollateral(ctx context.Context, req *CollateralValidationRequest) (*CollateralValidation, error)
	GetMintingCapacity(ctx context.Context, enterpriseID uuid.UUID, wrappedAsset string) (*MintingCapacity, error)

	// Burning Operations
	BurnWrappedAsset(ctx context.Context, req *BurningRequest) (*BurningResult, error)
	InitiateBurning(ctx context.Context, req *BurningRequest) (*BurningResult, error)
	ProcessBurning(ctx context.Context, burningID uuid.UUID) error

	// Collateral Management
	LockCollateral(ctx context.Context, req *CollateralLockRequest) (*CollateralLock, error)
	ReleaseCollateral(ctx context.Context, lockID uuid.UUID) error
	GetCollateralStatus(ctx context.Context, enterpriseID uuid.UUID) (*CollateralStatus, error)

	// Administrative Operations
	GetMintingHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*MintingOperation, error)
	GetBurningHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*BurningOperation, error)
	GetCollateralRatio(ctx context.Context, enterpriseID uuid.UUID, wrappedAsset string) (*CollateralRatio, error)
}

// MintingBurningService implements the minting and burning service interface
type MintingBurningService struct {
	assetRepo       repository.AssetRepositoryInterface
	balanceRepo     repository.BalanceRepositoryInterface
	treasuryService TreasuryServiceInterface
	xrplService     repository.XRPLServiceInterface
	messagingClient messaging.EventBus
}

// NewMintingBurningService creates a new minting and burning service instance
func NewMintingBurningService(
	assetRepo repository.AssetRepositoryInterface,
	balanceRepo repository.BalanceRepositoryInterface,
	treasuryService TreasuryServiceInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient messaging.EventBus,
) MintingBurningServiceInterface {
	return &MintingBurningService{
		assetRepo:       assetRepo,
		balanceRepo:     balanceRepo,
		treasuryService: treasuryService,
		xrplService:     xrplService,
		messagingClient: messagingClient,
	}
}

// MintingRequest represents a request to mint wrapped assets
type MintingRequest struct {
	EnterpriseID     uuid.UUID `json:"enterprise_id" validate:"required"`
	WrappedAsset     string    `json:"wrapped_asset" validate:"required"`     // e.g., "wUSDT", "wUSDC"
	CollateralAsset  string    `json:"collateral_asset" validate:"required"`  // e.g., "USDT", "USDC"
	CollateralAmount string    `json:"collateral_amount" validate:"required"` // Amount of collateral to lock
	MintAmount       string    `json:"mint_amount" validate:"required"`       // Amount of wrapped asset to mint
	CollateralTxHash string    `json:"collateral_tx_hash,omitempty"`          // External transaction hash for collateral
	Purpose          string    `json:"purpose,omitempty"`
	RequireApproval  bool      `json:"require_approval,omitempty"`
}

// BurningRequest represents a request to burn wrapped assets
type BurningRequest struct {
	EnterpriseID      uuid.UUID `json:"enterprise_id" validate:"required"`
	WrappedAsset      string    `json:"wrapped_asset" validate:"required"`
	BurnAmount        string    `json:"burn_amount" validate:"required"`
	RedemptionAddress string    `json:"redemption_address" validate:"required"` // Where to send redeemed collateral
	Purpose           string    `json:"purpose,omitempty"`
	RequireApproval   bool      `json:"require_approval,omitempty"`
}

// CollateralValidationRequest represents a request to validate collateral
type CollateralValidationRequest struct {
	EnterpriseID     uuid.UUID `json:"enterprise_id" validate:"required"`
	CollateralAsset  string    `json:"collateral_asset" validate:"required"`
	CollateralAmount string    `json:"collateral_amount" validate:"required"`
	TxHash           string    `json:"tx_hash,omitempty"`
	BlockHeight      *int64    `json:"block_height,omitempty"`
}

// CollateralLockRequest represents a request to lock collateral
type CollateralLockRequest struct {
	EnterpriseID     uuid.UUID `json:"enterprise_id" validate:"required"`
	CollateralAsset  string    `json:"collateral_asset" validate:"required"`
	CollateralAmount string    `json:"collateral_amount" validate:"required"`
	Purpose          string    `json:"purpose" validate:"required"`
	LockDuration     *int64    `json:"lock_duration,omitempty"` // Duration in seconds
}

// MintingResult contains the result of a minting operation
type MintingResult struct {
	MintingID        uuid.UUID     `json:"minting_id"`
	EnterpriseID     uuid.UUID     `json:"enterprise_id"`
	WrappedAsset     string        `json:"wrapped_asset"`
	CollateralAsset  string        `json:"collateral_asset"`
	CollateralAmount string        `json:"collateral_amount"`
	MintAmount       string        `json:"mint_amount"`
	Status           MintingStatus `json:"status"`
	CollateralLockID *uuid.UUID    `json:"collateral_lock_id,omitempty"`
	TransactionHash  *string       `json:"transaction_hash,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	ProcessedAt      *time.Time    `json:"processed_at,omitempty"`
}

// BurningResult contains the result of a burning operation
type BurningResult struct {
	BurningID         uuid.UUID     `json:"burning_id"`
	EnterpriseID      uuid.UUID     `json:"enterprise_id"`
	WrappedAsset      string        `json:"wrapped_asset"`
	BurnAmount        string        `json:"burn_amount"`
	RedemptionAmount  string        `json:"redemption_amount"`
	RedemptionAddress string        `json:"redemption_address"`
	Status            BurningStatus `json:"status"`
	TransactionHash   *string       `json:"transaction_hash,omitempty"`
	CreatedAt         time.Time     `json:"created_at"`
	ProcessedAt       *time.Time    `json:"processed_at,omitempty"`
}

// CollateralValidation contains the result of collateral validation
type CollateralValidation struct {
	IsValid           bool      `json:"is_valid"`
	CollateralAsset   string    `json:"collateral_asset"`
	CollateralAmount  string    `json:"collateral_amount"`
	ConfirmationCount int       `json:"confirmation_count"`
	RequiredConfirms  int       `json:"required_confirmations"`
	ValidationErrors  []string  `json:"validation_errors,omitempty"`
	ValidatedAt       time.Time `json:"validated_at"`
}

// CollateralLock represents a locked collateral
type CollateralLock struct {
	LockID          uuid.UUID            `json:"lock_id"`
	EnterpriseID    uuid.UUID            `json:"enterprise_id"`
	CollateralAsset string               `json:"collateral_asset"`
	LockedAmount    string               `json:"locked_amount"`
	Purpose         string               `json:"purpose"`
	Status          CollateralLockStatus `json:"status"`
	LockedAt        time.Time            `json:"locked_at"`
	ExpiresAt       *time.Time           `json:"expires_at,omitempty"`
	ReleasedAt      *time.Time           `json:"released_at,omitempty"`
}

// MintingCapacity represents the minting capacity for an enterprise
type MintingCapacity struct {
	EnterpriseID       uuid.UUID `json:"enterprise_id"`
	WrappedAsset       string    `json:"wrapped_asset"`
	MaxMintingCapacity string    `json:"max_minting_capacity"`
	CurrentlyMinted    string    `json:"currently_minted"`
	AvailableCapacity  string    `json:"available_capacity"`
	CollateralRequired string    `json:"collateral_required"`
	CollateralRatio    float64   `json:"collateral_ratio"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// CollateralStatus represents the collateral status for an enterprise
type CollateralStatus struct {
	EnterpriseID        uuid.UUID         `json:"enterprise_id"`
	TotalCollateral     map[string]string `json:"total_collateral"`
	LockedCollateral    map[string]string `json:"locked_collateral"`
	AvailableCollateral map[string]string `json:"available_collateral"`
	CollateralLocks     []*CollateralLock `json:"collateral_locks"`
	LastUpdated         time.Time         `json:"last_updated"`
}

// CollateralRatio represents the collateral ratio for a wrapped asset
type CollateralRatio struct {
	EnterpriseID    uuid.UUID `json:"enterprise_id"`
	WrappedAsset    string    `json:"wrapped_asset"`
	CollateralAsset string    `json:"collateral_asset"`
	RequiredRatio   float64   `json:"required_ratio"`
	CurrentRatio    float64   `json:"current_ratio"`
	IsHealthy       bool      `json:"is_healthy"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

// MintingOperation represents a historical minting operation
type MintingOperation struct {
	MintingID        uuid.UUID     `json:"minting_id"`
	EnterpriseID     uuid.UUID     `json:"enterprise_id"`
	WrappedAsset     string        `json:"wrapped_asset"`
	CollateralAsset  string        `json:"collateral_asset"`
	CollateralAmount string        `json:"collateral_amount"`
	MintAmount       string        `json:"mint_amount"`
	Status           MintingStatus `json:"status"`
	CreatedAt        time.Time     `json:"created_at"`
	ProcessedAt      *time.Time    `json:"processed_at,omitempty"`
}

// BurningOperation represents a historical burning operation
type BurningOperation struct {
	BurningID         uuid.UUID     `json:"burning_id"`
	EnterpriseID      uuid.UUID     `json:"enterprise_id"`
	WrappedAsset      string        `json:"wrapped_asset"`
	BurnAmount        string        `json:"burn_amount"`
	RedemptionAmount  string        `json:"redemption_amount"`
	RedemptionAddress string        `json:"redemption_address"`
	Status            BurningStatus `json:"status"`
	CreatedAt         time.Time     `json:"created_at"`
	ProcessedAt       *time.Time    `json:"processed_at,omitempty"`
}

// Status enums
type MintingStatus string

const (
	MintingStatusPending    MintingStatus = "pending"
	MintingStatusValidating MintingStatus = "validating"
	MintingStatusApproved   MintingStatus = "approved"
	MintingStatusMinting    MintingStatus = "minting"
	MintingStatusCompleted  MintingStatus = "completed"
	MintingStatusFailed     MintingStatus = "failed"
	MintingStatusCancelled  MintingStatus = "canceled"
)

type BurningStatus string

const (
	BurningStatusPending   BurningStatus = "pending"
	BurningStatusApproved  BurningStatus = "approved"
	BurningStatusBurning   BurningStatus = "burning"
	BurningStatusRedeeming BurningStatus = "redeeming"
	BurningStatusCompleted BurningStatus = "completed"
	BurningStatusFailed    BurningStatus = "failed"
	BurningStatusCancelled BurningStatus = "canceled"
)

// CollateralLockStatus represents the status of collateral locks
type CollateralLockStatus string

const (
	CollateralLockStatusActive   CollateralLockStatus = "active"
	CollateralLockStatusLocked   CollateralLockStatus = "locked"
	CollateralLockStatusReleased CollateralLockStatus = "released"
	CollateralLockStatusExpired  CollateralLockStatus = "expired"
)

// MintWrappedAsset processes a minting request for wrapped assets
func (s *MintingBurningService) MintWrappedAsset(ctx context.Context, req *MintingRequest) (*MintingResult, error) {
	// Validate wrapped asset is supported
	wrappedAsset, err := s.assetRepo.GetAssetByCurrency(ctx, req.WrappedAsset)
	if err != nil || wrappedAsset.AssetType != models.AssetTypeWrapped {
		return nil, fmt.Errorf("unsupported wrapped asset: %s", req.WrappedAsset)
	}

	// Validate collateral asset
	_, err = s.assetRepo.GetAssetByCurrency(ctx, req.CollateralAsset)
	if err != nil {
		return nil, fmt.Errorf("unsupported collateral asset: %s", req.CollateralAsset)
	}

	// Validate collateral if external transaction provided
	if req.CollateralTxHash != "" {
		validation, err := s.ValidateCollateral(ctx, &CollateralValidationRequest{
			EnterpriseID:     req.EnterpriseID,
			CollateralAsset:  req.CollateralAsset,
			CollateralAmount: req.CollateralAmount,
			TxHash:           req.CollateralTxHash,
		})
		if err != nil {
			return nil, fmt.Errorf("collateral validation failed: %w", err)
		}
		if !validation.IsValid {
			return nil, fmt.Errorf("invalid collateral: %v", validation.ValidationErrors)
		}
	}

	// Check collateral ratio requirements
	ratio, err := s.calculateRequiredCollateralRatio(req.WrappedAsset, req.CollateralAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate collateral ratio: %w", err)
	}

	// Validate collateral amount is sufficient
	if err := s.validateCollateralSufficiency(req.CollateralAmount, req.MintAmount, ratio); err != nil {
		return nil, fmt.Errorf("insufficient collateral: %w", err)
	}

	// Lock collateral
	collateralLock, err := s.LockCollateral(ctx, &CollateralLockRequest{
		EnterpriseID:     req.EnterpriseID,
		CollateralAsset:  req.CollateralAsset,
		CollateralAmount: req.CollateralAmount,
		Purpose:          fmt.Sprintf("Minting %s %s", req.MintAmount, req.WrappedAsset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lock collateral: %w", err)
	}

	// Create minting record
	mintingResult := &MintingResult{
		MintingID:        uuid.New(),
		EnterpriseID:     req.EnterpriseID,
		WrappedAsset:     req.WrappedAsset,
		CollateralAsset:  req.CollateralAsset,
		CollateralAmount: req.CollateralAmount,
		MintAmount:       req.MintAmount,
		Status:           MintingStatusPending,
		CollateralLockID: &collateralLock.LockID,
		CreatedAt:        time.Now(),
	}

	// Process minting if no approval required
	if !req.RequireApproval {
		if err := s.processMinting(ctx, mintingResult); err != nil {
			return nil, fmt.Errorf("failed to process minting: %w", err)
		}
	}

	// Publish minting event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "minting.requested",
			Source: "minting-burning-service",
			Data: map[string]interface{}{
				"minting_id":        mintingResult.MintingID.String(),
				"enterprise_id":     req.EnterpriseID.String(),
				"wrapped_asset":     req.WrappedAsset,
				"collateral_asset":  req.CollateralAsset,
				"mint_amount":       req.MintAmount,
				"requires_approval": req.RequireApproval,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish minting event: %v\n", err)
		}
	}

	return mintingResult, nil
}

// processMinting handles the actual minting process
func (s *MintingBurningService) processMinting(ctx context.Context, result *MintingResult) error {
	// Update status to minting
	result.Status = MintingStatusMinting

	// Create wrapped asset balance entry or update existing
	balance, err := s.balanceRepo.GetBalance(ctx, result.EnterpriseID, result.WrappedAsset)
	if err != nil {
		// Create new balance if doesn't exist
		balance = &models.EnterpriseBalance{
			ID:               uuid.New(),
			EnterpriseID:     result.EnterpriseID,
			CurrencyCode:     result.WrappedAsset,
			AvailableBalance: "0",
			ReservedBalance:  "0",
			TotalBalance:     "0",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
	}

	// Add minted amount to balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.AvailableBalance, 10)
	mintAmount := new(big.Int)
	mintAmount.SetString(result.MintAmount, 10)
	newBalance := new(big.Int).Add(currentBalance, mintAmount)

	balance.AvailableBalance = newBalance.String()
	balance.TotalBalance = newBalance.String()
	balance.UpdatedAt = time.Now()

	// Save balance
	if err := s.balanceRepo.UpdateBalance(ctx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create asset transaction record
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    result.EnterpriseID,
		CurrencyCode:    result.WrappedAsset,
		TransactionType: models.AssetTransactionTypeMint,
		Amount:          result.MintAmount,
		Status:          models.AssetTransactionStatusCompleted,
		Description:     stringPtr(fmt.Sprintf("Minted %s backed by %s %s", result.MintAmount, result.CollateralAmount, result.CollateralAsset)),
		ReferenceID:     stringPtr(result.MintingID.String()),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ProcessedAt:     timePtr(time.Now()),
	}

	if err := s.assetRepo.CreateAssetTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create asset transaction: %w", err)
	}

	// Update minting result
	result.Status = MintingStatusCompleted
	result.ProcessedAt = timePtr(time.Now())

	return nil
}

// validateCollateralSufficiency checks if collateral amount is sufficient for minting
func (s *MintingBurningService) validateCollateralSufficiency(collateralAmount, mintAmount string, requiredRatio float64) error {
	collateral := new(big.Int)
	collateral.SetString(collateralAmount, 10)
	mint := new(big.Int)
	mint.SetString(mintAmount, 10)

	// Calculate required collateral (mint amount * ratio)
	ratioFloat := new(big.Float).SetFloat64(requiredRatio)
	mintFloat := new(big.Float).SetInt(mint)
	requiredCollateralFloat := new(big.Float).Mul(mintFloat, ratioFloat)

	requiredCollateral, _ := requiredCollateralFloat.Int(nil)

	if collateral.Cmp(requiredCollateral) < 0 {
		return fmt.Errorf("insufficient collateral: provided %s, required %s (ratio: %.2f)",
			collateralAmount, requiredCollateral.String(), requiredRatio)
	}

	return nil
}

// calculateRequiredCollateralRatio returns the required collateral ratio for a wrapped asset
func (s *MintingBurningService) calculateRequiredCollateralRatio(wrappedAsset, collateralAsset string) (float64, error) {
	// Default ratios (can be made configurable)
	ratios := map[string]map[string]float64{
		"wUSDT": {
			"USDT": 1.0, // 1:1 for same asset
			"USDC": 1.1, // 10% over-collateralization for different stablecoin
			"XRP":  1.5, // 50% over-collateralization for volatile asset
		},
		"wUSDC": {
			"USDC": 1.0,
			"USDT": 1.1,
			"XRP":  1.5,
		},
		"we₹": {
			"e₹":   1.0,
			"USDT": 1.2, // 20% over-collateralization for foreign currency
			"USDC": 1.2,
			"XRP":  1.8, // 80% over-collateralization for volatile cross-currency
		},
	}

	if assetRatios, exists := ratios[wrappedAsset]; exists {
		if ratio, exists := assetRatios[collateralAsset]; exists {
			return ratio, nil
		}
	}

	return 0, fmt.Errorf("unsupported collateral asset %s for wrapped asset %s", collateralAsset, wrappedAsset)
}

// BurnWrappedAsset processes a burning request for wrapped assets
func (s *MintingBurningService) BurnWrappedAsset(ctx context.Context, req *BurningRequest) (*BurningResult, error) {
	// Validate wrapped asset is supported
	wrappedAsset, err := s.assetRepo.GetAssetByCurrency(ctx, req.WrappedAsset)
	if err != nil || wrappedAsset.AssetType != models.AssetTypeWrapped {
		return nil, fmt.Errorf("unsupported wrapped asset: %s", req.WrappedAsset)
	}

	// Check if enterprise has sufficient wrapped asset balance
	balance, err := s.balanceRepo.GetBalance(ctx, req.EnterpriseID, req.WrappedAsset)
	if err != nil {
		return nil, fmt.Errorf("no balance found for wrapped asset %s", req.WrappedAsset)
	}

	availableBalance := new(big.Int)
	availableBalance.SetString(balance.AvailableBalance, 10)
	burnAmount := new(big.Int)
	burnAmount.SetString(req.BurnAmount, 10)

	if availableBalance.Cmp(burnAmount) < 0 {
		return nil, fmt.Errorf("insufficient wrapped asset balance: have %s, need %s",
			balance.AvailableBalance, req.BurnAmount)
	}

	// Calculate redemption amount (typically 1:1 minus fees)
	redemptionAmount := s.calculateRedemptionAmount(req.BurnAmount, req.WrappedAsset)

	// Create burning record
	burningResult := &BurningResult{
		BurningID:         uuid.New(),
		EnterpriseID:      req.EnterpriseID,
		WrappedAsset:      req.WrappedAsset,
		BurnAmount:        req.BurnAmount,
		RedemptionAmount:  redemptionAmount,
		RedemptionAddress: req.RedemptionAddress,
		Status:            BurningStatusPending,
		CreatedAt:         time.Now(),
	}

	// Process burning if no approval required
	if !req.RequireApproval {
		if err := s.processBurning(ctx, burningResult); err != nil {
			return nil, fmt.Errorf("failed to process burning: %w", err)
		}
	}

	// Publish burning event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "burning.requested",
			Source: "minting-burning-service",
			Data: map[string]interface{}{
				"burning_id":         burningResult.BurningID.String(),
				"enterprise_id":      req.EnterpriseID.String(),
				"wrapped_asset":      req.WrappedAsset,
				"burn_amount":        req.BurnAmount,
				"redemption_amount":  redemptionAmount,
				"redemption_address": req.RedemptionAddress,
				"requires_approval":  req.RequireApproval,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish burning event: %v\n", err)
		}
	}

	return burningResult, nil
}

// InitiateBurning initiates the burning process with approval workflow
func (s *MintingBurningService) InitiateBurning(ctx context.Context, req *BurningRequest) (*BurningResult, error) {
	req.RequireApproval = true
	return s.BurnWrappedAsset(ctx, req)
}

// ProcessBurning processes an approved burning request
func (s *MintingBurningService) ProcessBurning(ctx context.Context, burningID uuid.UUID) error {
	// In a real implementation, you'd fetch the burning record from database
	// For now, this is a placeholder
	fmt.Printf("Processing burning request: %s\n", burningID.String())
	return nil
}

// ValidateCollateral validates collateral for minting operations
func (s *MintingBurningService) ValidateCollateral(ctx context.Context, req *CollateralValidationRequest) (*CollateralValidation, error) {
	validation := &CollateralValidation{
		CollateralAsset:   req.CollateralAsset,
		CollateralAmount:  req.CollateralAmount,
		ConfirmationCount: 0,
		RequiredConfirms:  6, // Default required confirmations
		ValidationErrors:  []string{},
		ValidatedAt:       time.Now(),
	}

	// Validate asset is supported
	_, err := s.assetRepo.GetAssetByCurrency(ctx, req.CollateralAsset)
	if err != nil {
		validation.ValidationErrors = append(validation.ValidationErrors,
			fmt.Sprintf("Unsupported collateral asset: %s", req.CollateralAsset))
	}

	// Validate amount format
	amount := new(big.Int)
	if _, ok := amount.SetString(req.CollateralAmount, 10); !ok {
		validation.ValidationErrors = append(validation.ValidationErrors,
			"Invalid collateral amount format")
	}

	// If transaction hash provided, validate on blockchain (simplified)
	if req.TxHash != "" {
		// In a real implementation, this would query the blockchain
		validation.ConfirmationCount = 12 // Simulated
		validation.IsValid = validation.ConfirmationCount >= validation.RequiredConfirms
	} else {
		validation.IsValid = len(validation.ValidationErrors) == 0
	}

	return validation, nil
}

// LockCollateral locks collateral for minting or other operations
func (s *MintingBurningService) LockCollateral(ctx context.Context, req *CollateralLockRequest) (*CollateralLock, error) {
	// Check if enterprise has sufficient balance
	balance, err := s.balanceRepo.GetBalance(ctx, req.EnterpriseID, req.CollateralAsset)
	if err != nil {
		return nil, fmt.Errorf("no balance found for collateral asset %s", req.CollateralAsset)
	}

	availableBalance := new(big.Int)
	availableBalance.SetString(balance.AvailableBalance, 10)
	lockAmount := new(big.Int)
	lockAmount.SetString(req.CollateralAmount, 10)

	if availableBalance.Cmp(lockAmount) < 0 {
		return nil, fmt.Errorf("insufficient balance to lock: have %s, need %s",
			balance.AvailableBalance, req.CollateralAmount)
	}

	// Move amount from available to reserved
	newAvailable := new(big.Int).Sub(availableBalance, lockAmount)
	reservedBalance := new(big.Int)
	reservedBalance.SetString(balance.ReservedBalance, 10)
	newReserved := new(big.Int).Add(reservedBalance, lockAmount)

	balance.AvailableBalance = newAvailable.String()
	balance.ReservedBalance = newReserved.String()
	balance.UpdatedAt = time.Now()

	if err := s.balanceRepo.UpdateBalance(ctx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Create collateral lock record
	lock := &CollateralLock{
		LockID:          uuid.New(),
		EnterpriseID:    req.EnterpriseID,
		CollateralAsset: req.CollateralAsset,
		LockedAmount:    req.CollateralAmount,
		Purpose:         req.Purpose,
		Status:          CollateralLockStatusLocked,
		LockedAt:        time.Now(),
	}

	// Set expiration if duration provided
	if req.LockDuration != nil && *req.LockDuration > 0 {
		expiresAt := time.Now().Add(time.Duration(*req.LockDuration) * time.Second)
		lock.ExpiresAt = &expiresAt
	}

	return lock, nil
}

// ReleaseCollateral releases locked collateral
func (s *MintingBurningService) ReleaseCollateral(ctx context.Context, lockID uuid.UUID) error {
	// In a real implementation, you'd fetch the lock from database and update balances
	fmt.Printf("Releasing collateral lock: %s\n", lockID.String())
	return nil
}

// GetCollateralStatus returns the collateral status for an enterprise
func (s *MintingBurningService) GetCollateralStatus(ctx context.Context, enterpriseID uuid.UUID) (*CollateralStatus, error) {
	balances, err := s.balanceRepo.GetEnterpriseBalances(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise balances: %w", err)
	}

	totalCollateral := make(map[string]string)
	lockedCollateral := make(map[string]string)
	availableCollateral := make(map[string]string)

	for _, balance := range balances {
		totalCollateral[balance.CurrencyCode] = balance.TotalBalance
		lockedCollateral[balance.CurrencyCode] = balance.ReservedBalance
		availableCollateral[balance.CurrencyCode] = balance.AvailableBalance
	}

	return &CollateralStatus{
		EnterpriseID:        enterpriseID,
		TotalCollateral:     totalCollateral,
		LockedCollateral:    lockedCollateral,
		AvailableCollateral: availableCollateral,
		CollateralLocks:     []*CollateralLock{}, // Would be populated from database
		LastUpdated:         time.Now(),
	}, nil
}

// GetMintingCapacity returns the minting capacity for an enterprise
func (s *MintingBurningService) GetMintingCapacity(ctx context.Context, enterpriseID uuid.UUID, wrappedAsset string) (*MintingCapacity, error) {
	// Get collateral status
	_, err := s.GetCollateralStatus(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collateral status: %w", err)
	}

	// Calculate capacity based on available collateral
	// This is simplified - in reality, you'd consider all collateral types and ratios
	maxCapacity := "1000000" // Example max capacity
	currentlyMinted := "0"   // Would query from database
	availableCapacity := maxCapacity

	return &MintingCapacity{
		EnterpriseID:       enterpriseID,
		WrappedAsset:       wrappedAsset,
		MaxMintingCapacity: maxCapacity,
		CurrentlyMinted:    currentlyMinted,
		AvailableCapacity:  availableCapacity,
		CollateralRequired: "1000000", // 1:1 ratio example
		CollateralRatio:    1.0,
		CalculatedAt:       time.Now(),
	}, nil
}

// GetCollateralRatio returns the collateral ratio for a wrapped asset
func (s *MintingBurningService) GetCollateralRatio(ctx context.Context, enterpriseID uuid.UUID, wrappedAsset string) (*CollateralRatio, error) {
	// This would calculate the actual collateral ratio from database records
	return &CollateralRatio{
		EnterpriseID:    enterpriseID,
		WrappedAsset:    wrappedAsset,
		CollateralAsset: "USDT", // Example
		RequiredRatio:   1.0,
		CurrentRatio:    1.2,
		IsHealthy:       true,
		CalculatedAt:    time.Now(),
	}, nil
}

// GetMintingHistory returns the minting history for an enterprise
func (s *MintingBurningService) GetMintingHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*MintingOperation, error) {
	// In a real implementation, this would query the database for minting records
	return []*MintingOperation{}, nil
}

// GetBurningHistory returns the burning history for an enterprise
func (s *MintingBurningService) GetBurningHistory(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*BurningOperation, error) {
	// In a real implementation, this would query the database for burning records
	return []*BurningOperation{}, nil
}

// Helper methods
func (s *MintingBurningService) calculateRedemptionAmount(burnAmount, _ string) string {
	// Simplified: 1:1 redemption minus small fee
	burnAmountBig := new(big.Int)
	burnAmountBig.SetString(burnAmount, 10)

	// Apply 0.1% burning fee
	fee := new(big.Int).Div(burnAmountBig, big.NewInt(1000))
	redemptionAmount := new(big.Int).Sub(burnAmountBig, fee)

	return redemptionAmount.String()
}

func (s *MintingBurningService) processBurning(ctx context.Context, result *BurningResult) error {
	// Update status to burning
	result.Status = BurningStatusBurning

	// Get wrapped asset balance
	balance, err := s.balanceRepo.GetBalance(ctx, result.EnterpriseID, result.WrappedAsset)
	if err != nil {
		return fmt.Errorf("balance not found: %w", err)
	}

	// Subtract burned amount from balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.AvailableBalance, 10)
	burnAmount := new(big.Int)
	burnAmount.SetString(result.BurnAmount, 10)
	newBalance := new(big.Int).Sub(currentBalance, burnAmount)

	if newBalance.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("insufficient balance for burning")
	}

	balance.AvailableBalance = newBalance.String()
	balance.TotalBalance = newBalance.String()
	balance.UpdatedAt = time.Now()

	// Save balance
	if err := s.balanceRepo.UpdateBalance(ctx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create asset transaction record
	transaction := &models.AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    result.EnterpriseID,
		CurrencyCode:    result.WrappedAsset,
		TransactionType: models.AssetTransactionTypeBurn,
		Amount:          result.BurnAmount,
		Status:          models.AssetTransactionStatusCompleted,
		Description: stringPtr(fmt.Sprintf("Burned %s for redemption of %s to %s",
			result.BurnAmount, result.RedemptionAmount, result.RedemptionAddress)),
		ReferenceID: stringPtr(result.BurningID.String()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ProcessedAt: timePtr(time.Now()),
	}

	if err := s.assetRepo.CreateAssetTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("failed to create asset transaction: %w", err)
	}

	// Update burning result
	result.Status = BurningStatusCompleted
	result.ProcessedAt = timePtr(time.Now())

	return nil
}
