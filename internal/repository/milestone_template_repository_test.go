package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestPostgresMilestoneTemplateRepository_CreateTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	template := &models.MilestoneTemplate{
		ID:              "template-123",
		Name:            "Standard Delivery Template",
		Description:     "Template for standard package delivery milestones",
		DefaultCategory: "delivery",
		DefaultPriority: 3,
		Variables:       []string{"delivery_address", "expected_date", "carrier"},
		Version:         "1.0",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mock.ExpectExec(`INSERT INTO milestone_templates`).
		WithArgs(
			template.ID,
			template.Name,
			template.Description,
			template.DefaultCategory,
			template.DefaultPriority,
			pq.Array(template.Variables),
			template.Version,
			template.CreatedAt,
			template.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateTemplate(context.Background(), template)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	expectedTemplate := &models.MilestoneTemplate{
		ID:              templateID,
		Name:            "Standard Delivery Template",
		Description:     "Template for standard package delivery milestones",
		DefaultCategory: "delivery",
		DefaultPriority: 3,
		Variables:       []string{"delivery_address", "expected_date", "carrier"},
		Version:         "1.0",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "default_category", "default_priority",
		"variables", "version", "created_at", "updated_at",
	}).AddRow(
		expectedTemplate.ID,
		expectedTemplate.Name,
		expectedTemplate.Description,
		expectedTemplate.DefaultCategory,
		expectedTemplate.DefaultPriority,
		pq.Array(expectedTemplate.Variables),
		expectedTemplate.Version,
		expectedTemplate.CreatedAt,
		expectedTemplate.UpdatedAt,
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(rows)

	result, err := repo.GetTemplateByID(context.Background(), templateID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTemplate.ID, result.ID)
	assert.Equal(t, expectedTemplate.Name, result.Name)
	assert.Equal(t, expectedTemplate.Description, result.Description)
	assert.Equal(t, expectedTemplate.DefaultCategory, result.DefaultCategory)
	assert.Equal(t, expectedTemplate.DefaultPriority, result.DefaultPriority)
	assert.Equal(t, expectedTemplate.Variables, result.Variables)
	assert.Equal(t, expectedTemplate.Version, result.Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "nonexistent-template"

	mock.ExpectQuery(`SELECT .+ FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetTemplateByID(context.Background(), templateID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_UpdateTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	template := &models.MilestoneTemplate{
		ID:              "template-123",
		Name:            "Updated Delivery Template",
		Description:     "Updated template for package delivery milestones",
		DefaultCategory: "approval",
		DefaultPriority: 5,
		Variables:       []string{"delivery_address", "expected_date", "carrier", "signature_required"},
		Version:         "1.1",
		UpdatedAt:       time.Now(),
	}

	mock.ExpectExec(`UPDATE milestone_templates SET`).
		WithArgs(
			template.ID,
			template.Name,
			template.Description,
			template.DefaultCategory,
			template.DefaultPriority,
			pq.Array(template.Variables),
			template.Version,
			template.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateTemplate(context.Background(), template)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_UpdateTemplate_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	template := &models.MilestoneTemplate{
		ID:        "nonexistent-template",
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec(`UPDATE milestone_templates SET`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	err = repo.UpdateTemplate(context.Background(), template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_DeleteTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM milestone_template_versions WHERE template_id = \$1`).
		WithArgs(templateID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM milestone_template_shares WHERE template_id = \$1`).
		WithArgs(templateID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.DeleteTemplate(context.Background(), templateID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	limit, offset := 10, 0

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "default_category", "default_priority",
		"variables", "version", "created_at", "updated_at",
	}).AddRow(
		"template-1", "Delivery Template", "Standard delivery", "delivery", 3,
		pq.Array([]string{"address", "date"}), "1.0", time.Now(), time.Now(),
	).AddRow(
		"template-2", "Approval Template", "Standard approval", "approval", 5,
		pq.Array([]string{"approver", "deadline"}), "1.0", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_templates ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(limit, offset).
		WillReturnRows(rows)

	results, err := repo.GetTemplates(context.Background(), limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "template-1", results[0].ID)
	assert.Equal(t, "template-2", results[1].ID)
	assert.Equal(t, "Delivery Template", results[0].Name)
	assert.Equal(t, "Approval Template", results[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_InstantiateTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	variables := map[string]interface{}{
		"description":           "Package delivery to customer",
		"verification_criteria": "Delivery confirmation received",
		"estimated_duration":    int64(24 * 60 * 60 * 1000000000), // 24 hours in nanoseconds
		"risk_level":            "low",
		"critical_path":         true,
	}

	// Mock the template retrieval
	templateRows := sqlmock.NewRows([]string{
		"id", "name", "description", "default_category", "default_priority",
		"variables", "version", "created_at", "updated_at",
	}).AddRow(
		templateID, "Delivery Template", "Standard delivery template", "delivery", 3,
		pq.Array([]string{"description", "verification_criteria"}), "1.0", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(templateRows)

	milestone, err := repo.InstantiateTemplate(context.Background(), templateID, variables)
	assert.NoError(t, err)
	assert.NotNil(t, milestone)
	assert.Equal(t, "delivery", milestone.Category)
	assert.Equal(t, 3, milestone.Priority)
	assert.Equal(t, "Package delivery to customer", milestone.TriggerConditions)
	assert.Equal(t, "Delivery confirmation received", milestone.VerificationCriteria)
	assert.Equal(t, time.Duration(24*60*60*1000000000), milestone.EstimatedDuration)
	assert.Equal(t, "low", milestone.RiskLevel)
	assert.True(t, milestone.CriticalPath)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_CustomizeTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	customizations := map[string]interface{}{
		"name":             "Customized Delivery Template",
		"description":      "Customized template for express delivery",
		"default_category": "express_delivery",
		"default_priority": 5,
		"variables":        []string{"express_address", "express_date", "priority_level"},
	}

	// Mock the template retrieval
	templateRows := sqlmock.NewRows([]string{
		"id", "name", "description", "default_category", "default_priority",
		"variables", "version", "created_at", "updated_at",
	}).AddRow(
		templateID, "Original Template", "Original description", "delivery", 3,
		pq.Array([]string{"address", "date"}), "1.0", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(templateRows)

	customTemplate, err := repo.CustomizeTemplate(context.Background(), templateID, customizations)
	assert.NoError(t, err)
	assert.NotNil(t, customTemplate)
	assert.Equal(t, "Customized Delivery Template", customTemplate.Name)
	assert.Equal(t, "Customized template for express delivery", customTemplate.Description)
	assert.Equal(t, "express_delivery", customTemplate.DefaultCategory)
	assert.Equal(t, 5, customTemplate.DefaultPriority)
	assert.Equal(t, []string{"express_address", "express_date", "priority_level"}, customTemplate.Variables)
	assert.Equal(t, "1.0", customTemplate.Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateVariables(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	expectedVariables := []string{"delivery_address", "expected_date", "carrier", "signature_required"}

	rows := sqlmock.NewRows([]string{"variables"}).
		AddRow(pq.Array(expectedVariables))

	mock.ExpectQuery(`SELECT variables FROM milestone_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(rows)

	variables, err := repo.GetTemplateVariables(context.Background(), templateID)
	assert.NoError(t, err)
	assert.Equal(t, expectedVariables, variables)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_CreateTemplateVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	version := &models.MilestoneTemplate{
		ID:              "version-456",
		Name:            "Delivery Template v2.0",
		Description:     "Updated delivery template with new features",
		DefaultCategory: "delivery",
		DefaultPriority: 4,
		Variables:       []string{"address", "date", "tracking_number"},
		Version:         "2.0",
		CreatedAt:       time.Now(),
	}

	mock.ExpectExec(`INSERT INTO milestone_template_versions`).
		WithArgs(
			version.ID,
			templateID,
			version.Version,
			version.Name,
			version.Description,
			version.DefaultCategory,
			version.DefaultPriority,
			pq.Array(version.Variables),
			version.CreatedAt,
			"system",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateTemplateVersion(context.Background(), templateID, version)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"

	rows := sqlmock.NewRows([]string{
		"id", "version", "name", "description", "default_category",
		"default_priority", "variables", "created_at",
	}).AddRow(
		"version-2", "2.0", "Template v2.0", "Version 2.0 description", "delivery",
		4, pq.Array([]string{"var1", "var2"}), time.Now(),
	).AddRow(
		"version-1", "1.0", "Template v1.0", "Version 1.0 description", "delivery",
		3, pq.Array([]string{"var1"}), time.Now().Add(-24*time.Hour),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_template_versions WHERE template_id = \$1 ORDER BY created_at DESC`).
		WithArgs(templateID).
		WillReturnRows(rows)

	versions, err := repo.GetTemplateVersions(context.Background(), templateID)
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
	assert.Equal(t, "version-2", versions[0].ID)
	assert.Equal(t, "2.0", versions[0].Version)
	assert.Equal(t, "version-1", versions[1].ID)
	assert.Equal(t, "1.0", versions[1].Version)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	version := "2.0"

	rows := sqlmock.NewRows([]string{
		"id", "version", "name", "description", "default_category",
		"default_priority", "variables", "created_at",
	}).AddRow(
		"version-2", version, "Template v2.0", "Version 2.0 description", "delivery",
		4, pq.Array([]string{"var1", "var2"}), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_template_versions WHERE template_id = \$1 AND version = \$2`).
		WithArgs(templateID, version).
		WillReturnRows(rows)

	result, err := repo.GetTemplateVersion(context.Background(), templateID, version)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "version-2", result.ID)
	assert.Equal(t, version, result.Version)
	assert.Equal(t, "Template v2.0", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetLatestTemplateVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"

	rows := sqlmock.NewRows([]string{
		"id", "version", "name", "description", "default_category",
		"default_priority", "variables", "created_at",
	}).AddRow(
		"version-latest", "3.0", "Template v3.0", "Latest version description", "delivery",
		5, pq.Array([]string{"var1", "var2", "var3"}), time.Now(),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_template_versions WHERE template_id = \$1 ORDER BY created_at DESC LIMIT 1`).
		WithArgs(templateID).
		WillReturnRows(rows)

	result, err := repo.GetLatestTemplateVersion(context.Background(), templateID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "version-latest", result.ID)
	assert.Equal(t, "3.0", result.Version)
	assert.Equal(t, "Template v3.0", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_CompareTemplateVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	version1 := "1.0"
	version2 := "2.0"

	// Mock first version query
	rows1 := sqlmock.NewRows([]string{
		"id", "version", "name", "description", "default_category",
		"default_priority", "variables", "created_at",
	}).AddRow(
		"version-1", version1, "Template v1.0", "Version 1.0 description", "delivery",
		3, pq.Array([]string{"var1"}), time.Now().Add(-48*time.Hour),
	)

	// Mock second version query
	rows2 := sqlmock.NewRows([]string{
		"id", "version", "name", "description", "default_category",
		"default_priority", "variables", "created_at",
	}).AddRow(
		"version-2", version2, "Template v2.0", "Version 2.0 description", "approval",
		4, pq.Array([]string{"var1", "var2"}), time.Now().Add(-24*time.Hour),
	)

	mock.ExpectQuery(`SELECT .+ FROM milestone_template_versions WHERE template_id = \$1 AND version = \$2`).
		WithArgs(templateID, version1).
		WillReturnRows(rows1)

	mock.ExpectQuery(`SELECT .+ FROM milestone_template_versions WHERE template_id = \$1 AND version = \$2`).
		WithArgs(templateID, version2).
		WillReturnRows(rows2)

	diff, err := repo.CompareTemplateVersions(context.Background(), templateID, version1, version2)
	assert.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Equal(t, templateID, diff.TemplateID)
	assert.Equal(t, version1, diff.Version1)
	assert.Equal(t, version2, diff.Version2)

	// Check for expected changes
	assert.Len(t, diff.Changes, 3) // name, description, default_category, default_priority, variables

	// Check for added variable
	assert.Contains(t, diff.AddedFields, "var2")

	// Check for modified fields
	assert.Contains(t, diff.ModifiedFields, "name")
	assert.Contains(t, diff.ModifiedFields, "description")
	assert.Contains(t, diff.ModifiedFields, "default_category")
	assert.Contains(t, diff.ModifiedFields, "default_priority")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_ShareTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	sharedWithUserID := "user-456"
	permissions := []string{"read", "instantiate"}

	mock.ExpectExec(`INSERT INTO milestone_template_shares .+ ON CONFLICT .+ DO UPDATE SET`).
		WithArgs(
			sqlmock.AnyArg(), // share ID (generated)
			templateID,
			sharedWithUserID,
			"system",
			pq.Array(permissions),
			sqlmock.AnyArg(),      // shared_at timestamp
			pq.Array(permissions), // for ON CONFLICT UPDATE
			sqlmock.AnyArg(),      // shared_at for UPDATE
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.ShareTemplate(context.Background(), templateID, sharedWithUserID, permissions)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_RevokeTemplateAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	userID := "user-456"

	mock.ExpectExec(`DELETE FROM milestone_template_shares WHERE template_id = \$1 AND shared_with = \$2`).
		WithArgs(templateID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RevokeTemplateAccess(context.Background(), templateID, userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetSharedTemplates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	userID := "user-456"
	limit, offset := 10, 0

	rows := sqlmock.NewRows([]string{
		"id", "name", "description", "default_category", "default_priority",
		"variables", "version", "created_at", "updated_at",
	}).AddRow(
		"template-1", "Shared Template 1", "Description 1", "delivery", 3,
		pq.Array([]string{"var1"}), "1.0", time.Now(), time.Now(),
	).AddRow(
		"template-2", "Shared Template 2", "Description 2", "approval", 4,
		pq.Array([]string{"var1", "var2"}), "1.1", time.Now(), time.Now(),
	)

	mock.ExpectQuery(`SELECT t\.id, t\.name, t\.description, t\.default_category, t\.default_priority, t\.variables, t\.version, t\.created_at, t\.updated_at FROM milestone_templates t JOIN milestone_template_shares s ON t\.id = s\.template_id WHERE s\.shared_with = \$1 AND \(s\.expires_at IS NULL OR s\.expires_at > NOW\(\)\) ORDER BY s\.shared_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	results, err := repo.GetSharedTemplates(context.Background(), userID, limit, offset)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "template-1", results[0].ID)
	assert.Equal(t, "template-2", results[1].ID)
	assert.Equal(t, "Shared Template 1", results[0].Name)
	assert.Equal(t, "Shared Template 2", results[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplatePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	userID := "user-456"
	expectedPermissions := []string{"read", "instantiate", "modify"}

	rows := sqlmock.NewRows([]string{"permissions"}).
		AddRow(pq.Array(expectedPermissions))

	mock.ExpectQuery(`SELECT permissions FROM milestone_template_shares WHERE template_id = \$1 AND shared_with = \$2 AND \(expires_at IS NULL OR expires_at > NOW\(\)\)`).
		WithArgs(templateID, userID).
		WillReturnRows(rows)

	permissions, err := repo.GetTemplatePermissions(context.Background(), templateID, userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedPermissions, permissions)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplatePermissions_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	userID := "user-456"

	mock.ExpectQuery(`SELECT permissions FROM milestone_template_shares WHERE template_id = \$1 AND shared_with = \$2 AND \(expires_at IS NULL OR expires_at > NOW\(\)\)`).
		WithArgs(templateID, userID).
		WillReturnError(sql.ErrNoRows)

	permissions, err := repo.GetTemplatePermissions(context.Background(), templateID, userID)
	assert.NoError(t, err)
	assert.Empty(t, permissions)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresMilestoneTemplateRepository_GetTemplateShareList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresMilestoneTemplateRepository(db)

	templateID := "template-123"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "template_id", "shared_with", "shared_by", "permissions", "shared_at", "expires_at",
	}).AddRow(
		"share-1", templateID, "user-1", "user-admin", pq.Array([]string{"read"}), time.Now(), nil,
	).AddRow(
		"share-2", templateID, "user-2", "user-admin", pq.Array([]string{"read", "instantiate"}), time.Now(), expiresAt,
	)

	mock.ExpectQuery(`SELECT id, template_id, shared_with, shared_by, permissions, shared_at, expires_at FROM milestone_template_shares WHERE template_id = \$1 ORDER BY shared_at DESC`).
		WithArgs(templateID).
		WillReturnRows(rows)

	shares, err := repo.GetTemplateShareList(context.Background(), templateID)
	assert.NoError(t, err)
	assert.Len(t, shares, 2)
	assert.Equal(t, "share-1", shares[0].ID)
	assert.Equal(t, "share-2", shares[1].ID)
	assert.Equal(t, "user-1", shares[0].SharedWith)
	assert.Equal(t, "user-2", shares[1].SharedWith)
	assert.Equal(t, []string{"read"}, shares[0].Permissions)
	assert.Equal(t, []string{"read", "instantiate"}, shares[1].Permissions)
	assert.Nil(t, shares[0].ExpiresAt)
	assert.NotNil(t, shares[1].ExpiresAt)
	assert.Equal(t, expiresAt, *shares[1].ExpiresAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}
