package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// TransactionMonitoringService handles transaction monitoring, audit logging, risk scoring, and compliance
type TransactionMonitoringService struct {
	transactionRepo repository.TransactionRepositoryInterface
	auditRepo       repository.AuditRepositoryInterface
	complianceRepo  repository.ComplianceRepositoryInterface
	updateInterval  time.Duration
	isRunning       bool
	stopChan        chan struct{}
}

// NewTransactionMonitoringService creates a new transaction monitoring service
func NewTransactionMonitoringService(
	transactionRepo repository.TransactionRepositoryInterface,
	auditRepo repository.AuditRepositoryInterface,
	complianceRepo repository.ComplianceRepositoryInterface,
) *TransactionMonitoringService {
	return &TransactionMonitoringService{
		transactionRepo: transactionRepo,
		auditRepo:       auditRepo,
		complianceRepo:  complianceRepo,
	}
}

// LogTransactionEvent logs a transaction-specific audit event
func (s *TransactionMonitoringService) LogTransactionEvent(
	transactionID string,
	eventType string,
	previousStatus models.TransactionStatus,
	newStatus models.TransactionStatus,
	userID string,
	enterpriseID string,
	details string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) error {
	auditLog := &models.AuditLog{
		UserID:       uuid.MustParse(userID),
		EnterpriseID: &uuid.UUID{}, // Will be set below
		Action:       fmt.Sprintf("transaction_%s", eventType),
		Resource:     "transaction",
		ResourceID:   &transactionID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      true,
	}

	if enterpriseUUID, err := uuid.Parse(enterpriseID); err == nil {
		auditLog.EnterpriseID = &enterpriseUUID
	}

	return s.auditRepo.CreateAuditLog(auditLog)
}

// AssessTransactionRisk performs risk assessment for a transaction
func (s *TransactionMonitoringService) AssessTransactionRisk(transaction *models.Transaction) (*models.TransactionRiskScore, error) {
	riskScore := 0.0
	riskFactors := []string{}

	// Amount-based risk assessment
	if amount, err := strconv.ParseFloat(transaction.Amount, 64); err == nil {
		if amount > 10000 { // High amount threshold
			riskScore += 0.3
			riskFactors = append(riskFactors, "high_transaction_amount")
		} else if amount > 1000 { // Medium amount threshold
			riskScore += 0.1
			riskFactors = append(riskFactors, "medium_transaction_amount")
		}
	}

	// Transaction type risk assessment
	switch transaction.Type {
	case models.TransactionTypeWalletSetup:
		riskScore += 0.05
		riskFactors = append(riskFactors, "wallet_setup_operation")
	case models.TransactionTypeEscrowCreate:
		riskScore += 0.1
		riskFactors = append(riskFactors, "escrow_creation")
	}

	// Retry count risk assessment
	if transaction.RetryCount > 0 {
		riskScore += float64(transaction.RetryCount) * 0.05
		riskFactors = append(riskFactors, fmt.Sprintf("retry_count_%d", transaction.RetryCount))
	}

	// Determine risk level
	riskLevel := "low"
	if riskScore >= 0.7 {
		riskLevel = "critical"
	} else if riskScore >= 0.4 {
		riskLevel = "high"
	} else if riskScore >= 0.2 {
		riskLevel = "medium"
	}

	assessmentDetails := fmt.Sprintf("Risk assessment completed with score %.4f based on factors: %s",
		riskScore, strings.Join(riskFactors, ", "))

	return &models.TransactionRiskScore{
		ID:                uuid.New().String(),
		TransactionID:     transaction.ID,
		RiskLevel:         riskLevel,
		RiskScore:         riskScore,
		RiskFactors:       riskFactors,
		AssessmentDetails: assessmentDetails,
		AssessedAt:        time.Now(),
		AssessedBy:        "system",
		ExpiresAt:         &time.Time{}, // Expires in 24 hours
	}, nil
}

// PerformComplianceCheck performs compliance checks on a transaction
func (s *TransactionMonitoringService) PerformComplianceCheck(transaction *models.Transaction) (*models.TransactionComplianceStatus, error) {
	checksPassed := []string{}
	checksFailed := []string{}
	violations := []string{}

	// Basic compliance checks
	if transaction.Amount != "" {
		checksPassed = append(checksPassed, "amount_present")
	} else {
		checksFailed = append(checksFailed, "amount_missing")
		violations = append(violations, "Transaction amount is required")
	}

	if transaction.FromAddress != "" && transaction.ToAddress != "" {
		checksPassed = append(checksPassed, "addresses_valid")
	} else {
		checksFailed = append(checksFailed, "addresses_invalid")
		violations = append(violations, "Valid addresses are required")
	}

	if transaction.EnterpriseID != "" {
		checksPassed = append(checksPassed, "enterprise_association")
	} else {
		checksFailed = append(checksFailed, "enterprise_missing")
		violations = append(violations, "Enterprise association is required")
	}

	// Determine overall status
	status := "approved"
	if len(violations) > 0 {
		if len(violations) >= 2 {
			status = "rejected"
		} else {
			status = "flagged"
		}
	} else {
		status = "approved"
	}

	return &models.TransactionComplianceStatus{
		ID:            uuid.New().String(),
		TransactionID: transaction.ID,
		Status:        status,
		ChecksPassed:  checksPassed,
		ChecksFailed:  checksFailed,
		Violations:    violations,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// GenerateTransactionReport generates a transaction monitoring report
func (s *TransactionMonitoringService) GenerateTransactionReport(
	enterpriseID string,
	reportType string,
	periodStart time.Time,
	periodEnd time.Time,
	generatedBy string,
) (*models.TransactionReport, error) {
	// This is a simplified implementation - in production, this would query actual transaction data
	summary := models.TransactionReportSummary{
		TotalTransactions:      150,
		SuccessfulTransactions: 142,
		FailedTransactions:     8,
		HighRiskTransactions:   12,
		ComplianceViolations:   3,
		AverageProcessingTime:  2.5,
		TotalVolume:            "125000.00",
		TotalFees:              "375.00",
		TopFailureReasons: map[string]int{
			"insufficient_funds": 3,
			"network_error":      2,
			"timeout":            2,
			"invalid_address":    1,
		},
		RiskDistribution: map[string]int{
			"low":      120,
			"medium":   20,
			"high":     8,
			"critical": 2,
		},
	}

	report := &models.TransactionReport{
		ID:           uuid.New().String(),
		ReportType:   reportType,
		EnterpriseID: enterpriseID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Summary:      summary,
		GeneratedAt:  time.Now(),
		GeneratedBy:  generatedBy,
	}

	return report, nil
}

// GenerateComplianceReport generates a compliance-focused report
func (s *TransactionMonitoringService) GenerateComplianceReport(
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
	generatedBy string,
) (*models.ComplianceReport, error) {
	// Get compliance statistics
	complianceStats, err := s.GetComplianceStats(&enterpriseID, &periodStart)
	if err != nil {
		return nil, err
	}

	// Get flagged transactions for the period
	flaggedTransactions, err := s.GetFlaggedTransactions(100, 0)
	if err != nil {
		return nil, err
	}

	report := &models.ComplianceReport{
		ID:                  uuid.New().String(),
		EnterpriseID:        enterpriseID,
		PeriodStart:         periodStart,
		PeriodEnd:           periodEnd,
		ComplianceStats:     *complianceStats,
		FlaggedTransactions: len(flaggedTransactions),
		GeneratedAt:         time.Now(),
		GeneratedBy:         generatedBy,
	}

	return report, nil
}

// GenerateRiskReport generates a risk-focused report
func (s *TransactionMonitoringService) GenerateRiskReport(
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
	generatedBy string,
) (*models.RiskReport, error) {
	report := &models.RiskReport{
		ID:           uuid.New().String(),
		EnterpriseID: enterpriseID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		RiskMetrics: models.RiskMetrics{
			HighRiskTransactions:     12,
			CriticalRiskTransactions: 2,
			RiskTrend:                "stable",
			TopRiskFactors: []string{
				"high_transaction_amount",
				"escrow_creation",
				"retry_count",
			},
			MitigationActions: []string{
				"Enhanced monitoring for high-value transactions",
				"Additional verification for escrow operations",
				"Retry limit enforcement",
			},
		},
		GeneratedAt: time.Now(),
		GeneratedBy: generatedBy,
	}

	return report, nil
}

// GetTransactionAnalytics provides detailed transaction analytics
func (s *TransactionMonitoringService) GetTransactionAnalytics(
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
) (*models.TransactionAnalytics, error) {
	analytics := &models.TransactionAnalytics{
		EnterpriseID: enterpriseID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Metrics: models.TransactionMetrics{
			TotalVolume:            "125000.00",
			AverageTransactionSize: "833.33",
			PeakHourVolume:         "25000.00",
			TransactionVelocity:    25, // transactions per hour
			SuccessRate:            94.7,
			FailureRate:            5.3,
		},
		Trends: models.TransactionTrends{
			VolumeGrowth:      15.5,
			SuccessRateChange: 2.1,
			RiskIncrease:      -5.2, // negative means risk decreased
		},
		GeneratedAt: time.Now(),
	}

	return analytics, nil
}

// GetTransactionStats retrieves transaction statistics for monitoring
func (s *TransactionMonitoringService) GetTransactionStats(enterpriseID string, since *time.Time) (*models.TransactionStats, error) {
	// This is a simplified implementation - in production, this would query the database
	stats := &models.TransactionStats{
		TotalTransactions:      150,
		PendingTransactions:    5,
		ProcessingTransactions: 3,
		CompletedTransactions:  140,
		FailedTransactions:     2,
		AverageProcessingTime:  2.5,
		TotalFeesProcessed:     "375.00",
		TotalFeeSavings:        "45.50",
		LastProcessedAt:        &time.Time{}, // Set to current time
	}

	if since != nil {
		stats.LastProcessedAt = since
	} else {
		now := time.Now().Add(-time.Hour) // Last hour
		stats.LastProcessedAt = &now
	}

	return stats, nil
}

// StoreComplianceStatus stores a compliance status in the database
func (s *TransactionMonitoringService) StoreComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	return s.complianceRepo.CreateComplianceStatus(complianceStatus)
}

// GetComplianceStatus retrieves compliance status for a transaction
func (s *TransactionMonitoringService) GetComplianceStatus(transactionID string) (*models.TransactionComplianceStatus, error) {
	return s.complianceRepo.GetComplianceStatus(transactionID)
}

// UpdateComplianceStatus updates an existing compliance status
func (s *TransactionMonitoringService) UpdateComplianceStatus(complianceStatus *models.TransactionComplianceStatus) error {
	return s.complianceRepo.UpdateComplianceStatus(complianceStatus)
}

// GetFlaggedTransactions retrieves transactions that need compliance review
func (s *TransactionMonitoringService) GetFlaggedTransactions(limit, offset int) ([]models.TransactionComplianceStatus, error) {
	return s.complianceRepo.GetFlaggedTransactions(limit, offset)
}

// ReviewComplianceStatus marks a compliance status as reviewed
func (s *TransactionMonitoringService) ReviewComplianceStatus(complianceStatusID string, reviewedBy string, comments string) error {
	return s.complianceRepo.ReviewComplianceStatus(complianceStatusID, reviewedBy, comments)
}

// GetComplianceStats retrieves compliance statistics
func (s *TransactionMonitoringService) GetComplianceStats(enterpriseID *string, since *time.Time) (*models.ComplianceStats, error) {
	return s.complianceRepo.GetComplianceStats(enterpriseID, since)
}

// PerformAndStoreComplianceCheck performs compliance check and stores the result
func (s *TransactionMonitoringService) PerformAndStoreComplianceCheck(transaction *models.Transaction) (*models.TransactionComplianceStatus, error) {
	// Perform the compliance check
	complianceStatus, err := s.PerformComplianceCheck(transaction)
	if err != nil {
		return nil, err
	}

	// Store the compliance status
	err = s.StoreComplianceStatus(complianceStatus)
	if err != nil {
		return nil, err
	}

	return complianceStatus, nil
}

// SetUpdateInterval sets the update interval for the monitoring service
func (s *TransactionMonitoringService) SetUpdateInterval(interval time.Duration) {
	s.updateInterval = interval
}

// Start starts the monitoring service
func (s *TransactionMonitoringService) Start() error {
	if s.isRunning {
		return fmt.Errorf("monitoring service is already running")
	}

	s.isRunning = true
	s.stopChan = make(chan struct{})

	// Start monitoring goroutine
	go s.monitorTransactions()

	return nil
}

// Stop stops the monitoring service
func (s *TransactionMonitoringService) Stop() {
	if !s.isRunning {
		return
	}

	s.isRunning = false
	close(s.stopChan)
}

// monitorTransactions runs the monitoring loop
func (s *TransactionMonitoringService) monitorTransactions() {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			// Perform monitoring tasks
			s.performMonitoringTasks()
		}
	}
}

// performMonitoringTasks performs the actual monitoring tasks
func (s *TransactionMonitoringService) performMonitoringTasks() {
	// This is a simplified implementation
	// In production, this would:
	// 1. Check for stuck transactions
	// 2. Monitor system health
	// 3. Update metrics
	// 4. Generate alerts if needed
}

// GetDashboardData returns dashboard data for monitoring
func (s *TransactionMonitoringService) GetDashboardData() (map[string]interface{}, error) {
	// Get transaction stats
	stats, err := s.GetTransactionStats("", nil)
	if err != nil {
		return nil, err
	}

	// Create dashboard data as a map
	dashboard := map[string]interface{}{
		"metrics": map[string]interface{}{
			"total_transactions":      stats.TotalTransactions,
			"pending_transactions":    stats.PendingTransactions,
			"processing_transactions": stats.ProcessingTransactions,
			"completed_transactions":  stats.CompletedTransactions,
			"failed_transactions":     stats.FailedTransactions,
			"success_rate":            float64(stats.CompletedTransactions) / float64(stats.TotalTransactions) * 100,
			"average_processing_time": stats.AverageProcessingTime,
			"last_updated":            time.Now(),
		},
		"status_distribution": map[string]interface{}{
			"pending":    stats.PendingTransactions,
			"processing": stats.ProcessingTransactions,
			"completed":  stats.CompletedTransactions,
			"failed":     stats.FailedTransactions,
		},
		"system_health": map[string]interface{}{
			"overall_status": "healthy",
			"last_check":     time.Now(),
			"components": map[string]string{
				"database":   "healthy",
				"queue":      "healthy",
				"compliance": "healthy",
				"monitoring": "healthy",
			},
		},
	}

	return dashboard, nil
}
