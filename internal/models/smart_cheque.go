package models

import (
	"time"
)

type SmartCheque struct {
	ID            string            `json:"id" db:"id"`
	PayerID       string            `json:"payer_id" db:"payer_id"`
	PayeeID       string            `json:"payee_id" db:"payee_id"`
	Amount        float64           `json:"amount" db:"amount"`
	Currency      Currency          `json:"currency" db:"currency"`
	Milestones    []Milestone       `json:"milestones"`
	EscrowAddress string            `json:"escrow_address" db:"escrow_address"`
	Status        SmartChequeStatus `json:"status" db:"status"`
	ContractHash  string            `json:"contract_hash" db:"contract_hash"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
}

type Currency string

const (
	CurrencyUSDT Currency = "USDT"
	CurrencyUSDC Currency = "USDC"
	CurrencyERupee Currency = "e₹"
)

type SmartChequeStatus string

const (
	SmartChequeStatusCreated    SmartChequeStatus = "created"
	SmartChequeStatusLocked     SmartChequeStatus = "locked"
	SmartChequeStatusInProgress SmartChequeStatus = "in_progress"
	SmartChequeStatusCompleted  SmartChequeStatus = "completed"
	SmartChequeStatusDisputed   SmartChequeStatus = "disputed"
)

type Milestone struct {
	ID                 string               `json:"id"`
	Description        string               `json:"description"`
	Amount             float64              `json:"amount"`
	VerificationMethod VerificationMethod   `json:"verification_method"`
	OracleConfig       *OracleConfig        `json:"oracle_config,omitempty"`
	Status             MilestoneStatus      `json:"status"`
	CompletedAt        *time.Time           `json:"completed_at,omitempty"`
}

type VerificationMethod string

const (
	VerificationMethodOracle VerificationMethod = "oracle"
	VerificationMethodManual VerificationMethod = "manual"
	VerificationMethodHybrid VerificationMethod = "hybrid"
)

type MilestoneStatus string

const (
	MilestoneStatusPending  MilestoneStatus = "pending"
	MilestoneStatusVerified MilestoneStatus = "verified"
	MilestoneStatusFailed   MilestoneStatus = "failed"
)

type OracleConfig struct {
	Type     string                 `json:"type"`
	Endpoint string                 `json:"endpoint"`
	Config   map[string]interface{} `json:"config"`
}