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
	TransactionStatusCancelled  TransactionStatus = "cancelled"
	TransactionStatusExpired    TransactionStatus = "expired"
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
	SmartChequeID *string `json:"smart_cheque_id,omitempty" gorm:"type:varchar(255);index"`
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
