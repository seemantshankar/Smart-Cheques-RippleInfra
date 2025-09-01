package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smart-payment-infrastructure/internal/models"
)

// AppError represents a custom application error
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// HTTPStatusCode returns the HTTP status code for the error
func (e *AppError) HTTPStatusCode() int {
	if e.Code == 0 {
		return http.StatusInternalServerError
	}
	return e.Code
}

// NewAppError creates a new AppError
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrorHandler is a global error handling middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Process all errors and respond with the most relevant one
			processErrors(c)
			return
		}
	}
}

// processErrors handles the error processing and response
func processErrors(c *gin.Context) {
	// Get all errors
	errs := c.Errors

	// Check if any of the errors is our custom AppError
	for _, err := range errs {
		var appErr *AppError
		if errors.As(err.Err, &appErr) {
			// Use our custom error
			c.JSON(appErr.HTTPStatusCode(), models.ErrorResponse{
				Error:     appErr.Message,
				ErrorCode: getErrorCode(appErr.HTTPStatusCode()),
				Details:   getErrorDetails(appErr),
			})
			return
		}
	}

	// If no custom AppError found, check for specific error types
	for _, err := range errs {
		// Handle validation errors (common in Gin with binding)
		if strings.Contains(err.Error(), "binding") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Err.Error(), "required") ||
			strings.Contains(err.Err.Error(), "email") ||
			strings.Contains(err.Err.Error(), "min") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:     "Invalid request data",
				ErrorCode: "BAD_REQUEST",
				Details:   err.Error(),
			})
			return
		}
	}

	// Handle other types of errors
	if len(errs) > 0 {
		lastErr := errs.Last()
		if lastErr != nil {
			// For generic errors, treat them as client errors (400) unless they are
			// explicitly marked as private errors with sensitive information
			if lastErr.Type == gin.ErrorTypePrivate {
				// Only return 500 for truly private errors that shouldn't be exposed
				// For most generic errors, treat them as client errors
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:     lastErr.Error(),
					ErrorCode: "BAD_REQUEST",
				})
			} else {
				// Handle public errors
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:     lastErr.Error(),
					ErrorCode: "BAD_REQUEST",
				})
			}
			return
		}
	}

	// Default error response
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error:     "Internal Server Error",
		ErrorCode: "INTERNAL_ERROR",
	})
}

// getErrorCode returns a standardized error code string
func getErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// getErrorDetails extracts details from an AppError
func getErrorDetails(appErr *AppError) string {
	if appErr.Err != nil {
		return appErr.Err.Error()
	}
	return ""
}
