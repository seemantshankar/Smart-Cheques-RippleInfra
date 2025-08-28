package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnomalyDetectionService_AnalyzeTransaction(t *testing.T) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	// Mock dependencies
	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	config := &AnomalyDetectionConfig{
		ZScoreThreshold:        3.0,
		PercentileThreshold:    95.0,
		VelocityThreshold:      10,
		ShortTermWindow:        time.Hour,
		MediumTermWindow:       24 * time.Hour,
		LongTermWindow:         30 * 24 * time.Hour,
		AutoHoldThreshold:      0.9,
		InvestigationThreshold: 0.7,
		NotificationThreshold:  0.5,
	}

	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, config)

	// Use a fixed timestamp within business hours to avoid time-of-day behavioral penalties
	businessHourTime := time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)

	testCases := []struct {
		name                   string
		request                *TransactionAnalysisRequest
		expectedRiskLevel      RiskLevel
		expectedRecommendation AnomalyAction
	}{
		{
			name: "Normal transaction - low risk",
			request: &TransactionAnalysisRequest{
				TransactionID:   uuid.New(),
				EnterpriseID:    uuid.New(),
				Amount:          "1000",
				CurrencyCode:    "USD",
				TransactionType: "transfer",
				Destination:     "normal_dest",
				Timestamp:       businessHourTime,
			},
			expectedRiskLevel:      RiskLevelLow,
			expectedRecommendation: AnomalyActionMonitor,
		},
		{
			name: "Large transaction - medium risk",
			request: &TransactionAnalysisRequest{
				TransactionID:   uuid.New(),
				EnterpriseID:    uuid.New(),
				Amount:          "50000",
				CurrencyCode:    "USD",
				TransactionType: "withdrawal",
				Destination:     "external_dest",
				Timestamp:       businessHourTime,
			},
			expectedRiskLevel:      RiskLevelModerate,
			expectedRecommendation: AnomalyActionAlert,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score, err := service.AnalyzeTransaction(ctx, tc.request)

			assert.NoError(t, err)
			assert.NotNil(t, score)
			assert.Equal(t, tc.request.TransactionID, score.TransactionID)
			assert.Equal(t, tc.expectedRiskLevel, score.RiskLevel)
			assert.Equal(t, tc.expectedRecommendation, score.Recommendation)
			assert.GreaterOrEqual(t, score.OverallScore, 0.0)
			assert.LessOrEqual(t, score.OverallScore, 1.0)
			assert.NotEmpty(t, score.Explanation)
			assert.NotEmpty(t, score.ModelVersion)
		})
	}
}

func TestAnomalyDetectionService_SetAnomalyThresholds(t *testing.T) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()
	createdBy := uuid.New()

	request := &AnomalyThresholdRequest{
		EnterpriseID:   enterpriseID,
		AnomalyType:    AnomalyTypeVelocitySpike,
		ThresholdValue: 0.8,
		Action:         AnomalyActionAlert,
		IsActive:       true,
		CreatedBy:      createdBy,
	}

	threshold, err := service.SetAnomalyThresholds(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, threshold)
	assert.Equal(t, request.EnterpriseID, threshold.EnterpriseID)
	assert.Equal(t, request.AnomalyType, threshold.AnomalyType)
	assert.Equal(t, request.ThresholdValue, threshold.ThresholdValue)
	assert.Equal(t, request.Action, threshold.Action)
	assert.Equal(t, request.IsActive, threshold.IsActive)
	assert.NotEqual(t, uuid.Nil, threshold.ID)
}

func TestAnomalyDetectionService_ModelManagement(t *testing.T) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, nil)

	// Test model training
	trainingReq := &ModelTrainingRequest{
		TrainingDataStart: time.Now().Add(-30 * 24 * time.Hour),
		TrainingDataEnd:   time.Now(),
		ModelType:         "statistical",
	}

	result, err := service.TrainDetectionModel(ctx, trainingReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, trainingReq.ModelType, result.ModelType)
	assert.NotEmpty(t, result.ModelVersion)
	assert.Greater(t, result.TrainingAccuracy, 0.0)
	assert.Greater(t, result.ValidationAccuracy, 0.0)

	// Test model performance
	performance, err := service.GetModelPerformance(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, performance)
	assert.NotEmpty(t, performance.ModelVersion)
	assert.Greater(t, performance.TotalPredictions, int64(0))
}

func TestAnomalyDetectionService_BatchAnalysis(t *testing.T) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	request := &BatchAnalysisRequest{
		EnterpriseID: &enterpriseID,
		CurrencyCode: "USD",
		StartDate:    time.Now().Add(-24 * time.Hour),
		EndDate:      time.Now(),
		AnalysisType: "statistical",
	}

	result, err := service.PerformBatchAnalysis(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.AnalysisType, result.AnalysisType)
	assert.NotEqual(t, uuid.Nil, result.AnalysisID)
	assert.NotNil(t, result.Anomalies)
	assert.NotNil(t, result.Recommendations)
}

func TestAnomalyDetectionService_ReportGeneration(t *testing.T) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, nil)

	enterpriseID := uuid.New()

	request := &AnomalyReportRequest{
		StartDate:       time.Now().Add(-7 * 24 * time.Hour),
		EndDate:         time.Now(),
		EnterpriseID:    &enterpriseID,
		MinSeverity:     AlertSeverity(TestAlertSeverityMedium),
		IncludeResolved: true,
	}

	report, err := service.GenerateAnomalyReport(ctx, request)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEqual(t, uuid.Nil, report.ID)
	assert.Equal(t, request.StartDate, report.ReportPeriod.Start)
	assert.Equal(t, request.EndDate, report.ReportPeriod.End)
	assert.NotNil(t, report.Recommendations)
}

// Benchmark tests for performance validation
func BenchmarkAnomalyDetectionService_AnalyzeTransaction(b *testing.B) {
	ctx := context.Background()
	mockAssetRepo := &TestMockAssetRepository{}
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	config := &AnomalyDetectionConfig{
		ZScoreThreshold:       3.0,
		VelocityThreshold:     10,
		NotificationThreshold: 0.5,
	}

	service := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &TransactionAnalysisRequest{
			TransactionID:   uuid.New(),
			EnterpriseID:    uuid.New(),
			Amount:          "10000",
			CurrencyCode:    "USD",
			TransactionType: "transfer",
			Timestamp:       time.Now(),
		}

		_, err := service.AnalyzeTransaction(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
