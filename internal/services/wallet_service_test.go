package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// Mock repositories
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Create(wallet *models.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) GetByID(id uuid.UUID) (*models.Wallet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetByAddress(address string) (*models.Wallet, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetByEnterpriseID(enterpriseID uuid.UUID) ([]*models.Wallet, error) {
	args := m.Called(enterpriseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetActiveByEnterpriseAndNetwork(enterpriseID uuid.UUID, networkType string) (*models.Wallet, error) {
	args := m.Called(enterpriseID, networkType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Update(wallet *models.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) UpdateLastActivity(walletID uuid.UUID) error {
	args := m.Called(walletID)
	return args.Error(0)
}

func (m *MockWalletRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWalletRepository) GetWhitelistedWallets() ([]*models.Wallet, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Wallet), args.Error(1)
}

// Mock enterprise repository
type MockEnterpriseRepository struct {
	mock.Mock
}

func (m *MockEnterpriseRepository) GetByID(id uuid.UUID) (*models.Enterprise, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

// Mock XRPL service
type MockXRPLService struct {
	mock.Mock
}

func (m *MockXRPLService) CreateWallet() (*xrpl.WalletInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*xrpl.WalletInfo), args.Error(1)
}

func (m *MockXRPLService) ValidateAddress(address string) bool {
	args := m.Called(address)
	return args.Bool(0)
}

func TestWalletService_CreateWalletForEnterprise(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012", // 32 bytes
		},
	)
	require.NoError(t, err)

	// Test data
	enterpriseID := uuid.New()
	networkType := "testnet"

	enterprise := &models.Enterprise{
		ID:        enterpriseID,
		LegalName: "Test Enterprise",
	}

	xrplWallet := &xrpl.WalletInfo{
		Address:    "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
		PublicKey:  "03ABC123",
		PrivateKey: "private_key_123",
		Seed:       "seed_123",
	}

	// Setup expectations
	mockEnterpriseRepo.On("GetByID", enterpriseID).Return(enterprise, nil)
	mockWalletRepo.On("GetActiveByEnterpriseAndNetwork", enterpriseID, networkType).Return(nil, assert.AnError)
	mockXRPLService.On("CreateWallet").Return(xrplWallet, nil)
	mockWalletRepo.On("Create", mock.AnythingOfType("*models.Wallet")).Return(nil)

	// Execute
	result, err := service.CreateWalletForEnterprise(enterpriseID, networkType)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, enterpriseID, result.EnterpriseID)
	assert.Equal(t, "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH", result.Address)
	assert.Equal(t, "03ABC123", result.PublicKey)
	assert.Equal(t, models.WalletStatusPending, result.Status)
	assert.False(t, result.IsWhitelisted)
	assert.Equal(t, networkType, result.NetworkType)

	// Verify mocks
	mockEnterpriseRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
	mockXRPLService.AssertExpectations(t)
}

func TestWalletService_CreateWalletForEnterprise_EnterpriseNotFound(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	enterpriseID := uuid.New()
	networkType := "testnet"

	// Setup expectations
	mockEnterpriseRepo.On("GetByID", enterpriseID).Return(nil, assert.AnError)

	// Execute
	result, err := service.CreateWalletForEnterprise(enterpriseID, networkType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get enterprise")

	// Verify mocks
	mockEnterpriseRepo.AssertExpectations(t)
}

func TestWalletService_CreateWalletForEnterprise_ExistingActiveWallet(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	enterpriseID := uuid.New()
	networkType := "testnet"

	enterprise := &models.Enterprise{
		ID:        enterpriseID,
		LegalName: "Test Enterprise",
	}

	existingWallet := &models.Wallet{
		ID:           uuid.New(),
		EnterpriseID: enterpriseID,
		Status:       models.WalletStatusActive,
		NetworkType:  networkType,
	}

	// Setup expectations
	mockEnterpriseRepo.On("GetByID", enterpriseID).Return(enterprise, nil)
	mockWalletRepo.On("GetActiveByEnterpriseAndNetwork", enterpriseID, networkType).Return(existingWallet, nil)

	// Execute
	result, err := service.CreateWalletForEnterprise(enterpriseID, networkType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already has an active wallet")

	// Verify mocks
	mockEnterpriseRepo.AssertExpectations(t)
	mockWalletRepo.AssertExpectations(t)
}

func TestWalletService_ActivateWallet(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	walletID := uuid.New()
	enterpriseID := uuid.New()

	wallet := &models.Wallet{
		ID:           walletID,
		EnterpriseID: enterpriseID,
		Status:       models.WalletStatusPending,
		NetworkType:  "testnet",
	}

	// Setup expectations
	mockWalletRepo.On("GetByID", walletID).Return(wallet, nil)
	mockWalletRepo.On("GetActiveByEnterpriseAndNetwork", enterpriseID, "testnet").Return(nil, assert.AnError)
	mockWalletRepo.On("Update", mock.AnythingOfType("*models.Wallet")).Return(nil)

	// Execute
	err = service.ActivateWallet(walletID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.WalletStatusActive, wallet.Status)

	// Verify mocks
	mockWalletRepo.AssertExpectations(t)
}

func TestWalletService_WhitelistWallet(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	walletID := uuid.New()

	wallet := &models.Wallet{
		ID:            walletID,
		Status:        models.WalletStatusActive,
		IsWhitelisted: false,
	}

	// Setup expectations
	mockWalletRepo.On("GetByID", walletID).Return(wallet, nil)
	mockWalletRepo.On("Update", mock.AnythingOfType("*models.Wallet")).Return(nil)

	// Execute
	err = service.WhitelistWallet(walletID)

	// Assert
	require.NoError(t, err)
	assert.True(t, wallet.IsWhitelisted)

	// Verify mocks
	mockWalletRepo.AssertExpectations(t)
}

func TestWalletService_WhitelistWallet_InactiveWallet(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	walletID := uuid.New()

	wallet := &models.Wallet{
		ID:            walletID,
		Status:        models.WalletStatusPending,
		IsWhitelisted: false,
	}

	// Setup expectations
	mockWalletRepo.On("GetByID", walletID).Return(wallet, nil)

	// Execute
	err = service.WhitelistWallet(walletID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot whitelist inactive wallet")

	// Verify mocks
	mockWalletRepo.AssertExpectations(t)
}

func TestWalletService_SuspendWallet(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	walletID := uuid.New()
	reason := "Suspicious activity detected"

	wallet := &models.Wallet{
		ID:       walletID,
		Status:   models.WalletStatusActive,
		Metadata: make(map[string]string),
	}

	// Setup expectations
	mockWalletRepo.On("GetByID", walletID).Return(wallet, nil)
	mockWalletRepo.On("Update", mock.AnythingOfType("*models.Wallet")).Return(nil)

	// Execute
	err = service.SuspendWallet(walletID, reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.WalletStatusSuspended, wallet.Status)
	assert.Equal(t, reason, wallet.Metadata["suspension_reason"])
	assert.NotEmpty(t, wallet.Metadata["suspended_at"])

	// Verify mocks
	mockWalletRepo.AssertExpectations(t)
}

func TestWalletService_ValidateWalletAddress(t *testing.T) {
	// Setup mocks
	mockWalletRepo := &MockWalletRepository{}
	mockEnterpriseRepo := &MockEnterpriseRepository{}
	mockXRPLService := &MockXRPLService{}

	// Create service
	service, err := NewWalletService(
		mockWalletRepo,
		mockEnterpriseRepo,
		mockXRPLService,
		WalletServiceConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	)
	require.NoError(t, err)

	// Test data
	address := "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"

	// Setup expectations
	mockXRPLService.On("ValidateAddress", address).Return(true)

	// Execute
	result := service.ValidateWalletAddress(address)

	// Assert
	assert.True(t, result)

	// Verify mocks
	mockXRPLService.AssertExpectations(t)
}