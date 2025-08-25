package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// MessagingMiddleware adds messaging service to the Gin context
func MessagingMiddleware(messagingService *messaging.MessagingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("messaging", messagingService)
		c.Next()
	}
}

// GetMessagingService retrieves the messaging service from Gin context
func GetMessagingService(c *gin.Context) (*messaging.MessagingService, bool) {
	service, exists := c.Get("messaging")
	if !exists {
		return nil, false
	}

	messagingService, ok := service.(*messaging.MessagingService)
	return messagingService, ok
}

// MessagingHealthCheck provides a health check endpoint for messaging
func MessagingHealthCheck(c *gin.Context) {
	messagingService, exists := GetMessagingService(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "messaging service not available",
		})
		return
	}

	if err := messagingService.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "messaging",
	})
}
