package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// OracleService handles oracle provider management and request orchestration
type OracleService struct {
	oracleRepo      repository.OracleRepositoryInterface
	messagingClient *messaging.Service
}

// NewOracleService creates a new oracle service
func NewOracleService(
	oracleRepo repository.OracleRepositoryInterface,
	messagingClient *messaging.Service,
) *OracleService {
	return &OracleService{
		oracleRepo:      oracleRepo,
		messagingClient: messagingClient,
	}
}

// RegisterProvider registers a new oracle provider
func (s *OracleService) RegisterProvider(ctx context.Context, provider *models.OracleProvider) error {
	// Validate provider configuration
	if err := s.validateProvider(provider); err != nil {
		return fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Set default values
	provider.ID = uuid.New()
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()
	provider.IsActive = true
	provider.Reliability = 1.0 // Start with perfect reliability

	// Create provider in repository
	if err := s.oracleRepo.CreateOracleProvider(ctx, provider); err != nil {
		return fmt.Errorf("failed to create oracle provider: %w", err)
	}

	// Publish event if messaging client is available
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "oracle_provider_registered",
			Source: "oracle_service",
			Data: map[string]interface{}{
				"event":         "oracle_provider_registered",
				"provider_id":   provider.ID.String(),
				"provider_type": provider.Type,
				"timestamp":     time.Now().Format(time.RFC3339),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			log.Printf("Warning: Failed to publish oracle provider registered event: %v", err)
		}
	}

	return nil
}

// GetProvider retrieves an oracle provider by ID
func (s *OracleService) GetProvider(ctx context.Context, providerID uuid.UUID) (*models.OracleProvider, error) {
	provider, err := s.oracleRepo.GetOracleProviderByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle provider: %w", err)
	}
	return provider, nil
}

// UpdateProvider updates an existing oracle provider
func (s *OracleService) UpdateProvider(ctx context.Context, provider *models.OracleProvider) error {
	// Validate provider configuration
	if err := s.validateProvider(provider); err != nil {
		return fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Update timestamp
	provider.UpdatedAt = time.Now()

	// Update provider in repository
	if err := s.oracleRepo.UpdateOracleProvider(ctx, provider); err != nil {
		return fmt.Errorf("failed to update oracle provider: %w", err)
	}

	// Publish event if messaging client is available
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "oracle_provider_updated",
			Source: "oracle_service",
			Data: map[string]interface{}{
				"event":       "oracle_provider_updated",
				"provider_id": provider.ID.String(),
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			log.Printf("Warning: Failed to publish oracle provider updated event: %v", err)
		}
	}

	return nil
}

// DeleteProvider removes an oracle provider
func (s *OracleService) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	// Delete provider from repository
	if err := s.oracleRepo.DeleteOracleProvider(ctx, providerID); err != nil {
		return fmt.Errorf("failed to delete oracle provider: %w", err)
	}

	// Publish event if messaging client is available
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "oracle_provider_deleted",
			Source: "oracle_service",
			Data: map[string]interface{}{
				"event":       "oracle_provider_deleted",
				"provider_id": providerID.String(),
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		if err := s.messagingClient.PublishEvent(event); err != nil {
			log.Printf("Warning: Failed to publish oracle provider deleted event: %v", err)
		}
	}

	return nil
}

// ListProviders lists all oracle providers
func (s *OracleService) ListProviders(ctx context.Context, limit, offset int) ([]*models.OracleProvider, error) {
	providers, err := s.oracleRepo.ListOracleProviders(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list oracle providers: %w", err)
	}
	return providers, nil
}

// GetActiveProviders retrieves all active oracle providers
func (s *OracleService) GetActiveProviders(ctx context.Context) ([]*models.OracleProvider, error) {
	providers, err := s.oracleRepo.GetActiveOracleProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active oracle providers: %w", err)
	}
	return providers, nil
}

// GetProvidersByType retrieves oracle providers by type
func (s *OracleService) GetProvidersByType(ctx context.Context, providerType models.OracleType) ([]*models.OracleProvider, error) {
	providers, err := s.oracleRepo.GetOracleProviderByType(ctx, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle providers by type: %w", err)
	}
	return providers, nil
}

// HealthCheck performs a health check on a provider
func (s *OracleService) HealthCheck(ctx context.Context, providerID uuid.UUID) (*models.OracleStatus, error) {
	status, err := s.oracleRepo.HealthCheckOracleProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to perform health check: %w", err)
	}
	return status, nil
}

// GetRequest retrieves an oracle request by ID
func (s *OracleService) GetRequest(ctx context.Context, requestID uuid.UUID) (*models.OracleRequest, error) {
	request, err := s.oracleRepo.GetOracleRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle request: %w", err)
	}
	return request, nil
}

// ListRequests lists oracle requests with filtering
func (s *OracleService) ListRequests(ctx context.Context, filter *repository.OracleRequestFilter, limit, offset int) ([]*models.OracleRequest, error) {
	requests, err := s.oracleRepo.ListOracleRequests(ctx, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list oracle requests: %w", err)
	}
	return requests, nil
}

// GetCachedResponse retrieves a cached response if available
func (s *OracleService) GetCachedResponse(ctx context.Context, condition string) (*models.OracleRequest, error) {
	cachedRequest, err := s.oracleRepo.GetCachedResponse(ctx, condition)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached response: %w", err)
	}

	// Check if cache is still valid
	if cachedRequest.CachedUntil != nil && time.Now().After(*cachedRequest.CachedUntil) {
		return nil, fmt.Errorf("cached response has expired")
	}

	return cachedRequest, nil
}

// CacheResponse caches a response
func (s *OracleService) CacheResponse(ctx context.Context, request *models.OracleRequest) error {
	// Set cache expiration (24 hours by default)
	expiration := time.Now().Add(24 * time.Hour)
	request.CachedUntil = &expiration

	if err := s.oracleRepo.CacheResponse(ctx, request); err != nil {
		return fmt.Errorf("failed to cache response: %w", err)
	}
	return nil
}

// ClearExpiredCache clears expired cached responses
func (s *OracleService) ClearExpiredCache(ctx context.Context) error {
	if err := s.oracleRepo.ClearExpiredCache(ctx); err != nil {
		return fmt.Errorf("failed to clear expired cache: %w", err)
	}
	return nil
}

// GetMetrics retrieves performance metrics for an oracle provider
func (s *OracleService) GetMetrics(ctx context.Context, providerID uuid.UUID) (*models.OracleMetrics, error) {
	metrics, err := s.oracleRepo.GetOracleMetrics(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle metrics: %w", err)
	}
	return metrics, nil
}

// GetReliabilityScore retrieves the reliability score for an oracle provider
func (s *OracleService) GetReliabilityScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	score, err := s.oracleRepo.GetOracleReliabilityScore(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get reliability score: %w", err)
	}
	return score, nil
}

// validateProvider validates oracle provider configuration
func (s *OracleService) validateProvider(provider *models.OracleProvider) error {
	if provider.Name == "" {
		return fmt.Errorf("provider name is required")
	}

	if provider.Type == "" {
		return fmt.Errorf("provider type is required")
	}

	// Validate endpoint for API and webhook oracles
	if (provider.Type == models.OracleTypeAPI || provider.Type == models.OracleTypeWebhook) && provider.Endpoint == "" {
		return fmt.Errorf("endpoint is required for API and webhook oracles")
	}

	// Validate rate limits
	if provider.RateLimitConfig.RequestsPerSecond <= 0 {
		provider.RateLimitConfig.RequestsPerSecond = 1 // Default to 1 request per second
	}

	if provider.RateLimitConfig.BurstLimit <= 0 {
		provider.RateLimitConfig.BurstLimit = 1 // Default to 1 burst request
	}

	return nil
}
