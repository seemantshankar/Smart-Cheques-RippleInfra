package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// EnhancedComplianceValidationService provides comprehensive compliance validation
type EnhancedComplianceValidationService struct {
	ruleEngine      *RegulatoryRuleEngine
	complianceRepo  repository.ComplianceRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	auditRepo       repository.AuditRepositoryInterface
}

// NewEnhancedComplianceValidationService creates a new enhanced compliance validation service
func NewEnhancedComplianceValidationService(
	ruleEngine *RegulatoryRuleEngine,
	complianceRepo repository.ComplianceRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	auditRepo repository.AuditRepositoryInterface,
) *EnhancedComplianceValidationService {
	return &EnhancedComplianceValidationService{
		ruleEngine:      ruleEngine,
		complianceRepo:  complianceRepo,
		transactionRepo: transactionRepo,
		auditRepo:       auditRepo,
	}
}

// ValidateTransactionCompliance performs comprehensive compliance validation for a transaction
func (s *EnhancedComplianceValidationService) ValidateTransactionCompliance(
	ctx context.Context,
	transaction *models.Transaction,
	validationLevel string,
) (*models.ComplianceValidationResult, error) {
	startTime := time.Now()

	// Create validation request
	request := &models.ComplianceValidationRequest{
		TransactionID:   transaction.ID,
		EnterpriseID:    transaction.EnterpriseID,
		Jurisdiction:    "US", // TODO: Get from enterprise profile
		TransactionType: string(transaction.Type),
		Amount:          transaction.Amount,
		Currency:        transaction.Currency,
		FromAddress:     transaction.FromAddress,
		ToAddress:       transaction.ToAddress,
		Metadata:        transaction.Metadata,
		ValidationLevel: validationLevel,
	}

	// Perform regulatory rule evaluation
	regulatoryResult, err := s.ruleEngine.EvaluateCompliance(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate regulatory compliance: %w", err)
	}

	// Perform basic compliance checks
	basicResult := s.performBasicComplianceChecks(transaction)

	// Perform enhanced checks based on validation level
	var enhancedResult *models.ComplianceValidationResult
	switch validationLevel {
	case "enhanced":
		enhancedResult = s.performEnhancedComplianceChecks(transaction)
	case "strict":
		enhancedResult = s.performStrictComplianceChecks(transaction)
	default:
		enhancedResult = &models.ComplianceValidationResult{
			IsCompliant: true,
			Status:      "approved",
		}
	}

	// Merge results
	finalResult := s.mergeComplianceResults(regulatoryResult, basicResult, enhancedResult)
	finalResult.ProcessingTime = float64(time.Since(startTime).Milliseconds())

	// Store compliance result
	err = s.storeComplianceResult(ctx, finalResult)
	if err != nil {
		return nil, fmt.Errorf("failed to store compliance result: %w", err)
	}

	// Log audit event
	err = s.logComplianceAuditEvent(transaction, finalResult)
	if err != nil {
		// Don't fail the validation for audit logging errors
		fmt.Printf("Warning: Failed to log compliance audit event: %v\n", err)
	}

	return finalResult, nil
}

// performBasicComplianceChecks performs basic compliance validation
func (s *EnhancedComplianceValidationService) performBasicComplianceChecks(transaction *models.Transaction) *models.ComplianceValidationResult {
	result := &models.ComplianceValidationResult{
		IsCompliant:     true,
		ComplianceScore: 1.0,
		Status:          "approved",
		RulesEvaluated:  0,
		RulesPassed:     0,
		RulesFailed:     0,
		RulesSkipped:    0,
		Violations:      []models.ComplianceViolation{},
		Warnings:        []string{},
		Recommendations: []string{},
	}

	// Basic validation checks
	if transaction.Amount == "" {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "basic_amount_check",
			RuleName:          "Amount Required",
			RuleCategory:      "basic",
			Severity:          "medium",
			Description:       "Transaction amount is required",
			ViolationType:     "missing_field",
			RecommendedAction: "Provide valid transaction amount",
			RequiresReview:    false,
		})
		result.IsCompliant = false
	}

	if transaction.FromAddress == "" || transaction.ToAddress == "" {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "basic_address_check",
			RuleName:          "Valid Addresses Required",
			RuleCategory:      "basic",
			Severity:          "medium",
			Description:       "Valid from and to addresses are required",
			ViolationType:     "missing_field",
			RecommendedAction: "Provide valid addresses",
			RequiresReview:    false,
		})
		result.IsCompliant = false
	}

	if transaction.EnterpriseID == "" {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "basic_enterprise_check",
			RuleName:          "Enterprise Association Required",
			RuleCategory:      "basic",
			Severity:          "high",
			Description:       "Enterprise association is required",
			ViolationType:     "missing_field",
			RecommendedAction: "Associate transaction with enterprise",
			RequiresReview:    true,
		})
		result.IsCompliant = false
	}

	// Update status based on violations
	if len(result.Violations) > 0 {
		result.Status = "flagged"
		result.ComplianceScore = 0.8
	}

	return result
}

// performEnhancedComplianceChecks performs enhanced compliance validation
func (s *EnhancedComplianceValidationService) performEnhancedComplianceChecks(transaction *models.Transaction) *models.ComplianceValidationResult {
	result := &models.ComplianceValidationResult{
		IsCompliant:     true,
		ComplianceScore: 1.0,
		Status:          "approved",
		RulesEvaluated:  0,
		RulesPassed:     0,
		RulesFailed:     0,
		RulesSkipped:    0,
		Violations:      []models.ComplianceViolation{},
		Warnings:        []string{},
		Recommendations: []string{},
	}

	// Enhanced checks
	if transaction.RetryCount > 2 {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "enhanced_retry_check",
			RuleName:          "High Retry Count",
			RuleCategory:      "enhanced",
			Severity:          "medium",
			Description:       fmt.Sprintf("Transaction has high retry count: %d", transaction.RetryCount),
			ViolationType:     "threshold_exceeded",
			Details:           map[string]interface{}{"retry_count": transaction.RetryCount},
			RecommendedAction: "Investigate transaction failures",
			RequiresReview:    true,
		})
		result.IsCompliant = false
	}

	// Check for unusual transaction patterns
	if s.isUnusualTransactionPattern(transaction) {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "enhanced_pattern_check",
			RuleName:          "Unusual Transaction Pattern",
			RuleCategory:      "enhanced",
			Severity:          "medium",
			Description:       "Transaction exhibits unusual patterns",
			ViolationType:     "pattern_detected",
			RecommendedAction: "Review transaction for potential risk",
			RequiresReview:    true,
		})
		result.IsCompliant = false
	}

	// Update status based on violations
	if len(result.Violations) > 0 {
		result.Status = "flagged"
		result.ComplianceScore = 0.7
	}

	return result
}

// performStrictComplianceChecks performs strict compliance validation
func (s *EnhancedComplianceValidationService) performStrictComplianceChecks(transaction *models.Transaction) *models.ComplianceValidationResult {
	result := &models.ComplianceValidationResult{
		IsCompliant:     true,
		ComplianceScore: 1.0,
		Status:          "approved",
		RulesEvaluated:  0,
		RulesPassed:     0,
		RulesFailed:     0,
		RulesSkipped:    0,
		Violations:      []models.ComplianceViolation{},
		Warnings:        []string{},
		Recommendations: []string{},
	}

	// Strict checks - more stringent requirements
	if transaction.RetryCount > 0 {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "strict_retry_check",
			RuleName:          "No Retries Allowed",
			RuleCategory:      "strict",
			Severity:          "high",
			Description:       "Retries are not allowed in strict mode",
			ViolationType:     "policy_violation",
			Details:           map[string]interface{}{"retry_count": transaction.RetryCount},
			RecommendedAction: "Resolve underlying issues before retry",
			RequiresReview:    true,
		})
		result.IsCompliant = false
	}

	// Check for any suspicious metadata
	if s.hasSuspiciousMetadata(transaction) {
		result.Violations = append(result.Violations, models.ComplianceViolation{
			RuleID:            "strict_metadata_check",
			RuleName:          "Suspicious Metadata",
			RuleCategory:      "strict",
			Severity:          "high",
			Description:       "Transaction contains suspicious metadata",
			ViolationType:     "suspicious_activity",
			RecommendedAction: "Review transaction metadata for compliance",
			RequiresReview:    true,
		})
		result.IsCompliant = false
	}

	// Update status based on violations
	if len(result.Violations) > 0 {
		result.Status = "rejected"
		result.ComplianceScore = 0.5
	}

	return result
}

// mergeComplianceResults merges multiple compliance validation results
func (s *EnhancedComplianceValidationService) mergeComplianceResults(
	regulatory *models.ComplianceValidationResult,
	basic *models.ComplianceValidationResult,
	enhanced *models.ComplianceValidationResult,
) *models.ComplianceValidationResult {
	result := &models.ComplianceValidationResult{
		ID:              uuid.New().String(),
		TransactionID:   regulatory.TransactionID,
		ValidationLevel: regulatory.ValidationLevel,
		IsCompliant:     true,
		ComplianceScore: 1.0,
		Status:          "approved",
		RulesEvaluated:  regulatory.RulesEvaluated + basic.RulesEvaluated + enhanced.RulesEvaluated,
		RulesPassed:     regulatory.RulesPassed + basic.RulesPassed + enhanced.RulesPassed,
		RulesFailed:     regulatory.RulesFailed + basic.RulesFailed + enhanced.RulesFailed,
		RulesSkipped:    regulatory.RulesSkipped + basic.RulesSkipped + enhanced.RulesSkipped,
		Violations:      []models.ComplianceViolation{},
		Warnings:        []string{},
		Recommendations: []string{},
		EvaluatedAt:     time.Now(),
		EvaluatedBy:     "enhanced_compliance_service",
	}

	// Merge violations
	result.Violations = append(result.Violations, regulatory.Violations...)
	result.Violations = append(result.Violations, basic.Violations...)
	result.Violations = append(result.Violations, enhanced.Violations...)

	// Merge warnings
	result.Warnings = append(result.Warnings, regulatory.Warnings...)
	result.Warnings = append(result.Warnings, basic.Warnings...)
	result.Warnings = append(result.Warnings, enhanced.Warnings...)

	// Merge recommendations
	result.Recommendations = append(result.Recommendations, regulatory.Recommendations...)
	result.Recommendations = append(result.Recommendations, basic.Recommendations...)
	result.Recommendations = append(result.Recommendations, enhanced.Recommendations...)

	// Calculate overall compliance score (weighted average)
	totalWeight := 0.5 + 0.3 + 0.2 // regulatory + basic + enhanced
	result.ComplianceScore = (regulatory.ComplianceScore*0.5 + basic.ComplianceScore*0.3 + enhanced.ComplianceScore*0.2) / totalWeight

	// Determine overall status
	if len(result.Violations) == 0 {
		result.Status = "approved"
		result.IsCompliant = true
	} else {
		// Check for critical violations
		hasCritical := false
		for _, violation := range result.Violations {
			if violation.Severity == "critical" {
				hasCritical = true
				break
			}
		}

		if hasCritical || result.ComplianceScore < 0.5 {
			result.Status = "rejected"
			result.IsCompliant = false
		} else if result.ComplianceScore < 0.8 || len(result.Violations) > 0 {
			result.Status = "flagged"
			result.IsCompliant = false
		}
	}

	return result
}

// isUnusualTransactionPattern checks for unusual transaction patterns
func (s *EnhancedComplianceValidationService) isUnusualTransactionPattern(transaction *models.Transaction) bool {
	// Check for unusual hours (outside 6 AM - 10 PM)
	hour := time.Now().Hour()
	if hour < 6 || hour > 22 {
		return true
	}

	// Check for round amounts over threshold
	if amount, err := strconv.ParseFloat(transaction.Amount, 64); err == nil {
		if amount == float64(int(amount)) && amount > 10000 {
			return true
		}
	}

	return false
}

// hasSuspiciousMetadata checks for suspicious metadata
func (s *EnhancedComplianceValidationService) hasSuspiciousMetadata(transaction *models.Transaction) bool {
	if transaction.Metadata == nil {
		return false
	}

	// Check for suspicious keywords in metadata
	suspiciousKeywords := []string{"urgent", "emergency", "rush", "immediate", "secret", "confidential"}
	for key, value := range transaction.Metadata {
		if str, ok := value.(string); ok {
			for _, keyword := range suspiciousKeywords {
				if strings.Contains(strings.ToLower(str), keyword) {
					return true
				}
			}
		}
		if strings.Contains(strings.ToLower(key), "urgent") {
			return true
		}
	}

	return false
}

// storeComplianceResult stores the compliance validation result
func (s *EnhancedComplianceValidationService) storeComplianceResult(_ context.Context, result *models.ComplianceValidationResult) error {
	// Convert to TransactionComplianceStatus for storage
	complianceStatus := &models.TransactionComplianceStatus{
		ID:            result.ID,
		TransactionID: result.TransactionID,
		Status:        result.Status,
		ChecksPassed:  []string{},
		ChecksFailed:  []string{},
		Violations:    []string{},
		CreatedAt:     result.EvaluatedAt,
		UpdatedAt:     result.EvaluatedAt,
	}

	// Convert violations to strings
	for _, violation := range result.Violations {
		complianceStatus.Violations = append(complianceStatus.Violations, violation.Description)
	}

	// Add passed/failed checks
	if result.RulesPassed > 0 {
		complianceStatus.ChecksPassed = append(complianceStatus.ChecksPassed, fmt.Sprintf("regulatory_rules_passed_%d", result.RulesPassed))
	}
	if result.RulesFailed > 0 {
		complianceStatus.ChecksFailed = append(complianceStatus.ChecksFailed, fmt.Sprintf("regulatory_rules_failed_%d", result.RulesFailed))
	}

	return s.complianceRepo.CreateComplianceStatus(complianceStatus)
}

// logComplianceAuditEvent logs the compliance validation event
func (s *EnhancedComplianceValidationService) logComplianceAuditEvent(transaction *models.Transaction, result *models.ComplianceValidationResult) error {
	// Use a fixed system UUID for compliance validation events
	systemUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	auditLog := &models.AuditLog{
		UserID:       systemUUID,
		EnterpriseID: &uuid.UUID{}, // Will be set below
		Action:       "compliance_validation",
		Resource:     "transaction",
		ResourceID:   &transaction.ID,
		Details:      fmt.Sprintf("Compliance validation completed. Status: %s, Score: %.2f", result.Status, result.ComplianceScore),
		IPAddress:    "system",
		UserAgent:    "enhanced_compliance_service",
		Success:      result.IsCompliant,
	}

	if enterpriseUUID, err := uuid.Parse(transaction.EnterpriseID); err == nil {
		auditLog.EnterpriseID = &enterpriseUUID
	}

	return s.auditRepo.CreateAuditLog(auditLog)
}
