package main

import (
	"fmt"
	"log"
	"time"

	"github.com/smart-payment-infrastructure/pkg/xrpl"
	"github.com/smart-payment-infrastructure/test/config"
)

const (
	TestnetURL = "https://s.altnet.rippletest.net:51234"
	TestnetID  = uint32(21338)
)

func main() {
	log.Println("🚀 XRPL Client Example")
	log.Println("=====================")

	// Load test wallet configuration
	testConfig := config.LoadTestConfig()
	log.Printf("Using testnet URL: %s", testConfig.NetworkURL)

	// Get source and destination wallet details
	sourceAddr, sourceSecret, _, _ := testConfig.GetTestWallet1()
	destAddr, _, _, _ := testConfig.GetTestWallet2()

	log.Printf("Source wallet: %s", sourceAddr)
	log.Printf("Destination wallet: %s", destAddr)

	// Create XRPL client
	client := xrpl.NewXRPLClient(TestnetURL, TestnetID)

	// Check account balances and sequence
	log.Println("\n💰 Checking Account Balances")
	sourceBalance, sourceSequence, err := client.GetAccountInfo(sourceAddr)
	if err != nil {
		log.Fatalf("❌ Failed to get source account info: %v", err)
	}
	log.Printf("Source balance: %s drops (%.6f XRP), sequence: %d",
		sourceBalance, float64(parseInt(sourceBalance))/1000000.0, sourceSequence)

	destBalance, destSequence, err := client.GetAccountInfo(destAddr)
	if err != nil {
		log.Fatalf("❌ Failed to get destination account info: %v", err)
	}
	log.Printf("Destination balance: %s drops (%.6f XRP), sequence: %d",
		destBalance, float64(parseInt(destBalance))/1000000.0, destSequence)

	// Create payment transaction
	amount := "1" // 1 drop = 0.000001 XRP
	log.Println("\n✍️ Creating Payment Transaction")
	payment, err := client.CreatePaymentTransaction(sourceAddr, destAddr, amount, sourceSequence)
	if err != nil {
		log.Fatalf("❌ Failed to create payment transaction: %v", err)
	}

	// Sign transaction
	log.Println("Signing transaction...")
	txBlob, txID, err := client.SignTransaction(payment, sourceSecret)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	log.Printf("✅ Transaction signed successfully")
	log.Printf("Transaction ID: %s", txID)
	log.Printf("Transaction blob length: %d", len(txBlob))

	// Submit transaction
	log.Println("\n🌐 Submitting Transaction to XRPL Testnet")
	txHash, err := client.SubmitTransaction(txBlob)
	if err != nil {
		log.Fatalf("❌ Failed to submit transaction: %v", err)
	}

	log.Printf("✅ Transaction submitted successfully")
	log.Printf("Transaction hash: %s", txHash)

	// Monitor transaction
	log.Println("\n⏳ Monitoring Transaction...")
	ledgerIndex, validated, err := client.MonitorTransaction(txHash, 60, 2*time.Second)
	if err != nil {
		log.Fatalf("❌ Failed to monitor transaction: %v", err)
	}

	log.Printf("✅ Transaction validation status: validated=%v, ledger_index=%d", validated, ledgerIndex)
	log.Printf("View transaction: https://testnet.xrpl.org/transactions/%s", txHash)

	// Check final balances
	log.Println("\n💰 Checking Final Account Balances")
	sourceBalanceAfter, _, err := client.GetAccountInfo(sourceAddr)
	if err != nil {
		log.Printf("❌ Failed to get source account info: %v", err)
	} else {
		log.Printf("Source balance after: %s drops (%.6f XRP)",
			sourceBalanceAfter, float64(parseInt(sourceBalanceAfter))/1000000.0)
		log.Printf("Change: %.6f XRP",
			(float64(parseInt(sourceBalanceAfter))-float64(parseInt(sourceBalance)))/1000000.0)
	}

	destBalanceAfter, _, err := client.GetAccountInfo(destAddr)
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

// Helper function to parse int64
func parseInt(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
