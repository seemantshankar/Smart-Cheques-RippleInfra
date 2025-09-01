package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// DisputeWebhookService handles webhook notifications for dispute events
type DisputeWebhookService struct {
	httpClient *http.Client
	baseURL    string
	secretKey  string
	webhooks   map[string]*WebhookConfig
}

// WebhookConfig represents a webhook endpoint configuration
type WebhookConfig struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Secret      string            `json:"secret"`
	Events      []string          `json:"events"`  // Events this webhook should receive
	Headers     map[string]string `json:"headers"` // Custom headers to include
	IsActive    bool              `json:"is_active"`
	RetryPolicy RetryPolicy       `json:"retry_policy"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RetryPolicy defines webhook retry behavior
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	MaxDelay      time.Duration `json:"max_delay"`
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	DisputeID string                 `json:"dispute_id"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID            string         `json:"id"`
	WebhookID     string         `json:"webhook_id"`
	EventType     string         `json:"event_type"`
	Payload       WebhookPayload `json:"payload"`
	URL           string         `json:"url"`
	StatusCode    int            `json:"status_code"`
	ResponseBody  string         `json:"response_body"`
	ErrorMessage  *string        `json:"error_message,omitempty"`
	AttemptNumber int            `json:"attempt_number"`
	SentAt        time.Time      `json:"sent_at"`
	Duration      time.Duration  `json:"duration"`
	Success       bool           `json:"success"`
}

// NewDisputeWebhookService creates a new webhook service instance
func NewDisputeWebhookService(baseURL, secretKey string) *DisputeWebhookService {
	return &DisputeWebhookService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   baseURL,
		secretKey: secretKey,
		webhooks:  make(map[string]*WebhookConfig),
	}
}

// RegisterWebhook registers a new webhook endpoint
func (s *DisputeWebhookService) RegisterWebhook(config *WebhookConfig) error {
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	config.UpdatedAt = time.Now()

	s.webhooks[config.ID] = config
	return nil
}

// UnregisterWebhook removes a webhook endpoint
func (s *DisputeWebhookService) UnregisterWebhook(webhookID string) {
	delete(s.webhooks, webhookID)
}

// SendWebhookNotification sends a webhook notification for a dispute event
func (s *DisputeWebhookService) SendWebhookNotification(ctx context.Context, dispute *models.Dispute, eventType string, additionalData map[string]interface{}) error {
	// Find active webhooks that subscribe to this event
	activeWebhooks := s.getActiveWebhooksForEvent(eventType)
	if len(activeWebhooks) == 0 {
		return nil // No webhooks to notify
	}

	// Prepare payload
	payload := s.prepareWebhookPayload(dispute, eventType, additionalData)

	// Send to all matching webhooks
	for _, webhook := range activeWebhooks {
		go s.sendWebhookWithRetry(ctx, webhook, payload)
	}

	return nil
}

// SendBulkWebhookNotifications sends notifications to multiple webhooks
func (s *DisputeWebhookService) SendBulkWebhookNotifications(ctx context.Context, dispute *models.Dispute, eventTypes []string, additionalData map[string]interface{}) error {
	for _, eventType := range eventTypes {
		if err := s.SendWebhookNotification(ctx, dispute, eventType, additionalData); err != nil {
			// Log error but continue with other event types
			fmt.Printf("Failed to send webhook for event %s: %v\n", eventType, err)
		}
	}
	return nil
}

// GetWebhookDeliveries retrieves delivery history for a webhook
func (s *DisputeWebhookService) GetWebhookDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookDelivery, error) {
	// This would typically query a database for delivery history
	// For now, return empty slice
	return []*WebhookDelivery{}, nil
}

// ValidateWebhookPayload validates incoming webhook payloads (for webhook endpoints)
func (s *DisputeWebhookService) ValidateWebhookPayload(payload []byte, signature string) (*WebhookPayload, error) {
	// Verify HMAC signature
	expectedSignature := s.generateHMAC(payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	// Parse payload
	var webhookPayload WebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return &webhookPayload, nil
}

// Private methods

func (s *DisputeWebhookService) getActiveWebhooksForEvent(eventType string) []*WebhookConfig {
	var matchingWebhooks []*WebhookConfig

	for _, webhook := range s.webhooks {
		if !webhook.IsActive {
			continue
		}

		// Check if webhook subscribes to this event or all events
		if s.webhookSubscribesToEvent(webhook, eventType) {
			matchingWebhooks = append(matchingWebhooks, webhook)
		}
	}

	return matchingWebhooks
}

func (s *DisputeWebhookService) webhookSubscribesToEvent(webhook *WebhookConfig, eventType string) bool {
	// Check if webhook subscribes to all events
	for _, event := range webhook.Events {
		if event == "*" || event == eventType {
			return true
		}
	}
	return false
}

func (s *DisputeWebhookService) prepareWebhookPayload(dispute *models.Dispute, eventType string, additionalData map[string]interface{}) WebhookPayload {
	data := map[string]interface{}{
		"dispute_id":       dispute.ID,
		"title":            dispute.Title,
		"description":      dispute.Description,
		"category":         dispute.Category,
		"priority":         dispute.Priority,
		"status":           dispute.Status,
		"initiator_id":     dispute.InitiatorID,
		"respondent_id":    dispute.RespondentID,
		"initiated_at":     dispute.InitiatedAt,
		"last_activity_at": dispute.LastActivityAt,
	}

	if dispute.DisputedAmount != nil {
		data["disputed_amount"] = *dispute.DisputedAmount
	}
	if dispute.Currency != nil {
		data["currency"] = *dispute.Currency
	}

	// Add additional data
	for key, value := range additionalData {
		data[key] = value
	}

	return WebhookPayload{
		ID:        uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
		DisputeID: dispute.ID,
		Data:      data,
		Metadata: map[string]interface{}{
			"source":      "dispute-management-service",
			"version":     "1.0",
			"environment": "production", // This would come from config
		},
	}
}

func (s *DisputeWebhookService) sendWebhookWithRetry(ctx context.Context, webhook *WebhookConfig, payload WebhookPayload) {
	policy := webhook.RetryPolicy
	if policy.MaxRetries == 0 {
		policy = RetryPolicy{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			BackoffFactor: 2.0,
			MaxDelay:      30 * time.Second,
		}
	}

	var lastError error
	delay := policy.InitialDelay

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		delivery := s.sendWebhookAttempt(ctx, webhook, payload, attempt+1)

		if delivery.Success {
			s.logWebhookDelivery(delivery)
			return
		}

		lastError = fmt.Errorf("attempt %d failed: %s", attempt+1, *delivery.ErrorMessage)

		// Don't retry on the last attempt
		if attempt < policy.MaxRetries {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * policy.BackoffFactor)
			if delay > policy.MaxDelay {
				delay = policy.MaxDelay
			}
		}
	}

	// Log final failure
	fmt.Printf("Webhook delivery failed after %d attempts: %v\n", policy.MaxRetries+1, lastError)
}

func (s *DisputeWebhookService) sendWebhookAttempt(ctx context.Context, webhook *WebhookConfig, payload WebhookPayload, attemptNumber int) *WebhookDelivery {
	delivery := &WebhookDelivery{
		ID:            uuid.New().String(),
		WebhookID:     webhook.ID,
		EventType:     payload.EventType,
		Payload:       payload,
		URL:           webhook.URL,
		AttemptNumber: attemptNumber,
		SentAt:        time.Now(),
	}

	// Prepare request
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to marshal payload: %v", err)
		delivery.ErrorMessage = &errorMsg
		return delivery
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		errorMsg := fmt.Sprintf("failed to create request: %v", err)
		delivery.ErrorMessage = &errorMsg
		return delivery
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Smart-Payment-Dispute-Service/1.0")
	req.Header.Set("X-Webhook-ID", webhook.ID)
	req.Header.Set("X-Event-Type", payload.EventType)
	req.Header.Set("X-Webhook-Signature", s.generateHMAC(payloadBytes))

	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	start := time.Now()
	resp, err := s.httpClient.Do(req)
	delivery.Duration = time.Since(start)

	if err != nil {
		errorMsg := fmt.Sprintf("request failed: %v", err)
		delivery.ErrorMessage = &errorMsg
		return delivery
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to read response: %v", err)
		delivery.ErrorMessage = &errorMsg
		delivery.StatusCode = resp.StatusCode
		return delivery
	}

	delivery.StatusCode = resp.StatusCode
	delivery.ResponseBody = string(body)

	// Check if delivery was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Success = true
	} else {
		errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		delivery.ErrorMessage = &errorMsg
	}

	return delivery
}

func (s *DisputeWebhookService) generateHMAC(payload []byte) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *DisputeWebhookService) logWebhookDelivery(delivery *WebhookDelivery) {
	// This would typically log to a database or monitoring system
	fmt.Printf("Webhook delivery successful: %s -> %s (HTTP %d, %.2fs)\n",
		delivery.WebhookID, delivery.URL, delivery.StatusCode, delivery.Duration.Seconds())
}

// Default webhook configurations for common dispute events
func (s *DisputeWebhookService) RegisterDefaultWebhooks() error {
	// This would register default webhooks from configuration
	// For now, just log that this method exists
	fmt.Println("RegisterDefaultWebhooks called - implement webhook configuration loading")
	return nil
}

// GetWebhookStats returns statistics about webhook deliveries
func (s *DisputeWebhookService) GetWebhookStats(ctx context.Context, webhookID string, since time.Time) (*WebhookStats, error) {
	// This would query delivery statistics from database
	return &WebhookStats{
		TotalDeliveries:      0,
		SuccessfulDeliveries: 0,
		FailedDeliveries:     0,
		AverageLatency:       0,
		LastDeliveryAt:       nil,
	}, nil
}

// WebhookStats represents webhook delivery statistics
type WebhookStats struct {
	TotalDeliveries      int64         `json:"total_deliveries"`
	SuccessfulDeliveries int64         `json:"successful_deliveries"`
	FailedDeliveries     int64         `json:"failed_deliveries"`
	AverageLatency       time.Duration `json:"average_latency"`
	LastDeliveryAt       *time.Time    `json:"last_delivery_at"`
}

// TestWebhookEndpoint tests a webhook endpoint configuration
func (s *DisputeWebhookService) TestWebhookEndpoint(ctx context.Context, webhook *WebhookConfig) error {
	testPayload := WebhookPayload{
		ID:        uuid.New().String(),
		EventType: "test",
		Timestamp: time.Now(),
		DisputeID: "test-dispute",
		Data: map[string]interface{}{
			"test":    true,
			"message": "This is a test webhook delivery",
		},
		Metadata: map[string]interface{}{
			"source": "webhook-test",
		},
	}

	delivery := s.sendWebhookAttempt(ctx, webhook, testPayload, 1)
	if !delivery.Success {
		return fmt.Errorf("webhook test failed: %s", *delivery.ErrorMessage)
	}

	return nil
}

// BulkRegisterWebhooks registers multiple webhook endpoints
func (s *DisputeWebhookService) BulkRegisterWebhooks(configs []*WebhookConfig) error {
	for _, config := range configs {
		if err := s.RegisterWebhook(config); err != nil {
			return fmt.Errorf("failed to register webhook %s: %w", config.URL, err)
		}
	}
	return nil
}

// GetRegisteredWebhooks returns all registered webhooks
func (s *DisputeWebhookService) GetRegisteredWebhooks() map[string]*WebhookConfig {
	// Return a copy to prevent external modifications
	result := make(map[string]*WebhookConfig)
	for id, config := range s.webhooks {
		configCopy := *config // Shallow copy
		result[id] = &configCopy
	}
	return result
}
