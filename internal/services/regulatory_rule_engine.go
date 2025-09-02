package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// RegulatoryRuleEngine handles dynamic compliance validation using regulatory rules
type RegulatoryRuleEngine struct {
	ruleRepo repository.RegulatoryRuleRepositoryInterface
}

// NewRegulatoryRuleEngine creates a new regulatory rule engine
func NewRegulatoryRuleEngine(ruleRepo repository.RegulatoryRuleRepositoryInterface) *RegulatoryRuleEngine {
	return &RegulatoryRuleEngine{
		ruleRepo: ruleRepo,
	}
}

// EvaluateCompliance evaluates compliance for a transaction using regulatory rules
func (e *RegulatoryRuleEngine) EvaluateCompliance(ctx context.Context, request *models.ComplianceValidationRequest) (*models.ComplianceValidationResult, error) {
	startTime := time.Now()

	// Get active regulatory rules for the jurisdiction
	rules, err := e.ruleRepo.GetActiveRegulatoryRules(ctx, request.Jurisdiction)
	if err != nil {
		return nil, fmt.Errorf("failed to get regulatory rules: %w", err)
	}

	result := &models.ComplianceValidationResult{
		ID:              uuid.New().String(),
		TransactionID:   request.TransactionID,
		ValidationLevel: request.ValidationLevel,
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
		EvaluatedAt:     time.Now(),
		EvaluatedBy:     "regulatory_rule_engine",
	}

	// Evaluate each rule
	for _, rule := range rules {
		violation, passed, skipped := e.evaluateRule(rule, request)

		result.RulesEvaluated++

		if skipped {
			result.RulesSkipped++
			continue
		}

		if passed {
			result.RulesPassed++
		} else {
			result.RulesFailed++
			result.Violations = append(result.Violations, *violation)

			// Update compliance score based on rule priority
			scoreReduction := float64(rule.Priority) * 0.1
			result.ComplianceScore -= scoreReduction
			if result.ComplianceScore < 0 {
				result.ComplianceScore = 0
			}
		}
	}

	// Determine overall status based on violations and compliance score
	result.Status = e.determineComplianceStatus(result)
	result.IsCompliant = result.Status == "approved"
	result.ProcessingTime = float64(time.Since(startTime).Milliseconds())

	// Add recommendations based on violations
	e.addRecommendations(result)

	return result, nil
}

// evaluateRule evaluates a single regulatory rule against the transaction
func (e *RegulatoryRuleEngine) evaluateRule(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	// Check if rule applies to this transaction type
	if !e.ruleAppliesToTransaction(rule, request) {
		return nil, false, true // skipped
	}

	// Evaluate based on rule type
	switch rule.RuleType {
	case "amount_limit":
		return e.evaluateAmountLimit(rule, request)
	case "frequency_limit":
		return e.evaluateFrequencyLimit(rule, request)
	case "pattern_detection":
		return e.evaluatePatternDetection(rule, request)
	case "blacklist_check":
		return e.evaluateBlacklistCheck(rule, request)
	case "threshold_check":
		return e.evaluateThresholdCheck(rule, request)
	default:
		return nil, false, true // unknown rule type, skip
	}
}

// ruleAppliesToTransaction checks if a rule applies to the given transaction
func (e *RegulatoryRuleEngine) ruleAppliesToTransaction(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) bool {
	// Check transaction type conditions
	if conditions, ok := rule.Conditions["transaction_types"]; ok {
		if allowedTypes, ok := conditions.([]interface{}); ok {
			found := false
			for _, t := range allowedTypes {
				if t.(string) == request.TransactionType {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check currency conditions
	if conditions, ok := rule.Conditions["currencies"]; ok {
		if allowedCurrencies, ok := conditions.([]interface{}); ok {
			found := false
			for _, c := range allowedCurrencies {
				if c.(string) == request.Currency {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// evaluateAmountLimit evaluates amount-based rules
func (e *RegulatoryRuleEngine) evaluateAmountLimit(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	amount, err := strconv.ParseFloat(request.Amount, 64)
	if err != nil {
		return nil, false, true // invalid amount, skip
	}

	// Get threshold from rule
	threshold, ok := rule.Thresholds["max_amount"]
	if !ok {
		return nil, false, true // no threshold defined, skip
	}

	maxAmount, ok := threshold.(float64)
	if !ok {
		return nil, false, true // invalid threshold, skip
	}

	if amount > maxAmount {
		violation := &models.ComplianceViolation{
			RuleID:            rule.ID,
			RuleName:          rule.Name,
			RuleCategory:      rule.Category,
			Severity:          e.getSeverityFromPriority(rule.Priority),
			Description:       fmt.Sprintf("Transaction amount %.2f exceeds maximum allowed amount %.2f", amount, maxAmount),
			ViolationType:     "threshold_exceeded",
			Details:           map[string]interface{}{"amount": amount, "max_amount": maxAmount},
			RecommendedAction: "Review transaction for potential risk",
			RequiresReview:    rule.Priority >= 4, // Critical priority requires review
		}
		return violation, false, false
	}

	return nil, true, false
}

// evaluateFrequencyLimit evaluates frequency-based rules
func (e *RegulatoryRuleEngine) evaluateFrequencyLimit(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	// This would typically check transaction history
	// For now, we'll implement a simple check based on metadata
	if metadata, ok := request.Metadata["transaction_count_24h"]; ok {
		if count, ok := metadata.(float64); ok {
			maxCount, ok := rule.Thresholds["max_transactions_24h"].(float64)
			if ok && count > maxCount {
				violation := &models.ComplianceViolation{
					RuleID:            rule.ID,
					RuleName:          rule.Name,
					RuleCategory:      rule.Category,
					Severity:          e.getSeverityFromPriority(rule.Priority),
					Description:       fmt.Sprintf("Transaction count %.0f in 24h exceeds maximum allowed %.0f", count, maxCount),
					ViolationType:     "threshold_exceeded",
					Details:           map[string]interface{}{"count": count, "max_count": maxCount},
					RecommendedAction: "Monitor transaction frequency",
					RequiresReview:    rule.Priority >= 3,
				}
				return violation, false, false
			}
		}
	}

	return nil, true, false
}

// evaluatePatternDetection evaluates pattern-based rules
func (e *RegulatoryRuleEngine) evaluatePatternDetection(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	// Check for suspicious patterns
	patterns, ok := rule.Conditions["suspicious_patterns"].([]interface{})
	if !ok {
		return nil, true, false
	}

	for _, pattern := range patterns {
		if patternStr, ok := pattern.(string); ok {
			if e.matchesPattern(request, patternStr) {
				violation := &models.ComplianceViolation{
					RuleID:            rule.ID,
					RuleName:          rule.Name,
					RuleCategory:      rule.Category,
					Severity:          e.getSeverityFromPriority(rule.Priority),
					Description:       fmt.Sprintf("Transaction matches suspicious pattern: %s", patternStr),
					ViolationType:     "pattern_detected",
					Details:           map[string]interface{}{"pattern": patternStr},
					RecommendedAction: "Investigate transaction pattern",
					RequiresReview:    true,
				}
				return violation, false, false
			}
		}
	}

	return nil, true, false
}

// evaluateBlacklistCheck evaluates blacklist-based rules
func (e *RegulatoryRuleEngine) evaluateBlacklistCheck(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	// Check if addresses are in blacklist
	blacklist, ok := rule.Conditions["blacklisted_addresses"].([]interface{})
	if !ok {
		return nil, true, false
	}

	for _, addr := range blacklist {
		if addrStr, ok := addr.(string); ok {
			if request.FromAddress == addrStr || request.ToAddress == addrStr {
				violation := &models.ComplianceViolation{
					RuleID:            rule.ID,
					RuleName:          rule.Name,
					RuleCategory:      rule.Category,
					Severity:          "critical",
					Description:       fmt.Sprintf("Transaction involves blacklisted address: %s", addrStr),
					ViolationType:     "blacklist_match",
					Details:           map[string]interface{}{"blacklisted_address": addrStr},
					RecommendedAction: "Immediately block transaction",
					RequiresReview:    true,
				}
				return violation, false, false
			}
		}
	}

	return nil, true, false
}

// evaluateThresholdCheck evaluates general threshold-based rules
func (e *RegulatoryRuleEngine) evaluateThresholdCheck(rule *models.RegulatoryRule, request *models.ComplianceValidationRequest) (*models.ComplianceViolation, bool, bool) {
	// Generic threshold evaluation
	for key, threshold := range rule.Thresholds {
		if value, exists := request.Metadata[key]; exists {
			if thresholdFloat, ok := threshold.(float64); ok {
				if valueFloat, ok := value.(float64); ok {
					if valueFloat > thresholdFloat {
						violation := &models.ComplianceViolation{
							RuleID:            rule.ID,
							RuleName:          rule.Name,
							RuleCategory:      rule.Category,
							Severity:          e.getSeverityFromPriority(rule.Priority),
							Description:       fmt.Sprintf("Value %.2f exceeds threshold %.2f for %s", valueFloat, thresholdFloat, key),
							ViolationType:     "threshold_exceeded",
							Details:           map[string]interface{}{"value": valueFloat, "threshold": thresholdFloat, "metric": key},
							RecommendedAction: "Review transaction metrics",
							RequiresReview:    rule.Priority >= 3,
						}
						return violation, false, false
					}
				}
			}
		}
	}

	return nil, true, false
}

// matchesPattern checks if transaction matches a suspicious pattern
func (e *RegulatoryRuleEngine) matchesPattern(request *models.ComplianceValidationRequest, pattern string) bool {
	// Simple pattern matching - in production this would be more sophisticated
	switch pattern {
	case "round_amount":
		amount, err := strconv.ParseFloat(request.Amount, 64)
		if err != nil {
			return false
		}
		return amount == float64(int(amount)) && amount > 1000
	case "high_frequency":
		if count, ok := request.Metadata["transaction_count_24h"].(float64); ok {
			return count > 10
		}
		return false
	case "unusual_hours":
		// Check if transaction is outside business hours
		hour := time.Now().Hour()
		return hour < 6 || hour > 22
	default:
		return false
	}
}

// getSeverityFromPriority converts rule priority to severity level
func (e *RegulatoryRuleEngine) getSeverityFromPriority(priority int) string {
	switch priority {
	case 1:
		return "low"
	case 2:
		return "low"
	case 3:
		return "medium"
	case 4:
		return "high"
	case 5:
		return "critical"
	default:
		return "medium"
	}
}

// determineComplianceStatus determines the overall compliance status
func (e *RegulatoryRuleEngine) determineComplianceStatus(result *models.ComplianceValidationResult) string {
	// Check for critical violations
	for _, violation := range result.Violations {
		if violation.Severity == "critical" {
			return "rejected"
		}
	}

	// Check compliance score
	if result.ComplianceScore < 0.5 {
		return "rejected"
	} else if result.ComplianceScore < 0.8 {
		return "flagged"
	} else if len(result.Violations) > 0 {
		return "flagged"
	}

	return "approved"
}

// addRecommendations adds recommendations based on violations
func (e *RegulatoryRuleEngine) addRecommendations(result *models.ComplianceValidationResult) {
	if len(result.Violations) == 0 {
		result.Recommendations = append(result.Recommendations, "Transaction appears compliant")
		return
	}

	// Add general recommendations
	if result.ComplianceScore < 0.8 {
		result.Recommendations = append(result.Recommendations, "Consider implementing additional monitoring")
	}

	// Add specific recommendations based on violation types
	hasAmountViolations := false
	hasPatternViolations := false
	hasBlacklistViolations := false

	for _, violation := range result.Violations {
		switch violation.ViolationType {
		case "threshold_exceeded":
			hasAmountViolations = true
		case "pattern_detected":
			hasPatternViolations = true
		case "blacklist_match":
			hasBlacklistViolations = true
		}
	}

	if hasAmountViolations {
		result.Recommendations = append(result.Recommendations, "Review transaction amount limits")
	}
	if hasPatternViolations {
		result.Recommendations = append(result.Recommendations, "Investigate transaction patterns")
	}
	if hasBlacklistViolations {
		result.Recommendations = append(result.Recommendations, "Immediate action required for blacklisted addresses")
	}
}
