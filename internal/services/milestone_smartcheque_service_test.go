package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockSmartChequeRepository implements the SmartChequeRepositoryInterface for testing
type mockSmartChequeRepository struct {
	mock.Mock
}

func (m *mockSmartChequeRepository) CreateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepository) GetSmartChequeByID(ctx context.Context, id string) (*models.SmartCheque, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) UpdateSmartCheque(ctx context.Context, smartCheque *models.SmartCheque) error {
	args := m.Called(ctx, smartCheque)
	return args.Error(0)
}

func (m *mockSmartChequeRepository) DeleteSmartCheque(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSmartChequeRepository) GetSmartChequesByPayer(ctx context.Context, payerID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payerID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequesByPayee(ctx context.Context, payeeID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, payeeID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequesByStatus(ctx context.Context, status models.SmartChequeStatus, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, status, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequesByContract(ctx context.Context, contractID string, limit, offset int) ([]*models.SmartCheque, error) {
	args := m.Called(ctx, contractID, limit, offset)
	return args.Get(0).([]*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequesByMilestone(ctx context.Context, milestoneID string) (*models.SmartCheque, error) {
	args := m.Called(ctx, milestoneID)
	return args.Get(0).(*models.SmartCheque), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequeCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockSmartChequeRepository) GetSmartChequeCountByStatus(ctx context.Context) (map[models.SmartChequeStatus]int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[models.SmartChequeStatus]int64), args.Error(1)
}

// mockVerificationWorkflowService implements the VerificationWorkflowServiceInterface for testing
type mockVerificationWorkflowService struct {
	mock.Mock
}

func (m *mockVerificationWorkflowService) GenerateVerificationRequest(ctx context.Context, milestone *models.ContractMilestone) (*VerificationRequest, error) {
	args := m.Called(ctx, milestone)
	return args.Get(0).(*VerificationRequest), args.Error(1)
}

func (m *mockVerificationWorkflowService) CollectVerificationEvidence(ctx context.Context, requestID string) error {
	args := m.Called(ctx, requestID)
	return args.Error(0)
}

func (m *mockVerificationWorkflowService) ExecuteMultiPartyApproval(ctx context.Context, requestID string) error {
	args := m.Called(ctx, requestID)
	return args.Error(0)
}

func (m *mockVerificationWorkflowService) CreateAuditTrail(ctx context.Context, requestID string, action string, details string) error {
	args := m.Called(ctx, requestID, action, details)
	return args.Error(0)
}

// mockDisputeHandlingService implements the DisputeHandlingServiceInterface for testing
type mockDisputeHandlingService struct {
	mock.Mock
}

func (m *mockDisputeHandlingService) InitiateMilestoneDispute(ctx context.Context, milestoneID string, reason string) error {
	args := m.Called(ctx, milestoneID, reason)
	return args.Error(0)
}

func (m *mockDisputeHandlingService) HoldMilestoneFunds(ctx context.Context, milestoneID string) error {
	args := m.Called(ctx, milestoneID)
	return args.Error(0)
}

func (m *mockDisputeHandlingService) ExecuteDisputeResolution(ctx context.Context, disputeID string) error {
	args := m.Called(ctx, disputeID)
	return args.Error(0)
}

func (m *mockDisputeHandlingService) EnforceDisputeOutcome(ctx context.Context, disputeID string, outcome string) error {
	args := m.Called(ctx, disputeID, outcome)
	return args.Error(0)
}

// mockXRPLService implements a mock XRPL service
type mockXRPLService struct {
	mock.Mock
}

func (m *mockXRPLService) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockXRPLService) CreateWallet() (*xrpl.WalletInfo, error) {
	args := m.Called()
	return args.Get(0).(*xrpl.WalletInfo), args.Error(1)
}

func (m *mockXRPLService) ValidateAddress(address string) bool {
	args := m.Called(address)
	return args.Bool(0)
}

func (m *mockXRPLService) GetAccountInfo(address string) (interface{}, error) {
	args := m.Called(address)
	return args.Get(0), args.Error(1)
}

func (m *mockXRPLService) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockXRPLService) CreateSmartChequeEscrow(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string) (*xrpl.TransactionResult, string, error) {
	args := m.Called(payerAddress, payeeAddress, amount, currency, milestoneSecret)
	return args.Get(0).(*xrpl.TransactionResult), args.String(1), args.Error(2)
}

func (m *mockXRPLService) CompleteSmartChequeMilestone(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string) (*xrpl.TransactionResult, error) {
	args := m.Called(payeeAddress, ownerAddress, sequence, condition, fulfillment)
	return args.Get(0).(*xrpl.TransactionResult), args.Error(1)
}

func (m *mockXRPLService) CancelSmartCheque(accountAddress, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error) {
	args := m.Called(accountAddress, ownerAddress, sequence)
	return args.Get(0).(*xrpl.TransactionResult), args.Error(1)
}

func (m *mockXRPLService) GetEscrowStatus(ownerAddress string, sequence string) (*xrpl.EscrowInfo, error) {
	args := m.Called(ownerAddress, sequence)
	return args.Get(0).(*xrpl.EscrowInfo), args.Error(1)
}

func (m *mockXRPLService) GenerateCondition(secret string) (condition string, fulfillment string, err error) {
	args := m.Called(secret)
	return args.String(0), args.String(1), args.Error(2)
}

func TestGenerateSmartChequeFromMilestone(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
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

	// Create test contract
	contract := &models.Contract{
		ID:      "test-contract-1",
		Parties: []string{"payer-1", "payee-1"},
	}

	// Set up mock expectations
	mockContractRepo.On("GetContractByID", mock.Anything, "test-contract-1").Return(contract, nil)
	mockSmartChequeRepo.On("CreateSmartCheque", mock.Anything, mock.AnythingOfType("*models.SmartCheque")).Return(nil)

	// Execute the method
	smartCheque, err := service.GenerateSmartChequeFromMilestone(context.Background(), milestone)

	// Assert results
	assert.NoError(t, err)
	assert.NotNil(t, smartCheque)
	assert.Equal(t, "sc-test-milestone-1", smartCheque.ID)
	assert.Equal(t, "test-contract-1", smartCheque.ContractHash)
	assert.Equal(t, "payer-1", smartCheque.PayerID)
	assert.Equal(t, "payee-1", smartCheque.PayeeID)
	assert.Equal(t, models.SmartChequeStatusCreated, smartCheque.Status)
	assert.Len(t, smartCheque.Milestones, 1)
	assert.Equal(t, "test-milestone-1", smartCheque.Milestones[0].ID)

	// Verify mock expectations
	mockContractRepo.AssertExpectations(t)
}

func TestTriggerPaymentRelease(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
	)

	// Create test milestone (completed)
	now := time.Now()
	milestone := &models.ContractMilestone{
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

	// Create test smart cheque
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
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, smartCheque).Return(nil)

	// Execute the method
	err := service.TriggerPaymentRelease(context.Background(), "test-milestone-1")

	// Assert results
	assert.NoError(t, err)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockSmartChequeRepo.AssertExpectations(t)
}

func TestTriggerPaymentReleaseIncompleteMilestone(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
	)

	// Create test milestone (incomplete)
	now := time.Now()
	milestone := &models.ContractMilestone{
		ID:                   "test-milestone-1",
		ContractID:           "test-contract-1",
		MilestoneID:          "m-1",
		SequenceNumber:       1,
		TriggerConditions:    "Test condition",
		VerificationCriteria: "Manual verification",
		PercentageComplete:   50.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Set up mock expectations
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "test-milestone-1").Return(milestone, nil)

	// Execute the method
	err := service.TriggerPaymentRelease(context.Background(), "test-milestone-1")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not completed yet")

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
}

func TestMapMilestoneToEscrow(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
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
	err := service.MapMilestoneToEscrow(context.Background(), "test-milestone-1", "escrow-address-123")

	// Assert results
	assert.NoError(t, err)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
}

func TestMapMilestoneToEscrowErrors(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
	)

	// Test with empty milestone ID
	err := service.MapMilestoneToEscrow(context.Background(), "", "escrow-address-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "milestone ID is required")

	// Test with empty escrow address
	err = service.MapMilestoneToEscrow(context.Background(), "test-milestone-1", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escrow address is required")

	// Test with milestone not found
	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, "non-existent-milestone").Return((*models.ContractMilestone)(nil), fmt.Errorf("milestone not found"))
	err = service.MapMilestoneToEscrow(context.Background(), "non-existent-milestone", "escrow-address-123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get milestone")
}

func TestTriggerPaymentReleaseWithSmartCheque(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
	)

	// Create test milestone (completed)
	now := time.Now()
	milestone := &models.ContractMilestone{
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

	// Create test smart cheque
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
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, smartCheque).Return(nil)

	// Execute the method
	err := service.TriggerPaymentRelease(context.Background(), "test-milestone-1")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, models.SmartChequeStatusCompleted, smartCheque.Status)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockSmartChequeRepo.AssertExpectations(t)
}

func TestHandleMilestoneFailure(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
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
		PercentageComplete:   50.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create test smart cheque
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
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)
	mockSmartChequeRepo.On("UpdateSmartCheque", mock.Anything, smartCheque).Return(nil)

	// Execute the method
	err := service.HandleMilestoneFailure(context.Background(), "test-milestone-1", "milestone failed due to external factors")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, models.SmartChequeStatusDisputed, smartCheque.Status)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockSmartChequeRepo.AssertExpectations(t)
}

func TestProcessPartialPayment(t *testing.T) {
	// Create mocks
	mockMilestoneRepo := &mockMilestoneRepository{}
	mockContractRepo := &mockContractRepository{}
	mockSmartChequeRepo := &mockSmartChequeRepository{}
	mockVerificationWorkflow := &mockVerificationWorkflowService{}
	mockDisputeHandling := &mockDisputeHandlingService{}
	mockXRPL := &mockXRPLService{}

	// Create service
	service := NewMilestoneSmartChequeService(
		mockMilestoneRepo,
		mockContractRepo,
		mockSmartChequeRepo,
		mockVerificationWorkflow,
		mockDisputeHandling,
		mockXRPL,
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
		PercentageComplete:   50.0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create test smart cheque
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
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, milestone).Return(nil)

	// Execute the method
	err := service.ProcessPartialPayment(context.Background(), "test-milestone-1", 75.0)

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, 75.0, milestone.PercentageComplete)

	// Verify mock expectations
	mockMilestoneRepo.AssertExpectations(t)
	mockSmartChequeRepo.AssertExpectations(t)
}
