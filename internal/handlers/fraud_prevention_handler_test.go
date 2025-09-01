package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// Mock services
type MockFraudDetectionService struct {
	mock.Mock
}

func (m *MockFraudDetectionService) AnalyzeTransaction(ctx context.Context, req *services.FraudAnalysisRequest) (*services.FraudAnalysisResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.FraudAnalysisResult), args.Error(1)
}

func (m *MockFraudDetectionService) DetectFraudPatterns(ctx context.Context, enterpriseID uuid.UUID) ([]*services.FraudPattern, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).([]*services.FraudPattern), args.Error(1)
}

func (m *MockFraudDetectionService) GetActiveRules(ctx context.Context) ([]*models.FraudRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.FraudRule), args.Error(1)
}

func (m *MockFraudDetectionService) CreateRule(ctx context.Context, rule *models.FraudRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFraudDetectionService) UpdateRule(ctx context.Context, rule *models.FraudRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFraudDetectionService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockFraudDetectionService) GetAlerts(ctx context.Context, filter *services.FraudAlertFilter) ([]*models.FraudAlert, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.FraudAlert), args.Error(1)
}

func (m *MockFraudDetectionService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, alertID, userID)
	return args.Error(0)
}

func (m *MockFraudDetectionService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error {
	args := m.Called(ctx, alertID, resolution, userID)
	return args.Error(0)
}

func (m *MockFraudDetectionService) CreateCase(ctx context.Context, req *services.FraudCaseRequest) (*models.FraudCase, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.FraudCase), args.Error(1)
}

func (m *MockFraudDetectionService) GetCase(ctx context.Context, caseID uuid.UUID) (*models.FraudCase, error) {
	args := m.Called(ctx, caseID)
	return args.Get(0).(*models.FraudCase), args.Error(1)
}

func (m *MockFraudDetectionService) UpdateCase(ctx context.Context, caseID uuid.UUID, updates *services.FraudCaseUpdate) error {
	args := m.Called(ctx, caseID, updates)
	return args.Error(0)
}

func (m *MockFraudDetectionService) CloseCase(ctx context.Context, caseID uuid.UUID, resolution *models.FraudCaseResolution, userID uuid.UUID) error {
	args := m.Called(ctx, caseID, resolution, userID)
	return args.Error(0)
}

func (m *MockFraudDetectionService) GetAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).(*models.AccountFraudStatus), args.Error(1)
}

func (m *MockFraudDetectionService) UpdateAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID, status models.AccountFraudStatusType, reason string, userID uuid.UUID) error {
	args := m.Called(ctx, enterpriseID, status, reason, userID)
	return args.Error(0)
}

func (m *MockFraudDetectionService) AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error {
	args := m.Called(ctx, enterpriseID, restriction)
	return args.Error(0)
}

func (m *MockFraudDetectionService) RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error {
	args := m.Called(ctx, enterpriseID, restrictionType)
	return args.Error(0)
}

func (m *MockFraudDetectionService) GenerateFraudReport(ctx context.Context, req *services.FraudReportRequest) (*services.FraudReport, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.FraudReport), args.Error(1)
}

func (m *MockFraudDetectionService) GetFraudMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*services.FraudMetrics, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).(*services.FraudMetrics), args.Error(1)
}

type MockFraudAlertingService struct {
	mock.Mock
}

func (m *MockFraudAlertingService) ProcessAlert(ctx context.Context, alert *models.FraudAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockFraudAlertingService) SendNotification(ctx context.Context, alertID uuid.UUID, channel services.FraudNotificationChannel) error {
	args := m.Called(ctx, alertID, channel)
	return args.Error(0)
}

func (m *MockFraudAlertingService) EscalateAlert(ctx context.Context, alertID uuid.UUID, reason string) error {
	args := m.Called(ctx, alertID, reason)
	return args.Error(0)
}

func (m *MockFraudAlertingService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID, notes string) error {
	args := m.Called(ctx, alertID, userID, notes)
	return args.Error(0)
}

func (m *MockFraudAlertingService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error {
	args := m.Called(ctx, alertID, resolution, userID)
	return args.Error(0)
}

func (m *MockFraudAlertingService) AssignAlert(ctx context.Context, alertID uuid.UUID, assignedTo uuid.UUID) error {
	args := m.Called(ctx, alertID, assignedTo)
	return args.Error(0)
}

func (m *MockFraudAlertingService) GetNotificationConfig(ctx context.Context, enterpriseID uuid.UUID) (*services.NotificationConfig, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).(*services.NotificationConfig), args.Error(1)
}

func (m *MockFraudAlertingService) UpdateNotificationConfig(ctx context.Context, enterpriseID uuid.UUID, config *services.NotificationConfig) error {
	args := m.Called(ctx, enterpriseID, config)
	return args.Error(0)
}

func (m *MockFraudAlertingService) GetEscalationRules(ctx context.Context) ([]*services.AlertEscalationRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*services.AlertEscalationRule), args.Error(1)
}

func (m *MockFraudAlertingService) UpdateEscalationRule(ctx context.Context, rule *services.AlertEscalationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFraudAlertingService) ConfigureNotificationChannels(ctx context.Context, enterpriseID uuid.UUID, channels []services.FraudNotificationChannel) error {
	args := m.Called(ctx, enterpriseID, channels)
	return args.Error(0)
}

func (m *MockFraudAlertingService) CorrelateAlerts(ctx context.Context, enterpriseID uuid.UUID, timeWindow time.Duration) ([]*services.AlertCorrelation, error) {
	args := m.Called(ctx, enterpriseID, timeWindow)
	return args.Get(0).([]*services.AlertCorrelation), args.Error(1)
}

func (m *MockFraudAlertingService) DeduplicateAlerts(ctx context.Context, alerts []*models.FraudAlert) ([]*models.FraudAlert, error) {
	args := m.Called(ctx, alerts)
	return args.Get(0).([]*models.FraudAlert), args.Error(1)
}

func (m *MockFraudAlertingService) CreateEscalationRule(ctx context.Context, rule *services.AlertEscalationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFraudAlertingService) DeleteEscalationRule(ctx context.Context, ruleID uuid.UUID) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

func (m *MockFraudAlertingService) GetNotificationChannels(ctx context.Context, enterpriseID uuid.UUID) ([]services.FraudNotificationChannel, error) {
	args := m.Called(ctx, enterpriseID)
	return args.Get(0).([]services.FraudNotificationChannel), args.Error(1)
}

func (m *MockFraudAlertingService) SendNotifications(ctx context.Context, alert *models.FraudAlert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

// Test setup
func setupTestRouter() (*gin.Engine, *MockFraudDetectionService, *MockFraudAlertingService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockDetectionService := &MockFraudDetectionService{}
	mockAlertingService := &MockFraudAlertingService{}

	handler := NewFraudPreventionHandler(mockDetectionService, mockAlertingService)
	handler.RegisterRoutes(router.Group("/api/v1"))

	return router, mockDetectionService, mockAlertingService
}

// Test cases
func TestAnalyzeTransaction(t *testing.T) {
	router, mockDetectionService, _ := setupTestRouter()

	// Test data
	enterpriseID := uuid.New()
	request := services.FraudAnalysisRequest{
		TransactionID:   "txn_123",
		EnterpriseID:    enterpriseID,
		Amount:          "1000.00",
		CurrencyCode:    "USD",
		TransactionType: "payment",
		Destination:     "dest_456",
		Timestamp:       time.Now(),
	}

	expectedResult := &services.FraudAnalysisResult{
		TransactionID:   "txn_123",
		EnterpriseID:    enterpriseID,
		RiskScore:       0.3,
		RiskLevel:       models.FraudRiskLevelLow,
		FraudDetected:   false,
		AlertGenerated:  false,
		RiskFactors:     []string{},
		Recommendations: []string{},
		Evidence:        map[string]interface{}{},
		ProcessingTime:  time.Millisecond * 50,
	}

	mockDetectionService.On("AnalyzeTransaction", mock.Anything, mock.AnythingOfType("*services.FraudAnalysisRequest")).Return(expectedResult, nil)

	// Create request
	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "result")
	mockDetectionService.AssertExpectations(t)
}

func TestGetActiveRules(t *testing.T) {
	router, mockDetectionService, _ := setupTestRouter()

	// Test data
	expectedRules := []*models.FraudRule{
		{
			ID:          uuid.New(),
			Name:        "High Amount Rule",
			Description: "Detects transactions above threshold",
			Category:    models.FraudRuleCategoryTransaction,
			RuleType:    models.FraudRuleTypeThreshold,
			Status:      models.FraudRuleStatusActive,
			Severity:    models.FraudSeverityHigh,
			BaseScore:   0.8,
			Confidence:  0.9,
		},
	}

	mockDetectionService.On("GetActiveRules", mock.Anything).Return(expectedRules, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/rules", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "rules")
	mockDetectionService.AssertExpectations(t)
}

func TestCreateRule(t *testing.T) {
	router, mockDetectionService, _ := setupTestRouter()

	// Test data
	rule := models.FraudRule{
		Name:        "Test Rule",
		Description: "Test rule for unit testing",
		Category:    models.FraudRuleCategoryTransaction,
		RuleType:    models.FraudRuleTypeThreshold,
		Status:      models.FraudRuleStatusActive,
		Severity:    models.FraudSeverityMedium,
		BaseScore:   0.5,
		Confidence:  0.7,
		Conditions: map[string]interface{}{
			"amount_threshold": 1000.0,
		},
		Thresholds: map[string]interface{}{
			"max_amount": 5000.0,
		},
		Actions: []models.FraudAction{
			models.FraudActionAlert,
		},
	}

	mockDetectionService.On("CreateRule", mock.Anything, mock.AnythingOfType("*models.FraudRule")).Return(nil)

	// Create request
	requestBody, _ := json.Marshal(rule)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/rules", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "rule")
	mockDetectionService.AssertExpectations(t)
}

func TestGetAlerts(t *testing.T) {
	router, mockDetectionService, _ := setupTestRouter()

	// Test data
	enterpriseID := uuid.New()
	expectedAlerts := []*models.FraudAlert{
		{
			ID:           uuid.New(),
			EnterpriseID: enterpriseID,
			AlertType:    models.FraudAlertTypeTransactionAnomaly,
			Severity:     models.FraudSeverityHigh,
			Status:       models.FraudAlertStatusNew,
			Score:        0.8,
			Confidence:   0.9,
			Title:        "High Risk Transaction",
			Description:  "Transaction amount exceeds normal patterns",
			DetectedAt:   time.Now(),
		},
	}

	mockDetectionService.On("GetAlerts", mock.Anything, mock.AnythingOfType("*services.FraudAlertFilter")).Return(expectedAlerts, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/alerts?enterprise_id="+enterpriseID.String(), nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "alerts")
	mockDetectionService.AssertExpectations(t)
}

func TestAcknowledgeAlert(t *testing.T) {
	router, _, mockAlertingService := setupTestRouter()

	// Test data
	alertID := uuid.New()
	userID := uuid.New()
	request := map[string]interface{}{
		"user_id": userID.String(),
		"notes":   "Acknowledged for investigation",
	}

	mockAlertingService.On("AcknowledgeAlert", mock.Anything, alertID, userID, "Acknowledged for investigation").Return(nil)

	// Create request
	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/alerts/"+alertID.String()+"/acknowledge", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "message")
	mockAlertingService.AssertExpectations(t)
}

func TestGetAccountFraudStatus(t *testing.T) {
	router, mockDetectionService, _ := setupTestRouter()

	// Test data
	enterpriseID := uuid.New()
	expectedStatus := &models.AccountFraudStatus{
		ID:              uuid.New(),
		EnterpriseID:    enterpriseID,
		Status:          models.AccountFraudStatusNormal,
		RiskScore:       0.2,
		RiskLevel:       models.FraudRiskLevelLow,
		RiskFactors:     []string{},
		Restrictions:    []models.AccountRestriction{},
		MonitoringLevel: models.MonitoringLevelStandard,
		NextReviewDate:  time.Now().AddDate(0, 0, 30),
		StatusChangedAt: time.Now(),
	}

	mockDetectionService.On("GetAccountFraudStatus", mock.Anything, enterpriseID).Return(expectedStatus, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/enterprises/"+enterpriseID.String()+"/status", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "status")
	mockDetectionService.AssertExpectations(t)
}

func TestInvalidUUID(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Test with invalid UUID
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/enterprises/invalid-uuid/status", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestInvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Test with invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
}
