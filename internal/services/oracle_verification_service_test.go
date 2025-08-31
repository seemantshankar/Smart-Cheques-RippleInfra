package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

const (
	oracleTestDeliveryCompletedCondition = "delivery_completed"
)

func TestOracleVerificationService_VerifyMilestone(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Test data
	milestoneID := "milestone-123"
	condition := oracleTestDeliveryCompletedCondition
	oracleConfig := &models.OracleConfig{
		Type:     "api",
		Endpoint: "https://api.example.com/verify",
		Config: map[string]interface{}{
			"api_key": "test-key",
		},
	}

	// Mock expectations for GetProvidersByType
	providers := []*models.OracleProvider{
		{
			ID:          uuid.New(),
			Name:        "Test API Oracle",
			Type:        models.OracleTypeAPI,
			Endpoint:    "https://api.example.com/verify",
			Reliability: 0.95,
		},
	}
	mockOracleRepo.On("GetOracleProviderByType", mock.Anything, models.OracleTypeAPI).Return(providers, nil)

	// Mock expectations for GetCachedResponse (should return error to simulate cache miss)
	mockOracleRepo.On("GetCachedResponse", mock.Anything, condition).Return(nil, assert.AnError)

	// Mock expectations for CreateOracleRequest
	mockOracleRepo.On("CreateOracleRequest", mock.Anything, mock.AnythingOfType("*models.OracleRequest")).Return(nil)

	// Execute
	ctx := context.Background()
	response, err := verificationService.VerifyMilestone(ctx, milestoneID, condition, oracleConfig)

	// Assert
	// Since we're using a mock API oracle that will fail, we expect an error
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "oracle verification failed")

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleVerificationService_GetVerificationResult(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Test data
	requestID := uuid.New()
	result := true
	confidence := 0.95
	verifiedAt := time.Now()
	proofHash := "test-proof-hash"

	expectedRequest := &models.OracleRequest{
		ID:         requestID,
		ProviderID: uuid.New(),
		Condition:  oracleTestDeliveryCompletedCondition,
		Status:     models.RequestStatusCompleted,
		Result:     &result,
		Confidence: &confidence,
		Evidence:   []byte("test evidence"),
		Metadata:   map[string]interface{}{"source": "api"},
		VerifiedAt: &verifiedAt,
		ProofHash:  &proofHash,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	expectedResponse := &models.OracleResponse{
		RequestID:  requestID,
		Condition:  oracleTestDeliveryCompletedCondition,
		Result:     result,
		Confidence: confidence,
		Evidence:   []byte("test evidence"),
		Metadata:   map[string]interface{}{"source": "api"},
		VerifiedAt: verifiedAt,
		ProofHash:  proofHash,
	}

	// Mock expectations
	mockOracleRepo.On("GetOracleRequestByID", mock.Anything, requestID).Return(expectedRequest, nil)

	// Execute
	ctx := context.Background()
	response, err := verificationService.GetVerificationResult(ctx, requestID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleVerificationService_GetVerificationResult_NotCompleted(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Test data
	requestID := uuid.New()

	expectedRequest := &models.OracleRequest{
		ID:         requestID,
		ProviderID: uuid.New(),
		Condition:  oracleTestDeliveryCompletedCondition,
		Status:     models.RequestStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockOracleRepo.On("GetOracleRequestByID", mock.Anything, requestID).Return(expectedRequest, nil)

	// Execute
	ctx := context.Background()
	response, err := verificationService.GetVerificationResult(ctx, requestID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "verification is not yet completed")

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleVerificationService_GetProof(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Test data
	requestID := uuid.New()
	expectedEvidence := []byte("test evidence data")

	expectedRequest := &models.OracleRequest{
		ID:         requestID,
		ProviderID: uuid.New(),
		Condition:  oracleTestDeliveryCompletedCondition,
		Status:     models.RequestStatusCompleted,
		Result:     boolPtr(true),
		Confidence: float64Ptr(0.95),
		Evidence:   expectedEvidence,
		Metadata:   map[string]interface{}{"source": "api"},
		VerifiedAt: timePtr(time.Now()),
		ProofHash:  stringPtr("test-proof-hash"),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Mock expectations
	mockOracleRepo.On("GetOracleRequestByID", mock.Anything, requestID).Return(expectedRequest, nil)

	// Execute
	ctx := context.Background()
	evidence, err := verificationService.GetProof(ctx, requestID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedEvidence, evidence)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleVerificationService_SelectBestProvider(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Test data
	providers := []*models.OracleProvider{
		{
			ID:          uuid.New(),
			Name:        "Low Reliability Oracle",
			Type:        models.OracleTypeAPI,
			Reliability: 0.75,
		},
		{
			ID:          uuid.New(),
			Name:        "High Reliability Oracle",
			Type:        models.OracleTypeAPI,
			Reliability: 0.95,
		},
		{
			ID:          uuid.New(),
			Name:        "Medium Reliability Oracle",
			Type:        models.OracleTypeAPI,
			Reliability: 0.85,
		},
	}

	// Execute
	bestProvider := verificationService.selectBestProvider(providers)

	// Assert
	require.NotNil(t, bestProvider)
	assert.Equal(t, "High Reliability Oracle", bestProvider.Name)
	assert.Equal(t, 0.95, bestProvider.Reliability)
}

func TestOracleVerificationService_SelectBestProvider_Empty(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	mockOracleService := NewOracleService(mockOracleRepo, mockMessaging)
	verificationService := NewOracleVerificationService(mockOracleService, mockOracleRepo, mockMessaging)

	// Execute
	bestProvider := verificationService.selectBestProvider([]*models.OracleProvider{})

	// Assert
	assert.Nil(t, bestProvider)
}

// Helper functions for pointer creation
func boolPtr(b bool) *bool {
	return &b
}

func float64Ptr(f float64) *float64 {
	return &f
}
