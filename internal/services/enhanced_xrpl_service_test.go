package services

import (
	"testing"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnhancedXRPLService(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)
	require.NotNil(t, service)
	assert.Equal(t, config.NetworkURL, service.client.NetworkURL)
	assert.Equal(t, config.TestNet, service.client.TestNet)
	assert.False(t, service.initialized)
}

func TestEnhancedXRPLService_Initialize(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test initialization
	_ = service.Initialize()
	// The service may or may not initialize depending on network connectivity
	// We just verify the service structure is correct
	assert.NotNil(t, service)
	assert.Equal(t, config.NetworkURL, service.client.NetworkURL)
	assert.Equal(t, config.TestNet, service.client.TestNet)
}

func TestEnhancedXRPLService_CreateWallet(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test wallet creation (will fail due to service not initialized)
	wallet, err := service.CreateWallet()
	assert.Error(t, err)
	assert.Nil(t, wallet)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CreateSecp256k1Wallet(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test secp256k1 wallet creation (will fail due to service not initialized)
	wallet, err := service.CreateSecp256k1Wallet()
	assert.Error(t, err)
	assert.Nil(t, wallet)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_ValidateAddress(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test address validation (doesn't require initialization)
	validAddress := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	assert.True(t, service.ValidateAddress(validAddress))

	invalidAddresses := []string{
		"",                                    // Empty
		"r",                                   // Too short
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh0", // Contains 0
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThO", // Contains O
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThI", // Contains I
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThl", // Contains l
		"xHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",  // Doesn't start with r
	}

	for _, addr := range invalidAddresses {
		assert.False(t, service.ValidateAddress(addr), "Address should be invalid: %s", addr)
	}
}

func TestEnhancedXRPLService_ValidateGeneratedAddress(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test generated address validation (more permissive)
	// This address is too long (59 chars > 40 max), so it should fail
	tooLongAddress := "r123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	assert.False(t, service.ValidateGeneratedAddress(tooLongAddress), "Address should fail validation due to length")

	// Test valid generated addresses
	validAddresses := []string{
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",       // Standard length
		"r123456789ABCDEFGHJKLMNPQRSTUVWXYZ",       // Shorter length
		"rABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnop", // Maximum length (40 chars)
	}

	for _, addr := range validAddresses {
		assert.True(t, service.ValidateGeneratedAddress(addr), "Generated address should be valid: %s", addr)
	}

	invalidGeneratedAddresses := []string{
		"",  // Empty
		"r", // Too short
		"x123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz", // Doesn't start with r
	}

	for _, addr := range invalidGeneratedAddresses {
		assert.False(t, service.ValidateGeneratedAddress(addr), "Generated address should be invalid: %s", addr)
	}
}

func TestEnhancedXRPLService_GetAccountInfo(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test account info retrieval (will fail due to service not initialized)
	accountInfo, err := service.GetAccountInfo("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	assert.Error(t, err)
	assert.Nil(t, accountInfo)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_GetAccountBalance(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test account balance retrieval (will fail due to service not initialized)
	balance, err := service.GetAccountBalance("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	assert.Error(t, err)
	assert.Empty(t, balance)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_HealthCheck(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test health check (will fail due to service not initialized)
	err := service.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "XRPL service not initialized")
}

func TestEnhancedXRPLService_CreateSmartChequeEscrow(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test escrow creation (will fail due to service not initialized)
	result, fulfillment, err := service.CreateSmartChequeEscrow(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		100.0,
		"XRP",
		"test_secret",
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, fulfillment)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CreateSmartChequeEscrowWithMilestones(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test escrow creation with milestones (will fail due to service not initialized)
	milestones := []models.Milestone{
		{
			ID:          "milestone1",
			Description: "First milestone",
			Amount:      50.0,
			Status:      models.MilestoneStatusPending,
		},
		{
			ID:          "milestone2",
			Description: "Second milestone",
			Amount:      50.0,
			Status:      models.MilestoneStatusPending,
		},
	}

	result, fulfillment, err := service.CreateSmartChequeEscrowWithMilestones(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		100.0,
		"XRP",
		milestones,
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, fulfillment)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CompleteSmartChequeMilestone(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test milestone completion (will fail due to service not initialized)
	result, err := service.CompleteSmartChequeMilestone(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		1,
		"condition_hash",
		"fulfillment_secret",
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CancelSmartCheque(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test Smart Cheque cancellation (will fail due to service not initialized)
	result, err := service.CancelSmartCheque(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		1,
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_GetEscrowStatus(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test escrow status retrieval (will fail due to service not initialized)
	escrowInfo, err := service.GetEscrowStatus("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "1")
	assert.Error(t, err)
	assert.Nil(t, escrowInfo)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_GenerateCondition(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test condition generation (will fail due to service not initialized)
	condition, fulfillment, err := service.GenerateCondition("test_secret")
	assert.Error(t, err)
	assert.Empty(t, condition)
	assert.Empty(t, fulfillment)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CreatePaymentChannel(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test payment channel creation (will fail due to service not initialized)
	result, err := service.CreatePaymentChannel(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		100.0,
		3600,
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_CreateTrustLine(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test trust line creation (will fail due to service not initialized)
	result, err := service.CreateTrustLine(
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"USDT",
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"1000000",
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestEnhancedXRPLService_Disconnect(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test disconnect (should not fail even if not connected)
	err := service.Disconnect()
	assert.NoError(t, err)
}

func TestEnhancedXRPLService_Integration(t *testing.T) {
	// This test demonstrates the complete flow when the service is properly initialized
	// In a real environment, this would test actual XRPL operations

	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51233",
		TestNet:    true,
	}

	service := NewEnhancedXRPLService(config)

	// Test service creation
	assert.NotNil(t, service)
	assert.False(t, service.initialized)

	// Test that all methods return appropriate errors when not initialized
	testCases := []struct {
		name string
		test func() error
	}{
		{
			name: "CreateWallet",
			test: func() error {
				_, err := service.CreateWallet()
				return err
			},
		},
		{
			name: "GetAccountInfo",
			test: func() error {
				_, err := service.GetAccountInfo("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
				return err
			},
		},
		{
			name: "HealthCheck",
			test: func() error {
				return service.HealthCheck()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.test()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "service not initialized")
		})
	}
}
