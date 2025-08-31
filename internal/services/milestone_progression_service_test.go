package services

import (
	"context"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMilestoneProgressionService_StartMilestone(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:           milestoneID,
		Status:       "pending",
		Dependencies: []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)

	err := service.StartMilestone(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.Equal(t, "in_progress", milestone.Status)
	assert.NotNil(t, milestone.ActualStartDate)

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_StartMilestone_InvalidStatus(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:           milestoneID,
		Status:       "in_progress", // Already in progress
		Dependencies: []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)

	err := service.StartMilestone(context.Background(), milestoneID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be started from status")

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_VerifyMilestone(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:        milestoneID,
		Status:    "in_progress",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)

	verificationData := map[string]interface{}{
		"evidence": "test-evidence",
	}

	err := service.VerifyMilestone(context.Background(), milestoneID, verificationData)
	assert.NoError(t, err)
	assert.Equal(t, "verified", milestone.Status)

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_CompleteMilestone(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:        milestoneID,
		Status:    "verified",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)
	// Add missing mock expectation for GetSmartChequesByMilestone
	mockSmartChequeRepo.On("GetSmartChequesByMilestone", mock.Anything, milestoneID).Return((*models.SmartCheque)(nil), nil)

	err := service.CompleteMilestone(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", milestone.Status)
	assert.Equal(t, float64(100.0), milestone.PercentageComplete)
	assert.NotNil(t, milestone.ActualEndDate)

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_FailMilestone(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:        milestoneID,
		Status:    "in_progress",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)
	// Add missing mock expectation for GetSmartChequesByMilestone
	mockSmartChequeRepo.On("GetSmartChequesByMilestone", mock.Anything, milestoneID).Return((*models.SmartCheque)(nil), nil)

	err := service.FailMilestone(context.Background(), milestoneID, "test failure reason")
	assert.NoError(t, err)
	assert.Equal(t, "failed", milestone.Status)

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_UpdateMilestoneProgress(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:        milestoneID,
		Status:    "in_progress",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)
	mockMilestoneRepo.On("UpdateMilestone", mock.Anything, mock.AnythingOfType("*models.ContractMilestone")).Return(nil)

	err := service.UpdateMilestoneProgress(context.Background(), milestoneID, 75.5)
	assert.NoError(t, err)
	assert.Equal(t, float64(75.5), milestone.PercentageComplete)

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_UpdateMilestoneProgress_InvalidPercentage(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"

	err := service.UpdateMilestoneProgress(context.Background(), milestoneID, 150.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid percentage")

	err = service.UpdateMilestoneProgress(context.Background(), milestoneID, -10.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid percentage")

	mockMilestoneRepo.AssertExpectations(t)
}

func TestMilestoneProgressionService_CheckDependencies_NoDependencies(t *testing.T) {
	mockMilestoneRepo := &mocks.MilestoneRepositoryInterface{}
	mockSmartChequeRepo := &mocks.SmartChequeRepositoryInterface{}
	service := NewMilestoneProgressionService(mockMilestoneRepo, mockSmartChequeRepo)

	milestoneID := "test-milestone-1"
	now := time.Now()

	milestone := &models.ContractMilestone{
		ID:           milestoneID,
		Status:       "pending",
		Dependencies: []string{}, // No dependencies
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockMilestoneRepo.On("GetMilestoneByID", mock.Anything, milestoneID).Return(milestone, nil)

	depsMet, err := service.CheckDependencies(context.Background(), milestoneID)
	assert.NoError(t, err)
	assert.True(t, depsMet)

	mockMilestoneRepo.AssertExpectations(t)
}
