package models

import (
	"encoding/json"
	"time"
)

// RegulatoryRule represents a compliance rule that can be applied to transactions
type RegulatoryRule struct {
	ID           string `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"type:varchar(255);not null"`
	Description  string `json:"description" gorm:"type:text"`
	Category     string `json:"category" gorm:"type:varchar(100);not null"` // aml, kyc, sanctions, tax, industry_specific
	Jurisdiction string `json:"jurisdiction" gorm:"type:varchar(100);not null"`
	Priority     int    `json:"priority" gorm:"type:int;not null;default:1"` // 1=low, 5=critical

	// Rule Configuration
	RuleType   string                 `json:"rule_type" gorm:"type:varchar(50);not null"` // amount_limit, frequency_limit, pattern_detection, blacklist_check
	Conditions map[string]interface{} `json:"conditions" gorm:"type:jsonb"`
	Thresholds map[string]interface{} `json:"thresholds" gorm:"type:jsonb"`
	Actions    []string               `json:"actions" gorm:"type:jsonb"` // flag, reject, hold, alert

	// Status and Versioning
	Status      string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // active, inactive, deprecated
	Version     int        `json:"version" gorm:"type:int;not null;default:1"`
	EffectiveAt time.Time  `json:"effective_at" gorm:"not null"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

	// Metadata
	CreatedBy string    `json:"created_by" gorm:"not null"`
	UpdatedBy string    `json:"updated_by" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// ComplianceValidationRequest represents a request to validate compliance
type ComplianceValidationRequest struct {
	TransactionID   string                 `json:"transaction_id"`
	EnterpriseID    string                 `json:"enterprise_id"`
	Jurisdiction    string                 `json:"jurisdiction"`
	TransactionType string                 `json:"transaction_type"`
	Amount          string                 `json:"amount"`
	Currency        string                 `json:"currency"`
	FromAddress     string                 `json:"from_address"`
	ToAddress       string                 `json:"to_address"`
	Metadata        map[string]interface{} `json:"metadata"`
	ValidationLevel string                 `json:"validation_level"` // basic, enhanced, strict
}

// ComplianceValidationResult represents the result of compliance validation
type ComplianceValidationResult struct {
	ID              string `json:"id" gorm:"primaryKey"`
	TransactionID   string `json:"transaction_id" gorm:"not null;index"`
	ValidationLevel string `json:"validation_level" gorm:"type:varchar(20);not null"`

	// Overall Result
	IsCompliant     bool    `json:"is_compliant"`
	ComplianceScore float64 `json:"compliance_score" gorm:"type:decimal(5,4)"`
	Status          string  `json:"status" gorm:"type:varchar(20);not null"` // approved, flagged, rejected, pending_review

	// Rule Evaluation Results
	RulesEvaluated int `json:"rules_evaluated"`
	RulesPassed    int `json:"rules_passed"`
	RulesFailed    int `json:"rules_failed"`
	RulesSkipped   int `json:"rules_skipped"`

	// Violations and Issues
	Violations      []ComplianceViolation `json:"violations" gorm:"type:jsonb"`
	Warnings        []string              `json:"warnings" gorm:"type:jsonb"`
	Recommendations []string              `json:"recommendations" gorm:"type:jsonb"`

	// Processing Details
	ProcessingTime float64    `json:"processing_time_ms"`
	EvaluatedAt    time.Time  `json:"evaluated_at" gorm:"autoCreateTime"`
	EvaluatedBy    string     `json:"evaluated_by" gorm:"not null"`
	NextReviewAt   *time.Time `json:"next_review_at,omitempty"`
}

// ComplianceViolation represents a specific compliance violation
type ComplianceViolation struct {
	RuleID            string                 `json:"rule_id"`
	RuleName          string                 `json:"rule_name"`
	RuleCategory      string                 `json:"rule_category"`
	Severity          string                 `json:"severity"` // low, medium, high, critical
	Description       string                 `json:"description"`
	ViolationType     string                 `json:"violation_type"` // threshold_exceeded, pattern_detected, blacklist_match
	Details           map[string]interface{} `json:"details"`
	RecommendedAction string                 `json:"recommended_action"`
	RequiresReview    bool                   `json:"requires_review"`
}

// ComplianceViolationAlert represents an alert for compliance violations
type ComplianceViolationAlert struct {
	ID            string `json:"id" gorm:"primaryKey"`
	ViolationID   string `json:"violation_id" gorm:"not null;index"`
	TransactionID string `json:"transaction_id" gorm:"not null;index"`
	EnterpriseID  string `json:"enterprise_id" gorm:"not null;index"`

	// Alert Details
	AlertType string                 `json:"alert_type" gorm:"type:varchar(50);not null"` // immediate, daily, weekly
	Severity  string                 `json:"severity" gorm:"type:varchar(20);not null"`   // low, medium, high, critical
	Message   string                 `json:"message" gorm:"type:text"`
	Details   map[string]interface{} `json:"details" gorm:"type:jsonb"`

	// Status and Escalation
	Status          string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // active, acknowledged, resolved, escalated
	EscalationLevel int        `json:"escalation_level" gorm:"type:int;not null;default:1"`
	AcknowledgedBy  *string    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedBy      *string    `json:"resolved_by,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// ComprehensiveComplianceReport represents a comprehensive compliance report
type ComprehensiveComplianceReport struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	EnterpriseID string    `json:"enterprise_id" gorm:"not null;index"`
	ReportType   string    `json:"report_type" gorm:"type:varchar(50);not null"` // daily, weekly, monthly, quarterly, annual
	PeriodStart  time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd    time.Time `json:"period_end" gorm:"not null"`

	// Compliance Metrics
	TotalTransactions        int64   `json:"total_transactions"`
	CompliantTransactions    int64   `json:"compliant_transactions"`
	NonCompliantTransactions int64   `json:"non_compliant_transactions"`
	ComplianceRate           float64 `json:"compliance_rate" gorm:"type:decimal(5,4)"`

	// Violation Summary
	TotalViolations    int64 `json:"total_violations"`
	CriticalViolations int64 `json:"critical_violations"`
	HighViolations     int64 `json:"high_violations"`
	MediumViolations   int64 `json:"medium_violations"`
	LowViolations      int64 `json:"low_violations"`

	// Rule Performance
	RulesEvaluated        int64   `json:"rules_evaluated"`
	RulesTriggered        int64   `json:"rules_triggered"`
	AverageProcessingTime float64 `json:"average_processing_time_ms"`

	// Risk Assessment
	RiskScore      float64  `json:"risk_score" gorm:"type:decimal(5,4)"`
	RiskTrend      string   `json:"risk_trend" gorm:"type:varchar(20)"` // increasing, decreasing, stable
	TopRiskFactors []string `json:"top_risk_factors" gorm:"type:jsonb"`

	// Recommendations
	Recommendations []string `json:"recommendations" gorm:"type:jsonb"`
	ActionItems     []string `json:"action_items" gorm:"type:jsonb"`

	// Metadata
	GeneratedBy  string    `json:"generated_by" gorm:"not null"`
	GeneratedAt  time.Time `json:"generated_at" gorm:"autoCreateTime"`
	ReportFormat string    `json:"report_format" gorm:"type:varchar(20);default:'json'"` // json, pdf, csv
}

// ComplianceTrend represents compliance trend analysis
type ComplianceTrend struct {
	EnterpriseID string    `json:"enterprise_id"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	TrendType    string    `json:"trend_type"` // compliance_rate, violation_count, risk_score

	// Trend Data
	DataPoints      []ComplianceDataPoint `json:"data_points"`
	TrendDirection  string                `json:"trend_direction"`  // up, down, stable
	TrendStrength   float64               `json:"trend_strength"`   // 0.0 to 1.0
	ConfidenceLevel float64               `json:"confidence_level"` // 0.0 to 1.0

	// Analysis
	KeyInsights []string `json:"key_insights"`
	Anomalies   []string `json:"anomalies"`
	Predictions []string `json:"predictions"`
}

// ComplianceDataPoint represents a single data point in compliance trend analysis
type ComplianceDataPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Baseline   float64   `json:"baseline"`
	IsAnomaly  bool      `json:"is_anomaly"`
	Confidence float64   `json:"confidence"`
}

// ComplianceConfiguration represents compliance system configuration
type ComplianceConfiguration struct {
	ID                string `json:"id" gorm:"primaryKey"`
	EnterpriseID      string `json:"enterprise_id" gorm:"not null;index"`
	ConfigurationType string `json:"configuration_type" gorm:"type:varchar(50);not null"` // validation_levels, alerting, escalation

	// Configuration Settings
	Settings         map[string]interface{} `json:"settings" gorm:"type:jsonb"`
	ValidationLevels map[string]interface{} `json:"validation_levels" gorm:"type:jsonb"`
	AlertThresholds  map[string]interface{} `json:"alert_thresholds" gorm:"type:jsonb"`
	EscalationRules  map[string]interface{} `json:"escalation_rules" gorm:"type:jsonb"`

	// Status and Versioning
	Status      string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Version     int        `json:"version" gorm:"type:int;not null;default:1"`
	EffectiveAt time.Time  `json:"effective_at" gorm:"not null"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

	// Metadata
	CreatedBy string    `json:"created_by" gorm:"not null"`
	UpdatedBy string    `json:"updated_by" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Value implements the driver.Valuer interface for GORM JSON storage
func (cv ComplianceViolation) Value() (interface{}, error) {
	return json.Marshal(cv)
}

// Scan implements the sql.Scanner interface for GORM JSON storage
func (cv *ComplianceViolation) Scan(value interface{}) error {
	if value == nil {
		*cv = ComplianceViolation{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, cv)
}
