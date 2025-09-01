package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
)

// FraudRepository implements the FraudRepositoryInterface
type FraudRepository struct {
	// In a real implementation, this would have database connections
	// For now, we'll use in-memory storage for testing
	alerts       map[uuid.UUID]*models.FraudAlert
	rules        map[uuid.UUID]*models.FraudRule
	cases        map[uuid.UUID]*models.FraudCase
	statuses     map[uuid.UUID]*models.AccountFraudStatus
	restrictions map[uuid.UUID][]models.AccountRestriction
}

// NewFraudRepository creates a new fraud repository instance
func NewFraudRepository() FraudRepositoryInterface {
	return &FraudRepository{
		alerts:       make(map[uuid.UUID]*models.FraudAlert),
		rules:        make(map[uuid.UUID]*models.FraudRule),
		cases:        make(map[uuid.UUID]*models.FraudCase),
		statuses:     make(map[uuid.UUID]*models.AccountFraudStatus),
		restrictions: make(map[uuid.UUID][]models.AccountRestriction),
	}
}

// Fraud alert operations
func (r *FraudRepository) CreateFraudAlert(ctx context.Context, alert *models.FraudAlert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	r.alerts[alert.ID] = alert
	return nil
}

func (r *FraudRepository) GetFraudAlertByID(ctx context.Context, id uuid.UUID) (*models.FraudAlert, error) {
	alert, exists := r.alerts[id]
	if !exists {
		return nil, fmt.Errorf("fraud alert not found: %s", id)
	}
	return alert, nil
}

func (r *FraudRepository) UpdateFraudAlert(ctx context.Context, alert *models.FraudAlert) error {
	if _, exists := r.alerts[alert.ID]; !exists {
		return fmt.Errorf("fraud alert not found: %s", alert.ID)
	}
	alert.UpdatedAt = time.Now()
	r.alerts[alert.ID] = alert
	return nil
}

func (r *FraudRepository) DeleteFraudAlert(ctx context.Context, id uuid.UUID) error {
	if _, exists := r.alerts[id]; !exists {
		return fmt.Errorf("fraud alert not found: %s", id)
	}
	delete(r.alerts, id)
	return nil
}

func (r *FraudRepository) ListFraudAlerts(ctx context.Context, filter *FraudAlertFilter, limit, offset int) ([]*models.FraudAlert, error) {
	var alerts []*models.FraudAlert

	for _, alert := range r.alerts {
		if filter != nil {
			if filter.EnterpriseID != nil && alert.EnterpriseID != *filter.EnterpriseID {
				continue
			}
			if filter.Status != nil && alert.Status != *filter.Status {
				continue
			}
			if filter.Severity != nil && alert.Severity != *filter.Severity {
				continue
			}
			if filter.AlertType != nil && alert.AlertType != *filter.AlertType {
				continue
			}
			if filter.StartDate != nil && alert.DetectedAt.Before(*filter.StartDate) {
				continue
			}
			if filter.EndDate != nil && alert.DetectedAt.After(*filter.EndDate) {
				continue
			}
		}
		alerts = append(alerts, alert)
	}

	// Simple pagination
	if offset >= len(alerts) {
		return []*models.FraudAlert{}, nil
	}
	end := offset + limit
	if end > len(alerts) {
		end = len(alerts)
	}

	return alerts[offset:end], nil
}

func (r *FraudRepository) GetFraudAlertsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.FraudAlert, error) {
	filter := &FraudAlertFilter{EnterpriseID: &enterpriseID}
	return r.ListFraudAlerts(ctx, filter, limit, offset)
}

func (r *FraudRepository) GetFraudAlertsByStatus(ctx context.Context, status models.FraudAlertStatus, limit, offset int) ([]*models.FraudAlert, error) {
	filter := &FraudAlertFilter{Status: &status}
	return r.ListFraudAlerts(ctx, filter, limit, offset)
}

func (r *FraudRepository) GetFraudAlertsBySeverity(ctx context.Context, severity models.FraudSeverity, limit, offset int) ([]*models.FraudAlert, error) {
	filter := &FraudAlertFilter{Severity: &severity}
	return r.ListFraudAlerts(ctx, filter, limit, offset)
}

// Fraud rule operations
func (r *FraudRepository) CreateFraudRule(ctx context.Context, rule *models.FraudRule) error {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	r.rules[rule.ID] = rule
	return nil
}

func (r *FraudRepository) GetFraudRuleByID(ctx context.Context, id uuid.UUID) (*models.FraudRule, error) {
	rule, exists := r.rules[id]
	if !exists {
		return nil, fmt.Errorf("fraud rule not found: %s", id)
	}
	return rule, nil
}

func (r *FraudRepository) UpdateFraudRule(ctx context.Context, rule *models.FraudRule) error {
	if _, exists := r.rules[rule.ID]; !exists {
		return fmt.Errorf("fraud rule not found: %s", rule.ID)
	}
	rule.UpdatedAt = time.Now()
	r.rules[rule.ID] = rule
	return nil
}

func (r *FraudRepository) DeleteFraudRule(ctx context.Context, id uuid.UUID) error {
	if _, exists := r.rules[id]; !exists {
		return fmt.Errorf("fraud rule not found: %s", id)
	}
	delete(r.rules, id)
	return nil
}

func (r *FraudRepository) ListFraudRules(ctx context.Context, filter *FraudRuleFilter, limit, offset int) ([]*models.FraudRule, error) {
	var rules []*models.FraudRule

	for _, rule := range r.rules {
		if filter != nil {
			if filter.Category != nil && rule.Category != *filter.Category {
				continue
			}
			if filter.RuleType != nil && rule.RuleType != *filter.RuleType {
				continue
			}
			if filter.Status != nil && rule.Status != *filter.Status {
				continue
			}
			if filter.Severity != nil && rule.Severity != *filter.Severity {
				continue
			}
			if filter.CreatedBy != nil && rule.CreatedBy != *filter.CreatedBy {
				continue
			}
		}
		rules = append(rules, rule)
	}

	// Simple pagination
	if offset >= len(rules) {
		return []*models.FraudRule{}, nil
	}
	end := offset + limit
	if end > len(rules) {
		end = len(rules)
	}

	return rules[offset:end], nil
}

func (r *FraudRepository) GetActiveFraudRules(ctx context.Context) ([]*models.FraudRule, error) {
	var activeRules []*models.FraudRule

	for _, rule := range r.rules {
		if rule.IsActive() {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules, nil
}

func (r *FraudRepository) GetFraudRulesByCategory(ctx context.Context, category models.FraudRuleCategory) ([]*models.FraudRule, error) {
	filter := &FraudRuleFilter{Category: &category}
	return r.ListFraudRules(ctx, filter, 100, 0)
}

// Fraud case operations
func (r *FraudRepository) CreateFraudCase(ctx context.Context, case_ *models.FraudCase) error {
	if case_.ID == uuid.Nil {
		case_.ID = uuid.New()
	}
	case_.CreatedAt = time.Now()
	case_.UpdatedAt = time.Now()
	r.cases[case_.ID] = case_
	return nil
}

func (r *FraudRepository) GetFraudCaseByID(ctx context.Context, id uuid.UUID) (*models.FraudCase, error) {
	case_, exists := r.cases[id]
	if !exists {
		return nil, fmt.Errorf("fraud case not found: %s", id)
	}
	return case_, nil
}

func (r *FraudRepository) UpdateFraudCase(ctx context.Context, case_ *models.FraudCase) error {
	if _, exists := r.cases[case_.ID]; !exists {
		return fmt.Errorf("fraud case not found: %s", case_.ID)
	}
	case_.UpdatedAt = time.Now()
	r.cases[case_.ID] = case_
	return nil
}

func (r *FraudRepository) DeleteFraudCase(ctx context.Context, id uuid.UUID) error {
	if _, exists := r.cases[id]; !exists {
		return fmt.Errorf("fraud case not found: %s", id)
	}
	delete(r.cases, id)
	return nil
}

func (r *FraudRepository) ListFraudCases(ctx context.Context, filter *FraudCaseFilter, limit, offset int) ([]*models.FraudCase, error) {
	var cases []*models.FraudCase

	for _, case_ := range r.cases {
		if filter != nil {
			if filter.EnterpriseID != nil && case_.EnterpriseID != *filter.EnterpriseID {
				continue
			}
			if filter.Status != nil && case_.Status != *filter.Status {
				continue
			}
			if filter.Priority != nil && case_.Priority != *filter.Priority {
				continue
			}
			if filter.Category != nil && case_.Category != *filter.Category {
				continue
			}
			if filter.AssignedTo != nil && (case_.AssignedTo == nil || *case_.AssignedTo != *filter.AssignedTo) {
				continue
			}
		}
		cases = append(cases, case_)
	}

	// Simple pagination
	if offset >= len(cases) {
		return []*models.FraudCase{}, nil
	}
	end := offset + limit
	if end > len(cases) {
		end = len(cases)
	}

	return cases[offset:end], nil
}

func (r *FraudRepository) GetFraudCasesByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.FraudCase, error) {
	filter := &FraudCaseFilter{EnterpriseID: &enterpriseID}
	return r.ListFraudCases(ctx, filter, limit, offset)
}

func (r *FraudRepository) GetFraudCasesByStatus(ctx context.Context, status models.FraudCaseStatus, limit, offset int) ([]*models.FraudCase, error) {
	filter := &FraudCaseFilter{Status: &status}
	return r.ListFraudCases(ctx, filter, limit, offset)
}

func (r *FraudRepository) GetFraudCasesByPriority(ctx context.Context, priority models.FraudCasePriority, limit, offset int) ([]*models.FraudCase, error) {
	filter := &FraudCaseFilter{Priority: &priority}
	return r.ListFraudCases(ctx, filter, limit, offset)
}

// Account fraud status operations
func (r *FraudRepository) CreateAccountFraudStatus(ctx context.Context, status *models.AccountFraudStatus) error {
	status.CreatedAt = time.Now()
	status.UpdatedAt = time.Now()
	r.statuses[status.EnterpriseID] = status
	return nil
}

func (r *FraudRepository) GetAccountFraudStatusByEnterprise(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error) {
	status, exists := r.statuses[enterpriseID]
	if !exists {
		// Return default status
		status = &models.AccountFraudStatus{
			ID:              uuid.New(),
			EnterpriseID:    enterpriseID,
			Status:          models.AccountFraudStatusNormal,
			RiskScore:       0.0,
			RiskLevel:       models.FraudRiskLevelLow,
			RiskFactors:     []string{},
			Restrictions:    []models.AccountRestriction{},
			MonitoringLevel: models.MonitoringLevelStandard,
			NextReviewDate:  time.Now().AddDate(0, 0, 30),
			StatusHistory:   []models.FraudStatusChange{},
			StatusChangedAt: time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}
	return status, nil
}

func (r *FraudRepository) UpdateAccountFraudStatus(ctx context.Context, status *models.AccountFraudStatus) error {
	status.UpdatedAt = time.Now()
	r.statuses[status.EnterpriseID] = status
	return nil
}

func (r *FraudRepository) DeleteAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) error {
	delete(r.statuses, enterpriseID)
	return nil
}

func (r *FraudRepository) ListAccountFraudStatus(ctx context.Context, filter *AccountFraudStatusFilter, limit, offset int) ([]*models.AccountFraudStatus, error) {
	var statuses []*models.AccountFraudStatus

	for _, status := range r.statuses {
		if filter != nil {
			if filter.Status != nil && status.Status != *filter.Status {
				continue
			}
			if filter.RiskLevel != nil && status.RiskLevel != *filter.RiskLevel {
				continue
			}
			if filter.MonitoringLevel != nil && status.MonitoringLevel != *filter.MonitoringLevel {
				continue
			}
		}
		statuses = append(statuses, status)
	}

	// Simple pagination
	if offset >= len(statuses) {
		return []*models.AccountFraudStatus{}, nil
	}
	end := offset + limit
	if end > len(statuses) {
		end = len(statuses)
	}

	return statuses[offset:end], nil
}

func (r *FraudRepository) GetRestrictedAccounts(ctx context.Context) ([]*models.AccountFraudStatus, error) {
	filter := &AccountFraudStatusFilter{}
	statuses, err := r.ListAccountFraudStatus(ctx, filter, 1000, 0)
	if err != nil {
		return nil, err
	}

	var restricted []*models.AccountFraudStatus
	for _, status := range statuses {
		if status.IsRestricted() {
			restricted = append(restricted, status)
		}
	}

	return restricted, nil
}

// Account restriction operations
func (r *FraudRepository) AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error {
	if restriction.EffectiveAt.IsZero() {
		restriction.EffectiveAt = time.Now()
	}

	restrictions := r.restrictions[enterpriseID]
	restrictions = append(restrictions, *restriction)
	r.restrictions[enterpriseID] = restrictions

	return nil
}

func (r *FraudRepository) RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error {
	restrictions := r.restrictions[enterpriseID]
	var filtered []models.AccountRestriction

	for _, restriction := range restrictions {
		if restriction.Type != restrictionType {
			filtered = append(filtered, restriction)
		}
	}

	r.restrictions[enterpriseID] = filtered
	return nil
}

func (r *FraudRepository) UpdateAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType, updates map[string]interface{}) error {
	restrictions := r.restrictions[enterpriseID]

	for i, restriction := range restrictions {
		if restriction.Type == restrictionType {
			// Update fields based on updates map
			if expiresAt, ok := updates["expires_at"]; ok {
				if expiresAtTime, ok := expiresAt.(time.Time); ok {
					restriction.ExpiresAt = &expiresAtTime
				}
			}
			if description, ok := updates["description"]; ok {
				if desc, ok := description.(string); ok {
					restriction.Description = desc
				}
			}
			if parameters, ok := updates["parameters"]; ok {
				if params, ok := parameters.(map[string]interface{}); ok {
					restriction.Parameters = params
				}
			}
			restrictions[i] = restriction
			break
		}
	}

	r.restrictions[enterpriseID] = restrictions
	return nil
}

func (r *FraudRepository) GetAccountRestrictions(ctx context.Context, enterpriseID uuid.UUID) ([]models.AccountRestriction, error) {
	restrictions, exists := r.restrictions[enterpriseID]
	if !exists {
		return []models.AccountRestriction{}, nil
	}
	return restrictions, nil
}

// Analytics and reporting
func (r *FraudRepository) GetFraudAlertSummary(ctx context.Context, enterpriseID *uuid.UUID, startDate, endDate time.Time) (*FraudAlertSummary, error) {
	// Implementation would aggregate data from alerts
	summary := &FraudAlertSummary{
		EnterpriseID:    uuid.Nil,
		TotalAlerts:     0,
		NewAlerts:       0,
		CriticalAlerts:  0,
		HighAlerts:      0,
		AverageScore:    0.0,
		AlertBySeverity: make(map[string]int),
		AlertByType:     make(map[string]int),
	}

	if enterpriseID != nil {
		summary.EnterpriseID = *enterpriseID
	}

	return summary, nil
}

func (r *FraudRepository) GetFraudMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*FraudMetrics, error) {
	// Implementation would calculate metrics from data
	metrics := &FraudMetrics{
		EnterpriseID:     uuid.Nil,
		CurrentRiskScore: 0.0,
		RiskLevel:        models.FraudRiskLevelLow,
		ActiveAlerts:     0,
		OpenCases:        0,
		RiskTrend:        "stable",
		TopRiskFactors:   []string{},
	}

	if enterpriseID != nil {
		metrics.EnterpriseID = *enterpriseID
	}

	return metrics, nil
}

func (r *FraudRepository) GetFraudTrends(ctx context.Context, enterpriseID *uuid.UUID, days int) (*FraudTrends, error) {
	// Implementation would calculate trends over time
	trends := &FraudTrends{
		EnterpriseID: uuid.Nil,
		Period:       fmt.Sprintf("%d days", days),
		DataPoints:   []FraudTrendDataPoint{},
		RiskTrend:    "stable",
		AlertTrend:   "stable",
		CaseTrend:    "stable",
	}

	if enterpriseID != nil {
		trends.EnterpriseID = *enterpriseID
	}

	return trends, nil
}
