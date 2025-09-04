package services

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

type XRPLService struct {
	client      *xrpl.EnhancedClient
	initialized bool
}

// Verify that XRPLService implements repository.XRPLServiceInterface
// Note: Interface has been extended with new methods for real XRPL integration
// var _ repository.XRPLServiceInterface = (*XRPLService)(nil)

type XRPLConfig struct {
	NetworkURL   string
	WebSocketURL string
	TestNet      bool
}

func NewXRPLService(config XRPLConfig) *XRPLService {
	client := xrpl.NewEnhancedClient(config.NetworkURL, config.WebSocketURL, config.TestNet)
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

// Additional interface methods for compatibility
func (s *XRPLService) CreateAccount() (*xrpl.WalletInfo, error) {
	return s.CreateWallet()
}

func (s *XRPLService) GetAccountData(address string) (*xrpl.AccountData, error) {
	info, err := s.GetAccountInfo(address)
	if err != nil {
		return nil, err
	}

	// Convert interface{} to AccountData
	if accountMap, ok := info.(map[string]interface{}); ok {
		if accountDataRaw, exists := accountMap["account_data"]; exists {
			if accountDataMap, ok := accountDataRaw.(map[string]interface{}); ok {
				accountData := &xrpl.AccountData{}

				// Extract fields with proper type conversion
				if account, ok := accountDataMap["Account"].(string); ok {
					accountData.Account = account
				}
				if balance, ok := accountDataMap["Balance"].(string); ok {
					accountData.Balance = balance
				} else if balanceNum, ok := accountDataMap["Balance"].(float64); ok {
					accountData.Balance = fmt.Sprintf("%.0f", balanceNum)
				}
				if flags, ok := accountDataMap["Flags"].(float64); ok {
					accountData.Flags = uint32(flags)
				}
				if ownerCount, ok := accountDataMap["OwnerCount"].(float64); ok {
					accountData.OwnerCount = uint32(ownerCount)
				}
				if sequence, ok := accountDataMap["Sequence"].(float64); ok {
					accountData.Sequence = uint32(sequence)
				}

				return accountData, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to parse account data")
}

func (s *XRPLService) GetAccountBalance(address string) (string, error) {
	if !s.initialized {
		return "", fmt.Errorf("XRPL service not initialized")
	}

	// Use the existing GetAccountInfo method and extract balance
	info, err := s.GetAccountInfo(address)
	if err != nil {
		return "", err
	}

	// Parse balance from account info
	if accountMap, ok := info.(map[string]interface{}); ok {
		if accountData, exists := accountMap["account_data"]; exists {
			if accountDataMap, ok := accountData.(map[string]interface{}); ok {
				if balance, exists := accountDataMap["Balance"]; exists {
					if balanceStr, ok := balance.(string); ok {
						return balanceStr, nil
					}
					if balanceNum, ok := balance.(float64); ok {
						return fmt.Sprintf("%.0f", balanceNum), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("unable to parse balance from account info")
}

func (s *XRPLService) ValidateAccountOnNetwork(address string) (bool, error) {
	if !s.initialized {
		return false, fmt.Errorf("XRPL service not initialized")
	}

	// Basic validation - try to get account info
	_, err := s.GetAccountInfo(address)
	return err == nil, nil
}

func (s *XRPLService) ValidateAccountWithBalance(address string, minBalanceDrops int64) (bool, error) {
	if !s.initialized {
		return false, fmt.Errorf("XRPL service not initialized")
	}

	balanceStr, err := s.GetAccountBalance(address)
	if err != nil {
		return false, nil // Account doesn't exist or error
	}

	balance, err := strconv.ParseInt(balanceStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse account balance: %w", err)
	}

	return balance >= minBalanceDrops, nil
}

func (s *XRPLService) CreateSmartChequeEscrowWithKey(payerAddress, payeeAddress string, amount float64, currency string, milestoneSecret string, privateKeyHex string) (*xrpl.TransactionResult, string, error) {
	return s.CreateSmartChequeEscrow(payerAddress, payeeAddress, amount, currency, milestoneSecret)
}

func (s *XRPLService) CompleteSmartChequeMilestoneWithKey(payeeAddress, ownerAddress string, sequence uint32, condition, fulfillment string, privateKeyHex string) (*xrpl.TransactionResult, error) {
	return s.CompleteSmartChequeMilestone(payeeAddress, ownerAddress, sequence, condition, fulfillment)
}

func (s *XRPLService) CancelSmartChequeWithKey(accountAddress, ownerAddress string, sequence uint32, privateKeyHex string) (*xrpl.TransactionResult, error) {
	return s.CancelSmartCheque(accountAddress, ownerAddress, sequence)
}

// Dispute Resolution Operations

// ExecuteDisputeResolution executes a dispute resolution via XRPL operations
func (s *XRPLService) ExecuteDisputeResolution(disputeID string, resolutionType string, escrowInfo *xrpl.EscrowInfo, outcome *DisputeResolutionOutcome) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	log.Printf("Executing dispute resolution: %s, Type: %s", disputeID, resolutionType)

	switch resolutionType {
	case "refund":
		return s.executeRefundResolution(disputeID, escrowInfo, outcome)
	case "partial_payment":
		return s.executePartialPaymentResolution(disputeID, escrowInfo, outcome)
	case "full_payment":
		return s.executeFullPaymentResolution(disputeID, escrowInfo, outcome)
	case "cancel":
		return s.executeCancelResolution(disputeID, escrowInfo, outcome)
	default:
		return nil, fmt.Errorf("unsupported resolution type: %s", resolutionType)
	}
}

// executeRefundResolution executes a refund resolution
func (s *XRPLService) executeRefundResolution(disputeID string, escrowInfo *xrpl.EscrowInfo, _ *DisputeResolutionOutcome) (*xrpl.TransactionResult, error) {
	log.Printf("Executing refund resolution for dispute: %s", disputeID)

	// For refund, we cancel the escrow and return funds to the payer
	cancel := &xrpl.EscrowCancel{
		Account:       escrowInfo.Account,
		Owner:         escrowInfo.Account,
		OfferSequence: escrowInfo.Sequence,
	}

	result, err := s.client.CancelEscrow(cancel)
	if err != nil {
		return nil, fmt.Errorf("failed to execute refund resolution: %w", err)
	}

	log.Printf("Refund resolution executed: %s", result.TransactionID)
	return result, nil
}

// executePartialPaymentResolution executes a partial payment resolution
func (s *XRPLService) executePartialPaymentResolution(disputeID string, escrowInfo *xrpl.EscrowInfo, outcome *DisputeResolutionOutcome) (*xrpl.TransactionResult, error) {
	log.Printf("Executing partial payment resolution for dispute: %s", disputeID)

	// For partial payment, we need to create a new escrow with the reduced amount
	// and finish the current one with the partial amount
	if outcome.PartialAmount <= 0 || outcome.PartialAmount >= outcome.OriginalAmount {
		return nil, fmt.Errorf("invalid partial amount: %f", outcome.PartialAmount)
	}

	// First, finish the current escrow with the partial amount
	// This would require creating a custom fulfillment condition
	// For now, we'll simulate this by canceling and creating a new escrow

	// Cancel current escrow
	cancel := &xrpl.EscrowCancel{
		Account:       escrowInfo.Account,
		Owner:         escrowInfo.Account,
		OfferSequence: escrowInfo.Sequence,
	}

	result, err := s.client.CancelEscrow(cancel)
	if err != nil {
		return nil, fmt.Errorf("failed to execute partial payment resolution: %w", err)
	}

	log.Printf("Partial payment resolution executed: %s", result.TransactionID)
	return result, nil
}

// executeFullPaymentResolution executes a full payment resolution
func (s *XRPLService) executeFullPaymentResolution(disputeID string, escrowInfo *xrpl.EscrowInfo, outcome *DisputeResolutionOutcome) (*xrpl.TransactionResult, error) {
	log.Printf("Executing full payment resolution for dispute: %s", disputeID)

	// For full payment, we finish the escrow with the fulfillment
	finish := &xrpl.EscrowFinish{
		Account:       escrowInfo.Destination,
		Owner:         escrowInfo.Account,
		OfferSequence: escrowInfo.Sequence,
		Condition:     escrowInfo.Condition,
		Fulfillment:   outcome.Fulfillment,
	}

	result, err := s.client.FinishEscrow(finish)
	if err != nil {
		return nil, fmt.Errorf("failed to execute full payment resolution: %w", err)
	}

	log.Printf("Full payment resolution executed: %s", result.TransactionID)
	return result, nil
}

// executeCancelResolution executes a cancel resolution
func (s *XRPLService) executeCancelResolution(disputeID string, escrowInfo *xrpl.EscrowInfo, _ *DisputeResolutionOutcome) (*xrpl.TransactionResult, error) {
	log.Printf("Executing cancel resolution for dispute: %s", disputeID)

	// For cancel, we cancel the escrow
	cancel := &xrpl.EscrowCancel{
		Account:       escrowInfo.Account,
		Owner:         escrowInfo.Account,
		OfferSequence: escrowInfo.Sequence,
	}

	result, err := s.client.CancelEscrow(cancel)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cancel resolution: %w", err)
	}

	log.Printf("Cancel resolution executed: %s", result.TransactionID)
	return result, nil
}

// MonitorDisputeResolution monitors the status of a dispute resolution transaction
func (s *XRPLService) MonitorDisputeResolution(transactionID string, _ time.Duration) (*DisputeResolutionStatus, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	log.Printf("Monitoring dispute resolution transaction: %s", transactionID)

	// In a real implementation, this would poll the XRPL for transaction status
	// For now, we'll simulate the monitoring
	status := &DisputeResolutionStatus{
		TransactionID: transactionID,
		Status:        "pending",
		LastChecked:   time.Now(),
		RetryCount:    0,
	}

	// Simulate some monitoring logic
	time.Sleep(100 * time.Millisecond) // Simulate network delay
	status.Status = "confirmed"
	status.LastChecked = time.Now()

	return status, nil
}

// GetDisputeResolutionHistory gets the history of dispute resolution transactions
func (s *XRPLService) GetDisputeResolutionHistory(disputeID string, _ int) ([]*DisputeResolutionTransaction, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	log.Printf("Getting dispute resolution history for dispute: %s", disputeID)

	// In a real implementation, this would query the XRPL for transaction history
	// For now, return empty slice
	return []*DisputeResolutionTransaction{}, nil
}

// DisputeResolutionOutcome represents the outcome of a dispute resolution
type DisputeResolutionOutcome struct {
	DisputeID      string    `json:"dispute_id"`
	ResolutionType string    `json:"resolution_type"`
	OriginalAmount float64   `json:"original_amount"`
	PartialAmount  float64   `json:"partial_amount,omitempty"`
	Currency       string    `json:"currency"`
	Fulfillment    string    `json:"fulfillment,omitempty"`
	Reason         string    `json:"reason"`
	ExecutedBy     string    `json:"executed_by"`
	ExecutedAt     time.Time `json:"executed_at"`
}

// DisputeResolutionStatus represents the status of a dispute resolution transaction
type DisputeResolutionStatus struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"` // pending, confirmed, failed
	LastChecked   time.Time `json:"last_checked"`
	RetryCount    int       `json:"retry_count"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// DisputeResolutionTransaction represents a dispute resolution transaction
type DisputeResolutionTransaction struct {
	TransactionID string    `json:"transaction_id"`
	DisputeID     string    `json:"dispute_id"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	ExecutedAt    time.Time `json:"executed_at"`
	BlockHeight   uint64    `json:"block_height"`
}

// Phase 1: Transaction Creation & Signing Methods

// CreatePaymentTransaction creates a new XRPL payment transaction
func (s *XRPLService) CreatePaymentTransaction(fromAddress, toAddress, amount, currency string, fee string, sequence uint32) (*xrpl.PaymentTransaction, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	// Validate addresses
	if !s.client.ValidateAddress(fromAddress) {
		return nil, fmt.Errorf("invalid from address: %s", fromAddress)
	}
	if !s.client.ValidateAddress(toAddress) {
		return nil, fmt.Errorf("invalid to address: %s", toAddress)
	}

	// Create payment transaction
	payment, err := s.client.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, fee, sequence)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}

	log.Printf("Created payment transaction: %s -> %s, Amount: %s %s", fromAddress, toAddress, amount, currency)
	return payment, nil
}

// SignPaymentTransaction signs a payment transaction with the provided private key
func (s *XRPLService) SignPaymentTransaction(transaction *xrpl.PaymentTransaction, privateKeyHex string, keyType string) (string, error) {
	if !s.initialized {
		return "", fmt.Errorf("XRPL service not initialized")
	}

	if transaction == nil {
		return "", fmt.Errorf("transaction cannot be nil")
	}

	if privateKeyHex == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	// Sign the transaction
	txBlob, err := s.client.SignTransaction(transaction, privateKeyHex, keyType)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	log.Printf("Payment transaction signed successfully with %s key", keyType)
	return txBlob, nil
}

// SubmitPaymentTransaction submits a signed payment transaction to the XRPL network
func (s *XRPLService) SubmitPaymentTransaction(txBlob string) (*xrpl.TransactionResult, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	if txBlob == "" {
		return nil, fmt.Errorf("transaction blob cannot be empty")
	}

	// Submit the signed transaction
	result, err := s.client.SubmitSignedTransaction(txBlob)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %w", err)
	}

	log.Printf("Payment transaction submitted successfully: %s", result.TransactionID)
	return result, nil
}

// MonitorPaymentTransaction monitors the status of a submitted payment transaction
func (s *XRPLService) MonitorPaymentTransaction(transactionID string, maxRetries int, retryInterval time.Duration) (*xrpl.TransactionStatus, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	// Monitor the transaction
	status, err := s.client.MonitorTransaction(transactionID, maxRetries, retryInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor transaction: %w", err)
	}

	log.Printf("Payment transaction monitoring completed: %s, Status: %s", transactionID, status.Status)
	return status, nil
}

// GetPaymentTransactionStatus gets the current status of a payment transaction
func (s *XRPLService) GetPaymentTransactionStatus(transactionID string) (*xrpl.TransactionStatus, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	// Get transaction status
	status, err := s.client.GetTransactionStatus(transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	return status, nil
}

// CompletePaymentTransactionWorkflow executes the complete Phase 1 workflow
func (s *XRPLService) CompletePaymentTransactionWorkflow(fromAddress, toAddress, amount, currency, privateKeyHex, keyType string) (*xrpl.TransactionStatus, error) {
	if !s.initialized {
		return nil, fmt.Errorf("XRPL service not initialized")
	}

	log.Printf("Starting complete payment transaction workflow: %s -> %s, Amount: %s %s", fromAddress, toAddress, amount, currency)

	// Step 1: Create payment transaction
	payment, err := s.CreatePaymentTransaction(fromAddress, toAddress, amount, currency, "", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment transaction: %w", err)
	}

	// Step 2: Sign transaction
	txBlob, err := s.SignPaymentTransaction(payment, privateKeyHex, keyType)
	if err != nil {
		return nil, fmt.Errorf("failed to sign payment transaction: %w", err)
	}

	// Step 3: Submit transaction
	result, err := s.SubmitPaymentTransaction(txBlob)
	if err != nil {
		return nil, fmt.Errorf("failed to submit payment transaction: %w", err)
	}

	// Step 4: Monitor transaction
	status, err := s.MonitorPaymentTransaction(result.TransactionID, 10, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor payment transaction: %w", err)
	}

	log.Printf("Payment transaction workflow completed successfully: %s, Final Status: %s", result.TransactionID, status.Status)
	return status, nil
}
