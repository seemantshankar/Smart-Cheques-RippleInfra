package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

// OracleRepository implements OracleRepositoryInterface for PostgreSQL
type OracleRepository struct {
	db *sql.DB
}

// NewOracleRepository creates a new oracle repository
func NewOracleRepository(db *sql.DB) *OracleRepository {
	return &OracleRepository{
		db: db,
	}
}

// scanOracleRequests scans oracle request rows and returns a slice of OracleRequest pointers
func (r *OracleRepository) scanOracleRequests(rows *sql.Rows) ([]*models.OracleRequest, error) {
	var requests []*models.OracleRequest

	for rows.Next() {
		var request models.OracleRequest
		var contextDataBytes, metadataBytes []byte

		err := rows.Scan(
			&request.ID,
			&request.ProviderID,
			&request.Condition,
			&contextDataBytes,
			&request.Status,
			&request.Result,
			&request.Confidence,
			&request.Evidence,
			&metadataBytes,
			&request.VerifiedAt,
			&request.ProofHash,
			&request.RetryCount,
			&request.ErrorMessage,
			&request.CreatedAt,
			&request.UpdatedAt,
			&request.CachedUntil,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan oracle request: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(contextDataBytes, &request.ContextData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal context data: %w", err)
		}

		if err := json.Unmarshal(metadataBytes, &request.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		requests = append(requests, &request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating oracle requests: %w", err)
	}

	return requests, nil
}

// CreateOracleProvider creates a new oracle provider
func (r *OracleRepository) CreateOracleProvider(ctx context.Context, provider *models.OracleProvider) error {
	query := `
		INSERT INTO oracle_providers (
			id, name, description, type, endpoint, auth_config, rate_limit_config,
			is_active, reliability, response_time, capabilities, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	authConfigBytes, err := json.Marshal(provider.AuthConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}

	rateLimitConfigBytes, err := json.Marshal(provider.RateLimitConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit config: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		provider.ID,
		provider.Name,
		provider.Description,
		string(provider.Type),
		provider.Endpoint,
		authConfigBytes,
		rateLimitConfigBytes,
		provider.IsActive,
		provider.Reliability,
		provider.ResponseTime.Nanoseconds(),
		pq.Array(provider.Capabilities),
		provider.CreatedAt,
		provider.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create oracle provider: %w", err)
	}

	return nil
}

// scanOracleProvider scans a single oracle provider from a row
func (r *OracleRepository) scanOracleProvider(provider *models.OracleProvider, authConfigBytes, rateLimitConfigBytes []byte, responseTimeNanos int64, capabilities []string) error {
	// Unmarshal JSON fields
	if err := json.Unmarshal(authConfigBytes, &provider.AuthConfig); err != nil {
		return fmt.Errorf("failed to unmarshal auth config: %w", err)
	}

	if err := json.Unmarshal(rateLimitConfigBytes, &provider.RateLimitConfig); err != nil {
		return fmt.Errorf("failed to unmarshal rate limit config: %w", err)
	}

	provider.ResponseTime = time.Duration(responseTimeNanos)
	provider.Capabilities = capabilities

	return nil
}

// GetOracleProviderByID retrieves an oracle provider by ID
func (r *OracleRepository) GetOracleProviderByID(ctx context.Context, id uuid.UUID) (*models.OracleProvider, error) {
	query := `
		SELECT 
			id, name, description, type, endpoint, auth_config, rate_limit_config,
			is_active, reliability, response_time, capabilities, created_at, updated_at
		FROM oracle_providers
		WHERE id = $1
	`

	var provider models.OracleProvider
	var authConfigBytes, rateLimitConfigBytes []byte
	var responseTimeNanos int64
	var capabilities []string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Description,
		&provider.Type,
		&provider.Endpoint,
		&authConfigBytes,
		&rateLimitConfigBytes,
		&provider.IsActive,
		&provider.Reliability,
		&responseTimeNanos,
		pq.Array(&capabilities),
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oracle provider not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oracle provider: %w", err)
	}

	if err := r.scanOracleProvider(&provider, authConfigBytes, rateLimitConfigBytes, responseTimeNanos, capabilities); err != nil {
		return nil, err
	}

	return &provider, nil
}

// GetOracleProviderByType retrieves oracle providers by type
func (r *OracleRepository) GetOracleProviderByType(ctx context.Context, providerType models.OracleType) ([]*models.OracleProvider, error) {
	query := `
		SELECT 
			id, name, description, type, endpoint, auth_config, rate_limit_config,
			is_active, reliability, response_time, capabilities, created_at, updated_at
		FROM oracle_providers
		WHERE type = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(providerType))
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle providers: %w", err)
	}
	defer rows.Close()

	var providers []*models.OracleProvider

	for rows.Next() {
		var provider models.OracleProvider
		var authConfigBytes, rateLimitConfigBytes []byte
		var responseTimeNanos int64
		var capabilities []string

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Description,
			&provider.Type,
			&provider.Endpoint,
			&authConfigBytes,
			&rateLimitConfigBytes,
			&provider.IsActive,
			&provider.Reliability,
			&responseTimeNanos,
			pq.Array(&capabilities),
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan oracle provider: %w", err)
		}

		if err := r.scanOracleProvider(&provider, authConfigBytes, rateLimitConfigBytes, responseTimeNanos, capabilities); err != nil {
			return nil, err
		}

		providers = append(providers, &provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating oracle providers: %w", err)
	}

	return providers, nil
}

// UpdateOracleProvider updates an existing oracle provider
func (r *OracleRepository) UpdateOracleProvider(ctx context.Context, provider *models.OracleProvider) error {
	query := `
		UPDATE oracle_providers SET
			name = $1, description = $2, type = $3, endpoint = $4, auth_config = $5,
			rate_limit_config = $6, is_active = $7, reliability = $8, response_time = $9,
			capabilities = $10, updated_at = $11
		WHERE id = $12
	`

	authConfigBytes, err := json.Marshal(provider.AuthConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}

	rateLimitConfigBytes, err := json.Marshal(provider.RateLimitConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit config: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		provider.Name,
		provider.Description,
		string(provider.Type),
		provider.Endpoint,
		authConfigBytes,
		rateLimitConfigBytes,
		provider.IsActive,
		provider.Reliability,
		provider.ResponseTime.Nanoseconds(),
		pq.Array(provider.Capabilities),
		provider.UpdatedAt,
		provider.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update oracle provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oracle provider not found: %s", provider.ID.String())
	}

	return nil
}

// DeleteOracleProvider deletes an oracle provider
func (r *OracleRepository) DeleteOracleProvider(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oracle_providers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oracle provider: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oracle provider not found: %s", id.String())
	}

	return nil
}

// ListOracleProviders lists all oracle providers
func (r *OracleRepository) ListOracleProviders(ctx context.Context, limit, offset int) ([]*models.OracleProvider, error) {
	query := `
		SELECT 
			id, name, description, type, endpoint, auth_config, rate_limit_config,
			is_active, reliability, response_time, capabilities, created_at, updated_at
		FROM oracle_providers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle providers: %w", err)
	}
	defer rows.Close()

	var providers []*models.OracleProvider

	for rows.Next() {
		var provider models.OracleProvider
		var authConfigBytes, rateLimitConfigBytes []byte
		var responseTimeNanos int64
		var capabilities []string

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Description,
			&provider.Type,
			&provider.Endpoint,
			&authConfigBytes,
			&rateLimitConfigBytes,
			&provider.IsActive,
			&provider.Reliability,
			&responseTimeNanos,
			pq.Array(&capabilities),
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan oracle provider: %w", err)
		}

		if err := r.scanOracleProvider(&provider, authConfigBytes, rateLimitConfigBytes, responseTimeNanos, capabilities); err != nil {
			return nil, err
		}

		providers = append(providers, &provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating oracle providers: %w", err)
	}

	return providers, nil
}

// GetActiveOracleProviders retrieves all active oracle providers
func (r *OracleRepository) GetActiveOracleProviders(ctx context.Context) ([]*models.OracleProvider, error) {
	query := `
		SELECT 
			id, name, description, type, endpoint, auth_config, rate_limit_config,
			is_active, reliability, response_time, capabilities, created_at, updated_at
		FROM oracle_providers
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active oracle providers: %w", err)
	}
	defer rows.Close()

	var providers []*models.OracleProvider

	for rows.Next() {
		var provider models.OracleProvider
		var authConfigBytes, rateLimitConfigBytes []byte
		var responseTimeNanos int64
		var capabilities []string

		err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Description,
			&provider.Type,
			&provider.Endpoint,
			&authConfigBytes,
			&rateLimitConfigBytes,
			&provider.IsActive,
			&provider.Reliability,
			&responseTimeNanos,
			pq.Array(&capabilities),
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan oracle provider: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(authConfigBytes, &provider.AuthConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal auth config: %w", err)
		}

		if err := json.Unmarshal(rateLimitConfigBytes, &provider.RateLimitConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rate limit config: %w", err)
		}

		provider.ResponseTime = time.Duration(responseTimeNanos)
		provider.Capabilities = capabilities

		providers = append(providers, &provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active oracle providers: %w", err)
	}

	return providers, nil
}

// HealthCheckOracleProvider performs a health check on a provider
func (r *OracleRepository) HealthCheckOracleProvider(ctx context.Context, id uuid.UUID) (*models.OracleStatus, error) {
	// This is a simplified implementation. In a real system, this would perform
	// an actual health check against the oracle provider.

	provider, err := r.GetOracleProviderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle provider: %w", err)
	}

	status := &models.OracleStatus{
		IsHealthy:    provider.IsActive,
		LastChecked:  time.Now(),
		ErrorRate:    0.0,     // This would be calculated from request history
		Capacity:     100,     // Placeholder
		Load:         0,       // Placeholder
		Version:      "1.0.0", // Placeholder
		Capabilities: provider.Capabilities,
		Reliability:  provider.Reliability,
	}

	return status, nil
}

// CreateOracleRequest creates a new oracle request
func (r *OracleRepository) CreateOracleRequest(ctx context.Context, request *models.OracleRequest) error {
	query := `
		INSERT INTO oracle_requests (
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	contextDataBytes, err := json.Marshal(request.ContextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	metadataBytes, err := json.Marshal(request.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		request.ID,
		request.ProviderID,
		request.Condition,
		contextDataBytes,
		string(request.Status),
		request.Result,
		request.Confidence,
		request.Evidence,
		metadataBytes,
		request.VerifiedAt,
		request.ProofHash,
		request.RetryCount,
		request.ErrorMessage,
		request.CreatedAt,
		request.UpdatedAt,
		request.CachedUntil,
	)

	if err != nil {
		return fmt.Errorf("failed to create oracle request: %w", err)
	}

	return nil
}

// GetOracleRequestByID retrieves an oracle request by ID
func (r *OracleRepository) GetOracleRequestByID(ctx context.Context, id uuid.UUID) (*models.OracleRequest, error) {
	query := `
		SELECT 
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		FROM oracle_requests
		WHERE id = $1
	`

	var request models.OracleRequest
	var contextDataBytes, metadataBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&request.ID,
		&request.ProviderID,
		&request.Condition,
		&contextDataBytes,
		&request.Status,
		&request.Result,
		&request.Confidence,
		&request.Evidence,
		&metadataBytes,
		&request.VerifiedAt,
		&request.ProofHash,
		&request.RetryCount,
		&request.ErrorMessage,
		&request.CreatedAt,
		&request.UpdatedAt,
		&request.CachedUntil,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oracle request not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oracle request: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(contextDataBytes, &request.ContextData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context data: %w", err)
	}

	if err := json.Unmarshal(metadataBytes, &request.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &request, nil
}

// UpdateOracleRequest updates an existing oracle request
func (r *OracleRepository) UpdateOracleRequest(ctx context.Context, request *models.OracleRequest) error {
	query := `
		UPDATE oracle_requests SET
			provider_id = $1, condition = $2, context_data = $3, status = $4, result = $5,
			confidence = $6, evidence = $7, metadata = $8, verified_at = $9, proof_hash = $10,
			retry_count = $11, error_message = $12, updated_at = $13, cached_until = $14
		WHERE id = $15
	`

	contextDataBytes, err := json.Marshal(request.ContextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	metadataBytes, err := json.Marshal(request.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		request.ProviderID,
		request.Condition,
		contextDataBytes,
		string(request.Status),
		request.Result,
		request.Confidence,
		request.Evidence,
		metadataBytes,
		request.VerifiedAt,
		request.ProofHash,
		request.RetryCount,
		request.ErrorMessage,
		request.UpdatedAt,
		request.CachedUntil,
		request.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update oracle request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oracle request not found: %s", request.ID.String())
	}

	return nil
}

// DeleteOracleRequest deletes an oracle request
func (r *OracleRepository) DeleteOracleRequest(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oracle_requests WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oracle request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oracle request not found: %s", id.String())
	}

	return nil
}

// ListOracleRequests lists oracle requests with filtering
func (r *OracleRepository) ListOracleRequests(ctx context.Context, filter *OracleRequestFilter, limit, offset int) ([]*models.OracleRequest, error) {
	// Build query dynamically based on filter
	query := `
		SELECT 
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		FROM oracle_requests
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.ProviderID != nil {
			query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
			args = append(args, *filter.ProviderID)
			argIndex++
		}

		if filter.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(*filter.Status))
			argIndex++
		}

		if filter.Condition != nil {
			query += fmt.Sprintf(" AND condition = $%d", argIndex)
			args = append(args, *filter.Condition)
			argIndex++
		}

		if filter.DateFrom != nil {
			query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
			args = append(args, *filter.DateFrom)
			argIndex++
		}

		if filter.DateTo != nil {
			query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
			args = append(args, *filter.DateTo)
			argIndex++
		}

		if filter.MinConfidence != nil {
			query += fmt.Sprintf(" AND confidence >= $%d", argIndex)
			args = append(args, *filter.MinConfidence)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle requests: %w", err)
	}
	defer rows.Close()

	return r.scanOracleRequests(rows)
}

// GetOracleRequestsByStatus retrieves oracle requests by status
func (r *OracleRepository) GetOracleRequestsByStatus(ctx context.Context, status models.RequestStatus, limit, offset int) ([]*models.OracleRequest, error) {
	query := `
		SELECT 
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		FROM oracle_requests
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle requests: %w", err)
	}
	defer rows.Close()

	return r.scanOracleRequests(rows)
}

// GetOracleRequestsByProvider retrieves oracle requests by provider
func (r *OracleRepository) GetOracleRequestsByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*models.OracleRequest, error) {
	query := `
		SELECT 
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		FROM oracle_requests
		WHERE provider_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, providerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query oracle requests: %w", err)
	}
	defer rows.Close()

	return r.scanOracleRequests(rows)
}

// GetCachedResponse retrieves a cached response if available
func (r *OracleRepository) GetCachedResponse(ctx context.Context, condition string) (*models.OracleRequest, error) {
	query := `
		SELECT 
			id, provider_id, condition, context_data, status, result, confidence,
			evidence, metadata, verified_at, proof_hash, retry_count, error_message,
			created_at, updated_at, cached_until
		FROM oracle_requests
		WHERE condition = $1 AND status = 'cached' AND cached_until > NOW()
		ORDER BY cached_until DESC
		LIMIT 1
	`

	var request models.OracleRequest
	var contextDataBytes, metadataBytes []byte

	err := r.db.QueryRowContext(ctx, query, condition).Scan(
		&request.ID,
		&request.ProviderID,
		&request.Condition,
		&contextDataBytes,
		&request.Status,
		&request.Result,
		&request.Confidence,
		&request.Evidence,
		&metadataBytes,
		&request.VerifiedAt,
		&request.ProofHash,
		&request.RetryCount,
		&request.ErrorMessage,
		&request.CreatedAt,
		&request.UpdatedAt,
		&request.CachedUntil,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no cached response found: %w", err)
		}
		return nil, fmt.Errorf("failed to get cached response: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(contextDataBytes, &request.ContextData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context data: %w", err)
	}

	if err := json.Unmarshal(metadataBytes, &request.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &request, nil
}

// CacheResponse caches a response
func (r *OracleRepository) CacheResponse(ctx context.Context, request *models.OracleRequest) error {
	// Update the request to mark it as cached
	request.Status = models.RequestStatusCached
	request.UpdatedAt = time.Now()

	return r.UpdateOracleRequest(ctx, request)
}

// ClearExpiredCache clears expired cached responses
func (r *OracleRepository) ClearExpiredCache(ctx context.Context) error {
	query := `
		DELETE FROM oracle_requests
		WHERE status = 'cached' AND cached_until <= NOW()
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clear expired cache: %w", err)
	}

	return nil
}

// GetOracleMetrics retrieves performance metrics for an oracle provider
func (r *OracleRepository) GetOracleMetrics(ctx context.Context, providerID uuid.UUID) (*models.OracleMetrics, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_requests,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_requests,
			AVG(CASE WHEN status = 'completed' THEN EXTRACT(EPOCH FROM (verified_at - created_at)) END) as avg_response_time
		FROM oracle_requests
		WHERE provider_id = $1
	`

	var metrics models.OracleMetrics

	var avgResponseTime *float64
	err := r.db.QueryRowContext(ctx, query, providerID).Scan(
		&metrics.TotalRequests,
		&metrics.SuccessfulRequests,
		&metrics.FailedRequests,
		&avgResponseTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get oracle metrics: %w", err)
	}

	if avgResponseTime != nil {
		metrics.AverageResponseTime = time.Duration(*avgResponseTime * float64(time.Second))
	}

	// Calculate cache hit rate
	cacheQuery := `
		SELECT 
			COUNT(CASE WHEN status = 'cached' THEN 1 END) as cached_requests,
			COUNT(*) as total_requests
		FROM oracle_requests
		WHERE provider_id = $1
	`

	var cachedRequests, totalRequests int64
	err = r.db.QueryRowContext(ctx, cacheQuery, providerID).Scan(&cachedRequests, &totalRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache metrics: %w", err)
	}

	if totalRequests > 0 {
		metrics.CacheHitRate = float64(cachedRequests) / float64(totalRequests)
	}

	return &metrics, nil
}

// GetOracleReliabilityScore retrieves the reliability score for an oracle provider
func (r *OracleRepository) GetOracleReliabilityScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	query := `
		SELECT 
			COALESCE(
				COUNT(CASE WHEN status = 'completed' THEN 1 END) * 1.0 / 
				NULLIF(COUNT(CASE WHEN status IN ('completed', 'failed') THEN 1 END), 0),
				1.0
			) as reliability_score
		FROM oracle_requests
		WHERE provider_id = $1
	`

	var reliabilityScore float64
	err := r.db.QueryRowContext(ctx, query, providerID).Scan(&reliabilityScore)
	if err != nil {
		return 0, fmt.Errorf("failed to get reliability score: %w", err)
	}

	return reliabilityScore, nil
}

// GetRequestStats retrieves statistics for oracle requests
func (r *OracleRepository) GetRequestStats(ctx context.Context, providerID *uuid.UUID, startDate, endDate *time.Time) (map[models.RequestStatus]int64, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM oracle_requests
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if providerID != nil {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, *providerID)
		argIndex++
	}

	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	query += " GROUP BY status"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query request stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[models.RequestStatus]int64)

	for rows.Next() {
		var status string
		var count int64

		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request stats: %w", err)
		}

		stats[models.RequestStatus(status)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating request stats: %w", err)
	}

	return stats, nil
}
