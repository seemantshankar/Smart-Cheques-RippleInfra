package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RedisEventBus struct {
	client *RedisClient
}

func NewRedisEventBus(redisClient *RedisClient) *RedisEventBus {
	return &RedisEventBus{
		client: redisClient,
	}
}

func (e *RedisEventBus) PublishEvent(ctx context.Context, event *Event) error {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	message := &Message{
		ID:        uuid.New().String(),
		Type:      "event",
		Payload:   map[string]interface{}{"event": event},
		Timestamp: time.Now(),
		Retries:   0,
	}

	channel := fmt.Sprintf("events.%s", event.Type)
	return e.client.Publish(channel, message)
}

func (e *RedisEventBus) SubscribeToEvent(ctx context.Context, eventType string, handler func(*Event) error) error {
	channel := fmt.Sprintf("events.%s", eventType)

	messageHandler := func(message *Message) error {
		// Extract event from message payload
		eventData, ok := message.Payload["event"]
		if !ok {
			return fmt.Errorf("no event data in message payload")
		}

		// Convert to Event struct
		eventBytes, err := json.Marshal(eventData)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		var event Event
		if err := json.Unmarshal(eventBytes, &event); err != nil {
			return fmt.Errorf("failed to unmarshal event: %w", err)
		}

		return handler(&event)
	}

	return e.client.Subscribe(channel, messageHandler)
}

func (e *RedisEventBus) HealthCheck() error {
	return e.client.HealthCheck()
}

func (e *RedisEventBus) Close() error {
	return e.client.Close()
}

// Common event types for the Smart Payment Infrastructure
const (
	EventTypeEnterpriseRegistered     = "enterprise.registered"
	EventTypeEnterpriseKYBUpdated     = "enterprise.kyb_updated"
	EventTypeSmartChequeCreated       = "smart_cheque.created"
	EventTypeSmartChequeLocked        = "smart_cheque.locked"
	EventTypeMilestoneCompleted       = "milestone.completed"
	EventTypeMilestoneVerified        = "milestone.verified"
	EventTypePaymentReleased          = "payment.released"
	EventTypeDisputeCreated           = "dispute.created"
	EventTypeDisputeResolved          = "dispute.resolved"
	EventTypeXRPLTransactionCreated   = "xrpl.transaction.created"
	EventTypeXRPLTransactionConfirmed = "xrpl.transaction.confirmed"
)

// Helper functions to create common events
func NewEnterpriseRegisteredEvent(enterpriseID, legalName string) *Event {
	return &Event{
		Type:   EventTypeEnterpriseRegistered,
		Source: "identity-service",
		Data: map[string]interface{}{
			"enterprise_id": enterpriseID,
			"legal_name":    legalName,
		},
	}
}

func NewSmartChequeCreatedEvent(chequeID, payerID, payeeID string, amount float64, currency string) *Event {
	return &Event{
		Type:   EventTypeSmartChequeCreated,
		Source: "orchestration-service",
		Data: map[string]interface{}{
			"cheque_id": chequeID,
			"payer_id":  payerID,
			"payee_id":  payeeID,
			"amount":    amount,
			"currency":  currency,
		},
	}
}

func NewMilestoneCompletedEvent(milestoneID, smartChequeID string, amount float64) *Event {
	return &Event{
		Type:   EventTypeMilestoneCompleted,
		Source: "orchestration-service",
		Data: map[string]interface{}{
			"milestone_id":    milestoneID,
			"smart_cheque_id": smartChequeID,
			"amount":          amount,
		},
	}
}

func NewXRPLTransactionCreatedEvent(transactionHash, escrowAddress string, amount float64) *Event {
	return &Event{
		Type:   EventTypeXRPLTransactionCreated,
		Source: "xrpl-service",
		Data: map[string]interface{}{
			"transaction_hash": transactionHash,
			"escrow_address":   escrowAddress,
			"amount":           amount,
		},
	}
}
