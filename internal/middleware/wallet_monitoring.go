package middleware

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// WalletMonitoringMiddleware tracks wallet activities
func WalletMonitoringMiddleware(walletService *services.WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Only track successful operations on wallet endpoints
		if c.Writer.Status() >= 400 {
			return
		}

		// Check if this is a wallet-related endpoint
		path := c.Request.URL.Path
		if !strings.Contains(path, "/wallets") {
			return
		}

		// Extract wallet ID from path or headers
		var walletID uuid.UUID
		var err error

		// Try to get wallet ID from path parameter
		if walletIDStr := c.Param("id"); walletIDStr != "" {
			walletID, err = uuid.Parse(walletIDStr)
			if err == nil {
				updateWalletActivity(walletService, walletID, c.Request.Method, path)
				return
			}
		}

		// Try to get wallet address from path and lookup wallet
		if address := c.Param("address"); address != "" {
			wallet, err := walletService.GetWalletByAddress(address)
			if err == nil {
				updateWalletActivity(walletService, wallet.ID, c.Request.Method, path)
				return
			}
		}

		// For enterprise wallet endpoints, we don't update activity since
		// it's a read operation on multiple wallets
	}
}

func updateWalletActivity(walletService *services.WalletService, walletID uuid.UUID, method, path string) {
	if err := walletService.UpdateWalletActivity(walletID); err != nil {
		log.Printf("Failed to update wallet activity for wallet %s: %v", walletID, err)
	} else {
		log.Printf("Updated activity for wallet %s (method: %s, path: %s)", walletID, method, path)
	}
}
