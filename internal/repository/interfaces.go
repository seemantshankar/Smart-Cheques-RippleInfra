package repository

import (
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	EmailExists(email string) (bool, error)
	CreateRefreshToken(token *models.RefreshToken) error
	GetRefreshToken(tokenString string) (*models.RefreshToken, error)
	RevokeRefreshToken(tokenString string) error
	RevokeAllUserRefreshTokens(userID uuid.UUID) error
}

// EnterpriseRepositoryInterface defines the interface for enterprise repository operations
type EnterpriseRepositoryInterface interface {
	CreateEnterprise(enterprise *models.Enterprise) error
	GetEnterpriseByID(id uuid.UUID) (*models.Enterprise, error)
	GetEnterpriseByRegistrationNumber(regNumber string) (*models.Enterprise, error)
	UpdateEnterpriseKYBStatus(id uuid.UUID, status models.KYBStatus) error
	UpdateEnterpriseComplianceStatus(id uuid.UUID, status models.ComplianceStatus) error
	UpdateEnterpriseXRPLWallet(id uuid.UUID, walletAddress string) error
	RegistrationNumberExists(regNumber string) (bool, error)
	CreateDocument(doc *models.EnterpriseDocument) error
	UpdateDocumentStatus(docID uuid.UUID, status models.DocumentStatus) error
}

// AuditRepositoryInterface defines the interface for audit repository operations
type AuditRepositoryInterface interface {
	CreateAuditLog(auditLog *models.AuditLog) error
	GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error)
	GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
	GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
}