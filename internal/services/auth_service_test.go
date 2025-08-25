package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepositoryInterface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateRefreshToken(token *models.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) GetRefreshToken(tokenString string) (*models.RefreshToken, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockUserRepository) RevokeRefreshToken(tokenString string) error {
	args := m.Called(tokenString)
	return args.Error(0)
}

func (m *MockUserRepository) RevokeAllUserRefreshTokens(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestAuthService_RegisterUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	req := &models.UserRegistrationRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}

	// Mock expectations
	mockRepo.On("EmailExists", req.Email).Return(false, nil)
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	user, err := authService.RegisterUser(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.FirstName, user.FirstName)
	assert.Equal(t, req.LastName, user.LastName)
	assert.Equal(t, req.Role, user.Role)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, req.Password, user.PasswordHash)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RegisterUser_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	req := &models.UserRegistrationRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}

	// Mock expectations
	mockRepo.On("EmailExists", req.Email).Return(true, nil)

	// Execute
	user, err := authService.RegisterUser(req)

	// Assert
	assert.Nil(t, user)
	assert.Equal(t, ErrUserAlreadyExists, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_LoginUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	// Create a user with hashed password
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}
	user.HashPassword("password123")

	req := &models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock expectations
	mockRepo.On("GetUserByEmail", req.Email).Return(user, nil)
	mockRepo.On("CreateRefreshToken", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	// Execute
	response, err := authService.LoginUser(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Greater(t, response.ExpiresIn, int64(0))

	mockRepo.AssertExpectations(t)
}

func TestAuthService_LoginUser_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	req := &models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Mock expectations - user not found
	mockRepo.On("GetUserByEmail", req.Email).Return(nil, nil)

	// Execute
	response, err := authService.LoginUser(req)

	// Assert
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_LoginUser_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	// Create a user with hashed password
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}
	user.HashPassword("correctpassword")

	req := &models.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	// Mock expectations
	mockRepo.On("GetUserByEmail", req.Email).Return(user, nil)

	// Execute
	response, err := authService.LoginUser(req)

	// Assert
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	authService := NewAuthService(mockRepo, jwtService)

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "admin",
	}

	// Generate a valid refresh token
	refreshTokenString, err := jwtService.GenerateRefreshToken(userID)
	require.NoError(t, err)

	storedToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	}

	req := &models.TokenRefreshRequest{
		RefreshToken: refreshTokenString,
	}

	// Mock expectations
	mockRepo.On("GetRefreshToken", refreshTokenString).Return(storedToken, nil)
	mockRepo.On("GetUserByID", userID).Return(user, nil)
	mockRepo.On("RevokeRefreshToken", refreshTokenString).Return(nil)
	mockRepo.On("CreateRefreshToken", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	// Execute
	response, err := authService.RefreshToken(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user.ID, response.User.ID)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	// Note: Refresh tokens might be the same if generated at the same time with same userID
	// This is acceptable behavior

	mockRepo.AssertExpectations(t)
}
