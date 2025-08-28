package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role string

const (
	RoleAdmin      Role = "admin"
	RoleFinance    Role = "finance"
	RoleCompliance Role = "compliance"
)

// Permission represents a permission in the system
type Permission string

const (
	// User management permissions
	PermissionCreateUser Permission = "user:create"
	PermissionReadUser   Permission = "user:read"
	PermissionUpdateUser Permission = "user:update"
	PermissionDeleteUser Permission = "user:delete"

	// Enterprise management permissions
	PermissionCreateEnterprise Permission = "enterprise:create"
	PermissionReadEnterprise   Permission = "enterprise:read"
	PermissionUpdateEnterprise Permission = "enterprise:update"
	PermissionDeleteEnterprise Permission = "enterprise:delete"

	// KYB management permissions
	PermissionManageKYB  Permission = "kyb:manage"
	PermissionApproveKYB Permission = "kyb:approve"
	PermissionRejectKYB  Permission = "kyb:reject"
	PermissionViewKYB    Permission = "kyb:view"

	// Document management permissions
	PermissionUploadDocument Permission = "document:upload"
	PermissionViewDocument   Permission = "document:view"
	PermissionVerifyDocument Permission = "document:verify"
	PermissionDeleteDocument Permission = "document:delete"

	// Compliance permissions
	PermissionManageCompliance Permission = "compliance:manage"
	PermissionViewCompliance   Permission = "compliance:view"
	PermissionRunChecks        Permission = "compliance:run_checks"

	// Smart Check permissions (for future use)
	PermissionCreateSmartCheque Permission = "smart_check:create"
	PermissionViewSmartCheque   Permission = "smart_check:view"
	PermissionApprovePayment    Permission = "payment:approve"
	PermissionProcessPayment    Permission = "payment:process"

	// System administration permissions
	PermissionViewAuditLogs Permission = "audit:view"
	PermissionManageSystem  Permission = "system:manage"
)

// RolePermissions defines the permissions for each role
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// User management
		PermissionCreateUser,
		PermissionReadUser,
		PermissionUpdateUser,
		PermissionDeleteUser,

		// Enterprise management
		PermissionCreateEnterprise,
		PermissionReadEnterprise,
		PermissionUpdateEnterprise,
		PermissionDeleteEnterprise,

		// KYB management
		PermissionManageKYB,
		PermissionApproveKYB,
		PermissionRejectKYB,
		PermissionViewKYB,

		// Document management
		PermissionUploadDocument,
		PermissionViewDocument,
		PermissionVerifyDocument,
		PermissionDeleteDocument,

		// Compliance
		PermissionManageCompliance,
		PermissionViewCompliance,
		PermissionRunChecks,

		// Smart Checks
		PermissionCreateSmartCheque,
		PermissionViewSmartCheque,
		PermissionApprovePayment,
		PermissionProcessPayment,

		// System
		PermissionViewAuditLogs,
		PermissionManageSystem,
	},
	RoleFinance: {
		// Enterprise management (limited)
		PermissionReadEnterprise,

		// Document management (limited)
		PermissionUploadDocument,
		PermissionViewDocument,

		// Smart Checks
		PermissionCreateSmartCheque,
		PermissionViewSmartCheque,
		PermissionApprovePayment,
		PermissionProcessPayment,

		// Compliance (view only)
		PermissionViewCompliance,
	},
	RoleCompliance: {
		// Enterprise management (limited)
		PermissionReadEnterprise,

		// KYB management
		PermissionManageKYB,
		PermissionApproveKYB,
		PermissionRejectKYB,
		PermissionViewKYB,

		// Document management
		PermissionViewDocument,
		PermissionVerifyDocument,

		// Compliance
		PermissionManageCompliance,
		PermissionViewCompliance,
		PermissionRunChecks,

		// Audit
		PermissionViewAuditLogs,
	},
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	EnterpriseID *uuid.UUID `json:"enterprise_id,omitempty" db:"enterprise_id"`
	Action       string     `json:"action" db:"action"`
	Resource     string     `json:"resource" db:"resource"`
	ResourceID   *string    `json:"resource_id,omitempty" db:"resource_id"`
	Details      string     `json:"details,omitempty" db:"details"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	Success      bool       `json:"success" db:"success"`
	ErrorMessage string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// AuditLogRequest represents a request to create an audit log
type AuditLogRequest struct {
	Action       string  `json:"action"`
	Resource     string  `json:"resource"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Details      string  `json:"details,omitempty"`
	Success      bool    `json:"success"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// HasPermission checks if a role has a specific permission
func (r Role) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetPermissions returns all permissions for a role
func (r Role) GetPermissions() []Permission {
	permissions, exists := RolePermissions[r]
	if !exists {
		return []Permission{}
	}

	return permissions
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	switch Role(role) {
	case RoleAdmin, RoleFinance, RoleCompliance:
		return true
	default:
		return false
	}
}
