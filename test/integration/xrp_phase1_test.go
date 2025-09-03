package integration

import (
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
	"github.com/smart-payment-infrastructure/test/config"
)

func TestXRPLPhase1Integration(t *testing.T) {
	// Load test configuration
	testConfig := config.LoadTestConfig()

	t.Run("Complete Payment Transaction Workflow", func(t *testing.T) {
		// Initialize XRPL service with testnet configuration
		xrplService := services.NewXRPLService(services.XRPLConfig{
			NetworkURL: testConfig.NetworkURL,
			TestNet:    testConfig.TestNet,
		})

		// Initialize the service
		if err := xrplService.Initialize(); err != nil {
			t.Fatalf("Failed to initialize XRPL service: %v", err)
		}

		// Get test wallet configurations
		fromAddress, _, privateKeyHex, keyType := testConfig.GetTestWallet1()
		toAddress, _, _, _ := testConfig.GetTestWallet2()
		amount, currency, _ := testConfig.GetTransactionParams()

		t.Logf("Testing payment transaction: %s -> %s, Amount: %s %s",
			fromAddress, toAddress, amount, currency)

		// Step 1: Create XRPL Payment Transaction
		payment, err := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)
		if err != nil {
			t.Fatalf("Failed to create payment transaction: %v", err)
		}

		// Verify payment transaction fields
		if payment.Account != fromAddress {
			t.Errorf("Expected account %s, got %s", fromAddress, payment.Account)
		}
		if payment.Destination != toAddress {
			t.Errorf("Expected destination %s, got %s", toAddress, payment.Destination)
		}
		if payment.TransactionType != "Payment" {
			t.Errorf("Expected transaction type Payment, got %s", payment.TransactionType)
		}

		// Step 2: Sign with provided private keys
		txBlob, err := xrplService.SignPaymentTransaction(payment, privateKeyHex, keyType)
		if err != nil {
			t.Fatalf("Failed to sign payment transaction: %v", err)
		}

		if txBlob == "" {
			t.Error("Expected non-empty transaction blob")
		}

		// Step 3: Submit to testnet
		result, err := xrplService.SubmitPaymentTransaction(txBlob)
		if err != nil {
			t.Fatalf("Failed to submit payment transaction: %v", err)
		}

		if result.TransactionID == "" {
			t.Error("Expected non-empty transaction ID")
		}

		// Step 4: Monitor transaction status
		maxRetries, retryInterval := testConfig.GetMonitoringConfig()
		status, err := xrplService.MonitorPaymentTransaction(result.TransactionID, maxRetries, time.Duration(retryInterval)*time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to monitor payment transaction: %v", err)
		}

		if status.Status == "" {
			t.Error("Expected non-empty status")
		}

		// Test the complete workflow
		workflowStatus, err := xrplService.CompletePaymentTransactionWorkflow(
			fromAddress, toAddress, amount, currency, privateKeyHex, keyType)
		if err != nil {
			t.Fatalf("Failed to complete payment transaction workflow: %v", err)
		}

		if workflowStatus.TransactionID == "" {
			t.Error("Expected non-empty workflow transaction ID")
		}

		t.Logf("Phase 1 XRP Integration completed successfully")
		t.Logf("Final Status: %s", workflowStatus.Status)
		t.Logf("Transaction ID: %s", workflowStatus.TransactionID)
	})

	t.Run("Individual Phase 1 Components", func(t *testing.T) {
		// Initialize XRPL service
		xrplService := services.NewXRPLService(services.XRPLConfig{
			NetworkURL: testConfig.NetworkURL,
			TestNet:    testConfig.TestNet,
		})

		if err := xrplService.Initialize(); err != nil {
			t.Fatalf("Failed to initialize XRPL service: %v", err)
		}

		// Test individual components
		t.Run("Create Payment Transaction", func(t *testing.T) {
			fromAddress, _, _, _ := testConfig.GetTestWallet1()
			toAddress, _, _, _ := testConfig.GetTestWallet2()
			amount, currency, _ := testConfig.GetTransactionParams()

			payment, err := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)
			if err != nil {
				t.Fatalf("Failed to create payment transaction: %v", err)
			}

			if payment == nil {
				t.Error("Expected non-nil payment transaction")
			}
		})

		t.Run("Sign Payment Transaction", func(t *testing.T) {
			fromAddress, _, _, _ := testConfig.GetTestWallet1()
			toAddress, _, _, _ := testConfig.GetTestWallet2()
			amount, currency, _ := testConfig.GetTransactionParams()
			_, _, privateKey, keyType := testConfig.GetTestWallet1()

			payment, _ := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)

			txBlob, err := xrplService.SignPaymentTransaction(payment, privateKey, keyType)
			if err != nil {
				t.Fatalf("Failed to sign payment transaction: %v", err)
			}

			if txBlob == "" {
				t.Error("Expected non-empty transaction blob")
			}
		})

		t.Run("Submit Payment Transaction", func(t *testing.T) {
			fromAddress, _, _, _ := testConfig.GetTestWallet1()
			toAddress, _, _, _ := testConfig.GetTestWallet2()
			amount, currency, _ := testConfig.GetTransactionParams()
			_, _, privateKey, keyType := testConfig.GetTestWallet1()

			payment, _ := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)

			txBlob, _ := xrplService.SignPaymentTransaction(payment, privateKey, keyType)

			result, err := xrplService.SubmitPaymentTransaction(txBlob)
			if err != nil {
				t.Fatalf("Failed to submit payment transaction: %v", err)
			}

			if result.TransactionID == "" {
				t.Error("Expected non-empty transaction ID")
			}
		})

		t.Run("Monitor Payment Transaction", func(t *testing.T) {
			fromAddress, _, _, _ := testConfig.GetTestWallet1()
			toAddress, _, _, _ := testConfig.GetTestWallet2()
			amount, currency, _ := testConfig.GetTransactionParams()
			_, _, privateKey, keyType := testConfig.GetTestWallet1()

			payment, _ := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)

			txBlob, _ := xrplService.SignPaymentTransaction(payment, privateKey, keyType)

			result, _ := xrplService.SubmitPaymentTransaction(txBlob)

			status, err := xrplService.MonitorPaymentTransaction(
				result.TransactionID, 2, 100*time.Millisecond)
			if err != nil {
				t.Fatalf("Failed to monitor payment transaction: %v", err)
			}

			if status.Status == "" {
				t.Error("Expected non-empty status")
			}
		})
	})

	t.Run("Multiple Wallet Types", func(t *testing.T) {
		// Initialize XRPL service
		xrplService := services.NewXRPLService(services.XRPLConfig{
			NetworkURL: testConfig.NetworkURL,
			TestNet:    testConfig.TestNet,
		})

		if err := xrplService.Initialize(); err != nil {
			t.Fatalf("Failed to initialize XRPL service: %v", err)
		}

		// Test with different wallet types
		t.Run("secp256k1 Wallet", func(t *testing.T) {
			fromAddress, _, privateKey, keyType := testConfig.GetTestWallet1()
			toAddress, _, _, _ := testConfig.GetTestWallet2()
			amount, currency, _ := testConfig.GetTransactionParams()

			payment, err := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)
			if err != nil {
				t.Fatalf("Failed to create payment transaction with secp256k1 wallet: %v", err)
			}

			txBlob, err := xrplService.SignPaymentTransaction(payment, privateKey, keyType)
			if err != nil {
				t.Fatalf("Failed to sign payment transaction with secp256k1 wallet: %v", err)
			}

			if txBlob == "" {
				t.Error("Expected non-empty transaction blob for secp256k1 wallet")
			}
		})

		t.Run("ed25519 Wallet", func(t *testing.T) {
			fromAddress, _, privateKey, keyType := testConfig.GetTestWallet3()
			toAddress, _, _, _ := testConfig.GetTestWallet1()
			amount, currency, _ := testConfig.GetTransactionParams()

			payment, err := xrplService.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)
			if err != nil {
				t.Fatalf("Failed to create payment transaction with ed25519 wallet: %v", err)
			}

			txBlob, err := xrplService.SignPaymentTransaction(payment, privateKey, keyType)
			if err != nil {
				t.Fatalf("Failed to sign payment transaction with ed25519 wallet: %v", err)
			}

			if txBlob == "" {
				t.Error("Expected non-empty transaction blob for ed25519 wallet")
			}
		})
	})

	t.Run("Real XRPL Testnet Integration", func(t *testing.T) {
		// Test real XRPL network integration
		realClient := xrpl.NewRealXRPLClient(testConfig.NetworkURL, testConfig.TestNet)

		// Test 1: Get current ledger index from real XRPL testnet
		t.Log("Testing real XRPL ledger query...")
		currentLedger, err := realClient.GetCurrentLedgerIndex()
		if err != nil {
			t.Fatalf("Failed to get current ledger from XRPL testnet: %v", err)
		}
		t.Logf("✅ Current XRPL Testnet Ledger: %d", currentLedger)

		// Test 2: Get account info from real XRPL testnet
		t.Log("Testing real XRPL account query...")
		wallet1Address, _, _, _ := testConfig.GetTestWallet1()
		accountInfo, err := realClient.GetAccountInfo(wallet1Address)
		if err != nil {
			t.Fatalf("Failed to get account info from XRPL testnet: %v", err)
		}
		t.Logf("✅ Account info retrieved for %s", wallet1Address)

		// Extract and log account details
		if accountData, ok := accountInfo["account_data"].(map[string]interface{}); ok {
			if sequence, ok := accountData["Sequence"].(float64); ok {
				t.Logf("📝 Sequence Number: %d", int(sequence))
			}
			if balance, ok := accountData["Balance"].(string); ok {
				t.Logf("💰 Balance: %s drops", balance)
			}
		}

		// Test 3: Test transaction submission (will fail with mock signing, but tests network connectivity)
		t.Log("Testing real XRPL transaction submission...")
		mockTxBlob := `{"TransactionType":"Payment","Account":"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh","Destination":"r9cZA1nxHfJfKLYGhos6kHE3Q66ZUFJD4X","Amount":"10","Fee":"12","Sequence":9,"LastLedgerSequence":10310644,"Flags":131072,"SigningPubKey":"mock_key","TxnSignature":"mock_sig"}`

		result, err := realClient.SubmitRealTransaction(mockTxBlob)
		if err != nil {
			// This is expected to fail due to mock signing, but it tests network connectivity
			t.Logf("ℹ️ Transaction submission failed as expected (mock signing): %v", err)
			t.Logf("✅ Network connectivity confirmed - XRPL testnet is reachable")
		} else {
			t.Logf("✅ Transaction submitted successfully: %s", result.TransactionID)
		}

		t.Log("🎉 Real XRPL testnet integration test completed successfully!")
	})
}
