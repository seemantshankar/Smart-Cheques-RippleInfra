package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/messaging"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
	"github.com/stretchr/testify/mock"
)

// AssetRepositoryInterface mock
type AssetRepositoryInterface struct {
	mock.Mock
}

func (m *AssetRepositoryInterface) GetAssetByCurrency(ctx context.Context, currency string) (*models.SupportedAsset, error) {
	args := m.Called(ctx, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportedAsset), args.Error(1)
}

func (m *AssetRepositoryInterface) CreateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *AssetRepositoryInterface) CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *AssetRepositoryInterface) UpdateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *AssetRepositoryInterface) GetAssetTransactionsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, enterpriseID, limit, offset)
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

// BalanceRepositoryInterface mock
type BalanceRepositoryInterface struct {
	mock.Mock
}

func (m *BalanceRepositoryInterface) GetBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnterpriseBalance), args.Error(1)
}

func (m *BalanceRepositoryInterface) UpdateBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *BalanceRepositoryInterface) GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).([]*models.EnterpriseBalance), args.Error(1)
}

func (m *BalanceRepositoryInterface) CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

// XRPLServiceInterface mock
type XRPLServiceInterface struct {
	mock.Mock
}

func (m *XRPLServiceInterface) CreateWallet() (*xrpl.WalletInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*xrpl.WalletInfo), args.Error(1)
}

func (m *XRPLServiceInterface) ValidateAddress(address string) bool {
	args := m.Called(address)
	return args.Bool(0)
}

func (m *XRPLServiceInterface) GetAccountInfo(address string) (interface{}, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *XRPLServiceInterface) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

// UserServiceInterface mock (referenced in withdrawal authorization service)
type UserServiceInterface struct {
	mock.Mock
}

func (m *UserServiceInterface) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserServiceInterface) GetEnterpriseUsers(ctx context.Context, enterpriseID uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *UserServiceInterface) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, userID, permission)
	return args.Bool(0), args.Error(1)
}

// EventBus mock
type EventBus struct {
	mock.Mock
}

func (m *EventBus) PublishEvent(ctx context.Context, event *messaging.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *EventBus) Subscribe(ctx context.Context, topic string, handler func(*messaging.Event) error) error {
	args := m.Called(ctx, topic, handler)
	return args.Error(0)
}

func (m *EventBus) Close() error {
	args := m.Called()
	return args.Error(0)
}
