package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/smart-payment-infrastructure/internal/config"
)

func main() {
	accountAddress := flag.String("account", "", "XRPL address to query escrows for")
	flag.Parse()

	if *accountAddress == "" {
		log.Fatal("account address is required")
	}

	cfg := config.Load()

	// Use HTTP URL for JSON-RPC instead of WebSocket URL
	networkURL := "https://s.altnet.rippletest.net:51234"
	if cfg.XRPL.NetworkURL != "" && !strings.Contains(cfg.XRPL.NetworkURL, "wss://") {
		networkURL = cfg.XRPL.NetworkURL
	}

	log.Printf("🔍 XRPL Escrow Query")
	log.Printf("=====================")
	log.Printf("Network URL: %s", networkURL)
	log.Printf("Account: %s", *accountAddress)

	// Query account objects (escrows)
	log.Printf("\n💰 Querying Account Objects (Escrows)...")

	// Create JSON-RPC request for account_objects
	requestBody := map[string]interface{}{
		"method": "account_objects",
		"params": []map[string]interface{}{
			{
				"account":      *accountAddress,
				"ledger_index": "validated",
				"type":         "escrow",
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
		log.Fatalf("❌ Failed to query account objects: %v", err)
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

		log.Printf("✅ Account objects queried successfully!")
		log.Printf("Result: %+v", result)

		// Check if there are any escrow objects
		if accountObjects, ok := result["account_objects"].([]interface{}); ok {
			log.Printf("\n📋 Found %d account objects", len(accountObjects))

			for i, obj := range accountObjects {
				if objMap, ok := obj.(map[string]interface{}); ok {
					log.Printf("\n--- Object %d ---", i+1)
					log.Printf("LedgerEntryType: %v", objMap["LedgerEntryType"])
					log.Printf("Index: %v", objMap["index"])

					if objMap["LedgerEntryType"] == "Escrow" {
						log.Printf("🎯 ESCROW FOUND!")
						log.Printf("Account: %v", objMap["Account"])
						log.Printf("Destination: %v", objMap["Destination"])
						log.Printf("Amount: %v", objMap["Amount"])
						log.Printf("FinishAfter: %v", objMap["FinishAfter"])
						log.Printf("CancelAfter: %v", objMap["CancelAfter"])
						log.Printf("Condition: %v", objMap["Condition"])
						log.Printf("Owner: %v", objMap["Owner"])
						log.Printf("PreviousTxnID: %v", objMap["PreviousTxnID"])
						log.Printf("PreviousTxnLgrSeq: %v", objMap["PreviousTxnLgrSeq"])
					}
				}
			}
		} else {
			log.Printf("No account objects found or invalid format")
		}
	} else {
		log.Printf("⚠️  Response format: %v", response)
	}
}
