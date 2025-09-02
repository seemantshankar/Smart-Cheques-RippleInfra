package services

import (
	"context"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repositories"
)

// CBDCServiceInterface defines the interface for CBDC operations
type CBDCServiceInterface interface {
	// Wallet Management
	CreateWallet(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error)
	GetWallet(ctx context.Context, walletID string) (*models.CBDCWallet, error)
	GetWalletByEnterprise(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error)
	UpdateWalletStatus(ctx context.Context, walletID string, status models.CBDCWalletStatus) error
	SuspendWallet(ctx context.Context, walletID string, reason string) error
	ActivateWallet(ctx context.Context, walletID string) error
	CloseWallet(ctx context.Context, walletID string, reason string) error

	// Balance Management
	GetBalance(ctx context.Context, walletID string) (*models.CBDCBalance, error)
	UpdateBalance(ctx context.Context, walletID string, available, reserved float64) error
	ReserveFunds(ctx context.Context, walletID string, amount float64) error
	ReleaseReservedFunds(ctx context.Context, walletID string, amount float64) error

	// Transaction Management
	CreateTransaction(ctx context.Context, tx *models.CBDCTransaction) error
	GetTransaction(ctx context.Context, transactionID string) (*models.CBDCTransaction, error)
	GetWalletTransactions(ctx context.Context, walletID string, limit, offset int) ([]*models.CBDCTransaction, error)
	UpdateTransactionStatus(ctx context.Context, transactionID string, status models.CBDCTransactionStatus) error
	ConfirmTransaction(ctx context.Context, transactionID string, hash string) error
	FailTransaction(ctx context.Context, transactionID string, reason string) error

	// Request Management
	CreateWalletRequest(ctx context.Context, req *models.CBDCWalletRequest) error
	GetWalletRequest(ctx context.Context, requestID string) (*models.CBDCWalletRequest, error)
	ApproveWalletRequest(ctx context.Context, requestID string, approverID string) error
	RejectWalletRequest(ctx context.Context, requestID string, approverID string, reason string) error
	GetPendingRequests(ctx context.Context, enterpriseID string) ([]*models.CBDCWalletRequest, error)

	// TSP Integration
	GetTSPConfig(ctx context.Context, tspID string) (*models.TSPConfig, error)
	UpdateTSPConfig(ctx context.Context, config *models.TSPConfig) error
	TestTSPConnection(ctx context.Context, tspID string) error
}

// CBDCService implements the CBDC service interface
type CBDCService struct {
	walletRepo      repositories.CBDCWalletRepositoryInterface
	transactionRepo repositories.CBDCTransactionRepositoryInterface
	balanceRepo     repositories.CBDCBalanceRepositoryInterface
	requestRepo     repositories.CBDCWalletRequestRepositoryInterface
	tspRepo         repositories.TSPConfigRepositoryInterface
}

// NewCBDCService creates a new CBDC service instance
func NewCBDCService(
	walletRepo repositories.CBDCWalletRepositoryInterface,
	transactionRepo repositories.CBDCTransactionRepositoryInterface,
	balanceRepo repositories.CBDCBalanceRepositoryInterface,
	requestRepo repositories.CBDCWalletRequestRepositoryInterface,
	tspRepo repositories.TSPConfigRepositoryInterface,
) *CBDCService {
	return &CBDCService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		balanceRepo:     balanceRepo,
		requestRepo:     requestRepo,
		tspRepo:         tspRepo,
	}
}

// CreateWallet creates a new CBDC wallet for an enterprise
func (s *CBDCService) CreateWallet(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error) {
	// Validate currency
	if currency != models.CurrencyERupee {
		return nil, ErrUnsupportedCurrency
	}

	// Check if wallet already exists
	existingWallet, err := s.walletRepo.GetByEnterprise(ctx, enterpriseID, currency)
	if err == nil && existingWallet != nil {
		return nil, ErrWalletAlreadyExists
	}

	// Generate wallet address (mock implementation)
	walletAddress := generateWalletAddress()

	wallet := &models.CBDCWallet{
		EnterpriseID:  enterpriseID,
		WalletAddress: walletAddress,
		Currency:      currency,
		Status:        models.CBDCWalletStatusPending,
		Balance:       0,
		Limit:         1000000, // Default limit of 1M e₹
	}

	// Create wallet in repository
	err = s.walletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, err
	}

	// Initialize balance
	balance := &models.CBDCBalance{
		WalletID:    wallet.ID,
		Available:   0,
		Reserved:    0,
		Total:       0,
		Currency:    currency,
		LastUpdated: wallet.CreatedAt,
	}

	err = s.balanceRepo.Create(ctx, balance)
	if err != nil {
		// Rollback wallet creation
		_ = s.walletRepo.Delete(ctx, wallet.ID)
		return nil, err
	}

	return wallet, nil
}

// GetWallet retrieves a CBDC wallet by ID
func (s *CBDCService) GetWallet(ctx context.Context, walletID string) (*models.CBDCWallet, error) {
	return s.walletRepo.GetByID(ctx, walletID)
}

// GetWalletByEnterprise retrieves a CBDC wallet by enterprise ID and currency
func (s *CBDCService) GetWalletByEnterprise(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error) {
	return s.walletRepo.GetByEnterprise(ctx, enterpriseID, currency)
}

// UpdateWalletStatus updates the status of a CBDC wallet
func (s *CBDCService) UpdateWalletStatus(ctx context.Context, walletID string, status models.CBDCWalletStatus) error {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return err
	}

	wallet.Status = status
	wallet.UpdatedAt = time.Now()

	// Set timestamp based on status
	switch status {
	case models.CBDCWalletStatusActive:
		now := time.Now()
		wallet.ActivatedAt = &now
	case models.CBDCWalletStatusSuspended:
		now := time.Now()
		wallet.SuspendedAt = &now
	}

	return s.walletRepo.Update(ctx, wallet)
}

// SuspendWallet suspends a CBDC wallet
func (s *CBDCService) SuspendWallet(ctx context.Context, walletID string, reason string) error {
	return s.UpdateWalletStatus(ctx, walletID, models.CBDCWalletStatusSuspended)
}

// ActivateWallet activates a CBDC wallet
func (s *CBDCService) ActivateWallet(ctx context.Context, walletID string) error {
	return s.UpdateWalletStatus(ctx, walletID, models.CBDCWalletStatusActive)
}

// CloseWallet closes a CBDC wallet
func (s *CBDCService) CloseWallet(ctx context.Context, walletID string, reason string) error {
	return s.UpdateWalletStatus(ctx, walletID, models.CBDCWalletStatusClosed)
}

// GetBalance retrieves the balance of a CBDC wallet
func (s *CBDCService) GetBalance(ctx context.Context, walletID string) (*models.CBDCBalance, error) {
	return s.balanceRepo.GetByWalletID(ctx, walletID)
}

// UpdateBalance updates the balance of a CBDC wallet
func (s *CBDCService) UpdateBalance(ctx context.Context, walletID string, available, reserved float64) error {
	balance, err := s.balanceRepo.GetByWalletID(ctx, walletID)
	if err != nil {
		return err
	}

	balance.Available = available
	balance.Reserved = reserved
	balance.Total = available + reserved
	balance.LastUpdated = time.Now()

	return s.balanceRepo.Update(ctx, balance)
}

// ReserveFunds reserves funds in a CBDC wallet
func (s *CBDCService) ReserveFunds(ctx context.Context, walletID string, amount float64) error {
	balance, err := s.balanceRepo.GetByWalletID(ctx, walletID)
	if err != nil {
		return err
	}

	if balance.Available < amount {
		return ErrInsufficientFunds
	}

	balance.Available -= amount
	balance.Reserved += amount
	balance.Total = balance.Available + balance.Reserved
	balance.LastUpdated = time.Now()

	return s.balanceRepo.Update(ctx, balance)
}

// ReleaseReservedFunds releases reserved funds in a CBDC wallet
func (s *CBDCService) ReleaseReservedFunds(ctx context.Context, walletID string, amount float64) error {
	balance, err := s.balanceRepo.GetByWalletID(ctx, walletID)
	if err != nil {
		return err
	}

	if balance.Reserved < amount {
		return ErrInsufficientReservedFunds
	}

	balance.Reserved -= amount
	balance.Available += amount
	balance.Total = balance.Available + balance.Reserved
	balance.LastUpdated = time.Now()

	return s.balanceRepo.Update(ctx, balance)
}

// CreateTransaction creates a new CBDC transaction
func (s *CBDCService) CreateTransaction(ctx context.Context, tx *models.CBDCTransaction) error {
	// Validate transaction
	if err := s.validateTransaction(tx); err != nil {
		return err
	}

	// Set initial status
	tx.Status = models.CBDCTransactionStatusPending
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	return s.transactionRepo.Create(ctx, tx)
}

// GetTransaction retrieves a CBDC transaction by ID
func (s *CBDCService) GetTransaction(ctx context.Context, transactionID string) (*models.CBDCTransaction, error) {
	return s.transactionRepo.GetByID(ctx, transactionID)
}

// GetWalletTransactions retrieves transactions for a wallet
func (s *CBDCService) GetWalletTransactions(ctx context.Context, walletID string, limit, offset int) ([]*models.CBDCTransaction, error) {
	return s.transactionRepo.GetByWalletID(ctx, walletID, limit, offset)
}

// UpdateTransactionStatus updates the status of a CBDC transaction
func (s *CBDCService) UpdateTransactionStatus(ctx context.Context, transactionID string, status models.CBDCTransactionStatus) error {
	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	tx.Status = status
	tx.UpdatedAt = time.Now()

	// Set timestamp based on status
	switch status {
	case models.CBDCTransactionStatusConfirmed:
		now := time.Now()
		tx.ConfirmedAt = &now
	case models.CBDCTransactionStatusFailed:
		now := time.Now()
		tx.FailedAt = &now
	}

	return s.transactionRepo.Update(ctx, tx)
}

// ConfirmTransaction confirms a CBDC transaction
func (s *CBDCService) ConfirmTransaction(ctx context.Context, transactionID string, hash string) error {
	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	tx.TransactionHash = hash
	tx.Status = models.CBDCTransactionStatusConfirmed
	tx.UpdatedAt = time.Now()
	now := time.Now()
	tx.ConfirmedAt = &now

	return s.transactionRepo.Update(ctx, tx)
}

// FailTransaction marks a CBDC transaction as failed
func (s *CBDCService) FailTransaction(ctx context.Context, transactionID string, reason string) error {
	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	tx.Status = models.CBDCTransactionStatusFailed
	tx.FailureReason = &reason
	tx.UpdatedAt = time.Now()
	now := time.Now()
	tx.FailedAt = &now

	return s.transactionRepo.Update(ctx, tx)
}

// CreateWalletRequest creates a new wallet request
func (s *CBDCService) CreateWalletRequest(ctx context.Context, req *models.CBDCWalletRequest) error {
	req.Status = models.CBDCRequestStatusPending
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	return s.requestRepo.Create(ctx, req)
}

// GetWalletRequest retrieves a wallet request by ID
func (s *CBDCService) GetWalletRequest(ctx context.Context, requestID string) (*models.CBDCWalletRequest, error) {
	return s.requestRepo.GetByID(ctx, requestID)
}

// ApproveWalletRequest approves a wallet request
func (s *CBDCService) ApproveWalletRequest(ctx context.Context, requestID string, approverID string) error {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	req.Status = models.CBDCRequestStatusApproved
	req.UpdatedAt = time.Now()
	now := time.Now()
	req.ApprovedAt = &now

	return s.requestRepo.Update(ctx, req)
}

// RejectWalletRequest rejects a wallet request
func (s *CBDCService) RejectWalletRequest(ctx context.Context, requestID string, approverID string, reason string) error {
	req, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	req.Status = models.CBDCRequestStatusRejected
	req.RejectionReason = &reason
	req.UpdatedAt = time.Now()
	now := time.Now()
	req.RejectedAt = &now

	return s.requestRepo.Update(ctx, req)
}

// GetPendingRequests retrieves pending requests for an enterprise
func (s *CBDCService) GetPendingRequests(ctx context.Context, enterpriseID string) ([]*models.CBDCWalletRequest, error) {
	return s.requestRepo.GetPendingByEnterprise(ctx, enterpriseID)
}

// GetTSPConfig retrieves TSP configuration
func (s *CBDCService) GetTSPConfig(ctx context.Context, tspID string) (*models.TSPConfig, error) {
	return s.tspRepo.GetByID(ctx, tspID)
}

// UpdateTSPConfig updates TSP configuration
func (s *CBDCService) UpdateTSPConfig(ctx context.Context, config *models.TSPConfig) error {
	config.UpdatedAt = time.Now()
	return s.tspRepo.Update(ctx, config)
}

// TestTSPConnection tests the connection to a TSP
func (s *CBDCService) TestTSPConnection(ctx context.Context, tspID string) error {
	config, err := s.tspRepo.GetByID(ctx, tspID)
	if err != nil {
		return err
	}

	// Mock TSP connection test
	if config.Status == models.TSPStatusError {
		return ErrTSPConnectionFailed
	}

	return nil
}

// validateTransaction validates a CBDC transaction
func (s *CBDCService) validateTransaction(tx *models.CBDCTransaction) error {
	if tx.WalletID == "" {
		return ErrInvalidWalletID
	}
	if tx.Amount <= 0 {
		return ErrInvalidAmount
	}
	if tx.FromAddress == "" {
		return ErrInvalidFromAddress
	}
	if tx.ToAddress == "" {
		return ErrInvalidToAddress
	}
	return nil
}

// generateWalletAddress generates a mock wallet address
func generateWalletAddress() string {
	// Mock implementation - in production this would integrate with TSP
	return "e₹_" + generateRandomString(32)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
