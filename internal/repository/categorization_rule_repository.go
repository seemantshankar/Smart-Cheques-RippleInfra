package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/smart-payment-infrastructure/internal/models"
)

// CategorizationRuleFilter represents filter criteria for rule queries
type CategorizationRuleFilter struct {
	Category      *models.DisputeCategory        `json:"category,omitempty"`
	Type          *models.CategorizationRuleType `json:"type,omitempty"`
	IsActive      *bool                          `json:"is_active,omitempty"`
	CreatedBy     *string                        `json:"created_by,omitempty"`
	NameContains  *string                        `json:"name_contains,omitempty"`
	PriorityOrder *int                           `json:"priority_order,omitempty"`
}

// CategorizationRuleGroupFilter represents filter criteria for rule group queries
type CategorizationRuleGroupFilter struct {
	Category     *models.DisputeCategory `json:"category,omitempty"`
	IsActive     *bool                   `json:"is_active,omitempty"`
	CreatedBy    *string                 `json:"created_by,omitempty"`
	NameContains *string                 `json:"name_contains,omitempty"`
}

// CategorizationRuleTemplateFilter represents filter criteria for rule template queries
type CategorizationRuleTemplateFilter struct {
	Category     *models.DisputeCategory        `json:"category,omitempty"`
	Type         *models.CategorizationRuleType `json:"type,omitempty"`
	IsPublic     *bool                          `json:"is_public,omitempty"`
	CreatedBy    *string                        `json:"created_by,omitempty"`
	NameContains *string                        `json:"name_contains,omitempty"`
}

// RuleUsageStat represents usage statistics for a rule
type RuleUsageStat struct {
	RuleID            string     `json:"rule_id"`
	RuleName          string     `json:"rule_name"`
	UseCount          int64      `json:"use_count"`
	SuccessCount      int64      `json:"success_count"`
	SuccessRate       float64    `json:"success_rate"`
	AverageConfidence float64    `json:"average_confidence"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
}

// CategorizationRuleRepository implements CategorizationRuleRepositoryInterface
type CategorizationRuleRepository struct {
	db *sql.DB
}

// NewCategorizationRuleRepository creates a new categorization rule repository
func NewCategorizationRuleRepository(db *sql.DB) *CategorizationRuleRepository {
	return &CategorizationRuleRepository{db: db}
}

// CreateRule creates a new categorization rule
func (r *CategorizationRuleRepository) CreateRule(ctx context.Context, rule *models.CategorizationRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	query := `
		INSERT INTO categorization_rules (
			id, name, description, type, category, priority,
			keywords, patterns, entities, semantic_keys, conditions,
			base_confidence, weight, min_confidence, max_confidence,
			is_active, priority_order, created_by, updated_by, created_at, updated_at,
			use_count, success_count, performance_score
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24)`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Category, rule.Priority,
		pq.Array(rule.Keywords), pq.Array(rule.Patterns), pq.Array(rule.Entities),
		pq.Array(rule.SemanticKeys), rule.Conditions,
		rule.BaseConfidence, rule.Weight, rule.MinConfidence, rule.MaxConfidence,
		rule.IsActive, rule.PriorityOrder, rule.CreatedBy, rule.UpdatedBy,
		rule.CreatedAt, rule.UpdatedAt, rule.UseCount, rule.SuccessCount, rule.PerformanceScore)

	return err
}

// GetRuleByID retrieves a categorization rule by ID
func (r *CategorizationRuleRepository) GetRuleByID(ctx context.Context, id string) (*models.CategorizationRule, error) {
	query := `
		SELECT id, name, description, type, category, priority,
			   keywords, patterns, entities, semantic_keys, conditions,
			   base_confidence, weight, min_confidence, max_confidence,
			   is_active, priority_order, created_by, updated_by, created_at, updated_at,
			   use_count, success_count, last_used_at, performance_score
		FROM categorization_rules WHERE id = $1`

	rule := &models.CategorizationRule{}
	var keywords, patterns, entities, semanticKeys pq.StringArray
	var conditions map[string]interface{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Category, &rule.Priority,
		&keywords, &patterns, &entities, &semanticKeys, &conditions,
		&rule.BaseConfidence, &rule.Weight, &rule.MinConfidence, &rule.MaxConfidence,
		&rule.IsActive, &rule.PriorityOrder, &rule.CreatedBy, &rule.UpdatedBy,
		&rule.CreatedAt, &rule.UpdatedAt, &rule.UseCount, &rule.SuccessCount,
		&rule.LastUsedAt, &rule.PerformanceScore)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	rule.Keywords = []string(keywords)
	rule.Patterns = []string(patterns)
	rule.Entities = []string(entities)
	rule.SemanticKeys = []string(semanticKeys)
	rule.Conditions = conditions

	return rule, nil
}

// UpdateRule updates an existing categorization rule
func (r *CategorizationRuleRepository) UpdateRule(ctx context.Context, rule *models.CategorizationRule) error {
	rule.UpdatedAt = time.Now()

	query := `
		UPDATE categorization_rules SET
			name = $2, description = $3, type = $4, category = $5, priority = $6,
			keywords = $7, patterns = $8, entities = $9, semantic_keys = $10, conditions = $11,
			base_confidence = $12, weight = $13, min_confidence = $14, max_confidence = $15,
			is_active = $16, priority_order = $17, updated_by = $18, updated_at = $19,
			use_count = $20, success_count = $21, last_used_at = $22, performance_score = $23
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Type, rule.Category, rule.Priority,
		pq.Array(rule.Keywords), pq.Array(rule.Patterns), pq.Array(rule.Entities),
		pq.Array(rule.SemanticKeys), rule.Conditions,
		rule.BaseConfidence, rule.Weight, rule.MinConfidence, rule.MaxConfidence,
		rule.IsActive, rule.PriorityOrder, rule.UpdatedBy, rule.UpdatedAt,
		rule.UseCount, rule.SuccessCount, rule.LastUsedAt, rule.PerformanceScore)

	return err
}

// DeleteRule deletes a categorization rule by ID
func (r *CategorizationRuleRepository) DeleteRule(ctx context.Context, id string) error {
	query := `DELETE FROM categorization_rules WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetRules retrieves categorization rules with filtering
func (r *CategorizationRuleRepository) GetRules(ctx context.Context, filter *CategorizationRuleFilter, limit, offset int) ([]*models.CategorizationRule, error) {
	query := `
		SELECT id, name, description, type, category, priority,
			   keywords, patterns, entities, semantic_keys, conditions,
			   base_confidence, weight, min_confidence, max_confidence,
			   is_active, priority_order, created_by, updated_by, created_at, updated_at,
			   use_count, success_count, last_used_at, performance_score
		FROM categorization_rules`

	var conditions []string
	var args []interface{}
	argCount := 0

	if filter != nil {
		if filter.Category != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("category = $%d", argCount))
			args = append(args, *filter.Category)
		}
		if filter.Type != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
			args = append(args, *filter.Type)
		}
		if filter.IsActive != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("is_active = $%d", argCount))
			args = append(args, *filter.IsActive)
		}
		if filter.CreatedBy != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("created_by = $%d", argCount))
			args = append(args, *filter.CreatedBy)
		}
		if filter.NameContains != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argCount))
			args = append(args, "%"+*filter.NameContains+"%")
		}
		if filter.PriorityOrder != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("priority_order = $%d", argCount))
			args = append(args, *filter.PriorityOrder)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY priority_order ASC, created_at DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.CategorizationRule
	for rows.Next() {
		rule := &models.CategorizationRule{}
		var keywords, patterns, entities, semanticKeys pq.StringArray
		var conditions map[string]interface{}

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Category, &rule.Priority,
			&keywords, &patterns, &entities, &semanticKeys, &conditions,
			&rule.BaseConfidence, &rule.Weight, &rule.MinConfidence, &rule.MaxConfidence,
			&rule.IsActive, &rule.PriorityOrder, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt, &rule.UseCount, &rule.SuccessCount,
			&rule.LastUsedAt, &rule.PerformanceScore)

		if err != nil {
			return nil, err
		}

		rule.Keywords = []string(keywords)
		rule.Patterns = []string(patterns)
		rule.Entities = []string(entities)
		rule.SemanticKeys = []string(semanticKeys)
		rule.Conditions = conditions

		rules = append(rules, rule)
	}

	return rules, nil
}

// BulkUpdateRuleStatus updates the status of multiple rules
func (r *CategorizationRuleRepository) BulkUpdateRuleStatus(ctx context.Context, ruleIDs []string, isActive bool, updatedBy string) error {
	query := `UPDATE categorization_rules SET is_active = $1, updated_by = $2, updated_at = $3 WHERE id = ANY($4)`

	_, err := r.db.ExecContext(ctx, query, isActive, updatedBy, time.Now(), pq.Array(ruleIDs))
	return err
}

// BulkDeleteRules deletes multiple rules
func (r *CategorizationRuleRepository) BulkDeleteRules(ctx context.Context, ruleIDs []string) error {
	query := `DELETE FROM categorization_rules WHERE id = ANY($1)`
	_, err := r.db.ExecContext(ctx, query, pq.Array(ruleIDs))
	return err
}

// GetTopPerformingRules retrieves the top performing rules for a category
func (r *CategorizationRuleRepository) GetTopPerformingRules(ctx context.Context, category models.DisputeCategory, limit int) ([]*models.CategorizationRule, error) {
	query := `
		SELECT id, name, description, type, category, priority,
			   keywords, patterns, entities, semantic_keys, conditions,
			   base_confidence, weight, min_confidence, max_confidence,
			   is_active, priority_order, created_by, updated_by, created_at, updated_at,
			   use_count, success_count, last_used_at, performance_score
		FROM categorization_rules
		WHERE category = $1 AND is_active = true
		ORDER BY performance_score DESC, success_count DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, category, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.CategorizationRule
	for rows.Next() {
		rule := &models.CategorizationRule{}
		var keywords, patterns, entities, semanticKeys pq.StringArray
		var conditions map[string]interface{}

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Category, &rule.Priority,
			&keywords, &patterns, &entities, &semanticKeys, &conditions,
			&rule.BaseConfidence, &rule.Weight, &rule.MinConfidence, &rule.MaxConfidence,
			&rule.IsActive, &rule.PriorityOrder, &rule.CreatedBy, &rule.UpdatedBy,
			&rule.CreatedAt, &rule.UpdatedAt, &rule.UseCount, &rule.SuccessCount,
			&rule.LastUsedAt, &rule.PerformanceScore)

		if err != nil {
			return nil, err
		}

		rule.Keywords = []string(keywords)
		rule.Patterns = []string(patterns)
		rule.Entities = []string(entities)
		rule.SemanticKeys = []string(semanticKeys)
		rule.Conditions = conditions

		rules = append(rules, rule)
	}

	return rules, nil
}

// Placeholder implementations for remaining methods (to be implemented as needed)
func (r *CategorizationRuleRepository) CreateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRuleGroupByID(ctx context.Context, id string) (*models.CategorizationRuleGroup, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) UpdateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) DeleteRuleGroup(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRuleGroups(ctx context.Context, filter *CategorizationRuleGroupFilter, limit, offset int) ([]*models.CategorizationRuleGroup, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) CreateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) UpdateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRulePerformance(ctx context.Context, ruleID string, periodStart, periodEnd time.Time) (*models.CategorizationRulePerformance, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) CreateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRuleTemplateByID(ctx context.Context, id string) (*models.CategorizationRuleTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) UpdateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) DeleteRuleTemplate(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRuleTemplates(ctx context.Context, filter *CategorizationRuleTemplateFilter, limit, offset int) ([]*models.CategorizationRuleTemplate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CategorizationRuleRepository) GetRuleUsageStats(ctx context.Context, startDate, endDate time.Time) ([]*RuleUsageStat, error) {
	return nil, fmt.Errorf("not implemented")
}
