package models

import (
	"fmt"
	"net/http"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorCode        string `json:"error_code,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	Details          string `json:"details,omitempty"`
}

// AppError represents a custom application error
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
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

// Common error types
var (
	ErrBadRequest          = &AppError{Code: http.StatusBadRequest, Message: "Bad Request"}
	ErrUnauthorized        = &AppError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden           = &AppError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrNotFound            = &AppError{Code: http.StatusNotFound, Message: "Not Found"}
	ErrInternalServerError = &AppError{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
	ErrConflict            = &AppError{Code: http.StatusConflict, Message: "Conflict"}
)
