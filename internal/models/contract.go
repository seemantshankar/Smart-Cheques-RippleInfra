package models

import (
	"time"
)

type Contract struct {
	ID                string           `json:"id" db:"id"`
	Parties           []string         `json:"parties"`
	Obligations       []Obligation     `json:"obligations"`
	PaymentTerms      []PaymentTerm    `json:"payment_terms"`
	DisputeResolution DisputeConfig    `json:"dispute_resolution"`
	AIAnalysis        ContractAnalysis `json:"ai_analysis"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

type Obligation struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Party       string    `json:"party"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
}

type PaymentTerm struct {
	ID         string    `json:"id"`
	Amount     float64   `json:"amount"`
	Currency   Currency  `json:"currency"`
	DueDate    time.Time `json:"due_date"`
	Conditions []string  `json:"conditions"`
}

type DisputeConfig struct {
	Method           string `json:"method"`
	ArbitrationRules string `json:"arbitration_rules"`
	Jurisdiction     string `json:"jurisdiction"`
}

type ContractAnalysis struct {
	ConfidenceScore float64           `json:"confidence_score"`
	ExtractedTerms  map[string]string `json:"extracted_terms"`
	RiskFactors     []string          `json:"risk_factors"`
	Recommendations []string          `json:"recommendations"`
	AnalyzedAt      time.Time         `json:"analyzed_at"`
}
