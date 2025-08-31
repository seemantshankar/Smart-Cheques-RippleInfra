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

func createTestBalance(enterpriseID uuid.UUID, currencyCode string, available, reserved string) *models.EnterpriseBalance {
	return &models.EnterpriseBalance{
		ID:               uuid.New(),
		EnterpriseID:     enterpriseID,
		CurrencyCode:     currencyCode,
		AvailableBalance: available,
		ReservedBalance:  reserved,
		TotalBalance:     available,
		XRPLBalance:      available,
		IsFrozen:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func TestBalanceService_InitializeEnterpriseBalance(t *testing.T) {
	tests := []struct {
		name          string
		currencyCode  string
		setupMocks    func(*MockAssetRepository, *MockBalanceRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:         "successful initialization",
			currencyCode: "USDT",
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("USDT", models.AssetTypeStablecoin, true)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, mock.AnythingOfType("uuid.UUID"), "USDT").Return(nil, errors.New("not found"))
				balanceRepo.On("CreateEnterpriseBalance", mock.Anything, mock.AnythingOfType("*models.EnterpriseBalance")).Return(nil)
			},
		},
		{
			name:         "unsupported currency",
			currencyCode: "INVALID",
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				assetRepo.On("GetAssetByCurrency", mock.Anything, "INVALID").Return(nil, errors.New("not found"))
			},
			expectError:   true,
			errorContains: "unsupported currency",
		},
		{
			name:         "deactivated currency",
			currencyCode: "INACTIVE",
			setupMocks: func(assetRepo *MockAssetRepository, _ *MockBalanceRepository) {
				asset := createTestAsset("INACTIVE", models.AssetTypeStablecoin, false)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "INACTIVE").Return(asset, nil)
			},
			expectError:   true,
			errorContains: "currently deactivated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewBalanceService(balanceRepo, assetRepo, nil)

			tt.setupMocks(assetRepo, balanceRepo)

			result, err := service.InitializeEnterpriseBalance(context.Background(), uuid.New(), tt.currencyCode)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.currencyCode, result.CurrencyCode)
				assert.Equal(t, "0", result.AvailableBalance)
			}

			assetRepo.AssertExpectations(t)
			balanceRepo.AssertExpectations(t)
		})
	}
}

func TestBalanceService_ValidateBalanceOperation(t *testing.T) {
	enterpriseID := uuid.New()

	tests := []struct {
		name        string
		request     *BalanceOperationRequest
		setupMocks  func(*MockAssetRepository, *MockBalanceRepository)
		expectError bool
		errorText   string
	}{
		{
			name: "valid deposit",
			request: &BalanceOperationRequest{
				EnterpriseID:  enterpriseID,
				CurrencyCode:  "USDT",
				Amount:        "10000",
				OperationType: models.AssetTransactionTypeDeposit,
			},
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("USDT", models.AssetTypeStablecoin, true)
				assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(nil, errors.New("not found"))
			},
		},
		{
			name: "insufficient balance for withdrawal",
			request: &BalanceOperationRequest{
				EnterpriseID:  enterpriseID,
				CurrencyCode:  "USDT",
				Amount:        "15000",
				OperationType: models.AssetTransactionTypeWithdrawal,
			},
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("USDT", models.AssetTypeStablecoin, true)
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectError: true,
			errorText:   "insufficient available balance",
		},
		{
			name: "frozen balance",
			request: &BalanceOperationRequest{
				EnterpriseID:  enterpriseID,
				CurrencyCode:  "USDT",
				Amount:        "5000",
				OperationType: models.AssetTransactionTypeWithdrawal,
			},
			setupMocks: func(assetRepo *MockAssetRepository, balanceRepo *MockBalanceRepository) {
				asset := createTestAsset("USDT", models.AssetTypeStablecoin, true)
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				balance.IsFrozen = true
				reason := "Security check"
				balance.FreezeReason = &reason
				assetRepo.On("GetAssetByCurrency", mock.Anything, "USDT").Return(asset, nil)
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectError: true,
			errorText:   "balance is frozen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewBalanceService(balanceRepo, assetRepo, nil)

			tt.setupMocks(assetRepo, balanceRepo)

			err := service.ValidateBalanceOperation(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				assert.NoError(t, err)
			}

			assetRepo.AssertExpectations(t)
			balanceRepo.AssertExpectations(t)
		})
	}
}

func TestBalanceService_CheckBalanceSufficiency(t *testing.T) {
	enterpriseID := uuid.New()

	tests := []struct {
		name             string
		requiredAmount   string
		setupMocks       func(*MockBalanceRepository)
		expectSufficient bool
		expectError      bool
	}{
		{
			name:           "sufficient balance",
			requiredAmount: "5000",
			setupMocks: func(balanceRepo *MockBalanceRepository) {
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectSufficient: true,
		},
		{
			name:           "insufficient balance",
			requiredAmount: "15000",
			setupMocks: func(balanceRepo *MockBalanceRepository) {
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectSufficient: false,
		},
		{
			name:           "balance not found",
			requiredAmount: "5000",
			setupMocks: func(balanceRepo *MockBalanceRepository) {
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(nil, errors.New("not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewBalanceService(balanceRepo, assetRepo, nil)

			tt.setupMocks(balanceRepo)

			result, err := service.CheckBalanceSufficiency(context.Background(), enterpriseID, "USDT", tt.requiredAmount)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectSufficient, result)
			}

			balanceRepo.AssertExpectations(t)
		})
	}
}

func TestBalanceService_ValidateBalanceConsistency(t *testing.T) {
	enterpriseID := uuid.New()

	tests := []struct {
		name             string
		setupMocks       func(*MockBalanceRepository)
		expectConsistent bool
		expectError      bool
	}{
		{
			name: "consistent balances",
			setupMocks: func(balanceRepo *MockBalanceRepository) {
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				balance.TotalBalance = "10000"
				balance.XRPLBalance = "10000"
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectConsistent: true,
		},
		{
			name: "inconsistent balances",
			setupMocks: func(balanceRepo *MockBalanceRepository) {
				balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
				balance.TotalBalance = "10000"
				balance.XRPLBalance = "8000"
				balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)
			},
			expectConsistent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assetRepo := &MockAssetRepository{}
			balanceRepo := &MockBalanceRepository{}
			service := NewBalanceService(balanceRepo, assetRepo, nil)

			tt.setupMocks(balanceRepo)

			report, err := service.ValidateBalanceConsistency(context.Background(), enterpriseID, "USDT")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectConsistent, report.IsConsistent)
				assert.Equal(t, enterpriseID, report.EnterpriseID)
			}

			balanceRepo.AssertExpectations(t)
		})
	}
}

// Benchmark tests
func BenchmarkBalanceService_CheckBalanceSufficiency(b *testing.B) {
	enterpriseID := uuid.New()
	assetRepo := &MockAssetRepository{}
	balanceRepo := &MockBalanceRepository{}
	service := NewBalanceService(balanceRepo, assetRepo, nil)

	balance := createTestBalance(enterpriseID, "USDT", "10000", "0")
	balanceRepo.On("GetEnterpriseBalance", mock.Anything, enterpriseID, "USDT").Return(balance, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CheckBalanceSufficiency(context.Background(), enterpriseID, "USDT", "5000")
	}
}
