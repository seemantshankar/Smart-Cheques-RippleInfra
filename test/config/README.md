# Test Configuration Setup

This directory contains configuration for XRP integration tests, including test wallet credentials and network settings.

## Configuration Files

### `test_wallets.go`
The main configuration file that loads test wallet credentials and settings. It supports both environment variables and default values.

### `.env.example` (Create this file manually)
A template for environment variables. Copy this structure to create your own `.env.test` file:

```bash
# XRPL Network Configuration
XRPL_NETWORK_URL=https://s.altnet.rippletest.net:51234
XRPL_TESTNET=true

# Test Wallet 1 (Source Account)
TEST_WALLET_1_ADDRESS=rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh
TEST_WALLET_1_SECRET=sEdTM1uX8pu2do5XvTnutH6HsouMaM2
TEST_WALLET_1_PRIVATE_KEY=1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
TEST_WALLET_1_KEY_TYPE=secp256k1

# Test Wallet 2 (Destination Account)
TEST_WALLET_2_ADDRESS=r9cZA1nxHfJfKLYGhos6kHE3Q66ZUFJD4X
TEST_WALLET_2_SECRET=sEdV19BLYQo9bXUUFu2NVFtxoaQqwN2
TEST_WALLET_2_PRIVATE_KEY=abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
TEST_WALLET_2_KEY_TYPE=secp256k1

# Test Wallet 3 (Additional Test Account - ed25519)
TEST_WALLET_3_ADDRESS=rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH
TEST_WALLET_3_SECRET=sEdSJHS4oiAdzPwRxs5MHq3HqHrFgMW
TEST_WALLET_3_PRIVATE_KEY=fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321
TEST_WALLET_3_KEY_TYPE=ed25519

# Test Transaction Parameters
TEST_TRANSACTION_AMOUNT=10
TEST_TRANSACTION_CURRENCY=XRP
TEST_TRANSACTION_FEE=12

# Test Monitoring Configuration
TEST_MONITOR_MAX_RETRIES=10
TEST_MONITOR_RETRY_INTERVAL=2000

# XRPL Testnet Faucet
XRPL_FAUCET_URL=https://faucet.altnet.rippletest.net/accounts
```

## Getting Test Wallets

### 1. XRPL Testnet Faucet
Visit [https://faucet.altnet.rippletest.net/accounts](https://faucet.altnet.rippletest.net/accounts) to get test wallets with test XRP.

### 2. Generate New Wallets
You can also generate new test wallets using the XRPL testnet tools or libraries.

## Setup Instructions

1. **Create Environment File**:
   ```bash
   cd test/config
   cp .env.example .env.test
   ```

2. **Customize Credentials**:
   - Replace wallet addresses with your actual test wallet addresses
   - Replace secrets with your actual test wallet secrets
   - Replace private keys with your actual test wallet private keys
   - Adjust transaction parameters as needed

3. **Set Environment Variables** (Alternative):
   ```bash
   export TEST_WALLET_1_ADDRESS="your_wallet_address"
   export TEST_WALLET_1_SECRET="your_wallet_secret"
   export TEST_WALLET_1_PRIVATE_KEY="your_private_key"
   # ... repeat for other wallets
   ```

## Security Notes

- **Never use test credentials in production**
- **Keep your test credentials secure**
- **Test wallets are for development/testing only**
- **Test XRP has no real value**

## Usage in Tests

The configuration is automatically loaded in tests:

```go
import "github.com/smart-payment-infrastructure/test/config"

func TestExample(t *testing.T) {
    testConfig := config.LoadTestConfig()
    
    // Get wallet 1 configuration
    address, secret, privateKey, keyType := testConfig.GetTestWallet1()
    
    // Get transaction parameters
    amount, currency, fee := testConfig.GetTransactionParams()
    
    // Use in your tests...
}
```

## Configuration Methods

The configuration supports multiple ways to set values:

1. **Environment Variables**: Set `TEST_WALLET_1_ADDRESS` etc.
2. **Default Values**: Hardcoded fallbacks for development
3. **Runtime Override**: Modify values programmatically

## Testing Different Scenarios

The configuration includes multiple wallet types to test different scenarios:

- **Wallet 1 & 2**: secp256k1 keys for traditional XRPL operations
- **Wallet 3**: ed25519 keys for modern, high-performance operations
- **Different Addresses**: Test various transaction patterns
- **Configurable Parameters**: Adjust amounts, fees, and monitoring settings
