package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// PaymentConfirmationHandler handles HTTP requests for payment confirmation operations
type PaymentConfirmationHandler struct {
	paymentConfirmationService services.PaymentConfirmationServiceInterface
}

// NewPaymentConfirmationHandler creates a new handler for payment confirmation
func NewPaymentConfirmationHandler(paymentConfirmationService services.PaymentConfirmationServiceInterface) *PaymentConfirmationHandler {
	return &PaymentConfirmationHandler{
		paymentConfirmationService: paymentConfirmationService,
	}
}

// RegisterRoutes registers the routes for payment confirmation endpoints
func (h *PaymentConfirmationHandler) RegisterRoutes(router *gin.RouterGroup) {
	confirmations := router.Group("/payment-confirmations")
	{
		// Transaction confirmation operations
		confirmations.POST("/start/:transactionId", h.StartTransactionConfirmation)
		confirmations.POST("/monitor", h.MonitorTransactionConfirmations)
		confirmations.GET("/status/:transactionId", h.GetTransactionConfirmationStatus)

		// Payment status updates
		confirmations.POST("/update-status/:transactionId", h.UpdatePaymentStatusFromConfirmation)
		confirmations.POST("/confirm/:executionId", h.ConfirmPaymentExecution)
		confirmations.POST("/fail/:executionId", h.FailPaymentExecution)

		// Confirmation history
		confirmations.GET("/history/:executionId", h.GetPaymentConfirmationHistory)

		// Blockchain interaction
		confirmations.GET("/transaction-status/:transactionId", h.GetTransactionStatus)
		confirmations.POST("/wait-confirmations/:transactionId", h.WaitForConfirmations)

		// Configuration
		confirmations.PUT("/requirements/:transactionId", h.UpdateConfirmationRequirements)
	}
}

// StartTransactionConfirmation starts monitoring confirmations for a transaction
func (h *PaymentConfirmationHandler) StartTransactionConfirmation(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	var req struct {
		PaymentExecutionID string `json:"payment_execution_id" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	paymentExecutionID, err := uuid.Parse(req.PaymentExecutionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment execution ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.paymentConfirmationService.StartTransactionConfirmation(ctx, transactionID, paymentExecutionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start transaction confirmation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Transaction confirmation monitoring started",
		"transaction_id":       transactionID,
		"payment_execution_id": paymentExecutionID,
	})
}

// MonitorTransactionConfirmations starts monitoring all active transaction confirmations
func (h *PaymentConfirmationHandler) MonitorTransactionConfirmations(c *gin.Context) {
	ctx, cancel := contextWithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	if err := h.paymentConfirmationService.MonitorTransactionConfirmations(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to monitor transaction confirmations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction confirmation monitoring completed",
	})
}

// GetTransactionConfirmationStatus gets the current confirmation status for a transaction
func (h *PaymentConfirmationHandler) GetTransactionConfirmationStatus(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentConfirmationService.GetTransactionConfirmationStatus(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Transaction confirmation status not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// UpdatePaymentStatusFromConfirmation updates payment status based on confirmation
func (h *PaymentConfirmationHandler) UpdatePaymentStatusFromConfirmation(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.paymentConfirmationService.UpdatePaymentStatusFromConfirmation(ctx, transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update payment status from confirmation",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Payment status updated from confirmation",
		"transaction_id": transactionID,
	})
}

// ConfirmPaymentExecution manually confirms a payment execution
func (h *PaymentConfirmationHandler) ConfirmPaymentExecution(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason,omitempty"`
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

	if err := h.paymentConfirmationService.ConfirmPaymentExecution(ctx, executionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to confirm payment execution",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Payment execution confirmed",
		"execution_id": executionID,
	})
}

// FailPaymentExecution marks a payment execution as failed
func (h *PaymentConfirmationHandler) FailPaymentExecution(c *gin.Context) {
	executionIDStr := c.Param("executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid execution ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" validate:"required"`
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

	if err := h.paymentConfirmationService.FailPaymentExecution(ctx, executionID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fail payment execution",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Payment execution marked as failed",
		"execution_id": executionID,
		"reason":       req.Reason,
	})
}

// GetPaymentConfirmationHistory gets the confirmation history for a payment execution
func (h *PaymentConfirmationHandler) GetPaymentConfirmationHistory(c *gin.Context) {
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

	history, err := h.paymentConfirmationService.GetPaymentConfirmationHistory(ctx, executionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get payment confirmation history",
			"details": err.Error(),
		})
		return
	}

	// Apply pagination to the results
	limit, offset := h.getPaginationParams(c)
	totalCount := len(history)

	// Apply pagination limits
	if offset >= totalCount {
		history = []*services.PaymentConfirmationRecord{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		history = history[offset:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"history":      history,
		"count":        len(history),
		"total_count":  totalCount,
		"offset":       offset,
		"limit":        limit,
		"execution_id": executionID,
	})
}

// GetTransactionStatus gets the current status of a transaction from blockchain
func (h *PaymentConfirmationHandler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentConfirmationService.GetTransactionStatus(ctx, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get transaction status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// WaitForConfirmations waits for a transaction to reach required confirmations
func (h *PaymentConfirmationHandler) WaitForConfirmations(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	var req struct {
		RequiredConfirmations int `json:"required_confirmations" validate:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	if err := h.paymentConfirmationService.WaitForConfirmations(ctx, transactionID, req.RequiredConfirmations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to wait for confirmations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Required confirmations reached",
		"transaction_id":         transactionID,
		"required_confirmations": req.RequiredConfirmations,
	})
}

// UpdateConfirmationRequirements updates the confirmation requirements for a transaction
func (h *PaymentConfirmationHandler) UpdateConfirmationRequirements(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	var req struct {
		RequiredConfirmations int `json:"required_confirmations" validate:"required,min=1"`
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

	if err := h.paymentConfirmationService.UpdateConfirmationRequirements(ctx, transactionID, req.RequiredConfirmations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update confirmation requirements",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Confirmation requirements updated",
		"transaction_id":         transactionID,
		"required_confirmations": req.RequiredConfirmations,
	})
}

// getPaginationParams extracts pagination parameters from the request
// Used for implementing pagination in list endpoints like GetPaymentConfirmationHistory
func (h *PaymentConfirmationHandler) getPaginationParams(c *gin.Context) (int, int) {
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
