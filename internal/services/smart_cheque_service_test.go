package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
)

func TestSmartChequeService_GetSmartChequeAnalytics(t *testing.T) {
	mockRepo := &mocks.SmartChequeRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	service := NewSmartChequeService(mockRepo, mockAuditRepo)

	ctx := context.Background()

	// Set up mock expectations
	countByStatus := map[models.SmartChequeStatus]int64{
		models.SmartChequeStatusCreated:    5,
		models.SmartChequeStatusInProgress: 3,
		models.SmartChequeStatusCompleted:  2,
	}
	mockRepo.On("GetSmartChequeCountByStatus", ctx).Return(countByStatus, nil)

	countByCurrency := map[models.Currency]int64{
		models.CurrencyUSDT: 6,
		models.CurrencyUSDC: 4,
	}
	mockRepo.On("GetSmartChequeCountByCurrency", ctx).Return(countByCurrency, nil)

	mockRepo.On("GetSmartChequeAmountStatistics", ctx).Return(10000.0, 1000.0, 5000.0, 100.0, nil)

	recentActivity := []*models.SmartCheque{
		{ID: "1", Amount: 1000.0, Currency: models.CurrencyUSDT},
		{ID: "2", Amount: 2000.0, Currency: models.CurrencyUSDC},
	}
	mockRepo.On("GetRecentSmartCheques", ctx, 10).Return(recentActivity, nil)

	statusTrends := map[string]int64{
		"2023-01-01": 5,
		"2023-01-02": 3,
	}
	mockRepo.On("GetSmartChequeTrends", ctx, 30).Return(statusTrends, nil)

	// Call the method under test
	analytics, err := service.GetSmartChequeAnalytics(ctx)

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, int64(10), analytics.TotalCount)
	assert.Equal(t, countByStatus, analytics.CountByStatus)
	assert.Equal(t, countByCurrency, analytics.CountByCurrency)
	assert.Equal(t, 1000.0, analytics.AverageAmount)
	assert.Equal(t, 10000.0, analytics.TotalAmount)
	assert.Equal(t, 5000.0, analytics.LargestAmount)
	assert.Equal(t, 100.0, analytics.SmallestAmount)
	assert.Equal(t, recentActivity, analytics.RecentActivity)
	assert.Equal(t, statusTrends, analytics.StatusTrends)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestSmartChequeService_CreateSmartChequeBatch(t *testing.T) {
	mockRepo := &mocks.SmartChequeRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	service := NewSmartChequeService(mockRepo, mockAuditRepo)

	ctx := context.Background()

	// Create test requests
	requests := []*CreateSmartChequeRequest{
		{
			PayerID:  "payer1",
			PayeeID:  "payee1",
			Amount:   1000.0,
			Currency: models.CurrencyUSDT,
			Milestones: []models.Milestone{
				{ID: "m1", Description: "Milestone 1", Amount: 1000.0, VerificationMethod: "manual", Status: models.MilestoneStatusPending},
			},
		},
		{
			PayerID:  "payer2",
			PayeeID:  "payee2",
			Amount:   2000.0,
			Currency: models.CurrencyUSDC,
			Milestones: []models.Milestone{
				{ID: "m2", Description: "Milestone 2", Amount: 2000.0, VerificationMethod: "manual", Status: models.MilestoneStatusPending},
			},
		},
	}

	// Set up mock expectations
	mockRepo.On("BatchCreateSmartCheques", ctx, mock.AnythingOfType("[]*models.SmartCheque")).Return(nil)
	mockAuditRepo.On("CreateAuditLog", mock.AnythingOfType("*models.AuditLog")).Return(nil)

	// Call the method under test
	result, err := service.CreateSmartChequeBatch(ctx, requests)

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, result.Results, 2)
	assert.True(t, result.Results[0].Success)
	assert.True(t, result.Results[1].Success)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

func TestSmartChequeService_UpdateSmartChequeBatch(t *testing.T) {
	mockRepo := &mocks.SmartChequeRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	service := NewSmartChequeService(mockRepo, mockAuditRepo)

	ctx := context.Background()

	// Create test updates
	status := models.SmartChequeStatusInProgress
	updates := map[string]*UpdateSmartChequeRequest{
		"1": {
			Status: &status,
		},
		"2": {
			Status: &status,
		},
	}

	// Set up mock expectations
	mockRepo.On("GetSmartChequeByID", ctx, "1").Return(&models.SmartCheque{
		ID:     "1",
		Status: models.SmartChequeStatusCreated,
	}, nil)
	mockRepo.On("GetSmartChequeByID", ctx, "2").Return(&models.SmartCheque{
		ID:     "2",
		Status: models.SmartChequeStatusCreated,
	}, nil)
	mockRepo.On("BatchUpdateSmartCheques", ctx, mock.AnythingOfType("[]*models.SmartCheque")).Return(nil)
	mockAuditRepo.On("CreateAuditLog", mock.AnythingOfType("*models.AuditLog")).Return(nil)

	// Call the method under test
	result, err := service.UpdateSmartChequeBatch(ctx, updates)

	// Assert results
	require.NoError(t, err)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailureCount)
	assert.Len(t, result.Results, 2)
	assert.True(t, result.Results[0].Success)
	assert.True(t, result.Results[1].Success)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockAuditRepo.AssertExpectations(t)
}

func TestSmartChequeService_CreateAuditLog(t *testing.T) {
	mockRepo := &mocks.SmartChequeRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	service := NewSmartChequeService(mockRepo, mockAuditRepo)

	ctx := context.Background()

	// Create test audit log entry
	entry := &AuditLogEntry{
		SmartChequeID: "test-id",
		Action:        "test_action",
		UserID:        uuid.New().String(),
		Details:       "Test audit log entry",
		Timestamp:     time.Now(),
		IPAddress:     "127.0.0.1",
		UserAgent:     "test-agent",
	}

	// Set up mock expectations
	mockAuditRepo.On("CreateAuditLog", mock.AnythingOfType("*models.AuditLog")).Return(nil)

	// Call the method under test
	err := service.CreateAuditLog(ctx, entry)

	// Assert results
	require.NoError(t, err)

	// Verify mock expectations
	mockAuditRepo.AssertExpectations(t)
}

func TestSmartChequeService_GetAuditTrail(t *testing.T) {
	mockRepo := &mocks.SmartChequeRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	service := NewSmartChequeService(mockRepo, mockAuditRepo)

	ctx := context.Background()

	// Set up mock expectations
	auditLogs := []models.AuditLog{
		{
			ID:         uuid.New(),
			Action:     "smart_cheque_created",
			Resource:   "smart_cheque",
			ResourceID: stringPtr("test-id"),
			Details:    "Smart cheque created",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Action:     "smart_cheque_updated",
			Resource:   "smart_cheque",
			ResourceID: stringPtr("test-id"),
			Details:    "Smart cheque updated",
			CreatedAt:  time.Now(),
		},
	}
	mockAuditRepo.On("GetAuditLogs", (*uuid.UUID)(nil), (*uuid.UUID)(nil), "", "smart_cheque", 10, 0).Return(auditLogs, nil)

	// Call the method under test
	trail, err := service.GetAuditTrail(ctx, "test-id", 10, 0)

	// Assert results
	require.NoError(t, err)
	assert.Len(t, trail, 2)
	assert.Equal(t, "test-id", trail[0].SmartChequeID)
	assert.Equal(t, "test-id", trail[1].SmartChequeID)

	// Verify mock expectations
	mockAuditRepo.AssertExpectations(t)
}
