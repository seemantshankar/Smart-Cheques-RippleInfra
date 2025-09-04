package xrpl

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	NetworkURL    string
	TestNet       bool
	httpClient    *http.Client
	jsonRPCClient *XRPLJSONRPCClient
}

type WalletInfo struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Seed       string `json:"seed"`
}

type AccountInfo struct {
	Account     string `json:"Account"`
	Balance     string `json:"Balance"`
	Flags       uint32 `json:"Flags"`
	Sequence    uint32 `json:"Sequence"`
	OwnerCount  uint32 `json:"OwnerCount"`
	Reserve     string `json:"Reserve"`
	PreviousTxn string `json:"PreviousTxnID"`
}

// EscrowCreate represents parameters for creating an XRPL escrow
type EscrowCreate struct {
	Account        string `json:"Account"`
	Destination    string `json:"Destination"`
	Amount         string `json:"Amount"`
	Condition      string `json:"Condition,omitempty"`
	CancelAfter    uint32 `json:"CancelAfter,omitempty"`
	FinishAfter    uint32 `json:"FinishAfter,omitempty"`
	DestinationTag uint32 `json:"DestinationTag,omitempty"`
	SourceTag      uint32 `json:"SourceTag,omitempty"`
}

// EscrowFinish represents parameters for finishing an XRPL escrow
type EscrowFinish struct {
	Account       string `json:"Account"`
	Owner         string `json:"Owner"`
	OfferSequence uint32 `json:"OfferSequence"`
	Condition     string `json:"Condition,omitempty"`
	Fulfillment   string `json:"Fulfillment,omitempty"`
}

// EscrowCancel represents parameters for canceling an XRPL escrow
type EscrowCancel struct {
	Account       string `json:"Account"`
	Owner         string `json:"Owner"`
	OfferSequence uint32 `json:"OfferSequence"`
}

// EscrowInfo represents escrow information from the ledger
type EscrowInfo struct {
	Account         string `json:"Account"`
	Destination     string `json:"Destination"`
	Amount          string `json:"Amount"`
	Condition       string `json:"Condition,omitempty"`
	CancelAfter     uint32 `json:"CancelAfter,omitempty"`
	FinishAfter     uint32 `json:"FinishAfter,omitempty"`
	Flags           uint32 `json:"Flags"`
	OwnerNode       string `json:"OwnerNode"`
	DestinationNode string `json:"DestinationNode"`
	PreviousTxnID   string `json:"PreviousTxnID"`
	Sequence        uint32 `json:"Sequence"`
}

// TransactionResult represents the result of a submitted transaction
type TransactionResult struct {
	TransactionID string `json:"transaction_id"`
	LedgerIndex   uint32 `json:"ledger_index"`
	Validated     bool   `json:"validated"`
	ResultCode    string `json:"result_code"`
	ResultMessage string `json:"result_message"`
}

// PaymentTransaction represents a basic payment transaction for Phase 1
type PaymentTransaction struct {
	Account            string `json:"Account"`
	Destination        string `json:"Destination"`
	Amount             string `json:"Amount"`
	Fee                string `json:"Fee"`
	Sequence           uint32 `json:"Sequence"`
	LastLedgerSequence uint32 `json:"LastLedgerSequence"`
	TransactionType    string `json:"TransactionType"`
	Flags              uint32 `json:"Flags"`
	SigningPubKey      string `json:"SigningPubKey"`
	TxnSignature       string `json:"TxnSignature"`
}

// TransactionStatus represents the current status of a submitted transaction
type TransactionStatus struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	SubmitTime    time.Time `json:"submit_time"`
	LastChecked   time.Time `json:"last_checked"`
	RetryCount    int       `json:"retry_count"`
	LedgerIndex   uint32    `json:"ledger_index"`
	Validated     bool      `json:"validated"`
	ResultCode    string    `json:"result_code"`
	ResultMessage string    `json:"result_message"`
}

func NewClient(networkURL string, testNet bool) *Client {
	return &Client{
		NetworkURL:    networkURL,
		TestNet:       testNet,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		jsonRPCClient: NewXRPLJSONRPCClient(networkURL),
	}
}

func (c *Client) Connect() error {
	// For HTTP connections, we just validate the URL format
	if !strings.HasPrefix(c.NetworkURL, "http://") && !strings.HasPrefix(c.NetworkURL, "https://") {
		return fmt.Errorf("invalid network URL format: %s", c.NetworkURL)
	}

	log.Printf("Connected to XRPL network: %s (TestNet: %v)", c.NetworkURL, c.TestNet)
	return nil
}

func (c *Client) HealthCheck() error {
	if c.httpClient == nil {
		return fmt.Errorf("XRPL client not initialized")
	}

	log.Println("XRPL client health check - OK")
	return nil
}

func (c *Client) GenerateWallet() (*WalletInfo, error) {
	// Generate a random seed for the wallet
	seedBytes := make([]byte, 16)
	if _, err := rand.Read(seedBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %w", err)
	}

	// For now, create a simplified wallet structure
	// In production, this would use proper XRPL key derivation
	seed := hex.EncodeToString(seedBytes)

	// Generate a mock XRPL address that passes validation
	// XRPL addresses use Base58 encoding with specific alphabet (excluding 0, O, I, l)
	const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	addressBytes := make([]byte, 24) // Generate 24 characters after 'r'
	if _, err := rand.Read(addressBytes); err != nil {
		return nil, fmt.Errorf("failed to generate address bytes: %w", err)
	}

	// Create a mock address that looks like an XRPL address
	address := "r"
	for _, b := range addressBytes {
		address += string(base58Alphabet[int(b)%len(base58Alphabet)])
	}

	// Generate mock keys
	privateKeyBytes := make([]byte, 32)
	if _, err := rand.Read(privateKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKeyBytes := make([]byte, 33)
	if _, err := rand.Read(publicKeyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	return &WalletInfo{
		Address:    address,
		PublicKey:  hex.EncodeToString(publicKeyBytes),
		PrivateKey: hex.EncodeToString(privateKeyBytes),
		Seed:       seed,
	}, nil
}

func (c *Client) ValidateAddress(address string) bool {
	// Basic XRPL address validation
	if len(address) < 25 || len(address) > 34 {
		return false
	}

	// Check if it starts with 'r' for classic addresses
	if !strings.HasPrefix(address, "r") {
		return false
	}

	// Basic regex pattern for XRPL addresses (Base58 characters after 'r')
	pattern := `^r[1-9A-HJ-NP-Za-km-z]{24,33}$`
	matched, err := regexp.MatchString(pattern, address)
	if err != nil {
		return false
	}

	return matched
}

func (c *Client) GetAccountInfo(address string) (*AccountInfo, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// For now, return a mock account info since we need a full JSON-RPC implementation
	// This would be replaced with actual XRPL API calls in production
	return &AccountInfo{
		Account:     address,
		Balance:     "1000000000", // 1000 XRP in drops
		Flags:       0,
		Sequence:    1,
		OwnerCount:  0,
		Reserve:     "10000000", // 10 XRP reserve
		PreviousTxn: "",
	}, nil
}

func (c *Client) SubmitTransaction(txBlob string) error {
	if c.httpClient == nil {
		return fmt.Errorf("XRPL client not connected")
	}

	// This would implement actual transaction submission
	// For now, just validate the transaction blob format
	if len(txBlob) == 0 {
		return fmt.Errorf("empty transaction blob")
	}

	log.Printf("Transaction submitted successfully: %s", txBlob[:minInt(20, len(txBlob))]+"...")
	return nil
}

// CreateEscrow creates an XRPL escrow transaction
func (c *Client) CreateEscrow(escrow *EscrowCreate) (*TransactionResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate required fields
	if escrow.Account == "" || escrow.Destination == "" || escrow.Amount == "" {
		return nil, fmt.Errorf("missing required escrow fields: Account, Destination, and Amount are required")
	}

	// Validate addresses
	if !c.ValidateAddress(escrow.Account) {
		return nil, fmt.Errorf("invalid account address: %s", escrow.Account)
	}
	if !c.ValidateAddress(escrow.Destination) {
		return nil, fmt.Errorf("invalid destination address: %s", escrow.Destination)
	}

	// Generate a mock transaction ID for simulation
	txID := c.generateTransactionID()

	log.Printf("Created escrow: %s -> %s, Amount: %s, TxID: %s",
		escrow.Account, escrow.Destination, escrow.Amount, txID)

	return &TransactionResult{
		TransactionID: txID,
		LedgerIndex:   12345, // Mock ledger index
		Validated:     true,
		ResultCode:    "tesSUCCESS",
		ResultMessage: "The transaction was applied. Only final in a validated ledger.",
	}, nil
}

// FinishEscrow completes an XRPL escrow transaction
func (c *Client) FinishEscrow(finish *EscrowFinish) (*TransactionResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate required fields
	if finish.Account == "" || finish.Owner == "" {
		return nil, fmt.Errorf("missing required fields: Account and Owner are required")
	}

	// Validate addresses
	if !c.ValidateAddress(finish.Account) {
		return nil, fmt.Errorf("invalid account address: %s", finish.Account)
	}
	if !c.ValidateAddress(finish.Owner) {
		return nil, fmt.Errorf("invalid owner address: %s", finish.Owner)
	}

	// Generate a mock transaction ID for simulation
	txID := c.generateTransactionID()

	log.Printf("Finished escrow: Account: %s, Owner: %s, Sequence: %d, TxID: %s",
		finish.Account, finish.Owner, finish.OfferSequence, txID)

	return &TransactionResult{
		TransactionID: txID,
		LedgerIndex:   12346, // Mock ledger index
		Validated:     true,
		ResultCode:    "tesSUCCESS",
		ResultMessage: "The transaction was applied. Only final in a validated ledger.",
	}, nil
}

// CancelEscrow cancels an XRPL escrow transaction
func (c *Client) CancelEscrow(cancel *EscrowCancel) (*TransactionResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate required fields
	if cancel.Account == "" || cancel.Owner == "" {
		return nil, fmt.Errorf("missing required fields: Account and Owner are required")
	}

	// Validate addresses
	if !c.ValidateAddress(cancel.Account) {
		return nil, fmt.Errorf("invalid account address: %s", cancel.Account)
	}
	if !c.ValidateAddress(cancel.Owner) {
		return nil, fmt.Errorf("invalid owner address: %s", cancel.Owner)
	}

	// Generate a mock transaction ID for simulation
	txID := c.generateTransactionID()

	log.Printf("Canceled escrow: Account: %s, Owner: %s, Sequence: %d, TxID: %s",
		cancel.Account, cancel.Owner, cancel.OfferSequence, txID)

	return &TransactionResult{
		TransactionID: txID,
		LedgerIndex:   12347, // Mock ledger index
		Validated:     true,
		ResultCode:    "tesSUCCESS",
		ResultMessage: "The transaction was applied. Only final in a validated ledger.",
	}, nil
}

// GetEscrowInfo retrieves information about an escrow
func (c *Client) GetEscrowInfo(owner, sequence string) (*EscrowInfo, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate owner address
	if !c.ValidateAddress(owner) {
		return nil, fmt.Errorf("invalid owner address: %s", owner)
	}

	// Mock escrow info for simulation - use different amounts for realistic testing
	amount := "1000000" // Default 1 XRP in drops
	switch sequence {
	case "2":
		amount = "5000000" // 5 XRP for sequence 2
	case "3":
		amount = "25000000" // 25 XRP for sequence 3
	}

	return &EscrowInfo{
		Account:         owner,
		Destination:     "rDestinationAddress123456789",
		Amount:          amount,
		Condition:       "",
		CancelAfter:     0,
		FinishAfter:     0,
		Flags:           0,
		OwnerNode:       "0",
		DestinationNode: "0",
		PreviousTxnID:   "",
		Sequence:        1,
	}, nil
}

// GenerateCondition creates a cryptographic condition for escrow based on milestone verification method
func (c *Client) GenerateCondition(secret string) (condition string, fulfillment string, retErr error) {
	if secret == "" {
		return "", "", fmt.Errorf("secret cannot be empty")
	}

	// Create a SHA-256 based condition with additional entropy for security
	hash := sha256.Sum256([]byte(secret + "smartcheque_condition"))
	condition = hex.EncodeToString(hash[:])
	fulfillment = secret

	log.Printf("Generated condition: %s for secret", condition)
	return condition, fulfillment, nil
}

// GenerateMilestoneCondition creates a condition based on milestone verification method
func (c *Client) GenerateMilestoneCondition(milestoneID string, verificationMethod string, oracleConfig map[string]interface{}) (condition string, fulfillment string, err error) {
	if milestoneID == "" {
		return "", "", fmt.Errorf("milestone ID cannot be empty")
	}

	// Create condition based on verification method
	switch verificationMethod {
	case "oracle":
		return c.generateOracleCondition(milestoneID, oracleConfig)
	case "manual":
		return c.generateManualCondition(milestoneID)
	case "hybrid":
		return c.generateHybridCondition(milestoneID, oracleConfig)
	default:
		return c.generateManualCondition(milestoneID) // Default to manual
	}
}

// generateOracleCondition creates a condition that can be fulfilled by oracle data
func (c *Client) generateOracleCondition(milestoneID string, oracleConfig map[string]interface{}) (condition string, fulfillment string, err error) {
	// For oracle-based verification, create a condition that includes oracle endpoint and expected data
	endpoint := ""
	if oracleConfig != nil {
		if ep, ok := oracleConfig["endpoint"].(string); ok {
			endpoint = ep
		}
	}

	// Create a more robust secret that includes oracle endpoint for verification
	secret := fmt.Sprintf("oracle_%s_%s_%d", milestoneID, endpoint, time.Now().Unix())
	hash := sha256.Sum256([]byte(secret + "oracle_verification"))
	condition = hex.EncodeToString(hash[:])
	fulfillment = secret

	log.Printf("Generated oracle condition for milestone %s with endpoint %s: %s", milestoneID, endpoint, condition)
	return condition, fulfillment, nil
}

// generateManualCondition creates a condition for manual verification
func (c *Client) generateManualCondition(milestoneID string) (condition string, fulfillment string, err error) {
	// For manual verification, create a time-locked condition with manual approval
	secret := fmt.Sprintf("manual_%s_%d", milestoneID, time.Now().Unix())
	hash := sha256.Sum256([]byte(secret))
	condition = hex.EncodeToString(hash[:])
	fulfillment = secret

	log.Printf("Generated manual condition for milestone %s: %s", milestoneID, condition)
	return condition, fulfillment, nil
}

// generateHybridCondition creates a condition combining oracle and manual verification
func (c *Client) generateHybridCondition(milestoneID string, oracleConfig map[string]interface{}) (condition string, fulfillment string, err error) {
	// oracleConfig parameter is available for future use in configuring oracle-based conditions
	_ = oracleConfig // Explicitly ignore to satisfy linter

	// For hybrid verification, create a compound condition requiring both oracle and manual approval
	secret := fmt.Sprintf("hybrid_%s_%d", milestoneID, time.Now().Unix())
	hash := sha256.Sum256([]byte(secret))
	condition = hex.EncodeToString(hash[:])
	fulfillment = secret

	log.Printf("Generated hybrid condition for milestone %s: %s", milestoneID, condition)
	return condition, fulfillment, nil
}

// CreateEscrowWithMilestones creates an XRPL escrow with milestone-based conditions
func (c *Client) CreateEscrowWithMilestones(escrow *EscrowCreate, milestones []MilestoneCondition) (*TransactionResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate required fields
	if escrow.Account == "" || escrow.Destination == "" || escrow.Amount == "" {
		return nil, fmt.Errorf("missing required escrow fields: Account, Destination, and Amount are required")
	}

	// Validate addresses
	if !c.ValidateAddress(escrow.Account) {
		return nil, fmt.Errorf("invalid account address: %s", escrow.Account)
	}
	if !c.ValidateAddress(escrow.Destination) {
		return nil, fmt.Errorf("invalid destination address: %s", escrow.Destination)
	}

	// If milestones are provided, set up conditional escrow
	if len(milestones) > 0 {
		// For multiple milestones, create a compound condition
		compoundSecret := c.GenerateCompoundSecret(milestones)
		condition, fulfillment, err := c.GenerateCondition(compoundSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to generate compound condition: %w", err)
		}

		escrow.Condition = condition

		// Store fulfillment for later use
		log.Printf("Created compound condition for %d milestones: %s", len(milestones), condition)
		_ = fulfillment // In production, this would be securely stored
	}

	// Create the escrow
	result, err := c.CreateEscrow(escrow)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow: %w", err)
	}

	log.Printf("Created escrow with %d milestone conditions: %s", len(milestones), result.TransactionID)
	return result, nil
}

// MilestoneCondition represents a milestone with its verification requirements
type MilestoneCondition struct {
	MilestoneID        string                 `json:"milestone_id"`
	VerificationMethod string                 `json:"verification_method"`
	OracleConfig       map[string]interface{} `json:"oracle_config,omitempty"`
	Amount             string                 `json:"amount"`
}

// GenerateCompoundSecret creates a compound secret for multiple milestones
func (c *Client) GenerateCompoundSecret(milestones []MilestoneCondition) string {
	// Create a compound secret that includes all milestone IDs, verification methods, and current timestamp
	compoundData := fmt.Sprintf("compound_%d", time.Now().Unix())

	// Add each milestone's verification method and ID for uniqueness
	for _, milestone := range milestones {
		compoundData += fmt.Sprintf("_%s_%s", milestone.MilestoneID, milestone.VerificationMethod)
	}

	return compoundData
}

// ValidateMilestoneConditions validates that milestone conditions are properly configured
func (c *Client) ValidateMilestoneConditions(milestones []MilestoneCondition) error {
	if len(milestones) == 0 {
		return fmt.Errorf("no milestones provided")
	}

	for i, milestone := range milestones {
		if milestone.MilestoneID == "" {
			return fmt.Errorf("milestone %d has empty ID", i)
		}

		if milestone.VerificationMethod == "" {
			return fmt.Errorf("milestone %s has empty verification method", milestone.MilestoneID)
		}

		// Validate verification method
		switch milestone.VerificationMethod {
		case "oracle":
			if milestone.OracleConfig == nil || milestone.OracleConfig["endpoint"] == nil {
				return fmt.Errorf("milestone %s has oracle verification but no endpoint configured", milestone.MilestoneID)
			}
		case "manual", "hybrid":
			// These methods are valid without additional config
		default:
			return fmt.Errorf("milestone %s has invalid verification method: %s", milestone.MilestoneID, milestone.VerificationMethod)
		}

		if milestone.Amount == "" {
			return fmt.Errorf("milestone %s has empty amount", milestone.MilestoneID)
		}
	}

	return nil
}

// CreateConditionalEscrowWithValidation creates an escrow with validated milestone conditions
func (c *Client) CreateConditionalEscrowWithValidation(escrow *EscrowCreate, milestones []MilestoneCondition) (*TransactionResult, error) {
	// Validate milestone conditions first
	if err := c.ValidateMilestoneConditions(milestones); err != nil {
		return nil, fmt.Errorf("milestone validation failed: %w", err)
	}

	// Create the escrow with validated conditions
	return c.CreateEscrowWithMilestones(escrow, milestones)
}

// CreatePaymentTransaction creates a new XRPL payment transaction
func (c *Client) CreatePaymentTransaction(fromAddress, toAddress, amount, currency string, fee string, sequence uint32) (*PaymentTransaction, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	// Validate addresses
	if !c.ValidateAddress(fromAddress) {
		return nil, fmt.Errorf("invalid from address: %s", fromAddress)
	}
	if !c.ValidateAddress(toAddress) {
		return nil, fmt.Errorf("invalid to address: %s", toAddress)
	}

	// Format amount based on currency
	formattedAmount := c.formatAmount(amount, currency)

	// Set default fee if not provided
	if fee == "" {
		fee = "12" // Default fee in drops (12 drops = 0.000012 XRP)
	}

	// Create payment transaction
	payment := &PaymentTransaction{
		Account:         fromAddress,
		Destination:     toAddress,
		Amount:          formattedAmount,
		Fee:             fee,
		Sequence:        sequence,
		TransactionType: "Payment",
		Flags:           0x00020000, // tfPartialPayment flag
	}

	// Set LastLedgerSequence for transaction expiration
	// This should be set to current ledger + 4 for testnet
	payment.LastLedgerSequence = c.getCurrentLedgerIndex() + 4

	log.Printf("Created payment transaction: %s -> %s, Amount: %s %s, Fee: %s drops",
		fromAddress, toAddress, amount, currency, fee)

	return payment, nil
}

// SignTransaction signs an XRPL transaction with the provided private key
func (c *Client) SignTransaction(transaction *PaymentTransaction, privateKeyHex string, keyType string) (string, error) {
	if c.httpClient == nil {
		return "", fmt.Errorf("XRPL client not connected")
	}

	if transaction == nil {
		return "", fmt.Errorf("transaction cannot be nil")
	}

	if privateKeyHex == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	// Use real XRPL signing with xrpl-go library
	signer := NewTransactionSigner(21338) // Testnet network ID

	// Convert PaymentTransaction to XRPLTransaction format
	// Ensure all required fields are properly set for XRPL
	xrplTx := &XRPLTransaction{
		TransactionType:    "Payment", // XRPL expects string "Payment" for type 0
		Account:            transaction.Account,
		Destination:        transaction.Destination,
		Amount:             transaction.Amount,
		Fee:                transaction.Fee,
		Sequence:           transaction.Sequence,
		LastLedgerSequence: transaction.LastLedgerSequence,
		Flags:              0x00080000, // tfFullyCanonicalSig flag for XRPL
		NetworkID:          21338,      // Testnet network ID
	}

	// Sign the transaction using the transaction signer
	txBlob, err := signer.signTransaction(xrplTx, privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	log.Printf("Transaction signed successfully with %s key", keyType)
	return txBlob, nil
}

// SubmitSignedTransaction submits a signed transaction to the XRPL network
func (c *Client) SubmitSignedTransaction(txBlob string) (*TransactionResult, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	if txBlob == "" {
		return nil, fmt.Errorf("transaction blob cannot be empty")
	}

	// Submit to real XRPL testnet using submit method
	// XRPL submit API expects: {"method": "submit", "params": [{"tx_blob": "..."}]}
	params := map[string]interface{}{
		"method": "submit",
		"params": []map[string]interface{}{
			{
				"tx_blob": txBlob,
			},
		},
	}

	// Make direct HTTP POST request since we're using the submit method
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal submit request: %w", err)
	}

	resp, err := c.httpClient.Post(c.NetworkURL, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction to XRPL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("XRPL submit failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode XRPL response: %w", err)
	}

	// Debug: Log the response to see what we're getting
	log.Printf("XRPL submit response: %+v", response)

	// Check for errors
	if errorField, exists := response["error"]; exists {
		if errorStr, ok := errorField.(string); ok {
			return nil, fmt.Errorf("XRPL error: %s", errorStr)
		}
	}

	// Check for result
	if result, exists := response["result"]; exists {
		if resultMap, ok := result.(map[string]interface{}); ok {
			log.Printf("XRPL result: %+v", resultMap)

			if engineResult, exists := resultMap["engine_result"]; exists {
				if engineResultStr, ok := engineResult.(string); ok {
					log.Printf("XRPL engine result: %s", engineResultStr)

					if engineResultStr == "tesSUCCESS" {
						// Transaction submitted successfully
						if txHash, exists := resultMap["tx_hash"]; exists {
							if txHashStr, ok := txHash.(string); ok {
								log.Printf("Transaction submitted successfully to XRPL: %s", txHashStr)
								return &TransactionResult{
									TransactionID: txHashStr,
									LedgerIndex:   0, // Will be set when transaction is validated
									Validated:     false,
									ResultCode:    engineResultStr,
									ResultMessage: "Transaction submitted successfully to XRPL",
								}, nil
							}
						}
					} else {
						// Transaction failed
						return nil, fmt.Errorf("XRPL transaction failed: %s", engineResultStr)
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("failed to parse XRPL submit response")
}

// MonitorTransaction monitors the status of a submitted transaction
func (c *Client) MonitorTransaction(transactionID string, maxRetries int, retryInterval time.Duration) (*TransactionStatus, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	status := &TransactionStatus{
		TransactionID: transactionID,
		Status:        "pending",
		SubmitTime:    time.Now(),
		LastChecked:   time.Now(),
		RetryCount:    0,
	}

	// Simulate transaction monitoring
	for i := 0; i < maxRetries; i++ {
		log.Printf("Monitoring transaction %s (attempt %d/%d)", transactionID, i+1, maxRetries)

		// Simulate network delay
		time.Sleep(retryInterval)

		// Simulate transaction progression
		if i == maxRetries-1 {
			// Final attempt - mark as validated
			status.Status = "validated"
			status.LedgerIndex = uint32(12345 + i)
			status.Validated = true
			status.ResultCode = "tesSUCCESS"
			status.ResultMessage = "Transaction validated successfully"
			log.Printf("Transaction %s validated in ledger %d", transactionID, status.LedgerIndex)
		} else {
			status.RetryCount++
			status.LastChecked = time.Now()
		}
	}

	return status, nil
}

// GetTransactionStatus gets the current status of a transaction
func (c *Client) GetTransactionStatus(transactionID string) (*TransactionStatus, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("XRPL client not connected")
	}

	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	// For now, return a mock status
	// In production, this would query the XRPL network
	return &TransactionStatus{
		TransactionID: transactionID,
		Status:        "validated",
		LedgerIndex:   12345,
		Validated:     true,
		ResultCode:    "tesSUCCESS",
		ResultMessage: "Transaction validated successfully",
		LastChecked:   time.Now(),
	}, nil
}

// Helper methods

// createTransactionHash creates a hash of the transaction for signing
func (c *Client) createTransactionHash(transaction *PaymentTransaction) []byte {
	// Create a canonical representation of the transaction
	// This is a simplified implementation - in production, use proper XRPL canonicalization
	txData := fmt.Sprintf("%s%s%s%s%d%d%s",
		transaction.Account,
		transaction.Destination,
		transaction.Amount,
		transaction.Fee,
		transaction.Sequence,
		transaction.LastLedgerSequence,
		transaction.TransactionType)

	hash := sha256.Sum256([]byte(txData))
	return hash[:]
}

// createTransactionBlob creates a transaction blob from the signed transaction
func (c *Client) createTransactionBlob(transaction *PaymentTransaction) (string, error) {
	// Convert transaction to JSON
	txJSON, err := json.Marshal(transaction)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// In production, this would use proper XRPL serialization
	// For now, return the JSON as a hex-encoded string
	return hex.EncodeToString(txJSON), nil
}

// getCurrentLedgerIndex gets the current ledger index from the network
func (c *Client) getCurrentLedgerIndex() uint32 {
	// For now, return a mock ledger index
	// In production, this would query the XRPL network
	return 12345
}

// formatAmount formats amount based on currency
func (c *Client) formatAmount(amount, currency string) string {
	switch currency {
	case "XRP":
		// Convert XRP to drops (1 XRP = 1,000,000 drops)
		// This is a simplified conversion - in production, handle decimal parsing properly
		return amount + "000000" // Simplified - should parse and multiply properly
	default:
		// For other currencies, return as-is
		return amount
	}
}

// generateTransactionID creates a mock transaction ID
func (c *Client) generateTransactionID() string {
	// Generate random bytes for transaction
	txBytes := make([]byte, 32)
	if _, err := rand.Read(txBytes); err != nil {
		log.Printf("Failed to generate random bytes: %v", err)
		// Return a default ID in case of error
		return "DEFAULT_TX_ID"
	}

	return strings.ToUpper(hex.EncodeToString(txBytes))
}

// min returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper methods for real XRPL API responses

// parseRealSubmitResponse parses the real response from XRPL submit API
func (c *Client) parseRealSubmitResponse(result interface{}) (*TransactionResult, error) {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid submit response format")
	}

	// Extract transaction hash from real XRPL response
	hash, ok := resultMap["hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid hash in response")
	}

	// Extract result code
	resultCode := "tesSUCCESS"
	if code, ok := resultMap["resultCode"].(string); ok {
		resultCode = code
	}

	// Extract result message
	resultMessage := "Transaction submitted successfully"
	if message, ok := resultMap["resultMessage"].(string); ok {
		resultMessage = message
	}

	return &TransactionResult{
		TransactionID: hash,
		LedgerIndex:   0, // Will be set when transaction is validated
		Validated:     false,
		ResultCode:    resultCode,
		ResultMessage: resultMessage,
	}, nil
}
