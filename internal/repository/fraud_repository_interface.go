package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// FraudRepositoryInterface defines the interface for fraud repository operations
type FraudRepositoryInterface interface {
	// Fraud alert operations
	CreateFraudAlert(ctx context.Context, alert *models.FraudAlert) error
	GetFraudAlertByID(ctx context.Context, id uuid.UUID) (*models.FraudAlert, error)
	UpdateFraudAlert(ctx context.Context, alert *models.FraudAlert) error
	DeleteFraudAlert(ctx context.Context, id uuid.UUID) error
	ListFraudAlerts(ctx context.Context, filter *FraudAlertFilter, limit, offset int) ([]*models.FraudAlert, error)
	GetFraudAlertsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.FraudAlert, error)
	GetFraudAlertsByStatus(ctx context.Context, status models.FraudAlertStatus, limit, offset int) ([]*models.FraudAlert, error)
	GetFraudAlertsBySeverity(ctx context.Context, severity models.FraudSeverity, limit, offset int) ([]*models.FraudAlert, error)

	// Fraud rule operations
	CreateFraudRule(ctx context.Context, rule *models.FraudRule) error
	GetFraudRuleByID(ctx context.Context, id uuid.UUID) (*models.FraudRule, error)
	UpdateFraudRule(ctx context.Context, rule *models.FraudRule) error
	DeleteFraudRule(ctx context.Context, id uuid.UUID) error
	ListFraudRules(ctx context.Context, filter *FraudRuleFilter, limit, offset int) ([]*models.FraudRule, error)
	GetActiveFraudRules(ctx context.Context) ([]*models.FraudRule, error)
	GetFraudRulesByCategory(ctx context.Context, category models.FraudRuleCategory) ([]*models.FraudRule, error)

	// Fraud case operations
	CreateFraudCase(ctx context.Context, case_ *models.FraudCase) error
	GetFraudCaseByID(ctx context.Context, id uuid.UUID) (*models.FraudCase, error)
	UpdateFraudCase(ctx context.Context, case_ *models.FraudCase) error
	DeleteFraudCase(ctx context.Context, id uuid.UUID) error
	ListFraudCases(ctx context.Context, filter *FraudCaseFilter, limit, offset int) ([]*models.FraudCase, error)
	GetFraudCasesByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.FraudCase, error)
	GetFraudCasesByStatus(ctx context.Context, status models.FraudCaseStatus, limit, offset int) ([]*models.FraudCase, error)
	GetFraudCasesByPriority(ctx context.Context, priority models.FraudCasePriority, limit, offset int) ([]*models.FraudCase, error)

	// Account fraud status operations
	CreateAccountFraudStatus(ctx context.Context, status *models.AccountFraudStatus) error
	GetAccountFraudStatusByEnterprise(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error)
	UpdateAccountFraudStatus(ctx context.Context, status *models.AccountFraudStatus) error
	DeleteAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) error
	ListAccountFraudStatus(ctx context.Context, filter *AccountFraudStatusFilter, limit, offset int) ([]*models.AccountFraudStatus, error)
	GetRestrictedAccounts(ctx context.Context) ([]*models.AccountFraudStatus, error)

	// Account restriction operations
	AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error
	RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error
	UpdateAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType, updates map[string]interface{}) error
	GetAccountRestrictions(ctx context.Context, enterpriseID uuid.UUID) ([]models.AccountRestriction, error)

	// Analytics and reporting
	GetFraudAlertSummary(ctx context.Context, enterpriseID *uuid.UUID, startDate, endDate time.Time) (*FraudAlertSummary, error)
	GetFraudMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*FraudMetrics, error)
	GetFraudTrends(ctx context.Context, enterpriseID *uuid.UUID, days int) (*FraudTrends, error)
}

// Filter types
type FraudAlertFilter struct {
	EnterpriseID *uuid.UUID               `json:"enterprise_id,omitempty"`
	Status       *models.FraudAlertStatus `json:"status,omitempty"`
	Severity     *models.FraudSeverity    `json:"severity,omitempty"`
	AlertType    *models.FraudAlertType   `json:"alert_type,omitempty"`
	StartDate    *time.Time               `json:"start_date,omitempty"`
	EndDate      *time.Time               `json:"end_date,omitempty"`
}

type FraudRuleFilter struct {
	Category  *models.FraudRuleCategory `json:"category,omitempty"`
	RuleType  *models.FraudRuleType     `json:"rule_type,omitempty"`
	Status    *models.FraudRuleStatus   `json:"status,omitempty"`
	Severity  *models.FraudSeverity     `json:"severity,omitempty"`
	CreatedBy *uuid.UUID                `json:"created_by,omitempty"`
}

type FraudCaseFilter struct {
	EnterpriseID *uuid.UUID                `json:"enterprise_id,omitempty"`
	Status       *models.FraudCaseStatus   `json:"status,omitempty"`
	Priority     *models.FraudCasePriority `json:"priority,omitempty"`
	Category     *models.FraudCaseCategory `json:"category,omitempty"`
	AssignedTo   *uuid.UUID                `json:"assigned_to,omitempty"`
	StartDate    *time.Time                `json:"start_date,omitempty"`
	EndDate      *time.Time                `json:"end_date,omitempty"`
}

type AccountFraudStatusFilter struct {
	Status          *models.AccountFraudStatusType `json:"status,omitempty"`
	RiskLevel       *models.FraudRiskLevel         `json:"risk_level,omitempty"`
	MonitoringLevel *models.MonitoringLevel        `json:"monitoring_level,omitempty"`
}

// Analytics types
type FraudAlertSummary struct {
	EnterpriseID    uuid.UUID      `json:"enterprise_id"`
	TotalAlerts     int            `json:"total_alerts"`
	NewAlerts       int            `json:"new_alerts"`
	CriticalAlerts  int            `json:"critical_alerts"`
	HighAlerts      int            `json:"high_alerts"`
	AverageScore    float64        `json:"average_score"`
	LatestAlert     *time.Time     `json:"latest_alert,omitempty"`
	AlertBySeverity map[string]int `json:"alert_by_severity"`
	AlertByType     map[string]int `json:"alert_by_type"`
}

type FraudMetrics struct {
	EnterpriseID     uuid.UUID             `json:"enterprise_id"`
	CurrentRiskScore float64               `json:"current_risk_score"`
	RiskLevel        models.FraudRiskLevel `json:"risk_level"`
	ActiveAlerts     int                   `json:"active_alerts"`
	OpenCases        int                   `json:"open_cases"`
	LastAlertDate    *time.Time            `json:"last_alert_date,omitempty"`
	RiskTrend        string                `json:"risk_trend"`
	TopRiskFactors   []string              `json:"top_risk_factors"`
}

type FraudTrends struct {
	EnterpriseID uuid.UUID             `json:"enterprise_id"`
	Period       string                `json:"period"`
	DataPoints   []FraudTrendDataPoint `json:"data_points"`
	RiskTrend    string                `json:"risk_trend"`
	AlertTrend   string                `json:"alert_trend"`
	CaseTrend    string                `json:"case_trend"`
}

type FraudTrendDataPoint struct {
	Date       time.Time `json:"date"`
	RiskScore  float64   `json:"risk_score"`
	AlertCount int       `json:"alert_count"`
	CaseCount  int       `json:"case_count"`
}
