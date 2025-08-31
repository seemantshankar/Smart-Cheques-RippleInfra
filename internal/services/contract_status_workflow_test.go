package services

import (
	"context"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestAllowedTransition_DraftToActive(t *testing.T) {
	svc := NewContractStatusWorkflowService()
	c := &models.Contract{ID: "c1", Status: "draft", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := svc.TransitionContract(context.Background(), c, "active"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if c.Status != "active" {
		t.Fatalf("expected status active, got %s", c.Status)
	}
}

func TestInvalidTransition_ActiveToDraft(t *testing.T) {
	svc := NewContractStatusWorkflowService()
	c := &models.Contract{ID: "c2", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := svc.TransitionContract(context.Background(), c, "draft"); err == nil {
		t.Fatalf("expected error for invalid transition active->draft")
	}
}

