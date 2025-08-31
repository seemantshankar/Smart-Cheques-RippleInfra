package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartChequeRepository_CreateSmartCheque(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	smartCheque := &models.SmartCheque{
		ID:            uuid.New().String(),
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        1000.0,
		Currency:      models.CurrencyUSDT,
		Milestones:    []models.Milestone{},
		EscrowAddress: "",
		Status:        models.SmartChequeStatusCreated,
		ContractHash:  "test-contract-hash",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mock.ExpectExec("INSERT INTO smart_cheques").
		WithArgs(
			smartCheque.ID,
			smartCheque.PayerID,
			smartCheque.PayeeID,
			smartCheque.Amount,
			string(smartCheque.Currency),
			sqlmock.AnyArg(), // milestones JSON
			smartCheque.EscrowAddress,
			string(smartCheque.Status),
			smartCheque.ContractHash,
			smartCheque.CreatedAt,
			smartCheque.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateSmartCheque(context.Background(), smartCheque)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequeByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	smartChequeID := uuid.New().String()
	expected := &models.SmartCheque{
		ID:            smartChequeID,
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        1000.0,
		Currency:      models.CurrencyUSDT,
		Milestones:    []models.Milestone{},
		EscrowAddress: "test-escrow-address",
		Status:        models.SmartChequeStatusLocked,
		ContractHash:  "test-contract-hash",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Access the Milestones field to resolve the unused write error
	_ = expected.Milestones

	rows := sqlmock.NewRows([]string{
		"id", "payer_id", "payee_id", "amount", "currency",
		"milestones", "escrow_address", "status", "contract_hash",
		"created_at", "updated_at",
	}).AddRow(
		expected.ID,
		expected.PayerID,
		expected.PayeeID,
		expected.Amount,
		string(expected.Currency),
		"[]", // milestones JSON
		expected.EscrowAddress,
		string(expected.Status),
		expected.ContractHash,
		expected.CreatedAt,
		expected.UpdatedAt,
	)

	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency").
		WithArgs(smartChequeID).
		WillReturnRows(rows)

	result, err := repo.GetSmartChequeByID(context.Background(), smartChequeID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.PayerID, result.PayerID)
	assert.Equal(t, expected.PayeeID, result.PayeeID)
	assert.Equal(t, expected.Amount, result.Amount)
	assert.Equal(t, expected.Currency, result.Currency)
	assert.Equal(t, expected.EscrowAddress, result.EscrowAddress)
	assert.Equal(t, expected.Status, result.Status)
	assert.Equal(t, expected.ContractHash, result.ContractHash)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequeByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	smartChequeID := uuid.New().String()

	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency").
		WithArgs(smartChequeID).
		WillReturnError(sqlmock.ErrCancelled)

	result, err := repo.GetSmartChequeByID(context.Background(), smartChequeID)
	assert.Error(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_UpdateSmartCheque(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	smartCheque := &models.SmartCheque{
		ID:            uuid.New().String(),
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        1500.0,
		Currency:      models.CurrencyUSDC,
		Milestones:    []models.Milestone{},
		EscrowAddress: "updated-escrow-address",
		Status:        models.SmartChequeStatusInProgress,
		ContractHash:  "updated-contract-hash",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mock.ExpectExec("UPDATE smart_cheques").
		WithArgs(
			smartCheque.PayerID,
			smartCheque.PayeeID,
			smartCheque.Amount,
			string(smartCheque.Currency),
			sqlmock.AnyArg(), // milestones JSON
			smartCheque.EscrowAddress,
			string(smartCheque.Status),
			smartCheque.ContractHash,
			smartCheque.UpdatedAt,
			smartCheque.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateSmartCheque(context.Background(), smartCheque)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_DeleteSmartCheque(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	smartChequeID := uuid.New().String()

	mock.ExpectExec("DELETE FROM smart_cheques").
		WithArgs(smartChequeID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeleteSmartCheque(context.Background(), smartChequeID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequesByPayer(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	payerID := uuid.New().String()

	rows := sqlmock.NewRows([]string{
		"id", "payer_id", "payee_id", "amount", "currency",
		"milestones", "escrow_address", "status", "contract_hash",
		"created_at", "updated_at",
	}).AddRow(
		uuid.New().String(),
		payerID,
		uuid.New().String(),
		1000.0,
		string(models.CurrencyUSDT),
		"[]", // milestones JSON
		"test-escrow-address-1",
		string(models.SmartChequeStatusCreated),
		"test-contract-hash-1",
		time.Now(),
		time.Now(),
	).AddRow(
		uuid.New().String(),
		payerID,
		uuid.New().String(),
		2000.0,
		string(models.CurrencyUSDC),
		"[]", // milestones JSON
		"test-escrow-address-2",
		string(models.SmartChequeStatusLocked),
		"test-contract-hash-2",
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery("SELECT id, payer_id, payee_id, amount, currency").
		WithArgs(payerID, 10, 0).
		WillReturnRows(rows)

	result, err := repo.GetSmartChequesByPayer(context.Background(), payerID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSmartChequeRepository_GetSmartChequeCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSmartChequeRepository(db)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM smart_cheques").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.GetSmartChequeCount(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}
