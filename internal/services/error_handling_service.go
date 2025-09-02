package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// ErrorHandlingServiceInterface defines the interface for error handling operations
type ErrorHandlingServiceInterface interface {
	// Error Handling
	HandleError(ctx context.Context, operation string, err error, metadata map[string]interface{}) *ErrorContext
	RetryOperation(ctx context.Context, operationID uuid.UUID, retryFunc RetryableOperation) error

	// Circuit Breaker
	GetCircuitBreakerStatus(ctx context.Context, serviceName string) *ErrorHandlingCircuitBreakerStatus
	ResetCircuitBreaker(ctx context.Context, serviceName string) error

	// Dead Letter Queue
	AddToDeadLetterQueue(ctx context.Context, operation string, data interface{}, err error) error
	ProcessDeadLetterQueue(ctx context.Context) error
	GetDeadLetterQueueItems(ctx context.Context, limit int) ([]*DeadLetterItem, error)

	// Recovery Strategies
	ExecuteRecoveryStrategy(ctx context.Context, strategy RecoveryStrategy, operationID uuid.UUID) error

	// Monitoring
	GetErrorMetrics(ctx context.Context, timeframe time.Duration) (*ErrorMetrics, error)
	GetErrorTrends(ctx context.Context, startDate, endDate time.Time) ([]*ErrorTrend, error)
}

// ErrorHandlingService implements the error handling service interface
type ErrorHandlingService struct {
	messagingClient    messaging.EventBus
	errorConfig        *ErrorHandlingConfig
	circuitBreakers    map[string]*ErrorHandlingCircuitBreaker
	deadLetterQueue    []*DeadLetterItem
	errorHistory       []*ErrorContext
	retryStrategies    map[string]RetryStrategy
	recoveryStrategies map[string]RecoveryStrategy
	mutex              sync.RWMutex
	queueMutex         sync.RWMutex
}

// NewErrorHandlingService creates a new error handling service instance
func NewErrorHandlingService(
	messagingClient messaging.EventBus,
	config *ErrorHandlingConfig,
) ErrorHandlingServiceInterface {
	service := &ErrorHandlingService{
		messagingClient:    messagingClient,
		errorConfig:        config,
		circuitBreakers:    make(map[string]*ErrorHandlingCircuitBreaker),
		deadLetterQueue:    make([]*DeadLetterItem, 0),
		errorHistory:       make([]*ErrorContext, 0),
		retryStrategies:    make(map[string]RetryStrategy),
		recoveryStrategies: make(map[string]RecoveryStrategy),
	}

	// Initialize default retry strategies
	service.initializeDefaultStrategies()

	// Start background processing if enabled
	if config.EnableBackgroundProcessing {
		go service.startBackgroundProcessing()
	}

	return service
}

// ErrorHandlingConfig defines configuration for error handling
type ErrorHandlingConfig struct {
	// Processing settings
	EnableBackgroundProcessing bool          `json:"enable_background_processing"`
	ProcessingInterval         time.Duration `json:"processing_interval"`
	MaxDeadLetterQueueSize     int           `json:"max_dead_letter_queue_size"`

	// Retry settings
	DefaultMaxRetries  int           `json:"default_max_retries"`
	DefaultRetryDelay  time.Duration `json:"default_retry_delay"`
	MaxRetryDelay      time.Duration `json:"max_retry_delay"`
	RetryBackoffFactor float64       `json:"retry_backoff_factor"`

	// Circuit breaker settings
	DefaultFailureThreshold int           `json:"default_failure_threshold"`
	DefaultRecoveryTimeout  time.Duration `json:"default_recovery_timeout"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`

	// Monitoring settings
	ErrorHistoryRetention     time.Duration `json:"error_history_retention"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
}

// ErrorContext represents the context of an error
type ErrorContext struct {
	ID          uuid.UUID              `json:"id"`
	Operation   string                 `json:"operation"`
	Error       error                  `json:"error"`
	Message     string                 `json:"message"`
	Code        ErrorCode              `json:"code"`
	Severity    ErrorSeverity          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	ServiceName string                 `json:"service_name"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	RequestID   *uuid.UUID             `json:"request_id,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// ErrorCode represents error codes
type ErrorCode string

const (
	ErrorCodeNetworkFailure        ErrorCode = "network_failure"
	ErrorCodeTimeout               ErrorCode = "timeout"
	ErrorCodeValidationFailure     ErrorCode = "validation_failure"
	ErrorCodeAuthenticationFailure ErrorCode = "authentication_failure"
	ErrorCodeAuthorizationFailure  ErrorCode = "authorization_failure"
	ErrorCodeResourceNotFound      ErrorCode = "resource_not_found"
	ErrorCodeResourceExhausted     ErrorCode = "resource_exhausted"
	ErrorCodeInternalError         ErrorCode = "internal_error"
	ErrorCodeExternalServiceError  ErrorCode = "external_service_error"
	ErrorCodeBlockchainError       ErrorCode = "blockchain_error"
	ErrorCodeDatabaseError         ErrorCode = "database_error"
	ErrorCodeConfigurationError    ErrorCode = "configuration_error"
)

// ErrorSeverity represents error severity levels
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorHandlingCircuitBreaker represents a circuit breaker for service protection
type ErrorHandlingCircuitBreaker struct {
	ServiceName      string                           `json:"service_name"`
	State            ErrorHandlingCircuitBreakerState `json:"state"`
	FailureCount     int                              `json:"failure_count"`
	LastFailureTime  *time.Time                       `json:"last_failure_time,omitempty"`
	NextRetryTime    *time.Time                       `json:"next_retry_time,omitempty"`
	FailureThreshold int                              `json:"failure_threshold"`
	RecoveryTimeout  time.Duration                    `json:"recovery_timeout"`
	mutex            sync.RWMutex
}

// ErrorHandlingCircuitBreakerState represents circuit breaker states
type ErrorHandlingCircuitBreakerState string

const (
	ErrorHandlingCircuitBreakerStateClosed   ErrorHandlingCircuitBreakerState = "closed"
	ErrorHandlingCircuitBreakerStateOpen     ErrorHandlingCircuitBreakerState = "open"
	ErrorHandlingCircuitBreakerStateHalfOpen ErrorHandlingCircuitBreakerState = "half_open"
)

// ErrorHandlingCircuitBreakerStatus represents the status of a circuit breaker
type ErrorHandlingCircuitBreakerStatus struct {
	ServiceName     string                           `json:"service_name"`
	State           ErrorHandlingCircuitBreakerState `json:"state"`
	FailureCount    int                              `json:"failure_count"`
	LastFailureTime *time.Time                       `json:"last_failure_time,omitempty"`
	NextRetryTime   *time.Time                       `json:"next_retry_time,omitempty"`
	IsAvailable     bool                             `json:"is_available"`
}

// DeadLetterItem represents an item in the dead letter queue
type DeadLetterItem struct {
	ID           uuid.UUID              `json:"id"`
	Operation    string                 `json:"operation"`
	Data         interface{}            `json:"data"`
	Error        error                  `json:"error"`
	ErrorContext *ErrorContext          `json:"error_context"`
	CreatedAt    time.Time              `json:"created_at"`
	RetryCount   int                    `json:"retry_count"`
	LastRetryAt  *time.Time             `json:"last_retry_at,omitempty"`
	NextRetryAt  *time.Time             `json:"next_retry_at,omitempty"`
	Status       DeadLetterStatus       `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DeadLetterStatus represents dead letter queue item status
type DeadLetterStatus string

const (
	DeadLetterStatusPending  DeadLetterStatus = "pending"
	DeadLetterStatusRetrying DeadLetterStatus = "retrying"
	DeadLetterStatusFailed   DeadLetterStatus = "failed"
	DeadLetterStatusResolved DeadLetterStatus = "resolved"
)

// RetryStrategy represents a retry strategy
type RetryStrategy struct {
	Name          string        `json:"name"`
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	JitterEnabled bool          `json:"jitter_enabled"`
}

// RecoveryStrategy represents a recovery strategy
type RecoveryStrategy string

const (
	RecoveryStrategyRetry        RecoveryStrategy = "retry"
	RecoveryStrategyCircuitBreak RecoveryStrategy = "circuit_break"
	RecoveryStrategyFallback     RecoveryStrategy = "fallback"
	RecoveryStrategyManualReview RecoveryStrategy = "manual_review"
	RecoveryStrategyIgnore       RecoveryStrategy = "ignore"
)

// RetryableOperation represents a function that can be retried
type RetryableOperation func() error

// ErrorMetrics represents error metrics
type ErrorMetrics struct {
	TimeRange         time.Duration           `json:"time_range"`
	TotalErrors       int64                   `json:"total_errors"`
	ErrorsByCode      map[ErrorCode]int64     `json:"errors_by_code"`
	ErrorsByService   map[string]int64        `json:"errors_by_service"`
	ErrorsBySeverity  map[ErrorSeverity]int64 `json:"errors_by_severity"`
	ErrorRate         float64                 `json:"error_rate"`
	RecoveryRate      float64                 `json:"recovery_rate"`
	AvgResolutionTime float64                 `json:"avg_resolution_time"`
}

// ErrorTrend represents error trends over time
type ErrorTrend struct {
	Date          time.Time `json:"date"`
	ErrorCount    int64     `json:"error_count"`
	ErrorRate     float64   `json:"error_rate"`
	TopErrorCodes []string  `json:"top_error_codes"`
	RecoveryRate  float64   `json:"recovery_rate"`
}

// HandleError handles an error and creates an error context
func (s *ErrorHandlingService) HandleError(ctx context.Context, operation string, err error, metadata map[string]interface{}) *ErrorContext {
	errorContext := &ErrorContext{
		ID:         uuid.New(),
		Operation:  operation,
		Error:      err,
		Message:    err.Error(),
		Code:       s.classifyError(err),
		Severity:   s.determineSeverity(err),
		Timestamp:  time.Now(),
		Metadata:   metadata,
		RetryCount: 0,
		Resolved:   false,
	}

	// Extract service name from operation
	if serviceName := s.extractServiceName(operation); serviceName != "" {
		errorContext.ServiceName = serviceName
	}

	// Extract user and request IDs from metadata if available
	if userIDStr, ok := metadata["user_id"].(string); ok {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			errorContext.UserID = &userID
		}
	}

	if requestIDStr, ok := metadata["request_id"].(string); ok {
		if requestID, err := uuid.Parse(requestIDStr); err == nil {
			errorContext.RequestID = &requestID
		}
	}

	// Store error in history
	s.mutex.Lock()
	s.errorHistory = append(s.errorHistory, errorContext)
	s.mutex.Unlock()

	// Update circuit breaker
	s.updateCircuitBreaker(errorContext.ServiceName, err)

	// Publish error event
	s.publishErrorEvent(ctx, "error.occurred", errorContext)

	log.Printf("Error handled: %s - %s (%s)", operation, err.Error(), errorContext.Code)

	return errorContext
}

// RetryOperation retries an operation with exponential backoff
func (s *ErrorHandlingService) RetryOperation(ctx context.Context, operationID uuid.UUID, retryFunc RetryableOperation) error {
	strategy, exists := s.retryStrategies["default"]
	if !exists {
		strategy = RetryStrategy{
			MaxRetries:    s.errorConfig.DefaultMaxRetries,
			InitialDelay:  s.errorConfig.DefaultRetryDelay,
			MaxDelay:      s.errorConfig.MaxRetryDelay,
			BackoffFactor: s.errorConfig.RetryBackoffFactor,
		}
	}

	var lastErr error
	delay := strategy.InitialDelay

	for attempt := 0; attempt <= strategy.MaxRetries; attempt++ {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the operation
		if err := retryFunc(); err != nil {
			lastErr = err

			// If this is the last attempt, don't wait
			if attempt == strategy.MaxRetries {
				break
			}

			// Wait before retrying
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * strategy.BackoffFactor)
				if delay > strategy.MaxDelay {
					delay = strategy.MaxDelay
				}
			}

			continue
		}

		// Success - operation completed
		return nil
	}

	// All retries failed - add to dead letter queue
	return s.AddToDeadLetterQueue(ctx, operationID.String(), nil, lastErr)
}

// GetCircuitBreakerStatus gets the status of a circuit breaker
func (s *ErrorHandlingService) GetCircuitBreakerStatus(ctx context.Context, serviceName string) *ErrorHandlingCircuitBreakerStatus {
	s.mutex.RLock()
	cb, exists := s.circuitBreakers[serviceName]
	s.mutex.RUnlock()

	if !exists {
		return &ErrorHandlingCircuitBreakerStatus{
			ServiceName: serviceName,
			State:       ErrorHandlingCircuitBreakerStateClosed,
			IsAvailable: true,
		}
	}

	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	status := &ErrorHandlingCircuitBreakerStatus{
		ServiceName:     cb.ServiceName,
		State:           cb.State,
		FailureCount:    cb.FailureCount,
		LastFailureTime: cb.LastFailureTime,
		NextRetryTime:   cb.NextRetryTime,
		IsAvailable:     cb.State == ErrorHandlingCircuitBreakerStateClosed,
	}

	return status
}

// ResetCircuitBreaker resets a circuit breaker to closed state
func (s *ErrorHandlingService) ResetCircuitBreaker(ctx context.Context, serviceName string) error {
	s.mutex.RLock()
	cb, exists := s.circuitBreakers[serviceName]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("circuit breaker not found for service: %s", serviceName)
	}

	cb.mutex.Lock()
	cb.State = ErrorHandlingCircuitBreakerStateClosed
	cb.FailureCount = 0
	cb.LastFailureTime = nil
	cb.NextRetryTime = nil
	cb.mutex.Unlock()

	log.Printf("Circuit breaker reset for service: %s", serviceName)
	return nil
}

// AddToDeadLetterQueue adds an item to the dead letter queue
func (s *ErrorHandlingService) AddToDeadLetterQueue(ctx context.Context, operation string, data interface{}, err error) error {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	// Check queue size limit
	if len(s.deadLetterQueue) >= s.errorConfig.MaxDeadLetterQueueSize {
		// Remove oldest item
		s.deadLetterQueue = s.deadLetterQueue[1:]
	}

	item := &DeadLetterItem{
		ID:         uuid.New(),
		Operation:  operation,
		Data:       data,
		Error:      err,
		CreatedAt:  time.Now(),
		RetryCount: 0,
		Status:     DeadLetterStatusPending,
	}

	s.deadLetterQueue = append(s.deadLetterQueue, item)

	log.Printf("Added item to dead letter queue: %s - %s", operation, err.Error())

	// Publish dead letter event
	s.publishDeadLetterEvent(ctx, "dead.letter.added", item)

	return nil
}

// ProcessDeadLetterQueue processes items in the dead letter queue
func (s *ErrorHandlingService) ProcessDeadLetterQueue(ctx context.Context) error {
	s.queueMutex.Lock()
	defer s.queueMutex.Unlock()

	processed := 0
	for _, item := range s.deadLetterQueue {
		if item.Status != DeadLetterStatusPending {
			continue
		}

		// Try to reprocess the item
		if err := s.reprocessDeadLetterItem(ctx, item); err != nil {
			log.Printf("Failed to reprocess dead letter item %s: %v", item.ID, err)
			continue
		}

		item.Status = DeadLetterStatusResolved
		processed++
	}

	log.Printf("Processed %d items from dead letter queue", processed)
	return nil
}

// GetDeadLetterQueueItems gets items from the dead letter queue
func (s *ErrorHandlingService) GetDeadLetterQueueItems(ctx context.Context, limit int) ([]*DeadLetterItem, error) {
	s.queueMutex.RLock()
	defer s.queueMutex.RUnlock()

	if limit <= 0 || limit > len(s.deadLetterQueue) {
		limit = len(s.deadLetterQueue)
	}

	items := make([]*DeadLetterItem, limit)
	copy(items, s.deadLetterQueue[len(s.deadLetterQueue)-limit:])

	return items, nil
}

// ExecuteRecoveryStrategy executes a recovery strategy for a failed operation
func (s *ErrorHandlingService) ExecuteRecoveryStrategy(ctx context.Context, strategy RecoveryStrategy, operationID uuid.UUID) error {
	log.Printf("Executing recovery strategy %s for operation %s", strategy, operationID)

	switch strategy {
	case RecoveryStrategyRetry:
		return s.retryFromDeadLetterQueue(ctx, operationID)
	case RecoveryStrategyCircuitBreak:
		return s.triggerCircuitBreaker(ctx, operationID)
	case RecoveryStrategyFallback:
		return s.executeFallbackStrategy(ctx, operationID)
	case RecoveryStrategyManualReview:
		return s.markForManualReview(ctx, operationID)
	case RecoveryStrategyIgnore:
		return s.ignoreOperation(ctx, operationID)
	default:
		return fmt.Errorf("unknown recovery strategy: %s", strategy)
	}
}

// GetErrorMetrics gets error metrics for a timeframe
func (s *ErrorHandlingService) GetErrorMetrics(ctx context.Context, timeframe time.Duration) (*ErrorMetrics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cutoff := time.Now().Add(-timeframe)
	metrics := &ErrorMetrics{
		TimeRange:        timeframe,
		ErrorsByCode:     make(map[ErrorCode]int64),
		ErrorsByService:  make(map[string]int64),
		ErrorsBySeverity: make(map[ErrorSeverity]int64),
	}

	for _, errCtx := range s.errorHistory {
		if errCtx.Timestamp.After(cutoff) {
			metrics.TotalErrors++
			metrics.ErrorsByCode[errCtx.Code]++
			metrics.ErrorsByService[errCtx.ServiceName]++
			metrics.ErrorsBySeverity[errCtx.Severity]++
		}
	}

	// Calculate rates (simplified)
	if timeframe > 0 {
		hours := timeframe.Hours()
		metrics.ErrorRate = float64(metrics.TotalErrors) / hours
	}

	return metrics, nil
}

// GetErrorTrends gets error trends over a date range
func (s *ErrorHandlingService) GetErrorTrends(ctx context.Context, startDate, endDate time.Time) ([]*ErrorTrend, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	trends := make([]*ErrorTrend, 0)
	current := startDate

	for current.Before(endDate) {
		next := current.Add(24 * time.Hour)
		errorCount := int64(0)

		for _, errCtx := range s.errorHistory {
			if errCtx.Timestamp.After(current) && errCtx.Timestamp.Before(next) {
				errorCount++
			}
		}

		trend := &ErrorTrend{
			Date:       current,
			ErrorCount: errorCount,
			ErrorRate:  float64(errorCount) / 24.0, // errors per hour
		}

		trends = append(trends, trend)
		current = next
	}

	return trends, nil
}

// Helper methods

func (s *ErrorHandlingService) initializeDefaultStrategies() {
	s.retryStrategies["default"] = RetryStrategy{
		Name:          "default",
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: true,
	}

	s.retryStrategies["aggressive"] = RetryStrategy{
		Name:          "aggressive",
		MaxRetries:    5,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 1.5,
		JitterEnabled: true,
	}

	s.retryStrategies["conservative"] = RetryStrategy{
		Name:          "conservative",
		MaxRetries:    2,
		InitialDelay:  5 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 3.0,
		JitterEnabled: false,
	}
}

func (s *ErrorHandlingService) startBackgroundProcessing() {
	log.Printf("Starting error handling background processing")

	ticker := time.NewTicker(s.errorConfig.ProcessingInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()

		// Process dead letter queue
		if err := s.ProcessDeadLetterQueue(ctx); err != nil {
			log.Printf("Error processing dead letter queue: %v", err)
		}

		// Clean up old error history
		s.cleanupErrorHistory()

		// Update circuit breakers
		s.updateCircuitBreakers(ctx)
	}
}

func (s *ErrorHandlingService) classifyError(err error) ErrorCode {
	if err == nil {
		return ErrorCodeInternalError
	}

	errStr := err.Error()

	// Classify based on error message patterns
	switch {
	case containsString(errStr, "timeout"):
		return ErrorCodeTimeout
	case containsString(errStr, "network") || containsString(errStr, "connection"):
		return ErrorCodeNetworkFailure
	case containsString(errStr, "unauthorized") || containsString(errStr, "forbidden"):
		return ErrorCodeAuthorizationFailure
	case containsString(errStr, "not found"):
		return ErrorCodeResourceNotFound
	case containsString(errStr, "validation"):
		return ErrorCodeValidationFailure
	case containsString(errStr, "blockchain") || containsString(errStr, "xrpl"):
		return ErrorCodeBlockchainError
	case containsString(errStr, "database") || containsString(errStr, "sql"):
		return ErrorCodeDatabaseError
	default:
		return ErrorCodeInternalError
	}
}

func (s *ErrorHandlingService) determineSeverity(err error) ErrorSeverity {
	code := s.classifyError(err)

	switch code {
	case ErrorCodeResourceExhausted:
		return ErrorSeverityCritical
	case ErrorCodeAuthorizationFailure, ErrorCodeBlockchainError, ErrorCodeDatabaseError:
		return ErrorSeverityHigh
	case ErrorCodeNetworkFailure, ErrorCodeTimeout:
		return ErrorSeverityMedium
	default:
		return ErrorSeverityLow
	}
}

func (s *ErrorHandlingService) extractServiceName(operation string) string {
	// Extract service name from operation string
	// Format: "service.operation" or "service.subservice.operation"
	parts := strings.Split(operation, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (s *ErrorHandlingService) updateCircuitBreaker(serviceName string, err error) {
	if err == nil {
		return // No error to handle
	}

	s.mutex.Lock()
	cb, exists := s.circuitBreakers[serviceName]
	if !exists {
		cb = &ErrorHandlingCircuitBreaker{
			ServiceName:      serviceName,
			State:            ErrorHandlingCircuitBreakerStateClosed,
			FailureCount:     0,
			FailureThreshold: s.errorConfig.DefaultFailureThreshold,
			RecoveryTimeout:  s.errorConfig.DefaultRecoveryTimeout,
		}
		s.circuitBreakers[serviceName] = cb
	}
	s.mutex.Unlock()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	cb.FailureCount++
	cb.LastFailureTime = &now

	// Check if circuit breaker should open
	if cb.FailureCount >= cb.FailureThreshold && cb.State == ErrorHandlingCircuitBreakerStateClosed {
		cb.State = ErrorHandlingCircuitBreakerStateOpen
		retryTime := now.Add(cb.RecoveryTimeout)
		cb.NextRetryTime = &retryTime
		log.Printf("Circuit breaker opened for service: %s due to error: %v", serviceName, err)
	}
}

func (s *ErrorHandlingService) updateCircuitBreakers(ctx context.Context) {
	// Check if context is canceled
	select {
	case <-ctx.Done():
		return
	default:
	}

	s.mutex.RLock()
	services := make([]string, 0, len(s.circuitBreakers))
	for serviceName := range s.circuitBreakers {
		services = append(services, serviceName)
	}
	s.mutex.RUnlock()

	for _, serviceName := range services {
		// Check if context is canceled before processing each service
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.mutex.RLock()
		cb := s.circuitBreakers[serviceName]
		s.mutex.RUnlock()

		if cb == nil {
			continue
		}

		cb.mutex.Lock()
		if cb.State == ErrorHandlingCircuitBreakerStateOpen && time.Now().After(*cb.NextRetryTime) {
			cb.State = ErrorHandlingCircuitBreakerStateHalfOpen
			log.Printf("Circuit breaker half-open for service: %s", serviceName)
		}
		cb.mutex.Unlock()
	}
}

func (s *ErrorHandlingService) reprocessDeadLetterItem(ctx context.Context, item *DeadLetterItem) error {
	// This is a simplified reprocessing implementation
	// In production, this would attempt to re-execute the original operation

	log.Printf("Reprocessing dead letter item: %s", item.ID)

	// Mark as retrying
	item.Status = DeadLetterStatusRetrying
	item.RetryCount++
	now := time.Now()
	item.LastRetryAt = &now

	// Simulate reprocessing with context timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Simulated processing completed
	}

	// For demonstration, we'll randomly succeed/fail
	// In production, this would be based on the actual operation result
	if item.RetryCount > 3 {
		item.Status = DeadLetterStatusFailed
		return fmt.Errorf("max retries exceeded for dead letter item: %s", item.ID)
	}

	// Simulate success
	item.Status = DeadLetterStatusResolved
	return nil
}

func (s *ErrorHandlingService) retryFromDeadLetterQueue(ctx context.Context, operationID uuid.UUID) error {
	// Find the item in dead letter queue and retry it
	s.queueMutex.RLock()
	defer s.queueMutex.RUnlock()

	for _, item := range s.deadLetterQueue {
		if item.ID == operationID {
			return s.reprocessDeadLetterItem(ctx, item)
		}
	}

	return fmt.Errorf("operation not found in dead letter queue: %s", operationID)
}

func (s *ErrorHandlingService) triggerCircuitBreaker(ctx context.Context, operationID uuid.UUID) error {
	// Check if context is canceled before proceeding
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find the operation and trigger circuit breaker for its service
	// This is a simplified implementation
	return fmt.Errorf("circuit breaker recovery not yet implemented for operation: %s", operationID)
}

func (s *ErrorHandlingService) executeFallbackStrategy(ctx context.Context, operationID uuid.UUID) error {
	// Check if context is canceled before proceeding
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Execute fallback strategy for the operation
	// This is a simplified implementation
	return fmt.Errorf("fallback recovery not yet implemented for operation: %s", operationID)
}

func (s *ErrorHandlingService) markForManualReview(ctx context.Context, operationID uuid.UUID) error {
	// Check if context is canceled before proceeding
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Mark the operation for manual review
	log.Printf("Operation %s marked for manual review", operationID)
	return nil
}

func (s *ErrorHandlingService) ignoreOperation(ctx context.Context, operationID uuid.UUID) error {
	// Check if context is canceled before proceeding
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Ignore the failed operation
	log.Printf("Operation %s ignored", operationID)
	return nil
}

func (s *ErrorHandlingService) cleanupErrorHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cutoff := time.Now().Add(-s.errorConfig.ErrorHistoryRetention)
	filtered := make([]*ErrorContext, 0)

	for _, errCtx := range s.errorHistory {
		if errCtx.Timestamp.After(cutoff) {
			filtered = append(filtered, errCtx)
		}
	}

	s.errorHistory = filtered
}

func (s *ErrorHandlingService) publishErrorEvent(ctx context.Context, eventType string, errorContext *ErrorContext) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "error-handling-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"error_id":      errorContext.ID,
			"operation":     errorContext.Operation,
			"error_code":    string(errorContext.Code),
			"severity":      string(errorContext.Severity),
			"service_name":  errorContext.ServiceName,
			"error_message": errorContext.Message,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		log.Printf("Failed to publish error event: %v", err)
	}
}

func (s *ErrorHandlingService) publishDeadLetterEvent(ctx context.Context, eventType string, item *DeadLetterItem) {
	event := &messaging.Event{
		Type:      eventType,
		Source:    "error-handling-service",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"item_id":     item.ID,
			"operation":   item.Operation,
			"error":       item.Error.Error(),
			"retry_count": item.RetryCount,
			"status":      string(item.Status),
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		log.Printf("Failed to publish dead letter event: %v", err)
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
