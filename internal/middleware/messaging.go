package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// MessagingMiddleware adds messaging service to the Gin context
func MessagingMiddleware(messagingService *messaging.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("messaging", messagingService)
		c.Next()
	}
}

// GetService retrieves the messaging service from Gin context
func GetService(c *gin.Context) (*messaging.Service, bool) {
	service, exists := c.Get("messaging")
	if !exists {
		return nil, false
	}

	messagingService, ok := service.(*messaging.Service)
	return messagingService, ok
}

// MessagingHealthCheck provides a health check endpoint for messaging
func MessagingHealthCheck(c *gin.Context) {
	messagingService, exists := GetService(c)
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
