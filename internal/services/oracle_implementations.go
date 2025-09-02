package services

import (
	"bytes"
	"context"
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

// OracleCallback defines the callback function signature for subscriptions
type OracleCallback func(response *models.OracleResponse, err error)

// OracleInterface defines the interface for verification services
type OracleInterface interface {
	// Verify evaluates a condition with context and returns the result
	Verify(ctx context.Context, condition string, contextData interface{}) (*models.OracleResponse, error)

	// GetProof returns verification evidence for a completed verification
	GetProof(ctx context.Context, requestID string) ([]byte, error)

	// GetStatus returns the health and availability status of the oracle
	GetStatus(ctx context.Context) (*models.OracleStatus, error)

	// Subscribe registers a callback for event-driven verification
	Subscribe(ctx context.Context, condition string, callback OracleCallback) (string, error)

	// Unsubscribe removes a subscription by ID
	Unsubscribe(ctx context.Context, subscriptionID string) error
}

// APIOracle implements oracle functionality for REST API verification
type APIOracle struct {
	provider   *models.OracleProvider
	httpClient *http.Client
}

// NewAPIOracle creates a new API oracle implementation
func NewAPIOracle(provider *models.OracleProvider) *APIOracle {
	return &APIOracle{
		provider: provider,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Verify evaluates a condition with context via API call
func (o *APIOracle) Verify(ctx context.Context, condition string, contextData interface{}) (*models.OracleResponse, error) {
	if o.provider.Endpoint == "" {
		return nil, fmt.Errorf("API oracle endpoint is not configured")
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"condition": condition,
		"context":   contextData,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", o.provider.Endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers
	if err := o.addAuthHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to add authentication headers: %w", err)
	}

	// Execute request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Result     bool        `json:"result"`
		Confidence float64     `json:"confidence"`
		Evidence   interface{} `json:"evidence"`
		Metadata   interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert evidence to bytes
	evidenceBytes, err := json.Marshal(apiResponse.Evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal evidence: %w", err)
	}

	// Generate proof hash
	proofHash := o.generateProofHash(condition, apiResponse.Result, evidenceBytes)

	response := &models.OracleResponse{
		RequestID:  uuid.New(),
		Condition:  condition,
		Result:     apiResponse.Result,
		Confidence: apiResponse.Confidence,
		Evidence:   evidenceBytes,
		Metadata:   apiResponse.Metadata,
		VerifiedAt: time.Now(),
		ProofHash:  proofHash,
	}

	return response, nil
}

// GetProof returns verification evidence for a completed verification
func (o *APIOracle) GetProof(_ context.Context, _ string) ([]byte, error) {
	// In a real implementation, this would retrieve the proof from storage
	// For now, we'll return a placeholder
	return []byte("proof_data"), nil
}

// GetStatus returns the health and availability status of the oracle
func (o *APIOracle) GetStatus(ctx context.Context) (*models.OracleStatus, error) {
	// Perform a simple health check
	req, err := http.NewRequestWithContext(ctx, "GET", o.provider.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return &models.OracleStatus{
			IsHealthy:   false,
			LastChecked: time.Now(),
			ErrorRate:   1.0,
		}, nil
	}
	defer resp.Body.Close()

	return &models.OracleStatus{
		IsHealthy:   resp.StatusCode == http.StatusOK,
		LastChecked: time.Now(),
		ErrorRate:   0.0,
	}, nil
}

// Subscribe registers a callback for event-driven verification
func (o *APIOracle) Subscribe(_ context.Context, _ string, _ OracleCallback) (string, error) {
	// API oracles typically don't support subscriptions directly
	// This would need to be implemented via webhooks on the API side
	return "", fmt.Errorf("API oracle does not support direct subscriptions")
}

// Unsubscribe removes a subscription by ID
func (o *APIOracle) Unsubscribe(_ context.Context, _ string) error {
	// Not supported for API oracles
	return nil
}

// addAuthHeaders adds authentication headers based on provider configuration
func (o *APIOracle) addAuthHeaders(req *http.Request) error {
	switch o.provider.AuthConfig.Type {
	case "bearer":
		token, ok := o.provider.AuthConfig.ConfigData["token"]
		if !ok {
			return fmt.Errorf("bearer token not configured")
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case "api_key":
		key, ok := o.provider.AuthConfig.ConfigData["key"]
		if !ok {
			return fmt.Errorf("API key not configured")
		}
		headerName, ok := o.provider.AuthConfig.ConfigData["header"]
		if !ok {
			headerName = "X-API-Key"
		}
		req.Header.Set(headerName, key)
	case "oauth":
		// OAuth implementation would go here
		return fmt.Errorf("OAuth authentication not yet implemented")
	}
	return nil
}

// generateProofHash generates a hash for proof integrity verification
func (o *APIOracle) generateProofHash(condition string, result bool, evidence []byte) string {
	data := fmt.Sprintf("%s:%t:%s", condition, result, string(evidence))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// WebhookOracle implements oracle functionality for event-driven verification
type WebhookOracle struct {
	provider *models.OracleProvider
	// callback OracleCallback  // Currently unused but kept for future implementation
}

// NewWebhookOracle creates a new webhook oracle implementation
func NewWebhookOracle(provider *models.OracleProvider) *WebhookOracle {
	return &WebhookOracle{
		provider: provider,
	}
}

// Verify evaluates a condition with context and sets up webhook notification
func (o *WebhookOracle) Verify(_ context.Context, _ string, _ interface{}) (*models.OracleResponse, error) {
	// For webhook oracles, we register the verification request and wait for callback
	// This is a simplified implementation
	return nil, fmt.Errorf("webhook oracle verification requires callback registration")
}

// GetProof returns verification evidence for a completed verification
func (o *WebhookOracle) GetProof(_ context.Context, _ string) ([]byte, error) {
	// Retrieve proof from storage
	return []byte("webhook_proof_data"), nil
}

// GetStatus returns the health and availability status of the oracle
func (o *WebhookOracle) GetStatus(_ context.Context) (*models.OracleStatus, error) {
	// Webhook oracles are typically always available since they're passive
	return &models.OracleStatus{
		IsHealthy:   true,
		LastChecked: time.Now(),
		ErrorRate:   0.0,
	}, nil
}

// Subscribe registers a callback for event-driven verification
func (o *WebhookOracle) Subscribe(_ context.Context, _ string, _ OracleCallback) (string, error) {
	// o.callback = callback  // Commented out as field is removed
	subscriptionID := uuid.New().String()

	// In a real implementation, this would register the webhook with the external service
	return subscriptionID, nil
}

// Unsubscribe removes a subscription by ID
func (o *WebhookOracle) Unsubscribe(_ context.Context, _ string) error {
	// o.callback = nil  // Commented out as field is removed
	return nil
}

// ManualOracle implements oracle functionality for human verification
type ManualOracle struct {
	provider *models.OracleProvider
}

// NewManualOracle creates a new manual oracle implementation
func NewManualOracle(provider *models.OracleProvider) *ManualOracle {
	return &ManualOracle{
		provider: provider,
	}
}

// Verify creates a manual verification task for human review
func (o *ManualOracle) Verify(_ context.Context, _ string, _ interface{}) (*models.OracleResponse, error) {
	// For manual oracles, we create a task for human verification
	// This is a simplified implementation that returns a pending response
	return nil, fmt.Errorf("manual oracle verification requires human intervention")
}

// GetProof returns verification evidence for a completed verification
func (o *ManualOracle) GetProof(_ context.Context, _ string) ([]byte, error) {
	// Retrieve proof from storage
	return []byte("manual_proof_data"), nil
}

// GetStatus returns the health and availability status of the oracle
func (o *ManualOracle) GetStatus(_ context.Context) (*models.OracleStatus, error) {
	// Manual oracles are always available (assuming humans are available)
	return &models.OracleStatus{
		IsHealthy:   true,
		LastChecked: time.Now(),
		ErrorRate:   0.0,
	}, nil
}

// Subscribe registers a callback for event-driven verification
func (o *ManualOracle) Subscribe(_ context.Context, _ string, _ OracleCallback) (string, error) {
	// Manual oracles don't typically support subscriptions
	return "", fmt.Errorf("manual oracle does not support subscriptions")
}

// Unsubscribe removes a subscription by ID
func (o *ManualOracle) Unsubscribe(_ context.Context, _ string) error {
	// Not applicable for manual oracles
	return nil
}

// verifyWebhookSignature verifies the signature of an incoming webhook
// func (o *WebhookOracle) verifyWebhookSignature(payload []byte, signature string) bool {
// 	secret, ok := o.provider.AuthConfig.ConfigData["webhook_secret"]
// 	if !ok {
// 		return false
// 	}
//
// 	// Calculate expected signature
// 	h := hmac.New(sha256.New, []byte(secret))
// 	h.Write(payload)
// 	expectedSignature := hex.EncodeToString(h.Sum(nil))
//
// 	return hmac.Equal([]byte(signature), []byte(expectedSignature))
// }
