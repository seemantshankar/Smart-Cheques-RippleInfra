package models

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetType(t *testing.T) {
	tests := []struct {
		name      string
		assetType AssetType
		expected  string
	}{
		{"Native", AssetTypeNative, "native"},
		{"Stablecoin", AssetTypeStablecoin, "stablecoin"},
		{"CBDC", AssetTypeCBDC, "cbdc"},
		{"Wrapped", AssetTypeWrapped, "wrapped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.assetType.String())

			// Test Value() method
			value, err := tt.assetType.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)

			// Test Scan() method
			var scanned AssetType
			err = scanned.Scan(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.assetType, scanned)
		})
	}
}

func TestAssetTransactionType(t *testing.T) {
	tests := []struct {
		name     string
		txType   AssetTransactionType
		expected string
		isDebit  bool
		isCredit bool
	}{
		{"Deposit", AssetTransactionTypeDeposit, "deposit", false, true},
		{"Withdrawal", AssetTransactionTypeWithdrawal, "withdrawal", true, false},
		{"TransferIn", AssetTransactionTypeTransferIn, "transfer_in", false, true},
		{"TransferOut", AssetTransactionTypeTransferOut, "transfer_out", true, false},
		{"EscrowLock", AssetTransactionTypeEscrowLock, "escrow_lock", true, false},
		{"EscrowRelease", AssetTransactionTypeEscrowRelease, "escrow_release", false, true},
		{"Fee", AssetTransactionTypeFee, "fee", true, false},
		{"Adjustment", AssetTransactionTypeAdjustment, "adjustment", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.txType.String())

			// Test Value() method
			value, err := tt.txType.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)

			// Test Scan() method
			var scanned AssetTransactionType
			err = scanned.Scan(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.txType, scanned)

			// Test transaction direction
			transaction := &AssetTransaction{TransactionType: tt.txType}
			assert.Equal(t, tt.isDebit, transaction.IsDebit(), "IsDebit check failed")
			assert.Equal(t, tt.isCredit, transaction.IsCredit(), "IsCredit check failed")
		})
	}
}

func TestAssetTransactionStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   AssetTransactionStatus
		expected string
	}{
		{"Pending", AssetTransactionStatusPending, "pending"},
		{"Processing", AssetTransactionStatusProcessing, "processing"},
		{"Completed", AssetTransactionStatusCompleted, "completed"},
		{"Failed", AssetTransactionStatusFailed, "failed"},
		{"Cancelled", AssetTransactionStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())

			// Test Value() method
			value, err := tt.status.Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)

			// Test Scan() method
			var scanned AssetTransactionStatus
			err = scanned.Scan(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.status, scanned)
		})
	}
}

func TestSupportedAsset(t *testing.T) {
	t.Run("IsNative", func(t *testing.T) {
		nativeAsset := &SupportedAsset{AssetType: AssetTypeNative}
		assert.True(t, nativeAsset.IsNative())

		stablecoinAsset := &SupportedAsset{AssetType: AssetTypeStablecoin}
		assert.False(t, stablecoinAsset.IsNative())
	})

	t.Run("RequiresTrustLine", func(t *testing.T) {
		// Native asset doesn't require trust line
		nativeAsset := &SupportedAsset{AssetType: AssetTypeNative}
		assert.False(t, nativeAsset.RequiresTrustLine())

		// Non-native without issuer doesn't require trust line
		assetWithoutIssuer := &SupportedAsset{AssetType: AssetTypeStablecoin}
		assert.False(t, assetWithoutIssuer.RequiresTrustLine())

		// Non-native with issuer requires trust line
		issuer := "rExampleIssuer123"
		assetWithIssuer := &SupportedAsset{
			AssetType:     AssetTypeStablecoin,
			IssuerAddress: &issuer,
		}
		assert.True(t, assetWithIssuer.RequiresTrustLine())
	})

	t.Run("GetMinimumAmountBigInt", func(t *testing.T) {
		asset := &SupportedAsset{MinimumAmount: "1000000"}
		amount, err := asset.GetMinimumAmountBigInt()
		require.NoError(t, err)

		expected := big.NewInt(1000000)
		assert.Equal(t, expected, amount)

		// Test invalid amount
		invalidAsset := &SupportedAsset{MinimumAmount: "invalid"}
		_, err = invalidAsset.GetMinimumAmountBigInt()
		assert.Error(t, err)
	})
}

func TestEnterpriseBalance(t *testing.T) {
	balance := &EnterpriseBalance{
		ID:               uuid.New(),
		EnterpriseID:     uuid.New(),
		CurrencyCode:     "USDT",
		AvailableBalance: "1000000", // 1 USDT in microunits
		ReservedBalance:  "500000",  // 0.5 USDT in microunits
		TotalBalance:     "1500000", // 1.5 USDT in microunits
	}

	t.Run("GetAvailableBalanceBigInt", func(t *testing.T) {
		amount, err := balance.GetAvailableBalanceBigInt()
		require.NoError(t, err)

		expected := big.NewInt(1000000)
		assert.Equal(t, expected, amount)
	})

	t.Run("GetReservedBalanceBigInt", func(t *testing.T) {
		amount, err := balance.GetReservedBalanceBigInt()
		require.NoError(t, err)

		expected := big.NewInt(500000)
		assert.Equal(t, expected, amount)
	})

	t.Run("GetTotalBalanceBigInt", func(t *testing.T) {
		amount, err := balance.GetTotalBalanceBigInt()
		require.NoError(t, err)

		expected := big.NewInt(1500000)
		assert.Equal(t, expected, amount)
	})

	t.Run("HasSufficientBalance", func(t *testing.T) {
		// Test sufficient balance
		sufficient, err := balance.HasSufficientBalance("500000") // 0.5 USDT
		require.NoError(t, err)
		assert.True(t, sufficient)

		// Test exactly available balance
		sufficient, err = balance.HasSufficientBalance("1000000") // 1 USDT
		require.NoError(t, err)
		assert.True(t, sufficient)

		// Test insufficient balance
		sufficient, err = balance.HasSufficientBalance("1500000") // 1.5 USDT (more than available)
		require.NoError(t, err)
		assert.False(t, sufficient)

		// Test invalid amount
		_, err = balance.HasSufficientBalance("invalid")
		assert.Error(t, err)
	})
}

func TestAssetTransaction(t *testing.T) {
	transaction := &AssetTransaction{
		ID:              uuid.New(),
		EnterpriseID:    uuid.New(),
		CurrencyCode:    "USDT",
		TransactionType: AssetTransactionTypeDeposit,
		Amount:          "1000000", // 1 USDT
		Fee:             "1000",    // 0.001 USDT
		Status:          AssetTransactionStatusCompleted,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	t.Run("GetAmountBigInt", func(t *testing.T) {
		amount, err := transaction.GetAmountBigInt()
		require.NoError(t, err)

		expected := big.NewInt(1000000)
		assert.Equal(t, expected, amount)

		// Test invalid amount
		invalidTransaction := &AssetTransaction{Amount: "invalid"}
		_, err = invalidTransaction.GetAmountBigInt()
		assert.Error(t, err)
	})

	t.Run("GetFeeBigInt", func(t *testing.T) {
		fee, err := transaction.GetFeeBigInt()
		require.NoError(t, err)

		expected := big.NewInt(1000)
		assert.Equal(t, expected, fee)

		// Test empty fee
		emptyFeeTransaction := &AssetTransaction{Fee: ""}
		fee, err = emptyFeeTransaction.GetFeeBigInt()
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(0), fee)

		// Test invalid fee
		invalidTransaction := &AssetTransaction{Fee: "invalid"}
		_, err = invalidTransaction.GetFeeBigInt()
		assert.Error(t, err)
	})

	t.Run("TransactionDirection", func(t *testing.T) {
		assert.False(t, transaction.IsDebit())
		assert.True(t, transaction.IsCredit())

		// Test debit transaction
		debitTransaction := &AssetTransaction{TransactionType: AssetTransactionTypeWithdrawal}
		assert.True(t, debitTransaction.IsDebit())
		assert.False(t, debitTransaction.IsCredit())
	})
}

func TestSupportedCurrencies(t *testing.T) {
	t.Run("IsSupportedCurrency", func(t *testing.T) {
		// Test supported currencies
		assert.True(t, IsSupportedCurrency("XRP"))
		assert.True(t, IsSupportedCurrency("xrp"))
		assert.True(t, IsSupportedCurrency("USDT"))
		assert.True(t, IsSupportedCurrency("usdt"))
		assert.True(t, IsSupportedCurrency("USDC"))
		assert.True(t, IsSupportedCurrency("usdc"))
		assert.True(t, IsSupportedCurrency("e₹"))
		assert.True(t, IsSupportedCurrency("E₹"))

		// Test unsupported currency
		assert.False(t, IsSupportedCurrency("BTC"))
		assert.False(t, IsSupportedCurrency("ETH"))
		assert.False(t, IsSupportedCurrency(""))
	})

	t.Run("GetSupportedCurrency", func(t *testing.T) {
		// Test getting XRP config
		asset, exists := GetSupportedCurrency("XRP")
		assert.True(t, exists)
		assert.Equal(t, "XRP", asset.CurrencyCode)
		assert.Equal(t, AssetTypeNative, asset.AssetType)
		assert.Equal(t, 6, asset.DecimalPlaces)

		// Test getting USDT config
		asset, exists = GetSupportedCurrency("usdt")
		assert.True(t, exists)
		assert.Equal(t, "USDT", asset.CurrencyCode)
		assert.Equal(t, AssetTypeStablecoin, asset.AssetType)
		assert.Equal(t, "10000", asset.MinimumAmount)

		// Test getting e₹ config
		asset, exists = GetSupportedCurrency("e₹")
		assert.True(t, exists)
		assert.Equal(t, "e₹", asset.CurrencyCode)
		assert.Equal(t, AssetTypeCBDC, asset.AssetType)
		assert.Equal(t, 2, asset.DecimalPlaces)

		// Test case insensitive for e₹
		asset, exists = GetSupportedCurrency("E₹")
		assert.True(t, exists)
		assert.Equal(t, "e₹", asset.CurrencyCode)

		// Test unsupported currency
		_, exists = GetSupportedCurrency("BTC")
		assert.False(t, exists)
	})

	t.Run("CurrencyConfigurations", func(t *testing.T) {
		// Verify all expected currencies are configured
		expectedCurrencies := []string{"XRP", "USDT", "USDC", "e₹"}

		for _, currency := range expectedCurrencies {
			asset, exists := GetSupportedCurrency(currency)
			assert.True(t, exists, "Currency %s should be supported", currency)
			assert.Equal(t, currency, asset.CurrencyCode)
			assert.True(t, asset.IsActive)
			assert.NotEmpty(t, asset.CurrencyName)
			assert.NotEmpty(t, asset.MinimumAmount)
			assert.Greater(t, asset.DecimalPlaces, -1)
			assert.LessOrEqual(t, asset.DecimalPlaces, 18)
		}

		// Verify XRP is native
		xrp, _ := GetSupportedCurrency("XRP")
		assert.True(t, xrp.IsNative())
		assert.False(t, xrp.RequiresTrustLine())

		// Verify stablecoins are configured correctly
		usdt, _ := GetSupportedCurrency("USDT")
		assert.Equal(t, AssetTypeStablecoin, usdt.AssetType)
		assert.False(t, usdt.IsNative())

		usdc, _ := GetSupportedCurrency("USDC")
		assert.Equal(t, AssetTypeStablecoin, usdc.AssetType)
		assert.False(t, usdc.IsNative())

		// Verify e₹ is CBDC
		eRupee, _ := GetSupportedCurrency("e₹")
		assert.Equal(t, AssetTypeCBDC, eRupee.AssetType)
		assert.False(t, eRupee.IsNative())
		assert.Equal(t, 2, eRupee.DecimalPlaces) // Currency precision
	})
}

func TestDatabaseInterfaceMethods(t *testing.T) {
	t.Run("AssetType Database Interface", func(t *testing.T) {
		var at AssetType

		// Test scanning valid value
		err := at.Scan("stablecoin")
		require.NoError(t, err)
		assert.Equal(t, AssetTypeStablecoin, at)

		// Test scanning nil
		err = at.Scan(nil)
		require.NoError(t, err)

		// Test scanning invalid type
		err = at.Scan(123)
		assert.Error(t, err)
	})

	t.Run("AssetTransactionType Database Interface", func(t *testing.T) {
		var att AssetTransactionType

		// Test scanning valid value
		err := att.Scan("deposit")
		require.NoError(t, err)
		assert.Equal(t, AssetTransactionTypeDeposit, att)

		// Test scanning nil
		err = att.Scan(nil)
		require.NoError(t, err)

		// Test scanning invalid type
		err = att.Scan(123)
		assert.Error(t, err)
	})

	t.Run("AssetTransactionStatus Database Interface", func(t *testing.T) {
		var ats AssetTransactionStatus

		// Test scanning valid value
		err := ats.Scan("completed")
		require.NoError(t, err)
		assert.Equal(t, AssetTransactionStatusCompleted, ats)

		// Test scanning nil
		err = ats.Scan(nil)
		require.NoError(t, err)

		// Test scanning invalid type
		err = ats.Scan(123)
		assert.Error(t, err)
	})
}

func TestBalanceOperations(t *testing.T) {
	t.Run("ValidBalanceArithmetic", func(t *testing.T) {
		balance := &EnterpriseBalance{
			AvailableBalance: "1000000000", // 1000 units
			ReservedBalance:  "500000000",  // 500 units
			TotalBalance:     "1500000000", // 1500 units
		}

		available, err := balance.GetAvailableBalanceBigInt()
		require.NoError(t, err)

		reserved, err := balance.GetReservedBalanceBigInt()
		require.NoError(t, err)

		total, err := balance.GetTotalBalanceBigInt()
		require.NoError(t, err)

		// Verify balance arithmetic
		calculatedTotal := new(big.Int).Add(available, reserved)
		assert.Equal(t, total, calculatedTotal)
	})

	t.Run("ZeroBalances", func(t *testing.T) {
		balance := &EnterpriseBalance{
			AvailableBalance: "0",
			ReservedBalance:  "0",
			TotalBalance:     "0",
		}

		available, err := balance.GetAvailableBalanceBigInt()
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(0), available)

		sufficient, err := balance.HasSufficientBalance("1")
		require.NoError(t, err)
		assert.False(t, sufficient)
	})
}
