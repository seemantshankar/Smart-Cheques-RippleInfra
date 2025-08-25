package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/services"
)

// AnomalyDetectionHandler handles HTTP requests for anomaly detection operations
type AnomalyDetectionHandler struct {
	anomalyService services.AnomalyDetectionServiceInterface
}

// NewAnomalyDetectionHandler creates a new anomaly detection handler
func NewAnomalyDetectionHandler(anomalyService services.AnomalyDetectionServiceInterface) *AnomalyDetectionHandler {
	return &AnomalyDetectionHandler{
		anomalyService: anomalyService,
	}
}

// RegisterRoutes registers all anomaly detection routes
func (h *AnomalyDetectionHandler) RegisterRoutes(router *gin.RouterGroup) {
	anomaly := router.Group("/anomaly-detection")
	{
		// Real-time analysis
		anomaly.POST("/analyze", h.AnalyzeTransaction)
		anomaly.GET("/enterprises/:enterpriseID/patterns", h.DetectPatternAnomalies)

		// Batch analysis
		anomaly.POST("/batch-analysis", h.PerformBatchAnalysis)
		anomaly.POST("/reports/generate", h.GenerateAnomalyReport)

		// Threshold management
		anomaly.POST("/thresholds", h.SetAnomalyThresholds)
		anomaly.GET("/enterprises/:enterpriseID/thresholds", h.GetAnomalyThresholds)

		// Investigation workflow
		anomaly.GET("/investigations/:anomalyID", h.InvestigateAnomaly)
		anomaly.POST("/feedback", h.SubmitFeedback)

		// Model management
		anomaly.POST("/models/train", h.TrainDetectionModel)
		anomaly.GET("/models/performance", h.GetModelPerformance)
	}
}

// AnalyzeTransaction analyzes a transaction for anomalies
func (h *AnomalyDetectionHandler) AnalyzeTransaction(c *gin.Context) {
	var req services.TransactionAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score, err := h.anomalyService.AnalyzeTransaction(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction analysis completed",
		"score":   score,
	})
}

// DetectPatternAnomalies detects pattern anomalies for an enterprise
func (h *AnomalyDetectionHandler) DetectPatternAnomalies(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	anomalies, err := h.anomalyService.DetectPatternAnomalies(c.Request.Context(), enterpriseID)
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

// PerformBatchAnalysis performs batch anomaly analysis
func (h *AnomalyDetectionHandler) PerformBatchAnalysis(c *gin.Context) {
	var req services.BatchAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.anomalyService.PerformBatchAnalysis(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch analysis completed",
		"result":  result,
	})
}

// GenerateAnomalyReport generates an anomaly report
func (h *AnomalyDetectionHandler) GenerateAnomalyReport(c *gin.Context) {
	var req services.AnomalyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.anomalyService.GenerateAnomalyReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// SetAnomalyThresholds sets anomaly detection thresholds
func (h *AnomalyDetectionHandler) SetAnomalyThresholds(c *gin.Context) {
	var req services.AnomalyThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	threshold, err := h.anomalyService.SetAnomalyThresholds(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Anomaly threshold created successfully",
		"threshold": threshold,
	})
}

// GetAnomalyThresholds gets anomaly thresholds for an enterprise
func (h *AnomalyDetectionHandler) GetAnomalyThresholds(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	thresholds, err := h.anomalyService.GetAnomalyThresholds(c.Request.Context(), enterpriseID)
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

// InvestigateAnomaly starts or gets an anomaly investigation
func (h *AnomalyDetectionHandler) InvestigateAnomaly(c *gin.Context) {
	anomalyIDStr := c.Param("anomalyID")
	anomalyID, err := uuid.Parse(anomalyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid anomaly ID"})
		return
	}

	investigation, err := h.anomalyService.InvestigateAnomaly(c.Request.Context(), anomalyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, investigation)
}

// SubmitFeedback submits feedback on anomaly detection accuracy
func (h *AnomalyDetectionHandler) SubmitFeedback(c *gin.Context) {
	var req services.AnomalyFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.anomalyService.SubmitFeedback(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Feedback submitted successfully",
	})
}

// TrainDetectionModel trains a new anomaly detection model
func (h *AnomalyDetectionHandler) TrainDetectionModel(c *gin.Context) {
	var req services.ModelTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.anomalyService.TrainDetectionModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model training completed",
		"result":  result,
	})
}

// GetModelPerformance gets current model performance metrics
func (h *AnomalyDetectionHandler) GetModelPerformance(c *gin.Context) {
	performance, err := h.anomalyService.GetModelPerformance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}
