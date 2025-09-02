package models

import (
	"time"
)

// CBDCWallet represents a CBDC wallet for an enterprise
type CBDCWallet struct {
	ID            string           `json:"id" db:"id"`
	EnterpriseID  string           `json:"enterprise_id" db:"enterprise_id"`
	WalletAddress string           `json:"wallet_address" db:"wallet_address"`
	Currency      Currency         `json:"currency" db:"currency"`
	Status        CBDCWalletStatus `json:"status" db:"status"`
	Balance       float64          `json:"balance" db:"balance"`
	Limit         float64          `json:"limit" db:"limit"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	ActivatedAt   *time.Time       `json:"activated_at,omitempty" db:"activated_at"`
	SuspendedAt   *time.Time       `json:"suspended_at,omitempty" db:"suspended_at"`
}

// CBDCWalletStatus represents the status of a CBDC wallet
type CBDCWalletStatus string

const (
	CBDCWalletStatusPending   CBDCWalletStatus = "pending"
	CBDCWalletStatusActive    CBDCWalletStatus = "active"
	CBDCWalletStatusSuspended CBDCWalletStatus = "suspended"
	CBDCWalletStatusFrozen    CBDCWalletStatus = "frozen"
	CBDCWalletStatusClosed    CBDCWalletStatus = "closed"
)

// CBDCTransaction represents a CBDC transaction
type CBDCTransaction struct {
	ID              string                 `json:"id" db:"id"`
	WalletID        string                 `json:"wallet_id" db:"wallet_id"`
	TransactionHash string                 `json:"transaction_hash" db:"transaction_hash"`
	Type            CBDCTransactionType    `json:"type" db:"type"`
	Amount          float64                `json:"amount" db:"amount"`
	Currency        Currency               `json:"currency" db:"currency"`
	Status          CBDCTransactionStatus  `json:"status" db:"status"`
	FromAddress     string                 `json:"from_address" db:"from_address"`
	ToAddress       string                 `json:"to_address" db:"to_address"`
	Description     string                 `json:"description" db:"description"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	ConfirmedAt     *time.Time             `json:"confirmed_at,omitempty" db:"confirmed_at"`
	FailedAt        *time.Time             `json:"failed_at,omitempty" db:"failed_at"`
	FailureReason   *string                `json:"failure_reason,omitempty" db:"failure_reason"`
}

// CBDCTransactionType represents the type of CBDC transaction
type CBDCTransactionType string

const (
	CBDCTransactionTypeTransfer CBDCTransactionType = "transfer"
	CBDCTransactionTypePayment  CBDCTransactionType = "payment"
	CBDCTransactionTypeRefund   CBDCTransactionType = "refund"
	CBDCTransactionTypeDeposit  CBDCTransactionType = "deposit"
	CBDCTransactionTypeWithdraw CBDCTransactionType = "withdraw"
)

// CBDCTransactionStatus represents the status of a CBDC transaction
type CBDCTransactionStatus string

const (
	CBDCTransactionStatusPending    CBDCTransactionStatus = "pending"
	CBDCTransactionStatusProcessing CBDCTransactionStatus = "processing"
	CBDCTransactionStatusConfirmed  CBDCTransactionStatus = "confirmed"
	CBDCTransactionStatusFailed     CBDCTransactionStatus = "failed"
	CBDCTransactionStatusCancelled  CBDCTransactionStatus = "canceled"
)

// CBDCBalance represents the balance of a CBDC wallet
type CBDCBalance struct {
	ID          string    `json:"id" db:"id"`
	WalletID    string    `json:"wallet_id" db:"wallet_id"`
	Available   float64   `json:"available" db:"available"`
	Reserved    float64   `json:"reserved" db:"reserved"`
	Total       float64   `json:"total" db:"total"`
	Currency    Currency  `json:"currency" db:"currency"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// CBDCWalletRequest represents a request for CBDC wallet operations
type CBDCWalletRequest struct {
	ID              string                 `json:"id" db:"id"`
	EnterpriseID    string                 `json:"enterprise_id" db:"enterprise_id"`
	Type            CBDCWalletRequestType  `json:"type" db:"type"`
	Status          CBDCRequestStatus      `json:"status" db:"status"`
	Amount          *float64               `json:"amount,omitempty" db:"amount"`
	Currency        Currency               `json:"currency" db:"currency"`
	Reason          string                 `json:"reason" db:"reason"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty" db:"approved_at"`
	RejectedAt      *time.Time             `json:"rejected_at,omitempty" db:"rejected_at"`
	RejectionReason *string                `json:"rejection_reason,omitempty" db:"rejection_reason"`
}

// CBDCWalletRequestType represents the type of CBDC wallet request
type CBDCWalletRequestType string

const (
	CBDCWalletRequestTypeCreate   CBDCWalletRequestType = "create"
	CBDCWalletRequestTypeActivate CBDCWalletRequestType = "activate"
	CBDCWalletRequestTypeSuspend  CBDCWalletRequestType = "suspend"
	CBDCWalletRequestTypeFreeze   CBDCWalletRequestType = "freeze"
	CBDCWalletRequestTypeIncrease CBDCWalletRequestType = "increase_limit"
	CBDCWalletRequestTypeDecrease CBDCWalletRequestType = "decrease_limit"
	CBDCWalletRequestTypeClose    CBDCWalletRequestType = "close"
)

// CBDCRequestStatus represents the status of a CBDC request
type CBDCRequestStatus string

const (
	CBDCRequestStatusPending   CBDCRequestStatus = "pending"
	CBDCRequestStatusApproved  CBDCRequestStatus = "approved"
	CBDCRequestStatusRejected  CBDCRequestStatus = "rejected"
	CBDCRequestStatusCancelled CBDCRequestStatus = "canceled"
)

// TSPConfig represents the configuration for a TSP (Technology Service Provider)
type TSPConfig struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Endpoint     string                 `json:"endpoint" db:"endpoint"`
	APIKey       string                 `json:"api_key" db:"api_key"`
	SecretKey    string                 `json:"secret_key" db:"secret_key"`
	Environment  string                 `json:"environment" db:"environment"`
	Status       TSPStatus              `json:"status" db:"status"`
	Capabilities []string               `json:"capabilities" db:"capabilities"`
	Config       map[string]interface{} `json:"config" db:"config"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// TSPStatus represents the status of a TSP
type TSPStatus string

const (
	TSPStatusActive   TSPStatus = "active"
	TSPStatusInactive TSPStatus = "inactive"
	TSPStatusTesting  TSPStatus = "testing"
	TSPStatusError    TSPStatus = "error"
)
