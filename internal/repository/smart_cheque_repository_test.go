package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestSmartChequeRepository_GetSmartChequeAnalyticsByPayee(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Mock the queries that GetSmartChequeAnalyticsByPayee will make
	mock.ExpectQuery("SELECT status, COUNT\\(\\*\\) FROM smart_cheques WHERE payee_id = \\$1 GROUP BY status").
		WithArgs("payee1").
		WillReturnRows(sqlmock.NewRows([]string{"status", "count"}).
			AddRow("created", 5).
			AddRow("in_progress", 3).
			AddRow("completed", 2))

	mock.ExpectQuery("SELECT currency, COUNT\\(\\*\\) FROM smart_cheques WHERE payee_id = \\$1 GROUP BY currency").
		WithArgs("payee1").
		WillReturnRows(sqlmock.NewRows([]string{"currency", "count"}).
			AddRow("USDT", 6).
			AddRow("USDC", 4))

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\), 0\\) as total_amount, COALESCE\\(AVG\\(amount\\), 0\\) as average_amount, COALESCE\\(MAX\\(amount\\), 0\\) as largest_amount, COALESCE\\(MIN\\(amount\\), 0\\) as smallest_amount FROM smart_cheques WHERE payee_id = \\$1").
		WithArgs("payee1").
		WillReturnRows(sqlmock.NewRows([]string{"total_amount", "average_amount", "largest_amount", "smallest_amount"}).
			AddRow(10000.0, 1000.0, 5000.0, 100.0))

	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency, milestones, escrow_address, status, contract_hash, created_at, updated_at FROM smart_cheques WHERE payee_id = \\$1 ORDER BY created_at DESC LIMIT \\$2 OFFSET \\$3").
		WithArgs("payee1", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payer_id", "payee_id", "amount", "currency", "milestones", "escrow_address", "status", "contract_hash", "created_at", "updated_at"}).
			AddRow("1", "payer1", "payee1", 1000.0, "USDT", []byte("[]"), "", "created", "", time.Now(), time.Now()).
			AddRow("2", "payer1", "payee1", 2000.0, "USDC", []byte("[]"), "", "in_progress", "", time.Now(), time.Now()))

	mock.ExpectQuery("SELECT DATE\\(created_at\\) as creation_date, COUNT\\(\\*\\) as count FROM smart_cheques WHERE payee_id = \\$1 AND created_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY DATE\\(created_at\\) ORDER BY creation_date").
		WithArgs("payee1").
		WillReturnRows(sqlmock.NewRows([]string{"creation_date", "count"}).
			AddRow("2023-01-01", 5).
			AddRow("2023-01-02", 3))

	// Test GetSmartChequeAnalyticsByPayee
	analytics, err := repo.GetSmartChequeAnalyticsByPayee(context.Background(), "payee1")
	require.NoError(t, err)
	assert.NotNil(t, analytics)
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

func TestSmartChequeRepository_BatchGetSmartCheques(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test BatchGetSmartCheques with empty IDs
	cheques, err := repo.BatchGetSmartCheques(context.Background(), []string{})
	require.NoError(t, err)
	assert.NotNil(t, cheques)
	assert.Len(t, cheques, 0)

	// Test BatchGetSmartCheques with IDs
	now := time.Now()
	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency.*").
		WithArgs("id1", "id2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "payer_id", "payee_id", "amount", "currency", "milestones", "escrow_address", "status", "contract_hash", "created_at", "updated_at"}).
			AddRow("id1", "payer1", "payee1", 1000.0, "USDT", []byte("[]"), "", "created", "", now, now).
			AddRow("id2", "payer2", "payee2", 2000.0, "USDC", []byte("[]"), "", "in_progress", "", now, now))

	cheques, err = repo.BatchGetSmartCheques(context.Background(), []string{"id1", "id2"})
	require.NoError(t, err)
	assert.NotNil(t, cheques)
	assert.Len(t, cheques, 2)
	assert.Equal(t, "id1", cheques[0].ID)
	assert.Equal(t, "id2", cheques[1].ID)
}

func TestSmartChequeRepository_BatchUpdateSmartChequeStatuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	// Test BatchUpdateSmartChequeStatuses with empty updates
	err = repo.BatchUpdateSmartChequeStatuses(context.Background(), map[string]models.SmartChequeStatus{})
	require.NoError(t, err)

	// Test BatchUpdateSmartChequeStatuses with updates
	mock.ExpectBegin()
	mock.ExpectPrepare("UPDATE smart_cheques SET status = \\$1, updated_at = \\$2 WHERE id = \\$3")
	mock.ExpectExec("UPDATE smart_cheque.*").
		WithArgs("completed", sqlmock.AnyArg(), "id1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE smart_cheque.*").
		WithArgs(string(models.SmartChequeStatusDisputed), sqlmock.AnyArg(), "id2").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	updates := map[string]models.SmartChequeStatus{
		"id1": models.SmartChequeStatusCompleted,
		"id2": models.SmartChequeStatusDisputed, // This might not exist, let's use a valid status
	}

	err = repo.BatchUpdateSmartChequeStatuses(context.Background(), updates)
	require.NoError(t, err)

	// Ensure all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}
