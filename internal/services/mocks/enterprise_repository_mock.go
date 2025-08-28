// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
)

// MockEnterpriseRepository implements the EnterpriseRepositoryInterface for testing
type MockEnterpriseRepository struct {
	mock.Mock
}

func (m *MockEnterpriseRepository) CreateEnterprise(ctx context.Context, enterprise *models.Enterprise) error {
	args := m.Called(ctx, enterprise)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) GetEnterpriseByID(ctx context.Context, id uuid.UUID) (*models.Enterprise, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) GetEnterpriseByLegalName(ctx context.Context, legalName string) (*models.Enterprise, error) {
	args := m.Called(ctx, legalName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) GetEnterpriseByRegistrationNumber(ctx context.Context, registrationNumber string) (*models.Enterprise, error) {
	args := m.Called(ctx, registrationNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) UpdateEnterprise(ctx context.Context, enterprise *models.Enterprise) error {
	args := m.Called(ctx, enterprise)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) DeleteEnterprise(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) ListEnterprises(ctx context.Context, limit, offset int) ([]*models.Enterprise, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) SearchEnterprises(ctx context.Context, query string, limit, offset int) ([]*models.Enterprise, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) GetEnterpriseCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
