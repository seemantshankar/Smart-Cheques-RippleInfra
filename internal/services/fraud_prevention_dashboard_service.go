package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// FraudPreventionDashboardService provides dashboard and monitoring capabilities for fraud prevention
type FraudPreventionDashboardService struct {
	fraudRepo       repository.FraudRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	enterpriseRepo  repository.EnterpriseRepositoryInterface
	messagingClient messaging.EventBus
	config          *FraudPreventionDashboardConfig
}

// FraudPreventionDashboardConfig defines configuration for the fraud prevention dashboard
type FraudPreventionDashboardConfig struct {
	// Dashboard refresh intervals
	RealTimeRefreshInterval time.Duration `json:"real_time_refresh_interval"` // e.g., 30 seconds
	MetricsRefreshInterval  time.Duration `json:"metrics_refresh_interval"`   // e.g., 5 minutes

	// Alert thresholds
	HighRiskThreshold     float64 `json:"high_risk_threshold"`     // e.g., 0.7
	CriticalRiskThreshold float64 `json:"critical_risk_threshold"` // e.g., 0.9

	// Dashboard limits
	MaxRecentAlerts       int `json:"max_recent_alerts"`       // e.g., 100
	MaxRecentCases        int `json:"max_recent_cases"`        // e.g., 50
	MaxRecentTransactions int `json:"max_recent_transactions"` // e.g., 1000

	// Time windows for analytics
	AnalyticsWindow time.Duration `json:"analytics_window"` // e.g., 24 hours
	TrendWindow     time.Duration `json:"trend_window"`     // e.g., 7 days
}

// NewFraudPreventionDashboardService creates a new fraud prevention dashboard service
func NewFraudPreventionDashboardService(
	fraudRepo repository.FraudRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	enterpriseRepo repository.EnterpriseRepositoryInterface,
	messagingClient messaging.EventBus,
	config *FraudPreventionDashboardConfig,
) *FraudPreventionDashboardService {
	if config == nil {
		config = &FraudPreventionDashboardConfig{
			RealTimeRefreshInterval: 30 * time.Second,
			MetricsRefreshInterval:  5 * time.Minute,
			HighRiskThreshold:       0.7,
			CriticalRiskThreshold:   0.9,
			MaxRecentAlerts:         100,
			MaxRecentCases:          50,
			MaxRecentTransactions:   1000,
			AnalyticsWindow:         24 * time.Hour,
			TrendWindow:             7 * 24 * time.Hour,
		}
	}

	return &FraudPreventionDashboardService{
		fraudRepo:       fraudRepo,
		transactionRepo: transactionRepo,
		enterpriseRepo:  enterpriseRepo,
		messagingClient: messagingClient,
		config:          config,
	}
}

// DashboardOverview provides a comprehensive overview of fraud prevention status
func (s *FraudPreventionDashboardService) DashboardOverview(ctx context.Context, enterpriseID *uuid.UUID) (*FraudDashboardOverview, error) {
	overview := &FraudDashboardOverview{
		GeneratedAt:  time.Now(),
		EnterpriseID: enterpriseID,
	}

	// Get real-time metrics
	metrics, err := s.GetRealTimeMetrics(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}
	overview.RealTimeMetrics = metrics

	// Get recent alerts
	alerts, err := s.GetRecentAlerts(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}
	overview.RecentAlerts = alerts

	// Get recent cases
	cases, err := s.GetRecentCases(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent cases: %w", err)
	}
	overview.RecentCases = cases

	// Get risk trends
	trends, err := s.GetRiskTrends(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk trends: %w", err)
	}
	overview.RiskTrends = trends

	// Get enterprise status
	if enterpriseID != nil {
		status, err := s.GetEnterpriseFraudStatus(ctx, *enterpriseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get enterprise fraud status: %w", err)
		}
		overview.EnterpriseStatus = status
	}

	return overview, nil
}

// GetRealTimeMetrics provides real-time fraud prevention metrics
func (s *FraudPreventionDashboardService) GetRealTimeMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*FraudRealTimeMetrics, error) {
	metrics := &FraudRealTimeMetrics{
		Timestamp: time.Now(),
	}

	// Get transaction counts
	transactionCounts, err := s.getTransactionCounts(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction counts: %w", err)
	}
	metrics.TransactionCounts = transactionCounts

	// Get alert counts
	alertCounts, err := s.getAlertCounts(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert counts: %w", err)
	}
	metrics.AlertCounts = alertCounts

	// Get case counts
	caseCounts, err := s.getCaseCounts(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get case counts: %w", err)
	}
	metrics.CaseCounts = caseCounts

	// Get risk score distribution
	riskDistribution, err := s.getRiskScoreDistribution(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score distribution: %w", err)
	}
	metrics.RiskScoreDistribution = riskDistribution

	return metrics, nil
}

// GetRecentAlerts gets recent fraud alerts
func (s *FraudPreventionDashboardService) GetRecentAlerts(ctx context.Context, enterpriseID *uuid.UUID) ([]*FraudAlertSummary, error) {
	filter := &repository.FraudAlertFilter{
		EnterpriseID: enterpriseID,
	}

	alerts, err := s.fraudRepo.ListFraudAlerts(ctx, filter, s.config.MaxRecentAlerts, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %w", err)
	}

	// Convert to summary format
	summaries := make([]*FraudAlertSummary, len(alerts))
	for i, alert := range alerts {
		summaries[i] = &FraudAlertSummary{
			ID:           alert.ID,
			AlertType:    alert.AlertType,
			Severity:     alert.Severity,
			Status:       alert.Status,
			Score:        alert.Score,
			Title:        alert.Title,
			DetectedAt:   alert.DetectedAt,
			EnterpriseID: alert.EnterpriseID,
		}
	}

	return summaries, nil
}

// GetRecentCases gets recent fraud cases
func (s *FraudPreventionDashboardService) GetRecentCases(ctx context.Context, enterpriseID *uuid.UUID) ([]*FraudCaseSummary, error) {
	filter := &repository.FraudCaseFilter{
		EnterpriseID: enterpriseID,
	}

	cases, err := s.fraudRepo.ListFraudCases(ctx, filter, s.config.MaxRecentCases, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent cases: %w", err)
	}

	// Convert to summary format
	summaries := make([]*FraudCaseSummary, len(cases))
	for i, caseItem := range cases {
		summaries[i] = &FraudCaseSummary{
			ID:           caseItem.ID,
			Title:        caseItem.Title,
			Category:     caseItem.Category,
			Priority:     caseItem.Priority,
			Status:       caseItem.Status,
			CreatedAt:    caseItem.CreatedAt,
			EnterpriseID: caseItem.EnterpriseID,
		}
	}

	return summaries, nil
}

// GetRiskTrends provides risk trend analysis
func (s *FraudPreventionDashboardService) GetRiskTrends(ctx context.Context, enterpriseID *uuid.UUID) (*FraudRiskTrends, error) {
	trends := &FraudRiskTrends{
		GeneratedAt: time.Now(),
		Window:      s.config.TrendWindow,
	}

	// Get risk score trends over time
	riskTrends, err := s.getRiskScoreTrends(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score trends: %w", err)
	}
	trends.RiskScoreTrends = riskTrends

	// Get alert frequency trends
	alertTrends, err := s.getAlertFrequencyTrends(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert frequency trends: %w", err)
	}
	trends.AlertFrequencyTrends = alertTrends

	// Get transaction volume trends
	transactionTrends, err := s.getTransactionVolumeTrends(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction volume trends: %w", err)
	}
	trends.TransactionVolumeTrends = transactionTrends

	return trends, nil
}

// GetEnterpriseFraudStatus gets the fraud status for a specific enterprise
func (s *FraudPreventionDashboardService) GetEnterpriseFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*FraudEnterpriseStatus, error) {
	status, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise fraud status: %w", err)
	}

	return &FraudEnterpriseStatus{
		EnterpriseID:    enterpriseID,
		Status:          status.Status,
		RiskScore:       status.RiskScore,
		Restrictions:    status.GetActiveRestrictions(),
		LastUpdated:     status.UpdatedAt,
		StatusChangedAt: status.StatusChangedAt,
	}, nil
}

// GetFraudPreventionConfig gets the current fraud prevention configuration
func (s *FraudPreventionDashboardService) GetFraudPreventionConfig(ctx context.Context) (*FraudPreventionConfig, error) {
	// Get active rules
	rules, err := s.fraudRepo.GetActiveFraudRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active rules: %w", err)
	}

	// Get system-wide settings
	settings, err := s.getSystemSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system settings: %w", err)
	}

	return &FraudPreventionConfig{
		Rules:     rules,
		Settings:  settings,
		UpdatedAt: time.Now(),
	}, nil
}

// UpdateFraudPreventionConfig updates the fraud prevention configuration
func (s *FraudPreventionDashboardService) UpdateFraudPreventionConfig(ctx context.Context, config *FraudPreventionConfig) error {
	// Update rules
	for _, rule := range config.Rules {
		if err := s.fraudRepo.UpdateFraudRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to update rule %s: %w", rule.ID, err)
		}
	}

	// Update system settings
	if err := s.updateSystemSettings(ctx, config.Settings); err != nil {
		return fmt.Errorf("failed to update system settings: %w", err)
	}

	// Publish configuration update event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "fraud_prevention.config_updated",
			Source: "fraud-prevention-dashboard",
			Data: map[string]interface{}{
				"updated_at":  time.Now(),
				"rules_count": len(config.Rules),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(ctx, event)
	}

	return nil
}

// Helper methods for data aggregation

func (s *FraudPreventionDashboardService) getTransactionCounts(ctx context.Context, enterpriseID *uuid.UUID) (*TransactionCounts, error) {
	// This would typically query the transaction repository
	// For now, return mock data
	return &TransactionCounts{
		Total:        1000,
		HighRisk:     50,
		FraudFlagged: 10,
		Blocked:      5,
		InProgress:   100,
		Completed:    800,
		Failed:       45,
	}, nil
}

func (s *FraudPreventionDashboardService) getAlertCounts(ctx context.Context, enterpriseID *uuid.UUID) (*AlertCounts, error) {
	// This would typically query the fraud repository
	// For now, return mock data
	return &AlertCounts{
		Total:         100,
		New:           20,
		Acknowledged:  30,
		Investigating: 25,
		Resolved:      20,
		FalsePositive: 5,
		BySeverity: map[string]int{
			"low":      10,
			"medium":   30,
			"high":     40,
			"critical": 20,
		},
	}, nil
}

func (s *FraudPreventionDashboardService) getCaseCounts(ctx context.Context, enterpriseID *uuid.UUID) (*CaseCounts, error) {
	// This would typically query the fraud repository
	// For now, return mock data
	return &CaseCounts{
		Total:         50,
		Open:          15,
		Investigating: 20,
		Resolved:      10,
		Closed:        5,
		ByPriority: map[string]int{
			"low":      10,
			"medium":   20,
			"high":     15,
			"critical": 5,
		},
	}, nil
}

func (s *FraudPreventionDashboardService) getRiskScoreDistribution(ctx context.Context, enterpriseID *uuid.UUID) (*RiskScoreDistribution, error) {
	// This would typically calculate from transaction data
	// For now, return mock data
	return &RiskScoreDistribution{
		Low:      600,
		Medium:   250,
		High:     100,
		Critical: 50,
		Average:  0.35,
		Median:   0.30,
	}, nil
}

func (s *FraudPreventionDashboardService) getRiskScoreTrends(ctx context.Context, enterpriseID *uuid.UUID) ([]*DashboardRiskTrendPoint, error) {
	// This would typically query historical data
	// For now, return mock data
	trends := make([]*DashboardRiskTrendPoint, 7)
	now := time.Now()

	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		trends[i] = &DashboardRiskTrendPoint{
			Date:             date,
			AverageScore:     0.3 + float64(i)*0.05,
			MaxScore:         0.8 + float64(i)*0.02,
			TransactionCount: 100 + i*10,
		}
	}

	return trends, nil
}

func (s *FraudPreventionDashboardService) getAlertFrequencyTrends(ctx context.Context, enterpriseID *uuid.UUID) ([]*AlertTrendPoint, error) {
	// This would typically query historical alert data
	// For now, return mock data
	trends := make([]*AlertTrendPoint, 7)
	now := time.Now()

	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		trends[i] = &AlertTrendPoint{
			Date:       date,
			AlertCount: 10 + i*2,
			BySeverity: map[string]int{
				"low":      2 + i,
				"medium":   4 + i,
				"high":     3 + i,
				"critical": 1,
			},
		}
	}

	return trends, nil
}

func (s *FraudPreventionDashboardService) getTransactionVolumeTrends(ctx context.Context, enterpriseID *uuid.UUID) ([]*TransactionTrendPoint, error) {
	// This would typically query historical transaction data
	// For now, return mock data
	trends := make([]*TransactionTrendPoint, 7)
	now := time.Now()

	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		trends[i] = &TransactionTrendPoint{
			Date:     date,
			Volume:   1000 + i*50,
			Amount:   50000 + float64(i)*2500,
			Currency: "USD",
		}
	}

	return trends, nil
}

func (s *FraudPreventionDashboardService) getSystemSettings(ctx context.Context) (map[string]interface{}, error) {
	// This would typically query system configuration
	// For now, return mock data
	return map[string]interface{}{
		"high_risk_threshold":     0.7,
		"critical_risk_threshold": 0.9,
		"auto_block_enabled":      true,
		"alert_channels":          []string{"email", "webhook", "slack"},
		"max_retries":             3,
		"timeout_seconds":         30,
	}, nil
}

func (s *FraudPreventionDashboardService) updateSystemSettings(ctx context.Context, settings map[string]interface{}) error {
	// This would typically update system configuration
	// For now, just return success
	return nil
}

// Response types for dashboard data

type FraudDashboardOverview struct {
	GeneratedAt      time.Time              `json:"generated_at"`
	EnterpriseID     *uuid.UUID             `json:"enterprise_id,omitempty"`
	RealTimeMetrics  *FraudRealTimeMetrics  `json:"real_time_metrics"`
	RecentAlerts     []*FraudAlertSummary   `json:"recent_alerts"`
	RecentCases      []*FraudCaseSummary    `json:"recent_cases"`
	RiskTrends       *FraudRiskTrends       `json:"risk_trends"`
	EnterpriseStatus *FraudEnterpriseStatus `json:"enterprise_status,omitempty"`
}

type FraudRealTimeMetrics struct {
	Timestamp             time.Time              `json:"timestamp"`
	TransactionCounts     *TransactionCounts     `json:"transaction_counts"`
	AlertCounts           *AlertCounts           `json:"alert_counts"`
	CaseCounts            *CaseCounts            `json:"case_counts"`
	RiskScoreDistribution *RiskScoreDistribution `json:"risk_score_distribution"`
}

type TransactionCounts struct {
	Total        int `json:"total"`
	HighRisk     int `json:"high_risk"`
	FraudFlagged int `json:"fraud_flagged"`
	Blocked      int `json:"blocked"`
	InProgress   int `json:"in_progress"`
	Completed    int `json:"completed"`
	Failed       int `json:"failed"`
}

type AlertCounts struct {
	Total         int            `json:"total"`
	New           int            `json:"new"`
	Acknowledged  int            `json:"acknowledged"`
	Investigating int            `json:"investigating"`
	Resolved      int            `json:"resolved"`
	FalsePositive int            `json:"false_positive"`
	BySeverity    map[string]int `json:"by_severity"`
}

type CaseCounts struct {
	Total         int            `json:"total"`
	Open          int            `json:"open"`
	Investigating int            `json:"investigating"`
	Resolved      int            `json:"resolved"`
	Closed        int            `json:"closed"`
	ByPriority    map[string]int `json:"by_priority"`
}

type RiskScoreDistribution struct {
	Low      int     `json:"low"`
	Medium   int     `json:"medium"`
	High     int     `json:"high"`
	Critical int     `json:"critical"`
	Average  float64 `json:"average"`
	Median   float64 `json:"median"`
}

type FraudAlertSummary struct {
	ID           uuid.UUID               `json:"id"`
	AlertType    models.FraudAlertType   `json:"alert_type"`
	Severity     models.FraudSeverity    `json:"severity"`
	Status       models.FraudAlertStatus `json:"status"`
	Score        float64                 `json:"score"`
	Title        string                  `json:"title"`
	DetectedAt   time.Time               `json:"detected_at"`
	EnterpriseID uuid.UUID               `json:"enterprise_id"`
}

type FraudCaseSummary struct {
	ID           uuid.UUID                `json:"id"`
	Title        string                   `json:"title"`
	Category     models.FraudCaseCategory `json:"category"`
	Priority     models.FraudCasePriority `json:"priority"`
	Status       models.FraudCaseStatus   `json:"status"`
	CreatedAt    time.Time                `json:"created_at"`
	EnterpriseID uuid.UUID                `json:"enterprise_id"`
}

type FraudRiskTrends struct {
	GeneratedAt             time.Time                  `json:"generated_at"`
	Window                  time.Duration              `json:"window"`
	RiskScoreTrends         []*DashboardRiskTrendPoint `json:"risk_score_trends"`
	AlertFrequencyTrends    []*AlertTrendPoint         `json:"alert_frequency_trends"`
	TransactionVolumeTrends []*TransactionTrendPoint   `json:"transaction_volume_trends"`
}

type DashboardRiskTrendPoint struct {
	Date             time.Time `json:"date"`
	AverageScore     float64   `json:"average_score"`
	MaxScore         float64   `json:"max_score"`
	TransactionCount int       `json:"transaction_count"`
}

type AlertTrendPoint struct {
	Date       time.Time      `json:"date"`
	AlertCount int            `json:"alert_count"`
	BySeverity map[string]int `json:"by_severity"`
}

type TransactionTrendPoint struct {
	Date     time.Time `json:"date"`
	Volume   int       `json:"volume"`
	Amount   float64   `json:"amount"`
	Currency string    `json:"currency"`
}

type FraudEnterpriseStatus struct {
	EnterpriseID    uuid.UUID                     `json:"enterprise_id"`
	Status          models.AccountFraudStatusType `json:"status"`
	RiskScore       float64                       `json:"risk_score"`
	Restrictions    []models.AccountRestriction   `json:"restrictions"`
	LastUpdated     time.Time                     `json:"last_updated"`
	StatusChangedAt time.Time                     `json:"status_changed_at"`
}

type FraudPreventionConfig struct {
	Rules     []*models.FraudRule    `json:"rules"`
	Settings  map[string]interface{} `json:"settings"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Filter types for repository queries

type FraudCaseFilter struct {
	EnterpriseID *uuid.UUID                `json:"enterprise_id,omitempty"`
	Status       *models.FraudCaseStatus   `json:"status,omitempty"`
	Category     *models.FraudCaseCategory `json:"category,omitempty"`
	Priority     *models.FraudCasePriority `json:"priority,omitempty"`
	StartDate    *time.Time                `json:"start_date,omitempty"`
	EndDate      *time.Time                `json:"end_date,omitempty"`
	Limit        int                       `json:"limit"`
	Offset       int                       `json:"offset"`
}
