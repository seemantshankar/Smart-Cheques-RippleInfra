package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/internal/services"
)

// MilestoneCompletionTriggerHandler handles HTTP requests for milestone completion triggers
type MilestoneCompletionTriggerHandler struct {
	triggerService services.MilestoneCompletionTriggerServiceInterface
}

// NewMilestoneCompletionTriggerHandler creates a new handler for milestone completion triggers
func NewMilestoneCompletionTriggerHandler(triggerService services.MilestoneCompletionTriggerServiceInterface) *MilestoneCompletionTriggerHandler {
	return &MilestoneCompletionTriggerHandler{
		triggerService: triggerService,
	}
}

// RegisterRoutes registers the routes for milestone completion trigger endpoints
func (h *MilestoneCompletionTriggerHandler) RegisterRoutes(router *gin.RouterGroup) {
	triggers := router.Group("/milestone-completion-triggers")
	{
		triggers.POST("/start", h.StartTriggerMonitoring)
		triggers.POST("/stop", h.StopTriggerMonitoring)
		triggers.GET("/status", h.GetTriggerStatus)
		triggers.POST("/process/:milestoneId", h.ProcessMilestoneCompletion)
		triggers.POST("/publish-event/:milestoneId", h.PublishMilestoneCompletedEvent)
	}
}

// StartTriggerMonitoring starts the milestone completion trigger monitoring
func (h *MilestoneCompletionTriggerHandler) StartTriggerMonitoring(c *gin.Context) {
	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.triggerService.StartTriggerMonitoring(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start trigger monitoring",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Milestone completion trigger monitoring started successfully",
		"status":  "running",
	})
}

// StopTriggerMonitoring stops the milestone completion trigger monitoring
func (h *MilestoneCompletionTriggerHandler) StopTriggerMonitoring(c *gin.Context) {
	if err := h.triggerService.StopTriggerMonitoring(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop trigger monitoring",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Milestone completion trigger monitoring stopped successfully",
		"status":  "stopped",
	})
}

// GetTriggerStatus returns the current status of the trigger monitoring system
func (h *MilestoneCompletionTriggerHandler) GetTriggerStatus(c *gin.Context) {
	status, err := h.triggerService.GetTriggerStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get trigger status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ProcessMilestoneCompletion manually processes a milestone completion
func (h *MilestoneCompletionTriggerHandler) ProcessMilestoneCompletion(c *gin.Context) {
	milestoneID := c.Param("milestoneId")
	if milestoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Milestone ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.triggerService.ProcessMilestoneCompletion(ctx, milestoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process milestone completion",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Milestone completion processed successfully",
		"milestone_id": milestoneID,
	})
}

// PublishMilestoneCompletedEvent publishes a milestone completed event
func (h *MilestoneCompletionTriggerHandler) PublishMilestoneCompletedEvent(c *gin.Context) {
	milestoneID := c.Param("milestoneId")
	if milestoneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Milestone ID is required",
		})
		return
	}

	ctx, cancel := contextWithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := h.triggerService.PublishMilestoneCompletedEvent(ctx, milestoneID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to publish milestone completed event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Milestone completed event published successfully",
		"milestone_id": milestoneID,
	})
}

// contextWithTimeout creates a context with timeout for request handling
func contextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
