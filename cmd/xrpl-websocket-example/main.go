package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

func main() {
	// Configuration
	testnetURL := "https://s.altnet.rippletest.net:51234"

	// Create enhanced XRPL client
	client := xrpl.NewEnhancedClient(testnetURL, true)

	// Connect to XRPL network
	log.Println("Connecting to XRPL testnet...")
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Check if WebSocket connection is available
	if !client.IsWebSocketConnected() {
		log.Println("WebSocket not available, using HTTP fallback")
		return
	}

	log.Println("WebSocket connection established successfully!")

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Subscribe to ledger stream
	log.Println("Subscribing to ledger stream...")
	ledgerSubID, err := client.SubscribeToLedgerStream(func(msg *xrpl.StreamMessage) error {
		log.Printf("Ledger Update: %s", string(msg.Data))
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to ledger stream: %v", err)
	} else {
		log.Printf("Subscribed to ledger stream with ID: %s", ledgerSubID)
	}

	// Subscribe to transaction stream
	log.Println("Subscribing to transaction stream...")
	txSubID, err := client.SubscribeToTransactionStream(func(msg *xrpl.StreamMessage) error {
		log.Printf("Transaction: %s", string(msg.Transaction))
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to transaction stream: %v", err)
	} else {
		log.Printf("Subscribed to transaction stream with ID: %s", txSubID)
	}

	// Subscribe to validation stream
	log.Println("Subscribing to validation stream...")
	valSubID, err := client.SubscribeToValidationStream(func(msg *xrpl.StreamMessage) error {
		log.Printf("Validation: %s", string(msg.Validation))
		return nil
	})
	if err != nil {
		log.Printf("Failed to subscribe to validation stream: %v", err)
	} else {
		log.Printf("Subscribed to validation stream with ID: %s", valSubID)
	}

	// Show active subscriptions
	activeSubs := client.GetActiveSubscriptions()
	log.Printf("Active subscriptions: %d", len(activeSubs))
	for _, sub := range activeSubs {
		log.Printf("  - %s: %v", sub.ID, sub.Streams)
	}

	// Demonstrate WebSocket call
	log.Println("Testing WebSocket call...")
	response, err := client.WebSocketCall("server_info", nil)
	if err != nil {
		log.Printf("WebSocket call failed: %v", err)
	} else {
		log.Printf("Server info response: %+v", response)
	}

	// Keep the program running and handle shutdown
	log.Println("Monitoring XRPL streams... Press Ctrl+C to exit")

	select {
	case <-ctx.Done():
		log.Println("Context cancelled")
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	// Graceful shutdown
	log.Println("Shutting down...")

	// Unsubscribe from streams
	if ledgerSubID != "" {
		if err := client.UnsubscribeFromStream(ledgerSubID); err != nil {
			log.Printf("Failed to unsubscribe from ledger stream: %v", err)
		}
	}

	if txSubID != "" {
		if err := client.UnsubscribeFromStream(txSubID); err != nil {
			log.Printf("Failed to unsubscribe from transaction stream: %v", err)
		}
	}

	if valSubID != "" {
		if err := client.UnsubscribeFromStream(valSubID); err != nil {
			log.Printf("Failed to unsubscribe from validation stream: %v", err)
		}
	}

	// Close WebSocket connection
	if err := client.CloseWebSocket(); err != nil {
		log.Printf("Failed to close WebSocket: %v", err)
	}

	log.Println("Shutdown complete")
}

// Example callback functions for different stream types
func handleLedgerUpdate(msg *xrpl.StreamMessage) error {
	var ledgerData map[string]interface{}
	if err := json.Unmarshal(msg.Data, &ledgerData); err != nil {
		return err
	}

	if ledgerIndex, exists := ledgerData["ledger_index"]; exists {
		log.Printf("New ledger closed: %v", ledgerIndex)
	}

	return nil
}

func handleTransaction(msg *xrpl.StreamMessage) error {
	var txData map[string]interface{}
	if err := json.Unmarshal(msg.Transaction, &txData); err != nil {
		return err
	}

	if txType, exists := txData["TransactionType"]; exists {
		log.Printf("New transaction: %v", txType)
	}

	return nil
}

func handleValidation(msg *xrpl.StreamMessage) error {
	var valData map[string]interface{}
	if err := json.Unmarshal(msg.Validation, &valData); err != nil {
		return err
	}

	if ledgerIndex, exists := valData["ledger_index"]; exists {
		log.Printf("New validation for ledger: %v", ledgerIndex)
	}

	return nil
}
