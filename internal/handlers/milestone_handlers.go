package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// MilestoneHandlers contains handlers for milestone-related operations
type MilestoneHandlers struct {
	milestoneRepo repository.MilestoneRepositoryInterface
	templateRepo  repository.MilestoneTemplateRepositoryInterface
}

// NewMilestoneHandlers creates a new instance of milestone handlers
func NewMilestoneHandlers(
	milestoneRepo repository.MilestoneRepositoryInterface,
	templateRepo repository.MilestoneTemplateRepositoryInterface,
) *MilestoneHandlers {
	return &MilestoneHandlers{
		milestoneRepo: milestoneRepo,
		templateRepo:  templateRepo,
	}
}

// RegisterRoutes registers all milestone-related routes
func (h *MilestoneHandlers) RegisterRoutes(router *gin.RouterGroup) {
	milestones := router.Group("/milestones")
	{
		// Milestone CRUD operations
		milestones.POST("", h.CreateMilestone)
		milestones.GET("/:id", h.GetMilestone)
		milestones.PUT("/:id", h.UpdateMilestone)
		milestones.DELETE("/:id", h.DeleteMilestone)

		// Milestone queries
		milestones.GET("", h.GetMilestones)
		milestones.GET("/status/:status", h.GetMilestonesByStatus)
		milestones.GET("/overdue", h.GetOverdueMilestones)
		milestones.GET("/search", h.SearchMilestones)

		// Milestone dependencies
		milestones.GET("/:id/dependencies", h.GetMilestoneDependencies)
		milestones.GET("/:id/dependents", h.GetMilestoneDependents)

		// Batch operations
		milestones.PUT("/batch/status", h.BatchUpdateStatus)
		milestones.PUT("/batch/progress", h.BatchUpdateProgress)
		milestones.POST("/batch", h.BatchCreateMilestones)
		milestones.DELETE("/batch", h.BatchDeleteMilestones)

		// Analytics and reporting
		milestones.GET("/performance-metrics", h.GetPerformanceMetrics)
		milestones.GET("/risk-analysis", h.GetRiskAnalysis)
		milestones.GET("/progress-trends", h.GetProgressTrends)
		milestones.GET("/delayed-report", h.GetDelayedReport)

		// Progress tracking
		milestones.POST("/:id/progress", h.CreateProgressEntry)
		milestones.GET("/:id/progress-history", h.GetProgressHistory)
		milestones.GET("/:id/latest-progress", h.GetLatestProgress)

		// Advanced filtering
		milestones.POST("/filter", h.FilterMilestones)
		milestones.GET("/upcoming", h.GetUpcomingMilestones)
	}

	// Contract-specific milestone routes
	contracts := router.Group("/contracts")
	{
		contracts.GET("/:contractId/milestones", h.GetMilestonesByContract)
		contracts.GET("/:contractId/milestones/critical-path", h.GetCriticalPathMilestones)
		contracts.GET("/:contractId/dependency-graph", h.GetDependencyGraph)
		contracts.GET("/:contractId/topological-order", h.GetTopologicalOrder)
		contracts.GET("/:contractId/milestone-stats", h.GetMilestoneStats)
		contracts.GET("/:contractId/timeline-analysis", h.GetTimelineAnalysis)
	}

	// Milestone dependencies
	dependencies := router.Group("/milestone-dependencies")
	{
		dependencies.POST("", h.CreateMilestoneDependency)
		dependencies.DELETE("/:id", h.DeleteMilestoneDependency)
	}

	// Template operations
	templates := router.Group("/milestone-templates")
	{
		templates.POST("", h.CreateTemplate)
		templates.GET("/:id", h.GetTemplate)
		templates.PUT("/:id", h.UpdateTemplate)
		templates.DELETE("/:id", h.DeleteTemplate)
		templates.GET("", h.GetTemplates)

		// Template instantiation and customization
		templates.POST("/:id/instantiate", h.InstantiateTemplate)
		templates.POST("/:id/customize", h.CustomizeTemplate)
		templates.GET("/:id/variables", h.GetTemplateVariables)

		// Template versioning
		templates.POST("/:id/versions", h.CreateTemplateVersion)
		templates.GET("/:id/versions", h.GetTemplateVersions)
		templates.GET("/:id/versions/:version", h.GetTemplateVersion)
		templates.GET("/:id/versions/latest", h.GetLatestTemplateVersion)
		templates.GET("/:id/versions/compare", h.CompareTemplateVersions)

		// Template sharing
		templates.POST("/:id/share", h.ShareTemplate)
		templates.DELETE("/:id/revoke/:userId", h.RevokeTemplateAccess)
		templates.GET("/:id/permissions/:userId", h.GetTemplatePermissions)
		templates.GET("/:id/shares", h.GetTemplateShares)
	}

	// User-specific template routes
	users := router.Group("/users")
	{
		users.GET("/:userId/shared-templates", h.GetSharedTemplates)
	}
}

// Milestone CRUD handlers

func (h *MilestoneHandlers) CreateMilestone(c *gin.Context) {
	var milestone models.ContractMilestone
	if err := c.ShouldBindJSON(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Set timestamps
	now := time.Now()
	milestone.CreatedAt = now
	milestone.UpdatedAt = now

	if err := h.milestoneRepo.CreateMilestone(c.Request.Context(), &milestone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create milestone: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, milestone)
}

func (h *MilestoneHandlers) GetMilestone(c *gin.Context) {
	milestoneID := c.Param("id")

	milestone, err := h.milestoneRepo.GetMilestoneByID(c.Request.Context(), milestoneID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestone: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestone)
}

func (h *MilestoneHandlers) UpdateMilestone(c *gin.Context) {
	milestoneID := c.Param("id")

	var milestone models.ContractMilestone
	if err := c.ShouldBindJSON(&milestone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	milestone.ID = milestoneID
	milestone.UpdatedAt = time.Now()

	if err := h.milestoneRepo.UpdateMilestone(c.Request.Context(), &milestone); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update milestone: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestone)
}

func (h *MilestoneHandlers) DeleteMilestone(c *gin.Context) {
	milestoneID := c.Param("id")

	if err := h.milestoneRepo.DeleteMilestone(c.Request.Context(), milestoneID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Milestone not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete milestone: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

// Milestone query handlers

func (h *MilestoneHandlers) GetMilestones(c *gin.Context) {
	limit, offset := h.getPaginationParams(c)

	// This would need a generic GetMilestones method in the interface
	// For now, return all milestones using a broad filter
	filter := &repository.MilestoneFilter{}
	milestones, err := h.milestoneRepo.FilterMilestones(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestones: %v", err)})
		return
	}

	// Apply pagination manually (better to do in database)
	start := offset
	end := offset + limit
	if start > len(milestones) {
		start = len(milestones)
	}
	if end > len(milestones) {
		end = len(milestones)
	}

	result := milestones[start:end]

	c.JSON(http.StatusOK, result)
}

func (h *MilestoneHandlers) GetMilestonesByContract(c *gin.Context) {
	contractID := c.Param("contractId")
	limit, offset := h.getPaginationParams(c)

	milestones, err := h.milestoneRepo.GetMilestonesByContract(c.Request.Context(), contractID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestones by contract: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *MilestoneHandlers) GetMilestonesByStatus(c *gin.Context) {
	status := c.Param("status")
	limit, offset := h.getPaginationParams(c)

	milestones, err := h.milestoneRepo.GetMilestonesByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestones by status: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *MilestoneHandlers) GetOverdueMilestones(c *gin.Context) {
	limit, offset := h.getPaginationParams(c)
	asOfDate := time.Now()

	// Allow custom date via query parameter
	if dateStr := c.Query("as_of_date"); dateStr != "" {
		if parsedDate, err := time.Parse(time.RFC3339, dateStr); err == nil {
			asOfDate = parsedDate
		}
	}

	milestones, err := h.milestoneRepo.GetOverdueMilestones(c.Request.Context(), asOfDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get overdue milestones: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *MilestoneHandlers) GetCriticalPathMilestones(c *gin.Context) {
	contractID := c.Param("contractId")

	milestones, err := h.milestoneRepo.GetCriticalPathMilestones(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get critical path milestones: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *MilestoneHandlers) SearchMilestones(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit, offset := h.getPaginationParams(c)

	milestones, err := h.milestoneRepo.SearchMilestones(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to search milestones: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

// Dependency handlers

func (h *MilestoneHandlers) GetMilestoneDependencies(c *gin.Context) {
	milestoneID := c.Param("id")

	dependencies, err := h.milestoneRepo.GetMilestoneDependencies(c.Request.Context(), milestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestone dependencies: %v", err)})
		return
	}

	c.JSON(http.StatusOK, dependencies)
}

func (h *MilestoneHandlers) GetMilestoneDependents(c *gin.Context) {
	milestoneID := c.Param("id")

	dependents, err := h.milestoneRepo.GetMilestoneDependents(c.Request.Context(), milestoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestone dependents: %v", err)})
		return
	}

	c.JSON(http.StatusOK, dependents)
}

func (h *MilestoneHandlers) CreateMilestoneDependency(c *gin.Context) {
	var dependency models.MilestoneDependency
	if err := c.ShouldBindJSON(&dependency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if err := h.milestoneRepo.CreateMilestoneDependency(c.Request.Context(), &dependency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create milestone dependency: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, dependency)
}

func (h *MilestoneHandlers) DeleteMilestoneDependency(c *gin.Context) {
	dependencyID := c.Param("id")

	if err := h.milestoneRepo.DeleteMilestoneDependency(c.Request.Context(), dependencyID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Dependency not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete milestone dependency: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MilestoneHandlers) GetDependencyGraph(c *gin.Context) {
	contractID := c.Param("contractId")

	graph, err := h.milestoneRepo.ResolveDependencyGraph(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get dependency graph: %v", err)})
		return
	}

	c.JSON(http.StatusOK, graph)
}

func (h *MilestoneHandlers) GetTopologicalOrder(c *gin.Context) {
	contractID := c.Param("contractId")

	order, err := h.milestoneRepo.GetTopologicalOrder(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get topological order: %v", err)})
		return
	}

	response := map[string]interface{}{
		"contract_id": contractID,
		"order":       order,
		"valid":       true, // If we got here, the graph is valid
	}

	c.JSON(http.StatusOK, response)
}

// Batch operation handlers

func (h *MilestoneHandlers) BatchUpdateStatus(c *gin.Context) {
	var request struct {
		MilestoneIDs []string `json:"milestone_ids"`
		Status       string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if len(request.MilestoneIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "milestone_ids cannot be empty"})
		return
	}

	if err := h.milestoneRepo.BatchUpdateMilestoneStatus(c.Request.Context(), request.MilestoneIDs, request.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to batch update status: %v", err)})
		return
	}

	response := map[string]interface{}{
		"updated_count": len(request.MilestoneIDs),
		"status":        request.Status,
	}

	c.JSON(http.StatusOK, response)
}

func (h *MilestoneHandlers) BatchUpdateProgress(c *gin.Context) {
	var updates []repository.MilestoneProgressUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Updates cannot be empty"})
		return
	}

	if err := h.milestoneRepo.BatchUpdateMilestoneProgress(c.Request.Context(), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to batch update progress: %v", err)})
		return
	}

	response := map[string]interface{}{
		"updated_count": len(updates),
	}

	c.JSON(http.StatusOK, response)
}

func (h *MilestoneHandlers) BatchCreateMilestones(c *gin.Context) {
	var milestones []*models.ContractMilestone
	if err := c.ShouldBindJSON(&milestones); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if len(milestones) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Milestones cannot be empty"})
		return
	}

	// Set timestamps for all milestones
	now := time.Now()
	for _, milestone := range milestones {
		milestone.CreatedAt = now
		milestone.UpdatedAt = now
	}

	if err := h.milestoneRepo.BatchCreateMilestones(c.Request.Context(), milestones); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to batch create milestones: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, milestones)
}

func (h *MilestoneHandlers) BatchDeleteMilestones(c *gin.Context) {
	var request struct {
		MilestoneIDs []string `json:"milestone_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if len(request.MilestoneIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "milestone_ids cannot be empty"})
		return
	}

	if err := h.milestoneRepo.BatchDeleteMilestones(c.Request.Context(), request.MilestoneIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to batch delete milestones: %v", err)})
		return
	}

	response := map[string]interface{}{
		"deleted_count": len(request.MilestoneIDs),
	}

	c.JSON(http.StatusOK, response)
}

// Analytics and reporting handlers

func (h *MilestoneHandlers) GetMilestoneStats(c *gin.Context) {
	contractID := c.Param("contractId")

	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &parsed
		}
	}

	stats, err := h.milestoneRepo.GetMilestoneCompletionStats(c.Request.Context(), &contractID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get milestone stats: %v", err)})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *MilestoneHandlers) GetPerformanceMetrics(c *gin.Context) {
	var contractID *string
	if contractIDParam := c.Query("contract_id"); contractIDParam != "" {
		contractID = &contractIDParam
	}

	metrics, err := h.milestoneRepo.GetMilestonePerformanceMetrics(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get performance metrics: %v", err)})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *MilestoneHandlers) GetTimelineAnalysis(c *gin.Context) {
	contractID := c.Param("contractId")

	analysis, err := h.milestoneRepo.GetMilestoneTimelineAnalysis(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get timeline analysis: %v", err)})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

func (h *MilestoneHandlers) GetRiskAnalysis(c *gin.Context) {
	var contractID *string
	if contractIDParam := c.Query("contract_id"); contractIDParam != "" {
		contractID = &contractIDParam
	}

	analysis, err := h.milestoneRepo.GetMilestoneRiskAnalysis(c.Request.Context(), contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get risk analysis: %v", err)})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

func (h *MilestoneHandlers) GetProgressTrends(c *gin.Context) {
	days := 30 // default
	if daysStr := c.Query("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}

	var contractID *string
	if contractIDParam := c.Query("contract_id"); contractIDParam != "" {
		contractID = &contractIDParam
	}

	trends, err := h.milestoneRepo.GetMilestoneProgressTrends(c.Request.Context(), contractID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get progress trends: %v", err)})
		return
	}

	c.JSON(http.StatusOK, trends)
}

func (h *MilestoneHandlers) GetDelayedReport(c *gin.Context) {
	threshold := 24 * time.Hour // default 1 day
	if thresholdStr := c.Query("threshold_hours"); thresholdStr != "" {
		if parsed, err := strconv.Atoi(thresholdStr); err == nil && parsed > 0 {
			threshold = time.Duration(parsed) * time.Hour
		}
	}

	report, err := h.milestoneRepo.GetDelayedMilestonesReport(c.Request.Context(), threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get delayed milestones report: %v", err)})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Progress tracking handlers

func (h *MilestoneHandlers) CreateProgressEntry(c *gin.Context) {
	milestoneID := c.Param("id")

	var entry repository.MilestoneProgressEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	entry.MilestoneID = milestoneID
	entry.CreatedAt = time.Now()
	if entry.RecordedAt.IsZero() {
		entry.RecordedAt = time.Now()
	}

	if err := h.milestoneRepo.CreateMilestoneProgressEntry(c.Request.Context(), &entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create progress entry: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *MilestoneHandlers) GetProgressHistory(c *gin.Context) {
	milestoneID := c.Param("id")
	limit, offset := h.getPaginationParams(c)

	history, err := h.milestoneRepo.GetMilestoneProgressHistory(c.Request.Context(), milestoneID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get progress history: %v", err)})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *MilestoneHandlers) GetLatestProgress(c *gin.Context) {
	milestoneID := c.Param("id")

	progress, err := h.milestoneRepo.GetLatestProgressUpdate(c.Request.Context(), milestoneID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "No progress updates found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get latest progress: %v", err)})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// Advanced filtering handlers

func (h *MilestoneHandlers) FilterMilestones(c *gin.Context) {
	var filter repository.MilestoneFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	milestones, err := h.milestoneRepo.FilterMilestones(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to filter milestones: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (h *MilestoneHandlers) GetUpcomingMilestones(c *gin.Context) {
	daysAhead := 7 // default 7 days
	if daysStr := c.Query("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			daysAhead = parsed
		}
	}

	limit, offset := h.getPaginationParams(c)

	milestones, err := h.milestoneRepo.GetUpcomingMilestones(c.Request.Context(), daysAhead, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get upcoming milestones: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

// Template handlers

func (h *MilestoneHandlers) CreateTemplate(c *gin.Context) {
	var template models.MilestoneTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	if err := h.templateRepo.CreateTemplate(c.Request.Context(), &template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create template: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *MilestoneHandlers) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	template, err := h.templateRepo.GetTemplateByID(c.Request.Context(), templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *MilestoneHandlers) UpdateTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var template models.MilestoneTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	template.ID = templateID
	template.UpdatedAt = time.Now()

	if err := h.templateRepo.UpdateTemplate(c.Request.Context(), &template); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *MilestoneHandlers) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")

	if err := h.templateRepo.DeleteTemplate(c.Request.Context(), templateID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete template: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MilestoneHandlers) GetTemplates(c *gin.Context) {
	limit, offset := h.getPaginationParams(c)

	templates, err := h.templateRepo.GetTemplates(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get templates: %v", err)})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *MilestoneHandlers) InstantiateTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var variables map[string]interface{}
	if err := c.ShouldBindJSON(&variables); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	milestone, err := h.templateRepo.InstantiateTemplate(c.Request.Context(), templateID, variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to instantiate template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, milestone)
}

func (h *MilestoneHandlers) CustomizeTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var customizations map[string]interface{}
	if err := c.ShouldBindJSON(&customizations); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	customizedTemplate, err := h.templateRepo.CustomizeTemplate(c.Request.Context(), templateID, customizations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to customize template: %v", err)})
		return
	}

	c.JSON(http.StatusOK, customizedTemplate)
}

func (h *MilestoneHandlers) GetTemplateVariables(c *gin.Context) {
	templateID := c.Param("id")

	variables, err := h.templateRepo.GetTemplateVariables(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template variables: %v", err)})
		return
	}

	response := map[string]interface{}{
		"template_id": templateID,
		"variables":   variables,
	}

	c.JSON(http.StatusOK, response)
}

// Template versioning handlers

func (h *MilestoneHandlers) CreateTemplateVersion(c *gin.Context) {
	templateID := c.Param("id")

	var version models.MilestoneTemplate
	if err := c.ShouldBindJSON(&version); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	version.CreatedAt = time.Now()

	if err := h.templateRepo.CreateTemplateVersion(c.Request.Context(), templateID, &version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create template version: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, version)
}

func (h *MilestoneHandlers) GetTemplateVersions(c *gin.Context) {
	templateID := c.Param("id")

	versions, err := h.templateRepo.GetTemplateVersions(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template versions: %v", err)})
		return
	}

	c.JSON(http.StatusOK, versions)
}

func (h *MilestoneHandlers) GetTemplateVersion(c *gin.Context) {
	templateID := c.Param("id")
	version := c.Param("version")

	templateVersion, err := h.templateRepo.GetTemplateVersion(c.Request.Context(), templateID, version)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template version: %v", err)})
		return
	}

	c.JSON(http.StatusOK, templateVersion)
}

func (h *MilestoneHandlers) GetLatestTemplateVersion(c *gin.Context) {
	templateID := c.Param("id")

	latestVersion, err := h.templateRepo.GetLatestTemplateVersion(c.Request.Context(), templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "No template versions found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get latest template version: %v", err)})
		return
	}

	c.JSON(http.StatusOK, latestVersion)
}

func (h *MilestoneHandlers) CompareTemplateVersions(c *gin.Context) {
	templateID := c.Param("id")
	version1 := c.Query("version1")
	version2 := c.Query("version2")

	if version1 == "" || version2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both version1 and version2 query parameters are required"})
		return
	}

	diff, err := h.templateRepo.CompareTemplateVersions(c.Request.Context(), templateID, version1, version2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to compare template versions: %v", err)})
		return
	}

	c.JSON(http.StatusOK, diff)
}

// Template sharing handlers

func (h *MilestoneHandlers) ShareTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var shareRequest struct {
		SharedWithUserID string   `json:"shared_with_user_id"`
		Permissions      []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&shareRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if shareRequest.SharedWithUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shared_with_user_id is required"})
		return
	}

	if err := h.templateRepo.ShareTemplate(c.Request.Context(), templateID, shareRequest.SharedWithUserID, shareRequest.Permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to share template: %v", err)})
		return
	}

	response := map[string]interface{}{
		"template_id":         templateID,
		"shared_with_user_id": shareRequest.SharedWithUserID,
		"permissions":         shareRequest.Permissions,
		"shared_at":           time.Now(),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *MilestoneHandlers) RevokeTemplateAccess(c *gin.Context) {
	templateID := c.Param("id")
	userID := c.Param("userId")

	if err := h.templateRepo.RevokeTemplateAccess(c.Request.Context(), templateID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to revoke template access: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MilestoneHandlers) GetSharedTemplates(c *gin.Context) {
	userID := c.Param("userId")
	limit, offset := h.getPaginationParams(c)

	templates, err := h.templateRepo.GetSharedTemplates(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get shared templates: %v", err)})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *MilestoneHandlers) GetTemplatePermissions(c *gin.Context) {
	templateID := c.Param("id")
	userID := c.Param("userId")

	permissions, err := h.templateRepo.GetTemplatePermissions(c.Request.Context(), templateID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template permissions: %v", err)})
		return
	}

	response := map[string]interface{}{
		"template_id": templateID,
		"user_id":     userID,
		"permissions": permissions,
	}

	c.JSON(http.StatusOK, response)
}

func (h *MilestoneHandlers) GetTemplateShares(c *gin.Context) {
	templateID := c.Param("id")

	shares, err := h.templateRepo.GetTemplateShareList(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get template shares: %v", err)})
		return
	}

	c.JSON(http.StatusOK, shares)
}

// Utility methods

func (h *MilestoneHandlers) getPaginationParams(c *gin.Context) (int, int) {
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
