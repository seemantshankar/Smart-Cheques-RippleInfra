package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// MilestoneOrchestrationServiceInterface defines the interface for milestone orchestration operations
type MilestoneOrchestrationServiceInterface interface {
	// CreateMilestonesFromContract creates milestones based on contract analysis
	CreateMilestonesFromContract(ctx context.Context, contract *models.Contract) ([]*models.ContractMilestone, error)

	// ResolveMilestoneDependencies automatically resolves dependencies between milestones
	ResolveMilestoneDependencies(ctx context.Context, contractID string) error

	// OptimizeMilestoneSchedule optimizes the timeline for milestone execution
	OptimizeMilestoneSchedule(ctx context.Context, contractID string) error

	// UpdateMilestoneProgress updates the progress of a milestone
	UpdateMilestoneProgress(ctx context.Context, milestoneID string, progress float64, notes string) error

	// ValidateMilestoneCompletion validates if a milestone has been properly completed
	ValidateMilestoneCompletion(ctx context.Context, milestoneID string) (bool, error)

	// GetMilestoneTimeline returns the timeline analysis for a contract's milestones
	GetMilestoneTimeline(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error)

	// GetMilestonePerformanceMetrics returns performance metrics for milestones
	GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error)

	// GetMilestoneRiskAnalysis returns risk analysis for milestones
	GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error)
}

// MilestoneNotificationServiceInterface defines the interface for milestone notifications
type MilestoneNotificationServiceInterface interface {
	// SendDeadlineAlert sends an alert for upcoming milestone deadlines
	SendDeadlineAlert(ctx context.Context, milestone *models.ContractMilestone) error

	// SendProgressUpdate sends a progress update notification
	SendProgressUpdate(ctx context.Context, milestone *models.ContractMilestone, progress float64) error

	// SendOverdueAlert sends an alert for overdue milestones
	SendOverdueAlert(ctx context.Context, milestone *models.ContractMilestone) error

	// SendCompletionNotification sends a notification when a milestone is completed
	SendCompletionNotification(ctx context.Context, milestone *models.ContractMilestone) error
}

// MilestoneAnalyticsServiceInterface defines the interface for milestone analytics
type MilestoneAnalyticsServiceInterface interface {
	// GetCompletionStats returns completion statistics for milestones
	GetCompletionStats(ctx context.Context, contractID *string, startDate, endDate *time.Time) (*repository.MilestoneStats, error)

	// GetProgressTrends returns progress trends for milestones
	GetProgressTrends(ctx context.Context, contractID *string, days int) ([]*repository.MilestoneProgressTrend, error)

	// GetDelayedMilestonesReport returns a report of delayed milestones
	GetDelayedMilestonesReport(ctx context.Context, threshold time.Duration) ([]*repository.DelayedMilestoneReport, error)
}

// milestoneOrchestrationService implements the MilestoneOrchestrationServiceInterface
type milestoneOrchestrationService struct {
	milestoneRepo   repository.MilestoneRepositoryInterface
	contractRepo    repository.ContractRepositoryInterface
	notificationSvc MilestoneNotificationServiceInterface
	analyticsSvc    MilestoneAnalyticsServiceInterface
}

// NewMilestoneOrchestrationService creates a new milestone orchestration service
func NewMilestoneOrchestrationService(
	milestoneRepo repository.MilestoneRepositoryInterface,
	contractRepo repository.ContractRepositoryInterface,
	notificationSvc MilestoneNotificationServiceInterface,
	analyticsSvc MilestoneAnalyticsServiceInterface,
) MilestoneOrchestrationServiceInterface {
	return &milestoneOrchestrationService{
		milestoneRepo:   milestoneRepo,
		contractRepo:    contractRepo,
		notificationSvc: notificationSvc,
		analyticsSvc:    analyticsSvc,
	}
}

// CreateMilestonesFromContract creates milestones based on contract analysis
func (s *milestoneOrchestrationService) CreateMilestonesFromContract(ctx context.Context, contract *models.Contract) ([]*models.ContractMilestone, error) {
	if contract == nil {
		return nil, fmt.Errorf("contract is nil")
	}

	// Pre-allocate milestones slice
	milestones := make([]*models.ContractMilestone, 0, len(contract.Obligations))

	// Create milestones based on contract obligations
	for i, obligation := range contract.Obligations {
		milestone := &models.ContractMilestone{
			ID:                   fmt.Sprintf("ms-%s-%d", contract.ID, i),
			ContractID:           contract.ID,
			MilestoneID:          fmt.Sprintf("m-%d", i),
			SequenceNumber:       i + 1,
			Category:             "obligation",
			Priority:             1,
			CriticalPath:         false,
			TriggerConditions:    fmt.Sprintf("Completion of obligation: %s", obligation.Description),
			VerificationCriteria: "Manual verification",
			EstimatedStartDate:   &contract.CreatedAt,
			EstimatedEndDate:     &obligation.DueDate,
			PercentageComplete:   0.0,
			RiskLevel:            "medium",
			CriticalityScore:     50,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		// Set estimated duration
		if obligation.DueDate.After(contract.CreatedAt) {
			milestone.EstimatedDuration = obligation.DueDate.Sub(contract.CreatedAt)
		}

		milestones = append(milestones, milestone)
	}

	// Save milestones to repository
	for _, milestone := range milestones {
		if err := s.milestoneRepo.CreateMilestone(ctx, milestone); err != nil {
			return nil, fmt.Errorf("failed to create milestone %s: %w", milestone.ID, err)
		}
	}

	return milestones, nil
}

// ResolveMilestoneDependencies automatically resolves dependencies between milestones
func (s *milestoneOrchestrationService) ResolveMilestoneDependencies(ctx context.Context, contractID string) error {
	// Get all milestones for the contract
	milestones, err := s.milestoneRepo.GetMilestonesByContract(ctx, contractID, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get milestones for contract %s: %w", contractID, err)
	}

	// For now, we'll create a simple sequential dependency chain
	// In a more complex implementation, this would use AI analysis or other logic
	for i := 1; i < len(milestones); i++ {
		dependency := &models.MilestoneDependency{
			ID:             fmt.Sprintf("dep-%s-%s", milestones[i].ID, milestones[i-1].ID),
			MilestoneID:    milestones[i].ID,
			DependsOnID:    milestones[i-1].ID,
			DependencyType: "prerequisite",
		}

		if err := s.milestoneRepo.CreateMilestoneDependency(ctx, dependency); err != nil {
			return fmt.Errorf("failed to create dependency for milestone %s: %w", milestones[i].ID, err)
		}
	}

	return nil
}

// OptimizeMilestoneSchedule optimizes the timeline for milestone execution
func (s *milestoneOrchestrationService) OptimizeMilestoneSchedule(ctx context.Context, contractID string) error {
	// Get all milestones for the contract
	milestones, err := s.milestoneRepo.GetMilestonesByContract(ctx, contractID, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get milestones for contract %s: %w", contractID, err)
	}

	// Get the dependency graph
	_, err = s.milestoneRepo.ResolveDependencyGraph(ctx, contractID)
	if err != nil {
		return fmt.Errorf("failed to resolve dependency graph for contract %s: %w", contractID, err)
	}

	// Get topological order
	order, err := s.milestoneRepo.GetTopologicalOrder(ctx, contractID)
	if err != nil {
		return fmt.Errorf("failed to get topological order for contract %s: %w", contractID, err)
	}

	// Update milestone dates based on dependencies and critical path analysis
	// This is a simplified implementation - a full implementation would use more sophisticated scheduling algorithms
	now := time.Now()
	for _, milestoneID := range order {
		for _, milestone := range milestones {
			if milestone.ID == milestoneID {
				// Set start date based on dependencies
				earliestStart := now
				dependencies, _ := s.milestoneRepo.GetMilestoneDependencies(ctx, milestoneID)
				for _, dep := range dependencies {
					depMilestone, err := s.milestoneRepo.GetMilestoneByID(ctx, dep.DependsOnID)
					if err == nil && depMilestone.ActualEndDate != nil {
						if depMilestone.ActualEndDate.After(earliestStart) {
							earliestStart = *depMilestone.ActualEndDate
						}
					} else if depMilestone.EstimatedEndDate != nil {
						if depMilestone.EstimatedEndDate.After(earliestStart) {
							earliestStart = *depMilestone.EstimatedEndDate
						}
					}
				}

				// Update milestone dates
				milestone.EstimatedStartDate = &earliestStart
				if milestone.EstimatedDuration > 0 {
					endDate := earliestStart.Add(milestone.EstimatedDuration)
					milestone.EstimatedEndDate = &endDate
				}

				milestone.UpdatedAt = time.Now()

				if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
					return fmt.Errorf("failed to update milestone %s: %w", milestone.ID, err)
				}

				break
			}
		}
	}

	return nil
}

// UpdateMilestoneProgress updates the progress of a milestone
func (s *milestoneOrchestrationService) UpdateMilestoneProgress(ctx context.Context, milestoneID string, progress float64, notes string) error {
	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Update progress
	milestone.PercentageComplete = progress
	milestone.UpdatedAt = time.Now()

	// If milestone is completed, set actual end date
	if progress >= 100.0 {
		now := time.Now()
		milestone.ActualEndDate = &now
	}

	// Update the milestone
	if err := s.milestoneRepo.UpdateMilestone(ctx, milestone); err != nil {
		return fmt.Errorf("failed to update milestone %s: %w", milestoneID, err)
	}

	// Create progress entry
	progressEntry := &repository.MilestoneProgressEntry{
		ID:                 fmt.Sprintf("pe-%s-%d", milestoneID, time.Now().Unix()),
		MilestoneID:        milestoneID,
		PercentageComplete: progress,
		Notes:              notes,
		RecordedBy:         "system",
		RecordedAt:         time.Now(),
		CreatedAt:          time.Now(),
	}

	if err := s.milestoneRepo.CreateMilestoneProgressEntry(ctx, progressEntry); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to create progress entry for milestone %s: %v\n", milestoneID, err)
	}

	// Send progress update notification
	if err := s.notificationSvc.SendProgressUpdate(ctx, milestone, progress); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to send progress update for milestone %s: %v\n", milestoneID, err)
	}

	// If milestone is completed, send completion notification
	if progress >= 100.0 {
		if err := s.notificationSvc.SendCompletionNotification(ctx, milestone); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to send completion notification for milestone %s: %v\n", milestoneID, err)
		}
	}

	return nil
}

// ValidateMilestoneCompletion validates if a milestone has been properly completed
func (s *milestoneOrchestrationService) ValidateMilestoneCompletion(ctx context.Context, milestoneID string) (bool, error) {
	// Get the milestone
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return false, fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	// Check if milestone is marked as completed
	if milestone.PercentageComplete < 100.0 {
		return false, nil
	}

	// In a real implementation, this would check verification criteria
	// For now, we'll just check if the percentage is 100
	return milestone.PercentageComplete >= 100.0, nil
}

// GetMilestoneTimeline returns the timeline analysis for a contract's milestones
func (s *milestoneOrchestrationService) GetMilestoneTimeline(ctx context.Context, contractID string) (*repository.MilestoneTimelineAnalysis, error) {
	return s.milestoneRepo.GetMilestoneTimelineAnalysis(ctx, contractID)
}

// GetMilestonePerformanceMetrics returns performance metrics for milestones
func (s *milestoneOrchestrationService) GetMilestonePerformanceMetrics(ctx context.Context, contractID *string) (*repository.MilestonePerformanceMetrics, error) {
	return s.milestoneRepo.GetMilestonePerformanceMetrics(ctx, contractID)
}

// GetMilestoneRiskAnalysis returns risk analysis for milestones
func (s *milestoneOrchestrationService) GetMilestoneRiskAnalysis(ctx context.Context, contractID *string) (*repository.MilestoneRiskAnalysis, error) {
	return s.milestoneRepo.GetMilestoneRiskAnalysis(ctx, contractID)
}
