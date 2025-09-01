package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// ComplianceHandler handles compliance-related HTTP requests
type ComplianceHandler struct {
	monitoringService *services.TransactionMonitoringService
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(monitoringService *services.TransactionMonitoringService) *ComplianceHandler {
	return &ComplianceHandler{
		monitoringService: monitoringService,
	}
}

// GetComplianceStatus retrieves compliance status for a transaction
func (h *ComplianceHandler) GetComplianceStatus(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	complianceStatus, err := h.monitoringService.GetComplianceStatus(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve compliance status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compliance_status": complianceStatus,
	})
}

// PerformComplianceCheck performs a compliance check on a transaction
func (h *ComplianceHandler) PerformComplianceCheck(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
		return
	}

	// In a real implementation, you'd fetch the transaction from the database
	// For now, create a mock transaction for demonstration
	transaction := &models.Transaction{
		ID:           transactionID,
		Amount:       "5000.00",
		FromAddress:  "rFromAddress123",
		ToAddress:    "rToAddress456",
		EnterpriseID: "enterprise-123",
	}

	complianceStatus, err := h.monitoringService.PerformAndStoreComplianceCheck(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to perform compliance check",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compliance_status": complianceStatus,
	})
}

// GetFlaggedTransactions retrieves transactions that need compliance review
func (h *ComplianceHandler) GetFlaggedTransactions(c *gin.Context) {
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

	flaggedTransactions, err := h.monitoringService.GetFlaggedTransactions(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve flagged transactions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"flagged_transactions": flaggedTransactions,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(flaggedTransactions),
		},
	})
}

// ReviewComplianceStatus marks a compliance status as reviewed
func (h *ComplianceHandler) ReviewComplianceStatus(c *gin.Context) {
	complianceStatusID := c.Param("complianceStatusId")
	if complianceStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Compliance status ID is required",
		})
		return
	}

	var request struct {
		ReviewedBy string `json:"reviewed_by" binding:"required"`
		Comments   string `json:"comments"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := h.monitoringService.ReviewComplianceStatus(complianceStatusID, request.ReviewedBy, request.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to review compliance status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Compliance status reviewed successfully",
	})
}

// GetComplianceStats retrieves compliance statistics
func (h *ComplianceHandler) GetComplianceStats(c *gin.Context) {
	enterpriseIDStr := c.Query("enterprise_id")
	var enterpriseID *string
	if enterpriseIDStr != "" {
		enterpriseID = &enterpriseIDStr
	}

	// Parse optional since parameter
	var since *time.Time
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &parsedTime
		}
	}

	stats, err := h.monitoringService.GetComplianceStats(enterpriseID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve compliance statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compliance_stats": stats,
	})
}

// UpdateComplianceStatus updates a compliance status
func (h *ComplianceHandler) UpdateComplianceStatus(c *gin.Context) {
	complianceStatusID := c.Param("complianceStatusId")
	if complianceStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Compliance status ID is required",
		})
		return
	}

	var request struct {
		Status       string   `json:"status" binding:"required"`
		ChecksPassed []string `json:"checks_passed"`
		ChecksFailed []string `json:"checks_failed"`
		Violations   []string `json:"violations"`
		ReviewedBy   *string  `json:"reviewed_by,omitempty"`
		Comments     string   `json:"comments"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	complianceStatus := &models.TransactionComplianceStatus{
		ID:           complianceStatusID,
		Status:       request.Status,
		ChecksPassed: request.ChecksPassed,
		ChecksFailed: request.ChecksFailed,
		Violations:   request.Violations,
		ReviewedBy:   request.ReviewedBy,
		Comments:     request.Comments,
		UpdatedAt:    time.Now(),
	}

	if request.ReviewedBy != nil {
		complianceStatus.ReviewedAt = &time.Time{}
		*complianceStatus.ReviewedAt = time.Now()
	}

	err := h.monitoringService.UpdateComplianceStatus(complianceStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update compliance status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Compliance status updated successfully",
		"compliance_status": complianceStatus,
	})
}
