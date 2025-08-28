package models

import (
	"time"
)

type Contract struct {
	ID                string             `json:"id" db:"id"`
	Parties           []string           `json:"parties"`
	Obligations       []Obligation       `json:"obligations"`
	PaymentTerms      []PaymentTerm      `json:"payment_terms"`
	DisputeResolution DisputeConfig      `json:"dispute_resolution"`
	AIAnalysis        ContractAnalysis   `json:"ai_analysis"`
	Status            string             `json:"status" db:"status"`               // draft, active, executed, terminated, disputed
	ContractType      string             `json:"contract_type" db:"contract_type"` // service_agreement, purchase_order, milestone_based
	Version           string             `json:"version" db:"version"`
	ParentContractID  *string            `json:"parent_contract_id" db:"parent_contract_id"`
	DocumentMetadata  DocumentMetadata   `json:"document_metadata" db:"-"`
	DigitalSignatures []DigitalSignature `json:"digital_signatures" db:"-"`
	Tags              []string           `json:"tags" db:"-"`
	Categories        []string           `json:"categories" db:"-"`
	ExpirationDate    *time.Time         `json:"expiration_date" db:"expiration_date"`
	RenewalTerms      string             `json:"renewal_terms" db:"renewal_terms"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
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

type DocumentMetadata struct {
	OriginalFilename string `json:"original_filename"`
	FileSize         int64  `json:"file_size"`
	MimeType         string `json:"mime_type"`
}

type DigitalSignature struct {
	SignerID    string    `json:"signer_id"`
	Signature   string    `json:"signature"`
	SignedAt    time.Time `json:"signed_at"`
	Verified    bool      `json:"verified"`
	Certificate string    `json:"certificate"`
}

type ContractMilestone struct {
	ID                   string         `json:"id" db:"id"`
	ContractID           string         `json:"contract_id" db:"contract_id"`
	MilestoneID          string         `json:"milestone_id" db:"milestone_id"`
	SequenceOrder        int            `json:"sequence_order" db:"sequence_order"`
	Dependencies         []string       `json:"dependencies" db:"-"`
	TriggerConditions    string         `json:"trigger_conditions" db:"trigger_conditions"`
	VerificationCriteria string         `json:"verification_criteria" db:"verification_criteria"`
	EstimatedDuration    time.Duration  `json:"estimated_duration" db:"estimated_duration"`
	ActualDuration       *time.Duration `json:"actual_duration" db:"actual_duration"`
	RiskLevel            string         `json:"risk_level" db:"risk_level"`
	CriticalityScore     int            `json:"criticality_score" db:"criticality_score"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}
