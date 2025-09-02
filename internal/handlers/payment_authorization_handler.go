package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// PaymentAuthorizationHandler handles HTTP requests for payment authorization operations
type PaymentAuthorizationHandler struct {
	paymentAuthService services.PaymentAuthorizationServiceInterface
}

// NewPaymentAuthorizationHandler creates a new handler for payment authorization
func NewPaymentAuthorizationHandler(paymentAuthService services.PaymentAuthorizationServiceInterface) *PaymentAuthorizationHandler {
	return &PaymentAuthorizationHandler{
		paymentAuthService: paymentAuthService,
	}
}

// RegisterRoutes registers the routes for payment authorization endpoints
func (h *PaymentAuthorizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	payments := router.Group("/payment-authorizations")
	{
		// Payment authorization CRUD operations
		payments.POST("", h.CreatePaymentAuthorization)
		payments.GET("/:requestId", h.GetPaymentAuthorization)
		payments.GET("/enterprise/:enterpriseId/pending", h.GetPendingPaymentAuthorizations)
		payments.GET("/smart-check/:smartChequeId", h.GetPaymentAuthorizationsBySmartCheque)

		// Approval workflow
		payments.POST("/:requestId/approve", h.ApprovePayment)
		payments.POST("/:requestId/reject", h.RejectPayment)
		payments.GET("/:requestId/status", h.CheckPaymentAuthorizationStatus)

		// Auto-approval
		payments.POST("/:requestId/auto-approve", h.ProcessAutoApproval)

		// Time-locked payments
		payments.POST("/time-lock", h.CreateTimeLockPayment)
		payments.POST("/time-lock/:lockId/release", h.ReleaseTimeLockPayment)
		payments.GET("/time-lock/:lockId/status", h.GetTimeLockPaymentStatus)

		// Risk assessment
		payments.POST("/risk-assessment", h.AssessPaymentRisk)
		payments.GET("/enterprise/:enterpriseId/risk-profile", h.GetPaymentRiskProfile)

		// Bulk operations
		payments.POST("/bulk-approve", h.BulkApprovePayments)
		payments.GET("/enterprise/:enterpriseId/history", h.GetPaymentAuthorizationHistory)

		// Milestone-triggered payments
		payments.POST("/milestone/:milestoneId/initiate", h.InitiatePaymentFromMilestone)
	}
}

// CreatePaymentAuthorization creates a new payment authorization request
func (h *PaymentAuthorizationHandler) CreatePaymentAuthorization(c *gin.Context) {
	var req services.PaymentAuthorizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	auth, err := h.paymentAuthService.CreatePaymentAuthorizationRequest(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create payment authorization",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, auth)
}

// GetPaymentAuthorization retrieves a payment authorization request by ID
func (h *PaymentAuthorizationHandler) GetPaymentAuthorization(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	auth, err := h.paymentAuthService.GetPaymentAuthorizationRequest(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Payment authorization not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, auth)
}

// GetPendingPaymentAuthorizations retrieves pending payment authorizations for an enterprise
func (h *PaymentAuthorizationHandler) GetPendingPaymentAuthorizations(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	auths, err := h.paymentAuthService.GetPendingPaymentAuthorizations(ctx, enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get pending payment authorizations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_authorizations": auths,
		"count":                  len(auths),
	})
}

// GetPaymentAuthorizationsBySmartCheque retrieves payment authorizations for a SmartCheque
func (h *PaymentAuthorizationHandler) GetPaymentAuthorizationsBySmartCheque(c *gin.Context) {
	smartChequeID := c.Param("smartChequeId")
	if smartChequeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "SmartCheque ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	auths, err := h.paymentAuthService.GetPaymentAuthorizationsBySmartCheque(ctx, smartChequeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get payment authorizations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_authorizations": auths,
		"count":                  len(auths),
	})
}

// ApprovePayment approves a payment authorization request
func (h *PaymentAuthorizationHandler) ApprovePayment(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	var req services.PaymentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}
	req.RequestID = requestID

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	approval, err := h.paymentAuthService.ApprovePayment(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to approve payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// RejectPayment rejects a payment authorization request
func (h *PaymentAuthorizationHandler) RejectPayment(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	var req services.PaymentRejectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}
	req.RequestID = requestID

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	approval, err := h.paymentAuthService.RejectPayment(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reject payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, approval)
}

// CheckPaymentAuthorizationStatus checks the status of a payment authorization
func (h *PaymentAuthorizationHandler) CheckPaymentAuthorizationStatus(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentAuthService.CheckPaymentAuthorizationStatus(ctx, requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check payment authorization status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ProcessAutoApproval processes auto-approval for a payment
func (h *PaymentAuthorizationHandler) ProcessAutoApproval(c *gin.Context) {
	requestIDStr := c.Param("requestId")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.paymentAuthService.ProcessAutoApproval(ctx, requestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process auto-approval",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Auto-approval processed successfully",
		"request_id": requestID,
	})
}

// CreateTimeLockPayment creates a time-locked payment
func (h *PaymentAuthorizationHandler) CreateTimeLockPayment(c *gin.Context) {
	var req services.TimeLockPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	lock, err := h.paymentAuthService.CreateTimeLockPayment(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create time-locked payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, lock)
}

// ReleaseTimeLockPayment releases a time-locked payment
func (h *PaymentAuthorizationHandler) ReleaseTimeLockPayment(c *gin.Context) {
	lockIDStr := c.Param("lockId")
	lockID, err := uuid.Parse(lockIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid lock ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.paymentAuthService.ReleaseTimeLockPayment(ctx, lockID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to release time-locked payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Time-locked payment released successfully",
		"lock_id": lockID,
	})
}

// GetTimeLockPaymentStatus gets the status of a time-locked payment
func (h *PaymentAuthorizationHandler) GetTimeLockPaymentStatus(c *gin.Context) {
	lockIDStr := c.Param("lockId")
	lockID, err := uuid.Parse(lockIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid lock ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	status, err := h.paymentAuthService.GetTimeLockPaymentStatus(ctx, lockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get time-locked payment status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// AssessPaymentRisk assesses the risk of a payment
func (h *PaymentAuthorizationHandler) AssessPaymentRisk(c *gin.Context) {
	var req services.PaymentRiskAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	riskScore, err := h.paymentAuthService.AssessPaymentRisk(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to assess payment risk",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, riskScore)
}

// GetPaymentRiskProfile gets the risk profile for an enterprise
func (h *PaymentAuthorizationHandler) GetPaymentRiskProfile(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	profile, err := h.paymentAuthService.GetPaymentRiskProfile(ctx, enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get payment risk profile",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// BulkApprovePayments approves multiple payments at once
func (h *PaymentAuthorizationHandler) BulkApprovePayments(c *gin.Context) {
	var req services.BulkPaymentApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON payload",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.paymentAuthService.BulkApprovePayments(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to bulk approve payments",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPaymentAuthorizationHistory gets the authorization history for an enterprise
func (h *PaymentAuthorizationHandler) GetPaymentAuthorizationHistory(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	limit, offset := h.getPaginationParams(c)

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	history, err := h.paymentAuthService.GetPaymentAuthorizationHistory(ctx, enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get payment authorization history",
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

// InitiatePaymentFromMilestone initiates payment authorization from a completed milestone
func (h *PaymentAuthorizationHandler) InitiatePaymentFromMilestone(c *gin.Context) {
	milestoneID := c.Param("milestoneId")
	if milestoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Milestone ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	auth, err := h.paymentAuthService.InitiatePaymentFromMilestone(ctx, milestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to initiate payment from milestone",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, auth)
}

// getPaginationParams extracts pagination parameters from the request
func (h *PaymentAuthorizationHandler) getPaginationParams(c *gin.Context) (int, int) {
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
