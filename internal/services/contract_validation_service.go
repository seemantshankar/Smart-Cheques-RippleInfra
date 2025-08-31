package services

import (
    "context"
    "fmt"

    "github.com/smart-payment-infrastructure/internal/models"
)

// ContractValidationService defines validation and parsing operations for contracts.
// This is a focused, testable service that will later be extended to call AI/LLM
// based parsers and richer validation rules.
type ContractValidationService interface {
    // ValidateContract performs domain-level validations on a Contract object.
    ValidateContract(ctx context.Context, contract *models.Contract) error

    // ParseContractFromStorage retrieves the contract document and parses it into
    // a Contract model. Parsing pipeline is intentionally minimal for now.
    ParseContractFromStorage(ctx context.Context, contractID string) (*models.Contract, error)
}

type contractValidationServiceImpl struct {
    storage   ContractStorageService
    contracts ContractServiceInterface
}

// NewContractValidationService constructs a new validation service instance.
func NewContractValidationService(storage ContractStorageService, contracts ContractServiceInterface) ContractValidationService {
    return &contractValidationServiceImpl{storage: storage, contracts: contracts}
}

// ValidateContract implements basic domain validations. Rules here should be
// conservative and fail-fast; richer checks (dates, amounts, jurisdictional
// rules) will be added incrementally.
func (s *contractValidationServiceImpl) ValidateContract(ctx context.Context, contract *models.Contract) error {
    if contract == nil {
        return fmt.Errorf("contract is nil")
    }
    if len(contract.Parties) < 2 {
        return fmt.Errorf("contract must include at least two parties")
    }
    if len(contract.PaymentTerms) == 0 && len(contract.Obligations) == 0 {
        return fmt.Errorf("contract must include obligations or payment terms")
    }
    // Ensure status has a sensible default
    if contract.Status == "" {
        contract.Status = "draft"
    }
    return nil
}

// ParseContractFromStorage is a placeholder for the parsing pipeline. It
// retrieves the stored document for the given contract and returns an error
// indicating the parsing pipeline is not yet implemented. The retrieval step
// is performed to validate integration points.
func (s *contractValidationServiceImpl) ParseContractFromStorage(ctx context.Context, contractID string) (*models.Contract, error) {
    rc, _, err := s.storage.Retrieve(ctx, contractID)
    if err != nil {
        return nil, fmt.Errorf("retrieve document: %w", err)
    }
    _ = rc
    // NOTE: actual parsing (OCR, NLP extraction) will be implemented in the
    // parsing pipeline task. For now, return a clear, actionable error so
    // callers know this feature is intentionally not available yet.
    return nil, fmt.Errorf("parsing pipeline not implemented")
}



