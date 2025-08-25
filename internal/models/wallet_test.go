package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWallet_ToResponse(t *testing.T) {
	now := time.Now()
	walletID := uuid.New()
	enterpriseID := uuid.New()

	wallet := &Wallet{
		ID:            walletID,
		EnterpriseID:  enterpriseID,
		Address:       "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
		PublicKey:     "03ABC123",
		Status:        WalletStatusActive,
		IsWhitelisted: true,
		NetworkType:   "testnet",
		CreatedAt:     now,
		UpdatedAt:     now,
		LastActivity:  &now,
		// Sensitive fields should not be in response
		EncryptedPrivateKey: "encrypted_private_key",
		EncryptedSeed:       "encrypted_seed",
	}

	response := wallet.ToResponse()

	assert.Equal(t, walletID, response.ID)
	assert.Equal(t, enterpriseID, response.EnterpriseID)
	assert.Equal(t, "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH", response.Address)
	assert.Equal(t, "03ABC123", response.PublicKey)
	assert.Equal(t, WalletStatusActive, response.Status)
	assert.True(t, response.IsWhitelisted)
	assert.Equal(t, "testnet", response.NetworkType)
	assert.Equal(t, now, response.CreatedAt)
	assert.Equal(t, now, response.UpdatedAt)
	assert.Equal(t, &now, response.LastActivity)
}

func TestWallet_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   WalletStatus
		expected bool
	}{
		{
			name:     "active wallet",
			status:   WalletStatusActive,
			expected: true,
		},
		{
			name:     "pending wallet",
			status:   WalletStatusPending,
			expected: false,
		},
		{
			name:     "suspended wallet",
			status:   WalletStatusSuspended,
			expected: false,
		},
		{
			name:     "deactivated wallet",
			status:   WalletStatusDeactivated,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := &Wallet{Status: tt.status}
			assert.Equal(t, tt.expected, wallet.IsActive())
		})
	}
}

func TestWallet_CanTransact(t *testing.T) {
	tests := []struct {
		name          string
		status        WalletStatus
		isWhitelisted bool
		expected      bool
	}{
		{
			name:          "active and whitelisted",
			status:        WalletStatusActive,
			isWhitelisted: true,
			expected:      true,
		},
		{
			name:          "active but not whitelisted",
			status:        WalletStatusActive,
			isWhitelisted: false,
			expected:      false,
		},
		{
			name:          "whitelisted but not active",
			status:        WalletStatusPending,
			isWhitelisted: true,
			expected:      false,
		},
		{
			name:          "neither active nor whitelisted",
			status:        WalletStatusSuspended,
			isWhitelisted: false,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet := &Wallet{
				Status:        tt.status,
				IsWhitelisted: tt.isWhitelisted,
			}
			assert.Equal(t, tt.expected, wallet.CanTransact())
		})
	}
}

func TestWalletStatus_Constants(t *testing.T) {
	assert.Equal(t, WalletStatus("pending"), WalletStatusPending)
	assert.Equal(t, WalletStatus("active"), WalletStatusActive)
	assert.Equal(t, WalletStatus("suspended"), WalletStatusSuspended)
	assert.Equal(t, WalletStatus("deactivated"), WalletStatusDeactivated)
}