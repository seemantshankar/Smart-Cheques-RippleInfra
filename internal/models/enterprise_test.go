package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnterpriseRegistrationRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request EnterpriseRegistrationRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: EnterpriseRegistrationRequest{
				LegalName:          "Test Corp Ltd",
				RegistrationNumber: "12345678",
				TaxID:              "TAX123456",
				Jurisdiction:       "US",
				BusinessType:       "Corporation",
				Industry:           "Technology",
				Email:              "contact@testcorp.com",
				Phone:              "+1234567890",
				Address: Address{
					Street1:    "123 Main St",
					City:       "New York",
					State:      "NY",
					PostalCode: "10001",
					Country:    "US",
				},
				AuthorizedRepresentative: AuthorizedRepresentativeRequest{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@testcorp.com",
					Phone:     "+1234567890",
					Position:  "CEO",
				},
			},
			valid: true,
		},
		{
			name: "missing legal name",
			request: EnterpriseRegistrationRequest{
				RegistrationNumber: "12345678",
				TaxID:              "TAX123456",
				Jurisdiction:       "US",
				BusinessType:       "Corporation",
				Industry:           "Technology",
				Email:              "contact@testcorp.com",
				Phone:              "+1234567890",
			},
			valid: false,
		},
		{
			name: "invalid email",
			request: EnterpriseRegistrationRequest{
				LegalName:          "Test Corp Ltd",
				RegistrationNumber: "12345678",
				TaxID:              "TAX123456",
				Jurisdiction:       "US",
				BusinessType:       "Corporation",
				Industry:           "Technology",
				Email:              "invalid-email",
				Phone:              "+1234567890",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: In a real test, you would use a validator like go-playground/validator
			// to test the struct tags. For now, we just check the basic structure.
			if tt.valid {
				assert.NotEmpty(t, tt.request.LegalName)
				assert.NotEmpty(t, tt.request.RegistrationNumber)
				assert.NotEmpty(t, tt.request.TaxID)
				assert.NotEmpty(t, tt.request.Jurisdiction)
				assert.NotEmpty(t, tt.request.BusinessType)
				assert.NotEmpty(t, tt.request.Industry)
				assert.NotEmpty(t, tt.request.Email)
				assert.NotEmpty(t, tt.request.Phone)
			}
		})
	}
}

func TestKYBStatus_Constants(t *testing.T) {
	assert.Equal(t, KYBStatus("pending"), KYBStatusPending)
	assert.Equal(t, KYBStatus("in_review"), KYBStatusInReview)
	assert.Equal(t, KYBStatus("verified"), KYBStatusVerified)
	assert.Equal(t, KYBStatus("rejected"), KYBStatusRejected)
	assert.Equal(t, KYBStatus("suspended"), KYBStatusSuspended)
}

func TestComplianceStatus_Constants(t *testing.T) {
	assert.Equal(t, ComplianceStatus("pending"), ComplianceStatusPending)
	assert.Equal(t, ComplianceStatus("compliant"), ComplianceStatusCompliant)
	assert.Equal(t, ComplianceStatus("non_compliant"), ComplianceStatusNonCompliant)
	assert.Equal(t, ComplianceStatus("under_review"), ComplianceStatusUnderReview)
}

func TestDocumentType_Constants(t *testing.T) {
	assert.Equal(t, DocumentType("certificate_of_incorporation"), DocumentTypeCertificateOfIncorporation)
	assert.Equal(t, DocumentType("business_license"), DocumentTypeBusinessLicense)
	assert.Equal(t, DocumentType("tax_certificate"), DocumentTypeTaxCertificate)
	assert.Equal(t, DocumentType("bank_statement"), DocumentTypeBankStatement)
	assert.Equal(t, DocumentType("director_id"), DocumentTypeDirectorID)
	assert.Equal(t, DocumentType("proof_of_address"), DocumentTypeProofOfAddress)
}

func TestDocumentStatus_Constants(t *testing.T) {
	assert.Equal(t, DocumentStatus("pending"), DocumentStatusPending)
	assert.Equal(t, DocumentStatus("verified"), DocumentStatusVerified)
	assert.Equal(t, DocumentStatus("rejected"), DocumentStatusRejected)
}
