package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/services"
)

func TestSmartChequeEscrowIntegration(t *testing.T) {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("No .env file found, using system environment variables: %v", err)
	}

	// This test demonstrates the complete Smart Check escrow workflow
	// using the XRPL testnet simulation

	// Setup XRPL service
	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	xrplService := services.NewXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err, "XRPL service should initialize successfully")

	t.Log("=== Smart Check Escrow Integration Test ===")

	// Step 1: Create enterprise wallets (simulate payer and payee enterprises)
	t.Log("Step 1: Creating enterprise wallets...")

	payerWallet, err := xrplService.CreateWallet()
	require.NoError(t, err, "Should create payer wallet successfully")
	t.Logf("  Payer wallet created: %s", payerWallet.Address)

	payeeWallet, err := xrplService.CreateWallet()
	require.NoError(t, err, "Should create payee wallet successfully")
	t.Logf("  Payee wallet created: %s", payeeWallet.Address)

	// Step 2: Create Smart Check escrow with milestone secret
	t.Log("Step 2: Creating Smart Check escrow...")

	milestoneSecret := "project_milestone_completion_secret_2024"
	escrowAmount := 100.0 // 100 XRP
	currency := "XRP"

	escrowResult, fulfillment, err := xrplService.CreateSmartChequeEscrow(
		payerWallet.Address, payeeWallet.Address, escrowAmount, currency, milestoneSecret)
	require.NoError(t, err, "Should create escrow successfully")
	require.NotNil(t, escrowResult, "Escrow result should not be nil")
	require.NotEmpty(t, fulfillment, "Fulfillment should not be empty")

	t.Logf("  Escrow created with Transaction ID: %s", escrowResult.TransactionID)
	t.Logf("  Escrow amount: %.0f %s", escrowAmount, currency)

	// Validate escrow creation result
	assert.Equal(t, "tesSUCCESS", escrowResult.ResultCode, "Escrow should be created successfully")
	assert.True(t, escrowResult.Validated, "Escrow transaction should be validated")
	assert.NotEmpty(t, escrowResult.TransactionID, "Transaction ID should not be empty")
	assert.Equal(t, 64, len(escrowResult.TransactionID), "Transaction ID should be 64 characters")

	// Step 3: Query escrow status
	t.Log("Step 3: Querying escrow status...")

	escrowInfo, err := xrplService.GetEscrowStatus(payerWallet.Address, "1")
	require.NoError(t, err, "Should query escrow status successfully")
	require.NotNil(t, escrowInfo, "Escrow info should not be nil")

	t.Logf("  Escrow status: Active")
	t.Logf("  Escrow owner: %s", escrowInfo.Account)
	t.Logf("  Escrow destination: %s", escrowInfo.Destination)
	t.Logf("  Escrow amount: %s", escrowInfo.Amount)

	// Validate escrow info
	assert.Equal(t, payerWallet.Address, escrowInfo.Account, "Escrow owner should match payer")
	assert.NotEmpty(t, escrowInfo.Destination, "Escrow destination should not be empty")
	assert.NotEmpty(t, escrowInfo.Amount, "Escrow amount should not be empty")

	// Step 4: Simulate milestone completion and escrow finish
	t.Log("Step 4: Completing milestone and finishing escrow...")

	// Generate the condition from the milestone secret
	condition, _, err := xrplService.GenerateCondition(milestoneSecret)
	require.NoError(t, err, "Should generate condition successfully")

	completionResult, err := xrplService.CompleteSmartChequeMilestone(
		payeeWallet.Address, payerWallet.Address, 1, condition, fulfillment)
	require.NoError(t, err, "Should complete milestone successfully")
	require.NotNil(t, completionResult, "Completion result should not be nil")

	t.Logf("  Milestone completed with Transaction ID: %s", completionResult.TransactionID)

	// Validate milestone completion
	assert.Equal(t, "tesSUCCESS", completionResult.ResultCode, "Milestone completion should be successful")
	assert.True(t, completionResult.Validated, "Completion transaction should be validated")
	assert.NotEmpty(t, completionResult.TransactionID, "Completion transaction ID should not be empty")
	assert.NotEqual(t, escrowResult.TransactionID, completionResult.TransactionID,
		"Completion transaction ID should differ from creation transaction ID")

	t.Log("Step 5: Integration test completed successfully!")
	t.Logf("  Total escrow operations: 2 (Create + Complete)")
	t.Logf("  Escrow lifecycle: Created -> Completed")
	t.Logf("  All transactions validated on XRPL testnet simulation")
}

func TestSmartChequeEscrowCancellation(t *testing.T) {
	// This test demonstrates escrow cancellation scenario

	// Setup XRPL service
	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	xrplService := services.NewXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err)

	t.Log("=== Smart Check Escrow Cancellation Test ===")

	// Create wallets
	payerWallet, err := xrplService.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := xrplService.CreateWallet()
	require.NoError(t, err)

	t.Logf("Created test wallets: Payer=%s, Payee=%s",
		payerWallet.Address, payeeWallet.Address)

	// Create escrow that will be canceled
	milestoneSecret := "canceled_project_secret_2024"
	escrowAmount := 50.0 // 50 XRP

	escrowResult, _, err := xrplService.CreateSmartChequeEscrow(
		payerWallet.Address, payeeWallet.Address, escrowAmount, "XRP", milestoneSecret)
	require.NoError(t, err)

	t.Logf("Created escrow for cancellation: %s", escrowResult.TransactionID)

	// Cancel the escrow (simulate failed milestone or timeout)
	cancelResult, err := xrplService.CancelSmartCheque(
		payeeWallet.Address, payerWallet.Address, 1)
	require.NoError(t, err)
	require.NotNil(t, cancelResult)

	t.Logf("Escrow canceled with Transaction ID: %s", cancelResult.TransactionID)

	// Validate cancellation
	assert.Equal(t, "tesSUCCESS", cancelResult.ResultCode, "Cancellation should be successful")
	assert.True(t, cancelResult.Validated, "Cancellation transaction should be validated")
	assert.NotEqual(t, escrowResult.TransactionID, cancelResult.TransactionID,
		"Cancellation transaction ID should differ from creation transaction ID")

	t.Log("Escrow cancellation test completed successfully!")
}

func TestMultipleCurrencyEscrows(t *testing.T) {
	// This test demonstrates escrows with different currencies

	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	xrplService := services.NewXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err)

	t.Log("=== Multiple Currency Escrow Test ===")

	// Create wallets
	payerWallet, err := xrplService.CreateWallet()
	require.NoError(t, err)

	payeeWallet, err := xrplService.CreateWallet()
	require.NoError(t, err)

	// Test different currency escrows
	currencies := []struct {
		name     string
		amount   float64
		currency string
	}{
		{"XRP Escrow", 25.5, "XRP"},
		{"USDT Escrow", 1000.75, "USDT"},
		{"USDC Escrow", 500.00, "USDC"},
	}

	for i, test := range currencies {
		t.Run(test.name, func(t *testing.T) {
			milestoneSecret := fmt.Sprintf("milestone_secret_%d_%s", i+1, test.currency)

			escrowResult, fulfillment, err := xrplService.CreateSmartChequeEscrow(
				payerWallet.Address, payeeWallet.Address,
				test.amount, test.currency, milestoneSecret)

			require.NoError(t, err, "Should create %s escrow successfully", test.currency)
			require.NotNil(t, escrowResult)
			require.NotEmpty(t, fulfillment)

			t.Logf("  %s escrow created: %.2f %s, TX: %s",
				test.currency, test.amount, test.currency, escrowResult.TransactionID)

			// Validate escrow creation
			assert.Equal(t, "tesSUCCESS", escrowResult.ResultCode)
			assert.True(t, escrowResult.Validated)
			assert.NotEmpty(t, escrowResult.TransactionID)
		})
	}

	t.Log("Multiple currency escrow test completed successfully!")
}

func TestEscrowConditionGeneration(t *testing.T) {
	// This test validates the cryptographic condition generation

	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}

	xrplService := services.NewXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err)

	t.Log("=== Escrow Condition Generation Test ===")

	// Test different milestone secrets
	secrets := []string{
		"milestone_1_delivery_complete",
		"milestone_2_quality_approved",
		"milestone_3_testing_passed",
		"milestone_4_deployment_successful",
	}

	conditions := make(map[string]string)

	for i, secret := range secrets {
		condition, fulfillment, err := xrplService.GenerateCondition(secret)
		require.NoError(t, err, "Should generate condition for secret %d", i+1)

		// Validate condition format
		assert.Equal(t, 64, len(condition), "Condition should be 64 character SHA-256 hash")
		assert.Equal(t, secret, fulfillment, "Fulfillment should match original secret")

		// Ensure conditions are unique
		for prevSecret, prevCondition := range conditions {
			assert.NotEqual(t, prevCondition, condition,
				"Condition for '%s' should differ from '%s'", secret, prevSecret)
		}

		conditions[secret] = condition

		t.Logf("  Secret: %s", secret)
		t.Logf("  Condition: %s", condition)
		t.Logf("  Fulfillment: %s", fulfillment)
	}

	// Test deterministic generation
	t.Log("Testing deterministic condition generation...")

	for secret, expectedCondition := range conditions {
		condition, fulfillment, err := xrplService.GenerateCondition(secret)
		require.NoError(t, err)

		assert.Equal(t, expectedCondition, condition,
			"Same secret should generate same condition")
		assert.Equal(t, secret, fulfillment,
			"Fulfillment should match original secret")
	}

	t.Log("Condition generation test completed successfully!")
}

// Helper function for formatting (needed for fmt.Sprintf)
func init() {
	// This ensures proper imports are available
	_ = time.Now()
}
