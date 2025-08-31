package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

// AuthServiceInterface is an interface that matches the methods of AuthService
type AuthServiceInterface struct {
	mock.Mock
}

// RegisterUser is a mock implementation
func (m *AuthServiceInterface) RegisterUser(req *models.UserRegistrationRequest) (*models.User, error) {
	args := m.Called(req)

	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}

	return user, args.Error(1)
}

// LoginUser is a mock implementation
func (m *AuthServiceInterface) LoginUser(req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	args := m.Called(req)

	var response *models.UserLoginResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*models.UserLoginResponse)
	}

	return response, args.Error(1)
}

// RefreshToken is a mock implementation
func (m *AuthServiceInterface) RefreshToken(req *models.TokenRefreshRequest) (*models.UserLoginResponse, error) {
	args := m.Called(req)

	var response *models.UserLoginResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*models.UserLoginResponse)
	}

	return response, args.Error(1)
}

// LogoutUser is a mock implementation
func (m *AuthServiceInterface) LogoutUser(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// ValidateAccessToken is a mock implementation
func (m *AuthServiceInterface) ValidateAccessToken(tokenString string) (*auth.JWTClaims, error) {
	args := m.Called(tokenString)

	var claims *auth.JWTClaims
	if args.Get(0) != nil {
		claims = args.Get(0).(*auth.JWTClaims)
	}

	return claims, args.Error(1)
}