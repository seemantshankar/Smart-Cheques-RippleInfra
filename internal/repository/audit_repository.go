package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// AuditRepository handles database operations for audit logs
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *AuditRepository) CreateAuditLog(auditLog *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, user_id, enterprise_id, action, resource, resource_id, details,
			ip_address, user_agent, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	auditLog.ID = uuid.New()
	auditLog.CreatedAt = time.Now()

	_, err := r.db.Exec(query,
		auditLog.ID,
		auditLog.UserID,
		auditLog.EnterpriseID,
		auditLog.Action,
		auditLog.Resource,
		auditLog.ResourceID,
		auditLog.Details,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.Success,
		auditLog.ErrorMessage,
		auditLog.CreatedAt,
	)

	return err
}

// GetAuditLogs retrieves audit logs with optional filtering
func (r *AuditRepository) GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT id, user_id, enterprise_id, action, resource, resource_id, details,
		       ip_address, user_agent, success, error_message, created_at
		FROM audit_logs
	`

	if userID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}

	if enterpriseID != nil {
		conditions = append(conditions, fmt.Sprintf("enterprise_id = $%d", argIndex))
		args = append(args, *enterpriseID)
		argIndex++
	}

	if action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, action)
		argIndex++
	}

	if resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, resource)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auditLogs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var enterpriseID sql.NullString
		var resourceID sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&enterpriseID,
			&log.Action,
			&log.Resource,
			&resourceID,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&log.Success,
			&log.ErrorMessage,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if enterpriseID.Valid {
			enterpriseUUID, err := uuid.Parse(enterpriseID.String)
			if err == nil {
				log.EnterpriseID = &enterpriseUUID
			}
		}

		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}

		auditLogs = append(auditLogs, log)
	}

	return auditLogs, nil
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (r *AuditRepository) GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	return r.GetAuditLogs(&userID, nil, "", "", limit, offset)
}

// GetAuditLogsByEnterprise retrieves audit logs for a specific enterprise
func (r *AuditRepository) GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	return r.GetAuditLogs(nil, &enterpriseID, "", "", limit, offset)
}
