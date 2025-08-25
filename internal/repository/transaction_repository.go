package repository

import (
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"gorm.io/gorm"
)

// TransactionRepository implements TransactionRepositoryInterface
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

// CreateTransaction creates a new transaction record
func (r *TransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	if err := r.db.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

// GetTransactionByID retrieves a transaction by its ID
func (r *TransactionRepository) GetTransactionByID(id string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.Where("id = ?", id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &transaction, nil
}

// UpdateTransaction updates an existing transaction
func (r *TransactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	transaction.UpdatedAt = time.Now()
	if err := r.db.Save(transaction).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	return nil
}

// DeleteTransaction deletes a transaction by ID
func (r *TransactionRepository) DeleteTransaction(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.Transaction{}).Error; err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	return nil
}

// GetTransactionsByStatus retrieves transactions by status with pagination
func (r *TransactionRepository) GetTransactionsByStatus(status models.TransactionStatus, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := r.db.Where("status = ?", status).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by status: %w", err)
	}
	return transactions, nil
}

// GetTransactionsByBatchID retrieves all transactions in a specific batch
func (r *TransactionRepository) GetTransactionsByBatchID(batchID string) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := r.db.Where("batch_id = ?", batchID).Order("created_at ASC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by batch ID: %w", err)
	}
	return transactions, nil
}

// GetTransactionsByEnterpriseID retrieves transactions for a specific enterprise
func (r *TransactionRepository) GetTransactionsByEnterpriseID(enterpriseID string, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := r.db.Where("enterprise_id = ?", enterpriseID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by enterprise ID: %w", err)
	}
	return transactions, nil
}

// GetTransactionsByUserID retrieves transactions for a specific user
func (r *TransactionRepository) GetTransactionsByUserID(userID string, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by user ID: %w", err)
	}
	return transactions, nil
}

// GetTransactionsByType retrieves transactions by type
func (r *TransactionRepository) GetTransactionsByType(txType models.TransactionType, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := r.db.Where("type = ?", txType).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transactions by type: %w", err)
	}
	return transactions, nil
}

// GetPendingTransactions retrieves pending transactions ready for processing
func (r *TransactionRepository) GetPendingTransactions(limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	query := r.db.Where("status IN ?", []models.TransactionStatus{
		models.TransactionStatusPending,
		models.TransactionStatusQueued,
	}).Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending transactions: %w", err)
	}
	return transactions, nil
}

// GetExpiredTransactions retrieves transactions that have expired
func (r *TransactionRepository) GetExpiredTransactions() ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := r.db.Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Where("status NOT IN ?", []models.TransactionStatus{
			models.TransactionStatusConfirmed,
			models.TransactionStatusCancelled,
			models.TransactionStatusExpired,
		}).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired transactions: %w", err)
	}
	return transactions, nil
}

// GetRetriableTransactions retrieves transactions that can be retried
func (r *TransactionRepository) GetRetriableTransactions() ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	if err := r.db.Where("status = ? AND retry_count < max_retries", models.TransactionStatusFailed).
		Order("priority DESC, created_at ASC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get retriable transactions: %w", err)
	}
	return transactions, nil
}

// CreateTransactionBatch creates a new transaction batch
func (r *TransactionRepository) CreateTransactionBatch(batch *models.TransactionBatch) error {
	if err := r.db.Create(batch).Error; err != nil {
		return fmt.Errorf("failed to create transaction batch: %w", err)
	}
	return nil
}

// GetTransactionBatchByID retrieves a transaction batch by its ID
func (r *TransactionRepository) GetTransactionBatchByID(id string) (*models.TransactionBatch, error) {
	var batch models.TransactionBatch
	if err := r.db.Preload("Transactions").Where("id = ?", id).First(&batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction batch not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get transaction batch: %w", err)
	}
	return &batch, nil
}

// UpdateTransactionBatch updates an existing transaction batch
func (r *TransactionRepository) UpdateTransactionBatch(batch *models.TransactionBatch) error {
	batch.UpdatedAt = time.Now()
	if err := r.db.Save(batch).Error; err != nil {
		return fmt.Errorf("failed to update transaction batch: %w", err)
	}
	return nil
}

// DeleteTransactionBatch deletes a transaction batch by ID
func (r *TransactionRepository) DeleteTransactionBatch(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&models.TransactionBatch{}).Error; err != nil {
		return fmt.Errorf("failed to delete transaction batch: %w", err)
	}
	return nil
}

// GetTransactionBatchesByStatus retrieves transaction batches by status
func (r *TransactionRepository) GetTransactionBatchesByStatus(status models.TransactionStatus, limit, offset int) ([]*models.TransactionBatch, error) {
	var batches []*models.TransactionBatch
	query := r.db.Where("status = ?", status).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&batches).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction batches by status: %w", err)
	}
	return batches, nil
}

// GetPendingBatches retrieves pending transaction batches
func (r *TransactionRepository) GetPendingBatches(limit int) ([]*models.TransactionBatch, error) {
	var batches []*models.TransactionBatch
	query := r.db.Where("status IN ?", []models.TransactionStatus{
		models.TransactionStatusPending,
		models.TransactionStatusBatching,
	}).Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&batches).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending batches: %w", err)
	}
	return batches, nil
}

// GetTransactionStats retrieves overall transaction statistics
func (r *TransactionRepository) GetTransactionStats() (*models.TransactionStats, error) {
	var stats models.TransactionStats

	// Count total transactions
	if err := r.db.Model(&models.Transaction{}).Count(&stats.TotalTransactions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total transactions: %w", err)
	}

	// Count by status
	statusCounts, err := r.GetTransactionCountByStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	stats.PendingTransactions = statusCounts[models.TransactionStatusPending] +
		statusCounts[models.TransactionStatusQueued] +
		statusCounts[models.TransactionStatusBatching]
	stats.ProcessingTransactions = statusCounts[models.TransactionStatusProcessing] +
		statusCounts[models.TransactionStatusSubmitted]
	stats.CompletedTransactions = statusCounts[models.TransactionStatusConfirmed]
	stats.FailedTransactions = statusCounts[models.TransactionStatusFailed] +
		statusCounts[models.TransactionStatusCancelled] +
		statusCounts[models.TransactionStatusExpired]

	// Calculate average processing time
	averageTime, err := r.GetAverageProcessingTime()
	if err == nil {
		stats.AverageProcessingTime = averageTime
	}

	// Get last processed transaction timestamp
	var lastTransaction models.Transaction
	if err := r.db.Where("confirmed_at IS NOT NULL").Order("confirmed_at DESC").
		Limit(1).First(&lastTransaction).Error; err == nil {
		stats.LastProcessedAt = lastTransaction.ConfirmedAt
	}

	// Calculate total fees (simplified)
	stats.TotalFeesProcessed = "0" // Would calculate actual fees in production
	stats.TotalFeeSavings = "0"    // Would calculate actual savings in production

	return &stats, nil
}

// GetTransactionStatsByDateRange retrieves transaction statistics for a specific date range
func (r *TransactionRepository) GetTransactionStatsByDateRange(start, end time.Time) (*models.TransactionStats, error) {
	var stats models.TransactionStats

	query := r.db.Model(&models.Transaction{}).Where("created_at BETWEEN ? AND ?", start, end)

	// Count total transactions in date range
	if err := query.Count(&stats.TotalTransactions).Error; err != nil {
		return nil, fmt.Errorf("failed to count transactions in date range: %w", err)
	}

	// Count by status in date range
	var statusResults []struct {
		Status models.TransactionStatus
		Count  int64
	}

	if err := query.Select("status, COUNT(*) as count").Group("status").Scan(&statusResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts for date range: %w", err)
	}

	statusCounts := make(map[models.TransactionStatus]int64)
	for _, result := range statusResults {
		statusCounts[result.Status] = result.Count
	}

	stats.PendingTransactions = statusCounts[models.TransactionStatusPending] +
		statusCounts[models.TransactionStatusQueued] +
		statusCounts[models.TransactionStatusBatching]
	stats.ProcessingTransactions = statusCounts[models.TransactionStatusProcessing] +
		statusCounts[models.TransactionStatusSubmitted]
	stats.CompletedTransactions = statusCounts[models.TransactionStatusConfirmed]
	stats.FailedTransactions = statusCounts[models.TransactionStatusFailed] +
		statusCounts[models.TransactionStatusCancelled] +
		statusCounts[models.TransactionStatusExpired]

	return &stats, nil
}

// GetTransactionCountByStatus retrieves transaction counts grouped by status
func (r *TransactionRepository) GetTransactionCountByStatus() (map[models.TransactionStatus]int64, error) {
	var results []struct {
		Status models.TransactionStatus
		Count  int64
	}

	if err := r.db.Model(&models.Transaction{}).Select("status, COUNT(*) as count").
		Group("status").Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction counts by status: %w", err)
	}

	counts := make(map[models.TransactionStatus]int64)
	for _, result := range results {
		counts[result.Status] = result.Count
	}

	return counts, nil
}

// GetAverageProcessingTime calculates the average processing time for confirmed transactions
func (r *TransactionRepository) GetAverageProcessingTime() (float64, error) {
	var result struct {
		AvgTime float64
	}

	if err := r.db.Model(&models.Transaction{}).
		Select("AVG(EXTRACT(EPOCH FROM (confirmed_at - created_at))) as avg_time").
		Where("confirmed_at IS NOT NULL").Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate average processing time: %w", err)
	}

	return result.AvgTime, nil
}
