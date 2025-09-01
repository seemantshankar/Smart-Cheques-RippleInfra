package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// OracleHandler handles HTTP requests for oracle functionality
type OracleHandler struct {
	oracleService       *services.OracleService
	verificationService *services.OracleVerificationService
	monitoringService   *services.OracleMonitoringService
}

// NewOracleHandler creates a new oracle handler
func NewOracleHandler(
	oracleService *services.OracleService,
	verificationService *services.OracleVerificationService,
	monitoringService *services.OracleMonitoringService,
) *OracleHandler {
	return &OracleHandler{
		oracleService:       oracleService,
		verificationService: verificationService,
		monitoringService:   monitoringService,
	}
}

// RegisterProvider registers a new oracle provider
func (h *OracleHandler) RegisterProvider(c *gin.Context) {
	var provider models.OracleProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid JSON", err))
		return
	}

	// Set default values
	provider.ID = uuid.New()
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()
	provider.IsActive = true
	provider.Reliability = 1.0

	if err := h.oracleService.RegisterProvider(c.Request.Context(), &provider); err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to register provider", err))
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetProvider retrieves an oracle provider by ID
func (h *OracleHandler) GetProvider(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid provider ID", err))
		return
	}

	provider, err := h.oracleService.GetProvider(c.Request.Context(), providerID)
	if err != nil {
		c.Error(models.NewAppError(http.StatusNotFound, "Failed to get provider", err))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// UpdateProvider updates an existing oracle provider
func (h *OracleHandler) UpdateProvider(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid provider ID", err))
		return
	}

	var provider models.OracleProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid JSON", err))
		return
	}

	// Set the ID from the URL
	provider.ID = providerID
	provider.UpdatedAt = time.Now()

	if err := h.oracleService.UpdateProvider(c.Request.Context(), &provider); err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to update provider", err))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider deletes an oracle provider
func (h *OracleHandler) DeleteProvider(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid provider ID", err))
		return
	}

	if err := h.oracleService.DeleteProvider(c.Request.Context(), providerID); err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to delete provider", err))
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProviders lists all oracle providers
func (h *OracleHandler) ListProviders(c *gin.Context) {
	// Parse query parameters
	limit := 100
	offset := 0

	providers, err := h.oracleService.ListProviders(c.Request.Context(), limit, offset)
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to list providers", err))
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetActiveProviders retrieves all active oracle providers
func (h *OracleHandler) GetActiveProviders(c *gin.Context) {
	providers, err := h.oracleService.GetActiveProviders(c.Request.Context())
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to get active providers", err))
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetProvidersByType retrieves oracle providers by type
func (h *OracleHandler) GetProvidersByType(c *gin.Context) {
	providerType := models.OracleType(c.Param("type"))

	providers, err := h.oracleService.GetProvidersByType(c.Request.Context(), providerType)
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to get providers by type", err))
		return
	}

	c.JSON(http.StatusOK, providers)
}

// HealthCheck performs a health check on a provider
func (h *OracleHandler) HealthCheck(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid provider ID", err))
		return
	}

	status, err := h.oracleService.HealthCheck(c.Request.Context(), providerID)
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to perform health check", err))
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetRequest retrieves an oracle request by ID
func (h *OracleHandler) GetRequest(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request ID", err))
		return
	}

	request, err := h.oracleService.GetRequest(c.Request.Context(), requestID)
	if err != nil {
		c.Error(models.NewAppError(http.StatusNotFound, "Failed to get request", err))
		return
	}

	c.JSON(http.StatusOK, request)
}

// VerifyMilestone verifies a milestone using oracle providers
func (h *OracleHandler) VerifyMilestone(c *gin.Context) {
	var req struct {
		MilestoneID  string               `json:"milestone_id"`
		Condition    string               `json:"condition"`
		OracleConfig *models.OracleConfig `json:"oracle_config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid JSON", err))
		return
	}

	if req.MilestoneID == "" || req.Condition == "" || req.OracleConfig == nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Missing required fields: milestone_id, condition, oracle_config", nil))
		return
	}

	response, err := h.verificationService.VerifyMilestone(c.Request.Context(), req.MilestoneID, req.Condition, req.OracleConfig)
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to verify milestone", err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetVerificationResult retrieves the result of a previous verification
func (h *OracleHandler) GetVerificationResult(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("request_id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request ID", err))
		return
	}

	response, err := h.verificationService.GetVerificationResult(c.Request.Context(), requestID)
	if err != nil {
		c.Error(models.NewAppError(http.StatusNotFound, "Failed to get verification result", err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProof retrieves verification evidence for a completed verification
func (h *OracleHandler) GetProof(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("request_id"))
	if err != nil {
		c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request ID", err))
		return
	}

	evidence, err := h.verificationService.GetProof(c.Request.Context(), requestID)
	if err != nil {
		c.Error(models.NewAppError(http.StatusNotFound, "Failed to get proof", err))
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", evidence)
}

// GetDashboardMetrics retrieves metrics for the oracle monitoring dashboard
func (h *OracleHandler) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.monitoringService.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to get dashboard metrics", err))
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSLAMonitoring retrieves SLA monitoring data for oracle providers
func (h *OracleHandler) GetSLAMonitoring(c *gin.Context) {
	report, err := h.monitoringService.GetSLAMonitoring(c.Request.Context())
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to get SLA monitoring report", err))
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetCostAnalysis retrieves cost analysis for oracle usage
func (h *OracleHandler) GetCostAnalysis(c *gin.Context) {
	report, err := h.monitoringService.GetCostAnalysis(c.Request.Context())
	if err != nil {
		c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to get cost analysis report", err))
		return
	}

	c.JSON(http.StatusOK, report)
}
