package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// MilestoneCompletionTriggerServiceInterface defines the interface for milestone completion triggers
type MilestoneCompletionTriggerServiceInterface interface {
	// StartTriggerMonitoring starts monitoring for milestone completion events
	StartTriggerMonitoring(ctx context.Context) error

	// StopTriggerMonitoring stops the monitoring process
	StopTriggerMonitoring() error

	// ProcessMilestoneCompletion manually processes a milestone completion (for testing/backfill)
	ProcessMilestoneCompletion(ctx context.Context, milestoneID string) error

	// GetTriggerStatus returns the current status of the trigger monitoring system
	GetTriggerStatus() (*TriggerStatus, error)

	// PublishMilestoneCompletedEvent publishes an event when a milestone is completed
	PublishMilestoneCompletedEvent(ctx context.Context, milestoneID string) error
}

// TriggerStatus represents the status of the trigger monitoring system
type TriggerStatus struct {
	IsMonitoring    bool      `json:"is_monitoring"`
	LastProcessedAt time.Time `json:"last_processed_at"`
	ProcessedCount  int64     `json:"processed_count"`
	ErrorCount      int64     `json:"error_count"`
	LastError       string    `json:"last_error,omitempty"`
	StartTime       time.Time `json:"start_time"`
}

// milestoneCompletionTriggerService implements MilestoneCompletionTriggerServiceInterface
type milestoneCompletionTriggerService struct {
	milestoneRepo   repository.MilestoneRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
	eventBus        messaging.EventBus
	contractRepo    repository.ContractRepositoryInterface
	isMonitoring    bool
	lastProcessedAt time.Time
	processedCount  int64
	errorCount      int64
	lastError       string
	startTime       time.Time
	stopChan        chan struct{}
}

// NewMilestoneCompletionTriggerService creates a new milestone completion trigger service
func NewMilestoneCompletionTriggerService(
	milestoneRepo repository.MilestoneRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	contractRepo repository.ContractRepositoryInterface,
	eventBus messaging.EventBus,
) MilestoneCompletionTriggerServiceInterface {
	return &milestoneCompletionTriggerService{
		milestoneRepo:   milestoneRepo,
		smartChequeRepo: smartChequeRepo,
		contractRepo:    contractRepo,
		eventBus:        eventBus,
		stopChan:        make(chan struct{}),
	}
}

// StartTriggerMonitoring starts monitoring for milestone completion events
func (s *milestoneCompletionTriggerService) StartTriggerMonitoring(ctx context.Context) error {
	if s.isMonitoring {
		return fmt.Errorf("trigger monitoring is already running")
	}

	s.isMonitoring = true
	s.startTime = time.Now()
	s.processedCount = 0
	s.errorCount = 0
	s.lastError = ""

	log.Printf("Starting milestone completion trigger monitoring")

	// Subscribe to milestone status update events
	// Note: In a real implementation, we might need to hook into the milestone update process
	// For now, we'll implement a polling mechanism that checks for completed milestones
	go s.monitorMilestones(ctx)

	log.Printf("Milestone completion trigger monitoring started successfully")
	return nil
}

// StopTriggerMonitoring stops the monitoring process
func (s *milestoneCompletionTriggerService) StopTriggerMonitoring() error {
	if !s.isMonitoring {
		return fmt.Errorf("trigger monitoring is not running")
	}

	log.Printf("Stopping milestone completion trigger monitoring")
	s.isMonitoring = false
	close(s.stopChan)
	return nil
}

// monitorMilestones continuously monitors milestones for completion
func (s *milestoneCompletionTriggerService) monitorMilestones(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	log.Printf("Starting milestone monitoring loop")

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping milestone monitoring")
			return
		case <-s.stopChan:
			log.Printf("Stop signal received, stopping milestone monitoring")
			return
		case <-ticker.C:
			if err := s.checkForCompletedMilestones(ctx); err != nil {
				s.errorCount++
				s.lastError = err.Error()
				log.Printf("Error checking for completed milestones: %v", err)
			}
		}
	}
}

// checkForCompletedMilestones checks for newly completed milestones
func (s *milestoneCompletionTriggerService) checkForCompletedMilestones(ctx context.Context) error {
	// Get recently updated milestones that might be completed
	// Note: This is a simplified approach. In production, you might want to:
	// 1. Use database triggers
	// 2. Hook into the milestone update process
	// 3. Use event-driven architecture with proper event sourcing

	// Get milestones that were updated in the last monitoring interval
	// This is a placeholder - the actual implementation would depend on your repository interface
	milestones, err := s.milestoneRepo.GetMilestonesByStatus(ctx, string(models.MilestoneStatusVerified), 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get verified milestones: %w", err)
	}

	processed := 0
	for _, milestone := range milestones {
		// Check if this milestone was recently completed (within the last monitoring interval)
		if s.wasRecentlyCompleted(*milestone) {
			if err := s.processCompletedMilestone(ctx, milestone.ID); err != nil {
				log.Printf("Error processing completed milestone %s: %v", milestone.ID, err)
				s.errorCount++
				s.lastError = err.Error()
				continue
			}
			processed++
			s.processedCount++
		}
	}

	if processed > 0 {
		s.lastProcessedAt = time.Now()
		log.Printf("Processed %d completed milestones", processed)
	}

	return nil
}

// wasRecentlyCompleted checks if a milestone was completed recently
func (s *milestoneCompletionTriggerService) wasRecentlyCompleted(milestone models.ContractMilestone) bool {
	// Check if the milestone has a completion timestamp and it's recent
	if milestone.UpdatedAt.IsZero() {
		return false
	}

	// Consider it "recent" if it was updated within the last 2 monitoring intervals
	timeThreshold := time.Now().Add(-60 * time.Second)
	return milestone.UpdatedAt.After(timeThreshold)
}

// processCompletedMilestone processes a completed milestone
func (s *milestoneCompletionTriggerService) processCompletedMilestone(ctx context.Context, milestoneID string) error {
	log.Printf("Processing completed milestone: %s", milestoneID)

	// Publish milestone completed event
	if err := s.PublishMilestoneCompletedEvent(ctx, milestoneID); err != nil {
		return fmt.Errorf("failed to publish milestone completed event: %w", err)
	}

	log.Printf("Successfully processed completed milestone: %s", milestoneID)
	return nil
}

// ProcessMilestoneCompletion manually processes a milestone completion
func (s *milestoneCompletionTriggerService) ProcessMilestoneCompletion(ctx context.Context, milestoneID string) error {
	log.Printf("Manually processing milestone completion: %s", milestoneID)

	// Validate the milestone exists and is completed
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone %s: %w", milestoneID, err)
	}

	if milestone.Status != string(models.MilestoneStatusVerified) {
		return fmt.Errorf("milestone %s is not in verified status (current: %s)", milestoneID, milestone.Status)
	}

	// Process the completion
	return s.processCompletedMilestone(ctx, milestoneID)
}

// PublishMilestoneCompletedEvent publishes an event when a milestone is completed
func (s *milestoneCompletionTriggerService) PublishMilestoneCompletedEvent(ctx context.Context, milestoneID string) error {
	// Get milestone details for the event
	milestone, err := s.milestoneRepo.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to get milestone details: %w", err)
	}

	// Find associated SmartCheque
	smartCheque, err := s.smartChequeRepo.GetSmartChequesByMilestone(ctx, milestoneID)
	if err != nil {
		log.Printf("Warning: Failed to get SmartCheque for milestone %s: %v", milestoneID, err)
	}

	var smartChequeID string
	var amount float64
	if smartCheque != nil {
		smartChequeID = smartCheque.ID
		amount = smartCheque.Amount
	}

	// Create and publish the event
	event := messaging.NewMilestoneCompletedEvent(milestoneID, smartChequeID, amount)
	event.Data["contract_id"] = milestone.ContractID
	event.Data["milestone_description"] = milestone.TriggerConditions
	event.Data["verified_at"] = milestone.UpdatedAt.Format(time.RFC3339)

	if err := s.eventBus.PublishEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to publish milestone completed event: %w", err)
	}

	log.Printf("Published milestone completed event for milestone %s", milestoneID)
	return nil
}

// GetTriggerStatus returns the current status of the trigger monitoring system
func (s *milestoneCompletionTriggerService) GetTriggerStatus() (*TriggerStatus, error) {
	return &TriggerStatus{
		IsMonitoring:    s.isMonitoring,
		LastProcessedAt: s.lastProcessedAt,
		ProcessedCount:  s.processedCount,
		ErrorCount:      s.errorCount,
		LastError:       s.lastError,
		StartTime:       s.startTime,
	}, nil
}
