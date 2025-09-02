package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smart-payment-infrastructure/internal/models"
)

func TestDisputeFundFreezingService_HelperMethods(t *testing.T) {
	// Test helper methods directly
	service := &DisputeFundFreezingService{}

	// Test canFreezeFunds helper method
	canFreeze := service.canFreezeFunds(models.DisputeStatusInitiated)
	assert.True(t, canFreeze)

	canFreeze = service.canFreezeFunds(models.DisputeStatusResolved)
	assert.False(t, canFreeze)

	// Test canUnfreezeFunds helper method
	canUnfreeze := service.canUnfreezeFunds(models.DisputeStatusResolved)
	assert.True(t, canUnfreeze)

	canUnfreeze = service.canUnfreezeFunds(models.DisputeStatusInitiated)
	assert.False(t, canUnfreeze)

	// Test isValidStatusTransition helper method
	isValid := service.isValidStatusTransition(FundFreezingStatusFrozen)
	assert.True(t, isValid)

	isValid = service.isValidStatusTransition("invalid_status")
	assert.False(t, isValid)
}

func TestFundFreezingStatusType_Constants(t *testing.T) {
	// Test that the constants are properly defined
	assert.Equal(t, "not_frozen", string(FundFreezingStatusNotFrozen))
	assert.Equal(t, "frozen", string(FundFreezingStatusFrozen))
	assert.Equal(t, "unfrozen", string(FundFreezingStatusUnfrozen))
	assert.Equal(t, "pending", string(FundFreezingStatusPending))
}
