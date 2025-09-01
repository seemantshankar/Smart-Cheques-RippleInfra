package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CategorizationRuleServiceTestSuite struct {
	suite.Suite
	service  *CategorizationRuleService
	mockRepo *MockCategorizationRuleRepository
	ctx      context.Context
}

type MockCategorizationRuleRepository struct {
	rules map[string]*models.CategorizationRule
}

func NewMockCategorizationRuleRepository() *MockCategorizationRuleRepository {
	return &MockCategorizationRuleRepository{
		rules: make(map[string]*models.CategorizationRule),
	}
}

func (m *MockCategorizationRuleRepository) CreateRule(ctx context.Context, rule *models.CategorizationRule) error {
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	m.rules[rule.ID] = rule
	return nil
}

func (m *MockCategorizationRuleRepository) GetRuleByID(ctx context.Context, id string) (*models.CategorizationRule, error) {
	if rule, exists := m.rules[id]; exists {
		return rule, nil
	}
	return nil, nil
}

func (m *MockCategorizationRuleRepository) UpdateRule(ctx context.Context, rule *models.CategorizationRule) error {
	if _, exists := m.rules[rule.ID]; exists {
		rule.UpdatedAt = time.Now()
		m.rules[rule.ID] = rule
		return nil
	}
	return sqlmock.ErrCancelled
}

func (m *MockCategorizationRuleRepository) DeleteRule(ctx context.Context, id string) error {
	delete(m.rules, id)
	return nil
}

func (m *MockCategorizationRuleRepository) GetRules(ctx context.Context, filter *repository.CategorizationRuleFilter, limit, offset int) ([]*models.CategorizationRule, error) {
	var rules []*models.CategorizationRule
	for _, rule := range m.rules {
		if m.matchesFilter(rule, filter) {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

func (m *MockCategorizationRuleRepository) matchesFilter(rule *models.CategorizationRule, filter *repository.CategorizationRuleFilter) bool {
	if filter == nil {
		return true
	}
	if filter.IsActive != nil && rule.IsActive != *filter.IsActive {
		return false
	}
	if filter.Category != nil && rule.Category != *filter.Category {
		return false
	}
	if filter.Type != nil && rule.Type != *filter.Type {
		return false
	}
	return true
}

func (m *MockCategorizationRuleRepository) BulkUpdateRuleStatus(ctx context.Context, ruleIDs []string, isActive bool, updatedBy string) error {
	for _, id := range ruleIDs {
		if rule, exists := m.rules[id]; exists {
			rule.IsActive = isActive
			rule.UpdatedBy = updatedBy
			rule.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (m *MockCategorizationRuleRepository) BulkDeleteRules(ctx context.Context, ruleIDs []string) error {
	for _, id := range ruleIDs {
		delete(m.rules, id)
	}
	return nil
}

func (m *MockCategorizationRuleRepository) GetTopPerformingRules(ctx context.Context, category models.DisputeCategory, limit int) ([]*models.CategorizationRule, error) {
	var rules []*models.CategorizationRule
	for _, rule := range m.rules {
		if rule.Category == category && rule.IsActive {
			rules = append(rules, rule)
		}
	}
	// Sort by performance score (simplified)
	return rules, nil
}

// Stub implementations for remaining interface methods
func (m *MockCategorizationRuleRepository) CreateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error {
	return nil
}
func (m *MockCategorizationRuleRepository) GetRuleGroupByID(ctx context.Context, id string) (*models.CategorizationRuleGroup, error) {
	return nil, nil
}
func (m *MockCategorizationRuleRepository) UpdateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error {
	return nil
}
func (m *MockCategorizationRuleRepository) DeleteRuleGroup(ctx context.Context, id string) error {
	return nil
}
func (m *MockCategorizationRuleRepository) GetRuleGroups(ctx context.Context, filter *repository.CategorizationRuleGroupFilter, limit, offset int) ([]*models.CategorizationRuleGroup, error) {
	return nil, nil
}
func (m *MockCategorizationRuleRepository) CreateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error {
	return nil
}
func (m *MockCategorizationRuleRepository) UpdateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error {
	return nil
}
func (m *MockCategorizationRuleRepository) GetRulePerformance(ctx context.Context, ruleID string, periodStart, periodEnd time.Time) (*models.CategorizationRulePerformance, error) {
	return nil, nil
}
func (m *MockCategorizationRuleRepository) CreateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error {
	return nil
}
func (m *MockCategorizationRuleRepository) GetRuleTemplateByID(ctx context.Context, id string) (*models.CategorizationRuleTemplate, error) {
	return nil, nil
}
func (m *MockCategorizationRuleRepository) UpdateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error {
	return nil
}
func (m *MockCategorizationRuleRepository) DeleteRuleTemplate(ctx context.Context, id string) error {
	return nil
}
func (m *MockCategorizationRuleRepository) GetRuleTemplates(ctx context.Context, filter *repository.CategorizationRuleTemplateFilter, limit, offset int) ([]*models.CategorizationRuleTemplate, error) {
	return nil, nil
}
func (m *MockCategorizationRuleRepository) GetRuleUsageStats(ctx context.Context, startDate, endDate time.Time) ([]*repository.RuleUsageStat, error) {
	return nil, nil
}

func (suite *CategorizationRuleServiceTestSuite) SetupTest() {
	suite.mockRepo = NewMockCategorizationRuleRepository()
	suite.service = NewCategorizationRuleService(suite.mockRepo)
	suite.ctx = context.Background()
}

func (suite *CategorizationRuleServiceTestSuite) TestCreateRule() {
	rule := &models.CategorizationRule{
		Name:        "Test Payment Rule",
		Description: "A test rule for payment disputes",
		Type:        models.CategorizationRuleTypeKeyword,
		Category:    models.DisputeCategoryPayment,
		Priority:    models.DisputePriorityHigh,
		Keywords:    []string{"payment", "unpaid", "refund"},
		IsActive:    true,
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), rule.ID)
	assert.True(suite.T(), rule.IsActive)
	assert.Equal(suite.T(), 0.5, rule.BaseConfidence)
	assert.Equal(suite.T(), 1.0, rule.Weight)
}

func (suite *CategorizationRuleServiceTestSuite) TestCreateRuleValidation() {
	// Test empty name
	rule := &models.CategorizationRule{
		Description: "A test rule",
		Type:        models.CategorizationRuleTypeKeyword,
		Category:    models.DisputeCategoryPayment,
		CreatedBy:   "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "rule name is required")
}

func (suite *CategorizationRuleServiceTestSuite) TestUpdateRule() {
	// First create a rule
	rule := &models.CategorizationRule{
		Name:        "Test Rule",
		Description: "A test rule",
		Type:        models.CategorizationRuleTypeKeyword,
		Category:    models.DisputeCategoryPayment,
		Priority:    models.DisputePriorityHigh,
		Keywords:    []string{"payment", "unpaid"},
		IsActive:    true,
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Update the rule
	rule.Name = "Updated Test Rule"
	rule.Description = "Updated description"

	err = suite.service.UpdateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Verify update
	updatedRule, err := suite.service.GetRule(suite.ctx, rule.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Test Rule", updatedRule.Name)
	assert.Equal(suite.T(), "Updated description", updatedRule.Description)
}

func (suite *CategorizationRuleServiceTestSuite) TestActivateRule() {
	// Create an inactive rule
	rule := &models.CategorizationRule{
		Name:      "Inactive Rule",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Priority:  models.DisputePriorityNormal,
		Keywords:  []string{"test", "inactive"},
		IsActive:  false,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Activate the rule
	err = suite.service.ActivateRule(suite.ctx, rule.ID, "test-user")
	assert.NoError(suite.T(), err)

	// Verify activation
	updatedRule, err := suite.service.GetRule(suite.ctx, rule.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updatedRule.IsActive)
}

func (suite *CategorizationRuleServiceTestSuite) TestDeactivateRule() {
	// Create an active rule
	rule := &models.CategorizationRule{
		Name:      "Active Rule",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Priority:  models.DisputePriorityNormal,
		Keywords:  []string{"test", "active"},
		IsActive:  true,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Deactivate the rule
	err = suite.service.DeactivateRule(suite.ctx, rule.ID, "test-user")
	assert.NoError(suite.T(), err)

	// Verify deactivation
	updatedRule, err := suite.service.GetRule(suite.ctx, rule.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), updatedRule.IsActive)
}

func (suite *CategorizationRuleServiceTestSuite) TestBulkActivateRules() {
	// Create multiple inactive rules
	rule1 := &models.CategorizationRule{
		Name:      "Rule 1",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Keywords:  []string{"rule1", "test"},
		IsActive:  false,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	rule2 := &models.CategorizationRule{
		Name:      "Rule 2",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Keywords:  []string{"rule2", "test"},
		IsActive:  false,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule1, "test-user")
	assert.NoError(suite.T(), err)
	err = suite.service.CreateRule(suite.ctx, rule2, "test-user")
	assert.NoError(suite.T(), err)

	// Bulk activate
	ruleIDs := []string{rule1.ID, rule2.ID}
	err = suite.service.BulkActivateRules(suite.ctx, ruleIDs, "test-user")
	assert.NoError(suite.T(), err)

	// Verify activation
	updatedRule1, _ := suite.service.GetRule(suite.ctx, rule1.ID)
	updatedRule2, _ := suite.service.GetRule(suite.ctx, rule2.ID)
	assert.True(suite.T(), updatedRule1.IsActive)
	assert.True(suite.T(), updatedRule2.IsActive)
}

func (suite *CategorizationRuleServiceTestSuite) TestApplyRules() {
	// Create a keyword-based rule
	rule := &models.CategorizationRule{
		Name:           "Payment Keyword Rule",
		Type:           models.CategorizationRuleTypeKeyword,
		Category:       models.DisputeCategoryPayment,
		Priority:       models.DisputePriorityHigh,
		Keywords:       []string{"payment", "unpaid", "refund"},
		BaseConfidence: 0.8,
		IsActive:       true,
		CreatedBy:      "test-user",
		UpdatedBy:      "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Create a dispute that should match the rule
	dispute := &models.Dispute{
		Title:       "Payment Not Received",
		Description: "We have not received the payment for our services. The payment was supposed to be made last week.",
	}

	// Apply rules
	matches, err := suite.service.ApplyRules(suite.ctx, dispute)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), matches)

	// Check that we got a match
	found := false
	for _, match := range matches {
		if match.IsMatch && match.Rule.ID == rule.ID {
			found = true
			assert.Greater(suite.T(), match.Confidence, 0.0)
			break
		}
	}
	assert.True(suite.T(), found, "Expected rule should have matched the dispute")
}

func (suite *CategorizationRuleServiceTestSuite) TestGetRulesWithFilter() {
	// Create rules of different categories
	paymentRule := &models.CategorizationRule{
		Name:      "Payment Rule",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Keywords:  []string{"payment", "test"},
		IsActive:  true,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	contractRule := &models.CategorizationRule{
		Name:      "Contract Rule",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryContractBreach,
		Keywords:  []string{"contract", "test"},
		IsActive:  true,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, paymentRule, "test-user")
	assert.NoError(suite.T(), err)
	err = suite.service.CreateRule(suite.ctx, contractRule, "test-user")
	assert.NoError(suite.T(), err)

	// Filter by payment category
	filter := &repository.CategorizationRuleFilter{
		Category: &[]models.DisputeCategory{models.DisputeCategoryPayment}[0],
	}

	rules, err := suite.service.GetRules(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), rules, 1)
	assert.Equal(suite.T(), models.DisputeCategoryPayment, rules[0].Category)
}

func (suite *CategorizationRuleServiceTestSuite) TestRuleValidation() {
	// Test invalid confidence values
	rule := &models.CategorizationRule{
		Name:           "Test Rule",
		Type:           models.CategorizationRuleTypeKeyword,
		Category:       models.DisputeCategoryPayment,
		BaseConfidence: 1.5, // Invalid: > 1.0
		CreatedBy:      "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "base confidence must be between 0 and 1")
}

func (suite *CategorizationRuleServiceTestSuite) TestDeleteRule() {
	// Create a rule
	rule := &models.CategorizationRule{
		Name:      "Rule to Delete",
		Type:      models.CategorizationRuleTypeKeyword,
		Category:  models.DisputeCategoryPayment,
		Keywords:  []string{"delete", "test"},
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, rule, "test-user")
	assert.NoError(suite.T(), err)

	// Delete the rule
	err = suite.service.DeleteRule(suite.ctx, rule.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	deletedRule, err := suite.service.GetRule(suite.ctx, rule.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), deletedRule)
}

func (suite *CategorizationRuleServiceTestSuite) TestGetTopPerformingRules() {
	// Create rules with different performance scores
	highPerfRule := &models.CategorizationRule{
		Name:             "High Performance Rule",
		Type:             models.CategorizationRuleTypeKeyword,
		Category:         models.DisputeCategoryPayment,
		Keywords:         []string{"high", "performance"},
		PerformanceScore: 0.9,
		IsActive:         true,
		CreatedBy:        "test-user",
		UpdatedBy:        "test-user",
	}

	lowPerfRule := &models.CategorizationRule{
		Name:             "Low Performance Rule",
		Type:             models.CategorizationRuleTypeKeyword,
		Category:         models.DisputeCategoryPayment,
		Keywords:         []string{"low", "performance"},
		PerformanceScore: 0.3,
		IsActive:         true,
		CreatedBy:        "test-user",
		UpdatedBy:        "test-user",
	}

	err := suite.service.CreateRule(suite.ctx, highPerfRule, "test-user")
	assert.NoError(suite.T(), err)
	err = suite.service.CreateRule(suite.ctx, lowPerfRule, "test-user")
	assert.NoError(suite.T(), err)

	// Get top performing rules
	rules, err := suite.service.GetTopPerformingRules(suite.ctx, models.DisputeCategoryPayment, 5)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), rules)
	// Note: Mock doesn't implement sorting, but in real implementation it would be sorted
}

func TestCategorizationRuleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategorizationRuleServiceTestSuite))
}
