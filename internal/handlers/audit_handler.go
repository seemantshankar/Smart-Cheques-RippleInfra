package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

// AuditHandler handles audit-related HTTP requests
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs retrieves audit logs with optional filtering
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	action := c.Query("action")
	resource := c.Query("resource")
	userIDStr := c.Query("user_id")
	enterpriseIDStr := c.Query("enterprise_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var userID *uuid.UUID
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	var enterpriseID *uuid.UUID
	if enterpriseIDStr != "" {
		if eid, err := uuid.Parse(enterpriseIDStr); err == nil {
			enterpriseID = &eid
		}
	}

	// Get audit logs
	auditLogs, err := h.auditService.GetAuditLogs(userID, enterpriseID, action, resource, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(auditLogs),
		},
	})
}

// GetUserAuditLogs retrieves audit logs for the current user
func (h *AuditHandler) GetUserAuditLogs(c *gin.Context) {
	// Get user claims from context
	claims, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userClaims, ok := claims.(*auth.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user claims",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get audit logs for the user
	auditLogs, err := h.auditService.GetAuditLogsByUser(userClaims.UserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(auditLogs),
		},
	})
}

// GetEnterpriseAuditLogs retrieves audit logs for a specific enterprise
func (h *AuditHandler) GetEnterpriseAuditLogs(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get audit logs for the enterprise
	auditLogs, err := h.auditService.GetAuditLogsByEnterprise(enterpriseID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(auditLogs),
		},
	})
}
