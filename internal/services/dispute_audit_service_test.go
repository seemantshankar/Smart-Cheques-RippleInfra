package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDisputeAuditService_EventTypes(t *testing.T) {
	// Test audit event type constants
	assert.Equal(t, "initiated", string(DisputeResolutionAuditEventTypeInitiated))
	assert.Equal(t, "ai_review", string(DisputeResolutionAuditEventTypeAIReview))
	assert.Equal(t, "stakeholder_review", string(DisputeResolutionAuditEventTypeStakeholderReview))
	assert.Equal(t, "human_arbitration", string(DisputeResolutionAuditEventTypeHumanArbitration))
	assert.Equal(t, "resolution_executed", string(DisputeResolutionAuditEventTypeResolutionExecuted))
	assert.Equal(t, "funds_frozen", string(DisputeResolutionAuditEventTypeFundsFrozen))
	assert.Equal(t, "funds_unfrozen", string(DisputeResolutionAuditEventTypeFundsUnfrozen))
	assert.Equal(t, "refund_processed", string(DisputeResolutionAuditEventTypeRefundProcessed))
	assert.Equal(t, "partial_payment_processed", string(DisputeResolutionAuditEventTypePartialPaymentProcessed))
	assert.Equal(t, "compliance_violation", string(DisputeResolutionAuditEventTypeComplianceViolation))
	assert.Equal(t, "risk_threshold_exceeded", string(DisputeResolutionAuditEventTypeRiskThresholdExceeded))
}

func TestDisputeAuditService_ComplianceSeverity(t *testing.T) {
	// Test compliance severity constants
	assert.Equal(t, "low", string(DisputeComplianceSeverityLow))
	assert.Equal(t, "medium", string(DisputeComplianceSeverityMedium))
	assert.Equal(t, "high", string(DisputeComplianceSeverityHigh))
	assert.Equal(t, "critical", string(DisputeComplianceSeverityCritical))
}

func TestDisputeAuditService_ComplianceReportType(t *testing.T) {
	// Test compliance report type constants
	assert.Equal(t, "regulatory", string(DisputeComplianceReportTypeRegulatory))
	assert.Equal(t, "internal", string(DisputeComplianceReportTypeInternal))
	assert.Equal(t, "audit", string(DisputeComplianceReportTypeAudit))
	assert.Equal(t, "risk", string(DisputeComplianceReportTypeRisk))
}

func TestDisputeAuditService_DashboardViewType(t *testing.T) {
	// Test dashboard view type constants
	assert.Equal(t, "overview", string(DisputeDashboardViewTypeOverview))
	assert.Equal(t, "performance", string(DisputeDashboardViewTypePerformance))
	assert.Equal(t, "compliance", string(DisputeDashboardViewTypeCompliance))
	assert.Equal(t, "risk", string(DisputeDashboardViewTypeRisk))
}

func TestDisputeAuditService_RecommendationType(t *testing.T) {
	// Test recommendation type constants
	assert.Equal(t, "process", string(DisputeRecommendationTypeProcess))
	assert.Equal(t, "technology", string(DisputeRecommendationTypeTechnology))
	assert.Equal(t, "training", string(DisputeRecommendationTypeTraining))
	assert.Equal(t, "policy", string(DisputeRecommendationTypePolicy))
}

func TestDisputeAuditService_RecommendationPriority(t *testing.T) {
	// Test recommendation priority constants
	assert.Equal(t, "low", string(DisputeRecommendationPriorityLow))
	assert.Equal(t, "medium", string(DisputeRecommendationPriorityMedium))
	assert.Equal(t, "high", string(DisputeRecommendationPriorityHigh))
	assert.Equal(t, "critical", string(DisputeRecommendationPriorityCritical))
}

func TestDisputeAuditService_RecommendationStatus(t *testing.T) {
	// Test recommendation status constants
	assert.Equal(t, "open", string(DisputeRecommendationStatusOpen))
	assert.Equal(t, "in_progress", string(DisputeRecommendationStatusInProgress))
	assert.Equal(t, "completed", string(DisputeRecommendationStatusCompleted))
	assert.Equal(t, "rejected", string(DisputeRecommendationStatusRejected))
}

func TestDisputeAuditService_MitigationStatus(t *testing.T) {
	// Test mitigation status constants
	assert.Equal(t, "planned", string(DisputeMitigationStatusPlanned))
	assert.Equal(t, "in_progress", string(DisputeMitigationStatusInProgress))
	assert.Equal(t, "completed", string(DisputeMitigationStatusCompleted))
	assert.Equal(t, "failed", string(DisputeMitigationStatusFailed))
}

func TestDisputeAuditService_ViolationStatus(t *testing.T) {
	// Test violation status constants
	assert.Equal(t, "open", string(DisputeViolationStatusOpen))
	assert.Equal(t, "in_progress", string(DisputeViolationStatusInProgress))
	assert.Equal(t, "resolved", string(DisputeViolationStatusResolved))
	assert.Equal(t, "closed", string(DisputeViolationStatusClosed))
}

func TestDisputeAuditService_IncidentStatus(t *testing.T) {
	// Test incident status constants
	assert.Equal(t, "open", string(DisputeIncidentStatusOpen))
	assert.Equal(t, "in_progress", string(DisputeIncidentStatusInProgress))
	assert.Equal(t, "resolved", string(DisputeIncidentStatusResolved))
	assert.Equal(t, "closed", string(DisputeIncidentStatusClosed))
}

func TestDisputeAuditService_ValidateAuditEvent(t *testing.T) {
	service := &DisputeAuditService{}

	// Valid event
	validEvent := &DisputeResolutionAuditEvent{
		DisputeID:    "dispute-123",
		EventType:    DisputeResolutionAuditEventTypeInitiated,
		UserID:       "user-123",
		EnterpriseID: "enterprise-123",
	}

	err := service.validateAuditEvent(validEvent)
	assert.NoError(t, err)

	// Invalid: missing dispute ID
	invalidDisputeID := &DisputeResolutionAuditEvent{
		EventType:    DisputeResolutionAuditEventTypeInitiated,
		UserID:       "user-123",
		EnterpriseID: "enterprise-123",
	}

	err = service.validateAuditEvent(invalidDisputeID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute ID is required")

	// Invalid: missing event type
	invalidEventType := &DisputeResolutionAuditEvent{
		DisputeID:    "dispute-123",
		UserID:       "user-123",
		EnterpriseID: "enterprise-123",
	}

	err = service.validateAuditEvent(invalidEventType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event type is required")

	// Invalid: missing user ID
	invalidUserID := &DisputeResolutionAuditEvent{
		DisputeID:    "dispute-123",
		EventType:    DisputeResolutionAuditEventTypeInitiated,
		EnterpriseID: "enterprise-123",
	}

	err = service.validateAuditEvent(invalidUserID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")

	// Invalid: missing enterprise ID
	invalidEnterpriseID := &DisputeResolutionAuditEvent{
		DisputeID: "dispute-123",
		EventType: DisputeResolutionAuditEventTypeInitiated,
		UserID:    "user-123",
	}

	err = service.validateAuditEvent(invalidEnterpriseID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enterprise ID is required")
}

func TestDisputeAuditService_DataRetentionPolicy(t *testing.T) {
	// Test DataRetentionPolicy struct creation
	now := time.Now()
	policy := &DataRetentionPolicy{
		PolicyID:        "policy-123",
		PolicyName:      "Dispute Resolution Retention",
		Description:     "Retention policy for dispute resolution data",
		RetentionPeriod: 7 * 24 * time.Hour, // 7 days
		DataTypes:       []string{"audit_logs", "compliance_reports", "performance_metrics"},
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, "policy-123", policy.PolicyID)
	assert.Equal(t, "Dispute Resolution Retention", policy.PolicyName)
	assert.Equal(t, "Retention policy for dispute resolution data", policy.Description)
	assert.Equal(t, 7*24*time.Hour, policy.RetentionPeriod)
	assert.Equal(t, 3, len(policy.DataTypes))
	assert.Equal(t, now, policy.CreatedAt)
	assert.Equal(t, now, policy.UpdatedAt)
}

func TestDisputeAuditService_ComplianceRule(t *testing.T) {
	// Test DisputeComplianceRule struct creation
	rule := &DisputeComplianceRule{
		RuleID:          "rule-123",
		Regulation:      "GDPR",
		Requirement:     "Data retention for 7 years",
		RetentionPeriod: 7 * 365 * 24 * time.Hour, // 7 years
		Description:     "GDPR requires retention of financial data for 7 years",
	}

	assert.Equal(t, "rule-123", rule.RuleID)
	assert.Equal(t, "GDPR", rule.Regulation)
	assert.Equal(t, "Data retention for 7 years", rule.Requirement)
	assert.Equal(t, 7*365*24*time.Hour, rule.RetentionPeriod)
	assert.Equal(t, "GDPR requires retention of financial data for 7 years", rule.Description)
}

func TestDisputeAuditService_TimeSeriesData(t *testing.T) {
	// Test DisputeTimeSeriesData struct creation
	now := time.Now()
	timeSeriesData := &DisputeTimeSeriesData{
		Timestamp: now,
		Value:     85.5,
		Label:     "Success Rate",
	}

	assert.Equal(t, now, timeSeriesData.Timestamp)
	assert.Equal(t, 85.5, timeSeriesData.Value)
	assert.Equal(t, "Success Rate", timeSeriesData.Label)
}

func TestDisputeAuditService_TopPerformer(t *testing.T) {
	// Test DisputeTopPerformer struct creation
	topPerformer := &DisputeTopPerformer{
		UserID:                "user-123",
		UserName:              "John Doe",
		PerformanceScore:      95.5,
		DisputesResolved:      25,
		AverageResolutionTime: 2 * time.Hour,
	}

	assert.Equal(t, "user-123", topPerformer.UserID)
	assert.Equal(t, "John Doe", topPerformer.UserName)
	assert.Equal(t, 95.5, topPerformer.PerformanceScore)
	assert.Equal(t, int64(25), topPerformer.DisputesResolved)
	assert.Equal(t, 2*time.Hour, topPerformer.AverageResolutionTime)
}
