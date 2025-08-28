package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/auth"
)

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   repository.UserRepositoryInterface
	jwtService *auth.JWTService
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtService *auth.JWTService) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(req *models.UserRegistrationRequest) (*models.User, error) {
	// Check if user already exists
	exists, err := s.userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
	}

	// Hash password
	if err := user.HashPassword(req.Password); err != nil {
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser authenticates a user and returns tokens
func (s *AuthService) LoginUser(req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role, user.EnterpriseID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshTokenString, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 days
	}

	if err := s.userRepo.CreateRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	return &models.UserLoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64((15 * time.Minute).Seconds()), // Access token expires in 15 minutes
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(req *models.TokenRefreshRequest) (*models.UserLoginResponse, error) {
	// Validate refresh token
	userID, err := s.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if refresh token exists in database and is not revoked
	storedToken, err := s.userRepo.GetRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	if storedToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Revoke old refresh token
	if err := s.userRepo.RevokeRefreshToken(req.RefreshToken); err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role, user.EnterpriseID)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshTokenString, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Save new refresh token to database
	newRefreshToken := &models.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 days
	}

	if err := s.userRepo.CreateRefreshToken(newRefreshToken); err != nil {
		return nil, err
	}

	return &models.UserLoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		ExpiresIn:    int64((15 * time.Minute).Seconds()),
	}, nil
}

// LogoutUser revokes all refresh tokens for a user
func (s *AuthService) LogoutUser(userID uuid.UUID) error {
	return s.userRepo.RevokeAllUserRefreshTokens(userID)
}

// ValidateAccessToken validates an access token and returns user claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*auth.JWTClaims, error) {
	return s.jwtService.ValidateAccessToken(tokenString)
}
