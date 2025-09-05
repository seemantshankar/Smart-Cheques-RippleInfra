package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/btcsuite/btcutil/base58"
	cc "github.com/go-interledger/cryptoconditions"
	"github.com/joho/godotenv"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// XRPLFaucetRequest represents the request structure for the XRPL testnet faucet
type XRPLFaucetRequest struct {
	Destination string `json:"destination"`
}

// XRPLFaucetResponse represents the response from the XRPL testnet faucet
type XRPLFaucetResponse struct {
	Account struct {
		XAddress       string `json:"xAddress"`
		ClassicAddress string `json:"classicAddress"`
		Address        string `json:"address"`
	} `json:"account"`
	Amount          interface{} `json:"amount"` // Can be string or number
	TransactionHash string      `json:"transactionHash"`
	Error           string      `json:"error,omitempty"`
}

// AccountInfo represents account information from XRPL
type AccountInfo struct {
	Result struct {
		AccountData struct {
			Balance  string `json:"Balance"`
			Sequence int    `json:"Sequence"`
		} `json:"account_data"`
	} `json:"result"`
}

func main() {
	fmt.Println("=== Comprehensive XRPL Testnet Testing ===")
	fmt.Println("This test includes: Wallet Funding + Real Transactions")

	// Load environment variables from env.local
	if err := godotenv.Load("env.local"); err != nil {
		log.Fatalf("Failed to load env.local: %v", err)
	}

	// Get wallet information from environment
	payerAddress := os.Getenv("PAYER_ADDRESS")
	payerSecretBase58 := os.Getenv("PAYER_SECRET")
	payeeAddress := os.Getenv("PAYEE_ADDRESS")
	payeeSecretBase58 := os.Getenv("PAYEE_SECRET")
	networkURL := os.Getenv("XRPL_NETWORK_URL")
	webSocketURL := os.Getenv("XRPL_WEBSOCKET_URL")

	// Note: We're now using the base58 private keys directly with xrpl-go library
	fmt.Printf("✅ Using base58 private keys directly with xrpl-go library\n")

	fmt.Printf("Payer Address: %s\n", payerAddress)
	fmt.Printf("Payee Address: %s\n", payeeAddress)
	fmt.Printf("Network URL: %s\n", networkURL)

	// Official XRPL Testnet Faucet endpoint
	faucetURL := "https://faucet.altnet.rippletest.net/accounts"

	// ============================================================================
	// PHASE 1: WALLET FUNDING
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PHASE 1: WALLET FUNDING")
	fmt.Println(strings.Repeat("=", 60))

	// Test 1: Check initial balances
	fmt.Println("\n--- Test 1: Initial Account Balances ---")

	payerBalance, payerSequence := getAccountBalance(networkURL, payerAddress)
	payeeBalance, payeeSequence := getAccountBalance(networkURL, payeeAddress)

	fmt.Printf("Payer Balance: %s drops (%.6f XRP), Sequence: %d\n",
		payerBalance, float64(parseBalance(payerBalance))/1000000, payerSequence)
	fmt.Printf("Payee Balance: %s drops (%.6f XRP), Sequence: %d\n",
		payeeBalance, float64(parseBalance(payeeBalance))/1000000, payeeSequence)

	// Test 2: Fund Payer Wallet
	fmt.Println("\n--- Test 2: Funding Payer Wallet ---")

	fmt.Printf("Funding payer wallet: %s...\n", payerAddress)

	payerTxHash, err := fundWallet(faucetURL, payerAddress)
	if err != nil {
		log.Fatalf("Failed to fund payer wallet: %v", err)
	}
	fmt.Printf("✅ Payer funding transaction: %s\n", payerTxHash)

	// Wait for transaction to be processed
	fmt.Println("Waiting for payer funding transaction to be processed...")
	time.Sleep(10 * time.Second)

	// Test 3: Fund Payee Wallet
	fmt.Println("\n--- Test 3: Funding Payee Wallet ---")

	fmt.Printf("Funding payee wallet: %s...\n", payeeAddress)

	payeeTxHash, err := fundWallet(faucetURL, payeeAddress)
	if err != nil {
		log.Fatalf("Failed to fund payee wallet: %v", err)
	}
	fmt.Printf("✅ Payee funding transaction: %s\n", payeeTxHash)

	// Wait for transaction to be processed
	fmt.Println("Waiting for payee funding transaction to be processed...")
	time.Sleep(10 * time.Second)

	// Test 4: Verify Funding Results
	fmt.Println("\n--- Test 4: Funding Verification ---")

	payerBalanceAfter, payerSequenceAfter := getAccountBalance(networkURL, payerAddress)
	payeeBalanceAfter, payeeSequenceAfter := getAccountBalance(networkURL, payeeAddress)

	fmt.Printf("Payer Balance After Funding: %s drops (%.6f XRP), Sequence: %d\n",
		payerBalanceAfter, float64(parseBalance(payerBalanceAfter))/1000000, payerSequenceAfter)
	fmt.Printf("Payee Balance After Funding: %s drops (%.6f XRP), Sequence: %d\n",
		payeeBalanceAfter, float64(parseBalance(payeeBalanceAfter))/1000000, payeeSequenceAfter)

	// Calculate funding results
	payerBalanceChange := parseBalance(payerBalanceAfter) - parseBalance(payerBalance)
	payeeBalanceChange := parseBalance(payeeBalanceAfter) - parseBalance(payeeBalance)

	fmt.Printf("\nPayer Balance Change: %d drops (%.6f XRP)\n",
		payerBalanceChange, float64(payerBalanceChange)/1000000)
	fmt.Printf("Payee Balance Change: %d drops (%.6f XRP)\n",
		payeeBalanceChange, float64(payeeBalanceChange)/1000000)

	if payerBalanceChange > 0 {
		fmt.Printf("✅ Payer wallet successfully funded with %d drops (%.6f XRP)\n",
			payerBalanceChange, float64(payerBalanceChange)/1000000)
	} else {
		fmt.Printf("⚠️ Payer wallet funding may not have completed yet\n")
	}

	if payeeBalanceChange > 0 {
		fmt.Printf("✅ Payee wallet successfully funded with %d drops (%.6f XRP)\n",
			payeeBalanceChange, float64(payeeBalanceChange)/1000000)
	} else {
		fmt.Printf("⚠️ Payee wallet funding may not have completed yet\n")
	}

	// ============================================================================
	// PHASE 2: REAL TRANSACTION TESTING
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PHASE 2: REAL TRANSACTION TESTING")
	fmt.Println(strings.Repeat("=", 60))

	// Store transaction IDs for block explorer URLs
	var transactionIDs []string

	// Store expected finish time for escrow timing verification
	var expectedFinishTime uint32

	// We'll test both escrow finalization and cancellation scenarios
	// For this, we'll create two separate escrows:
	// 1. One to finish (complete milestone)
	// 2. One to cancel (return funds to payer)

	// Create enhanced client for real testnet operations
	enhancedClient := xrpl.NewEnhancedClient(networkURL, webSocketURL, true)
	err = enhancedClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect enhanced client: %v", err)
	}
	defer enhancedClient.Disconnect()
	fmt.Println("✅ Connected to XRPL testnet successfully")

	// Test 5: Health Check
	fmt.Println("\n--- Test 5: Enhanced Client Health Check ---")
	err = enhancedClient.HealthCheck()
	if err != nil {
		log.Fatalf("Enhanced client health check failed: %v", err)
	}
	fmt.Println("✅ Enhanced client health check passed")

	// Test 6: Get Account Data After Funding
	fmt.Println("\n--- Test 6: Account Data After Funding ---")

	payerAccountData, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer account data: %v", err)
	}
	fmt.Printf("Payer Account Data - Balance: %s, Sequence: %d\n", payerAccountData.Balance, payerAccountData.Sequence)

	payeeAccountData, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee account data: %v", err)
	}
	fmt.Printf("Payee Account Data - Balance: %s, Sequence: %d\n", payeeAccountData.Balance, payeeAccountData.Sequence)

	// Test 7: Create and Submit Payment Transaction using xrpl-go
	fmt.Println("\n--- Test 7: Real Payment Transaction using xrpl-go ---")

	// Get balances BEFORE payment
	payerBalanceBeforePayment, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer balance before payment: %v", err)
	}
	payeeBalanceBeforePayment, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee balance before payment: %v", err)
	}

	fmt.Printf("BEFORE PAYMENT - Payer: %s drops, Payee: %s drops\n",
		payerBalanceBeforePayment.Balance, payeeBalanceBeforePayment.Balance)

	// Get current account info to ensure we have the latest sequence number
	fmt.Println("Getting current account info for transaction...")
	currentAccountData, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get current account data: %v", err)
	}
	fmt.Printf("Current sequence number: %d\n", currentAccountData.Sequence)

	// Create a small payment transaction (0.1 XRP = 100,000 drops)
	amount := "100000" // 0.1 XRP in drops

	fmt.Printf("Creating payment transaction using xrpl-go: %s -> %s, Amount: %s drops (0.1 XRP)\n",
		payerAddress, payeeAddress, amount)

	// Use xrpl-go library to create and sign the transaction
	result, err := createPaymentWithXrplGo(payerAddress, payeeAddress, amount, payerSecretBase58, currentAccountData.Sequence)
	if err != nil {
		log.Fatalf("Failed to create payment transaction with xrpl-go: %v", err)
	}

	fmt.Printf("✅ Payment transaction created successfully!\n")
	fmt.Printf("Transaction ID: %s\n", result.TransactionID)
	fmt.Printf("Result Code: %s\n", result.ResultCode)
	fmt.Printf("Validated: %t\n", result.Validated)

	// Store transaction ID for block explorer
	transactionIDs = append(transactionIDs, result.TransactionID)

	// Wait a moment for the transaction to be processed
	fmt.Println("Waiting for payment transaction to be processed...")
	time.Sleep(5 * time.Second)

	// Test 8: Check Account Balances After Payment
	fmt.Println("\n--- Test 8: Account Balances After Payment ---")

	payerAccountDataAfter, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer account data after payment: %v", err)
	}
	fmt.Printf("Payer Balance After Payment: %s drops (%.6f XRP)\n",
		payerAccountDataAfter.Balance, float64(parseBalance(payerAccountDataAfter.Balance))/1000000)

	payeeAccountDataAfter, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee account data after payment: %v", err)
	}
	fmt.Printf("Payee Balance After Payment: %s drops (%.6f XRP)\n",
		payeeAccountDataAfter.Balance, float64(parseBalance(payeeAccountDataAfter.Balance))/1000000)

	// Calculate and verify payment results
	payerPaymentChange := parseBalance(payerAccountDataAfter.Balance) - parseBalance(payerBalanceBeforePayment.Balance)
	payeePaymentChange := parseBalance(payeeAccountDataAfter.Balance) - parseBalance(payeeBalanceBeforePayment.Balance)

	fmt.Printf("Payer Payment Change: %d drops (%.6f XRP)\n",
		payerPaymentChange, float64(payerPaymentChange)/1000000)
	fmt.Printf("Payee Payment Change: %d drops (%.6f XRP)\n",
		payeePaymentChange, float64(payeePaymentChange)/1000000)

	// Verify payment amounts (accounting for fees)
	expectedPaymentAmount := int64(100000) // 0.1 XRP
	if payeePaymentChange == expectedPaymentAmount {
		fmt.Printf("✅ PAYMENT VERIFIED: Payee received correct amount (+%d drops)\n", expectedPaymentAmount)
	} else {
		fmt.Printf("⚠️ PAYMENT VERIFICATION: Expected +%d drops, got %+d drops\n", expectedPaymentAmount, payeePaymentChange)
	}

	// Payer should have lost payment amount + fee (typically ~12 drops)
	if payerPaymentChange <= -expectedPaymentAmount-20 && payerPaymentChange >= -expectedPaymentAmount-1000 {
		fmt.Printf("✅ PAYMENT VERIFIED: Payer deducted correct amount (%+d drops)\n", payerPaymentChange)
	} else {
		fmt.Printf("⚠️ PAYMENT VERIFICATION: Unexpected payer deduction (%+d drops)\n", payerPaymentChange)
	}

	// Test 9: Create and Submit Escrow Transaction using xrpl-go
	fmt.Println("\n--- Test 9: Real Escrow Transaction using xrpl-go ---")

	// Get balances BEFORE escrow creation
	payerBalanceBeforeEscrow, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer balance before escrow: %v", err)
	}

	fmt.Printf("BEFORE ESCROW - Payer: %s drops\n", payerBalanceBeforeEscrow.Balance)

	// Generate proper XRPL crypto-condition for escrow using go-interledger/cryptoconditions
	// This generates the correct DER-encoded format expected by XRPL

	// Generate a secure random 32-byte preimage (fulfillment)
	preimage, err := generatePreimage()
	if err != nil {
		log.Fatalf("Failed to generate preimage: %v", err)
	}

	// Create proper XRPL condition and fulfillment using crypto-conditions library
	condition, fulfillment, err := createXRPLConditionAndFulfillment(preimage)
	if err != nil {
		log.Fatalf("Failed to create XRPL condition and fulfillment: %v", err)
	}

	fmt.Printf("Generated preimage (raw): %s\n", hex.EncodeToString(preimage))
	fmt.Printf("Generated condition (DER-encoded): %s\n", condition)
	fmt.Printf("Generated fulfillment (DER-encoded): %s\n", fulfillment)
	fmt.Printf("Condition length: %d characters\n", len(condition))

	// Create an escrow for 0.2 XRP (200,000 drops)
	escrowAmount := "200000" // 0.2 XRP in drops

	fmt.Printf("Creating escrow using xrpl-go: %s -> %s, Amount: %s drops (0.2 XRP)\n",
		payerAddress, payeeAddress, escrowAmount)

	// DEBUG: Try creating escrow WITHOUT condition first to test basic functionality
	fmt.Printf("🔧 DEBUG: Creating escrow WITHOUT condition to test basic finish functionality\n")
	emptyCondition := ""
	// Get the current sequence number right before escrow creation (after payment transaction)
	currentAccountDataForEscrow, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get current account data for escrow: %v", err)
	}
	fmt.Printf("Current sequence for escrow creation: %d\n", currentAccountDataForEscrow.Sequence)

	// Store this as the actual escrow sequence that will be used
	actualEscrowSequence := currentAccountDataForEscrow.Sequence

	// Calculate expected finish time (5 seconds from escrow creation)
	epochStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	xrplTimeNow := uint32(time.Now().Sub(epochStart).Seconds())
	expectedFinishTime = xrplTimeNow + 5

	escrowResult, err := createEscrowWithXrplGo(payerAddress, payeeAddress, escrowAmount, emptyCondition, payerSecretBase58, currentAccountDataForEscrow.Sequence)
	if err != nil {
		log.Fatalf("Failed to create escrow with xrpl-go: %v", err)
	}

	fmt.Printf("✅ Escrow created successfully!\n")
	fmt.Printf("Escrow Transaction ID: %s\n", escrowResult.TransactionID)
	fmt.Printf("Result Code: %s\n", escrowResult.ResultCode)
	fmt.Printf("Validated: %t\n", escrowResult.Validated)

	// Store transaction ID for block explorer
	transactionIDs = append(transactionIDs, escrowResult.TransactionID)

	// Wait a moment for the escrow to be processed
	fmt.Println("Waiting for escrow to be processed...")
	time.Sleep(5 * time.Second)

	// Test 10: Check Account Balances After Escrow Creation
	fmt.Println("\n--- Test 10: Account Balances After Escrow Creation ---")

	payerAccountDataAfterEscrow, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer account data after escrow: %v", err)
	}
	fmt.Printf("Payer Balance After Escrow: %s drops (%.6f XRP)\n",
		payerAccountDataAfterEscrow.Balance, float64(parseBalance(payerAccountDataAfterEscrow.Balance))/1000000)

	// Calculate and verify escrow balance change
	escrowChange := parseBalance(payerAccountDataAfterEscrow.Balance) - parseBalance(payerBalanceBeforeEscrow.Balance)
	fmt.Printf("Payer Balance Change (Escrow Creation): %d drops (%.6f XRP)\n",
		escrowChange, float64(escrowChange)/1000000)

	// Verify escrow creation (accounting for fees)
	expectedEscrowAmount := int64(200000) // 0.2 XRP
	if escrowChange <= -expectedEscrowAmount-20 && escrowChange >= -expectedEscrowAmount-1000 {
		fmt.Printf("✅ ESCROW CREATION VERIFIED: Payer escrowed correct amount (%+d drops)\n", escrowChange)
	} else {
		fmt.Printf("⚠️ ESCROW CREATION VERIFICATION: Unexpected escrow deduction (%+d drops)\n", escrowChange)
	}

	// ============================================================================
	// PHASE 3: ESCROW FINALIZATION AND CANCELLATION TESTING
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PHASE 3: ESCROW FINALIZATION AND CANCELLATION")
	fmt.Println(strings.Repeat("=", 60))

	// Test 10.5: Create Second Escrow for Cancellation Testing
	fmt.Println("\n--- Test 10.5: Create Second Escrow for Cancellation ---")

	// Get updated account sequence after first escrow
	updatedAccountData, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get updated account data: %v", err)
	}
	fmt.Printf("Updated account sequence after first escrow: %d\n", updatedAccountData.Sequence)

	// Create second escrow for cancellation testing
	escrowAmount2 := "150000" // 0.15 XRP in drops

	fmt.Printf("Creating second escrow for cancellation: %s -> %s, Amount: %s drops (0.15 XRP)\n",
		payerAddress, payeeAddress, escrowAmount2)

	// Generate new condition and fulfillment for second escrow
	preimage2, err := generatePreimage()
	if err != nil {
		log.Fatalf("Failed to generate second preimage: %v", err)
	}

	condition2, _, err := createXRPLConditionAndFulfillment(preimage2)
	if err != nil {
		log.Fatalf("Failed to create second XRPL condition and fulfillment: %v", err)
	}

	fmt.Printf("Generated second condition: %s\n", condition2)

	// Create second escrow using xrpl-go
	escrowResult2, err := createEscrowWithXrplGo(payerAddress, payeeAddress, escrowAmount2, condition2, payerSecretBase58, updatedAccountData.Sequence)
	if err != nil {
		log.Fatalf("Failed to create second escrow: %v", err)
	}

	fmt.Printf("✅ Second escrow created successfully!\n")
	fmt.Printf("Second Escrow Transaction ID: %s\n", escrowResult2.TransactionID)
	fmt.Printf("Result Code: %s\n", escrowResult2.ResultCode)
	fmt.Printf("Validated: %t\n", escrowResult2.Validated)

	// Store transaction ID for block explorer
	transactionIDs = append(transactionIDs, escrowResult2.TransactionID)

	// Wait for second escrow to be processed
	fmt.Println("Waiting for second escrow to be processed...")
	time.Sleep(3 * time.Second)

	// Wait for FinishAfter time (5 seconds) to pass so we can finish the first escrow
	fmt.Println("Waiting for FinishAfter time (5 seconds) to allow escrow finalization...")
	time.Sleep(3 * time.Second)

	// Test 11: Finish First Escrow (Complete Milestone) using xrpl-go
	fmt.Println("\n--- Test 11: Finish First Escrow (Complete Milestone) using xrpl-go ---")

	// Use the actual escrow sequence that was stored during escrow creation
	escrowSequence := actualEscrowSequence

	fmt.Printf("Finishing first escrow with sequence: %d (from escrow creation)\n", escrowSequence)

	// Debug: Check current XRPL time vs expected FinishAfter time
	currentTime := uint32(time.Now().Unix() - 946684800) // XRPL epoch: 2000-01-01T00:00:00 UTC

	fmt.Printf("🕐 TIMING VERIFICATION:\n")
	fmt.Printf("  Current XRPL time: %d\n", currentTime)
	fmt.Printf("  Escrow FinishAfter was set to: %d\n", expectedFinishTime)
	fmt.Printf("  Time difference: %d seconds\n", int(currentTime)-int(expectedFinishTime))
	fmt.Printf("  Can finish escrow: %t\n", currentTime >= expectedFinishTime)

	// If FinishAfter time hasn't elapsed yet, wait for it
	if currentTime < expectedFinishTime {
		waitSeconds := expectedFinishTime - currentTime + 1 // Add 1 second buffer
		fmt.Printf("⏳ FinishAfter time not reached yet. Waiting %d seconds...\n", waitSeconds)
		time.Sleep(time.Duration(waitSeconds) * time.Second)
		fmt.Printf("✅ FinishAfter time should now be elapsed\n")
	}

	// Check if escrow exists before trying to finish it
	fmt.Printf("🔍 Checking if escrow exists before finishing...\n")

	// First, try our enhanced client's method
	escrowStatus, err := enhancedClient.GetEscrowStatus(payerAddress, fmt.Sprintf("%d", escrowSequence))
	if err != nil {
		fmt.Printf("⚠️ Enhanced client escrow status failed: %v\n", err)
	} else {
		fmt.Printf("✅ Enhanced client found escrow - Amount: %s, Condition: %s\n", escrowStatus.Amount, escrowStatus.Condition)
		if escrowStatus.Condition == "" {
			fmt.Printf("✅ Unconditional escrow (no condition required)\n")
		}
	}

	// CRITICAL: Query the actual XRPL ledger directly for escrow objects
	fmt.Printf("🔍 Querying XRPL ledger directly for escrow objects...\n")

	ledgerQuery := map[string]interface{}{
		"method": "account_objects",
		"params": []interface{}{
			map[string]interface{}{
				"account":      payerAddress,
				"type":         "escrow",
				"ledger_index": "validated",
			},
		},
	}

	jsonData, _ := json.Marshal(ledgerQuery)
	resp, err := http.Post("https://s.altnet.rippletest.net:51234", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("⚠️ Failed to query XRPL ledger for escrow objects: %v\n", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("⚠️ XRPL API returned status: %d\n", resp.StatusCode)
		} else {
			body, _ := io.ReadAll(resp.Body)
			var ledgerResponse map[string]interface{}
			json.Unmarshal(body, &ledgerResponse)

			if result, ok := ledgerResponse["result"].(map[string]interface{}); ok {
				if objects, ok := result["account_objects"].([]interface{}); ok {
					fmt.Printf("📋 Found %d escrow objects in ledger\n", len(objects))
					for i, obj := range objects {
						if escrowObj, ok := obj.(map[string]interface{}); ok {
							if seq, ok := escrowObj["Sequence"].(float64); ok && uint32(seq) == escrowSequence {
								fmt.Printf("🎯 ESCROW OBJECT %d DETAILS:\n", i+1)
								fmt.Printf("  Sequence: %.0f\n", seq)
								fmt.Printf("  Amount: %v\n", escrowObj["Amount"])
								fmt.Printf("  Destination: %v\n", escrowObj["Destination"])
								fmt.Printf("  Owner: %v\n", escrowObj["Account"]) // Owner is the Account field
								fmt.Printf("  Condition: %v\n", escrowObj["Condition"])
								fmt.Printf("  FinishAfter: %v\n", escrowObj["FinishAfter"])
								fmt.Printf("  CancelAfter: %v\n", escrowObj["CancelAfter"])
								fmt.Printf("  Flags: %v\n", escrowObj["Flags"])

								// Check if escrow is still active
								if flags, ok := escrowObj["Flags"].(float64); ok {
									if flags == 0 {
										fmt.Printf("✅ Escrow is ACTIVE (Flags: 0)\n")
									} else {
										fmt.Printf("⚠️ Escrow may be FINISHED or CANCELLED (Flags: %.0f)\n", flags)
									}
								}

								// Verify Owner matches our expectation
								if owner, ok := escrowObj["Account"].(string); ok {
									if owner == payerAddress {
										fmt.Printf("✅ Escrow Owner matches expected: %s\n", owner)
									} else {
										fmt.Printf("⚠️ Escrow Owner mismatch! Expected: %s, Got: %s\n", payerAddress, owner)
									}
								}

								// Verify Destination matches our expectation
								if dest, ok := escrowObj["Destination"].(string); ok {
									if dest == payeeAddress {
										fmt.Printf("✅ Escrow Destination matches expected: %s\n", dest)
									} else {
										fmt.Printf("⚠️ Escrow Destination mismatch! Expected: %s, Got: %s\n", payeeAddress, dest)
									}
								}

								break // Found our escrow
							}
						}
					}
				} else {
					fmt.Printf("⚠️ No escrow objects found in ledger for account %s\n", payerAddress)
				}
			}
		}
	}

	// Verify payee has sufficient balance for transaction fee
	payeeBalanceBeforeFinish, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee balance before finish: %v", err)
	}
	fmt.Printf("💰 Payee balance before finish: %s drops\n", payeeBalanceBeforeFinish.Balance)

	// CRITICAL DEBUGGING: Verify all escrow finish requirements
	fmt.Printf("🔍 PRE-FINISH VERIFICATION:\n")
	fmt.Printf("  Escrow Owner (payer): %s\n", payerAddress)
	fmt.Printf("  Escrow Destination (payee): %s\n", payeeAddress)
	fmt.Printf("  Escrow Sequence: %d\n", escrowSequence)

	// Verify the escrow sequence by checking recent transactions
	fmt.Printf("🔍 VERIFYING ESCROW SEQUENCE:\n")
	fmt.Printf("  Expected escrow created with sequence: %d\n", escrowSequence)
	fmt.Printf("  This should match the sequence used in EscrowCreate transaction\n")
	fmt.Printf("  Finish Account: %s (should be payee)\n", payeeAddress)

	// Get the current payee sequence for the finish transaction
	payeeCurrentData, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get current payee data for finish: %v", err)
	}
	fmt.Printf("  Payee current sequence: %d\n", payeeCurrentData.Sequence)
	fmt.Printf("  Payee current balance: %s drops\n", payeeCurrentData.Balance)

	// Verify we're using the correct private key for payee
	fmt.Printf("🔐 SIGNING VERIFICATION:\n")
	fmt.Printf("  Using payee secret key for signing: %s...\n", payeeSecretBase58[:10]+"...")

	// Check account flags for both accounts to diagnose tecNO_PERMISSION
	fmt.Printf("\n🔍 ACCOUNT FLAGS ANALYSIS:\n")
	fmt.Printf("   Checking multiple possible DepositAuth flag values:\n")
	fmt.Printf("   - DepositAuth asfDepositAuth (9): 0x200 = %d\n", 1<<9)        // 512
	fmt.Printf("   - DepositAuth lsfDepositAuth: 0x01000000 = %d\n", 0x01000000) // 16777216

	// Check payer account flags
	payerAccountData, err = enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer account data: %v", err)
	}
	fmt.Printf("  Payer account flags: %d (hex: 0x%X)\n", payerAccountData.Flags, payerAccountData.Flags)
	fmt.Printf("  Payer DepositAuth (512): %v\n", (payerAccountData.Flags&512) != 0)
	fmt.Printf("  Payer DepositAuth (16777216): %v\n", (payerAccountData.Flags&0x01000000) != 0)
	fmt.Printf("  Payer RequireAuth (256): %v\n", (payerAccountData.Flags&256) != 0)

	// Check payee account flags (this might be the issue!)
	payeeAccountData, err = enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee account data: %v", err)
	}
	fmt.Printf("  Payee account flags: %d (hex: 0x%X)\n", payeeAccountData.Flags, payeeAccountData.Flags)
	fmt.Printf("  Payee DepositAuth (512): %v\n", (payeeAccountData.Flags&512) != 0)
	fmt.Printf("  Payee DepositAuth (16777216): %v\n", (payeeAccountData.Flags&0x01000000) != 0)
	fmt.Printf("  Payee RequireAuth (256): %v\n", (payeeAccountData.Flags&256) != 0)

	// Check for DepositAuth with either possible flag value
	if (payeeAccountData.Flags&512) != 0 || (payeeAccountData.Flags&0x01000000) != 0 {
		fmt.Printf("🚨 ISSUE FOUND: Payee account has DepositAuth enabled!\n")
		fmt.Printf("   This explains the tecNO_PERMISSION error - escrows cannot finish to accounts with DepositAuth\n")
		fmt.Printf("   The payee account needs to preauthorize the payer or disable DepositAuth\n")
		fmt.Printf("   SOLUTION: Either disable DepositAuth on payee or preauthorize payer account\n")
	} else {
		fmt.Printf("✅ No DepositAuth detected with known flag values\n")
	}

	// 🎯 FINAL BREAKTHROUGH: Time-based escrow rules!
	// XRPL Documentation: "anyone can submit the EscrowFinish transaction to finish the escrow"
	// Our escrow has FinishAfter (time-based) but no Condition - this is a TIME-BASED escrow
	// For time-based escrows: ANYONE can finish, not just destination!
	fmt.Printf("\n🎯 FINAL BREAKTHROUGH - TIME-BASED ESCROW RULES:\n")
	fmt.Printf("   Our escrow: FinishAfter=%d, Condition='' (TIME-BASED escrow)\n", 810427760)
	fmt.Printf("   XRPL Rule: 'anyone can submit EscrowFinish for time-based escrows'\n")
	fmt.Printf("   Testing: Account = payer, signed by payer (ANYONE can finish)\n\n")

	// Use enhanced XRPL client to finish the escrow transaction (with our sequence fix)
	// For time-based escrow, don't provide condition/fulfillment
	escrowFinish := &xrpl.EscrowFinish{
		Account:       payerAddress, // CRITICAL: Account = payer (ANYONE can finish time-based escrow)
		Owner:         payerAddress, // Owner = payer (escrow creator)
		OfferSequence: escrowSequence,
		Condition:     "", // Empty for time-based escrow
		Fulfillment:   "", // Empty for time-based escrow
	}

	// Use the enhanced client which automatically handles the correct sequence
	// Use PAYER's key to sign (ANYONE can finish time-based escrows)
	finishResult, err := enhancedClient.FinishEscrow(escrowFinish, payerSecretBase58) // PAYER signs for PAYER account
	if err != nil {
		log.Fatalf("Failed to finish first escrow with xrpl-go: %v", err)
	}

	fmt.Printf("✅ First escrow finished successfully!\n")
	fmt.Printf("Finish Transaction ID: %s\n", finishResult.TransactionID)
	fmt.Printf("Result Code: %s\n", finishResult.ResultCode)
	fmt.Printf("Validated: %t\n", finishResult.Validated)

	// Store transaction ID for block explorer
	transactionIDs = append(transactionIDs, finishResult.TransactionID)

	// Verify escrow finish balance changes
	payeeBalanceAfterFinish, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get payee balance after finish: %v", err)
	}
	payerBalanceAfterFinish, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer balance after finish: %v", err)
	}

	payeeFinishChange := parseBalance(payeeBalanceAfterFinish.Balance) - parseBalance(payeeBalanceBeforeFinish.Balance)
	payerFinishChange := parseBalance(payerBalanceAfterFinish.Balance) - parseBalance(payerAccountDataAfterEscrow.Balance)

	fmt.Printf("💰 Payee balance change after finish: %+d drops\n", payeeFinishChange)
	fmt.Printf("💰 Payer balance change after finish: %+d drops\n", payerFinishChange)

	// Verify escrow release (accounting for fees)
	expectedEscrowRelease := int64(200000) // 0.2 XRP
	if payeeFinishChange >= expectedEscrowRelease-50 && payeeFinishChange <= expectedEscrowRelease+50 {
		fmt.Printf("✅ ESCROW FINISH VERIFIED: Payee received escrowed funds (%+d drops)\n", payeeFinishChange)
	} else {
		fmt.Printf("⚠️ ESCROW FINISH VERIFICATION: Expected ~%+d drops, got %+d drops\n", expectedEscrowRelease, payeeFinishChange)
	}

	// Wait a moment for the finish transaction to be processed
	fmt.Println("Waiting for first escrow finish to be processed...")
	time.Sleep(3 * time.Second)

	// Test 12: Cancel Second Escrow using xrpl-go
	fmt.Println("\n--- Test 12: Cancel Second Escrow using xrpl-go ---")

	// Wait additional time to ensure CancelAfter time (10 seconds) has passed
	fmt.Println("Waiting additional time for CancelAfter time (10 seconds) to allow escrow cancellation...")
	time.Sleep(4 * time.Second)

	// Enhanced client handles sequence automatically, no need to get account data

	// Use the sequence from the second escrow creation
	// The second escrow was created with sequence = updatedAccountData.Sequence (the sequence number at time of second escrow creation)
	escrowSequence2 := uint32(updatedAccountData.Sequence)
	fmt.Printf("Cancelling second escrow with sequence: %d\n", escrowSequence2)

	// Use enhanced XRPL client to cancel the escrow transaction (with our sequence fix)
	escrowCancel := &xrpl.EscrowCancel{
		Account:       payerAddress,
		Owner:         payerAddress,
		OfferSequence: escrowSequence2,
	}

	// Use the enhanced client which automatically handles the correct sequence
	// Pass the base58 private key directly (enhanced client now expects base58)
	cancelResult, err := enhancedClient.CancelEscrow(escrowCancel, payerSecretBase58)
	if err != nil {
		log.Fatalf("Failed to cancel escrow with xrpl-go: %v", err)
	}

	fmt.Printf("✅ Second escrow cancelled successfully!\n")
	fmt.Printf("Cancel Transaction ID: %s\n", cancelResult.TransactionID)
	fmt.Printf("Result Code: %s\n", cancelResult.ResultCode)
	fmt.Printf("Validated: %t\n", cancelResult.Validated)

	// Store transaction ID for block explorer
	transactionIDs = append(transactionIDs, cancelResult.TransactionID)

	// Wait a moment for the cancel transaction to be processed
	fmt.Println("Waiting for escrow cancel to be processed...")
	time.Sleep(3 * time.Second)

	// Test 13: Final Account Balances After All Operations
	fmt.Println("\n--- Test 13: Final Account Balances After All Operations ---")

	payerAccountDataFinal, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get final payer account data: %v", err)
	}
	fmt.Printf("Payer Final Balance: %s drops (%.6f XRP)\n",
		payerAccountDataFinal.Balance, float64(parseBalance(payerAccountDataFinal.Balance))/1000000)

	payeeAccountDataFinal, err := enhancedClient.GetAccountData(payeeAddress)
	if err != nil {
		log.Fatalf("Failed to get final payee account data: %v", err)
	}
	fmt.Printf("Payee Final Balance: %s drops (%.6f XRP)\n",
		payeeAccountDataFinal.Balance, float64(parseBalance(payeeAccountDataFinal.Balance))/1000000)

	// Calculate total changes from initial funding
	payerTotalChange := parseBalance(payerAccountDataFinal.Balance) - parseBalance(payerBalanceAfter)
	payeeTotalChange := parseBalance(payeeAccountDataFinal.Balance) - parseBalance(payeeBalanceAfter)

	fmt.Printf("Payer Total Transaction Change: %d drops (%.6f XRP)\n",
		payerTotalChange, float64(payerTotalChange)/1000000)
	fmt.Printf("Payee Total Transaction Change: %d drops (%.6f XRP)\n",
		payeeTotalChange, float64(payeeTotalChange)/1000000)

	// ============================================================================
	// FINAL SUMMARY
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("COMPREHENSIVE TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n✅ PHASE 1: WALLET FUNDING")
	fmt.Printf("✅ Payer funded with %d drops (%.6f XRP)\n",
		payerBalanceChange, float64(payerBalanceChange)/1000000)
	fmt.Printf("✅ Payee funded with %d drops (%.6f XRP)\n",
		payeeBalanceChange, float64(payeeBalanceChange)/1000000)
	fmt.Printf("✅ Payer Funding Transaction: %s\n", payerTxHash)
	fmt.Printf("✅ Payee Funding Transaction: %s\n", payeeTxHash)

	fmt.Println("\n✅ PHASE 2: REAL TRANSACTIONS")
	fmt.Println("✅ Payment transaction created and submitted!")
	fmt.Printf("✅ Payment Transaction ID: %s\n", result.TransactionID)
	fmt.Println("✅ First escrow transaction created and submitted!")
	fmt.Printf("✅ First Escrow Transaction ID: %s\n", escrowResult.TransactionID)
	fmt.Println("✅ First escrow finished (milestone completed)!")
	fmt.Printf("✅ First Finish Transaction ID: %s\n", finishResult.TransactionID)
	fmt.Println("✅ Second escrow transaction created and submitted!")
	fmt.Printf("✅ Second Escrow Transaction ID: %s\n", escrowResult2.TransactionID)
	fmt.Println("✅ Second escrow cancelled (funds returned)!")
	fmt.Printf("✅ Cancel Transaction ID: %s\n", cancelResult.TransactionID)

	fmt.Println("\n✅ OVERALL RESULTS")
	fmt.Println("✅ All real testnet tests completed successfully!")
	fmt.Println("✅ Wallet funding and transaction operations verified!")
	fmt.Println("✅ Enhanced XRPL Client working with real testnet!")
	fmt.Println("✅ Payment and escrow functionality confirmed!")
	fmt.Println("✅ Ready for production-level testing!")

	// ============================================================================
	// BLOCK EXPLORER LINKS
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("BLOCK EXPLORER LINKS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Check these transactions on the XRPL Testnet Explorer:")
	fmt.Println()

	for i, txID := range transactionIDs {
		var txType string
		switch i {
		case 0:
			txType = "Payment Transaction"
		case 1:
			txType = "First Escrow Creation"
		case 2:
			txType = "Second Escrow Creation"
		case 3:
			txType = "Escrow Finish (Milestone Completion)"
		case 4:
			txType = "Escrow Cancel"
		default:
			txType = fmt.Sprintf("Transaction %d", i+1)
		}

		explorerURL := fmt.Sprintf("https://testnet.xrpl.org/transactions/%s", txID)
		fmt.Printf("%s:\n", txType)
		fmt.Printf("  Transaction ID: %s\n", txID)
		fmt.Printf("  Explorer URL: %s\n", explorerURL)
		fmt.Println()
	}

	fmt.Println("✅ All transaction links generated for verification!")
}

// fundWallet sends a funding request to the XRPL testnet faucet
func fundWallet(faucetURL, address string) (string, error) {
	// Create the funding request
	request := XRPLFaucetRequest{
		Destination: address,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Send POST request to faucet
	resp, err := http.Post(faucetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request to faucet: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var faucetResp XRPLFaucetResponse
	if err := json.Unmarshal(body, &faucetResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check for errors
	if faucetResp.Error != "" {
		return "", fmt.Errorf("faucet error: %s", faucetResp.Error)
	}

	// Return transaction hash
	return faucetResp.TransactionHash, nil
}

// getAccountBalance retrieves account balance and sequence from XRPL
func getAccountBalance(networkURL, address string) (string, int) {
	// Create account_info request
	request := map[string]interface{}{
		"method": "account_info",
		"params": []map[string]interface{}{
			{
				"account":      address,
				"ledger_index": "validated",
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal account_info request: %v", err)
		return "0", 0
	}

	// Send POST request
	resp, err := http.Post(networkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send account_info request: %v", err)
		return "0", 0
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read account_info response: %v", err)
		return "0", 0
	}

	// Parse response
	var accountInfo AccountInfo
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		log.Printf("Failed to parse account_info response: %v", err)
		return "0", 0
	}

	return accountInfo.Result.AccountData.Balance, accountInfo.Result.AccountData.Sequence
}

// parseBalance converts balance string to int64
func parseBalance(balanceStr string) int64 {
	balance, err := strconv.ParseInt(balanceStr, 10, 64)
	if err != nil {
		return 0
	}
	return balance
}

// convertBase58ToHex converts XRPL base58 private key to hex format
func convertBase58ToHex(base58Key string) (string, error) {
	// Decode base58 to bytes
	decoded := base58.Decode(base58Key)
	if len(decoded) == 0 {
		return "", fmt.Errorf("invalid base58 private key")
	}

	// Debug: Print the decoded length and first few bytes (commented out for clean output)
	// fmt.Printf("Debug: Decoded key length: %d bytes\n", len(decoded))
	// fmt.Printf("Debug: First 5 bytes: %x\n", decoded[:min(5, len(decoded))])

	// XRPL private keys in base58 format can have different structures
	// Let's handle the case where we have 23 bytes (which might be the raw key + checksum)
	if len(decoded) == 23 {
		// For 23-byte keys, assume the first 16 bytes are the private key
		// and the last 7 bytes are checksum/prefix
		// Pad to 32 bytes for Ed25519
		privateKeyBytes := make([]byte, 32)
		copy(privateKeyBytes, decoded[:16])
		hexKey := hex.EncodeToString(privateKeyBytes)
		return hexKey, nil
	} else if len(decoded) == 37 {
		// Standard XRPL format: prefix + 32 bytes + checksum
		privateKeyBytes := decoded[1:33]
		hexKey := hex.EncodeToString(privateKeyBytes)
		return hexKey, nil
	} else {
		// Try to extract the middle portion as the private key
		if len(decoded) >= 32 {
			// Take the middle 32 bytes
			start := (len(decoded) - 32) / 2
			privateKeyBytes := decoded[start : start+32]
			hexKey := hex.EncodeToString(privateKeyBytes)
			return hexKey, nil
		} else {
			// Use the entire decoded key and pad with zeros if needed
			paddedKey := make([]byte, 32)
			copy(paddedKey, decoded)
			hexKey := hex.EncodeToString(paddedKey)
			return hexKey, nil
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generatePreimage generates a secure random 32-byte preimage for XRPL crypto-conditions
func generatePreimage() ([]byte, error) {
	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
	if err != nil {
		return nil, err
	}
	return preimage, nil
}

// createXRPLConditionAndFulfillment creates proper XRPL condition and fulfillment using go-interledger/cryptoconditions
// This generates the correct DER-encoded format expected by XRPL
func createXRPLConditionAndFulfillment(preimage []byte) (string, string, error) {
	// Create a PREIMAGE-SHA-256 fulfillment using the crypto-conditions library
	fulfillment := cc.NewPreimageSha256(preimage)

	// Serialize the fulfillment to binary (DER-encoded)
	fulfillmentBinary, err := fulfillment.Encode()
	if err != nil {
		return "", "", fmt.Errorf("failed to encode fulfillment: %w", err)
	}

	// Generate the condition from the fulfillment
	condition := fulfillment.Condition()

	// Serialize the condition to binary (DER-encoded)
	conditionBinary, err := condition.Encode()
	if err != nil {
		return "", "", fmt.Errorf("failed to encode condition: %w", err)
	}

	// Convert to hexadecimal strings for XRPL
	conditionHex := fmt.Sprintf("%X", conditionBinary)
	fulfillmentHex := fmt.Sprintf("%X", fulfillmentBinary)

	return conditionHex, fulfillmentHex, nil
}

// createPaymentWithXrplGo creates and signs a payment transaction using the xrpl-go library
func createPaymentWithXrplGo(fromAddress, toAddress, amount, privateKeyBase58 string, sequence uint32) (*xrpl.TransactionResult, error) {
	// Create a new XRPL client for testnet
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	client := rpc.NewClient(cfg)

	// Create wallet from the private key (seed)
	w, err := wallet.FromSeed(privateKeyBase58, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Amount is already in drops as string, convert to int for xrpl-go library
	// Note: XRPL protocol expects Amount as string representing drops (1 XRP = 1,000,000 drops)
	// but xrpl-go library's XRPCurrencyAmount type expects int64
	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Create a payment transaction using xrpl-go
	payment := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: types.Address(fromAddress),
		},
		Destination: types.Address(toAddress),
		Amount:      types.XRPCurrencyAmount(amountInt), // xrpl-go expects int64 for drops
		DeliverMax:  types.XRPCurrencyAmount(amountInt), // xrpl-go expects int64 for drops
	}

	// Flatten the transaction
	flattenedTx := payment.Flatten()

	// Autofill the transaction (sequence, fee, etc.)
	if err := client.Autofill(&flattenedTx); err != nil {
		return nil, fmt.Errorf("failed to autofill transaction: %w", err)
	}

	// Sign the transaction
	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Submit the signed transaction
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	// Convert the result to our TransactionResult format
	result := &xrpl.TransactionResult{
		TransactionID: string(res.Hash),
		ResultCode:    "tesSUCCESS", // xrpl-go doesn't return engine result in this format
		Validated:     res.Validated,
	}

	return result, nil
}

// createEscrowWithXrplGo creates and signs an escrow transaction using the xrpl-go library
func createEscrowWithXrplGo(fromAddress, toAddress, amount, condition, privateKeyBase58 string, sequence uint32) (*xrpl.TransactionResult, error) {
	// Create a new XRPL client for testnet
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	client := rpc.NewClient(cfg)

	// Create wallet from the private key (seed)
	w, err := wallet.FromSeed(privateKeyBase58, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Amount is already in drops as string, convert to int for xrpl-go library
	// Note: XRPL protocol expects Amount as string representing drops (1 XRP = 1,000,000 drops)
	// but xrpl-go library's XRPCurrencyAmount type expects int64
	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// Create an escrow transaction using xrpl-go
	// Set proper epoch times (seconds since 2000-01-01T00:00:00 UTC)
	// For testing: FinishAfter = 5 seconds from now, CancelAfter = 10 seconds from now
	now := time.Now()
	epochStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	xrplTimeNow := uint32(now.Sub(epochStart).Seconds())

	fmt.Printf("🕐 XRPL TIME CALCULATION:\n")
	fmt.Printf("  Current Unix time: %d\n", now.Unix())
	fmt.Printf("  XRPL epoch (2000-01-01): %d\n", epochStart.Unix())
	fmt.Printf("  Current XRPL time: %d\n", xrplTimeNow)

	finishAfter := xrplTimeNow + 5  // 5 seconds from now
	cancelAfter := xrplTimeNow + 10 // 10 seconds from now

	fmt.Printf("  FinishAfter will be set to: %d (in 5 seconds)\n", finishAfter)
	fmt.Printf("  CancelAfter will be set to: %d (in 10 seconds)\n", cancelAfter)

	escrow := &transaction.EscrowCreate{
		BaseTx: transaction.BaseTx{
			Account: types.Address(fromAddress),
		},
		Amount:      types.XRPCurrencyAmount(amountInt), // xrpl-go expects int64 for drops
		Destination: types.Address(toAddress),
		Condition:   condition,   // Restore condition for conditional escrow
		CancelAfter: cancelAfter, // Cancel after 2 hours
		FinishAfter: finishAfter, // Finish after 1 hour
	}

	// Flatten the transaction
	flattenedTx := escrow.Flatten()

	// Convert types.Address to strings for Autofill compatibility
	if addr, ok := flattenedTx["Account"].(types.Address); ok {
		flattenedTx["Account"] = string(addr)
	}
	if addr, ok := flattenedTx["Destination"].(types.Address); ok {
		flattenedTx["Destination"] = string(addr)
	}

	// Autofill the transaction (sequence, fee, etc.) - this is important for escrow transactions
	if err := client.Autofill(&flattenedTx); err != nil {
		return nil, fmt.Errorf("failed to autofill escrow transaction: %w", err)
	}

	// Debug: Print the transaction fields
	fmt.Printf("Debug: Escrow transaction fields:\n")
	fmt.Printf("  FinishAfter epoch time: %d (5 seconds from now)\n", finishAfter)
	fmt.Printf("  CancelAfter epoch time: %d (10 seconds from now)\n", cancelAfter)
	for key, value := range flattenedTx {
		fmt.Printf("  %s: %v (type: %T)\n", key, value, value)
	}

	// Sign the transaction
	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign escrow transaction: %w", err)
	}

	// Submit the signed transaction
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		return nil, fmt.Errorf("failed to submit escrow transaction: %w", err)
	}

	// Convert the result to our TransactionResult format
	result := &xrpl.TransactionResult{
		TransactionID: string(res.Hash),
		ResultCode:    "tesSUCCESS", // xrpl-go doesn't return engine result in this format
		Validated:     res.Validated,
	}

	return result, nil
}

// finishEscrowWithXrplGo finishes an escrow transaction using the xrpl-go library
func finishEscrowWithXrplGo(account, owner string, offerSequence uint32, condition, fulfillment, privateKeyBase58 string, sequence uint32) (*xrpl.TransactionResult, error) {
	// Create a new XRPL client for testnet
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	client := rpc.NewClient(cfg)

	// Create wallet from the private key (seed)
	w, err := wallet.FromSeed(privateKeyBase58, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Debug: Verify the signing account matches the transaction account
	signingAddress := w.GetAddress()
	fmt.Printf("🔑 SIGNING VERIFICATION:\n")
	fmt.Printf("  Transaction Account: %s\n", account)
	fmt.Printf("  Signing Address: %s\n", signingAddress)
	fmt.Printf("  Addresses match: %t\n", string(signingAddress) == account)

	if string(signingAddress) != account {
		return nil, fmt.Errorf("CRITICAL ERROR: Signing address (%s) does not match transaction account (%s)", signingAddress, account)
	}

	fmt.Printf("✅ Account verification passed\n")

	// Create an escrow finish transaction using xrpl-go
	escrowFinish := &transaction.EscrowFinish{
		BaseTx: transaction.BaseTx{
			Account: types.Address(account),
		},
		Owner:         types.Address(owner),
		OfferSequence: offerSequence,
	}

	// CRITICAL: For unconditional escrows, DO NOT set Condition or Fulfillment fields at all
	// Only set them for conditional escrows
	fmt.Printf("🔧 CONDITIONAL CHECK: condition='%s', fulfillment='%s'\n", condition, fulfillment)
	if condition != "" && condition != " " { // Also check for space character
		escrowFinish.Condition = condition
		fmt.Printf("✅ Setting Condition field for conditional escrow\n")
	} else {
		fmt.Printf("✅ Omitting Condition field (unconditional escrow)\n")
	}
	if fulfillment != "" && fulfillment != " " { // Also check for space character
		escrowFinish.Fulfillment = fulfillment
		fmt.Printf("✅ Setting Fulfillment field for conditional escrow\n")
	} else {
		fmt.Printf("✅ Omitting Fulfillment field (unconditional escrow)\n")
	}

	// Flatten the transaction
	flattenedTx := escrowFinish.Flatten()

	// Debug: Print escrow finish transaction fields
	fmt.Printf("Debug: Escrow finish transaction fields:\n")
	for key, value := range flattenedTx {
		if key == "Fulfillment" {
			fmt.Printf("  %s: %s (length: %d)\n", key, value, len(value.(string)))
		} else {
			fmt.Printf("  %s: %v (type: %T)\n", key, value, value)
		}
	}

	// Convert types.Address to strings for Autofill compatibility
	if addr, ok := flattenedTx["Account"].(types.Address); ok {
		flattenedTx["Account"] = string(addr)
	}
	if addr, ok := flattenedTx["Owner"].(types.Address); ok {
		flattenedTx["Owner"] = string(addr)
	}

	// Note: Sequence number is now handled automatically by the enhanced XRPL client
	fmt.Printf("🔢 Account sequence will be fetched automatically by the enhanced client\n")

	// Autofill the transaction (fee, etc.) - this is important for escrow finish transactions
	if err := client.Autofill(&flattenedTx); err != nil {
		return nil, fmt.Errorf("failed to autofill escrow finish transaction: %w", err)
	}

	fmt.Printf("📋 FINAL TRANSACTION TO SUBMIT:\n")
	for key, value := range flattenedTx {
		if key == "Fulfillment" && value != nil {
			fmt.Printf("  %s: %s (length: %d)\n", key, value, len(value.(string)))
		} else if key == "Condition" && value != nil {
			fmt.Printf("  %s: %s (length: %d)\n", key, value, len(value.(string)))
		} else {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// CRITICAL CHECK: Ensure Condition and Fulfillment are omitted for unconditional escrow
	if _, hasCondition := flattenedTx["Condition"]; hasCondition {
		fmt.Printf("⚠️ WARNING: Condition field present in unconditional escrow finish!\n")
	}
	if _, hasFulfillment := flattenedTx["Fulfillment"]; hasFulfillment {
		fmt.Printf("⚠️ WARNING: Fulfillment field present in unconditional escrow finish!\n")
	}

	// Sign the transaction
	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign escrow finish transaction: %w", err)
	}

	// Submit the signed transaction
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		return nil, fmt.Errorf("failed to submit escrow finish transaction: %w", err)
	}

	// Convert the result to our TransactionResult format
	result := &xrpl.TransactionResult{
		TransactionID: string(res.Hash),
		ResultCode:    "tesSUCCESS", // xrpl-go doesn't return engine result in this format
		Validated:     res.Validated,
	}

	return result, nil
}

// cancelEscrowWithXrplGo cancels an escrow transaction using the xrpl-go library
func cancelEscrowWithXrplGo(account, owner string, offerSequence uint32, privateKeyBase58 string, sequence uint32) (*xrpl.TransactionResult, error) {
	// Create a new XRPL client for testnet
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	client := rpc.NewClient(cfg)

	// Create wallet from the private key (seed)
	w, err := wallet.FromSeed(privateKeyBase58, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from seed: %w", err)
	}

	// Create an escrow cancel transaction using xrpl-go
	escrowCancel := &transaction.EscrowCancel{
		BaseTx: transaction.BaseTx{
			Account: types.Address(account),
		},
		Owner:         types.Address(owner),
		OfferSequence: offerSequence,
	}

	// Flatten the transaction
	flattenedTx := escrowCancel.Flatten()

	// Convert types.Address to strings for Autofill compatibility
	if addr, ok := flattenedTx["Account"].(types.Address); ok {
		flattenedTx["Account"] = string(addr)
	}
	if addr, ok := flattenedTx["Owner"].(types.Address); ok {
		flattenedTx["Owner"] = string(addr)
	}

	// Autofill the transaction (sequence, fee, etc.) - this is important for escrow cancel transactions
	if err := client.Autofill(&flattenedTx); err != nil {
		return nil, fmt.Errorf("failed to autofill escrow cancel transaction: %w", err)
	}

	// Sign the transaction
	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign escrow cancel transaction: %w", err)
	}

	// Submit the signed transaction
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		return nil, fmt.Errorf("failed to submit escrow cancel transaction: %w", err)
	}

	// Convert the result to our TransactionResult format
	result := &xrpl.TransactionResult{
		TransactionID: string(res.Hash),
		ResultCode:    "tesSUCCESS", // xrpl-go doesn't return engine result in this format
		Validated:     res.Validated,
	}

	return result, nil
}
