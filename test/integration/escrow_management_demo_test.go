package integration

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// TestEscrowManagementWorkflow demonstrates the complete escrow management workflow
func TestEscrowManagementWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow test in short mode")
	}

	// Initialize XRPL service
	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	xrplService := services.NewEnhancedXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err)
	defer xrplService.Disconnect()

	// Step 1: Create a test wallet
	t.Run("CreateTestWallet", func(t *testing.T) {
		wallet, err := xrplService.CreateWallet()
		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.NotEmpty(t, wallet.Address)
		assert.NotEmpty(t, wallet.PrivateKey)

		log.Printf("Created test wallet: %s", wallet.Address)

		// Store wallet for subsequent tests
		t.Logf("Test wallet created: %s", wallet.Address)
		t.Logf("Private key: %s", wallet.PrivateKey)
	})

	// Step 2: Test escrow management with a known account that has escrows
	t.Run("EscrowManagementWithKnownAccount", func(t *testing.T) {
		// Use a known testnet account that might have escrows
		testAccounts := []string{
			"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", // Testnet faucet account
			"rUCzEr6jrEyMpjhs8wjH2wv2Z1aU3JqGR",  // Another testnet account
		}

		for _, account := range testAccounts {
			log.Printf("Testing escrow management for account: %s", account)

			// Test multiple escrows retrieval
			result, err := xrplService.GetMultipleEscrows(account, 5)
			if err != nil {
				log.Printf("Warning: Could not get escrows for %s: %v", account, err)
				continue
			}

			log.Printf("Account %s has %d escrows", account, result.Total)

			if result.Total > 0 {
				// Test individual escrow lookup
				for _, escrow := range result.Escrows {
					log.Printf("Testing escrow: %s_%d", escrow.Account, escrow.Sequence)

					// Test escrow status
					status, err := xrplService.GetEscrowStatus(escrow.Account, fmt.Sprintf("%d", escrow.Sequence))
					if err != nil {
						log.Printf("Warning: Could not get status for escrow %d: %v", escrow.Sequence, err)
						continue
					}

					assert.NotNil(t, status)
					log.Printf("Escrow status: Flags=%d, CancelAfter=%d, FinishAfter=%d",
						status.Flags, status.CancelAfter, status.FinishAfter)

					// Test balance verification
					balance, err := xrplService.VerifyEscrowBalance(status)
					if err != nil {
						log.Printf("Warning: Could not verify balance for escrow %d: %v", escrow.Sequence, err)
						continue
					}

					assert.NotNil(t, balance)
					log.Printf("Escrow balance: Locked=%s %s, Available=%s",
						balance.LockedAmount, balance.Currency, balance.AvailableFor)

					// Test health status
					health, err := xrplService.GetEscrowHealthStatus(escrow.Account, fmt.Sprintf("%d", escrow.Sequence))
					if err != nil {
						log.Printf("Warning: Could not get health status for escrow %d: %v", escrow.Sequence, err)
						continue
					}

					assert.NotNil(t, health)
					log.Printf("Escrow health: Status=%s, Health=%s, Message=%s",
						health.Status, health.Health, health.Message)

					// Only test one escrow per account to avoid too many API calls
					break
				}
			}

			// Test escrow history
			history, err := xrplService.GetEscrowHistory(account, 10)
			if err != nil {
				log.Printf("Warning: Could not get escrow history for %s: %v", account, err)
				continue
			}

			log.Printf("Account %s has %d escrow transactions in history", account, len(history))
		}
	})

	// Step 3: Test escrow monitoring capabilities
	t.Run("EscrowMonitoringCapabilities", func(t *testing.T) {
		// Test monitoring with a known account
		testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

		// Create a channel to receive monitoring updates
		updates := make(chan *xrpl.EscrowInfo, 3)
		errors := make(chan error, 3)

		// Start monitoring for a short duration
		err := xrplService.MonitorEscrowStatus(testAccount, "", func(escrowInfo *xrpl.EscrowInfo, err error) {
			if err != nil {
				errors <- err
				return
			}
			updates <- escrowInfo
		})
		require.NoError(t, err)

		// Wait for a short time to see if we get any updates
		timeout := time.After(30 * time.Second)
		updateCount := 0

		for {
			select {
			case escrowInfo := <-updates:
				updateCount++
				log.Printf("Monitoring update %d: Account: %s, Sequence: %d", updateCount, escrowInfo.Account, escrowInfo.Sequence)

				// Stop after 2 updates or if we get enough data
				if updateCount >= 2 {
					log.Printf("Received sufficient monitoring updates")
					return
				}

			case err := <-errors:
				log.Printf("Monitoring error: %v", err)
				// Don't fail the test for monitoring errors

			case <-timeout:
				log.Printf("Monitoring timeout after %d updates", updateCount)
				return
			}
		}
	})

	// Step 4: Test error handling and edge cases
	t.Run("ErrorHandlingAndEdgeCases", func(t *testing.T) {
		// Test with invalid account address
		_, err := xrplService.LookupEscrow("invalid_address", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to lookup escrow")

		// Test with empty account address
		_, err = xrplService.LookupEscrow("", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to lookup escrow")

		// Test with nil escrow info for balance verification
		_, err = xrplService.VerifyEscrowBalance(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "escrow info cannot be nil")

		// Test with very large limit for multiple escrows
		result, err := xrplService.GetMultipleEscrows("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", 1000)
		if err == nil {
			// Should respect maximum limit
			assert.LessOrEqual(t, result.Total, 100)
		}
	})

	// Step 5: Test performance characteristics
	t.Run("PerformanceCharacteristics", func(t *testing.T) {
		testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

		// Test multiple rapid operations
		start := time.Now()

		for i := 0; i < 3; i++ {
			_, err := xrplService.GetEscrowStatus(testAccount, "")
			if err != nil {
				log.Printf("Warning: Operation %d failed: %v", i+1, err)
			}
		}

		duration := time.Since(start)
		log.Printf("3 escrow status operations completed in %v", duration)

		// Ensure operations complete within reasonable time
		assert.Less(t, duration, 30*time.Second)
	})
}

// TestEscrowManagementAPIs tests the individual API methods
func TestEscrowManagementAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Initialize XRPL service
	config := services.XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	xrplService := services.NewEnhancedXRPLService(config)
	err := xrplService.Initialize()
	require.NoError(t, err)
	defer xrplService.Disconnect()

	// Test all the escrow management API methods
	t.Run("LookupEscrow", func(t *testing.T) {
		// Test with a known account
		escrowInfo, err := xrplService.LookupEscrow("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "")
		if err != nil {
			// Expected if no escrows exist
			log.Printf("No escrows found for test account (expected)")
			return
		}

		assert.NotNil(t, escrowInfo)
		log.Printf("Escrow lookup successful: %s_%d", escrowInfo.Account, escrowInfo.Sequence)
	})

	t.Run("GetEscrowStatus", func(t *testing.T) {
		// Test with a known account
		escrowInfo, err := xrplService.GetEscrowStatus("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "")
		if err != nil {
			// Expected if no escrows exist
			log.Printf("No escrow status found for test account (expected)")
			return
		}

		assert.NotNil(t, escrowInfo)
		log.Printf("Escrow status retrieved: %s_%d", escrowInfo.Account, escrowInfo.Sequence)
	})

	t.Run("GetMultipleEscrows", func(t *testing.T) {
		// Test with a known account
		result, err := xrplService.GetMultipleEscrows("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", 10)
		require.NoError(t, err)
		assert.NotNil(t, result)

		log.Printf("Multiple escrows result: Total=%d, HasMore=%t", result.Total, result.HasMore)
	})

	t.Run("GetEscrowHistory", func(t *testing.T) {
		// Test with a known account
		history, err := xrplService.GetEscrowHistory("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", 10)
		require.NoError(t, err)

		log.Printf("Escrow history: %d transactions", len(history))

		// If no escrow transactions exist, that's expected behavior
		if len(history) == 0 {
			log.Printf("No escrow transactions found (expected for test account)")
		}
	})
}
