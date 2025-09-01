package services

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// CategorizationRuleService provides business logic for managing categorization rules
type CategorizationRuleService struct {
	ruleRepo repository.CategorizationRuleRepositoryInterface
}

// NewCategorizationRuleService creates a new categorization rule service
func NewCategorizationRuleService(ruleRepo repository.CategorizationRuleRepositoryInterface) *CategorizationRuleService {
	return &CategorizationRuleService{
		ruleRepo: ruleRepo,
	}
}

// CreateRule creates a new categorization rule
func (s *CategorizationRuleService) CreateRule(ctx context.Context, rule *models.CategorizationRule, userID string) error {
	// Validate rule
	if err := s.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// Set metadata
	rule.CreatedBy = userID
	rule.UpdatedBy = userID
	rule.IsActive = true // Default to active
	rule.UseCount = 0
	rule.SuccessCount = 0
	rule.PerformanceScore = 0.0

	// Set default values if not provided
	if rule.BaseConfidence == 0 {
		rule.BaseConfidence = 0.5
	}
	if rule.Weight == 0 {
		rule.Weight = 1.0
	}
	if rule.MinConfidence == 0 {
		rule.MinConfidence = 0.1
	}
	if rule.MaxConfidence == 0 {
		rule.MaxConfidence = 0.95
	}

	return s.ruleRepo.CreateRule(ctx, rule)
}

// UpdateRule updates an existing categorization rule
func (s *CategorizationRuleService) UpdateRule(ctx context.Context, rule *models.CategorizationRule, userID string) error {
	// Validate rule
	if err := s.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// Get existing rule
	existing, err := s.ruleRepo.GetRuleByID(ctx, rule.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing rule: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("rule not found")
	}

	// Preserve performance metrics
	rule.UseCount = existing.UseCount
	rule.SuccessCount = existing.SuccessCount
	rule.LastUsedAt = existing.LastUsedAt
	rule.PerformanceScore = existing.PerformanceScore
	rule.UpdatedBy = userID

	return s.ruleRepo.UpdateRule(ctx, rule)
}

// DeleteRule deletes a categorization rule
func (s *CategorizationRuleService) DeleteRule(ctx context.Context, ruleID string) error {
	return s.ruleRepo.DeleteRule(ctx, ruleID)
}

// GetRule retrieves a rule by ID
func (s *CategorizationRuleService) GetRule(ctx context.Context, ruleID string) (*models.CategorizationRule, error) {
	return s.ruleRepo.GetRuleByID(ctx, ruleID)
}

// GetRules retrieves rules with filtering
func (s *CategorizationRuleService) GetRules(ctx context.Context, filter *repository.CategorizationRuleFilter, limit, offset int) ([]*models.CategorizationRule, error) {
	return s.ruleRepo.GetRules(ctx, filter, limit, offset)
}

// ActivateRule activates a categorization rule
func (s *CategorizationRuleService) ActivateRule(ctx context.Context, ruleID string, userID string) error {
	rules := []string{ruleID}
	return s.ruleRepo.BulkUpdateRuleStatus(ctx, rules, true, userID)
}

// DeactivateRule deactivates a categorization rule
func (s *CategorizationRuleService) DeactivateRule(ctx context.Context, ruleID string, userID string) error {
	rules := []string{ruleID}
	return s.ruleRepo.BulkUpdateRuleStatus(ctx, rules, false, userID)
}

// BulkActivateRules activates multiple rules
func (s *CategorizationRuleService) BulkActivateRules(ctx context.Context, ruleIDs []string, userID string) error {
	return s.ruleRepo.BulkUpdateRuleStatus(ctx, ruleIDs, true, userID)
}

// BulkDeactivateRules deactivates multiple rules
func (s *CategorizationRuleService) BulkDeactivateRules(ctx context.Context, ruleIDs []string, userID string) error {
	return s.ruleRepo.BulkUpdateRuleStatus(ctx, ruleIDs, false, userID)
}

// BulkDeleteRules deletes multiple rules
func (s *CategorizationRuleService) BulkDeleteRules(ctx context.Context, ruleIDs []string) error {
	return s.ruleRepo.BulkDeleteRules(ctx, ruleIDs)
}

// ApplyRules applies categorization rules to a dispute and returns matching rules
func (s *CategorizationRuleService) ApplyRules(ctx context.Context, dispute *models.Dispute) ([]*RuleMatchResult, error) {
	// Get all active rules
	filter := &repository.CategorizationRuleFilter{
		IsActive: &[]bool{true}[0],
	}
	rules, err := s.ruleRepo.GetRules(ctx, filter, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	var matches []*RuleMatchResult
	content := strings.ToLower(dispute.Title + " " + dispute.Description)

	for _, rule := range rules {
		matchResult := s.evaluateRule(rule, content, dispute)
		if matchResult.IsMatch {
			matches = append(matches, matchResult)

			// Update rule usage statistics
			rule.UseCount++
			rule.LastUsedAt = &[]time.Time{time.Now()}[0]
			if matchResult.Confidence > 0.7 {
				rule.SuccessCount++
			}
			s.updateRulePerformanceScore(rule)

			// Update rule in background (don't block)
			go func(r *models.CategorizationRule) {
				ctx := context.Background()
				_ = s.ruleRepo.UpdateRule(ctx, r)
			}(rule)
		}
	}

	// Sort matches by confidence (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	return matches, nil
}

// GetTopPerformingRules retrieves the top performing rules for a category
func (s *CategorizationRuleService) GetTopPerformingRules(ctx context.Context, category models.DisputeCategory, limit int) ([]*models.CategorizationRule, error) {
	return s.ruleRepo.GetTopPerformingRules(ctx, category, limit)
}

// GetRuleUsageStats retrieves usage statistics for rules
func (s *CategorizationRuleService) GetRuleUsageStats(ctx context.Context, startDate, endDate time.Time) ([]*repository.RuleUsageStat, error) {
	return s.ruleRepo.GetRuleUsageStats(ctx, startDate, endDate)
}

// RuleMatchResult represents the result of applying a rule to a dispute
type RuleMatchResult struct {
	Rule       *models.CategorizationRule `json:"rule"`
	IsMatch    bool                       `json:"is_match"`
	Confidence float64                    `json:"confidence"`
	Reason     string                     `json:"reason"`
	MatchedOn  []string                   `json:"matched_on"`
}

// evaluateRule evaluates if a rule matches the dispute content
func (s *CategorizationRuleService) evaluateRule(rule *models.CategorizationRule, content string, dispute *models.Dispute) *RuleMatchResult {
	result := &RuleMatchResult{
		Rule:       rule,
		IsMatch:    false,
		Confidence: rule.BaseConfidence,
		MatchedOn:  make([]string, 0),
	}

	switch rule.Type {
	case models.CategorizationRuleTypeKeyword:
		result = s.evaluateKeywordRule(rule, content, result)
	case models.CategorizationRuleTypePattern:
		result = s.evaluatePatternRule(rule, content, result)
	case models.CategorizationRuleTypeSemantic:
		result = s.evaluateSemanticRule(rule, content, result)
	case models.CategorizationRuleTypeEntity:
		result = s.evaluateEntityRule(rule, content, result)
	case models.CategorizationRuleTypeComposite:
		result = s.evaluateCompositeRule(rule, content, dispute, result)
	}

	// Apply confidence bounds
	if result.Confidence < rule.MinConfidence {
		result.IsMatch = false
	} else if result.Confidence > rule.MaxConfidence {
		result.Confidence = rule.MaxConfidence
	}

	// Build reason string
	if result.IsMatch {
		result.Reason = fmt.Sprintf("Matched %s rule '%s' with %.1f%% confidence",
			rule.Type, rule.Name, result.Confidence*100)
	}

	return result
}

// evaluateKeywordRule evaluates keyword-based rules
func (s *CategorizationRuleService) evaluateKeywordRule(rule *models.CategorizationRule, content string, result *RuleMatchResult) *RuleMatchResult {
	matches := 0
	totalKeywords := len(rule.Keywords)

	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			matches++
			result.MatchedOn = append(result.MatchedOn, fmt.Sprintf("keyword: %s", keyword))
		}
	}

	if matches > 0 {
		result.IsMatch = true
		// Boost confidence based on keyword match ratio
		result.Confidence += (float64(matches) / float64(totalKeywords)) * 0.3
	}

	return result
}

// evaluatePatternRule evaluates regex pattern-based rules
func (s *CategorizationRuleService) evaluatePatternRule(rule *models.CategorizationRule, content string, result *RuleMatchResult) *RuleMatchResult {
	for _, patternStr := range rule.Patterns {
		pattern, err := regexp.Compile("(?i)" + patternStr)
		if err != nil {
			continue // Skip invalid patterns
		}

		if pattern.MatchString(content) {
			result.IsMatch = true
			result.Confidence += 0.4 // Pattern matches are highly confident
			result.MatchedOn = append(result.MatchedOn, fmt.Sprintf("pattern: %s", patternStr))
			break
		}
	}

	return result
}

// evaluateSemanticRule evaluates semantic-based rules
func (s *CategorizationRuleService) evaluateSemanticRule(rule *models.CategorizationRule, content string, result *RuleMatchResult) *RuleMatchResult {
	// This would integrate with semantic analysis - simplified for now
	for _, semanticKey := range rule.SemanticKeys {
		switch semanticKey {
		case "payment":
			if strings.Contains(content, "payment") || strings.Contains(content, "paid") || strings.Contains(content, "unpaid") {
				result.IsMatch = true
				result.Confidence += 0.3
				result.MatchedOn = append(result.MatchedOn, "semantic: payment")
			}
		case "contract":
			if strings.Contains(content, "contract") || strings.Contains(content, "agreement") || strings.Contains(content, "breach") {
				result.IsMatch = true
				result.Confidence += 0.3
				result.MatchedOn = append(result.MatchedOn, "semantic: contract")
			}
		case "milestone":
			if strings.Contains(content, "milestone") || strings.Contains(content, "deliverable") || strings.Contains(content, "deadline") {
				result.IsMatch = true
				result.Confidence += 0.3
				result.MatchedOn = append(result.MatchedOn, "semantic: milestone")
			}
		}
	}

	return result
}

// evaluateEntityRule evaluates entity-based rules
func (s *CategorizationRuleService) evaluateEntityRule(rule *models.CategorizationRule, content string, result *RuleMatchResult) *RuleMatchResult {
	for _, entity := range rule.Entities {
		switch entity {
		case "currency":
			if s.containsCurrency(content) {
				result.IsMatch = true
				result.Confidence += 0.4
				result.MatchedOn = append(result.MatchedOn, "entity: currency")
			}
		case "date":
			if s.containsDate(content) {
				result.IsMatch = true
				result.Confidence += 0.2
				result.MatchedOn = append(result.MatchedOn, "entity: date")
			}
		case "contract_ref":
			if s.containsContractReference(content) {
				result.IsMatch = true
				result.Confidence += 0.5
				result.MatchedOn = append(result.MatchedOn, "entity: contract_reference")
			}
		}
	}

	return result
}

// evaluateCompositeRule evaluates complex rules with conditions
func (s *CategorizationRuleService) evaluateCompositeRule(rule *models.CategorizationRule, content string, dispute *models.Dispute, result *RuleMatchResult) *RuleMatchResult {
	// Evaluate conditions from rule.Conditions map
	if rule.Conditions != nil {
		conditionsMet := 0
		totalConditions := len(rule.Conditions)

		for conditionKey, conditionValue := range rule.Conditions {
			if s.evaluateCondition(conditionKey, conditionValue, dispute, content) {
				conditionsMet++
				result.MatchedOn = append(result.MatchedOn, fmt.Sprintf("condition: %s", conditionKey))
			}
		}

		if conditionsMet == totalConditions {
			result.IsMatch = true
			result.Confidence += 0.5
		} else if conditionsMet > 0 {
			result.IsMatch = true
			result.Confidence += float64(conditionsMet) / float64(totalConditions) * 0.5
		}
	}

	return result
}

// evaluateCondition evaluates a single condition
func (s *CategorizationRuleService) evaluateCondition(conditionKey string, conditionValue interface{}, dispute *models.Dispute, content string) bool {
	switch conditionKey {
	case "min_amount":
		if dispute.DisputedAmount != nil {
			if minAmount, ok := conditionValue.(float64); ok {
				return *dispute.DisputedAmount >= minAmount
			}
		}
	case "contains_keyword":
		if keyword, ok := conditionValue.(string); ok {
			return strings.Contains(content, strings.ToLower(keyword))
		}
	case "title_contains":
		if keyword, ok := conditionValue.(string); ok {
			return strings.Contains(strings.ToLower(dispute.Title), keyword)
		}
	case "has_entity":
		if entity, ok := conditionValue.(string); ok {
			switch entity {
			case "currency":
				return s.containsCurrency(content)
			case "date":
				return s.containsDate(content)
			case "contract":
				return s.containsContractReference(content)
			}
		}
	}
	return false
}

// Helper methods for entity detection
func (s *CategorizationRuleService) containsCurrency(content string) bool {
	currencyPattern := regexp.MustCompile(`(?i)(\$|usd|eur|gbp|inr|e₹|₹|usdt|usdc)\s*\d{1,3}(?:,\d{3})*(?:\.\d{2})?`)
	return currencyPattern.MatchString(content)
}

func (s *CategorizationRuleService) containsDate(content string) bool {
	datePattern := regexp.MustCompile(`(?i)\b\d{1,2}[/-]\d{1,2}[/-]\d{2,4}|\b\d{2,4}[/-]\d{1,2}[/-]\d{1,2}|\b(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\s+\d{1,2}`)
	return datePattern.MatchString(content)
}

func (s *CategorizationRuleService) containsContractReference(content string) bool {
	contractPattern := regexp.MustCompile(`(?i)(contract|agreement|po|purchase order|order)\s+#?\w+-\d+`)
	return contractPattern.MatchString(content)
}

// validateRule validates a categorization rule
func (s *CategorizationRuleService) validateRule(rule *models.CategorizationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if len(rule.Name) < 3 || len(rule.Name) > 100 {
		return fmt.Errorf("rule name must be between 3 and 100 characters")
	}
	if rule.Description != "" && len(rule.Description) > 500 {
		return fmt.Errorf("rule description must be less than 500 characters")
	}
	if rule.BaseConfidence < 0 || rule.BaseConfidence > 1 {
		return fmt.Errorf("base confidence must be between 0 and 1")
	}
	if rule.Weight < 0 || rule.Weight > 1 {
		return fmt.Errorf("weight must be between 0 and 1")
	}
	if rule.MinConfidence < 0 || rule.MinConfidence > rule.MaxConfidence {
		return fmt.Errorf("min confidence must be between 0 and max confidence")
	}
	if rule.MaxConfidence < rule.MinConfidence || rule.MaxConfidence > 1 {
		return fmt.Errorf("max confidence must be between min confidence and 1")
	}

	// Validate based on rule type
	switch rule.Type {
	case models.CategorizationRuleTypeKeyword:
		if len(rule.Keywords) == 0 {
			return fmt.Errorf("keyword rules must have at least one keyword")
		}
	case models.CategorizationRuleTypePattern:
		if len(rule.Patterns) == 0 {
			return fmt.Errorf("pattern rules must have at least one pattern")
		}
		// Validate regex patterns
		for _, pattern := range rule.Patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
			}
		}
	case models.CategorizationRuleTypeEntity:
		if len(rule.Entities) == 0 {
			return fmt.Errorf("entity rules must have at least one entity")
		}
		validEntities := []string{"currency", "date", "contract_ref", "milestone_ref"}
		for _, entity := range rule.Entities {
			found := false
			for _, validEntity := range validEntities {
				if entity == validEntity {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid entity type '%s'", entity)
			}
		}
	}

	return nil
}

// updateRulePerformanceScore calculates and updates the performance score for a rule
func (s *CategorizationRuleService) updateRulePerformanceScore(rule *models.CategorizationRule) {
	if rule.UseCount == 0 {
		rule.PerformanceScore = 0.0
		return
	}

	// Calculate success rate
	successRate := float64(rule.SuccessCount) / float64(rule.UseCount)

	// Factor in recency (more recent usage has higher weight)
	recencyBonus := 0.0
	if rule.LastUsedAt != nil {
		hoursSinceLastUse := time.Since(*rule.LastUsedAt).Hours()
		if hoursSinceLastUse < 24 {
			recencyBonus = 0.1
		} else if hoursSinceLastUse < 168 { // 7 days
			recencyBonus = 0.05
		}
	}

	rule.PerformanceScore = successRate + recencyBonus

	// Cap at 1.0
	if rule.PerformanceScore > 1.0 {
		rule.PerformanceScore = 1.0
	}
}
