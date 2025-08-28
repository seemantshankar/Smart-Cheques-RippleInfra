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

// WithdrawalAuthorizationHandler handles HTTP requests for withdrawal authorization operations
type WithdrawalAuthorizationHandler struct {
	authorizationService services.WithdrawalAuthorizationServiceInterface
}

// NewWithdrawalAuthorizationHandler creates a new withdrawal authorization handler
func NewWithdrawalAuthorizationHandler(authorizationService services.WithdrawalAuthorizationServiceInterface) *WithdrawalAuthorizationHandler {
	return &WithdrawalAuthorizationHandler{
		authorizationService: authorizationService,
	}
}

// RegisterRoutes registers all withdrawal authorization routes
func (h *WithdrawalAuthorizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	authorization := router.Group("/withdrawal-authorization")
	{
		// Withdrawal request management
		authorization.POST("/enterprises/:enterpriseID/requests", h.CreateWithdrawalRequest)
		authorization.GET("/requests/:requestID", h.GetWithdrawalRequest)
		authorization.GET("/enterprises/:enterpriseID/pending", h.GetPendingWithdrawalRequests)
		authorization.GET("/enterprises/:enterpriseID/history", h.GetAuthorizationHistory)

		// Authorization workflow
		authorization.POST("/requests/:requestID/approve", h.ApproveWithdrawal)
		authorization.POST("/requests/:requestID/reject", h.RejectWithdrawal)
		authorization.GET("/requests/:requestID/status", h.CheckAuthorizationStatus)

		// Time-locked withdrawals
		authorization.POST("/requests/:requestID/time-lock", h.CreateTimeLockWithdrawal)
		authorization.DELETE("/time-locks/:lockID", h.ReleaseTimeLockWithdrawal)
		authorization.GET("/time-locks/:lockID/status", h.GetTimeLockStatus)

		// Risk assessment
		authorization.POST("/enterprises/:enterpriseID/risk-assessment", h.AssessWithdrawalRisk)
		authorization.GET("/enterprises/:enterpriseID/risk-profile", h.GetRiskProfile)

		// Bulk operations
		authorization.POST("/bulk-approve", h.BulkApproveWithdrawals)
	}
}

// parseWithdrawalUUIDParam parses a UUID parameter from the request context
func parseWithdrawalUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	paramStr := c.Param(paramName)
	id, err := uuid.Parse(paramStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s", paramName)})
		return uuid.Nil, false
	}
	return id, true
}

// CreateWithdrawalRequest handles withdrawal authorization request creation
func (h *WithdrawalAuthorizationHandler) CreateWithdrawalRequest(c *gin.Context) {
	enterpriseID, valid := parseWithdrawalUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.WithdrawalAuthorizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	authorization, err := h.authorizationService.CreateWithdrawalRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Withdrawal authorization request created successfully",
		"authorization": authorization,
	})
}

// GetWithdrawalRequest retrieves a specific withdrawal request
func (h *WithdrawalAuthorizationHandler) GetWithdrawalRequest(c *gin.Context) {
	requestID, valid := parseWithdrawalUUIDParam(c, "requestID")
	if !valid {
		return
	}

	authorization, err := h.authorizationService.GetWithdrawalRequest(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, authorization)
}

// GetPendingWithdrawalRequests retrieves pending withdrawal requests for an enterprise
func (h *WithdrawalAuthorizationHandler) GetPendingWithdrawalRequests(c *gin.Context) {
	enterpriseID, valid := parseWithdrawalUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	requests, err := h.authorizationService.GetPendingWithdrawalRequests(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_requests": requests,
		"count":            len(requests),
	})
}

// ApproveWithdrawal handles withdrawal approval requests
func (h *WithdrawalAuthorizationHandler) ApproveWithdrawal(c *gin.Context) {
	requestID, valid := parseWithdrawalUUIDParam(c, "requestID")
	if !valid {
		return
	}

	var req services.WithdrawalApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set request ID from URL parameter
	req.RequestID = requestID

	approval, err := h.authorizationService.ApproveWithdrawal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Withdrawal approved successfully",
		"approval": approval,
	})
}

// RejectWithdrawal handles withdrawal rejection requests
func (h *WithdrawalAuthorizationHandler) RejectWithdrawal(c *gin.Context) {
	requestID, valid := parseWithdrawalUUIDParam(c, "requestID")
	if !valid {
		return
	}

	var req services.WithdrawalRejectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set request ID from URL parameter
	req.RequestID = requestID

	rejection, err := h.authorizationService.RejectWithdrawal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Withdrawal rejected successfully",
		"rejection": rejection,
	})
}

// CheckAuthorizationStatus checks the authorization status of a withdrawal request
func (h *WithdrawalAuthorizationHandler) CheckAuthorizationStatus(c *gin.Context) {
	requestID, valid := parseWithdrawalUUIDParam(c, "requestID")
	if !valid {
		return
	}

	status, err := h.authorizationService.CheckAuthorizationStatus(c.Request.Context(), requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// CreateTimeLockWithdrawal creates a time-locked withdrawal
func (h *WithdrawalAuthorizationHandler) CreateTimeLockWithdrawal(c *gin.Context) {
	requestID, valid := parseWithdrawalUUIDParam(c, "requestID")
	if !valid {
		return
	}

	var req services.TimeLockWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set request ID from URL parameter
	req.WithdrawalRequestID = requestID

	timeLock, err := h.authorizationService.CreateTimeLockWithdrawal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Time lock created successfully",
		"time_lock": timeLock,
	})
}

// ReleaseTimeLockWithdrawal releases a time lock early
func (h *WithdrawalAuthorizationHandler) ReleaseTimeLockWithdrawal(c *gin.Context) {
	lockID, valid := parseWithdrawalUUIDParam(c, "lockID")
	if !valid {
		return
	}

	err := h.authorizationService.ReleaseTimeLockWithdrawal(c.Request.Context(), lockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Time lock released successfully",
	})
}

// GetTimeLockStatus gets the status of a time lock
func (h *WithdrawalAuthorizationHandler) GetTimeLockStatus(c *gin.Context) {
	lockIDStr := c.Param("lockID")
	lockID, err := uuid.Parse(lockIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lock ID"})
		return
	}

	status, err := h.authorizationService.GetTimeLockStatus(c.Request.Context(), lockID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// AssessWithdrawalRisk performs risk assessment for a withdrawal
func (h *WithdrawalAuthorizationHandler) AssessWithdrawalRisk(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	var req services.WithdrawalRiskAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter and current time if not provided
	req.EnterpriseID = enterpriseID
	if req.TimeOfDay.IsZero() {
		req.TimeOfDay = time.Now()
	}

	riskScore, err := h.authorizationService.AssessWithdrawalRisk(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, riskScore)
}

// GetRiskProfile gets the risk profile for an enterprise
func (h *WithdrawalAuthorizationHandler) GetRiskProfile(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	profile, err := h.authorizationService.GetRiskProfile(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// BulkApproveWithdrawals processes multiple withdrawal approvals at once
func (h *WithdrawalAuthorizationHandler) BulkApproveWithdrawals(c *gin.Context) {
	var req services.BulkApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authorizationService.BulkApproveWithdrawals(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk approval processed successfully",
		"result":  result,
	})
}

// GetAuthorizationHistory returns authorization history for an enterprise
func (h *WithdrawalAuthorizationHandler) GetAuthorizationHistory(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	history, err := h.authorizationService.GetAuthorizationHistory(c.Request.Context(), enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(history),
		},
	})
}
