package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// PaymentExecutionHandler handles HTTP requests for payment execution operations
type PaymentExecutionHandler struct {
	paymentExecService services.PaymentExecutionServiceInterface
}

// NewPaymentExecutionHandler creates a new handler for payment execution
func NewPaymentExecutionHandler(paymentExecService services.PaymentExecutionServiceInterface) *PaymentExecutionHandler {
	return &PaymentExecutionHandler{
		paymentExecService: paymentExecService,
	}
}

// RegisterRoutes registers the routes for payment execution endpoints
func (h *PaymentExecutionHandler) RegisterRoutes(router *gin.RouterGroup) {
	executions := router.Group("/payment-executions")
	{
		// Payment execution operations
		executions.POST("", h.ExecutePayment)
		executions.POST("/from-authorization/:authId", h.ExecutePaymentFromAuthorization)
		executions.POST("/bulk", h.ExecuteBulkPayments)

		// Execution monitoring
		executions.GET("/:executionId/status", h.GetPaymentExecutionStatus)
		executions.GET("/:executionId/monitor", h.MonitorPaymentExecution)
		executions.GET("/request/:requestId/history", h.GetPaymentExecutionHistory)

		// Retry and recovery
		executions.POST("/:executionId/retry", h.RetryFailedPayment)
		executions.POST("/:executionId/cancel", h.CancelPaymentExecution)
		executions.PUT("/:executionId/status", h.UpdatePaymentExecutionStatus)

		// Condition management
		executions.POST("/fulfillment", h.GeneratePaymentFulfillment)
		executions.POST("/validate-condition", h.ValidatePaymentCondition)
	}
}

// ExecutePayment executes a payment based on a payment request ID
func (h *PaymentExecutionHandler) ExecutePayment(c *gin.Context) {
	var req struct {
		PaymentRequestID string `json:"payment_request_id" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	requestID, err := uuid.Parse(req.PaymentRequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment request ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := h.paymentExecService.ExecutePayment(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExecutePaymentFromAuthorization executes a payment from an approved authorization
func (h *PaymentExecutionHandler) ExecutePaymentFromAuthorization(c *gin.Context) {
	authIDStr := c.Param("authId")
	authID, err := uuid.Parse(authIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid authorization ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := h.paymentExecService.ExecutePaymentFromAuthorization(ctx, authID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute payment from authorization",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExecuteBulkPayments executes multiple payments in batch
func (h *PaymentExecutionHandler) ExecuteBulkPayments(c *gin.Context) {
	var req struct {
		PaymentRequestIDs []string `json:"payment_request_ids" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	requestIDs := make([]uuid.UUID, len(req.PaymentRequestIDs))
	for i, idStr := range req.PaymentRequestIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid payment request ID format",
				"details": err.Error(),
			})
			return
		}
		requestIDs[i] = id
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	result, err := h.paymentExecService.ExecuteBulkPayments(ctx, requestIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute bulk payments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPaymentExecutionStatus gets the current status of a payment execution
func (h *PaymentExecutionHandler) GetPaymentExecutionStatus(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentExecService.GetPaymentExecutionStatus(ctx, executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Payment execution status not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// MonitorPaymentExecution monitors the status of a payment execution
func (h *PaymentExecutionHandler) MonitorPaymentExecution(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentExecService.MonitorPaymentExecution(ctx, executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Payment execution not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetPaymentExecutionHistory gets the execution history for a payment request
func (h *PaymentExecutionHandler) GetPaymentExecutionHistory(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment request ID format",
		})
		return
	}

	limit, offset := h.getPaginationParams(c)

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	history, err := h.paymentExecService.GetPaymentExecutionHistory(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get payment execution history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
		"limit":   limit,
		"offset":  offset,
	})
}

// RetryFailedPayment retries a failed payment execution
func (h *PaymentExecutionHandler) RetryFailedPayment(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := h.paymentExecService.RetryFailedPayment(ctx, executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retry payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelPaymentExecution cancels a payment execution
func (h *PaymentExecutionHandler) CancelPaymentExecution(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.paymentExecService.CancelPaymentExecution(ctx, executionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cancel payment execution",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment execution canceled successfully",
	})
}

// UpdatePaymentExecutionStatus updates the status of a payment execution
func (h *PaymentExecutionHandler) UpdatePaymentExecutionStatus(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	var req struct {
		Status  string `json:"status" validate:"required"`
		Details string `json:"details,omitempty"`
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

	status := services.PaymentExecutionStatusType(req.Status)
	if err := h.paymentExecService.UpdatePaymentExecutionStatus(ctx, executionID, status, req.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update payment execution status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment execution status updated successfully",
	})
}

// GeneratePaymentFulfillment generates the condition and fulfillment for payment release
func (h *PaymentExecutionHandler) GeneratePaymentFulfillment(c *gin.Context) {
	var req struct {
		SmartChequeID string `json:"smart_check_id" validate:"required"`
		MilestoneID   string `json:"milestone_id" validate:"required"`
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

	fulfillment, err := h.paymentExecService.GeneratePaymentFulfillment(ctx, req.SmartChequeID, req.MilestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate payment fulfillment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, fulfillment)
}

// ValidatePaymentCondition validates that a payment condition can be fulfilled
func (h *PaymentExecutionHandler) ValidatePaymentCondition(c *gin.Context) {
	var req struct {
		SmartChequeID string `json:"smart_check_id" validate:"required"`
		MilestoneID   string `json:"milestone_id" validate:"required"`
		Condition     string `json:"condition" validate:"required"`
		Fulfillment   string `json:"fulfillment" validate:"required"`
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

	if err := h.paymentExecService.ValidatePaymentCondition(ctx, req.SmartChequeID, req.MilestoneID, req.Condition, req.Fulfillment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Payment condition validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment condition validated successfully",
		"valid":   true,
	})
}

// getPaginationParams extracts pagination parameters from the request
func (h *PaymentExecutionHandler) getPaginationParams(c *gin.Context) (int, int) {
	limit := 10 // default limit
	offset := 0 // default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			if parsed > 100 { // max limit
				parsed = 100
			}
			limit = parsed
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
