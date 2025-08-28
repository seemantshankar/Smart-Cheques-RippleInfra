package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// ReconciliationHandler handles HTTP requests for reconciliation operations
type ReconciliationHandler struct {
	reconciliationService services.ReconciliationServiceInterface
}

// NewReconciliationHandler creates a new reconciliation handler
func NewReconciliationHandler(reconciliationService services.ReconciliationServiceInterface) *ReconciliationHandler {
	return &ReconciliationHandler{
		reconciliationService: reconciliationService,
	}
}

// RegisterRoutes registers all reconciliation routes
func (h *ReconciliationHandler) RegisterRoutes(router *gin.RouterGroup) {
	reconciliation := router.Group("/reconciliation")
	{
		// Automated reconciliation
		reconciliation.POST("/perform", h.PerformReconciliation)
		reconciliation.POST("/schedule", h.ScheduleReconciliation)
		reconciliation.GET("/status/:reconciliationID", h.GetReconciliationStatus)

		// Discrepancy management
		reconciliation.GET("/discrepancies", h.GetDiscrepancies)
		reconciliation.GET("/enterprises/:enterpriseID/discrepancies", h.GetEnterpriseDiscrepancies)
		reconciliation.POST("/discrepancies/:discrepancyID/resolve", h.ResolveDiscrepancy)
		reconciliation.POST("/discrepancies/bulk-resolve", h.BulkResolveDiscrepancies)

		// Reporting and analytics
		reconciliation.POST("/reports/generate", h.GenerateReconciliationReport)
		reconciliation.GET("/history", h.GetReconciliationHistory)
		reconciliation.GET("/enterprises/:enterpriseID/history", h.GetEnterpriseReconciliationHistory)
		reconciliation.GET("/metrics/:period", h.GetReconciliationMetrics)

		// Manual operations
		reconciliation.POST("/enterprises/:enterpriseID/manual", h.PerformManualReconciliation)
		reconciliation.POST("/overrides", h.CreateReconciliationOverride)
		reconciliation.GET("/overrides/pending", h.GetPendingOverrides)
	}
}

// PerformReconciliation handles reconciliation requests
func (h *ReconciliationHandler) PerformReconciliation(c *gin.Context) {
	var req services.ReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.reconciliationService.PerformReconciliation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reconciliation completed successfully",
		"result":  result,
	})
}

// ScheduleReconciliation handles reconciliation scheduling
func (h *ReconciliationHandler) ScheduleReconciliation(c *gin.Context) {
	var schedule services.ReconciliationSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default values
	schedule.ID = uuid.New()
	schedule.CreatedAt = time.Now()
	schedule.IsActive = true

	err := h.reconciliationService.ScheduleReconciliation(c.Request.Context(), &schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Reconciliation scheduled successfully",
		"schedule": schedule,
	})
}

// GetReconciliationStatus gets the status of a reconciliation
func (h *ReconciliationHandler) GetReconciliationStatus(c *gin.Context) {
	reconciliationIDStr := c.Param("reconciliationID")
	reconciliationID, err := uuid.Parse(reconciliationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reconciliation ID"})
		return
	}

	status, err := h.reconciliationService.GetReconciliationStatus(c.Request.Context(), reconciliationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reconciliation_id": reconciliationID,
		"status":            status,
	})
}

// GetDiscrepancies gets all discrepancies with optional enterprise filter
func (h *ReconciliationHandler) GetDiscrepancies(c *gin.Context) {
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

	discrepancies, err := h.reconciliationService.GetDiscrepancies(c.Request.Context(), enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"discrepancies": discrepancies,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(discrepancies),
		},
	})
}

// GetEnterpriseDiscrepancies gets discrepancies for a specific enterprise
func (h *ReconciliationHandler) GetEnterpriseDiscrepancies(c *gin.Context) {
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

	discrepancies, err := h.reconciliationService.GetDiscrepancies(c.Request.Context(), &enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"discrepancies": discrepancies,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(discrepancies),
		},
	})
}

// ResolveDiscrepancy handles discrepancy resolution
func (h *ReconciliationHandler) ResolveDiscrepancy(c *gin.Context) {
	discrepancyIDStr := c.Param("discrepancyID")
	discrepancyID, err := uuid.Parse(discrepancyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid discrepancy ID"})
		return
	}

	var req services.DiscrepancyResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set discrepancy ID from URL parameter
	req.DiscrepancyID = discrepancyID

	resolution, err := h.reconciliationService.ResolveDiscrepancy(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Discrepancy resolved successfully",
		"resolution": resolution,
	})
}

// BulkResolveDiscrepancies handles bulk discrepancy resolution
func (h *ReconciliationHandler) BulkResolveDiscrepancies(c *gin.Context) {
	var req services.BulkDiscrepancyResolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.reconciliationService.BulkResolveDiscrepancies(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk resolution processed successfully",
		"result":  result,
	})
}

// GenerateReconciliationReport generates a reconciliation report
func (h *ReconciliationHandler) GenerateReconciliationReport(c *gin.Context) {
	var req services.ReconciliationReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reconciliationService.GenerateReconciliationReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetReconciliationHistory gets global reconciliation history
func (h *ReconciliationHandler) GetReconciliationHistory(c *gin.Context) {
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

	history, err := h.reconciliationService.GetReconciliationHistory(c.Request.Context(), enterpriseID, limit, offset)
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

// GetEnterpriseReconciliationHistory gets reconciliation history for a specific enterprise
func (h *ReconciliationHandler) GetEnterpriseReconciliationHistory(c *gin.Context) {
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

	history, err := h.reconciliationService.GetReconciliationHistory(c.Request.Context(), &enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"history":       history,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(history),
		},
	})
}

// GetReconciliationMetrics gets reconciliation metrics for a period
func (h *ReconciliationHandler) GetReconciliationMetrics(c *gin.Context) {
	periodStr := c.Param("period")
	var period services.ReconciliationPeriod

	switch periodStr {
	case "daily":
		period = services.ReconciliationPeriodDaily
	case "weekly":
		period = services.ReconciliationPeriodWeekly
	case "monthly":
		period = services.ReconciliationPeriodMonthly
	case "quarterly":
		period = services.ReconciliationPeriodQuarterly
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use: daily, weekly, monthly, quarterly"})
		return
	}

	metrics, err := h.reconciliationService.GetReconciliationMetrics(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// PerformManualReconciliation handles manual reconciliation requests
func (h *ReconciliationHandler) PerformManualReconciliation(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	var req services.ManualReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	result, err := h.reconciliationService.PerformManualReconciliation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Manual reconciliation completed successfully",
		"result":  result,
	})
}

// CreateReconciliationOverride creates a reconciliation override
func (h *ReconciliationHandler) CreateReconciliationOverride(c *gin.Context) {
	var req services.ReconciliationOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	override, err := h.reconciliationService.CreateReconciliationOverride(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Reconciliation override created successfully",
		"override": override,
	})
}

// GetPendingOverrides gets pending reconciliation overrides
func (h *ReconciliationHandler) GetPendingOverrides(c *gin.Context) {
	overrides, err := h.reconciliationService.GetPendingOverrides(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_overrides": overrides,
		"count":             len(overrides),
	})
}
