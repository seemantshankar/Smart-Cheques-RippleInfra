package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

// RequirePermission creates a middleware that checks if the user has the required permission
func RequirePermission(permission models.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context (set by auth middleware)
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no user claims found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - invalid user claims",
			})
			c.Abort()
			return
		}

		// Check if user role has the required permission
		userRole := models.Role(userClaims.Role)
		if !userRole.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Forbidden - insufficient permissions",
				"required":  string(permission),
				"user_role": string(userRole),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates a middleware that checks if the user has one of the required roles
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context (set by auth middleware)
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no user claims found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - invalid user claims",
			})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		userRole := models.Role(userClaims.Role)
		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error":          "Forbidden - insufficient role",
			"required_roles": roles,
			"user_role":      string(userRole),
		})
		c.Abort()
	}
}

// RequireAdminRole creates a middleware that requires admin role
func RequireAdminRole() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireComplianceRole creates a middleware that requires compliance or admin role
func RequireComplianceRole() gin.HandlerFunc {
	return RequireRole(models.RoleCompliance, models.RoleAdmin)
}

// RequireFinanceRole creates a middleware that requires finance or admin role
func RequireFinanceRole() gin.HandlerFunc {
	return RequireRole(models.RoleFinance, models.RoleAdmin)
}

// EnterpriseOwnership creates a middleware that checks if the user belongs to the enterprise
func EnterpriseOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no user claims found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*auth.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - invalid user claims",
			})
			c.Abort()
			return
		}

		// Admin users can access any enterprise
		if userClaims.Role == string(models.RoleAdmin) {
			c.Next()
			return
		}

		// Get enterprise ID from URL parameter
		enterpriseID := c.Param("id")
		if enterpriseID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Enterprise ID is required",
			})
			c.Abort()
			return
		}

		// Check if user belongs to the enterprise
		if userClaims.EnterpriseID == nil || userClaims.EnterpriseID.String() != enterpriseID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Forbidden - you can only access your own enterprise",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
