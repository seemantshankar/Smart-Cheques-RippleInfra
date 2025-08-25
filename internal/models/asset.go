package models

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AssetType represents different types of supported assets
type AssetType string

const (
	AssetTypeNative     AssetType = "native"
	AssetTypeStablecoin AssetType = "stablecoin"
	AssetTypeCBDC       AssetType = "cbdc"
	AssetTypeWrapped    AssetType = "wrapped"
)

// String returns string representation of AssetType
func (at AssetType) String() string {
	return string(at)
}

// Value implements driver.Valuer interface for database storage
func (at AssetType) Value() (driver.Value, error) {
	return string(at), nil
}

// Scan implements sql.Scanner interface for database retrieval
func (at *AssetType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		*at = AssetType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into AssetType", value)
}

// AssetTransactionType represents different types of asset transactions
type AssetTransactionType string

const (
	AssetTransactionTypeDeposit       AssetTransactionType = "deposit"
	AssetTransactionTypeWithdrawal    AssetTransactionType = "withdrawal"
	AssetTransactionTypeTransferIn    AssetTransactionType = "transfer_in"
	AssetTransactionTypeTransferOut   AssetTransactionType = "transfer_out"
	AssetTransactionTypeEscrowLock    AssetTransactionType = "escrow_lock"
	AssetTransactionTypeEscrowRelease AssetTransactionType = "escrow_release"
	AssetTransactionTypeMint          AssetTransactionType = "mint"
	AssetTransactionTypeBurn          AssetTransactionType = "burn"
	AssetTransactionTypeFee           AssetTransactionType = "fee"
	AssetTransactionTypeAdjustment    AssetTransactionType = "adjustment"
)

// String returns string representation of AssetTransactionType
func (att AssetTransactionType) String() string {
	return string(att)
}

// Value implements driver.Valuer interface
func (att AssetTransactionType) Value() (driver.Value, error) {
	return string(att), nil
}

// Scan implements sql.Scanner interface
func (att *AssetTransactionType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		*att = AssetTransactionType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into AssetTransactionType", value)
}

// AssetTransactionStatus represents the status of an asset transaction
type AssetTransactionStatus string

const (
	AssetTransactionStatusPending    AssetTransactionStatus = "pending"
	AssetTransactionStatusProcessing AssetTransactionStatus = "processing"
	AssetTransactionStatusCompleted  AssetTransactionStatus = "completed"
	AssetTransactionStatusFailed     AssetTransactionStatus = "failed"
	AssetTransactionStatusCancelled  AssetTransactionStatus = "cancelled"
)

// String returns string representation of AssetTransactionStatus
func (ats AssetTransactionStatus) String() string {
	return string(ats)
}

// Value implements driver.Valuer interface
func (ats AssetTransactionStatus) Value() (driver.Value, error) {
	return string(ats), nil
}

// Scan implements sql.Scanner interface
func (ats *AssetTransactionStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if str, ok := value.(string); ok {
		*ats = AssetTransactionStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into AssetTransactionStatus", value)
}

// SupportedAsset represents a currency/asset supported by the platform
type SupportedAsset struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CurrencyCode string    `json:"currency_code" db:"currency_code" validate:"required,min=1,max=10"`
	CurrencyName string    `json:"currency_name" db:"currency_name" validate:"required,min=1,max=100"`
	AssetType    AssetType `json:"asset_type" db:"asset_type" validate:"required"`

	// XRPL specific fields
	IssuerAddress  *string `json:"issuer_address,omitempty" db:"issuer_address"`
	CurrencyHex    *string `json:"currency_hex,omitempty" db:"currency_hex"`
	TrustLineLimit *string `json:"trust_line_limit,omitempty" db:"trust_line_limit"`
	TransferFee    float64 `json:"transfer_fee" db:"transfer_fee"`
	GlobalFreeze   bool    `json:"global_freeze" db:"global_freeze"`
	NoFreeze       bool    `json:"no_freeze" db:"no_freeze"`

	// Configuration
	DecimalPlaces int     `json:"decimal_places" db:"decimal_places" validate:"min=0,max=18"`
	MinimumAmount string  `json:"minimum_amount" db:"minimum_amount" validate:"required"`
	MaximumAmount *string `json:"maximum_amount,omitempty" db:"maximum_amount"`
	IsActive      bool    `json:"is_active" db:"is_active"`

	// Metadata
	Description      *string `json:"description,omitempty" db:"description"`
	IconURL          *string `json:"icon_url,omitempty" db:"icon_url"`
	DocumentationURL *string `json:"documentation_url,omitempty" db:"documentation_url"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsNative returns true if this is a native blockchain asset (like XRP)
func (sa *SupportedAsset) IsNative() bool {
	return sa.AssetType == AssetTypeNative
}

// RequiresTrustLine returns true if this asset requires a trust line on XRPL
func (sa *SupportedAsset) RequiresTrustLine() bool {
	return !sa.IsNative() && sa.IssuerAddress != nil
}

// GetMinimumAmountBigInt returns the minimum amount as *big.Int
func (sa *SupportedAsset) GetMinimumAmountBigInt() (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(sa.MinimumAmount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid minimum amount: %s", sa.MinimumAmount)
	}
	return amount, nil
}

// EnterpriseBalance represents the balance of an enterprise for a specific currency
type EnterpriseBalance struct {
	ID           uuid.UUID `json:"id" db:"id"`
	EnterpriseID uuid.UUID `json:"enterprise_id" db:"enterprise_id" validate:"required"`
	CurrencyCode string    `json:"currency_code" db:"currency_code" validate:"required"`

	// Balance tracking (stored as strings to avoid precision issues)
	AvailableBalance string `json:"available_balance" db:"available_balance"`
	ReservedBalance  string `json:"reserved_balance" db:"reserved_balance"`
	TotalBalance     string `json:"total_balance" db:"total_balance"`

	// XRPL synchronization
	XRPLBalance  string     `json:"xrpl_balance" db:"xrpl_balance"`
	LastXRPLSync *time.Time `json:"last_xrpl_sync,omitempty" db:"last_xrpl_sync"`

	// Limits and controls
	DailyLimit           *string `json:"daily_limit,omitempty" db:"daily_limit"`
	MonthlyLimit         *string `json:"monthly_limit,omitempty" db:"monthly_limit"`
	MaxTransactionAmount *string `json:"max_transaction_amount,omitempty" db:"max_transaction_amount"`

	// Status
	IsFrozen          bool       `json:"is_frozen" db:"is_frozen"`
	FreezeReason      *string    `json:"freeze_reason,omitempty" db:"freeze_reason"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty" db:"last_transaction_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data (populated via joins)
	Enterprise     *Enterprise     `json:"enterprise,omitempty"`
	SupportedAsset *SupportedAsset `json:"supported_asset,omitempty"`
}

// GetAvailableBalanceBigInt returns available balance as *big.Int
func (eb *EnterpriseBalance) GetAvailableBalanceBigInt() (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(eb.AvailableBalance, 10)
	if !ok {
		return nil, fmt.Errorf("invalid available balance: %s", eb.AvailableBalance)
	}
	return amount, nil
}

// GetReservedBalanceBigInt returns reserved balance as *big.Int
func (eb *EnterpriseBalance) GetReservedBalanceBigInt() (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(eb.ReservedBalance, 10)
	if !ok {
		return nil, fmt.Errorf("invalid reserved balance: %s", eb.ReservedBalance)
	}
	return amount, nil
}

// GetTotalBalanceBigInt returns total balance as *big.Int
func (eb *EnterpriseBalance) GetTotalBalanceBigInt() (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(eb.TotalBalance, 10)
	if !ok {
		return nil, fmt.Errorf("invalid total balance: %s", eb.TotalBalance)
	}
	return amount, nil
}

// HasSufficientBalance checks if there's enough available balance for a transaction
func (eb *EnterpriseBalance) HasSufficientBalance(amount string) (bool, error) {
	available, err := eb.GetAvailableBalanceBigInt()
	if err != nil {
		return false, err
	}

	required := new(big.Int)
	_, ok := required.SetString(amount, 10)
	if !ok {
		return false, fmt.Errorf("invalid amount: %s", amount)
	}

	return available.Cmp(required) >= 0, nil
}

// AssetTransaction represents a transaction affecting enterprise balances
type AssetTransaction struct {
	ID              uuid.UUID            `json:"id" db:"id"`
	EnterpriseID    uuid.UUID            `json:"enterprise_id" db:"enterprise_id" validate:"required"`
	CurrencyCode    string               `json:"currency_code" db:"currency_code" validate:"required"`
	TransactionType AssetTransactionType `json:"transaction_type" db:"transaction_type" validate:"required"`

	// Transaction details
	Amount         string  `json:"amount" db:"amount" validate:"required"`
	Fee            string  `json:"fee" db:"fee"`
	ReferenceID    *string `json:"reference_id,omitempty" db:"reference_id"`
	ExternalTxHash *string `json:"external_tx_hash,omitempty" db:"external_tx_hash"`

	// Balance tracking
	BalanceBefore string `json:"balance_before" db:"balance_before"`
	BalanceAfter  string `json:"balance_after" db:"balance_after"`

	// Status and metadata
	Status      AssetTransactionStatus `json:"status" db:"status"`
	Description *string                `json:"description,omitempty" db:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	// Approval workflow
	ApprovedBy *uuid.UUID `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at,omitempty" db:"approved_at"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`

	// Related data
	Enterprise     *Enterprise     `json:"enterprise,omitempty"`
	SupportedAsset *SupportedAsset `json:"supported_asset,omitempty"`
}

// GetAmountBigInt returns transaction amount as *big.Int
func (at *AssetTransaction) GetAmountBigInt() (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(at.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", at.Amount)
	}
	return amount, nil
}

// GetFeeBigInt returns transaction fee as *big.Int
func (at *AssetTransaction) GetFeeBigInt() (*big.Int, error) {
	if at.Fee == "" {
		return big.NewInt(0), nil
	}
	fee := new(big.Int)
	_, ok := fee.SetString(at.Fee, 10)
	if !ok {
		return nil, fmt.Errorf("invalid fee: %s", at.Fee)
	}
	return fee, nil
}

// IsDebit returns true if this transaction debits the enterprise balance
func (at *AssetTransaction) IsDebit() bool {
	switch at.TransactionType {
	case AssetTransactionTypeWithdrawal, AssetTransactionTypeTransferOut, AssetTransactionTypeEscrowLock, AssetTransactionTypeFee:
		return true
	default:
		return false
	}
}

// IsCredit returns true if this transaction credits the enterprise balance
func (at *AssetTransaction) IsCredit() bool {
	switch at.TransactionType {
	case AssetTransactionTypeDeposit, AssetTransactionTypeTransferIn, AssetTransactionTypeEscrowRelease:
		return true
	default:
		return false
	}
}

// EnterpriseBalanceSummary represents aggregated balance information for reporting
type EnterpriseBalanceSummary struct {
	EnterpriseID      uuid.UUID  `json:"enterprise_id" db:"enterprise_id"`
	EnterpriseName    string     `json:"enterprise_name" db:"enterprise_name"`
	CurrencyCode      string     `json:"currency_code" db:"currency_code"`
	CurrencyName      string     `json:"currency_name" db:"currency_name"`
	AvailableBalance  string     `json:"available_balance" db:"available_balance"`
	ReservedBalance   string     `json:"reserved_balance" db:"reserved_balance"`
	TotalBalance      string     `json:"total_balance" db:"total_balance"`
	XRPLBalance       string     `json:"xrpl_balance" db:"xrpl_balance"`
	IsFrozen          bool       `json:"is_frozen" db:"is_frozen"`
	LastTransactionAt *time.Time `json:"last_transaction_at,omitempty" db:"last_transaction_at"`
	LastXRPLSync      *time.Time `json:"last_xrpl_sync,omitempty" db:"last_xrpl_sync"`
}

// CurrencyConfig holds configuration for supported currencies
var SupportedCurrencies = map[string]SupportedAsset{
	"XRP": {
		CurrencyCode:  "XRP",
		CurrencyName:  "XRP",
		AssetType:     AssetTypeNative,
		DecimalPlaces: 6,
		MinimumAmount: "1",
		IsActive:      true,
		Description:   stringPtr("Native XRP cryptocurrency"),
	},
	"USDT": {
		CurrencyCode:  "USDT",
		CurrencyName:  "Tether USD",
		AssetType:     AssetTypeStablecoin,
		DecimalPlaces: 6,
		MinimumAmount: "10000", // 0.01 USDT in microunits
		IsActive:      true,
		Description:   stringPtr("USD-pegged stablecoin issued by Tether"),
	},
	"USDC": {
		CurrencyCode:  "USDC",
		CurrencyName:  "USD Coin",
		AssetType:     AssetTypeStablecoin,
		DecimalPlaces: 6,
		MinimumAmount: "10000", // 0.01 USDC in microunits
		IsActive:      true,
		Description:   stringPtr("USD-pegged stablecoin issued by Centre"),
	},
	"e₹": {
		CurrencyCode:  "e₹",
		CurrencyName:  "Digital Rupee",
		AssetType:     AssetTypeCBDC,
		DecimalPlaces: 2,
		MinimumAmount: "1", // 0.01 INR in paisa
		IsActive:      true,
		Description:   stringPtr("Central Bank Digital Currency issued by Reserve Bank of India"),
	},
}

// helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// IsSupportedCurrency checks if a currency code is supported
func IsSupportedCurrency(currencyCode string) bool {
	currencyCode = strings.ToUpper(currencyCode)
	if currencyCode == "E₹" {
		currencyCode = "e₹" // Handle the special case
	}
	_, exists := SupportedCurrencies[currencyCode]
	return exists
}

// GetSupportedCurrency returns configuration for a supported currency
func GetSupportedCurrency(currencyCode string) (SupportedAsset, bool) {
	currencyCode = strings.ToUpper(currencyCode)
	if currencyCode == "E₹" {
		currencyCode = "e₹" // Handle the special case
	}
	asset, exists := SupportedCurrencies[currencyCode]
	return asset, exists
}
