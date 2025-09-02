package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// MilestoneOrchestrationHandlers contains handlers for milestone orchestration operations
type MilestoneOrchestrationHandlers struct {
	milestoneOrchestrationService services.MilestoneOrchestrationServiceInterface
}

// NewMilestoneOrchestrationHandlers creates a new instance of milestone orchestration handlers
func NewMilestoneOrchestrationHandlers(
	milestoneOrchestrationService services.MilestoneOrchestrationServiceInterface,
) *MilestoneOrchestrationHandlers {
	return &MilestoneOrchestrationHandlers{
		milestoneOrchestrationService: milestoneOrchestrationService,
	}
}

// RegisterRoutes registers all milestone orchestration routes
func (h *MilestoneOrchestrationHandlers) RegisterRoutes(router *gin.RouterGroup) {
	orchestration := router.Group("/milestone-orchestration")
	{
		// Milestone orchestration operations
		orchestration.POST("/contracts/:contractId/milestones", h.CreateMilestonesFromContract)
		orchestration.POST("/contracts/:contractId/resolve-dependencies", h.ResolveMilestoneDependencies)
		orchestration.POST("/contracts/:contractId/optimize-schedule", h.OptimizeMilestoneSchedule)
		orchestration.PUT("/milestones/:milestoneId/progress", h.UpdateMilestoneProgress)
		orchestration.GET("/milestones/:milestoneId/validate", h.ValidateMilestoneCompletion)

		// Analytics endpoints
		orchestration.GET("/contracts/:contractId/timeline", h.GetMilestoneTimeline)
		orchestration.GET("/contracts/:contractId/performance-metrics", h.GetMilestonePerformanceMetrics)
		orchestration.GET("/contracts/:contractId/risk-analysis", h.GetMilestoneRiskAnalysis)
	}
}

// CreateMilestonesFromContract creates milestones based on contract analysis
func (h *MilestoneOrchestrationHandlers) CreateMilestonesFromContract(c *gin.Context) {
	contractID := c.Param("contractId")

	// In a real implementation, we would fetch the contract from the repository
	// For now, we'll create a mock contract and call the service
	// TODO: Integrate with the contract service to fetch the actual contract

	// Create a mock contract for demonstration
	now := time.Now()
	contract := &models.Contract{
		ID:        contractID,
		CreatedAt: now,
		Obligations: []models.Obligation{
			{
				ID:          "ob-1",
				Description: "Initial payment milestone",
				DueDate:     now.Add(30 * 24 * time.Hour), // 30 days from now
			},
			{
				ID:          "ob-2",
				Description: "Delivery milestone",
				DueDate:     now.Add(60 * 24 * time.Hour), // 60 days from now
			},
			{
				ID:          "ob-3",
				Description: "Final payment milestone",
				DueDate:     now.Add(90 * 24 * time.Hour), // 90 days from now
			},
		},
	}

	// Call the service to create milestones
	milestones, err := h.milestoneOrchestrationService.CreateMilestonesFromContract(c.Request.Context(), contract)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Milestones created successfully",
		"contractId": contractID,
		"milestones": milestones,
	})
}

// ResolveMilestoneDependencies automatically resolves dependencies between milestones
func (h *MilestoneOrchestrationHandlers) ResolveMilestoneDependencies(c *gin.Context) {
	contractID := c.Param("contractId")

	// Call the service to resolve dependencies
	if err := h.milestoneOrchestrationService.ResolveMilestoneDependencies(c.Request.Context(), contractID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Milestone dependencies resolved successfully",
		"contractId": contractID,
	})
}

// OptimizeMilestoneSchedule optimizes the timeline for milestone execution
func (h *MilestoneOrchestrationHandlers) OptimizeMilestoneSchedule(c *gin.Context) {
	contractID := c.Param("contractId")

	// Call the service to optimize the schedule
	if err := h.milestoneOrchestrationService.OptimizeMilestoneSchedule(c.Request.Context(), contractID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Milestone schedule optimized successfully",
		"contractId": contractID,
	})
}

// UpdateMilestoneProgress updates the progress of a milestone
func (h *MilestoneOrchestrationHandlers) UpdateMilestoneProgress(c *gin.Context) {
	milestoneID := c.Param("milestoneId")

	// Parse request body
	var request struct {
		Progress float64 `json:"progress" binding:"required"`
		Notes    string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Call the service to update progress
	if err := h.milestoneOrchestrationService.UpdateMilestoneProgress(c.Request.Context(), milestoneID, request.Progress, request.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Milestone progress updated successfully",
		"milestoneId": milestoneID,
		"progress":    request.Progress,
	})
}

// ValidateMilestoneCompletion validates if a milestone has been properly completed
func (h *MilestoneOrchestrationHandlers) ValidateMilestoneCompletion(c *gin.Context) {
	milestoneID := c.Param("milestoneId")

	// Call the service to validate completion
	isValid, err := h.milestoneOrchestrationService.ValidateMilestoneCompletion(c.Request.Context(), milestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"milestoneId": milestoneID,
		"isValid":     isValid,
	})
}

// GetMilestoneTimeline returns the timeline analysis for a contract's milestones
func (h *MilestoneOrchestrationHandlers) GetMilestoneTimeline(c *gin.Context) {
	contractID := c.Param("contractId")

	// Call the service to get timeline analysis
	timeline, err := h.milestoneOrchestrationService.GetMilestoneTimeline(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timeline)
}

// GetMilestonePerformanceMetrics returns performance metrics for milestones
func (h *MilestoneOrchestrationHandlers) GetMilestonePerformanceMetrics(c *gin.Context) {
	contractID := c.Param("contractId")

	// Call the service to get performance metrics
	metrics, err := h.milestoneOrchestrationService.GetMilestonePerformanceMetrics(c.Request.Context(), &contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetMilestoneRiskAnalysis returns risk analysis for milestones
func (h *MilestoneOrchestrationHandlers) GetMilestoneRiskAnalysis(c *gin.Context) {
	contractID := c.Param("contractId")

	// Call the service to get risk analysis
	riskAnalysis, err := h.milestoneOrchestrationService.GetMilestoneRiskAnalysis(c.Request.Context(), &contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, riskAnalysis)
}
