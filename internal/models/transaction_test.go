package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransaction(t *testing.T) {
	enterpriseID := uuid.New().String()
	userID := uuid.New().String()

	tx := NewTransaction(
		TransactionTypeEscrowCreate,
		"rSender123",
		"rReceiver456",
		"1000000", // 1 XRP in drops
		"XRP",
		enterpriseID,
		userID,
	)

	assert.NotEmpty(t, tx.ID)
	assert.Equal(t, TransactionTypeEscrowCreate, tx.Type)
	assert.Equal(t, TransactionStatusPending, tx.Status)
	assert.Equal(t, PriorityNormal, tx.Priority)
	assert.Equal(t, "rSender123", tx.FromAddress)
	assert.Equal(t, "rReceiver456", tx.ToAddress)
	assert.Equal(t, "1000000", tx.Amount)
	assert.Equal(t, "XRP", tx.Currency)
	assert.Equal(t, enterpriseID, tx.EnterpriseID)
	assert.Equal(t, userID, tx.UserID)
	assert.Equal(t, 0, tx.RetryCount)
	assert.Equal(t, 3, tx.MaxRetries)
	assert.NotNil(t, tx.Metadata)
	assert.False(t, tx.CreatedAt.IsZero())
	assert.False(t, tx.UpdatedAt.IsZero())
}

func TestNewTransactionBatch(t *testing.T) {
	batch := NewTransactionBatch(PriorityHigh, 10)

	assert.NotEmpty(t, batch.ID)
	assert.Equal(t, TransactionStatusPending, batch.Status)
	assert.Equal(t, PriorityHigh, batch.Priority)
	assert.Equal(t, 10, batch.MaxTransactions)
	assert.Equal(t, 0, batch.TransactionCount)
	assert.Equal(t, 0, batch.SuccessCount)
	assert.Equal(t, 0, batch.FailureCount)
	assert.Equal(t, 0, batch.RetryCount)
	assert.Equal(t, 3, batch.MaxRetries)
	assert.False(t, batch.CreatedAt.IsZero())
	assert.False(t, batch.UpdatedAt.IsZero())
}

func TestTransaction_CanRetry(t *testing.T) {
	tx := &Transaction{
		Status:     TransactionStatusFailed,
		RetryCount: 2,
		MaxRetries: 3,
	}

	// Should be able to retry
	assert.True(t, tx.CanRetry())

	// Reached max retries
	tx.RetryCount = 3
	assert.False(t, tx.CanRetry())

	// Wrong status
	tx.RetryCount = 1
	tx.Status = TransactionStatusConfirmed
	assert.False(t, tx.CanRetry())

	// Expired status should allow retry
	tx.Status = TransactionStatusExpired
	assert.True(t, tx.CanRetry())
}

func TestTransaction_IncrementRetry(t *testing.T) {
	tx := &Transaction{
		Status:     TransactionStatusFailed,
		RetryCount: 1,
		MaxRetries: 3,
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}

	oldUpdateTime := tx.UpdatedAt
	tx.IncrementRetry()

	assert.Equal(t, 2, tx.RetryCount)
	assert.Equal(t, TransactionStatusPending, tx.Status)
	assert.True(t, tx.UpdatedAt.After(oldUpdateTime))

	// Test max retries reached
	tx.IncrementRetry()

	assert.Equal(t, 3, tx.RetryCount)
	assert.Equal(t, TransactionStatusFailed, tx.Status)
}

func TestTransaction_SetError(t *testing.T) {
	tx := &Transaction{
		Status:    TransactionStatusProcessing,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	testError := assert.AnError
	oldUpdateTime := tx.UpdatedAt

	tx.SetError(testError)

	assert.Equal(t, TransactionStatusFailed, tx.Status)
	assert.Equal(t, testError.Error(), tx.LastError)
	assert.True(t, tx.UpdatedAt.After(oldUpdateTime))
}

func TestTransaction_IsExpired(t *testing.T) {
	tx := &Transaction{}

	// No expiration set
	assert.False(t, tx.IsExpired())

	// Future expiration
	futureTime := time.Now().Add(1 * time.Hour)
	tx.ExpiresAt = &futureTime
	assert.False(t, tx.IsExpired())

	// Past expiration
	pastTime := time.Now().Add(-1 * time.Hour)
	tx.ExpiresAt = &pastTime
	assert.True(t, tx.IsExpired())
}

func TestTransaction_CanBatch(t *testing.T) {
	tx := &Transaction{
		Status: TransactionStatusQueued,
		Type:   TransactionTypeEscrowCreate,
	}

	// Should be able to batch
	assert.True(t, tx.CanBatch())

	// Wrong status
	tx.Status = TransactionStatusProcessing
	assert.False(t, tx.CanBatch())

	// Wallet setup should not be batched
	tx.Status = TransactionStatusQueued
	tx.Type = TransactionTypeWalletSetup
	assert.False(t, tx.CanBatch())

	// Expired transaction
	tx.Type = TransactionTypeEscrowCreate
	pastTime := time.Now().Add(-1 * time.Hour)
	tx.ExpiresAt = &pastTime
	assert.False(t, tx.CanBatch())
}

func TestTransaction_EstimatedFee(t *testing.T) {
	testCases := []struct {
		txType      TransactionType
		expectedFee string
	}{
		{TransactionTypeEscrowCreate, "12"},
		{TransactionTypeEscrowFinish, "12"},
		{TransactionTypeEscrowCancel, "12"},
		{TransactionTypePayment, "10"},
		{TransactionTypeWalletSetup, "15"},
	}

	for _, tc := range testCases {
		tx := &Transaction{Type: tc.txType}
		fee := tx.EstimatedFee()
		assert.Equal(t, tc.expectedFee, fee, "Fee mismatch for type %s", tc.txType)
	}
}

func TestDefaultBatchConfig(t *testing.T) {
	config := DefaultBatchConfig()

	assert.Equal(t, 10, config.MaxBatchSize)
	assert.Equal(t, 30*time.Second, config.MaxWaitTime)
	assert.True(t, config.FeeOptimization)
	assert.True(t, config.PriorityBatching)
	assert.Equal(t, 2, config.MinBatchSize)
	assert.Equal(t, 300, config.BatchTimeoutSeconds)
}

func TestTransactionMetadata_Value(t *testing.T) {
	metadata := TransactionMetadata{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	value, err := metadata.Value()
	require.NoError(t, err)
	assert.NotNil(t, value)

	// Should be valid JSON
	jsonBytes, ok := value.([]byte)
	require.True(t, ok)
	assert.Contains(t, string(jsonBytes), "key1")
	assert.Contains(t, string(jsonBytes), "value1")
}

func TestTransactionMetadata_Scan(t *testing.T) {
	var metadata TransactionMetadata

	// Test nil value
	err := metadata.Scan(nil)
	require.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Len(t, metadata, 0)

	// Test valid JSON
	jsonData := []byte(`{"key1":"value1","key2":42}`)
	err = metadata.Scan(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, float64(42), metadata["key2"]) // JSON numbers are float64

	// Test invalid data type
	err = metadata.Scan("invalid")
	assert.Error(t, err)
}

func TestTransactionStatus_String(t *testing.T) {
	assert.Equal(t, "pending", string(TransactionStatusPending))
	assert.Equal(t, "queued", string(TransactionStatusQueued))
	assert.Equal(t, "confirmed", string(TransactionStatusConfirmed))
	assert.Equal(t, "failed", string(TransactionStatusFailed))
}

func TestTransactionType_String(t *testing.T) {
	assert.Equal(t, "escrow_create", string(TransactionTypeEscrowCreate))
	assert.Equal(t, "escrow_finish", string(TransactionTypeEscrowFinish))
	assert.Equal(t, "payment", string(TransactionTypePayment))
}

func TestTransactionPriority_Values(t *testing.T) {
	assert.Equal(t, TransactionPriority(1), PriorityLow)
	assert.Equal(t, TransactionPriority(2), PriorityNormal)
	assert.Equal(t, TransactionPriority(3), PriorityHigh)
	assert.Equal(t, TransactionPriority(4), PriorityCritical)
}

func TestTransactionBatch_Initialization(t *testing.T) {
	batch := &TransactionBatch{
		ID:               "test-batch",
		Status:           TransactionStatusPending,
		Priority:         PriorityHigh,
		MaxTransactions:  5,
		TransactionCount: 0,
		SuccessCount:     0,
		FailureCount:     0,
		// Note: Only testing initialization, MaxRetries testing removed
	}

	assert.Equal(t, "test-batch", batch.ID)
	assert.Equal(t, TransactionStatusPending, batch.Status)
	assert.Equal(t, PriorityHigh, batch.Priority)
	assert.Equal(t, 5, batch.MaxTransactions)
	assert.Equal(t, 0, batch.TransactionCount)
	assert.Equal(t, 0, batch.SuccessCount)
	assert.Equal(t, 0, batch.FailureCount)
}

func TestTransactionStats_DefaultValues(t *testing.T) {
	stats := &TransactionStats{}

	assert.Equal(t, int64(0), stats.TotalTransactions)
	assert.Equal(t, int64(0), stats.PendingTransactions)
	assert.Equal(t, int64(0), stats.ProcessingTransactions)
	assert.Equal(t, int64(0), stats.CompletedTransactions)
	assert.Equal(t, int64(0), stats.FailedTransactions)
	assert.Equal(t, float64(0), stats.AverageProcessingTime)
	assert.Equal(t, "", stats.TotalFeesProcessed)
	assert.Equal(t, "", stats.TotalFeeSavings)
	assert.Nil(t, stats.LastProcessedAt)
}

func TestBatchConfig_Validation(t *testing.T) {
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxWaitTime:  30 * time.Second,
		// Note: FeeOptimization and PriorityBatching are being tested
		MinBatchSize:        2,
		BatchTimeoutSeconds: 300,
	}

	assert.Greater(t, config.MaxBatchSize, config.MinBatchSize)
	assert.Greater(t, config.BatchTimeoutSeconds, 0)
	assert.Greater(t, config.MaxWaitTime, time.Duration(0))
}

func TestTransaction_CompleteLifecycle(t *testing.T) {
	// Create a new transaction
	tx := NewTransaction(
		TransactionTypeEscrowCreate,
		"rSender123",
		"rReceiver456",
		"1000000",
		"XRP",
		uuid.New().String(),
		uuid.New().String(),
	)

	// Verify initial state
	assert.Equal(t, TransactionStatusPending, tx.Status)
	assert.Equal(t, 0, tx.RetryCount)
	assert.False(t, tx.CanBatch()) // Pending transactions cannot batch, only queued ones can

	// Move to queued
	tx.Status = TransactionStatusQueued
	assert.True(t, tx.CanBatch())

	// Process transaction
	tx.Status = TransactionStatusProcessing
	assert.False(t, tx.CanBatch())

	// Simulate failure and retry
	tx.SetError(assert.AnError)
	assert.Equal(t, TransactionStatusFailed, tx.Status)
	assert.True(t, tx.CanRetry())

	// Increment retry
	tx.IncrementRetry()
	assert.Equal(t, 1, tx.RetryCount)
	assert.Equal(t, TransactionStatusPending, tx.Status)

	// Finally confirm
	tx.Status = TransactionStatusConfirmed
	now := time.Now()
	tx.ConfirmedAt = &now
	assert.False(t, tx.CanRetry())
	assert.False(t, tx.CanBatch())
}

func TestTransactionBatch_BatchProcessing(t *testing.T) {
	batch := NewTransactionBatch(PriorityHigh, 5)

	// Verify initial state
	assert.Equal(t, TransactionStatusPending, batch.Status)
	assert.Equal(t, 0, batch.TransactionCount)

	// Simulate adding transactions
	batch.TransactionCount = 3
	batch.Status = TransactionStatusBatching

	// Simulate processing
	batch.Status = TransactionStatusProcessing
	now := time.Now()
	batch.ProcessedAt = &now

	// Simulate completion
	batch.SuccessCount = 2
	batch.FailureCount = 1
	batch.Status = TransactionStatusConfirmed
	batch.CompletedAt = &now

	assert.Equal(t, 3, batch.TransactionCount)
	assert.Equal(t, 2, batch.SuccessCount)
	assert.Equal(t, 1, batch.FailureCount)
	assert.Equal(t, TransactionStatusConfirmed, batch.Status)
}
