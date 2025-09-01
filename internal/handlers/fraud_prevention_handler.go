package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// FraudPreventionHandler handles HTTP requests for fraud prevention operations
type FraudPreventionHandler struct {
	fraudDetectionService services.FraudDetectionServiceInterface
	fraudAlertingService  services.FraudAlertingServiceInterface
}

// NewFraudPreventionHandler creates a new fraud prevention handler
func NewFraudPreventionHandler(
	fraudDetectionService services.FraudDetectionServiceInterface,
	fraudAlertingService services.FraudAlertingServiceInterface,
) *FraudPreventionHandler {
	return &FraudPreventionHandler{
		fraudDetectionService: fraudDetectionService,
		fraudAlertingService:  fraudAlertingService,
	}
}

// RegisterRoutes registers all fraud prevention routes
func (h *FraudPreventionHandler) RegisterRoutes(router *gin.RouterGroup) {
	fraud := router.Group("/fraud-prevention")
	{
		// Transaction analysis
		fraud.POST("/analyze", h.AnalyzeTransaction)
		fraud.GET("/enterprises/:enterpriseID/patterns", h.DetectFraudPatterns)

		// Fraud rules
		fraud.GET("/rules", h.GetActiveRules)
		fraud.POST("/rules", h.CreateRule)
		fraud.PUT("/rules/:ruleID", h.UpdateRule)
		fraud.DELETE("/rules/:ruleID", h.DeleteRule)

		// Fraud alerts
		fraud.GET("/alerts", h.GetAlerts)
		fraud.POST("/alerts/:alertID/acknowledge", h.AcknowledgeAlert)
		fraud.POST("/alerts/:alertID/resolve", h.ResolveAlert)
		fraud.POST("/alerts/:alertID/assign", h.AssignAlert)

		// Fraud cases
		fraud.POST("/cases", h.CreateCase)
		fraud.GET("/cases/:caseID", h.GetCase)
		fraud.PUT("/cases/:caseID", h.UpdateCase)
		fraud.POST("/cases/:caseID/close", h.CloseCase)

		// Account fraud status
		fraud.GET("/enterprises/:enterpriseID/status", h.GetAccountFraudStatus)
		fraud.PUT("/enterprises/:enterpriseID/status", h.UpdateAccountFraudStatus)
		fraud.POST("/enterprises/:enterpriseID/restrictions", h.AddAccountRestriction)
		fraud.DELETE("/enterprises/:enterpriseID/restrictions/:restrictionType", h.RemoveAccountRestriction)

		// Reporting and analytics
		fraud.POST("/reports", h.GenerateFraudReport)
		fraud.GET("/metrics", h.GetFraudMetrics)
	}
}

// AnalyzeTransaction analyzes a transaction for fraud
func (h *FraudPreventionHandler) AnalyzeTransaction(c *gin.Context) {
	var req services.FraudAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}
	// Basic validation for required fields
	if req.TransactionID == "" || req.EnterpriseID == uuid.Nil || req.Amount == "" || req.CurrencyCode == "" || req.TransactionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: transaction_id, enterprise_id, amount, currency_code, transaction_type",
		})
		return
	}

	// Set default timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	result, err := h.fraudDetectionService.AnalyzeTransaction(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// DetectFraudPatterns detects fraud patterns for an enterprise
func (h *FraudPreventionHandler) DetectFraudPatterns(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	patterns, err := h.fraudDetectionService.DetectFraudPatterns(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to detect fraud patterns",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patterns": patterns,
	})
}

// GetActiveRules gets all active fraud rules
func (h *FraudPreventionHandler) GetActiveRules(c *gin.Context) {
	rules, err := h.fraudDetectionService.GetActiveRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get active rules",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
	})
}

// CreateRule creates a new fraud rule
func (h *FraudPreventionHandler) CreateRule(c *gin.Context) {
	var rule models.FraudRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	err := h.fraudDetectionService.CreateRule(c.Request.Context(), &rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create rule",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"rule": rule,
	})
}

// UpdateRule updates an existing fraud rule
func (h *FraudPreventionHandler) UpdateRule(c *gin.Context) {
	ruleIDStr := c.Param("ruleID")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid rule ID",
		})
		return
	}

	var rule models.FraudRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	rule.ID = ruleID
	err = h.fraudDetectionService.UpdateRule(c.Request.Context(), &rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update rule",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rule": rule,
	})
}

// DeleteRule deletes a fraud rule
func (h *FraudPreventionHandler) DeleteRule(c *gin.Context) {
	ruleIDStr := c.Param("ruleID")
	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid rule ID",
		})
		return
	}

	err = h.fraudDetectionService.DeleteRule(c.Request.Context(), ruleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete rule",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rule deleted successfully",
	})
}

// GetAlerts gets fraud alerts with optional filtering
func (h *FraudPreventionHandler) GetAlerts(c *gin.Context) {
	var filter services.FraudAlertFilter

	// Parse query parameters
	if enterpriseIDStr := c.Query("enterprise_id"); enterpriseIDStr != "" {
		if enterpriseID, err := uuid.Parse(enterpriseIDStr); err == nil {
			filter.EnterpriseID = &enterpriseID
		}
	}

	if status := c.Query("status"); status != "" {
		fraudAlertStatus := models.FraudAlertStatus(status)
		filter.Status = &fraudAlertStatus
	}

	if severity := c.Query("severity"); severity != "" {
		fraudSeverity := models.FraudSeverity(severity)
		filter.Severity = &fraudSeverity
	}

	if alertType := c.Query("alert_type"); alertType != "" {
		fraudAlertType := models.FraudAlertType(alertType)
		filter.AlertType = &fraudAlertType
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	// Set defaults
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	alerts, err := h.fraudDetectionService.GetAlerts(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

// AcknowledgeAlert acknowledges a fraud alert
func (h *FraudPreventionHandler) AcknowledgeAlert(c *gin.Context) {
	alertIDStr := c.Param("alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	err = h.fraudAlertingService.AcknowledgeAlert(c.Request.Context(), alertID, userID, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to acknowledge alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged successfully",
	})
}

// ResolveAlert resolves a fraud alert
func (h *FraudPreventionHandler) ResolveAlert(c *gin.Context) {
	alertIDStr := c.Param("alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	var req struct {
		UserID     string `json:"user_id" binding:"required"`
		Resolution string `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	err = h.fraudAlertingService.ResolveAlert(c.Request.Context(), alertID, req.Resolution, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// AssignAlert assigns a fraud alert to a user
func (h *FraudPreventionHandler) AssignAlert(c *gin.Context) {
	alertIDStr := c.Param("alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	var req struct {
		AssignedTo string `json:"assigned_to" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	assignedTo, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid assigned_to user ID",
		})
		return
	}

	err = h.fraudAlertingService.AssignAlert(c.Request.Context(), alertID, assignedTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert assigned successfully",
	})
}

// CreateCase creates a new fraud case
func (h *FraudPreventionHandler) CreateCase(c *gin.Context) {
	var req services.FraudCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	fraudCase, err := h.fraudDetectionService.CreateCase(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create fraud case",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"case": fraudCase,
	})
}

// GetCase gets a fraud case by ID
func (h *FraudPreventionHandler) GetCase(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid case ID",
		})
		return
	}

	fraudCase, err := h.fraudDetectionService.GetCase(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get fraud case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"case": fraudCase,
	})
}

// UpdateCase updates a fraud case
func (h *FraudPreventionHandler) UpdateCase(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid case ID",
		})
		return
	}

	var updates services.FraudCaseUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	err = h.fraudDetectionService.UpdateCase(c.Request.Context(), caseID, &updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update fraud case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case updated successfully",
	})
}

// CloseCase closes a fraud case
func (h *FraudPreventionHandler) CloseCase(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid case ID",
		})
		return
	}

	var req struct {
		UserID     string                     `json:"user_id" binding:"required"`
		Resolution models.FraudCaseResolution `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	err = h.fraudDetectionService.CloseCase(c.Request.Context(), caseID, &req.Resolution, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to close fraud case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case closed successfully",
	})
}

// GetAccountFraudStatus gets the fraud status for an enterprise
func (h *FraudPreventionHandler) GetAccountFraudStatus(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	status, err := h.fraudDetectionService.GetAccountFraudStatus(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get account fraud status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// UpdateAccountFraudStatus updates the fraud status for an enterprise
func (h *FraudPreventionHandler) UpdateAccountFraudStatus(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	status := models.AccountFraudStatusType(req.Status)
	err = h.fraudDetectionService.UpdateAccountFraudStatus(c.Request.Context(), enterpriseID, status, req.Reason, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update account fraud status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account fraud status updated successfully",
	})
}

// AddAccountRestriction adds a restriction to an enterprise account
func (h *FraudPreventionHandler) AddAccountRestriction(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	var restriction models.AccountRestriction
	if err := c.ShouldBindJSON(&restriction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	err = h.fraudDetectionService.AddAccountRestriction(c.Request.Context(), enterpriseID, &restriction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add account restriction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account restriction added successfully",
	})
}

// RemoveAccountRestriction removes a restriction from an enterprise account
func (h *FraudPreventionHandler) RemoveAccountRestriction(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	restrictionType := models.RestrictionType(c.Param("restrictionType"))

	err = h.fraudDetectionService.RemoveAccountRestriction(c.Request.Context(), enterpriseID, restrictionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove account restriction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account restriction removed successfully",
	})
}

// GenerateFraudReport generates a fraud report
func (h *FraudPreventionHandler) GenerateFraudReport(c *gin.Context) {
	var req services.FraudReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	report, err := h.fraudDetectionService.GenerateFraudReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate fraud report",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// GetFraudMetrics gets fraud metrics
func (h *FraudPreventionHandler) GetFraudMetrics(c *gin.Context) {
	var enterpriseID *uuid.UUID
	if enterpriseIDStr := c.Query("enterprise_id"); enterpriseIDStr != "" {
		if id, err := uuid.Parse(enterpriseIDStr); err == nil {
			enterpriseID = &id
		}
	}

	metrics, err := h.fraudDetectionService.GetFraudMetrics(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get fraud metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}
