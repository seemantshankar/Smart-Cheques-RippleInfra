package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestDisputeRefundService_RefundTypes(t *testing.T) {
	// Test refund type constants
	assert.Equal(t, "full", string(RefundTypeFull))
	assert.Equal(t, "partial", string(RefundTypePartial))
	assert.Equal(t, "penalty", string(RefundTypePenalty))
}

func TestDisputeRefundService_RefundPriorities(t *testing.T) {
	// Test refund priority constants
	assert.Equal(t, "low", string(RefundPriorityLow))
	assert.Equal(t, "normal", string(RefundPriorityNormal))
	assert.Equal(t, "high", string(RefundPriorityHigh))
	assert.Equal(t, "urgent", string(RefundPriorityUrgent))
}

func TestDisputeRefundService_RefundStatuses(t *testing.T) {
	// Test refund status constants
	assert.Equal(t, "pending", string(RefundStatusPending))
	assert.Equal(t, "approved", string(RefundStatusApproved))
	assert.Equal(t, "rejected", string(RefundStatusRejected))
	assert.Equal(t, "executing", string(RefundStatusExecuting))
	assert.Equal(t, "completed", string(RefundStatusCompleted))
	assert.Equal(t, "failed", string(RefundStatusFailed))
}

func TestDisputeRefundService_PaymentStatuses(t *testing.T) {
	// Test payment status constants
	assert.Equal(t, "pending", string(PaymentStatusPending))
	assert.Equal(t, "approved", string(PaymentStatusApproved))
	assert.Equal(t, "rejected", string(PaymentStatusRejected))
	assert.Equal(t, "executing", string(PaymentStatusExecuting))
	assert.Equal(t, "completed", string(PaymentStatusCompleted))
	assert.Equal(t, "failed", string(PaymentStatusFailed))
}

func TestDisputeRefundService_ValidateRefundRequest(t *testing.T) {
	service := &DisputeRefundService{}

	// Valid request
	validRequest := &RefundRequest{
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		RefundAmount:  100.0,
		Currency:      models.CurrencyUSDT,
		RefundReason:  "Service not delivered",
		RefundType:    RefundTypeFull,
		RequestedBy:   "user-123",
		Priority:      RefundPriorityNormal,
	}

	err := service.validateRefundRequest(validRequest)
	assert.NoError(t, err)

	// Invalid: zero amount
	invalidAmount := &RefundRequest{
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		RefundAmount:  0.0,
		Currency:      models.CurrencyUSDT,
		RefundReason:  "Service not delivered",
		RefundType:    RefundTypeFull,
		RequestedBy:   "user-123",
		Priority:      RefundPriorityNormal,
	}

	err = service.validateRefundRequest(invalidAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refund amount must be greater than 0")

	// Invalid: negative amount
	negativeAmount := &RefundRequest{
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		RefundAmount:  -50.0,
		Currency:      models.CurrencyUSDT,
		RefundReason:  "Service not delivered",
		RefundType:    RefundTypeFull,
		RequestedBy:   "user-123",
		Priority:      RefundPriorityNormal,
	}

	err = service.validateRefundRequest(negativeAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refund amount must be greater than 0")

	// Invalid: empty reason
	emptyReason := &RefundRequest{
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		RefundAmount:  100.0,
		Currency:      models.CurrencyUSDT,
		RefundReason:  "",
		RefundType:    RefundTypeFull,
		RequestedBy:   "user-123",
		Priority:      RefundPriorityNormal,
	}

	err = service.validateRefundRequest(emptyReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refund reason is required")

	// Invalid: empty requester
	emptyRequester := &RefundRequest{
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		RefundAmount:  100.0,
		Currency:      models.CurrencyUSDT,
		RefundReason:  "Service not delivered",
		RefundType:    RefundTypeFull,
		RequestedBy:   "",
		Priority:      RefundPriorityNormal,
	}

	err = service.validateRefundRequest(emptyRequester)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requester ID is required")
}

func TestDisputeRefundService_ValidatePartialPaymentRequest(t *testing.T) {
	service := &DisputeRefundService{}

	// Valid request
	validRequest := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      50.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err := service.validatePartialPaymentRequest(validRequest)
	assert.NoError(t, err)

	// Invalid: partial amount equals original amount
	equalAmount := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      100.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(equalAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial amount must be less than original amount")

	// Invalid: partial amount greater than original amount
	greaterAmount := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      150.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(greaterAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial amount must be less than original amount")

	// Invalid: zero partial amount
	zeroPartialAmount := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      0.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(zeroPartialAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial amount must be greater than 0")

	// Invalid: zero original amount
	zeroOriginalAmount := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      50.0,
		OriginalAmount:     0.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(zeroOriginalAmount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original amount must be greater than 0")

	// Invalid: empty payment reason
	emptyReason := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      50.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(emptyReason)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment reason is required")

	// Invalid: empty requester
	emptyRequester := &PartialPaymentRequest{
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		PartialAmount:      50.0,
		OriginalAmount:     100.0,
		Currency:           models.CurrencyUSDT,
		PaymentReason:      "Partial milestone completion",
		MilestoneCompleted: "milestone-1",
		RequestedBy:        "",
		Priority:           RefundPriorityNormal,
	}

	err = service.validatePartialPaymentRequest(emptyRequester)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requester ID is required")
}

func TestDisputeRefundService_CanProcessRefund(t *testing.T) {
	service := &DisputeRefundService{}

	// Should allow refunds for resolved disputes
	assert.True(t, service.canProcessRefund(models.DisputeStatusResolved))

	// Should allow refunds for closed disputes
	assert.True(t, service.canProcessRefund(models.DisputeStatusClosed))

	// Should not allow refunds for initiated disputes
	assert.False(t, service.canProcessRefund(models.DisputeStatusInitiated))

	// Should not allow refunds for under review disputes
	assert.False(t, service.canProcessRefund(models.DisputeStatusUnderReview))

	// Should not allow refunds for escalated disputes
	assert.False(t, service.canProcessRefund(models.DisputeStatusEscalated))
}

func TestDisputeRefundService_CanProcessPartialPayment(t *testing.T) {
	service := &DisputeRefundService{}

	// Should allow partial payments for resolved disputes
	assert.True(t, service.canProcessPartialPayment(models.DisputeStatusResolved))

	// Should allow partial payments for closed disputes
	assert.True(t, service.canProcessPartialPayment(models.DisputeStatusClosed))

	// Should not allow partial payments for initiated disputes
	assert.False(t, service.canProcessPartialPayment(models.DisputeStatusInitiated))

	// Should not allow partial payments for under review disputes
	assert.False(t, service.canProcessPartialPayment(models.DisputeStatusUnderReview))

	// Should not allow partial payments for escalated disputes
	assert.False(t, service.canProcessPartialPayment(models.DisputeStatusEscalated))
}

func TestDisputeRefundService_RefundRecord(t *testing.T) {
	// Test RefundRecord struct creation
	now := time.Now()
	refundRecord := &RefundRecord{
		ID:            "refund-123",
		DisputeID:     "dispute-123",
		SmartChequeID: "check-123",
		Amount:        100.0,
		Currency:      models.CurrencyUSDT,
		Status:        RefundStatusPending,
		RefundType:    RefundTypeFull,
		RequestedAt:   now,
		RequestedBy:   "user-123",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "refund-123", refundRecord.ID)
	assert.Equal(t, "dispute-123", refundRecord.DisputeID)
	assert.Equal(t, "check-123", refundRecord.SmartChequeID)
	assert.Equal(t, 100.0, refundRecord.Amount)
	assert.Equal(t, models.CurrencyUSDT, refundRecord.Currency)
	assert.Equal(t, RefundStatusPending, refundRecord.Status)
	assert.Equal(t, RefundTypeFull, refundRecord.RefundType)
	assert.Equal(t, now, refundRecord.RequestedAt)
	assert.Equal(t, "user-123", refundRecord.RequestedBy)
	assert.Equal(t, now, refundRecord.CreatedAt)
	assert.Equal(t, now, refundRecord.UpdatedAt)
}

func TestDisputeRefundService_PartialPaymentResult(t *testing.T) {
	// Test PartialPaymentResult struct creation
	now := time.Now()
	partialPaymentResult := &PartialPaymentResult{
		PaymentID:          "payment-123",
		DisputeID:          "dispute-123",
		SmartChequeID:      "check-123",
		Amount:             50.0,
		Currency:           models.CurrencyUSDT,
		Status:             PaymentStatusPending,
		MilestoneCompleted: "milestone-1",
		RequestedAt:        now,
		RequestedBy:        "user-123",
		Priority:           RefundPriorityNormal,
	}

	assert.Equal(t, "payment-123", partialPaymentResult.PaymentID)
	assert.Equal(t, "dispute-123", partialPaymentResult.DisputeID)
	assert.Equal(t, "check-123", partialPaymentResult.SmartChequeID)
	assert.Equal(t, 50.0, partialPaymentResult.Amount)
	assert.Equal(t, models.CurrencyUSDT, partialPaymentResult.Currency)
	assert.Equal(t, PaymentStatusPending, partialPaymentResult.Status)
	assert.Equal(t, "milestone-1", partialPaymentResult.MilestoneCompleted)
	assert.Equal(t, now, partialPaymentResult.RequestedAt)
	assert.Equal(t, "user-123", partialPaymentResult.RequestedBy)
	assert.Equal(t, RefundPriorityNormal, partialPaymentResult.Priority)
}
