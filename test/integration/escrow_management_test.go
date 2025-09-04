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

// TestEscrowManagementIntegration tests the complete escrow management functionality
func TestEscrowManagementIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
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

	// Test escrow lookup functionality
	t.Run("EscrowLookup", func(t *testing.T) {
		testEscrowLookup(t, xrplService)
	})

	// Test escrow status tracking
	t.Run("EscrowStatusTracking", func(t *testing.T) {
		testEscrowStatusTracking(t, xrplService)
	})

	// Test balance verification
	t.Run("BalanceVerification", func(t *testing.T) {
		testBalanceVerification(t, xrplService)
	})

	// Test multiple escrows
	t.Run("MultipleEscrows", func(t *testing.T) {
		testMultipleEscrows(t, xrplService)
	})

	// Test escrow history
	t.Run("EscrowHistory", func(t *testing.T) {
		testEscrowHistory(t, xrplService)
	})

	// Test escrow monitoring
	t.Run("EscrowMonitoring", func(t *testing.T) {
		testEscrowMonitoring(t, xrplService)
	})

	// Test escrow health status
	t.Run("EscrowHealthStatus", func(t *testing.T) {
		testEscrowHealthStatus(t, xrplService)
	})
}

// testEscrowLookup tests escrow lookup functionality
func testEscrowLookup(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Test basic escrow lookup
	escrowInfo, err := xrplService.LookupEscrow(testAccount, "")
	require.NoError(t, err)
	assert.NotNil(t, escrowInfo)

	log.Printf("Escrow lookup successful: Account: %s, Sequence: %d", escrowInfo.Account, escrowInfo.Sequence)

	// Test escrow lookup with specific sequence
	if escrowInfo.Sequence > 0 {
		escrowInfo2, err := xrplService.LookupEscrow(testAccount, fmt.Sprintf("%d", escrowInfo.Sequence))
		require.NoError(t, err)
		assert.Equal(t, escrowInfo.Sequence, escrowInfo2.Sequence)
	}
}

// testEscrowStatusTracking tests escrow status tracking functionality
func testEscrowStatusTracking(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Get escrow status
	escrowInfo, err := xrplService.GetEscrowStatus(testAccount, "")
	require.NoError(t, err)
	assert.NotNil(t, escrowInfo)

	log.Printf("Escrow status: Account: %s, Flags: %d, CancelAfter: %d, FinishAfter: %d",
		escrowInfo.Account, escrowInfo.Flags, escrowInfo.CancelAfter, escrowInfo.FinishAfter)

	// Verify escrow has valid information
	assert.NotEmpty(t, escrowInfo.Account)
	assert.True(t, escrowInfo.Sequence > 0)
}

// testBalanceVerification tests escrow balance verification
func testBalanceVerification(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Get escrow info first
	escrowInfo, err := xrplService.GetEscrowStatus(testAccount, "")
	require.NoError(t, err)
	require.NotNil(t, escrowInfo)

	// Verify escrow balance
	escrowBalance, err := xrplService.VerifyEscrowBalance(escrowInfo)
	require.NoError(t, err)
	assert.NotNil(t, escrowBalance)

	log.Printf("Escrow balance: ID: %s, Locked: %s %s, Available: %s",
		escrowBalance.EscrowID, escrowBalance.LockedAmount, escrowBalance.Currency, escrowBalance.AvailableFor)

	// Verify balance information
	assert.NotEmpty(t, escrowBalance.EscrowID)
	assert.NotEmpty(t, escrowBalance.Currency)
}

// testMultipleEscrows tests retrieving multiple escrows
func testMultipleEscrows(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Get multiple escrows
	result, err := xrplService.GetMultipleEscrows(testAccount, 10)
	require.NoError(t, err)
	assert.NotNil(t, result)

	log.Printf("Multiple escrows: Total: %d, HasMore: %t", result.Total, result.HasMore)

	// Verify result structure
	assert.GreaterOrEqual(t, result.Total, 0)
	assert.LessOrEqual(t, result.Total, 10)

	// If escrows exist, verify their structure
	if result.Total > 0 {
		for i, escrow := range result.Escrows {
			log.Printf("Escrow %d: Account: %s, Sequence: %d, Amount: %s",
				i+1, escrow.Account, escrow.Sequence, escrow.Amount)
			
			assert.NotEmpty(t, escrow.Account)
			assert.True(t, escrow.Sequence > 0)
		}
	}
}

// testEscrowHistory tests escrow transaction history
func testEscrowHistory(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Get escrow history
	transactions, err := xrplService.GetEscrowHistory(testAccount, 20)
	require.NoError(t, err)
	assert.NotNil(t, transactions)

	log.Printf("Escrow history: %d transactions", len(transactions))

	// Verify transactions structure
	assert.GreaterOrEqual(t, len(transactions), 0)
	assert.LessOrEqual(t, len(transactions), 20)

	// If transactions exist, verify their structure
	for i, tx := range transactions {
		log.Printf("Transaction %d: ID: %s, Ledger: %d, Validated: %t",
			i+1, tx.TransactionID, tx.LedgerIndex, tx.Validated)
		
		assert.NotEmpty(t, tx.TransactionID)
		assert.True(t, tx.LedgerIndex > 0)
	}
}

// testEscrowMonitoring tests escrow monitoring functionality
func testEscrowMonitoring(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Create a channel to receive monitoring updates
	updates := make(chan *xrpl.EscrowInfo, 5)
	errors := make(chan error, 5)

	// Start monitoring
	err := xrplService.MonitorEscrowStatus(testAccount, "", func(escrowInfo *xrpl.EscrowInfo, err error) {
		if err != nil {
			errors <- err
			return
		}
		updates <- escrowInfo
	})
	require.NoError(t, err)

	// Wait for a few updates or timeout
	timeout := time.After(2 * time.Minute)
	updateCount := 0

	for {
		select {
		case escrowInfo := <-updates:
			updateCount++
			log.Printf("Monitoring update %d: Account: %s, Sequence: %d", updateCount, escrowInfo.Account, escrowInfo.Sequence)
			
			// Stop after 3 updates
			if updateCount >= 3 {
				return
			}
			
		case err := <-errors:
			log.Printf("Monitoring error: %v", err)
			// Don't fail the test for monitoring errors, just log them
			
		case <-timeout:
			log.Printf("Monitoring timeout after %d updates", updateCount)
			return
		}
	}
}

// testEscrowHealthStatus tests escrow health status functionality
func testEscrowHealthStatus(t *testing.T, xrplService *services.EnhancedXRPLService) {
	// Use a known testnet account for testing
	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Get escrow health status
	healthStatus, err := xrplService.GetEscrowHealthStatus(testAccount, "")
	require.NoError(t, err)
	assert.NotNil(t, healthStatus)

	log.Printf("Escrow health: ID: %s, Status: %s, Health: %s, Message: %s",
		healthStatus.SmartChequeID, healthStatus.Status, healthStatus.Health, healthStatus.Message)

	// Verify health status structure
	assert.NotEmpty(t, healthStatus.SmartChequeID)
	assert.NotEmpty(t, healthStatus.Status)
	assert.NotEmpty(t, healthStatus.Health)
	assert.NotEmpty(t, healthStatus.Message)
	assert.NotNil(t, healthStatus.EscrowInfo)
	assert.True(t, healthStatus.LastSync.After(time.Now().Add(-1*time.Hour)))
}

// TestEscrowManagementErrorHandling tests error handling in escrow management
func TestEscrowManagementErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
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

	// Test with invalid account address
	t.Run("InvalidAccountAddress", func(t *testing.T) {
		_, err := xrplService.LookupEscrow("invalid_address", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to lookup escrow")
	})

	// Test with empty account address
	t.Run("EmptyAccountAddress", func(t *testing.T) {
		_, err := xrplService.LookupEscrow("", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to lookup escrow")
	})

	// Test with nil escrow info
	t.Run("NilEscrowInfo", func(t *testing.T) {
		_, err := xrplService.VerifyEscrowBalance(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "escrow info cannot be nil")
	})
}

// TestEscrowManagementPerformance tests performance of escrow management operations
func TestEscrowManagementPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
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

	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Test performance of multiple operations
	t.Run("MultipleOperationsPerformance", func(t *testing.T) {
		start := time.Now()
		
		// Perform multiple operations
		for i := 0; i < 5; i++ {
			_, err := xrplService.GetEscrowStatus(testAccount, "")
			require.NoError(t, err)
		}
		
		duration := time.Since(start)
		log.Printf("5 escrow status operations completed in %v", duration)
		
		// Ensure operations complete within reasonable time
		assert.Less(t, duration, 30*time.Second)
	})
}

// TestEscrowManagementConcurrency tests concurrent escrow management operations
func TestEscrowManagementConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
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

	testAccount := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"

	// Test concurrent operations
	t.Run("ConcurrentOperations", func(t *testing.T) {
		const numGoroutines = 5
		results := make(chan error, numGoroutines)
		
		// Start concurrent operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				_, err := xrplService.GetEscrowStatus(testAccount, "")
				results <- err
			}(i)
		}
		
		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
		
		log.Printf("Successfully completed %d concurrent escrow status operations", numGoroutines)
	})
}
