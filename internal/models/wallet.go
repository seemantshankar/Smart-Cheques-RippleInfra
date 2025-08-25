package models

import (
	"time"

	"github.com/google/uuid"
)

type WalletStatus string

const (
	WalletStatusPending    WalletStatus = "pending"
	WalletStatusActive     WalletStatus = "active"
	WalletStatusSuspended  WalletStatus = "suspended"
	WalletStatusDeactivated WalletStatus = "deactivated"
)

type Wallet struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	EnterpriseID uuid.UUID    `json:"enterprise_id" db:"enterprise_id"`
	Address      string       `json:"address" db:"address"`
	PublicKey    string       `json:"public_key" db:"public_key"`
	Status       WalletStatus `json:"status" db:"status"`
	IsWhitelisted bool        `json:"is_whitelisted" db:"is_whitelisted"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	LastActivity *time.Time   `json:"last_activity,omitempty" db:"last_activity"`
	
	// Encrypted private key and seed (stored separately for security)
	EncryptedPrivateKey string `json:"-" db:"encrypted_private_key"`
	EncryptedSeed       string `json:"-" db:"encrypted_seed"`
	
	// Metadata
	NetworkType string            `json:"network_type" db:"network_type"` // testnet, mainnet
	Metadata    map[string]string `json:"metadata,omitempty" db:"metadata"`
}

type WalletCreateRequest struct {
	EnterpriseID uuid.UUID `json:"enterprise_id" validate:"required"`
	NetworkType  string    `json:"network_type" validate:"required,oneof=testnet mainnet"`
}

type WalletUpdateRequest struct {
	Status        *WalletStatus `json:"status,omitempty"`
	IsWhitelisted *bool         `json:"is_whitelisted,omitempty"`
}

type WalletResponse struct {
	ID            uuid.UUID    `json:"id"`
	EnterpriseID  uuid.UUID    `json:"enterprise_id"`
	Address       string       `json:"address"`
	PublicKey     string       `json:"public_key"`
	Status        WalletStatus `json:"status"`
	IsWhitelisted bool         `json:"is_whitelisted"`
	NetworkType   string       `json:"network_type"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	LastActivity  *time.Time   `json:"last_activity,omitempty"`
}

func (w *Wallet) ToResponse() *WalletResponse {
	return &WalletResponse{
		ID:            w.ID,
		EnterpriseID:  w.EnterpriseID,
		Address:       w.Address,
		PublicKey:     w.PublicKey,
		Status:        w.Status,
		IsWhitelisted: w.IsWhitelisted,
		NetworkType:   w.NetworkType,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
		LastActivity:  w.LastActivity,
	}
}

func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

func (w *Wallet) CanTransact() bool {
	return w.IsActive() && w.IsWhitelisted
}