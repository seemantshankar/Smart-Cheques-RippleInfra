package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// WalletHandler handles wallet-related HTTP requests
type WalletHandler struct {
	walletService     *services.WalletService
	monitoringService *services.WalletMonitoringService
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletService *services.WalletService, monitoringService *services.WalletMonitoringService) *WalletHandler {
	return &WalletHandler{
		walletService:     walletService,
		monitoringService: monitoringService,
	}
}

// CreateWallet handles wallet creation for an enterprise
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req models.WalletCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	wallet, err := h.walletService.CreateWalletForEnterprise(req.EnterpriseID, req.NetworkType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create wallet",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Wallet created successfully",
		"wallet":  wallet,
	})
}

// GetWallet handles retrieving a wallet by ID
func (h *WalletHandler) GetWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	wallet, err := h.walletService.GetWalletByID(walletID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Wallet not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet": wallet,
	})
}

// GetWalletByAddress handles retrieving a wallet by address
func (h *WalletHandler) GetWalletByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	wallet, err := h.walletService.GetWalletByAddress(address)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Wallet not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet": wallet,
	})
}

// GetEnterpriseWallets handles retrieving all wallets for an enterprise
func (h *WalletHandler) GetEnterpriseWallets(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	wallets, err := h.walletService.GetWalletsForEnterprise(enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get wallets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallets": wallets,
	})
}

// ActivateWallet handles wallet activation
func (h *WalletHandler) ActivateWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	if err := h.walletService.ActivateWallet(walletID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate wallet",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet activated successfully",
	})
}

// WhitelistWallet handles wallet whitelisting
func (h *WalletHandler) WhitelistWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	if err := h.walletService.WhitelistWallet(walletID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to whitelist wallet",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet whitelisted successfully",
	})
}

// SuspendWallet handles wallet suspension
func (h *WalletHandler) SuspendWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	if err := h.walletService.SuspendWallet(walletID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to suspend wallet",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet suspended successfully",
	})
}

// GetWhitelistedWallets handles retrieving all whitelisted wallets
func (h *WalletHandler) GetWhitelistedWallets(c *gin.Context) {
	wallets, err := h.walletService.GetWhitelistedWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get whitelisted wallets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallets": wallets,
	})
}

// ValidateAddress handles XRPL address validation
func (h *WalletHandler) ValidateAddress(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	isValid := h.walletService.ValidateWalletAddress(req.Address)

	c.JSON(http.StatusOK, gin.H{
		"address":  req.Address,
		"is_valid": isValid,
	})
}

// UpdateWallet handles wallet updates
func (h *WalletHandler) UpdateWallet(c *gin.Context) {
	walletIDStr := c.Param("id")
	_, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	var req models.WalletUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// For now, we'll handle specific update operations through dedicated endpoints
	// This is a placeholder for future generic update functionality
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Generic wallet updates not implemented. Use specific endpoints like /activate, /whitelist, /suspend",
	})
}

// GetWalletHealthStatus handles wallet health status monitoring
func (h *WalletHandler) GetWalletHealthStatus(c *gin.Context) {
	status, err := h.monitoringService.GetWalletHealthStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get wallet health status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"health_status": status,
	})
}

// PerformWalletHealthCheck handles individual wallet health checks
func (h *WalletHandler) PerformWalletHealthCheck(c *gin.Context) {
	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet ID format",
		})
		return
	}

	if err := h.monitoringService.PerformWalletHealthCheck(walletID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Wallet health check failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Wallet health check passed",
		"wallet_id": walletID,
	})
}

// GetInactiveWallets handles monitoring of inactive wallets
func (h *WalletHandler) GetInactiveWallets(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid days parameter. Must be a positive integer",
		})
		return
	}

	inactiveWallets, err := h.monitoringService.MonitorInactiveWallets(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get inactive wallets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inactive_wallets": inactiveWallets,
		"threshold_days":   days,
		"count":            len(inactiveWallets),
	})
}

// GetWalletMetrics handles wallet metrics endpoint
func (h *WalletHandler) GetWalletMetrics(c *gin.Context) {
	metrics, err := h.monitoringService.GetWalletMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get wallet metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}
