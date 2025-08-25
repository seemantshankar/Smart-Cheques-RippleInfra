package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MessagingService provides a high-level interface for messaging operations
type MessagingService struct {
	redisClient *RedisClient
	eventBus    *RedisEventBus
	subscribers map[string]context.CancelFunc
	mu          sync.RWMutex
}

// NewMessagingService creates a new messaging service
func NewMessagingService(redisAddr, redisPassword string, redisDB int) (*MessagingService, error) {
	redisClient, err := NewRedisClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	eventBus := NewRedisEventBus(redisClient)

	return &MessagingService{
		redisClient: redisClient,
		eventBus:    eventBus,
		subscribers: make(map[string]context.CancelFunc),
	}, nil
}

// PublishEvent publishes an event to the event bus
func (m *MessagingService) PublishEvent(event *Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return m.eventBus.PublishEvent(ctx, event)
}

// SubscribeToEvent subscribes to events of a specific type
func (m *MessagingService) SubscribeToEvent(eventType string, handler func(*Event) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	
	m.mu.Lock()
	m.subscribers[eventType] = cancel
	m.mu.Unlock()

	go func() {
		if err := m.eventBus.SubscribeToEvent(ctx, eventType, handler); err != nil {
			log.Printf("Error subscribing to event type %s: %v", eventType, err)
		}
	}()

	return nil
}

// UnsubscribeFromEvent unsubscribes from events of a specific type
func (m *MessagingService) UnsubscribeFromEvent(eventType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if cancel, exists := m.subscribers[eventType]; exists {
		cancel()
		delete(m.subscribers, eventType)
		log.Printf("Unsubscribed from event type: %s", eventType)
	}
}

// SendMessage sends a message to a specific queue with retry logic
func (m *MessagingService) SendMessage(queue string, messageType string, payload map[string]interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Type:      messageType,
		Payload:   payload,
		Timestamp: time.Now(),
		Retries:   0,
	}

	return m.redisClient.EnqueueWithRetry(queue, message, 3) // Default 3 retries
}

// ProcessMessages processes messages from a queue with retry logic
func (m *MessagingService) ProcessMessages(queue string, handler func(*Message) error) error {
	return m.redisClient.DequeueWithRetry(queue, handler, 3) // Default 3 retries
}

// HealthCheck checks the health of the messaging service
func (m *MessagingService) HealthCheck() error {
	return m.redisClient.HealthCheck()
}

// GetQueueStats returns statistics about a queue
func (m *MessagingService) GetQueueStats(queue string) (QueueStats, error) {
	length, err := m.redisClient.GetQueueLength(queue)
	if err != nil {
		return QueueStats{}, err
	}

	dlqLength, err := m.redisClient.GetQueueLength(queue + "_dlq")
	if err != nil {
		dlqLength = 0 // DLQ might not exist yet
	}

	return QueueStats{
		QueueName:         queue,
		MessageCount:      length,
		DeadLetterCount:   dlqLength,
		LastChecked:       time.Now(),
	}, nil
}

// Close closes the messaging service and all subscriptions
func (m *MessagingService) Close() error {
	// Cancel all subscriptions
	m.mu.Lock()
	for eventType, cancel := range m.subscribers {
		cancel()
		log.Printf("Cancelled subscription for event type: %s", eventType)
	}
	m.subscribers = make(map[string]context.CancelFunc)
	m.mu.Unlock()

	// Close Redis client
	return m.redisClient.Close()
}

// QueueStats represents statistics about a message queue
type QueueStats struct {
	QueueName       string    `json:"queue_name"`
	MessageCount    int64     `json:"message_count"`
	DeadLetterCount int64     `json:"dead_letter_count"`
	LastChecked     time.Time `json:"last_checked"`
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// Common queue names for the Smart Payment Infrastructure
const (
	QueueEnterpriseEvents    = "enterprise_events"
	QueueSmartChequeEvents   = "smart_cheque_events"
	QueueMilestoneEvents     = "milestone_events"
	QueuePaymentEvents       = "payment_events"
	QueueXRPLTransactions    = "xrpl_transactions"
	QueueNotifications       = "notifications"
	QueueAuditLogs          = "audit_logs"
)