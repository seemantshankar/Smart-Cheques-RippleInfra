package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/smart-payment-infrastructure/internal/models"
)

// MockEnterpriseRepository is a mock implementation of EnterpriseRepositoryInterface
type MockEnterpriseRepository struct {
	mock.Mock
}

func (m *MockEnterpriseRepository) CreateEnterprise(enterprise *models.Enterprise) error {
	args := m.Called(enterprise)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) GetEnterpriseByID(id uuid.UUID) (*models.Enterprise, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) GetEnterpriseByRegistrationNumber(regNumber string) (*models.Enterprise, error) {
	args := m.Called(regNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enterprise), args.Error(1)
}

func (m *MockEnterpriseRepository) UpdateEnterpriseKYBStatus(id uuid.UUID, status models.KYBStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) UpdateEnterpriseComplianceStatus(id uuid.UUID, status models.ComplianceStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) RegistrationNumberExists(regNumber string) (bool, error) {
	args := m.Called(regNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockEnterpriseRepository) CreateDocument(doc *models.EnterpriseDocument) error {
	args := m.Called(doc)
	return args.Error(0)
}

func (m *MockEnterpriseRepository) UpdateDocumentStatus(docID uuid.UUID, status models.DocumentStatus) error {
	args := m.Called(docID, status)
	return args.Error(0)
}

func TestEnterpriseService_RegisterEnterprise_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	req := &models.EnterpriseRegistrationRequest{
		LegalName:          "Test Corp Ltd",
		RegistrationNumber: "12345678",
		TaxID:              "TAX123456",
		Jurisdiction:       "US",
		BusinessType:       "Corporation",
		Industry:           "Technology",
		Email:              "contact@testcorp.com",
		Phone:              "+1234567890",
		Address: models.Address{
			Street1:    "123 Main St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "US",
		},
		AuthorizedRepresentative: models.AuthorizedRepresentativeRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@testcorp.com",
			Phone:     "+1234567890",
			Position:  "CEO",
		},
	}

	// Mock expectations
	mockRepo.On("RegistrationNumberExists", req.RegistrationNumber).Return(false, nil)
	mockRepo.On("CreateEnterprise", mock.AnythingOfType("*models.Enterprise")).Return(nil)

	// Execute
	enterprise, err := enterpriseService.RegisterEnterprise(req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, req.LegalName, enterprise.LegalName)
	assert.Equal(t, req.RegistrationNumber, enterprise.RegistrationNumber)
	assert.Equal(t, req.TaxID, enterprise.TaxID)
	assert.Equal(t, req.Jurisdiction, enterprise.Jurisdiction)
	assert.Equal(t, req.BusinessType, enterprise.BusinessType)
	assert.Equal(t, req.Industry, enterprise.Industry)
	assert.Equal(t, req.Email, enterprise.Email)
	assert.Equal(t, req.Phone, enterprise.Phone)
	assert.Equal(t, models.KYBStatusPending, enterprise.KYBStatus)
	assert.Equal(t, models.ComplianceStatusPending, enterprise.ComplianceStatus)
	assert.True(t, enterprise.IsActive)
	assert.Len(t, enterprise.AuthorizedRepresentatives, 1)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_RegisterEnterprise_AlreadyExists(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	req := &models.EnterpriseRegistrationRequest{
		LegalName:          "Test Corp Ltd",
		RegistrationNumber: "12345678",
		TaxID:              "TAX123456",
		Jurisdiction:       "US",
		BusinessType:       "Corporation",
		Industry:           "Technology",
		Email:              "contact@testcorp.com",
		Phone:              "+1234567890",
		Address: models.Address{
			Street1:    "123 Main St",
			City:       "New York",
			State:      "NY",
			PostalCode: "10001",
			Country:    "US",
		},
		AuthorizedRepresentative: models.AuthorizedRepresentativeRequest{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@testcorp.com",
			Phone:     "+1234567890",
			Position:  "CEO",
		},
	}

	// Mock expectations
	mockRepo.On("RegistrationNumberExists", req.RegistrationNumber).Return(true, nil)

	// Execute
	enterprise, err := enterpriseService.RegisterEnterprise(req)

	// Assert
	assert.Nil(t, enterprise)
	assert.Equal(t, ErrEnterpriseAlreadyExists, err)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_GetEnterpriseByID_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()
	expectedEnterprise := &models.Enterprise{
		ID:                 enterpriseID,
		LegalName:          "Test Corp Ltd",
		RegistrationNumber: "12345678",
		KYBStatus:          models.KYBStatusVerified,
		ComplianceStatus:   models.ComplianceStatusCompliant,
	}

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(expectedEnterprise, nil)

	// Execute
	enterprise, err := enterpriseService.GetEnterpriseByID(enterpriseID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedEnterprise, enterprise)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_GetEnterpriseByID_NotFound(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(nil, nil)

	// Execute
	enterprise, err := enterpriseService.GetEnterpriseByID(enterpriseID)

	// Assert
	assert.Nil(t, enterprise)
	assert.Equal(t, ErrEnterpriseNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_UpdateKYBStatus_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()
	enterprise := &models.Enterprise{
		ID:        enterpriseID,
		LegalName: "Test Corp Ltd",
		KYBStatus: models.KYBStatusPending,
	}

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(enterprise, nil)
	mockRepo.On("UpdateEnterpriseKYBStatus", enterpriseID, models.KYBStatusVerified).Return(nil)

	// Execute
	err := enterpriseService.UpdateKYBStatus(enterpriseID, models.KYBStatusVerified, "Approved")

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_UpdateComplianceStatus_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()
	enterprise := &models.Enterprise{
		ID:               enterpriseID,
		LegalName:        "Test Corp Ltd",
		ComplianceStatus: models.ComplianceStatusPending,
	}

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(enterprise, nil)
	mockRepo.On("UpdateEnterpriseComplianceStatus", enterpriseID, models.ComplianceStatusCompliant).Return(nil)

	// Execute
	err := enterpriseService.UpdateComplianceStatus(enterpriseID, models.ComplianceStatusCompliant)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_PerformKYBCheck_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()
	enterprise := &models.Enterprise{
		ID:        enterpriseID,
		LegalName: "Test Corp Ltd",
		KYBStatus: models.KYBStatusPending,
	}

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(enterprise, nil)
	mockRepo.On("UpdateEnterpriseKYBStatus", enterpriseID, models.KYBStatusInReview).Return(nil)

	// Execute
	err := enterpriseService.PerformKYBCheck(enterpriseID)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEnterpriseService_PerformComplianceCheck_Success(t *testing.T) {
	mockRepo := new(MockEnterpriseRepository)
	enterpriseService := NewEnterpriseService(mockRepo)

	enterpriseID := uuid.New()
	enterprise := &models.Enterprise{
		ID:               enterpriseID,
		LegalName:        "Test Corp Ltd",
		ComplianceStatus: models.ComplianceStatusPending,
	}

	// Mock expectations
	mockRepo.On("GetEnterpriseByID", enterpriseID).Return(enterprise, nil)
	mockRepo.On("UpdateEnterpriseComplianceStatus", enterpriseID, models.ComplianceStatusUnderReview).Return(nil)

	// Execute
	err := enterpriseService.PerformComplianceCheck(enterpriseID)

	// Assert
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}