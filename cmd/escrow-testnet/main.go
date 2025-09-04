package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/smart-payment-infrastructure/internal/config"
)

func main() {
	payerSecret := flag.String("payer-secret", "", "Base58-encoded secret for payer account")
	payerAddress := flag.String("payer-address", "", "XRPL address of payer")
	payeeAddress := flag.String("payee-address", "", "XRPL address of payee")
	amount := flag.Float64("amount", 5.0, "Amount in XRP to lock")
	flag.Parse()

	if *payerSecret == "" || *payerAddress == "" || *payeeAddress == "" {
		log.Fatal("payer-secret, payer-address, payee-address are required")
	}

	cfg := config.Load()

	// Use HTTP URL for JSON-RPC instead of WebSocket URL
	networkURL := "https://s.altnet.rippletest.net:51234"
	if cfg.XRPL.NetworkURL != "" && !strings.Contains(cfg.XRPL.NetworkURL, "wss://") {
		networkURL = cfg.XRPL.NetworkURL
	}

	log.Printf("🚀 XRPL Escrow Test")
	log.Printf("==================")
	log.Printf("Network URL: %s", networkURL)
	log.Printf("TestNet: %v", cfg.XRPL.TestNet)
	log.Printf("Payer: %s", *payerAddress)
	log.Printf("Payee: %s", *payeeAddress)
	log.Printf("Amount: %.6f XRP", *amount)

	// Create escrow using the working xrpl-go library pattern
	log.Printf("\n💰 Creating Escrow...")

	// Create wallet from secret
	w, err := wallet.FromSecret(*payerSecret)
	if err != nil {
		log.Fatalf("❌ Failed to create wallet from secret: %v", err)
	}

	// Get current account sequence number
	accountInfo, err := getAccountInfo(*payerAddress, networkURL)
	if err != nil {
		log.Fatalf("❌ Failed to get account info: %v", err)
	}

	// Get current ledger index for LastLedgerSequence
	currentLedger, err := getCurrentLedger(networkURL)
	if err != nil {
		log.Fatalf("❌ Failed to get current ledger: %v", err)
	}

	// Convert time to Ripple epoch (seconds since 2000-01-01T00:00:00Z)
	rippleEpoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	// Set FinishAfter to at least 15 seconds from now (following XRPL best practices)
	// This ensures at least 3-5 ledger close intervals to avoid timing race conditions
	finishAfter := uint32(now.Add(15 * time.Second).Sub(rippleEpoch).Seconds())

	// Set CancelAfter to 1 hour from now (sufficiently later than FinishAfter)
	// This allows time for finishing while providing a reasonable cancellation window
	cancelAfter := uint32(now.Add(1 * time.Hour).Sub(rippleEpoch).Seconds())

	// Create escrow transaction as a map (compatible with xrpl-go)
	// According to XRPL guidelines, we need at least one of FinishAfter, Condition, or CancelAfter
	escrowTx := map[string]interface{}{
		"TransactionType":    "EscrowCreate",
		"Account":            *payerAddress,
		"Destination":        *payeeAddress,
		"Amount":             fmt.Sprintf("%d", int64(*amount*1000000)), // Convert XRP to drops
		"FinishAfter":        finishAfter,                               // Allow finish after 1 hour
		"CancelAfter":        cancelAfter,                               // Allow cancel after 24 hours
		"Fee":                "12",
		"Sequence":           uint32(accountInfo.Sequence),
		"LastLedgerSequence": uint32(currentLedger + 4), // Transaction expiration
		"Flags":              uint32(2147483648),        // tfFullyCanonicalSig flag
	}

	// Sign the transaction using the library
	log.Println("Signing escrow transaction with xrpl-go...")
	txBlob, txID, err := w.Sign(escrowTx)
	if err != nil {
		log.Fatalf("❌ Failed to sign escrow transaction: %v", err)
	}

	log.Printf("✅ Escrow transaction signed successfully!")
	log.Printf("Transaction ID: %s", txID)
	log.Printf("Transaction blob length: %d bytes", len(txBlob))
	log.Printf("FinishAfter: %d (Ripple epoch seconds)", finishAfter)
	log.Printf("CancelAfter: %d (Ripple epoch seconds)", cancelAfter)

	// Submit the transaction to XRPL testnet
	log.Println("\n🌐 Submitting escrow transaction to XRPL testnet...")

	// Create JSON-RPC request
	requestBody := map[string]interface{}{
		"method": "submit",
		"params": []map[string]interface{}{
			{
				"tx_blob": txBlob,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("❌ Failed to marshal request: %v", err)
	}

	// Submit to XRPL
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(networkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("❌ Failed to submit transaction: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("❌ Failed to decode response: %v", err)
	}

	// Check for errors
	if result, ok := response["result"].(map[string]interface{}); ok {
		if result["error"] != nil {
			log.Fatalf("❌ XRPL error: %v", result["error"])
		}

		log.Printf("✅ Escrow transaction submitted successfully!")
		log.Printf("Result: %v", result)
	} else {
		log.Printf("⚠️  Response format: %v", response)
	}
}

// getAccountInfo gets account information from XRPL
func getAccountInfo(address, networkURL string) (*AccountInfo, error) {
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(networkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if result["error"] != nil {
			return nil, fmt.Errorf("XRPL error: %v", result["error"])
		}

		// Debug: log the available fields
		log.Printf("Account info response fields: %+v", result)

		// Check for account_sequence_available first (this should be the correct next sequence)
		if availableSequence, ok := result["account_sequence_available"].(float64); ok {
			log.Printf("Using account_sequence_available: %f", availableSequence)
			return &AccountInfo{
				Sequence: uint32(availableSequence),
			}, nil
		}

		// Fallback to account_sequence_next
		if nextSequence, ok := result["account_sequence_next"].(float64); ok {
			log.Printf("Using account_sequence_next: %f", nextSequence)
			return &AccountInfo{
				Sequence: uint32(nextSequence),
			}, nil
		}

		// Last resort: current sequence (not +1)
		if accountData, ok := result["account_data"].(map[string]interface{}); ok {
			sequence, _ := accountData["Sequence"].(float64)
			log.Printf("Using account_data.Sequence: %f", sequence)
			return &AccountInfo{
				Sequence: uint32(sequence),
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to parse account info response")
}

// getCurrentLedger gets the current ledger index from XRPL
func getCurrentLedger(networkURL string) (uint32, error) {
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

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(networkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to get ledger info: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if result["error"] != nil {
			return 0, fmt.Errorf("XRPL error: %v", result["error"])
		}

		// Debug: log the response structure
		log.Printf("Ledger response: %+v", result)

		// Check if ledger_index is at the top level of result
		if ledgerIndex, ok := result["ledger_index"].(float64); ok {
			return uint32(ledgerIndex), nil
		}

		// Also check inside the ledger object
		if ledger, ok := result["ledger"].(map[string]interface{}); ok {
			if ledgerIndex, ok := ledger["ledger_index"].(float64); ok {
				return uint32(ledgerIndex), nil
			}
		}
	}

	return 0, fmt.Errorf("failed to parse ledger response: %+v", response)
}

// AccountInfo represents basic account information
type AccountInfo struct {
	Sequence uint32
}
