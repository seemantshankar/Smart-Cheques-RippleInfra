package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		permission Permission
		expected   bool
	}{
		{
			name:       "admin has create user permission",
			role:       RoleAdmin,
			permission: PermissionCreateUser,
			expected:   true,
		},
		{
			name:       "finance has create smart check permission",
			role:       RoleFinance,
			permission: PermissionCreateSmartCheque,
			expected:   true,
		},
		{
			name:       "finance does not have create user permission",
			role:       RoleFinance,
			permission: PermissionCreateUser,
			expected:   false,
		},
		{
			name:       "compliance has approve KYB permission",
			role:       RoleCompliance,
			permission: PermissionApproveKYB,
			expected:   true,
		},
		{
			name:       "compliance does not have create smart check permission",
			role:       RoleCompliance,
			permission: PermissionCreateSmartCheque,
			expected:   false,
		},
		{
			name:       "invalid role has no permissions",
			role:       Role("invalid"),
			permission: PermissionCreateUser,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.HasPermission(tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRole_GetPermissions(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected int // number of permissions
	}{
		{
			name:     "admin has all permissions",
			role:     RoleAdmin,
			expected: len(RolePermissions[RoleAdmin]),
		},
		{
			name:     "finance has limited permissions",
			role:     RoleFinance,
			expected: len(RolePermissions[RoleFinance]),
		},
		{
			name:     "compliance has specific permissions",
			role:     RoleCompliance,
			expected: len(RolePermissions[RoleCompliance]),
		},
		{
			name:     "invalid role has no permissions",
			role:     Role("invalid"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := tt.role.GetPermissions()
			assert.Len(t, permissions, tt.expected)
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{
			name:     "admin is valid",
			role:     "admin",
			expected: true,
		},
		{
			name:     "finance is valid",
			role:     "finance",
			expected: true,
		},
		{
			name:     "compliance is valid",
			role:     "compliance",
			expected: true,
		},
		{
			name:     "invalid role",
			role:     "invalid",
			expected: false,
		},
		{
			name:     "empty role",
			role:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRolePermissions_Coverage(t *testing.T) {
	// Test that all roles have at least one permission
	for role, permissions := range RolePermissions {
		t.Run(string(role)+" has permissions", func(t *testing.T) {
			assert.NotEmpty(t, permissions, "Role %s should have at least one permission", role)
		})
	}

	// Test that admin has the most permissions
	adminPermissions := len(RolePermissions[RoleAdmin])
	financePermissions := len(RolePermissions[RoleFinance])
	compliancePermissions := len(RolePermissions[RoleCompliance])

	assert.Greater(t, adminPermissions, financePermissions, "Admin should have more permissions than finance")
	assert.Greater(t, adminPermissions, compliancePermissions, "Admin should have more permissions than compliance")
}

func TestPermissionConstants(t *testing.T) {
	// Test that permission constants are properly defined
	permissions := []Permission{
		PermissionCreateUser,
		PermissionReadUser,
		PermissionUpdateUser,
		PermissionDeleteUser,
		PermissionCreateEnterprise,
		PermissionReadEnterprise,
		PermissionUpdateEnterprise,
		PermissionDeleteEnterprise,
		PermissionManageKYB,
		PermissionApproveKYB,
		PermissionRejectKYB,
		PermissionViewKYB,
		PermissionUploadDocument,
		PermissionViewDocument,
		PermissionVerifyDocument,
		PermissionDeleteDocument,
		PermissionManageCompliance,
		PermissionViewCompliance,
		PermissionRunChecks,
		PermissionCreateSmartCheque,
		PermissionViewSmartCheque,
		PermissionApprovePayment,
		PermissionProcessPayment,
		PermissionViewAuditLogs,
		PermissionManageSystem,
	}

	for _, permission := range permissions {
		assert.NotEmpty(t, string(permission), "Permission should not be empty")
		assert.Contains(t, string(permission), ":", "Permission should contain a colon separator")
	}
}
