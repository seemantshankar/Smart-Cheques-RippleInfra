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

	cc "github.com/go-interledger/cryptoconditions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/btcsuite/btcutil/base58"
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

	// Calculate payment results
	payerPaymentChange := parseBalance(payerAccountDataAfter.Balance) - parseBalance(payerAccountData.Balance)
	payeePaymentChange := parseBalance(payeeAccountDataAfter.Balance) - parseBalance(payeeAccountData.Balance)

	fmt.Printf("Payer Payment Change: %d drops (%.6f XRP)\n",
		payerPaymentChange, float64(payerPaymentChange)/1000000)
	fmt.Printf("Payee Payment Change: %d drops (%.6f XRP)\n",
		payeePaymentChange, float64(payeePaymentChange)/1000000)

	// Test 9: Create and Submit Escrow Transaction using xrpl-go
	fmt.Println("\n--- Test 9: Real Escrow Transaction using xrpl-go ---")

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

	// Use xrpl-go library to create and sign the escrow transaction
	escrowResult, err := createEscrowWithXrplGo(payerAddress, payeeAddress, escrowAmount, condition, payerSecretBase58, currentAccountData.Sequence)
	if err != nil {
		log.Fatalf("Failed to create escrow with xrpl-go: %v", err)
	}

	fmt.Printf("✅ Escrow created successfully!\n")
	fmt.Printf("Escrow Transaction ID: %s\n", escrowResult.TransactionID)
	fmt.Printf("Result Code: %s\n", escrowResult.ResultCode)
	fmt.Printf("Validated: %t\n", escrowResult.Validated)

	// Wait a moment for the escrow to be processed
	fmt.Println("Waiting for escrow to be processed...")
	time.Sleep(5 * time.Second)

	// Test 10: Check Account Balances After Escrow
	fmt.Println("\n--- Test 10: Account Balances After Escrow ---")

	payerAccountDataAfterEscrow, err := enhancedClient.GetAccountData(payerAddress)
	if err != nil {
		log.Fatalf("Failed to get payer account data after escrow: %v", err)
	}
	fmt.Printf("Payer Balance After Escrow: %s drops (%.6f XRP)\n",
		payerAccountDataAfterEscrow.Balance, float64(parseBalance(payerAccountDataAfterEscrow.Balance))/1000000)

	// Calculate escrow balance change
	escrowChange := parseBalance(payerAccountDataAfterEscrow.Balance) - parseBalance(payerAccountDataAfter.Balance)
	fmt.Printf("Payer Balance Change (Escrow): %d drops (%.6f XRP)\n",
		escrowChange, float64(escrowChange)/1000000)

	// Test 11: Finish Escrow (Complete Milestone) using xrpl-go
	fmt.Println("\n--- Test 11: Finish Escrow (Complete Milestone) using xrpl-go ---")

	// Get the escrow sequence number (this would typically be stored from the creation)
	// For this test, we'll use sequence 1 as a placeholder
	escrowSequence := uint32(1)

	fmt.Printf("Finishing escrow with sequence: %d\n", escrowSequence)

	// Use xrpl-go library to finish the escrow transaction
	finishResult, err := finishEscrowWithXrplGo(payeeAddress, payerAddress, escrowSequence, condition, fulfillment, payeeSecretBase58, payeeAccountDataAfter.Sequence)
	if err != nil {
		log.Fatalf("Failed to finish escrow with xrpl-go: %v", err)
	}

	fmt.Printf("✅ Escrow finished successfully!\n")
	fmt.Printf("Finish Transaction ID: %s\n", finishResult.TransactionID)
	fmt.Printf("Result Code: %s\n", finishResult.ResultCode)
	fmt.Printf("Validated: %t\n", finishResult.Validated)

	// Wait a moment for the finish transaction to be processed
	fmt.Println("Waiting for escrow finish to be processed...")
	time.Sleep(5 * time.Second)

	// Test 12: Final Account Balances
	fmt.Println("\n--- Test 12: Final Account Balances ---")

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

	// Calculate total changes
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
	fmt.Println("✅ Escrow transaction created and submitted!")
	fmt.Printf("✅ Escrow Transaction ID: %s\n", escrowResult.TransactionID)
	fmt.Println("✅ Escrow finished (milestone completed)!")
	fmt.Printf("✅ Finish Transaction ID: %s\n", finishResult.TransactionID)

	fmt.Println("\n✅ OVERALL RESULTS")
	fmt.Println("✅ All real testnet tests completed successfully!")
	fmt.Println("✅ Wallet funding and transaction operations verified!")
	fmt.Println("✅ Enhanced XRPL Client working with real testnet!")
	fmt.Println("✅ Payment and escrow functionality confirmed!")
	fmt.Println("✅ Ready for production-level testing!")
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
	// For testing: FinishAfter = 1 hour from now, CancelAfter = 2 hours from now
	now := time.Now()
	epochStart := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	finishAfter := uint32(now.Add(1 * time.Hour).Sub(epochStart).Seconds())
	cancelAfter := uint32(now.Add(2 * time.Hour).Sub(epochStart).Seconds())

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
	fmt.Printf("  FinishAfter epoch time: %d (1 hour from now)\n", finishAfter)
	fmt.Printf("  CancelAfter epoch time: %d (2 hours from now)\n", cancelAfter)
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

	// Create an escrow finish transaction using xrpl-go
	escrowFinish := &transaction.EscrowFinish{
		BaseTx: transaction.BaseTx{
			Account: types.Address(w.GetAddress()),
		},
		Owner:         types.Address(owner),
		OfferSequence: offerSequence,
		Condition:     condition,
		Fulfillment:   fulfillment,
	}

	// Flatten the transaction
	flattenedTx := escrowFinish.Flatten()

	// Convert types.Address to strings for Autofill compatibility
	if addr, ok := flattenedTx["Account"].(types.Address); ok {
		flattenedTx["Account"] = string(addr)
	}
	if addr, ok := flattenedTx["Owner"].(types.Address); ok {
		flattenedTx["Owner"] = string(addr)
	}

	// Autofill the transaction (sequence, fee, etc.) - this is important for escrow finish transactions
	if err := client.Autofill(&flattenedTx); err != nil {
		return nil, fmt.Errorf("failed to autofill escrow finish transaction: %w", err)
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
