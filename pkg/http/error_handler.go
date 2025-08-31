package http
package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code,omitempty"`
	Details   string `json:"details,omitempty"`
}

// ErrorHandler is a middleware for handling errors in HTTP handlers
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to capture status codes
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Pass the wrapped writer to the next handler
		next.ServeHTTP(wrapped, r)
	})
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorCode := getErrorCode(statusCode)
	
	details := ""
	if err != nil {
		details = err.Error()
	}
	
	response := ErrorResponse{
		Error:     message,
		ErrorCode: errorCode,
		Details:   details,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// responseWriter wraps http.ResponseWriter to capture status codes
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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