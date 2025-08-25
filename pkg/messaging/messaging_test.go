package messaging

import (
	"testing"
	"time"
)

func TestNewRedisClient(t *testing.T) {
	// This test would require a running Redis instance
	// For now, we'll just test that the function exists and can be called
	
	// Test with invalid connection string
	_, err := NewRedisClient("invalid-host:6379", "", 0)
	if err == nil {
		t.Error("Expected error for invalid Redis connection")
	}
}

func TestMessage(t *testing.T) {
	// Test message creation
	message := &Message{
		ID:        "test-123",
		Type:      "test-event",
		Payload:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
		Retries:   0,
	}

	if message.ID != "test-123" {
		t.Errorf("Expected message ID 'test-123', got '%s'", message.ID)
	}

	if message.Type != "test-event" {
		t.Errorf("Expected message type 'test-event', got '%s'", message.Type)
	}
}

func TestEventCreation(t *testing.T) {
	// Test event helper functions
	event := NewEnterpriseRegisteredEvent("ent-123", "Test Corp")
	
	if event.Type != EventTypeEnterpriseRegistered {
		t.Errorf("Expected event type '%s', got '%s'", EventTypeEnterpriseRegistered, event.Type)
	}

	if event.Source != "identity-service" {
		t.Errorf("Expected event source 'identity-service', got '%s'", event.Source)
	}

	enterpriseID, ok := event.Data["enterprise_id"].(string)
	if !ok || enterpriseID != "ent-123" {
		t.Errorf("Expected enterprise_id 'ent-123', got '%v'", enterpriseID)
	}
}

func TestSmartChequeEventCreation(t *testing.T) {
	event := NewSmartChequeCreatedEvent("sc-123", "payer-1", "payee-1", 1000.0, "USDT")
	
	if event.Type != EventTypeSmartChequeCreated {
		t.Errorf("Expected event type '%s', got '%s'", EventTypeSmartChequeCreated, event.Type)
	}

	amount, ok := event.Data["amount"].(float64)
	if !ok || amount != 1000.0 {
		t.Errorf("Expected amount 1000.0, got '%v'", amount)
	}

	currency, ok := event.Data["currency"].(string)
	if !ok || currency != "USDT" {
		t.Errorf("Expected currency 'USDT', got '%v'", currency)
	}
}