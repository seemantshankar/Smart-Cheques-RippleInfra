// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
)

// MockAssetRepository implements the AssetRepository interface for testing
type MockAssetRepository struct {
	mock.Mock
}

func (m *MockAssetRepository) CreateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *MockAssetRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*models.SupportedAsset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportedAsset), args.Error(1)
}

func (m *MockAssetRepository) GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error) {
	args := m.Called(ctx, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportedAsset), args.Error(1)
}

func (m *MockAssetRepository) UpdateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *MockAssetRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAssetRepository) GetAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error) {
	args := m.Called(ctx, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SupportedAsset), args.Error(1)
}

func (m *MockAssetRepository) GetAssetsByType(ctx context.Context, assetType models.AssetType) ([]*models.SupportedAsset, error) {
	args := m.Called(ctx, assetType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SupportedAsset), args.Error(1)
}

func (m *MockAssetRepository) GetAssetCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAssetRepository) GetActiveAssetCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockBalanceRepository implements the BalanceRepository interface for testing
type MockBalanceRepository struct {
	mock.Mock
}

func (m *MockBalanceRepository) IsAssetInUse(ctx context.Context, currencyCode string) (bool, error) {
	args := m.Called(ctx, currencyCode)
	return args.Bool(0), args.Error(1)
}

// Add other required methods as stubs for compilation
func (m *MockBalanceRepository) CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *MockBalanceRepository) GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnterpriseBalance), args.Error(1)
}

func (m *MockBalanceRepository) GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalance), args.Error(1)
}

func (m *MockBalanceRepository) UpdateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *MockBalanceRepository) GetEnterpriseBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error) {
	args := m.Called(ctx, enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalanceSummary), args.Error(1)
}

func (m *MockBalanceRepository) GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalanceSummary), args.Error(1)
}

func (m *MockBalanceRepository) CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockBalanceRepository) GetAssetTransaction(ctx context.Context, id uuid.UUID) (*models.AssetTransaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssetTransaction), args.Error(1)
}

func (m *MockBalanceRepository) GetAssetTransactionsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, enterpriseID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

func (m *MockBalanceRepository) GetAssetTransactionsByCurrency(ctx context.Context, currencyCode string, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, currencyCode, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

func (m *MockBalanceRepository) GetAssetTransactionsByType(ctx context.Context, txType models.AssetTransactionType, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, txType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

func (m *MockBalanceRepository) UpdateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockBalanceRepository) UpdateBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, amount string, txType models.AssetTransactionType, referenceID *string) error {
	args := m.Called(ctx, enterpriseID, currencyCode, amount, txType, referenceID)
	return args.Error(0)
}

func (m *MockBalanceRepository) FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error {
	args := m.Called(ctx, enterpriseID, currencyCode, reason)
	return args.Error(0)
}

func (m *MockBalanceRepository) UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error {
	args := m.Called(ctx, enterpriseID, currencyCode)
	return args.Error(0)
}
