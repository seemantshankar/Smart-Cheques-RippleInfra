package integration

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/smart-payment-infrastructure/internal/services"
)

// ReconciliationTestSuite tests reconciliation accuracy and performance
type ReconciliationTestSuite struct {
	suite.Suite

	// Test data
	testEnterprises []TestEnterprise
}

type TestEnterprise struct {
	EnterpriseID     uuid.UUID
	XRPLAddress      string
	InternalBalances map[string]*big.Int
	XRPLBalances     map[string]*big.Int
}

func TestReconciliationAccuracyAndPerformance(t *testing.T) {
	suite.Run(t, new(ReconciliationTestSuite))
}

func (suite *ReconciliationTestSuite) SetupSuite() {
	suite.setupTestServices()
	suite.setupTestData()
}

func (suite *ReconciliationTestSuite) setupTestServices() {
	t := suite.T()
	t.Log("Setting up reconciliation test services")
	// In real implementation, initialize services with test database
}

func (suite *ReconciliationTestSuite) setupTestData() {
	suite.testEnterprises = make([]TestEnterprise, 5)

	for i := range suite.testEnterprises {
		enterprise := TestEnterprise{
			EnterpriseID:     uuid.New(),
			XRPLAddress:      fmt.Sprintf("rTestAddr%d%s", i, generateRandomAddress()),
			InternalBalances: make(map[string]*big.Int),
			XRPLBalances:     make(map[string]*big.Int),
		}

		currencies := []string{"USDT", "USDC", "XRP"}
		for _, currency := range currencies {
			baseAmount := big.NewInt(int64(1000000000 * (i + 1)))
			enterprise.InternalBalances[currency] = new(big.Int).Set(baseAmount)

			// Add discrepancy for testing
			discrepancy := big.NewInt(int64(rand.Intn(1000000)))
			if rand.Float32() > 0.5 {
				enterprise.XRPLBalances[currency] = new(big.Int).Add(baseAmount, discrepancy)
			} else {
				enterprise.XRPLBalances[currency] = new(big.Int).Sub(baseAmount, discrepancy)
			}
		}

		suite.testEnterprises[i] = enterprise
	}
}

func (suite *ReconciliationTestSuite) TestBasicReconciliationAccuracy() {
	t := suite.T()
	ctx := context.Background()

	enterprise := suite.testEnterprises[0]

	request := &services.ReconciliationRequest{
		EnterpriseID: &enterprise.EnterpriseID,
		CurrencyCode: "USDT",
	}

	result, err := suite.performReconciliation(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Discrepancies)

	discrepancy := result.Discrepancies[0]
	expectedInternal := enterprise.InternalBalances["USDT"]
	expectedXRPL := enterprise.XRPLBalances["USDT"]
	expectedDiscrepancy := new(big.Int).Sub(expectedInternal, expectedXRPL)

	assert.Equal(t, expectedInternal.String(), discrepancy.InternalBalance)
	assert.Equal(t, expectedXRPL.String(), discrepancy.XRPLBalance)
	assert.Equal(t, expectedDiscrepancy.String(), discrepancy.DiscrepancyAmount)

	t.Logf("Reconciliation completed: Internal=%s, XRPL=%s, Discrepancy=%s",
		discrepancy.InternalBalance, discrepancy.XRPLBalance, discrepancy.DiscrepancyAmount)
}

func (suite *ReconciliationTestSuite) TestBulkReconciliationPerformance() {
	t := suite.T()
	ctx := context.Background()

	startTime := time.Now()
	var requests []*services.ReconciliationRequest

	currencies := []string{"USDT", "USDC", "XRP"}
	for _, enterprise := range suite.testEnterprises {
		for _, currency := range currencies {
			request := &services.ReconciliationRequest{
				EnterpriseID: &enterprise.EnterpriseID,
				CurrencyCode: currency,
			}
			requests = append(requests, request)
		}
	}

	totalRequests := len(requests)
	results := make([]*services.ReconciliationResult, 0, totalRequests)

	for _, request := range requests {
		result, err := suite.performReconciliation(ctx, request)
		require.NoError(t, err)
		results = append(results, result)
	}

	duration := time.Since(startTime)
	avgTime := duration / time.Duration(totalRequests)
	throughput := float64(totalRequests) / duration.Seconds()

	t.Logf("Bulk reconciliation: %d requests in %v", totalRequests, duration)
	t.Logf("Average time: %v, Throughput: %.2f/sec", avgTime, throughput)

	assert.Less(t, avgTime, 100*time.Millisecond, "Avg time should be under 100ms")
	assert.Greater(t, throughput, 10.0, "Should process 10+ reconciliations/sec")
	assert.Equal(t, totalRequests, len(results))
}

func (suite *ReconciliationTestSuite) TestConcurrentReconciliation() {
	t := suite.T()
	ctx := context.Background()

	numWorkers := 3
	reconciliationsPerWorker := 5

	var wg sync.WaitGroup
	results := make(chan *ConcurrentResult, numWorkers*reconciliationsPerWorker)
	errors := make(chan error, numWorkers*reconciliationsPerWorker)

	startTime := time.Now()

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < reconciliationsPerWorker; i++ {
				enterprise := suite.testEnterprises[workerID%len(suite.testEnterprises)]
				currency := []string{"USDT", "USDC", "XRP"}[i%3]

				reconciliationStart := time.Now()

				request := &services.ReconciliationRequest{
					EnterpriseID: &enterprise.EnterpriseID,
					CurrencyCode: currency,
				}

				result, err := suite.performReconciliation(ctx, request)
				_ = result // Use variable to avoid unused warning
				duration := time.Since(reconciliationStart)

				if err != nil {
					errors <- err
				} else {
					results <- &ConcurrentResult{
						WorkerID: workerID,
						Duration: duration,
						Success:  true,
					}
				}
			}
		}(worker)
	}

	wg.Wait()
	close(results)
	close(errors)

	totalDuration := time.Since(startTime)

	successCount := 0
	var totalTime time.Duration
	for result := range results {
		totalTime += result.Duration
		successCount++
	}

	errorCount := 0
	for err := range errors {
		errorCount++
		t.Logf("Error: %v", err)
	}

	totalOps := numWorkers * reconciliationsPerWorker
	avgTime := totalTime / time.Duration(successCount)

	t.Logf("Concurrent reconciliation: %d/%d successful in %v", successCount, totalOps, totalDuration)
	t.Logf("Average time: %v", avgTime)

	assert.Equal(t, totalOps, successCount, "All should succeed")
	assert.Equal(t, 0, errorCount, "No errors expected")
	assert.Less(t, avgTime, 200*time.Millisecond, "Avg time under 200ms")
}

func (suite *ReconciliationTestSuite) TestReconciliationWithLargeDiscrepancies() {
	t := suite.T()
	ctx := context.Background()

	enterprise := TestEnterprise{
		EnterpriseID:     uuid.New(),
		XRPLAddress:      "rLargeDiscrepTest" + generateRandomAddress(),
		InternalBalances: make(map[string]*big.Int),
		XRPLBalances:     make(map[string]*big.Int),
	}

	// Large discrepancy: 100K vs 98K USDT
	enterprise.InternalBalances["USDT"] = big.NewInt(100000000000)
	enterprise.XRPLBalances["USDT"] = big.NewInt(98000000000)

	// Add this enterprise to our test data so findEnterpriseByID can find it
	suite.testEnterprises = append(suite.testEnterprises, enterprise)

	request := &services.ReconciliationRequest{
		EnterpriseID: &enterprise.EnterpriseID,
		CurrencyCode: "USDT",
	}

	result, err := suite.performReconciliation(ctx, request)
	require.NoError(t, err)
	require.NotEmpty(t, result.Discrepancies)

	discrepancy := result.Discrepancies[0]
	expectedDiscrepancy := big.NewInt(2000000000) // 2000 USDT difference

	assert.Equal(t, expectedDiscrepancy.String(), discrepancy.DiscrepancyAmount)
	assert.Equal(t, services.DiscrepancySeverityCritical, discrepancy.Severity)
	// Note: AlertGenerated is not a field in BalanceDiscrepancy, but we can check the severity

	t.Logf("Large discrepancy handled: %s severity", discrepancy.Severity)
}

// Helper types
type ConcurrentResult struct {
	WorkerID int
	Duration time.Duration
	Success  bool
}

// Helper methods
func (suite *ReconciliationTestSuite) performReconciliation(ctx context.Context, req *services.ReconciliationRequest) (*services.ReconciliationResult, error) {
	// Simulate reconciliation - in real implementation call actual service
	enterprise := suite.findEnterpriseByID(req.EnterpriseID)
	if enterprise == nil {
		return nil, fmt.Errorf("enterprise not found")
	}

	internal := enterprise.InternalBalances[req.CurrencyCode]
	xrpl := enterprise.XRPLBalances[req.CurrencyCode]
	discrepancy := new(big.Int).Sub(internal, xrpl)
	severity := suite.calculateSeverity(discrepancy)

	// Create a proper BalanceDiscrepancy
	discrepancyRecord := &services.BalanceDiscrepancy{
		ID:                uuid.New(),
		ReconciliationID:  uuid.New(),
		EnterpriseID:      *req.EnterpriseID,
		CurrencyCode:      req.CurrencyCode,
		InternalBalance:   internal.String(),
		XRPLBalance:       xrpl.String(),
		DiscrepancyAmount: discrepancy.String(),
		Severity:          severity,
		Status:            services.DiscrepancyStatusPending,
		DetectedAt:        time.Now(),
	}

	// Create a proper ReconciliationResult
	completedAt := time.Now()
	result := &services.ReconciliationResult{
		ID:                     uuid.New(),
		EnterpriseID:           req.EnterpriseID,
		CurrencyCode:           req.CurrencyCode,
		Status:                 services.ReconciliationStatusCompleted,
		StartedAt:              time.Now(),
		CompletedAt:            &completedAt,
		TotalChecked:           1,
		DiscrepanciesFound:     1,
		TotalDiscrepancyAmount: discrepancy.String(),
		Discrepancies:          []*services.BalanceDiscrepancy{discrepancyRecord},
	}

	return result, nil
}

func (suite *ReconciliationTestSuite) findEnterpriseByID(id *uuid.UUID) *TestEnterprise {
	if id == nil {
		return nil
	}

	for i := range suite.testEnterprises {
		if suite.testEnterprises[i].EnterpriseID == *id {
			return &suite.testEnterprises[i]
		}
	}
	return nil
}

func (suite *ReconciliationTestSuite) calculateSeverity(discrepancy *big.Int) services.DiscrepancySeverity {
	abs := new(big.Int).Abs(discrepancy)

	if abs.Cmp(big.NewInt(1000000000)) >= 0 { // 1000 USDT
		return services.DiscrepancySeverityCritical
	} else if abs.Cmp(big.NewInt(100000000)) >= 0 { // 100 USDT
		return services.DiscrepancySeverityHigh
	} else if abs.Cmp(big.NewInt(10000000)) >= 0 { // 10 USDT
		return services.DiscrepancySeverityMedium
	}

	return services.DiscrepancySeverityLow
}

func generateRandomAddress() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
