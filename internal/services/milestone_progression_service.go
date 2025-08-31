package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// MilestoneProgressionServiceInterface defines the interface for milestone progression operations
type MilestoneProgressionServiceInterface interface {
	// StartMilestone starts a milestone and updates its status
	StartMilestone(ctx context.Context, milestoneID string) error

	// VerifyMilestone verifies a milestone and updates its status
	VerifyMilestone(ctx context.Context, milestoneID string, verificationData map[string]interface{}) error

	// CompleteMilestone marks a milestone as completed
	CompleteMilestone(ctx context.Context, milestoneID string) error

	// FailMilestone marks a milestone as failed
	FailMilestone(ctx context.Context, milestoneID string, reason string) error

	// UpdateMilestoneProgress updates the progress of a milestone
	UpdateMilestoneProgress(ctx context.Context, milestoneID string, percentageComplete float64) error

	// CheckDependencies checks if all dependencies for a milestone are met
	CheckDependencies(ctx context.Context, milestoneID string) (bool, error)

	// GetMilestoneProgress calculates the overall progress of a milestone
	GetMilestoneProgress(ctx context.Context, milestoneID string) (*MilestoneProgress, error)
}

// MilestoneProgress represents the progress of a milestone
type MilestoneProgress struct {
	PercentageComplete float64   `json:"percentage_complete"`
	Status             string    `json:"status"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	EstimatedDuration  string    `json:"estimated_duration"`
	ActualDuration     string    `json:"actual_duration"`
}

// milestoneProgressionService implements MilestoneProgressionServiceInterface
type milestoneProgressionService struct {
	milestoneRepo   repository.MilestoneRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
}

// NewMilestoneProgressionService creates a new milestone progression service
func NewMilestoneProgressionService(
	milestoneRepo repository.MilestoneRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
) MilestoneProgressionServiceInterface {
	return &milestoneProgressionService{
		milestoneRepo:   milestoneRepo,
		smartChequeRepo: smartChequeRepo,
	}
}

// StartMilestone starts a milestone and updates its status
func (s *milestoneProgressionService) StartMilestone(ctx context.Context, milestoneID string) error {
	if milestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Check if milestone can be started
	if milestone.Status != "pending" {
		return fmt.Errorf("milestone %s cannot be started from status %s", milestoneID, milestone.Status)
	}

	// Check dependencies
	depsMet, err := s.CheckDependencies(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to check dependencies for milestone %s: %w", milestoneID, err)
	}

	if !depsMet {
		return fmt.Errorf("dependencies not met for milestone %s", milestoneID)
	}

	// Update milestone status
	milestone.Status = "in_progress"
	if milestone.ActualStartDate == nil {
		now := time.Now()
		milestone.ActualStartDate = &now
	}
	milestone.UpdatedAt = time.Now()

	// Save the updated milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	return nil
}

// VerifyMilestone verifies a milestone and updates its status
func (s *milestoneProgressionService) VerifyMilestone(ctx context.Context, milestoneID string, verificationData map[string]interface{}) error {
	if milestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Check if milestone can be verified
	if milestone.Status != "in_progress" {
		return fmt.Errorf("milestone %s cannot be verified from status %s", milestoneID, milestone.Status)
	}

	// In a real implementation, we would:
	// 1. Validate the verification data based on the verification criteria
	// 2. Execute the verification workflow (manual, oracle, hybrid)
	// 3. Update milestone with verification results

	// For now, we'll assume verification is successful
	milestone.Status = "verified"
	milestone.UpdatedAt = time.Now()

	// Save the updated milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	return nil
}

// CompleteMilestone marks a milestone as completed
func (s *milestoneProgressionService) CompleteMilestone(ctx context.Context, milestoneID string) error {
	if milestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Check if milestone can be completed
	if milestone.Status != "verified" {
		return fmt.Errorf("milestone %s cannot be completed from status %s", milestoneID, milestone.Status)
	}

	// Update milestone status
	milestone.Status = "completed"
	milestone.PercentageComplete = 100.0
	if milestone.ActualEndDate == nil {
		now := time.Now()
		milestone.ActualEndDate = &now
	}
	milestone.UpdatedAt = time.Now()

	// Save the updated milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	// Trigger payment release if associated with a smart cheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequesByMilestone(ctx, milestoneID)
	if err == nil && smartCheque != nil {
		// Update smart cheque status
		smartCheque.Status = models.SmartChequeStatusCompleted
		smartCheque.UpdatedAt = time.Now()
		if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to update smart cheque %s: %v\n", smartCheque.ID, err)
		}
	}

	return nil
}

// FailMilestone marks a milestone as failed
func (s *milestoneProgressionService) FailMilestone(ctx context.Context, milestoneID string, reason string) error {
	if milestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Update milestone status
	milestone.Status = "failed"
	milestone.UpdatedAt = time.Now()

	// Save the updated milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	// Update associated smart cheque if it exists
	smartCheque, err := s.smartChequeRepo.GetSmartChequesByMilestone(ctx, milestoneID)
	if err == nil && smartCheque != nil {
		// Update smart cheque status
		smartCheque.Status = models.SmartChequeStatusDisputed
		smartCheque.UpdatedAt = time.Now()
		if err := s.smartChequeRepo.UpdateSmartCheque(ctx, smartCheque); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to update smart cheque %s: %v\n", smartCheque.ID, err)
		}
	}

	return nil
}

// UpdateMilestoneProgress updates the progress of a milestone
func (s *milestoneProgressionService) UpdateMilestoneProgress(ctx context.Context, milestoneID string, percentageComplete float64) error {
	if milestoneID == "" {
		return fmt.Errorf("milestone ID is required")
	}

	// Validate percentage
	if percentageComplete < 0 || percentageComplete > 100 {
		return fmt.Errorf("invalid percentage: %f", percentageComplete)
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Update milestone progress
	milestone.PercentageComplete = percentageComplete
	milestone.UpdatedAt = time.Now()

	// Save the updated milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	return nil
}

// CheckDependencies checks if all dependencies for a milestone are met
func (s *milestoneProgressionService) CheckDependencies(ctx context.Context, milestoneID string) (bool, error) {
	if milestoneID == "" {
		return false, fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return false, fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// If no dependencies, return true
	if len(milestone.Dependencies) == 0 {
		return true, nil
	}

	// Check each dependency
	for _, depID := range milestone.Dependencies {
		depMilestone, err := s.milestoneRepo.GetMilestoneByID(ctx, depID)
		if err != nil {
			return false, fmt.Errorf("failed to get dependency milestone %s: %w", depID, err)
		}

		// Check if dependency is completed
		if depMilestone.Status != "completed" {
			return false, nil
		}
	}

	return true, nil
}

// GetMilestoneProgress calculates the overall progress of a milestone
func (s *milestoneProgressionService) GetMilestoneProgress(ctx context.Context, milestoneID string) (*MilestoneProgress, error) {
	if milestoneID == "" {
		return nil, fmt.Errorf("milestone ID is required")
	}

	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	progress := &MilestoneProgress{
		PercentageComplete: milestone.PercentageComplete,
		Status:             milestone.Status,
	}

	// Set start date
	if milestone.ActualStartDate != nil {
		progress.StartDate = *milestone.ActualStartDate
	} else if milestone.EstimatedStartDate != nil {
		progress.StartDate = *milestone.EstimatedStartDate
	}

	// Set end date
	if milestone.ActualEndDate != nil {
		progress.EndDate = *milestone.ActualEndDate
	} else if milestone.EstimatedEndDate != nil {
		progress.EndDate = *milestone.EstimatedEndDate
	}

	// Set durations
	if milestone.EstimatedDuration > 0 {
		progress.EstimatedDuration = milestone.EstimatedDuration.String()
	}

	if milestone.ActualDuration != nil {
		progress.ActualDuration = milestone.ActualDuration.String()
	}

	return progress, nil
}
