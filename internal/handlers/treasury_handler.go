package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// TreasuryHandler handles HTTP requests for treasury operations
type TreasuryHandler struct {
	treasuryService services.TreasuryServiceInterface
}

// NewTreasuryHandler creates a new treasury handler
func NewTreasuryHandler(treasuryService services.TreasuryServiceInterface) *TreasuryHandler {
	return &TreasuryHandler{
		treasuryService: treasuryService,
	}
}

// RegisterRoutes registers all treasury routes
func (h *TreasuryHandler) RegisterRoutes(router *gin.RouterGroup) {
	treasury := router.Group("/treasury")
	{
		// Fund management
		treasury.POST("/enterprises/:enterpriseID/fund", h.FundEnterprise)
		treasury.POST("/enterprises/:enterpriseID/withdraw", h.WithdrawFunds)
		treasury.POST("/transfer", h.TransferFunds)

		// Treasury operations
		treasury.GET("/enterprises/:enterpriseID/balance", h.GetTreasuryBalance)
		treasury.GET("/enterprises/:enterpriseID/history", h.GetFundingHistory)

		// Liquidity management
		treasury.POST("/enterprises/:enterpriseID/rebalance", h.RebalanceLiquidity)
		treasury.GET("/liquidity/alerts", h.CheckLiquidityThresholds)

		// Analytics and reporting
		treasury.GET("/enterprises/:enterpriseID/report/:period", h.GenerateTreasuryReport)
		treasury.GET("/enterprises/:enterpriseID/capacity/:currencyCode", h.CalculateFundingCapacity)
	}
}

// parseTreasuryUUIDParam parses a UUID parameter from the request context
func parseTreasuryUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	paramStr := c.Param(paramName)
	id, err := uuid.Parse(paramStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s", paramName)})
		return uuid.Nil, false
	}
	return id, true
}

// FundEnterprise handles enterprise funding requests
func (h *TreasuryHandler) FundEnterprise(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.FundingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	transaction, err := h.treasuryService.FundEnterprise(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Funding request processed successfully",
		"transaction": transaction,
	})
}

// WithdrawFunds handles withdrawal requests
func (h *TreasuryHandler) WithdrawFunds(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	transaction, err := h.treasuryService.WithdrawFunds(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Withdrawal request processed successfully",
		"transaction": transaction,
	})
}

// TransferFunds handles internal fund transfers
func (h *TreasuryHandler) TransferFunds(c *gin.Context) {
	var req services.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactions, err := h.treasuryService.TransferFunds(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Transfer completed successfully",
		"transactions": transactions,
	})
}

// GetTreasuryBalance returns treasury balance summary
func (h *TreasuryHandler) GetTreasuryBalance(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	balance, err := h.treasuryService.GetTreasuryBalance(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetFundingHistory returns transaction history
func (h *TreasuryHandler) GetFundingHistory(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
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

	transactions, err := h.treasuryService.GetFundingHistory(c.Request.Context(), enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(transactions),
		},
	})
}

// RebalanceLiquidity performs liquidity rebalancing
func (h *TreasuryHandler) RebalanceLiquidity(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	result, err := h.treasuryService.RebalanceLiquidity(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Liquidity rebalancing completed successfully",
		"result":  result,
	})
}

// CheckLiquidityThresholds checks liquidity alerts
func (h *TreasuryHandler) CheckLiquidityThresholds(c *gin.Context) {
	alerts, err := h.treasuryService.CheckLiquidityThresholds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GenerateTreasuryReport generates treasury analytics report
func (h *TreasuryHandler) GenerateTreasuryReport(c *gin.Context) {
	enterpriseID, valid := parseTreasuryUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	periodStr := c.Param("period")
	var period services.ReportPeriod

	switch periodStr {
	case "daily":
		period = services.ReportPeriodDaily
	case "weekly":
		period = services.ReportPeriodWeekly
	case "monthly":
		period = services.ReportPeriodMonthly
	case "quarterly":
		period = services.ReportPeriodQuarterly
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report period. Use: daily, weekly, monthly, quarterly"})
		return
	}

	report, err := h.treasuryService.GenerateTreasuryReport(c.Request.Context(), enterpriseID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// CalculateFundingCapacity calculates funding capacity for currency
func (h *TreasuryHandler) CalculateFundingCapacity(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	currencyCode := c.Param("currencyCode")
	if currencyCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Currency code is required"})
		return
	}

	capacity, err := h.treasuryService.CalculateFundingCapacity(c.Request.Context(), enterpriseID, currencyCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, capacity)
}
