package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// CircuitBreakerServiceInterface defines the interface for circuit breaker operations
type CircuitBreakerServiceInterface interface {
	// Circuit breaker management
	RegisterCircuitBreaker(ctx context.Context, req *CircuitBreakerRequest) (*CircuitBreaker, error)
	GetCircuitBreaker(ctx context.Context, name string) (*CircuitBreaker, error)
	UpdateCircuitBreaker(ctx context.Context, req *UpdateCircuitBreakerRequest) error

	// Circuit operations
	ExecuteWithCircuitBreaker(ctx context.Context, name string, operation func() error) error
	TripCircuitBreaker(ctx context.Context, name string, reason string) error
	ResetCircuitBreaker(ctx context.Context, name string) error

	// Monitoring
	GetCircuitBreakerStatus(ctx context.Context) ([]*CircuitBreakerStatus, error)
	GetCircuitBreakerMetrics(ctx context.Context, name string) (*CircuitBreakerMetrics, error)
}

// CircuitBreakerService implements circuit breaker functionality
type CircuitBreakerService struct {
	circuitBreakers map[string]*CircuitBreaker
	mutex           sync.RWMutex
	messagingClient messaging.EventBus
}

// NewCircuitBreakerService creates a new circuit breaker service
func NewCircuitBreakerService(messagingClient messaging.EventBus) CircuitBreakerServiceInterface {
	return &CircuitBreakerService{
		circuitBreakers: make(map[string]*CircuitBreaker),
		messagingClient: messagingClient,
	}
}

// Types and structs
type CircuitBreakerRequest struct {
	Name               string        `json:"name" validate:"required"`
	FailureThreshold   int           `json:"failure_threshold" validate:"required"`
	SuccessThreshold   int           `json:"success_threshold"`
	Timeout            time.Duration `json:"timeout" validate:"required"`
	MaxConcurrentCalls int           `json:"max_concurrent_calls"`
	VolumeThreshold    int           `json:"volume_threshold"`
	ErrorRateThreshold float64       `json:"error_rate_threshold"`
}

type UpdateCircuitBreakerRequest struct {
	Name               string         `json:"name" validate:"required"`
	FailureThreshold   *int           `json:"failure_threshold,omitempty"`
	SuccessThreshold   *int           `json:"success_threshold,omitempty"`
	Timeout            *time.Duration `json:"timeout,omitempty"`
	MaxConcurrentCalls *int           `json:"max_concurrent_calls,omitempty"`
}

type CircuitBreaker struct {
	Name               string              `json:"name"`
	State              CircuitBreakerState `json:"state"`
	FailureThreshold   int                 `json:"failure_threshold"`
	SuccessThreshold   int                 `json:"success_threshold"`
	Timeout            time.Duration       `json:"timeout"`
	MaxConcurrentCalls int                 `json:"max_concurrent_calls"`
	VolumeThreshold    int                 `json:"volume_threshold"`
	ErrorRateThreshold float64             `json:"error_rate_threshold"`

	// Runtime state
	FailureCount    int        `json:"failure_count"`
	SuccessCount    int        `json:"success_count"`
	LastFailureTime *time.Time `json:"last_failure_time,omitempty"`
	NextAttemptTime *time.Time `json:"next_attempt_time,omitempty"`
	CurrentCalls    int        `json:"current_calls"`
	TotalCalls      int64      `json:"total_calls"`
	TotalFailures   int64      `json:"total_failures"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	mutex sync.RWMutex
}

type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half_open"
)

type CircuitBreakerStatus struct {
	Name         string              `json:"name"`
	State        CircuitBreakerState `json:"state"`
	FailureCount int                 `json:"failure_count"`
	SuccessCount int                 `json:"success_count"`
	ErrorRate    float64             `json:"error_rate"`
	LastFailure  *time.Time          `json:"last_failure,omitempty"`
}

type CircuitBreakerMetrics struct {
	Name                string              `json:"name"`
	State               CircuitBreakerState `json:"state"`
	TotalCalls          int64               `json:"total_calls"`
	TotalFailures       int64               `json:"total_failures"`
	TotalSuccesses      int64               `json:"total_successes"`
	ErrorRate           float64             `json:"error_rate"`
	SuccessRate         float64             `json:"success_rate"`
	AverageResponseTime time.Duration       `json:"average_response_time"`
	RecentCalls         []CallResult        `json:"recent_calls"`
}

type CallResult struct {
	Timestamp time.Time     `json:"timestamp"`
	Success   bool          `json:"success"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// RegisterCircuitBreaker creates a new circuit breaker
func (s *CircuitBreakerService) RegisterCircuitBreaker(ctx context.Context, req *CircuitBreakerRequest) (*CircuitBreaker, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.circuitBreakers[req.Name]; exists {
		return nil, fmt.Errorf("circuit breaker %s already exists", req.Name)
	}

	cb := &CircuitBreaker{
		Name:               req.Name,
		State:              StateClosed,
		FailureThreshold:   req.FailureThreshold,
		SuccessThreshold:   req.SuccessThreshold,
		Timeout:            req.Timeout,
		MaxConcurrentCalls: req.MaxConcurrentCalls,
		VolumeThreshold:    req.VolumeThreshold,
		ErrorRateThreshold: req.ErrorRateThreshold,
		FailureCount:       0,
		SuccessCount:       0,
		CurrentCalls:       0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	s.circuitBreakers[req.Name] = cb

	s.publishCircuitBreakerEvent(ctx, "circuit_breaker.registered", cb.Name, cb.State)

	return cb, nil
}

// GetCircuitBreaker retrieves a circuit breaker by name
func (s *CircuitBreakerService) GetCircuitBreaker(ctx context.Context, name string) (*CircuitBreaker, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cb, exists := s.circuitBreakers[name]
	if !exists {
		return nil, fmt.Errorf("circuit breaker %s not found", name)
	}

	return cb, nil
}

// ExecuteWithCircuitBreaker executes an operation with circuit breaker protection
func (s *CircuitBreakerService) ExecuteWithCircuitBreaker(ctx context.Context, name string, operation func() error) error {
	cb, err := s.GetCircuitBreaker(ctx, name)
	if err != nil {
		return err
	}

	// Check if circuit breaker allows the call
	if !s.canExecute(cb) {
		return fmt.Errorf("circuit breaker %s is open", name)
	}

	// Increment concurrent calls
	cb.mutex.Lock()
	cb.CurrentCalls++
	cb.TotalCalls++
	cb.mutex.Unlock()

	start := time.Now()

	// Execute the operation
	err = operation()

	duration := time.Since(start)

	// Update circuit breaker state based on result
	cb.mutex.Lock()
	cb.CurrentCalls--

	if err != nil {
		cb.FailureCount++
		cb.TotalFailures++
		now := time.Now()
		cb.LastFailureTime = &now
		s.checkStateTransition(cb)
	} else {
		cb.SuccessCount++
		s.checkStateTransition(cb)
	}

	cb.UpdatedAt = time.Now()
	cb.mutex.Unlock()

	// Publish metrics event
	s.publishCallEvent(ctx, name, err == nil, duration)

	return err
}

// Helper methods
func (s *CircuitBreakerService) canExecute(cb *CircuitBreaker) bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.State {
	case StateClosed:
		return cb.CurrentCalls < cb.MaxConcurrentCalls
	case StateOpen:
		if cb.NextAttemptTime != nil && time.Now().After(*cb.NextAttemptTime) {
			cb.State = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return cb.CurrentCalls == 0 // Only allow one call in half-open state
	}

	return false
}

func (s *CircuitBreakerService) checkStateTransition(cb *CircuitBreaker) {
	switch cb.State {
	case StateClosed:
		if cb.FailureCount >= cb.FailureThreshold {
			cb.State = StateOpen
			nextAttempt := time.Now().Add(cb.Timeout)
			cb.NextAttemptTime = &nextAttempt
			s.publishCircuitBreakerEvent(context.Background(), "circuit_breaker.opened", cb.Name, cb.State)
		}
	case StateHalfOpen:
		if cb.SuccessCount >= cb.SuccessThreshold {
			cb.State = StateClosed
			cb.FailureCount = 0
			cb.SuccessCount = 0
			s.publishCircuitBreakerEvent(context.Background(), "circuit_breaker.closed", cb.Name, cb.State)
		} else if cb.FailureCount > 0 {
			cb.State = StateOpen
			nextAttempt := time.Now().Add(cb.Timeout)
			cb.NextAttemptTime = &nextAttempt
			s.publishCircuitBreakerEvent(context.Background(), "circuit_breaker.reopened", cb.Name, cb.State)
		}
	}
}

func (s *CircuitBreakerService) publishCircuitBreakerEvent(ctx context.Context, eventType, name string, state CircuitBreakerState) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   eventType,
		Source: "circuit-breaker-service",
		Data: map[string]interface{}{
			"circuit_breaker": name,
			"state":           state,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish circuit breaker event: %v\n", err)
	}
}

func (s *CircuitBreakerService) publishCallEvent(ctx context.Context, name string, success bool, duration time.Duration) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   "circuit_breaker.call_executed",
		Source: "circuit-breaker-service",
		Data: map[string]interface{}{
			"circuit_breaker": name,
			"success":         success,
			"duration_ms":     duration.Milliseconds(),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish call event: %v\n", err)
	}
}

// UpdateCircuitBreaker updates circuit breaker configuration
func (s *CircuitBreakerService) UpdateCircuitBreaker(ctx context.Context, req *UpdateCircuitBreakerRequest) error {
	cb, err := s.GetCircuitBreaker(ctx, req.Name)
	if err != nil {
		return err
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if req.FailureThreshold != nil {
		cb.FailureThreshold = *req.FailureThreshold
	}
	if req.SuccessThreshold != nil {
		cb.SuccessThreshold = *req.SuccessThreshold
	}
	if req.Timeout != nil {
		cb.Timeout = *req.Timeout
	}
	if req.MaxConcurrentCalls != nil {
		cb.MaxConcurrentCalls = *req.MaxConcurrentCalls
	}

	cb.UpdatedAt = time.Now()

	return nil
}

// TripCircuitBreaker manually trips a circuit breaker
func (s *CircuitBreakerService) TripCircuitBreaker(ctx context.Context, name string, _ string) error {
	cb, err := s.GetCircuitBreaker(ctx, name)
	if err != nil {
		return err
	}

	cb.mutex.Lock()
	cb.State = StateOpen
	nextAttempt := time.Now().Add(cb.Timeout)
	cb.NextAttemptTime = &nextAttempt
	cb.UpdatedAt = time.Now()
	cb.mutex.Unlock()

	s.publishCircuitBreakerEvent(ctx, "circuit_breaker.manual_trip", name, StateOpen)

	return nil
}

// ResetCircuitBreaker manually resets a circuit breaker
func (s *CircuitBreakerService) ResetCircuitBreaker(ctx context.Context, name string) error {
	cb, err := s.GetCircuitBreaker(ctx, name)
	if err != nil {
		return err
	}

	cb.mutex.Lock()
	cb.State = StateClosed
	cb.FailureCount = 0
	cb.SuccessCount = 0
	cb.NextAttemptTime = nil
	cb.UpdatedAt = time.Now()
	cb.mutex.Unlock()

	s.publishCircuitBreakerEvent(ctx, "circuit_breaker.manual_reset", name, StateClosed)

	return nil
}

// GetCircuitBreakerStatus returns status of all circuit breakers
func (s *CircuitBreakerService) GetCircuitBreakerStatus(ctx context.Context) ([]*CircuitBreakerStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	statuses := make([]*CircuitBreakerStatus, 0, len(s.circuitBreakers))

	for _, cb := range s.circuitBreakers {
		cb.mutex.RLock()

		var errorRate float64
		if cb.TotalCalls > 0 {
			errorRate = float64(cb.TotalFailures) / float64(cb.TotalCalls) * 100
		}

		status := &CircuitBreakerStatus{
			Name:         cb.Name,
			State:        cb.State,
			FailureCount: cb.FailureCount,
			SuccessCount: cb.SuccessCount,
			ErrorRate:    errorRate,
			LastFailure:  cb.LastFailureTime,
		}

		statuses = append(statuses, status)
		cb.mutex.RUnlock()
	}

	return statuses, nil
}

// GetCircuitBreakerMetrics returns detailed metrics for a specific circuit breaker
func (s *CircuitBreakerService) GetCircuitBreakerMetrics(ctx context.Context, name string) (*CircuitBreakerMetrics, error) {
	cb, err := s.GetCircuitBreaker(ctx, name)
	if err != nil {
		return nil, err
	}

	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	var errorRate, successRate float64
	if cb.TotalCalls > 0 {
		errorRate = float64(cb.TotalFailures) / float64(cb.TotalCalls) * 100
		successRate = float64(cb.TotalCalls-cb.TotalFailures) / float64(cb.TotalCalls) * 100
	}

	metrics := &CircuitBreakerMetrics{
		Name:           cb.Name,
		State:          cb.State,
		TotalCalls:     cb.TotalCalls,
		TotalFailures:  cb.TotalFailures,
		TotalSuccesses: cb.TotalCalls - cb.TotalFailures,
		ErrorRate:      errorRate,
		SuccessRate:    successRate,
		RecentCalls:    []CallResult{}, // In a real implementation, this would store recent call history
	}

	return metrics, nil
}
