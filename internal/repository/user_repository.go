package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, enterprise_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	user.ID = uuid.New()
	user.IsActive = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.EnterpriseID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, enterprise_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = true
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.EnterpriseID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, enterprise_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.EnterpriseID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// EmailExists checks if an email already exists in the database
func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CreateRefreshToken creates a new refresh token in the database
func (r *UserRepository) CreateRefreshToken(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, is_revoked)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	token.IsRevoked = false

	_, err := r.db.Exec(query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
		token.IsRevoked,
	)

	return err
}

// GetRefreshToken retrieves a refresh token by token string
func (r *UserRepository) GetRefreshToken(tokenString string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, is_revoked
		FROM refresh_tokens
		WHERE token = $1 AND is_revoked = false AND expires_at > NOW()
	`

	token := &models.RefreshToken{}
	err := r.db.QueryRow(query, tokenString).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.IsRevoked,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *UserRepository) RevokeRefreshToken(tokenString string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE token = $1`
	_, err := r.db.Exec(query, tokenString)
	return err
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func (r *UserRepository) RevokeAllUserRefreshTokens(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}
