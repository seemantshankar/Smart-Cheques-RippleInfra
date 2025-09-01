package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/services"
)

// ErrorHandlingHandler handles HTTP requests for error handling operations
type ErrorHandlingHandler struct {
	errorHandlingService services.ErrorHandlingServiceInterface
}

// NewErrorHandlingHandler creates a new handler for error handling
func NewErrorHandlingHandler(errorHandlingService services.ErrorHandlingServiceInterface) *ErrorHandlingHandler {
	return &ErrorHandlingHandler{
		errorHandlingService: errorHandlingService,
	}
}

// RegisterRoutes registers the routes for error handling endpoints
func (h *ErrorHandlingHandler) RegisterRoutes(router *gin.RouterGroup) {
	errors := router.Group("/error-handling")
	{
		// Error handling operations
		errors.POST("/handle", h.HandleError)
		errors.POST("/retry/:operationId", h.RetryOperation)

		// Circuit breaker operations
		errors.GET("/circuit-breaker/:serviceName/status", h.GetCircuitBreakerStatus)
		errors.POST("/circuit-breaker/:serviceName/reset", h.ResetCircuitBreaker)

		// Dead letter queue operations
		errors.POST("/dead-letter", h.AddToDeadLetterQueue)
		errors.POST("/dead-letter/process", h.ProcessDeadLetterQueue)
		errors.GET("/dead-letter", h.GetDeadLetterQueueItems)

		// Recovery operations
		errors.POST("/recovery/:operationId/:strategy", h.ExecuteRecoveryStrategy)

		// Monitoring operations
		errors.GET("/metrics", h.GetErrorMetrics)
		errors.GET("/trends", h.GetErrorTrends)
	}
}

// HandleError handles an error and creates an error context
func (h *ErrorHandlingHandler) HandleError(c *gin.Context) {
	var req struct {
		Operation string                 `json:"operation" validate:"required"`
		Error     string                 `json:"error" validate:"required"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Create a mock error for demonstration
	mockErr := fmt.Errorf(req.Error)
	errorContext := h.errorHandlingService.HandleError(ctx, req.Operation, mockErr, req.Metadata)

	c.JSON(http.StatusOK, errorContext)
}

// RetryOperation retries an operation with exponential backoff
func (h *ErrorHandlingHandler) RetryOperation(c *gin.Context) {
	operationIDStr := c.Param("operationId")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Mock retry function for demonstration
	retryFunc := func() error {
		// Simulate an operation that might fail
		return fmt.Errorf("simulated operation failure")
	}

	if err := h.errorHandlingService.RetryOperation(ctx, operationID, retryFunc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retry operation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Operation retry completed",
		"operation_id": operationID,
	})
}

// GetCircuitBreakerStatus gets the status of a circuit breaker
func (h *ErrorHandlingHandler) GetCircuitBreakerStatus(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status := h.errorHandlingService.GetCircuitBreakerStatus(ctx, serviceName)

	c.JSON(http.StatusOK, status)
}

// ResetCircuitBreaker resets a circuit breaker to closed state
func (h *ErrorHandlingHandler) ResetCircuitBreaker(c *gin.Context) {
	serviceName := c.Param("serviceName")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.errorHandlingService.ResetCircuitBreaker(ctx, serviceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset circuit breaker",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Circuit breaker reset successfully",
		"service_name": serviceName,
	})
}

// AddToDeadLetterQueue adds an item to the dead letter queue
func (h *ErrorHandlingHandler) AddToDeadLetterQueue(c *gin.Context) {
	var req struct {
		Operation string      `json:"operation" validate:"required"`
		Error     string      `json:"error" validate:"required"`
		Data      interface{} `json:"data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	mockErr := fmt.Errorf(req.Error)
	if err := h.errorHandlingService.AddToDeadLetterQueue(ctx, req.Operation, req.Data, mockErr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to add to dead letter queue",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Added to dead letter queue",
		"operation": req.Operation,
	})
}

// ProcessDeadLetterQueue processes items in the dead letter queue
func (h *ErrorHandlingHandler) ProcessDeadLetterQueue(c *gin.Context) {
	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	if err := h.errorHandlingService.ProcessDeadLetterQueue(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process dead letter queue",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Dead letter queue processed successfully",
	})
}

// GetDeadLetterQueueItems gets items from the dead letter queue
func (h *ErrorHandlingHandler) GetDeadLetterQueueItems(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	items, err := h.errorHandlingService.GetDeadLetterQueueItems(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get dead letter queue items",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
		"limit": limit,
	})
}

// ExecuteRecoveryStrategy executes a recovery strategy for a failed operation
func (h *ErrorHandlingHandler) ExecuteRecoveryStrategy(c *gin.Context) {
	operationIDStr := c.Param("operationId")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid operation ID format",
		})
		return
	}

	strategyStr := c.Param("strategy")
	var strategy services.RecoveryStrategy
	switch strategyStr {
	case "retry":
		strategy = services.RecoveryStrategyRetry
	case "circuit_break":
		strategy = services.RecoveryStrategyCircuitBreak
	case "fallback":
		strategy = services.RecoveryStrategyFallback
	case "manual_review":
		strategy = services.RecoveryStrategyManualReview
	case "ignore":
		strategy = services.RecoveryStrategyIgnore
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid recovery strategy",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	if err := h.errorHandlingService.ExecuteRecoveryStrategy(ctx, strategy, operationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute recovery strategy",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Recovery strategy executed successfully",
		"operation_id": operationID,
		"strategy":     strategyStr,
	})
}

// GetErrorMetrics gets error metrics for a timeframe
func (h *ErrorHandlingHandler) GetErrorMetrics(c *gin.Context) {
	timeframeStr := c.DefaultQuery("timeframe", "1h")
	var timeframe time.Duration

	switch timeframeStr {
	case "1h":
		timeframe = time.Hour
	case "24h":
		timeframe = 24 * time.Hour
	case "7d":
		timeframe = 7 * 24 * time.Hour
	default:
		timeframe = time.Hour
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	metrics, err := h.errorHandlingService.GetErrorMetrics(ctx, timeframe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get error metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetErrorTrends gets error trends over a date range
func (h *ErrorHandlingHandler) GetErrorTrends(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		// Use default range (last 7 days)
		endDate := time.Now()
		startDate := endDate.Add(-7 * 24 * time.Hour)

		ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		trends, err := h.errorHandlingService.GetErrorTrends(ctx, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get error trends",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"trends":     trends,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return
	}

	// Parse provided dates (simplified for demo)
	endDate := time.Now()
	startDate := endDate.Add(-7 * 24 * time.Hour)

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	trends, err := h.errorHandlingService.GetErrorTrends(ctx, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get error trends",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trends":     trends,
		"start_date": startDate,
		"end_date":   endDate,
	})
}
