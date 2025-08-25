package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	userID := uuid.New()
	email := "test@example.com"
	role := "admin"
	enterpriseID := uuid.New()

	// Generate access token
	token, err := jwtService.GenerateAccessToken(userID, email, role, &enterpriseID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate access token
	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, &enterpriseID, claims.EnterpriseID)
}

func TestJWTService_GenerateAndValidateRefreshToken(t *testing.T) {
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	userID := uuid.New()

	// Generate refresh token
	token, err := jwtService.GenerateRefreshToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate refresh token
	extractedUserID, err := jwtService.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}

func TestJWTService_ValidateExpiredToken(t *testing.T) {
	secretKey := "test-secret-key"
	accessTokenDuration := -1 * time.Hour // Expired token
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	userID := uuid.New()
	email := "test@example.com"
	role := "admin"

	// Generate expired access token
	token, err := jwtService.GenerateAccessToken(userID, email, role, nil)
	require.NoError(t, err)

	// Try to validate expired token
	_, err = jwtService.ValidateAccessToken(token)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestJWTService_ValidateInvalidToken(t *testing.T) {
	secretKey := "test-secret-key"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// Try to validate invalid token
	_, err := jwtService.ValidateAccessToken("invalid-token")
	assert.Equal(t, ErrInvalidToken, err)
}

func TestJWTService_ValidateTokenWithWrongSecret(t *testing.T) {
	secretKey1 := "test-secret-key-1"
	secretKey2 := "test-secret-key-2"
	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService1 := NewJWTService(secretKey1, accessTokenDuration, refreshTokenDuration)
	jwtService2 := NewJWTService(secretKey2, accessTokenDuration, refreshTokenDuration)

	userID := uuid.New()
	email := "test@example.com"
	role := "admin"

	// Generate token with first service
	token, err := jwtService1.GenerateAccessToken(userID, email, role, nil)
	require.NoError(t, err)

	// Try to validate with second service (different secret)
	_, err = jwtService2.ValidateAccessToken(token)
	assert.Equal(t, ErrInvalidToken, err)
}