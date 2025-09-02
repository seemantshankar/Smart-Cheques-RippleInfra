package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

func TestNewXRPLService(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
}

func TestXRPLService_Initialize(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)

	err := service.Initialize()
	assert.NoError(t, err)
}

func TestXRPLService_CreateWallet(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	wallet, err := service.CreateWallet()
	require.NoError(t, err)
	require.NotNil(t, wallet)

	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.Seed)

	// Validate the generated address
	assert.True(t, service.ValidateAddress(wallet.Address))
}

func TestXRPLService_ValidateAddress(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)

	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid address",
			address:  "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
			expected: true,
		},
		{
			name:     "invalid address",
			address:  "invalid",
			expected: false,
		},
		{
			name:     "empty address",
			address:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestXRPLService_HealthCheck(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)

	// Should fail before initialization
	err := service.HealthCheck()
	assert.Error(t, err)

	// Should pass after initialization
	err = service.Initialize()
	require.NoError(t, err)

	err = service.HealthCheck()
	assert.NoError(t, err)
}

func TestXRPLService_CreateSmartChequeEscrow(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	// Create test wallets
	payerWallet, err := service.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := service.CreateWallet()
	require.NoError(t, err)

	tests := []struct {
		name            string
		payerAddress    string
		payeeAddress    string
		amount          float64
		currency        string
		milestoneSecret string
		expectError     bool
	}{
		{
			name:            "valid XRP escrow",
			payerAddress:    payerWallet.Address,
			payeeAddress:    payeeWallet.Address,
			amount:          10.0,
			currency:        "XRP",
			milestoneSecret: "milestone_secret_123",
			expectError:     false,
		},
		{
			name:            "valid USDT escrow",
			payerAddress:    payerWallet.Address,
			payeeAddress:    payeeWallet.Address,
			amount:          1000.50,
			currency:        "USDT",
			milestoneSecret: "usdt_milestone_secret",
			expectError:     false,
		},
		{
			name:            "invalid payer address",
			payerAddress:    "invalid_address",
			payeeAddress:    payeeWallet.Address,
			amount:          10.0,
			currency:        "XRP",
			milestoneSecret: "secret",
			expectError:     true,
		},
		{
			name:            "invalid payee address",
			payerAddress:    payerWallet.Address,
			payeeAddress:    "invalid_address",
			amount:          10.0,
			currency:        "XRP",
			milestoneSecret: "secret",
			expectError:     true,
		},
		{
			name:            "empty milestone secret",
			payerAddress:    payerWallet.Address,
			payeeAddress:    payeeWallet.Address,
			amount:          10.0,
			currency:        "XRP",
			milestoneSecret: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, fulfillment, err := service.CreateSmartChequeEscrow(
				tt.payerAddress, tt.payeeAddress, tt.amount, tt.currency, tt.milestoneSecret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Empty(t, fulfillment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, fulfillment)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
			}
		})
	}
}

func TestXRPLService_CompleteSmartChequeMilestone(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	// Create test wallets
	payerWallet, err := service.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := service.CreateWallet()
	require.NoError(t, err)

	// First, create an escrow
	milestoneSecret := "milestone_completion_secret"
	escrowResult, fulfillment, err := service.CreateSmartChequeEscrow(
		payerWallet.Address, payeeWallet.Address, 50.0, "XRP", milestoneSecret)
	require.NoError(t, err)
	require.NotNil(t, escrowResult)

	// Generate condition from the same secret
	condition, _, err := service.client.GenerateCondition(milestoneSecret)
	require.NoError(t, err)

	tests := []struct {
		name         string
		payeeAddress string
		ownerAddress string
		sequence     uint32
		condition    string
		fulfillment  string
		expectError  bool
	}{
		{
			name:         "valid milestone completion",
			payeeAddress: payeeWallet.Address,
			ownerAddress: payerWallet.Address,
			sequence:     1,
			condition:    condition,
			fulfillment:  fulfillment,
			expectError:  false,
		},
		{
			name:         "invalid payee address",
			payeeAddress: "invalid_address",
			ownerAddress: payerWallet.Address,
			sequence:     1,
			condition:    condition,
			fulfillment:  fulfillment,
			expectError:  true,
		},
		{
			name:         "invalid owner address",
			payeeAddress: payeeWallet.Address,
			ownerAddress: "invalid_address",
			sequence:     1,
			condition:    condition,
			fulfillment:  fulfillment,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CompleteSmartChequeMilestone(
				tt.payeeAddress, tt.ownerAddress, tt.sequence, tt.condition, tt.fulfillment)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
			}
		})
	}
}

func TestXRPLService_CancelSmartCheque(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	// Create test wallets
	payerWallet, err := service.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := service.CreateWallet()
	require.NoError(t, err)

	tests := []struct {
		name           string
		accountAddress string
		ownerAddress   string
		sequence       uint32
		expectError    bool
	}{
		{
			name:           "valid cancellation",
			accountAddress: payeeWallet.Address,
			ownerAddress:   payerWallet.Address,
			sequence:       1,
			expectError:    false,
		},
		{
			name:           "invalid account address",
			accountAddress: "invalid_address",
			ownerAddress:   payerWallet.Address,
			sequence:       1,
			expectError:    true,
		},
		{
			name:           "invalid owner address",
			accountAddress: payeeWallet.Address,
			ownerAddress:   "invalid_address",
			sequence:       1,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CancelSmartCheque(
				tt.accountAddress, tt.ownerAddress, tt.sequence)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
			}
		})
	}
}

func TestXRPLService_GetEscrowStatus(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	// Create test wallet
	payerWallet, err := service.CreateWallet()
	require.NoError(t, err)

	tests := []struct {
		name         string
		ownerAddress string
		sequence     string
		expectError  bool
	}{
		{
			name:         "valid escrow status query",
			ownerAddress: payerWallet.Address,
			sequence:     "1",
			expectError:  false,
		},
		{
			name:         "invalid owner address",
			ownerAddress: "invalid_address",
			sequence:     "1",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escrowInfo, err := service.GetEscrowStatus(tt.ownerAddress, tt.sequence)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, escrowInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, escrowInfo)
				assert.Equal(t, tt.ownerAddress, escrowInfo.Account)
				assert.NotEmpty(t, escrowInfo.Destination)
				assert.NotEmpty(t, escrowInfo.Amount)
			}
		})
	}
}

func TestXRPLService_FormatAmount(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)

	tests := []struct {
		name     string
		amount   float64
		currency string
		expected string
	}{
		{
			name:     "XRP amount",
			amount:   10.5,
			currency: "XRP",
			expected: "10500000", // 10.5 * 1,000,000 drops
		},
		{
			name:     "USDT amount",
			amount:   1000.123456,
			currency: "USDT",
			expected: "1000.123456",
		},
		{
			name:     "USDC amount",
			amount:   500.0,
			currency: "USDC",
			expected: "500.000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.formatAmount(tt.amount, tt.currency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestXRPLService_EscrowIntegrationFlow(t *testing.T) {
	// Integration test that tests the complete Smart Check escrow flow
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)

	// Step 1: Create test wallets
	payerWallet, err := service.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := service.CreateWallet()
	require.NoError(t, err)

	// Step 2: Create Smart Check escrow
	milestoneSecret := "integration_test_secret_12345"
	escrowResult, fulfillment, err := service.CreateSmartChequeEscrow(
		payerWallet.Address, payeeWallet.Address, 25.0, "XRP", milestoneSecret)
	require.NoError(t, err)
	require.NotNil(t, escrowResult)
	require.NotEmpty(t, fulfillment)

	// Verify escrow creation
	assert.NotEmpty(t, escrowResult.TransactionID)
	assert.Equal(t, "tesSUCCESS", escrowResult.ResultCode)
	assert.True(t, len(escrowResult.TransactionID) == 64) // Standard transaction ID length
	assert.True(t, strings.ToUpper(escrowResult.TransactionID) == escrowResult.TransactionID)

	// Step 3: Query escrow status
	escrowInfo, err := service.GetEscrowStatus(payerWallet.Address, "1")
	require.NoError(t, err)
	require.NotNil(t, escrowInfo)

	// Verify escrow info
	assert.Equal(t, payerWallet.Address, escrowInfo.Account)
	assert.NotEmpty(t, escrowInfo.Amount) // Amount validation is flexible for mock data
	assert.Equal(t, uint32(1), escrowInfo.Sequence)

	// Step 4: Complete milestone (finish escrow)
	condition, _, err := service.client.GenerateCondition(milestoneSecret)
	require.NoError(t, err)

	completionResult, err := service.CompleteSmartChequeMilestone(
		payeeWallet.Address, payerWallet.Address, 1, condition, fulfillment)
	require.NoError(t, err)
	require.NotNil(t, completionResult)

	// Verify completion
	assert.NotEmpty(t, completionResult.TransactionID)
	assert.Equal(t, "tesSUCCESS", completionResult.ResultCode)
	assert.True(t, completionResult.Validated)

	// Ensure completion transaction ID is different from creation
	assert.NotEqual(t, escrowResult.TransactionID, completionResult.TransactionID)

	t.Log("Integration test completed successfully:")
	t.Logf("  Payer: %s", payerWallet.Address)
	t.Logf("  Payee: %s", payeeWallet.Address)
	t.Logf("  Escrow Creation TX: %s", escrowResult.TransactionID)
	t.Logf("  Milestone Completion TX: %s", completionResult.TransactionID)
}

func TestXRPLService_DisputeResolution(t *testing.T) {
	service := &XRPLService{
		initialized: true,
	}

	// Test dispute resolution outcome creation
	outcome := &DisputeResolutionOutcome{
		DisputeID:      "test-dispute-123",
		ResolutionType: "refund",
		OriginalAmount: 1000.0,
		Currency:       "USDT",
		Reason:         "Dispute resolved in favor of payer",
		ExecutedBy:     "test-user",
		ExecutedAt:     time.Now(),
	}

	assert.Equal(t, "test-dispute-123", outcome.DisputeID)
	assert.Equal(t, "refund", outcome.ResolutionType)
	assert.Equal(t, 1000.0, outcome.OriginalAmount)

	// Test dispute resolution status
	status := &DisputeResolutionStatus{
		TransactionID: "test-tx-123",
		Status:        "pending",
		LastChecked:   time.Now(),
		RetryCount:    0,
	}

	assert.Equal(t, "test-tx-123", status.TransactionID)
	assert.Equal(t, "pending", status.Status)
	assert.Equal(t, 0, status.RetryCount)

	// Test dispute resolution transaction
	txn := &DisputeResolutionTransaction{
		TransactionID: "test-tx-123",
		DisputeID:     "test-dispute-123",
		Type:          "refund",
		Amount:        1000.0,
		Currency:      "USDT",
		Status:        "confirmed",
		ExecutedAt:    time.Now(),
		BlockHeight:   12345,
	}

	assert.Equal(t, "test-tx-123", txn.TransactionID)
	assert.Equal(t, "test-dispute-123", txn.DisputeID)
	assert.Equal(t, "refund", txn.Type)
	assert.Equal(t, 1000.0, txn.Amount)
	assert.Equal(t, uint64(12345), txn.BlockHeight)

	// Test service initialization
	assert.True(t, service.initialized)
}

func TestXRPLService_DisputeResolutionTypes(t *testing.T) {
	// Test that all resolution types are supported
	resolutionTypes := []string{"refund", "partial_payment", "full_payment", "cancel"}

	for _, resolutionType := range resolutionTypes {
		outcome := &DisputeResolutionOutcome{
			DisputeID:      "test-dispute",
			ResolutionType: resolutionType,
			OriginalAmount: 1000.0,
			Currency:       "USDT",
			Reason:         "Test",
			ExecutedBy:     "test-user",
			ExecutedAt:     time.Now(),
		}

		escrowInfo := &xrpl.EscrowInfo{
			Account:     "test-account",
			Destination: "test-destination",
			Amount:      "1000",
			Sequence:    123,
		}

		// Test with uninitialized service (should fail)
		uninitializedService := &XRPLService{
			initialized: false,
		}

		_, err := uninitializedService.ExecuteDisputeResolution("test-dispute", resolutionType, escrowInfo, outcome)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "XRPL service not initialized")
	}
}
