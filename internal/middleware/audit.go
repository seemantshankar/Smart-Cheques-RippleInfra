package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

// AuditMiddleware creates a middleware that logs actions to the audit trail
func AuditMiddleware(auditService *services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for health checks and non-sensitive endpoints
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Capture request body for audit logging
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture response status
		writer := &auditResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
		}
		c.Writer = writer

		// Process the request
		c.Next()

		// Log the action after request processing
		logAuditAction(c, auditService, writer.statusCode, requestBody)
	}
}

// auditResponseWriter wraps gin.ResponseWriter to capture status code
type auditResponseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// shouldSkipAudit determines if an endpoint should be skipped for audit logging
func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/health/messaging",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// logAuditAction logs the action to the audit trail
func logAuditAction(c *gin.Context, auditService *services.AuditService, statusCode int, requestBody []byte) {
	// Get user information from context
	claims, exists := c.Get("user_claims")
	if !exists {
		// Skip audit logging if no user claims (unauthenticated requests)
		return
	}

	userClaims, ok := claims.(*auth.JWTClaims)
	if !ok {
		return
	}

	// Determine action and resource from the request
	action := getActionFromMethod(c.Request.Method)
	resource := getResourceFromPath(c.Request.URL.Path)
	resourceID := getResourceIDFromPath(c.Request.URL.Path)

	// Create audit log request
	auditReq := &models.AuditLogRequest{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Success:    statusCode < 400,
	}

	// Add error message if request failed
	if statusCode >= 400 {
		auditReq.ErrorMessage = http.StatusText(statusCode)
	}

	// Add request details for sensitive operations
	if isSensitiveOperation(action, resource) {
		auditReq.Details = getSanitizedRequestDetails(c.Request.Method, c.Request.URL.Path, requestBody)
	}

	// Log the action
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	_ = auditService.LogAction(
		userClaims.UserID,
		userClaims.EnterpriseID,
		auditReq,
		ipAddress,
		userAgent,
	)
}

// getActionFromMethod maps HTTP methods to audit actions
func getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}

// getResourceFromPath extracts the resource type from the URL path
func getResourceFromPath(path string) string {
	// Simple path parsing - in a real implementation, you might use a more sophisticated approach
	if contains(path, "/auth/") {
		return "authentication"
	}
	if contains(path, "/enterprises") {
		return "enterprise"
	}
	if contains(path, "/documents") {
		return "document"
	}
	if contains(path, "/users") {
		return "user"
	}
	if contains(path, "/kyb") {
		return "kyb"
	}
	if contains(path, "/compliance") {
		return "compliance"
	}

	return "unknown"
}

// getResourceIDFromPath extracts the resource ID from the URL path
func getResourceIDFromPath(path string) *string {
	// Extract UUID from path segments
	segments := splitPath(path)
	for _, segment := range segments {
		if isUUID(segment) {
			return &segment
		}
	}
	return nil
}

// isSensitiveOperation determines if an operation should have detailed logging
func isSensitiveOperation(action, resource string) bool {
	sensitiveOperations := map[string][]string{
		"enterprise": {"create", "update"},
		"user":       {"create", "update", "delete"},
		"kyb":        {"update"},
		"document":   {"create", "update"},
	}

	actions, exists := sensitiveOperations[resource]
	if !exists {
		return false
	}

	for _, a := range actions {
		if a == action {
			return true
		}
	}

	return false
}

// getSanitizedRequestDetails creates sanitized request details for audit logging
func getSanitizedRequestDetails(method, path string, body []byte) string {
	details := method + " " + path

	// Add sanitized body for certain operations (remove sensitive data)
	if len(body) > 0 && len(body) < 1000 { // Only log small request bodies
		// In a real implementation, you would sanitize sensitive fields like passwords
		details += " - Request body length: " + string(rune(len(body))) + " bytes"
	}

	return details
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

func splitPath(path string) []string {
	segments := []string{}
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		segments = append(segments, current)
	}
	return segments
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
