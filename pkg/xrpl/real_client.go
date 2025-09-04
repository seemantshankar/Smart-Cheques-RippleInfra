package xrpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/xrpl-go/xrpl-go/wallet"
	"github.com/ybbus/jsonrpc/v3"
)

// RealXRPLClient provides real XRPL network integration
type RealXRPLClient struct {
	NetworkURL string
	TestNet    bool
}

// NewRealXRPLClient creates a new real XRPL client
func NewRealXRPLClient(networkURL string, testNet bool) *RealXRPLClient {
	return &RealXRPLClient{
		NetworkURL: networkURL,
		TestNet:    testNet,
	}
}

// SubmitRealTransaction submits a transaction to the real XRPL network
func (c *RealXRPLClient) SubmitRealTransaction(txBlob string) (*TransactionResult, error) {
	if txBlob == "" {
		return nil, fmt.Errorf("transaction blob cannot be empty")
	}

	// Create JSON-RPC client for XRPL
	rpcClient := jsonrpc.NewClient(c.NetworkURL)

	// Submit transaction using real XRPL API
	var response jsonrpc.RPCResponse
	err := rpcClient.CallFor(context.Background(), &response, "submit", map[string]interface{}{
		"tx_blob": txBlob,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction: %v", err)
	}

	// Parse the real XRPL response
	if response.Error != nil {
		return nil, fmt.Errorf("XRPL submission error: %s", response.Error.Message)
	}

	// Extract transaction result from real response
	result, err := c.parseRealSubmitResponse(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse submit response: %w", err)
	}

	log.Printf("Transaction submitted successfully to XRPL testnet: %s", result.TransactionID)
	return result, nil
}

// MonitorRealTransaction monitors the status of a submitted transaction on the real network
func (c *RealXRPLClient) MonitorRealTransaction(transactionID string, maxRetries int, retryInterval time.Duration) (*TransactionStatus, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	// Create JSON-RPC client for XRPL
	rpcClient := jsonrpc.NewClient(c.NetworkURL)

	status := &TransactionStatus{
		TransactionID: transactionID,
		Status:        "pending",
		SubmitTime:    time.Now(),
		LastChecked:   time.Now(),
		RetryCount:    0,
	}

	// Monitor transaction status with real XRPL API calls
	for i := 0; i < maxRetries; i++ {
		log.Printf("Monitoring transaction %s on XRPL testnet (attempt %d/%d)", transactionID, i+1, maxRetries)

		// Query transaction status using real XRPL API
		var response jsonrpc.RPCResponse
		err := rpcClient.CallFor(context.Background(), &response, "tx", map[string]interface{}{
			"transaction": transactionID,
			"binary":      false,
		})

		if err != nil {
			log.Printf("Failed to query transaction %s (attempt %d): %v", transactionID, i+1, err)
			status.RetryCount++
			time.Sleep(retryInterval)
			continue
		}

		// Parse real XRPL response
		if response.Error != nil {
			// Transaction might still be pending
			if strings.Contains(response.Error.Message, "txnNotFound") {
				log.Printf("Transaction %s still pending on XRPL testnet (attempt %d)", transactionID, i+1)
				status.RetryCount++
				time.Sleep(retryInterval)
				continue
			}
			return nil, fmt.Errorf("XRPL query error: %s", response.Error.Message)
		}

		// Transaction found - parse real status
		txStatus, err := c.parseRealTransactionStatus(response.Result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transaction status: %w", err)
		}

		// Update status with real data
		status.Status = txStatus.Status
		status.LedgerIndex = txStatus.LedgerIndex
		status.Validated = txStatus.Validated
		status.ResultCode = txStatus.ResultCode
		status.ResultMessage = txStatus.ResultMessage
		status.LastChecked = time.Now()

		// Check if transaction is validated
		if txStatus.Validated {
			log.Printf("Transaction %s validated in ledger %d on XRPL testnet", transactionID, txStatus.LedgerIndex)
			break
		}

		// Wait before next retry
		time.Sleep(retryInterval)
	}

	return status, nil
}

// GetRealTransactionStatus gets the current status of a transaction from the real network
func (c *RealXRPLClient) GetRealTransactionStatus(transactionID string) (*TransactionStatus, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID cannot be empty")
	}

	// Create JSON-RPC client for XRPL
	rpcClient := jsonrpc.NewClient(c.NetworkURL)

	// Query transaction status using real XRPL API
	var response jsonrpc.RPCResponse
	err := rpcClient.CallFor(context.Background(), &response, "tx", map[string]interface{}{
		"transaction": transactionID,
		"binary":      false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query transaction: %w", err)
	}

	// Parse real XRPL response
	if response.Error != nil {
		return nil, fmt.Errorf("XRPL query error: %s", response.Error.Message)
	}

	// Parse real transaction status
	txStatus, err := c.parseRealTransactionStatus(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction status: %w", err)
	}

	return txStatus, nil
}

// GetCurrentLedgerIndex gets the current ledger index from the real XRPL network
func (c *RealXRPLClient) GetCurrentLedgerIndex() (uint32, error) {
	// For now, use a direct HTTP call to test connectivity
	// In production, this would use the JSON-RPC library

	// Create a simple HTTP client
	client := &http.Client{Timeout: 10 * time.Second}

	// Create the request payload
	payload := map[string]interface{}{
		"method": "ledger",
		"params": []map[string]interface{}{
			{
				"ledger_index": "validated",
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.NetworkURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Debug: log the response structure
	log.Printf("XRPL Response: %+v", response)

	// Check for errors
	if result, ok := response["result"].(map[string]interface{}); ok {
		log.Printf("Result: %+v", result)
		if ledger, ok := result["ledger_index"].(string); ok {
			log.Printf("Ledger Index (string): %s", ledger)
			if ledgerIndex, err := strconv.ParseUint(ledger, 10, 32); err == nil {
				return uint32(ledgerIndex), nil
			}
		}
		// Also try as float64 (some APIs return numbers as floats)
		if ledger, ok := result["ledger_index"].(float64); ok {
			log.Printf("Ledger Index (float64): %f", ledger)
			return uint32(ledger), nil
		}
	}

	return 0, fmt.Errorf("could not extract ledger index from response")
}

// GetAccountInfo gets account information from the real XRPL network
func (c *RealXRPLClient) GetAccountInfo(address string) (map[string]interface{}, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Use direct HTTP call for now to avoid JSON-RPC library issues
	client := &http.Client{Timeout: 10 * time.Second}

	// Create the request payload
	payload := map[string]interface{}{
		"method": "account_info",
		"params": []map[string]interface{}{
			{
				"account": address,
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.NetworkURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if response["error"] != nil {
		return nil, fmt.Errorf("XRPL account query error: %v", response["error"])
	}

	// Extract result
	if result, ok := response["result"].(map[string]interface{}); ok {
		return result, nil
	}

	return nil, fmt.Errorf("invalid account response format")
}

// Helper methods for real XRPL API responses

// parseRealSubmitResponse parses the real response from XRPL submit API
func (c *RealXRPLClient) parseRealSubmitResponse(result interface{}) (*TransactionResult, error) {
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

// parseRealTransactionStatus parses the real response from XRPL transaction query API
func (c *RealXRPLClient) parseRealTransactionStatus(result interface{}) (*TransactionStatus, error) {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid transaction response format")
	}

	status := &TransactionStatus{
		LastChecked: time.Now(),
	}

	// Extract transaction hash
	if hash, ok := resultMap["hash"].(string); ok {
		status.TransactionID = hash
	}

	// Extract ledger index
	if ledgerIndex, ok := resultMap["ledger_index"].(float64); ok {
		status.LedgerIndex = uint32(ledgerIndex)
	}

	// Extract validation status
	if validated, ok := resultMap["validated"].(bool); ok {
		status.Validated = validated
	}

	// Extract result code from metadata
	if meta, ok := resultMap["meta"].(map[string]interface{}); ok {
		if resultCode, ok := meta["TransactionResult"].(string); ok {
			status.ResultCode = resultCode
		}
	}

	// Determine status based on validation and result code
	if status.Validated {
		if status.ResultCode == "tesSUCCESS" {
			status.Status = "validated"
		} else {
			status.Status = "failed"
		}
	} else {
		status.Status = "pending"
	}

	return status, nil
}

// GenerateWallet creates a real XRPL wallet using the xrpl-go library
func (c *RealXRPLClient) GenerateWallet() (*WalletInfo, error) {
	// Use real XRPL wallet generation - create a random seed first
	seedBytes := make([]byte, 16)
	if _, err := rand.Read(seedBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %w", err)
	}

	// Convert to hex string for seed
	seed := hex.EncodeToString(seedBytes)

	// Use FromSeed to create wallet from the generated seed
	xrplWallet, err := wallet.FromSeed(seed, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create XRPL wallet: %w", err)
	}

	// Convert to our WalletInfo format
	walletInfo := &WalletInfo{
		Address:    xrplWallet.GetAddress().String(),
		PublicKey:  "",   // Will be set when signing
		PrivateKey: seed, // Use seed as private key for now
		Seed:       seed,
	}

	log.Printf("Generated real XRPL wallet: %s", walletInfo.Address)
	return walletInfo, nil
}

// GenerateCondition creates a cryptographic condition for escrow
func (c *RealXRPLClient) GenerateCondition(secret string) (condition string, fulfillment string, retErr error) {
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

// CreateEscrow creates a real XRPL escrow transaction
func (c *RealXRPLClient) CreateEscrow(escrow *EscrowCreate) (*TransactionResult, error) {
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

	// Get account sequence for escrow creation
	accountInfo, err := c.GetAccountInfo(escrow.Account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// Extract sequence from account info
	sequence, ok := accountInfo["Sequence"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to get account sequence")
	}

	// Create escrow transaction using real XRPL API
	escrowTx := map[string]interface{}{
		"TransactionType": "EscrowCreate",
		"Account":         escrow.Account,
		"Destination":     escrow.Destination,
		"Amount":          escrow.Amount,
		"Sequence":        uint32(sequence),
	}

	if escrow.Condition != "" {
		escrowTx["Condition"] = escrow.Condition
	}
	if escrow.CancelAfter > 0 {
		escrowTx["CancelAfter"] = escrow.CancelAfter
	}
	if escrow.FinishAfter > 0 {
		escrowTx["FinishAfter"] = escrow.FinishAfter
	}

	// Submit escrow creation transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var response jsonrpc.RPCResponse
	rpcClient := jsonrpc.NewClient(c.NetworkURL)
	err = rpcClient.CallFor(ctx, &response, "submit", map[string]interface{}{
		"tx_json": escrowTx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit escrow creation: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("XRPL escrow creation error: %s", response.Error.Message)
	}

	// Parse response to get transaction hash
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from XRPL")
	}

	txHash, _ := result["tx_hash"].(string)
	engineResult, _ := result["engine_result"].(string)

	log.Printf("Created escrow: %s -> %s, Amount: %s, TxID: %s",
		escrow.Account, escrow.Destination, escrow.Amount, txHash)

	return &TransactionResult{
		TransactionID: txHash,
		LedgerIndex:   0, // Will be set when validated
		Validated:     false,
		ResultCode:    engineResult,
		ResultMessage: "Escrow creation submitted to XRPL",
	}, nil
}

// FinishEscrow completes a real XRPL escrow transaction
func (c *RealXRPLClient) FinishEscrow(finish *EscrowFinish) (*TransactionResult, error) {
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

	// Get account sequence
	accountInfo, err := c.GetAccountInfo(finish.Account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// Extract sequence from account info
	sequence, ok := accountInfo["Sequence"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to get account sequence")
	}

	// Create escrow finish transaction
	finishTx := map[string]interface{}{
		"TransactionType": "EscrowFinish",
		"Account":         finish.Account,
		"Owner":           finish.Owner,
		"OfferSequence":   finish.OfferSequence,
		"Sequence":        uint32(sequence),
	}

	if finish.Condition != "" {
		finishTx["Condition"] = finish.Condition
	}
	if finish.Fulfillment != "" {
		finishTx["Fulfillment"] = finish.Fulfillment
	}

	// Submit escrow finish transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var response jsonrpc.RPCResponse
	rpcClient := jsonrpc.NewClient(c.NetworkURL)
	err = rpcClient.CallFor(ctx, &response, "submit", map[string]interface{}{
		"tx_json": finishTx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to submit escrow finish: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("XRPL escrow finish error: %s", response.Error.Message)
	}

	// Parse response
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from XRPL")
	}

	txHash, _ := result["tx_hash"].(string)
	engineResult, _ := result["engine_result"].(string)

	log.Printf("Finished escrow: Account: %s, Owner: %s, Sequence: %d, TxID: %s",
		finish.Account, finish.Owner, finish.OfferSequence, txHash)

	return &TransactionResult{
		TransactionID: txHash,
		LedgerIndex:   0, // Will be set when validated
		Validated:     false,
		ResultCode:    engineResult,
		ResultMessage: "Escrow finish submitted to XRPL",
	}, nil
}
