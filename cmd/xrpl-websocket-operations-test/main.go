package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

type TestResults struct {
	WebSocketConnected bool
	LedgerSubscribed   bool
	WalletCreated      bool
	BasicTransaction   bool
	EscrowCreated      bool
	EscrowFinalization bool
	EscrowCancellation bool
	EscrowFinish       bool
}

func main() {
	// Configuration from env.local
	testnetURL := "https://s.altnet.rippletest.net:51234"

	// Test wallet credentials from env.local
	payerAddress := "rHTwC9fofmbDCnjzwEzVRiFwRKjp9E7j1u"
	payerSecret := "sEdT6TjAdXVSxcGD1Zu3nBAc9cmBkz8"
	payeeAddress := "rwAYkUefDLHTv3nskfjACz3L96tcrYTbM4"

	log.Println("=== XRPL WebSocket Operations Comprehensive Test ===")
	log.Printf("Using testnet URL: %s", testnetURL)
	log.Printf("Payer Address: %s", payerAddress)
	log.Printf("Payee Address: %s", payeeAddress)

	// Create enhanced XRPL client
	client := xrpl.NewEnhancedClient(testnetURL, true)

	// Track results and failures
	results := &TestResults{}
	failedSteps := []string{}
	failedMap := map[string]bool{}
	recordFail := func(name string) {
		if !failedMap[name] {
			failedMap[name] = true
			failedSteps = append(failedSteps, name)
		}
	}

	// Connect to XRPL network
	log.Println("\n1. Connecting to XRPL testnet...")
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// Check if WebSocket connection is available
	if !client.IsWebSocketConnected() {
		log.Fatalf("❌ WebSocket not available, cannot proceed with WebSocket tests")
	}

	log.Println("✅ WebSocket connection established successfully!")
	results.WebSocketConnected = true

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Subscribe to ledger stream for real-time updates (limited output)
	log.Println("\n2. Subscribing to ledger stream...")
	ledgerUpdateCount := 0
	ledgerSubID, err := client.SubscribeToLedgerStream(func(msg *xrpl.StreamMessage) error {
		// Only show first few ledger updates to avoid spam
		if ledgerUpdateCount < 3 {
			if msg.LedgerIndex > 0 {
				log.Printf("📊 Ledger #%d - Hash: %s", msg.LedgerIndex, msg.LedgerHash)
			}
			ledgerUpdateCount++
		} else if ledgerUpdateCount == 3 {
			log.Printf("📊 Ledger streaming active (further updates suppressed)...")
			ledgerUpdateCount++
		}
		return nil
	})
	if err != nil {
		log.Printf("⚠️  Failed to subscribe to ledger stream: %v", err)
	} else {
		log.Printf("✅ Subscribed to ledger stream with ID: %s", ledgerSubID)
		results.LedgerSubscribed = true
	}

	// Test 1: Wallet Creation
	log.Println("\n3. Testing Wallet Creation...")
	log.Println("   Creating and funding new test wallet via Testnet faucet (HTTP) and WS monitoring...")

	// Add timeout for wallet creation (funding + validation can take time)
	walletCtx, walletCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer walletCancel()

	walletChan := make(chan *xrpl.WalletInfo, 1)
	walletErrChan := make(chan error, 1)

	go func() {
		// Use official faucet (HTTP) to fund and WS/JSON-RPC to confirm
		wallet, err := client.CreateAndFundWalletWithFaucet(60 * time.Second)
		if err != nil {
			walletErrChan <- err
			return
		}
		walletChan <- wallet
	}()

	select {
	case newWallet := <-walletChan:
		log.Printf("✅ New wallet created and funded successfully!")
		log.Printf("   Address: %s", newWallet.Address)
		log.Printf("   Public Key: %s", newWallet.PublicKey)
		log.Printf("   Private Key: %s", newWallet.PrivateKey)
		log.Printf("   Seed: %s", newWallet.Seed)
		log.Println("   ✅ Wallet creation test completed!")
		results.WalletCreated = true
	case err := <-walletErrChan:
		log.Printf("❌ Wallet creation failed: %v", err)
		log.Println("   ⚠️  Continuing with next test...")
		recordFail("Wallet creation")
	case <-walletCtx.Done():
		log.Printf("❌ Wallet creation timed out after 120 seconds")
		log.Println("   ⚠️  Continuing with next test...")
		recordFail("Wallet creation")
	}

	// Test 2: Basic Transaction between test wallets
	log.Println("\n4. Testing Basic Transaction between test wallets...")

	// Check account balances
	log.Println("   Checking account balances...")
	payerBalance, err := client.GetAccountBalance(payerAddress)
	if err != nil {
		log.Printf("❌ Failed to get payer balance: %v", err)
		log.Println("   ⚠️  Skipping transaction test...")
	} else {
		log.Printf("   Payer balance: %s drops", payerBalance)

		payeeBalance, err := client.GetAccountBalance(payeeAddress)
		if err != nil {
			log.Printf("❌ Failed to get payee balance: %v", err)
			log.Println("   ⚠️  Skipping transaction test...")
			recordFail("Basic transaction")
		} else {
			log.Printf("   Payee balance: %s drops", payeeBalance)

			// Create and submit a small payment transaction
			log.Println("   Creating payment transaction...")
			paymentAmount := "1000000" // 1 XRP in drops

			// Get current sequence numbers
			payerAccountData, err := client.GetAccountData(payerAddress)
			if err != nil {
				log.Printf("❌ Failed to get payer account data: %v", err)
			} else {
				log.Printf("   Payer sequence: %d", payerAccountData.Sequence)

				// Submit payment transaction
				log.Println("   Submitting payment transaction...")
				paymentResult, err := client.CreatePaymentTransaction(payerAddress, payeeAddress, paymentAmount, payerSecret)
				if err != nil {
					log.Printf("❌ Payment transaction failed: %v", err)
					recordFail("Basic transaction")
				} else {
					log.Printf("   Transaction ID: %s", paymentResult.TransactionID)
					log.Printf("   Result Code: %s", paymentResult.ResultCode)
					log.Printf("   Ledger Index: %d", paymentResult.LedgerIndex)
					if paymentResult.ResultCode == "tesSUCCESS" {
						log.Printf("✅ Payment transaction submitted successfully!")
						results.BasicTransaction = true
					} else {
						log.Printf("❌ Payment transaction returned non-success result: %s", paymentResult.ResultCode)
						recordFail("Basic transaction")
					}
				}

				// Wait a bit for transaction to be processed
				log.Println("   Waiting for transaction to be processed...")
				time.Sleep(5 * time.Second)

				// Check updated balances
				log.Println("   Checking updated balances...")
				newPayerBalance, err := client.GetAccountBalance(payerAddress)
				if err != nil {
					log.Printf("❌ Failed to get updated payer balance: %v", err)
				} else {
					log.Printf("   Updated payer balance: %s drops", newPayerBalance)
				}

				newPayeeBalance, err := client.GetAccountBalance(payeeAddress)
				if err != nil {
					log.Printf("❌ Failed to get updated payee balance: %v", err)
				} else {
					log.Printf("   Updated payee balance: %s drops", newPayeeBalance)
				}

				log.Println("   ✅ Basic transaction test completed!")
			}
		}
	}

	// Test 3: Escrow Creation
	log.Println("\n5. Testing Escrow Creation...")

	// Generate escrow condition and fulfillment
	log.Println("   Generating escrow condition...")
	escrowSecret := "test_escrow_secret_123"
	condition, fulfillment, err := client.GenerateCondition(escrowSecret)
	if err != nil {
		log.Printf("❌ Failed to generate escrow condition: %v", err)
		log.Println("   ⚠️  Skipping escrow tests...")
	} else {
		log.Printf("✅ Escrow condition generated!")
		log.Printf("   Condition: %s", condition)
		log.Printf("   Fulfillment: %s", fulfillment)

		// Create escrow
		log.Println("   Creating escrow...")
		escrowAmount := "500000" // 0.5 XRP in drops

		// Calculate finish and cancel times (1 hour from now)
		finishAfter := uint32(time.Now().Add(time.Hour).Unix())
		cancelAfter := uint32(time.Now().Add(2 * time.Hour).Unix())

		escrowCreate := &xrpl.EscrowCreate{
			Account:     payerAddress,
			Amount:      escrowAmount,
			Destination: payeeAddress,
			FinishAfter: finishAfter,
			CancelAfter: cancelAfter,
			Condition:   condition,
		}

		escrowResult, err := client.CreateEscrow(escrowCreate, payerSecret)
		if err != nil {
			log.Printf("❌ Escrow creation failed: %v", err)
			recordFail("Escrow creation")
		} else {
			log.Printf("   Transaction ID: %s", escrowResult.TransactionID)
			log.Printf("   Result Code: %s", escrowResult.ResultCode)
			log.Printf("   Ledger Index: %d", escrowResult.LedgerIndex)
			if escrowResult.ResultCode == "tesSUCCESS" {
				log.Printf("✅ Escrow created successfully!")
				results.EscrowCreated = true
			} else {
				log.Printf("❌ Escrow creation returned non-success result: %s", escrowResult.ResultCode)
				recordFail("Escrow creation")
			}
		}

		// Wait for escrow to be processed
		log.Println("   Waiting for escrow to be processed...")
		time.Sleep(5 * time.Second)

		// Lookup escrow
		log.Println("   Looking up created escrow...")
		payerAccountData, err := client.GetAccountData(payerAddress)
		if err != nil {
			log.Printf("❌ Failed to get payer account data: %v", err)
		} else {
			escrowInfo, err := client.LookupEscrow(payerAddress, strconv.FormatUint(uint64(payerAccountData.Sequence), 10))
			if err != nil {
				log.Printf("❌ Escrow lookup failed: %v", err)
			} else {
				log.Printf("✅ Escrow lookup successful!")
				log.Printf("   Account: %s", escrowInfo.Account)
				log.Printf("   Amount: %s", escrowInfo.Amount)
				log.Printf("   Destination: %s", escrowInfo.Destination)
				log.Printf("   Finish After: %d", escrowInfo.FinishAfter)
				log.Printf("   Cancel After: %d", escrowInfo.CancelAfter)
				log.Printf("   Condition: %s", escrowInfo.Condition)
			}
		}

		log.Println("   ✅ Escrow creation test completed!")
	}

	// Test 4: Escrow Finalization
	log.Println("\n6. Testing Escrow Finalization...")

	// For testing purposes, we'll create a new escrow with shorter times
	log.Println("   Creating escrow for finalization test...")
	finishEscrowAmount := "300000" // 0.3 XRP in drops

	// Use shorter times for testing
	finishFinishAfter := uint32(time.Now().Add(30 * time.Second).Unix())
	finishCancelAfter := uint32(time.Now().Add(1 * time.Minute).Unix())

	finishEscrowCreate := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Amount:      finishEscrowAmount,
		Destination: payeeAddress,
		FinishAfter: finishFinishAfter,
		CancelAfter: finishCancelAfter,
		Condition:   condition,
	}

	finishEscrowResult, err := client.CreateEscrow(finishEscrowCreate, payerSecret)
	if err != nil {
		log.Printf("❌ Finish escrow creation failed: %v", err)
		log.Println("   ⚠️  Skipping finalization test...")
		recordFail("Escrow finalization")
	} else {
		log.Printf("✅ Finish escrow created successfully!")
		log.Printf("   Transaction ID: %s", finishEscrowResult.TransactionID)
		log.Printf("   Result Code: %s", finishEscrowResult.ResultCode)
		log.Printf("   Ledger Index: %d", finishEscrowResult.LedgerIndex)

		// Wait for escrow to be processed
		log.Println("   Waiting for finish escrow to be processed...")
		time.Sleep(5 * time.Second)

		// Wait for finish time to be reached
		log.Println("   Waiting for escrow finish time...")
		time.Sleep(35 * time.Second)

		// Finish escrow
		log.Println("   Finalizing escrow...")
		payerAccountData, err := client.GetAccountData(payerAddress)
		if err != nil {
			log.Printf("❌ Failed to get payer account data: %v", err)
		} else {
			escrowFinish := &xrpl.EscrowFinish{
				Owner:         payerAddress,
				OfferSequence: payerAccountData.Sequence,
				Condition:     condition,
				Fulfillment:   fulfillment,
			}

			finishResult, err := client.FinishEscrow(escrowFinish, payerSecret)
			if err != nil {
				log.Printf("❌ Escrow finalization failed: %v", err)
				recordFail("Escrow finalization")
			} else {
				log.Printf("   Transaction ID: %s", finishResult.TransactionID)
				log.Printf("   Result Code: %s", finishResult.ResultCode)
				log.Printf("   Ledger Index: %d", finishResult.LedgerIndex)
				if finishResult.ResultCode == "tesSUCCESS" {
					log.Printf("✅ Escrow finalized successfully!")
					results.EscrowFinalization = true
				} else {
					log.Printf("❌ Escrow finalization returned non-success result: %s", finishResult.ResultCode)
					recordFail("Escrow finalization")
				}
			}

			// Wait for finalization to be processed
			log.Println("   Waiting for finalization to be processed...")
			time.Sleep(5 * time.Second)
		}

		log.Println("   ✅ Escrow finalization test completed!")
	}

	// Test 5: Escrow Cancellation
	log.Println("\n7. Testing Escrow Cancellation...")

	// Create another escrow for cancellation test
	log.Println("   Creating escrow for cancellation test...")
	cancelEscrowAmount := "200000" // 0.2 XRP in drops

	// Use shorter times for testing
	cancelFinishAfter := uint32(time.Now().Add(30 * time.Second).Unix())
	cancelCancelAfter := uint32(time.Now().Add(1 * time.Minute).Unix())

	cancelEscrowCreate := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Amount:      cancelEscrowAmount,
		Destination: payeeAddress,
		FinishAfter: cancelFinishAfter,
		CancelAfter: cancelCancelAfter,
		Condition:   condition,
	}

	cancelEscrowResult, err := client.CreateEscrow(cancelEscrowCreate, payerSecret)
	if err != nil {
		log.Printf("❌ Cancel escrow creation failed: %v", err)
		log.Println("   ⚠️  Skipping cancellation test...")
		recordFail("Escrow cancellation")
	} else {
		log.Printf("✅ Cancel escrow created successfully!")
		log.Printf("   Transaction ID: %s", cancelEscrowResult.TransactionID)
		log.Printf("   Result Code: %s", cancelEscrowResult.ResultCode)
		log.Printf("   Ledger Index: %d", cancelEscrowResult.LedgerIndex)

		// Wait for escrow to be processed
		log.Println("   Waiting for cancel escrow to be processed...")
		time.Sleep(5 * time.Second)

		// Cancel escrow
		log.Println("   Cancelling escrow...")
		payerAccountData, err := client.GetAccountData(payerAddress)
		if err != nil {
			log.Printf("❌ Failed to get payer account data: %v", err)
		} else {
			escrowCancel := &xrpl.EscrowCancel{
				Owner:         payerAddress,
				OfferSequence: payerAccountData.Sequence,
			}

			cancelResult, err := client.CancelEscrow(escrowCancel, payerSecret)
			if err != nil {
				log.Printf("❌ Escrow cancellation failed: %v", err)
				recordFail("Escrow cancellation")
			} else {
				log.Printf("   Transaction ID: %s", cancelResult.TransactionID)
				log.Printf("   Result Code: %s", cancelResult.ResultCode)
				log.Printf("   Ledger Index: %d", cancelResult.LedgerIndex)
				if cancelResult.ResultCode == "tesSUCCESS" {
					log.Printf("✅ Escrow cancelled successfully!")
					results.EscrowCancellation = true
				} else {
					log.Printf("❌ Escrow cancellation returned non-success result: %s", cancelResult.ResultCode)
					recordFail("Escrow cancellation")
				}
			}

			// Wait for cancellation to be processed
			log.Println("   Waiting for cancellation to be processed...")
			time.Sleep(5 * time.Second)
		}

		log.Println("   ✅ Escrow cancellation test completed!")
	}

	// Test 6: Escrow Finish
	log.Println("\n8. Testing Escrow Finish...")

	// Create final escrow for finish test
	log.Println("   Creating escrow for finish test...")
	finalEscrowAmount := "100000" // 0.1 XRP in drops

	// Use very short times for testing
	finalFinishAfter := uint32(time.Now().Add(10 * time.Second).Unix())
	finalCancelAfter := uint32(time.Now().Add(2 * time.Minute).Unix())

	finalEscrowCreate := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Amount:      finalEscrowAmount,
		Destination: payeeAddress,
		FinishAfter: finalFinishAfter,
		CancelAfter: finalCancelAfter,
		Condition:   condition,
	}

	finalEscrowResult, err := client.CreateEscrow(finalEscrowCreate, payerSecret)
	if err != nil {
		log.Printf("❌ Final escrow creation failed: %v", err)
		log.Println("   ⚠️  Skipping finish test...")
		recordFail("Escrow finish")
	} else {
		log.Printf("✅ Final escrow created successfully!")
		log.Printf("   Transaction ID: %s", finalEscrowResult.TransactionID)
		log.Printf("   Result Code: %s", finalEscrowResult.ResultCode)
		log.Printf("   Ledger Index: %d", finalEscrowResult.LedgerIndex)

		// Wait for escrow to be processed
		log.Println("   Waiting for final escrow to be processed...")
		time.Sleep(5 * time.Second)

		// Wait for finish time to be reached
		log.Println("   Waiting for escrow finish time...")
		time.Sleep(15 * time.Second)

		// Finish escrow
		log.Println("   Finishing escrow...")
		payerAccountData, err := client.GetAccountData(payerAddress)
		if err != nil {
			log.Printf("❌ Failed to get payer account data: %v", err)
		} else {
			finalEscrowFinish := &xrpl.EscrowFinish{
				Owner:         payerAddress,
				OfferSequence: payerAccountData.Sequence,
				Condition:     condition,
				Fulfillment:   fulfillment,
			}

			finalFinishResult, err := client.FinishEscrow(finalEscrowFinish, payerSecret)
			if err != nil {
				log.Printf("❌ Final escrow finish failed: %v", err)
				recordFail("Escrow finish")
			} else {
				log.Printf("   Transaction ID: %s", finalFinishResult.TransactionID)
				log.Printf("   Result Code: %s", finalFinishResult.ResultCode)
				log.Printf("   Ledger Index: %d", finalFinishResult.LedgerIndex)
				if finalFinishResult.ResultCode == "tesSUCCESS" {
					log.Printf("✅ Final escrow finished successfully!")
					results.EscrowFinish = true
				} else {
					log.Printf("❌ Final escrow finish returned non-success result: %s", finalFinishResult.ResultCode)
					recordFail("Escrow finish")
				}
			}

			// Wait for finish to be processed
			log.Println("   Waiting for final finish to be processed...")
			time.Sleep(5 * time.Second)
		}

		log.Println("   ✅ Escrow finish test completed!")
	}

	// Final balance check
	log.Println("\n9. Final Balance Check...")
	finalPayerBalance, err := client.GetAccountBalance(payerAddress)
	if err != nil {
		log.Printf("❌ Failed to get final payer balance: %v", err)
	} else {
		log.Printf("   Final payer balance: %s drops", finalPayerBalance)
	}

	finalPayeeBalance, err := client.GetAccountBalance(payeeAddress)
	if err != nil {
		log.Printf("❌ Failed to get final payee balance: %v", err)
	} else {
		log.Printf("   Final payee balance: %s drops", finalPayeeBalance)
	}

	// Keep the program running to capture real-time updates
	log.Println("\n🔄 Monitoring XRPL operations... Press Ctrl+C to exit")
	log.Println("   (This will run for 30 seconds to complete all operations)")

	// Set a timeout for the test
	testTimer := time.NewTimer(30 * time.Second)

	select {
	case <-ctx.Done():
		log.Println("Context cancelled")
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-testTimer.C:
		log.Println("Test timeout reached")
	}

	// Graceful shutdown
	log.Println("\n🛑 Shutting down...")

	// Unsubscribe from streams
	if ledgerSubID != "" {
		if err := client.UnsubscribeFromStream(ledgerSubID); err != nil {
			log.Printf("⚠️  Failed to unsubscribe from ledger stream: %v", err)
		} else {
			log.Printf("✅ Unsubscribed from ledger stream")
		}
	}

	// Close WebSocket connection
	if err := client.CloseWebSocket(); err != nil {
		log.Printf("⚠️  Failed to close WebSocket: %v", err)
	} else {
		log.Printf("✅ WebSocket connection closed")
	}

	log.Println("✅ Shutdown complete")
	log.Println("\n=== Test Summary ===")
	if results.WebSocketConnected {
		log.Println("✅ WebSocket connection established")
	} else {
		log.Println("❌ WebSocket connection failed")
	}
	if results.WalletCreated {
		log.Println("✅ Wallet creation tested")
	} else {
		log.Println("❌ Wallet creation failed")
	}
	if results.BasicTransaction {
		log.Println("✅ Basic transaction tested")
	} else {
		log.Println("❌ Basic transaction failed")
	}
	if results.EscrowCreated {
		log.Println("✅ Escrow creation tested")
	} else {
		log.Println("❌ Escrow creation failed")
	}
	if results.EscrowFinalization {
		log.Println("✅ Escrow finalization tested")
	} else {
		log.Println("❌ Escrow finalization failed")
	}
	if results.EscrowCancellation {
		log.Println("✅ Escrow cancellation tested")
	} else {
		log.Println("❌ Escrow cancellation failed")
	}
	if results.EscrowFinish {
		log.Println("✅ Escrow finish tested")
	} else {
		log.Println("❌ Escrow finish failed")
	}
	if results.LedgerSubscribed {
		log.Println("✅ Real-time ledger streaming")
	} else {
		log.Println("❌ Real-time ledger streaming failed")
	}
	if len(failedSteps) == 0 {
		log.Println("✅ All operations completed via WebSocket")
	} else {
		log.Printf("❌ %d test(s) failed: %v", len(failedSteps), failedSteps)
	}
	if len(failedSteps) > 0 {
		os.Exit(1)
	}
}
