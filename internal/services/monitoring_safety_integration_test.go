package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Integration test suite for monitoring and safety mechanisms
type MonitoringSafetyIntegrationSuite struct {
	balanceMonitoring BalanceMonitoringServiceInterface
	anomalyDetection  AnomalyDetectionServiceInterface
	circuitBreaker    CircuitBreakerServiceInterface
	mockBalanceRepo   *TestMockBalanceRepository
	mockAssetRepo     *TestMockAssetRepository
	mockEventBus      *TestMockEventBus
}

func setupIntegrationSuite() *MonitoringSafetyIntegrationSuite {
	mockBalanceRepo := &TestMockBalanceRepository{}
	mockAssetRepo := &TestMockAssetRepository{}
	mockEventBus := &TestMockEventBus{}

	// Setup mocks
	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	balanceConfig := &MonitoringConfig{
		CheckInterval:       time.Second * 2,
		BatchSize:           10,
		AlertCooldown:       time.Minute * 1,
		MaxAlertsPerHour:    100,
		EnableTrendAnalysis: true,
		TrendWindow:         time.Hour,
	}

	anomalyConfig := &AnomalyDetectionConfig{
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

	balanceMonitoring := NewBalanceMonitoringService(mockBalanceRepo, mockEventBus, balanceConfig)
	anomalyDetection := NewAnomalyDetectionService(mockAssetRepo, mockBalanceRepo, mockEventBus, anomalyConfig)
	circuitBreaker := NewCircuitBreakerService(mockEventBus)

	return &MonitoringSafetyIntegrationSuite{
		balanceMonitoring: balanceMonitoring,
		anomalyDetection:  anomalyDetection,
		circuitBreaker:    circuitBreaker,
		mockBalanceRepo:   mockBalanceRepo,
		mockAssetRepo:     mockAssetRepo,
		mockEventBus:      mockEventBus,
	}
}

func TestMonitoringSafetyIntegration_AlertTriggering(t *testing.T) {
	suite := setupIntegrationSuite()
	ctx := context.Background()

	// Test scenario: Balance drops below threshold, triggers anomaly detection
	enterpriseID := uuid.New()

	// Set up balance threshold
	thresholdReq := &BalanceThresholdRequest{
		EnterpriseID:   enterpriseID,
		CurrencyCode:   "USD",
		ThresholdType:  ThresholdTypeMinimum,
		ThresholdValue: "10000",
		AlertSeverity:  AlertSeverity(TestAlertSeverityHigh),
		IsActive:       true,
		CreatedBy:      uuid.New(),
	}

	threshold, err := suite.balanceMonitoring.SetBalanceThreshold(ctx, thresholdReq)
	assert.NoError(t, err)
	assert.NotNil(t, threshold)

	// Simulate transaction that should trigger anomaly
	transactionReq := &TransactionAnalysisRequest{
		TransactionID:   uuid.New(),
		EnterpriseID:    enterpriseID,
		Amount:          "15000", // Large withdrawal that would drop balance below threshold
		CurrencyCode:    "USD",
		TransactionType: "withdrawal",
		Destination:     "external_wallet",
		Timestamp:       time.Now(),
	}

	// Analyze transaction for anomalies
	anomalyScore, err := suite.anomalyDetection.AnalyzeTransaction(ctx, transactionReq)
	assert.NoError(t, err)
	assert.NotNil(t, anomalyScore)

	// Verify anomaly detection results
	assert.Greater(t, anomalyScore.OverallScore, 0.0)
	assert.Contains(t, []RiskLevel{RiskLevelModerate, RiskLevelHigh, RiskLevelCritical}, anomalyScore.RiskLevel)

	// Verify events were published
	events := suite.mockEventBus.GetPublishedEvents()
	assert.Greater(t, len(events), 0)
}

func TestMonitoringSafetyIntegration_CircuitBreakerActivation(t *testing.T) {
	suite := setupIntegrationSuite()
	ctx := context.Background()

	// Register circuit breaker for payment processing
	cbReq := &CircuitBreakerRequest{
		Name:               "payment-processor",
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            time.Second * 5,
		MaxConcurrentCalls: 10,
	}

	cb, err := suite.circuitBreaker.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State)

	// Simulate scenario where multiple transactions fail due to anomalies
	failedTransactions := 0
	for i := 0; i < 5; i++ {
		err := suite.circuitBreaker.ExecuteWithCircuitBreaker(ctx, "payment-processor", func() error {
			// Simulate payment processing
			transactionReq := &TransactionAnalysisRequest{
				TransactionID:   uuid.New(),
				EnterpriseID:    uuid.New(),
				Amount:          "50000", // Large amount
				CurrencyCode:    "USD",
				TransactionType: "withdrawal",
				Timestamp:       time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC), // Unusual time
			}

			score, err := suite.anomalyDetection.AnalyzeTransaction(ctx, transactionReq)
			if err != nil {
				return err
			}

			// If high risk, simulate payment failure
			if score.RiskLevel == RiskLevelHigh || score.RiskLevel == RiskLevelCritical {
				failedTransactions++
				return fmt.Errorf("payment blocked due to high risk score: %f", score.OverallScore)
			}

			return nil
		})

		if err != nil {
			t.Logf("Transaction %d failed: %v", i+1, err)
		}
	}

	// Verify circuit breaker opened due to failures
	cb, err = suite.circuitBreaker.GetCircuitBreaker(ctx, "payment-processor")
	assert.NoError(t, err)

	// Should be open if enough failures occurred
	if failedTransactions >= 3 {
		assert.Equal(t, StateOpen, cb.State)

		// Test that subsequent calls are rejected
		err = suite.circuitBreaker.ExecuteWithCircuitBreaker(ctx, "payment-processor", func() error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker payment-processor is open")
	}
}

func TestMonitoringSafetyIntegration_EmergencyProcedures(t *testing.T) {
	suite := setupIntegrationSuite()
	ctx := context.Background()

	// Simulate emergency scenario: Multiple high-risk transactions detected
	enterpriseID := uuid.New()

	// Set up anomaly threshold for automatic blocking
	thresholdReq := &AnomalyThresholdRequest{
		EnterpriseID:   enterpriseID,
		AnomalyType:    AnomalyTypeSuddenIncrease,
		ThresholdValue: 0.8, // High threshold
		Action:         AnomalyActionHold,
		IsActive:       true,
		CreatedBy:      uuid.New(),
	}

	_, err := suite.anomalyDetection.SetAnomalyThresholds(ctx, thresholdReq)
	assert.NoError(t, err)

	// Register emergency circuit breaker
	cbReq := &CircuitBreakerRequest{
		Name:               "emergency-system",
		FailureThreshold:   1,                // Trip immediately on failure
		SuccessThreshold:   5,                // Require multiple successes to recover
		Timeout:            time.Minute * 10, // Long timeout for emergency
		MaxConcurrentCalls: 1,
	}

	_, err = suite.circuitBreaker.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)

	// Simulate emergency scenario
	highRiskTransactions := []struct {
		amount     string
		timeHour   int
		expectTrip bool
	}{
		{"100000", 3, true}, // Very large amount at 3 AM
		{"75000", 4, false}, // Should be blocked by circuit breaker
		{"50000", 2, false}, // Should be blocked by circuit breaker
	}

	for i, tx := range highRiskTransactions {
		t.Run(fmt.Sprintf("Emergency_Transaction_%d", i+1), func(t *testing.T) {
			err := suite.circuitBreaker.ExecuteWithCircuitBreaker(ctx, "emergency-system", func() error {
				transactionReq := &TransactionAnalysisRequest{
					TransactionID:   uuid.New(),
					EnterpriseID:    enterpriseID,
					Amount:          tx.amount,
					CurrencyCode:    "USD",
					TransactionType: "withdrawal",
					Timestamp:       time.Date(2023, 1, 1, tx.timeHour, 0, 0, 0, time.UTC),
				}

				score, err := suite.anomalyDetection.AnalyzeTransaction(ctx, transactionReq)
				if err != nil {
					return err
				}

				// Emergency response: Block if very high risk
				if score.OverallScore >= 0.8 {
					return fmt.Errorf("EMERGENCY: Transaction blocked - risk score: %f", score.OverallScore)
				}

				return nil
			})

			if tx.expectTrip {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "EMERGENCY")
			} else {
				// Should be blocked by circuit breaker being open
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "circuit breaker emergency-system is open")
			}
		})
	}
}

func TestMonitoringSafetyIntegration_HighLoadStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load integration test in short mode")
	}

	suite := setupIntegrationSuite()
	ctx := context.Background()

	enterpriseID := uuid.New()

	// Start balance monitoring
	err := suite.balanceMonitoring.StartMonitoring(ctx, nil)
	assert.NoError(t, err)

	// Register circuit breakers for different services
	services := []string{"payment-service", "xrpl-service", "notification-service"}
	for _, serviceName := range services {
		cbReq := &CircuitBreakerRequest{
			Name:               serviceName,
			FailureThreshold:   10,
			SuccessThreshold:   5,
			Timeout:            time.Second * 2,
			MaxConcurrentCalls: 100,
		}
		_, err := suite.circuitBreaker.RegisterCircuitBreaker(ctx, cbReq)
		assert.NoError(t, err)
	}

	// Simulate high load with concurrent transactions
	numGoroutines := 20
	transactionsPerGoroutine := 10

	var wg sync.WaitGroup
	results := make(chan struct {
		Service   string
		Success   bool
		Error     string
		RiskScore float64
		Duration  time.Duration
	}, numGoroutines*transactionsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < transactionsPerGoroutine; j++ {
				serviceName := services[j%len(services)]
				start := time.Now()

				err := suite.circuitBreaker.ExecuteWithCircuitBreaker(ctx, serviceName, func() error {
					// Vary transaction parameters to create diverse scenarios
					amount := fmt.Sprintf("%d", 1000+goroutineID*100+j*10)
					hour := 9 + (j % 12) // Business hours mostly

					// Occasionally create high-risk transactions
					if (goroutineID+j)%20 == 0 {
						amount = "75000"
						hour = 3 // Off hours
					}

					transactionReq := &TransactionAnalysisRequest{
						TransactionID:   uuid.New(),
						EnterpriseID:    enterpriseID,
						Amount:          amount,
						CurrencyCode:    "USD",
						TransactionType: "transfer",
						Timestamp:       time.Date(2023, 1, 1, hour, 0, 0, 0, time.UTC),
					}

					score, err := suite.anomalyDetection.AnalyzeTransaction(ctx, transactionReq)
					if err != nil {
						return err
					}

					// Simulate service failure for high-risk transactions
					if score.OverallScore > 0.8 {
						return fmt.Errorf("service rejected high-risk transaction")
					}

					results <- struct {
						Service   string
						Success   bool
						Error     string
						RiskScore float64
						Duration  time.Duration
					}{
						Service:   serviceName,
						Success:   true,
						RiskScore: score.OverallScore,
						Duration:  time.Since(start),
					}

					return nil
				})

				if err != nil {
					results <- struct {
						Service   string
						Success   bool
						Error     string
						RiskScore float64
						Duration  time.Duration
					}{
						Service:  serviceName,
						Success:  false,
						Error:    err.Error(),
						Duration: time.Since(start),
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Analyze results
	serviceStats := make(map[string]struct {
		Total    int
		Success  int
		Failed   int
		Rejected int
	})

	totalDuration := time.Duration(0)
	count := 0

	for result := range results {
		stats := serviceStats[result.Service]
		stats.Total++

		if result.Success {
			stats.Success++
		} else if result.Error == "service rejected high-risk transaction" {
			stats.Failed++
		} else {
			stats.Rejected++
		}

		serviceStats[result.Service] = stats
		totalDuration += result.Duration
		count++
	}

	// Verify system handled high load appropriately
	totalTransactions := numGoroutines * transactionsPerGoroutine
	assert.Equal(t, totalTransactions, count)

	averageDuration := totalDuration / time.Duration(count)
	assert.Less(t, averageDuration, time.Second, "Average transaction duration should be reasonable")

	// Verify each service handled load appropriately
	for serviceName, stats := range serviceStats {
		t.Logf("Service %s: %d total, %d success, %d failed, %d rejected",
			serviceName, stats.Total, stats.Success, stats.Failed, stats.Rejected)

		assert.Greater(t, stats.Total, 0)
		assert.GreaterOrEqual(t, stats.Success, 0)

		// Should have some successful transactions
		successRate := float64(stats.Success) / float64(stats.Total)
		assert.Greater(t, successRate, 0.3, "Success rate should be reasonable for %s", serviceName)
	}

	// Stop monitoring
	err = suite.balanceMonitoring.StopMonitoring(ctx)
	assert.NoError(t, err)
}
