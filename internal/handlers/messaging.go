package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// GetQueueStats returns statistics for all queues
func GetQueueStats(c *gin.Context) {
	messagingService, exists := middleware.GetMessagingService(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "messaging service not available",
		})
		return
	}

	// Get stats for common queues
	queues := []string{
		messaging.QueueEnterpriseEvents,
		messaging.QueueSmartChequeEvents,
		messaging.QueueMilestoneEvents,
		messaging.QueuePaymentEvents,
		messaging.QueueXRPLTransactions,
		messaging.QueueNotifications,
		messaging.QueueAuditLogs,
	}

	stats := make(map[string]messaging.QueueStats)
	for _, queue := range queues {
		queueStats, err := messagingService.GetQueueStats(queue)
		if err != nil {
			// Queue might not exist yet, that's okay
			queueStats = messaging.QueueStats{
				QueueName:       queue,
				MessageCount:    0,
				DeadLetterCount: 0,
			}
		}
		stats[queue] = queueStats
	}

	c.JSON(http.StatusOK, gin.H{
		"queue_stats": stats,
		"status":      "healthy",
	})
}

// PublishTestEvent publishes a test event for debugging
func PublishTestEvent(c *gin.Context) {
	messagingService, exists := middleware.GetMessagingService(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "messaging service not available",
		})
		return
	}

	var request struct {
		EventType string                 `json:"event_type" binding:"required"`
		Data      map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	event := &messaging.Event{
		Type:   request.EventType,
		Source: "test-endpoint",
		Data:   request.Data,
	}

	if err := messagingService.PublishEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to publish event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "event published successfully",
		"event_type": request.EventType,
		"source":     "test-endpoint",
	})
}
