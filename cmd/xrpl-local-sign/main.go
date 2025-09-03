package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/smart-payment-infrastructure/test/config"
)

const (
	TestnetURL = "https://s.altnet.rippletest.net:51234"
)

func main() {
	log.Println("🚀 XRPL Local Signing Test")
	log.Println("==========================")

	// Load test wallet configuration
	testConfig := config.LoadTestConfig()
	log.Printf("Using testnet URL: %s", testConfig.NetworkURL)

	// Get source and destination wallet details
	sourceAddr, sourceSecret, _, _ := testConfig.GetTestWallet1()
	destAddr, _, _, _ := testConfig.GetTestWallet2()

	log.Printf("Source wallet: %s", sourceAddr)
	log.Printf("Destination wallet: %s", destAddr)

	// Check account balances and sequence
	log.Println("\n💰 Checking Account Balances")
	sourceBalance, sourceSequence, err := getAccountInfo(sourceAddr)
	if err != nil {
		log.Fatalf("❌ Failed to get source account info: %v", err)
	}
	log.Printf("Source balance: %s drops (%.6f XRP), sequence: %d",
		sourceBalance, float64(parseInt(sourceBalance))/1000000.0, sourceSequence)

	destBalance, destSequence, err := getAccountInfo(destAddr)
	if err != nil {
		log.Fatalf("❌ Failed to get destination account info: %v", err)
	}
	log.Printf("Destination balance: %s drops (%.6f XRP), sequence: %d",
		destBalance, float64(parseInt(destBalance))/1000000.0, destSequence)

	// Get current ledger for LastLedgerSequence
	currentLedger, err := getCurrentLedgerIndex()
	if err != nil {
		log.Fatalf("❌ Failed to get current ledger index: %v", err)
	}
	log.Printf("Current ledger index: %d", currentLedger)

	// Create and sign transaction using xrpl-go library
	log.Println("\n✍️ Creating and Signing Transaction Locally")

	// Create wallet from secret
	w, err := wallet.FromSecret(sourceSecret)
	if err != nil {
		log.Fatalf("❌ Failed to create wallet from secret: %v", err)
	}

	// Create payment transaction
	amount := "1" // 1 drop = 0.000001 XRP
	fee := "12"   // Standard fee

	// Create payment transaction as a map (compatible with any version of xrpl-go)
	// Note: Convert all integer values to uint32 to avoid type conversion issues
	// Remove NetworkID as it's not supported by the testnet
	payment := map[string]interface{}{
		"TransactionType":    "Payment",
		"Account":            sourceAddr,
		"Destination":        destAddr,
		"Amount":             amount,
		"Fee":                fee,
		"Sequence":           uint32(sourceSequence),
		"LastLedgerSequence": uint32(currentLedger + 4),
		"Flags":              uint32(2147483648), // tfFullyCanonicalSig flag
	}

	// Sign transaction locally
	log.Println("Signing transaction locally...")
	txBlob, txID, err := w.Sign(payment)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	log.Printf("✅ Transaction signed successfully")
	log.Printf("Transaction ID: %s", txID)
	log.Printf("Transaction blob length: %d", len(txBlob))
	log.Printf("Transaction blob: %s", txBlob)

	// Submit transaction
	log.Println("\n🌐 Submitting Transaction to XRPL Testnet")
	txHash, err := submitTransaction(txBlob)
	if err != nil {
		log.Fatalf("❌ Failed to submit transaction: %v", err)
	}

	log.Printf("✅ Transaction submitted successfully")
	log.Printf("Transaction hash: %s", txHash)

	// Monitor transaction
	log.Println("\n⏳ Monitoring Transaction...")
	ledgerIndex, validated, err := monitorTransaction(txHash, 60, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Failed to monitor transaction: %v", err)
	}

	log.Printf("✅ Transaction validation status: validated=%v, ledger_index=%d", validated, ledgerIndex)
	log.Printf("View transaction: https://testnet.xrpl.org/transactions/%s", txHash)

	// Check final balances
	log.Println("\n💰 Checking Final Account Balances")
	sourceBalanceAfter, _, err := getAccountInfo(sourceAddr)
	if err != nil {
		log.Printf("❌ Failed to get source account info: %v", err)
	} else {
		log.Printf("Source balance after: %s drops (%.6f XRP)",
			sourceBalanceAfter, float64(parseInt(sourceBalanceAfter))/1000000.0)
		log.Printf("Change: %.6f XRP",
			(float64(parseInt(sourceBalanceAfter))-float64(parseInt(sourceBalance)))/1000000.0)
	}

	destBalanceAfter, _, err := getAccountInfo(destAddr)
	if err != nil {
		log.Printf("❌ Failed to get destination account info: %v", err)
	} else {
		log.Printf("Destination balance after: %s drops (%.6f XRP)",
			destBalanceAfter, float64(parseInt(destBalanceAfter))/1000000.0)
		log.Printf("Change: %.6f XRP",
			(float64(parseInt(destBalanceAfter))-float64(parseInt(destBalance)))/1000000.0)
	}

	log.Println("\n🎉 Transaction test completed!")
	log.Printf("Transaction hash: %s", txHash)
	log.Printf("Ledger index: %d", ledgerIndex)
}

// getAccountInfo retrieves account information from XRPL
func getAccountInfo(address string) (balance string, sequence int, err error) {
	client := &http.Client{Timeout: 10 * time.Second}

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
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post(TestnetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("invalid response format")
	}

	if result["error"] != nil {
		return "", 0, fmt.Errorf("XRPL error: %v", result["error"])
	}

	accountData, ok := result["account_data"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("invalid account data format")
	}

	balance = toString(accountData["Balance"])
	sequence = toInt(accountData["Sequence"])

	return balance, sequence, nil
}

// getCurrentLedgerIndex gets the current ledger index
func getCurrentLedgerIndex() (int, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	requestBody := map[string]interface{}{
		"method": "ledger",
		"params": []map[string]interface{}{
			{
				"ledger_index": "validated",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post(TestnetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	ledgerIndex := toInt(result["ledger_index"])
	return ledgerIndex, nil
}

// submitTransaction submits a signed transaction blob to XRPL
func submitTransaction(txBlob string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	requestBody := map[string]interface{}{
		"method": "submit",
		"params": []map[string]interface{}{
			{
				"tx_blob":   txBlob,
				"fail_hard": false,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post(TestnetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	// Print the entire result for debugging
	resultBytes, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("Submit result: %s", string(resultBytes))

	if result["error"] != nil {
		return "", fmt.Errorf("XRPL error: %v", result["error"])
	}

	engineResult := toString(result["engine_result"])
	if engineResult != "tesSUCCESS" {
		return "", fmt.Errorf("XRPL engine result: %s - %s",
			engineResult, toString(result["engine_result_message"]))
	}

	txID := toString(result["tx_json"].(map[string]interface{})["hash"])
	return txID, nil
}

// monitorTransaction polls the rippled 'tx' method until transaction is validated or retries exhausted
func monitorTransaction(txHash string, maxRetries int, interval time.Duration) (int, bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < maxRetries; i++ {
		requestBody := map[string]interface{}{
			"method": "tx",
			"params": []map[string]interface{}{
				{
					"transaction": txHash,
					"binary":      false,
				},
			},
		}

		jsonData, _ := json.Marshal(requestBody)
		resp, err := client.Post(TestnetURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			// wait and retry
			time.Sleep(interval)
			continue
		}
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			time.Sleep(interval)
			continue
		}
		resp.Body.Close()

		if response["result"] == nil {
			time.Sleep(interval)
			continue
		}
		resMap, _ := response["result"].(map[string]interface{})
		if resMap["error"] != nil {
			// txn not found -> keep waiting
			if errStr, ok := resMap["error"].(string); ok && errStr == "txnNotFound" {
				log.Printf("Transaction not found yet (retry %d/%d)...", i+1, maxRetries)
				time.Sleep(interval)
				continue
			}
			return 0, false, fmt.Errorf("rippled tx error: %v", resMap["error"])
		}

		// Extract ledger_index and validated
		ledgerIndex := toInt(resMap["ledger_index"])
		validated := false
		if v, ok := resMap["validated"].(bool); ok {
			validated = v
		}

		// If found and validated (or not), return
		if ledgerIndex != 0 {
			return ledgerIndex, validated, nil
		}

		// Wait before next retry
		time.Sleep(interval)
	}

	return 0, false, fmt.Errorf("transaction not found after %d retries", maxRetries)
}

// Helper functions
func parseInt(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result
	default:
		return 0
	}
}
