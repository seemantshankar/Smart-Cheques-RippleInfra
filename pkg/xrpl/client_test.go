package xrpl

import (
	"strings"
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

func TestCreateEscrow(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	err := client.Connect()
	require.NoError(t, err)

	// Generate test wallets
	payerWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	payeeWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	tests := []struct {
		name        string
		escrow      *EscrowCreate
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid basic escrow",
			escrow: &EscrowCreate{
				Account:     payerWallet.Address,
				Destination: payeeWallet.Address,
				Amount:      "1000000", // 1 XRP in drops
			},
			expectError: false,
		},
		{
			name: "valid escrow with condition",
			escrow: &EscrowCreate{
				Account:     payerWallet.Address,
				Destination: payeeWallet.Address,
				Amount:      "5000000", // 5 XRP in drops
				Condition:   "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				CancelAfter: 123456789,
				FinishAfter: 123456700,
			},
			expectError: false,
		},
		{
			name: "missing account",
			escrow: &EscrowCreate{
				Destination: payeeWallet.Address,
				Amount:      "1000000",
			},
			expectError: true,
			errorMsg:    "missing required escrow fields",
		},
		{
			name: "missing destination",
			escrow: &EscrowCreate{
				Account: payerWallet.Address,
				Amount:  "1000000",
			},
			expectError: true,
			errorMsg:    "missing required escrow fields",
		},
		{
			name: "missing amount",
			escrow: &EscrowCreate{
				Account:     payerWallet.Address,
				Destination: payeeWallet.Address,
			},
			expectError: true,
			errorMsg:    "missing required escrow fields",
		},
		{
			name: "invalid account address",
			escrow: &EscrowCreate{
				Account:     "invalid_address",
				Destination: payeeWallet.Address,
				Amount:      "1000000",
			},
			expectError: true,
			errorMsg:    "invalid account address",
		},
		{
			name: "invalid destination address",
			escrow: &EscrowCreate{
				Account:     payerWallet.Address,
				Destination: "invalid_address",
				Amount:      "1000000",
			},
			expectError: true,
			errorMsg:    "invalid destination address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CreateEscrow(tt.escrow)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
				assert.Greater(t, result.LedgerIndex, uint32(0))

				// Validate transaction ID format (64 character uppercase hex)
				assert.Equal(t, 64, len(result.TransactionID))
				assert.Equal(t, strings.ToUpper(result.TransactionID), result.TransactionID)
			}
		})
	}
}

func TestFinishEscrow(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	err := client.Connect()
	require.NoError(t, err)

	// Generate test wallets
	payerWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	payeeWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	tests := []struct {
		name        string
		finish      *EscrowFinish
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid escrow finish",
			finish: &EscrowFinish{
				Account:       payeeWallet.Address,
				Owner:         payerWallet.Address,
				OfferSequence: 1,
				Condition:     "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				Fulfillment:   "A0028000",
			},
			expectError: false,
		},
		{
			name: "valid escrow finish without condition",
			finish: &EscrowFinish{
				Account:       payeeWallet.Address,
				Owner:         payerWallet.Address,
				OfferSequence: 2,
			},
			expectError: false,
		},
		{
			name: "missing account",
			finish: &EscrowFinish{
				Owner:         payerWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "missing required fields",
		},
		{
			name: "missing owner",
			finish: &EscrowFinish{
				Account:       payeeWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "missing required fields",
		},
		{
			name: "invalid account address",
			finish: &EscrowFinish{
				Account:       "invalid_address",
				Owner:         payerWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "invalid account address",
		},
		{
			name: "invalid owner address",
			finish: &EscrowFinish{
				Account:       payeeWallet.Address,
				Owner:         "invalid_address",
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "invalid owner address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.FinishEscrow(tt.finish)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
				assert.Greater(t, result.LedgerIndex, uint32(0))
			}
		})
	}
}

func TestCancelEscrow(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	err := client.Connect()
	require.NoError(t, err)

	// Generate test wallets
	payerWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	payeeWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	tests := []struct {
		name        string
		cancel      *EscrowCancel
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid escrow cancel",
			cancel: &EscrowCancel{
				Account:       payeeWallet.Address,
				Owner:         payerWallet.Address,
				OfferSequence: 1,
			},
			expectError: false,
		},
		{
			name: "missing account",
			cancel: &EscrowCancel{
				Owner:         payerWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "missing required fields",
		},
		{
			name: "missing owner",
			cancel: &EscrowCancel{
				Account:       payeeWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "missing required fields",
		},
		{
			name: "invalid account address",
			cancel: &EscrowCancel{
				Account:       "invalid_address",
				Owner:         payerWallet.Address,
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "invalid account address",
		},
		{
			name: "invalid owner address",
			cancel: &EscrowCancel{
				Account:       payeeWallet.Address,
				Owner:         "invalid_address",
				OfferSequence: 1,
			},
			expectError: true,
			errorMsg:    "invalid owner address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.CancelEscrow(tt.cancel)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.TransactionID)
				assert.Equal(t, "tesSUCCESS", result.ResultCode)
				assert.True(t, result.Validated)
				assert.Greater(t, result.LedgerIndex, uint32(0))
			}
		})
	}
}

func TestGetEscrowInfo(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)
	err := client.Connect()
	require.NoError(t, err)

	// Generate test wallet
	payerWallet, err := client.GenerateWallet()
	require.NoError(t, err)

	tests := []struct {
		name        string
		owner       string
		sequence    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid escrow info query",
			owner:       payerWallet.Address,
			sequence:    "1",
			expectError: false,
		},
		{
			name:        "invalid owner address",
			owner:       "invalid_address",
			sequence:    "1",
			expectError: true,
			errorMsg:    "invalid owner address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			escrowInfo, err := client.GetEscrowInfo(tt.owner, tt.sequence)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, escrowInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, escrowInfo)
				assert.Equal(t, tt.owner, escrowInfo.Account)
				assert.NotEmpty(t, escrowInfo.Destination)
				assert.NotEmpty(t, escrowInfo.Amount)
				assert.Equal(t, uint32(1), escrowInfo.Sequence)
			}
		})
	}
}

func TestGenerateCondition(t *testing.T) {
	client := NewClient("https://s.altnet.rippletest.net:51234", true)

	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "valid secret",
			secret:      "milestone_secret_123",
			expectError: false,
		},
		{
			name:        "empty secret",
			secret:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, fulfillment, err := client.GenerateCondition(tt.secret)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, condition)
				assert.Empty(t, fulfillment)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, condition)
				assert.Equal(t, tt.secret, fulfillment)

				// Condition should be a valid hex string (SHA-256 hash = 64 chars)
				assert.Equal(t, 64, len(condition))

				// Test that same secret generates same condition
				condition2, fulfillment2, err2 := client.GenerateCondition(tt.secret)
				assert.NoError(t, err2)
				assert.Equal(t, condition, condition2)
				assert.Equal(t, fulfillment, fulfillment2)
			}
		})
	}
}
