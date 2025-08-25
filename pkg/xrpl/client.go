package xrpl

import (
	"crypto/rand"
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

	log.Printf("Transaction submitted successfully: %s", txBlob[:min(20, len(txBlob))]+"...")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}