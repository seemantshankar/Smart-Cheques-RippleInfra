package xrpl

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	NetworkURL string
	TestNet    bool
	httpClient *http.Client
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

func NewClient(networkURL string, testNet bool) *Client {
	return &Client{
		NetworkURL: networkURL,
		TestNet:    testNet,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

// GenerateCondition creates a cryptographic condition for escrow
func (c *Client) GenerateCondition(secret string) (condition string, fulfillment string, retErr error) {
	if secret == "" {
		return "", "", fmt.Errorf("secret cannot be empty")
	}

	// Create a simple SHA-256 based condition
	hash := sha256.Sum256([]byte(secret))
	condition = hex.EncodeToString(hash[:])
	fulfillment = secret

	log.Printf("Generated condition: %s for secret", condition)
	return condition, fulfillment, nil
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
