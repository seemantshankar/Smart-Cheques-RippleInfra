package services

import (
	"context"

	"github.com/smart-payment-infrastructure/internal/models"
	repo "github.com/smart-payment-infrastructure/internal/repository"
)

// ContractServiceInterface defines service-level operations for contracts and milestones
// It composes contract and milestone repository operations and is the integration seam
// for higher-level orchestration and validation in the future.
type ContractServiceInterface interface {
	// Contract operations
	CreateContract(ctx context.Context, contract *models.Contract) error
	GetContractByID(ctx context.Context, id string) (*models.Contract, error)
	UpdateContract(ctx context.Context, contract *models.Contract) error
	DeleteContract(ctx context.Context, id string) error

	GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error)
	GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error)
	GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error)

	// Milestone operations
	CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error)
	UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error
	DeleteMilestone(ctx context.Context, id string) error
	GetMilestonesByContractID(ctx context.Context, contractID string) ([]*models.ContractMilestone, error)
}

// ContractService is a thin orchestration layer over repositories.
// Future enhancements: validation, status transitions, audit/events, permissions.
type ContractService struct {
	contracts  repo.ContractRepositoryInterface
	milestones repo.ContractMilestoneRepositoryInterface
}

func NewContractService(contracts repo.ContractRepositoryInterface, milestones repo.ContractMilestoneRepositoryInterface) *ContractService {
	return &ContractService{contracts: contracts, milestones: milestones}
}

// Contract methods
func (s *ContractService) CreateContract(ctx context.Context, contract *models.Contract) error {
	// TODO: add domain validations (e.g., required fields, status transitions)
	return s.contracts.CreateContract(ctx, contract)
}

func (s *ContractService) GetContractByID(ctx context.Context, id string) (*models.Contract, error) {
	return s.contracts.GetContractByID(ctx, id)
}

func (s *ContractService) UpdateContract(ctx context.Context, contract *models.Contract) error {
	return s.contracts.UpdateContract(ctx, contract)
}

func (s *ContractService) DeleteContract(ctx context.Context, id string) error {
	return s.contracts.DeleteContract(ctx, id)
}

func (s *ContractService) GetContractsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Contract, error) {
	return s.contracts.GetContractsByStatus(ctx, status, limit, offset)
}

func (s *ContractService) GetContractsByType(ctx context.Context, contractType string, limit, offset int) ([]*models.Contract, error) {
	return s.contracts.GetContractsByType(ctx, contractType, limit, offset)
}

func (s *ContractService) GetContractsByParty(ctx context.Context, party string, limit, offset int) ([]*models.Contract, error) {
	return s.contracts.GetContractsByParty(ctx, party, limit, offset)
}

// Milestone methods
func (s *ContractService) CreateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	return s.milestones.CreateMilestone(ctx, milestone)
}

func (s *ContractService) GetMilestoneByID(ctx context.Context, id string) (*models.ContractMilestone, error) {
	return s.milestones.GetMilestoneByID(ctx, id)
}

func (s *ContractService) UpdateMilestone(ctx context.Context, milestone *models.ContractMilestone) error {
	return s.milestones.UpdateMilestone(ctx, milestone)
}

func (s *ContractService) DeleteMilestone(ctx context.Context, id string) error {
	return s.milestones.DeleteMilestone(ctx, id)
}

func (s *ContractService) GetMilestonesByContractID(ctx context.Context, contractID string) ([]*models.ContractMilestone, error) {
	return s.milestones.GetMilestonesByContractID(ctx, contractID)
}
