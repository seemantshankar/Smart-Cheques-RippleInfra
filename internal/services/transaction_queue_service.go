package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// TransactionQueueService manages transaction queuing, batching, and processing
type TransactionQueueService struct {
	transactionRepo       repository.TransactionRepositoryInterface
	xrplService           *XRPLService
	messagingService      *messaging.Service
	fraudDetectionService FraudDetectionServiceInterface
	batchConfig           models.BatchConfig

	// Queue management
	processingQueue chan *models.Transaction
	batchingQueue   chan *models.Transaction

	// Batch management
	activeBatches map[string]*models.TransactionBatch
	batchMutex    sync.RWMutex

	// Processing control
	isRunning   bool
	stopChannel chan struct{}
	stopMutex   sync.Mutex // Mutex to protect stopChannel operations
	wg          sync.WaitGroup

	// Statistics
	stats      *models.TransactionStats
	statsMutex sync.RWMutex
}

// NewTransactionQueueService creates a new transaction queue service
func NewTransactionQueueService(
	transactionRepo repository.TransactionRepositoryInterface,
	xrplService *XRPLService,
	messagingService *messaging.Service,
	fraudDetectionService FraudDetectionServiceInterface,
	config models.BatchConfig,
) *TransactionQueueService {
	return &TransactionQueueService{
		transactionRepo:       transactionRepo,
		xrplService:           xrplService,
		messagingService:      messagingService,
		fraudDetectionService: fraudDetectionService,
		batchConfig:           config,
		processingQueue:       make(chan *models.Transaction, 1000),
		batchingQueue:         make(chan *models.Transaction, 1000),
		activeBatches:         make(map[string]*models.TransactionBatch),
		stopChannel:           make(chan struct{}),
		stats:                 &models.TransactionStats{},
	}
}

// Start begins the transaction queue processing
func (s *TransactionQueueService) Start() error {
	s.stopMutex.Lock()
	defer s.stopMutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("transaction queue service is already running")
	}

	s.isRunning = true
	log.Println("Starting Transaction Queue Service...")

	// Start background workers
	s.wg.Add(4)
	go s.queueProcessor()
	go s.batchProcessor()
	go s.transactionProcessor()
	go s.statusMonitor()

	// Subscribe to transaction events
	if err := s.subscribeToBatchingEvents(); err != nil {
		log.Printf("Warning: Failed to subscribe to batching events: %v", err)
	}

	log.Println("Transaction Queue Service started successfully")
	return nil
}

// Stop gracefully shuts down the transaction queue service
func (s *TransactionQueueService) Stop() {
	s.stopMutex.Lock()
	defer s.stopMutex.Unlock()

	if !s.isRunning {
		return
	}

	log.Println("Stopping Transaction Queue Service...")
	s.isRunning = false

	// Close stop channel only if it hasn't been closed already
	select {
	case <-s.stopChannel:
		// Channel already closed
	default:
		close(s.stopChannel)
	}

	s.wg.Wait()

	// Process any remaining transactions in batches
	s.processPendingBatches()

	log.Println("Transaction Queue Service stopped")
}

// EnqueueTransaction adds a transaction to the processing queue
func (s *TransactionQueueService) EnqueueTransaction(transaction *models.Transaction) error {
	// Validate transaction
	if err := s.validateTransaction(transaction); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Set initial status and timestamps
	transaction.Status = models.TransactionStatusQueued
	transaction.UpdatedAt = time.Now()

	// Set expiration if not set
	if transaction.ExpiresAt == nil {
		expiresAt := time.Now().Add(24 * time.Hour) // Default 24-hour expiration
		transaction.ExpiresAt = &expiresAt
	}

	// Save to database if repository is available
	if s.transactionRepo != nil {
		if err := s.transactionRepo.CreateTransaction(transaction); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}
	}

	// Add to appropriate queue based on batching capability
	if transaction.CanBatch() && s.batchConfig.FeeOptimization {
		select {
		case s.batchingQueue <- transaction:
			log.Printf("Transaction %s queued for batching", transaction.ID)
		default:
			// If batching queue is full, process individually
			select {
			case s.processingQueue <- transaction:
				log.Printf("Transaction %s queued for individual processing (batching queue full)", transaction.ID)
			default:
				return fmt.Errorf("processing queues are full")
			}
		}
	} else {
		select {
		case s.processingQueue <- transaction:
			log.Printf("Transaction %s queued for individual processing", transaction.ID)
		default:
			return fmt.Errorf("processing queue is full")
		}
	}

	// Publish queuing event
	if s.messagingService != nil {
		if err := s.publishTransactionEvent(transaction, "transaction_queued"); err != nil {
			log.Printf("Warning: Failed to publish transaction queued event: %v", err)
		}
	}

	return nil
}

// GetTransactionStatus retrieves the current status of a transaction
func (s *TransactionQueueService) GetTransactionStatus(transactionID string) (*models.Transaction, error) {
	if s.transactionRepo == nil {
		return nil, fmt.Errorf("transaction repository not initialized")
	}
	return s.transactionRepo.GetTransactionByID(transactionID)
}

// GetBatchStatus retrieves the current status of a transaction batch
func (s *TransactionQueueService) GetBatchStatus(batchID string) (*models.TransactionBatch, error) {
	if s.transactionRepo == nil {
		return nil, fmt.Errorf("transaction repository not initialized")
	}
	return s.transactionRepo.GetTransactionBatchByID(batchID)
}

// GetQueueStats returns current queue statistics
func (s *TransactionQueueService) GetQueueStats() (*models.TransactionStats, error) {
	if s.transactionRepo == nil {
		return s.stats, nil // Return cached stats if repository not available
	}

	// Get fresh stats from database
	stats, err := s.transactionRepo.GetTransactionStats()
	if err != nil {
		return s.stats, nil // Return cached stats if database query fails
	}

	s.stats = stats
	return stats, nil
}

// RetryFailedTransactions attempts to retry failed transactions
func (s *TransactionQueueService) RetryFailedTransactions() error {
	if s.transactionRepo == nil {
		return fmt.Errorf("transaction repository not initialized")
	}

	retriableTransactions, err := s.transactionRepo.GetRetriableTransactions()
	if err != nil {
		return fmt.Errorf("failed to get retriable transactions: %w", err)
	}

	retryCount := 0
	for _, tx := range retriableTransactions {
		if tx.CanRetry() {
			tx.Status = models.TransactionStatusQueued
			tx.RetryCount++
			tx.UpdatedAt = time.Now()

			if err := s.transactionRepo.UpdateTransaction(tx); err != nil {
				log.Printf("Failed to update retry transaction %s: %v", tx.ID, err)
				continue
			}

			// Re-enqueue the transaction
			select {
			case s.processingQueue <- tx:
				retryCount++
				log.Printf("Retrying transaction %s", tx.ID)
			default:
				log.Printf("Failed to re-queue transaction %s: queue full", tx.ID)
			}
		}
	}

	log.Printf("Retried %d transactions", retryCount)
	return nil
}

// ExpireOldTransactions marks expired transactions as expired
func (s *TransactionQueueService) ExpireOldTransactions() error {
	if s.transactionRepo == nil {
		return fmt.Errorf("transaction repository not initialized")
	}

	expiredTransactions, err := s.transactionRepo.GetExpiredTransactions()
	if err != nil {
		return fmt.Errorf("failed to get expired transactions: %w", err)
	}

	expiredCount := 0
	for _, tx := range expiredTransactions {
		// Only expire transactions that are still in a queued or processing state
		if tx.Status == models.TransactionStatusQueued ||
			tx.Status == models.TransactionStatusProcessing ||
			tx.Status == models.TransactionStatusBatched {
			tx.Status = models.TransactionStatusExpired
			tx.UpdatedAt = time.Now()

			if err := s.transactionRepo.UpdateTransaction(tx); err != nil {
				log.Printf("Failed to update expired transaction %s: %v", tx.ID, err)
				continue
			}

			expiredCount++
			log.Printf("Expired transaction %s", tx.ID)
		}
	}

	if expiredCount > 0 {
		log.Printf("Expired %d transactions", expiredCount)
	}

	return nil
}

// queueProcessor handles incoming transactions for queuing decisions
func (s *TransactionQueueService) queueProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChannel:
			return
		default:
			// Process any pending transactions from database
			// Skip if transactionRepo is not initialized
			if s.transactionRepo != nil {
				pendingTransactions, err := s.transactionRepo.GetPendingTransactions(100)
				if err != nil {
					log.Printf("Error retrieving pending transactions: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				for _, tx := range pendingTransactions {
					if tx.CanBatch() && s.batchConfig.FeeOptimization {
						select {
						case s.batchingQueue <- tx:
						case <-s.stopChannel:
							return
						}
					} else {
						select {
						case s.processingQueue <- tx:
						case <-s.stopChannel:
							return
						}
					}
				}
			}

			time.Sleep(10 * time.Second) // Check every 10 seconds
		}
	}
}

// batchProcessor handles transaction batching logic
func (s *TransactionQueueService) batchProcessor() {
	defer s.wg.Done()

	batchTimeout := time.NewTicker(time.Duration(s.batchConfig.BatchTimeoutSeconds) * time.Second)
	defer batchTimeout.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case transaction := <-s.batchingQueue:
			s.addTransactionToBatch(transaction)
		case <-batchTimeout.C:
			s.processReadyBatches()
		}
	}
}

// transactionProcessor handles individual transaction processing
func (s *TransactionQueueService) transactionProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopChannel:
			return
		case transaction := <-s.processingQueue:
			s.processTransaction(transaction)
		}
	}
}

// statusMonitor periodically updates transaction and batch statuses
func (s *TransactionQueueService) statusMonitor() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case <-ticker.C:
			// Expire old transactions
			if err := s.ExpireOldTransactions(); err != nil {
				log.Printf("Error expiring transactions: %v", err)
			}

			// Update statistics
			if stats, err := s.transactionRepo.GetTransactionStats(); err == nil {
				s.statsMutex.Lock()
				s.stats = stats
				s.statsMutex.Unlock()
			}
		}
	}
}

// addTransactionToBatch adds a transaction to an appropriate batch
func (s *TransactionQueueService) addTransactionToBatch(transaction *models.Transaction) {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	// Find suitable batch or create new one
	var targetBatch *models.TransactionBatch

	for _, batch := range s.activeBatches {
		if s.canAddToBatch(batch, transaction) {
			targetBatch = batch
			break
		}
	}

	// Create new batch if none suitable
	if targetBatch == nil {
		targetBatch = models.NewTransactionBatch(transaction.Priority, s.batchConfig.MaxBatchSize)
		targetBatch.Status = models.TransactionStatusBatching

		if err := s.transactionRepo.CreateTransactionBatch(targetBatch); err != nil {
			log.Printf("Failed to create transaction batch: %v", err)
			// Fall back to individual processing
			s.processingQueue <- transaction
			return
		}

		s.activeBatches[targetBatch.ID] = targetBatch
		log.Printf("Created new transaction batch: %s", targetBatch.ID)
	}

	// Add transaction to batch
	transaction.BatchID = &targetBatch.ID
	transaction.Status = models.TransactionStatusBatched
	transaction.UpdatedAt = time.Now()

	if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
		log.Printf("Failed to update transaction for batching: %v", err)
		return
	}

	// Update batch
	targetBatch.TransactionCount++
	targetBatch.UpdatedAt = time.Now()

	if err := s.transactionRepo.UpdateTransactionBatch(targetBatch); err != nil {
		log.Printf("Failed to update transaction batch: %v", err)
	}

	log.Printf("Added transaction %s to batch %s (%d/%d)",
		transaction.ID, targetBatch.ID, targetBatch.TransactionCount, targetBatch.MaxTransactions)

	// Check if batch is ready for processing
	if targetBatch.TransactionCount >= s.batchConfig.MinBatchSize &&
		(targetBatch.TransactionCount >= targetBatch.MaxTransactions ||
			time.Since(targetBatch.CreatedAt) >= s.batchConfig.MaxWaitTime) {
		s.processBatch(targetBatch)
	}
}

// canAddToBatch checks if a transaction can be added to a specific batch
func (s *TransactionQueueService) canAddToBatch(batch *models.TransactionBatch, transaction *models.Transaction) bool {
	return batch.Status == models.TransactionStatusBatching &&
		batch.TransactionCount < batch.MaxTransactions &&
		batch.Priority == transaction.Priority &&
		time.Since(batch.CreatedAt) < s.batchConfig.MaxWaitTime
}

// processReadyBatches processes batches that are ready for submission
func (s *TransactionQueueService) processReadyBatches() {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	for batchID, batch := range s.activeBatches {
		if batch.TransactionCount >= s.batchConfig.MinBatchSize &&
			time.Since(batch.CreatedAt) >= s.batchConfig.MaxWaitTime {
			s.processBatch(batch)
			delete(s.activeBatches, batchID)
		}
	}
}

// processBatch processes a complete transaction batch
func (s *TransactionQueueService) processBatch(batch *models.TransactionBatch) {
	log.Printf("Processing transaction batch: %s with %d transactions", batch.ID, batch.TransactionCount)

	// Get all transactions in the batch
	transactions, err := s.transactionRepo.GetTransactionsByBatchID(batch.ID)
	if err != nil {
		log.Printf("Failed to get transactions for batch %s: %v", batch.ID, err)
		return
	}

	// Update batch status
	batch.Status = models.TransactionStatusProcessing
	batch.UpdatedAt = time.Now()
	now := time.Now()
	batch.ProcessedAt = &now

	// Calculate optimized fees
	if s.batchConfig.FeeOptimization {
		s.optimizeBatchFees(batch, transactions)
	}

	if err := s.transactionRepo.UpdateTransactionBatch(batch); err != nil {
		log.Printf("Failed to update batch status: %v", err)
	}

	// Process each transaction in the batch
	successCount := 0
	failureCount := 0

	for _, tx := range transactions {
		if s.processTransaction(tx) {
			successCount++
		} else {
			failureCount++
		}
	}

	// Update batch completion status
	batch.SuccessCount = successCount
	batch.FailureCount = failureCount
	batch.UpdatedAt = time.Now()
	now = time.Now()
	batch.CompletedAt = &now

	if failureCount == 0 {
		batch.Status = models.TransactionStatusConfirmed
	} else if successCount == 0 {
		batch.Status = models.TransactionStatusFailed
	} else {
		batch.Status = models.TransactionStatusConfirmed // Partial success still considered confirmed
	}

	if err := s.transactionRepo.UpdateTransactionBatch(batch); err != nil {
		log.Printf("Failed to update final batch status: %v", err)
	}

	log.Printf("Completed batch %s: %d success, %d failures", batch.ID, successCount, failureCount)

	// Publish batch completion event
	if err := s.publishBatchEvent(batch, "batch_completed"); err != nil {
		log.Printf("Warning: Failed to publish batch completed event: %v", err)
	}
}

// processTransaction processes a single transaction
func (s *TransactionQueueService) processTransaction(transaction *models.Transaction) bool {
	transaction.Status = models.TransactionStatusProcessing
	transaction.UpdatedAt = time.Now()
	now := time.Now()
	transaction.ProcessedAt = &now

	if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
		log.Printf("Failed to update transaction status to processing: %v", err)
	}

	// Perform fraud detection analysis
	if s.fraudDetectionService != nil {
		// Convert transaction to fraud analysis request
		enterpriseID, err := uuid.Parse(transaction.EnterpriseID)
		if err != nil {
			log.Printf("Invalid enterprise ID for transaction %s: %v", transaction.ID, err)
			transaction.SetError(fmt.Errorf("invalid enterprise ID: %w", err))
			if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
				log.Printf("Failed to update transaction with enterprise ID error: %v", err)
			}
			return false
		}

		fraudRequest := &FraudAnalysisRequest{
			TransactionID:   transaction.ID,
			EnterpriseID:    enterpriseID,
			Amount:          transaction.Amount,
			CurrencyCode:    transaction.Currency,
			TransactionType: string(transaction.Type),
			Destination:     transaction.ToAddress,
			Timestamp:       time.Now(),
			Metadata:        transaction.Metadata,
		}

		fraudResult, err := s.fraudDetectionService.AnalyzeTransaction(context.Background(), fraudRequest)
		if err != nil {
			log.Printf("Fraud detection failed for transaction %s: %v", transaction.ID, err)
			transaction.SetError(fmt.Errorf("fraud detection failed: %w", err))
			if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
				log.Printf("Failed to update transaction with fraud detection error: %v", err)
			}
			return false
		}

		// Check if fraud is detected
		if fraudResult.FraudDetected {
			log.Printf("Transaction %s flagged as fraud with risk score %.2f: %v",
				transaction.ID, fraudResult.RiskScore, fraudResult.RiskFactors)
			transaction.Status = models.TransactionStatusFraud
			transaction.UpdatedAt = time.Now()
			if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
				log.Printf("Failed to update transaction to fraud status: %v", err)
			}
			return false
		}

		// Log high-risk transactions for monitoring
		if fraudResult.RiskScore >= 0.6 {
			log.Printf("High-risk transaction %s detected with risk score %.2f",
				transaction.ID, fraudResult.RiskScore)
		}
	}

	// Calculate fee if not already set
	if transaction.Fee == "" {
		feeCalculator := NewFeeCalculator()
		networkLoad := feeCalculator.EstimateNetworkLoad()
		fee, err := feeCalculator.CalculateTransactionFee(transaction, networkLoad)
		if err != nil {
			log.Printf("Failed to calculate fee for transaction %s: %v", transaction.ID, err)
			transaction.SetError(fmt.Errorf("fee calculation failed: %w", err))
			if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
				log.Printf("Failed to update transaction with fee calculation error: %v", err)
			}
			return false
		}
		transaction.Fee = fee
	}

	// Process based on transaction type
	var err error
	switch transaction.Type {
	case models.TransactionTypeEscrowCreate:
		err = s.processEscrowCreate(transaction)
	case models.TransactionTypeEscrowFinish:
		err = s.processEscrowFinish(transaction)
	case models.TransactionTypeEscrowCancel:
		err = s.processEscrowCancel(transaction)
	case models.TransactionTypePayment:
		err = s.processPayment(transaction)
	case models.TransactionTypeWalletSetup:
		err = s.processWalletSetup(transaction)
	default:
		err = fmt.Errorf("unsupported transaction type: %s", transaction.Type)
	}

	if err != nil {
		log.Printf("Transaction %s failed: %v", transaction.ID, err)
		transaction.SetError(err)
		if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
			log.Printf("Failed to update transaction with error: %v", err)
		}

		// Publish failure event
		if err := s.publishTransactionEvent(transaction, "transaction_failed"); err != nil {
			log.Printf("Failed to publish transaction failed event: %v", err)
		}
		return false
	}

	// Mark as confirmed
	transaction.Status = models.TransactionStatusConfirmed
	transaction.UpdatedAt = time.Now()
	now = time.Now()
	transaction.ConfirmedAt = &now

	if err := s.transactionRepo.UpdateTransaction(transaction); err != nil {
		log.Printf("Failed to update transaction to confirmed: %v", err)
	}

	log.Printf("Transaction %s processed successfully", transaction.ID)

	// Publish success event
	if err := s.publishTransactionEvent(transaction, "transaction_confirmed"); err != nil {
		log.Printf("Failed to publish transaction confirmed event: %v", err)
	}
	return true
}

// processEscrowCreate handles escrow creation transactions
func (s *TransactionQueueService) processEscrowCreate(transaction *models.Transaction) error {
	if !s.xrplService.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}

	// Extract milestone secret from metadata
	milestoneSecret, ok := transaction.Metadata["milestone_secret"].(string)
	if !ok {
		milestoneSecret = "default_milestone_secret"
	}

	// Parse amount
	amount, err := strconv.ParseFloat(transaction.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	// Create escrow
	result, fulfillment, err := s.xrplService.CreateSmartChequeEscrow(
		transaction.FromAddress,
		transaction.ToAddress,
		amount,
		transaction.Currency,
		milestoneSecret,
	)
	if err != nil {
		return fmt.Errorf("failed to create escrow: %w", err)
	}

	// Update transaction with XRPL details
	transaction.TransactionHash = result.TransactionID
	transaction.LedgerIndex = &result.LedgerIndex
	transaction.Fulfillment = fulfillment

	return nil
}

// processEscrowFinish handles escrow finish transactions
func (s *TransactionQueueService) processEscrowFinish(transaction *models.Transaction) error {
	if !s.xrplService.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}

	if transaction.OfferSequence == nil {
		return fmt.Errorf("offer sequence is required for escrow finish")
	}

	result, err := s.xrplService.CompleteSmartChequeMilestone(
		transaction.ToAddress,
		transaction.FromAddress,
		*transaction.OfferSequence,
		transaction.Condition,
		transaction.Fulfillment,
	)
	if err != nil {
		return fmt.Errorf("failed to finish escrow: %w", err)
	}

	transaction.TransactionHash = result.TransactionID
	transaction.LedgerIndex = &result.LedgerIndex

	return nil
}

// processEscrowCancel handles escrow cancellation transactions
func (s *TransactionQueueService) processEscrowCancel(transaction *models.Transaction) error {
	if !s.xrplService.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}

	if transaction.OfferSequence == nil {
		return fmt.Errorf("offer sequence is required for escrow cancel")
	}

	result, err := s.xrplService.CancelSmartCheque(
		transaction.FromAddress,
		transaction.FromAddress,
		*transaction.OfferSequence,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel escrow: %w", err)
	}

	transaction.TransactionHash = result.TransactionID
	transaction.LedgerIndex = &result.LedgerIndex

	return nil
}

// processPayment handles regular payment transactions
// Returns error which is always nil in this implementation
func (s *TransactionQueueService) processPayment(transaction *models.Transaction) error {
	// For now, simulate payment processing
	// In production, this would use XRPL payment transactions
	log.Printf("Processing payment: %s %s from %s to %s",
		transaction.Amount, transaction.Currency,
		transaction.FromAddress, transaction.ToAddress)

	// Simulate transaction hash
	transaction.TransactionHash = fmt.Sprintf("payment_%s_%d", transaction.ID, time.Now().Unix())

	return nil
}

// processWalletSetup handles wallet setup transactions
func (s *TransactionQueueService) processWalletSetup(transaction *models.Transaction) error {
	if !s.xrplService.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}

	// Create new wallet
	walletInfo, err := s.xrplService.CreateWallet()
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Store wallet info in transaction metadata
	transaction.Metadata["wallet_address"] = walletInfo.Address
	transaction.Metadata["wallet_public_key"] = walletInfo.PublicKey

	// Update transaction addresses
	transaction.ToAddress = walletInfo.Address
	transaction.TransactionHash = fmt.Sprintf("wallet_setup_%s", walletInfo.Address)

	return nil
}

// optimizeBatchFees optimizes fees for a batch of transactions
func (s *TransactionQueueService) optimizeBatchFees(batch *models.TransactionBatch, transactions []*models.Transaction) {
	feeCalculator := NewFeeCalculator()
	networkLoad := feeCalculator.EstimateNetworkLoad()

	if err := feeCalculator.OptimizeBatchFees(batch, transactions, networkLoad); err != nil {
		log.Printf("Failed to optimize batch fees: %v", err)
	}
}

// validateTransaction validates a transaction before processing
func (s *TransactionQueueService) validateTransaction(transaction *models.Transaction) error {
	if transaction.FromAddress == "" {
		return fmt.Errorf("from address is required")
	}

	if transaction.ToAddress == "" {
		return fmt.Errorf("to address is required")
	}

	if transaction.Amount == "" {
		return fmt.Errorf("amount is required")
	}

	if transaction.EnterpriseID == "" {
		return fmt.Errorf("enterprise ID is required")
	}

	if transaction.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate addresses using XRPL service
	if s.xrplService != nil {
		if !s.xrplService.ValidateAddress(transaction.FromAddress) {
			return fmt.Errorf("invalid from address: %s", transaction.FromAddress)
		}

		if !s.xrplService.ValidateAddress(transaction.ToAddress) {
			return fmt.Errorf("invalid to address: %s", transaction.ToAddress)
		}
	}

	return nil
}

// processPendingBatches processes any remaining batches during shutdown
func (s *TransactionQueueService) processPendingBatches() {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	for batchID, batch := range s.activeBatches {
		if batch.TransactionCount > 0 {
			log.Printf("Processing pending batch %s with %d transactions during shutdown",
				batchID, batch.TransactionCount)
			s.processBatch(batch)
		}
		delete(s.activeBatches, batchID)
	}
}

// subscribeToBatchingEvents sets up event subscriptions for batch processing
func (s *TransactionQueueService) subscribeToBatchingEvents() error {
	// Subscribe to transaction events that might affect batching
	if err := s.messagingService.SubscribeToEvent("transaction_priority_changed", s.handlePriorityChange); err != nil {
		return fmt.Errorf("failed to subscribe to priority change events: %w", err)
	}

	if err := s.messagingService.SubscribeToEvent("transaction_canceled", s.handleTransactionCancellation); err != nil {
		return fmt.Errorf("failed to subscribe to cancellation events: %w", err)
	}

	return nil
}

// handlePriorityChange handles transaction priority change events
func (s *TransactionQueueService) handlePriorityChange(event *messaging.Event) error {
	transactionID, ok := event.Data["transaction_id"].(string)
	if !ok {
		return fmt.Errorf("invalid transaction ID in priority change event")
	}

	newPriority, ok := event.Data["new_priority"].(float64)
	if !ok {
		return fmt.Errorf("invalid priority in priority change event")
	}

	log.Printf("Handling priority change for transaction %s to %d", transactionID, int(newPriority))

	// Update transaction priority in database
	transaction, err := s.transactionRepo.GetTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction.Priority = models.TransactionPriority(int(newPriority))
	transaction.UpdatedAt = time.Now()

	return s.transactionRepo.UpdateTransaction(transaction)
}

// handleTransactionCancellation handles transaction cancellation events
func (s *TransactionQueueService) handleTransactionCancellation(event *messaging.Event) error {
	transactionID, ok := event.Data["transaction_id"].(string)
	if !ok {
		return fmt.Errorf("invalid transaction ID in cancellation event")
	}

	log.Printf("Handling cancellation for transaction %s", transactionID)

	// Update transaction status
	transaction, err := s.transactionRepo.GetTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction.Status = models.TransactionStatusCancelled
	transaction.UpdatedAt = time.Now()

	return s.transactionRepo.UpdateTransaction(transaction)
}

// publishTransactionEvent publishes a transaction-related event
func (s *TransactionQueueService) publishTransactionEvent(transaction *models.Transaction, eventType string) error {
	event := &messaging.Event{
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"transaction_id":   transaction.ID,
			"transaction_type": transaction.Type,
			"status":           transaction.Status,
			"enterprise_id":    transaction.EnterpriseID,
			"user_id":          transaction.UserID,
			"amount":           transaction.Amount,
			"currency":         transaction.Currency,
			"batch_id":         transaction.BatchID,
		},
	}

	return s.messagingService.PublishEvent(event)
}

// publishBatchEvent publishes a batch-related event
func (s *TransactionQueueService) publishBatchEvent(batch *models.TransactionBatch, eventType string) error {
	event := &messaging.Event{
		Type:      eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"batch_id":          batch.ID,
			"status":            batch.Status,
			"transaction_count": batch.TransactionCount,
			"success_count":     batch.SuccessCount,
			"failure_count":     batch.FailureCount,
			"total_fee":         batch.TotalFee,
			"optimized_fee":     batch.OptimizedFee,
			"fee_savings":       batch.FeeSavings,
		},
	}

	return s.messagingService.PublishEvent(event)
}
