package services

import (
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// Test that ConsolidatedXRPLService implements XRPLServiceInterface
func TestConsolidatedXRPLServiceImplementsInterface(t *testing.T) {
	// This will cause a compilation error if ConsolidatedXRPLService doesn't implement XRPLServiceInterface
	var _ repository.XRPLServiceInterface = (*ConsolidatedXRPLService)(nil)
}

// Test service initialization
func TestNewConsolidatedXRPLService(t *testing.T) {
	config := ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	service := NewConsolidatedXRPLService(config)
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.client == nil {
		t.Fatal("Expected client to be initialized, got nil")
	}

	if service.initialized {
		t.Fatal("Expected service to be uninitialized initially")
	}
}

// Test utility methods
func TestConsolidatedXRPLServiceUtilityMethods(t *testing.T) {
	config := ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	service := NewConsolidatedXRPLService(config)

	// Test formatAmount
	amountStr := service.formatAmount(100.0, "XRP")
	expectedDrops := "100000000" // 100 XRP = 100,000,000 drops
	if amountStr != expectedDrops {
		t.Errorf("Expected %s, got %s", expectedDrops, amountStr)
	}

	// Test getLedgerTimeOffset
	ledgerOffset := service.getLedgerTimeOffset(24 * time.Hour)
	if ledgerOffset == 0 {
		t.Error("Expected non-zero ledger offset for 24 hours")
	}

	// Test calculateCancelAfter
	milestones := []models.Milestone{
		{
			ID:                 "milestone1",
			EstimatedDuration:  48 * time.Hour,
			Amount:             50.0,
			VerificationMethod: models.VerificationMethodManual,
		},
		{
			ID:                 "milestone2",
			EstimatedDuration:  72 * time.Hour,
			Amount:             50.0,
			VerificationMethod: models.VerificationMethodManual,
		},
	}

	cancelAfter := service.calculateCancelAfter(milestones)
	if cancelAfter == 0 {
		t.Error("Expected non-zero cancel after time")
	}

	// Test with empty milestones
	emptyCancelAfter := service.calculateCancelAfter([]models.Milestone{})
	if emptyCancelAfter == 0 {
		t.Error("Expected non-zero cancel after time for empty milestones")
	}
}

// Test interface method signatures
func TestConsolidatedXRPLServiceMethodSignatures(t *testing.T) {
	config := ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	service := NewConsolidatedXRPLService(config)

	// Test that all required methods exist and have correct signatures
	// This is a compile-time check, so if this compiles, the signatures are correct

	// Test CreateSmartChequeEscrow
	_, _, err := service.CreateSmartChequeEscrow("payer", "payee", 100.0, "XRP", "secret")
	if err == nil {
		t.Error("Expected error for CreateSmartChequeEscrow without private key")
	}

	// Test CompleteSmartChequeMilestone
	_, err = service.CompleteSmartChequeMilestone("payee", "owner", 1, "condition", "fulfillment")
	if err == nil {
		t.Error("Expected error for CompleteSmartChequeMilestone without private key")
	}

	// Test that the service can be used as the interface type
	var _ repository.XRPLServiceInterface = service
}

// Test error handling for uninitialized service
func TestConsolidatedXRPLServiceUninitialized(t *testing.T) {
	config := ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	service := NewConsolidatedXRPLService(config)

	// Test that methods return error when service is not initialized
	_, err := service.CreateWallet()
	if err == nil {
		t.Error("Expected error when service is not initialized")
	}

	_, err = service.GetAccountInfo("test")
	if err == nil {
		t.Error("Expected error when service is not initialized")
	}

	err = service.HealthCheck()
	if err == nil {
		t.Error("Expected error when service is not initialized")
	}
}

// Test milestone processing
func TestConsolidatedXRPLServiceMilestoneProcessing(t *testing.T) {
	config := ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234",
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",
		TestNet:      true,
	}

	service := NewConsolidatedXRPLService(config)

	milestones := []models.Milestone{
		{
			ID:                 "milestone1",
			Description:        "First milestone",
			Amount:             25.0,
			VerificationMethod: models.VerificationMethodManual,
			EstimatedDuration:  24 * time.Hour,
			Status:             models.MilestoneStatusPending,
		},
		{
			ID:                 "milestone2",
			Description:        "Second milestone",
			Amount:             25.0,
			VerificationMethod: models.VerificationMethodOracle,
			EstimatedDuration:  48 * time.Hour,
			Status:             models.MilestoneStatusPending,
			OracleConfig: &models.OracleConfig{
				Type:     "http",
				Endpoint: "https://oracle.example.com/verify",
				Config: map[string]interface{}{
					"timeout": 30,
					"retries": 3,
				},
			},
		},
	}

	// Test calculateCancelAfter with these milestones
	cancelAfter := service.calculateCancelAfter(milestones)
	if cancelAfter == 0 {
		t.Error("Expected non-zero cancel after time")
	}

	// The cancel after should be based on the longest milestone duration (48 hours) + buffer
	expectedMinLedgers := int((48*time.Hour + 24*time.Hour).Seconds() / 3.5)
	if int(cancelAfter) < expectedMinLedgers {
		t.Errorf("Expected cancel after to be at least %d ledgers, got %d", expectedMinLedgers, cancelAfter)
	}
}
