package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TransactionStatus represents the current status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusQueued     TransactionStatus = "queued"
	TransactionStatusBatching   TransactionStatus = "batching"
	TransactionStatusBatched    TransactionStatus = "batched"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusSubmitted  TransactionStatus = "submitted"
	TransactionStatusConfirmed  TransactionStatus = "confirmed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusCancelled  TransactionStatus = "canceled"
	TransactionStatusExpired    TransactionStatus = "expired"
	TransactionStatusFraud      TransactionStatus = "fraud"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeEscrowCreate TransactionType = "escrow_create"
	TransactionTypeEscrowFinish TransactionType = "escrow_finish"
	TransactionTypeEscrowCancel TransactionType = "escrow_cancel"
	TransactionTypePayment      TransactionType = "payment"
	TransactionTypeWalletSetup  TransactionType = "wallet_setup"
)

// TransactionPriority represents the priority level of a transaction
type TransactionPriority int

const (
	PriorityLow      TransactionPriority = 1
	PriorityNormal   TransactionPriority = 2
	PriorityHigh     TransactionPriority = 3
	PriorityCritical TransactionPriority = 4
)

// Transaction represents a blockchain transaction in the queue system
type Transaction struct {
	ID       string              `json:"id" gorm:"primaryKey"`
	Type     TransactionType     `json:"type" gorm:"type:varchar(50);not null"`
	Status   TransactionStatus   `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`
	Priority TransactionPriority `json:"priority" gorm:"type:int;not null;default:2"`
	BatchID  *string             `json:"batch_id,omitempty" gorm:"type:varchar(255);index"`

	// Transaction Details
	FromAddress string `json:"from_address" gorm:"type:varchar(255);not null"`
	ToAddress   string `json:"to_address" gorm:"type:varchar(255);not null"`
	Amount      string `json:"amount" gorm:"type:varchar(255);not null"`
	Currency    string `json:"currency" gorm:"type:varchar(10);not null;default:'XRP'"`
	Fee         string `json:"fee" gorm:"type:varchar(255)"`

	// XRPL Specific Fields
	Sequence        *uint32 `json:"sequence,omitempty" gorm:"type:int"`
	LedgerIndex     *uint32 `json:"ledger_index,omitempty" gorm:"type:int"`
	TransactionHash string  `json:"transaction_hash,omitempty" gorm:"type:varchar(255);index"`

	// Escrow Specific Fields
	Condition     string  `json:"condition,omitempty" gorm:"type:text"`
	Fulfillment   string  `json:"fulfillment,omitempty" gorm:"type:text"`
	CancelAfter   *uint32 `json:"cancel_after,omitempty" gorm:"type:int"`
	FinishAfter   *uint32 `json:"finish_after,omitempty" gorm:"type:int"`
	OfferSequence *uint32 `json:"offer_sequence,omitempty" gorm:"type:int"`

	// Business Context
	SmartChequeID *string `json:"smart_check_id,omitempty" gorm:"type:varchar(255);index"`
	MilestoneID   *string `json:"milestone_id,omitempty" gorm:"type:varchar(255);index"`
	EnterpriseID  string  `json:"enterprise_id" gorm:"type:varchar(255);not null;index"`
	UserID        string  `json:"user_id" gorm:"type:varchar(255);not null;index"`

	// Error Handling
	RetryCount int    `json:"retry_count" gorm:"type:int;not null;default:0"`
	MaxRetries int    `json:"max_retries" gorm:"type:int;not null;default:3"`
	LastError  string `json:"last_error,omitempty" gorm:"type:text"`

	// Metadata
	Metadata TransactionMetadata `json:"metadata" gorm:"type:jsonb"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// TransactionMetadata stores additional transaction-specific data
type TransactionMetadata map[string]interface{}

// Value implements the driver.Valuer interface for GORM JSON storage
func (tm TransactionMetadata) Value() (driver.Value, error) {
	return json.Marshal(tm)
}

// Scan implements the sql.Scanner interface for GORM JSON storage
func (tm *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		*tm = make(TransactionMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TransactionMetadata: %v", value)
	}

	return json.Unmarshal(bytes, tm)
}

// TransactionBatch represents a collection of transactions processed together
type TransactionBatch struct {
	ID       string              `json:"id" gorm:"primaryKey"`
	Status   TransactionStatus   `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`
	Priority TransactionPriority `json:"priority" gorm:"type:int;not null;default:2"`

	// Batch Configuration
	MaxTransactions int    `json:"max_transactions" gorm:"type:int;not null;default:10"`
	TotalFee        string `json:"total_fee" gorm:"type:varchar(255)"`
	OptimizedFee    string `json:"optimized_fee" gorm:"type:varchar(255)"`
	FeeSavings      string `json:"fee_savings" gorm:"type:varchar(255)"`

	// Processing Details
	TransactionCount int `json:"transaction_count" gorm:"type:int;not null;default:0"`
	SuccessCount     int `json:"success_count" gorm:"type:int;not null;default:0"`
	FailureCount     int `json:"failure_count" gorm:"type:int;not null;default:0"`

	// Error Handling
	RetryCount int    `json:"retry_count" gorm:"type:int;not null;default:0"`
	MaxRetries int    `json:"max_retries" gorm:"type:int;not null;default:3"`
	LastError  string `json:"last_error,omitempty" gorm:"type:text"`

	// Relationships
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:BatchID"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// TransactionStats represents transaction processing statistics
type TransactionStats struct {
	TotalTransactions      int64      `json:"total_transactions"`
	PendingTransactions    int64      `json:"pending_transactions"`
	ProcessingTransactions int64      `json:"processing_transactions"`
	CompletedTransactions  int64      `json:"completed_transactions"`
	FailedTransactions     int64      `json:"failed_transactions"`
	AverageProcessingTime  float64    `json:"average_processing_time_seconds"`
	TotalFeesProcessed     string     `json:"total_fees_processed"`
	TotalFeeSavings        string     `json:"total_fee_savings"`
	LastProcessedAt        *time.Time `json:"last_processed_at"`
}

// TransactionAuditLog represents a transaction-specific audit log entry
type TransactionAuditLog struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	TransactionID  string                 `json:"transaction_id" gorm:"not null;index"`
	EventType      string                 `json:"event_type" gorm:"not null"` // created, status_changed, retried, failed, completed
	PreviousStatus TransactionStatus      `json:"previous_status" gorm:"type:varchar(50)"`
	NewStatus      TransactionStatus      `json:"new_status" gorm:"type:varchar(50)"`
	UserID         string                 `json:"user_id" gorm:"not null"`
	EnterpriseID   string                 `json:"enterprise_id" gorm:"not null"`
	Details        string                 `json:"details,omitempty" gorm:"type:text"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt      time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// TransactionRiskScore represents a risk assessment for a transaction
type TransactionRiskScore struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	TransactionID     string     `json:"transaction_id" gorm:"not null;uniqueIndex"`
	RiskLevel         string     `json:"risk_level" gorm:"type:varchar(20);not null"` // low, medium, high, critical
	RiskScore         float64    `json:"risk_score" gorm:"type:decimal(5,4);not null"`
	RiskFactors       []string   `json:"risk_factors" gorm:"type:jsonb"`
	AssessmentDetails string     `json:"assessment_details" gorm:"type:text"`
	AssessedAt        time.Time  `json:"assessed_at" gorm:"autoCreateTime"`
	AssessedBy        string     `json:"assessed_by" gorm:"not null"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// TransactionComplianceStatus represents compliance status for a transaction
type TransactionComplianceStatus struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	TransactionID string     `json:"transaction_id" gorm:"not null;uniqueIndex"`
	Status        string     `json:"status" gorm:"type:varchar(20);not null"` // pending, approved, rejected, flagged
	ChecksPassed  []string   `json:"checks_passed" gorm:"type:jsonb"`
	ChecksFailed  []string   `json:"checks_failed" gorm:"type:jsonb"`
	Violations    []string   `json:"violations" gorm:"type:jsonb"`
	ReviewedBy    *string    `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	Comments      string     `json:"comments,omitempty" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TransactionReport represents a transaction monitoring report
type TransactionReport struct {
	ID           string                   `json:"id" gorm:"primaryKey"`
	ReportType   string                   `json:"report_type" gorm:"type:varchar(50);not null"` // daily, weekly, monthly, custom
	EnterpriseID string                   `json:"enterprise_id" gorm:"not null;index"`
	PeriodStart  time.Time                `json:"period_start" gorm:"not null"`
	PeriodEnd    time.Time                `json:"period_end" gorm:"not null"`
	Summary      TransactionReportSummary `json:"summary" gorm:"type:jsonb"`
	GeneratedAt  time.Time                `json:"generated_at" gorm:"autoCreateTime"`
	GeneratedBy  string                   `json:"generated_by" gorm:"not null"`
}

// TransactionReportSummary contains aggregated data for reports
type TransactionReportSummary struct {
	TotalTransactions      int64          `json:"total_transactions"`
	SuccessfulTransactions int64          `json:"successful_transactions"`
	FailedTransactions     int64          `json:"failed_transactions"`
	HighRiskTransactions   int64          `json:"high_risk_transactions"`
	ComplianceViolations   int64          `json:"compliance_violations"`
	AverageProcessingTime  float64        `json:"average_processing_time_seconds"`
	TotalVolume            string         `json:"total_volume"`
	TotalFees              string         `json:"total_fees"`
	TopFailureReasons      map[string]int `json:"top_failure_reasons"`
	RiskDistribution       map[string]int `json:"risk_distribution"`
}

// ComplianceStats represents compliance statistics
type ComplianceStats struct {
	TotalTransactions    int64 `json:"total_transactions"`
	ApprovedTransactions int64 `json:"approved_transactions"`
	RejectedTransactions int64 `json:"rejected_transactions"`
	FlaggedTransactions  int64 `json:"flagged_transactions"`
	PendingTransactions  int64 `json:"pending_transactions"`
	ReviewedTransactions int64 `json:"reviewed_transactions"`
}

// ComplianceReport represents a compliance-focused report
type ComplianceReport struct {
	ID                  string          `json:"id" gorm:"primaryKey"`
	EnterpriseID        string          `json:"enterprise_id" gorm:"not null;index"`
	PeriodStart         time.Time       `json:"period_start" gorm:"not null"`
	PeriodEnd           time.Time       `json:"period_end" gorm:"not null"`
	ComplianceStats     ComplianceStats `json:"compliance_stats" gorm:"type:jsonb"`
	FlaggedTransactions int             `json:"flagged_transactions"`
	GeneratedAt         time.Time       `json:"generated_at" gorm:"autoCreateTime"`
	GeneratedBy         string          `json:"generated_by" gorm:"not null"`
}

// RiskReport represents a risk-focused report
type RiskReport struct {
	ID           string      `json:"id" gorm:"primaryKey"`
	EnterpriseID string      `json:"enterprise_id" gorm:"not null;index"`
	PeriodStart  time.Time   `json:"period_start" gorm:"not null"`
	PeriodEnd    time.Time   `json:"period_end" gorm:"not null"`
	RiskMetrics  RiskMetrics `json:"risk_metrics" gorm:"type:jsonb"`
	GeneratedAt  time.Time   `json:"generated_at" gorm:"autoCreateTime"`
	GeneratedBy  string      `json:"generated_by" gorm:"not null"`
}

// RiskMetrics contains risk-related metrics
type RiskMetrics struct {
	HighRiskTransactions     int      `json:"high_risk_transactions"`
	CriticalRiskTransactions int      `json:"critical_risk_transactions"`
	RiskTrend                string   `json:"risk_trend"`
	TopRiskFactors           []string `json:"top_risk_factors"`
	MitigationActions        []string `json:"mitigation_actions"`
}

// TransactionAnalytics represents detailed transaction analytics
type TransactionAnalytics struct {
	EnterpriseID string             `json:"enterprise_id"`
	PeriodStart  time.Time          `json:"period_start"`
	PeriodEnd    time.Time          `json:"period_end"`
	Metrics      TransactionMetrics `json:"metrics"`
	Trends       TransactionTrends  `json:"trends"`
	GeneratedAt  time.Time          `json:"generated_at"`
}

// TransactionMetrics contains transaction performance metrics
type TransactionMetrics struct {
	TotalVolume            string  `json:"total_volume"`
	AverageTransactionSize string  `json:"average_transaction_size"`
	PeakHourVolume         string  `json:"peak_hour_volume"`
	TransactionVelocity    int     `json:"transaction_velocity"`
	SuccessRate            float64 `json:"success_rate"`
	FailureRate            float64 `json:"failure_rate"`
}

// TransactionTrends contains transaction trend data
type TransactionTrends struct {
	VolumeGrowth      float64 `json:"volume_growth"`
	SuccessRateChange float64 `json:"success_rate_change"`
	RiskIncrease      float64 `json:"risk_increase"`
}

// BatchConfig represents configuration for transaction batching
type BatchConfig struct {
	MaxBatchSize        int           `json:"max_batch_size"`
	MaxWaitTime         time.Duration `json:"max_wait_time"`
	FeeOptimization     bool          `json:"fee_optimization"`
	PriorityBatching    bool          `json:"priority_batching"`
	MinBatchSize        int           `json:"min_batch_size"`
	BatchTimeoutSeconds int           `json:"batch_timeout_seconds"`
}

// DefaultBatchConfig returns a sensible default batch configuration
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxBatchSize:        10,
		MaxWaitTime:         30 * time.Second,
		FeeOptimization:     true,
		PriorityBatching:    true,
		MinBatchSize:        2,
		BatchTimeoutSeconds: 300, // 5 minutes
	}
}

// NewTransaction creates a new transaction with default values
func NewTransaction(txType TransactionType, fromAddr, toAddr, amount, currency, enterpriseID, userID string) *Transaction {
	return &Transaction{
		ID:           uuid.New().String(),
		Type:         txType,
		Status:       TransactionStatusPending,
		Priority:     PriorityNormal,
		FromAddress:  fromAddr,
		ToAddress:    toAddr,
		Amount:       amount,
		Currency:     currency,
		EnterpriseID: enterpriseID,
		UserID:       userID,
		RetryCount:   0,
		MaxRetries:   3,
		Metadata:     make(TransactionMetadata),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewTransactionBatch creates a new transaction batch
func NewTransactionBatch(priority TransactionPriority, maxTransactions int) *TransactionBatch {
	return &TransactionBatch{
		ID:               uuid.New().String(),
		Status:           TransactionStatusPending,
		Priority:         priority,
		MaxTransactions:  maxTransactions,
		TransactionCount: 0,
		SuccessCount:     0,
		FailureCount:     0,
		RetryCount:       0,
		MaxRetries:       3,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// CanRetry checks if the transaction can be retried
func (t *Transaction) CanRetry() bool {
	return t.RetryCount < t.MaxRetries &&
		(t.Status == TransactionStatusFailed || t.Status == TransactionStatusExpired)
}

// IncrementRetry increments the retry count and updates status
func (t *Transaction) IncrementRetry() {
	t.RetryCount++
	t.UpdatedAt = time.Now()
	if t.RetryCount >= t.MaxRetries {
		t.Status = TransactionStatusFailed
	} else {
		t.Status = TransactionStatusPending
	}
}

// SetError sets the error message and updates status
func (t *Transaction) SetError(err error) {
	t.LastError = err.Error()
	t.Status = TransactionStatusFailed
	t.UpdatedAt = time.Now()
}

// IsExpired checks if the transaction has expired
func (t *Transaction) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// CanBatch checks if the transaction can be included in a batch
func (t *Transaction) CanBatch() bool {
	return t.Status == TransactionStatusQueued &&
		t.Type != TransactionTypeWalletSetup && // Wallet setup should be processed individually
		!t.IsExpired()
}

// EstimatedFee calculates the estimated transaction fee (mock implementation)
func (t *Transaction) EstimatedFee() string {
	// Basic fee calculation - in production this would use real XRPL fee calculations
	switch t.Type {
	case TransactionTypeEscrowCreate:
		return "12" // 12 drops
	case TransactionTypeEscrowFinish:
		return "12" // 12 drops
	case TransactionTypeEscrowCancel:
		return "12" // 12 drops
	case TransactionTypePayment:
		return "10" // 10 drops
	case TransactionTypeWalletSetup:
		return "15" // 15 drops
	default:
		return "10" // 10 drops default
	}
}
