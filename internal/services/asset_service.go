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

// AssetService handles asset management operations
type AssetService struct {
	assetRepo       repository.AssetRepository
	balanceRepo     repository.BalanceRepository
	messagingClient *messaging.MessagingService
}

// NewAssetService creates a new asset service instance
func NewAssetService(assetRepo repository.AssetRepository, balanceRepo repository.BalanceRepository, messagingClient *messaging.MessagingService) *AssetService {
	return &AssetService{
		assetRepo:       assetRepo,
		balanceRepo:     balanceRepo,
		messagingClient: messagingClient,
	}
}

// AssetRegistryRequest represents a request to register a new asset
type AssetRegistryRequest struct {
	CurrencyCode     string  `json:"currency_code" validate:"required,min=1,max=10"`
	CurrencyName     string  `json:"currency_name" validate:"required,min=1,max=100"`
	AssetType        string  `json:"asset_type" validate:"required,oneof=native stablecoin cbdc wrapped"`
	IssuerAddress    *string `json:"issuer_address,omitempty"`
	CurrencyHex      *string `json:"currency_hex,omitempty"`
	DecimalPlaces    int     `json:"decimal_places" validate:"min=0,max=18"`
	MinimumAmount    string  `json:"minimum_amount" validate:"required"`
	MaximumAmount    *string `json:"maximum_amount,omitempty"`
	TrustLineLimit   *string `json:"trust_line_limit,omitempty"`
	TransferFee      float64 `json:"transfer_fee" validate:"min=0,max=1"`
	GlobalFreeze     bool    `json:"global_freeze"`
	NoFreeze         bool    `json:"no_freeze"`
	Description      *string `json:"description,omitempty"`
	IconURL          *string `json:"icon_url,omitempty"`
	DocumentationURL *string `json:"documentation_url,omitempty"`
}

// RegisterAsset registers a new supported asset
func (s *AssetService) RegisterAsset(ctx context.Context, req *AssetRegistryRequest) (*models.SupportedAsset, error) {
	// Validate currency code doesn't already exist
	existing, err := s.assetRepo.GetAssetByCurrency(ctx, req.CurrencyCode)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("asset with currency code %s already exists", req.CurrencyCode)
	}

	// Validate minimum amount format
	minAmount := new(big.Int)
	_, ok := minAmount.SetString(req.MinimumAmount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid minimum amount format: %s", req.MinimumAmount)
	}

	// Validate maximum amount if provided
	if req.MaximumAmount != nil {
		maxAmount := new(big.Int)
		_, ok := maxAmount.SetString(*req.MaximumAmount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid maximum amount format: %s", *req.MaximumAmount)
		}

		// Ensure minimum <= maximum
		if minAmount.Cmp(maxAmount) > 0 {
			return nil, fmt.Errorf("minimum amount cannot be greater than maximum amount")
		}
	}

	// Create asset model
	asset := &models.SupportedAsset{
		ID:               uuid.New(),
		CurrencyCode:     strings.ToUpper(req.CurrencyCode),
		CurrencyName:     req.CurrencyName,
		AssetType:        models.AssetType(req.AssetType),
		IssuerAddress:    req.IssuerAddress,
		CurrencyHex:      req.CurrencyHex,
		DecimalPlaces:    req.DecimalPlaces,
		MinimumAmount:    req.MinimumAmount,
		MaximumAmount:    req.MaximumAmount,
		IsActive:         true,
		TrustLineLimit:   req.TrustLineLimit,
		TransferFee:      req.TransferFee,
		GlobalFreeze:     req.GlobalFreeze,
		NoFreeze:         req.NoFreeze,
		Description:      req.Description,
		IconURL:          req.IconURL,
		DocumentationURL: req.DocumentationURL,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	if err := s.assetRepo.CreateAsset(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	// Publish asset registration event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "asset.registered",
			Source: "asset-service",
			Data: map[string]interface{}{
				"currency_code": asset.CurrencyCode,
				"asset_type":    asset.AssetType,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to publish asset registration event: %v\n", err)
		}
	}

	return asset, nil
}

// GetSupportedAssets returns all supported assets
func (s *AssetService) GetSupportedAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error) {
	return s.assetRepo.GetAssets(ctx, activeOnly)
}

// GetAssetByCurrency returns asset information for a specific currency
func (s *AssetService) GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error) {
	currencyCode = strings.ToUpper(currencyCode)
	if currencyCode == "E₹" {
		currencyCode = "e₹" // Handle special case
	}

	return s.assetRepo.GetAssetByCurrency(ctx, currencyCode)
}

// UpdateAsset updates an existing asset's configuration
func (s *AssetService) UpdateAsset(ctx context.Context, currencyCode string, req *AssetRegistryRequest) (*models.SupportedAsset, error) {
	// Get existing asset
	asset, err := s.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	// Update fields
	asset.CurrencyName = req.CurrencyName
	asset.IssuerAddress = req.IssuerAddress
	asset.CurrencyHex = req.CurrencyHex
	asset.DecimalPlaces = req.DecimalPlaces
	asset.MinimumAmount = req.MinimumAmount
	asset.MaximumAmount = req.MaximumAmount
	asset.TrustLineLimit = req.TrustLineLimit
	asset.TransferFee = req.TransferFee
	asset.GlobalFreeze = req.GlobalFreeze
	asset.NoFreeze = req.NoFreeze
	asset.Description = req.Description
	asset.IconURL = req.IconURL
	asset.DocumentationURL = req.DocumentationURL
	asset.UpdatedAt = time.Now()

	// Save changes
	if err := s.assetRepo.UpdateAsset(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	// Publish update event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "asset.updated",
			Source: "asset-service",
			Data: map[string]interface{}{
				"currency_code": asset.CurrencyCode,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(event); err != nil {
			fmt.Printf("Warning: Failed to publish asset update event: %v\n", err)
		}
	}

	return asset, nil
}

// DeactivateAsset deactivates an asset (soft delete)
func (s *AssetService) DeactivateAsset(ctx context.Context, currencyCode string) error {
	asset, err := s.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		return fmt.Errorf("asset not found: %w", err)
	}

	// Check if asset is in use
	inUse, err := s.balanceRepo.IsAssetInUse(ctx, currencyCode)
	if err != nil {
		return fmt.Errorf("failed to check asset usage: %w", err)
	}

	if inUse {
		return fmt.Errorf("cannot deactivate asset %s: it is currently in use", currencyCode)
	}

	// Deactivate asset
	asset.IsActive = false
	asset.UpdatedAt = time.Now()

	if err := s.assetRepo.UpdateAsset(ctx, asset); err != nil {
		return fmt.Errorf("failed to deactivate asset: %w", err)
	}

	// Publish deactivation event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "asset.deactivated",
			Source: "asset-service",
			Data: map[string]interface{}{
				"currency_code": asset.CurrencyCode,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(event); err != nil {
			fmt.Printf("Warning: Failed to publish asset deactivation event: %v\n", err)
		}
	}

	return nil
}

// ValidateAssetForTransaction validates if an asset can be used for a transaction
func (s *AssetService) ValidateAssetForTransaction(ctx context.Context, currencyCode string, amount string) error {
	asset, err := s.GetAssetByCurrency(ctx, currencyCode)
	if err != nil {
		return fmt.Errorf("unsupported currency: %s", currencyCode)
	}

	if !asset.IsActive {
		return fmt.Errorf("currency %s is currently deactivated", currencyCode)
	}

	// Validate amount format
	amountBig := new(big.Int)
	_, ok := amountBig.SetString(amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	// Check minimum amount
	minAmount, err := asset.GetMinimumAmountBigInt()
	if err != nil {
		return fmt.Errorf("invalid minimum amount configuration: %w", err)
	}

	if amountBig.Cmp(minAmount) < 0 {
		return fmt.Errorf("amount %s is below minimum %s for currency %s", amount, asset.MinimumAmount, currencyCode)
	}

	// Check maximum amount if configured
	if asset.MaximumAmount != nil {
		maxAmount := new(big.Int)
		_, ok := maxAmount.SetString(*asset.MaximumAmount, 10)
		if !ok {
			return fmt.Errorf("invalid maximum amount configuration for currency %s", currencyCode)
		}

		if amountBig.Cmp(maxAmount) > 0 {
			return fmt.Errorf("amount %s exceeds maximum %s for currency %s", amount, *asset.MaximumAmount, currencyCode)
		}
	}

	return nil
}

// GetAssetConfiguration returns configuration details for supported assets
func (s *AssetService) GetAssetConfiguration(ctx context.Context) (map[string]*models.SupportedAsset, error) {
	assets, err := s.GetSupportedAssets(ctx, true)
	if err != nil {
		return nil, err
	}

	config := make(map[string]*models.SupportedAsset)
	for _, asset := range assets {
		config[asset.CurrencyCode] = asset
	}

	return config, nil
}

// InitializeDefaultAssets creates the default supported assets if they don't exist
func (s *AssetService) InitializeDefaultAssets(ctx context.Context) error {
	// Check which assets already exist
	existingAssets, err := s.GetSupportedAssets(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get existing assets: %w", err)
	}

	existingCodes := make(map[string]bool)
	for _, asset := range existingAssets {
		existingCodes[asset.CurrencyCode] = true
	}

	// Create default assets that don't exist
	defaultAssets := []AssetRegistryRequest{
		{
			CurrencyCode:  "XRP",
			CurrencyName:  "XRP",
			AssetType:     "native",
			DecimalPlaces: 6,
			MinimumAmount: "1",
			Description:   stringPtr("Native XRP cryptocurrency"),
		},
		{
			CurrencyCode:  "USDT",
			CurrencyName:  "Tether USD",
			AssetType:     "stablecoin",
			DecimalPlaces: 6,
			MinimumAmount: "10000", // 0.01 USDT in microunits
			Description:   stringPtr("USD-pegged stablecoin issued by Tether"),
		},
		{
			CurrencyCode:  "USDC",
			CurrencyName:  "USD Coin",
			AssetType:     "stablecoin",
			DecimalPlaces: 6,
			MinimumAmount: "10000", // 0.01 USDC in microunits
			Description:   stringPtr("USD-pegged stablecoin issued by Centre"),
		},
		{
			CurrencyCode:  "e₹",
			CurrencyName:  "Digital Rupee",
			AssetType:     "cbdc",
			DecimalPlaces: 2,
			MinimumAmount: "1", // 0.01 INR in paisa
			Description:   stringPtr("Central Bank Digital Currency issued by Reserve Bank of India"),
		},
	}

	for _, req := range defaultAssets {
		if !existingCodes[req.CurrencyCode] {
			_, err := s.RegisterAsset(ctx, &req)
			if err != nil {
				return fmt.Errorf("failed to create default asset %s: %w", req.CurrencyCode, err)
			}
		}
	}

	return nil
}

// GetAssetStatistics returns usage statistics for assets
func (s *AssetService) GetAssetStatistics(ctx context.Context) (map[string]interface{}, error) {
	assets, err := s.GetSupportedAssets(ctx, false)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_assets":      len(assets),
		"active_assets":     0,
		"assets_by_type":    make(map[string]int),
		"native_assets":     0,
		"stablecoin_assets": 0,
		"cbdc_assets":       0,
		"wrapped_assets":    0,
	}

	assetsByType := stats["assets_by_type"].(map[string]int)

	for _, asset := range assets {
		if asset.IsActive {
			stats["active_assets"] = stats["active_assets"].(int) + 1
		}

		assetType := string(asset.AssetType)
		assetsByType[assetType]++

		switch asset.AssetType {
		case models.AssetTypeNative:
			stats["native_assets"] = stats["native_assets"].(int) + 1
		case models.AssetTypeStablecoin:
			stats["stablecoin_assets"] = stats["stablecoin_assets"].(int) + 1
		case models.AssetTypeCBDC:
			stats["cbdc_assets"] = stats["cbdc_assets"].(int) + 1
		case models.AssetTypeWrapped:
			stats["wrapped_assets"] = stats["wrapped_assets"].(int) + 1
		}
	}

	return stats, nil
}

// helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
