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

// FraudPreventionConfigService manages fraud prevention configuration
type FraudPreventionConfigService struct {
	fraudRepo       repository.FraudRepositoryInterface
	messagingClient messaging.EventBus
	config          *FraudPreventionConfigServiceConfig
}

// FraudPreventionConfigServiceConfig defines configuration for the fraud prevention config service
type FraudPreventionConfigServiceConfig struct {
	// Configuration validation
	MaxRulesPerEnterprise int `json:"max_rules_per_enterprise"` // e.g., 100
	MaxRuleConditions     int `json:"max_rule_conditions"`      // e.g., 20
	MaxRuleActions        int `json:"max_rule_actions"`         // e.g., 10

	// Rule validation
	MinRuleScore      float64 `json:"min_rule_score"`      // e.g., 0.1
	MaxRuleScore      float64 `json:"max_rule_score"`      // e.g., 1.0
	MinRuleConfidence float64 `json:"min_rule_confidence"` // e.g., 0.5
	MaxRuleConfidence float64 `json:"max_rule_confidence"` // e.g., 1.0

	// Configuration limits
	MaxWhitelistedEnterprises int `json:"max_whitelisted_enterprises"` // e.g., 1000
	MaxAlertChannels          int `json:"max_alert_channels"`          // e.g., 10

	// Audit settings
	AuditConfigurationChanges bool `json:"audit_configuration_changes"`
	AuditRetentionDays        int  `json:"audit_retention_days"` // e.g., 90
}

// NewFraudPreventionConfigService creates a new fraud prevention configuration service
func NewFraudPreventionConfigService(
	fraudRepo repository.FraudRepositoryInterface,
	messagingClient messaging.EventBus,
	config *FraudPreventionConfigServiceConfig,
) *FraudPreventionConfigService {
	if config == nil {
		config = &FraudPreventionConfigServiceConfig{
			MaxRulesPerEnterprise:     100,
			MaxRuleConditions:         20,
			MaxRuleActions:            10,
			MinRuleScore:              0.1,
			MaxRuleScore:              1.0,
			MinRuleConfidence:         0.5,
			MaxRuleConfidence:         1.0,
			MaxWhitelistedEnterprises: 1000,
			MaxAlertChannels:          10,
			AuditConfigurationChanges: true,
			AuditRetentionDays:        90,
		}
	}

	return &FraudPreventionConfigService{
		fraudRepo:       fraudRepo,
		messagingClient: messagingClient,
		config:          config,
	}
}

// GetFraudPreventionConfiguration gets the complete fraud prevention configuration
func (s *FraudPreventionConfigService) GetFraudPreventionConfiguration(ctx context.Context, enterpriseID *uuid.UUID) (*FraudPreventionConfiguration, error) {
	config := &FraudPreventionConfiguration{
		GeneratedAt:  time.Now(),
		EnterpriseID: enterpriseID,
	}

	// Get fraud rules
	rules, err := s.getFraudRules(ctx, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fraud rules: %w", err)
	}
	config.Rules = rules

	// Get system settings
	settings, err := s.getSystemSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system settings: %w", err)
	}
	config.SystemSettings = settings

	// Get enterprise-specific settings
	if enterpriseID != nil {
		enterpriseSettings, err := s.getEnterpriseSettings(ctx, *enterpriseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get enterprise settings: %w", err)
		}
		config.EnterpriseSettings = enterpriseSettings
	}

	// Get whitelist configuration
	whitelist, err := s.getWhitelistConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelist configuration: %w", err)
	}
	config.WhitelistConfiguration = whitelist

	// Get alert configuration
	alertConfig, err := s.getAlertConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert configuration: %w", err)
	}
	config.AlertConfiguration = alertConfig

	return config, nil
}

// UpdateFraudPreventionConfiguration updates the fraud prevention configuration
func (s *FraudPreventionConfigService) UpdateFraudPreventionConfiguration(ctx context.Context, config *FraudPreventionConfiguration, updatedBy uuid.UUID) error {
	// Validate configuration
	if err := s.validateConfiguration(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Update fraud rules
	if err := s.updateFraudRules(ctx, config.Rules, updatedBy); err != nil {
		return fmt.Errorf("failed to update fraud rules: %w", err)
	}

	// Update system settings
	if err := s.updateSystemSettings(ctx, config.SystemSettings, updatedBy); err != nil {
		return fmt.Errorf("failed to update system settings: %w", err)
	}

	// Update enterprise settings
	if config.EnterpriseID != nil {
		if err := s.updateEnterpriseSettings(ctx, *config.EnterpriseID, config.EnterpriseSettings, updatedBy); err != nil {
			return fmt.Errorf("failed to update enterprise settings: %w", err)
		}
	}

	// Update whitelist configuration
	if err := s.updateWhitelistConfiguration(ctx, config.WhitelistConfiguration, updatedBy); err != nil {
		return fmt.Errorf("failed to update whitelist configuration: %w", err)
	}

	// Update alert configuration
	if err := s.updateAlertConfiguration(ctx, config.AlertConfiguration, updatedBy); err != nil {
		return fmt.Errorf("failed to update alert configuration: %w", err)
	}

	// Audit configuration change
	if s.config.AuditConfigurationChanges {
		if err := s.auditConfigurationChange(ctx, config, updatedBy); err != nil {
			return fmt.Errorf("failed to audit configuration change: %w", err)
		}
	}

	// Publish configuration update event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "fraud_prevention.configuration_updated",
			Source: "fraud-prevention-config-service",
			Data: map[string]interface{}{
				"enterprise_id": config.EnterpriseID,
				"updated_by":    updatedBy,
				"updated_at":    time.Now(),
				"rules_count":   len(config.Rules),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(ctx, event)
	}

	return nil
}

// CreateFraudRule creates a new fraud rule
func (s *FraudPreventionConfigService) CreateFraudRule(ctx context.Context, rule *models.FraudRule, createdBy uuid.UUID) error {
	// Validate rule
	if err := s.validateFraudRule(rule); err != nil {
		return fmt.Errorf("fraud rule validation failed: %w", err)
	}

	// Set creation metadata
	rule.ID = uuid.New()
	rule.CreatedBy = createdBy
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Create rule
	if err := s.fraudRepo.CreateFraudRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to create fraud rule: %w", err)
	}

	// Publish rule creation event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "fraud_prevention.rule_created",
			Source: "fraud-prevention-config-service",
			Data: map[string]interface{}{
				"rule_id":    rule.ID,
				"rule_name":  rule.Name,
				"created_by": createdBy,
				"created_at": rule.CreatedAt,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(ctx, event)
	}

	return nil
}

// UpdateFraudRule updates an existing fraud rule
func (s *FraudPreventionConfigService) UpdateFraudRule(ctx context.Context, rule *models.FraudRule, updatedBy uuid.UUID) error {
	// Validate rule
	if err := s.validateFraudRule(rule); err != nil {
		return fmt.Errorf("fraud rule validation failed: %w", err)
	}

	// Get existing rule
	existingRule, err := s.fraudRepo.GetFraudRuleByID(ctx, rule.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing rule: %w", err)
	}

	// Update metadata
	rule.UpdatedAt = time.Now()
	rule.UpdatedBy = updatedBy

	// Update rule
	if err := s.fraudRepo.UpdateFraudRule(ctx, rule); err != nil {
		return fmt.Errorf("failed to update fraud rule: %w", err)
	}

	// Publish rule update event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "fraud_prevention.rule_updated",
			Source: "fraud-prevention-config-service",
			Data: map[string]interface{}{
				"rule_id":         rule.ID,
				"rule_name":       rule.Name,
				"updated_by":      updatedBy,
				"updated_at":      rule.UpdatedAt,
				"previous_status": existingRule.Status,
				"new_status":      rule.Status,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(ctx, event)
	}

	return nil
}

// DeleteFraudRule deletes a fraud rule
func (s *FraudPreventionConfigService) DeleteFraudRule(ctx context.Context, ruleID uuid.UUID, deletedBy uuid.UUID) error {
	// Get existing rule
	rule, err := s.fraudRepo.GetFraudRuleByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// Delete rule
	if err := s.fraudRepo.DeleteFraudRule(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete fraud rule: %w", err)
	}

	// Publish rule deletion event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "fraud_prevention.rule_deleted",
			Source: "fraud-prevention-config-service",
			Data: map[string]interface{}{
				"rule_id":    ruleID,
				"rule_name":  rule.Name,
				"deleted_by": deletedBy,
				"deleted_at": time.Now(),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		_ = s.messagingClient.PublishEvent(ctx, event)
	}

	return nil
}

// GetConfigurationHistory gets the configuration change history
func (s *FraudPreventionConfigService) GetConfigurationHistory(ctx context.Context, filter *ConfigurationHistoryFilter) ([]*ConfigurationHistoryEntry, error) {
	// This would typically query a configuration audit log
	// For now, return mock data
	entries := make([]*ConfigurationHistoryEntry, 0)

	// Mock history entries
	now := time.Now()
	for i := 0; i < 5; i++ {
		entries = append(entries, &ConfigurationHistoryEntry{
			ID:          uuid.New(),
			ChangeType:  "rule_updated",
			EntityID:    uuid.New(),
			EntityName:  fmt.Sprintf("Rule %d", i+1),
			ChangedBy:   uuid.New(),
			ChangedAt:   now.AddDate(0, 0, -i),
			Description: fmt.Sprintf("Updated rule configuration %d", i+1),
		})
	}

	return entries, nil
}

// Helper methods

func (s *FraudPreventionConfigService) getFraudRules(ctx context.Context, enterpriseID *uuid.UUID) ([]*models.FraudRule, error) {
	// For now, get all active rules since there's no enterprise-specific method
	// In a real implementation, you would filter by enterprise
	return s.fraudRepo.GetActiveFraudRules(ctx)
}

func (s *FraudPreventionConfigService) getSystemSettings(ctx context.Context) (map[string]interface{}, error) {
	// This would typically query system configuration
	// For now, return mock data
	return map[string]interface{}{
		"high_risk_threshold":     0.7,
		"critical_risk_threshold": 0.9,
		"auto_block_enabled":      true,
		"max_retries":             3,
		"timeout_seconds":         30,
		"default_currency":        "USD",
		"supported_currencies":    []string{"USD", "EUR", "GBP", "JPY"},
	}, nil
}

func (s *FraudPreventionConfigService) getEnterpriseSettings(ctx context.Context, enterpriseID uuid.UUID) (map[string]interface{}, error) {
	// This would typically query enterprise-specific configuration
	// For now, return mock data
	return map[string]interface{}{
		"custom_risk_threshold":  0.6,
		"alert_channels":         []string{"email", "webhook"},
		"notification_frequency": "immediate",
		"timezone":               "UTC",
	}, nil
}

func (s *FraudPreventionConfigService) getWhitelistConfiguration(ctx context.Context) (*WhitelistConfiguration, error) {
	// This would typically query whitelist configuration
	// For now, return mock data
	return &WhitelistConfiguration{
		WhitelistedEnterprises: []uuid.UUID{uuid.New(), uuid.New()},
		WhitelistedIPs:         []string{"192.168.1.1", "10.0.0.1"},
		WhitelistedDomains:     []string{"trusted.com", "verified.org"},
		AutoWhitelistEnabled:   true,
		WhitelistExpiryDays:    30,
	}, nil
}

func (s *FraudPreventionConfigService) getAlertConfiguration(ctx context.Context) (*AlertConfiguration, error) {
	// This would typically query alert configuration
	// For now, return mock data
	return &AlertConfiguration{
		Channels: []AlertChannel{
			{Type: "email", Enabled: true, Config: map[string]interface{}{"recipients": []string{"admin@company.com"}}},
			{Type: "webhook", Enabled: true, Config: map[string]interface{}{"url": "https://webhook.site/abc123"}},
			{Type: "slack", Enabled: false, Config: map[string]interface{}{"channel": "#fraud-alerts"}},
		},
		EscalationRules: []ConfigEscalationRule{
			{SeverityLevel: "high", DelayMinutes: 30, EscalateTo: "manager"},
			{SeverityLevel: "critical", DelayMinutes: 5, EscalateTo: "emergency"},
		},
		Throttling: AlertThrottling{
			MaxAlertsPerHour: 100,
			MaxAlertsPerDay:  1000,
			ThrottleWindow:   1 * time.Hour,
		},
	}, nil
}

func (s *FraudPreventionConfigService) validateConfiguration(config *FraudPreventionConfiguration) error {
	// Validate rules
	if len(config.Rules) > s.config.MaxRulesPerEnterprise {
		return fmt.Errorf("too many rules: %d exceeds maximum %d", len(config.Rules), s.config.MaxRulesPerEnterprise)
	}

	// Validate each rule
	for _, rule := range config.Rules {
		if err := s.validateFraudRule(rule); err != nil {
			return fmt.Errorf("invalid rule %s: %w", rule.Name, err)
		}
	}

	// Validate whitelist configuration
	if config.WhitelistConfiguration != nil {
		if len(config.WhitelistConfiguration.WhitelistedEnterprises) > s.config.MaxWhitelistedEnterprises {
			return fmt.Errorf("too many whitelisted enterprises: %d exceeds maximum %d",
				len(config.WhitelistConfiguration.WhitelistedEnterprises), s.config.MaxWhitelistedEnterprises)
		}
	}

	// Validate alert configuration
	if config.AlertConfiguration != nil {
		if len(config.AlertConfiguration.Channels) > s.config.MaxAlertChannels {
			return fmt.Errorf("too many alert channels: %d exceeds maximum %d",
				len(config.AlertConfiguration.Channels), s.config.MaxAlertChannels)
		}
	}

	return nil
}

func (s *FraudPreventionConfigService) validateFraudRule(rule *models.FraudRule) error {
	// Validate basic fields
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if rule.Description == "" {
		return fmt.Errorf("rule description is required")
	}

	// Validate score range
	if rule.BaseScore < s.config.MinRuleScore || rule.BaseScore > s.config.MaxRuleScore {
		return fmt.Errorf("rule score %.2f is outside valid range [%.2f, %.2f]",
			rule.BaseScore, s.config.MinRuleScore, s.config.MaxRuleScore)
	}

	// Validate confidence range
	if rule.Confidence < s.config.MinRuleConfidence || rule.Confidence > s.config.MaxRuleConfidence {
		return fmt.Errorf("rule confidence %.2f is outside valid range [%.2f, %.2f]",
			rule.Confidence, s.config.MinRuleConfidence, s.config.MaxRuleConfidence)
	}

	// Validate conditions
	if len(rule.Conditions) > s.config.MaxRuleConditions {
		return fmt.Errorf("too many conditions: %d exceeds maximum %d",
			len(rule.Conditions), s.config.MaxRuleConditions)
	}

	// Validate actions
	if len(rule.Actions) > s.config.MaxRuleActions {
		return fmt.Errorf("too many actions: %d exceeds maximum %d",
			len(rule.Actions), s.config.MaxRuleActions)
	}

	return nil
}

func (s *FraudPreventionConfigService) updateFraudRules(ctx context.Context, rules []*models.FraudRule, updatedBy uuid.UUID) error {
	for _, rule := range rules {
		if err := s.fraudRepo.UpdateFraudRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to update rule %s: %w", rule.ID, err)
		}
	}
	return nil
}

func (s *FraudPreventionConfigService) updateSystemSettings(ctx context.Context, settings map[string]interface{}, updatedBy uuid.UUID) error {
	// This would typically update system configuration
	// For now, just return success
	return nil
}

func (s *FraudPreventionConfigService) updateEnterpriseSettings(ctx context.Context, enterpriseID uuid.UUID, settings map[string]interface{}, updatedBy uuid.UUID) error {
	// This would typically update enterprise-specific configuration
	// For now, just return success
	return nil
}

func (s *FraudPreventionConfigService) updateWhitelistConfiguration(ctx context.Context, config *WhitelistConfiguration, updatedBy uuid.UUID) error {
	// This would typically update whitelist configuration
	// For now, just return success
	return nil
}

func (s *FraudPreventionConfigService) updateAlertConfiguration(ctx context.Context, config *AlertConfiguration, updatedBy uuid.UUID) error {
	// This would typically update alert configuration
	// For now, just return success
	return nil
}

func (s *FraudPreventionConfigService) auditConfigurationChange(ctx context.Context, config *FraudPreventionConfiguration, updatedBy uuid.UUID) error {
	// This would typically log configuration changes to an audit log
	// For now, just return success
	return nil
}

// Response types

type FraudPreventionConfiguration struct {
	GeneratedAt            time.Time               `json:"generated_at"`
	EnterpriseID           *uuid.UUID              `json:"enterprise_id,omitempty"`
	Rules                  []*models.FraudRule     `json:"rules"`
	SystemSettings         map[string]interface{}  `json:"system_settings"`
	EnterpriseSettings     map[string]interface{}  `json:"enterprise_settings,omitempty"`
	WhitelistConfiguration *WhitelistConfiguration `json:"whitelist_configuration"`
	AlertConfiguration     *AlertConfiguration     `json:"alert_configuration"`
}

type WhitelistConfiguration struct {
	WhitelistedEnterprises []uuid.UUID `json:"whitelisted_enterprises"`
	WhitelistedIPs         []string    `json:"whitelisted_ips"`
	WhitelistedDomains     []string    `json:"whitelisted_domains"`
	AutoWhitelistEnabled   bool        `json:"auto_whitelist_enabled"`
	WhitelistExpiryDays    int         `json:"whitelist_expiry_days"`
}

type AlertConfiguration struct {
	Channels        []AlertChannel         `json:"channels"`
	EscalationRules []ConfigEscalationRule `json:"escalation_rules"`
	Throttling      AlertThrottling        `json:"throttling"`
}

type AlertChannel struct {
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

type ConfigEscalationRule struct {
	SeverityLevel string `json:"severity_level"`
	DelayMinutes  int    `json:"delay_minutes"`
	EscalateTo    string `json:"escalate_to"`
}

type AlertThrottling struct {
	MaxAlertsPerHour int           `json:"max_alerts_per_hour"`
	MaxAlertsPerDay  int           `json:"max_alerts_per_day"`
	ThrottleWindow   time.Duration `json:"throttle_window"`
}

type ConfigurationHistoryFilter struct {
	EnterpriseID *uuid.UUID `json:"enterprise_id,omitempty"`
	ChangeType   *string    `json:"change_type,omitempty"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

type ConfigurationHistoryEntry struct {
	ID          uuid.UUID `json:"id"`
	ChangeType  string    `json:"change_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	EntityName  string    `json:"entity_name"`
	ChangedBy   uuid.UUID `json:"changed_by"`
	ChangedAt   time.Time `json:"changed_at"`
	Description string    `json:"description"`
}
