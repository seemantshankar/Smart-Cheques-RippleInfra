package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestComplianceWorkflowIntegration(t *testing.T) {
	t.Run("Complete Compliance Workflow - Compliant Transaction", func(t *testing.T) {
		// Setup mocks
		mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		ruleEngine := NewRegulatoryRuleEngine(mockRuleRepo)
		enhancedValidation := NewEnhancedComplianceValidationService(
			ruleEngine,
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)
		violationDetection := NewComplianceViolationDetectionService(
			mockComplianceRepo,
			mockAuditRepo,
		)

		// Setup test transaction
		transaction := &models.Transaction{
			ID:           "tx-compliant-123",
			Type:         models.TransactionTypePayment,
			Status:       models.TransactionStatusPending,
			FromAddress:  "addr1",
			ToAddress:    "addr2",
			Amount:       "1000.00",
			Currency:     "USDT",
			EnterpriseID: "ent-123",
			UserID:       "user-123",
			Metadata:     map[string]interface{}{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Setup mock expectations for compliant transaction
		mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, "US").Return([]*models.RegulatoryRule{}, nil)
		mockComplianceRepo.On("CreateComplianceStatus", mock.Anything).Return(nil)
		mockAuditRepo.On("CreateAuditLog", mock.Anything).Return(nil)

		// Execute compliance validation
		result, err := enhancedValidation.ValidateTransactionCompliance(
			context.Background(),
			transaction,
			"basic",
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsCompliant)
		assert.Equal(t, "approved", result.Status)
		assert.True(t, result.ComplianceScore >= 0.8)

		// Verify no violations detected
		alerts, err := violationDetection.DetectViolations(context.Background(), result)
		assert.NoError(t, err)
		assert.Empty(t, alerts)

		// Verify mock expectations
		mockRuleRepo.AssertExpectations(t)
		mockComplianceRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Complete Compliance Workflow - Non-Compliant Transaction", func(t *testing.T) {
		// Setup mocks
		mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		ruleEngine := NewRegulatoryRuleEngine(mockRuleRepo)
		enhancedValidation := NewEnhancedComplianceValidationService(
			ruleEngine,
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)
		violationDetection := NewComplianceViolationDetectionService(
			mockComplianceRepo,
			mockAuditRepo,
		)

		// Setup test transaction with violations
		transaction := &models.Transaction{
			ID:           "tx-violation-123",
			Type:         models.TransactionTypePayment,
			Status:       models.TransactionStatusPending,
			FromAddress:  "blacklisted-addr",
			ToAddress:    "addr2",
			Amount:       "15000.00",
			Currency:     "USDT",
			EnterpriseID: "ent-123",
			UserID:       "user-123",
			Metadata:     map[string]interface{}{"transaction_count_24h": 15.0},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Setup mock expectations for non-compliant transaction
		mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, "US").Return([]*models.RegulatoryRule{
			{
				ID:           "rule-1",
				Name:         "Blacklist Check",
				Category:     "sanctions",
				Jurisdiction: "US",
				Priority:     5,
				RuleType:     "blacklist_check",
				Conditions:   map[string]interface{}{"blacklisted_addresses": []interface{}{"blacklisted-addr"}},
				Thresholds:   map[string]interface{}{},
				Actions:      []string{"reject"},
				Status:       "active",
				EffectiveAt:  time.Now().Add(-24 * time.Hour),
			},
			{
				ID:           "rule-2",
				Name:         "Amount Limit",
				Category:     "aml",
				Jurisdiction: "US",
				Priority:     3,
				RuleType:     "amount_limit",
				Conditions:   map[string]interface{}{},
				Thresholds:   map[string]interface{}{"max_amount": 10000.0},
				Actions:      []string{"flag"},
				Status:       "active",
				EffectiveAt:  time.Now().Add(-24 * time.Hour),
			},
		}, nil)
		mockComplianceRepo.On("CreateComplianceStatus", mock.Anything).Return(nil)
		mockAuditRepo.On("CreateAuditLog", mock.Anything).Return(nil)

		// Execute compliance validation
		result, err := enhancedValidation.ValidateTransactionCompliance(
			context.Background(),
			transaction,
			"strict",
		)

		// Debug output
		t.Logf("Compliance Result: %+v", result)
		if result != nil {
			t.Logf("Violations: %+v", result.Violations)
			t.Logf("Compliance Score: %.2f", result.ComplianceScore)
			t.Logf("Status: %s", result.Status)
		}

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsCompliant)
		assert.Equal(t, "rejected", result.Status)
		assert.True(t, result.ComplianceScore < 0.8)
		assert.NotEmpty(t, result.Violations)

		// Verify violations detected
		alerts, err := violationDetection.DetectViolations(context.Background(), result)
		assert.NoError(t, err)
		assert.NotEmpty(t, alerts)

		// Check for critical alert
		hasCriticalAlert := false
		for _, alert := range alerts {
			if alert.Severity == "critical" {
				hasCriticalAlert = true
				break
			}
		}
		assert.True(t, hasCriticalAlert, "Should have critical alert for blacklist violation")

		// Verify mock expectations
		mockRuleRepo.AssertExpectations(t)
		mockComplianceRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Compliance Reporting Integration", func(t *testing.T) {
		// Setup mocks
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		reportingService := NewComplianceReportingService(
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)

		// Setup mock expectations for reporting
		mockComplianceRepo.On("GetComplianceStats", mock.Anything, mock.Anything).Return(&models.ComplianceStats{
			TotalTransactions:    100,
			ApprovedTransactions: 85,
			RejectedTransactions: 5,
			FlaggedTransactions:  10,
			PendingTransactions:  0,
			ReviewedTransactions: 95,
		}, nil)
		mockComplianceRepo.On("GetComplianceStatusesByStatus", "flagged", 1000, 0).Return([]models.TransactionComplianceStatus{}, nil)

		// Generate compliance report
		report, err := reportingService.GenerateComplianceReport(
			context.Background(),
			"ent-123",
			"daily",
			time.Now().AddDate(0, 0, -1),
			time.Now(),
			"test-user",
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "ent-123", report.EnterpriseID)
		assert.Equal(t, "daily", report.ReportType)
		assert.Equal(t, 0.85, report.ComplianceRate)
		assert.Equal(t, int64(100), report.TotalTransactions)
		assert.Equal(t, int64(85), report.CompliantTransactions)
		assert.Equal(t, int64(15), report.NonCompliantTransactions)

		// Verify mock expectations
		mockComplianceRepo.AssertExpectations(t)
	})

	t.Run("Compliance Trend Analysis Integration", func(t *testing.T) {
		// Setup mocks
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		reportingService := NewComplianceReportingService(
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)

		// Setup mock expectations for trend analysis
		// The function calls GetComplianceStats for each day in the period
		// Use varying data to generate insights
		statsData := []*models.ComplianceStats{
			{TotalTransactions: 50, ApprovedTransactions: 45, RejectedTransactions: 2, FlaggedTransactions: 3, PendingTransactions: 0, ReviewedTransactions: 48}, // 90%
			{TotalTransactions: 60, ApprovedTransactions: 50, RejectedTransactions: 5, FlaggedTransactions: 5, PendingTransactions: 0, ReviewedTransactions: 55}, // 83%
			{TotalTransactions: 40, ApprovedTransactions: 30, RejectedTransactions: 5, FlaggedTransactions: 5, PendingTransactions: 0, ReviewedTransactions: 35}, // 75%
			{TotalTransactions: 55, ApprovedTransactions: 45, RejectedTransactions: 3, FlaggedTransactions: 7, PendingTransactions: 0, ReviewedTransactions: 52}, // 82%
			{TotalTransactions: 45, ApprovedTransactions: 40, RejectedTransactions: 2, FlaggedTransactions: 3, PendingTransactions: 0, ReviewedTransactions: 43}, // 89%
			{TotalTransactions: 50, ApprovedTransactions: 45, RejectedTransactions: 2, FlaggedTransactions: 3, PendingTransactions: 0, ReviewedTransactions: 48}, // 90%
			{TotalTransactions: 60, ApprovedTransactions: 55, RejectedTransactions: 2, FlaggedTransactions: 3, PendingTransactions: 0, ReviewedTransactions: 58}, // 92%
		}

		// Set up mock to return different values for each call (7 days)
		for i := 0; i < 7; i++ {
			mockComplianceRepo.On("GetComplianceStats", mock.Anything, mock.Anything).Return(statsData[i], nil).Once()
		}
		// Allow any extra calls to be handled gracefully with last day's data (defensive for time boundary)
		mockComplianceRepo.On("GetComplianceStats", mock.Anything, mock.Anything).Return(statsData[len(statsData)-1], nil).Maybe()

		// Generate compliance trend
		trend, err := reportingService.GenerateComplianceTrend(
			context.Background(),
			"ent-123",
			time.Now().AddDate(0, 0, -7),
			time.Now(),
			"compliance_rate",
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, trend)
		assert.Equal(t, "ent-123", trend.EnterpriseID)
		assert.Equal(t, "compliance_rate", trend.TrendType)
		assert.NotEmpty(t, trend.DataPoints)
		assert.NotEmpty(t, trend.KeyInsights)

		// Verify mock expectations
		mockComplianceRepo.AssertExpectations(t)
	})

	t.Run("Compliance Dashboard Integration", func(t *testing.T) {
		// Setup mocks
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		reportingService := NewComplianceReportingService(
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)

		// Setup mock expectations for dashboard
		mockComplianceRepo.On("GetComplianceStats", mock.Anything, mock.Anything).Return(&models.ComplianceStats{
			TotalTransactions:    200,
			ApprovedTransactions: 180,
			RejectedTransactions: 10,
			FlaggedTransactions:  10,
			PendingTransactions:  0,
			ReviewedTransactions: 190,
		}, nil)
		mockComplianceRepo.On("GetComplianceStatusesByStatus", "flagged", 10, 0).Return([]models.TransactionComplianceStatus{}, nil)

		// Generate dashboard data
		dashboard, err := reportingService.GenerateComplianceDashboard(
			context.Background(),
			"ent-123",
		)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, dashboard)
		assert.Equal(t, "ent-123", dashboard["enterprise_id"])
		assert.Equal(t, 0.9, dashboard["compliance_rate"])
		assert.Equal(t, int64(200), dashboard["total_transactions"])
		assert.Equal(t, int64(180), dashboard["approved_transactions"])

		// Verify mock expectations
		mockComplianceRepo.AssertExpectations(t)
	})
}

func TestComplianceWorkflowPerformance(t *testing.T) {
	// Setup mocks
	mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
	mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
	mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
	mockAuditRepo := &mocks.AuditRepositoryInterface{}

	// Create services
	ruleEngine := NewRegulatoryRuleEngine(mockRuleRepo)
	enhancedValidation := NewEnhancedComplianceValidationService(
		ruleEngine,
		mockComplianceRepo,
		mockTransactionRepo,
		mockAuditRepo,
	)

	t.Run("Performance Test - Multiple Transactions", func(t *testing.T) {
		// Setup mock expectations
		mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, "US").Return([]*models.RegulatoryRule{}, nil)
		mockComplianceRepo.On("CreateComplianceStatus", mock.Anything).Return(nil)
		mockAuditRepo.On("CreateAuditLog", mock.Anything).Return(nil)

		// Test with multiple transactions
		startTime := time.Now()
		for i := 0; i < 10; i++ {
			transaction := &models.Transaction{
				ID:           fmt.Sprintf("tx-perf-%d", i),
				Type:         models.TransactionTypePayment,
				Status:       models.TransactionStatusPending,
				FromAddress:  "addr1",
				ToAddress:    "addr2",
				Amount:       "1000.00",
				Currency:     "USDT",
				EnterpriseID: "ent-123",
				UserID:       "user-123",
				Metadata:     map[string]interface{}{},
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			result, err := enhancedValidation.ValidateTransactionCompliance(
				context.Background(),
				transaction,
				"basic",
			)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsCompliant)
		}

		duration := time.Since(startTime)
		assert.True(t, duration < 5*time.Second, "Performance test should complete within 5 seconds")

		// Verify mock expectations
		mockRuleRepo.AssertExpectations(t)
		mockComplianceRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})
}

func TestComplianceWorkflowErrorHandling(t *testing.T) {
	t.Run("Error Handling - Repository Failure", func(t *testing.T) {
		// Setup mocks
		mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		ruleEngine := NewRegulatoryRuleEngine(mockRuleRepo)
		enhancedValidation := NewEnhancedComplianceValidationService(
			ruleEngine,
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)

		// Setup transaction
		transaction := &models.Transaction{
			ID:           "tx-error-123",
			Type:         models.TransactionTypePayment,
			Status:       models.TransactionStatusPending,
			FromAddress:  "addr1",
			ToAddress:    "addr2",
			Amount:       "1000.00",
			Currency:     "USDT",
			EnterpriseID: "ent-123",
			UserID:       "user-123",
			Metadata:     map[string]interface{}{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Setup mock to return error
		mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, "US").Return([]*models.RegulatoryRule{}, assert.AnError)

		// Execute compliance validation
		result, err := enhancedValidation.ValidateTransactionCompliance(
			context.Background(),
			transaction,
			"basic",
		)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to evaluate regulatory compliance")

		// Verify mock expectations
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Error Handling - Invalid Transaction", func(t *testing.T) {
		// Setup mocks
		mockRuleRepo := &mocks.RegulatoryRuleRepositoryInterface{}
		mockComplianceRepo := &mocks.ComplianceRepositoryInterface{}
		mockTransactionRepo := &mocks.TransactionRepositoryInterface{}
		mockAuditRepo := &mocks.AuditRepositoryInterface{}

		// Create services
		ruleEngine := NewRegulatoryRuleEngine(mockRuleRepo)
		enhancedValidation := NewEnhancedComplianceValidationService(
			ruleEngine,
			mockComplianceRepo,
			mockTransactionRepo,
			mockAuditRepo,
		)

		// Setup invalid transaction
		transaction := &models.Transaction{
			ID:           "tx-invalid-123",
			Type:         models.TransactionTypePayment,
			Status:       models.TransactionStatusPending,
			FromAddress:  "", // Invalid: empty address
			ToAddress:    "", // Invalid: empty address
			Amount:       "", // Invalid: empty amount
			Currency:     "USDT",
			EnterpriseID: "", // Invalid: empty enterprise
			UserID:       "user-123",
			Metadata:     map[string]interface{}{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Setup mock expectations
		mockRuleRepo.On("GetActiveRegulatoryRules", mock.Anything, "US").Return([]*models.RegulatoryRule{}, nil)
		mockComplianceRepo.On("CreateComplianceStatus", mock.Anything).Return(nil)
		mockAuditRepo.On("CreateAuditLog", mock.Anything).Return(nil)

		// Execute compliance validation
		result, err := enhancedValidation.ValidateTransactionCompliance(
			context.Background(),
			transaction,
			"strict",
		)

		// Assertions
		assert.NoError(t, err) // Should not error, but should flag violations
		assert.NotNil(t, result)
		assert.False(t, result.IsCompliant)
		assert.Equal(t, "flagged", result.Status)
		assert.NotEmpty(t, result.Violations)

		// Check for specific violations
		hasAddressViolation := false
		hasAmountViolation := false
		hasEnterpriseViolation := false

		for _, violation := range result.Violations {
			if violation.RuleID == "basic_address_check" {
				hasAddressViolation = true
			}
			if violation.RuleID == "basic_amount_check" {
				hasAmountViolation = true
			}
			if violation.RuleID == "basic_enterprise_check" {
				hasEnterpriseViolation = true
			}
		}

		assert.True(t, hasAddressViolation, "Should have address violation")
		assert.True(t, hasAmountViolation, "Should have amount violation")
		assert.True(t, hasEnterpriseViolation, "Should have enterprise violation")

		// Verify mock expectations
		mockRuleRepo.AssertExpectations(t)
		mockComplianceRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})
}
