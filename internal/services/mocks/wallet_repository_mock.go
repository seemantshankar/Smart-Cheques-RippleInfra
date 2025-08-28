// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
)

// MockWalletRepositoryInterface implements the WalletRepositoryInterface for testing
type MockWalletRepositoryInterface struct {
	mock.Mock
}

func (m *MockWalletRepositoryInterface) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockWalletRepositoryInterface) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryInterface) GetWalletByAddress(ctx context.Context, address string) (*models.Wallet, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryInterface) GetWalletsByEnterpriseID(ctx context.Context, enterpriseID uuid.UUID) ([]*models.Wallet, error) {
	args := m.Called(ctx, enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryInterface) UpdateWallet(ctx context.Context, wallet *models.Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockWalletRepositoryInterface) DeleteWallet(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWalletRepositoryInterface) ListWallets(ctx context.Context, limit, offset int) ([]*models.Wallet, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Wallet), args.Error(1)
}

func (m *MockWalletRepositoryInterface) GetWalletCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockEnterpriseRepositoryInterface implements another interface for testing
type MockEnterpriseRepositoryInterface struct {
	mock.Mock
}

// Add methods as needed based on the actual interface

// MockXRPLService implements the XRPLServiceInterface for testing
type MockXRPLService struct {
	mock.Mock
}

// Add methods as needed based on the actual interface
