package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	ownerSecret := flag.String("owner-secret", "", "Base58-encoded secret for owner account (who created the escrow)")
	ownerAddress := flag.String("owner-address", "", "XRPL address of owner (who created the escrow)")
	payeeAddress := flag.String("payee-address", "", "XRPL address of payee (who will receive the funds)")
	sequence := flag.Uint("sequence", 0, "Transaction sequence number of the escrow creation")
	condition := flag.String("condition", "", "Escrow condition (SHA-256 hash)")
	fulfillment := flag.String("fulfillment", "", "Escrow fulfillment (the secret that satisfies the condition)")
	flag.Parse()

	if *ownerSecret == "" || *ownerAddress == "" || *payeeAddress == "" || *sequence == 0 {
		log.Fatal("owner-secret, owner-address, payee-address, and sequence are required")
	}

	// Use HTTP URL for JSON-RPC
	networkURL := "https://s.altnet.rippletest.net:51234"

	log.Printf("🚀 XRPL Escrow Finish Test")
	log.Printf("==========================")
	log.Printf("Network URL: %s", networkURL)
	log.Printf("Owner: %s", *ownerAddress)
	log.Printf("Payee: %s", *payeeAddress)
	log.Printf("Sequence: %d", *sequence)
	log.Printf("Condition: %s", *condition)
	log.Printf("Fulfillment: %s", *fulfillment)

	// First, let's check the escrow status
	log.Printf("\n🔍 Checking Escrow Status...")
	escrowInfo, err := getEscrowInfo(*ownerAddress, networkURL)
	if err != nil {
		log.Fatalf("❌ Failed to get escrow info: %v", err)
	}

	log.Printf("✅ Escrow found:")
	log.Printf("   Amount: %s drops", escrowInfo.Amount)
	log.Printf("   Destination: %s", escrowInfo.Destination)
	log.Printf("   FinishAfter: %d", escrowInfo.FinishAfter)
	log.Printf("   CancelAfter: %d", escrowInfo.CancelAfter)
	log.Printf("   Flags: %d", escrowInfo.Flags)

	// Check if escrow is ready to finish
	currentTime := getCurrentRippleTime()
	log.Printf("Current Ripple time: %d", currentTime)

	if currentTime < escrowInfo.FinishAfter {
		log.Printf("⏳ Escrow not ready to finish yet. FinishAfter: %d, Current: %d", escrowInfo.FinishAfter, currentTime)
		log.Printf("   Waiting %d seconds...", escrowInfo.FinishAfter-currentTime)
		return
	}

	log.Printf("✅ Escrow is ready to finish!")

	// Now let's finish the escrow
	log.Printf("\n💰 Finishing Escrow...")

	// Create wallet from secret
	w, err := wallet.FromSecret(*ownerSecret)
	if err != nil {
		log.Fatalf("❌ Failed to create wallet from secret: %v", err)
	}

	// Get current account sequence number for the owner
	accountInfo, err := getAccountInfo(*ownerAddress, networkURL)
	if err != nil {
		log.Fatalf("❌ Failed to get owner account info: %v", err)
	}

	// Create escrow finish transaction
	escrowFinishTx := map[string]interface{}{
		"TransactionType": "EscrowFinish",
		"Account":         *ownerAddress,     // The owner account submits the finish transaction
		"Owner":           *ownerAddress,     // The account that created the escrow
		"OfferSequence":   uint32(*sequence), // Use the sequence parameter (should be the escrow creation sequence)
		"Fee":             "400",
		"Sequence":        accountInfo.Sequence,
		"Flags":           uint32(2147483648), // tfFullyCanonicalSig flag
	}

	// Only add Condition and Fulfillment if they are provided and not empty
	if *condition != "" && *fulfillment != "" {
		escrowFinishTx["Condition"] = *condition
		escrowFinishTx["Fulfillment"] = *fulfillment
	}

	// Sign the transaction
	log.Println("Signing escrow finish transaction with xrpl-go...")
	txBlob, txID, err := w.Sign(escrowFinishTx)
	if err != nil {
		log.Fatalf("❌ Failed to sign escrow finish transaction: %v", err)
	}

	log.Printf("✅ Escrow finish transaction signed successfully!")
	log.Printf("Transaction ID: %s", txID)
	log.Printf("Transaction blob length: %d bytes", len(txBlob))

	// Submit the transaction to XRPL testnet
	log.Println("\n🌐 Submitting escrow finish transaction to XRPL testnet...")

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
			log.Printf("❌ XRPL error: %v", result["error"])

			// Check for specific error types
			if engineResult, ok := result["engine_result"].(string); ok {
				log.Printf("Engine Result: %s", engineResult)

				switch engineResult {
				case "tecNO_PERMISSION":
					log.Printf("🔒 Permission Error: The escrow cannot be finished yet or you don't have permission")
					log.Printf("   This usually means:")
					log.Printf("   - FinishAfter time hasn't passed yet")
					log.Printf("   - You're not the owner of the escrow")
					log.Printf("   - The escrow has already been finished or cancelled")
				case "tecNO_TARGET":
					log.Printf("🎯 Target Error: Escrow not found")
				case "temMALFORMED":
					log.Printf("📝 Malformed Error: Transaction format issue")
				default:
					log.Printf("❓ Unknown error type: %s", engineResult)
				}
			}

			if engineResultMessage, ok := result["engine_result_message"].(string); ok {
				log.Printf("Engine Result Message: %s", engineResultMessage)
			}

			return
		}

		log.Printf("✅ Escrow finish transaction submitted successfully!")
		log.Printf("Result: %v", result)

		// Check if transaction was applied
		if applied, ok := result["applied"].(bool); ok && applied {
			log.Printf("✅ Transaction applied to ledger!")
		}

		if validatedLedger, ok := result["validated_ledger_index"].(float64); ok {
			log.Printf("✅ Validated in ledger: %.0f", validatedLedger)
		}
	} else {
		log.Printf("⚠️  Response format: %v", response)
	}
}

// getEscrowInfo gets escrow information from XRPL and finds the first escrow ready to finish
func getEscrowInfo(ownerAddress string, networkURL string) (*EscrowInfo, error) {
	requestBody := map[string]interface{}{
		"method": "account_objects",
		"params": []map[string]interface{}{
			{
				"account":      ownerAddress,
				"ledger_index": "validated",
				"type":         "escrow",
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
		return nil, fmt.Errorf("failed to get escrow info: %w", err)
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

		if accountObjects, ok := result["account_objects"].([]interface{}); ok {
			currentTime := getCurrentRippleTime()

			for _, obj := range accountObjects {
				if escrowObj, ok := obj.(map[string]interface{}); ok {
					// Check if this is an escrow
					if escrowObj["LedgerEntryType"] == "Escrow" {
						// Check if it's ready to finish
						if finishAfter, ok := escrowObj["FinishAfter"].(float64); ok {
							if uint32(finishAfter) <= currentTime {
								log.Printf("Found escrow ready to finish: PreviousTxnLgrSeq: %v", escrowObj["PreviousTxnLgrSeq"])
								return parseEscrowInfo(escrowObj), nil
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no escrow ready to finish found")
}

// parseEscrowInfo parses escrow information from XRPL response
func parseEscrowInfo(escrowObj map[string]interface{}) *EscrowInfo {
	escrow := &EscrowInfo{}

	if amount, ok := escrowObj["Amount"].(string); ok {
		escrow.Amount = amount
	}
	if destination, ok := escrowObj["Destination"].(string); ok {
		escrow.Destination = destination
	}
	if finishAfter, ok := escrowObj["FinishAfter"].(float64); ok {
		escrow.FinishAfter = uint32(finishAfter)
	}
	if cancelAfter, ok := escrowObj["CancelAfter"].(float64); ok {
		escrow.CancelAfter = uint32(cancelAfter)
	}
	if flags, ok := escrowObj["Flags"].(float64); ok {
		escrow.Flags = uint32(flags)
	}
	if prevTxnLgrSeq, ok := escrowObj["PreviousTxnLgrSeq"].(float64); ok {
		escrow.Sequence = uint32(prevTxnLgrSeq)
	}

	return escrow
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

		// Check for account_sequence_available first
		if availableSequence, ok := result["account_sequence_available"].(float64); ok {
			return &AccountInfo{
				Sequence: uint32(availableSequence),
			}, nil
		}

		// Fallback to account_sequence_next
		if nextSequence, ok := result["account_sequence_next"].(float64); ok {
			return &AccountInfo{
				Sequence: uint32(nextSequence),
			}, nil
		}

		// Last resort: current sequence (not +1)
		if accountData, ok := result["account_data"].(map[string]interface{}); ok {
			if sequence, ok := accountData["Sequence"].(float64); ok {
				return &AccountInfo{
					Sequence: uint32(sequence),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse account info response")
}

// getCurrentRippleTime gets the current time in Ripple epoch seconds
func getCurrentRippleTime() uint32 {
	rippleEpoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	return uint32(now.Sub(rippleEpoch).Seconds())
}

// EscrowInfo represents escrow information from XRPL
type EscrowInfo struct {
	Amount      string `json:"Amount"`
	Destination string `json:"Destination"`
	FinishAfter uint32 `json:"FinishAfter"`
	CancelAfter uint32 `json:"CancelAfter"`
	Flags       uint32 `json:"Flags"`
	Sequence    uint32 `json:"Sequence"` // PreviousTxnLgrSeq from the escrow
}

// AccountInfo represents basic account information
type AccountInfo struct {
	Sequence uint32
}
