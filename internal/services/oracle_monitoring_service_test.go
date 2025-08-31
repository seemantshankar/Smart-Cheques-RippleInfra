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

func TestOracleMonitoringService_GetDashboardMetrics(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	monitoringService := NewOracleMonitoringService(mockOracleRepo)

	// Test data - create fixed UUIDs for testing
	provider1ID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	provider2ID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	provider3ID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	provider1 := &models.OracleProvider{
		ID:          provider1ID,
		Name:        "API Oracle",
		Type:        models.OracleTypeAPI,
		IsActive:    true,
		Reliability: 0.95,
	}

	provider2 := &models.OracleProvider{
		ID:          provider2ID,
		Name:        "Webhook Oracle",
		Type:        models.OracleTypeWebhook,
		IsActive:    true,
		Reliability: 0.85,
	}

	provider3 := &models.OracleProvider{
		ID:          provider3ID,
		Name:        "Manual Oracle",
		Type:        models.OracleTypeManual,
		IsActive:    false,
		Reliability: 1.0,
	}

	providers := []*models.OracleProvider{provider1, provider2, provider3}

	stats1 := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 95,
		models.RequestStatusFailed:    5,
	}

	stats2 := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 85,
		models.RequestStatusFailed:    15,
	}

	// Stats for inactive provider
	stats3 := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 0,
		models.RequestStatusFailed:    0,
	}

	// Mock expectations
	mockOracleRepo.On("ListOracleProviders", mock.Anything, 100, 0).Return(providers, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &provider1ID, mock.Anything, mock.Anything).Return(stats1, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &provider2ID, mock.Anything, mock.Anything).Return(stats2, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &provider3ID, mock.Anything, mock.Anything).Return(stats3, nil)

	// Execute
	ctx := context.Background()
	metrics, err := monitoringService.GetDashboardMetrics(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.TotalProviders)
	assert.Equal(t, int64(2), metrics.ActiveProviders)
	assert.Equal(t, int64(200), metrics.TotalRequests)       // 100 + 100 + 0
	assert.Equal(t, int64(180), metrics.SuccessfulRequests)  // 95 + 85 + 0
	assert.Equal(t, int64(20), metrics.FailedRequests)       // 5 + 15 + 0
	assert.InDelta(t, 90.0, metrics.OverallSuccessRate, 0.1) // (95+85+0)/(100+100+0) * 100 = 90%
	assert.InDelta(t, 10.0, metrics.ErrorRate, 0.1)          // (5+15+0)/(100+100+0) * 100 = 10%
	assert.Len(t, metrics.ProviderMetrics, 3)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleMonitoringService_GetSLAMonitoring(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	monitoringService := NewOracleMonitoringService(mockOracleRepo)

	// Test data
	providerID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	provider := &models.OracleProvider{
		ID:   providerID,
		Name: "API Oracle",
		Type: models.OracleTypeAPI,
	}

	hourlyStats := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 50,
		models.RequestStatusFailed:    2,
	}

	dailyStats := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 400,
		models.RequestStatusFailed:    15,
	}

	weeklyStats := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 2800,
		models.RequestStatusFailed:    100,
	}

	// Mock expectations
	mockOracleRepo.On("ListOracleProviders", mock.Anything, 100, 0).Return([]*models.OracleProvider{provider}, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &providerID, mock.Anything, mock.Anything).Return(hourlyStats, nil).Once()
	mockOracleRepo.On("GetRequestStats", mock.Anything, &providerID, mock.Anything, mock.Anything).Return(dailyStats, nil).Once()
	mockOracleRepo.On("GetRequestStats", mock.Anything, &providerID, mock.Anything, mock.Anything).Return(weeklyStats, nil).Once()

	// Execute
	ctx := context.Background()
	report, err := monitoringService.GetSLAMonitoring(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.WithinDuration(t, time.Now(), report.ReportGenerated, time.Second)
	assert.Len(t, report.ProviderSLAs, 1)

	providerSLA := report.ProviderSLAs[0]
	assert.Equal(t, provider.ID, providerSLA.ProviderID)
	assert.Equal(t, provider.Name, providerSLA.ProviderName)
	assert.Equal(t, string(provider.Type), providerSLA.ProviderType)
	assert.NotNil(t, providerSLA.HourlySLA)
	assert.NotNil(t, providerSLA.DailySLA)
	assert.NotNil(t, providerSLA.WeeklySLA)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleMonitoringService_GetCostAnalysis(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	monitoringService := NewOracleMonitoringService(mockOracleRepo)

	// Test data
	provider1ID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	provider2ID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	provider1 := &models.OracleProvider{
		ID:   provider1ID,
		Name: "API Oracle",
		Type: models.OracleTypeAPI,
	}

	provider2 := &models.OracleProvider{
		ID:   provider2ID,
		Name: "Webhook Oracle",
		Type: models.OracleTypeWebhook,
	}

	providers := []*models.OracleProvider{provider1, provider2}

	stats1 := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 1000,
		models.RequestStatusFailed:    50,
	}

	stats2 := map[models.RequestStatus]int64{
		models.RequestStatusCompleted: 500,
		models.RequestStatusFailed:    25,
	}

	// Mock expectations
	mockOracleRepo.On("ListOracleProviders", mock.Anything, 100, 0).Return(providers, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &provider1ID, mock.Anything, mock.Anything).Return(stats1, nil)
	mockOracleRepo.On("GetRequestStats", mock.Anything, &provider2ID, mock.Anything, mock.Anything).Return(stats2, nil)

	// Execute
	ctx := context.Background()
	report, err := monitoringService.GetCostAnalysis(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.WithinDuration(t, time.Now(), report.ReportGenerated, time.Second)
	assert.InDelta(t, 1.575, report.TotalCost, 0.01) // (1050 * 0.001) + (525 * 0.001) = 1.05 + 0.525 = 1.575
	assert.Len(t, report.ProviderCosts, 2)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}
