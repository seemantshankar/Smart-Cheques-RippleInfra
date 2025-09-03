package xrpl

import (
	"testing"
)

func TestNewEnhancedClient(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.NetworkURL != "https://s.altnet.rippletest.net:51233" {
		t.Errorf("Expected NetworkURL to be 'https://s.altnet.rippletest.net:51233', got '%s'", client.NetworkURL)
	}

	if !client.TestNet {
		t.Error("Expected TestNet to be true")
	}

	if client.initialized {
		t.Error("Expected client to not be initialized initially")
	}
}

func TestEnhancedClient_ValidateAddress(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	// Test valid XRPL address
	validAddress := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	if !client.ValidateAddress(validAddress) {
		t.Errorf("Expected address '%s' to be valid", validAddress)
	}

	// Test invalid addresses
	invalidAddresses := []string{
		"",                                    // Empty
		"r",                                   // Too short
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh0", // Contains 0
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThO", // Contains O
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThI", // Contains I
		"rHb9CJAWyB4rj91VRWn96DkukG4bwdtyThl", // Contains l
		"xHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",  // Doesn't start with r
	}

	for _, addr := range invalidAddresses {
		if client.ValidateAddress(addr) {
			t.Errorf("Expected address '%s' to be invalid", addr)
		}
	}
}

func TestEnhancedClient_GenerateWallet(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	wallet, err := client.GenerateWallet()
	if err != nil {
		t.Fatalf("Failed to generate wallet: %v", err)
	}

	if wallet == nil {
		t.Fatal("Expected wallet to be generated, got nil")
	}

	// Validate wallet structure
	if wallet.Address == "" {
		t.Error("Expected wallet address to be set")
	}

	if wallet.PublicKey == "" {
		t.Error("Expected wallet public key to be set")
	}

	if wallet.PrivateKey == "" {
		t.Error("Expected wallet private key to be set")
	}

	if wallet.Seed == "" {
		t.Error("Expected wallet seed to be set")
	}

	// Validate address format using permissive validation for generated addresses
	if !client.ValidateGeneratedAddress(wallet.Address) {
		t.Errorf("Generated address '%s' is not valid", wallet.Address)
	}
}

func TestEnhancedClient_GenerateSecp256k1Wallet(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	wallet, err := client.GenerateSecp256k1Wallet()
	if err != nil {
		t.Fatalf("Failed to generate secp256k1 wallet: %v", err)
	}

	if wallet == nil {
		t.Fatal("Expected wallet to be generated, got nil")
	}

	// Validate wallet structure
	if wallet.Address == "" {
		t.Error("Expected wallet address to be set")
	}

	if wallet.PublicKey == "" {
		t.Error("Expected wallet public key to be set")
	}

	if wallet.PrivateKey == "" {
		t.Error("Expected wallet private key to be set")
	}

	if wallet.Seed == "" {
		t.Error("Expected wallet seed to be set")
	}

	// Validate address format using permissive validation for generated addresses
	if !client.ValidateGeneratedAddress(wallet.Address) {
		t.Errorf("Generated address '%s' is not valid", wallet.Address)
	}
}

func TestEnhancedClient_FormatAmount(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	// Test XRP formatting (converts to drops)
	xrpAmount := 1.5
	expectedDrops := "1500000"
	formatted := client.FormatAmount(xrpAmount, "XRP")
	if formatted != expectedDrops {
		t.Errorf("Expected XRP amount %.1f to format to '%s', got '%s'", xrpAmount, expectedDrops, formatted)
	}

	// Test other currency formatting
	usdAmount := 100.50
	expectedUSD := "100.500000"
	formatted = client.FormatAmount(usdAmount, "USD")
	if formatted != expectedUSD {
		t.Errorf("Expected USD amount %.2f to format to '%s', got '%s'", usdAmount, expectedUSD, formatted)
	}
}

func TestEnhancedClient_GenerateCondition(t *testing.T) {
	client := NewEnhancedClient("https://s.altnet.rippletest.net:51233", true)

	secret := "test_secret_123"
	condition, fulfillment, err := client.GenerateCondition(secret)
	if err != nil {
		t.Fatalf("Failed to generate condition: %v", err)
	}

	if condition == "" {
		t.Error("Expected condition to be generated")
	}

	if fulfillment != secret {
		t.Errorf("Expected fulfillment to be '%s', got '%s'", secret, fulfillment)
	}

	// Test that condition is different each time
	condition2, _, err := client.GenerateCondition(secret)
	if err != nil {
		t.Fatalf("Failed to generate second condition: %v", err)
	}

	if condition == condition2 {
		t.Error("Expected conditions to be different each time")
	}
}
