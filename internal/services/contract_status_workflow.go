package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

// ContractStatusWorkflowService manages deterministic contract status transitions.
// It encapsulates the allowed transitions and applies them to Contract models.
type ContractStatusWorkflowService interface {
	// CanTransition returns nil if a contract in 'from' may transition to 'to'.
	CanTransition(ctx context.Context, from, to string) error

	// TransitionContract applies the transition to the provided contract (in-memory)
	// and updates timestamps. It returns an error if transition is invalid.
	TransitionContract(ctx context.Context, contract *models.Contract, to string) error
}

type contractStatusWorkflowServiceImpl struct {
	// allowedTransitions maps a source status to allowed destination statuses.
	allowedTransitions map[string]map[string]struct{}
}

// NewContractStatusWorkflowService constructs a deterministic workflow service.
func NewContractStatusWorkflowService() ContractStatusWorkflowService {
	allowed := map[string]map[string]struct{}{
		"draft": {
			"active":     {},
			"terminated": {},
		},
		"active": {
			"executed":   {},
			"disputed":   {},
			"terminated": {},
		},
		"executed": {
			"disputed":   {},
			"terminated": {},
		},
		"disputed": {
			"executed":   {},
			"terminated": {},
		},
	}
	return &contractStatusWorkflowServiceImpl{allowedTransitions: allowed}
}

// CanTransition checks allowedTransitions for the pair.
func (s *contractStatusWorkflowServiceImpl) CanTransition(ctx context.Context, from, to string) error {
	if from == to {
		return fmt.Errorf("no-op transition: %s -> %s", from, to)
	}
	if dsts, ok := s.allowedTransitions[from]; ok {
		if _, ok2 := dsts[to]; ok2 {
			return nil
		}
		return fmt.Errorf("invalid transition: %s -> %s", from, to)
	}
	return fmt.Errorf("unknown source status: %s", from)
}

// TransitionContract applies the transition to the contract in-place.
func (s *contractStatusWorkflowServiceImpl) TransitionContract(ctx context.Context, contract *models.Contract, to string) error {
	if contract == nil {
		return fmt.Errorf("contract is nil")
	}
	if err := s.CanTransition(ctx, contract.Status, to); err != nil {
		return err
	}
	// update status and timestamps
	contract.Status = to
	now := time.Now()
	contract.UpdatedAt = now
	// if moving to executed, set expiration date to now + 1 year as an example
	if to == "executed" && contract.ExpirationDate == nil {
		exp := now.Add(365 * 24 * time.Hour)
		contract.ExpirationDate = &exp
	}
	return nil
}

