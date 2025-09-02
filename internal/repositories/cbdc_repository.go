package repositories

import (
	"context"

	"github.com/smart-payment-infrastructure/internal/models"
)

// CBDCWalletRepositoryInterface defines the interface for CBDC wallet repository operations
type CBDCWalletRepositoryInterface interface {
	Create(ctx context.Context, wallet *models.CBDCWallet) error
	GetByID(ctx context.Context, id string) (*models.CBDCWallet, error)
	GetByEnterprise(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error)
	Update(ctx context.Context, wallet *models.CBDCWallet) error
	Delete(ctx context.Context, id string) error
	ListByEnterprise(ctx context.Context, enterpriseID string) ([]*models.CBDCWallet, error)
	ListByStatus(ctx context.Context, status models.CBDCWalletStatus) ([]*models.CBDCWallet, error)
}

// CBDCTransactionRepositoryInterface defines the interface for CBDC transaction repository operations
type CBDCTransactionRepositoryInterface interface {
	Create(ctx context.Context, transaction *models.CBDCTransaction) error
	GetByID(ctx context.Context, id string) (*models.CBDCTransaction, error)
	GetByWalletID(ctx context.Context, walletID string, limit, offset int) ([]*models.CBDCTransaction, error)
	Update(ctx context.Context, transaction *models.CBDCTransaction) error
	Delete(ctx context.Context, id string) error
	ListByStatus(ctx context.Context, status models.CBDCTransactionStatus) ([]*models.CBDCTransaction, error)
	ListByType(ctx context.Context, txType models.CBDCTransactionType) ([]*models.CBDCTransaction, error)
}

// CBDCBalanceRepositoryInterface defines the interface for CBDC balance repository operations
type CBDCBalanceRepositoryInterface interface {
	Create(ctx context.Context, balance *models.CBDCBalance) error
	GetByWalletID(ctx context.Context, walletID string) (*models.CBDCBalance, error)
	Update(ctx context.Context, balance *models.CBDCBalance) error
	Delete(ctx context.Context, id string) error
}

// CBDCWalletRequestRepositoryInterface defines the interface for CBDC wallet request repository operations
type CBDCWalletRequestRepositoryInterface interface {
	Create(ctx context.Context, request *models.CBDCWalletRequest) error
	GetByID(ctx context.Context, id string) (*models.CBDCWalletRequest, error)
	Update(ctx context.Context, request *models.CBDCWalletRequest) error
	Delete(ctx context.Context, id string) error
	GetPendingByEnterprise(ctx context.Context, enterpriseID string) ([]*models.CBDCWalletRequest, error)
	ListByStatus(ctx context.Context, status models.CBDCRequestStatus) ([]*models.CBDCWalletRequest, error)
}

// TSPConfigRepositoryInterface defines the interface for TSP configuration repository operations
type TSPConfigRepositoryInterface interface {
	Create(ctx context.Context, config *models.TSPConfig) error
	GetByID(ctx context.Context, id string) (*models.TSPConfig, error)
	Update(ctx context.Context, config *models.TSPConfig) error
	Delete(ctx context.Context, id string) error
	ListByStatus(ctx context.Context, status models.TSPStatus) ([]*models.TSPConfig, error)
	ListByEnvironment(ctx context.Context, environment string) ([]*models.TSPConfig, error)
}
