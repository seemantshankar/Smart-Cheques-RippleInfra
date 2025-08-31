package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/internal/models"
)

func TestGlobalErrorHandler_Middleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with our error handling middleware
	r := gin.New()
	r.Use(middleware.ErrorHandler())

	// Add a test route that deliberately causes an error
	r.POST("/test-error", func(c *gin.Context) {
		// Simulate an error being added to the context
		c.Error(middleware.NewAppError(http.StatusBadRequest, "Test error message", nil))
	})

	// Create a request
	req, _ := http.NewRequest("POST", "/test-error", nil)
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse the response
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test error message", response.Error)
	assert.Equal(t, "BAD_REQUEST", response.ErrorCode)
}

func TestGlobalErrorHandler_ValidationErrors(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with our error handling middleware
	r := gin.New()
	r.Use(middleware.ErrorHandler())

	// Add a test route that uses binding
	r.POST("/test-validation", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a request with invalid data
	reqBody := []byte(`{
		"email": "invalid-email",
		"password": "123"
	}`)
	req, _ := http.NewRequest("POST", "/test-validation", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse the response
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request data", response.Error)
	assert.Equal(t, "BAD_REQUEST", response.ErrorCode)
}

func TestGlobalErrorHandler_UnhandledErrors(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router with our error handling middleware
	r := gin.New()
	r.Use(middleware.ErrorHandler())

	// Add a test route that adds a generic error
	r.POST("/test-generic-error", func(c *gin.Context) {
		// Simulate a generic error being added to the context
		c.Error(errors.New("generic error"))
	})

	// Create a request
	req, _ := http.NewRequest("POST", "/test-generic-error", nil)
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse the response
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}
