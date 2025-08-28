package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// AssetHandler handles HTTP requests for asset operations
type AssetHandler struct {
	assetService   *services.AssetService
	balanceService *services.BalanceService
}

// NewAssetHandler creates a new asset handler
func NewAssetHandler(assetService *services.AssetService, balanceService *services.BalanceService) *AssetHandler {
	return &AssetHandler{
		assetService:   assetService,
		balanceService: balanceService,
	}
}

// DepositRequest represents a request to deposit funds
type DepositRequest struct {
	EnterpriseID   uuid.UUID `json:"enterprise_id" validate:"required"`
	CurrencyCode   string    `json:"currency_code" validate:"required"`
	Amount         string    `json:"amount" validate:"required"`
	ExternalTxHash *string   `json:"external_tx_hash,omitempty"`
	Description    *string   `json:"description,omitempty"`
	Source         string    `json:"source" validate:"required"`
}

// WithdrawalRequest represents a request to withdraw funds
type WithdrawalRequest struct {
	EnterpriseID    uuid.UUID `json:"enterprise_id" validate:"required"`
	CurrencyCode    string    `json:"currency_code" validate:"required"`
	Amount          string    `json:"amount" validate:"required"`
	DestinationAddr string    `json:"destination_address" validate:"required"`
	Description     *string   `json:"description,omitempty"`
}

// DepositResponse represents the response for a deposit request
type DepositResponse struct {
	TransactionID uuid.UUID                 `json:"transaction_id"`
	Status        string                    `json:"status"`
	Message       string                    `json:"message"`
	Balance       *models.EnterpriseBalance `json:"balance,omitempty"`
}

// WithdrawalResponse represents the response for a withdrawal request
type WithdrawalResponse struct {
	TransactionID uuid.UUID                 `json:"transaction_id"`
	Status        string                    `json:"status"`
	Message       string                    `json:"message"`
	Balance       *models.EnterpriseBalance `json:"balance,omitempty"`
}

// ProcessDeposit handles deposit requests
func (h *AssetHandler) ProcessDeposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate enterprise exists (this would typically check against enterprise service)
	if req.EnterpriseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	// Validate currency is supported
	asset, err := h.assetService.GetAssetByCurrency(c.Request.Context(), req.CurrencyCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Unsupported currency",
			"details": err.Error(),
		})
		return
	}

	if !asset.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Currency is currently deactivated",
		})
		return
	}

	// Create balance operation request
	balanceReq := &services.BalanceOperationRequest{
		EnterpriseID:  req.EnterpriseID,
		CurrencyCode:  req.CurrencyCode,
		Amount:        req.Amount,
		OperationType: models.AssetTransactionTypeDeposit,
		Description:   req.Description,
		Metadata: map[string]interface{}{
			"source":           req.Source,
			"external_tx_hash": req.ExternalTxHash,
		},
	}

	// Process the deposit
	transaction, err := h.balanceService.ProcessBalanceOperation(c.Request.Context(), balanceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to process deposit",
			"details": err.Error(),
		})
		return
	}

	// Get updated balance
	balance, err := h.balanceService.GetEnterpriseBalance(c.Request.Context(), req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		// Don't fail the response if we can't get the balance, just log it
		balance = nil
	}

	response := DepositResponse{
		TransactionID: transaction.ID,
		Status:        "success",
		Message:       "Deposit processed successfully",
		Balance:       balance,
	}

	c.JSON(http.StatusOK, response)
}

// ProcessWithdrawal handles withdrawal requests
func (h *AssetHandler) ProcessWithdrawal(c *gin.Context) {
	var req WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate enterprise exists
	if req.EnterpriseID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	// Validate currency is supported
	asset, err := h.assetService.GetAssetByCurrency(c.Request.Context(), req.CurrencyCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Unsupported currency",
			"details": err.Error(),
		})
		return
	}

	if !asset.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Currency is currently deactivated",
		})
		return
	}

	// Check balance sufficiency
	sufficient, err := h.balanceService.CheckBalanceSufficiency(c.Request.Context(), req.EnterpriseID, req.CurrencyCode, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to check balance",
			"details": err.Error(),
		})
		return
	}

	if !sufficient {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Insufficient balance for withdrawal",
		})
		return
	}

	// Create balance operation request
	balanceReq := &services.BalanceOperationRequest{
		EnterpriseID:  req.EnterpriseID,
		CurrencyCode:  req.CurrencyCode,
		Amount:        req.Amount,
		OperationType: models.AssetTransactionTypeWithdrawal,
		Description:   req.Description,
		Metadata: map[string]interface{}{
			"destination_address": req.DestinationAddr,
		},
	}

	// Process the withdrawal
	transaction, err := h.balanceService.ProcessBalanceOperation(c.Request.Context(), balanceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to process withdrawal",
			"details": err.Error(),
		})
		return
	}

	// Get updated balance
	balance, err := h.balanceService.GetEnterpriseBalance(c.Request.Context(), req.EnterpriseID, req.CurrencyCode)
	if err != nil {
		balance = nil
	}

	response := WithdrawalResponse{
		TransactionID: transaction.ID,
		Status:        "success",
		Message:       "Withdrawal processed successfully",
		Balance:       balance,
	}

	c.JSON(http.StatusOK, response)
}

// GetBalance retrieves balance for an enterprise and currency
func (h *AssetHandler) GetBalance(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	currencyCode := c.Param("currencyCode")

	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	balance, err := h.balanceService.GetEnterpriseBalance(c.Request.Context(), enterpriseID, currencyCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Balance not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetAllBalances retrieves all balances for an enterprise
func (h *AssetHandler) GetAllBalances(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")

	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	balances, err := h.balanceService.GetEnterpriseBalances(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get balances",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"balances":      balances,
	})
}

// GetBalanceSummary retrieves balance summary for an enterprise
func (h *AssetHandler) GetBalanceSummary(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")

	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	summary, err := h.balanceService.GetBalanceSummary(c.Request.Context(), enterpriseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get balance summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"summary":       summary,
	})
}

// ReserveBalance reserves funds for a specific purpose
func (h *AssetHandler) ReserveBalance(c *gin.Context) {
	var req struct {
		EnterpriseID uuid.UUID `json:"enterprise_id" validate:"required"`
		CurrencyCode string    `json:"currency_code" validate:"required"`
		Amount       string    `json:"amount" validate:"required"`
		ReferenceID  *string   `json:"reference_id,omitempty"`
		Reason       *string   `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := h.balanceService.ReserveBalance(c.Request.Context(), req.EnterpriseID, req.CurrencyCode, req.Amount, req.ReferenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to reserve balance",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Balance reserved successfully",
	})
}

// ReleaseReservedBalance releases previously reserved funds
func (h *AssetHandler) ReleaseReservedBalance(c *gin.Context) {
	var req struct {
		EnterpriseID uuid.UUID `json:"enterprise_id" validate:"required"`
		CurrencyCode string    `json:"currency_code" validate:"required"`
		Amount       string    `json:"amount" validate:"required"`
		ReferenceID  *string   `json:"reference_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := h.balanceService.ReleaseReservedBalance(c.Request.Context(), req.EnterpriseID, req.CurrencyCode, req.Amount, req.ReferenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to release reserved balance",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Reserved balance released successfully",
	})
}

// GetSupportedAssets returns all supported assets
func (h *AssetHandler) GetSupportedAssets(c *gin.Context) {
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	assets, err := h.assetService.GetSupportedAssets(c.Request.Context(), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get supported assets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assets": assets,
	})
}

// GetAssetConfiguration returns configuration for a specific asset
func (h *AssetHandler) GetAssetConfiguration(c *gin.Context) {
	currencyCode := c.Param("currencyCode")

	asset, err := h.assetService.GetAssetByCurrency(c.Request.Context(), currencyCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Asset not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// GetTransactionHistory returns transaction history for an enterprise
func (h *AssetHandler) GetTransactionHistory(c *gin.Context) {
	enterpriseIDStr := c.Param("enterpriseId")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID format",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// This would need to be implemented in the balance service
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"enterprise_id": enterpriseID,
		"transactions":  []interface{}{},
		"limit":         limit,
		"offset":        offset,
		"message":       "Transaction history endpoint - implementation pending",
	})
}

// RegisterRoutes registers all asset-related routes
func (h *AssetHandler) RegisterRoutes(router *gin.Engine) {
	assetGroup := router.Group("/api/v1/assets")
	{
		// Asset information endpoints
		assetGroup.GET("/supported", h.GetSupportedAssets)
		assetGroup.GET("/config/:currencyCode", h.GetAssetConfiguration)

		// Balance endpoints
		assetGroup.GET("/balance/:enterpriseId/:currencyCode", h.GetBalance)
		assetGroup.GET("/balances/:enterpriseId", h.GetAllBalances)
		assetGroup.GET("/summary/:enterpriseId", h.GetBalanceSummary)

		// Transaction endpoints
		assetGroup.POST("/deposit", h.ProcessDeposit)
		assetGroup.POST("/withdraw", h.ProcessWithdrawal)
		assetGroup.POST("/reserve", h.ReserveBalance)
		assetGroup.POST("/release", h.ReleaseReservedBalance)

		// Transaction history
		assetGroup.GET("/transactions/:enterpriseId", h.GetTransactionHistory)
	}
}
