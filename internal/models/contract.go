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
	SequenceNumber       int            `json:"sequence_number" db:"sequence_number"` // numeric sequence for ordering
	Dependencies         []string       `json:"dependencies" db:"-"`
	Category             string         `json:"category" db:"category"` // delivery, payment, approval, compliance
	Priority             int            `json:"priority" db:"priority"` // higher = more important
	CriticalPath         bool           `json:"critical_path" db:"critical_path"`
	TriggerConditions    string         `json:"trigger_conditions" db:"trigger_conditions"`
	VerificationCriteria string         `json:"verification_criteria" db:"verification_criteria"`
	EstimatedStartDate   *time.Time     `json:"estimated_start_date" db:"estimated_start_date"`
	EstimatedEndDate     *time.Time     `json:"estimated_end_date" db:"estimated_end_date"`
	ActualStartDate      *time.Time     `json:"actual_start_date" db:"actual_start_date"`
	ActualEndDate        *time.Time     `json:"actual_end_date" db:"actual_end_date"`
	EstimatedDuration    time.Duration  `json:"estimated_duration" db:"estimated_duration"`
	ActualDuration       *time.Duration `json:"actual_duration" db:"actual_duration"`
	PercentageComplete   float64        `json:"percentage_complete" db:"percentage_complete"`
	RiskLevel            string         `json:"risk_level" db:"risk_level"`
	ContingencyPlans     []string       `json:"contingency_plans" db:"-"`
	CriticalityScore     int            `json:"criticality_score" db:"criticality_score"`
	Status               string         `json:"status" db:"status"` // Add Status field
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
}

// MilestoneTemplate defines reusable milestone patterns that can be instantiated
// into ContractMilestone entries. Templates support variables and versioning.
type MilestoneTemplate struct {
	ID              string    `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	DefaultCategory string    `json:"default_category" db:"default_category"`
	DefaultPriority int       `json:"default_priority" db:"default_priority"`
	Variables       []string  `json:"variables" db:"-"`
	Version         string    `json:"version" db:"version"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// MilestoneDependency represents a dependency between two milestones.
type MilestoneDependency struct {
	ID             string `json:"id" db:"id"`
	MilestoneID    string `json:"milestone_id" db:"milestone_id"`
	DependsOnID    string `json:"depends_on_id" db:"depends_on_id"`
	DependencyType string `json:"dependency_type" db:"dependency_type"` // prerequisite, parallel, conditional
}
