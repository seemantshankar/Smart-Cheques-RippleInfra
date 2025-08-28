package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	EmailExists(email string) (bool, error)
	CreateRefreshToken(token *models.RefreshToken) error
	GetRefreshToken(tokenString string) (*models.RefreshToken, error)
	RevokeRefreshToken(tokenString string) error
	RevokeAllUserRefreshTokens(userID uuid.UUID) error
}

// EnterpriseRepositoryInterface defines the interface for enterprise repository operations
type EnterpriseRepositoryInterface interface {
	CreateEnterprise(enterprise *models.Enterprise) error
	GetEnterpriseByID(id uuid.UUID) (*models.Enterprise, error)
	GetEnterpriseByRegistrationNumber(regNumber string) (*models.Enterprise, error)
	UpdateEnterpriseKYBStatus(id uuid.UUID, status models.KYBStatus) error
	UpdateEnterpriseComplianceStatus(id uuid.UUID, status models.ComplianceStatus) error
	UpdateEnterpriseXRPLWallet(id uuid.UUID, walletAddress string) error
	RegistrationNumberExists(regNumber string) (bool, error)
	CreateDocument(doc *models.EnterpriseDocument) error
	UpdateDocumentStatus(docID uuid.UUID, status models.DocumentStatus) error
}

// WalletRepositoryInterface defines the interface for wallet repository operations
type WalletRepositoryInterface interface {
	Create(wallet *models.Wallet) error
	GetByID(id uuid.UUID) (*models.Wallet, error)
	GetByAddress(address string) (*models.Wallet, error)
	GetByEnterpriseID(enterpriseID uuid.UUID) ([]*models.Wallet, error)
	GetActiveByEnterpriseAndNetwork(enterpriseID uuid.UUID, networkType string) (*models.Wallet, error)
	Update(wallet *models.Wallet) error
	UpdateLastActivity(walletID uuid.UUID) error
	Delete(id uuid.UUID) error
	GetAllWallets() ([]*models.Wallet, error)
	GetWhitelistedWallets() ([]*models.Wallet, error)
}

// XRPLServiceInterface defines the interface for XRPL service operations
type XRPLServiceInterface interface {
	CreateWallet() (*xrpl.WalletInfo, error)
	ValidateAddress(address string) bool
	GetAccountInfo(address string) (interface{}, error)
	HealthCheck() error
}

// AuditRepositoryInterface defines the interface for audit repository operations
type AuditRepositoryInterface interface {
	CreateAuditLog(auditLog *models.AuditLog) error
	GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error)
	GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
	GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error)
}

// TransactionRepositoryInterface defines the interface for transaction repository operations
type TransactionRepositoryInterface interface {
	// Transaction CRUD operations
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionByID(id string) (*models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
	DeleteTransaction(id string) error

	// Transaction queries
	GetTransactionsByStatus(status models.TransactionStatus, limit, offset int) ([]*models.Transaction, error)
	GetTransactionsByBatchID(batchID string) ([]*models.Transaction, error)
	GetTransactionsByEnterpriseID(enterpriseID string, limit, offset int) ([]*models.Transaction, error)
	GetTransactionsByUserID(userID string, limit, offset int) ([]*models.Transaction, error)
	GetTransactionsByType(txType models.TransactionType, limit, offset int) ([]*models.Transaction, error)
	GetPendingTransactions(limit int) ([]*models.Transaction, error)
	GetExpiredTransactions() ([]*models.Transaction, error)
	GetRetriableTransactions() ([]*models.Transaction, error)

	// Batch operations
	CreateTransactionBatch(batch *models.TransactionBatch) error
	GetTransactionBatchByID(id string) (*models.TransactionBatch, error)
	UpdateTransactionBatch(batch *models.TransactionBatch) error
	DeleteTransactionBatch(id string) error
	GetTransactionBatchesByStatus(status models.TransactionStatus, limit, offset int) ([]*models.TransactionBatch, error)
	GetPendingBatches(limit int) ([]*models.TransactionBatch, error)

	// Statistics and monitoring
	GetTransactionStats() (*models.TransactionStats, error)
	GetTransactionStatsByDateRange(start, end time.Time) (*models.TransactionStats, error)
	GetTransactionCountByStatus() (map[models.TransactionStatus]int64, error)
	GetAverageProcessingTime() (float64, error)
}

// AssetRepositoryInterface defines the interface for asset repository operations
type AssetRepositoryInterface interface {
	// Asset CRUD operations
	CreateAsset(ctx context.Context, asset *models.SupportedAsset) error
	GetAssetByID(ctx context.Context, id uuid.UUID) (*models.SupportedAsset, error)
	GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error)
	UpdateAsset(ctx context.Context, asset *models.SupportedAsset) error
	DeleteAsset(ctx context.Context, id uuid.UUID) error

	// Asset queries
	GetAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error)
	GetAssetsByType(ctx context.Context, assetType models.AssetType) ([]*models.SupportedAsset, error)

	// Asset statistics
	GetAssetCount(ctx context.Context) (int64, error)
	GetActiveAssetCount(ctx context.Context) (int64, error)

	// Asset transaction operations
	CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error
	GetAssetTransaction(ctx context.Context, id uuid.UUID) (*models.AssetTransaction, error)
	GetAssetTransactionsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error)
	GetAssetTransactionsByCurrency(ctx context.Context, currencyCode string, limit, offset int) ([]*models.AssetTransaction, error)
	GetAssetTransactionsByType(ctx context.Context, txType models.AssetTransactionType, limit, offset int) ([]*models.AssetTransaction, error)
	UpdateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error
}

// AssetRepository defines the interface for asset repository operations
type AssetRepository interface {
	// Asset CRUD operations
	CreateAsset(ctx context.Context, asset *models.SupportedAsset) error
	GetAssetByID(ctx context.Context, id uuid.UUID) (*models.SupportedAsset, error)
	GetAssetByCurrency(ctx context.Context, currencyCode string) (*models.SupportedAsset, error)
	UpdateAsset(ctx context.Context, asset *models.SupportedAsset) error
	DeleteAsset(ctx context.Context, id uuid.UUID) error

	// Asset queries
	GetAssets(ctx context.Context, activeOnly bool) ([]*models.SupportedAsset, error)
	GetAssetsByType(ctx context.Context, assetType models.AssetType) ([]*models.SupportedAsset, error)

	// Asset statistics
	GetAssetCount(ctx context.Context) (int64, error)
	GetActiveAssetCount(ctx context.Context) (int64, error)
}

// BalanceRepositoryInterface defines the interface for balance repository operations
type BalanceRepositoryInterface interface {
	// Enterprise balance operations
	CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error
	GetBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error)
	GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error)
	GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error)
	UpdateBalance(ctx context.Context, balance *models.EnterpriseBalance) error
	UpdateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error

	// Balance queries
	GetEnterpriseBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error)
	GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error)
	IsAssetInUse(ctx context.Context, currencyCode string) (bool, error)

	// Balance operations
	FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error
	UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error
}

// BalanceRepository defines the interface for balance repository operations
type BalanceRepository interface {
	// Enterprise balance operations
	CreateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error
	GetEnterpriseBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) (*models.EnterpriseBalance, error)
	GetEnterpriseBalances(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalance, error)
	UpdateEnterpriseBalance(ctx context.Context, balance *models.EnterpriseBalance) error

	// Balance queries
	GetEnterpriseBalanceSummary(ctx context.Context, enterpriseID uuid.UUID) ([]*models.EnterpriseBalanceSummary, error)
	GetAllBalanceSummaries(ctx context.Context) ([]*models.EnterpriseBalanceSummary, error)
	IsAssetInUse(ctx context.Context, currencyCode string) (bool, error)

	// Asset transaction operations
	CreateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error
	GetAssetTransaction(ctx context.Context, id uuid.UUID) (*models.AssetTransaction, error)
	GetAssetTransactionsByEnterprise(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*models.AssetTransaction, error)
	GetAssetTransactionsByCurrency(ctx context.Context, currencyCode string, limit, offset int) ([]*models.AssetTransaction, error)
	GetAssetTransactionsByType(ctx context.Context, txType models.AssetTransactionType, limit, offset int) ([]*models.AssetTransaction, error)
	UpdateAssetTransaction(ctx context.Context, transaction *models.AssetTransaction) error

	// Balance operations
	UpdateBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, amount string, txType models.AssetTransactionType, referenceID *string) error
	FreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, reason string) error
	UnfreezeBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string) error
}

// ContractRepositoryInterface defines the interface for contract repository operations
type ContractRepositoryInterface interface {
	// Contract CRUD operations
	CreateContract(ctx context.Context, contract *models.Contract) error
	GetContractByID(ctx context.Context, id string) (*models.Contract, error)
	UpdateContract(ctx context.Context, contract *models.Contract) error
	DeleteContract(ctx context.Context, id string) error

	// Contract queries
	GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error)
	GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error)
	GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error)
}

// ContractMilestoneRepositoryInterface defines the interface for contract milestone repository operations
type ContractMilestoneRepositoryInterface interface {
	// Milestone CRUD operations
	CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	DeleteMilestone(ctx context.Context, id string) error

	// Milestone queries
	GetMilestonesByContractID(ctx context.Context, contractID string) ([]*models.ContractMilestone, error)
}
