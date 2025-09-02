package services

import (
	"context"
	"fmt"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

// MockCBDCWalletRepository implements CBDCWalletRepositoryInterface
type MockCBDCWalletRepository struct {
	wallets map[string]*models.CBDCWallet
}

func NewMockCBDCWalletRepository() *MockCBDCWalletRepository {
	return &MockCBDCWalletRepository{
		wallets: make(map[string]*models.CBDCWallet),
	}
}

func (m *MockCBDCWalletRepository) Create(ctx context.Context, wallet *models.CBDCWallet) error {
	wallet.ID = generateMockID("wallet")
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	m.wallets[wallet.ID] = wallet
	return nil
}

func (m *MockCBDCWalletRepository) GetByID(ctx context.Context, id string) (*models.CBDCWallet, error) {
	wallet, exists := m.wallets[id]
	if !exists {
		return nil, fmt.Errorf("wallet not found: %s", id)
	}
	return wallet, nil
}

func (m *MockCBDCWalletRepository) GetByEnterprise(ctx context.Context, enterpriseID string, currency models.Currency) (*models.CBDCWallet, error) {
	for _, wallet := range m.wallets {
		if wallet.EnterpriseID == enterpriseID && wallet.Currency == currency {
			return wallet, nil
		}
	}
	return nil, fmt.Errorf("wallet not found for enterprise: %s, currency: %s", enterpriseID, currency)
}

func (m *MockCBDCWalletRepository) Update(ctx context.Context, wallet *models.CBDCWallet) error {
	wallet.UpdatedAt = time.Now()
	m.wallets[wallet.ID] = wallet
	return nil
}

func (m *MockCBDCWalletRepository) Delete(ctx context.Context, id string) error {
	delete(m.wallets, id)
	return nil
}

func (m *MockCBDCWalletRepository) ListByEnterprise(ctx context.Context, enterpriseID string) ([]*models.CBDCWallet, error) {
	var wallets []*models.CBDCWallet
	for _, wallet := range m.wallets {
		if wallet.EnterpriseID == enterpriseID {
			wallets = append(wallets, wallet)
		}
	}
	return wallets, nil
}

func (m *MockCBDCWalletRepository) ListByStatus(ctx context.Context, status models.CBDCWalletStatus) ([]*models.CBDCWallet, error) {
	var wallets []*models.CBDCWallet
	for _, wallet := range m.wallets {
		if wallet.Status == status {
			wallets = append(wallets, wallet)
		}
	}
	return wallets, nil
}

// MockCBDCTransactionRepository implements CBDCTransactionRepositoryInterface
type MockCBDCTransactionRepository struct {
	transactions map[string]*models.CBDCTransaction
}

func NewMockCBDCTransactionRepository() *MockCBDCTransactionRepository {
	return &MockCBDCTransactionRepository{
		transactions: make(map[string]*models.CBDCTransaction),
	}
}

func (m *MockCBDCTransactionRepository) Create(ctx context.Context, transaction *models.CBDCTransaction) error {
	transaction.ID = generateMockID("tx")
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockCBDCTransactionRepository) GetByID(ctx context.Context, id string) (*models.CBDCTransaction, error) {
	tx, exists := m.transactions[id]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", id)
	}
	return tx, nil
}

func (m *MockCBDCTransactionRepository) GetByWalletID(ctx context.Context, walletID string, limit, offset int) ([]*models.CBDCTransaction, error) {
	var transactions []*models.CBDCTransaction
	count := 0

	for _, tx := range m.transactions {
		if tx.WalletID == walletID {
			if count >= offset {
				transactions = append(transactions, tx)
				if len(transactions) >= limit {
					break
				}
			}
			count++
		}
	}

	return transactions, nil
}

func (m *MockCBDCTransactionRepository) Update(ctx context.Context, transaction *models.CBDCTransaction) error {
	transaction.UpdatedAt = time.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockCBDCTransactionRepository) Delete(ctx context.Context, id string) error {
	delete(m.transactions, id)
	return nil
}

func (m *MockCBDCTransactionRepository) ListByStatus(ctx context.Context, status models.CBDCTransactionStatus) ([]*models.CBDCTransaction, error) {
	var transactions []*models.CBDCTransaction
	for _, tx := range m.transactions {
		if tx.Status == status {
			transactions = append(transactions, tx)
		}
	}
	return transactions, nil
}

func (m *MockCBDCTransactionRepository) ListByType(ctx context.Context, txType models.CBDCTransactionType) ([]*models.CBDCTransaction, error) {
	var transactions []*models.CBDCTransaction
	for _, tx := range m.transactions {
		if tx.Type == txType {
			transactions = append(transactions, tx)
		}
	}
	return transactions, nil
}

// MockCBDCBalanceRepository implements CBDCBalanceRepositoryInterface
type MockCBDCBalanceRepository struct {
	balances map[string]*models.CBDCBalance
}

func NewMockCBDCBalanceRepository() *MockCBDCBalanceRepository {
	return &MockCBDCBalanceRepository{
		balances: make(map[string]*models.CBDCBalance),
	}
}

func (m *MockCBDCBalanceRepository) Create(ctx context.Context, balance *models.CBDCBalance) error {
	balance.ID = generateMockID("balance")
	balance.LastUpdated = time.Now()
	m.balances[balance.WalletID] = balance
	return nil
}

func (m *MockCBDCBalanceRepository) GetByWalletID(ctx context.Context, walletID string) (*models.CBDCBalance, error) {
	balance, exists := m.balances[walletID]
	if !exists {
		return nil, fmt.Errorf("balance not found for wallet: %s", walletID)
	}
	return balance, nil
}

func (m *MockCBDCBalanceRepository) Update(ctx context.Context, balance *models.CBDCBalance) error {
	balance.LastUpdated = time.Now()
	m.balances[balance.WalletID] = balance
	return nil
}

func (m *MockCBDCBalanceRepository) Delete(ctx context.Context, id string) error {
	// Find and remove balance by wallet ID
	for walletID, balance := range m.balances {
		if balance.ID == id {
			delete(m.balances, walletID)
			break
		}
	}
	return nil
}

// MockCBDCWalletRequestRepository implements CBDCWalletRequestRepositoryInterface
type MockCBDCWalletRequestRepository struct {
	requests map[string]*models.CBDCWalletRequest
}

func NewMockCBDCWalletRequestRepository() *MockCBDCWalletRequestRepository {
	return &MockCBDCWalletRequestRepository{
		requests: make(map[string]*models.CBDCWalletRequest),
	}
}

func (m *MockCBDCWalletRequestRepository) Create(ctx context.Context, request *models.CBDCWalletRequest) error {
	request.ID = generateMockID("req")
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()
	m.requests[request.ID] = request
	return nil
}

func (m *MockCBDCWalletRequestRepository) GetByID(ctx context.Context, id string) (*models.CBDCWalletRequest, error) {
	req, exists := m.requests[id]
	if !exists {
		return nil, fmt.Errorf("request not found: %s", id)
	}
	return req, nil
}

func (m *MockCBDCWalletRequestRepository) Update(ctx context.Context, request *models.CBDCWalletRequest) error {
	request.UpdatedAt = time.Now()
	m.requests[request.ID] = request
	return nil
}

func (m *MockCBDCWalletRequestRepository) Delete(ctx context.Context, id string) error {
	delete(m.requests, id)
	return nil
}

func (m *MockCBDCWalletRequestRepository) GetPendingByEnterprise(ctx context.Context, enterpriseID string) ([]*models.CBDCWalletRequest, error) {
	var pendingRequests []*models.CBDCWalletRequest

	for _, req := range m.requests {
		if req.EnterpriseID == enterpriseID && req.Status == models.CBDCRequestStatusPending {
			pendingRequests = append(pendingRequests, req)
		}
	}

	return pendingRequests, nil
}

func (m *MockCBDCWalletRequestRepository) ListByStatus(ctx context.Context, status models.CBDCRequestStatus) ([]*models.CBDCWalletRequest, error) {
	var requests []*models.CBDCWalletRequest
	for _, req := range m.requests {
		if req.Status == status {
			requests = append(requests, req)
		}
	}
	return requests, nil
}

// MockTSPConfigRepository implements TSPConfigRepositoryInterface
type MockTSPConfigRepository struct {
	configs map[string]*models.TSPConfig
}

func NewMockTSPConfigRepository() *MockTSPConfigRepository {
	return &MockTSPConfigRepository{
		configs: make(map[string]*models.TSPConfig),
	}
}

func (m *MockTSPConfigRepository) Create(ctx context.Context, config *models.TSPConfig) error {
	config.ID = generateMockID("tsp")
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	m.configs[config.ID] = config
	return nil
}

func (m *MockTSPConfigRepository) GetByID(ctx context.Context, id string) (*models.TSPConfig, error) {
	config, exists := m.configs[id]
	if !exists {
		return nil, fmt.Errorf("TSP config not found: %s", id)
	}
	return config, nil
}

func (m *MockTSPConfigRepository) Update(ctx context.Context, config *models.TSPConfig) error {
	config.UpdatedAt = time.Now()
	m.configs[config.ID] = config
	return nil
}

func (m *MockTSPConfigRepository) Delete(ctx context.Context, id string) error {
	delete(m.configs, id)
	return nil
}

func (m *MockTSPConfigRepository) ListByStatus(ctx context.Context, status models.TSPStatus) ([]*models.TSPConfig, error) {
	var configs []*models.TSPConfig
	for _, config := range m.configs {
		if config.Status == status {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

func (m *MockTSPConfigRepository) ListByEnvironment(ctx context.Context, environment string) ([]*models.TSPConfig, error) {
	var configs []*models.TSPConfig
	for _, config := range m.configs {
		if config.Environment == environment {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

// Helper functions
func generateMockID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
