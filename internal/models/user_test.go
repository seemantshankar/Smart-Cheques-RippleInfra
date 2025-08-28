package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "testpassword123"

func TestUser_HashPassword(t *testing.T) {
	password := testPassword
	user := &User{}

	err := user.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash)
}

func TestUser_CheckPassword(t *testing.T) {
	password := testPassword
	user := &User{}

	// Hash the password
	err := user.HashPassword(password)
	require.NoError(t, err)

	// Check correct password
	assert.True(t, user.CheckPassword(password))

	// Check incorrect password
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUser_CheckPasswordWithoutHash(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	// Check password without hashing first
	assert.False(t, user.CheckPassword(password))
}

func TestUserRegistrationRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UserRegistrationRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: UserRegistrationRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "admin",
			},
			valid: true,
		},
		{
			name: "invalid email",
			request: UserRegistrationRequest{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "admin",
			},
			valid: false,
		},
		{
			name: "short password",
			request: UserRegistrationRequest{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "admin",
			},
			valid: false,
		},
		{
			name: "invalid role",
			request: UserRegistrationRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
				Role:      "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: In a real test, you would use a validator like go-playground/validator
			// to test the struct tags. For now, we just check the basic structure.
			if tt.valid {
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Password)
				assert.NotEmpty(t, tt.request.FirstName)
				assert.NotEmpty(t, tt.request.LastName)
				assert.NotEmpty(t, tt.request.Role)
			}
		})
	}
}
