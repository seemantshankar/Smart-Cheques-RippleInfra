package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

// Test helper functions
func createTestAsset(currencyCode string, assetType models.AssetType, isActive bool) *models.SupportedAsset {
	return &models.SupportedAsset{
		ID:            uuid.New(),
		CurrencyCode:  currencyCode,
		CurrencyName:  currencyCode + " Test",
		AssetType:     assetType,
		DecimalPlaces: 6,
		MinimumAmount: "1000",
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestAssetService_RegisterAsset(t *testing.T) {
	tests := []struct {
		name           string
		request        *AssetRegistryRequest
		setupMocks     func(*MockAssetRepository, *MockBalanceRepository)
		expectedError  string
		validateResult func(*testing.T, *models.SupportedAsset)
	}{
		{
			name: "successful asset registration",
			request: &AssetRegistryRequest{
				CurrencyCode:  "TEST",
				CurrencyName:  "Test Currency",
				AssetType:     "stablecoin",
				DecimalPlaces: 6,
				MinimumAmount: "1000",
				Description:   stringPtr("Test description"),
			},
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				assetRepo.On("GetAssetByCurrency", mock.Anything, "TEST").Return(nil, errors.New("not found"))
				assetRepo.On("CreateAsset", mock.Anything, mock.AnythingOfType("*models.SupportedAsset")).Return(nil)
			},
			validateResult: func(t *testing.T, asset *models.SupportedAsset) {
				assert.Equal(t, "TEST", asset.CurrencyCode)
				assert.Equal(t, "Test Currency", asset.CurrencyName)
				assert.Equal(t, models.AssetTypeStablecoin, asset.AssetType)
				assert.True(t, asset.IsActive)
			},
		},
		{
			name: "asset already exists",
			request: &AssetRegistryRequest{
				CurrencyCode:  "USDT",
				CurrencyName:  "Tether USD",
				AssetType:     "stablecoin",
				DecimalPlaces: 6,
				MinimumAmount: "1000",
			},
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				existingAsset := createTestAsset("USDT", models.AssetTypeStablecoin, true)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(existingAsset, nil)
			},
			expectedError: "asset with currency code USDT already exists",
		},
		{
			name: "invalid minimum amount format",
			request: &AssetRegistryRequest{
				CurrencyCode:  "INVALID",
				CurrencyName:  "Invalid Currency",
				AssetType:     "stablecoin",
				DecimalPlaces: 6,
				MinimumAmount: "invalid_amount",
			},
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				assetRepo.On("GetAssetByCurrency", mock.Anything, "INVALID").Return(nil, errors.New("not found"))
			},
			expectedError: "invalid minimum amount format: invalid_amount",
		},
		{
			name: "minimum amount greater than maximum",
			request: &AssetRegistryRequest{
				CurrencyCode:  "TEST",
				CurrencyName:  "Test Currency",
				AssetType:     "stablecoin",
				DecimalPlaces: 6,
				MinimumAmount: "1000",
				MaximumAmount: stringPtr("500"),
			},
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				assetRepo.On("GetAssetByCurrency", mock.Anything, "TEST").Return(nil, errors.New("not found"))
			},
			expectedError: "minimum amount cannot be greater than maximum amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewAssetService(assetRepo, balanceRepo, nil)

			tt.setupMocks(assetRepo, balanceRepo)

			// Act
			result, err := service.RegisterAsset(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			assetRepo.AssertExpectations(t)
			balanceRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_GetSupportedAssets(t *testing.T) {
	tests := []struct {
		name       string
		activeOnly bool
		setupMocks func(*MockAssetRepository)
		expected   int
		expectErr  bool
	}{
		{
			name:       "get all assets",
			activeOnly: false,
			setupMocks: func(repo *MockAssetRepository) {
				assets := []*models.SupportedAsset{
					createTestAsset("USDT", models.AssetTypeStablecoin, true),
					createTestAsset("USDC", models.AssetTypeStablecoin, true),
					createTestAsset("INACTIVE", models.AssetTypeStablecoin, false),
				}
				repo.On("GetAssets", mock.Anything, false).Return(assets, nil)
			},
			expected: 3,
		},
		{
			name:       "get active assets only",
			activeOnly: true,
			setupMocks: func(repo *MockAssetRepository) {
				assets := []*models.SupportedAsset{
					createTestAsset("USDT", models.AssetTypeStablecoin, true),
					createTestAsset("USDC", models.AssetTypeStablecoin, true),
				}
				repo.On("GetAssets", mock.Anything, true).Return(assets, nil)
			},
			expected: 2,
		},
		{
			name:       "repository error",
			activeOnly: true,
			setupMocks: func(repo *MockAssetRepository) {
				repo.On("GetAssets", mock.Anything, true).Return(nil, errors.New("database error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewAssetService(assetRepo, balanceRepo, nil)

			tt.setupMocks(assetRepo)

			// Act
			result, err := service.GetSupportedAssets(context.Background(), tt.activeOnly)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expected)
			}

			assetRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_DeactivateAsset(t *testing.T) {
	tests := []struct {
		name          string
		currencyCode  string
		setupMocks    func(*MockAssetRepository, *MockBalanceRepository)
		expectedError string
	}{
		{
			name:         "successful deactivation",
			currencyCode: "TEST",
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("TEST", models.AssetTypeStablecoin, true)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "TEST").Return(asset, nil)
				balanceRepo.On("IsAssetInUse", mock.Anything, "TEST").Return(false, nil)
				assetRepo.On("UpdateAsset", mock.Anything, mock.AnythingOfType("*models.SupportedAsset")).Return(nil)
			},
		},
		{
			name:         "asset not found",
			currencyCode: "NOTFOUND",
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				assetRepo.On("GetAssetByCurrency", mock.Anything, "NOTFOUND").Return(nil, errors.New("not found"))
			},
			expectedError: "asset not found",
		},
		{
			name:         "asset in use",
			currencyCode: "INUSE",
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("INUSE", models.AssetTypeStablecoin, true)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "INUSE").Return(asset, nil)
				balanceRepo.On("IsAssetInUse", mock.Anything, "INUSE").Return(true, nil)
			},
			expectedError: "cannot deactivate asset INUSE: it is currently in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewAssetService(assetRepo, balanceRepo, nil)

			tt.setupMocks(assetRepo, balanceRepo)

			// Act
			err := service.DeactivateAsset(context.Background(), tt.currencyCode)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assetRepo.AssertExpectations(t)
			balanceRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_ValidateAssetForTransaction(t *testing.T) {
	tests := []struct {
		name          string
		currencyCode  string
		amount        string
		setupMocks    func(*MockAssetRepository)
		expectedError string
	}{
		{
			name:         "valid transaction",
			currencyCode: "USDT",
			amount:       "10000",
			setupMocks: func(repo *MockAssetRepository) {
				asset := &models.SupportedAsset{
					CurrencyCode:  "USDT",
					MinimumAmount: "1000",
					IsActive:      true,
				}
				repo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
			},
		},
		{
			name:         "unsupported currency",
			currencyCode: "INVALID",
			amount:       "10000",
			setupMocks: func(repo *MockAssetRepository) {
				repo.On("GetAssetByCurrency", mock.Anything, "INVALID").Return(nil, errors.New("not found"))
			},
			expectedError: "unsupported currency: INVALID",
		},
		{
			name:         "inactive currency",
			currencyCode: "INACTIVE",
			amount:       "10000",
			setupMocks: func(repo *MockAssetRepository) {
				asset := &models.SupportedAsset{
					CurrencyCode: "INACTIVE",
					IsActive:     false,
				}
				repo.On("GetAssetByCurrency", mock.Anything, "INACTIVE").Return(asset, nil)
			},
			expectedError: "currency INACTIVE is currently deactivated",
		},
		{
			name:         "invalid amount format",
			currencyCode: "USDT",
			amount:       "invalid",
			setupMocks: func(repo *MockAssetRepository) {
				asset := &models.SupportedAsset{
					CurrencyCode: "USDT",
					IsActive:     true,
				}
				repo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
			},
			expectedError: "invalid amount format: invalid",
		},
		{
			name:         "amount below minimum",
			currencyCode: "USDT",
			amount:       "500",
			setupMocks: func(repo *MockAssetRepository) {
				asset := &models.SupportedAsset{
					CurrencyCode:  "USDT",
					MinimumAmount: "1000",
					IsActive:      true,
				}
				repo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
			},
			expectedError: "amount 500 is below minimum 1000 for currency USDT",
		},
		{
			name:         "amount exceeds maximum",
			currencyCode: "USDT",
			amount:       "200000",
			setupMocks: func(repo *MockAssetRepository) {
				maxAmount := "100000"
				asset := &models.SupportedAsset{
					CurrencyCode:  "USDT",
					MinimumAmount: "1000",
					MaximumAmount: &maxAmount,
					IsActive:      true,
				}
				repo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
			},
			expectedError: "amount 200000 exceeds maximum 100000 for currency USDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewAssetService(assetRepo, balanceRepo, nil)

			tt.setupMocks(assetRepo)

			// Act
			err := service.ValidateAssetForTransaction(context.Background(), tt.currencyCode, tt.amount)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assetRepo.AssertExpectations(t)
		})
	}
}

func TestAssetService_InitializeDefaultAssets(t *testing.T) {
	t.Run("initialize default assets successfully", func(t *testing.T) {
		// Arrange
		assetRepo := &MockAssetRepository{}
		balanceRepo := &MockBalanceRepository{}
		service := NewAssetService(assetRepo, balanceRepo, nil)

		// Mock no existing assets
		assetRepo.On("GetAssets", mock.Anything, false).Return([]*models.SupportedAsset{}, nil)

		// Mock GetAssetByCurrency calls for duplicate check in RegisterAsset
		assetRepo.On("GetAssetByCurrency", mock.Anything, "XRP").Return(nil, errors.New("not found"))
		assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(nil, errors.New("not found"))
		assetRepo.On("GetAssetByCurrency", mock.Anything, "USDC").Return(nil, errors.New("not found"))
		assetRepo.On("GetAssetByCurrency", mock.Anything, "e₹").Return(nil, errors.New("not found"))

		// Mock successful creation of default assets
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "XRP"
		})).Return(nil)
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "USDT"
		})).Return(nil)
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "USDC"
		})).Return(nil)
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "E₹"
		})).Return(nil)

		// Act
		err := service.InitializeDefaultAssets(context.Background())

		// Assert
		assert.NoError(t, err)
		assetRepo.AssertExpectations(t)
	})

	t.Run("skip existing assets", func(t *testing.T) {
		// Arrange
		assetRepo := &MockAssetRepository{}
		balanceRepo := &MockBalanceRepository{}
		service := NewAssetService(assetRepo, balanceRepo, nil)

		// Mock existing XRP asset
		existingAssets := []*models.SupportedAsset{
			createTestAsset("XRP", models.AssetTypeNative, true),
		}
		assetRepo.On("GetAssets", mock.Anything, false).Return(existingAssets, nil)

		// Mock GetAssetByCurrency calls for duplicate check in RegisterAsset
		// XRP already exists, so no GetAssetByCurrency call for it
		assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(nil, errors.New("not found"))
		assetRepo.On("GetAssetByCurrency", mock.Anything, "USDC").Return(nil, errors.New("not found"))
		assetRepo.On("GetAssetByCurrency", mock.Anything, "e₹").Return(nil, errors.New("not found"))

		// Only expect creation of non-existing assets
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "USDT"
		})).Return(nil)
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "USDC"
		})).Return(nil)
		assetRepo.On("CreateAsset", mock.Anything, mock.MatchedBy(func(asset *models.SupportedAsset) bool {
			return asset.CurrencyCode == "E₹"
		})).Return(nil)

		// Act
		err := service.InitializeDefaultAssets(context.Background())

		// Assert
		assert.NoError(t, err)
		assetRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkAssetService_ValidateAssetForTransaction(b *testing.B) {
	assetRepo := &MockAssetRepository{}
	balanceRepo := &MockBalanceRepository{}
	service := NewAssetService(assetRepo, balanceRepo, nil)

	asset := &models.SupportedAsset{
		CurrencyCode:  "USDT",
		MinimumAmount: "1000",
		IsActive:      true,
	}
	assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.ValidateAssetForTransaction(context.Background(), "USDT", "10000")
	}
}
