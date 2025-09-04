package config

import (
	"fmt"
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

// GetTestWallet1 returns the first test wallet configuration with generated private key
func (c *TestWalletConfig) GetTestWallet1() (address, secret, privateKey, keyType string) {
	addr, sec, priv, typ, err := c.GetWalletWithPrivateKey(c.Wallet1Address, c.Wallet1Secret, c.Wallet1KeyType)
	if err != nil {
		// Return original values if generation fails (for backward compatibility)
		return c.Wallet1Address, c.Wallet1Secret, c.Wallet1PrivateKey, c.Wallet1KeyType
	}
	return addr, sec, priv, typ
}

// GetTestWallet2 returns the second test wallet configuration with generated private key
func (c *TestWalletConfig) GetTestWallet2() (address, secret, privateKey, keyType string) {
	addr, sec, priv, typ, err := c.GetWalletWithPrivateKey(c.Wallet2Address, c.Wallet2Secret, c.Wallet2KeyType)
	if err != nil {
		// Return original values if generation fails (for backward compatibility)
		return c.Wallet2Address, c.Wallet2Secret, c.Wallet2PrivateKey, c.Wallet2KeyType
	}
	return addr, sec, priv, typ
}

// GetTestWallet3 returns the third test wallet configuration with generated private key
func (c *TestWalletConfig) GetTestWallet3() (address, secret, privateKey, keyType string) {
	addr, sec, priv, typ, err := c.GetWalletWithPrivateKey(c.Wallet3Address, c.Wallet3Secret, c.Wallet3KeyType)
	if err != nil {
		// Return original values if generation fails (for backward compatibility)
		return c.Wallet3Address, c.Wallet3Secret, c.Wallet3PrivateKey, c.Wallet3KeyType
	}
	return addr, sec, priv, typ
}

// GetTransactionParams returns the test transaction parameters
func (c *TestWalletConfig) GetTransactionParams() (amount, currency, fee string) {
	return c.TransactionAmount, c.TransactionCurrency, c.TransactionFee
}

// GetMonitoringConfig returns the test monitoring configuration
func (c *TestWalletConfig) GetMonitoringConfig() (maxRetries, retryInterval int) {
	return c.MonitorMaxRetries, c.MonitorRetryInterval
}

// GeneratePrivateKeyFromSecret generates private key hex from XRPL secret
// Note: For now, we'll return the secret directly as the service expects private keys
// In a production implementation, this would extract the actual private key from the wallet
func GeneratePrivateKeyFromSecret(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	// For now, return the secret as-is since the current service implementation
	// may expect secrets rather than extracted private keys
	return secret, nil
}

// GetWalletWithPrivateKey returns wallet info with generated private key
func (c *TestWalletConfig) GetWalletWithPrivateKey(address, secret, keyType string) (addressOut, secretOut, privateKey, keyTypeOut string, err error) {
	if secret != "" {
		privateKey, err = GeneratePrivateKeyFromSecret(secret)
		if err != nil {
			return "", "", "", "", fmt.Errorf("failed to generate private key for %s: %w", address, err)
		}
	}
	return address, secret, privateKey, keyType, nil
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
