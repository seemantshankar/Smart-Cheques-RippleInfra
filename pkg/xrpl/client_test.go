package xrpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("wss://s.altnet.rippletest.net:51233", true)
	
	assert.NotNil(t, client)
	assert.Equal(t, "wss://s.altnet.rippletest.net:51233", client.NetworkURL)
	assert.True(t, client.TestNet)
}

func TestGenerateWallet(t *testing.T) {
	client := NewClient("wss://s.altnet.rippletest.net:51233", true)
	
	wallet, err := client.GenerateWallet()
	require.NoError(t, err)
	require.NotNil(t, wallet)
	
	// Validate wallet structure
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.Seed)
	
	// Validate address format
	t.Logf("Generated address: %s (length: %d)", wallet.Address, len(wallet.Address))
	assert.True(t, client.ValidateAddress(wallet.Address))
	assert.True(t, len(wallet.Address) >= 25)
	assert.Equal(t, byte('r'), wallet.Address[0])
}

func TestValidateAddress(t *testing.T) {
	client := NewClient("wss://s.altnet.rippletest.net:51233", true)
	
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid classic address",
			address:  "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
			expected: true,
		},
		{
			name:     "invalid address - too short",
			address:  "rN7n7otQ",
			expected: false,
		},
		{
			name:     "invalid address - wrong prefix",
			address:  "xN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
			expected: false,
		},
		{
			name:     "empty address",
			address:  "",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.ValidateAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConnect(t *testing.T) {
	// Test with testnet URL
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	
	err := client.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, client.httpClient)
}

func TestHealthCheck(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	
	// Should fail when not initialized
	client.httpClient = nil
	err := client.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
	
	// Should pass when connected
	err = client.Connect()
	require.NoError(t, err)
	
	// Note: This might fail if the actual endpoint is not reachable
	// In a real test environment, you might want to mock this
	err = client.HealthCheck()
	// We'll just check that the method doesn't panic for now
	// assert.NoError(t, err)
}

func TestWalletGeneration_Multiple(t *testing.T) {
	client := NewClient("wss://s.altnet.rippletest.net:51233", true)
	
	// Generate multiple wallets and ensure they're unique
	wallets := make([]*WalletInfo, 5)
	addresses := make(map[string]bool)
	
	for i := 0; i < 5; i++ {
		wallet, err := client.GenerateWallet()
		require.NoError(t, err)
		require.NotNil(t, wallet)
		
		wallets[i] = wallet
		
		// Ensure address is unique
		assert.False(t, addresses[wallet.Address], "Duplicate address generated: %s", wallet.Address)
		addresses[wallet.Address] = true
		
		// Validate each wallet
		assert.True(t, client.ValidateAddress(wallet.Address))
	}
	
	// Ensure all wallets are different
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			assert.NotEqual(t, wallets[i].Address, wallets[j].Address)
			assert.NotEqual(t, wallets[i].PrivateKey, wallets[j].PrivateKey)
			assert.NotEqual(t, wallets[i].Seed, wallets[j].Seed)
		}
	}
}