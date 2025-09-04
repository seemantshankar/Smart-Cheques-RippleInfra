# Consolidated XRPL Service

The `ConsolidatedXRPLService` provides a unified interface for all XRPL (XRP Ledger) operations, consolidating functionality from multiple existing services into a single, maintainable service.

## Overview

This service implements the `XRPLServiceInterface` and provides a comprehensive set of methods for:
- Wallet management and creation
- Account operations and validation
- Smart Cheque escrow management
- Payment transactions
- Transaction monitoring
- Health checks and network validation

## Key Features

### 1. **Unified Interface**
- Single service implementing all XRPL operations
- Consistent error handling and logging
- Standardized method signatures

### 2. **Enhanced Client Integration**
- Built on top of the `EnhancedClient` for robust XRPL interactions
- Supports both HTTP and WebSocket connections
- Automatic connection management and health checks

### 3. **Smart Cheque Support**
- Milestone-based escrow creation
- Conditional escrow with validation
- Oracle configuration support
- Dynamic timing based on milestone durations

### 4. **Security Features**
- Private key management for transaction signing
- Escrow condition and fulfillment generation
- Secure wallet creation and management

## Usage

### Basic Setup

```go
import "github.com/smart-payment-infrastructure/internal/services"

// Create configuration
config := services.ConsolidatedXRPLConfig{
    NetworkURL:   "https://s.altnet.rippletest.net:51234",
    WebSocketURL: "wss://s.altnet.rippletest.net:51233",
    TestNet:      true,
}

// Create and initialize service
service := services.NewConsolidatedXRPLService(config)
if err := service.Initialize(); err != nil {
    log.Fatal(err)
}
```

### Wallet Operations

```go
// Create a new wallet
wallet, err := service.CreateWallet()
if err != nil {
    log.Fatal(err)
}

// Create secp256k1 wallet
secpWallet, err := service.CreateSecp256k1Wallet()

// Create funded account (testnet)
account, err := service.CreateAccount()
```

### Smart Cheque Operations

```go
// Create escrow with private key
result, fulfillment, err := service.CreateSmartChequeEscrowWithKey(
    payerAddress,
    payeeAddress,
    100.0,
    "XRP",
    milestoneSecret,
    privateKeyHex,
)

// Complete milestone with private key
result, err := service.CompleteSmartChequeMilestoneWithKey(
    payeeAddress,
    ownerAddress,
    sequence,
    condition,
    fulfillment,
    privateKeyHex,
)

// Cancel escrow
result, err := service.CancelSmartChequeWithKey(
    accountAddress,
    ownerAddress,
    sequence,
    privateKeyHex,
)
```

### Account Management

```go
// Get account information
accountInfo, err := service.GetAccountInfo(address)

// Get account balance
balance, err := service.GetAccountBalance(address)

// Validate account on network
exists, err := service.ValidateAccountOnNetwork(address)

// Check account balance
hasBalance, err := service.ValidateAccountWithBalance(address, minBalanceDrops)
```

### Payment Operations

```go
// Submit payment
result, err := service.SubmitPayment(
    senderWallet,
    recipientAddress,
    amount,
    currency,
)

// Monitor transaction
status, err := service.MonitorTransaction(
    transactionID,
    maxRetries,
    retryInterval,
)
```

## Method Reference

### Core Methods
- `Initialize()` - Initialize the service and establish connections
- `HealthCheck()` - Check service health and network connectivity
- `CreateWallet()` - Generate new XRPL wallet
- `CreateSecp256k1Wallet()` - Generate secp256k1 wallet
- `CreateAccount()` - Create funded testnet account

### Smart Cheque Methods
- `CreateSmartChequeEscrowWithKey()` - Create escrow with private key
- `CreateSmartChequeEscrowWithMilestones()` - Create escrow with milestone conditions
- `CompleteSmartChequeMilestoneWithKey()` - Complete milestone with private key
- `CancelSmartChequeWithKey()` - Cancel escrow with private key
- `GetEscrowStatus()` - Get escrow status from ledger

### Account Methods
- `GetAccountInfo()` - Get account information
- `GetAccountData()` - Get structured account data
- `GetAccountBalance()` - Get account balance
- `ValidateAddress()` - Validate XRPL address format
- `ValidateAccountOnNetwork()` - Check if account exists
- `ValidateAccountWithBalance()` - Validate account balance

### Utility Methods
- `GenerateCondition()` - Generate escrow condition and fulfillment
- `formatAmount()` - Format amount for XRPL (XRP to drops)
- `getLedgerTimeOffset()` - Convert duration to ledger time

## Migration Guide

### From Existing Services

1. **Replace service initialization:**
   ```go
   // Old
   xrplService := services.NewXRPLService()
   
   // New
   config := services.ConsolidatedXRPLConfig{
       NetworkURL:   "https://s.altnet.rippletest.net:51234",
       WebSocketURL: "wss://s.altnet.rippletest.net:51233",
       TestNet:      true,
   }
   xrplService := services.NewConsolidatedXRPLService(config)
   ```

2. **Update method calls:**
   ```go
   // Old methods that required WalletInfo now use string addresses + private keys
   // Use the "WithKey" versions for operations requiring private keys
   ```

3. **Initialize the service:**
   ```go
   if err := xrplService.Initialize(); err != nil {
       log.Fatal(err)
   }
   ```

## Configuration

### Network Configuration
- **TestNet**: Use testnet URLs for development and testing
- **MainNet**: Use mainnet URLs for production (requires proper credentials)
- **Custom Networks**: Support for custom XRPL node configurations

### Connection Settings
- **HTTP Endpoint**: For REST API calls
- **WebSocket Endpoint**: For real-time updates and subscriptions
- **Health Check**: Automatic connection validation

## Error Handling

The service provides consistent error handling:
- All methods return descriptive error messages
- Network errors are wrapped with context
- Validation errors include specific field information
- Connection errors are automatically retried where appropriate

## Testing

Run the service tests:
```bash
go test ./internal/services/consolidated_xrpl_service_test.go ./internal/services/consolidated_xrpl_service.go
```

## Dependencies

- `github.com/smart-payment-infrastructure/pkg/xrpl` - Enhanced XRPL client
- `github.com/smart-payment-infrastructure/internal/models` - Data models
- Standard Go libraries: `fmt`, `log`, `strconv`, `time`

## Future Enhancements

- **Batch Operations**: Support for multiple transactions
- **Advanced Milestones**: Complex milestone dependencies and conditions
- **Performance Monitoring**: Metrics and performance tracking
- **Caching**: Intelligent caching for frequently accessed data
- **Rate Limiting**: Built-in rate limiting for API calls
