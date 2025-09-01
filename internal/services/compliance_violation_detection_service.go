package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// ComplianceViolationDetectionService provides violation detection and alerting functionality
type ComplianceViolationDetectionService struct {
	complianceRepo  repository.ComplianceRepositoryInterface
	auditRepo       repository.AuditRepositoryInterface
	alertThresholds map[string]interface{}
}

// NewComplianceViolationDetectionService creates a new violation detection service
func NewComplianceViolationDetectionService(
	complianceRepo repository.ComplianceRepositoryInterface,
	auditRepo repository.AuditRepositoryInterface,
) *ComplianceViolationDetectionService {
	return &ComplianceViolationDetectionService{
		complianceRepo:  complianceRepo,
		auditRepo:       auditRepo,
		alertThresholds: make(map[string]interface{}),
	}
}

// DetectViolations detects compliance violations in real-time
func (s *ComplianceViolationDetectionService) DetectViolations(
	ctx context.Context,
	complianceResult *models.ComplianceValidationResult,
) ([]*models.ComplianceViolationAlert, error) {
	var alerts []*models.ComplianceViolationAlert

	// Check each violation for alert conditions
	for _, violation := range complianceResult.Violations {
		alert, shouldAlert := s.shouldCreateAlert(violation, complianceResult)
		if shouldAlert {
			alerts = append(alerts, alert)
		}
	}

	// Check for overall compliance score violations
	if complianceResult.ComplianceScore < 0.8 {
		alert := s.createScoreAlert(complianceResult)
		alerts = append(alerts, alert)
	}

	// Check for critical status violations
	if complianceResult.Status == "rejected" {
		alert := s.createStatusAlert(complianceResult)
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// shouldCreateAlert determines if a violation should trigger an alert
func (s *ComplianceViolationDetectionService) shouldCreateAlert(
	violation models.ComplianceViolation,
	result *models.ComplianceValidationResult,
) (*models.ComplianceViolationAlert, bool) {
	// Always alert for critical violations
	if violation.Severity == "critical" {
		return s.createViolationAlert(violation, result, "immediate"), true
	}

	// Alert for high severity violations
	if violation.Severity == "high" {
		return s.createViolationAlert(violation, result, "immediate"), true
	}

	// Alert for medium severity violations that require review
	if violation.Severity == "medium" && violation.RequiresReview {
		return s.createViolationAlert(violation, result, "daily"), true
	}

	// Check threshold-based alerting
	if s.exceedsThreshold(violation, result) {
		return s.createViolationAlert(violation, result, "weekly"), true
	}

	return nil, false
}

// createViolationAlert creates an alert for a specific violation
func (s *ComplianceViolationDetectionService) createViolationAlert(
	violation models.ComplianceViolation,
	result *models.ComplianceValidationResult,
	alertType string,
) *models.ComplianceViolationAlert {
	alert := &models.ComplianceViolationAlert{
		ID:              uuid.New().String(),
		ViolationID:     violation.RuleID,
		TransactionID:   result.TransactionID,
		EnterpriseID:    "enterprise-123", // TODO: Get from transaction
		AlertType:       alertType,
		Severity:        violation.Severity,
		Message:         fmt.Sprintf("Compliance violation detected: %s", violation.Description),
		Details:         violation.Details,
		Status:          "active",
		EscalationLevel: s.getEscalationLevel(violation.Severity),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set expiration based on alert type
	switch alertType {
	case "immediate":
		alert.ExpiresAt = &time.Time{}
		*alert.ExpiresAt = time.Now().Add(24 * time.Hour)
	case "daily":
		alert.ExpiresAt = &time.Time{}
		*alert.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	case "weekly":
		alert.ExpiresAt = &time.Time{}
		*alert.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	}

	return alert
}

// createScoreAlert creates an alert for low compliance score
func (s *ComplianceViolationDetectionService) createScoreAlert(
	result *models.ComplianceValidationResult,
) *models.ComplianceViolationAlert {
	alert := &models.ComplianceViolationAlert{
		ID:              uuid.New().String(),
		ViolationID:     "score_violation",
		TransactionID:   result.TransactionID,
		EnterpriseID:    "enterprise-123", // TODO: Get from transaction
		AlertType:       "daily",
		Severity:        "medium",
		Message:         fmt.Sprintf("Low compliance score detected: %.2f", result.ComplianceScore),
		Details:         map[string]interface{}{"compliance_score": result.ComplianceScore},
		Status:          "active",
		EscalationLevel: 2,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	alert.ExpiresAt = &expiresAt

	return alert
}

// createStatusAlert creates an alert for rejected status
func (s *ComplianceViolationDetectionService) createStatusAlert(
	result *models.ComplianceValidationResult,
) *models.ComplianceViolationAlert {
	alert := &models.ComplianceViolationAlert{
		ID:              uuid.New().String(),
		ViolationID:     "status_violation",
		TransactionID:   result.TransactionID,
		EnterpriseID:    "enterprise-123", // TODO: Get from transaction
		AlertType:       "immediate",
		Severity:        "critical",
		Message:         "Transaction compliance status rejected",
		Details:         map[string]interface{}{"status": result.Status},
		Status:          "active",
		EscalationLevel: 3,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	alert.ExpiresAt = &expiresAt

	return alert
}

// exceedsThreshold checks if a violation exceeds configured thresholds
func (s *ComplianceViolationDetectionService) exceedsThreshold(
	violation models.ComplianceViolation,
	_ *models.ComplianceValidationResult,
) bool {
	// Check if we have threshold configuration for this violation type
	if _, exists := s.alertThresholds[violation.ViolationType]; exists {
		// This would implement threshold checking logic
		// For now, return false to avoid excessive alerts
		return false
	}

	return false
}

// getEscalationLevel determines the escalation level based on severity
func (s *ComplianceViolationDetectionService) getEscalationLevel(severity string) int {
	switch severity {
	case "critical":
		return 3
	case "high":
		return 2
	case "medium":
		return 1
	default:
		return 1
	}
}

// MonitorComplianceViolations continuously monitors for compliance violations
func (s *ComplianceViolationDetectionService) MonitorComplianceViolations(
	_ context.Context,
	enterpriseID string,
) error {
	// Get recent compliance statuses
	flaggedTransactions, err := s.complianceRepo.GetComplianceStatusesByStatus("flagged", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get flagged transactions: %w", err)
	}

	// Check for violations that need escalation
	for _, tx := range flaggedTransactions {
		if s.needsEscalation(tx) {
			err := s.escalateViolation(context.Background(), tx)
			if err != nil {
				// Log error but continue processing
				fmt.Printf("Failed to escalate violation %s: %v\n", tx.ID, err)
			}
		}
	}

	return nil
}

// needsEscalation determines if a compliance status needs escalation
func (s *ComplianceViolationDetectionService) needsEscalation(
	status models.TransactionComplianceStatus,
) bool {
	// Check if status has been flagged for more than 24 hours without review
	if status.ReviewedBy == nil {
		timeSinceCreation := time.Since(status.CreatedAt)
		return timeSinceCreation > 24*time.Hour
	}

	return false
}

// escalateViolation escalates a compliance violation
func (s *ComplianceViolationDetectionService) escalateViolation(
	ctx context.Context,
	status models.TransactionComplianceStatus,
) error {
	// Create escalation alert
	alert := &models.ComplianceViolationAlert{
		ID:              uuid.New().String(),
		ViolationID:     "escalation",
		TransactionID:   status.TransactionID,
		EnterpriseID:    "enterprise-123", // TODO: Get from transaction
		AlertType:       "immediate",
		Severity:        "high",
		Message:         "Compliance violation requires immediate attention",
		Details:         map[string]interface{}{"escalation_reason": "unreviewed_violation"},
		Status:          "active",
		EscalationLevel: 3,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	expiresAt := time.Now().Add(12 * time.Hour)
	alert.ExpiresAt = &expiresAt

	// Log escalation event
	auditLog := &models.AuditLog{
		UserID:       uuid.MustParse("system"),
		EnterpriseID: &uuid.UUID{}, // Will be set below
		Action:       "compliance_escalation",
		Resource:     "compliance_violation",
		ResourceID:   &status.ID,
		Details:      "Compliance violation escalated due to lack of review",
		IPAddress:    "system",
		UserAgent:    "violation_detection_service",
		Success:      true,
	}

	// TODO: Set enterprise ID properly
	// if enterpriseUUID, err := uuid.Parse(enterpriseID); err == nil {
	//     auditLog.EnterpriseID = &enterpriseUUID
	// }

	return s.auditRepo.CreateAuditLog(auditLog)
}

// GetActiveAlerts retrieves active compliance alerts
func (s *ComplianceViolationDetectionService) GetActiveAlerts(
	ctx context.Context,
	enterpriseID string,
	severity string,
) ([]*models.ComplianceViolationAlert, error) {
	// This would typically query a database for active alerts
	// For now, return an empty slice
	return []*models.ComplianceViolationAlert{}, nil
}

// AcknowledgeAlert acknowledges a compliance alert
func (s *ComplianceViolationDetectionService) AcknowledgeAlert(
	ctx context.Context,
	alertID string,
	acknowledgedBy string,
) error {
	// This would typically update the alert status in the database
	// For now, just log the acknowledgment
	fmt.Printf("Alert %s acknowledged by %s\n", alertID, acknowledgedBy)
	return nil
}

// ResolveAlert resolves a compliance alert
func (s *ComplianceViolationDetectionService) ResolveAlert(
	ctx context.Context,
	alertID string,
	resolvedBy string,
	resolutionDetails string,
) error {
	// This would typically update the alert status in the database
	// For now, just log the resolution
	fmt.Printf("Alert %s resolved by %s: %s\n", alertID, resolvedBy, resolutionDetails)
	return nil
}

// SetAlertThresholds sets alert thresholds for different violation types
func (s *ComplianceViolationDetectionService) SetAlertThresholds(thresholds map[string]interface{}) {
	s.alertThresholds = thresholds
}

// GetAlertStatistics retrieves alert statistics
func (s *ComplianceViolationDetectionService) GetAlertStatistics(
	ctx context.Context,
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
) (map[string]interface{}, error) {
	// This would typically calculate alert statistics from the database
	// For now, return mock statistics
	stats := map[string]interface{}{
		"total_alerts":            0,
		"active_alerts":           0,
		"acknowledged_alerts":     0,
		"resolved_alerts":         0,
		"critical_alerts":         0,
		"high_alerts":             0,
		"medium_alerts":           0,
		"low_alerts":              0,
		"average_resolution_time": 0,
		"escalation_count":        0,
	}

	return stats, nil
}
