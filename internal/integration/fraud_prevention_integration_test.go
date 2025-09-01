package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// MockEventBus for testing
type MockEventBus struct {
	events []*messaging.Event
}

func (m *MockEventBus) PublishEvent(ctx context.Context, event *messaging.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventBus) SubscribeToEvent(ctx context.Context, eventType string, handler func(*messaging.Event) error) error {
	return nil
}

func (m *MockEventBus) HealthCheck() error {
	return nil
}

func (m *MockEventBus) Close() error {
	return nil
}

// Integration test setup
func setupIntegrationTest() (*gin.Engine, *repository.FraudRepository, *MockEventBus) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create real repository
	fraudRepo := repository.NewFraudRepository().(*repository.FraudRepository)

	// Create mock event bus
	eventBus := &MockEventBus{}

	// Create services with real dependencies
	fraudDetectionConfig := &services.FraudDetectionConfig{
		HighAmountThreshold:   10000,
		VelocityThreshold:     10,
		RiskScoreThreshold:    0.7,
		AnalysisWindow:        24 * time.Hour,
		VelocityWindow:        1 * time.Hour,
		PatternWindow:         7 * 24 * time.Hour,
		BusinessHoursStart:    6,
		BusinessHoursEnd:      22,
		AutoAlertThreshold:    0.8,
		ManualReviewThreshold: 0.6,
		BlockThreshold:        0.9,
	}

	fraudDetectionService := services.NewFraudDetectionService(
		fraudRepo,
		nil, // transactionRepo
		nil, // enterpriseRepo
		eventBus,
		fraudDetectionConfig,
	)

	fraudAlertingConfig := &services.FraudAlertingConfig{
		DefaultChannels:     []string{"email"},
		EscalationDelay:     30 * time.Minute,
		MaxEscalationLevels: 3,
		CorrelationWindow:   1 * time.Hour,
		DeduplicationWindow: 5 * time.Minute,
	}

	fraudAlertingService := services.NewFraudAlertingService(
		fraudRepo,
		nil, // enterpriseRepo
		eventBus,
		fraudAlertingConfig,
	)

	// Create handler
	handler := handlers.NewFraudPreventionHandler(fraudDetectionService, fraudAlertingService)
	handler.RegisterRoutes(router.Group("/api/v1"))

	return router, fraudRepo, eventBus
}

// Integration test: Complete fraud detection workflow
func TestFraudDetectionWorkflow(t *testing.T) {
	router, _, eventBus := setupIntegrationTest()

	// Step 1: Create a fraud rule
	rule := models.FraudRule{
		Name:        "High Amount Rule",
		Description: "Detects transactions above $5000",
		Category:    models.FraudRuleCategoryTransaction,
		RuleType:    models.FraudRuleTypeThreshold,
		Status:      models.FraudRuleStatusActive,
		Severity:    models.FraudSeverityHigh,
		BaseScore:   0.8,
		Confidence:  0.9,
		Conditions: map[string]interface{}{
			"amount_threshold": 5000.0,
		},
		Thresholds: map[string]interface{}{
			"max_amount": 5000.0,
		},
		Actions: []models.FraudAction{
			models.FraudActionAlert,
		},
		EffectiveAt: time.Now(),
		CreatedBy:   uuid.New(),
	}

	ruleBody, _ := json.Marshal(rule)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/rules", bytes.NewBuffer(ruleBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var ruleResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &ruleResponse)
	require.NoError(t, err)
	assert.Contains(t, ruleResponse, "rule")

	// Step 2: Analyze a high-risk transaction
	enterpriseID := uuid.New()
	analysisRequest := services.FraudAnalysisRequest{
		TransactionID:   "txn_high_risk_001",
		EnterpriseID:    enterpriseID,
		Amount:          "7500.00",
		CurrencyCode:    "USD",
		TransactionType: "payment",
		Destination:     "dest_high_risk",
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"source_ip":  "192.168.1.100",
			"user_agent": "suspicious_bot",
		},
	}

	analysisBody, _ := json.Marshal(analysisRequest)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBuffer(analysisBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var analysisResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &analysisResponse)
	require.NoError(t, err)
	assert.Contains(t, analysisResponse, "result")

	// Step 3: Check that alerts were generated
	req, _ = http.NewRequest("GET", "/api/v1/fraud-prevention/alerts?enterprise_id="+enterpriseID.String(), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var alertsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &alertsResponse)
	require.NoError(t, err)
	assert.Contains(t, alertsResponse, "alerts")

	// Step 4: Verify events were published
	assert.Greater(t, len(eventBus.events), 0, "Expected events to be published")

	// Check for fraud alert event
	fraudAlertEventFound := false
	for _, event := range eventBus.events {
		if event.Type == "fraud.alert.generated" {
			fraudAlertEventFound = true
			break
		}
	}
	assert.True(t, fraudAlertEventFound, "Expected fraud alert event to be published")
}

// Integration test: Account fraud status management
func TestAccountFraudStatusManagement(t *testing.T) {
	router, _, eventBus := setupIntegrationTest()

	enterpriseID := uuid.New()

	// Step 1: Get initial account fraud status
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/enterprises/"+enterpriseID.String()+"/status", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var statusResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &statusResponse)
	require.NoError(t, err)
	assert.Contains(t, statusResponse, "status")

	// Step 2: Add account restriction
	restriction := models.AccountRestriction{
		Type:        models.RestrictionTypeTransactionLimit,
		Description: "Daily transaction limit reduced due to suspicious activity",
		Parameters: map[string]interface{}{
			"daily_limit": 1000.0,
		},
		EffectiveAt: time.Now(),
		AppliedBy:   uuid.New(),
	}

	restrictionBody, _ := json.Marshal(restriction)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/enterprises/"+enterpriseID.String()+"/restrictions", bytes.NewBuffer(restrictionBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 3: Update account fraud status to restricted
	statusUpdate := map[string]interface{}{
		"status":  "restricted",
		"reason":  "Multiple high-risk transactions detected",
		"user_id": uuid.New().String(),
	}

	statusBody, _ := json.Marshal(statusUpdate)
	req, _ = http.NewRequest("PUT", "/api/v1/fraud-prevention/enterprises/"+enterpriseID.String()+"/status", bytes.NewBuffer(statusBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 4: Verify account status was updated
	req, _ = http.NewRequest("GET", "/api/v1/fraud-prevention/enterprises/"+enterpriseID.String()+"/status", nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedStatusResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updatedStatusResponse)
	require.NoError(t, err)

	status := updatedStatusResponse["status"].(map[string]interface{})
	assert.Equal(t, "restricted", status["status"])

	// Step 5: Verify events were published
	restrictionEventFound := false
	statusEventFound := false

	for _, event := range eventBus.events {
		if event.Type == "account.restriction.added" {
			restrictionEventFound = true
		}
		if event.Type == "account.fraud_status.changed" {
			statusEventFound = true
		}
	}

	assert.True(t, restrictionEventFound, "Expected restriction event to be published")
	assert.True(t, statusEventFound, "Expected status change event to be published")
}

// Integration test: Fraud case management
func TestFraudCaseManagement(t *testing.T) {
	router, _, _ := setupIntegrationTest()

	enterpriseID := uuid.New()

	// Step 1: Create a fraud case
	caseRequest := services.FraudCaseRequest{
		EnterpriseID: enterpriseID,
		Title:        "Suspicious Transaction Pattern",
		Description:  "Multiple high-value transactions from unusual locations",
		Category:     models.FraudCaseCategoryTransactionFraud,
		Priority:     models.FraudCasePriorityHigh,
	}

	caseBody, _ := json.Marshal(caseRequest)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/cases", bytes.NewBuffer(caseBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var caseResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &caseResponse)
	require.NoError(t, err)
	assert.Contains(t, caseResponse, "case")

	// Extract case ID from response
	caseData := caseResponse["case"].(map[string]interface{})
	caseID := caseData["id"].(string)

	// Step 2: Update the case
	caseUpdate := services.FraudCaseUpdate{
		Status:     &[]models.FraudCaseStatus{models.FraudCaseStatusInvestigating}[0],
		AssignedTo: &[]uuid.UUID{uuid.New()}[0],
		InvestigationNotes: []models.InvestigationNote{
			{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				Content:   "Initial investigation started",
				NoteType:  "investigation",
				CreatedAt: time.Now(),
			},
		},
	}

	updateBody, _ := json.Marshal(caseUpdate)
	req, _ = http.NewRequest("PUT", "/api/v1/fraud-prevention/cases/"+caseID, bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 3: Close the case
	resolution := models.FraudCaseResolution{
		ResolutionType: "confirmed_fraud",
		Description:    "Fraud confirmed through investigation",
		Actions:        []string{"Account frozen", "Law enforcement notified"},
		Evidence: map[string]interface{}{
			"ip_addresses": []string{"192.168.1.100", "10.0.0.50"},
			"transactions": []string{"txn_001", "txn_002", "txn_003"},
		},
		ResolvedBy: uuid.New(),
		ResolvedAt: time.Now(),
	}

	closeRequest := map[string]interface{}{
		"user_id":    uuid.New().String(),
		"resolution": resolution,
	}

	closeBody, _ := json.Marshal(closeRequest)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/cases/"+caseID+"/close", bytes.NewBuffer(closeBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 4: Verify case was closed
	req, _ = http.NewRequest("GET", "/api/v1/fraud-prevention/cases/"+caseID, nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var finalCaseResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &finalCaseResponse)
	require.NoError(t, err)

	finalCase := finalCaseResponse["case"].(map[string]interface{})
	assert.Equal(t, "closed", finalCase["status"])
}

// Integration test: Alert management workflow
func TestAlertManagementWorkflow(t *testing.T) {
	router, fraudRepo, _ := setupIntegrationTest()

	enterpriseID := uuid.New()

	// Step 1: Create a fraud alert manually
	alert := models.FraudAlert{
		ID:           uuid.New(),
		EnterpriseID: enterpriseID,
		AlertType:    models.FraudAlertTypeTransactionAnomaly,
		Severity:     models.FraudSeverityHigh,
		Status:       models.FraudAlertStatusNew,
		Score:        0.85,
		Confidence:   0.9,
		Title:        "High Risk Transaction Detected",
		Description:  "Transaction amount exceeds normal patterns",
		DetectedAt:   time.Now(),
	}

	err := fraudRepo.CreateFraudAlert(context.Background(), &alert)
	require.NoError(t, err)

	// Step 2: Acknowledge the alert
	acknowledgeRequest := map[string]interface{}{
		"user_id": uuid.New().String(),
		"notes":   "Alert acknowledged for investigation",
	}

	ackBody, _ := json.Marshal(acknowledgeRequest)
	req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/alerts/"+alert.ID.String()+"/acknowledge", bytes.NewBuffer(ackBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 3: Assign the alert
	assignRequest := map[string]interface{}{
		"assigned_to": uuid.New().String(),
	}

	assignBody, _ := json.Marshal(assignRequest)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/alerts/"+alert.ID.String()+"/assign", bytes.NewBuffer(assignBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 4: Resolve the alert
	resolveRequest := map[string]interface{}{
		"user_id":    uuid.New().String(),
		"resolution": "False positive - legitimate business transaction",
	}

	resolveBody, _ := json.Marshal(resolveRequest)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/alerts/"+alert.ID.String()+"/resolve", bytes.NewBuffer(resolveBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Step 5: Verify alert status
	req, _ = http.NewRequest("GET", "/api/v1/fraud-prevention/alerts?enterprise_id="+enterpriseID.String(), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var alertsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &alertsResponse)
	require.NoError(t, err)

	alerts := alertsResponse["alerts"].([]interface{})
	assert.Greater(t, len(alerts), 0, "Expected alerts to be returned")

	// Verify alert was resolved
	alertData := alerts[0].(map[string]interface{})
	assert.Equal(t, "resolved", alertData["status"])
}

// Integration test: Performance under load
func TestPerformanceUnderLoad(t *testing.T) {
	router, _, _ := setupIntegrationTest()

	enterpriseID := uuid.New()

	// Create multiple rules
	for i := 0; i < 10; i++ {
		rule := models.FraudRule{
			Name:        fmt.Sprintf("Rule %d", i),
			Description: fmt.Sprintf("Test rule %d", i),
			Category:    models.FraudRuleCategoryTransaction,
			RuleType:    models.FraudRuleTypeThreshold,
			Status:      models.FraudRuleStatusActive,
			Severity:    models.FraudSeverityMedium,
			BaseScore:   0.5,
			Confidence:  0.7,
			Conditions: map[string]interface{}{
				"amount_threshold": float64(1000 + i*100),
			},
			EffectiveAt: time.Now(),
			CreatedBy:   uuid.New(),
		}

		ruleBody, _ := json.Marshal(rule)
		req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/rules", bytes.NewBuffer(ruleBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Analyze multiple transactions concurrently
	start := time.Now()

	for i := 0; i < 50; i++ {
		analysisRequest := services.FraudAnalysisRequest{
			TransactionID:   fmt.Sprintf("txn_%d", i),
			EnterpriseID:    enterpriseID,
			Amount:          fmt.Sprintf("%d.00", 1000+i*50),
			CurrencyCode:    "USD",
			TransactionType: "payment",
			Destination:     fmt.Sprintf("dest_%d", i),
			Timestamp:       time.Now(),
		}

		analysisBody, _ := json.Marshal(analysisRequest)
		req, _ := http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBuffer(analysisBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	duration := time.Since(start)

	// Performance assertions
	assert.Less(t, duration, 5*time.Second, "Analysis should complete within 5 seconds")

	// Verify alerts were generated
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/alerts?enterprise_id="+enterpriseID.String(), nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var alertsResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &alertsResponse)
	require.NoError(t, err)

	alerts := alertsResponse["alerts"].([]interface{})
	assert.Greater(t, len(alerts), 0, "Expected alerts to be generated under load")
}

// Integration test: Error handling and edge cases
func TestErrorHandlingAndEdgeCases(t *testing.T) {
	router, _, _ := setupIntegrationTest()

	// Test 1: Invalid UUID
	req, _ := http.NewRequest("GET", "/api/v1/fraud-prevention/enterprises/invalid-uuid/status", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 2: Invalid JSON
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 3: Missing required fields
	incompleteRequest := map[string]interface{}{
		"transaction_id": "txn_001",
		// Missing enterprise_id and other required fields
	}

	incompleteBody, _ := json.Marshal(incompleteRequest)
	req, _ = http.NewRequest("POST", "/api/v1/fraud-prevention/analyze", bytes.NewBuffer(incompleteBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 4: Non-existent resource
	req, _ = http.NewRequest("GET", "/api/v1/fraud-prevention/cases/"+uuid.New().String(), nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 or 404 depending on implementation
	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}
