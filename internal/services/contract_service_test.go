package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/smart-payment-infrastructure/internal/models"
	repomocks "github.com/smart-payment-infrastructure/internal/repository/mocks"
)

func TestContractService_ContractCRUDAndQueries(t *testing.T) {
	ctx := context.Background()
	contractRepo := &repomocks.ContractRepositoryInterface{}
	milestoneRepo := &repomocks.ContractMilestoneRepositoryInterface{}
	svc := NewContractService(contractRepo, milestoneRepo)

	contractID := uuid.NewString()
	contract := &models.Contract{
		ID:           contractID,
		Status:       "active",
		ContractType: "service_agreement",
		Version:      "v1",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create
	contractRepo.On("CreateContract", ctx, contract).Return(nil).Once()
	err := svc.CreateContract(ctx, contract)
	assert.NoError(t, err)

	// GetByID
	contractRepo.On("GetContractByID", ctx, contractID).Return(contract, nil).Once()
	got, err := svc.GetContractByID(ctx, contractID)
	assert.NoError(t, err)
	assert.Equal(t, contract, got)

	// Update
	contract.Version = "v2"
	contractRepo.On("UpdateContract", ctx, contract).Return(nil).Once()
	err = svc.UpdateContract(ctx, contract)
	assert.NoError(t, err)

	// Queries
	contractRepo.On("GetContractsByStatus", ctx, "active", 10, 0).Return([]*models.Contract{contract}, nil).Once()
	byStatus, err := svc.GetContractsByStatus(ctx, "active", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, byStatus, 1)

	contractRepo.On("GetContractsByType", ctx, "service_agreement", 5, 0).Return([]*models.Contract{contract}, nil).Once()
	byType, err := svc.GetContractsByType(ctx, "service_agreement", 5, 0)
	assert.NoError(t, err)
	assert.Len(t, byType, 1)

	contractRepo.On("GetContractsByParty", ctx, "partyA", 3, 0).Return([]*models.Contract{}, nil).Once()
	byParty, err := svc.GetContractsByParty(ctx, "partyA", 3, 0)
	assert.NoError(t, err)
	assert.NotNil(t, byParty)

	// Delete
	contractRepo.On("DeleteContract", ctx, contractID).Return(nil).Once()
	err = svc.DeleteContract(ctx, contractID)
	assert.NoError(t, err)

	contractRepo.AssertExpectations(t)
	milestoneRepo.AssertExpectations(t)
}

func TestContractService_MilestoneCRUDAndQueries(t *testing.T) {
	ctx := context.Background()
	contractRepo := &repomocks.ContractRepositoryInterface{}
	milestoneRepo := &repomocks.ContractMilestoneRepositoryInterface{}
	svc := NewContractService(contractRepo, milestoneRepo)

	contractID := uuid.NewString()
	milestoneID := uuid.NewString()
	milestone := &models.ContractMilestone{
		ID:                   milestoneID,
		ContractID:           contractID,
		MilestoneID:          "M-001",
		SequenceOrder:        1,
		TriggerConditions:    "on-approval",
		VerificationCriteria: "deliverables-v1",
		EstimatedDuration:    time.Hour * 24 * 30,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Create
	milestoneRepo.On("CreateMilestone", ctx, milestone).Return(nil).Once()
	err := svc.CreateMilestone(ctx, milestone)
	assert.NoError(t, err)

	// GetByID
	milestoneRepo.On("GetMilestoneByID", ctx, milestoneID).Return(milestone, nil).Once()
	got, err := svc.GetMilestoneByID(ctx, milestoneID)
	assert.NoError(t, err)
	assert.Equal(t, milestone, got)

	// Update
	milestone.VerificationCriteria = "deliverables-v2"
	milestoneRepo.On("UpdateMilestone", ctx, milestone).Return(nil).Once()
	err = svc.UpdateMilestone(ctx, milestone)
	assert.NoError(t, err)

	// By ContractID
	milestoneRepo.On("GetMilestonesByContractID", ctx, contractID).Return([]*models.ContractMilestone{milestone}, nil).Once()
	list, err := svc.GetMilestonesByContractID(ctx, contractID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// Delete
	milestoneRepo.On("DeleteMilestone", ctx, milestoneID).Return(nil).Once()
	err = svc.DeleteMilestone(ctx, milestoneID)
	assert.NoError(t, err)

	contractRepo.AssertExpectations(t)
	milestoneRepo.AssertExpectations(t)
}
