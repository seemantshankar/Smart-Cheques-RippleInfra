package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCircuitBreakerService_RegisterCircuitBreaker(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	testCases := []struct {
		name        string
		request     *CircuitBreakerRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid circuit breaker registration",
			request: &CircuitBreakerRequest{
				Name:               "payment-service",
				FailureThreshold:   5,
				SuccessThreshold:   3,
				Timeout:            time.Minute * 5,
				MaxConcurrentCalls: 100,
				VolumeThreshold:    10,
				ErrorRateThreshold: 50.0,
			},
			expectError: false,
		},
		{
			name: "Duplicate circuit breaker registration",
			request: &CircuitBreakerRequest{
				Name:               "payment-service", // Same name as above
				FailureThreshold:   3,
				SuccessThreshold:   2,
				Timeout:            time.Minute * 3,
				MaxConcurrentCalls: 50,
			},
			expectError: true,
			errorMsg:    "already exists",
		},
		{
			name: "Another valid circuit breaker",
			request: &CircuitBreakerRequest{
				Name:               "xrpl-service",
				FailureThreshold:   3,
				SuccessThreshold:   2,
				Timeout:            time.Minute * 2,
				MaxConcurrentCalls: 50,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cb, err := service.RegisterCircuitBreaker(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, cb)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cb)
				assert.Equal(t, tc.request.Name, cb.Name)
				assert.Equal(t, tc.request.FailureThreshold, cb.FailureThreshold)
				assert.Equal(t, tc.request.SuccessThreshold, cb.SuccessThreshold)
				assert.Equal(t, tc.request.Timeout, cb.Timeout)
				assert.Equal(t, StateClosed, cb.State)
				assert.Equal(t, 0, cb.FailureCount)
				assert.Equal(t, 0, cb.SuccessCount)
			}
		})
	}
}

func TestCircuitBreakerService_CircuitBreakerStates(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	// Register a circuit breaker
	cbReq := &CircuitBreakerRequest{
		Name:               "test-service",
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:            time.Second * 5,
		MaxConcurrentCalls: 10,
	}

	cb, err := service.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State)

	// Test successful operations (should stay closed)
	for i := 0; i < 5; i++ {
		err := service.ExecuteWithCircuitBreaker(ctx, "test-service", func() error {
			return nil // Success
		})
		assert.NoError(t, err)
	}

	// Verify still closed
	cb, err = service.GetCircuitBreaker(ctx, "test-service")
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State)
	assert.Equal(t, 5, cb.SuccessCount)
	assert.Equal(t, 0, cb.FailureCount)

	// Test failures to trigger opening
	for i := 0; i < 3; i++ {
		err := service.ExecuteWithCircuitBreaker(ctx, "test-service", func() error {
			return errors.New("operation failed")
		})
		assert.Error(t, err)
	}

	// Should now be open
	cb, err = service.GetCircuitBreaker(ctx, "test-service")
	assert.NoError(t, err)
	assert.Equal(t, StateOpen, cb.State)
	assert.Equal(t, 3, cb.FailureCount)

	// Test that operations are rejected when open
	err = service.ExecuteWithCircuitBreaker(ctx, "test-service", func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker test-service is open")
}

func TestCircuitBreakerService_ManualControl(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	// Register circuit breaker
	cbReq := &CircuitBreakerRequest{
		Name:               "manual-test-service",
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            time.Minute,
		MaxConcurrentCalls: 10,
	}

	_, err := service.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)

	// Test manual trip
	err = service.TripCircuitBreaker(ctx, "manual-test-service", "Manual test")
	assert.NoError(t, err)

	// Verify circuit is open
	cb, err := service.GetCircuitBreaker(ctx, "manual-test-service")
	assert.NoError(t, err)
	assert.Equal(t, StateOpen, cb.State)

	// Operations should be rejected
	err = service.ExecuteWithCircuitBreaker(ctx, "manual-test-service", func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker manual-test-service is open")

	// Test manual reset
	err = service.ResetCircuitBreaker(ctx, "manual-test-service")
	assert.NoError(t, err)

	// Verify circuit is closed and reset
	cb, err = service.GetCircuitBreaker(ctx, "manual-test-service")
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State)
	assert.Equal(t, 0, cb.FailureCount)
	assert.Equal(t, 0, cb.SuccessCount)

	// Operations should now work
	err = service.ExecuteWithCircuitBreaker(ctx, "manual-test-service", func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestCircuitBreakerService_Metrics(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	// Register circuit breaker
	cbReq := &CircuitBreakerRequest{
		Name:               "metrics-test-service",
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            time.Minute,
		MaxConcurrentCalls: 10,
	}

	_, err := service.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)

	// Execute some successful operations
	for i := 0; i < 7; i++ {
		err := service.ExecuteWithCircuitBreaker(ctx, "metrics-test-service", func() error {
			return nil
		})
		assert.NoError(t, err)
	}

	// Execute some failed operations
	for i := 0; i < 3; i++ {
		err := service.ExecuteWithCircuitBreaker(ctx, "metrics-test-service", func() error {
			return errors.New("test failure")
		})
		assert.Error(t, err)
	}

	// Get metrics
	metrics, err := service.GetCircuitBreakerMetrics(ctx, "metrics-test-service")
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, "metrics-test-service", metrics.Name)
	assert.Equal(t, int64(10), metrics.TotalCalls)
	assert.Equal(t, int64(3), metrics.TotalFailures)
	assert.Equal(t, int64(7), metrics.TotalSuccesses)
	assert.Equal(t, 30.0, metrics.ErrorRate)   // 3/10 * 100
	assert.Equal(t, 70.0, metrics.SuccessRate) // 7/10 * 100
}

func TestCircuitBreakerService_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	service := NewCircuitBreakerService(mockEventBus)

	// Test operations on non-existent circuit breaker
	testCases := []struct {
		name      string
		operation func() error
		expectErr string
	}{
		{
			name: "Get non-existent circuit breaker",
			operation: func() error {
				_, err := service.GetCircuitBreaker(ctx, "non-existent")
				return err
			},
			expectErr: "not found",
		},
		{
			name: "Execute on non-existent circuit breaker",
			operation: func() error {
				return service.ExecuteWithCircuitBreaker(ctx, "non-existent", func() error {
					return nil
				})
			},
			expectErr: "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.operation()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestCircuitBreakerService_ConcurrentCalls(t *testing.T) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}

	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	// Register circuit breaker with low concurrent call limit
	cbReq := &CircuitBreakerRequest{
		Name:               "concurrent-test-service",
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:            time.Minute,
		MaxConcurrentCalls: 3, // Low limit for testing
	}

	_, err := service.RegisterCircuitBreaker(ctx, cbReq)
	assert.NoError(t, err)

	// Test concurrent execution
	var wg sync.WaitGroup
	results := make(chan error, 6)

	// Start more concurrent calls than allowed
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := service.ExecuteWithCircuitBreaker(ctx, "concurrent-test-service", func() error {
				time.Sleep(time.Millisecond * 100) // Simulate work
				return nil
			})
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	// Count successful vs rejected calls
	successCount := 0
	rejectedCount := 0

	for err := range results {
		if err == nil {
			successCount++
		} else {
			rejectedCount++
		}
	}

	// Should have some successful calls and possibly some rejected due to concurrency limit
	assert.True(t, successCount > 0, "Should have some successful calls")
	assert.True(t, successCount <= 6, "Should not exceed total calls")
}

// Performance benchmark
func BenchmarkCircuitBreakerService_ExecuteWithCircuitBreaker(b *testing.B) {
	ctx := context.Background()
	mockEventBus := &TestMockEventBus{}
	mockEventBus.On("Close").Return(nil)
	mockEventBus.On("PublishEvent", mock.Anything, mock.Anything).Return(nil)

	service := NewCircuitBreakerService(mockEventBus)

	// Register circuit breaker
	cbReq := &CircuitBreakerRequest{
		Name:               "benchmark-service",
		FailureThreshold:   1000,
		SuccessThreshold:   3,
		Timeout:            time.Minute,
		MaxConcurrentCalls: 1000,
	}
	_, err := service.RegisterCircuitBreaker(ctx, cbReq)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.ExecuteWithCircuitBreaker(ctx, "benchmark-service", func() error {
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
