package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// SmartChequeHandler handles HTTP requests for smart checks
type SmartChequeHandler struct {
	smartChequeService services.SmartChequeServiceInterface
}

// NewSmartChequeHandler creates a new smart check handler
func NewSmartChequeHandler(smartChequeService services.SmartChequeServiceInterface) *SmartChequeHandler {
	return &SmartChequeHandler{
		smartChequeService: smartChequeService,
	}
}

// CreateSmartCheque creates a new smart check
// @Summary Create a new smart check
// @Description Create a new smart check with the provided details
// @Tags SmartCheques
// @Accept json
// @Produce json
// @Param smartCheque body services.CreateSmartChequeRequest true "Smart Check details"
// @Success 201 {object} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques [post]
func (h *SmartChequeHandler) CreateSmartCheque(c *gin.Context) {
	var request services.CreateSmartChequeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	smartCheque, err := h.smartChequeService.CreateSmartCheque(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, smartCheque)
}

// GetSmartCheque retrieves a smart check by ID
// @Summary Get a smart check by ID
// @Description Retrieve a smart check by its ID
// @Tags SmartCheques
// @Produce json
// @Param id path string true "Smart Check ID"
// @Success 200 {object} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/{id} [get]
func (h *SmartChequeHandler) GetSmartCheque(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	smartCheque, err := h.smartChequeService.GetSmartCheque(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, smartCheque)
}

// UpdateSmartCheque updates an existing smart check
// @Summary Update a smart check
// @Description Update an existing smart check with the provided details
// @Tags SmartCheques
// @Accept json
// @Produce json
// @Param id path string true "Smart Check ID"
// @Param smartCheque body services.UpdateSmartChequeRequest true "Smart Check update details"
// @Success 200 {object} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/{id} [put]
func (h *SmartChequeHandler) UpdateSmartCheque(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var request services.UpdateSmartChequeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	smartCheque, err := h.smartChequeService.UpdateSmartCheque(c.Request.Context(), id, &request)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, smartCheque)
}

// DeleteSmartCheque deletes a smart check
// @Summary Delete a smart check
// @Description Delete a smart check by its ID
// @Tags SmartCheques
// @Produce json
// @Param id path string true "Smart Check ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/{id} [delete]
func (h *SmartChequeHandler) DeleteSmartCheque(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.smartChequeService.DeleteSmartCheque(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSmartChequesByPayer lists smart checks by payer ID
// @Summary List smart checks by payer
// @Description List smart checks for a specific payer
// @Tags SmartCheques
// @Produce json
// @Param payer_id query string true "Payer ID"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/payer [get]
func (h *SmartChequeHandler) ListSmartChequesByPayer(c *gin.Context) {
	payerID := c.Query("payer_id")
	if payerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payer_id is required"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	smartCheques, err := h.smartChequeService.ListSmartChequesByPayer(c.Request.Context(), payerID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, smartCheques)
}

// ListSmartChequesByPayee lists smart checks by payee ID
// @Summary List smart checks by payee
// @Description List smart checks for a specific payee
// @Tags SmartCheques
// @Produce json
// @Param payee_id query string true "Payee ID"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/payee [get]
func (h *SmartChequeHandler) ListSmartChequesByPayee(c *gin.Context) {
	payeeID := c.Query("payee_id")
	if payeeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payee_id is required"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	smartCheques, err := h.smartChequeService.ListSmartChequesByPayee(c.Request.Context(), payeeID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, smartCheques)
}

// ListSmartChequesByStatus lists smart checks by status
// @Summary List smart checks by status
// @Description List smart checks with a specific status
// @Tags SmartCheques
// @Produce json
// @Param status query string true "Smart Check Status"
// @Param limit query int false "Limit (default: 10, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {array} models.SmartCheque
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/status [get]
func (h *SmartChequeHandler) ListSmartChequesByStatus(c *gin.Context) {
	statusStr := c.Query("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	status := models.SmartChequeStatus(statusStr)
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	smartCheques, err := h.smartChequeService.ListSmartChequesByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, smartCheques)
}

// UpdateSmartChequeStatus updates the status of a smart check
// @Summary Update smart check status
// @Description Update the status of a smart check
// @Tags SmartCheques
// @Produce json
// @Param id path string true "Smart Check ID"
// @Param status query string true "New Status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/{id}/status [put]
func (h *SmartChequeHandler) UpdateSmartChequeStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	statusStr := c.Query("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	status := models.SmartChequeStatus(statusStr)
	if err := h.smartChequeService.UpdateSmartChequeStatus(c.Request.Context(), id, status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// GetSmartChequeStatistics retrieves statistics about smart checks
// @Summary Get smart check statistics
// @Description Get statistics about all smart checks
// @Tags SmartCheques
// @Produce json
// @Success 200 {object} services.SmartChequeStatistics
// @Failure 500 {object} map[string]interface{}
// @Router /smart-cheques/statistics [get]
func (h *SmartChequeHandler) GetSmartChequeStatistics(c *gin.Context) {
	statistics, err := h.smartChequeService.GetSmartChequeStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statistics)
}
