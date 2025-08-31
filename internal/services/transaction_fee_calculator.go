package services

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/smart-payment-infrastructure/internal/models"
)

// FeeCalculator handles transaction fee calculation and optimization
type FeeCalculator struct {
	baseFeeDrops     int64   // Base fee in drops
	reserveIncrement int64   // Reserve increment for account objects
	maxFeeDrops      int64   // Maximum fee allowed
	feeMultiplier    float64 // Multiplier for fee escalation
}

// NewFeeCalculator creates a new fee calculator
func NewFeeCalculator() *FeeCalculator {
	return &FeeCalculator{
		baseFeeDrops:     10,      // 10 drops base fee
		reserveIncrement: 5000000, // 5 XRP in drops
		maxFeeDrops:      10000,   // 0.01 XRP maximum fee
		feeMultiplier:    1.2,     // 20% fee escalation for retries
	}
}

// CalculateTransactionFee calculates the optimal fee for a single transaction
func (f *FeeCalculator) CalculateTransactionFee(transaction *models.Transaction, networkLoad float64) (string, error) {
	baseFee := f.baseFeeDrops

	// Adjust fee based on transaction type
	switch transaction.Type {
	case models.TransactionTypeEscrowCreate:
		baseFee = 12 // Escrow creation requires slightly higher fee
	case models.TransactionTypeEscrowFinish:
		baseFee = 12 // Escrow finish requires slightly higher fee
	case models.TransactionTypeEscrowCancel:
		baseFee = 12 // Escrow cancel requires slightly higher fee
	case models.TransactionTypePayment:
		baseFee = 10 // Standard payment fee
	case models.TransactionTypeWalletSetup:
		baseFee = 15 // Wallet setup may require account creation
	}

	// Adjust fee based on network load (0.1 = low, 1.0 = high)
	networkMultiplier := 1.0 + networkLoad
	adjustedFee := float64(baseFee) * networkMultiplier

	// Adjust fee based on priority
	priorityMultiplier := f.getPriorityMultiplier(transaction.Priority)
	adjustedFee *= priorityMultiplier

	// Apply retry multiplier if this is a retry
	if transaction.RetryCount > 0 {
		retryMultiplier := math.Pow(f.feeMultiplier, float64(transaction.RetryCount))
		adjustedFee *= retryMultiplier
	}

	// Ensure fee doesn't exceed maximum
	finalFee := int64(math.Ceil(adjustedFee))
	if finalFee > f.maxFeeDrops {
		finalFee = f.maxFeeDrops
	}

	log.Printf("Calculated fee for transaction %s: %d drops (base: %d, network: %.2f, priority: %.2f, retries: %d)",
		transaction.ID, finalFee, baseFee, networkLoad, priorityMultiplier, transaction.RetryCount)

	return strconv.FormatInt(finalFee, 10), nil
}

// OptimizeBatchFees calculates optimized fees for a batch of transactions
func (f *FeeCalculator) OptimizeBatchFees(batch *models.TransactionBatch, transactions []*models.Transaction, networkLoad float64) error {
	if len(transactions) == 0 {
		return fmt.Errorf("no transactions in batch")
	}

	// Calculate individual fees
	totalIndividualFee := int64(0)
	totalOptimizedFee := int64(0)

	for _, tx := range transactions {
		// Calculate what the individual fee would be
		individualFeeStr, err := f.CalculateTransactionFee(tx, networkLoad)
		if err != nil {
			return fmt.Errorf("failed to calculate fee for transaction %s: %w", tx.ID, err)
		}

		individualFee, err := strconv.ParseInt(individualFeeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse individual fee: %w", err)
		}

		totalIndividualFee += individualFee

		// Apply batch optimization (reduced fee for batched transactions)
		optimizedFee := f.applyBatchOptimization(individualFee, len(transactions), tx.Priority)
		totalOptimizedFee += optimizedFee

		// Update transaction fee
		tx.Fee = strconv.FormatInt(optimizedFee, 10)
		log.Printf("Transaction %s fee optimized: %d -> %d drops", tx.ID, individualFee, optimizedFee)
	}

	// Calculate savings
	savings := totalIndividualFee - totalOptimizedFee
	if savings < 0 {
		savings = 0
	}

	// Update batch fee information
	batch.TotalFee = strconv.FormatInt(totalIndividualFee, 10)
	batch.OptimizedFee = strconv.FormatInt(totalOptimizedFee, 10)
	batch.FeeSavings = strconv.FormatInt(savings, 10)

	log.Printf("Batch %s fee optimization: %d -> %d drops (saved %d drops, %.2f%%)",
		batch.ID, totalIndividualFee, totalOptimizedFee, savings,
		float64(savings)/float64(totalIndividualFee)*100)

	return nil
}

// EstimateNetworkLoad estimates current network load based on recent transaction data
func (f *FeeCalculator) EstimateNetworkLoad() float64 {
	// This is a simplified network load estimation
	// In production, this would analyze:
	// - Recent transaction confirmation times
	// - Current fee levels being paid
	// - Network ledger close times
	// - Queue depth

	// For now, return a random value between 0.1 and 0.8
	// representing low to moderate network load
	return 0.3 // Assuming moderate network load
}

// CalculateFeeForRetry calculates an escalated fee for a retry attempt
func (f *FeeCalculator) CalculateFeeForRetry(originalTransaction *models.Transaction, retryCount int) (string, error) {
	// Parse original fee
	originalFee, err := strconv.ParseInt(originalTransaction.Fee, 10, 64)
	if err != nil {
		// If no original fee, calculate base fee
		networkLoad := f.EstimateNetworkLoad()
		return f.CalculateTransactionFee(originalTransaction, networkLoad)
	}

	// Escalate fee for retry
	escalationMultiplier := math.Pow(f.feeMultiplier, float64(retryCount))
	escalatedFee := float64(originalFee) * escalationMultiplier

	// Ensure fee doesn't exceed maximum
	finalFee := int64(math.Ceil(escalatedFee))
	if finalFee > f.maxFeeDrops {
		finalFee = f.maxFeeDrops
	}

	log.Printf("Escalated fee for retry %d of transaction %s: %d -> %d drops",
		retryCount, originalTransaction.ID, originalFee, finalFee)

	return strconv.FormatInt(finalFee, 10), nil
}

// ValidateFee validates if a fee is reasonable for a transaction
func (f *FeeCalculator) ValidateFee(_ *models.Transaction, proposedFee string) error {
	fee, err := strconv.ParseInt(proposedFee, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid fee format: %w", err)
	}

	if fee < f.baseFeeDrops {
		return fmt.Errorf("fee %d drops is below minimum %d drops", fee, f.baseFeeDrops)
	}

	if fee > f.maxFeeDrops {
		return fmt.Errorf("fee %d drops exceeds maximum %d drops", fee, f.maxFeeDrops)
	}

	return nil
}

// GetFeeRecommendation provides fee recommendations based on urgency
func (f *FeeCalculator) GetFeeRecommendation(urgency string) (string, string, error) {
	networkLoad := f.EstimateNetworkLoad()
	baseFee := float64(f.baseFeeDrops)

	var lowFee, highFee int64

	switch strings.ToLower(urgency) {
	case "low":
		// Low urgency: minimal fee
		lowFee = int64(baseFee * (1.0 + networkLoad*0.5))
		highFee = int64(baseFee * (1.0 + networkLoad*0.8))
	case "normal":
		// Normal urgency: standard fee with network adjustment
		lowFee = int64(baseFee * (1.0 + networkLoad))
		highFee = int64(baseFee * (1.0 + networkLoad*1.5))
	case "high":
		// High urgency: premium fee
		lowFee = int64(baseFee * (1.0 + networkLoad*1.5))
		highFee = int64(baseFee * (1.0 + networkLoad*2.0))
	case "critical":
		// Critical urgency: maximum reasonable fee
		lowFee = int64(baseFee * (1.0 + networkLoad*2.0))
		highFee = f.maxFeeDrops
	default:
		return "", "", fmt.Errorf("invalid urgency level: %s", urgency)
	}

	// Ensure fees don't exceed maximum
	if lowFee > f.maxFeeDrops {
		lowFee = f.maxFeeDrops
	}
	if highFee > f.maxFeeDrops {
		highFee = f.maxFeeDrops
	}

	return strconv.FormatInt(lowFee, 10), strconv.FormatInt(highFee, 10), nil
}

// getPriorityMultiplier returns fee multiplier based on transaction priority
func (f *FeeCalculator) getPriorityMultiplier(priority models.TransactionPriority) float64 {
	switch priority {
	case models.PriorityLow:
		return 0.8 // 20% discount for low priority
	case models.PriorityNormal:
		return 1.0 // Standard fee
	case models.PriorityHigh:
		return 1.5 // 50% premium for high priority
	case models.PriorityCritical:
		return 2.0 // 100% premium for critical priority
	default:
		return 1.0
	}
}

// applyBatchOptimization applies fee optimization for batched transactions
func (f *FeeCalculator) applyBatchOptimization(individualFee int64, batchSize int, priority models.TransactionPriority) int64 {
	// Base batch discount
	batchDiscount := f.calculateBatchDiscount(batchSize)

	// Priority affects batch optimization
	priorityFactor := 1.0
	switch priority {
	case models.PriorityLow:
		priorityFactor = 1.2 // More aggressive optimization for low priority
	case models.PriorityNormal:
		priorityFactor = 1.0 // Standard optimization
	case models.PriorityHigh:
		priorityFactor = 0.8 // Less aggressive optimization for high priority
	case models.PriorityCritical:
		priorityFactor = 0.6 // Minimal optimization for critical priority
	}

	optimizedFee := float64(individualFee) * (1.0 - batchDiscount*priorityFactor)

	// Ensure optimized fee is not below minimum
	if optimizedFee < float64(f.baseFeeDrops) {
		optimizedFee = float64(f.baseFeeDrops)
	}

	return int64(math.Ceil(optimizedFee))
}

// calculateBatchDiscount calculates discount percentage based on batch size
func (f *FeeCalculator) calculateBatchDiscount(batchSize int) float64 {
	switch {
	case batchSize >= 10:
		return 0.15 // 15% discount for large batches
	case batchSize >= 5:
		return 0.10 // 10% discount for medium batches
	case batchSize >= 3:
		return 0.05 // 5% discount for small batches
	default:
		return 0.0 // No discount for very small batches
	}
}

// AnalyzeFeeEfficiency analyzes the fee efficiency of completed transactions
func (f *FeeCalculator) AnalyzeFeeEfficiency(transactions []*models.Transaction) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"total_transactions": 0,
			"analysis":           "No transactions to analyze",
		}
	}

	totalFee := int64(0)
	overpaidCount := 0
	underpaidCount := 0
	optimalCount := 0

	for _, tx := range transactions {
		if tx.Fee == "" {
			continue
		}

		fee, err := strconv.ParseInt(tx.Fee, 10, 64)
		if err != nil {
			continue
		}

		totalFee += fee

		// Compare with optimal fee (simplified analysis)
		optimalFee := f.baseFeeDrops
		switch tx.Type {
		case models.TransactionTypeEscrowCreate, models.TransactionTypeEscrowFinish, models.TransactionTypeEscrowCancel:
			optimalFee = 12
		case models.TransactionTypeWalletSetup:
			optimalFee = 15
		}

		feeRatio := float64(fee) / float64(optimalFee)
		if feeRatio > 1.5 {
			overpaidCount++
		} else if feeRatio < 0.8 {
			underpaidCount++
		} else {
			optimalCount++
		}
	}

	return map[string]interface{}{
		"total_transactions": len(transactions),
		"total_fee_drops":    totalFee,
		"average_fee_drops":  totalFee / int64(len(transactions)),
		"optimal_count":      optimalCount,
		"overpaid_count":     overpaidCount,
		"underpaid_count":    underpaidCount,
		"efficiency_score":   float64(optimalCount) / float64(len(transactions)) * 100,
	}
}
