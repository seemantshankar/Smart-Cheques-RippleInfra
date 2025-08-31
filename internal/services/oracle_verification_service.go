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

// OracleVerificationService handles milestone verification using oracle providers
type OracleVerificationService struct {
	oracleService   *OracleService
	oracleRepo      repository.OracleRepositoryInterface
	messagingClient *messaging.Service
}

// NewOracleVerificationService creates a new oracle verification service
func NewOracleVerificationService(
	oracleService *OracleService,
	oracleRepo repository.OracleRepositoryInterface,
	messagingClient *messaging.Service,
) *OracleVerificationService {
	return &OracleVerificationService{
		oracleService:   oracleService,
		oracleRepo:      oracleRepo,
		messagingClient: messagingClient,
	}
}

// VerifyMilestone evaluates a milestone condition using the appropriate oracle provider
func (s *OracleVerificationService) VerifyMilestone(ctx context.Context, milestoneID string, condition string, oracleConfig *models.OracleConfig) (*models.OracleResponse, error) {
	if oracleConfig == nil {
		return nil, fmt.Errorf("oracle configuration is required for verification")
	}

	// Check cache first
	cachedRequest, err := s.oracleService.GetCachedResponse(ctx, condition)
	if err == nil && cachedRequest != nil {
		response := &models.OracleResponse{
			RequestID:  cachedRequest.ID,
			Condition:  cachedRequest.Condition,
			Result:     *cachedRequest.Result,
			Confidence: *cachedRequest.Confidence,
			Evidence:   cachedRequest.Evidence,
			Metadata:   cachedRequest.Metadata,
			VerifiedAt: *cachedRequest.VerifiedAt,
			ProofHash:  *cachedRequest.ProofHash,
		}
		return response, nil
	}

	// Find appropriate oracle provider
	providers, err := s.oracleService.GetProvidersByType(ctx, models.OracleType(oracleConfig.Type))
	if err != nil {
		return nil, fmt.Errorf("failed to get oracle providers: %w", err)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no oracle providers found for type: %s", oracleConfig.Type)
	}

	// Select the best provider based on reliability
	bestProvider := s.selectBestProvider(providers)

	// Create oracle implementation based on provider type
	var oracle OracleInterface
	switch bestProvider.Type {
	case models.OracleTypeAPI:
		oracle = NewAPIOracle(bestProvider)
	case models.OracleTypeWebhook:
		oracle = NewWebhookOracle(bestProvider)
	case models.OracleTypeManual:
		oracle = NewManualOracle(bestProvider)
	default:
		return nil, fmt.Errorf("unsupported oracle type: %s", bestProvider.Type)
	}

	// Perform verification
	response, err := oracle.Verify(ctx, condition, oracleConfig.Config)
	if err != nil {
		// Log the error but don't fail completely - create a failed request record
		log.Printf("Oracle verification failed for milestone %s: %v", milestoneID, err)

		// Create failed request record
		errStr := err.Error()
		failedRequest := &models.OracleRequest{
			ID:           uuid.New(),
			ProviderID:   bestProvider.ID,
			Condition:    condition,
			ContextData:  oracleConfig.Config,
			Status:       models.RequestStatusFailed,
			ErrorMessage: &errStr,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.oracleRepo.CreateOracleRequest(ctx, failedRequest); err != nil {
			log.Printf("Warning: Failed to record failed oracle request: %v", err)
		}

		return nil, fmt.Errorf("oracle verification failed: %w", err)
	}

	// Create successful request record
	request := &models.OracleRequest{
		ID:          uuid.New(),
		ProviderID:  bestProvider.ID,
		Condition:   condition,
		ContextData: oracleConfig.Config,
		Status:      models.RequestStatusCompleted,
		Result:      &response.Result,
		Confidence:  &response.Confidence,
		Evidence:    response.Evidence,
		Metadata:    response.Metadata,
		VerifiedAt:  &response.VerifiedAt,
		ProofHash:   &response.ProofHash,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.oracleRepo.CreateOracleRequest(ctx, request); err != nil {
		log.Printf("Warning: Failed to record successful oracle request: %v", err)
	}

	// Cache the response for future use
	if err := s.oracleService.CacheResponse(ctx, request); err != nil {
		log.Printf("Warning: Failed to cache oracle response: %v", err)
	}

	// Publish verification event
	event := &messaging.Event{
		Type:   "milestone_verified",
		Source: "oracle_verification_service",
		Data: map[string]interface{}{
			"event":        "milestone_verified",
			"milestone_id": milestoneID,
			"condition":    condition,
			"result":       response.Result,
			"confidence":   response.Confidence,
			"provider_id":  bestProvider.ID.String(),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(event); err != nil {
		log.Printf("Warning: Failed to publish milestone verification event: %v", err)
	}

	return response, nil
}

// GetVerificationResult retrieves the result of a previous verification
func (s *OracleVerificationService) GetVerificationResult(ctx context.Context, requestID uuid.UUID) (*models.OracleResponse, error) {
	request, err := s.oracleService.GetRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification result: %w", err)
	}

	if request.Status != models.RequestStatusCompleted {
		return nil, fmt.Errorf("verification is not yet completed, current status: %s", request.Status)
	}

	response := &models.OracleResponse{
		RequestID:  request.ID,
		Condition:  request.Condition,
		Result:     *request.Result,
		Confidence: *request.Confidence,
		Evidence:   request.Evidence,
		Metadata:   request.Metadata,
		VerifiedAt: *request.VerifiedAt,
		ProofHash:  *request.ProofHash,
	}

	return response, nil
}

// GetProof retrieves verification evidence for a completed verification
func (s *OracleVerificationService) GetProof(ctx context.Context, requestID uuid.UUID) ([]byte, error) {
	request, err := s.oracleService.GetRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification proof: %w", err)
	}

	return request.Evidence, nil
}

// selectBestProvider selects the best oracle provider based on reliability and performance
func (s *OracleVerificationService) selectBestProvider(providers []*models.OracleProvider) *models.OracleProvider {
	if len(providers) == 0 {
		return nil
	}

	// For now, simply select the provider with the highest reliability score
	// In a more sophisticated implementation, this could consider response time,
	// error rates, and other performance metrics
	bestProvider := providers[0]
	for _, provider := range providers {
		if provider.Reliability > bestProvider.Reliability {
			bestProvider = provider
		}
	}

	return bestProvider
}

// handleWebhookCallback handles incoming webhook callbacks from oracle providers
/*
func (s *OracleVerificationService) handleWebhookCallback(ctx context.Context, payload map[string]interface{}) error {
	// Extract verification result from webhook payload
	requestID, ok := payload["request_id"].(string)
	if !ok {
		return fmt.Errorf("missing request_id in webhook payload")
	}

	result, ok := payload["result"].(bool)
	if !ok {
		return fmt.Errorf("missing result in webhook payload")
	}

	// Update the oracle request with the verification result
	requestUUID, err := uuid.Parse(requestID)
	if err != nil {
		return fmt.Errorf("invalid request ID format: %w", err)
	}

	request, err := s.oracleService.GetRequest(ctx, requestUUID)
	if err != nil {
		return fmt.Errorf("failed to get oracle request: %w", err)
	}

	// Update request status and result
	request.Status = models.RequestStatusCompleted
	request.Result = &result
	now := time.Now()
	request.VerifiedAt = &now
	request.UpdatedAt = time.Now()

	if err := s.oracleRepo.UpdateOracleRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to update oracle request: %w", err)
	}

	// Cache the response
	if err := s.oracleService.CacheResponse(ctx, request); err != nil {
		log.Printf("Warning: Failed to cache oracle response: %v", err)
	}

	// Publish verification completion event
	event := &messaging.Event{
		Type:   "oracle_verification_completed",
		Source: "oracle_verification_service",
		Data: map[string]interface{}{
			"event":      "oracle_verification_completed",
			"request_id": requestID,
			"result":     result,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(event); err != nil {
		log.Printf("Warning: Failed to publish verification completion event: %v", err)
	}

	return nil
}
*/

// GetDashboardMetrics retrieves dashboard metrics for oracle verification
func (s *OracleVerificationService) GetDashboardMetrics(_ context.Context) (map[string]interface{}, error) {
	// This is a placeholder implementation
	// In a real implementation, this would query the database for metrics
	metrics := map[string]interface{}{
		"total_verifications":      0,
		"successful_verifications": 0,
		"failed_verifications":     0,
		"pending_verifications":    0,
		"average_response_time":    0.0,
	}

	return metrics, nil
}
