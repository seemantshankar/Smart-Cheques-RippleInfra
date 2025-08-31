package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// mockVerificationMilestoneRepository implements the MilestoneRepositoryInterface for testing
type mockVerificationMilestoneRepository struct {
	mock.Mock
}

func (m *mockVerificationMilestoneRepository) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, asOfDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, priority, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, riskLevel, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	args := m.Called(ctx, contractID)
	return args.Bool(0), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	args := m.Called(ctx, dependency)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) DeleteMilestoneDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	args := m.Called(ctx, milestoneIDs, status)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) BatchUpdateMilestoneProgress(ctx context.Context, updates []repository.MilestoneProgressUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	args := m.Called(ctx, milestones)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	args := m.Called(ctx, milestoneIDs)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	return args.Get(0).(*repository.MilestoneStats), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestonePerformanceMetrics), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneTimelineAnalysis), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneRiskAnalysis), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) CreateMilestoneProgressEntry(ctx context.Context, entry *repository.MilestoneProgressEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockVerificationMilestoneRepository) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID, limit, offset)
	return args.Get(0).([]*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).(*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) FilterMilestones(ctx context.Context, filter *repository.MilestoneFilter) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockVerificationMilestoneRepository) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, daysAhead, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// mockVerificationAuditRepository implements the AuditRepositoryInterface for testing
type mockVerificationAuditRepository struct {
	mock.Mock
}

func (m *mockVerificationAuditRepository) CreateAuditLog(auditLog *models.AuditLog) error {
	args := m.Called(auditLog)
	return args.Error(0)
}

func (m *mockVerificationAuditRepository) GetAuditLogs(userID *uuid.UUID, enterpriseID *uuid.UUID, action, resource string, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, enterpriseID, action, resource, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *mockVerificationAuditRepository) GetAuditLogsByUser(userID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *mockVerificationAuditRepository) GetAuditLogsByEnterprise(enterpriseID uuid.UUID, limit, offset int) ([]models.AuditLog, error) {
	args := m.Called(enterpriseID, limit, offset)
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func TestVerificationWorkflowService_GenerateVerificationRequest(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
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

	// Execute the method
	request, err := service.GenerateVerificationRequest(context.Background(), milestone)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, request)
	assert.NotEmpty(t, request.ID)
	assert.Equal(t, "test-milestone-1", request.MilestoneID)
	assert.Equal(t, "pending", request.Status)
	assert.Equal(t, "system", request.Requester)
	assert.NotNil(t, request.CreatedAt)
	assert.NotNil(t, request.UpdatedAt)
}

func TestVerificationWorkflowService_GenerateVerificationRequest_NilMilestone(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with nil milestone
	request, err := service.GenerateVerificationRequest(context.Background(), nil)

	// Assert results
	assert.Error(t, err)
	assert.Nil(t, request)
	assert.Contains(t, err.Error(), "milestone is required")
}

func TestVerificationWorkflowService_CollectVerificationEvidence(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method
	err := service.CollectVerificationEvidence(context.Background(), "test-request-1")

	// Assert results
	assert.NoError(t, err)
}

func TestVerificationWorkflowService_CollectVerificationEvidence_EmptyRequestID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty request ID
	err := service.CollectVerificationEvidence(context.Background(), "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request ID is required")
}

func TestVerificationWorkflowService_ExecuteMultiPartyApproval(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method
	err := service.ExecuteMultiPartyApproval(context.Background(), "test-request-1")

	// Assert results
	assert.NoError(t, err)
}

func TestVerificationWorkflowService_ExecuteMultiPartyApproval_EmptyRequestID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty request ID
	err := service.ExecuteMultiPartyApproval(context.Background(), "")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request ID is required")
}

func TestVerificationWorkflowService_CreateAuditTrail(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method
	err := service.CreateAuditTrail(context.Background(), "test-request-1", "test-action", "test-details")

	// Assert results
	assert.NoError(t, err)
}

func TestVerificationWorkflowService_CreateAuditTrail_EmptyRequestID(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockVerificationMilestoneRepository{}
	mockOracleService := &OracleVerificationService{}
	mockMessaging := &messaging.Service{}
	mockAuditRepo := &mockVerificationAuditRepository{}

	// Create service
	service := NewVerificationWorkflowService(
		mockMilestoneRepo,
		mockOracleService,
		mockMessaging,
		mockAuditRepo,
	)

	// Execute the method with empty request ID
	err := service.CreateAuditTrail(context.Background(), "", "test-action", "test-details")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request ID is required")
}
