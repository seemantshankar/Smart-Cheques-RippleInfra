package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

// MilestoneService provides simple operations for ContractMilestone management.
type MilestoneService interface {
	CreateMilestone(ctx context.Context, m *models.ContractMilestone) error
	ValidateMilestone(ctx context.Context, m *models.ContractMilestone) error
}

type milestoneServiceImpl struct{}

func NewMilestoneService() MilestoneService {
	return &milestoneServiceImpl{}
}

func (s *milestoneServiceImpl) CreateMilestone(ctx context.Context, m *models.ContractMilestone) error {
	if err := s.ValidateMilestone(ctx, m); err != nil {
		return err
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	return nil
}

func (s *milestoneServiceImpl) ValidateMilestone(ctx context.Context, m *models.ContractMilestone) error {
	if m == nil {
		return fmt.Errorf("milestone is nil")
	}
	if m.ContractID == "" {
		return fmt.Errorf("ContractID required")
	}
	if m.MilestoneID == "" {
		return fmt.Errorf("MilestoneID required")
	}
	if m.SequenceOrder <= 0 && m.SequenceNumber <= 0 {
		return fmt.Errorf("sequence order or sequence number required")
	}
	return nil
}

