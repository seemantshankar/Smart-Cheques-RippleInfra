package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// TransactionMonitoringHandler handles transaction monitoring HTTP requests
type TransactionMonitoringHandler struct {
	monitoringService *services.TransactionMonitoringService
}

// NewTransactionMonitoringHandler creates a new transaction monitoring handler
func NewTransactionMonitoringHandler(monitoringService *services.TransactionMonitoringService) *TransactionMonitoringHandler {
	return &TransactionMonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetTransactionStats retrieves transaction statistics
func (h *TransactionMonitoringHandler) GetTransactionStats(c *gin.Context) {
	enterpriseID := c.Param("enterpriseId")
	if enterpriseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Enterprise ID is required",
		})
		return
	}

	// Parse optional since parameter
	var since *time.Time
	if sinceStr := c.Query("since"); sinceStr != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = &parsedTime
		}
	}

	stats, err := h.monitoringService.GetTransactionStats(enterpriseID, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve transaction statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GenerateTransactionReport generates a transaction monitoring report
func (h *TransactionMonitoringHandler) GenerateTransactionReport(c *gin.Context) {
	enterpriseID := c.Param("enterpriseId")
	if enterpriseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Enterprise ID is required",
		})
		return
	}

	// Parse report parameters
	reportType := c.DefaultQuery("type", "daily")

	// Parse date range
	periodStartStr := c.Query("start_date")
	periodEndStr := c.Query("end_date")

	if periodStartStr == "" || periodEndStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Both start_date and end_date are required",
		})
		return
	}

	periodStart, err := time.Parse("2006-01-02", periodStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	periodEnd, err := time.Parse("2006-01-02", periodEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Get user ID from context (assuming middleware sets this)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // Fallback
	}

	report, err := h.monitoringService.GenerateTransactionReport(
		enterpriseID,
		reportType,
		periodStart,
		periodEnd,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate transaction report",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// AssessTransactionRisk performs risk assessment on a transaction
func (h *TransactionMonitoringHandler) AssessTransactionRisk(c *gin.Context) {
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
		ID:         transactionID,
		Type:       models.TransactionTypePayment,
		Status:     models.TransactionStatusPending,
		Amount:     "5000.00", // Mock amount
		Currency:   "XRP",
		RetryCount: 0,
	}

	riskScore, err := h.monitoringService.AssessTransactionRisk(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assess transaction risk",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id":  transactionID,
		"risk_assessment": riskScore,
	})
}

// CheckTransactionCompliance performs compliance check on a transaction
func (h *TransactionMonitoringHandler) CheckTransactionCompliance(c *gin.Context) {
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
		FromAddress:  "rExampleFromAddress",
		ToAddress:    "rExampleToAddress",
		EnterpriseID: "enterprise-123",
	}

	complianceStatus, err := h.monitoringService.PerformComplianceCheck(transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check transaction compliance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_id":    transactionID,
		"compliance_status": complianceStatus,
	})
}

// GetAuditLogs retrieves transaction audit logs
func (h *TransactionMonitoringHandler) GetAuditLogs(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Transaction ID is required",
		})
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

	// In a real implementation, you'd query audit logs filtered by transaction ID
	// For now, return a mock response
	c.JSON(http.StatusOK, gin.H{
		"transaction_id": transactionID,
		"audit_logs": []map[string]interface{}{
			{
				"id":         "audit-1",
				"event_type": "created",
				"details":    "Transaction created",
				"created_at": time.Now().Format(time.RFC3339),
			},
		},
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  1,
		},
	})
}
