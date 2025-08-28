package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

func setupTestDB(t *testing.T) *sql.DB {
	// For integration tests, you would typically use a test database
	// For now, we'll skip the actual database setup and use mocks in unit tests
	// This is a placeholder for when you have a test database available
	t.Skip("Integration tests require a test database setup")
	return nil
}

func setupAuthTestRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Initialize services
	jwtService := auth.NewJWTService("test-secret-key", 15*time.Minute, 24*time.Hour)
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, jwtService)
	authHandler := handlers.NewAuthHandler(authService)

	r := gin.New()

	// Authentication endpoints
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", handlers.AuthMiddleware(authService), authHandler.Logout)
		auth.GET("/me", handlers.AuthMiddleware(authService), authHandler.Me)
	}

	return r
}

func TestAuthEndpoints_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupAuthTestRouter(db)

	// Test user registration
	t.Run("Register User", func(t *testing.T) {
		reqBody := models.UserRegistrationRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			Role:      "admin",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "User registered successfully", response["message"])
	})

	// Test user login
	t.Run("Login User", func(t *testing.T) {
		reqBody := models.UserLoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.UserLoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "test@example.com", response.User.Email)
	})

	// Test protected endpoint
	t.Run("Access Protected Endpoint", func(t *testing.T) {
		// First login to get token
		loginReq := models.UserLoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		jsonBody, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse models.UserLoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)

		// Use token to access protected endpoint
		req, _ = http.NewRequest("GET", "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		user := response["user"].(map[string]interface{})
		assert.Equal(t, "test@example.com", user["email"])
	})
}
