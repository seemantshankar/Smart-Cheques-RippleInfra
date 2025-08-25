package services

import (
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// AuditService handles audit logging business logic
type AuditService struct {
	auditRepo repository.AuditRepositoryInterface
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo repository.AuditRepositoryInterface) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// LogAction logs an action to the audit trail
func (s *AuditService) LogAction(userID uuid.UUID, enterpriseID *uuid.UUID, req *models.AuditLogRequest, ipAddress, userAgent string) error {
	auditLog := &models.AuditLog{
		UserID:       userID,
		EnterpriseID: enterpriseID,
		Action:       req.Action,
		Resource:     req.Resource,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      req.Success,
		ErrorMessage: req.ErrorMessage,
	}

	return s.auditRepo.CreateAuditLog(auditLog)
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	return s.auditRepo.GetAuditLogs(userID, enterpriseID, action, resource, limit, offset)
}

// GetAuditLogsByUser retrieves audit logs for a specific user
func (s *AuditService) GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	return s.auditRepo.GetAuditLogsByUser(userID, limit, offset)
}

// GetAuditLogsByEnterprise retrieves audit logs for a specific enterprise
func (s *AuditService) GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	return s.auditRepo.GetAuditLogsByEnterprise(enterpriseID, limit, offset)
}