package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/config"
	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize messaging service
	messagingService, err := messaging.NewService(
		cfg.Redis.URL,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize messaging service: %v", err)
	}
	defer messagingService.Close()

	r := gin.New()

	// Add global error handling middleware first
	r.Use(middleware.ErrorHandler())

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add messaging middleware
	r.Use(middleware.MessagingMiddleware(messagingService))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "api-gateway",
			"status":  "healthy",
		})
	})

	// Messaging health check
	r.GET("/health/messaging", middleware.MessagingHealthCheck)

	// Messaging monitoring endpoints
	r.GET("/admin/queue-stats", handlers.GetQueueStats)
	r.POST("/admin/test-event", handlers.PublishTestEvent)

	log.Println("API Gateway starting on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
