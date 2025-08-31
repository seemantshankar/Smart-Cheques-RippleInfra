package services

import (
	"context"
	"testing"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestValidateMilestone_RequiresContractAndID(t *testing.T) {
	svc := NewMilestoneService()
	m := &models.ContractMilestone{ContractID: "", MilestoneID: "", SequenceOrder: 0}
	if err := svc.ValidateMilestone(context.Background(), m); err == nil {
		t.Fatalf("expected validation error for missing fields")
	}
}

func TestCreateMilestone_SetsTimestamps(t *testing.T) {
	svc := NewMilestoneService()
	m := &models.ContractMilestone{ContractID: "c1", MilestoneID: "m1", SequenceOrder: 1}
	if err := svc.CreateMilestone(context.Background(), m); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps to be set")
	}
}
