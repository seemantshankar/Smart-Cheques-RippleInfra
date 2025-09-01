package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// Mock implementations with unique names to avoid conflicts
type mockSmartChequeRepoXRPL struct {
	mock.Mock
}

func (m *mockSmartChequeRepoXRPL) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepoXRPL) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	args := m.Called(ctx, id)
	smartCheque, _ := args.Get(0).(*models.SmartCheque)
	return smartCheque, args.Error(1)
}

func (m *mockSmartChequeRepoXRPL) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepoXRPL) DeleteSmartCheque(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSmartChequeRepoXRPL) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payerID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepoXRPL) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payeeID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepoXRPL) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepoXRPL) GetSmartChequeStatistics(ctx context.Context) (*SmartChequeStatistics, error) {
	args := m.Called(ctx)
	stats, _ := args.Get(0).(*SmartChequeStatistics)
	return stats, args.Error(1)
}

// Add missing BatchCreateSmartCheques method
func (m *mockSmartChequeRepoXRPL) BatchCreateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

// Add missing BatchDeleteSmartCheques method
func (m *mockSmartChequeRepoXRPL) BatchDeleteSmartCheques(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// Add missing GetSmartChequeAnalyticsByPayer method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeAnalyticsByPayer(ctx context.Context, payerID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payerID)
	analytics, _ := args.Get(0).(*repository.SmartChequeAnalytics)
	return analytics, args.Error(1)
}

// Add missing GetSmartChequeAnalyticsByPayee method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeAnalyticsByPayee(ctx context.Context, payeeID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payeeID)
	analytics, _ := args.Get(0).(*repository.SmartChequeAnalytics)
	return analytics, args.Error(1)
}

// Add missing GetSmartChequePerformanceMetrics method
func (m *mockSmartChequeRepoXRPL) GetSmartChequePerformanceMetrics(ctx context.Context, filters *repository.SmartChequeFilter) (*repository.SmartChequePerformanceMetrics, error) {
	args := m.Called(ctx, filters)
	metrics, _ := args.Get(0).(*repository.SmartChequePerformanceMetrics)
	return metrics, args.Error(1)
}

// Add missing GetSmartChequesByContract method
func (m *mockSmartChequeRepoXRPL) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

// Add missing GetSmartChequesByMilestone method
func (m *mockSmartChequeRepoXRPL) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	smartCheque, _ := args.Get(0).(*models.SmartCheque)
	return smartCheque, args.Error(1)
}

// Add missing GetSmartChequeCount method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Add missing GetSmartChequeCountByStatus method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.SmartChequeStatus]int64), args.Error(1)
}

// Add missing GetSmartChequeCountByCurrency method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeCountByCurrency(ctx context.Context) (map[models.Currency]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.Currency]int64), args.Error(1)
}

// Add missing GetSmartChequeAmountStatistics method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeAmountStatistics(ctx context.Context) (totalAmount, averageAmount, largestAmount, smallestAmount float64, err error) {
	args := m.Called(ctx)
	totalAmount, _ = args.Get(0).(float64)
	averageAmount, _ = args.Get(1).(float64)
	largestAmount, _ = args.Get(2).(float64)
	smallestAmount, _ = args.Get(3).(float64)
	err = args.Error(4)
	return
}

// Add missing GetSmartChequeTrends method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeTrends(ctx context.Context, days int) (map[string]int64, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(map[string]int64), args.Error(1)
}

// Add missing GetRecentSmartCheques method
func (m *mockSmartChequeRepoXRPL) GetRecentSmartCheques(ctx context.Context, limit int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

// Add missing SearchSmartCheques method
func (m *mockSmartChequeRepoXRPL) SearchSmartCheques(ctx context.Context, query string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

// Add missing BatchUpdateSmartCheques method
func (m *mockSmartChequeRepoXRPL) BatchUpdateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

// Add missing BatchUpdateSmartChequeStatus method
func (m *mockSmartChequeRepoXRPL) BatchUpdateSmartChequeStatus(ctx context.Context, ids []string, status models.SmartChequeStatus) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// Add missing BatchGetSmartCheques method
func (m *mockSmartChequeRepoXRPL) BatchGetSmartCheques(ctx context.Context, ids []string) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

// Add missing BatchUpdateSmartChequeStatuses method
func (m *mockSmartChequeRepoXRPL) BatchUpdateSmartChequeStatuses(ctx context.Context, updates map[string]models.SmartChequeStatus) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

// Add missing GetSmartChequeAuditTrail method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(ctx, smartChequeID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

// Add missing GetSmartChequeComplianceReport method
func (m *mockSmartChequeRepoXRPL) GetSmartChequeComplianceReport(ctx context.Context, smartChequeID string) (*repository.SmartChequeComplianceReport, error) {
	args := m.Called(ctx, smartChequeID)
	report, _ := args.Get(0).(*repository.SmartChequeComplianceReport)
	return report, args.Error(1)
}

type mockTransactionRepoXRPL struct {
	mock.Mock
}

func (m *mockTransactionRepoXRPL) CreateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *mockTransactionRepoXRPL) GetTransactionByID(id string) (*models.Transaction, error) {
	args := m.Called(id)
	transaction, _ := args.Get(0).(*models.Transaction)
	return transaction, args.Error(1)
}

func (m *mockTransactionRepoXRPL) UpdateTransaction(transaction *models.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *mockTransactionRepoXRPL) DeleteTransaction(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockTransactionRepoXRPL) GetTransactionsBySmartChequeID(smartChequeID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(smartChequeID, limit, offset)
	transactions, _ := args.Get(0).([]*models.Transaction)
	return transactions, args.Error(1)
}

func (m *mockTransactionRepoXRPL) GetTransactionStats() (*models.TransactionStats, error) {
	args := m.Called()
	stats, _ := args.Get(0).(*models.TransactionStats)
	return stats, args.Error(1)
}

// Add missing CreateTransactionBatch method
func (m *mockTransactionRepoXRPL) CreateTransactionBatch(batch *models.TransactionBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

// Add missing DeleteTransactionBatch method
func (m *mockTransactionRepoXRPL) DeleteTransactionBatch(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Add missing GetTransactionBatchByID method
func (m *mockTransactionRepoXRPL) GetTransactionBatchByID(id string) (*models.TransactionBatch, error) {
	args := m.Called(id)
	batch, _ := args.Get(0).(*models.TransactionBatch)
	return batch, args.Error(1)
}

// Add missing UpdateTransactionBatch method
func (m *mockTransactionRepoXRPL) UpdateTransactionBatch(batch *models.TransactionBatch) error {
	args := m.Called(batch)
	return args.Error(0)
}

// Add missing GetTransactionBatchesByStatus method
func (m *mockTransactionRepoXRPL) GetTransactionBatchesByStatus(status models.TransactionStatus, limit, offset int) ([]*models.TransactionBatch, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.TransactionBatch), args.Error(1)
}

// Add missing GetPendingBatches method
func (m *mockTransactionRepoXRPL) GetPendingBatches(limit int) ([]*models.TransactionBatch, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.TransactionBatch), args.Error(1)
}

// Add missing GetTransactionStatsByDateRange method
func (m *mockTransactionRepoXRPL) GetTransactionStatsByDateRange(start, end time.Time) (*models.TransactionStats, error) {
	args := m.Called(start, end)
	stats, _ := args.Get(0).(*models.TransactionStats)
	return stats, args.Error(1)
}

// Add missing GetTransactionCountByStatus method
func (m *mockTransactionRepoXRPL) GetTransactionCountByStatus() (map[models.TransactionStatus]int64, error) {
	args := m.Called()
	return args.Get(0).(map[models.TransactionStatus]int64), args.Error(1)
}

// Add missing GetTransactionsByStatus method
func (m *mockTransactionRepoXRPL) GetTransactionsByStatus(status models.TransactionStatus, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetTransactionsByBatchID method
func (m *mockTransactionRepoXRPL) GetTransactionsByBatchID(batchID string) ([]*models.Transaction, error) {
	args := m.Called(batchID)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetTransactionsByEnterpriseID method
func (m *mockTransactionRepoXRPL) GetTransactionsByEnterpriseID(enterpriseID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetTransactionsByUserID method
func (m *mockTransactionRepoXRPL) GetTransactionsByUserID(userID string, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetTransactionsByType method
func (m *mockTransactionRepoXRPL) GetTransactionsByType(txType models.TransactionType, limit, offset int) ([]*models.Transaction, error) {
	args := m.Called(txType, limit, offset)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetPendingTransactions method
func (m *mockTransactionRepoXRPL) GetPendingTransactions(limit int) ([]*models.Transaction, error) {
	args := m.Called(limit)
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetExpiredTransactions method
func (m *mockTransactionRepoXRPL) GetExpiredTransactions() ([]*models.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

// Add missing GetRetriableTransactions method
func (m *mockTransactionRepoXRPL) GetRetriableTransactions() ([]*models.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*models.Transaction), args.Error(1)
}

type mockXRPLServiceXRPL struct {
	mock.Mock
}

func (m *mockXRPLServiceXRPL) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockXRPLServiceXRPL) CreateWallet() (*xrpl.WalletInfo, error) {
	args := m.Called()
	wallet, _ := args.Get(0).(*xrpl.WalletInfo)
	return wallet, args.Error(1)
}

func (m *mockXRPLServiceXRPL) ValidateAddress(address string) bool {
	args := m.Called(address)
	return args.Bool(0)
}

func (m *mockXRPLServiceXRPL) GetAccountInfo(address string) (interface{}, error) {
	args := m.Called(address)
	return args.Get(0), args.Error(1)
}

func (m *mockXRPLServiceXRPL) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockXRPLServiceXRPL) CreateSmartChequeEscrow(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string) (*xrpl.TransactionResult, string, error) {
	args := m.Called(payerAddress, payeeAddress, amount, currency, milestoneSecret)
	result, _ := args.Get(0).(*xrpl.TransactionResult)
	fulfillment, _ := args.Get(1).(string)
	return result, fulfillment, args.Error(2)
}

func (m *mockXRPLServiceXRPL) CreateSmartChequeEscrowWithMilestones(payerAddress, payeeAddress string, amount float64, currency string, milestones []models.Milestone) (*xrpl.TransactionResult, string, error) {
	args := m.Called(payerAddress, payeeAddress, amount, currency, milestones)
	result, _ := args.Get(0).(*xrpl.TransactionResult)
	fulfillment, _ := args.Get(1).(string)
	return result, fulfillment, args.Error(2)
}

func (m *mockXRPLServiceXRPL) CompleteSmartChequeMilestone(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string) (*xrpl.TransactionResult, error) {
	args := m.Called(payeeAddress, ownerAddress, sequence, condition, fulfillment)
	result, _ := args.Get(0).(*xrpl.TransactionResult)
	return result, args.Error(1)
}

func (m *mockXRPLServiceXRPL) CancelSmartCheque(accountAddress, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error) {
	args := m.Called(accountAddress, ownerAddress, sequence)
	result, _ := args.Get(0).(*xrpl.TransactionResult)
	return result, args.Error(1)
}

func (m *mockXRPLServiceXRPL) GetEscrowStatus(ownerAddress string, sequence string) (*xrpl.EscrowInfo, error) {
	args := m.Called(ownerAddress, sequence)
	escrowInfo, _ := args.Get(0).(*xrpl.EscrowInfo)
	return escrowInfo, args.Error(1)
}

func (m *mockXRPLServiceXRPL) GenerateCondition(secret string) (condition string, fulfillment string, err error) {
	args := m.Called(secret)
	condition, _ = args.Get(0).(string)
	fulfillment, _ = args.Get(1).(string)
	return condition, fulfillment, args.Error(2)
}

type mockMilestoneRepoXRPL struct {
	mock.Mock
}

// Fix the CreateMilestone method to use the correct type
func (m *mockMilestoneRepoXRPL) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockMilestoneRepoXRPL) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	milestone, _ := args.Get(0).(*models.ContractMilestone)
	return milestone, args.Error(1)
}

// Fix the UpdateMilestone method to use the correct type
func (m *mockMilestoneRepoXRPL) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

// Fix the DeleteMilestone method to use the correct type
func (m *mockMilestoneRepoXRPL) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMilestoneRepoXRPL) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID, limit, offset)
	milestones, _ := args.Get(0).([]*models.ContractMilestone)
	return milestones, args.Error(1)
}

func (m *mockMilestoneRepoXRPL) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, asOfDate, limit, offset)
	milestones, _ := args.Get(0).([]*models.ContractMilestone)
	return milestones, args.Error(1)
}

// Fix the BatchCreateMilestones method to use the correct type
func (m *mockMilestoneRepoXRPL) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	args := m.Called(ctx, milestones)
	return args.Error(0)
}

// Add missing BatchDeleteMilestones method
func (m *mockMilestoneRepoXRPL) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	args := m.Called(ctx, milestoneIDs)
	return args.Error(0)
}

// Add missing GetMilestonesByStatus method
func (m *mockMilestoneRepoXRPL) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetMilestonesByPriority method
func (m *mockMilestoneRepoXRPL) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, priority, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetMilestonesByCategory method
func (m *mockMilestoneRepoXRPL) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetMilestonesByRiskLevel method
func (m *mockMilestoneRepoXRPL) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, riskLevel, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetCriticalPathMilestones method
func (m *mockMilestoneRepoXRPL) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing SearchMilestones method
func (m *mockMilestoneRepoXRPL) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetMilestoneDependencies method
func (m *mockMilestoneRepoXRPL) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

// Add missing GetMilestoneDependents method
func (m *mockMilestoneRepoXRPL) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

// Add missing ResolveDependencyGraph method
func (m *mockMilestoneRepoXRPL) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(map[string][]string), args.Error(1)
}

// Add missing ValidateDependencyGraph method
func (m *mockMilestoneRepoXRPL) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	args := m.Called(ctx, contractID)
	return args.Bool(0), args.Error(1)
}

// Add missing GetTopologicalOrder method
func (m *mockMilestoneRepoXRPL) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]string), args.Error(1)
}

// Add missing CreateMilestoneDependency method
func (m *mockMilestoneRepoXRPL) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	args := m.Called(ctx, dependency)
	return args.Error(0)
}

// Add missing DeleteMilestoneDependency method
func (m *mockMilestoneRepoXRPL) DeleteMilestoneDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Add missing BatchUpdateMilestoneStatus method
func (m *mockMilestoneRepoXRPL) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	args := m.Called(ctx, milestoneIDs, status)
	return args.Error(0)
}

// Add missing BatchUpdateMilestoneProgress method
func (m *mockMilestoneRepoXRPL) BatchUpdateMilestoneProgress(ctx context.Context, updates []repository.MilestoneProgressUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

// Add missing GetMilestoneCompletionStats method
func (m *mockMilestoneRepoXRPL) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	stats, _ := args.Get(0).(*repository.MilestoneStats)
	return stats, args.Error(1)
}

// Add missing GetMilestonePerformanceMetrics method
func (m *mockMilestoneRepoXRPL) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	args := m.Called(ctx, contractID)
	metrics, _ := args.Get(0).(*repository.MilestonePerformanceMetrics)
	return metrics, args.Error(1)
}

// Add missing GetMilestoneTimelineAnalysis method
func (m *mockMilestoneRepoXRPL) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	args := m.Called(ctx, contractID)
	analysis, _ := args.Get(0).(*repository.MilestoneTimelineAnalysis)
	return analysis, args.Error(1)
}

// Add missing GetMilestoneRiskAnalysis method
func (m *mockMilestoneRepoXRPL) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	args := m.Called(ctx, contractID)
	analysis, _ := args.Get(0).(*repository.MilestoneRiskAnalysis)
	return analysis, args.Error(1)
}

// Add missing GetMilestoneProgressTrends method
func (m *mockMilestoneRepoXRPL) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

// Add missing GetDelayedMilestonesReport method
func (m *mockMilestoneRepoXRPL) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

// Add missing CreateMilestoneProgressEntry method
func (m *mockMilestoneRepoXRPL) CreateMilestoneProgressEntry(ctx context.Context, entry *repository.MilestoneProgressEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// Add missing GetMilestoneProgressHistory method
func (m *mockMilestoneRepoXRPL) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID, limit, offset)
	return args.Get(0).([]*repository.MilestoneProgressEntry), args.Error(1)
}

// Add missing GetLatestProgressUpdate method
func (m *mockMilestoneRepoXRPL) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID)
	entry, _ := args.Get(0).(*repository.MilestoneProgressEntry)
	return entry, args.Error(1)
}

// Add missing FilterMilestones method
func (m *mockMilestoneRepoXRPL) FilterMilestones(ctx context.Context, filter *repository.MilestoneFilter) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetMilestonesByDateRange method
func (m *mockMilestoneRepoXRPL) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// Add missing GetUpcomingMilestones method
func (m *mockMilestoneRepoXRPL) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, daysAhead, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// TestSmartChequeXRPLService_CreateEscrowForSmartCheque tests the CreateEscrowForSmartCheque method
func TestSmartChequeXRPLService_CreateEscrowForSmartCheque(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	// Test data
	smartChequeID := uuid.New().String()
	payerAddress := "rPayerAddress123456789"
	payeeAddress := "rPayeeAddress123456789"

	smartCheque := &models.SmartCheque{
		ID:       smartChequeID,
		PayerID:  uuid.New().String(),
		PayeeID:  uuid.New().String(),
		Amount:   100.0,
		Currency: models.CurrencyUSDT,
		Status:   models.SmartChequeStatusCreated,
	}

	// Set up mock expectations
	mockSmartChequeRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(smartCheque, nil)
	mockXRPLService.On("ValidateAddress", payerAddress).Return(true)
	mockXRPLService.On("ValidateAddress", payeeAddress).Return(true)

	transactionResult := &xrpl.TransactionResult{
		TransactionID: uuid.New().String(),
		ResultCode:    "tesSUCCESS",
		Validated:     true,
		LedgerIndex:   12345,
	}

	mockXRPLService.On("CreateSmartChequeEscrowWithMilestones", payerAddress, payeeAddress, 100.0, "USDT", smartCheque.Milestones).Return(transactionResult, "fulfillment123", nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, mock.MatchedBy(func(sc *models.SmartCheque) bool {
		return sc.EscrowAddress == transactionResult.TransactionID && sc.Status == models.SmartChequeStatusLocked
	})).Return(nil)

	mockTransactionRepo.On("CreateTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
		return tx.Type == models.TransactionTypeEscrowCreate && tx.TransactionHash == transactionResult.TransactionID
	})).Return(nil)

	// Execute the method
	err := service.CreateEscrowForSmartCheque(context.Background(), smartChequeID, payerAddress, payeeAddress)

	// Assert results
	assert.NoError(t, err)
	mockSmartChequeRepo.AssertExpectations(t)
	mockXRPLService.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

// TestSmartChequeXRPLService_CreateEscrowForSmartCheque_SmartChequeNotFound tests the case when smart cheque is not found
func TestSmartChequeXRPLService_CreateEscrowForSmartCheque_SmartChequeNotFound(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	// Test data
	smartChequeID := uuid.New().String()
	payerAddress := "rPayerAddress123456789"
	payeeAddress := "rPayeeAddress123456789"

	// Set up mock expectations
	mockSmartChequeRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return((*models.SmartCheque)(nil), nil)

	// Execute the method
	err := service.CreateEscrowForSmartCheque(context.Background(), smartChequeID, payerAddress, payeeAddress)

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smart cheque not found")
	mockSmartChequeRepo.AssertExpectations(t)
}

// TestSmartChequeXRPLService_CompleteMilestonePayment tests the CompleteMilestonePayment method
func TestSmartChequeXRPLService_CompleteMilestonePayment(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	// Test data
	smartChequeID := uuid.New().String()
	milestoneID := uuid.New().String()

	smartCheque := &models.SmartCheque{
		ID:            smartChequeID,
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        100.0,
		Currency:      models.CurrencyUSDT,
		Status:        models.SmartChequeStatusLocked,
		EscrowAddress: "rEscrowAddress123456789",
		Milestones: []models.Milestone{
			{
				ID:     milestoneID,
				Amount: 100.0,
				Status: models.MilestoneStatusPending,
			},
		},
	}

	// Set up mock expectations
	mockSmartChequeRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(smartCheque, nil)
	mockXRPLService.On("GenerateCondition", mock.AnythingOfType("string")).Return("condition123", "fulfillment123", nil)

	transactionResult := &xrpl.TransactionResult{
		TransactionID: uuid.New().String(),
		ResultCode:    "tesSUCCESS",
		Validated:     true,
		LedgerIndex:   12345,
	}

	mockXRPLService.On("CompleteSmartChequeMilestone", "rEscrowAddress123456789", "rEscrowAddress123456789", uint32(1), "condition123", "fulfillment123").Return(transactionResult, nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, mock.MatchedBy(func(sc *models.SmartCheque) bool {
		return sc.Status == models.SmartChequeStatusCompleted
	})).Return(nil)

	mockTransactionRepo.On("CreateTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
		return tx.Type == models.TransactionTypeEscrowFinish && tx.TransactionHash == transactionResult.TransactionID
	})).Return(nil)

	// Execute the method
	err := service.CompleteMilestonePayment(context.Background(), smartChequeID, milestoneID)

	// Assert results
	assert.NoError(t, err)
	mockSmartChequeRepo.AssertExpectations(t)
	mockXRPLService.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

// TestSmartChequeXRPLService_CancelSmartChequeEscrow tests the CancelSmartChequeEscrow method
func TestSmartChequeXRPLService_CancelSmartChequeEscrow(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	// Test data
	smartChequeID := uuid.New().String()

	smartCheque := &models.SmartCheque{
		ID:            smartChequeID,
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        100.0,
		Currency:      models.CurrencyUSDT,
		Status:        models.SmartChequeStatusLocked,
		EscrowAddress: "rEscrowAddress123456789",
	}

	// Set up mock expectations
	mockSmartChequeRepo.On("GetSmartChequeByID", mock.Anything, smartChequeID).Return(smartCheque, nil)

	transactionResult := &xrpl.TransactionResult{
		TransactionID: uuid.New().String(),
		ResultCode:    "tesSUCCESS",
		Validated:     true,
		LedgerIndex:   12345,
	}

	mockXRPLService.On("CancelSmartCheque", "rEscrowAddress123456789", "rEscrowAddress123456789", uint32(1)).Return(transactionResult, nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, mock.MatchedBy(func(sc *models.SmartCheque) bool {
		return sc.Status == models.SmartChequeStatusDisputed
	})).Return(nil)

	mockTransactionRepo.On("CreateTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
		return tx.Type == models.TransactionTypeEscrowCancel && tx.TransactionHash == transactionResult.TransactionID
	})).Return(nil)

	// Execute the method
	err := service.CancelSmartChequeEscrow(context.Background(), smartChequeID)

	// Assert results
	assert.NoError(t, err)
	mockSmartChequeRepo.AssertExpectations(t)
	mockXRPLService.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

// TestSmartChequeXRPLService_GetXRPLTransactionHistory tests the GetXRPLTransactionHistory method
func TestSmartChequeXRPLService_GetXRPLTransactionHistory(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	// Test data
	smartChequeID := uuid.New().String()

	expectedTransactions := []*models.Transaction{
		{
			ID:       uuid.New().String(),
			Type:     models.TransactionTypeEscrowCreate,
			Status:   models.TransactionStatusConfirmed,
			Amount:   "100.0",
			Currency: "USDT",
		},
	}

	// Set up mock expectations
	mockTransactionRepo.On("GetTransactionsBySmartChequeID", smartChequeID, 100, 0).Return(expectedTransactions, nil)

	// Execute the method
	transactions, err := service.GetXRPLTransactionHistory(context.Background(), smartChequeID)

	// Assert results
	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, expectedTransactions[0].ID, transactions[0].ID)
	assert.Equal(t, expectedTransactions[0].Type, transactions[0].Type)
	mockTransactionRepo.AssertExpectations(t)
}

// TestSmartChequeXRPLService_FullLifecycleIntegration tests the complete Smart Cheque lifecycle with XRPL operations
func TestSmartChequeXRPLService_FullLifecycleIntegration(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	ctx := context.Background()
	smartChequeID := uuid.New().String()
	payerAddress := "rPayerAddress123456789"
	payeeAddress := "rPayeeAddress123456789"
	milestoneID1 := uuid.New().String()
	milestoneID2 := uuid.New().String()

	// Test data - Smart Cheque with multiple milestones
	smartCheque := &models.SmartCheque{
		ID:            smartChequeID,
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        1000.0,
		Currency:      models.CurrencyUSDT,
		Status:        models.SmartChequeStatusCreated,
		EscrowAddress: "",
		Milestones: []models.Milestone{
			{
				ID:                 milestoneID1,
				Description:        "Deliver goods",
				Amount:             500.0,
				VerificationMethod: models.VerificationMethodManual,
				Status:             models.MilestoneStatusPending,
			},
			{
				ID:                 milestoneID2,
				Description:        "Complete installation",
				Amount:             500.0,
				VerificationMethod: models.VerificationMethodOracle,
				OracleConfig:       &models.OracleConfig{Type: "api", Endpoint: "https://api.example.com/verify"},
				Status:             models.MilestoneStatusPending,
			},
		},
	}

	// Phase 1: Create Escrow
	t.Run("Phase 1: Create Escrow", func(t *testing.T) {
		// Set up mock expectations
		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)
		mockXRPLService.On("ValidateAddress", payerAddress).Return(true)
		mockXRPLService.On("ValidateAddress", payeeAddress).Return(true)

		escrowResult := &xrpl.TransactionResult{
			TransactionID: uuid.New().String(),
			ResultCode:    "tesSUCCESS",
			Validated:     true,
		}
		mockXRPLService.On("CreateSmartChequeEscrowWithMilestones", payerAddress, payeeAddress, 1000.0, "USDT", smartCheque.Milestones).Return(escrowResult, "fulfillment123", nil)

		mockSmartChequeRepo.On("UpdateSmartCheque", ctx, mock.MatchedBy(func(sc *models.SmartCheque) bool {
			return sc.EscrowAddress == escrowResult.TransactionID && sc.Status == models.SmartChequeStatusLocked
		})).Return(nil)

		mockTransactionRepo.On("CreateTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
			return tx.Type == models.TransactionTypeEscrowCreate && tx.TransactionHash == escrowResult.TransactionID
		})).Return(nil)

		// Execute escrow creation
		err := service.CreateEscrowForSmartCheque(ctx, smartChequeID, payerAddress, payeeAddress)

		// Assert results
		assert.NoError(t, err)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	// Phase 2: Complete First Milestone
	t.Run("Phase 2: Complete First Milestone", func(t *testing.T) {
		// Update Smart Cheque status to locked with escrow
		smartCheque.Status = models.SmartChequeStatusLocked
		smartCheque.EscrowAddress = "escrow_tx_123"

		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)
		mockXRPLService.On("GenerateCondition", mock.AnythingOfType("string")).Return("condition123", "fulfillment123", nil)

		finishResult := &xrpl.TransactionResult{
			TransactionID: uuid.New().String(),
			ResultCode:    "tesSUCCESS",
			Validated:     true,
		}
		mockXRPLService.On("CompleteSmartChequeMilestone", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(finishResult, nil)

		mockSmartChequeRepo.On("UpdateSmartCheque", ctx, mock.Anything).Return(nil)

		mockTransactionRepo.On("CreateTransaction", mock.MatchedBy(func(tx *models.Transaction) bool {
			return tx.Type == models.TransactionTypeEscrowFinish
		})).Return(nil)

		// Execute milestone completion
		err := service.CompleteMilestonePayment(ctx, smartChequeID, milestoneID1)

		// Assert results
		assert.NoError(t, err)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	// Phase 3: Complete Second Milestone (Full Completion)
	t.Run("Phase 3: Complete Second Milestone", func(t *testing.T) {
		// Update first milestone as completed
		smartCheque.Milestones[0].Status = models.MilestoneStatusVerified
		smartCheque.Status = models.SmartChequeStatusInProgress

		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)
		mockXRPLService.On("GenerateCondition", mock.AnythingOfType("string")).Return("condition456", "fulfillment456", nil)

		finishResult := &xrpl.TransactionResult{
			TransactionID: uuid.New().String(),
			ResultCode:    "tesSUCCESS",
			Validated:     true,
		}
		mockXRPLService.On("CompleteSmartChequeMilestone", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(finishResult, nil)

		mockSmartChequeRepo.On("UpdateSmartCheque", ctx, mock.Anything).Return(nil)

		mockTransactionRepo.On("CreateTransaction", mock.Anything).Return(nil)

		// Execute final milestone completion
		err := service.CompleteMilestonePayment(ctx, smartChequeID, milestoneID2)

		// Assert results
		assert.NoError(t, err)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	// Phase 4: Verify Health Status
	t.Run("Phase 4: Verify Health Status", func(t *testing.T) {
		// Update Smart Cheque to completed state
		smartCheque.Status = models.SmartChequeStatusCompleted
		smartCheque.Milestones[0].Status = models.MilestoneStatusVerified
		smartCheque.Milestones[1].Status = models.MilestoneStatusVerified

		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)

		escrowInfo := &xrpl.EscrowInfo{
			Account:     "escrow_tx_123",
			Destination: payeeAddress,
			Amount:      "1000000000", // 1000 XRP in drops
			Flags:       1,            // Finished
		}
		mockXRPLService.On("GetEscrowStatus", "escrow_tx_123", "escrow_tx_123").Return(escrowInfo, nil)

		// Execute health check
		healthStatus, err := service.GetEscrowHealthStatus(ctx, smartChequeID)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, healthStatus)
		assert.Contains(t, []string{"active", "inactive", "ready_for_release"}, healthStatus.Health)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
	})
}

// TestSmartChequeXRPLService_CancellationWorkflow tests the cancellation workflow with refund calculation
func TestSmartChequeXRPLService_CancellationWorkflow(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	ctx := context.Background()
	smartChequeID := uuid.New().String()

	// Test data - Smart Cheque with one completed milestone
	smartCheque := &models.SmartCheque{
		ID:            smartChequeID,
		PayerID:       uuid.New().String(),
		PayeeID:       uuid.New().String(),
		Amount:        1000.0,
		Currency:      models.CurrencyUSDT,
		Status:        models.SmartChequeStatusInProgress,
		EscrowAddress: "escrow_tx_123",
		Milestones: []models.Milestone{
			{
				ID:                 uuid.New().String(),
				Description:        "Partial work completed",
				Amount:             600.0,
				VerificationMethod: models.VerificationMethodManual,
				Status:             models.MilestoneStatusVerified, // Completed
			},
			{
				ID:                 uuid.New().String(),
				Description:        "Remaining work",
				Amount:             400.0,
				VerificationMethod: models.VerificationMethodManual,
				Status:             models.MilestoneStatusPending, // Not completed
			},
		},
	}

	t.Run("Cancellation with Partial Refund", func(t *testing.T) {
		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)

		cancelResult := &xrpl.TransactionResult{
			TransactionID: uuid.New().String(),
			ResultCode:    "tesSUCCESS",
			Validated:     true,
		}
		mockXRPLService.On("CancelSmartCheque", "escrow_tx_123", "escrow_tx_123", uint32(1)).Return(cancelResult, nil)

		mockSmartChequeRepo.On("UpdateSmartCheque", ctx, mock.Anything).Return(nil)

		mockTransactionRepo.On("CreateTransaction", mock.Anything).Return(nil)

		// Execute cancellation with reason
		err := service.CancelSmartChequeEscrowWithReason(ctx, smartChequeID, CancellationReasonMutualAgreement, "Mutual agreement to cancel")

		// Assert results
		assert.NoError(t, err)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("Partial Refund Validation", func(t *testing.T) {
		// Test partial refund
		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartCheque, nil)
		mockXRPLService.On("CancelSmartCheque", "escrow_tx_123", "escrow_tx_123", uint32(1)).Return(&xrpl.TransactionResult{
			TransactionID: uuid.New().String(),
			ResultCode:    "tesSUCCESS",
			Validated:     true,
		}, nil)

		mockSmartChequeRepo.On("UpdateSmartCheque", ctx, mock.Anything).Return(nil)
		mockTransactionRepo.On("CreateTransaction", mock.Anything).Return(nil)

		// Execute partial refund
		err := service.PartialRefundEscrow(ctx, smartChequeID, 60.0) // 60% refund

		// Assert results
		assert.NoError(t, err)
		mockSmartChequeRepo.AssertExpectations(t)
		mockXRPLService.AssertExpectations(t)
	})
}

// TestSmartChequeXRPLService_ErrorScenarios tests various error scenarios
func TestSmartChequeXRPLService_ErrorScenarios(t *testing.T) {
	// Create mocks
	mockSmartChequeRepo := &mockSmartChequeRepoXRPL{}
	mockTransactionRepo := &mockTransactionRepoXRPL{}
	mockXRPLService := &mockXRPLServiceXRPL{}
	mockMilestoneRepo := &mockMilestoneRepoXRPL{}

	// Create service
	service := NewSmartChequeXRPLService(
		mockSmartChequeRepo,
		mockTransactionRepo,
		mockXRPLService,
		mockMilestoneRepo,
	)

	ctx := context.Background()
	smartChequeID := uuid.New().String()

	t.Run("Cancel Completed Smart Cheque", func(t *testing.T) {
		completedSmartCheque := &models.SmartCheque{
			ID:            smartChequeID,
			Status:        models.SmartChequeStatusCompleted,
			EscrowAddress: "escrow_tx_123",
		}

		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(completedSmartCheque, nil)

		err := service.CancelSmartChequeEscrow(ctx, smartChequeID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel completed smart cheque")
		mockSmartChequeRepo.AssertExpectations(t)
	})

	t.Run("Partial Refund Without Completed Milestones", func(t *testing.T) {
		smartChequeWithoutCompleted := &models.SmartCheque{
			ID:            smartChequeID,
			Status:        models.SmartChequeStatusLocked,
			EscrowAddress: "escrow_tx_123",
			Milestones: []models.Milestone{
				{
					Status: models.MilestoneStatusPending,
				},
			},
		}

		mockSmartChequeRepo.On("GetSmartChequeByID", ctx, smartChequeID).Return(smartChequeWithoutCompleted, nil)

		err := service.PartialRefundEscrow(ctx, smartChequeID, 50.0)

		assert.Error(t, err)
		// The error could be either about no completed milestones or validation failure
		assert.True(t, strings.Contains(err.Error(), "no completed milestones") ||
			strings.Contains(err.Error(), "partial refund validation failed"))
		mockSmartChequeRepo.AssertExpectations(t)
	})
}
