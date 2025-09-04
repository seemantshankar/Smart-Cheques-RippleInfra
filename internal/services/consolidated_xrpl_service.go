package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

// ConsolidatedXRPLService provides a unified interface for XRPL operations
// using the enhanced client implementation
type ConsolidatedXRPLService struct {
	client      *xrpl.EnhancedClient
	initialized bool
}

// ConsolidatedXRPLConfig holds configuration for XRPL service
type ConsolidatedXRPLConfig struct {
	NetworkURL   string
	WebSocketURL string
	TestNet      bool
}

// NewConsolidatedXRPLService creates a new consolidated XRPL service
func NewConsolidatedXRPLService(config ConsolidatedXRPLConfig) *ConsolidatedXRPLService {
	client := xrpl.NewEnhancedClient(config.NetworkURL, config.WebSocketURL, config.TestNet)
	return &ConsolidatedXRPLService{
		client: client,
	}
}

// Initialize initializes the XRPL service
func (s *ConsolidatedXRPLService) Initialize() error {
	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to XRPL: %w", err)
	}

	if err := s.client.HealthCheck(); err != nil {
		return fmt.Errorf("XRPL health check failed: %w", err)
	}

	s.initialized = true
	log.Println("Consolidated XRPL service initialized successfully")
	return nil
}

// CreateWallet creates a new XRPL wallet
func (s *ConsolidatedXRPLService) CreateWallet() (*xrpl.WalletInfo, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	wallet, err := s.client.GenerateWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	log.Printf("Created new XRPL wallet: %s", wallet.Address)
	return wallet, nil
}

// CreateSecp256k1Wallet creates a new XRPL wallet using secp256k1
func (s *ConsolidatedXRPLService) CreateSecp256k1Wallet() (*xrpl.WalletInfo, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	wallet, err := s.client.GenerateSecp256k1Wallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create secp256k1 wallet: %w", err)
	}

	log.Printf("Created new XRPL secp256k1 wallet: %s", wallet.Address)
	return wallet, nil
}

// CreateAccount creates a new XRPL account and funds it using the testnet faucet
func (s *ConsolidatedXRPLService) CreateAccount() (*xrpl.WalletInfo, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	wallet, err := s.client.CreateAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	log.Printf("Created new XRPL account: %s", wallet.Address)
	return wallet, nil
}

// ValidateAddress validates an XRPL address
func (s *ConsolidatedXRPLService) ValidateAddress(address string) bool {
	return s.client.ValidateAddress(address)
}

// GetAccountInfo retrieves account information from XRPL
func (s *ConsolidatedXRPLService) GetAccountInfo(address string) (interface{}, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	accountInfo, err := s.client.GetAccountInfo(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info for %s: %w", address, err)
	}

	return accountInfo, nil
}

// GetAccountData retrieves structured account data from XRPL
func (s *ConsolidatedXRPLService) GetAccountData(address string) (*xrpl.AccountData, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	accountData, err := s.client.GetAccountData(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account data for %s: %w", address, err)
	}

	return accountData, nil
}

// HealthCheck performs a health check on the XRPL service
func (s *ConsolidatedXRPLService) HealthCheck() error {
	if !s.initialized {
		return fmt.Errorf("service not initialized")
	}
	return s.client.HealthCheck()
}

// CreateSmartChequeEscrow creates an escrow for a Smart Check with basic milestone support
func (s *ConsolidatedXRPLService) CreateSmartChequeEscrow(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string) (*xrpl.TransactionResult, string, error) {
	if !s.initialized {
		return nil, "", fmt.Errorf("service not initialized")
	}

	// This method requires a private key, so we'll use the WithKey version
	// For now, return an error indicating that the WithKey version should be used
	return nil, "", fmt.Errorf("CreateSmartChequeEscrow requires a private key, use CreateSmartChequeEscrowWithKey instead")
}

// CreateSmartChequeEscrowWithMilestones creates an escrow with milestone-based conditions
func (s *ConsolidatedXRPLService) CreateSmartChequeEscrowWithMilestones(payerAddress, payeeAddress string, amount float64, currency string, milestones []models.Milestone) (*xrpl.TransactionResult, string, error) {
	if !s.initialized {
		return nil, "", fmt.Errorf("service not initialized")
	}
	amountStr := s.formatAmount(amount, currency)
	xrplMilestones := make([]xrpl.MilestoneCondition, len(milestones))
	for i, milestone := range milestones {
		oracleConfigStr := ""
		if milestone.OracleConfig != nil && milestone.OracleConfig.Config != nil {
			if configBytes, err := json.Marshal(milestone.OracleConfig.Config); err == nil {
				oracleConfigStr = string(configBytes)
			}
		}
		xrplMilestones[i] = xrpl.MilestoneCondition{
			MilestoneID:        milestone.ID,
			VerificationMethod: string(milestone.VerificationMethod),
			OracleConfig:       oracleConfigStr,
			Amount:             s.formatAmount(milestone.Amount, currency),
		}
	}
	escrow := &xrpl.EscrowCreate{
		Account:     payerAddress,
		Destination: payeeAddress,
		Amount:      amountStr,
		CancelAfter: s.calculateCancelAfter(milestones),
		FinishAfter: s.getLedgerTimeOffset(1 * time.Hour),
	}
	result, err := s.client.CreateConditionalEscrowWithValidation(escrow, xrplMilestones)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create validated escrow with milestones: %w", err)
	}
	compoundSecret := s.client.GenerateCompoundSecret(xrplMilestones)
	_, fulfillment, err := s.client.GenerateCondition(compoundSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate fulfillment: %w", err)
	}
	log.Printf("Smart Check escrow with %d validated milestones created: %s, Amount: %s %s", len(milestones), result.TransactionID, amountStr, currency)
	return result, fulfillment, nil
}

// CompleteSmartChequeMilestone releases funds for a completed milestone
func (s *ConsolidatedXRPLService) CompleteSmartChequeMilestone(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	// This method requires a private key, so we'll use the WithKey version
	// For now, return an error indicating that the WithKey version should be used
	return nil, fmt.Errorf("CompleteSmartChequeMilestone requires a private key, use CompleteSmartChequeMilestoneWithKey instead")
}

// CancelSmartChequeEscrow cancels an escrow and returns funds to the payer
func (s *ConsolidatedXRPLService) CancelSmartChequeEscrow(payerWallet *xrpl.WalletInfo, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	if payerWallet == nil || payerWallet.PrivateKey == "" {
		return nil, fmt.Errorf("payer wallet must have a private key")
	}

	// Create escrow cancel transaction
	cancel := &xrpl.EscrowCancel{
		Account:       payerWallet.Address,
		Owner:         ownerAddress,
		OfferSequence: sequence,
	}

	// Cancel the escrow using the enhanced client
	result, err := s.client.CancelEscrow(cancel, payerWallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel escrow: %w", err)
	}

	log.Printf("Smart Check escrow cancelled: %s", result.TransactionID)
	return result, nil
}

// SubmitPayment submits a payment transaction to the XRPL network
func (s *ConsolidatedXRPLService) SubmitPayment(senderWallet *xrpl.WalletInfo, recipientAddress string, amount float64, currency string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	if senderWallet == nil || senderWallet.PrivateKey == "" {
		return nil, fmt.Errorf("sender wallet must have a private key")
	}

	// Convert amount to drops (for XRP) or appropriate format
	amountStr := s.formatAmount(amount, currency)

	// Submit the payment using the enhanced client
	result, err := s.client.CreatePaymentTransaction(senderWallet.Address, recipientAddress, amountStr, senderWallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to submit payment: %w", err)
	}

	log.Printf("Payment submitted successfully: %s, Amount: %s %s", result.TransactionID, amountStr, currency)
	return result, nil
}

// MonitorTransaction monitors the status of a submitted transaction
func (s *ConsolidatedXRPLService) MonitorTransaction(transactionID string, maxRetries int, retryInterval time.Duration) (*xrpl.TransactionStatus, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	return s.client.MonitorTransaction(transactionID, maxRetries, retryInterval)
}

// formatAmount converts amount to XRPL format (drops for XRP)
func (s *ConsolidatedXRPLService) formatAmount(amount float64, currency string) string {
	if currency == "XRP" {
		// Convert XRP to drops (1 XRP = 1,000,000 drops)
		drops := int64(amount * 1000000)
		return strconv.FormatInt(drops, 10)
	}
	// For other currencies, return as string (assuming proper format)
	return strconv.FormatFloat(amount, 'f', -1, 64)
}

// getLedgerTimeOffset converts duration to ledger time offset
func (s *ConsolidatedXRPLService) getLedgerTimeOffset(duration time.Duration) uint32 {
	// Approximate: 1 ledger = 3.5 seconds
	ledgers := int(duration.Seconds() / 3.5)
	if ledgers < 1 {
		ledgers = 1
	}
	return uint32(ledgers)
}

// calculateCancelAfter calculates the CancelAfter ledger time for a Smart Check
func (s *ConsolidatedXRPLService) calculateCancelAfter(milestones []models.Milestone) uint32 {
	if len(milestones) == 0 {
		return s.getLedgerTimeOffset(24 * time.Hour) // Default to 24 hours if no milestones
	}

	// Find the maximum duration among milestones
	var maxDuration time.Duration
	for _, milestone := range milestones {
		if milestone.EstimatedDuration > maxDuration {
			maxDuration = milestone.EstimatedDuration
		}
	}

	return s.getLedgerTimeOffset(maxDuration + 24*time.Hour)
}

// GetAccountBalance retrieves the account balance from XRPL
func (s *ConsolidatedXRPLService) GetAccountBalance(address string) (string, error) {
	if !s.initialized {
		return "", fmt.Errorf("service not initialized")
	}

	accountData, err := s.client.GetAccountData(address)
	if err != nil {
		return "", fmt.Errorf("failed to get account balance for %s: %w", address, err)
	}

	if accountData.Balance == "" {
		return "0", nil
	}

	return accountData.Balance, nil
}

// ValidateAccountOnNetwork validates if an account exists on the XRPL network
func (s *ConsolidatedXRPLService) ValidateAccountOnNetwork(address string) (bool, error) {
	if !s.initialized {
		return false, fmt.Errorf("service not initialized")
	}

	_, err := s.client.GetAccountInfo(address)
	if err != nil {
		return false, nil // Account doesn't exist
	}

	return true, nil
}

// ValidateAccountWithBalance validates if an account has sufficient balance
func (s *ConsolidatedXRPLService) ValidateAccountWithBalance(address string, minBalanceDrops int64) (bool, error) {
	if !s.initialized {
		return false, fmt.Errorf("service not initialized")
	}

	balanceStr, err := s.GetAccountBalance(address)
	if err != nil {
		return false, fmt.Errorf("failed to get account balance: %w", err)
	}

	balance, err := strconv.ParseInt(balanceStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance >= minBalanceDrops, nil
}

// CreateSmartChequeEscrowWithKey creates an escrow with private key for milestone completion
func (s *ConsolidatedXRPLService) CreateSmartChequeEscrowWithKey(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string, privateKeyHex string) (*xrpl.TransactionResult, string, error) {
	if !s.initialized {
		return nil, "", fmt.Errorf("service not initialized")
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

	// Create the escrow using the enhanced client
	result, err := s.client.CreateEscrow(escrow, privateKeyHex)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create escrow: %w", err)
	}

	log.Printf("Smart Check escrow created with key: %s, Amount: %s %s", result.TransactionID, amountStr, currency)
	return result, fulfillment, nil
}

// CompleteSmartChequeMilestoneWithKey completes a milestone with private key
func (s *ConsolidatedXRPLService) CompleteSmartChequeMilestoneWithKey(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string, privateKeyHex string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	// Create escrow finish transaction
	finish := &xrpl.EscrowFinish{
		Account:       payeeAddress,
		Owner:         ownerAddress,
		OfferSequence: sequence,
		Condition:     condition,
		Fulfillment:   fulfillment,
	}

	// Finish the escrow using the enhanced client
	result, err := s.client.FinishEscrow(finish, privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to finish escrow: %w", err)
	}

	log.Printf("Smart Check milestone completed with key: %s", result.TransactionID)
	return result, nil
}

// CancelSmartCheque cancels an escrow and returns funds to the payer
func (s *ConsolidatedXRPLService) CancelSmartCheque(accountAddress, ownerAddress string, sequence uint32) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	// Create escrow cancel transaction
	cancel := &xrpl.EscrowCancel{
		Account:       accountAddress,
		Owner:         ownerAddress,
		OfferSequence: sequence,
	}

	// Cancel the escrow using the enhanced client
	result, err := s.client.CancelEscrow(cancel, "")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel escrow: %w", err)
	}

	log.Printf("Smart Check escrow cancelled: %s", result.TransactionID)
	return result, nil
}

// CancelSmartChequeWithKey cancels an escrow with private key
func (s *ConsolidatedXRPLService) CancelSmartChequeWithKey(accountAddress, ownerAddress string, sequence uint32, privateKeyHex string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	// Cancel the escrow using the enhanced client
	result, err := s.client.CancelEscrow(&xrpl.EscrowCancel{
		Account:       accountAddress,
		Owner:         ownerAddress,
		OfferSequence: sequence,
	}, privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel escrow: %w", err)
	}

	log.Printf("Smart Check escrow cancelled with key: %s", result.TransactionID)
	return result, nil
}

// GetEscrowStatus retrieves escrow status from XRPL ledger
func (s *ConsolidatedXRPLService) GetEscrowStatus(ownerAddress string, sequence string) (*xrpl.EscrowInfo, error) {
	if !s.initialized {
		return nil, fmt.Errorf("service not initialized")
	}

	return s.client.GetEscrowStatus(ownerAddress, sequence)
}

// GenerateCondition generates escrow condition and fulfillment
func (s *ConsolidatedXRPLService) GenerateCondition(secret string) (string, string, error) {
	if !s.initialized {
		return "", "", fmt.Errorf("service not initialized")
	}

	return s.client.GenerateCondition(secret)
}
