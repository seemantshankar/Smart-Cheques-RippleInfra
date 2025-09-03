package config

import (
	"os"
	"strconv"
)

// TestWalletConfig contains configuration for test wallets
type TestWalletConfig struct {
	NetworkURL string
	TestNet    bool

	// Test Wallet 1 (Source Account)
	Wallet1Address    string
	Wallet1Secret     string
	Wallet1PrivateKey string
	Wallet1KeyType    string

	// Test Wallet 2 (Destination Account)
	Wallet2Address    string
	Wallet2Secret     string
	Wallet2PrivateKey string
	Wallet2KeyType    string

	// Test Wallet 3 (Additional Test Account)
	Wallet3Address    string
	Wallet3Secret     string
	Wallet3PrivateKey string
	Wallet3KeyType    string

	// Test Transaction Parameters
	TransactionAmount   string
	TransactionCurrency string
	TransactionFee      string

	// Test Monitoring Configuration
	MonitorMaxRetries    int
	MonitorRetryInterval int

	// XRPL Testnet Faucet
	FaucetURL string
}

// LoadTestConfig loads test configuration from environment variables or defaults
func LoadTestConfig() *TestWalletConfig {
	config := &TestWalletConfig{
		// Network Configuration
		NetworkURL: getEnvOrDefault("XRPL_NETWORK_URL", "https://s.altnet.rippletest.net:51234"),
		TestNet:    getEnvBoolOrDefault("XRPL_TESTNET", true),

		// Test Wallet 1 (Source Account)
		Wallet1Address:    getEnvOrDefault("TEST_WALLET_1_ADDRESS", "r3HhM6gecjrzZQXRaLNZnL82K8vxRgdSGe"),
		Wallet1Secret:     getEnvOrDefault("TEST_WALLET_1_SECRET", "sEdVK6HJp45224vWuQCLiXQ93bq2EZm"),
		Wallet1PrivateKey: getEnvOrDefault("TEST_WALLET_1_PRIVATE_KEY", ""),
		Wallet1KeyType:    getEnvOrDefault("TEST_WALLET_1_KEY_TYPE", "ed25519"),

		// Test Wallet 2 (Destination Account)
		Wallet2Address:    getEnvOrDefault("TEST_WALLET_2_ADDRESS", "rabLpuxj8Z2gjy1d6K5t81vBysNoy3mPGk"),
		Wallet2Secret:     getEnvOrDefault("TEST_WALLET_2_SECRET", "sEdVTRMeZzbozpePJz2fmk279AgbcEP"),
		Wallet2PrivateKey: getEnvOrDefault("TEST_WALLET_2_PRIVATE_KEY", ""),
		Wallet2KeyType:    getEnvOrDefault("TEST_WALLET_2_KEY_TYPE", "ed25519"),

		// Test Wallet 3 (Mock Wallet - Not on ripple testnet)
		Wallet3Address:    getEnvOrDefault("TEST_WALLET_3_ADDRESS", "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"),
		Wallet3Secret:     getEnvOrDefault("TEST_WALLET_3_SECRET", "sEdSJHS4oiAdzPwRxs5MHq3HqHrFgMW"),
		Wallet3PrivateKey: getEnvOrDefault("TEST_WALLET_3_PRIVATE_KEY", "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"),
		Wallet3KeyType:    getEnvOrDefault("TEST_WALLET_3_KEY_TYPE", "ed25519"),

		// Test Transaction Parameters
		TransactionAmount:   getEnvOrDefault("TEST_TRANSACTION_AMOUNT", "10"),
		TransactionCurrency: getEnvOrDefault("TEST_TRANSACTION_CURRENCY", "XRP"),
		TransactionFee:      getEnvOrDefault("TEST_TRANSACTION_FEE", "12"),

		// Test Monitoring Configuration
		MonitorMaxRetries:    getEnvIntOrDefault("TEST_MONITOR_MAX_RETRIES", 10),
		MonitorRetryInterval: getEnvIntOrDefault("TEST_MONITOR_RETRY_INTERVAL", 2000),

		// XRPL Testnet Faucet
		FaucetURL: getEnvOrDefault("XRPL_FAUCET_URL", "https://faucet.altnet.rippletest.net/accounts"),
	}

	return config
}

// GetTestWallet1 returns the first test wallet configuration
func (c *TestWalletConfig) GetTestWallet1() (address, secret, privateKey, keyType string) {
	return c.Wallet1Address, c.Wallet1Secret, c.Wallet1PrivateKey, c.Wallet1KeyType
}

// GetTestWallet2 returns the second test wallet configuration
func (c *TestWalletConfig) GetTestWallet2() (address, secret, privateKey, keyType string) {
	return c.Wallet2Address, c.Wallet2Secret, c.Wallet2PrivateKey, c.Wallet2KeyType
}

// GetTestWallet3 returns the third test wallet configuration
func (c *TestWalletConfig) GetTestWallet3() (address, secret, privateKey, keyType string) {
	return c.Wallet3Address, c.Wallet3Secret, c.Wallet3PrivateKey, c.Wallet3KeyType
}

// GetTransactionParams returns the test transaction parameters
func (c *TestWalletConfig) GetTransactionParams() (amount, currency, fee string) {
	return c.TransactionAmount, c.TransactionCurrency, c.TransactionFee
}

// GetMonitoringConfig returns the test monitoring configuration
func (c *TestWalletConfig) GetMonitoringConfig() (maxRetries, retryInterval int) {
	return c.MonitorMaxRetries, c.MonitorRetryInterval
}

// Helper functions for environment variable handling

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
