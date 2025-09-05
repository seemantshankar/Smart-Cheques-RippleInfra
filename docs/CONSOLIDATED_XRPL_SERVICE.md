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

The consolidated XRPL service is tested through the comprehensive integration test:
```bash
go run test_comprehensive_xrpl.go
```

This test covers:
- Wallet funding via XRPL testnet faucet
- Real payment transactions  
- Escrow creation with crypto-conditions
- Full transaction lifecycle on XRPL testnet

### **Test Results Summary**

**✅ SUCCESSFUL OPERATIONS:**
- **Wallet Funding**: Successfully funded wallets with 10 XRP each via testnet faucet
- **Payment Transactions**: 0.1 XRP payments working correctly with proper fees
- **Escrow Creation**: 0.2 XRP escrows with DER-encoded conditions successful
- **Amount Handling**: Proper conversion from drops strings to xrpl-go types
- **Condition Generation**: Crypto-conditions library producing valid XRPL conditions

**📊 VERIFIED TRANSACTION EXAMPLES:**
```
✅ Payment Transaction: 857E5FCBAD2DF23E1E7A1D6A76E5C231B87D8184F16A254880A5168936D9BEEB
   - Amount: 100000 drops (0.1 XRP)
   - Fee: 12 drops
   - Result: tesSUCCESS

✅ Escrow Transaction: 314E3E730402915FCDAFD42C4337812B11FABD1006868E6D9A6A60240BF93160
   - Amount: 200000 drops (0.2 XRP)  
   - Condition: A025802026D7D43C574EF5169EDCE18862A4B5B3F46059189990F4A451BBEF6EFD65BF26810120
   - Result: tesSUCCESS
```

## Dependencies

- `github.com/smart-payment-infrastructure/pkg/xrpl` - Enhanced XRPL client
- `github.com/smart-payment-infrastructure/internal/models` - Data models
- Standard Go libraries: `fmt`, `log`, `strconv`, `time`

## Key Learnings & Common Issues

### **Amount Type Requirements**

**CRITICAL**: XRPL expects specific amount formats depending on the library layer:

```go
// ❌ WRONG: Raw string passed to xrpl-go types
Amount: "200000" // This causes type conversion errors

// ✅ CORRECT: Convert string to int64 for xrpl-go library
amountInt, err := strconv.ParseInt(amount, 10, 64)
if err != nil {
    return nil, fmt.Errorf("failed to parse amount: %w", err)
}
escrow.Amount = types.XRPCurrencyAmount(amountInt)
```

**Key Points:**
- Input amounts should be strings representing drops (1 XRP = 1,000,000 drops)
- xrpl-go library's `types.XRPCurrencyAmount` expects `int64`
- The library internally converts back to proper JSON string format for XRPL
- Example: "200000" drops = 0.2 XRP

### **Crypto-Conditions Implementation**

**CRITICAL**: XRPL requires proper DER-encoded crypto-conditions, not raw SHA-256 hashes.

#### ❌ **Common Mistakes**
```go
// Raw SHA-256 hash - causes temMALFORMED
hash := sha256.Sum256(preimage)
condition := hex.EncodeToString(hash[:])

// Manual condition format - often incorrect
condition := "A0258020" + hashHex + "8114" + hashHex + "80"
```

#### ✅ **Correct Implementation**
```go
import cc "github.com/go-interledger/cryptoconditions"

func createXRPLConditionAndFulfillment(preimage []byte) (string, string, error) {
    // Create PREIMAGE-SHA-256 fulfillment using crypto-conditions library
    fulfillment := cc.NewPreimageSha256(preimage)
    
    // Serialize to binary (DER-encoded)
    fulfillmentBinary, err := fulfillment.Encode()
    if err != nil {
        return "", "", fmt.Errorf("failed to encode fulfillment: %w", err)
    }
    
    // Generate condition from fulfillment
    condition := fulfillment.Condition()
    conditionBinary, err := condition.Encode()
    if err != nil {
        return "", "", fmt.Errorf("failed to encode condition: %w", err)
    }
    
    // Convert to hexadecimal strings for XRPL
    conditionHex := fmt.Sprintf("%X", conditionBinary)
    fulfillmentHex := fmt.Sprintf("%X", fulfillmentBinary)
    
    return conditionHex, fulfillmentHex, nil
}
```

**Required Dependency:**
```bash
go get github.com/go-interledger/cryptoconditions
```

### **Transaction Autofill Requirements**

**CRITICAL**: Address type conversion required before Autofill

```go
// ❌ WRONG: Autofill with types.Address causes panic
flattenedTx := escrow.Flatten()
client.Autofill(&flattenedTx) // PANIC: interface conversion error

// ✅ CORRECT: Convert addresses to strings first
flattenedTx := escrow.Flatten()

// Convert types.Address to strings for Autofill compatibility
if addr, ok := flattenedTx["Account"].(types.Address); ok {
    flattenedTx["Account"] = string(addr)
}
if addr, ok := flattenedTx["Destination"].(types.Address); ok {
    flattenedTx["Destination"] = string(addr)
}

// Now Autofill works correctly
if err := client.Autofill(&flattenedTx); err != nil {
    return nil, fmt.Errorf("failed to autofill transaction: %w", err)
}
```

### **Common Error Codes & Solutions**

#### **temMALFORMED**
- **Cause**: Incorrect condition format, wrong amount type, or malformed transaction fields
- **Solution**: Use proper DER-encoded conditions and correct amount types

#### **tecNO_TARGET**
- **Cause**: Escrow object not found (wrong sequence number or timing issues)
- **Solution**: Use correct escrow sequence from creation transaction

#### **Interface Conversion Panics**
- **Cause**: xrpl-go Autofill expects string addresses, not types.Address
- **Solution**: Convert address types before calling Autofill

### **Best Practices**

#### **1. Amount Handling**
```go
// Always validate and convert amounts properly
func formatAmountForXRPL(amount string) (int64, error) {
    amountInt, err := strconv.ParseInt(amount, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid amount format: %w", err)
    }
    return amountInt, nil
}
```

#### **2. Condition Generation**
```go
// Always use crypto-conditions library for proper DER encoding
func generateSecureCondition() (string, string, error) {
    // Generate secure random 32-byte preimage
    preimage := make([]byte, 32)
    if _, err := rand.Read(preimage); err != nil {
        return "", "", err
    }
    
    return createXRPLConditionAndFulfillment(preimage)
}
```

#### **3. Transaction Processing**
```go
// Always use proper sequence management
func createEscrowTransaction(client *rpc.Client, escrow *transaction.EscrowCreate) error {
    flattenedTx := escrow.Flatten()
    
    // Convert address types before Autofill
    convertAddressTypes(&flattenedTx)
    
    // Use Autofill for proper sequence, fee, and ledger sequence
    if err := client.Autofill(&flattenedTx); err != nil {
        return fmt.Errorf("autofill failed: %w", err)
    }
    
    // Sign and submit transaction
    return submitTransaction(client, flattenedTx)
}
```

### **Debugging Tips**

#### **1. Enable Debug Output**
```go
// Add debug output to understand transaction structure
fmt.Printf("Debug: Transaction fields:\n")
for key, value := range flattenedTx {
    fmt.Printf("  %s: %v (type: %T)\n", key, value, value)
}
```

#### **2. Validate Conditions**
```go
// Verify condition length and format
fmt.Printf("Condition length: %d characters\n", len(condition))
fmt.Printf("Expected format: DER-encoded hex string\n")
```

#### **3. Check Network Responses**
```go
// Log actual XRPL responses for debugging
log.Printf("XRPL submit response: %+v", response)
```

## Troubleshooting Guide

### **Error Scenario 1: Type Conversion Failures**

**Symptoms:**
```
cannot convert amount (variable of type string) to type XRPCurrencyAmount
```

**Root Cause:** Passing string directly to `types.XRPCurrencyAmount`

**Solution:**
```go
// Convert string to int64 first
amountInt, err := strconv.ParseInt(amountStr, 10, 64)
if err != nil {
    return fmt.Errorf("invalid amount: %w", err)
}
escrow.Amount = types.XRPCurrencyAmount(amountInt)
```

### **Error Scenario 2: temMALFORMED with Conditions**

**Symptoms:**
```
transaction failed to submit with engine result: temMALFORMED
```

**Root Cause:** Using raw SHA-256 hash instead of proper DER-encoded condition

**Solution:**
```go
// Replace manual condition generation
import cc "github.com/go-interledger/cryptoconditions"

fulfillment := cc.NewPreimageSha256(preimage)
conditionBinary, _ := fulfillment.Condition().Encode()
condition := fmt.Sprintf("%X", conditionBinary)
```

### **Error Scenario 3: Interface Conversion Panics**

**Symptoms:**
```
panic: interface conversion: interface {} is types.Address, not string
```

**Root Cause:** Autofill expects string addresses, not types.Address

**Solution:**
```go
// Convert address types before Autofill
if addr, ok := flattenedTx["Account"].(types.Address); ok {
    flattenedTx["Account"] = string(addr)
}
if addr, ok := flattenedTx["Destination"].(types.Address); ok {
    flattenedTx["Destination"] = string(addr)
}

// Now safe to call Autofill
err := client.Autofill(&flattenedTx)
```

### **Error Scenario 4: tecNO_TARGET on EscrowFinish**

**Symptoms:**
```
transaction failed to submit with engine result: tecNO_TARGET
```

**Root Cause:** Incorrect escrow sequence number or escrow not ready

**Solutions:**
1. Use actual sequence from EscrowCreate transaction
2. Ensure FinishAfter time has passed
3. Verify escrow still exists and hasn't been cancelled

### **Validation Checklist**

Before submitting XRPL transactions, verify:

- [ ] **Amount Format**: String input converted to int64 for xrpl-go
- [ ] **Condition Format**: DER-encoded using crypto-conditions library  
- [ ] **Address Types**: Converted to strings before Autofill
- [ ] **Sequence Numbers**: Current account sequence obtained from network
- [ ] **Timing**: FinishAfter/CancelAfter times properly calculated
- [ ] **Network**: Connected to correct XRPL network (testnet/mainnet)

### **Performance Considerations**

- **Connection Pooling**: Reuse XRPL client connections
- **Batch Processing**: Group multiple operations when possible
- **Error Retry**: Implement exponential backoff for network errors
- **Rate Limiting**: Respect XRPL network rate limits

## Future Enhancements

- **Batch Operations**: Support for multiple transactions
- **Advanced Milestones**: Complex milestone dependencies and conditions
- **Performance Monitoring**: Metrics and performance tracking
- **Caching**: Intelligent caching for frequently accessed data
- **Rate Limiting**: Built-in rate limiting for API calls
