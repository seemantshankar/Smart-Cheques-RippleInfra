package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionMonitoringService_AssessTransactionRisk(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	tests := []struct {
		name            string
		transaction     *models.Transaction
		expectedRisk    string
		expectedScore   float64
		expectedFactors []string
	}{
		{
			name: "Low risk transaction",
			transaction: &models.Transaction{
				ID:         uuid.New().String(),
				Type:       models.TransactionTypePayment,
				Amount:     "500.00",
				RetryCount: 0,
			},
			expectedRisk:    "low",
			expectedScore:   0.0,
			expectedFactors: []string{},
		},
		{
			name: "Medium risk transaction - medium amount",
			transaction: &models.Transaction{
				ID:         uuid.New().String(),
				Type:       models.TransactionTypePayment,
				Amount:     "5000.00",
				RetryCount: 0,
			},
			expectedRisk:    "low", // 0.1 is still low risk
			expectedScore:   0.1,
			expectedFactors: []string{"medium_transaction_amount"},
		},
		{
			name: "High risk transaction - high amount",
			transaction: &models.Transaction{
				ID:         uuid.New().String(),
				Type:       models.TransactionTypePayment,
				Amount:     "15000.00",
				RetryCount: 0,
			},
			expectedRisk:    "medium", // 0.3 is medium risk
			expectedScore:   0.3,
			expectedFactors: []string{"high_transaction_amount"},
		},
		{
			name: "Critical risk transaction - high amount with retries",
			transaction: &models.Transaction{
				ID:         uuid.New().String(),
				Type:       models.TransactionTypeEscrowCreate,
				Amount:     "15000.00",
				RetryCount: 2,
			},
			expectedRisk:    "high", // 0.5 is high risk
			expectedScore:   0.5,    // 0.3 (high amount) + 0.1 (escrow) + 0.1 (retries)
			expectedFactors: []string{"high_transaction_amount", "escrow_creation", "retry_count_2"},
		},
		{
			name: "Wallet setup transaction",
			transaction: &models.Transaction{
				ID:         uuid.New().String(),
				Type:       models.TransactionTypeWalletSetup,
				Amount:     "100.00",
				RetryCount: 0,
			},
			expectedRisk:    "low",
			expectedScore:   0.05,
			expectedFactors: []string{"wallet_setup_operation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskScore, err := service.AssessTransactionRisk(tt.transaction)

			assert.NoError(t, err)
			assert.NotNil(t, riskScore)
			assert.Equal(t, tt.expectedRisk, riskScore.RiskLevel)
			assert.Equal(t, tt.expectedScore, riskScore.RiskScore)
			assert.Equal(t, tt.expectedFactors, riskScore.RiskFactors)
			assert.Equal(t, tt.transaction.ID, riskScore.TransactionID)
			assert.Equal(t, "system", riskScore.AssessedBy)
		})
	}
}

func TestTransactionMonitoringService_PerformComplianceCheck(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	tests := []struct {
		name                 string
		transaction          *models.Transaction
		expectedStatus       string
		expectedViolations   []string
		expectedChecksPassed []string
		expectedChecksFailed []string
	}{
		{
			name: "Fully compliant transaction",
			transaction: &models.Transaction{
				ID:           uuid.New().String(),
				Amount:       "1000.00",
				FromAddress:  "rFromAddress123",
				ToAddress:    "rToAddress456",
				EnterpriseID: "enterprise-123",
			},
			expectedStatus:       "approved",
			expectedViolations:   []string{},
			expectedChecksPassed: []string{"amount_present", "addresses_valid", "enterprise_association"},
			expectedChecksFailed: []string{},
		},
		{
			name: "Transaction with missing amount",
			transaction: &models.Transaction{
				ID:           uuid.New().String(),
				Amount:       "",
				FromAddress:  "rFromAddress123",
				ToAddress:    "rToAddress456",
				EnterpriseID: "enterprise-123",
			},
			expectedStatus:       "flagged",
			expectedViolations:   []string{"Transaction amount is required"},
			expectedChecksPassed: []string{"addresses_valid", "enterprise_association"},
			expectedChecksFailed: []string{"amount_missing"},
		},
		{
			name: "Transaction with invalid addresses",
			transaction: &models.Transaction{
				ID:           uuid.New().String(),
				Amount:       "1000.00",
				FromAddress:  "",
				ToAddress:    "",
				EnterpriseID: "enterprise-123",
			},
			expectedStatus:       "flagged",
			expectedViolations:   []string{"Valid addresses are required"},
			expectedChecksPassed: []string{"amount_present", "enterprise_association"},
			expectedChecksFailed: []string{"addresses_invalid"},
		},
		{
			name: "Transaction with multiple violations",
			transaction: &models.Transaction{
				ID:           uuid.New().String(),
				Amount:       "",
				FromAddress:  "",
				ToAddress:    "",
				EnterpriseID: "",
			},
			expectedStatus: "rejected",
			expectedViolations: []string{
				"Transaction amount is required",
				"Valid addresses are required",
				"Enterprise association is required",
			},
			expectedChecksPassed: []string{},
			expectedChecksFailed: []string{"amount_missing", "addresses_invalid", "enterprise_missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complianceStatus, err := service.PerformComplianceCheck(tt.transaction)

			assert.NoError(t, err)
			assert.NotNil(t, complianceStatus)
			assert.Equal(t, tt.expectedStatus, complianceStatus.Status)
			assert.Equal(t, tt.expectedViolations, complianceStatus.Violations)
			assert.Equal(t, tt.expectedChecksPassed, complianceStatus.ChecksPassed)
			assert.Equal(t, tt.expectedChecksFailed, complianceStatus.ChecksFailed)
			assert.Equal(t, tt.transaction.ID, complianceStatus.TransactionID)
		})
	}
}

func TestTransactionMonitoringService_LogTransactionEvent(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	transactionID := uuid.New().String()
	eventType := "status_changed"
	previousStatus := models.TransactionStatusPending
	newStatus := models.TransactionStatusProcessing
	userID := uuid.New().String()
	enterpriseID := uuid.New().String()
	details := "Transaction status changed from pending to processing"
	ipAddress := "192.168.1.100"
	userAgent := "TestAgent/1.0"
	metadata := map[string]interface{}{
		"custom_field": "custom_value",
	}

	// Mock the audit repository
	enterpriseUUID := uuid.MustParse(enterpriseID)
	expectedAuditLog := &models.AuditLog{
		UserID:       uuid.MustParse(userID),
		EnterpriseID: &enterpriseUUID,
		Action:       "transaction_status_changed",
		Resource:     "transaction",
		ResourceID:   &transactionID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      true,
	}
	mockAuditRepo.On("CreateAuditLog", mock.AnythingOfType("*models.AuditLog")).Return(nil).Run(func(args mock.Arguments) {
		actualLog := args.Get(0).(*models.AuditLog)
		assert.Equal(t, expectedAuditLog.UserID, actualLog.UserID)
		assert.Equal(t, expectedAuditLog.Action, actualLog.Action)
		assert.Equal(t, expectedAuditLog.Resource, actualLog.Resource)
		assert.Equal(t, expectedAuditLog.ResourceID, actualLog.ResourceID)
		assert.Equal(t, expectedAuditLog.Details, actualLog.Details)
		assert.Equal(t, expectedAuditLog.IPAddress, actualLog.IPAddress)
		assert.Equal(t, expectedAuditLog.UserAgent, actualLog.UserAgent)
		assert.Equal(t, expectedAuditLog.Success, actualLog.Success)
	})

	// Execute
	err := service.LogTransactionEvent(
		transactionID,
		eventType,
		previousStatus,
		newStatus,
		userID,
		enterpriseID,
		details,
		ipAddress,
		userAgent,
		metadata,
	)

	// Assert
	assert.NoError(t, err)
	mockAuditRepo.AssertExpectations(t)
}

func TestTransactionMonitoringService_GenerateTransactionReport(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	enterpriseID := uuid.New().String()
	reportType := "daily"
	periodStart := time.Now().AddDate(0, 0, -1)
	periodEnd := time.Now()
	generatedBy := "test-user"

	report, err := service.GenerateTransactionReport(
		enterpriseID,
		reportType,
		periodStart,
		periodEnd,
		generatedBy,
	)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, reportType, report.ReportType)
	assert.Equal(t, enterpriseID, report.EnterpriseID)
	assert.Equal(t, periodStart, report.PeriodStart)
	assert.Equal(t, periodEnd, report.PeriodEnd)
	assert.Equal(t, generatedBy, report.GeneratedBy)

	// Check summary data
	summary := report.Summary
	assert.Equal(t, int64(150), summary.TotalTransactions)
	assert.Equal(t, int64(142), summary.SuccessfulTransactions)
	assert.Equal(t, int64(8), summary.FailedTransactions)
	assert.Equal(t, int64(12), summary.HighRiskTransactions)
	assert.Equal(t, int64(3), summary.ComplianceViolations)
	assert.Equal(t, 2.5, summary.AverageProcessingTime)
	assert.Equal(t, "125000.00", summary.TotalVolume)
	assert.Equal(t, "375.00", summary.TotalFees)

	// Check failure reasons
	expectedFailureReasons := map[string]int{
		"insufficient_funds": 3,
		"network_error":      2,
		"timeout":            2,
		"invalid_address":    1,
	}
	assert.Equal(t, expectedFailureReasons, summary.TopFailureReasons)

	// Check risk distribution
	expectedRiskDistribution := map[string]int{
		"low":      120,
		"medium":   20,
		"high":     8,
		"critical": 2,
	}
	assert.Equal(t, expectedRiskDistribution, summary.RiskDistribution)
}

func TestTransactionMonitoringService_GetTransactionStats(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	enterpriseID := uuid.New().String()
	since := time.Now().Add(-time.Hour)

	stats, err := service.GetTransactionStats(enterpriseID, &since)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(150), stats.TotalTransactions)
	assert.Equal(t, int64(5), stats.PendingTransactions)
	assert.Equal(t, int64(3), stats.ProcessingTransactions)
	assert.Equal(t, int64(140), stats.CompletedTransactions)
	assert.Equal(t, int64(2), stats.FailedTransactions)
	assert.Equal(t, 2.5, stats.AverageProcessingTime)
	assert.Equal(t, "375.00", stats.TotalFeesProcessed)
	assert.Equal(t, "45.50", stats.TotalFeeSavings)
	assert.Equal(t, &since, stats.LastProcessedAt)
}

func TestTransactionMonitoringService_GetTransactionStats_WithoutSince(t *testing.T) {
	// Setup
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	service := NewTransactionMonitoringService(mockTransactionRepo, mockAuditRepo, mockComplianceRepo)

	enterpriseID := uuid.New().String()

	stats, err := service.GetTransactionStats(enterpriseID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.LastProcessedAt)

	// Verify that LastProcessedAt is set to approximately 1 hour ago
	expectedTime := time.Now().Add(-time.Hour)
	timeDiff := stats.LastProcessedAt.Sub(expectedTime)
	assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute, "LastProcessedAt should be approximately 1 hour ago")
}
