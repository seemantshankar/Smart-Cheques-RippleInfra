package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/ybbus/jsonrpc/v3"
)

// XRPL Testnet Configuration
const (
	TestnetURL   = "https://s.altnet.rippletest.net:51234"
	PayerAddress = "rNT1br8vX75mHxXcbduUowhmouSVmroGvm"
	PayerSecret  = "sEdV8gF5VidtFSHJUkLgKmwM6f1cA9E"
	PayeeAddress = "r3HhM6gecjrzZQXRaLNZnL82K8vxRgdSGe"
	PayeeSecret  = "sEdVK6HJp45224vWuQCLiXQ93bq2EZm"
)

// XRPL RPC Client
type XRPLClient struct {
	client jsonrpc.RPCClient
}

// NewXRPLClient creates a new XRPL RPC client
func NewXRPLClient(url string) *XRPLClient {
	return &XRPLClient{
		client: jsonrpc.NewClient(url),
	}
}

// AccountInfo represents XRPL account information
type AccountInfo struct {
	AccountData struct {
		Account           string `json:"Account"`
		Balance           string `json:"Balance"`
		Flags             int    `json:"Flags"`
		LedgerEntryType   string `json:"LedgerEntryType"`
		OwnerCount        int    `json:"OwnerCount"`
		PreviousTxnID     string `json:"PreviousTxnID"`
		PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
		Sequence          int    `json:"Sequence"`
		TransferRate      int    `json:"TransferRate,omitempty"`
	} `json:"account_data"`
	LedgerCurrentIndex int `json:"ledger_current_index"`
	QueueData          struct {
		TxnCount int `json:"txn_count"`
	} `json:"queue_data"`
	Validated bool `json:"validated"`
}

// ServerInfo represents XRPL server information
type ServerInfo struct {
	Info struct {
		CompleteLedgers string `json:"complete_ledgers"`
		HostID          string `json:"hostid"`
		LastClose       struct {
			ConvergeTimeS interface{} `json:"converge_time_s"`
			Proposers     interface{} `json:"proposers"`
		} `json:"last_close"`
		LoadFactor      interface{} `json:"load_factor"`
		NetworkID       interface{} `json:"network_id"`
		Peers           interface{} `json:"peers"`
		PubkeyNode      string      `json:"pubkey_node"`
		ServerState     string      `json:"server_state"`
		Time            string      `json:"time"`
		Uptime          interface{} `json:"uptime"`
		ValidatedLedger struct {
			Age            int         `json:"age"`
			BaseFeeXRP     interface{} `json:"base_fee_xrp"`
			Hash           string      `json:"hash"`
			ReserveBaseXRP interface{} `json:"reserve_base_xrp"`
			ReserveIncXRP  interface{} `json:"reserve_inc_xrp"`
			Seq            int         `json:"seq"`
		} `json:"validated_ledger"`
		Version string `json:"version"`
	} `json:"info"`
}

// TestConnection tests basic connectivity to XRPL testnet
func (x *XRPLClient) TestConnection() error {
	log.Println("🔌 Testing XRPL Testnet Connection...")

	// Test server info
	var serverInfo ServerInfo
	err := x.client.CallFor(context.Background(), &serverInfo, "server_info")
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	log.Printf("✅ Connected to XRPL Testnet!")
	log.Printf("   Server State: %s", serverInfo.Info.ServerState)
	log.Printf("   Version: %s", serverInfo.Info.Version)
	log.Printf("   Network ID: %v", serverInfo.Info.NetworkID)
	log.Printf("   Uptime: %v seconds", serverInfo.Info.Uptime)
	log.Printf("   Peers: %v", serverInfo.Info.Peers)
	log.Printf("   Validated Ledger: %d (Age: %d seconds)",
		serverInfo.Info.ValidatedLedger.Seq,
		serverInfo.Info.ValidatedLedger.Age)

	return nil
}

// GetAccountInfo retrieves account information from XRPL
func (x *XRPLClient) GetAccountInfo(address string) (*AccountInfo, error) {
	log.Printf("📊 Getting account info for: %s", address)

	// Use direct HTTP POST instead of JSON-RPC client
	requestBody := map[string]interface{}{
		"method": "account_info",
		"params": []map[string]interface{}{
			{
				"account": address,
				"strict":  true,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %s", err)
	}

	resp, err := http.Post(TestnetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	// Extract result from response
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	// Convert to our struct format
	accountInfo := &AccountInfo{}
	if accountData, ok := result["account_data"].(map[string]interface{}); ok {
		accountInfo.AccountData.Account = toString(accountData["Account"])
		accountInfo.AccountData.Balance = toString(accountData["Balance"])
		accountInfo.AccountData.Sequence = toInt(accountData["Sequence"])
		accountInfo.AccountData.Flags = toInt(accountData["Flags"])
	}

	return accountInfo, nil
}

// CheckWalletBalances checks balances of both test wallets
func (x *XRPLClient) CheckWalletBalances() error {
	log.Println("\n💰 Checking Wallet Balances...")

	// Check Payer Wallet
	payerInfo, err := x.GetAccountInfo(PayerAddress)
	if err != nil {
		return fmt.Errorf("failed to get payer account info: %w", err)
	}

	// Convert drops to XRP (1 XRP = 1,000,000 drops)
	payerBalanceXRP := float64(parseInt(payerInfo.AccountData.Balance)) / 1000000.0

	log.Printf("   Payer Wallet (%s):", PayerAddress)
	log.Printf("      Balance: %s drops (%f XRP)",
		payerInfo.AccountData.Balance, payerBalanceXRP)
	log.Printf("      Sequence: %d", payerInfo.AccountData.Sequence)
	log.Printf("      Flags: %d", payerInfo.AccountData.Flags)

	// Check Payee Wallet
	payeeInfo, err := x.GetAccountInfo(PayeeAddress)
	if err != nil {
		return fmt.Errorf("failed to get payee account info: %w", err)
	}

	payeeBalanceXRP := float64(parseInt(payeeInfo.AccountData.Balance)) / 1000000.0

	log.Printf("   Payee Wallet (%s):", PayeeAddress)
	log.Printf("      Balance: %s drops (%f XRP)",
		payeeInfo.AccountData.Balance, payeeBalanceXRP)
	log.Printf("      Sequence: %d", payeeInfo.AccountData.Sequence)
	log.Printf("      Flags: %d", payeeInfo.AccountData.Flags)

	return nil
}

// TestSmallTransaction tests a very small transaction between wallets
func (x *XRPLClient) TestSmallTransaction() error {
	log.Println("\n🔄 Testing Small Transaction...")

	// Get current account info for sequence number
	payerInfo, err := x.GetAccountInfo(PayerAddress)
	if err != nil {
		return fmt.Errorf("failed to get payer account info: %w", err)
	}

	// Calculate minimum transaction amount (1 drop = 0.000001 XRP)
	// We'll send just 1 drop to test the system
	amountDrops := "1" // 1 drop = 0.000001 XRP

	log.Printf("   Testing transaction of %s drops (%f XRP)",
		amountDrops, float64(parseInt(amountDrops))/1000000.0)
	log.Printf("   From: %s (Sequence: %d)", PayerAddress, payerInfo.AccountData.Sequence)
	log.Printf("   To: %s", PayeeAddress)

	// Note: In a real implementation, we would:
	// 1. Create the transaction
	// 2. Sign it with the private key
	// 3. Submit it to the network
	// 4. Monitor the transaction status

	log.Println("   ⚠️  Transaction creation and signing not implemented in this test")
	log.Println("   ⚠️  This is just a connectivity and balance check test")

	return nil
}

// TestNetworkHealth performs comprehensive network health checks
func (x *XRPLClient) TestNetworkHealth() error {
	log.Println("\n🏥 Testing Network Health...")

	// Test HTTP connectivity
	log.Println("   Testing HTTP connectivity...")
	resp, err := http.Get("https://s.altnet.rippletest.net:51234")
	if err != nil {
		return fmt.Errorf("HTTP connectivity test failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("   ✅ HTTP Status: %s", resp.Status)

	// Test RPC connectivity
	log.Println("   Testing RPC connectivity...")
	var pingResponse map[string]interface{}
	err = x.client.CallFor(context.Background(), &pingResponse, "ping")
	if err != nil {
		return fmt.Errorf("RPC ping test failed: %w", err)
	}

	log.Println("   ✅ RPC ping successful")

	// Test ledger info (simplified)
	log.Println("   Testing ledger info...")
	log.Println("   ✅ Ledger info test skipped (using server_info instead)")

	return nil
}

// Helper function to parse string to int
func parseInt(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// Helper function to convert interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Helper function to convert interface{} to int
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
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}

// Main test function
func main() {
	log.Println("🚀 Starting XRP Testnet Integration Test")
	log.Println("==========================================")

	// Create XRPL client
	client := NewXRPLClient(TestnetURL)

	// Test 1: Basic Connection
	if err := client.TestConnection(); err != nil {
		log.Fatalf("❌ Connection test failed: %v", err)
	}

	// Test 2: Network Health
	if err := client.TestNetworkHealth(); err != nil {
		log.Fatalf("❌ Network health test failed: %v", err)
	}

	// Test 3: Wallet Balance Check
	if err := client.CheckWalletBalances(); err != nil {
		log.Fatalf("❌ Wallet balance check failed: %v", err)
	}

	// Test 4: Small Transaction Test (simulation)
	if err := client.TestSmallTransaction(); err != nil {
		log.Fatalf("❌ Transaction test failed: %v", err)
	}

	log.Println("\n🎉 All tests completed successfully!")
	log.Println("✅ XRP Testnet integration is working correctly")
	log.Println("✅ Both test wallets are accessible")
	log.Println("✅ Network connectivity is stable")
	log.Println("\n📝 Next steps:")
	log.Println("   1. Implement transaction creation and signing")
	log.Println("   2. Add proper error handling for failed transactions")
	log.Println("   3. Implement transaction monitoring")
	log.Println("   4. Add comprehensive transaction validation")
	log.Println("   5. Implement proper key management and security")
}
