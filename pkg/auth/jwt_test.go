package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSecretKey = "test-secret-key"
	testEmail     = "test@example.com"
	testRole      = "admin"
)

func TestGenerateToken(t *testing.T) {
	secretKey := testSecretKey
	email := testEmail
	role := testRole
	id := uuid.New()

	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// Generate access token
	token, err := jwtService.GenerateAccessToken(id, email, role, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Generate refresh token
	token, err = jwtService.GenerateRefreshToken(id)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	secretKey := testSecretKey
	email := testEmail
	role := testRole
	id := uuid.New()

	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// Generate access token
	token, err := jwtService.GenerateAccessToken(id, email, role, nil)
	require.NoError(t, err)

	// Validate access token
	claims, err := jwtService.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, id, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)

	// Generate refresh token
	token, err = jwtService.GenerateRefreshToken(id)
	require.NoError(t, err)

	// Validate refresh token
	extractedID, err := jwtService.ValidateRefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, id, extractedID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secretKey := testSecretKey

	accessTokenDuration := 15 * time.Minute
	refreshTokenDuration := 24 * time.Hour

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// Try to validate invalid token
	_, err := jwtService.ValidateAccessToken("invalid-token")
	assert.Equal(t, ErrInvalidToken, err)

	// Try to validate invalid token
	_, err = jwtService.ValidateRefreshToken("invalid-token")
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secretKey := testSecretKey
	email := testEmail
	role := testRole
	id := uuid.New()

	accessTokenDuration := -2 * time.Hour  // Expired token (2 hours ago to ensure it's definitely expired)
	refreshTokenDuration := -1 * time.Hour // Also expired

	jwtService := NewJWTService(secretKey, accessTokenDuration, refreshTokenDuration)

	// Generate expired access token
	token, err := jwtService.GenerateAccessToken(id, email, role, nil)
	require.NoError(t, err)

	// Try to validate expired token
	_, err = jwtService.ValidateAccessToken(token)
	assert.Equal(t, ErrExpiredToken, err)

	// Generate expired refresh token
	token, err = jwtService.GenerateRefreshToken(id)
	require.NoError(t, err)

	// Try to validate expired token
	_, err = jwtService.ValidateRefreshToken(token)
	assert.Equal(t, ErrExpiredToken, err)
}
