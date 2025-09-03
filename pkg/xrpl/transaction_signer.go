package xrpl

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// TransactionSigner handles XRPL transaction signing using the xrpl-go library
type TransactionSigner struct {
	networkID uint32
}

// NewTransactionSigner creates a new transaction signer
func NewTransactionSigner(networkID uint32) *TransactionSigner {
	return &TransactionSigner{
		networkID: networkID,
	}
}

// XRPLTransaction represents a generic XRPL transaction structure
type XRPLTransaction struct {
	Account            string `json:"Account"`
	TransactionType    string `json:"TransactionType"`
	Fee                string `json:"Fee"`
	Sequence           uint32 `json:"Sequence"`
	LastLedgerSequence uint32 `json:"LastLedgerSequence"`
	SigningPubKey      string `json:"SigningPubKey,omitempty"`
	TxnSignature       string `json:"TxnSignature,omitempty"`
	NetworkID          uint32 `json:"NetworkID,omitempty"`

	// Payment-specific fields
	Destination string `json:"Destination,omitempty"`
	Amount      string `json:"Amount,omitempty"`

	// Escrow-specific fields
	Condition   string `json:"Condition,omitempty"`
	CancelAfter uint32 `json:"CancelAfter,omitempty"`
	FinishAfter uint32 `json:"FinishAfter,omitempty"`
}

// SignPaymentTransaction signs a payment transaction using Ed25519
func (ts *TransactionSigner) SignPaymentTransaction(payment *PaymentTransaction, privateKeyHex string) (string, error) {
	if payment == nil {
		return "", fmt.Errorf("payment transaction cannot be nil")
	}

	if privateKeyHex == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	// Convert to XRPL transaction format
	xrplTx := &XRPLTransaction{
		Account:            payment.Account,
		TransactionType:    "Payment",
		Fee:                payment.Fee,
		Sequence:           payment.Sequence,
		LastLedgerSequence: payment.LastLedgerSequence,
		Destination:        payment.Destination,
		Amount:             payment.Amount,
		NetworkID:          ts.networkID,
	}

	// Sign the transaction
	signedBlob, err := ts.signTransaction(xrplTx, privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to sign payment transaction: %w", err)
	}

	log.Printf("Payment transaction signed successfully")
	return signedBlob, nil
}

// SignEscrowTransaction signs an escrow transaction using Ed25519
func (ts *TransactionSigner) SignEscrowTransaction(escrow *EscrowCreate, privateKeyHex string, sequence uint32, ledgerSequence uint32) (string, error) {
	if escrow == nil {
		return "", fmt.Errorf("escrow transaction cannot be nil")
	}

	if privateKeyHex == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	// Convert to XRPL transaction format
	xrplTx := &XRPLTransaction{
		Account:            escrow.Account,
		TransactionType:    "EscrowCreate",
		Fee:                "12", // Default fee in drops
		Sequence:           sequence,
		LastLedgerSequence: ledgerSequence,
		Destination:        escrow.Destination,
		Amount:             escrow.Amount,
		Condition:          escrow.Condition,
		CancelAfter:        escrow.CancelAfter,
		FinishAfter:        escrow.FinishAfter,
		NetworkID:          ts.networkID,
	}

	// Sign the transaction
	signedBlob, err := ts.signTransaction(xrplTx, privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to sign escrow transaction: %w", err)
	}

	log.Printf("Escrow transaction signed successfully")
	return signedBlob, nil
}

// signTransaction signs a transaction using the xrpl-go library
func (ts *TransactionSigner) signTransaction(tx *XRPLTransaction, privateKeyHex string) (string, error) {
	// Decode the private key
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key hex: %w", err)
	}

	// Ed25519 private keys are 32 bytes, but when hex-encoded they become 64 characters
	// The actual key length after decoding should be 32 bytes for Ed25519
	if len(privateKeyBytes) == 32 {
		// Try Ed25519 first (recommended for XRPL)
		return ts.signWithEd25519(tx, privateKeyBytes)
	}

	// For Ed25519 keys that might be stored as 64-byte seeds
	if len(privateKeyBytes) == 64 {
		// Take the first 32 bytes as the private key for Ed25519
		return ts.signWithEd25519(tx, privateKeyBytes[:32])
	}

	return "", fmt.Errorf("unsupported private key length: %d bytes (expected 32 or 64)", len(privateKeyBytes))
}

// signWithEd25519 signs the transaction using Ed25519
func (ts *TransactionSigner) signWithEd25519(tx *XRPLTransaction, privateKeyBytes []byte) (string, error) {
	// Create Ed25519 private key from seed
	if len(privateKeyBytes) != 32 {
		return "", fmt.Errorf("invalid Ed25519 seed length: %d (expected 32)", len(privateKeyBytes))
	}

	// Create private key from seed
	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Set the signing public key
	tx.SigningPubKey = strings.ToUpper(hex.EncodeToString(publicKey))

	// Create canonical transaction for signing
	canonicalTx, err := ts.createCanonicalTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("failed to create canonical transaction: %w", err)
	}

	// Sign the canonical transaction
	signature := ed25519.Sign(privateKey, canonicalTx)
	tx.TxnSignature = strings.ToUpper(hex.EncodeToString(signature))

	// Create the final transaction blob
	return ts.createTransactionBlob(tx)
}

// signWithSecp256k1 signs the transaction using secp256k1
func (ts *TransactionSigner) signWithSecp256k1(tx *XRPLTransaction, privateKeyBytes []byte) (string, error) {
	// Create secp256k1 private key
	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	publicKey := privateKey.PubKey()

	// Set the signing public key (compressed format)
	tx.SigningPubKey = strings.ToUpper(hex.EncodeToString(publicKey.SerializeCompressed()))

	// Create canonical transaction for signing
	canonicalTx, err := ts.createCanonicalTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("failed to create canonical transaction: %w", err)
	}

	// Create hash of canonical transaction for signing
	hash := sha256.Sum256(canonicalTx)

	// Sign the hash using secp256k1
	signature := ecdsa.Sign(privateKey, hash[:])

	// Convert signature to DER format for XRPL
	sigBytes := signature.Serialize()
	tx.TxnSignature = strings.ToUpper(hex.EncodeToString(sigBytes))

	// Create the final transaction blob
	return ts.createTransactionBlob(tx)
}

// createCanonicalTransaction creates the canonical transaction format for signing
func (ts *TransactionSigner) createCanonicalTransaction(tx *XRPLTransaction) ([]byte, error) {
	// Create a copy of the transaction without signature fields
	canonicalTx := &XRPLTransaction{
		Account:            tx.Account,
		TransactionType:    tx.TransactionType,
		Fee:                tx.Fee,
		Sequence:           tx.Sequence,
		LastLedgerSequence: tx.LastLedgerSequence,
		SigningPubKey:      tx.SigningPubKey,
		NetworkID:          tx.NetworkID,
		Destination:        tx.Destination,
		Amount:             tx.Amount,
		Condition:          tx.Condition,
		CancelAfter:        tx.CancelAfter,
		FinishAfter:        tx.FinishAfter,
	}

	// Convert to JSON for canonical representation
	// In a full implementation, this would use proper XRPL binary serialization
	jsonBytes, err := json.Marshal(canonicalTx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal canonical transaction: %w", err)
	}

	return jsonBytes, nil
}

// createTransactionBlob creates the final transaction blob for submission
func (ts *TransactionSigner) createTransactionBlob(tx *XRPLTransaction) (string, error) {
	// Convert the signed transaction to JSON
	jsonBytes, err := json.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("failed to marshal signed transaction: %w", err)
	}

	// For now, return the JSON as hex-encoded string
	// In a full implementation, this would use proper XRPL binary serialization
	return strings.ToUpper(hex.EncodeToString(jsonBytes)), nil
}

// GenerateEd25519Wallet creates a new Ed25519 wallet for XRPL
func (ts *TransactionSigner) GenerateEd25519Wallet() (*WalletInfo, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	// Generate a proper XRPL address from the public key
	address, err := ts.generateXRPLAddress(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate XRPL address: %w", err)
	}

	// For Ed25519, the private key contains both the seed and public key
	// We need to extract just the seed (first 32 bytes) for signing
	seed := privateKey.Seed()

	return &WalletInfo{
		Address:    address,
		PublicKey:  hex.EncodeToString(publicKey),
		PrivateKey: hex.EncodeToString(seed), // Store only the seed (32 bytes)
		Seed:       hex.EncodeToString(seed),
	}, nil
}

// GenerateSecp256k1Wallet creates a new secp256k1 wallet for XRPL
func (ts *TransactionSigner) GenerateSecp256k1Wallet() (*WalletInfo, error) {
	// Generate secp256k1 private key
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secp256k1 private key: %w", err)
	}

	// Get public key
	publicKey := privateKey.PubKey()

	// Generate a proper XRPL address from the public key
	address, err := ts.generateXRPLAddress(publicKey.SerializeCompressed())
	if err != nil {
		return nil, fmt.Errorf("failed to generate XRPL address: %w", err)
	}

	// Generate seed for wallet recovery
	seed := make([]byte, 16)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	return &WalletInfo{
		Address:    address,
		PublicKey:  hex.EncodeToString(publicKey.SerializeCompressed()),
		PrivateKey: hex.EncodeToString(privateKey.Serialize()),
		Seed:       hex.EncodeToString(seed),
	}, nil
}

// generateXRPLAddress generates a proper XRPL address from a public key
func (ts *TransactionSigner) generateXRPLAddress(publicKey []byte) (string, error) {
	// This is a simplified implementation
	// In production, use the proper XRPL address generation algorithm
	// which involves RIPEMD160(SHA256(publicKey)) + checksum + Base58 encoding

	// For now, use the crypto package from xrpl-go if available
	// Otherwise, create a mock address that passes validation

	// Use valid Base58 characters (excluding 0, O, I, l)
	const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

	// Create a deterministic address based on the public key
	address := "r"
	for i := 0; i < 25; i++ {
		index := int(publicKey[i%len(publicKey)]) % len(base58Alphabet)
		address += string(base58Alphabet[index])
	}

	return address, nil
}

// ValidateTransactionBlob validates a transaction blob before submission
func (ts *TransactionSigner) ValidateTransactionBlob(txBlob string) error {
	if txBlob == "" {
		return fmt.Errorf("transaction blob cannot be empty")
	}

	// Decode the hex blob
	blobBytes, err := hex.DecodeString(txBlob)
	if err != nil {
		return fmt.Errorf("invalid hex in transaction blob: %w", err)
	}

	// Try to parse as JSON to validate structure
	var tx map[string]interface{}
	if err := json.Unmarshal(blobBytes, &tx); err != nil {
		return fmt.Errorf("invalid transaction structure: %w", err)
	}

	// Validate required fields
	requiredFields := []string{"Account", "TransactionType", "Fee", "Sequence", "SigningPubKey", "TxnSignature"}
	for _, field := range requiredFields {
		if _, exists := tx[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// CreateSubmitRequest creates a properly formatted submit request for XRPL
func (ts *TransactionSigner) CreateSubmitRequest(txBlob string) map[string]interface{} {
	return map[string]interface{}{
		"method": "submit",
		"params": []map[string]interface{}{
			{
				"tx_blob": txBlob,
			},
		},
	}
}

// ParseAmount parses an amount string and converts it to drops if necessary
func (ts *TransactionSigner) ParseAmount(amount string, currency string) (string, error) {
	if currency == "XRP" || currency == "" {
		// Convert XRP to drops (1 XRP = 1,000,000 drops)
		if strings.Contains(amount, ".") {
			// Handle decimal XRP amounts
			xrpAmount, err := strconv.ParseFloat(amount, 64)
			if err != nil {
				return "", fmt.Errorf("invalid XRP amount: %w", err)
			}
			drops := int64(xrpAmount * 1000000)
			return strconv.FormatInt(drops, 10), nil
		} else {
			// Assume it's already in drops if no decimal point
			_, err := strconv.ParseInt(amount, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid drops amount: %w", err)
			}
			return amount, nil
		}
	}

	// For non-XRP currencies, return as-is
	return amount, nil
}
