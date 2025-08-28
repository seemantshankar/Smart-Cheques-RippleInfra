package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// MintingBurningHandler handles HTTP requests for minting and burning operations
type MintingBurningHandler struct {
	mintingBurningService services.MintingBurningServiceInterface
}

// NewMintingBurningHandler creates a new minting and burning handler
func NewMintingBurningHandler(mintingBurningService services.MintingBurningServiceInterface) *MintingBurningHandler {
	return &MintingBurningHandler{
		mintingBurningService: mintingBurningService,
	}
}

// RegisterRoutes registers all minting and burning routes
func (h *MintingBurningHandler) RegisterRoutes(router *gin.RouterGroup) {
	minting := router.Group("/minting-burning")
	{
		// Minting operations
		minting.POST("/enterprises/:enterpriseID/mint", h.MintWrappedAsset)
		minting.POST("/enterprises/:enterpriseID/validate-collateral", h.ValidateCollateral)
		minting.GET("/enterprises/:enterpriseID/minting-capacity/:wrappedAsset", h.GetMintingCapacity)
		minting.GET("/enterprises/:enterpriseID/minting-history", h.GetMintingHistory)

		// Burning operations
		minting.POST("/enterprises/:enterpriseID/burn", h.BurnWrappedAsset)
		minting.POST("/enterprises/:enterpriseID/initiate-burn", h.InitiateBurning)
		minting.POST("/burning/:burningID/process", h.ProcessBurning)
		minting.GET("/enterprises/:enterpriseID/burning-history", h.GetBurningHistory)

		// Collateral management
		minting.POST("/enterprises/:enterpriseID/collateral/lock", h.LockCollateral)
		minting.DELETE("/collateral/locks/:lockID", h.ReleaseCollateral)
		minting.GET("/enterprises/:enterpriseID/collateral/status", h.GetCollateralStatus)
		minting.GET("/enterprises/:enterpriseID/collateral/ratio/:wrappedAsset", h.GetCollateralRatio)
	}
}

// parseMintingUUIDParam parses a UUID parameter from the request context
func parseMintingUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	paramStr := c.Param(paramName)
	id, err := uuid.Parse(paramStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s", paramName)})
		return uuid.Nil, false
	}
	return id, true
}

// MintWrappedAsset handles wrapped asset minting requests
func (h *MintingBurningHandler) MintWrappedAsset(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.MintingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	result, err := h.mintingBurningService.MintWrappedAsset(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Minting request processed successfully",
		"result":  result,
	})
}

// ValidateCollateral handles collateral validation requests
func (h *MintingBurningHandler) ValidateCollateral(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.CollateralValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	validation, err := h.mintingBurningService.ValidateCollateral(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, validation)
}

// GetMintingCapacity returns minting capacity for an enterprise
func (h *MintingBurningHandler) GetMintingCapacity(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	wrappedAsset := c.Param("wrappedAsset")
	if wrappedAsset == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrapped asset is required"})
		return
	}

	capacity, err := h.mintingBurningService.GetMintingCapacity(c.Request.Context(), enterpriseID, wrappedAsset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, capacity)
}

// BurnWrappedAsset handles wrapped asset burning requests
func (h *MintingBurningHandler) BurnWrappedAsset(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.BurningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	result, err := h.mintingBurningService.BurnWrappedAsset(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Burning request processed successfully",
		"result":  result,
	})
}

// InitiateBurning handles burning initiation with approval workflow
func (h *MintingBurningHandler) InitiateBurning(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.BurningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	result, err := h.mintingBurningService.InitiateBurning(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Burning request initiated successfully",
		"result":  result,
	})
}

// ProcessBurning handles processing of approved burning requests
func (h *MintingBurningHandler) ProcessBurning(c *gin.Context) {
	burningID, valid := parseMintingUUIDParam(c, "burningID")
	if !valid {
		return
	}

	err := h.mintingBurningService.ProcessBurning(c.Request.Context(), burningID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Burning processed successfully",
	})
}

// LockCollateral handles collateral locking requests
func (h *MintingBurningHandler) LockCollateral(c *gin.Context) {
	enterpriseID, valid := parseMintingUUIDParam(c, "enterpriseID")
	if !valid {
		return
	}

	var req services.CollateralLockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set enterprise ID from URL parameter
	req.EnterpriseID = enterpriseID

	lock, err := h.mintingBurningService.LockCollateral(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Collateral locked successfully",
		"lock":    lock,
	})
}

// ReleaseCollateral handles collateral release requests
func (h *MintingBurningHandler) ReleaseCollateral(c *gin.Context) {
	lockID, valid := parseMintingUUIDParam(c, "lockID")
	if !valid {
		return
	}

	err := h.mintingBurningService.ReleaseCollateral(c.Request.Context(), lockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collateral released successfully",
	})
}

// GetCollateralStatus returns collateral status for an enterprise
func (h *MintingBurningHandler) GetCollateralStatus(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	status, err := h.mintingBurningService.GetCollateralStatus(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetCollateralRatio returns collateral ratio for a wrapped asset
func (h *MintingBurningHandler) GetCollateralRatio(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseID")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
		return
	}

	wrappedAsset := c.Param("wrappedAsset")
	if wrappedAsset == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wrapped asset is required"})
		return
	}

	ratio, err := h.mintingBurningService.GetCollateralRatio(c.Request.Context(), enterpriseID, wrappedAsset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ratio)
}

// GetMintingHistory returns minting history for an enterprise
func (h *MintingBurningHandler) GetMintingHistory(c *gin.Context) {
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

	history, err := h.mintingBurningService.GetMintingHistory(c.Request.Context(), enterpriseID, limit, offset)
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

// GetBurningHistory returns burning history for an enterprise
func (h *MintingBurningHandler) GetBurningHistory(c *gin.Context) {
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

	history, err := h.mintingBurningService.GetBurningHistory(c.Request.Context(), enterpriseID, limit, offset)
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
