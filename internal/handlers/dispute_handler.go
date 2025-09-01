package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
)

// DisputeHandler handles HTTP requests for dispute operations
type DisputeHandler struct {
	disputeService  services.DisputeManagementService
	disputeRepo     repository.DisputeRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
	categorization  *services.DisputeCategorizationService
	routing         *services.ResolutionRoutingService
}

// NewDisputeHandler creates a new dispute handler
func NewDisputeHandler(
	disputeService services.DisputeManagementService,
	disputeRepo repository.DisputeRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
) *DisputeHandler {
	return &DisputeHandler{
		disputeService:  disputeService,
		disputeRepo:     disputeRepo,
		smartChequeRepo: smartChequeRepo,
	}
}

// WithCategorization injects categorization service for recompute endpoint
func (h *DisputeHandler) WithCategorization(c *services.DisputeCategorizationService) *DisputeHandler {
	h.categorization = c
	return h
}

// WithRouting injects resolution routing service for route suggestion endpoint
func (h *DisputeHandler) WithRouting(r *services.ResolutionRoutingService) *DisputeHandler {
	h.routing = r
	return h
}

// CreateDispute creates a new dispute
// @Summary Create a new dispute
// @Description Create a new dispute with the provided details and optional evidence
// @Tags Disputes
// @Accept json
// @Produce json
// @Param dispute body services.CreateDisputeRequest true "Dispute details"
// @Success 201 {object} models.Dispute
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes [post]
func (h *DisputeHandler) CreateDispute(c *gin.Context) {
	var request services.CreateDisputeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (TODO: implement proper authentication middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // fallback for development
	}

	dispute, err := h.disputeService.CreateDispute(c.Request.Context(), &request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dispute)
}

// GetDisputes retrieves disputes with optional filtering
// @Summary Get disputes
// @Description Retrieve disputes with optional filtering by status, category, etc.
// @Tags Disputes
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param category query string false "Filter by category"
// @Param initiator_id query string false "Filter by initiator ID"
// @Param respondent_id query string false "Filter by respondent ID"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.Dispute
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes [get]
func (h *DisputeHandler) GetDisputes(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	// Build filter from query parameters
	filter := &models.DisputeFilter{}
	if status := c.Query("status"); status != "" {
		filter.Status = (*models.DisputeStatus)(&status)
	}
	if category := c.Query("category"); category != "" {
		filter.Category = (*models.DisputeCategory)(&category)
	}
	if initiatorID := c.Query("initiator_id"); initiatorID != "" {
		filter.InitiatorID = &initiatorID
	}
	if respondentID := c.Query("respondent_id"); respondentID != "" {
		filter.RespondentID = &respondentID
	}
	if smartChequeID := c.Query("smart_cheque_id"); smartChequeID != "" {
		filter.SmartChequeID = &smartChequeID
	}
	if milestoneID := c.Query("milestone_id"); milestoneID != "" {
		filter.MilestoneID = &milestoneID
	}
	if contractID := c.Query("contract_id"); contractID != "" {
		filter.ContractID = &contractID
	}
	if searchText := c.Query("search_text"); searchText != "" {
		filter.SearchText = &searchText
	}

	disputes, err := h.disputeRepo.GetDisputes(c.Request.Context(), filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, disputes)
}

// GetDisputeByID retrieves a specific dispute by ID
// @Summary Get dispute by ID
// @Description Retrieve a specific dispute by its ID with full details
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Success 200 {object} models.Dispute
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id} [get]
func (h *DisputeHandler) GetDisputeByID(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	dispute, err := h.disputeRepo.GetDisputeByID(c.Request.Context(), disputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dispute == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dispute not found"})
		return
	}

	c.JSON(http.StatusOK, dispute)
}

// UpdateDisputeStatus updates the status of a dispute
// @Summary Update dispute status
// @Description Update the status of a dispute with reason
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param status_update body UpdateDisputeStatusRequest true "Status update details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/status [put]
func (h *DisputeHandler) UpdateDisputeStatus(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	var request UpdateDisputeStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // fallback for development
	}

	err := h.disputeService.UpdateDisputeStatus(c.Request.Context(), disputeID, request.Status, userID, request.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dispute status updated successfully"})
}

// AddDisputeEvidence adds evidence to a dispute
// @Summary Add evidence to dispute
// @Description Upload and attach evidence files to a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param evidence body models.DisputeEvidence true "Evidence details"
// @Success 201 {object} models.DisputeEvidence
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/evidence [post]
func (h *DisputeHandler) AddDisputeEvidence(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	var evidence models.DisputeEvidence
	if err := c.ShouldBindJSON(&evidence); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // fallback for development
	}

	evidence.DisputeID = disputeID
	evidence.UploadedBy = userID

	err := h.disputeService.AddDisputeEvidence(c.Request.Context(), disputeID, &evidence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, evidence)
}

// GetDisputeEvidence retrieves evidence for a dispute
// @Summary Get dispute evidence
// @Description Retrieve all evidence files attached to a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Success 200 {array} models.DisputeEvidence
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/evidence [get]
func (h *DisputeHandler) GetDisputeEvidence(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	evidence, err := h.disputeRepo.GetEvidenceByDisputeID(c.Request.Context(), disputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, evidence)
}

// CreateDisputeResolution creates a resolution for a dispute
// @Summary Create dispute resolution
// @Description Create a resolution proposal for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param resolution body models.DisputeResolution true "Resolution details"
// @Success 201 {object} models.DisputeResolution
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/resolution [post]
func (h *DisputeHandler) CreateDisputeResolution(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	var resolution models.DisputeResolution
	if err := c.ShouldBindJSON(&resolution); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resolution.DisputeID = disputeID

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // fallback for development
	}

	err := h.disputeService.CreateDisputeResolution(c.Request.Context(), &resolution, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resolution)
}

// ExecuteDisputeResolution executes an approved dispute resolution
// @Summary Execute dispute resolution
// @Description Execute the approved resolution for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/resolution/execute [post]
func (h *DisputeHandler) ExecuteDisputeResolution(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	// Get user ID from context
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system" // fallback for development
	}

	err := h.disputeService.ExecuteDisputeResolution(c.Request.Context(), disputeID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dispute resolution executed successfully"})
}

// SuggestResolutionRoute suggests an intelligent resolution route for a dispute
// @Summary Suggest resolution route
// @Description Suggest the best resolution method and next actions for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/resolution/suggest [get]
func (h *DisputeHandler) SuggestResolutionRoute(c *gin.Context) {
	if h.routing == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "resolution routing service not configured"})
		return
	}

	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	route, err := h.routing.SuggestRoute(c.Request.Context(), disputeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"method":       route.Method,
		"confidence":   route.Confidence,
		"reasons":      route.Reasons,
		"next_actions": route.NextActions,
		"sla_by":       route.SLABy,
		"alternatives": route.Alternatives,
	})
}

// RecomputePriority triggers a recomputation of dispute priority
// @Summary Recompute dispute priority
// @Description Recompute and, if changed, update the priority for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/priority/recompute [post]
func (h *DisputeHandler) RecomputePriority(c *gin.Context) {
	if h.categorization == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categorization service not configured"})
		return
	}

	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	changed, newPriority, err := h.categorization.RecomputeAndApplyPriority(c.Request.Context(), disputeID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"changed":      changed,
		"new_priority": newPriority,
	})
}

// GetDisputeComments retrieves comments for a dispute
// @Summary Get dispute comments
// @Description Retrieve all comments and notes for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.DisputeComment
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/comments [get]
func (h *DisputeHandler) GetDisputeComments(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	comments, err := h.disputeRepo.GetCommentsByDisputeID(c.Request.Context(), disputeID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// GetDisputeAuditLogs retrieves audit logs for a dispute
// @Summary Get dispute audit logs
// @Description Retrieve audit trail for a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.DisputeAuditLog
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/audit [get]
func (h *DisputeHandler) GetDisputeAuditLogs(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	auditLogs, err := h.disputeRepo.GetAuditLogsByDisputeID(c.Request.Context(), disputeID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, auditLogs)
}

// GetDisputeStats retrieves dispute statistics
// @Summary Get dispute statistics
// @Description Retrieve aggregated statistics about disputes
// @Tags Disputes
// @Accept json
// @Produce json
// @Success 200 {object} models.DisputeStats
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/stats [get]
func (h *DisputeHandler) GetDisputeStats(c *gin.Context) {
	stats, err := h.disputeRepo.GetDisputeStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AddDisputeComment adds a comment to a dispute
// @Summary Add comment to dispute
// @Description Add a comment or note to a dispute
// @Tags Disputes
// @Accept json
// @Produce json
// @Param id path string true "Dispute ID"
// @Param comment body AddCommentRequest true "Comment details"
// @Success 201 {object} models.DisputeComment
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /disputes/{id}/comments [post]
func (h *DisputeHandler) AddDisputeComment(c *gin.Context) {
	disputeID := c.Param("id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dispute ID is required"})
		return
	}

	var request AddCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and type from context
	userID := c.GetString("user_id")
	userType := c.GetString("user_type")
	if userID == "" {
		userID = "system" // fallback for development
	}
	if userType == "" {
		userType = "user" // fallback for development
	}

	comment := &models.DisputeComment{
		DisputeID:  disputeID,
		AuthorID:   userID,
		AuthorType: userType,
		Content:    request.Content,
		IsInternal: request.IsInternal,
	}

	err := h.disputeRepo.CreateComment(c.Request.Context(), comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// Request structs for API endpoints
type UpdateDisputeStatusRequest struct {
	Status models.DisputeStatus `json:"status" validate:"required"`
	Reason string               `json:"reason" validate:"required,min=5,max=500"`
}

type AddCommentRequest struct {
	Content    string `json:"content" validate:"required,min=1,max=1000"`
	IsInternal bool   `json:"is_internal"`
}
