package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request payload", err))
		return
	}

	user, err := h.authService.RegisterUser(&req)
	if err != nil {
		switch err {
		case services.ErrUserAlreadyExists:
			_ = c.Error(models.NewAppError(http.StatusConflict, "User already exists", err))
		default:
			_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to register user", err))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request payload", err))
		return
	}

	response, err := h.authService.LoginUser(&req)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Invalid credentials", err))
		default:
			_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to login", err))
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(models.NewAppError(http.StatusBadRequest, "Invalid request payload", err))
		return
	}

	response, err := h.authService.RefreshToken(&req)
	if err != nil {
		switch err {
		case services.ErrInvalidRefreshToken:
			_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Invalid refresh token", err))
		default:
			_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to refresh token", err))
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Unauthorized", nil))
		return
	}

	uid, ok := userID.(uuid.UUID)
	if !ok {
		_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Invalid user ID", nil))
		return
	}

	if err := h.authService.LogoutUser(uid); err != nil {
		_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Failed to logout", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me returns the current user's information
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user claims from context (set by auth middleware)
	claims, exists := c.Get("user_claims")
	if !exists {
		_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Unauthorized", nil))
		return
	}

	userClaims, ok := claims.(*auth.JWTClaims)
	if !ok {
		_ = c.Error(models.NewAppError(http.StatusInternalServerError, "Invalid user claims", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            userClaims.UserID,
			"email":         userClaims.Email,
			"role":          userClaims.Role,
			"enterprise_id": userClaims.EnterpriseID,
		},
	})
}

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Authorization header required", nil))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Invalid authorization header format", nil))
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := authService.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case auth.ErrExpiredToken:
				_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Token has expired", err))
			case auth.ErrInvalidToken:
				_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Invalid token", err))
			default:
				_ = c.Error(models.NewAppError(http.StatusUnauthorized, "Token validation failed", err))
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("enterprise_id", claims.EnterpriseID)
		c.Set("user_claims", claims)

		c.Next()
	}
}
