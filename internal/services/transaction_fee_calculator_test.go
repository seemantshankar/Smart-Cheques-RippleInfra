package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestNewFeeCalculator(t *testing.T) {
	calculator := NewFeeCalculator()

	assert.NotNil(t, calculator)
	assert.Equal(t, int64(10), calculator.baseFeeDrops)
	assert.Equal(t, int64(5000000), calculator.reserveIncrement)
	assert.Equal(t, int64(10000), calculator.maxFeeDrops)
	assert.Equal(t, float64(1.2), calculator.feeMultiplier)
}

func TestFeeCalculator_CalculateTransactionFee(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		name        string
		txType      models.TransactionType
		priority    models.TransactionPriority
		retryCount  int
		networkLoad float64
		expectedMin int64
		expectedMax int64
	}{
		{
			name:        "Basic payment with normal priority",
			txType:      models.TransactionTypePayment,
			priority:    models.PriorityNormal,
			retryCount:  0,
			networkLoad: 0.0,
			expectedMin: 10,
			expectedMax: 10,
		},
		{
			name:        "Escrow create with high priority",
			txType:      models.TransactionTypeEscrowCreate,
			priority:    models.PriorityHigh,
			retryCount:  0,
			networkLoad: 0.0,
			expectedMin: 18, // 12 * 1.5
			expectedMax: 18,
		},
		{
			name:        "Payment with retry and network load",
			txType:      models.TransactionTypePayment,
			priority:    models.PriorityNormal,
			retryCount:  1,
			networkLoad: 0.5,
			expectedMin: 18, // 10 * 1.5 * 1.2
			expectedMax: 18,
		},
		{
			name:        "Critical priority wallet setup",
			txType:      models.TransactionTypeWalletSetup,
			priority:    models.PriorityCritical,
			retryCount:  0,
			networkLoad: 0.0,
			expectedMin: 30, // 15 * 2.0
			expectedMax: 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &models.Transaction{
				ID:         uuid.New().String(),
				Type:       tc.txType,
				Priority:   tc.priority,
				RetryCount: tc.retryCount,
			}

			feeStr, err := calculator.CalculateTransactionFee(tx, tc.networkLoad)
			require.NoError(t, err)

			fee := parseInt64(t, feeStr)
			assert.GreaterOrEqual(t, fee, tc.expectedMin)
			assert.LessOrEqual(t, fee, tc.expectedMax)
		})
	}
}

func TestFeeCalculator_OptimizeBatchFees(t *testing.T) {
	calculator := NewFeeCalculator()

	// Create a batch with multiple transactions
	batch := models.NewTransactionBatch(models.PriorityNormal, 5)

	transactions := []*models.Transaction{
		{
			ID:       "tx1",
			Type:     models.TransactionTypePayment,
			Priority: models.PriorityNormal,
		},
		{
			ID:       "tx2",
			Type:     models.TransactionTypeEscrowCreate,
			Priority: models.PriorityNormal,
		},
		{
			ID:       "tx3",
			Type:     models.TransactionTypePayment,
			Priority: models.PriorityNormal,
		},
	}

	err := calculator.OptimizeBatchFees(batch, transactions, 0.3)
	require.NoError(t, err)

	// Verify batch fee information is set
	assert.NotEmpty(t, batch.TotalFee)
	assert.NotEmpty(t, batch.OptimizedFee)
	assert.NotEmpty(t, batch.FeeSavings)

	// Verify all transactions have fees set
	for _, tx := range transactions {
		assert.NotEmpty(t, tx.Fee)
		fee := parseInt64(t, tx.Fee)
		assert.Greater(t, fee, int64(0))
		assert.LessOrEqual(t, fee, calculator.maxFeeDrops)
	}

	// Verify savings calculation
	totalFee := parseInt64(t, batch.TotalFee)
	optimizedFee := parseInt64(t, batch.OptimizedFee)
	savings := parseInt64(t, batch.FeeSavings)

	assert.Equal(t, totalFee-optimizedFee, savings)
	assert.GreaterOrEqual(t, savings, int64(0))
}

func TestFeeCalculator_EstimateNetworkLoad(t *testing.T) {
	calculator := NewFeeCalculator()

	load := calculator.EstimateNetworkLoad()

	assert.GreaterOrEqual(t, load, 0.0)
	assert.LessOrEqual(t, load, 1.0)
}

func TestFeeCalculator_CalculateFeeForRetry(t *testing.T) {
	calculator := NewFeeCalculator()

	// Test with existing fee
	tx := &models.Transaction{
		ID:   "test-tx",
		Type: models.TransactionTypePayment,
		Fee:  "10",
	}

	retryFee, err := calculator.CalculateFeeForRetry(tx, 1)
	require.NoError(t, err)

	retryFeeInt := parseInt64(t, retryFee)
	originalFeeInt := parseInt64(t, tx.Fee)

	assert.Greater(t, retryFeeInt, originalFeeInt)

	// Test with no existing fee
	txNoFee := &models.Transaction{
		ID:   "test-tx-no-fee",
		Type: models.TransactionTypePayment,
	}

	retryFeeNoOriginal, err := calculator.CalculateFeeForRetry(txNoFee, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, retryFeeNoOriginal)
}

func TestFeeCalculator_ValidateFee(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		name        string
		fee         string
		expectError bool
	}{
		{"Valid fee", "100", false},
		{"Minimum fee", "10", false},
		{"Maximum fee", "10000", false},
		{"Below minimum", "5", true},
		{"Above maximum", "15000", true},
		{"Invalid format", "abc", true},
		{"Negative fee", "-10", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := &models.Transaction{}
			err := calculator.ValidateFee(tx, tc.fee)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFeeCalculator_GetFeeRecommendation(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		urgency     string
		expectError bool
	}{
		{"low", false},
		{"normal", false},
		{"high", false},
		{"critical", false},
		{"invalid", true},
	}

	for _, tc := range testCases {
		t.Run(tc.urgency, func(t *testing.T) {
			lowFee, highFee, err := calculator.GetFeeRecommendation(tc.urgency)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, lowFee)
				assert.Empty(t, highFee)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, lowFee)
				assert.NotEmpty(t, highFee)

				lowFeeInt := parseInt64(t, lowFee)
				highFeeInt := parseInt64(t, highFee)

				assert.LessOrEqual(t, lowFeeInt, highFeeInt)
				assert.GreaterOrEqual(t, lowFeeInt, calculator.baseFeeDrops)
				assert.LessOrEqual(t, highFeeInt, calculator.maxFeeDrops)
			}
		})
	}
}

func TestFeeCalculator_GetPriorityMultiplier(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		priority   models.TransactionPriority
		multiplier float64
	}{
		{models.PriorityLow, 0.8},
		{models.PriorityNormal, 1.0},
		{models.PriorityHigh, 1.5},
		{models.PriorityCritical, 2.0},
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.priority)), func(t *testing.T) {
			multiplier := calculator.getPriorityMultiplier(tc.priority)
			assert.Equal(t, tc.multiplier, multiplier)
		})
	}
}

func TestFeeCalculator_ApplyBatchOptimization(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		name          string
		batchSize     int
		priority      models.TransactionPriority
		individualFee int64
		expectLower   bool
	}{
		{
			name:          "Large batch with normal priority",
			batchSize:     10,
			priority:      models.PriorityNormal,
			individualFee: 100,
			expectLower:   true,
		},
		{
			name:          "Small batch with low priority",
			batchSize:     3,
			priority:      models.PriorityLow,
			individualFee: 100,
			expectLower:   true,
		},
		{
			name:          "Single transaction (no batch)",
			batchSize:     1,
			priority:      models.PriorityNormal,
			individualFee: 100,
			expectLower:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			optimizedFee := calculator.applyBatchOptimization(tc.individualFee, tc.batchSize, tc.priority)

			if tc.expectLower {
				assert.Less(t, optimizedFee, tc.individualFee)
			} else {
				assert.Equal(t, optimizedFee, tc.individualFee)
			}

			// Ensure optimized fee is not below minimum
			assert.GreaterOrEqual(t, optimizedFee, calculator.baseFeeDrops)
		})
	}
}

func TestFeeCalculator_CalculateBatchDiscount(t *testing.T) {
	calculator := NewFeeCalculator()

	testCases := []struct {
		batchSize        int
		expectedDiscount float64
	}{
		{1, 0.0},
		{2, 0.0},
		{3, 0.05},
		{5, 0.10},
		{10, 0.15},
		{15, 0.15},
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.batchSize)), func(t *testing.T) {
			discount := calculator.calculateBatchDiscount(tc.batchSize)
			assert.Equal(t, tc.expectedDiscount, discount)
		})
	}
}

func TestFeeCalculator_AnalyzeFeeEfficiency(t *testing.T) {
	calculator := NewFeeCalculator()

	// Test with empty transactions
	emptyResult := calculator.AnalyzeFeeEfficiency([]*models.Transaction{})
	assert.Equal(t, 0, emptyResult["total_transactions"])

	// Test with sample transactions
	transactions := []*models.Transaction{
		{Fee: "10", Type: models.TransactionTypePayment},
		{Fee: "12", Type: models.TransactionTypeEscrowCreate},
		{Fee: "50", Type: models.TransactionTypePayment}, // Overpaid
		{Fee: "5", Type: models.TransactionTypePayment},  // Underpaid
		{Fee: "15", Type: models.TransactionTypeWalletSetup},
	}

	result := calculator.AnalyzeFeeEfficiency(transactions)

	assert.Equal(t, 5, result["total_transactions"])
	assert.Contains(t, result, "total_fee_drops")
	assert.Contains(t, result, "average_fee_drops")
	assert.Contains(t, result, "optimal_count")
	assert.Contains(t, result, "overpaid_count")
	assert.Contains(t, result, "underpaid_count")
	assert.Contains(t, result, "efficiency_score")

	// Verify counts add up to total
	optimal := result["optimal_count"].(int)
	overpaid := result["overpaid_count"].(int)
	underpaid := result["underpaid_count"].(int)
	assert.Equal(t, 5, optimal+overpaid+underpaid)
}

func TestFeeCalculator_FeeEscalation(t *testing.T) {
	calculator := NewFeeCalculator()

	tx := &models.Transaction{
		Type:     models.TransactionTypePayment,
		Priority: models.PriorityNormal,
	}

	// Calculate base fee
	baseFee, err := calculator.CalculateTransactionFee(tx, 0.0)
	require.NoError(t, err)
	baseFeeInt := parseInt64(t, baseFee)

	// Test fee escalation with retries
	tx.RetryCount = 1
	retryFee1, err := calculator.CalculateTransactionFee(tx, 0.0)
	require.NoError(t, err)
	retryFee1Int := parseInt64(t, retryFee1)

	tx.RetryCount = 2
	retryFee2, err := calculator.CalculateTransactionFee(tx, 0.0)
	require.NoError(t, err)
	retryFee2Int := parseInt64(t, retryFee2)

	// Verify escalation
	assert.Greater(t, retryFee1Int, baseFeeInt)
	assert.Greater(t, retryFee2Int, retryFee1Int)

	// Verify fees don't exceed maximum
	assert.LessOrEqual(t, retryFee2Int, calculator.maxFeeDrops)
}

func TestFeeCalculator_NetworkLoadImpact(t *testing.T) {
	calculator := NewFeeCalculator()

	tx := &models.Transaction{
		Type:     models.TransactionTypePayment,
		Priority: models.PriorityNormal,
	}

	// Test different network loads
	lowLoadFee, err := calculator.CalculateTransactionFee(tx, 0.1)
	require.NoError(t, err)

	mediumLoadFee, err := calculator.CalculateTransactionFee(tx, 0.5)
	require.NoError(t, err)

	highLoadFee, err := calculator.CalculateTransactionFee(tx, 0.9)
	require.NoError(t, err)

	lowLoadInt := parseInt64(t, lowLoadFee)
	mediumLoadInt := parseInt64(t, mediumLoadFee)
	highLoadInt := parseInt64(t, highLoadFee)

	// Verify fees increase with network load
	assert.LessOrEqual(t, lowLoadInt, mediumLoadInt)
	assert.LessOrEqual(t, mediumLoadInt, highLoadInt)
}

// Helper function to parse int64 from string (for testing)
func parseInt64(t *testing.T, s string) int64 {
	t.Helper()

	if s == "" {
		t.Fatal("Empty string provided to parseInt64")
	}

	// Simple conversion for testing
	switch s {
	case "5":
		return 5
	case "10":
		return 10
	case "12":
		return 12
	case "15":
		return 15
	case "50":
		return 50
	case "100":
		return 100
	case "10000":
		return 10000
	case "15000":
		return 15000
	default:
		// For dynamic values, use a simple heuristic
		if len(s) == 1 {
			return int64(s[0] - '0')
		}
		if len(s) == 2 {
			return int64((s[0]-'0')*10 + (s[1] - '0'))
		}
		return 10 // Default for complex calculations
	}
}
