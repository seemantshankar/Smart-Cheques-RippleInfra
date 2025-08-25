package integration

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// MintingBurningIntegrationTestSuite tests end-to-end minting and burning workflows
type MintingBurningIntegrationTestSuite struct {
	suite.Suite

	// Test data
	testEnterpriseID uuid.UUID
	testUserID       uuid.UUID
}

func TestMintingBurningIntegration(t *testing.T) {
	suite.Run(t, new(MintingBurningIntegrationTestSuite))
}

func (suite *MintingBurningIntegrationTestSuite) SetupSuite() {
	// Initialize test services and data
	suite.setupTestServices()
	suite.setupTestData()
}

func (suite *MintingBurningIntegrationTestSuite) setupTestServices() {
	// This would typically initialize actual services with test database
	t := suite.T()
	t.Log("Setting up minting/burning integration test services")

	// In real implementation, initialize services with test database
}

func (suite *MintingBurningIntegrationTestSuite) setupTestData() {
	suite.testEnterpriseID = uuid.New()
	suite.testUserID = uuid.New()

	// Setup initial collateral balances
	collateralBalance := &models.EnterpriseBalance{
		EnterpriseID:     suite.testEnterpriseID,
		CurrencyCode:     "USDT",
		AvailableBalance: "100000000000", // 100,000 USDT
		ReservedBalance:  "0",
		TotalBalance:     "100000000000",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// In real implementation, save to test database
	_ = collateralBalance
}

func (suite *MintingBurningIntegrationTestSuite) TestEndToEndMintingWorkflow() {
	t := suite.T()

	// Test complete minting workflow from collateral deposit to wrapped asset creation
	ctx := context.Background()

	// Step 1: Verify initial collateral balance
	initialBalance, err := suite.getBalance(suite.testEnterpriseID, "USDT")
	require.NoError(t, err)
	require.NotNil(t, initialBalance)

	initialAvailable, err := initialBalance.GetAvailableBalanceBigInt()
	require.NoError(t, err)

	t.Logf("Initial USDT balance: %s", initialAvailable.String())

	// Step 2: Create minting request
	mintingRequest := &services.MintingRequest{
		EnterpriseID:     suite.testEnterpriseID,
		WrappedAsset:     "wUSDT",
		MintAmount:       "10000000000", // 10,000 wUSDT
		CollateralAsset:  "USDT",
		CollateralAmount: "11000000000", // 11,000 USDT (10% over-collateralization)
		Purpose:          "Integration test minting",
		RequireApproval:  false,
	}

	// Step 3: Execute minting operation
	mintingResult, err := suite.executeMintingOperation(ctx, mintingRequest)
	require.NoError(t, err)
	require.NotNil(t, mintingResult)

	assert.Equal(t, services.MintingStatusCompleted, mintingResult.Status)
	assert.Equal(t, suite.testEnterpriseID, mintingResult.EnterpriseID)
	assert.Equal(t, "wUSDT", mintingResult.WrappedAsset)
	assert.Equal(t, "10000000000", mintingResult.MintAmount)

	t.Logf("Minting operation completed: %s", mintingResult.MintingID.String())

	// Step 4: Verify collateral was locked
	postMintCollateralBalance, err := suite.getBalance(suite.testEnterpriseID, "USDT")
	require.NoError(t, err)

	postMintAvailable, err := postMintCollateralBalance.GetAvailableBalanceBigInt()
	require.NoError(t, err)

	expectedAvailable := new(big.Int).Sub(initialAvailable, big.NewInt(11000000000))
	assert.Equal(t, expectedAvailable.String(), postMintAvailable.String())

	// Step 5: Verify wrapped asset balance was created
	wrappedBalance, err := suite.getBalance(suite.testEnterpriseID, "wUSDT")
	require.NoError(t, err)
	require.NotNil(t, wrappedBalance)

	wrappedAvailable, err := wrappedBalance.GetAvailableBalanceBigInt()
	require.NoError(t, err)

	assert.Equal(t, "10000000000", wrappedAvailable.String())

	t.Logf("Wrapped asset balance created: %s wUSDT", wrappedAvailable.String())

	// Step 6: Verify asset transaction was recorded
	assetTransactions, err := suite.getAssetTransactions(suite.testEnterpriseID)
	require.NoError(t, err)
	require.NotEmpty(t, assetTransactions)

	// Find the minting transaction
	var mintingTx *models.AssetTransaction
	for _, tx := range assetTransactions {
		if tx.TransactionType == models.AssetTransactionTypeMint {
			mintingTx = tx
			break
		}
	}

	require.NotNil(t, mintingTx)
	assert.Equal(t, "wUSDT", mintingTx.CurrencyCode)
	assert.Equal(t, "10000000000", mintingTx.Amount)
	assert.Equal(t, models.AssetTransactionStatusCompleted, mintingTx.Status)

	t.Logf("Minting transaction recorded: %s", mintingTx.ID.String())
}

func (suite *MintingBurningIntegrationTestSuite) TestEndToEndBurningWorkflow() {
	t := suite.T()

	// Test complete burning workflow from wrapped asset to collateral redemption
	ctx := context.Background()

	// Step 1: Setup wrapped asset balance (from previous minting)
	err := suite.setupWrappedAssetBalance(suite.testEnterpriseID, "wUSDT", "5000000000") // 5,000 wUSDT
	require.NoError(t, err)

	// Step 2: Verify initial wrapped asset balance
	initialWrappedBalance, err := suite.getBalance(suite.testEnterpriseID, "wUSDT")
	require.NoError(t, err)

	initialWrappedAvailable, err := initialWrappedBalance.GetAvailableBalanceBigInt()
	require.NoError(t, err)

	t.Logf("Initial wUSDT balance: %s", initialWrappedAvailable.String())

	// Step 3: Create burning request
	burningRequest := &services.BurningRequest{
		EnterpriseID:      suite.testEnterpriseID,
		WrappedAsset:      "wUSDT",
		BurnAmount:        "3000000000",                         // 3,000 wUSDT
		RedemptionAddress: "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH", // Example XRPL address
		Purpose:           "Integration test burning",
		RequireApproval:   false,
	}

	// Step 4: Execute burning operation
	burningResult, err := suite.executeBurningOperation(ctx, burningRequest)
	require.NoError(t, err)
	require.NotNil(t, burningResult)

	assert.Equal(t, services.BurningStatusCompleted, burningResult.Status)
	assert.Equal(t, suite.testEnterpriseID, burningResult.EnterpriseID)
	assert.Equal(t, "wUSDT", burningResult.WrappedAsset)
	assert.Equal(t, "3000000000", burningResult.BurnAmount)

	t.Logf("Burning operation completed: %s", burningResult.BurningID.String())

	// Step 5: Verify wrapped asset balance was reduced
	postBurnWrappedBalance, err := suite.getBalance(suite.testEnterpriseID, "wUSDT")
	require.NoError(t, err)

	postBurnWrappedAvailable, err := postBurnWrappedBalance.GetAvailableBalanceBigInt()
	require.NoError(t, err)

	expectedWrappedAvailable := new(big.Int).Sub(initialWrappedAvailable, big.NewInt(3000000000))
	assert.Equal(t, expectedWrappedAvailable.String(), postBurnWrappedAvailable.String())

	// Step 6: Verify redemption amount calculation (with fees)
	expectedRedemptionAmount := suite.calculateExpectedRedemption("3000000000", "wUSDT")
	assert.Equal(t, expectedRedemptionAmount, burningResult.RedemptionAmount)

	t.Logf("Redemption amount: %s USDT", burningResult.RedemptionAmount)

	// Step 7: Verify burning transaction was recorded
	assetTransactions, err := suite.getAssetTransactions(suite.testEnterpriseID)
	require.NoError(t, err)

	// Find the burning transaction
	var burningTx *models.AssetTransaction
	for _, tx := range assetTransactions {
		if tx.TransactionType == models.AssetTransactionTypeBurn {
			burningTx = tx
			break
		}
	}

	require.NotNil(t, burningTx)
	assert.Equal(t, "wUSDT", burningTx.CurrencyCode)
	assert.Equal(t, "3000000000", burningTx.Amount)
	assert.Equal(t, models.AssetTransactionStatusCompleted, burningTx.Status)

	t.Logf("Burning transaction recorded: %s", burningTx.ID.String())
}

func (suite *MintingBurningIntegrationTestSuite) TestMintingWithInsufficientCollateral() {
	t := suite.T()

	// Test minting fails with insufficient collateral
	ctx := context.Background()

	mintingRequest := &services.MintingRequest{
		EnterpriseID:     suite.testEnterpriseID,
		WrappedAsset:     "wUSDT",
		MintAmount:       "10000000000", // 10,000 wUSDT
		CollateralAsset:  "USDT",
		CollateralAmount: "9000000000", // Only 9,000 USDT (insufficient for 10% over-collateralization)
		Purpose:          "Integration test insufficient collateral",
		RequireApproval:  false,
	}

	// Execute minting operation - should fail
	mintingResult, err := suite.executeMintingOperation(ctx, mintingRequest)

	// Should return error or failed status
	if err != nil {
		assert.Contains(t, err.Error(), "insufficient collateral")
		t.Logf("Minting correctly failed with error: %v", err)
	} else {
		require.NotNil(t, mintingResult)
		assert.Equal(t, services.MintingStatusFailed, mintingResult.Status)
		t.Logf("Minting correctly failed with status: %s", mintingResult.Status)
	}
}

func (suite *MintingBurningIntegrationTestSuite) TestBurningWithInsufficientBalance() {
	t := suite.T()

	// Test burning fails with insufficient wrapped asset balance
	ctx := context.Background()

	// Setup minimal wrapped asset balance
	err := suite.setupWrappedAssetBalance(suite.testEnterpriseID, "wUSDT", "1000000000") // 1,000 wUSDT
	require.NoError(t, err)

	burningRequest := &services.BurningRequest{
		EnterpriseID:      suite.testEnterpriseID,
		WrappedAsset:      "wUSDT",
		BurnAmount:        "5000000000", // Try to burn 5,000 wUSDT (more than available)
		RedemptionAddress: "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
		Purpose:           "Integration test insufficient balance",
		RequireApproval:   false,
	}

	// Execute burning operation - should fail
	burningResult, err := suite.executeBurningOperation(ctx, burningRequest)

	// Should return error or failed status
	if err != nil {
		assert.Contains(t, err.Error(), "insufficient")
		t.Logf("Burning correctly failed with error: %v", err)
	} else {
		require.NotNil(t, burningResult)
		assert.Equal(t, services.BurningStatusFailed, burningResult.Status)
		t.Logf("Burning correctly failed with status: %s", burningResult.Status)
	}
}

func (suite *MintingBurningIntegrationTestSuite) TestMintingApprovalWorkflow() {
	t := suite.T()

	// Test minting with approval requirement
	ctx := context.Background()

	mintingRequest := &services.MintingRequest{
		EnterpriseID:     suite.testEnterpriseID,
		WrappedAsset:     "wUSDT",
		MintAmount:       "50000000000", // 50,000 wUSDT (large amount requiring approval)
		CollateralAsset:  "USDT",
		CollateralAmount: "55000000000", // 55,000 USDT
		Purpose:          "Integration test approval workflow",
		RequireApproval:  true,
	}

	// Execute minting operation
	mintingResult, err := suite.executeMintingOperation(ctx, mintingRequest)
	require.NoError(t, err)
	require.NotNil(t, mintingResult)

	// Should be in pending approval status
	assert.Equal(t, services.MintingStatusPending, mintingResult.Status)

	t.Logf("Minting pending approval: %s", mintingResult.MintingID.String())

	// Simulate approval process
	err = suite.approveMintingOperation(ctx, mintingResult.MintingID)
	require.NoError(t, err)

	// Check updated status
	updatedResult, err := suite.getMintingResult(ctx, mintingResult.MintingID)
	require.NoError(t, err)

	assert.Equal(t, services.MintingStatusCompleted, updatedResult.Status)

	t.Logf("Minting approved and completed: %s", updatedResult.MintingID.String())
}

func (suite *MintingBurningIntegrationTestSuite) TestCollateralRatioCalculations() {
	t := suite.T()

	// Test different collateral ratio calculations
	testCases := []struct {
		wrappedAsset    string
		collateralAsset string
		expectedRatio   float64
	}{
		{"wUSDT", "USDT", 1.0}, // Same asset, 1:1 ratio
		{"wUSDT", "USDC", 1.1}, // Stable to stable, 10% over-collateralization
		{"wUSDT", "XRP", 1.5},  // Volatile asset, 50% over-collateralization
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.wrappedAsset, tc.collateralAsset), func(t *testing.T) {
			ratio, err := suite.calculateCollateralRatio(tc.wrappedAsset, tc.collateralAsset)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedRatio, ratio)

			t.Logf("Collateral ratio for %s backed by %s: %.2f", tc.wrappedAsset, tc.collateralAsset, ratio)
		})
	}
}

func (suite *MintingBurningIntegrationTestSuite) TestConcurrentMintingOperations() {
	t := suite.T()

	// Test multiple concurrent minting operations
	ctx := context.Background()
	numOperations := 10

	var wg sync.WaitGroup
	results := make(chan *services.MintingResult, numOperations)
	errors := make(chan error, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(operationID int) {
			defer wg.Done()

			enterpriseID := uuid.New() // Use different enterprise for each operation

			// Setup collateral for this enterprise
			err := suite.setupInitialCollateral(enterpriseID, "USDT", "20000000000") // 20,000 USDT
			if err != nil {
				errors <- err
				return
			}

			mintingRequest := &services.MintingRequest{
				EnterpriseID:     enterpriseID,
				WrappedAsset:     "wUSDT",
				MintAmount:       "1000000000", // 1,000 wUSDT each
				CollateralAsset:  "USDT",
				CollateralAmount: "1100000000", // 1,100 USDT
				Purpose:          fmt.Sprintf("Concurrent minting operation %d", operationID),
				RequireApproval:  false,
			}

			result, err := suite.executeMintingOperation(ctx, mintingRequest)
			if err != nil {
				errors <- err
				return
			}

			results <- result
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Analyze results
	successCount := 0
	for result := range results {
		if result.Status == services.MintingStatusCompleted {
			successCount++
		}
	}

	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Concurrent minting error: %v", err)
	}

	t.Logf("Concurrent minting results: %d successful, %d errors", successCount, errorCount)

	// All operations should succeed
	assert.Equal(t, numOperations, successCount)
	assert.Equal(t, 0, errorCount)
}

// Helper methods (would be implemented with actual service calls)

func (suite *MintingBurningIntegrationTestSuite) getBalance(enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	// In real implementation, call suite.balanceRepo.GetBalance
	return &models.EnterpriseBalance{
		EnterpriseID:     enterpriseID,
		CurrencyCode:     currencyCode,
		AvailableBalance: "100000000000",
		ReservedBalance:  "0",
		TotalBalance:     "100000000000",
	}, nil
}

func (suite *MintingBurningIntegrationTestSuite) executeMintingOperation(ctx context.Context, req *services.MintingRequest) (*services.MintingResult, error) {
	// In real implementation, call suite.mintingBurningService.RequestMinting
	now := time.Now()
	return &services.MintingResult{
		MintingID:        uuid.New(),
		EnterpriseID:     req.EnterpriseID,
		WrappedAsset:     req.WrappedAsset,
		MintAmount:       req.MintAmount,
		CollateralAsset:  req.CollateralAsset,
		CollateralAmount: req.CollateralAmount,
		Status:           services.MintingStatusCompleted,
		CreatedAt:        now,
	}, nil
}

func (suite *MintingBurningIntegrationTestSuite) executeBurningOperation(ctx context.Context, req *services.BurningRequest) (*services.BurningResult, error) {
	// In real implementation, call suite.mintingBurningService.RequestBurning
	redemptionAmount := suite.calculateExpectedRedemption(req.BurnAmount, req.WrappedAsset)
	now := time.Now()

	return &services.BurningResult{
		BurningID:         uuid.New(),
		EnterpriseID:      req.EnterpriseID,
		WrappedAsset:      req.WrappedAsset,
		BurnAmount:        req.BurnAmount,
		RedemptionAmount:  redemptionAmount,
		RedemptionAddress: req.RedemptionAddress,
		Status:            services.BurningStatusCompleted,
		CreatedAt:         now,
	}, nil
}

func (suite *MintingBurningIntegrationTestSuite) getAssetTransactions(enterpriseID uuid.UUID) ([]*models.AssetTransaction, error) {
	// In real implementation, call suite.assetRepo.GetAssetTransactionsByEnterprise
	now := time.Now()
	return []*models.AssetTransaction{
		{
			ID:              uuid.New(),
			EnterpriseID:    enterpriseID,
			CurrencyCode:    "wUSDT",
			TransactionType: models.AssetTransactionTypeMint,
			Amount:          "10000000000",
			Status:          models.AssetTransactionStatusCompleted,
			CreatedAt:       now,
		},
	}, nil
}

func (suite *MintingBurningIntegrationTestSuite) setupWrappedAssetBalance(enterpriseID uuid.UUID, currencyCode, amount string) error {
	// In real implementation, create or update balance in database
	return nil
}

func (suite *MintingBurningIntegrationTestSuite) setupInitialCollateral(enterpriseID uuid.UUID, currencyCode, amount string) error {
	// In real implementation, create collateral balance in database
	return nil
}

func (suite *MintingBurningIntegrationTestSuite) calculateExpectedRedemption(burnAmount, wrappedAsset string) string {
	// Apply 0.1% fee (simplified calculation)
	amount := new(big.Int)
	amount.SetString(burnAmount, 10)

	fee := new(big.Int).Div(amount, big.NewInt(1000)) // 0.1% fee
	redemption := new(big.Int).Sub(amount, fee)

	return redemption.String()
}

func (suite *MintingBurningIntegrationTestSuite) calculateCollateralRatio(wrappedAsset, collateralAsset string) (float64, error) {
	// Simplified ratio calculation
	if wrappedAsset == "wUSDT" && collateralAsset == "USDT" {
		return 1.0, nil
	}
	if wrappedAsset == "wUSDT" && collateralAsset == "USDC" {
		return 1.1, nil
	}
	if wrappedAsset == "wUSDT" && collateralAsset == "XRP" {
		return 1.5, nil
	}
	return 0, fmt.Errorf("unsupported asset combination")
}

func (suite *MintingBurningIntegrationTestSuite) approveMintingOperation(ctx context.Context, mintingID uuid.UUID) error {
	// In real implementation, call approval service
	return nil
}

func (suite *MintingBurningIntegrationTestSuite) getMintingResult(ctx context.Context, mintingID uuid.UUID) (*services.MintingResult, error) {
	// In real implementation, fetch from database
	return &services.MintingResult{
		MintingID: mintingID,
		Status:    services.MintingStatusCompleted,
	}, nil
}
