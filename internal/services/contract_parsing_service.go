package services

import (
	"context"
	"fmt"

	"github.com/smart-payment-infrastructure/internal/models"
)

// ContractParsingService defines a simple, deterministic parsing pipeline
// that extracts obligations and milestones from known document metadata.
// This is intentionally non-ML and serves as a deterministic stub until the
// LLM-based parser is implemented.
type ContractParsingService interface {
	// ParseFromMetadata creates a Contract model given document metadata and
	// optional hints. It extracts obligations and simple milestones deterministically.
	ParseFromMetadata(ctx context.Context, contractID string, meta *models.DocumentMetadata, hints map[string]string) (*models.Contract, error)
}

type contractParsingServiceImpl struct{}

func NewContractParsingService() ContractParsingService {
	return &contractParsingServiceImpl{}
}

// ParseFromMetadata implements a simple rule-based extraction. For example:
//   - if the filename contains 'milestone' or 'milestones' we create a single
//     milestone derived from filename; otherwise we create a generic obligation.
//
// This function is small, deterministic, and fully testable.
func (s *contractParsingServiceImpl) ParseFromMetadata(_ context.Context, contractID string, meta *models.DocumentMetadata, _ map[string]string) (*models.Contract, error) {
	if meta == nil {
		return nil, fmt.Errorf("metadata required")
	}

	c := &models.Contract{
		ID:        contractID,
		Parties:   []string{},
		Status:    "draft",
		CreatedAt: models.TimeNow(),
		UpdatedAt: models.TimeNow(),
	}

	// Simple heuristic: filename-driven milestone extraction
	fn := meta.OriginalFilename
	if fn == "" {
		// fallback: create a single obligation
		ob := models.Obligation{ID: "ob-1", Description: "Auto-extracted obligation", Status: "pending", Party: "unknown"}
		c.Obligations = []models.Obligation{ob}
		return c, nil
	}

	// If filename contains 'milestone' create a milestone entry
	if containsInsensitive(fn, "milestone") {
		m := models.ContractMilestone{
			ID:                   "ms-1",
			ContractID:           contractID,
			MilestoneID:          "m-1",
			SequenceOrder:        1,
			TriggerConditions:    "filename-hint",
			VerificationCriteria: "file-based",
			CreatedAt:            models.TimeNow(),
			UpdatedAt:            models.TimeNow(),
		}
		cob := models.Obligation{ID: "ob-1", Description: "Milestone based obligation", Status: "pending", Party: "unknown"}
		c.Obligations = []models.Obligation{cob}
		// store milestone in a separate slice via interface (contract model has PaymentTerms/Obligations only now)
		// we attach milestone data to Tags so callers can inspect it until DB model is extended
		c.Tags = append(c.Tags, "milestone:ms-1")
		_ = m
		return c, nil
	}

	// Default fallback: obligation from filename
	ob := models.Obligation{ID: "ob-1", Description: fmt.Sprintf("Auto-extracted from %s", fn), Status: "pending", Party: "unknown"}
	c.Obligations = []models.Obligation{ob}
	return c, nil
}

// containsInsensitive checks substring presence case-insensitively.
func containsInsensitive(s, substr string) bool {
	// simple implementation without extra imports to keep code small and testable
	// use basic lowercasing
	ls := makeLower(s)
	lsub := makeLower(substr)
	return contains(ls, lsub)
}

func makeLower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] = b[i] + ('a' - 'A')
		}
	}
	return string(b)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	// naive O(n*m) search sufficient for small strings in tests
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
