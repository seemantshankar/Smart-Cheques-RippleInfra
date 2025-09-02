package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// AccountFraudStatusServiceInterface defines the interface for account fraud status management
type AccountFraudStatusServiceInterface interface {
	// Status management
	GetAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error)
	UpdateAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID, status models.AccountFraudStatusType, reason string, userID uuid.UUID) error
	TransitionStatus(ctx context.Context, enterpriseID uuid.UUID, newStatus models.AccountFraudStatusType, reason string, userID uuid.UUID) error

	// Restriction management
	AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error
	RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error
	GetActiveRestrictions(ctx context.Context, enterpriseID uuid.UUID) ([]models.AccountRestriction, error)
	UpdateRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType, updates map[string]interface{}) error

	// Risk assessment
	CalculateRiskScore(ctx context.Context, enterpriseID uuid.UUID) (float64, error)
	UpdateRiskFactors(ctx context.Context, enterpriseID uuid.UUID, riskFactors []string) error
	GetRiskHistory(ctx context.Context, enterpriseID uuid.UUID, days int) ([]RiskHistoryEntry, error)

	// Monitoring and review
	SetMonitoringLevel(ctx context.Context, enterpriseID uuid.UUID, level models.MonitoringLevel) error
	ScheduleReview(ctx context.Context, enterpriseID uuid.UUID, reviewDate time.Time, frequency time.Duration) error
	GetPendingReviews(ctx context.Context) ([]ReviewEntry, error)
	MarkReviewComplete(ctx context.Context, enterpriseID uuid.UUID, reviewNotes string, userID uuid.UUID) error

	// Recovery procedures
	InitiateRecovery(ctx context.Context, enterpriseID uuid.UUID, recoveryType string, userID uuid.UUID) error
	GetRecoveryStatus(ctx context.Context, enterpriseID uuid.UUID) (*RecoveryStatus, error)
	CompleteRecovery(ctx context.Context, enterpriseID uuid.UUID, outcome string, userID uuid.UUID) error

	// Reporting
	GenerateStatusReport(ctx context.Context, enterpriseID uuid.UUID) (*AccountStatusReport, error)
	GetStatusHistory(ctx context.Context, enterpriseID uuid.UUID, limit int) ([]models.FraudStatusChange, error)
}

// AccountFraudStatusService implements account fraud status management
type AccountFraudStatusService struct {
	fraudRepo       repository.FraudRepositoryInterface
	enterpriseRepo  repository.EnterpriseRepositoryInterface
	messagingClient messaging.EventBus
	config          *AccountFraudStatusConfig
}

// NewAccountFraudStatusService creates a new account fraud status service
func NewAccountFraudStatusService(
	fraudRepo repository.FraudRepositoryInterface,
	enterpriseRepo repository.EnterpriseRepositoryInterface,
	messagingClient messaging.EventBus,
	config *AccountFraudStatusConfig,
) AccountFraudStatusServiceInterface {
	return &AccountFraudStatusService{
		fraudRepo:       fraudRepo,
		enterpriseRepo:  enterpriseRepo,
		messagingClient: messagingClient,
		config:          config,
	}
}

// Configuration types
type AccountFraudStatusConfig struct {
	// Risk scoring
	BaseRiskScore      float64 `json:"base_risk_score"`       // e.g., 0.1
	RiskScoreDecayRate float64 `json:"risk_score_decay_rate"` // e.g., 0.05 per day
	MaxRiskScore       float64 `json:"max_risk_score"`        // e.g., 1.0

	// Review scheduling
	DefaultReviewFrequency time.Duration `json:"default_review_frequency"` // e.g., 30 days
	UrgentReviewFrequency  time.Duration `json:"urgent_review_frequency"`  // e.g., 7 days
	MaxReviewDelay         time.Duration `json:"max_review_delay"`         // e.g., 90 days

	// Recovery settings
	RecoveryTimeLimit     time.Duration `json:"recovery_time_limit"`     // e.g., 30 days
	AutoRecoveryThreshold float64       `json:"auto_recovery_threshold"` // e.g., 0.3
}

// Supporting types
type RiskHistoryEntry struct {
	Date      time.Time             `json:"date"`
	RiskScore float64               `json:"risk_score"`
	RiskLevel models.FraudRiskLevel `json:"risk_level"`
	Factors   []string              `json:"factors"`
}

type ReviewEntry struct {
	EnterpriseID   uuid.UUID  `json:"enterprise_id"`
	ReviewDate     time.Time  `json:"review_date"`
	Status         string     `json:"status"` // "pending", "overdue", "completed"
	LastReviewDate *time.Time `json:"last_review_date,omitempty"`
	RiskScore      float64    `json:"risk_score"`
}

type RecoveryStatus struct {
	EnterpriseID uuid.UUID  `json:"enterprise_id"`
	RecoveryType string     `json:"recovery_type"`
	Status       string     `json:"status"` // "initiated", "in_progress", "completed", "failed"
	InitiatedAt  time.Time  `json:"initiated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	InitiatedBy  uuid.UUID  `json:"initiated_by"`
	CompletedBy  *uuid.UUID `json:"completed_by,omitempty"`
	Outcome      string     `json:"outcome"`
	Notes        string     `json:"notes"`
}

type AccountStatusReport struct {
	EnterpriseID       uuid.UUID                     `json:"enterprise_id"`
	GeneratedAt        time.Time                     `json:"generated_at"`
	CurrentStatus      models.AccountFraudStatusType `json:"current_status"`
	CurrentRiskScore   float64                       `json:"current_risk_score"`
	CurrentRiskLevel   models.FraudRiskLevel         `json:"current_risk_level"`
	ActiveRestrictions int                           `json:"active_restrictions"`
	LastStatusChange   time.Time                     `json:"last_status_change"`
	NextReviewDate     time.Time                     `json:"next_review_date"`
	StatusHistory      []models.FraudStatusChange    `json:"status_history"`
	RiskTrend          string                        `json:"risk_trend"` // "increasing", "decreasing", "stable"
	Recommendations    []string                      `json:"recommendations"`
}

// Implementation methods

// GetAccountFraudStatus retrieves the current fraud status for an enterprise
func (s *AccountFraudStatusService) GetAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error) {
	status, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account fraud status: %w", err)
	}

	// Calculate current risk score if not recent
	if time.Since(status.UpdatedAt) > 24*time.Hour {
		riskScore, err := s.CalculateRiskScore(ctx, enterpriseID)
		if err == nil {
			status.RiskScore = riskScore
			status.RiskLevel = s.calculateRiskLevel(riskScore)
		}
	}

	return status, nil
}

// UpdateAccountFraudStatus updates the fraud status for an enterprise
func (s *AccountFraudStatusService) UpdateAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID, status models.AccountFraudStatusType, reason string, userID uuid.UUID) error {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get current account status: %w", err)
	}

	// Validate status transition
	if err := accountStatus.CanTransition(status); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Perform transition
	if err := accountStatus.TransitionTo(status, reason, userID); err != nil {
		return fmt.Errorf("failed to transition status: %w", err)
	}

	// Update in database
	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, accountStatus); err != nil {
		return fmt.Errorf("failed to update account fraud status: %w", err)
	}

	// Publish status change event
	event := &messaging.Event{
		Type: "account.fraud_status.changed",
		Data: map[string]interface{}{
			"enterprise_id": enterpriseID,
			"old_status":    accountStatus.StatusHistory[len(accountStatus.StatusHistory)-2].FromStatus,
			"new_status":    status,
			"reason":        reason,
			"changed_by":    userID,
			"changed_at":    time.Now(),
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish status change event: %v\n", err)
	}

	return nil
}

// TransitionStatus is an alias for UpdateAccountFraudStatus for clarity
func (s *AccountFraudStatusService) TransitionStatus(ctx context.Context, enterpriseID uuid.UUID, newStatus models.AccountFraudStatusType, reason string, userID uuid.UUID) error {
	return s.UpdateAccountFraudStatus(ctx, enterpriseID, newStatus, reason, userID)
}

// AddAccountRestriction adds a restriction to an enterprise account
func (s *AccountFraudStatusService) AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error {
	// Set default effective time if not provided
	if restriction.EffectiveAt.IsZero() {
		restriction.EffectiveAt = time.Now()
	}

	// Validate restriction parameters
	if err := s.validateRestriction(restriction); err != nil {
		return fmt.Errorf("invalid restriction: %w", err)
	}

	// TODO: Implement AddAccountRestriction in repository
	// For now, just publish an event
	event := &messaging.Event{
		Type: "account.restriction.added",
		Data: map[string]interface{}{
			"enterprise_id": enterpriseID,
			"restriction":   restriction,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish restriction event: %v\n", err)
	}

	return nil
}

// RemoveAccountRestriction removes a restriction from an enterprise account
func (s *AccountFraudStatusService) RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error {
	if err := s.fraudRepo.RemoveAccountRestriction(ctx, enterpriseID, restrictionType); err != nil {
		return fmt.Errorf("failed to remove account restriction: %w", err)
	}

	// Publish restriction removal event
	event := &messaging.Event{
		Type: "account.restriction.removed",
		Data: map[string]interface{}{
			"enterprise_id":    enterpriseID,
			"restriction_type": restrictionType,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish restriction removal event: %v\n", err)
	}

	return nil
}

// GetActiveRestrictions retrieves all active restrictions for an enterprise
func (s *AccountFraudStatusService) GetActiveRestrictions(ctx context.Context, enterpriseID uuid.UUID) ([]models.AccountRestriction, error) {
	restrictions, err := s.fraudRepo.GetAccountRestrictions(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account restrictions: %w", err)
	}

	// Filter active restrictions
	var active []models.AccountRestriction
	now := time.Now()

	for _, restriction := range restrictions {
		if restriction.EffectiveAt.Before(now) &&
			(restriction.ExpiresAt == nil || restriction.ExpiresAt.After(now)) {
			active = append(active, restriction)
		}
	}

	return active, nil
}

// UpdateRestriction updates an existing restriction
func (s *AccountFraudStatusService) UpdateRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType, updates map[string]interface{}) error {
	if err := s.fraudRepo.UpdateAccountRestriction(ctx, enterpriseID, restrictionType, updates); err != nil {
		return fmt.Errorf("failed to update account restriction: %w", err)
	}

	// Publish restriction update event
	event := &messaging.Event{
		Type: "account.restriction.updated",
		Data: map[string]interface{}{
			"enterprise_id":    enterpriseID,
			"restriction_type": restrictionType,
			"updates":          updates,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Failed to publish restriction update event: %v\n", err)
	}

	return nil
}

// CalculateRiskScore calculates the current risk score for an enterprise
func (s *AccountFraudStatusService) CalculateRiskScore(ctx context.Context, enterpriseID uuid.UUID) (float64, error) {
	// Get recent fraud alerts
	startDate := time.Now().Add(-30 * 24 * time.Hour)
	alerts, err := s.fraudRepo.ListFraudAlerts(ctx, &repository.FraudAlertFilter{
		EnterpriseID: &enterpriseID,
		StartDate:    &startDate,
	}, 100, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent alerts: %w", err)
	}

	// Calculate base risk score
	riskScore := s.config.BaseRiskScore

	// Add risk from alerts
	for _, alert := range alerts {
		severityWeight := s.getSeverityWeight(alert.Severity)
		timeDecay := s.calculateTimeDecay(alert.DetectedAt)
		riskScore += alert.Score * severityWeight * timeDecay
	}

	// Apply decay rate
	daysSinceLastActivity := s.getDaysSinceLastActivity(ctx, enterpriseID)
	riskScore *= math.Pow(1-s.config.RiskScoreDecayRate, float64(daysSinceLastActivity))

	// Cap at maximum
	if riskScore > s.config.MaxRiskScore {
		riskScore = s.config.MaxRiskScore
	}

	return riskScore, nil
}

// UpdateRiskFactors updates the risk factors for an enterprise
func (s *AccountFraudStatusService) UpdateRiskFactors(ctx context.Context, enterpriseID uuid.UUID, riskFactors []string) error {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get account fraud status: %w", err)
	}

	accountStatus.RiskFactors = riskFactors
	accountStatus.UpdatedAt = time.Now()

	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, accountStatus); err != nil {
		return fmt.Errorf("failed to update risk factors: %w", err)
	}

	return nil
}

// GetRiskHistory retrieves risk history for an enterprise
func (s *AccountFraudStatusService) GetRiskHistory(ctx context.Context, enterpriseID uuid.UUID, days int) ([]RiskHistoryEntry, error) {
	// This would typically query a time-series database or risk history table
	// For now, we'll return a simplified implementation
	entries := make([]RiskHistoryEntry, 0)

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		riskScore, err := s.CalculateRiskScore(ctx, enterpriseID)
		if err != nil {
			continue
		}

		entries = append(entries, RiskHistoryEntry{
			Date:      date,
			RiskScore: riskScore,
			RiskLevel: s.calculateRiskLevel(riskScore),
			Factors:   []string{}, // Would be populated from historical data
		})
	}

	return entries, nil
}

// SetMonitoringLevel sets the monitoring level for an enterprise
func (s *AccountFraudStatusService) SetMonitoringLevel(ctx context.Context, enterpriseID uuid.UUID, level models.MonitoringLevel) error {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get account fraud status: %w", err)
	}

	accountStatus.MonitoringLevel = level
	accountStatus.UpdatedAt = time.Now()

	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, accountStatus); err != nil {
		return fmt.Errorf("failed to update monitoring level: %w", err)
	}

	return nil
}

// ScheduleReview schedules a review for an enterprise
func (s *AccountFraudStatusService) ScheduleReview(ctx context.Context, enterpriseID uuid.UUID, reviewDate time.Time, frequency time.Duration) error {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get account fraud status: %w", err)
	}

	accountStatus.NextReviewDate = reviewDate
	accountStatus.ReviewFrequency = frequency
	accountStatus.UpdatedAt = time.Now()

	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, accountStatus); err != nil {
		return fmt.Errorf("failed to schedule review: %w", err)
	}

	return nil
}

// GetPendingReviews retrieves all pending reviews
func (s *AccountFraudStatusService) GetPendingReviews(ctx context.Context) ([]ReviewEntry, error) {
	// This would query for accounts with overdue or upcoming reviews
	// For now, return empty list
	return []ReviewEntry{}, nil
}

// MarkReviewComplete marks a review as complete
func (s *AccountFraudStatusService) MarkReviewComplete(ctx context.Context, enterpriseID uuid.UUID, reviewNotes string, userID uuid.UUID) error {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get account fraud status: %w", err)
	}

	// Schedule next review
	nextReview := time.Now().Add(accountStatus.ReviewFrequency)
	accountStatus.NextReviewDate = nextReview
	accountStatus.UpdatedAt = time.Now()

	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, accountStatus); err != nil {
		return fmt.Errorf("failed to mark review complete: %w", err)
	}

	return nil
}

// InitiateRecovery initiates a recovery procedure for an enterprise
func (s *AccountFraudStatusService) InitiateRecovery(ctx context.Context, enterpriseID uuid.UUID, recoveryType string, userID uuid.UUID) error {
	recoveryStatus := &RecoveryStatus{
		EnterpriseID: enterpriseID,
		RecoveryType: recoveryType,
		Status:       "initiated",
		InitiatedAt:  time.Now(),
		InitiatedBy:  userID,
	}

	// Store recovery status (would need a recovery_status table)
	// For now, just publish an event

	event := messaging.Event{
		Type: "account.recovery.initiated",
		Data: map[string]interface{}{
			"recovery_status": recoveryStatus,
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, &event); err != nil {
		return fmt.Errorf("failed to publish recovery initiation event: %w", err)
	}

	return nil
}

// GetRecoveryStatus retrieves the recovery status for an enterprise
func (s *AccountFraudStatusService) GetRecoveryStatus(ctx context.Context, enterpriseID uuid.UUID) (*RecoveryStatus, error) {
	// This would query a recovery_status table
	// For now, return nil (no active recovery)
	return nil, nil
}

// CompleteRecovery completes a recovery procedure
func (s *AccountFraudStatusService) CompleteRecovery(ctx context.Context, enterpriseID uuid.UUID, outcome string, userID uuid.UUID) error {
	event := messaging.Event{
		Type: "account.recovery.completed",
		Data: map[string]interface{}{
			"enterprise_id": enterpriseID,
			"outcome":       outcome,
			"completed_by":  userID,
			"completed_at":  time.Now(),
		},
	}

	if err := s.messagingClient.PublishEvent(ctx, &event); err != nil {
		return fmt.Errorf("failed to publish recovery completion event: %w", err)
	}

	return nil
}

// GenerateStatusReport generates a comprehensive status report for an enterprise
func (s *AccountFraudStatusService) GenerateStatusReport(ctx context.Context, enterpriseID uuid.UUID) (*AccountStatusReport, error) {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account fraud status: %w", err)
	}

	activeRestrictions := len(accountStatus.GetActiveRestrictions())

	// Calculate risk trend
	riskHistory, err := s.GetRiskHistory(ctx, enterpriseID, 7)
	var riskTrend string
	if err == nil && len(riskHistory) >= 2 {
		if riskHistory[0].RiskScore > riskHistory[len(riskHistory)-1].RiskScore {
			riskTrend = RiskTrendIncreasing
		} else if riskHistory[0].RiskScore < riskHistory[len(riskHistory)-1].RiskScore {
			riskTrend = RiskTrendDecreasing
		} else {
			riskTrend = RiskTrendStable
		}
	} else {
		riskTrend = RiskTrendUnknown
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(accountStatus)

	report := &AccountStatusReport{
		EnterpriseID:       enterpriseID,
		GeneratedAt:        time.Now(),
		CurrentStatus:      accountStatus.Status,
		CurrentRiskScore:   accountStatus.RiskScore,
		CurrentRiskLevel:   accountStatus.RiskLevel,
		ActiveRestrictions: activeRestrictions,
		LastStatusChange:   accountStatus.StatusChangedAt,
		NextReviewDate:     accountStatus.NextReviewDate,
		StatusHistory:      accountStatus.StatusHistory,
		RiskTrend:          riskTrend,
		Recommendations:    recommendations,
	}

	return report, nil
}

// GetStatusHistory retrieves the status change history for an enterprise
func (s *AccountFraudStatusService) GetStatusHistory(ctx context.Context, enterpriseID uuid.UUID, limit int) ([]models.FraudStatusChange, error) {
	accountStatus, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account fraud status: %w", err)
	}

	if limit > 0 && len(accountStatus.StatusHistory) > limit {
		return accountStatus.StatusHistory[len(accountStatus.StatusHistory)-limit:], nil
	}

	return accountStatus.StatusHistory, nil
}

// Helper methods

func (s *AccountFraudStatusService) validateRestriction(restriction *models.AccountRestriction) error {
	if restriction.Type == "" {
		return fmt.Errorf("restriction type is required")
	}

	if restriction.Description == "" {
		return fmt.Errorf("restriction description is required")
	}

	return nil
}

func (s *AccountFraudStatusService) getSeverityWeight(severity models.FraudSeverity) float64 {
	switch severity {
	case models.FraudSeverityLow:
		return 0.25
	case models.FraudSeverityMedium:
		return 0.5
	case models.FraudSeverityHigh:
		return 0.75
	case models.FraudSeverityCritical:
		return 1.0
	default:
		return 0.5
	}
}

func (s *AccountFraudStatusService) calculateTimeDecay(detectedAt time.Time) float64 {
	daysSince := time.Since(detectedAt).Hours() / 24
	return math.Exp(-daysSince / 30) // Decay over 30 days
}

func (s *AccountFraudStatusService) calculateRiskLevel(riskScore float64) models.FraudRiskLevel {
	switch {
	case riskScore < 0.25:
		return models.FraudRiskLevelLow
	case riskScore < 0.5:
		return models.FraudRiskLevelMedium
	case riskScore < 0.75:
		return models.FraudRiskLevelHigh
	default:
		return models.FraudRiskLevelCritical
	}
}

func (s *AccountFraudStatusService) getDaysSinceLastActivity(_ context.Context, _ uuid.UUID) int {
	// This would query transaction history to find the last activity
	// For now, return 0 (assume recent activity)
	return 0
}

func (s *AccountFraudStatusService) generateRecommendations(status *models.AccountFraudStatus) []string {
	var recommendations []string

	if status.RiskScore > 0.7 {
		recommendations = append(recommendations, "Consider increasing monitoring frequency")
		recommendations = append(recommendations, "Review recent transactions for suspicious patterns")
	}

	if status.Status == models.AccountFraudStatusRestricted {
		recommendations = append(recommendations, "Monitor restriction effectiveness")
		recommendations = append(recommendations, "Consider additional verification measures")
	}

	if len(status.GetActiveRestrictions()) == 0 && status.RiskScore > 0.5 {
		recommendations = append(recommendations, "Consider implementing transaction limits")
	}

	if status.NextReviewDate.Before(time.Now().Add(7 * 24 * time.Hour)) {
		recommendations = append(recommendations, "Schedule account review soon")
	}

	return recommendations
}
