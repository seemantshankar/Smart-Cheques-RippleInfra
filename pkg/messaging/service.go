package messaging

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Service provides a high-level interface for messaging operations
type Service struct {
	redisClient *RedisClient
	eventBus    *RedisEventBus
	subscribers map[string]context.CancelFunc
	mu          sync.RWMutex
}

// NewService creates a new messaging service
func NewService(redisAddr, redisPassword string, redisDB int) (*Service, error) {
	redisClient, err := NewRedisClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	eventBus := NewRedisEventBus(redisClient)

	return &Service{
		redisClient: redisClient,
		eventBus:    eventBus,
		subscribers: make(map[string]context.CancelFunc),
	}, nil
}

// PublishEvent publishes an event to the event bus
func (s *Service) PublishEvent(event *Event) error {
	// If service is nil or eventBus is nil, skip publishing
	if s == nil || s.eventBus == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.eventBus.PublishEvent(ctx, event)
}

// SubscribeToEvent subscribes to events of a specific type
func (s *Service) SubscribeToEvent(eventType string, handler func(*Event) error) error {
	// If service is nil or eventBus is nil, skip subscribing
	if s == nil || s.eventBus == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.subscribers[eventType] = cancel
	s.mu.Unlock()

	go func() {
		if err := s.eventBus.SubscribeToEvent(ctx, eventType, handler); err != nil {
			log.Printf("Error subscribing to event type %s: %v", eventType, err)
		}
	}()

	return nil
}

// UnsubscribeFromEvent unsubscribes from events of a specific type
func (s *Service) UnsubscribeFromEvent(eventType string) {
	// If service is nil, skip unsubscribing
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.subscribers[eventType]; exists {
		cancel()
		delete(s.subscribers, eventType)
		log.Printf("Unsubscribed from event type: %s", eventType)
	}
}

// SendMessage sends a message to a specific queue with retry logic
func (s *Service) SendMessage(queue string, messageType string, payload map[string]interface{}) error {
	// If service is nil or redisClient is nil, skip sending
	if s == nil || s.redisClient == nil {
		return nil
	}

	message := &Message{
		ID:        generateMessageID(),
		Type:      messageType,
		Payload:   payload,
		Timestamp: time.Now(),
		Retries:   0,
	}

	return s.redisClient.EnqueueWithRetry(queue, message, 3) // Default 3 retries
}

// ProcessMessages processes messages from a queue with retry logic
func (s *Service) ProcessMessages(queue string, handler func(*Message) error) error {
	// If service is nil or redisClient is nil, skip processing
	if s == nil || s.redisClient == nil {
		return nil
	}

	return s.redisClient.DequeueWithRetry(queue, handler, 3) // Default 3 retries
}

// HealthCheck checks the health of the messaging service
func (s *Service) HealthCheck() error {
	// If service is nil or redisClient is nil, return nil (healthy)
	if s == nil || s.redisClient == nil {
		return nil
	}

	return s.redisClient.HealthCheck()
}

// GetQueueStats returns statistics about a queue
func (s *Service) GetQueueStats(queue string) (QueueStats, error) {
	// If service is nil or redisClient is nil, return empty stats
	if s == nil || s.redisClient == nil {
		return QueueStats{}, nil
	}

	length, err := s.redisClient.GetQueueLength(queue)
	if err != nil {
		return QueueStats{}, err
	}

	dlqLength, err := s.redisClient.GetQueueLength(queue + "_dlq")
	if err != nil {
		dlqLength = 0 // DLQ might not exist yet
	}

	return QueueStats{
		QueueName:       queue,
		MessageCount:    length,
		DeadLetterCount: dlqLength,
		LastChecked:     time.Now(),
	}, nil
}

// Close closes the messaging service and all subscriptions
func (s *Service) Close() error {
	// If service is nil, nothing to close
	if s == nil {
		return nil
	}

	// Cancel all subscriptions
	s.mu.Lock()
	for eventType, cancel := range s.subscribers {
		cancel()
		log.Printf("Canceled subscription for event type: %s", eventType)
	}
	s.subscribers = make(map[string]context.CancelFunc)
	s.mu.Unlock()

	// Close Redis client if it exists
	if s.redisClient != nil {
		return s.redisClient.Close()
	}

	return nil
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
	QueueEnterpriseEvents  = "enterprise_events"
	QueueSmartChequeEvents = "smart_check_events"
	QueueMilestoneEvents   = "milestone_events"
	QueuePaymentEvents     = "payment_events"
	QueueXRPLTransactions  = "xrpl_transactions"
	QueueNotifications     = "notifications"
	QueueAuditLogs         = "audit_logs"
)
