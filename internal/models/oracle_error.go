package models

import (
	"fmt"
	"net/http"
)

// OracleError represents a custom application error for the oracle service
type OracleError struct {
	Code    int
	Message string
	Err     error
}

// Error implements the error interface
func (e *OracleError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// HTTPStatusCode returns the HTTP status code for the error
func (e *OracleError) HTTPStatusCode() int {
	if e.Code == 0 {
		return http.StatusInternalServerError
	}
	return e.Code
}

// NewOracleError creates a new OracleError
func NewOracleError(code int, message string, err error) *OracleError {
	return &OracleError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error types
var (
	ErrOracleBadRequest          = &OracleError{Code: http.StatusBadRequest, Message: "Bad Request"}
	ErrOracleUnauthorized        = &OracleError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrOracleForbidden           = &OracleError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrOracleNotFound            = &OracleError{Code: http.StatusNotFound, Message: "Not Found"}
	ErrOracleInternalServerError = &OracleError{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
	ErrOracleConflict            = &OracleError{Code: http.StatusConflict, Message: "Conflict"}
)
