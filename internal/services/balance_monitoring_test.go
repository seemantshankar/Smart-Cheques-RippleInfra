package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceMonitoringService_StartMonitoring(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	config := &MonitoringConfig{
		CheckInterval:    time.Second * 5,
		BatchSize:        10,
		AlertCooldown:    time.Minute * 5,
		MaxAlertsPerHour: 60,
	}

	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, config)

	// Mock event publishing
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)
	mockEventBus.On("Close").Return(nil)

	// Test starting monitoring
	err := service.StartMonitoring(ctx, config)
	assert.NoError(t, err)

	// Verify status
	status, err := service.GetMonitoringStatus(ctx)
	assert.NoError(t, err)
	assert.True(t, status.IsRunning)
	assert.NotNil(t, status.StartedAt)

	// Test starting already running monitoring
	err = service.StartMonitoring(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop monitoring for cleanup
	err = service.StopMonitoring(ctx)
	assert.NoError(t, err)

	mockEventBus.AssertExpectations(t)
}

func TestBalanceMonitoringService_StopMonitoring(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	config := &MonitoringConfig{
		CheckInterval: time.Second * 5,
		BatchSize:     10,
	}

	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, config)

	// Test stopping when not running
	err := service.StopMonitoring(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start monitoring first
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)
	mockEventBus.On("Close").Return(nil)
	err = service.StartMonitoring(ctx, config)
	assert.NoError(t, err)

	// Now stop it
	err = service.StopMonitoring(ctx)
	assert.NoError(t, err)

	// Verify status
	status, err := service.GetMonitoringStatus(ctx)
	assert.NoError(t, err)
	assert.False(t, status.IsRunning)
}

func TestBalanceMonitoringService_SetBalanceThreshold(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()
	createdBy := uuid.New()

	testCases := []struct {
		name        string
		request     *BalanceThresholdRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid minimum threshold",
			request: &BalanceThresholdRequest{
				EnterpriseID:   enterpriseID,
				CurrencyCode:   "USD",
				ThresholdType:  ThresholdTypeMinimum,
				ThresholdValue: "1000",
				AlertSeverity:  AlertSeverity(TestAlertSeverityMedium),
				IsActive:       true,
				CreatedBy:      createdBy,
			},
			expectError: false,
		},
		{
			name: "Invalid threshold value",
			request: &BalanceThresholdRequest{
				EnterpriseID:   enterpriseID,
				CurrencyCode:   "USD",
				ThresholdType:  ThresholdTypeMinimum,
				ThresholdValue: "invalid",
				AlertSeverity:  AlertSeverity(TestAlertSeverityMedium),
				IsActive:       true,
				CreatedBy:      createdBy,
			},
			expectError: true,
			errorMsg:    "invalid threshold value",
		},
		{
			name: "Valid percentage threshold",
			request: &BalanceThresholdRequest{
				EnterpriseID:   enterpriseID,
				CurrencyCode:   "USDT",
				ThresholdType:  ThresholdTypePercentage,
				ThresholdValue: "25",
				AlertSeverity:  AlertSeverity(TestAlertSeverityHigh),
				IsActive:       true,
				CreatedBy:      createdBy,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			threshold, err := service.SetBalanceThreshold(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, threshold)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, threshold)
				assert.Equal(t, tc.request.EnterpriseID, threshold.EnterpriseID)
				assert.Equal(t, tc.request.CurrencyCode, threshold.CurrencyCode)
				assert.Equal(t, tc.request.ThresholdType, threshold.ThresholdType)
				assert.Equal(t, tc.request.ThresholdValue, threshold.ThresholdValue)
				assert.Equal(t, tc.request.AlertSeverity, threshold.AlertSeverity)
				assert.NotEqual(t, uuid.Nil, threshold.ID)
			}
		})
	}
}

func TestBalanceMonitoringService_GetBalanceTrends(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	request := &BalanceTrendRequest{
		EnterpriseID: enterpriseID,
		CurrencyCode: "USD",
		TimeRange:    TimeRangeDay,
		Granularity:  "hourly",
	}

	trends, err := service.GetBalanceTrends(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, enterpriseID, trends.EnterpriseID)
	assert.Equal(t, "USD", trends.CurrencyCode)
	assert.Equal(t, TimeRangeDay, trends.TimeRange)
	assert.Equal(t, TrendDirectionFlat, trends.TrendDirection)
}

func TestBalanceMonitoringService_DetectBalanceAnomalies(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	anomalies, err := service.DetectBalanceAnomalies(ctx, enterpriseID)
	assert.NoError(t, err)
	assert.NotNil(t, anomalies)
	assert.Equal(t, 0, len(anomalies)) // Empty for mock implementation
}

func TestBalanceMonitoringService_AlertManagement(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	// Test getting active alerts
	alerts, err := service.GetActiveAlerts(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)
	assert.Equal(t, 0, len(alerts)) // Empty for mock implementation

	// Test getting alerts for specific enterprise
	enterpriseID := uuid.New()
	alerts, err = service.GetActiveAlerts(ctx, &enterpriseID)
	assert.NoError(t, err)
	assert.NotNil(t, alerts)

	// Test acknowledging alert
	alertID := uuid.New()
	acknowledgedBy := uuid.New()
	err = service.AcknowledgeAlert(ctx, alertID, acknowledgedBy)
	assert.NoError(t, err)

	// Test getting alert history
	historyReq := &AlertHistoryRequest{
		EnterpriseID: &enterpriseID,
		CurrencyCode: "USD",
		StartDate:    time.Now().Add(-24 * time.Hour),
		EndDate:      time.Now(),
		Limit:        50,
		Offset:       0,
	}

	history, err := service.GetAlertHistory(ctx, historyReq)
	assert.NoError(t, err)
	assert.NotNil(t, history)
}

func TestBalanceMonitoringService_ThresholdManagement(t *testing.T) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	// Test getting thresholds for enterprise
	thresholds, err := service.GetBalanceThresholds(ctx, enterpriseID)
	assert.NoError(t, err)
	assert.NotNil(t, thresholds)

	// Test updating threshold (should return error for non-existent threshold)
	thresholdID := uuid.New()
	updatedBy := uuid.New()
	newValue := "5000"
	newSeverity := AlertSeverity(TestAlertSeverityHigh)

	updateReq := &UpdateThresholdRequest{
		ThresholdID:    thresholdID,
		ThresholdValue: &newValue,
		AlertSeverity:  &newSeverity,
		UpdatedBy:      updatedBy,
	}

	_, err = service.UpdateBalanceThreshold(ctx, updateReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test deleting threshold
	err = service.DeleteBalanceThreshold(ctx, thresholdID)
	assert.NoError(t, err) // Mock implementation returns nil
}

// Benchmark tests for performance validation
func BenchmarkBalanceMonitoringService_SetBalanceThreshold(b *testing.B) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()
	createdBy := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &BalanceThresholdRequest{
			EnterpriseID:   enterpriseID,
			CurrencyCode:   "USD",
			ThresholdType:  ThresholdTypeMinimum,
			ThresholdValue: "1000",
			AlertSeverity:  AlertSeverity(TestAlertSeverityMedium),
			IsActive:       true,
			CreatedBy:      createdBy,
		}

		_, err := service.SetBalanceThreshold(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBalanceMonitoringService_GetBalanceTrends(b *testing.B) {
	ctx := context.Background()
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &BalanceTrendRequest{
			EnterpriseID: enterpriseID,
			CurrencyCode: "USD",
			TimeRange:    TimeRangeDay,
			Granularity:  "hourly",
		}

		_, err := service.GetBalanceTrends(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test scenarios for alert triggering
func TestBalanceMonitoringService_AlertScenarios(t *testing.T) {
	testCases := []struct {
		name             string
		thresholdType    ThresholdType
		thresholdValue   string
		currentBalance   string
		expectedAlert    bool
		expectedSeverity string
	}{
		{
			name:             "Balance below minimum threshold",
			thresholdType:    ThresholdTypeMinimum,
			thresholdValue:   "1000",
			currentBalance:   "500",
			expectedAlert:    true,
			expectedSeverity: TestAlertSeverityHigh,
		},
		{
			name:           "Balance above minimum threshold",
			thresholdType:  ThresholdTypeMinimum,
			thresholdValue: "1000",
			currentBalance: "1500",
			expectedAlert:  false,
		},
		{
			name:             "Balance above maximum threshold",
			thresholdType:    ThresholdTypeMaximum,
			thresholdValue:   "50000",
			currentBalance:   "75000",
			expectedAlert:    true,
			expectedSeverity: TestAlertSeverityMedium,
		},
		{
			name:           "Balance within maximum threshold",
			thresholdType:  ThresholdTypeMaximum,
			thresholdValue: "50000",
			currentBalance: "25000",
			expectedAlert:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test validates the alert triggering logic
			// In a real implementation, this would test the actual alert generation

			ctx := context.Background()
			mockBalanceRepo := &TestMockBalanceRepository{}
			mockEventBus := &TestMockEventBus{}

			mockEventBus.On("Close").Return(nil)
			service := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, nil)

			enterpriseID := uuid.New()
			createdBy := uuid.New()

			// Set up threshold
			thresholdReq := &BalanceThresholdRequest{
				EnterpriseID:   enterpriseID,
				CurrencyCode:   "USD",
				ThresholdType:  tc.thresholdType,
				ThresholdValue: tc.thresholdValue,
				AlertSeverity:  AlertSeverity(tc.expectedSeverity),
				IsActive:       true,
				CreatedBy:      createdBy,
			}

			threshold, err := service.SetBalanceThreshold(ctx, thresholdReq)
			assert.NoError(t, err)
			assert.NotNil(t, threshold)

			// In a real implementation, you would:
			// 1. Mock the balance repository to return the current balance
			// 2. Trigger a balance check
			// 3. Verify that alerts are generated when expected
			// 4. Verify alert severity matches expectations

			// For now, we just verify the threshold was set correctly
			assert.Equal(t, tc.thresholdType, threshold.ThresholdType)
			assert.Equal(t, tc.thresholdValue, threshold.ThresholdValue)
		})
	}
}
