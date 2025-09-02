package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestSmartChequeRepository_GetSmartChequeAuditTrail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test with empty smart check ID
	_, err = repo.GetSmartChequeAuditTrail(context.Background(), "", 10, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smart check ID is required")

	// Test successful query
	smartChequeID := "test-id"
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "enterprise_id", "action", "resource", "resource_id",
		"details", "ip_address", "user_agent", "success", "created_at",
	}).AddRow(
		uuid.New(),
		uuid.New(),
		uuid.New(),
		"create",
		"smart_check",
		smartChequeID,
		"Created smart check",
		"127.0.0.1",
		"test-agent",
		true,
		time.Now(),
	)

	mock.ExpectQuery("SELECT id, user_id, enterprise_id, action, resource, resource_id, details, ip_address, user_agent, success, created_at FROM audit_logs WHERE resource = 'smart_cheque' AND resource_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs(smartChequeID, 10, 0).
		WillReturnRows(rows)

	auditLogs, err := repo.GetSmartChequeAuditTrail(context.Background(), smartChequeID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, auditLogs, 1)
	assert.Equal(t, smartChequeID, *auditLogs[0].ResourceID)
	assert.Equal(t, "smart_check", auditLogs[0].Resource)
	assert.Equal(t, "create", auditLogs[0].Action)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequeComplianceReport(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test with empty smart check ID
	_, err = repo.GetSmartChequeComplianceReport(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smart check ID is required")

	// Test successful query
	smartChequeID := "test-id"

	// Mock transaction count query
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) as total_transactions, COUNT\\(CASE WHEN status = 'completed' THEN 1 END\\) as compliant_tx_count, COUNT\\(CASE WHEN status IN \\('failed', 'cancelled', 'rejected'\\) THEN 1 END\\) as non_compliant_tx_count FROM transactions WHERE smart_check_id = \\$1").
		WithArgs(smartChequeID).
		WillReturnRows(sqlmock.NewRows([]string{"total_transactions", "compliant_tx_count", "non_compliant_tx_count"}).
			AddRow(100, 95, 5))

	// Mock audit log date query
	now := time.Now()
	mock.ExpectQuery("SELECT MAX\\(created_at\\) as last_audit_date FROM audit_logs WHERE resource = 'smart_cheque' AND resource_id = \\$1").
		WithArgs(smartChequeID).
		WillReturnRows(sqlmock.NewRows([]string{"last_audit_date"}).
			AddRow(now))

	report, err := repo.GetSmartChequeComplianceReport(context.Background(), smartChequeID)
	require.NoError(t, err)
	assert.Equal(t, smartChequeID, report.SmartChequeID)
	assert.Equal(t, int64(100), report.TotalTransactions)
	assert.Equal(t, int64(95), report.CompliantTxCount)
	assert.Equal(t, int64(5), report.NonCompliantTxCount)
	assert.InDelta(t, 0.95, report.ComplianceRate, 0.01)
	assert.Equal(t, "compliant", report.RegulatoryStatus)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequeAnalyticsByPayer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test with empty payer ID
	_, err = repo.GetSmartChequeAnalyticsByPayer(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payer ID is required")

	// Test successful query
	payerID := "payer1"

	// Mock count by status query
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM smart_checks WHERE payer_id = \\$1 GROUP BY status").
		WithArgs(payerID).
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("created", 5).
			AddRow("in_progress", 3).
			AddRow("completed", 2))

	// Mock count by currency query
	mock.ExpectQuery("SELECT currency, COUNT\\(\\*\\) FROM smart_checks WHERE payer_id = \\$1 GROUP BY currency").
		WithArgs(payerID).
		WillReturnRows(sqlmock.NewRows([]string{"currency", "count"}).
			AddRow("USDT", 6).
			AddRow("USDC", 4))

	// Mock amount statistics query
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) as total_amount, COALESCE\\(AVG\\(amount\\), 0\\) as average_amount, COALESCE\\(MAX\\(amount\\), 0\\) as largest_amount, COALESCE\\(MIN\\(amount\\), 0\\) as smallest_amount FROM smart_checks WHERE payer_id = \\$1").
		WithArgs(payerID).
		WillReturnRows(sqlmock.NewRows([]string{"total_amount", "average_amount", "largest_amount", "smallest_amount"}).
			AddRow(10000.0, 1000.0, 5000.0, 100.0))

	// Mock recent activity query (GetSmartChequesByPayer)
	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency, milestones, escrow_address, status, contract_hash, created_at, updated_at FROM smart_checks WHERE payer_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs(payerID, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payer_id", "payee_id", "amount", "currency", "milestones", "escrow_address", "status", "contract_hash", "created_at", "updated_at"}).
			AddRow("1", payerID, "payee1", 1000.0, "USDT", []byte("[]"), "", "created", "", time.Now(), time.Now()).
			AddRow("2", payerID, "payee2", 2000.0, "USDC", []byte("[]"), "", "in_progress", "", time.Now(), time.Now()))

	// Mock trends query
	mock.ExpectQuery("SELECT DATE\\(created_at\\) as creation_date, COUNT\\(\\*\\) as count FROM smart_checks WHERE payer_id = \\$1 AND created_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY DATE\\(created_at\\) ORDER BY creation_date").
		WithArgs(payerID).
		WillReturnRows(sqlmock.NewRows([]string{"creation_date", "count"}).
			AddRow("2023-01-01", 5).
			AddRow("2023-01-02", 3))

	analytics, err := repo.GetSmartChequeAnalyticsByPayer(context.Background(), payerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10), analytics.TotalCount) // 5 + 3 + 2
	assert.Len(t, analytics.CountByStatus, 3)
	assert.Len(t, analytics.CountByCurrency, 2)
	assert.Equal(t, 10000.0, analytics.TotalAmount)
	assert.Equal(t, 1000.0, analytics.AverageAmount)
	assert.Len(t, analytics.RecentActivity, 2)
	assert.Len(t, analytics.StatusTrends, 2)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequePerformanceMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test successful query with no filters
	mock.ExpectQuery("SELECT COALESCE\\(AVG\\(EXTRACT\\(EPOCH FROM \\(t.updated_at - t.created_at\\)\\)\\), 0\\) as avg_processing_time, COALESCE\\(SUM\\(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END\\) \\* 1.0 / COUNT\\(\\*\\), 1\\) as success_rate, COALESCE\\(SUM\\(CASE WHEN t.status IN \\('failed', 'cancelled', 'rejected'\\) THEN 1 ELSE 0 END\\) \\* 1.0 / COUNT\\(\\*\\), 0\\) as failure_rate, COALESCE\\(AVG\\(s.amount\\), 0\\) as average_amount, COALESCE\\(SUM\\(s.amount\\), 0\\) as total_volume, COALESCE\\(MAX\\(s.amount\\), 0\\) as peak_hour_volume FROM smart_checks s LEFT JOIN transactions t ON s.id = t.smart_check_id WHERE 1=1").
		WillReturnRows(sqlmock.NewRows([]string{"avg_processing_time", "success_rate", "failure_rate", "average_amount", "total_volume", "peak_hour_volume"}).
			AddRow(30.0, 0.95, 0.05, 1000.0, 100000.0, 5000.0))

	metrics, err := repo.GetSmartChequePerformanceMetrics(context.Background(), nil)
	require.NoError(t, err)
	assert.NotZero(t, metrics.AverageProcessingTime)
	assert.InDelta(t, 0.95, metrics.SuccessRate, 0.01)
	assert.InDelta(t, 0.05, metrics.FailureRate, 0.01)
	assert.Equal(t, 1000.0, metrics.AverageAmount)
	assert.Equal(t, 100000.0, metrics.TotalVolume)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequePerformanceMetrics_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test with filters
	payerID := "payer1"
	status := models.SmartChequeStatusCompleted
	currency := models.CurrencyUSDT
	dateFrom := time.Now().Add(-24 * time.Hour)
	dateTo := time.Now()
	minAmount := 100.0
	maxAmount := 10000.0
	contractHash := "contract-hash"

	filters := &SmartChequeFilter{
		PayerID:      &payerID,
		Status:       &status,
		Currency:     &currency,
		DateFrom:     &dateFrom,
		DateTo:       &dateTo,
		MinAmount:    &minAmount,
		MaxAmount:    &maxAmount,
		ContractHash: &contractHash,
	}

	mock.ExpectQuery("SELECT COALESCE\\(AVG\\(EXTRACT\\(EPOCH FROM \\(t.updated_at - t.created_at\\)\\)\\), 0\\) as avg_processing_time, COALESCE\\(SUM\\(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END\\) \\* 1.0 / COUNT\\(\\*\\), 1\\) as success_rate, COALESCE\\(SUM\\(CASE WHEN t.status IN \\('failed', 'cancelled', 'rejected'\\) THEN 1 ELSE 0 END\\) \\* 1.0 / COUNT\\(\\*\\), 0\\) as failure_rate, COALESCE\\(AVG\\(s.amount\\), 0\\) as average_amount, COALESCE\\(SUM\\(s.amount\\), 0\\) as total_volume, COALESCE\\(MAX\\(s.amount\\), 0\\) as peak_hour_volume FROM smart_checks s LEFT JOIN transactions t ON s.id = t.smart_check_id WHERE 1=1 AND s.payer_id = \\$1 AND s.status = \\$2 AND s.currency = \\$3 AND s.created_at >= \\$4 AND s.created_at <= \\$5 AND s.amount >= \\$6 AND s.amount <= \\$7 AND s.contract_hash = \\$8").
		WithArgs(payerID, string(status), string(currency), dateFrom, dateTo, minAmount, maxAmount, contractHash).
		WillReturnRows(sqlmock.NewRows([]string{"avg_processing_time", "success_rate", "failure_rate", "average_amount", "total_volume", "peak_hour_volume"}).
			AddRow(25.0, 0.98, 0.02, 1500.0, 150000.0, 7500.0))

	metrics, err := repo.GetSmartChequePerformanceMetrics(context.Background(), filters)
	require.NoError(t, err)
	assert.NotZero(t, metrics.AverageProcessingTime)
	assert.InDelta(t, 0.98, metrics.SuccessRate, 0.01)
	assert.InDelta(t, 0.02, metrics.FailureRate, 0.01)
	assert.Equal(t, 1500.0, metrics.AverageAmount)
	assert.Equal(t, 150000.0, metrics.TotalVolume)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_UpdateSmartChequeStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	updates := map[string]models.SmartChequeStatus{
		"id1": models.SmartChequeStatusCompleted,
		"id2": models.SmartChequeStatusDisputed,
	}

	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE smart_checks SET status = \\$1, updated_at = \\$2 WHERE id = \\$3")
	mock.ExpectExec("UPDATE smart_check.*").
		WithArgs(string(models.SmartChequeStatusCompleted), sqlmock.AnyArg(), "id1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE smart_check.*").
		WithArgs(string(models.SmartChequeStatusDisputed), sqlmock.AnyArg(), "id2").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.BatchUpdateSmartChequeStatuses(context.Background(), updates)
	require.NoError(t, err)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}
