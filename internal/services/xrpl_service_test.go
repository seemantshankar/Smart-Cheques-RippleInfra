package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewXRPLService(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	
	service := NewXRPLService(config)
	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
}

func TestXRPLService_Initialize(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	
	service := NewXRPLService(config)
	
	err := service.Initialize()
	assert.NoError(t, err)
}

func TestXRPLService_CreateWallet(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	
	service := NewXRPLService(config)
	err := service.Initialize()
	require.NoError(t, err)
	
	wallet, err := service.CreateWallet()
	require.NoError(t, err)
	require.NotNil(t, wallet)
	
	assert.NotEmpty(t, wallet.Address)
	assert.NotEmpty(t, wallet.PublicKey)
	assert.NotEmpty(t, wallet.PrivateKey)
	assert.NotEmpty(t, wallet.Seed)
	
	// Validate the generated address
	assert.True(t, service.ValidateAddress(wallet.Address))
}

func TestXRPLService_ValidateAddress(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	
	service := NewXRPLService(config)
	
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "valid address",
			address:  "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
			expected: true,
		},
		{
			name:     "invalid address",
			address:  "invalid",
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
			result := service.ValidateAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestXRPLService_HealthCheck(t *testing.T) {
	config := XRPLConfig{
		NetworkURL: "https://s.altnet.rippletest.net:51234",
		TestNet:    true,
	}
	
	service := NewXRPLService(config)
	
	// Should fail before initialization
	err := service.HealthCheck()
	assert.Error(t, err)
	
	// Should pass after initialization
	err = service.Initialize()
	require.NoError(t, err)
	
	err = service.HealthCheck()
	assert.NoError(t, err)
}