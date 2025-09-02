package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

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

// TransactionRepositoryInterface mock
type TransactionRepositoryInterface struct {
	mock.Mock
}

// Transaction CRUD operations
func (m *TransactionRepositoryInterface) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}
func (m *TransactionRepositoryInterface) GetTransactionByID(id string) (*models.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) UpdateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}
func (m *TransactionRepositoryInterface) DeleteTransaction(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Transaction queries
func (m *TransactionRepositoryInterface) GetTransactionsByStatus(status models.TransactionStatus, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionsByBatchID(batchID string) ([]*models.Transaction, error) {
	args := m.Called(batchID)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionsByEnterpriseID(enterpriseID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionsByUserID(userID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionsByType(txType models.TransactionType, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(txType, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionsBySmartChequeID(smartChequeID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(smartChequeID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetPendingTransactions(limit int) ([]*models.Transaction, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetExpiredTransactions() ([]*models.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*models.Transaction), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetRetriableTransactions() ([]*models.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Batch operations
func (m *TransactionRepositoryInterface) CreateTransactionBatch(batch *models.TransactionBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}
func (m *TransactionRepositoryInterface) GetTransactionBatchByID(id string) (*models.TransactionBatch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionBatch), args.Error(1)
}
func (m *TransactionRepositoryInterface) UpdateTransactionBatch(batch *models.TransactionBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}
func (m *TransactionRepositoryInterface) DeleteTransactionBatch(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *TransactionRepositoryInterface) GetTransactionBatchesByStatus(status models.TransactionStatus, limit, offset int) ([]*models.TransactionBatch, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.TransactionBatch), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetPendingBatches(limit int) ([]*models.TransactionBatch, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.TransactionBatch), args.Error(1)
}

// Statistics and monitoring
func (m *TransactionRepositoryInterface) GetTransactionStats() (*models.TransactionStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionStats), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionStatsByDateRange(start, end time.Time) (*models.TransactionStats, error) {
	args := m.Called(start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionStats), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetTransactionCountByStatus() (map[models.TransactionStatus]int64, error) {
	args := m.Called()
	return args.Get(0).(map[models.TransactionStatus]int64), args.Error(1)
}
func (m *TransactionRepositoryInterface) GetAverageProcessingTime() (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

// ContractRepositoryInterface mock
type ContractRepositoryInterface struct {
	mock.Mock
}

func (m *ContractRepositoryInterface) CreateContract(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}
func (m *ContractRepositoryInterface) GetContractByID(ctx context.Context, id string) (*models.Contract, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Contract), args.Error(1)
}
func (m *ContractRepositoryInterface) UpdateContract(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}
func (m *ContractRepositoryInterface) DeleteContract(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *ContractRepositoryInterface) GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}
func (m *ContractRepositoryInterface) GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, contractType, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}
func (m *ContractRepositoryInterface) GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, party, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}

// ContractMilestoneRepositoryInterface mock
type ContractMilestoneRepositoryInterface struct {
	mock.Mock
}

func (m *ContractMilestoneRepositoryInterface) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}
func (m *ContractMilestoneRepositoryInterface) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ContractMilestone), args.Error(1)
}
func (m *ContractMilestoneRepositoryInterface) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}
func (m *ContractMilestoneRepositoryInterface) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *ContractMilestoneRepositoryInterface) GetMilestonesByContractID(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// MilestoneRepositoryInterface mock
type MilestoneRepositoryInterface struct {
	mock.Mock
}

// CRUD operations for milestones
func (m *MilestoneRepositoryInterface) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Query methods
func (m *MilestoneRepositoryInterface) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, asOfDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, priority, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, riskLevel, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Dependency resolution methods
func (m *MilestoneRepositoryInterface) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *MilestoneRepositoryInterface) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (m *MilestoneRepositoryInterface) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	args := m.Called(ctx, contractID)
	return args.Bool(0), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MilestoneRepositoryInterface) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	args := m.Called(ctx, dependency)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) DeleteMilestoneDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Batch operations for milestone updates
func (m *MilestoneRepositoryInterface) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	args := m.Called(ctx, milestoneIDs, status)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) BatchUpdateMilestoneProgress(ctx context.Context, updates []repository.MilestoneProgressUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	args := m.Called(ctx, milestones)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	args := m.Called(ctx, milestoneIDs)
	return args.Error(0)
}

// Milestone analytics and reporting queries
func (m *MilestoneRepositoryInterface) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MilestoneStats), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	args := m.Called(ctx, contractID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MilestonePerformanceMetrics), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	args := m.Called(ctx, contractID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MilestoneTimelineAnalysis), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	args := m.Called(ctx, contractID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MilestoneRiskAnalysis), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

// Progress tracking and history
func (m *MilestoneRepositoryInterface) CreateMilestoneProgressEntry(ctx context.Context, entry *repository.MilestoneProgressEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MilestoneRepositoryInterface) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID, limit, offset)
	return args.Get(0).([]*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MilestoneProgressEntry), args.Error(1)
}

// Search and filtering capabilities
func (m *MilestoneRepositoryInterface) FilterMilestones(ctx context.Context, filter *repository.MilestoneFilter) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *MilestoneRepositoryInterface) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, daysAhead, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// OracleRepositoryInterface mock
type OracleRepositoryInterface struct {
	mock.Mock
}

func (m *OracleRepositoryInterface) CreateOracleProvider(ctx context.Context, provider *models.OracleProvider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) GetOracleProviderByID(ctx context.Context, id uuid.UUID) (*models.OracleProvider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OracleProvider), args.Error(1)
}

func (m *OracleRepositoryInterface) GetOracleProviderByType(ctx context.Context, providerType models.OracleType) ([]*models.OracleProvider, error) {
	args := m.Called(ctx, providerType)
	return args.Get(0).([]*models.OracleProvider), args.Error(1)
}

func (m *OracleRepositoryInterface) UpdateOracleProvider(ctx context.Context, provider *models.OracleProvider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) DeleteOracleProvider(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) ListOracleProviders(ctx context.Context, limit, offset int) ([]*models.OracleProvider, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.OracleProvider), args.Error(1)
}

func (m *OracleRepositoryInterface) GetActiveOracleProviders(ctx context.Context) ([]*models.OracleProvider, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.OracleProvider), args.Error(1)
}

func (m *OracleRepositoryInterface) HealthCheckOracleProvider(ctx context.Context, id uuid.UUID) (*models.OracleStatus, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OracleStatus), args.Error(1)
}

func (m *OracleRepositoryInterface) CreateOracleRequest(ctx context.Context, request *models.OracleRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) GetOracleRequestByID(ctx context.Context, id uuid.UUID) (*models.OracleRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OracleRequest), args.Error(1)
}

func (m *OracleRepositoryInterface) UpdateOracleRequest(ctx context.Context, request *models.OracleRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) DeleteOracleRequest(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) ListOracleRequests(ctx context.Context, filter *repository.OracleRequestFilter, limit, offset int) ([]*models.OracleRequest, error) {
	args := m.Called(ctx, filter, limit, offset)
	return args.Get(0).([]*models.OracleRequest), args.Error(1)
}

func (m *OracleRepositoryInterface) GetOracleRequestsByStatus(ctx context.Context, status models.RequestStatus, limit, offset int) ([]*models.OracleRequest, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.OracleRequest), args.Error(1)
}

func (m *OracleRepositoryInterface) GetOracleRequestsByProvider(ctx context.Context, providerID uuid.UUID, limit, offset int) ([]*models.OracleRequest, error) {
	args := m.Called(ctx, providerID, limit, offset)
	return args.Get(0).([]*models.OracleRequest), args.Error(1)
}

func (m *OracleRepositoryInterface) GetCachedResponse(ctx context.Context, condition string) (*models.OracleRequest, error) {
	args := m.Called(ctx, condition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OracleRequest), args.Error(1)
}

func (m *OracleRepositoryInterface) CacheResponse(ctx context.Context, request *models.OracleRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) ClearExpiredCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *OracleRepositoryInterface) GetOracleMetrics(ctx context.Context, providerID uuid.UUID) (*models.OracleMetrics, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OracleMetrics), args.Error(1)
}

func (m *OracleRepositoryInterface) GetOracleReliabilityScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	args := m.Called(ctx, providerID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *OracleRepositoryInterface) GetRequestStats(ctx context.Context, providerID *uuid.UUID, startDate, endDate *time.Time) (map[models.RequestStatus]int64, error) {
	args := m.Called(ctx, providerID, startDate, endDate)
	return args.Get(0).(map[models.RequestStatus]int64), args.Error(1)
}

// SmartChequeRepositoryInterface mock
type SmartChequeRepositoryInterface struct {
	mock.Mock
}

func (m *SmartChequeRepositoryInterface) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) DeleteSmartCheque(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payerID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payeeID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.SmartChequeStatus]int64), args.Error(1)
}

// Add the missing methods for the new interface
func (m *SmartChequeRepositoryInterface) GetSmartChequeCountByCurrency(ctx context.Context) (map[models.Currency]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.Currency]int64), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeAmountStatistics(ctx context.Context) (totalAmount, averageAmount, largestAmount, smallestAmount float64, err error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Get(3).(float64), args.Error(4)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeTrends(ctx context.Context, days int) (map[string]int64, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetRecentSmartCheques(ctx context.Context, limit int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) SearchSmartCheques(ctx context.Context, query string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) BatchCreateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) BatchUpdateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) BatchDeleteSmartCheques(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *SmartChequeRepositoryInterface) BatchUpdateSmartChequeStatus(ctx context.Context, ids []string, status models.SmartChequeStatus) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// Additional batch operations for performance optimization
func (m *SmartChequeRepositoryInterface) BatchGetSmartCheques(ctx context.Context, ids []string) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) BatchUpdateSmartChequeStatuses(ctx context.Context, updates map[string]models.SmartChequeStatus) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

// Audit trail and compliance tracking
func (m *SmartChequeRepositoryInterface) GetSmartChequeAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(ctx, smartChequeID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeComplianceReport(ctx context.Context, smartChequeID string) (*repository.SmartChequeComplianceReport, error) {
	args := m.Called(ctx, smartChequeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeComplianceReport), args.Error(1)
}

// Advanced analytics and reporting
func (m *SmartChequeRepositoryInterface) GetSmartChequeAnalyticsByPayer(ctx context.Context, payerID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeAnalytics), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequeAnalyticsByPayee(ctx context.Context, payeeID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payeeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeAnalytics), args.Error(1)
}

func (m *SmartChequeRepositoryInterface) GetSmartChequePerformanceMetrics(ctx context.Context, filters *repository.SmartChequeFilter) (*repository.SmartChequePerformanceMetrics, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequePerformanceMetrics), args.Error(1)
}

// AuditRepositoryInterface mock
type AuditRepositoryInterface struct {
	mock.Mock
}

func (m *AuditRepositoryInterface) CreateAuditLog(auditLog *models.AuditLog) error {
	args := m.Called(auditLog)
	return args.Error(0)
}

func (m *AuditRepositoryInterface) GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, enterpriseID, action, resource, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *AuditRepositoryInterface) GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *AuditRepositoryInterface) GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

// RegulatoryRuleRepositoryInterface mock
type RegulatoryRuleRepositoryInterface struct {
	mock.Mock
}

func (m *RegulatoryRuleRepositoryInterface) CreateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *RegulatoryRuleRepositoryInterface) GetRegulatoryRule(ctx context.Context, ruleID string) (*models.RegulatoryRule, error) {
	args := m.Called(ctx, ruleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RegulatoryRule), args.Error(1)
}

func (m *RegulatoryRuleRepositoryInterface) GetActiveRegulatoryRules(ctx context.Context, jurisdiction string) ([]*models.RegulatoryRule, error) {
	args := m.Called(ctx, jurisdiction)
	return args.Get(0).([]*models.RegulatoryRule), args.Error(1)
}

func (m *RegulatoryRuleRepositoryInterface) GetRegulatoryRulesByCategory(ctx context.Context, jurisdiction, category string) ([]*models.RegulatoryRule, error) {
	args := m.Called(ctx, jurisdiction, category)
	return args.Get(0).([]*models.RegulatoryRule), args.Error(1)
}

func (m *RegulatoryRuleRepositoryInterface) UpdateRegulatoryRule(ctx context.Context, rule *models.RegulatoryRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *RegulatoryRuleRepositoryInterface) DeleteRegulatoryRule(ctx context.Context, ruleID string, deletedBy string) error {
	args := m.Called(ctx, ruleID, deletedBy)
	return args.Error(0)
}

func (m *RegulatoryRuleRepositoryInterface) GetRegulatoryRuleStats(ctx context.Context, jurisdiction string) (map[string]interface{}, error) {
	args := m.Called(ctx, jurisdiction)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *RegulatoryRuleRepositoryInterface) SearchRegulatoryRules(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.RegulatoryRule, error) {
	args := m.Called(ctx, filters, limit, offset)
	return args.Get(0).([]*models.RegulatoryRule), args.Error(1)
}

// ComplianceRepositoryInterface mock
type ComplianceRepositoryInterface struct {
	mock.Mock
}

func (m *ComplianceRepositoryInterface) CreateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	args := m.Called(complianceStatus)
	return args.Error(0)
}

func (m *ComplianceRepositoryInterface) GetComplianceStatus(transactionID string) (*models.TransactionComplianceStatus, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionComplianceStatus), args.Error(1)
}

func (m *ComplianceRepositoryInterface) UpdateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	args := m.Called(complianceStatus)
	return args.Error(0)
}

func (m *ComplianceRepositoryInterface) GetComplianceStatusesByStatus(status string, limit, offset int) ([]models.TransactionComplianceStatus, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]models.TransactionComplianceStatus), args.Error(1)
}

func (m *ComplianceRepositoryInterface) GetComplianceStatusesByEnterprise(enterpriseID string, limit, offset int) ([]models.TransactionComplianceStatus, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]models.TransactionComplianceStatus), args.Error(1)
}

func (m *ComplianceRepositoryInterface) GetFlaggedTransactions(limit, offset int) ([]models.TransactionComplianceStatus, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.TransactionComplianceStatus), args.Error(1)
}

func (m *ComplianceRepositoryInterface) ReviewComplianceStatus(complianceStatusID string, reviewedBy string, comments string) error {
	args := m.Called(complianceStatusID, reviewedBy, comments)
	return args.Error(0)
}

func (m *ComplianceRepositoryInterface) GetComplianceStats(enterpriseID *string, since *time.Time) (*models.ComplianceStats, error) {
	args := m.Called(enterpriseID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ComplianceStats), args.Error(1)
}
