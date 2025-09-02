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
	Initialize() error
	CreateWallet() (*xrpl.WalletInfo, error)
	ValidateAddress(address string) bool
	GetAccountInfo(address string) (interface{}, error)
	HealthCheck() error
	CreateSmartChequeEscrow(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string) (*xrpl.TransactionResult, string, error)
	CreateSmartChequeEscrowWithMilestones(payerAddress, payeeAddress string, amount float64, currency string, milestones []models.Milestone) (*xrpl.TransactionResult, string, error)
	CompleteSmartChequeMilestone(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string) (*xrpl.TransactionResult, error)
	CancelSmartCheque(accountAddress, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error)
	GetEscrowStatus(ownerAddress string, sequence string) (*xrpl.EscrowInfo, error)
	GenerateCondition(secret string) (condition string, fulfillment string, err error)
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
	GetTransactionsBySmartChequeID(smartChequeID string, limit, offset int) ([]*models.Transaction, error)
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

// Supporting types for milestone repository operations
type MilestoneProgressUpdate struct {
	MilestoneID        string  `json:"milestone_id"`
	PercentageComplete float64 `json:"percentage_complete"`
	Status             string  `json:"status,omitempty"`
	Notes              string  `json:"notes,omitempty"`
}

type MilestoneStats struct {
	TotalMilestones     int     `json:"total_milestones"`
	CompletedMilestones int     `json:"completed_milestones"`
	PendingMilestones   int     `json:"pending_milestones"`
	OverdueMilestones   int     `json:"overdue_milestones"`
	CompletionRate      float64 `json:"completion_rate"`
	AverageCompletion   float64 `json:"average_completion"`
}

type MilestonePerformanceMetrics struct {
	AverageCompletionTime time.Duration `json:"average_completion_time"`
	OnTimeCompletionRate  float64       `json:"on_time_completion_rate"`
	EarlyCompletionRate   float64       `json:"early_completion_rate"`
	DelayedCompletionRate float64       `json:"delayed_completion_rate"`
	AverageDelay          time.Duration `json:"average_delay"`
	EfficiencyScore       float64       `json:"efficiency_score"`
}

type MilestoneTimelineAnalysis struct {
	ContractID           string                   `json:"contract_id"`
	TotalDuration        time.Duration            `json:"total_duration"`
	CriticalPathDuration time.Duration            `json:"critical_path_duration"`
	SlackTime            time.Duration            `json:"slack_time"`
	Milestones           []MilestoneTimelineEntry `json:"milestones"`
}

type MilestoneTimelineEntry struct {
	MilestoneID    string        `json:"milestone_id"`
	EarliestStart  time.Time     `json:"earliest_start"`
	LatestStart    time.Time     `json:"latest_start"`
	EarliestFinish time.Time     `json:"earliest_finish"`
	LatestFinish   time.Time     `json:"latest_finish"`
	Slack          time.Duration `json:"slack"`
	IsCritical     bool          `json:"is_critical"`
}

type MilestoneRiskAnalysis struct {
	OverallRiskScore  float64              `json:"overall_risk_score"`
	HighRiskCount     int                  `json:"high_risk_count"`
	MediumRiskCount   int                  `json:"medium_risk_count"`
	LowRiskCount      int                  `json:"low_risk_count"`
	RiskDistribution  map[string]int       `json:"risk_distribution"`
	RiskMilestones    []MilestoneRiskEntry `json:"risk_milestones"`
	MitigationActions []string             `json:"mitigation_actions"`
}

type MilestoneRiskEntry struct {
	MilestoneID      string   `json:"milestone_id"`
	RiskLevel        string   `json:"risk_level"`
	RiskScore        float64  `json:"risk_score"`
	RiskFactors      []string `json:"risk_factors"`
	ContingencyPlans []string `json:"contingency_plans"`
}

type MilestoneProgressTrend struct {
	Date                time.Time `json:"date"`
	CompletionRate      float64   `json:"completion_rate"`
	MilestonesCompleted int       `json:"milestones_completed"`
	TotalMilestones     int       `json:"total_milestones"`
}

type DelayedMilestoneReport struct {
	MilestoneID     string        `json:"milestone_id"`
	ContractID      string        `json:"contract_id"`
	Description     string        `json:"description"`
	OriginalDueDate time.Time     `json:"original_due_date"`
	CurrentDueDate  *time.Time    `json:"current_due_date"`
	DelayDuration   time.Duration `json:"delay_duration"`
	DelayReason     string        `json:"delay_reason"`
	ImpactLevel     string        `json:"impact_level"`
}

type MilestoneProgressEntry struct {
	ID                 string    `json:"id" db:"id"`
	MilestoneID        string    `json:"milestone_id" db:"milestone_id"`
	PercentageComplete float64   `json:"percentage_complete" db:"percentage_complete"`
	Status             string    `json:"status" db:"status"`
	Notes              string    `json:"notes" db:"notes"`
	RecordedBy         string    `json:"recorded_by" db:"recorded_by"`
	RecordedAt         time.Time `json:"recorded_at" db:"recorded_at"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

type MilestoneFilter struct {
	ContractID      *string    `json:"contract_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Category        *string    `json:"category,omitempty"`
	Priority        *int       `json:"priority,omitempty"`
	RiskLevel       *string    `json:"risk_level,omitempty"`
	CriticalPath    *bool      `json:"critical_path,omitempty"`
	StartDateFrom   *time.Time `json:"start_date_from,omitempty"`
	StartDateTo     *time.Time `json:"start_date_to,omitempty"`
	DueDateFrom     *time.Time `json:"due_date_from,omitempty"`
	DueDateTo       *time.Time `json:"due_date_to,omitempty"`
	MinCompletion   *float64   `json:"min_completion,omitempty"`
	MaxCompletion   *float64   `json:"max_completion,omitempty"`
	SearchText      *string    `json:"search_text,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
	Dependencies    []string   `json:"dependencies,omitempty"`
	ExcludeStatuses []string   `json:"exclude_statuses,omitempty"`
}

type TemplateVersionDiff struct {
	TemplateID     string                `json:"template_id"`
	Version1       string                `json:"version1"`
	Version2       string                `json:"version2"`
	Changes        []TemplateFieldChange `json:"changes"`
	AddedFields    []string              `json:"added_fields"`
	RemovedFields  []string              `json:"removed_fields"`
	ModifiedFields []string              `json:"modified_fields"`
}

type TemplateFieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"` // added, removed, modified
}

type TemplateShare struct {
	ID          string     `json:"id" db:"id"`
	TemplateID  string     `json:"template_id" db:"template_id"`
	SharedWith  string     `json:"shared_with" db:"shared_with"`
	SharedBy    string     `json:"shared_by" db:"shared_by"`
	Permissions []string   `json:"permissions" db:"-"`
	SharedAt    time.Time  `json:"shared_at" db:"shared_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}

// SmartChequeRepositoryInterface defines the interface for smart check repository operations
type SmartChequeRepositoryInterface interface {
	// SmartCheque CRUD operations
	CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error
	GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error)
	UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error
	DeleteSmartCheque(ctx context.Context, id string) error

	// SmartCheque queries
	GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error)
	GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error)
	GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error)
	GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error)
	GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error)

	// SmartCheque statistics
	GetSmartChequeCount(ctx context.Context) (int64, error)
	GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error)

	// SmartCheque analytics and complex queries
	GetSmartChequeCountByCurrency(ctx context.Context) (map[models.Currency]int64, error)
	GetSmartChequeAmountStatistics(ctx context.Context) (totalAmount, averageAmount, largestAmount, smallestAmount float64, err error)
	GetSmartChequeTrends(ctx context.Context, days int) (map[string]int64, error)
	GetRecentSmartCheques(ctx context.Context, limit int) ([]*models.SmartCheque, error)
	SearchSmartCheques(ctx context.Context, query string, limit, offset int) ([]*models.SmartCheque, error)

	// SmartCheque batch operations
	BatchCreateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error
	BatchUpdateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error
	BatchDeleteSmartCheques(ctx context.Context, ids []string) error
	BatchUpdateSmartChequeStatus(ctx context.Context, ids []string, status models.SmartChequeStatus) error

	// Additional batch operations for performance optimization
	BatchGetSmartCheques(ctx context.Context, ids []string) ([]*models.SmartCheque, error)
	BatchUpdateSmartChequeStatuses(ctx context.Context, updates map[string]models.SmartChequeStatus) error

	// Audit trail and compliance tracking
	GetSmartChequeAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]models.AuditLog, error)
	GetSmartChequeComplianceReport(ctx context.Context, smartChequeID string) (*SmartChequeComplianceReport, error)

	// Advanced analytics and reporting
	GetSmartChequeAnalyticsByPayer(ctx context.Context, payerID string) (*SmartChequeAnalytics, error)
	GetSmartChequeAnalyticsByPayee(ctx context.Context, payeeID string) (*SmartChequeAnalytics, error)
	GetSmartChequePerformanceMetrics(ctx context.Context, filters *SmartChequeFilter) (*SmartChequePerformanceMetrics, error)
}

// SmartChequeComplianceReport represents a compliance report for a smart check
type SmartChequeComplianceReport struct {
	SmartChequeID       string                    `json:"smart_check_id"`
	TotalTransactions   int64                     `json:"total_transactions"`
	CompliantTxCount    int64                     `json:"compliant_tx_count"`
	NonCompliantTxCount int64                     `json:"non_compliant_tx_count"`
	ComplianceRate      float64                   `json:"compliance_rate"`
	LastAuditDate       time.Time                 `json:"last_audit_date"`
	AuditFindings       []SmartChequeAuditFinding `json:"audit_findings"`
	RegulatoryStatus    string                    `json:"regulatory_status"`
}

// SmartChequeAuditFinding represents a finding from an audit
type SmartChequeAuditFinding struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
}

// SmartChequeAnalytics represents detailed analytics for smart checks
type SmartChequeAnalytics struct {
	TotalCount           int64                              `json:"total_count"`
	CountByStatus        map[models.SmartChequeStatus]int64 `json:"count_by_status"`
	CountByCurrency      map[models.Currency]int64          `json:"count_by_currency"`
	AverageAmount        float64                            `json:"average_amount"`
	TotalAmount          float64                            `json:"total_amount"`
	LargestAmount        float64                            `json:"largest_amount"`
	SmallestAmount       float64                            `json:"smallest_amount"`
	RecentActivity       []*models.SmartCheque              `json:"recent_activity"`
	StatusTrends         map[string]int64                   `json:"status_trends"`
	CurrencyDistribution map[models.Currency]float64        `json:"currency_distribution"`
}

// SmartChequePerformanceMetrics represents performance metrics for smart checks
type SmartChequePerformanceMetrics struct {
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	SuccessRate           float64       `json:"success_rate"`
	FailureRate           float64       `json:"failure_rate"`
	AverageAmount         float64       `json:"average_amount"`
	TotalVolume           float64       `json:"total_volume"`
	PeakHourVolume        float64       `json:"peak_hour_volume"`
}

// SmartChequeFilter represents filter criteria for smart check queries
type SmartChequeFilter struct {
	PayerID      *string                   `json:"payer_id,omitempty"`
	PayeeID      *string                   `json:"payee_id,omitempty"`
	Status       *models.SmartChequeStatus `json:"status,omitempty"`
	Currency     *models.Currency          `json:"currency,omitempty"`
	DateFrom     *time.Time                `json:"date_from,omitempty"`
	DateTo       *time.Time                `json:"date_to,omitempty"`
	MinAmount    *float64                  `json:"min_amount,omitempty"`
	MaxAmount    *float64                  `json:"max_amount,omitempty"`
	ContractHash *string                   `json:"contract_hash,omitempty"`
	Tags         []string                  `json:"tags,omitempty"`
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

// MilestoneRepositoryInterface defines comprehensive interface for milestone repository operations
type MilestoneRepositoryInterface interface {
	// CRUD operations for milestones
	CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	DeleteMilestone(ctx context.Context, id string) error

	// Query methods
	GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error)
	GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error)
	GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error)
	GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error)
	GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error)
	GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error)
	GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error)
	SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error)

	// Dependency resolution methods
	GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error)
	GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error)
	ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error)
	ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error)
	GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error)
	CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error
	DeleteMilestoneDependency(ctx context.Context, id string) error

	// Batch operations for milestone updates
	BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error
	BatchUpdateMilestoneProgress(ctx context.Context, updates []MilestoneProgressUpdate) error
	BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error
	BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error

	// Milestone analytics and reporting queries
	GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*MilestoneStats, error)
	GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*MilestonePerformanceMetrics, error)
	GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*MilestoneTimelineAnalysis, error)
	GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*MilestoneRiskAnalysis, error)
	GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*MilestoneProgressTrend, error)
	GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*DelayedMilestoneReport, error)

	// Progress tracking and history
	CreateMilestoneProgressEntry(ctx context.Context, entry *MilestoneProgressEntry) error
	GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*MilestoneProgressEntry, error)
	GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*MilestoneProgressEntry, error)

	// Search and filtering capabilities
	FilterMilestones(ctx context.Context, filter *MilestoneFilter) ([]*models.ContractMilestone, error)
	GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error)
	GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error)
}

// MilestoneTemplateRepositoryInterface defines interface for template management
type MilestoneTemplateRepositoryInterface interface {
	// Template CRUD operations
	CreateTemplate(ctx context.Context, template *models.MilestoneTemplate) error
	GetTemplateByID(ctx context.Context, id string) (*models.MilestoneTemplate, error)
	UpdateTemplate(ctx context.Context, template *models.MilestoneTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	GetTemplates(ctx context.Context, limit, offset int) ([]*models.MilestoneTemplate, error)

	// Template instantiation and customization
	InstantiateTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (*models.ContractMilestone, error)
	CustomizeTemplate(ctx context.Context, templateID string, customizations map[string]interface{}) (*models.MilestoneTemplate, error)
	GetTemplateVariables(ctx context.Context, templateID string) ([]string, error)

	// Template versioning and change tracking
	CreateTemplateVersion(ctx context.Context, templateID string, version *models.MilestoneTemplate) error
	GetTemplateVersions(ctx context.Context, templateID string) ([]*models.MilestoneTemplate, error)
	GetTemplateVersion(ctx context.Context, templateID, version string) (*models.MilestoneTemplate, error)
	GetLatestTemplateVersion(ctx context.Context, templateID string) (*models.MilestoneTemplate, error)
	CompareTemplateVersions(ctx context.Context, templateID, version1, version2 string) (*TemplateVersionDiff, error)

	// Template sharing and permission management
	ShareTemplate(ctx context.Context, templateID, sharedWithUserID string, permissions []string) error
	RevokeTemplateAccess(ctx context.Context, templateID, userID string) error
	GetSharedTemplates(ctx context.Context, userID string, limit, offset int) ([]*models.MilestoneTemplate, error)
	GetTemplatePermissions(ctx context.Context, templateID, userID string) ([]string, error)
	GetTemplateShareList(ctx context.Context, templateID string) ([]*TemplateShare, error)
}

// RegulatoryRuleRepositoryInterface defines the interface for regulatory rule repository operations
type RegulatoryRuleRepositoryInterface interface {
	CreateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error
	GetRegulatoryRule(ctx context.Context, ruleID string) (*models.RegulatoryRule, error)
	GetActiveRegulatoryRules(ctx context.Context, jurisdiction string) ([]*models.RegulatoryRule, error)
	GetRegulatoryRulesByCategory(ctx context.Context, jurisdiction, category string) ([]*models.RegulatoryRule, error)
	UpdateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error
	DeleteRegulatoryRule(ctx context.Context, ruleID string, deletedBy string) error
	GetRegulatoryRuleStats(ctx context.Context, jurisdiction string) (map[string]interface{}, error)
	SearchRegulatoryRules(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.RegulatoryRule, error)
}

// ComplianceRepositoryInterface defines the interface for compliance repository operations
type ComplianceRepositoryInterface interface {
	CreateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error
	GetComplianceStatus(transactionID string) (*models.TransactionComplianceStatus, error)
	UpdateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error
	GetComplianceStatusesByStatus(status string, limit, offset int) ([]models.TransactionComplianceStatus, error)
	GetComplianceStatusesByEnterprise(enterpriseID string, limit, offset int) ([]models.TransactionComplianceStatus, error)
	GetFlaggedTransactions(limit, offset int) ([]models.TransactionComplianceStatus, error)
	ReviewComplianceStatus(complianceStatusID string, reviewedBy string, comments string) error
	GetComplianceStats(enterpriseID *string, since *time.Time) (*models.ComplianceStats, error)
}

// DisputeRepositoryInterface defines the interface for dispute repository operations
type DisputeRepositoryInterface interface {
	// Dispute CRUD operations
	CreateDispute(ctx context.Context, dispute *models.Dispute) error
	GetDisputeByID(ctx context.Context, id string) (*models.Dispute, error)
	UpdateDispute(ctx context.Context, dispute *models.Dispute) error
	DeleteDispute(ctx context.Context, id string) error

	// Dispute queries
	GetDisputes(ctx context.Context, filter *models.DisputeFilter, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByInitiator(ctx context.Context, initiatorID string, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByRespondent(ctx context.Context, respondentID string, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByStatus(ctx context.Context, status models.DisputeStatus, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByCategory(ctx context.Context, category models.DisputeCategory, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByPriority(ctx context.Context, priority models.DisputePriority, limit, offset int) ([]*models.Dispute, error)
	GetDisputesBySmartCheque(ctx context.Context, smartChequeID string, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByMilestone(ctx context.Context, milestoneID string, limit, offset int) ([]*models.Dispute, error)
	GetDisputesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.Dispute, error)
	GetActiveDisputes(ctx context.Context, limit, offset int) ([]*models.Dispute, error)
	GetOverdueDisputes(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.Dispute, error)
	SearchDisputes(ctx context.Context, query string, limit, offset int) ([]*models.Dispute, error)

	// Dispute statistics and analytics
	GetDisputeCount(ctx context.Context) (int64, error)
	GetDisputeCountByStatus(ctx context.Context) (map[models.DisputeStatus]int64, error)
	GetDisputeCountByCategory(ctx context.Context) (map[models.DisputeCategory]int64, error)
	GetDisputeCountByPriority(ctx context.Context) (map[models.DisputePriority]int64, error)
	GetDisputeStats(ctx context.Context) (*models.DisputeStats, error)

	// Evidence operations
	CreateEvidence(ctx context.Context, evidence *models.DisputeEvidence) error
	GetEvidenceByID(ctx context.Context, id string) (*models.DisputeEvidence, error)
	GetEvidenceByDisputeID(ctx context.Context, disputeID string) ([]*models.DisputeEvidence, error)
	UpdateEvidence(ctx context.Context, evidence *models.DisputeEvidence) error
	DeleteEvidence(ctx context.Context, id string) error

	// Resolution operations
	CreateResolution(ctx context.Context, resolution *models.DisputeResolution) error
	GetResolutionByID(ctx context.Context, id string) (*models.DisputeResolution, error)
	GetResolutionByDisputeID(ctx context.Context, disputeID string) (*models.DisputeResolution, error)
	UpdateResolution(ctx context.Context, resolution *models.DisputeResolution) error
	DeleteResolution(ctx context.Context, id string) error

	// Comment operations
	CreateComment(ctx context.Context, comment *models.DisputeComment) error
	GetCommentByID(ctx context.Context, id string) (*models.DisputeComment, error)
	GetCommentsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeComment, error)
	UpdateComment(ctx context.Context, comment *models.DisputeComment) error
	DeleteComment(ctx context.Context, id string) error

	// Audit operations
	CreateAuditLog(ctx context.Context, auditLog *models.DisputeAuditLog) error
	GetAuditLogsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeAuditLog, error)

	// Notification operations
	CreateNotification(ctx context.Context, notification *models.DisputeNotification) error
	GetNotificationByID(ctx context.Context, id string) (*models.DisputeNotification, error)
	GetNotificationsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.DisputeNotification, error)
	UpdateNotification(ctx context.Context, notification *models.DisputeNotification) error
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.DisputeNotification, error)
}

// CategorizationRuleRepositoryInterface defines methods for categorization rule data operations
type CategorizationRuleRepositoryInterface interface {
	CreateRule(ctx context.Context, rule *models.CategorizationRule) error
	GetRuleByID(ctx context.Context, id string) (*models.CategorizationRule, error)
	UpdateRule(ctx context.Context, rule *models.CategorizationRule) error
	DeleteRule(ctx context.Context, id string) error
	GetRules(ctx context.Context, filter *CategorizationRuleFilter, limit, offset int) ([]*models.CategorizationRule, error)
	BulkUpdateRuleStatus(ctx context.Context, ruleIDs []string, isActive bool, updatedBy string) error
	BulkDeleteRules(ctx context.Context, ruleIDs []string) error
	GetTopPerformingRules(ctx context.Context, category models.DisputeCategory, limit int) ([]*models.CategorizationRule, error)
	CreateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error
	GetRuleGroupByID(ctx context.Context, id string) (*models.CategorizationRuleGroup, error)
	UpdateRuleGroup(ctx context.Context, group *models.CategorizationRuleGroup) error
	DeleteRuleGroup(ctx context.Context, id string) error
	GetRuleGroups(ctx context.Context, filter *CategorizationRuleGroupFilter, limit, offset int) ([]*models.CategorizationRuleGroup, error)
	CreateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error
	UpdateRulePerformance(ctx context.Context, performance *models.CategorizationRulePerformance) error
	GetRulePerformance(ctx context.Context, ruleID string, periodStart, periodEnd time.Time) (*models.CategorizationRulePerformance, error)
	CreateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error
	GetRuleTemplateByID(ctx context.Context, id string) (*models.CategorizationRuleTemplate, error)
	UpdateRuleTemplate(ctx context.Context, template *models.CategorizationRuleTemplate) error
	DeleteRuleTemplate(ctx context.Context, id string) error
	GetRuleTemplates(ctx context.Context, filter *CategorizationRuleTemplateFilter, limit, offset int) ([]*models.CategorizationRuleTemplate, error)
	GetRuleUsageStats(ctx context.Context, startDate, endDate time.Time) ([]*RuleUsageStat, error)
}

// CategorizationMLModelRepositoryInterface defines methods for ML model data operations
type CategorizationMLModelRepositoryInterface interface {
	CreateModel(ctx context.Context, model *models.CategorizationMLModel) error
	GetModelByID(ctx context.Context, id string) (*models.CategorizationMLModel, error)
	UpdateModel(ctx context.Context, model *models.CategorizationMLModel) error
	DeleteModel(ctx context.Context, id string) error
	GetModels(ctx context.Context, filter *MLModelFilter, limit, offset int) ([]*models.CategorizationMLModel, error)
	GetLatestDeployedModel(ctx context.Context) (*models.CategorizationMLModel, error)
	GetModelsByStatus(ctx context.Context, status models.MLModelStatus, limit, offset int) ([]*models.CategorizationMLModel, error)
}

// CategorizationTrainingDataRepositoryInterface defines methods for training data operations
type CategorizationTrainingDataRepositoryInterface interface {
	CreateTrainingData(ctx context.Context, data *models.CategorizationTrainingData) error
	GetTrainingDataByID(ctx context.Context, id string) (*models.CategorizationTrainingData, error)
	UpdateTrainingData(ctx context.Context, data *models.CategorizationTrainingData) error
	DeleteTrainingData(ctx context.Context, id string) error
	GetTrainingData(ctx context.Context, filter *TrainingDataFilter, limit, offset int) ([]*models.CategorizationTrainingData, error)
	GetValidatedTrainingData(ctx context.Context, limit int) ([]*models.CategorizationTrainingData, error)
	GetTrainingDataCount(ctx context.Context) (int64, error)
	GetValidatedTrainingDataCount(ctx context.Context) (int64, error)
	GetTrainingDataCategoryDistribution(ctx context.Context) (map[string]int64, error)
	BulkCreateTrainingData(ctx context.Context, data []*models.CategorizationTrainingData) error
}

// CategorizationPredictionRepositoryInterface defines methods for prediction data operations
type CategorizationPredictionRepositoryInterface interface {
	CreatePrediction(ctx context.Context, prediction *models.CategorizationPrediction) error
	GetPredictionByID(ctx context.Context, id string) (*models.CategorizationPrediction, error)
	UpdatePrediction(ctx context.Context, prediction *models.CategorizationPrediction) error
	GetPredictionsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.CategorizationPrediction, error)
	GetPredictionsByModelID(ctx context.Context, modelID string, limit, offset int) ([]*models.CategorizationPrediction, error)
	GetUnvalidatedPredictions(ctx context.Context, limit, offset int) ([]*models.CategorizationPrediction, error)
	GetPredictionAccuracy(ctx context.Context, modelID string, startDate, endDate time.Time) (float64, error)
}

// MLModelMetricsRepositoryInterface defines methods for ML model metrics operations
type MLModelMetricsRepositoryInterface interface {
	CreateModelMetrics(ctx context.Context, metrics *models.MLModelMetrics) error
	GetModelMetrics(ctx context.Context, modelID string, startDate, endDate time.Time) (*models.MLModelMetrics, error)
	UpdateModelMetrics(ctx context.Context, metrics *models.MLModelMetrics) error
	GetModelMetricsHistory(ctx context.Context, modelID string, limit, offset int) ([]*models.MLModelMetrics, error)
	GetLatestModelMetrics(ctx context.Context, modelID string) (*models.MLModelMetrics, error)
}

// MLModelFilter represents filter criteria for ML model queries
type MLModelFilter struct {
	Status       *models.MLModelStatus `json:"status,omitempty"`
	Algorithm    *string               `json:"algorithm,omitempty"`
	CreatedBy    *string               `json:"created_by,omitempty"`
	NameContains *string               `json:"name_contains,omitempty"`
	MinAccuracy  *float64              `json:"min_accuracy,omitempty"`
}

// TrainingDataFilter represents filter criteria for training data queries
type TrainingDataFilter struct {
	Category      *models.DisputeCategory `json:"category,omitempty"`
	IsValidated   *bool                   `json:"is_validated,omitempty"`
	ValidatedBy   *string                 `json:"validated_by,omitempty"`
	CreatedAfter  *time.Time              `json:"created_after,omitempty"`
	CreatedBefore *time.Time              `json:"created_before,omitempty"`
}
