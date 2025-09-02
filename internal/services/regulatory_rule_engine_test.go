package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
)

func TestRegulatoryRuleEngine_EvaluateCompliance(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		request        *models.ComplianceValidationRequest
		mockRules      []*models.RegulatoryRule
		expectedStatus string
		expectedScore  float64
		expectedPassed int
		expectedFailed int
	}{
		{
			name: "Compliant transaction with no rules",
			request: &models.ComplianceValidationRequest{
				TransactionID:   "tx-123",
				EnterpriseID:    "ent-123",
				Jurisdiction:    "US",
				TransactionType: "payment",
				Amount:          "1000.00",
				Currency:        "USDT",
				FromAddress:     "addr1",
				ToAddress:       "addr2",
				ValidationLevel: "basic",
				Metadata:        map[string]interface{}{},
			},
			mockRules:      []*models.RegulatoryRule{},
			expectedStatus: "approved",
			expectedScore:  1.0,
			expectedPassed: 0,
			expectedFailed: 0,
		},
		{
			name: "Transaction with amount limit violation",
			request: &models.ComplianceValidationRequest{
				TransactionID:   "tx-124",
				EnterpriseID:    "ent-123",
				Jurisdiction:    "US",
				TransactionType: "payment",
				Amount:          "15000.00",
				Currency:        "USDT",
				FromAddress:     "addr1",
				ToAddress:       "addr2",
				ValidationLevel: "basic",
				Metadata:        map[string]interface{}{},
			},
			mockRules: []*models.RegulatoryRule{
				{
					ID:           "rule-1",
					Name:         "High Amount Limit",
					Category:     "aml",
					Jurisdiction: "US",
					Priority:     3,
					RuleType:     "amount_limit",
					Conditions:   map[string]interface{}{},
					Thresholds:   map[string]interface{}{"max_amount": 10000.0},
					Actions:      []string{"flag"},
					Status:       "active",
					EffectiveAt:  time.Now().Add(-24 * time.Hour),
				},
			},
			expectedStatus: "flagged",
			expectedScore:  0.7, // 1.0 - (3 * 0.1)
			expectedPassed: 0,
			expectedFailed: 1,
		},
		{
			name: "Transaction with blacklist violation",
			request: &models.ComplianceValidationRequest{
				TransactionID:   "tx-125",
				EnterpriseID:    "ent-123",
				Jurisdiction:    "US",
				TransactionType: "payment",
				Amount:          "1000.00",
				Currency:        "USDT",
				FromAddress:     "blacklisted-addr",
				ToAddress:       "addr2",
				ValidationLevel: "strict",
				Metadata:        map[string]interface{}{},
			},
			mockRules: []*models.RegulatoryRule{
				{
					ID:           "rule-2",
					Name:         "Blacklist Check",
					Category:     "sanctions",
					Jurisdiction: "US",
					Priority:     5,
					RuleType:     "blacklist_check",
					Conditions:   map[string]interface{}{"blacklisted_addresses": []interface{}{"blacklisted-addr"}},
					Thresholds:   map[string]interface{}{},
					Actions:      []string{"reject"},
					Status:       "active",
					EffectiveAt:  time.Now().Add(-24 * time.Hour),
				},
			},
			expectedStatus: "rejected",
			expectedScore:  0.5, // 1.0 - (5 * 0.1)
			expectedPassed: 0,
			expectedFailed: 1,
		},
		{
			name: "Transaction with pattern detection",
			request: &models.ComplianceValidationRequest{
				TransactionID:   "tx-126",
				EnterpriseID:    "ent-123",
				Jurisdiction:    "US",
				TransactionType: "payment",
				Amount:          "5000.00",
				Currency:        "USDT",
				FromAddress:     "addr1",
				ToAddress:       "addr2",
				ValidationLevel: "enhanced",
				Metadata:        map[string]interface{}{"transaction_count_24h": 15.0},
			},
			mockRules: []*models.RegulatoryRule{
				{
					ID:           "rule-3",
					Name:         "High Frequency Detection",
					Category:     "aml",
					Jurisdiction: "US",
					Priority:     4,
					RuleType:     "frequency_limit",
					Conditions:   map[string]interface{}{},
					Thresholds:   map[string]interface{}{"max_transactions_24h": 10.0},
					Actions:      []string{"flag"},
					Status:       "active",
					EffectiveAt:  time.Now().Add(-24 * time.Hour),
				},
			},
			expectedStatus: "flagged",
			expectedScore:  0.6, // 1.0 - (4 * 0.1)
			expectedPassed: 0,
			expectedFailed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh mock for each test case
			mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
			engine := NewRegulatoryRuleEngine(mockRuleRepo)

			// Setup mock expectations
			mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, tt.request.Jurisdiction).Return(tt.mockRules, nil)

			// Execute
			result, err := engine.EvaluateCompliance(context.Background(), tt.request)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedScore, result.ComplianceScore)
			assert.Equal(t, tt.expectedPassed, result.RulesPassed)
			assert.Equal(t, tt.expectedFailed, result.RulesFailed)
			assert.Equal(t, tt.request.TransactionID, result.TransactionID)
			assert.Equal(t, tt.request.ValidationLevel, result.ValidationLevel)

			// Verify mock expectations
			mockRuleRepo.AssertExpectations(t)
		})
	}
}

func TestRegulatoryRuleEngine_RuleAppliesToTransaction(t *testing.T) {
	engine := &RegulatoryRuleEngine{}

	tests := []struct {
		name     string
		rule     *models.RegulatoryRule
		request  *models.ComplianceValidationRequest
		expected bool
	}{
		{
			name: "Rule applies to all transactions",
			rule: &models.RegulatoryRule{
				Conditions: map[string]interface{}{},
			},
			request: &models.ComplianceValidationRequest{
				TransactionType: "payment",
				Currency:        "USDT",
			},
			expected: true,
		},
		{
			name: "Rule applies to specific transaction type",
			rule: &models.RegulatoryRule{
				Conditions: map[string]interface{}{
					"transaction_types": []interface{}{"payment", "escrow_create"},
				},
			},
			request: &models.ComplianceValidationRequest{
				TransactionType: "payment",
				Currency:        "USDT",
			},
			expected: true,
		},
		{
			name: "Rule does not apply to transaction type",
			rule: &models.RegulatoryRule{
				Conditions: map[string]interface{}{
					"transaction_types": []interface{}{"escrow_create"},
				},
			},
			request: &models.ComplianceValidationRequest{
				TransactionType: "payment",
				Currency:        "USDT",
			},
			expected: false,
		},
		{
			name: "Rule applies to specific currency",
			rule: &models.RegulatoryRule{
				Conditions: map[string]interface{}{
					"currencies": []interface{}{"USDT", "USDC"},
				},
			},
			request: &models.ComplianceValidationRequest{
				TransactionType: "payment",
				Currency:        "USDT",
			},
			expected: true,
		},
		{
			name: "Rule does not apply to currency",
			rule: &models.RegulatoryRule{
				Conditions: map[string]interface{}{
					"currencies": []interface{}{"USDC"},
				},
			},
			request: &models.ComplianceValidationRequest{
				TransactionType: "payment",
				Currency:        "USDT",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.ruleAppliesToTransaction(tt.rule, tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegulatoryRuleEngine_GetSeverityFromPriority(t *testing.T) {
	engine := &RegulatoryRuleEngine{}

	tests := []struct {
		priority int
		expected string
	}{
		{1, "low"},
		{2, "low"},
		{3, "medium"},
		{4, "high"},
		{5, "critical"},
		{6, "medium"}, // default case
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Priority_%d", tt.priority), func(t *testing.T) {
			result := engine.getSeverityFromPriority(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegulatoryRuleEngine_DetermineComplianceStatus(t *testing.T) {
	engine := &RegulatoryRuleEngine{}

	tests := []struct {
		name           string
		result         *models.ComplianceValidationResult
		expectedStatus string
	}{
		{
			name: "Approved with high score and no violations",
			result: &models.ComplianceValidationResult{
				ComplianceScore: 0.9,
				Violations:      []models.ComplianceViolation{},
			},
			expectedStatus: "approved",
		},
		{
			name: "Flagged with medium score",
			result: &models.ComplianceValidationResult{
				ComplianceScore: 0.7,
				Violations:      []models.ComplianceViolation{},
			},
			expectedStatus: "flagged",
		},
		{
			name: "Rejected with low score",
			result: &models.ComplianceValidationResult{
				ComplianceScore: 0.4,
				Violations:      []models.ComplianceViolation{},
			},
			expectedStatus: "rejected",
		},
		{
			name: "Rejected with critical violation",
			result: &models.ComplianceValidationResult{
				ComplianceScore: 0.9,
				Violations: []models.ComplianceViolation{
					{Severity: "critical"},
				},
			},
			expectedStatus: "rejected",
		},
		{
			name: "Flagged with violations",
			result: &models.ComplianceValidationResult{
				ComplianceScore: 0.9,
				Violations: []models.ComplianceViolation{
					{Severity: "medium"},
				},
			},
			expectedStatus: "flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.determineComplianceStatus(tt.result)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}

func TestRegulatoryRuleEngine_MatchesPattern(t *testing.T) {
	engine := &RegulatoryRuleEngine{}

	tests := []struct {
		name     string
		pattern  string
		request  *models.ComplianceValidationRequest
		expected bool
	}{
		{
			name:    "Round amount pattern - matches",
			pattern: "round_amount",
			request: &models.ComplianceValidationRequest{
				Amount: "5000.00",
			},
			expected: true,
		},
		{
			name:    "Round amount pattern - no match",
			pattern: "round_amount",
			request: &models.ComplianceValidationRequest{
				Amount: "5000.50",
			},
			expected: false,
		},
		{
			name:    "High frequency pattern - matches",
			pattern: "high_frequency",
			request: &models.ComplianceValidationRequest{
				Metadata: map[string]interface{}{"transaction_count_24h": 15.0},
			},
			expected: true,
		},
		{
			name:    "High frequency pattern - no match",
			pattern: "high_frequency",
			request: &models.ComplianceValidationRequest{
				Metadata: map[string]interface{}{"transaction_count_24h": 5.0},
			},
			expected: false,
		},
		{
			name:     "Unknown pattern",
			pattern:  "unknown_pattern",
			request:  &models.ComplianceValidationRequest{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.matchesPattern(tt.request, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
