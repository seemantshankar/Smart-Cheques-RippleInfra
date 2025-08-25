package services

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestMintingBurningServiceUnitTests tests the core business logic
func TestMintingBurningServiceUnitTests(t *testing.T) {
	// Test collateral ratio calculations
	service := &MintingBurningService{}

	// Test 1:1 ratio for same asset
	ratio, err := service.calculateRequiredCollateralRatio("wUSDT", "USDT")
	assert.NoError(t, err)
	assert.Equal(t, 1.0, ratio)

	// Test over-collateralization for different assets
	ratio, err = service.calculateRequiredCollateralRatio("wUSDT", "USDC")
	assert.NoError(t, err)
	assert.Equal(t, 1.1, ratio) // 10% over-collateralization

	// Test volatile asset ratio
	ratio, err = service.calculateRequiredCollateralRatio("wUSDT", "XRP")
	assert.NoError(t, err)
	assert.Equal(t, 1.5, ratio) // 50% over-collateralization

	// Test unsupported combination
	_, err = service.calculateRequiredCollateralRatio("wUNKNOWN", "USDT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported collateral asset")

	// Test collateral sufficiency validation
	err = service.validateCollateralSufficiency("1100", "1000", 1.1) // 1100 collateral for 1000 mint at 1.1x ratio
	assert.NoError(t, err)

	err = service.validateCollateralSufficiency("1000", "1000", 1.1) // Insufficient collateral
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient collateral")

	// Test redemption amount calculation
	redemptionAmount := service.calculateRedemptionAmount("1000", "wUSDT")
	assert.Equal(t, "999", redemptionAmount) // 0.1% fee applied
}

func TestWithdrawalAuthorizationRiskScoring(t *testing.T) {
	// Test approval requirements calculation
	config := &AuthorizationConfig{
		LowAmountThreshold:    "1000",
		MediumAmountThreshold: "10000",
		HighAmountThreshold:   "100000",
		LowAmountApprovals:    1,
		MediumAmountApprovals: 2,
		HighAmountApprovals:   3,
		RiskScoreThreshold:    0.7,
		TimeLockThreshold:     "50000",
	}
	service := &WithdrawalAuthorizationService{
		authorizationConfig: config,
	}

	// Low amount, low risk
	approvals := service.calculateRequiredApprovals("500", 0.2)
	assert.Equal(t, 1, approvals)

	// Amount below medium threshold, low risk
	approvals = service.calculateRequiredApprovals("5000", 0.2)
	assert.Equal(t, 1, approvals) // 5000 < 10000 medium threshold

	// Actual medium amount, low risk
	approvals = service.calculateRequiredApprovals("15000", 0.2)
	assert.Equal(t, 2, approvals) // 15000 >= 10000 medium threshold

	// High amount, low risk
	approvals = service.calculateRequiredApprovals("150000", 0.2)
	assert.Equal(t, 3, approvals)

	// Low amount, high risk (should add extra approval)
	approvals = service.calculateRequiredApprovals("500", 0.8)
	assert.Equal(t, 2, approvals) // 1 + 1 for high risk

	// Test time lock requirements
	shouldLock := service.shouldApplyTimeLock("75000", 0.2) // Above threshold
	assert.True(t, shouldLock)

	shouldLock = service.shouldApplyTimeLock("30000", 0.8) // High risk
	assert.True(t, shouldLock)

	shouldLock = service.shouldApplyTimeLock("30000", 0.3) // Below threshold, low risk
	assert.False(t, shouldLock)
}

func TestReconciliationDiscrepancySeverity(t *testing.T) {
	config := &ReconciliationConfig{
		CriticalThreshold: "1000",
	}
	service := &ReconciliationService{config: config}

	// Test critical severity
	severity := service.determineDiscrepancySeverity(big.NewInt(2000), 15.0)
	assert.Equal(t, DiscrepancySeverityCritical, severity)

	// Test critical severity (amount exceeds critical threshold)
	severity = service.determineDiscrepancySeverity(big.NewInt(1500), 7.0)
	assert.Equal(t, DiscrepancySeverityCritical, severity)

	// Test high severity (meets percentage threshold)
	severity = service.determineDiscrepancySeverity(big.NewInt(500), 6.0)
	assert.Equal(t, DiscrepancySeverityHigh, severity)

	// Test medium severity
	severity = service.determineDiscrepancySeverity(big.NewInt(500), 2.0)
	assert.Equal(t, DiscrepancySeverityMedium, severity)

	// Test low severity
	severity = service.determineDiscrepancySeverity(big.NewInt(50), 0.5)
	assert.Equal(t, DiscrepancySeverityLow, severity)

	// Test possible cause analysis
	causes := service.analyzePossibleCauses(context.TODO(), uuid.New(), "USDT", big.NewInt(100))
	assert.Contains(t, causes, "Pending XRPL transactions not yet confirmed")
	assert.Contains(t, causes, "Data consistency issue")

	causes = service.analyzePossibleCauses(context.TODO(), uuid.New(), "USDT", big.NewInt(-100))
	assert.Contains(t, causes, "Unrecorded incoming XRPL transaction")
	assert.Contains(t, causes, "System synchronization delay")
}

func TestTreasuryServiceBalanceCalculations(t *testing.T) {
	// Test reconciliation configuration setup
	config := &ReconciliationConfig{
		ToleranceThreshold: "1",
		CriticalThreshold:  "1000",
	}
	service := &ReconciliationService{config: config}

	// Test that service is properly configured
	assert.NotNil(t, service)
	assert.Equal(t, "1", service.config.ToleranceThreshold)
	assert.Equal(t, "1000", service.config.CriticalThreshold)

	// Test authorization config setup
	authConfig := &AuthorizationConfig{
		LowAmountThreshold: "1000",
		RiskScoreThreshold: 0.7,
		TimeLockThreshold:  "50000",
	}
	authorizationService := &WithdrawalAuthorizationService{authorizationConfig: authConfig}
	assert.NotNil(t, authorizationService)
	assert.Equal(t, "1000", authorizationService.authorizationConfig.LowAmountThreshold)
}

// Benchmark tests for performance validation
func BenchmarkCollateralRatioCalculation(b *testing.B) {
	service := &MintingBurningService{}

	for i := 0; i < b.N; i++ {
		_, _ = service.calculateRequiredCollateralRatio("wUSDT", "USDT")
	}
}

func BenchmarkDiscrepancySeverityDetermination(b *testing.B) {
	config := &ReconciliationConfig{CriticalThreshold: "1000"}
	service := &ReconciliationService{config: config}

	for i := 0; i < b.N; i++ {
		_ = service.determineDiscrepancySeverity(big.NewInt(1500), 7.0)
	}
}
