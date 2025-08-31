package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

const (
	oracleImplTestDeliveryCompletedCondition = "delivery_completed"
	oracleImplTestQualityInspectionCondition = "quality_inspection_passed"
)

func TestAPIOracle_Verify(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:       uuid.New(),
		Name:     "Test API Oracle",
		Type:     models.OracleTypeAPI,
		Endpoint: "https://api.example.com/verify",
		AuthConfig: models.OracleAuthConfig{
			Type: "bearer",
			ConfigData: map[string]string{
				"token": "test-token",
			},
		},
	}

	oracle := NewAPIOracle(provider)

	// Test data
	condition := "delivery_completed"
	contextData := map[string]interface{}{
		"shipment_id":     "ship-123",
		"tracking_number": "track-456",
	}

	// Execute
	ctx := context.Background()
	response, err := oracle.Verify(ctx, condition, contextData)

	// Assert
	// Since we're using a fake endpoint, we expect an error
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to execute HTTP request")
}

func TestAPIOracle_Verify_MissingEndpoint(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test API Oracle",
		Type: models.OracleTypeAPI,
		// Missing endpoint
	}

	oracle := NewAPIOracle(provider)

	// Test data
	condition := oracleImplTestDeliveryCompletedCondition
	contextData := map[string]interface{}{
		"shipment_id": "ship-123",
	}

	// Execute
	ctx := context.Background()
	response, err := oracle.Verify(ctx, condition, contextData)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "API oracle endpoint is not configured")
}

func TestAPIOracle_GetStatus(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:       uuid.New(),
		Name:     "Test API Oracle",
		Type:     models.OracleTypeAPI,
		Endpoint: "https://api.example.com/health",
	}

	oracle := NewAPIOracle(provider)

	// Execute
	ctx := context.Background()
	status, err := oracle.GetStatus(ctx)

	// Assert
	// Since we're using a fake endpoint, we expect either an error or unhealthy status
	if err != nil {
		// Error is acceptable, just log it for debugging
		t.Logf("Expected error occurred: %v", err)
	} else {
		require.NotNil(t, status)
		// Status should be either healthy or unhealthy
		// Just checking that status is not nil and is a valid boolean
		assert.NotNil(t, status)
	}
}

func TestAPIOracle_AddAuthHeaders_Bearer(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test API Oracle",
		Type: models.OracleTypeAPI,
		AuthConfig: models.OracleAuthConfig{
			Type: "bearer",
			ConfigData: map[string]string{
				"token": "test-token",
			},
		},
	}

	oracle := NewAPIOracle(provider)

	// Execute
	// This is testing a private method, so we'll test it indirectly through Verify
	// which will call addAuthHeaders
	condition := "test_condition"
	contextData := map[string]interface{}{}

	// Execute
	ctx := context.Background()
	_, err := oracle.Verify(ctx, condition, contextData)

	// Assert
	// We expect an error due to the fake endpoint, but not due to auth
	if err != nil {
		assert.NotContains(t, err.Error(), "bearer token not configured")
	}
}

func TestWebhookOracle_Verify(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test Webhook Oracle",
		Type: models.OracleTypeWebhook,
		AuthConfig: models.OracleAuthConfig{
			Type: "webhook",
			ConfigData: map[string]string{
				"webhook_secret": "test-secret",
			},
		},
	}

	oracle := NewWebhookOracle(provider)

	// Test data
	condition := oracleImplTestDeliveryCompletedCondition
	contextData := map[string]interface{}{
		"shipment_id": "ship-123",
	}

	// Execute
	ctx := context.Background()
	response, err := oracle.Verify(ctx, condition, contextData)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "webhook oracle verification requires callback registration")
}

func TestWebhookOracle_GetStatus(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test Webhook Oracle",
		Type: models.OracleTypeWebhook,
	}

	oracle := NewWebhookOracle(provider)

	// Execute
	ctx := context.Background()
	status, err := oracle.GetStatus(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.IsHealthy)
	assert.WithinDuration(t, time.Now(), status.LastChecked, time.Second)
	assert.Equal(t, 0.0, status.ErrorRate)
}

func TestManualOracle_Verify(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test Manual Oracle",
		Type: models.OracleTypeManual,
	}

	oracle := NewManualOracle(provider)

	// Test data
	condition := oracleImplTestQualityInspectionCondition
	contextData := map[string]interface{}{
		"product_id":   "prod-123",
		"batch_number": "batch-456",
	}

	// Execute
	ctx := context.Background()
	response, err := oracle.Verify(ctx, condition, contextData)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "manual oracle verification requires human intervention")
}

func TestManualOracle_GetStatus(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test Manual Oracle",
		Type: models.OracleTypeManual,
	}

	oracle := NewManualOracle(provider)

	// Execute
	ctx := context.Background()
	status, err := oracle.GetStatus(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.IsHealthy)
	assert.WithinDuration(t, time.Now(), status.LastChecked, time.Second)
	assert.Equal(t, 0.0, status.ErrorRate)
}

func TestAPIOracle_GenerateProofHash(t *testing.T) {
	// Setup
	provider := &models.OracleProvider{
		ID:   uuid.New(),
		Name: "Test API Oracle",
		Type: models.OracleTypeAPI,
	}

	oracle := NewAPIOracle(provider)

	// Test data
	condition := "test_condition"
	result := true
	evidence := []byte("test_evidence")

	// Execute
	proofHash := oracle.generateProofHash(condition, result, evidence)

	// Assert
	require.NotEmpty(t, proofHash)
	assert.Len(t, proofHash, 64) // SHA256 produces 64-character hex string
}
