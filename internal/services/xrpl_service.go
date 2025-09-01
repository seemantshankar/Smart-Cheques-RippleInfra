package services

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

type XRPLService struct {
	client      *xrpl.Client
	initialized bool
}

// Verify that XRPLService implements repository.XRPLServiceInterface
var _ repository.XRPLServiceInterface = (*XRPLService)(nil)

type XRPLConfig struct {
	NetworkURL string
	TestNet    bool
}

func NewXRPLService(config XRPLConfig) *XRPLService {
	client := xrpl.NewClient(config.NetworkURL, config.TestNet)
	return &XRPLService{
		client: client,
	}
}

func (s *XRPLService) Initialize() error {
	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to XRPL: %w", err)
	}

	if err := s.client.HealthCheck(); err != nil {
		return fmt.Errorf("XRPL health check failed: %w", err)
	}

	s.initialized = true
	log.Println("XRPL service initialized successfully")
	return nil
}

func (s *XRPLService) CreateWallet() (*xrpl.WalletInfo, error) {
	wallet, err := s.client.GenerateWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	log.Printf("Created new XRPL wallet: %s", wallet.Address)
	return wallet, nil
}

func (s *XRPLService) ValidateAddress(address string) bool {
	return s.client.ValidateAddress(address)
}

func (s *XRPLService) GetAccountInfo(address string) (interface{}, error) {
	accountInfo, err := s.client.GetAccountInfo(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info for %s: %w", address, err)
	}

	return accountInfo, nil
}

func (s *XRPLService) HealthCheck() error {
	if !s.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}
	return s.client.HealthCheck()
}

// CreateSmartChequeEscrow creates an escrow for a Smart Check with basic milestone support
func (s *XRPLService) CreateSmartChequeEscrow(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string) (*xrpl.TransactionResult, string, error) {
	if !s.initialized {
		return nil, "", fmt.Errorf("XRPL service not initialized")
	}

	// Convert amount to drops (for XRP) or appropriate format
	amountStr := s.formatAmount(amount, currency)

	// Generate condition and fulfillment for milestone completion
	condition, fulfillment, err := s.client.GenerateCondition(milestoneSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate escrow condition: %w", err)
	}

	// Set escrow parameters
	escrow := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Destination: payeeAddress,
		Amount:      amountStr,
		Condition:   condition,
		// Set cancel after 30 days (approximate ledger time)
		CancelAfter: s.getLedgerTimeOffset(30 * 24 * time.Hour),
		// Allow finish after 1 hour minimum
		FinishAfter: s.getLedgerTimeOffset(1 * time.Hour),
	}

	// Create the escrow
	result, err := s.client.CreateEscrow(escrow)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create escrow: %w", err)
	}

	log.Printf("Smart Check escrow created: %s, Amount: %s %s", result.TransactionID, amountStr, currency)
	return result, fulfillment, nil
}

// CreateSmartChequeEscrowWithMilestones creates an escrow with milestone-based conditions
func (s *XRPLService) CreateSmartChequeEscrowWithMilestones(payerAddress, payeeAddress string, amount float64, currency string, milestones []models.Milestone) (*xrpl.TransactionResult, string, error) {
	if !s.initialized {
		return nil, "", fmt.Errorf("XRPL service not initialized")
	}

	// Convert amount to drops (for XRP) or appropriate format
	amountStr := s.formatAmount(amount, currency)

	// Convert milestones to XRPL milestone conditions
	xrplMilestones := make([]xrpl.MilestoneCondition, len(milestones))
	for i, milestone := range milestones {
		xrplMilestones[i] = xrpl.MilestoneCondition{
			MilestoneID:        milestone.ID,
			VerificationMethod: string(milestone.VerificationMethod),
			OracleConfig:       milestone.OracleConfig.Config,
			Amount:             s.formatAmount(milestone.Amount, currency),
		}
	}

	// Set escrow parameters with dynamic timing based on milestones
	escrow := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Destination: payeeAddress,
		Amount:      amountStr,
		// Set cancel after based on longest milestone duration
		CancelAfter: s.calculateCancelAfter(milestones),
		// Allow finish after 1 hour minimum
		FinishAfter: s.getLedgerTimeOffset(1 * time.Hour),
	}

	// Create the escrow with validated milestone conditions
	result, err := s.client.CreateConditionalEscrowWithValidation(escrow, xrplMilestones)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create validated escrow with milestones: %w", err)
	}

	// Generate fulfillment for the compound condition
	compoundSecret := s.client.GenerateCompoundSecret(xrplMilestones)
	_, fulfillment, err := s.client.GenerateCondition(compoundSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate fulfillment: %w", err)
	}

	log.Printf("Smart Check escrow with %d validated milestones created: %s, Amount: %s %s", len(milestones), result.TransactionID, amountStr, currency)
	return result, fulfillment, nil
}

// calculateCancelAfter calculates the cancel time based on milestone durations
func (s *XRPLService) calculateCancelAfter(milestones []models.Milestone) uint32 {
	// Default cancel after 30 days
	defaultDuration := 30 * 24 * time.Hour

	// If we have milestones with end dates, use the latest one plus buffer
	var latestEndDate *time.Time
	for _, milestone := range milestones {
		if milestone.EstimatedEndDate != nil {
			if latestEndDate == nil || milestone.EstimatedEndDate.After(*latestEndDate) {
				latestEndDate = milestone.EstimatedEndDate
			}
		}
	}

	if latestEndDate != nil {
		// Add 7 days buffer to the latest milestone end date
		cancelTime := latestEndDate.Add(7 * 24 * time.Hour)
		if cancelTime.After(time.Now()) {
			duration := time.Until(cancelTime)
			return s.getLedgerTimeOffset(duration)
		}
	}

	// Fall back to default
	return s.getLedgerTimeOffset(defaultDuration)
}

// CompleteSmartChequeMilestone releases funds for a completed milestone
func (s *XRPLService) CompleteSmartChequeMilestone(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	// Create escrow finish transaction
	finish := &xrpl.EscrowFinish{
		Account:       payeeAddress,
		Owner:         ownerAddress,
		OfferSequence: sequence,
		Condition:     condition,
		Fulfillment:   fulfillment,
	}

	// Finish the escrow
	result, err := s.client.FinishEscrow(finish)
	if err != nil {
		return nil, fmt.Errorf("failed to finish escrow: %w", err)
	}

	log.Printf("Smart Check milestone completed: %s, Sequence: %d", result.TransactionID, sequence)
	return result, nil
}

// CancelSmartCheque cancels a Smart Check escrow
func (s *XRPLService) CancelSmartCheque(accountAddress, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	// Create escrow cancel transaction
	cancel := &xrpl.EscrowCancel{
		Account:       accountAddress,
		Owner:         ownerAddress,
		OfferSequence: sequence,
	}

	// Cancel the escrow
	result, err := s.client.CancelEscrow(cancel)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel escrow: %w", err)
	}

	log.Printf("Smart Check canceled: %s, Sequence: %d", result.TransactionID, sequence)
	return result, nil
}

// GetEscrowStatus retrieves the current status of an escrow
func (s *XRPLService) GetEscrowStatus(ownerAddress string, sequence string) (*xrpl.EscrowInfo, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	escrowInfo, err := s.client.GetEscrowInfo(ownerAddress, sequence)
	if err != nil {
		return nil, fmt.Errorf("failed to get escrow info: %w", err)
	}

	return escrowInfo, nil
}

// formatAmount converts amount to appropriate format based on currency
func (s *XRPLService) formatAmount(amount float64, currency string) string {
	switch currency {
	case "XRP":
		// Convert XRP to drops (1 XRP = 1,000,000 drops)
		drops := int64(amount * 1000000)
		return strconv.FormatInt(drops, 10)
	default:
		// For other currencies (USDT, USDC, etc.), use the amount as-is
		// In production, this would handle currency-specific formatting
		return fmt.Sprintf("%.6f", amount)
	}
}

// getLedgerTimeOffset calculates ledger time offset (mock implementation)
func (s *XRPLService) getLedgerTimeOffset(duration time.Duration) uint32 {
	// XRPL uses seconds since January 1, 2000 (00:00 UTC) as "Ripple Epoch"
	// This is a simplified calculation for testing
	rippleEpoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	futureTime := time.Now().Add(duration)
	offset := futureTime.Sub(rippleEpoch).Seconds()
	return uint32(offset)
}

// GenerateCondition creates a cryptographic condition for escrow
func (s *XRPLService) GenerateCondition(secret string) (condition string, fulfillment string, err error) {
	if !s.initialized {
		return "", "", fmt.Errorf("XRPL service not initialized")
	}

	return s.client.GenerateCondition(secret)
}
