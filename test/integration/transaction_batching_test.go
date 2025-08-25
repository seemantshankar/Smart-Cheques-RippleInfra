package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/database"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// TransactionBatchingIntegrationTestSuite tests transaction batching functionality
type TransactionBatchingIntegrationTestSuite struct {
	suite.Suite

	// Services
	queueService      *services.TransactionQueueService
	monitoringService *services.TransactionMonitoringService
	xrplService       *services.XRPLService
	messagingService  *messaging.MessagingService

	// Repositories
	transactionRepo repository.TransactionRepositoryInterface

	// Database
	db *database.PostgresDB

	// Test data
	testEnterpriseID string
	testUserID       string
}

func TestTransactionBatchingIntegration(t *testing.T) {
	suite.Run(t, new(TransactionBatchingIntegrationTestSuite))
}

func (suite *TransactionBatchingIntegrationTestSuite) SetupSuite() {
	// Initialize test database (in-memory or test instance)
	suite.db = setupTestDatabase()

	// Initialize repositories (using mock for now due to type mismatch)
	// suite.transactionRepo = repository.NewTransactionRepository(suite.db.DB)
	// TODO: Fix database type mismatch - PostgresDB.DB is *sql.DB but NewTransactionRepository expects *gorm.DB
	suite.transactionRepo = nil // Will be mocked in actual tests

	// Initialize messaging service
	messagingService, err := messaging.NewMessagingService("localhost:6379", "", 1)
	require.NoError(suite.T(), err)
	suite.messagingService = messagingService

	// Initialize XRPL service
	suite.xrplService = services.NewXRPLService(services.XRPLConfig{
		NetworkURL: "http://localhost:5005", // Test XRPL server
		TestNet:    true,
	})
	err = suite.xrplService.Initialize()
	require.NoError(suite.T(), err)

	// Initialize queue service
	batchConfig := models.DefaultBatchConfig()
	batchConfig.MaxBatchSize = 5
	batchConfig.MinBatchSize = 2
	batchConfig.MaxWaitTime = 5 * time.Second

	suite.queueService = services.NewTransactionQueueService(
		suite.transactionRepo,
		suite.xrplService,
		suite.messagingService,
		batchConfig,
	)

	// Initialize monitoring service
	suite.monitoringService = services.NewTransactionMonitoringService(
		suite.transactionRepo,
		suite.messagingService,
	)

	// Setup test data
	suite.testEnterpriseID = uuid.New().String()
	suite.testUserID = uuid.New().String()
}

func (suite *TransactionBatchingIntegrationTestSuite) TearDownSuite() {
	if suite.queueService != nil {
		suite.queueService.Stop()
	}
	if suite.monitoringService != nil {
		suite.monitoringService.Stop()
	}
	if suite.messagingService != nil {
		suite.messagingService.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *TransactionBatchingIntegrationTestSuite) SetupTest() {
	// Start services for each test
	err := suite.queueService.Start()
	require.NoError(suite.T(), err)

	err = suite.monitoringService.Start()
	require.NoError(suite.T(), err)
}

func (suite *TransactionBatchingIntegrationTestSuite) TearDownTest() {
	// Stop services after each test
	suite.queueService.Stop()
	suite.monitoringService.Stop()
}

func (suite *TransactionBatchingIntegrationTestSuite) TestSingleTransactionProcessing() {
	t := suite.T()

	// Create wallet addresses for testing
	senderWallet, err := suite.xrplService.CreateWallet()
	require.NoError(t, err)

	receiverWallet, err := suite.xrplService.CreateWallet()
	require.NoError(t, err)

	// Create a single transaction
	transaction := models.NewTransaction(
		models.TransactionTypePayment,
		senderWallet.Address,
		receiverWallet.Address,
		"1000000", // 1 XRP
		"XRP",
		suite.testEnterpriseID,
		suite.testUserID,
	)
	transaction.Priority = models.PriorityHigh

	// Enqueue the transaction
	err = suite.queueService.EnqueueTransaction(transaction)
	require.NoError(t, err)

	// Wait for processing
	suite.waitForTransactionStatus(transaction.ID, models.TransactionStatusConfirmed, 30*time.Second)

	// Verify transaction was processed
	processedTx, err := suite.transactionRepo.GetTransactionByID(transaction.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TransactionStatusConfirmed, processedTx.Status)
	assert.NotEmpty(t, processedTx.Fee)
	assert.NotNil(t, processedTx.ProcessedAt)
	assert.NotNil(t, processedTx.ConfirmedAt)
}

func (suite *TransactionBatchingIntegrationTestSuite) TestBatchTransactionProcessing() {
	t := suite.T()

	// Create multiple transactions for batching
	transactions := make([]*models.Transaction, 4)

	for i := 0; i < 4; i++ {
		senderWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		receiverWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		tx := models.NewTransaction(
			models.TransactionTypeEscrowCreate,
			senderWallet.Address,
			receiverWallet.Address,
			fmt.Sprintf("%d000000", i+1), // Different amounts
			"XRP",
			suite.testEnterpriseID,
			suite.testUserID,
		)
		tx.Priority = models.PriorityNormal
		transactions[i] = tx

		// Enqueue transaction
		err = suite.queueService.EnqueueTransaction(tx)
		require.NoError(t, err)
	}

	// Wait for all transactions to be processed
	for _, tx := range transactions {
		suite.waitForTransactionStatus(tx.ID, models.TransactionStatusConfirmed, 60*time.Second)
	}

	// Verify transactions were batched (check if they share a batch ID)
	batchIDs := make(map[string]int)
	for _, tx := range transactions {
		processedTx, err := suite.transactionRepo.GetTransactionByID(tx.ID)
		require.NoError(t, err)

		assert.Equal(t, models.TransactionStatusConfirmed, processedTx.Status)

		if processedTx.BatchID != nil {
			batchIDs[*processedTx.BatchID]++
		}
	}

	// Verify batching occurred (at least one batch with multiple transactions)
	foundBatch := false
	for _, count := range batchIDs {
		if count >= 2 {
			foundBatch = true
			break
		}
	}
	assert.True(t, foundBatch, "Expected transactions to be batched together")
}

func (suite *TransactionBatchingIntegrationTestSuite) TestTransactionRetryMechanism() {
	t := suite.T()

	// Create a transaction that will initially fail
	transaction := models.NewTransaction(
		models.TransactionTypeEscrowCreate,
		"rInvalidSender", // Invalid address to cause failure
		"rInvalidReceiver",
		"1000000",
		"XRP",
		suite.testEnterpriseID,
		suite.testUserID,
	)
	transaction.MaxRetries = 2

	// Enqueue the transaction
	err := suite.queueService.EnqueueTransaction(transaction)
	require.NoError(t, err)

	// Wait for failure and retry attempts
	time.Sleep(10 * time.Second)

	// Verify transaction eventually failed after retries
	processedTx, err := suite.transactionRepo.GetTransactionByID(transaction.ID)
	require.NoError(t, err)

	assert.Equal(t, models.TransactionStatusFailed, processedTx.Status)
	assert.Equal(t, 2, processedTx.RetryCount)
	assert.NotEmpty(t, processedTx.LastError)
}

func (suite *TransactionBatchingIntegrationTestSuite) TestFeeOptimization() {
	t := suite.T()

	// Create multiple transactions of the same type for batch optimization
	transactions := make([]*models.Transaction, 3)

	for i := 0; i < 3; i++ {
		senderWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		receiverWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		tx := models.NewTransaction(
			models.TransactionTypePayment,
			senderWallet.Address,
			receiverWallet.Address,
			"500000", // 0.5 XRP
			"XRP",
			suite.testEnterpriseID,
			suite.testUserID,
		)
		tx.Priority = models.PriorityLow // Low priority for better optimization
		transactions[i] = tx

		err = suite.queueService.EnqueueTransaction(tx)
		require.NoError(t, err)
	}

	// Wait for processing
	for _, tx := range transactions {
		suite.waitForTransactionStatus(tx.ID, models.TransactionStatusConfirmed, 30*time.Second)
	}

	// Check if transactions were batched and fees were optimized
	var batchID *string
	for _, tx := range transactions {
		processedTx, err := suite.transactionRepo.GetTransactionByID(tx.ID)
		require.NoError(t, err)

		if processedTx.BatchID != nil {
			batchID = processedTx.BatchID
			break
		}
	}

	if batchID != nil {
		// Verify batch optimization
		batch, err := suite.transactionRepo.GetTransactionBatchByID(*batchID)
		require.NoError(t, err)

		assert.NotEmpty(t, batch.TotalFee)
		assert.NotEmpty(t, batch.OptimizedFee)
		assert.NotEmpty(t, batch.FeeSavings)

		// Parse fees to verify optimization
		// In a real implementation, you'd parse the string values
		// For now, just verify they're set
	}
}

func (suite *TransactionBatchingIntegrationTestSuite) TestTransactionExpiration() {
	t := suite.T()

	// Create a transaction with short expiration
	senderWallet, err := suite.xrplService.CreateWallet()
	require.NoError(t, err)

	receiverWallet, err := suite.xrplService.CreateWallet()
	require.NoError(t, err)

	transaction := models.NewTransaction(
		models.TransactionTypePayment,
		senderWallet.Address,
		receiverWallet.Address,
		"1000000",
		"XRP",
		suite.testEnterpriseID,
		suite.testUserID,
	)

	// Set short expiration
	expiresAt := time.Now().Add(2 * time.Second)
	transaction.ExpiresAt = &expiresAt

	err = suite.queueService.EnqueueTransaction(transaction)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(5 * time.Second)

	// Trigger expiration check
	err = suite.queueService.ExpireOldTransactions()
	require.NoError(t, err)

	// Verify transaction was expired
	expiredTx, err := suite.transactionRepo.GetTransactionByID(transaction.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TransactionStatusExpired, expiredTx.Status)
}

func (suite *TransactionBatchingIntegrationTestSuite) TestMonitoringDashboard() {
	t := suite.T()

	// Create some transactions for monitoring
	for i := 0; i < 3; i++ {
		senderWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		receiverWallet, err := suite.xrplService.CreateWallet()
		require.NoError(t, err)

		tx := models.NewTransaction(
			models.TransactionTypePayment,
			senderWallet.Address,
			receiverWallet.Address,
			"1000000",
			"XRP",
			suite.testEnterpriseID,
			suite.testUserID,
		)

		err = suite.queueService.EnqueueTransaction(tx)
		require.NoError(t, err)
	}

	// Wait a bit for metrics to update
	time.Sleep(5 * time.Second)

	// Get dashboard data
	dashboardData, err := suite.monitoringService.GetDashboardData()
	require.NoError(t, err)

	// Verify dashboard data structure
	assert.NotNil(t, dashboardData.Metrics)
	assert.NotNil(t, dashboardData.StatusDistribution)
	assert.NotNil(t, dashboardData.SystemHealth)

	// Verify some metrics are populated
	assert.False(t, dashboardData.Metrics.LastUpdated.IsZero())
	assert.Equal(t, "healthy", dashboardData.SystemHealth.OverallStatus)
}

func (suite *TransactionBatchingIntegrationTestSuite) TestPriorityBasedProcessing() {
	t := suite.T()

	// Create transactions with different priorities
	lowPriorityTx := suite.createTestTransaction(models.PriorityLow)
	normalPriorityTx := suite.createTestTransaction(models.PriorityNormal)
	highPriorityTx := suite.createTestTransaction(models.PriorityHigh)
	criticalPriorityTx := suite.createTestTransaction(models.PriorityCritical)

	// Enqueue in reverse priority order
	err := suite.queueService.EnqueueTransaction(lowPriorityTx)
	require.NoError(t, err)

	err = suite.queueService.EnqueueTransaction(normalPriorityTx)
	require.NoError(t, err)

	err = suite.queueService.EnqueueTransaction(highPriorityTx)
	require.NoError(t, err)

	err = suite.queueService.EnqueueTransaction(criticalPriorityTx)
	require.NoError(t, err)

	// Wait for processing
	transactions := []*models.Transaction{lowPriorityTx, normalPriorityTx, highPriorityTx, criticalPriorityTx}
	for _, tx := range transactions {
		suite.waitForTransactionStatus(tx.ID, models.TransactionStatusConfirmed, 30*time.Second)
	}

	// Verify all transactions were processed
	for _, tx := range transactions {
		processedTx, err := suite.transactionRepo.GetTransactionByID(tx.ID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusConfirmed, processedTx.Status)
	}
}

// Helper methods

func (suite *TransactionBatchingIntegrationTestSuite) createTestTransaction(priority models.TransactionPriority) *models.Transaction {
	senderWallet, _ := suite.xrplService.CreateWallet()
	receiverWallet, _ := suite.xrplService.CreateWallet()

	tx := models.NewTransaction(
		models.TransactionTypePayment,
		senderWallet.Address,
		receiverWallet.Address,
		"1000000",
		"XRP",
		suite.testEnterpriseID,
		suite.testUserID,
	)
	tx.Priority = priority
	return tx
}

func (suite *TransactionBatchingIntegrationTestSuite) waitForTransactionStatus(
	transactionID string,
	expectedStatus models.TransactionStatus,
	timeout time.Duration,
) {
	t := suite.T()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for transaction %s to reach status %s", transactionID, expectedStatus)
		case <-ticker.C:
			tx, err := suite.transactionRepo.GetTransactionByID(transactionID)
			if err != nil {
				continue
			}

			if tx.Status == expectedStatus {
				return
			}

			// If transaction failed, don't wait further
			if tx.Status == models.TransactionStatusFailed ||
				tx.Status == models.TransactionStatusCancelled ||
				tx.Status == models.TransactionStatusExpired {
				if expectedStatus != tx.Status {
					t.Logf("Transaction %s reached final status %s instead of expected %s",
						transactionID, tx.Status, expectedStatus)
				}
				return
			}
		}
	}
}

// setupTestDatabase creates a test database instance
func setupTestDatabase() *database.PostgresDB {
	// This would setup a test database
	// For now, return a mock or use existing setup
	return nil // Would be properly implemented in real scenario
}
