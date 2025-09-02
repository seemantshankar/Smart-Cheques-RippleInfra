package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/models"
)

// createMockCBDCService creates a CBDC service with mock repositories
func createMockCBDCService() *CBDCService {
	mockWalletRepo := NewMockCBDCWalletRepository()
	mockTransactionRepo := NewMockCBDCTransactionRepository()
	mockBalanceRepo := NewMockCBDCBalanceRepository()
	mockRequestRepo := NewMockCBDCWalletRequestRepository()
	mockTSPRepo := NewMockTSPConfigRepository()

	return &CBDCService{
		walletRepo:      mockWalletRepo,
		transactionRepo: mockTransactionRepo,
		balanceRepo:     mockBalanceRepo,
		requestRepo:     mockRequestRepo,
		tspRepo:         mockTSPRepo,
	}
}

func TestCBDCService_CreateWallet(t *testing.T) {
	// Create mock repositories
	mockWalletRepo := NewMockCBDCWalletRepository()
	mockTransactionRepo := NewMockCBDCTransactionRepository()
	mockBalanceRepo := NewMockCBDCBalanceRepository()
	mockRequestRepo := NewMockCBDCWalletRequestRepository()
	mockTSPRepo := NewMockTSPConfigRepository()

	// Create CBDC service with mock dependencies
	service := &CBDCService{
		walletRepo:      mockWalletRepo,
		transactionRepo: mockTransactionRepo,
		balanceRepo:     mockBalanceRepo,
		requestRepo:     mockRequestRepo,
		tspRepo:         mockTSPRepo,
	}

	ctx := context.Background()
	enterpriseID := "test-enterprise-123"
	currency := models.CurrencyERupee

	// Test successful wallet creation
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.Equal(t, enterpriseID, wallet.EnterpriseID)
	assert.Equal(t, currency, wallet.Currency)
	assert.Equal(t, models.CBDCWalletStatusPending, wallet.Status)
	assert.Equal(t, 1000000.0, wallet.Limit)
	assert.NotEmpty(t, wallet.WalletAddress)
	assert.True(t, wallet.CreatedAt.After(time.Now().Add(-time.Second)))
	assert.True(t, wallet.UpdatedAt.After(time.Now().Add(-time.Second)))

	// Test unsupported currency
	_, err = service.CreateWallet(ctx, enterpriseID, models.CurrencyUSDT)
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedCurrency, err)

	// Test duplicate wallet creation
	_, err = service.CreateWallet(ctx, enterpriseID, currency)
	assert.Error(t, err)
	assert.Equal(t, ErrWalletAlreadyExists, err)
}

func TestCBDCService_GetWallet(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-456"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test getting wallet by ID
	retrievedWallet, err := service.GetWallet(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, wallet.ID, retrievedWallet.ID)
	assert.Equal(t, wallet.EnterpriseID, retrievedWallet.EnterpriseID)

	// Test getting non-existent wallet
	_, err = service.GetWallet(ctx, "non-existent-id")
	assert.Error(t, err)
}

func TestCBDCService_GetWalletByEnterprise(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-789"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test getting wallet by enterprise
	retrievedWallet, err := service.GetWalletByEnterprise(ctx, enterpriseID, currency)
	require.NoError(t, err)
	assert.Equal(t, wallet.ID, retrievedWallet.ID)
	assert.Equal(t, enterpriseID, retrievedWallet.EnterpriseID)

	// Test getting non-existent wallet
	_, err = service.GetWalletByEnterprise(ctx, "non-existent-enterprise", currency)
	assert.Error(t, err)
}

func TestCBDCService_UpdateWalletStatus(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-status"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test activating wallet
	err = service.ActivateWallet(ctx, wallet.ID)
	require.NoError(t, err)

	// Verify status was updated
	updatedWallet, err := service.GetWallet(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCWalletStatusActive, updatedWallet.Status)
	assert.NotNil(t, updatedWallet.ActivatedAt)

	// Test suspending wallet
	err = service.SuspendWallet(ctx, wallet.ID, "test suspension")
	require.NoError(t, err)

	// Verify status was updated
	updatedWallet, err = service.GetWallet(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCWalletStatusSuspended, updatedWallet.Status)
	assert.NotNil(t, updatedWallet.SuspendedAt)
}

func TestCBDCService_BalanceManagement(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-balance"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test getting initial balance
	balance, err := service.GetBalance(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, balance.Available)
	assert.Equal(t, 0.0, balance.Reserved)
	assert.Equal(t, 0.0, balance.Total)

	// Test updating balance
	err = service.UpdateBalance(ctx, wallet.ID, 1000.0, 200.0)
	require.NoError(t, err)

	// Verify balance was updated
	updatedBalance, err := service.GetBalance(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, 1000.0, updatedBalance.Available)
	assert.Equal(t, 200.0, updatedBalance.Reserved)
	assert.Equal(t, 1200.0, updatedBalance.Total)

	// Test reserving funds
	err = service.ReserveFunds(ctx, wallet.ID, 300.0)
	require.NoError(t, err)

	// Verify funds were reserved
	balanceAfterReserve, err := service.GetBalance(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, 700.0, balanceAfterReserve.Available)
	assert.Equal(t, 500.0, balanceAfterReserve.Reserved)
	assert.Equal(t, 1200.0, balanceAfterReserve.Total)

	// Test releasing reserved funds
	err = service.ReleaseReservedFunds(ctx, wallet.ID, 200.0)
	require.NoError(t, err)

	// Verify funds were released
	balanceAfterRelease, err := service.GetBalance(ctx, wallet.ID)
	require.NoError(t, err)
	assert.Equal(t, 900.0, balanceAfterRelease.Available)
	assert.Equal(t, 300.0, balanceAfterRelease.Reserved)
	assert.Equal(t, 1200.0, balanceAfterRelease.Total)

	// Test insufficient funds error
	err = service.ReserveFunds(ctx, wallet.ID, 1000.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)

	// Test insufficient reserved funds error
	err = service.ReleaseReservedFunds(ctx, wallet.ID, 500.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientReservedFunds, err)
}

func TestCBDCService_TransactionManagement(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-tx"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test creating transaction
	tx := &models.CBDCTransaction{
		WalletID:    wallet.ID,
		Amount:      500.0,
		Currency:    currency,
		FromAddress: wallet.WalletAddress,
		ToAddress:   "e₹_recipient_123",
		Description: "Test transaction",
	}

	err = service.CreateTransaction(ctx, tx)
	require.NoError(t, err)
	assert.NotEmpty(t, tx.ID)
	assert.Equal(t, models.CBDCTransactionStatusPending, tx.Status)
	assert.True(t, tx.CreatedAt.After(time.Now().Add(-time.Second)))

	// Test getting transaction
	retrievedTx, err := service.GetTransaction(ctx, tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.ID, retrievedTx.ID)
	assert.Equal(t, tx.Amount, retrievedTx.Amount)

	// Test updating transaction status
	err = service.UpdateTransactionStatus(ctx, tx.ID, models.CBDCTransactionStatusProcessing)
	require.NoError(t, err)

	// Verify status was updated
	updatedTx, err := service.GetTransaction(ctx, tx.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCTransactionStatusProcessing, updatedTx.Status)

	// Test confirming transaction
	err = service.ConfirmTransaction(ctx, tx.ID, "tx_hash_123")
	require.NoError(t, err)

	// Verify transaction was confirmed
	confirmedTx, err := service.GetTransaction(ctx, tx.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCTransactionStatusConfirmed, confirmedTx.Status)
	assert.Equal(t, "tx_hash_123", confirmedTx.TransactionHash)
	assert.NotNil(t, confirmedTx.ConfirmedAt)

	// Test failing transaction
	err = service.FailTransaction(ctx, tx.ID, "test failure")
	require.NoError(t, err)

	// Verify transaction was failed
	failedTx, err := service.GetTransaction(ctx, tx.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCTransactionStatusFailed, failedTx.Status)
	assert.Equal(t, "test failure", *failedTx.FailureReason)
	assert.NotNil(t, failedTx.FailedAt)
}

func TestCBDCService_TransactionValidation(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-validation"
	currency := models.CurrencyERupee

	// Create a wallet first
	wallet, err := service.CreateWallet(ctx, enterpriseID, currency)
	require.NoError(t, err)

	// Test invalid wallet ID
	tx := &models.CBDCTransaction{
		WalletID:    "",
		Amount:      100.0,
		Currency:    currency,
		FromAddress: wallet.WalletAddress,
		ToAddress:   "e₹_recipient_123",
	}
	err = service.CreateTransaction(ctx, tx)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidWalletID, err)

	// Test invalid amount
	tx = &models.CBDCTransaction{
		WalletID:    wallet.ID,
		Amount:      -100.0,
		Currency:    currency,
		FromAddress: wallet.WalletAddress,
		ToAddress:   "e₹_recipient_123",
	}
	err = service.CreateTransaction(ctx, tx)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)

	// Test invalid from address
	tx = &models.CBDCTransaction{
		WalletID:    wallet.ID,
		Amount:      100.0,
		Currency:    currency,
		FromAddress: "",
		ToAddress:   "e₹_recipient_123",
	}
	err = service.CreateTransaction(ctx, tx)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFromAddress, err)

	// Test invalid to address
	tx = &models.CBDCTransaction{
		WalletID:    wallet.ID,
		Amount:      100.0,
		Currency:    currency,
		FromAddress: wallet.WalletAddress,
		ToAddress:   "",
	}
	err = service.CreateTransaction(ctx, tx)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToAddress, err)
}

func TestCBDCService_WalletRequests(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()
	enterpriseID := "test-enterprise-requests"
	currency := models.CurrencyERupee

	// Test creating wallet request
	req := &models.CBDCWalletRequest{
		EnterpriseID: enterpriseID,
		Type:         models.CBDCWalletRequestTypeCreate,
		Amount:       nil,
		Currency:     currency,
		Reason:       "Test wallet creation",
	}

	err := service.CreateWalletRequest(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, req.ID)
	assert.Equal(t, models.CBDCRequestStatusPending, req.Status)
	assert.True(t, req.CreatedAt.After(time.Now().Add(-time.Second)))

	// Test getting wallet request
	retrievedReq, err := service.GetWalletRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, req.ID, retrievedReq.ID)
	assert.Equal(t, enterpriseID, retrievedReq.EnterpriseID)

	// Test approving wallet request
	err = service.ApproveWalletRequest(ctx, req.ID, "approver-123")
	require.NoError(t, err)

	// Verify request was approved
	approvedReq, err := service.GetWalletRequest(ctx, req.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCRequestStatusApproved, approvedReq.Status)
	assert.NotNil(t, approvedReq.ApprovedAt)

	// Test rejecting wallet request
	req2 := &models.CBDCWalletRequest{
		EnterpriseID: enterpriseID,
		Type:         models.CBDCWalletRequestTypeSuspend,
		Amount:       nil,
		Currency:     currency,
		Reason:       "Test suspension",
	}

	err = service.CreateWalletRequest(ctx, req2)
	require.NoError(t, err)

	err = service.RejectWalletRequest(ctx, req2.ID, "rejector-123", "test rejection")
	require.NoError(t, err)

	// Verify request was rejected
	rejectedReq, err := service.GetWalletRequest(ctx, req2.ID)
	require.NoError(t, err)
	assert.Equal(t, models.CBDCRequestStatusRejected, rejectedReq.Status)
	assert.Equal(t, "test rejection", *rejectedReq.RejectionReason)
	assert.NotNil(t, rejectedReq.RejectedAt)

	// Test getting pending requests
	pendingRequests, err := service.GetPendingRequests(ctx, enterpriseID)
	require.NoError(t, err)
	// Should be 0 since we approved one and rejected the other
	assert.Len(t, pendingRequests, 0)
}

func TestCBDCService_TSPIntegration(t *testing.T) {
	service := createMockCBDCService()

	ctx := context.Background()

	// Test TSP connection - this should fail since no TSP config exists
	err := service.TestTSPConnection(ctx, "test-tsp-123")
	assert.Error(t, err) // Should fail because TSP config doesn't exist
}
