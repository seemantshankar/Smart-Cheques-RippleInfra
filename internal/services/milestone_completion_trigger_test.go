package services

import (
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

// Test wasRecentlyCompleted tests the logic for determining if a milestone was recently completed
func TestWasRecentlyCompleted(t *testing.T) {
	service := &milestoneCompletionTriggerService{}

	// Test milestone completed recently (within 60 seconds)
	recentMilestone := models.ContractMilestone{
		UpdatedAt: time.Now().Add(-30 * time.Second), // 30 seconds ago
	}

	if !service.wasRecentlyCompleted(recentMilestone) {
		t.Error("Expected milestone to be considered recently completed")
	}

	// Test milestone completed long ago (more than 60 seconds)
	oldMilestone := models.ContractMilestone{
		UpdatedAt: time.Now().Add(-120 * time.Second), // 2 minutes ago
	}

	if service.wasRecentlyCompleted(oldMilestone) {
		t.Error("Expected milestone to NOT be considered recently completed")
	}

	// Test milestone with zero time (never updated)
	zeroMilestone := models.ContractMilestone{
		UpdatedAt: time.Time{},
	}

	if service.wasRecentlyCompleted(zeroMilestone) {
		t.Error("Expected milestone with zero time to NOT be considered recently completed")
	}
}

// Test TriggerStatus tests the trigger status structure
func TestTriggerStatus(t *testing.T) {
	service := &milestoneCompletionTriggerService{
		isMonitoring:    true,
		startTime:       time.Now().Add(-5 * time.Minute),
		processedCount:  10,
		errorCount:      2,
		lastError:       "test error",
		lastProcessedAt: time.Now().Add(-1 * time.Minute),
	}

	status, err := service.GetTriggerStatus()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !status.IsMonitoring {
		t.Error("Expected IsMonitoring to be true")
	}

	if status.ProcessedCount != 10 {
		t.Errorf("Expected ProcessedCount to be 10, got %d", status.ProcessedCount)
	}

	if status.ErrorCount != 2 {
		t.Errorf("Expected ErrorCount to be 2, got %d", status.ErrorCount)
	}

	if status.LastError != "test error" {
		t.Errorf("Expected LastError to be 'test error', got '%s'", status.LastError)
	}
}
