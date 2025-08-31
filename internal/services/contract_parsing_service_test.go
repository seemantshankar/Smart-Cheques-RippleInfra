package services

import (
	"context"
	"testing"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestParseFromMetadata_FilenameMilestoneCreatesMilestoneTag(t *testing.T) {
	svc := NewContractParsingService()
	meta := &models.DocumentMetadata{OriginalFilename: "project_milestones_v1.pdf", FileSize: 1234, MimeType: "application/pdf"}
	c, err := svc.ParseFromMetadata(context.Background(), "c-123", meta, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(c.Obligations) == 0 {
		t.Fatalf("expected obligations to be extracted")
	}
	found := false
	for _, tag := range c.Tags {
		if tag == "milestone:ms-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected milestone tag to be present in Tags")
	}
}

func TestParseFromMetadata_NoFilenameCreatesObligation(t *testing.T) {
	svc := NewContractParsingService()
	meta := &models.DocumentMetadata{OriginalFilename: "", FileSize: 0, MimeType: "application/octet-stream"}
	c, err := svc.ParseFromMetadata(context.Background(), "c-456", meta, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(c.Obligations) != 1 {
		t.Fatalf("expected one obligation fallback, got %d", len(c.Obligations))
	}
}

