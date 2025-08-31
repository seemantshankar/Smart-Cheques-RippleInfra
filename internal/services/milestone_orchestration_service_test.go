package services

import (
	"context"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockMilestoneRepository implements the MilestoneRepositoryInterface for testing
type mockMilestoneRepository struct {
	mock.Mock
}

func (m *mockMilestoneRepository) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockMilestoneRepository) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockMilestoneRepository) DeleteMilestone(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMilestoneRepository) GetMilestonesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetOverdueMilestones(ctx context.Context, asOfDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, asOfDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonesByPriority(ctx context.Context, priority int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, priority, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonesByCategory(ctx context.Context, category string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, category, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonesByRiskLevel(ctx context.Context, riskLevel string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, riskLevel, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetCriticalPathMilestones(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) SearchMilestones(ctx context.Context, query string, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestoneDependencies(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestoneDependents(ctx context.Context, milestoneID string) ([]*models.MilestoneDependency, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).([]*models.MilestoneDependency), args.Error(1)
}

func (m *mockMilestoneRepository) ResolveDependencyGraph(ctx context.Context, contractID string) (map[string][]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func (m *mockMilestoneRepository) ValidateDependencyGraph(ctx context.Context, contractID string) (bool, error) {
	args := m.Called(ctx, contractID)
	return args.Bool(0), args.Error(1)
}

func (m *mockMilestoneRepository) GetTopologicalOrder(ctx context.Context, contractID string) ([]string, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockMilestoneRepository) CreateMilestoneDependency(ctx context.Context, dependency *models.MilestoneDependency) error {
	args := m.Called(ctx, dependency)
	return args.Error(0)
}

func (m *mockMilestoneRepository) DeleteMilestoneDependency(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMilestoneRepository) BatchUpdateMilestoneStatus(ctx context.Context, milestoneIDs []string, status string) error {
	args := m.Called(ctx, milestoneIDs, status)
	return args.Error(0)
}

func (m *mockMilestoneRepository) BatchUpdateMilestoneProgress(ctx context.Context, updates []repository.MilestoneProgressUpdate) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *mockMilestoneRepository) BatchCreateMilestones(ctx context.Context, milestones []*models.ContractMilestone) error {
	args := m.Called(ctx, milestones)
	return args.Error(0)
}

func (m *mockMilestoneRepository) BatchDeleteMilestones(ctx context.Context, milestoneIDs []string) error {
	args := m.Called(ctx, milestoneIDs)
	return args.Error(0)
}

func (m *mockMilestoneRepository) GetMilestoneCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	return args.Get(0).(*repository.MilestoneStats), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestonePerformanceMetrics), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestoneTimelineAnalysis(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneTimelineAnalysis), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	args := m.Called(ctx, contractID)
	return args.Get(0).(*repository.MilestoneRiskAnalysis), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestoneProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

func (m *mockMilestoneRepository) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

func (m *mockMilestoneRepository) CreateMilestoneProgressEntry(ctx context.Context, entry *repository.MilestoneProgressEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *mockMilestoneRepository) GetMilestoneProgressHistory(ctx context.Context, milestoneID string, limit, offset int) ([]*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID, limit, offset)
	return args.Get(0).([]*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockMilestoneRepository) GetLatestProgressUpdate(ctx context.Context, milestoneID string) (*repository.MilestoneProgressEntry, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).(*repository.MilestoneProgressEntry), args.Error(1)
}

func (m *mockMilestoneRepository) FilterMilestones(ctx context.Context, filter *repository.MilestoneFilter) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetMilestonesByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, startDate, endDate, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

func (m *mockMilestoneRepository) GetUpcomingMilestones(ctx context.Context, daysAhead int, limit, offset int) ([]*models.ContractMilestone, error) {
	args := m.Called(ctx, daysAhead, limit, offset)
	return args.Get(0).([]*models.ContractMilestone), args.Error(1)
}

// mockContractRepository implements the ContractRepositoryInterface for testing
type mockContractRepository struct {
	mock.Mock
}

func (m *mockContractRepository) CreateContract(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}

func (m *mockContractRepository) GetContractByID(ctx context.Context, id string) (*models.Contract, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Contract), args.Error(1)
}

func (m *mockContractRepository) UpdateContract(ctx context.Context, contract *models.Contract) error {
	args := m.Called(ctx, contract)
	return args.Error(0)
}

func (m *mockContractRepository) DeleteContract(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockContractRepository) GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}

func (m *mockContractRepository) GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, contractType, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}

func (m *mockContractRepository) GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error) {
	args := m.Called(ctx, party, limit, offset)
	return args.Get(0).([]*models.Contract), args.Error(1)
}

// mockNotificationService implements the MilestoneNotificationServiceInterface for testing
type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) SendDeadlineAlert(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockNotificationService) SendProgressUpdate(ctx context.Context, milestone *models.ContractMilestone, progress float64) error {
	args := m.Called(ctx, milestone, progress)
	return args.Error(0)
}

func (m *mockNotificationService) SendOverdueAlert(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

func (m *mockNotificationService) SendCompletionNotification(ctx context.Context, milestone *models.ContractMilestone) error {
	args := m.Called(ctx, milestone)
	return args.Error(0)
}

// mockAnalyticsService implements the MilestoneAnalyticsServiceInterface for testing
type mockAnalyticsService struct {
	mock.Mock
}

func (m *mockAnalyticsService) GetCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error) {
	args := m.Called(ctx, contractID, startDate, endDate)
	return args.Get(0).(*repository.MilestoneStats), args.Error(1)
}

func (m *mockAnalyticsService) GetProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error) {
	args := m.Called(ctx, contractID, days)
	return args.Get(0).([]*repository.MilestoneProgressTrend), args.Error(1)
}

func (m *mockAnalyticsService) GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]*repository.DelayedMilestoneReport), args.Error(1)
}

func TestCreateMilestonesFromContract(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockNotificationSvc := &mockNotificationService{}
	mockAnalyticsSvc := &mockAnalyticsService{}

	// Create service
	service := NewMilestoneOrchestrationService(
		mockMilestoneRepo,
		mockContractRepo,
		mockNotificationSvc,
		mockAnalyticsSvc,
	)

	// Create test contract
	now := time.Now()
	contract := &models.Contract{
		ID:        "test-contract-1",
		CreatedAt: now,
		Obligations: []models.Obligation{
			{
				ID:          "ob-1",
				Description: "Test obligation 1",
				DueDate:     now.Add(24 * time.Hour),
			},
			{
				ID:          "ob-2",
				Description: "Test obligation 2",
				DueDate:     now.Add(48 * time.Hour),
			},
		},
	}

	// Set up mock expectations
	mockMilestoneRepo.On("CreateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil).Twice()

	// Execute the method
	milestones, err := service.CreateMilestonesFromContract(context.Background(), contract)

	// Assert results
	assert.NoError(t, err)
	assert.Len(t, milestones, 2)
	assert.Equal(t, "ms-test-contract-1-0", milestones[0].ID)
	assert.Equal(t, "ms-test-contract-1-1", milestones[1].ID)
	assert.Contains(t, milestones[0].TriggerConditions, "Test obligation 1")
	assert.Contains(t, milestones[1].TriggerConditions, "Test obligation 2")

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
}

func TestUpdateMilestoneProgress(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockNotificationSvc := &mockNotificationService{}
	mockAnalyticsSvc := &mockAnalyticsService{}

	// Create service
	service := NewMilestoneOrchestrationService(
		mockMilestoneRepo,
		mockContractRepo,
		mockNotificationSvc,
		mockAnalyticsSvc,
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
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)
	mockMilestoneRepo.On("CreateMilestoneProgressEntry", mock.Anything, mock.AnythingOfType("*repository.MilestoneProgressEntry")).Return(nil)
	mockNotificationSvc.On("SendProgressUpdate", mock.Anything, milestone, 50.0).Return(nil)
	mockNotificationSvc.On("SendProgressUpdate", mock.Anything, milestone, 100.0).Return(nil)
	mockNotificationSvc.On("SendCompletionNotification", mock.Anything, milestone).Return(nil)

	// Test updating progress to 50%
	err := service.UpdateMilestoneProgress(context.Background(), "test-milestone-1", 50.0, "Halfway done")
	assert.NoError(t, err)

	// Test updating progress to 100%
	err = service.UpdateMilestoneProgress(context.Background(), "test-milestone-1", 100.0, "Completed")
	assert.NoError(t, err)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockNotificationSvc.AssertExpectations(t)
}

func TestValidateMilestoneCompletion(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockNotificationSvc := &mockNotificationService{}
	mockAnalyticsSvc := &mockAnalyticsService{}

	// Create service
	service := NewMilestoneOrchestrationService(
		mockMilestoneRepo,
		mockContractRepo,
		mockNotificationSvc,
		mockAnalyticsSvc,
	)

	// Create test milestone (completed)
	now := time.Now()
	completedMilestone := &models.ContractMilestone{
		ID:                   "test-milestone-1",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-1",
		SequenceNumber:       1,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		PercentageComplete:   100.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create test milestone (incomplete)
	incompleteMilestone := &models.ContractMilestone{
		ID:                   "test-milestone-2",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-2",
		SequenceNumber:       2,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		PercentageComplete:   50.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-1").Return(completedMilestone, nil)
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-2").Return(incompleteMilestone, nil)

	// Test completed milestone
	isValid, err := service.ValidateMilestoneCompletion(context.Background(), "test-milestone-1")
	assert.NoError(t, err)
	assert.True(t, isValid)

	// Test incomplete milestone
	isValid, err = service.ValidateMilestoneCompletion(context.Background(), "test-milestone-2")
	assert.NoError(t, err)
	assert.False(t, isValid)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
}
