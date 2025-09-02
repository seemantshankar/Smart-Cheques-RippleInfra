package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// mockDisputeSmartChequeRepository implements the SmartChequeRepositoryInterface for testing
type mockDisputeSmartChequeRepository struct {
	mock.Mock
}

func (m *mockDisputeSmartChequeRepository) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) DeleteSmartCheque(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payerID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payeeID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.SmartChequeStatus]int64), args.Error(1)
}

// Add the missing methods for the new interface
func (m *mockDisputeSmartChequeRepository) GetSmartChequeCountByCurrency(ctx context.Context) (map[models.Currency]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.Currency]int64), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeAmountStatistics(ctx context.Context) (totalAmount, averageAmount, largestAmount, smallestAmount float64, err error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Get(3).(float64), args.Error(4)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeTrends(ctx context.Context, days int) (map[string]int64, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetRecentSmartCheques(ctx context.Context, limit int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) SearchSmartCheques(ctx context.Context, query string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) BatchCreateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) BatchUpdateSmartCheques(ctx context.Context, smartCheques []*models.SmartCheque) error {
	args := m.Called(ctx, smartCheques)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) BatchDeleteSmartCheques(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *mockDisputeSmartChequeRepository) BatchUpdateSmartChequeStatus(ctx context.Context, ids []string, status models.SmartChequeStatus) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// Additional batch operations for performance optimization
func (m *mockDisputeSmartChequeRepository) BatchGetSmartCheques(ctx context.Context, ids []string) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) BatchUpdateSmartChequeStatuses(ctx context.Context, updates map[string]models.SmartChequeStatus) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

// Audit trail and compliance tracking
func (m *mockDisputeSmartChequeRepository) GetSmartChequeAuditTrail(ctx context.Context, smartChequeID string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(ctx, smartChequeID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeComplianceReport(ctx context.Context, smartChequeID string) (*repository.SmartChequeComplianceReport, error) {
	args := m.Called(ctx, smartChequeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeComplianceReport), args.Error(1)
}

// Advanced analytics and reporting
func (m *mockDisputeSmartChequeRepository) GetSmartChequeAnalyticsByPayer(ctx context.Context, payerID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeAnalytics), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequeAnalyticsByPayee(ctx context.Context, payeeID string) (*repository.SmartChequeAnalytics, error) {
	args := m.Called(ctx, payeeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequeAnalytics), args.Error(1)
}

func (m *mockDisputeSmartChequeRepository) GetSmartChequePerformanceMetrics(ctx context.Context, filters *repository.SmartChequeFilter) (*repository.SmartChequePerformanceMetrics, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SmartChequePerformanceMetrics), args.Error(1)
}

// mockDisputeMilestoneRepository implements the MilestoneRepositoryInterface for testing
type mockDisputeMilestoneRepository struct {
	mock.Mock
}

func (m *mockDisputeMilestoneRepository) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, asOfDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, priority, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, riskLevel, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	args := m.Called(ctx, contractID)
	return args.Bool(0), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	args := m.Called(ctx, dependency)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) DeleteMilestoneDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	args := m.Called(ctx, milestoneIDs, status)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) BatchUpdateMilestoneProgress(ctx context.Context, updates []repository.MilestoneProgressUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	args := m.Called(ctx, milestones)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	args := m.Called(ctx, milestoneIDs)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	return args.Get(0).(*repository.MilestoneStats), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestonePerformanceMetrics), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneTimelineAnalysis), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneRiskAnalysis), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) CreateMilestoneProgressEntry(ctx context.Context, entry *repository.MilestoneProgressEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockDisputeMilestoneRepository) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID, limit, offset)
	return args.Get(0).([]*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).(*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) FilterMilestones(ctx context.Context, filter *repository.MilestoneFilter) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockDisputeMilestoneRepository) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, daysAhead, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// mockAuditRepository implements the AuditRepositoryInterface for testing
type mockAuditRepository struct {
	mock.Mock
}

func (m *mockAuditRepository) CreateAuditLog(auditLog *models.AuditLog) error {
	args := m.Called(auditLog)
	return args.Error(0)
}

func (m *mockAuditRepository) GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, enterpriseID, action, resource, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *mockAuditRepository) GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *mockAuditRepository) GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func TestDisputeHandlingService_InitiateMilestoneDispute(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Create test milestone
	now := time.Now()
	milestone := &models.ContractMilestone{
		ID:                   "test-milestone-1",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-1",
		SequenceNumber:       1,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-1").Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)

	// Execute the method
	err := service.InitiateMilestoneDispute(context.Background(), "test-milestone-1", "Test dispute reason")

	// Assert results
	assert.NoError(t, err)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
}

func TestDisputeHandlingService_InitiateMilestoneDispute_EmptyMilestoneID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty milestone ID
	err := service.InitiateMilestoneDispute(context.Background(), "", "Test dispute reason")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "milestone ID is required")
}

func TestDisputeHandlingService_InitiateMilestoneDispute_EmptyReason(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty reason
	err := service.InitiateMilestoneDispute(context.Background(), "test-milestone-1", "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute reason is required")
}

func TestDisputeHandlingService_InitiateMilestoneDispute_MilestoneNotFound(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "non-existent-milestone").Return((*models.ContractMilestone)(nil), fmt.Errorf("milestone not found"))

	// Execute the method
	err := service.InitiateMilestoneDispute(context.Background(), "non-existent-milestone", "Test dispute reason")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get milestone")
}

func TestDisputeHandlingService_HoldMilestoneFunds(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Create test milestone
	now := time.Now()
	milestone := &models.ContractMilestone{
		ID:                   "test-milestone-1",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-1",
		SequenceNumber:       1,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create test smart check
	smartCheque := &models.SmartCheque{
		ID:            "sc-test-milestone-1",
		PayerID:       "payer-1",
		PayeeID:       "payee-1",
		Amount:        1000.0,
		Currency:      models.CurrencyUSDT,
		EscrowAddress: "escrow-address-123",
		Status:        models.SmartChequeStatusLocked,
		ContractHash:  "test-contract-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-1").Return(milestone, nil)
	mockSmartChequeRepo.On("GetSmartChequesByMilestone", mock.Anything, "test-milestone-1").Return(smartCheque, nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, smartCheque).Return(nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)

	// Execute the method
	err := service.HoldMilestoneFunds(context.Background(), "test-milestone-1")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, models.SmartChequeStatusDisputed, smartCheque.Status)
	assert.Equal(t, "high", milestone.RiskLevel)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockSmartChequeRepo.AssertExpectations(t)
}

func TestDisputeHandlingService_HoldMilestoneFunds_EmptyMilestoneID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty milestone ID
	err := service.HoldMilestoneFunds(context.Background(), "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "milestone ID is required")
}

func TestDisputeHandlingService_HoldMilestoneFunds_MilestoneNotFound(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "non-existent-milestone").Return((*models.ContractMilestone)(nil), fmt.Errorf("milestone not found"))

	// Execute the method
	err := service.HoldMilestoneFunds(context.Background(), "non-existent-milestone")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get milestone")
}

func TestDisputeHandlingService_HoldMilestoneFunds_SmartChequeNotFound(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Create test milestone
	now := time.Now()
	milestone := &models.ContractMilestone{
		ID:                   "test-milestone-1",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-1",
		SequenceNumber:       1,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-1").Return(milestone, nil)
	mockSmartChequeRepo.On("GetSmartChequesByMilestone", mock.Anything, "test-milestone-1").Return((*models.SmartCheque)(nil), fmt.Errorf("smart check not found"))

	// Execute the method
	err := service.HoldMilestoneFunds(context.Background(), "test-milestone-1")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get smart check")
}

func TestDisputeHandlingService_ExecuteDisputeResolution(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method
	err := service.ExecuteDisputeResolution(context.Background(), "test-dispute-1")

	// Assert results
	assert.NoError(t, err)
}

func TestDisputeHandlingService_ExecuteDisputeResolution_EmptyDisputeID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty dispute ID
	err := service.ExecuteDisputeResolution(context.Background(), "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute ID is required")
}

func TestDisputeHandlingService_EnforceDisputeOutcome(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method
	err := service.EnforceDisputeOutcome(context.Background(), "test-dispute-1", "resolved_in_favor_of_payee")

	// Assert results
	assert.NoError(t, err)
}

func TestDisputeHandlingService_EnforceDisputeOutcome_EmptyDisputeID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty dispute ID
	err := service.EnforceDisputeOutcome(context.Background(), "", "resolved_in_favor_of_payee")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute ID is required")
}

func TestDisputeHandlingService_EnforceDisputeOutcome_EmptyOutcome(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockDisputeMilestoneRepository{}
	mockSmartChequeRepo := &mockDisputeSmartChequeRepository{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockAuditRepository{}

	// Create service
	service := NewDisputeHandlingService(
		mockMilestoneRepo,
		mockSmartChequeRepo,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty outcome
	err := service.EnforceDisputeOutcome(context.Background(), "test-dispute-1", "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispute outcome is required")
}
