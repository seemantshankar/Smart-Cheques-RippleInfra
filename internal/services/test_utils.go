package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/messaging"
	"github.com/stretchr/testify/mock"
)

// TestBalance represents a balance for testing
type TestBalance struct {
	EnterpriseID     uuid.UUID `json:"enterprise_id"`
	CurrencyCode     string    `json:"currency_code"`
	TotalBalance     string    `json:"total_balance"`
	AvailableBalance string    `json:"available_balance"`
	PendingBalance   string    `json:"pending_balance"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TestMockBalanceRepository for testing balance operations
type TestMockBalanceRepository struct {
	mock.Mock
}

func (m *TestMockBalanceRepository) GetBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnterpriseBalance), args.Error(1)
}

func (m *TestMockBalanceRepository) UpdateBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *TestMockBalanceRepository) GetBalanceHistory(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, limit int) ([]*TestBalance, error) {
	args := m.Called(ctx, enterpriseID, currencyCode, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TestBalance), args.Error(1)
}

func (m *TestMockBalanceRepository) CreateBalance(ctx context.Context, balance *TestBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

// Additional methods required by BalanceRepositoryInterface
func (m *TestMockBalanceRepository) CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *TestMockBalanceRepository) GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EnterpriseBalance), args.Error(1)
}

func (m *TestMockBalanceRepository) GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error) {
	args := m.Called(ctx, enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalance), args.Error(1)
}

func (m *TestMockBalanceRepository) UpdateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error {
	args := m.Called(ctx, balance)
	return args.Error(0)
}

func (m *TestMockBalanceRepository) GetEnterpriseBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error) {
	args := m.Called(ctx, enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalanceSummary), args.Error(1)
}

func (m *TestMockBalanceRepository) GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EnterpriseBalanceSummary), args.Error(1)
}

func (m *TestMockBalanceRepository) IsAssetInUse(ctx context.Context, currencyCode string) (bool, error) {
	args := m.Called(ctx, currencyCode)
	return args.Bool(0), args.Error(1)
}

func (m *TestMockBalanceRepository) FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error {
	args := m.Called(ctx, enterpriseID, currencyCode, reason)
	return args.Error(0)
}

func (m *TestMockBalanceRepository) UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error {
	args := m.Called(ctx, enterpriseID, currencyCode)
	return args.Error(0)
}

// TestMockAssetRepository for testing asset operations
type TestMockAssetRepository struct {
	mock.Mock
}

func (m *TestMockAssetRepository) GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error) {
	args := m.Called(ctx, currencyCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportedAsset), args.Error(1)
}

func (m *TestMockAssetRepository) CreateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *TestMockAssetRepository) UpdateAsset(ctx context.Context, asset *models.SupportedAsset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *TestMockAssetRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *TestMockAssetRepository) GetAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error) {
	args := m.Called(ctx, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SupportedAsset), args.Error(1)
}

func (m *TestMockAssetRepository) CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *TestMockAssetRepository) GetAssetTransactionsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, enterpriseID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

func (m *TestMockAssetRepository) UpdateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *TestMockAssetRepository) GetAssetByID(ctx context.Context, assetID uuid.UUID) (*models.SupportedAsset, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SupportedAsset), args.Error(1)
}

// Additional methods required by AssetRepositoryInterface
func (m *TestMockAssetRepository) GetAssetsByType(ctx context.Context, assetType models.AssetType) ([]*models.SupportedAsset, error) {
	args := m.Called(ctx, assetType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SupportedAsset), args.Error(1)
}

func (m *TestMockAssetRepository) GetAssetCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *TestMockAssetRepository) GetActiveAssetCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *TestMockAssetRepository) GetAssetTransaction(ctx context.Context, id uuid.UUID) (*models.AssetTransaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssetTransaction), args.Error(1)
}

func (m *TestMockAssetRepository) GetAssetTransactionsByCurrency(ctx context.Context, currencyCode string, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, currencyCode, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

func (m *TestMockAssetRepository) GetAssetTransactionsByType(ctx context.Context, txType models.AssetTransactionType, limit, offset int) ([]*models.AssetTransaction, error) {
	args := m.Called(ctx, txType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AssetTransaction), args.Error(1)
}

// TestMockEventBus for testing event publishing
type TestMockEventBus struct {
	mock.Mock
	publishedEvents []*messaging.Event
}

func (m *TestMockEventBus) PublishEvent(ctx context.Context, event *messaging.Event) error {
	m.publishedEvents = append(m.publishedEvents, event)
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *TestMockEventBus) SubscribeToEvent(ctx context.Context, eventType string, handler func(event *messaging.Event) error) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

func (m *TestMockEventBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *TestMockEventBus) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *TestMockEventBus) GetPublishedEvents() []*messaging.Event {
	return m.publishedEvents
}

func (m *TestMockEventBus) ClearEvents() {
	m.publishedEvents = []*messaging.Event{}
}

// Test alert severity constants
const (
	TestAlertSeverityLow      = "low"
	TestAlertSeverityMedium   = "medium"
	TestAlertSeverityHigh     = "high"
	TestAlertSeverityCritical = "critical"
)

// TestBalanceMonitoringServiceInterface for testing
type TestBalanceMonitoringServiceInterface interface {
	StartMonitoring(ctx context.Context, config *MonitoringConfig) error
	StopMonitoring(ctx context.Context) error
	GetMonitoringStatus(ctx context.Context) (*MonitoringStatus, error)
	SetBalanceThreshold(ctx context.Context, req *BalanceThresholdRequest) (*BalanceThreshold, error)
	GetBalanceThresholds(ctx context.Context, enterpriseID uuid.UUID) ([]*BalanceThreshold, error)
}

// TestAnomalyDetectionServiceInterface for testing
type TestAnomalyDetectionServiceInterface interface {
	AnalyzeTransaction(ctx context.Context, req *TransactionAnalysisRequest) (*AnomalyScore, error)
	SetAnomalyThresholds(ctx context.Context, req *AnomalyThresholdRequest) (*AnomalyThreshold, error)
	GetAnomalyThresholds(ctx context.Context, enterpriseID uuid.UUID) ([]*AnomalyThreshold, error)
}

// TestCircuitBreakerServiceInterface for testing
type TestCircuitBreakerServiceInterface interface {
	RegisterCircuitBreaker(ctx context.Context, req *CircuitBreakerRequest) (*CircuitBreaker, error)
	GetCircuitBreaker(ctx context.Context, name string) (*CircuitBreaker, error)
	ExecuteWithCircuitBreaker(ctx context.Context, name string, operation func() error) error
	TripCircuitBreaker(ctx context.Context, name string, reason string) error
	ResetCircuitBreaker(ctx context.Context, name string) error
	GetCircuitBreakerMetrics(ctx context.Context, name string) (*CircuitBreakerMetrics, error)
}


