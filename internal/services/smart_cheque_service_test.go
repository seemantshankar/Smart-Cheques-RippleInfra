package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSmartChequeRepository implements the SmartChequeRepositoryInterface for testing
type mockSmartChequeRepositorySmartCheque struct {
	mock.Mock
}

func (m *mockSmartChequeRepositorySmartCheque) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepositorySmartCheque) DeleteSmartCheque(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payerID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payeeID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequeCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSmartChequeRepositorySmartCheque) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.SmartChequeStatus]int64), args.Error(1)
}

func TestSmartChequeService_CreateSmartCheque(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	request := &CreateSmartChequeRequest{
		PayerID:      uuid.New().String(),
		PayeeID:      uuid.New().String(),
		Amount:       1000.0,
		Currency:     models.CurrencyUSDT,
		Milestones:   []models.Milestone{},
		ContractHash: "test-contract-hash",
	}

	mockRepo.On("CreateSmartCheque", mock.Anything, mock.AnythingOfType("*models.SmartCheque")).Return(nil)

	smartCheque, err := service.CreateSmartCheque(context.Background(), request)
	assert.NoError(t, err)
	assert.NotNil(t, smartCheque)
	assert.Equal(t, request.PayerID, smartCheque.PayerID)
	assert.Equal(t, request.PayeeID, smartCheque.PayeeID)
	assert.Equal(t, request.Amount, smartCheque.Amount)
	assert.Equal(t, request.Currency, smartCheque.Currency)
	assert.Equal(t, models.SmartChequeStatusCreated, smartCheque.Status)

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_CreateSmartCheque_ValidationErrors(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	// Test missing payer ID
	request := &CreateSmartChequeRequest{
		PayeeID:      uuid.New().String(),
		Amount:       1000.0,
		Currency:     models.CurrencyUSDT,
		Milestones:   []models.Milestone{},
		ContractHash: "test-contract-hash",
	}

	_, err := service.CreateSmartCheque(context.Background(), request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payer_id is required")

	// Test missing payee ID
	request = &CreateSmartChequeRequest{
		PayerID:      uuid.New().String(),
		Amount:       1000.0,
		Currency:     models.CurrencyUSDT,
		Milestones:   []models.Milestone{},
		ContractHash: "test-contract-hash",
	}

	_, err = service.CreateSmartCheque(context.Background(), request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payee_id is required")

	// Test invalid amount
	request = &CreateSmartChequeRequest{
		PayerID:      uuid.New().String(),
		PayeeID:      uuid.New().String(),
		Amount:       0,
		Currency:     models.CurrencyUSDT,
		Milestones:   []models.Milestone{},
		ContractHash: "test-contract-hash",
	}

	_, err = service.CreateSmartCheque(context.Background(), request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than 0")

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_GetSmartCheque(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	smartChequeID := uuid.New().String()
	expected := &models.SmartCheque{
		ID:            smartChequeID,
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

	mockRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(expected, nil)

	result, err := service.GetSmartCheque(context.Background(), smartChequeID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_UpdateSmartCheque(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	smartChequeID := uuid.New().String()
	existing := &models.SmartCheque{
		ID:            smartChequeID,
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

	updatedEscrowAddress := "updated-escrow-address"
	request := &UpdateSmartChequeRequest{
		EscrowAddress: &updatedEscrowAddress,
		Status:        &[]models.SmartChequeStatus{models.SmartChequeStatusLocked}[0],
	}

	mockRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(existing, nil)
	mockRepo.On("UpdateSmartCheque", mock.Anything, mock.AnythingOfType("*models.SmartCheque")).Return(nil)

	result, err := service.UpdateSmartCheque(context.Background(), smartChequeID, request)
	assert.NoError(t, err)
	assert.Equal(t, updatedEscrowAddress, result.EscrowAddress)
	assert.Equal(t, models.SmartChequeStatusLocked, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_DeleteSmartCheque(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	smartChequeID := uuid.New().String()
	existing := &models.SmartCheque{
		ID: smartChequeID,
	}

	mockRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(existing, nil)
	mockRepo.On("DeleteSmartCheque", mock.Anything, smartChequeID).Return(nil)

	err := service.DeleteSmartCheque(context.Background(), smartChequeID)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_ListSmartChequesByPayer(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	payerID := uuid.New().String()
	expected := []*models.SmartCheque{
		{
			ID:            uuid.New().String(),
			PayerID:       payerID,
			PayeeID:       uuid.New().String(),
			Amount:        1000.0,
			Currency:      models.CurrencyUSDT,
			Milestones:    []models.Milestone{},
			EscrowAddress: "",
			Status:        models.SmartChequeStatusCreated,
			ContractHash:  "test-contract-hash-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            uuid.New().String(),
			PayerID:       payerID,
			PayeeID:       uuid.New().String(),
			Amount:        2000.0,
			Currency:      models.CurrencyUSDC,
			Milestones:    []models.Milestone{},
			EscrowAddress: "",
			Status:        models.SmartChequeStatusLocked,
			ContractHash:  "test-contract-hash-2",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("GetSmartChequesByPayer", mock.Anything, payerID, 10, 0).Return(expected, nil)

	result, err := service.ListSmartChequesByPayer(context.Background(), payerID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_GetSmartChequeStatistics(t *testing.T) {
	mockRepo := &mockSmartChequeRepositorySmartCheque{}
	service := NewSmartChequeService(mockRepo)

	expectedCount := int64(10)
	expectedCountByStatus := map[models.SmartChequeStatus]int64{
		models.SmartChequeStatusCreated:    3,
		models.SmartChequeStatusLocked:     4,
		models.SmartChequeStatusInProgress: 2,
		models.SmartChequeStatusCompleted:  1,
	}

	mockRepo.On("GetSmartChequeCount", mock.Anything).Return(expectedCount, nil)
	mockRepo.On("GetSmartChequeCountByStatus", mock.Anything).Return(expectedCountByStatus, nil)

	result, err := service.GetSmartChequeStatistics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result.TotalCount)
	assert.Equal(t, expectedCountByStatus, result.CountByStatus)

	mockRepo.AssertExpectations(t)
}
