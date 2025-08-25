package messaging

import (
	"context"
)

// Publisher interface for publishing messages
type Publisher interface {
	Publish(channel string, message *Message) error
	EnqueueWithRetry(queue string, message *Message, maxRetries int) error
	HealthCheck() error
	Close() error
}

// Subscriber interface for subscribing to messages
type Subscriber interface {
	Subscribe(channel string, handler func(*Message) error) error
	DequeueWithRetry(queue string, handler func(*Message) error, maxRetries int) error
	HealthCheck() error
	Close() error
}

// MessageQueue combines both publisher and subscriber interfaces
type MessageQueue interface {
	Publisher
	Subscriber
	GetQueueLength(queue string) (int64, error)
}

// MessageHandler is a function type for handling messages
type MessageHandler func(*Message) error

// Event represents a domain event in the system
type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

// EventBus interface for event-driven communication
type EventBus interface {
	PublishEvent(ctx context.Context, event *Event) error
	SubscribeToEvent(ctx context.Context, eventType string, handler func(*Event) error) error
	HealthCheck() error
	Close() error
}
