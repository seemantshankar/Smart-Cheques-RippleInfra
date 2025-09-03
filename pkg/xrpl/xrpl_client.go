package xrpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// XRPLClient provides methods for interacting with the XRP Ledger
type XRPLClient struct {
	NetworkURL string
	NetworkID  uint32
	httpClient *http.Client
}

// NewXRPLClient creates a new XRPLClient
func NewXRPLClient(networkURL string, networkID uint32) *XRPLClient {
	return &XRPLClient{
		NetworkURL: networkURL,
		NetworkID:  networkID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccountInfo retrieves account information from XRPL
func (c *XRPLClient) GetAccountInfo(address string) (balance string, sequence int, err error) {
	requestBody := map[string]interface{}{
		"method": "account_info",
		"params": []map[string]interface{}{
			{
				"account":      address,
				"ledger_index": "validated",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.NetworkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("invalid response format")
	}

	if result["error"] != nil {
		return "", 0, fmt.Errorf("XRPL error: %v", result["error"])
	}

	accountData, ok := result["account_data"].(map[string]interface{})
	if !ok {
		return "", 0, fmt.Errorf("invalid account data format")
	}

	balance = toString(accountData["Balance"])
	sequence = toInt(accountData["Sequence"])

	return balance, sequence, nil
}

// GetCurrentLedgerIndex gets the current ledger index
func (c *XRPLClient) GetCurrentLedgerIndex() (int, error) {
	requestBody := map[string]interface{}{
		"method": "ledger",
		"params": []map[string]interface{}{
			{
				"ledger_index": "validated",
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.NetworkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	ledgerIndex := toInt(result["ledger_index"])
	return ledgerIndex, nil
}

// CreatePaymentTransaction creates a payment transaction
func (c *XRPLClient) CreatePaymentTransaction(sourceAddr, destAddr, amount string, sequence int) (map[string]interface{}, error) {
	// Get current ledger for LastLedgerSequence
	currentLedger, err := c.GetCurrentLedgerIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get current ledger index: %w", err)
	}

	// Create payment transaction as a map
	payment := map[string]interface{}{
		"TransactionType":    "Payment",
		"Account":            sourceAddr,
		"Destination":        destAddr,
		"Amount":             amount,
		"Fee":                "12", // Standard fee
		"Sequence":           uint32(sequence),
		"LastLedgerSequence": uint32(currentLedger + 4),
		"Flags":              uint32(2147483648), // tfFullyCanonicalSig flag
	}

	return payment, nil
}

// SignTransaction signs a transaction using a wallet
func (c *XRPLClient) SignTransaction(tx map[string]interface{}, secret string) (string, string, error) {
	// Create wallet from secret
	w, err := wallet.FromSecret(secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to create wallet from secret: %w", err)
	}

	// Sign transaction locally
	txBlob, txID, err := w.Sign(tx)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	return txBlob, txID, nil
}

// SubmitTransaction submits a signed transaction blob to XRPL
func (c *XRPLClient) SubmitTransaction(txBlob string) (string, error) {
	requestBody := map[string]interface{}{
		"method": "submit",
		"params": []map[string]interface{}{
			{
				"tx_blob":   txBlob,
				"fail_hard": false,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.NetworkURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	if result["error"] != nil {
		return "", fmt.Errorf("XRPL error: %v", result["error"])
	}

	engineResult := toString(result["engine_result"])
	if engineResult != "tesSUCCESS" {
		return "", fmt.Errorf("XRPL engine result: %s - %s",
			engineResult, toString(result["engine_result_message"]))
	}

	txID := toString(result["tx_json"].(map[string]interface{})["hash"])
	return txID, nil
}

// MonitorTransaction polls the rippled 'tx' method until transaction is validated or retries exhausted
func (c *XRPLClient) MonitorTransaction(txHash string, maxRetries int, interval time.Duration) (int, bool, error) {
	for i := 0; i < maxRetries; i++ {
		requestBody := map[string]interface{}{
			"method": "tx",
			"params": []map[string]interface{}{
				{
					"transaction": txHash,
					"binary":      false,
				},
			},
		}

		jsonData, _ := json.Marshal(requestBody)
		resp, err := c.httpClient.Post(c.NetworkURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			// wait and retry
			time.Sleep(interval)
			continue
		}
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			time.Sleep(interval)
			continue
		}
		resp.Body.Close()

		if response["result"] == nil {
			time.Sleep(interval)
			continue
		}
		resMap, _ := response["result"].(map[string]interface{})
		if resMap["error"] != nil {
			// txn not found -> keep waiting
			if errStr, ok := resMap["error"].(string); ok && errStr == "txnNotFound" {
				time.Sleep(interval)
				continue
			}
			return 0, false, fmt.Errorf("rippled tx error: %v", resMap["error"])
		}

		// Extract ledger_index and validated
		ledgerIndex := toInt(resMap["ledger_index"])
		validated := false
		if v, ok := resMap["validated"].(bool); ok {
			validated = v
		}

		// If found and validated (or not), return
		if ledgerIndex != 0 {
			return ledgerIndex, validated, nil
		}

		// Wait before next retry
		time.Sleep(interval)
	}

	return 0, false, fmt.Errorf("transaction not found after %d retries", maxRetries)
}

// Helper functions
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result
	default:
		return 0
	}
}
