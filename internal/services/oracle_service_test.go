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

func TestOracleService_RegisterProvider(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	provider := &models.OracleProvider{
		Name:        "Test API Oracle",
		Description: "Test API Oracle for testing",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.example.com/verify",
		AuthConfig: models.OracleAuthConfig{
			Type: "bearer",
			ConfigData: map[string]string{
				"token": "test-token",
			},
		},
		RateLimitConfig: models.OracleRateLimitConfig{
			RequestsPerSecond: 10,
			BurstLimit:        20,
		},
		Capabilities: []string{"delivery_verification", "quality_check"},
	}

	// Mock expectations
	mockOracleRepo.On("CreateOracleProvider", mock.Anything, mock.MatchedBy(func(p *models.OracleProvider) bool {
		return p.Name == provider.Name && p.Type == provider.Type
	})).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.RegisterProvider(ctx, provider)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, provider.ID)
	assert.WithinDuration(t, time.Now(), provider.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), provider.UpdatedAt, time.Second)
	assert.True(t, provider.IsActive)
	assert.Equal(t, 1.0, provider.Reliability)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_RegisterProvider_InvalidConfig(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data with invalid config (missing name)
	provider := &models.OracleProvider{
		Type: models.OracleTypeAPI,
	}

	// Execute
	ctx := context.Background()
	err := service.RegisterProvider(ctx, provider)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider name is required")

	// Verify no repository calls were made
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_GetProvider(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	providerID := uuid.New()
	expectedProvider := &models.OracleProvider{
		ID:          providerID,
		Name:        "Test Oracle",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.example.com/verify",
		IsActive:    true,
		Reliability: 0.95,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expectations
	mockOracleRepo.On("GetOracleProviderByID", mock.Anything, providerID).Return(expectedProvider, nil)

	// Execute
	ctx := context.Background()
	provider, err := service.GetProvider(ctx, providerID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProvider, provider)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_UpdateProvider(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	provider := &models.OracleProvider{
		ID:          uuid.New(),
		Name:        "Updated Test Oracle",
		Description: "Updated description",
		Type:        models.OracleTypeAPI,
		Endpoint:    "https://api.updated-example.com/verify",
		IsActive:    false,
		Reliability: 0.85,
	}

	// Mock expectations
	mockOracleRepo.On("UpdateOracleProvider", mock.Anything, mock.MatchedBy(func(p *models.OracleProvider) bool {
		return p.ID == provider.ID && p.Name == provider.Name
	})).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.UpdateProvider(ctx, provider)

	// Assert
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), provider.UpdatedAt, time.Second)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_DeleteProvider(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	providerID := uuid.New()

	// Mock expectations
	mockOracleRepo.On("DeleteOracleProvider", mock.Anything, providerID).Return(nil)

	// Execute
	ctx := context.Background()
	err := service.DeleteProvider(ctx, providerID)

	// Assert
	require.NoError(t, err)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_ListProviders(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	expectedProviders := []*models.OracleProvider{
		{
			ID:   uuid.New(),
			Name: "API Oracle",
			Type: models.OracleTypeAPI,
		},
		{
			ID:   uuid.New(),
			Name: "Webhook Oracle",
			Type: models.OracleTypeWebhook,
		},
	}

	// Mock expectations
	mockOracleRepo.On("ListOracleProviders", mock.Anything, 10, 0).Return(expectedProviders, nil)

	// Execute
	ctx := context.Background()
	providers, err := service.ListProviders(ctx, 10, 0)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProviders, providers)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_GetActiveProviders(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	expectedProviders := []*models.OracleProvider{
		{
			ID:       uuid.New(),
			Name:     "Active API Oracle",
			Type:     models.OracleTypeAPI,
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Name:     "Active Webhook Oracle",
			Type:     models.OracleTypeWebhook,
			IsActive: true,
		},
	}

	// Mock expectations
	mockOracleRepo.On("GetActiveOracleProviders", mock.Anything).Return(expectedProviders, nil)

	// Execute
	ctx := context.Background()
	providers, err := service.GetActiveProviders(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProviders, providers)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}

func TestOracleService_GetProvidersByType(t *testing.T) {
	// Setup
	mockOracleRepo := &mocks.OracleRepositoryInterface{}
	mockMessaging := &messaging.Service{}

	service := NewOracleService(mockOracleRepo, mockMessaging)

	// Test data
	providerType := models.OracleTypeAPI
	expectedProviders := []*models.OracleProvider{
		{
			ID:   uuid.New(),
			Name: "API Oracle 1",
			Type: models.OracleTypeAPI,
		},
		{
			ID:   uuid.New(),
			Name: "API Oracle 2",
			Type: models.OracleTypeAPI,
		},
	}

	// Mock expectations
	mockOracleRepo.On("GetOracleProviderByType", mock.Anything, providerType).Return(expectedProviders, nil)

	// Execute
	ctx := context.Background()
	providers, err := service.GetProvidersByType(ctx, providerType)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedProviders, providers)

	// Verify mock expectations
	mockOracleRepo.AssertExpectations(t)
}
