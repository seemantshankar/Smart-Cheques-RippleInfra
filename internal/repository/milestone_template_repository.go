package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/models"
)

// PostgresMilestoneTemplateRepository implements MilestoneTemplateRepositoryInterface using PostgreSQL
type PostgresMilestoneTemplateRepository struct {
	db *sql.DB
}

func NewPostgresMilestoneTemplateRepository(db *sql.DB) MilestoneTemplateRepositoryInterface {
	return &PostgresMilestoneTemplateRepository{db: db}
}

// Template CRUD operations

// CreateTemplate creates a new milestone template
func (r *PostgresMilestoneTemplateRepository) CreateTemplate(ctx context.Context, template *models.MilestoneTemplate) error {
	query := `
		INSERT INTO milestone_templates (
			id, name, description, default_category, default_priority, variables, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	_, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Description,
		template.DefaultCategory,
		template.DefaultPriority,
		pq.Array(template.Variables),
		template.Version,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create milestone template: %w", err)
	}
	return nil
}

// GetTemplateByID retrieves a template by ID
func (r *PostgresMilestoneTemplateRepository) GetTemplateByID(ctx context.Context, id string) (*models.MilestoneTemplate, error) {
	query := `
		SELECT id, name, description, default_category, default_priority, variables, version, created_at, updated_at
		FROM milestone_templates
		WHERE id = $1`

	template := &models.MilestoneTemplate{}
	var variables pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.DefaultCategory,
		&template.DefaultPriority,
		&variables,
		&template.Version,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone template with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get milestone template by ID: %w", err)
	}

	template.Variables = []string(variables)
	return template, nil
}

// UpdateTemplate updates an existing template
func (r *PostgresMilestoneTemplateRepository) UpdateTemplate(ctx context.Context, template *models.MilestoneTemplate) error {
	query := `
		UPDATE milestone_templates SET
			name = $2, description = $3, default_category = $4, default_priority = $5,
			variables = $6, version = $7, updated_at = $8
		WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Description,
		template.DefaultCategory,
		template.DefaultPriority,
		pq.Array(template.Variables),
		template.Version,
		template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update milestone template: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("milestone template with ID %s not found", template.ID)
	}

	return nil
}

// DeleteTemplate deletes a template by ID
func (r *PostgresMilestoneTemplateRepository) DeleteTemplate(ctx context.Context, id string) error {
	// Start transaction to delete template and related data
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete template versions
	_, err = tx.ExecContext(ctx, "DELETE FROM milestone_template_versions WHERE template_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete template versions: %w", err)
	}

	// Delete template shares
	_, err = tx.ExecContext(ctx, "DELETE FROM milestone_template_shares WHERE template_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete template shares: %w", err)
	}

	// Delete main template
	res, err := tx.ExecContext(ctx, "DELETE FROM milestone_templates WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete milestone template: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("milestone template with ID %s not found", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTemplates retrieves templates with pagination
func (r *PostgresMilestoneTemplateRepository) GetTemplates(ctx context.Context, limit, offset int) ([]*models.MilestoneTemplate, error) {
	query := `
		SELECT id, name, description, default_category, default_priority, variables, version, created_at, updated_at
		FROM milestone_templates
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.MilestoneTemplate
	for rows.Next() {
		template := &models.MilestoneTemplate{}
		var variables pq.StringArray

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.DefaultCategory,
			&template.DefaultPriority,
			&variables,
			&template.Version,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone template: %w", err)
		}

		template.Variables = []string(variables)
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

// Template instantiation and customization

// InstantiateTemplate creates a milestone from a template with variables
func (r *PostgresMilestoneTemplateRepository) InstantiateTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*models.ContractMilestone, error) {
	// Get the template
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create milestone from template
	milestone := &models.ContractMilestone{
		Category:           template.DefaultCategory,
		Priority:           template.DefaultPriority,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		SequenceNumber:     1, // Default value, should be set by caller
		CriticalPath:       false,
		PercentageComplete: 0,
		RiskLevel:          "medium",
		CriticalityScore:   50,
	}

	// Apply variable substitutions
	if description, exists := variables["description"]; exists {
		if desc, ok := description.(string); ok {
			milestone.TriggerConditions = desc
		}
	}

	if verification, exists := variables["verification_criteria"]; exists {
		if criteria, ok := verification.(string); ok {
			milestone.VerificationCriteria = criteria
		}
	}

	if duration, exists := variables["estimated_duration"]; exists {
		if dur, ok := duration.(int64); ok {
			milestone.EstimatedDuration = time.Duration(dur)
		}
	}

	if startDate, exists := variables["estimated_start_date"]; exists {
		if date, ok := startDate.(time.Time); ok {
			milestone.EstimatedStartDate = &date
		}
	}

	if endDate, exists := variables["estimated_end_date"]; exists {
		if date, ok := endDate.(time.Time); ok {
			milestone.EstimatedEndDate = &date
		}
	}

	if riskLevel, exists := variables["risk_level"]; exists {
		if risk, ok := riskLevel.(string); ok {
			milestone.RiskLevel = risk
		}
	}

	if criticalPath, exists := variables["critical_path"]; exists {
		if critical, ok := criticalPath.(bool); ok {
			milestone.CriticalPath = critical
		}
	}

	if contingency, exists := variables["contingency_plans"]; exists {
		if plans, ok := contingency.([]string); ok {
			milestone.ContingencyPlans = plans
		}
	}

	return milestone, nil
}

// CustomizeTemplate creates a customized version of a template
func (r *PostgresMilestoneTemplateRepository) CustomizeTemplate(ctx context.Context, templateID string, customizations map[string]interface{}) (*models.MilestoneTemplate, error) {
	// Get the original template
	template, err := r.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create a copy with customizations
	customTemplate := &models.MilestoneTemplate{
		Name:            template.Name,
		Description:     template.Description,
		DefaultCategory: template.DefaultCategory,
		DefaultPriority: template.DefaultPriority,
		Variables:       make([]string, len(template.Variables)),
		Version:         "1.0", // Start with version 1.0 for customized template
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	copy(customTemplate.Variables, template.Variables)

	// Apply customizations
	if name, exists := customizations["name"]; exists {
		if nameStr, ok := name.(string); ok {
			customTemplate.Name = nameStr
		}
	}

	if description, exists := customizations["description"]; exists {
		if descStr, ok := description.(string); ok {
			customTemplate.Description = descStr
		}
	}

	if category, exists := customizations["default_category"]; exists {
		if catStr, ok := category.(string); ok {
			customTemplate.DefaultCategory = catStr
		}
	}

	if priority, exists := customizations["default_priority"]; exists {
		if priInt, ok := priority.(int); ok {
			customTemplate.DefaultPriority = priInt
		}
	}

	if variables, exists := customizations["variables"]; exists {
		if vars, ok := variables.([]string); ok {
			customTemplate.Variables = vars
		}
	}

	return customTemplate, nil
}

// GetTemplateVariables retrieves variables for a template
func (r *PostgresMilestoneTemplateRepository) GetTemplateVariables(ctx context.Context, templateID string) ([]string, error) {
	query := `SELECT variables FROM milestone_templates WHERE id = $1`

	var variables pq.StringArray
	err := r.db.QueryRowContext(ctx, query, templateID).Scan(&variables)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone template with ID %s not found", templateID)
		}
		return nil, fmt.Errorf("failed to get template variables: %w", err)
	}

	return []string(variables), nil
}

// Template versioning and change tracking

// CreateTemplateVersion creates a new version of a template
func (r *PostgresMilestoneTemplateRepository) CreateTemplateVersion(ctx context.Context, templateID string, version *models.MilestoneTemplate) error {
	query := `
		INSERT INTO milestone_template_versions (
			id, template_id, version, name, description, default_category, default_priority,
			variables, created_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	_, err := r.db.ExecContext(ctx, query,
		version.ID,
		templateID,
		version.Version,
		version.Name,
		version.Description,
		version.DefaultCategory,
		version.DefaultPriority,
		pq.Array(version.Variables),
		version.CreatedAt,
		"system", // Should be passed as parameter in real implementation
	)
	if err != nil {
		return fmt.Errorf("failed to create template version: %w", err)
	}
	return nil
}

// GetTemplateVersions retrieves all versions of a template
func (r *PostgresMilestoneTemplateRepository) GetTemplateVersions(ctx context.Context, templateID string) ([]*models.MilestoneTemplate, error) {
	query := `
		SELECT id, version, name, description, default_category, default_priority, variables, created_at
		FROM milestone_template_versions
		WHERE template_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.MilestoneTemplate
	for rows.Next() {
		version := &models.MilestoneTemplate{}
		var variables pq.StringArray

		err := rows.Scan(
			&version.ID,
			&version.Version,
			&version.Name,
			&version.Description,
			&version.DefaultCategory,
			&version.DefaultPriority,
			&variables,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template version: %w", err)
		}

		version.Variables = []string(variables)
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// GetTemplateVersion retrieves a specific version of a template
func (r *PostgresMilestoneTemplateRepository) GetTemplateVersion(ctx context.Context, templateID, version string) (*models.MilestoneTemplate, error) {
	query := `
		SELECT id, version, name, description, default_category, default_priority, variables, created_at
		FROM milestone_template_versions
		WHERE template_id = $1 AND version = $2`

	template := &models.MilestoneTemplate{}
	var variables pq.StringArray

	err := r.db.QueryRowContext(ctx, query, templateID, version).Scan(
		&template.ID,
		&template.Version,
		&template.Name,
		&template.Description,
		&template.DefaultCategory,
		&template.DefaultPriority,
		&variables,
		&template.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template version %s for template %s not found", version, templateID)
		}
		return nil, fmt.Errorf("failed to get template version: %w", err)
	}

	template.Variables = []string(variables)
	return template, nil
}

// GetLatestTemplateVersion retrieves the latest version of a template
func (r *PostgresMilestoneTemplateRepository) GetLatestTemplateVersion(ctx context.Context, templateID string) (*models.MilestoneTemplate, error) {
	query := `
		SELECT id, version, name, description, default_category, default_priority, variables, created_at
		FROM milestone_template_versions
		WHERE template_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	template := &models.MilestoneTemplate{}
	var variables pq.StringArray

	err := r.db.QueryRowContext(ctx, query, templateID).Scan(
		&template.ID,
		&template.Version,
		&template.Name,
		&template.Description,
		&template.DefaultCategory,
		&template.DefaultPriority,
		&variables,
		&template.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no versions found for template %s", templateID)
		}
		return nil, fmt.Errorf("failed to get latest template version: %w", err)
	}

	template.Variables = []string(variables)
	return template, nil
}

// CompareTemplateVersions compares two versions of a template
func (r *PostgresMilestoneTemplateRepository) CompareTemplateVersions(ctx context.Context, templateID, version1, version2 string) (*TemplateVersionDiff, error) {
	// Get both versions
	v1, err := r.GetTemplateVersion(ctx, templateID, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version1, err)
	}

	v2, err := r.GetTemplateVersion(ctx, templateID, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", version2, err)
	}

	diff := &TemplateVersionDiff{
		TemplateID: templateID,
		Version1:   version1,
		Version2:   version2,
		Changes:    []TemplateFieldChange{},
	}

	// Compare fields
	if v1.Name != v2.Name {
		diff.Changes = append(diff.Changes, TemplateFieldChange{
			Field:    "name",
			OldValue: v1.Name,
			NewValue: v2.Name,
			Type:     "modified",
		})
		diff.ModifiedFields = append(diff.ModifiedFields, "name")
	}

	if v1.Description != v2.Description {
		diff.Changes = append(diff.Changes, TemplateFieldChange{
			Field:    "description",
			OldValue: v1.Description,
			NewValue: v2.Description,
			Type:     "modified",
		})
		diff.ModifiedFields = append(diff.ModifiedFields, "description")
	}

	if v1.DefaultCategory != v2.DefaultCategory {
		diff.Changes = append(diff.Changes, TemplateFieldChange{
			Field:    "default_category",
			OldValue: v1.DefaultCategory,
			NewValue: v2.DefaultCategory,
			Type:     "modified",
		})
		diff.ModifiedFields = append(diff.ModifiedFields, "default_category")
	}

	if v1.DefaultPriority != v2.DefaultPriority {
		diff.Changes = append(diff.Changes, TemplateFieldChange{
			Field:    "default_priority",
			OldValue: v1.DefaultPriority,
			NewValue: v2.DefaultPriority,
			Type:     "modified",
		})
		diff.ModifiedFields = append(diff.ModifiedFields, "default_priority")
	}

	// Compare variables arrays
	v1VarSet := make(map[string]bool)
	for _, v := range v1.Variables {
		v1VarSet[v] = true
	}

	v2VarSet := make(map[string]bool)
	for _, v := range v2.Variables {
		v2VarSet[v] = true
	}

	// Find added variables
	for _, v := range v2.Variables {
		if !v1VarSet[v] {
			diff.AddedFields = append(diff.AddedFields, v)
			diff.Changes = append(diff.Changes, TemplateFieldChange{
				Field:    "variables",
				OldValue: nil,
				NewValue: v,
				Type:     "added",
			})
		}
	}

	// Find removed variables
	for _, v := range v1.Variables {
		if !v2VarSet[v] {
			diff.RemovedFields = append(diff.RemovedFields, v)
			diff.Changes = append(diff.Changes, TemplateFieldChange{
				Field:    "variables",
				OldValue: v,
				NewValue: nil,
				Type:     "removed",
			})
		}
	}

	return diff, nil
}

// Template sharing and permission management

// ShareTemplate shares a template with another user
func (r *PostgresMilestoneTemplateRepository) ShareTemplate(ctx context.Context, templateID, sharedWithUserID string, permissions []string) error {
	query := `
		INSERT INTO milestone_template_shares (id, template_id, shared_with, shared_by, permissions, shared_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (template_id, shared_with)
		DO UPDATE SET permissions = $5, shared_at = $6`

	shareID := fmt.Sprintf("share_%s_%s_%d", templateID, sharedWithUserID, time.Now().Unix())

	_, err := r.db.ExecContext(ctx, query,
		shareID,
		templateID,
		sharedWithUserID,
		"system", // Should be passed as parameter
		pq.Array(permissions),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to share template: %w", err)
	}
	return nil
}

// RevokeTemplateAccess revokes access to a template for a user
func (r *PostgresMilestoneTemplateRepository) RevokeTemplateAccess(ctx context.Context, templateID, userID string) error {
	query := `DELETE FROM milestone_template_shares WHERE template_id = $1 AND shared_with = $2`

	res, err := r.db.ExecContext(ctx, query, templateID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke template access: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no access found for user %s on template %s", userID, templateID)
	}

	return nil
}

// GetSharedTemplates retrieves templates shared with a user
func (r *PostgresMilestoneTemplateRepository) GetSharedTemplates(ctx context.Context, userID string, limit, offset int) ([]*models.MilestoneTemplate, error) {
	query := `
		SELECT t.id, t.name, t.description, t.default_category, t.default_priority,
		       t.variables, t.version, t.created_at, t.updated_at
		FROM milestone_templates t
		JOIN milestone_template_shares s ON t.id = s.template_id
		WHERE s.shared_with = $1 AND (s.expires_at IS NULL OR s.expires_at > NOW())
		ORDER BY s.shared_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.MilestoneTemplate
	for rows.Next() {
		template := &models.MilestoneTemplate{}
		var variables pq.StringArray

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.DefaultCategory,
			&template.DefaultPriority,
			&variables,
			&template.Version,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared template: %w", err)
		}

		template.Variables = []string(variables)
		templates = append(templates, template)
	}

	return templates, rows.Err()
}

// GetTemplatePermissions retrieves permissions for a user on a template
func (r *PostgresMilestoneTemplateRepository) GetTemplatePermissions(ctx context.Context, templateID, userID string) ([]string, error) {
	query := `
		SELECT permissions
		FROM milestone_template_shares
		WHERE template_id = $1 AND shared_with = $2 AND (expires_at IS NULL OR expires_at > NOW())`

	var permissions pq.StringArray
	err := r.db.QueryRowContext(ctx, query, templateID, userID).Scan(&permissions)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil // No permissions found
		}
		return nil, fmt.Errorf("failed to get template permissions: %w", err)
	}

	return []string(permissions), nil
}

// GetTemplateShareList retrieves the list of users a template is shared with
func (r *PostgresMilestoneTemplateRepository) GetTemplateShareList(ctx context.Context, templateID string) ([]*TemplateShare, error) {
	query := `
		SELECT id, template_id, shared_with, shared_by, permissions, shared_at, expires_at
		FROM milestone_template_shares
		WHERE template_id = $1
		ORDER BY shared_at DESC`

	rows, err := r.db.QueryContext(ctx, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template share list: %w", err)
	}
	defer rows.Close()

	var shares []*TemplateShare
	for rows.Next() {
		share := &TemplateShare{}
		var permissions pq.StringArray
		var expiresAt sql.NullTime

		err := rows.Scan(
			&share.ID,
			&share.TemplateID,
			&share.SharedWith,
			&share.SharedBy,
			&permissions,
			&share.SharedAt,
			&expiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template share: %w", err)
		}

		share.Permissions = []string(permissions)
		if expiresAt.Valid {
			share.ExpiresAt = &expiresAt.Time
		}

		shares = append(shares, share)
	}

	return shares, rows.Err()
}
