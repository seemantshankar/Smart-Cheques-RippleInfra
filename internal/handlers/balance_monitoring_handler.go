package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/services"
)

// BalanceMonitoringHandler handles HTTP requests for balance monitoring operations
type BalanceMonitoringHandler struct {
	monitoringService services.BalanceMonitoringServiceInterface
}

// NewBalanceMonitoringHandler creates a new balance monitoring handler
func NewBalanceMonitoringHandler(monitoringService services.BalanceMonitoringServiceInterface) *BalanceMonitoringHandler {
	return &BalanceMonitoringHandler{
		monitoringService: monitoringService,
	}
}

// RegisterRoutes registers all balance monitoring routes
func (h *BalanceMonitoringHandler) RegisterRoutes(router *gin.RouterGroup) {
	monitoring := router.Group("/balance-monitoring")
	{
		// Monitoring control
		monitoring.POST("/start", h.StartMonitoring)
		monitoring.POST("/stop", h.StopMonitoring)
		monitoring.GET("/status", h.GetMonitoringStatus)

		// Threshold management
		monitoring.POST("/thresholds", h.SetBalanceThreshold)
		monitoring.GET("/enterprises/:enterpriseID/thresholds", h.GetBalanceThresholds)
		monitoring.PUT("/thresholds/:thresholdID", h.UpdateBalanceThreshold)
		monitoring.DELETE("/thresholds/:thresholdID", h.DeleteBalanceThreshold)

		// Alert management
		monitoring.GET("/alerts", h.GetActiveAlerts)
		monitoring.GET("/enterprises/:enterpriseID/alerts", h.GetEnterpriseAlerts)
		monitoring.POST("/alerts/:alertID/acknowledge", h.AcknowledgeAlert)
		monitoring.GET("/alerts/history", h.GetAlertHistory)

		// Analytics and trends
		monitoring.POST("/trends", h.GetBalanceTrends)
		monitoring.GET("/enterprises/:enterpriseID/anomalies", h.DetectBalanceAnomalies)
	}
}

// StartMonitoring starts the balance monitoring service
func (h *BalanceMonitoringHandler) StartMonitoring(c *gin.Context) {
	var config services.MonitoringConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.monitoringService.StartMonitoring(c.Request.Context(), &config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Balance monitoring started successfully",
		"config":  config,
	})
}

// StopMonitoring stops the balance monitoring service
func (h *BalanceMonitoringHandler) StopMonitoring(c *gin.Context) {
	err := h.monitoringService.StopMonitoring(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Balance monitoring stopped successfully",
	})
}

// GetMonitoringStatus gets the current monitoring status
func (h *BalanceMonitoringHandler) GetMonitoringStatus(c *gin.Context) {
	status, err := h.monitoringService.GetMonitoringStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SetBalanceThreshold creates a new balance threshold
func (h *BalanceMonitoringHandler) SetBalanceThreshold(c *gin.Context) {
	var req services.BalanceThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	threshold, err := h.monitoringService.SetBalanceThreshold(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Balance threshold created successfully",
		"threshold": threshold,
	})
}

// GetBalanceThresholds gets balance thresholds for an enterprise
func (h *BalanceMonitoringHandler) GetBalanceThresholds(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	thresholds, err := h.monitoringService.GetBalanceThresholds(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"thresholds":    thresholds,
		"count":         len(thresholds),
	})
}

// UpdateBalanceThreshold updates an existing balance threshold
func (h *BalanceMonitoringHandler) UpdateBalanceThreshold(c *gin.Context) {
	thresholdIDStr := c.Param("thresholdID")
	thresholdID, err := uuid.Parse(thresholdIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid threshold ID"})
		return
	}

	var req services.UpdateThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set threshold ID from URL parameter
	req.ThresholdID = thresholdID

	threshold, err := h.monitoringService.UpdateBalanceThreshold(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Balance threshold updated successfully",
		"threshold": threshold,
	})
}

// DeleteBalanceThreshold deletes a balance threshold
func (h *BalanceMonitoringHandler) DeleteBalanceThreshold(c *gin.Context) {
	thresholdIDStr := c.Param("thresholdID")
	thresholdID, err := uuid.Parse(thresholdIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid threshold ID"})
		return
	}

	err = h.monitoringService.DeleteBalanceThreshold(c.Request.Context(), thresholdID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Balance threshold deleted successfully",
	})
}

// GetActiveAlerts gets active balance alerts
func (h *BalanceMonitoringHandler) GetActiveAlerts(c *gin.Context) {
	// Parse optional enterprise ID filter
	var enterpriseID *uuid.UUID
	enterpriseIDStr := c.Query("enterprise_id")
	if enterpriseIDStr != "" {
		id, err := uuid.Parse(enterpriseIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
			return
		}
		enterpriseID = &id
	}

	alerts, err := h.monitoringService.GetActiveAlerts(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_alerts": alerts,
		"count":         len(alerts),
	})
}

// GetEnterpriseAlerts gets active alerts for a specific enterprise
func (h *BalanceMonitoringHandler) GetEnterpriseAlerts(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	alerts, err := h.monitoringService.GetActiveAlerts(c.Request.Context(), &enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"alerts":        alerts,
		"count":         len(alerts),
	})
}

// AcknowledgeAlert acknowledges a balance alert
func (h *BalanceMonitoringHandler) AcknowledgeAlert(c *gin.Context) {
	alertIDStr := c.Param("alertID")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var req struct {
		AcknowledgedBy uuid.UUID `json:"acknowledged_by" binding:"required"`
		Comments       string    `json:"comments,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.monitoringService.AcknowledgeAlert(c.Request.Context(), alertID, req.AcknowledgedBy)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged successfully",
	})
}

// GetAlertHistory gets alert history
func (h *BalanceMonitoringHandler) GetAlertHistory(c *gin.Context) {
	var req services.AlertHistoryRequest

	// Parse optional enterprise ID filter
	enterpriseIDStr := c.Query("enterprise_id")
	if enterpriseIDStr != "" {
		id, err := uuid.Parse(enterpriseIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
			return
		}
		req.EnterpriseID = &id
	}

	// Parse other query parameters
	req.CurrencyCode = c.Query("currency_code")

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

	req.Limit = limit
	req.Offset = offset

	history, err := h.monitoringService.GetAlertHistory(c.Request.Context(), &req)
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

// GetBalanceTrends gets balance trend analysis
func (h *BalanceMonitoringHandler) GetBalanceTrends(c *gin.Context) {
	var req services.BalanceTrendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trends, err := h.monitoringService.GetBalanceTrends(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// DetectBalanceAnomalies detects balance anomalies for an enterprise
func (h *BalanceMonitoringHandler) DetectBalanceAnomalies(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	anomalies, err := h.monitoringService.DetectBalanceAnomalies(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"anomalies":     anomalies,
		"count":         len(anomalies),
	})
}
