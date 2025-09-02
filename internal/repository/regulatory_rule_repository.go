package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
)

// RegulatoryRuleRepository handles database operations for regulatory rules
type RegulatoryRuleRepository struct {
	db *sql.DB
}

// NewRegulatoryRuleRepository creates a new regulatory rule repository
func NewRegulatoryRuleRepository(db *sql.DB) *RegulatoryRuleRepository {
	return &RegulatoryRuleRepository{db: db}
}

// CreateRegulatoryRule creates a new regulatory rule
func (r *RegulatoryRuleRepository) CreateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error {
	query := `
		INSERT INTO regulatory_rules (
			id, name, description, category, jurisdiction, priority, rule_type,
			conditions, thresholds, actions, status, version, effective_at, expires_at,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.Category,
		rule.Jurisdiction,
		rule.Priority,
		rule.RuleType,
		rule.Conditions,
		rule.Thresholds,
		rule.Actions,
		rule.Status,
		rule.Version,
		rule.EffectiveAt,
		rule.ExpiresAt,
		rule.CreatedBy,
		rule.UpdatedBy,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	return err
}

// GetRegulatoryRule retrieves a regulatory rule by ID
func (r *RegulatoryRuleRepository) GetRegulatoryRule(ctx context.Context, ruleID string) (*models.RegulatoryRule, error) {
	query := `
		SELECT id, name, description, category, jurisdiction, priority, rule_type,
		       conditions, thresholds, actions, status, version, effective_at, expires_at,
		       created_by, updated_by, created_at, updated_at
		FROM regulatory_rules
		WHERE id = $1
	`

	var rule models.RegulatoryRule
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Description,
		&rule.Category,
		&rule.Jurisdiction,
		&rule.Priority,
		&rule.RuleType,
		&rule.Conditions,
		&rule.Thresholds,
		&rule.Actions,
		&rule.Status,
		&rule.Version,
		&rule.EffectiveAt,
		&expiresAt,
		&rule.CreatedBy,
		&rule.UpdatedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		rule.ExpiresAt = &expiresAt.Time
	}

	return &rule, nil
}

// GetActiveRegulatoryRules retrieves all active regulatory rules for a jurisdiction
func (r *RegulatoryRuleRepository) GetActiveRegulatoryRules(ctx context.Context, jurisdiction string) ([]*models.RegulatoryRule, error) {
	query := `
		SELECT id, name, description, category, jurisdiction, priority, rule_type,
		       conditions, thresholds, actions, status, version, effective_at, expires_at,
		       created_by, updated_by, created_at, updated_at
		FROM regulatory_rules
		WHERE jurisdiction = $1 AND status = 'active' AND effective_at <= $2
		AND (expires_at IS NULL OR expires_at > $2)
		ORDER BY priority DESC, created_at ASC
	`

	now := time.Now()
	rows, err := r.db.QueryContext(ctx, query, jurisdiction, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.RegulatoryRule
	for rows.Next() {
		var rule models.RegulatoryRule
		var expiresAt sql.NullTime

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.Category,
			&rule.Jurisdiction,
			&rule.Priority,
			&rule.RuleType,
			&rule.Conditions,
			&rule.Thresholds,
			&rule.Actions,
			&rule.Status,
			&rule.Version,
			&rule.EffectiveAt,
			&expiresAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			rule.ExpiresAt = &expiresAt.Time
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetRegulatoryRulesByCategory retrieves regulatory rules by category
func (r *RegulatoryRuleRepository) GetRegulatoryRulesByCategory(ctx context.Context, jurisdiction, category string) ([]*models.RegulatoryRule, error) {
	query := `
		SELECT id, name, description, category, jurisdiction, priority, rule_type,
		       conditions, thresholds, actions, status, version, effective_at, expires_at,
		       created_by, updated_by, created_at, updated_at
		FROM regulatory_rules
		WHERE jurisdiction = $1 AND category = $2 AND status = 'active' AND effective_at <= $3
		AND (expires_at IS NULL OR expires_at > $3)
		ORDER BY priority DESC, created_at ASC
	`

	now := time.Now()
	rows, err := r.db.QueryContext(ctx, query, jurisdiction, category, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.RegulatoryRule
	for rows.Next() {
		var rule models.RegulatoryRule
		var expiresAt sql.NullTime

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.Category,
			&rule.Jurisdiction,
			&rule.Priority,
			&rule.RuleType,
			&rule.Conditions,
			&rule.Thresholds,
			&rule.Actions,
			&rule.Status,
			&rule.Version,
			&rule.EffectiveAt,
			&expiresAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			rule.ExpiresAt = &expiresAt.Time
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// UpdateRegulatoryRule updates an existing regulatory rule
func (r *RegulatoryRuleRepository) UpdateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error {
	query := `
		UPDATE regulatory_rules
		SET name = $1, description = $2, category = $3, jurisdiction = $4, priority = $5,
		    rule_type = $6, conditions = $7, thresholds = $8, actions = $9, status = $10,
		    version = $11, effective_at = $12, expires_at = $13, updated_by = $14, updated_at = $15
		WHERE id = $16
	`

	rule.UpdatedAt = time.Now()
	rule.Version++

	_, err := r.db.ExecContext(ctx, query,
		rule.Name,
		rule.Description,
		rule.Category,
		rule.Jurisdiction,
		rule.Priority,
		rule.RuleType,
		rule.Conditions,
		rule.Thresholds,
		rule.Actions,
		rule.Status,
		rule.Version,
		rule.EffectiveAt,
		rule.ExpiresAt,
		rule.UpdatedBy,
		rule.UpdatedAt,
		rule.ID,
	)

	return err
}

// DeleteRegulatoryRule soft deletes a regulatory rule
func (r *RegulatoryRuleRepository) DeleteRegulatoryRule(ctx context.Context, ruleID string, deletedBy string) error {
	query := `
		UPDATE regulatory_rules
		SET status = 'deprecated', updated_by = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, deletedBy, time.Now(), ruleID)
	return err
}

// GetRegulatoryRuleStats retrieves statistics about regulatory rules
func (r *RegulatoryRuleRepository) GetRegulatoryRuleStats(ctx context.Context, jurisdiction string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_rules,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_rules,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_rules,
			COUNT(CASE WHEN status = 'deprecated' THEN 1 END) as deprecated_rules,
			COUNT(CASE WHEN category = 'aml' THEN 1 END) as aml_rules,
			COUNT(CASE WHEN category = 'kyc' THEN 1 END) as kyc_rules,
			COUNT(CASE WHEN category = 'sanctions' THEN 1 END) as sanctions_rules,
			COUNT(CASE WHEN category = 'tax' THEN 1 END) as tax_rules,
			COUNT(CASE WHEN category = 'industry_specific' THEN 1 END) as industry_rules
		FROM regulatory_rules
		WHERE jurisdiction = $1
	`

	var stats struct {
		TotalRules      int64 `db:"total_rules"`
		ActiveRules     int64 `db:"active_rules"`
		InactiveRules   int64 `db:"inactive_rules"`
		DeprecatedRules int64 `db:"deprecated_rules"`
		AMLRules        int64 `db:"aml_rules"`
		KYCRules        int64 `db:"kyc_rules"`
		SanctionsRules  int64 `db:"sanctions_rules"`
		TaxRules        int64 `db:"tax_rules"`
		IndustryRules   int64 `db:"industry_rules"`
	}

	err := r.db.QueryRowContext(ctx, query, jurisdiction).Scan(
		&stats.TotalRules,
		&stats.ActiveRules,
		&stats.InactiveRules,
		&stats.DeprecatedRules,
		&stats.AMLRules,
		&stats.KYCRules,
		&stats.SanctionsRules,
		&stats.TaxRules,
		&stats.IndustryRules,
	)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_rules":      stats.TotalRules,
		"active_rules":     stats.ActiveRules,
		"inactive_rules":   stats.InactiveRules,
		"deprecated_rules": stats.DeprecatedRules,
		"aml_rules":        stats.AMLRules,
		"kyc_rules":        stats.KYCRules,
		"sanctions_rules":  stats.SanctionsRules,
		"tax_rules":        stats.TaxRules,
		"industry_rules":   stats.IndustryRules,
	}, nil
}

// SearchRegulatoryRules searches for regulatory rules with filters
func (r *RegulatoryRuleRepository) SearchRegulatoryRules(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.RegulatoryRule, error) {
	query := `
		SELECT id, name, description, category, jurisdiction, priority, rule_type,
		       conditions, thresholds, actions, status, version, effective_at, expires_at,
		       created_by, updated_by, created_at, updated_at
		FROM regulatory_rules
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	// Add filters
	if jurisdiction, ok := filters["jurisdiction"].(string); ok && jurisdiction != "" {
		query += fmt.Sprintf(" AND jurisdiction = $%d", argIndex)
		args = append(args, jurisdiction)
		argIndex++
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if ruleType, ok := filters["rule_type"].(string); ok && ruleType != "" {
		query += fmt.Sprintf(" AND rule_type = $%d", argIndex)
		args = append(args, ruleType)
		argIndex++
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY priority DESC, created_at ASC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.RegulatoryRule
	for rows.Next() {
		var rule models.RegulatoryRule
		var expiresAt sql.NullTime

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.Category,
			&rule.Jurisdiction,
			&rule.Priority,
			&rule.RuleType,
			&rule.Conditions,
			&rule.Thresholds,
			&rule.Actions,
			&rule.Status,
			&rule.Version,
			&rule.EffectiveAt,
			&expiresAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			rule.ExpiresAt = &expiresAt.Time
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}
