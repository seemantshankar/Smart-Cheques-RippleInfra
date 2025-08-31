package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// OracleRepositoryInterface defines the interface for oracle repository operations
type OracleRepositoryInterface interface {
	// Oracle provider management
	CreateOracleProvider(ctx context.Context, provider *models.OracleProvider) error
	GetOracleProviderByID(ctx context.Context, id uuid.UUID) (*models.OracleProvider, error)
	GetOracleProviderByType(ctx context.Context, providerType models.OracleType) ([]*models.OracleProvider, error)
	UpdateOracleProvider(ctx context.Context, provider *models.OracleProvider) error
	DeleteOracleProvider(ctx context.Context, id uuid.UUID) error
	ListOracleProviders(ctx context.Context, limit, offset int) ([]*models.OracleProvider, error)
	GetActiveOracleProviders(ctx context.Context) ([]*models.OracleProvider, error)
	HealthCheckOracleProvider(ctx context.Context, id uuid.UUID) (*models.OracleStatus, error)

	// Oracle request management
	CreateOracleRequest(ctx context.Context, request *models.OracleRequest) error
	GetOracleRequestByID(ctx context.Context, id uuid.UUID) (*models.OracleRequest, error)
	UpdateOracleRequest(ctx context.Context, request *models.OracleRequest) error
	DeleteOracleRequest(ctx context.Context, id uuid.UUID) error
	ListOracleRequests(ctx context.Context, filter *OracleRequestFilter, limit, offset int) ([]*models.OracleRequest, error)
	GetOracleRequestsByStatus(ctx context.Context, status models.RequestStatus, limit, offset int) ([]*models.OracleRequest, error)
	GetOracleRequestsByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*models.OracleRequest, error)

	// Caching operations
	GetCachedResponse(ctx context.Context, condition string) (*models.OracleRequest, error)
	CacheResponse(ctx context.Context, request *models.OracleRequest) error
	ClearExpiredCache(ctx context.Context) error

	// Analytics and metrics
	GetOracleMetrics(ctx context.Context, providerID uuid.UUID) (*models.OracleMetrics, error)
	GetOracleReliabilityScore(ctx context.Context, providerID uuid.UUID) (float64, error)
	GetRequestStats(ctx context.Context, providerID *uuid.UUID, startDate, endDate *time.Time) (map[models.RequestStatus]int64, error)
}

// OracleRequestFilter defines filtering options for oracle requests
type OracleRequestFilter struct {
	ProviderID    *uuid.UUID            `json:"provider_id,omitempty"`
	Status        *models.RequestStatus `json:"status,omitempty"`
	Condition     *string               `json:"condition,omitempty"`
	DateFrom      *time.Time            `json:"date_from,omitempty"`
	DateTo        *time.Time            `json:"date_to,omitempty"`
	MinConfidence *float64              `json:"min_confidence,omitempty"`
}
