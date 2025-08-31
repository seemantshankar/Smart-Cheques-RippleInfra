package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/config"
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

	// Subscribe to relevant events
	err = messagingService.SubscribeToEvent(messaging.EventTypeEnterpriseRegistered, handleEnterpriseRegistered)
	if err != nil {
		log.Printf("Failed to subscribe to enterprise registered events: %v", err)
	}

	err = messagingService.SubscribeToEvent(messaging.EventTypeMilestoneCompleted, handleMilestoneCompleted)
	if err != nil {
		log.Printf("Failed to subscribe to milestone completed events: %v", err)
	}

	r := gin.Default()

	// Add messaging middleware
	r.Use(middleware.MessagingMiddleware(messagingService))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "orchestration-service",
			"status":  "healthy",
		})
	})

	// Messaging health check
	r.GET("/health/messaging", middleware.MessagingHealthCheck)

	// Smart Check creation endpoint (example)
	r.POST("/smart-checks", createSmartCheque)

	// Milestone completion endpoint (example)
	r.POST("/milestones/:id/complete", completeMilestone)

	log.Println("Orchestration Service starting on :8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}

func handleEnterpriseRegistered(event *messaging.Event) error {
	log.Printf("Orchestration service handling enterprise registered event: %+v", event)
	// TODO: Initialize enterprise-specific orchestration workflows
	return nil
}

func handleMilestoneCompleted(event *messaging.Event) error {
	log.Printf("Orchestration service handling milestone completed event: %+v", event)
	// TODO: Trigger payment release workflow
	return nil
}

func createSmartCheque(c *gin.Context) {
	messagingService, exists := middleware.GetService(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "messaging service not available",
		})
		return
	}

	var request struct {
		PayerID  string  `json:"payer_id" binding:"required"`
		PayeeID  string  `json:"payee_id" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
		Currency string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Implement actual Smart Check creation logic
	chequeID := "sc_" + request.PayerID + "_" + request.PayeeID // Simplified for demo

	// Publish Smart Check created event
	event := messaging.NewSmartChequeCreatedEvent(
		chequeID,
		request.PayerID,
		request.PayeeID,
		request.Amount,
		request.Currency,
	)
	if err := messagingService.PublishEvent(event); err != nil {
		log.Printf("Failed to publish Smart Check created event: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"check_id": chequeID,
		"payer_id": request.PayerID,
		"payee_id": request.PayeeID,
		"amount":   request.Amount,
		"currency": request.Currency,
		"status":   "created",
	})
}

func completeMilestone(c *gin.Context) {
	messagingService, exists := middleware.GetService(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "messaging service not available",
		})
		return
	}

	milestoneID := c.Param("id")

	var request struct {
		SmartChequeID string  `json:"smart_check_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Implement actual milestone completion logic

	// Publish milestone completed event
	event := messaging.NewMilestoneCompletedEvent(
		milestoneID,
		request.SmartChequeID,
		request.Amount,
	)
	if err := messagingService.PublishEvent(event); err != nil {
		log.Printf("Failed to publish milestone completed event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"milestone_id":   milestoneID,
		"smart_check_id": request.SmartChequeID,
		"amount":         request.Amount,
		"status":         "completed",
	})
}
