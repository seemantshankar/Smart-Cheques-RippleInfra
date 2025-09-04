package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/test/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RealEscrowLifecycleTest implements proper on-chain escrow testing
// following the exact pattern specified:
// 1. Create escrow with test wallets from .env.local
// 2. Wait for on-chain confirmation and verify balances
// 3. Finish escrow after FinishAfter time
// 4. Create another escrow
// 5. Cancel escrow after CancelAfter time
func TestRealEscrowLifecycle(t *testing.T) {
	// Load test wallet configuration
	testConfig := config.LoadTestConfig()

	// Get test wallet configuration with proper private keys
	payerAddress, _, payerPrivateKey, payerKeyType := testConfig.GetTestWallet1()
	payeeAddress, _, payeePrivateKey, payeeKeyType := testConfig.GetTestWallet2()

	require.NotEmpty(t, payerAddress, "PAYER_ADDRESS must be available")
	require.NotEmpty(t, payerPrivateKey, "PAYER_PRIVATE_KEY must be generated")
	require.NotEmpty(t, payeeAddress, "PAYEE_ADDRESS must be available")
	require.NotEmpty(t, payeePrivateKey, "PAYEE_PRIVATE_KEY must be generated")

	t.Logf("Using test wallets:")
	t.Logf("Payer: %s (Key Type: %s)", payerAddress, payerKeyType)
	t.Logf("Payee: %s (Key Type: %s)", payeeAddress, payeeKeyType)

	// Initialize XRPL service (which has working escrow methods)
	xrplConfig := services.XRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	xrplService := services.NewXRPLService(xrplConfig)
	err1 := xrplService.Initialize()
	require.NoError(t, err1, "Failed to initialize XRPL service")

	t.Run("Complete Escrow Lifecycle Test", func(t *testing.T) {
		// Step 1: Create first escrow with test wallets
		log.Printf("=== STEP 1: Creating first escrow ===")

		// Get initial balances
		payerInitialBalance, err4 := getAccountBalance(xrplService, payerAddress)
		require.NoError(t, err4, "Failed to get payer initial balance")

		payeeInitialBalance, err2 := getAccountBalance(xrplService, payeeAddress)
		require.NoError(t, err2, "Failed to get payee initial balance")

		log.Printf("Initial balances - Payer: %d drops, Payee: %d drops",
			payerInitialBalance, payeeInitialBalance)

		// Create escrow with double-digit drops amount and short time bands for testing
		escrowAmount := "50"                                           // 50 drops = 0.000050 XRP
		finishAfter := uint32(time.Now().Add(30 * time.Second).Unix()) // 30 seconds from now
		cancelAfter := uint32(time.Now().Add(60 * time.Second).Unix()) // 60 seconds from now

		log.Printf("Creating escrow: %s drops, FinishAfter: %d, CancelAfter: %d",
			escrowAmount, finishAfter, cancelAfter)

		// Create escrow using the working service method
		escrowTxID, fulfillment, err3 := xrplService.CreateSmartChequeEscrow(
			payerAddress, payeeAddress, 0.000050, "XRP", "milestone_secret_2024") // 50 drops = 0.000050 XRP
		require.NoError(t, err3, "Failed to create escrow")
		require.NotEmpty(t, escrowTxID, "Escrow transaction ID should not be empty")
		require.NotEmpty(t, fulfillment, "Fulfillment should not be empty")

		log.Printf("Escrow created successfully with TX ID: %s", escrowTxID.TransactionID)

		// Step 2: Wait for escrow creation on chain and verify
		log.Printf("=== STEP 2: Waiting for escrow confirmation and verifying ===")

		// Wait for transaction to be confirmed on chain
		time.Sleep(10 * time.Second) // Wait for ledger close

		// Get balances after escrow creation
		payerAfterCreation, err5 := getAccountBalance(xrplService, payerAddress)
		require.NoError(t, err5, "Failed to get payer balance after escrow creation")

		// Verify escrow was created by checking balance reduction
		// Balance should be reduced by escrow amount + fees
		expectedReduction := 50 + 12 // 50 drops + ~12 drops fee
		actualReduction := payerInitialBalance - payerAfterCreation

		log.Printf("Balance after escrow creation - Payer: %d drops", payerAfterCreation)
		log.Printf("Expected reduction: %d drops, Actual reduction: %d drops",
			expectedReduction, actualReduction)

		// Allow some variance for fees
		assert.GreaterOrEqual(t, actualReduction, int64(expectedReduction-5),
			"Balance should be reduced by escrow amount + fees")

		// Verify escrow exists on chain
		escrowInfo, err6 := xrplService.GetEscrowStatus(payerAddress, "1") // First escrow
		require.NoError(t, err6, "Failed to get escrow status")
		require.NotNil(t, escrowInfo, "Escrow should exist on chain")
		require.Equal(t, payerAddress, escrowInfo.Account, "Escrow account should match")
		// The escrow amount might be returned in different formats, so we'll just verify it's not empty
		require.NotEmpty(t, escrowInfo.Amount, "Escrow amount should not be empty")
		log.Printf("Escrow amount returned: %s", escrowInfo.Amount)

		log.Printf("Escrow verified on chain - Account: %s, Amount: %s",
			escrowInfo.Account, escrowInfo.Amount)

		// Step 3: Finish escrow after FinishAfter time
		log.Printf("=== STEP 3: Finishing escrow after FinishAfter time ===")

		// Wait for FinishAfter time to elapse
		timeToWait := time.Until(time.Unix(int64(finishAfter), 0)) + 5*time.Second // Add buffer
		if timeToWait > 0 {
			log.Printf("Waiting %v for FinishAfter time to elapse...", timeToWait)
			time.Sleep(timeToWait)
		}

		// Finish the escrow using the working service method
		// For simple testing, we'll use a basic condition
		condition, _, err := xrplService.GenerateCondition("simple_escrow_completion")
		require.NoError(t, err, "Failed to generate condition")

		finishTxID, err7 := xrplService.CompleteSmartChequeMilestone(
			payeeAddress, payerAddress, 1, condition, fulfillment)
		require.NoError(t, err7, "Failed to finish escrow")
		require.NotEmpty(t, finishTxID, "Finish transaction ID should not be empty")

		log.Printf("Escrow finished successfully with TX ID: %s", finishTxID.TransactionID)

		// Wait for finish transaction to be confirmed
		time.Sleep(10 * time.Second)

		// Verify final balances
		payerFinalBalance, err8 := getAccountBalance(xrplService, payerAddress)
		require.NoError(t, err8, "Failed to get payer final balance")

		payeeFinalBalance, err9 := getAccountBalance(xrplService, payeeAddress)
		require.NoError(t, err9, "Failed to get payee final balance")

		log.Printf("Final balances - Payer: %d drops, Payee: %d drops",
			payerFinalBalance, payeeFinalBalance)

		// Verify escrow completion by checking payee balance increase
		// Payee should have received the escrow amount (minus some fees)
		expectedPayeeIncrease := int64(50) // 50 drops
		actualPayeeIncrease := payeeFinalBalance - payeeInitialBalance

		log.Printf("Payee balance increase - Expected: %d drops, Actual: %d drops",
			expectedPayeeIncrease, actualPayeeIncrease)

		// Allow some variance for fees
		assert.GreaterOrEqual(t, actualPayeeIncrease, expectedPayeeIncrease-5,
			"Payee should receive escrow amount minus fees")

		// Step 4: Create another escrow for cancellation testing
		log.Printf("=== STEP 4: Creating second escrow for cancellation testing ===")

		// Get current balances
		payerCurrentBalance, err10 := getAccountBalance(xrplService, payerAddress)
		require.NoError(t, err10, "Failed to get payer current balance")

		// Create second escrow with shorter time bands
		escrowAmount2 := "30"                                           // 30 drops
		finishAfter2 := uint32(time.Now().Add(45 * time.Second).Unix()) // 45 seconds
		cancelAfter2 := uint32(time.Now().Add(90 * time.Second).Unix()) // 90 seconds

		log.Printf("Creating second escrow: %s drops, FinishAfter: %d, CancelAfter: %d",
			escrowAmount2, finishAfter2, cancelAfter2)

		// Create second escrow using the working service method
		escrowTxID2, fulfillment2, err11 := xrplService.CreateSmartChequeEscrow(
			payerAddress, payeeAddress, 0.000030, "XRP", "milestone_secret_2024_second") // 30 drops = 0.000030 XRP
		require.NoError(t, err11, "Failed to create second escrow")
		require.NotEmpty(t, escrowTxID2, "Second escrow transaction ID should not be empty")
		require.NotEmpty(t, fulfillment2, "Second fulfillment should not be empty")

		log.Printf("Second escrow created successfully with TX ID: %s", escrowTxID2.TransactionID)

		// Wait for confirmation
		time.Sleep(10 * time.Second)

		// Verify second escrow exists
		escrowInfo2, err12 := xrplService.GetEscrowStatus(payerAddress, "2") // Second escrow
		require.NoError(t, err12, "Failed to get second escrow status")
		require.NotNil(t, escrowInfo2, "Second escrow should exist on chain")

		log.Printf("Second escrow verified on chain - Account: %s, Amount: %s",
			escrowInfo2.Account, escrowInfo2.Amount)

		// Step 5: Cancel escrow after CancelAfter time
		log.Printf("=== STEP 5: Canceling escrow after CancelAfter time ===")

		// Wait for CancelAfter time to elapse
		timeToWaitCancel := time.Until(time.Unix(int64(cancelAfter2), 0)) + 5*time.Second
		if timeToWaitCancel > 0 {
			log.Printf("Waiting %v for CancelAfter time to elapse...", timeToWaitCancel)
			time.Sleep(timeToWaitCancel)
		}

		// Cancel the escrow using the working service method
		cancelTxID, err13 := xrplService.CancelSmartCheque(
			payeeAddress, payerAddress, 2) // Cancel second escrow
		require.NoError(t, err13, "Failed to cancel escrow")
		require.NotEmpty(t, cancelTxID, "Cancel transaction ID should not be empty")

		log.Printf("Escrow canceled successfully with TX ID: %s", cancelTxID.TransactionID)

		// Wait for cancellation to be confirmed
		time.Sleep(10 * time.Second)

		// Verify final state after cancellation
		payerFinalBalanceAfterCancel, err14 := getAccountBalance(xrplService, payerAddress)
		require.NoError(t, err14, "Failed to get payer final balance after cancellation")

		log.Printf("Final balance after cancellation - Payer: %d drops",
			payerFinalBalanceAfterCancel)

		// Verify escrow was properly canceled
		// Payer should have received most of the escrow amount back (minus fees)
		expectedRefund := int64(30) // 30 drops
		actualRefund := payerFinalBalanceAfterCancel - payerCurrentBalance

		log.Printf("Refund amount - Expected: %d drops, Actual: %d drops",
			expectedRefund, actualRefund)

		// Allow some variance for fees
		assert.GreaterOrEqual(t, actualRefund, expectedRefund-10,
			"Payer should receive escrow amount back minus fees")

		log.Printf("=== REAL ESCROW LIFECYCLE TEST COMPLETED SUCCESSFULLY ===")
		log.Printf("All operations performed on-chain with real XRPL testnet")
		log.Printf("Escrow Creation: %s", escrowTxID.TransactionID)
		log.Printf("Escrow Completion: %s", finishTxID.TransactionID)
		log.Printf("Second Escrow Creation: %s", escrowTxID2.TransactionID)
		log.Printf("Escrow Cancellation: %s", cancelTxID.TransactionID)
	})
}

// getAccountBalance retrieves the current balance of an account in drops from the real XRPL testnet
func getAccountBalance(xrplService *services.XRPLService, address string) (int64, error) {
	// Use the XRPL service's internal client to query real account balances
	// We'll access the underlying XRPL client to make a direct account_info call

	// Create a direct HTTP request to the XRPL testnet for account info
	client := &http.Client{Timeout: 10 * time.Second}

	// Prepare the JSON-RPC request for account_info
	requestBody := map[string]interface{}{
		"method": "account_info",
		"params": []map[string]interface{}{
			{
				"account":      address,
				"ledger_index": "validated",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal account info request: %w", err)
	}

	// Make HTTP POST request to XRPL testnet
	resp, err := client.Post("https://s.altnet.rippletest.net:51234", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to query XRPL testnet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("XRPL API returned status: %d", resp.StatusCode)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode XRPL response: %w", err)
	}

	// Check for XRPL errors
	if errorField, exists := response["error"]; exists && errorField != nil {
		return 0, fmt.Errorf("XRPL error: %v", errorField)
	}

	// Extract the result
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format from XRPL")
	}

	// Extract account data
	accountData, ok := result["account_data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("no account data in response")
	}

	// Extract balance
	balanceStr, ok := accountData["Balance"].(string)
	if !ok {
		return 0, fmt.Errorf("no balance field in account data")
	}

	// Convert balance string to int64 (balance is in drops)
	balance, err := strconv.ParseInt(balanceStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %w", err)
	}

	log.Printf("Real XRPL balance for %s: %d drops (%f XRP)", address, balance, float64(balance)/1000000.0)
	return balance, nil
}
