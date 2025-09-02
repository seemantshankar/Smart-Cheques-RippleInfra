package services

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// FraudDetectionServiceInterface defines the interface for fraud detection operations

// FraudDetectionServiceInterface defines the interface for fraud detection operations
type FraudDetectionServiceInterface interface {
	// Transaction analysis
	AnalyzeTransaction(ctx context.Context, req *FraudAnalysisRequest) (*FraudAnalysisResult, error)
	DetectFraudPatterns(ctx context.Context, enterpriseID uuid.UUID) ([]*FraudPattern, error)

	// Rule management
	GetActiveRules(ctx context.Context) ([]*models.FraudRule, error)
	CreateRule(ctx context.Context, rule *models.FraudRule) error
	UpdateRule(ctx context.Context, rule *models.FraudRule) error
	DeleteRule(ctx context.Context, ruleID uuid.UUID) error

	// Alert management
	GetAlerts(ctx context.Context, filter *FraudAlertFilter) ([]*models.FraudAlert, error)
	AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error
	ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error

	// Case management
	CreateCase(ctx context.Context, req *FraudCaseRequest) (*models.FraudCase, error)
	GetCase(ctx context.Context, caseID uuid.UUID) (*models.FraudCase, error)
	UpdateCase(ctx context.Context, caseID uuid.UUID, updates *FraudCaseUpdate) error
	CloseCase(ctx context.Context, caseID uuid.UUID, resolution *models.FraudCaseResolution, userID uuid.UUID) error

	// Account status management
	GetAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error)
	UpdateAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID, status models.AccountFraudStatusType, reason string, userID uuid.UUID) error
	AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error
	RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error

	// Reporting and analytics
	GenerateFraudReport(ctx context.Context, req *FraudReportRequest) (*FraudReport, error)
	GetFraudMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*FraudMetrics, error)
}

// FraudDetectionService implements the fraud detection service interface
type FraudDetectionService struct {
	fraudRepo       repository.FraudRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	enterpriseRepo  repository.EnterpriseRepositoryInterface
	messagingClient messaging.EventBus
	config          *FraudDetectionConfig
}

// NewFraudDetectionService creates a new fraud detection service instance
func NewFraudDetectionService(
	fraudRepo repository.FraudRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	enterpriseRepo repository.EnterpriseRepositoryInterface,
	messagingClient messaging.EventBus,
	config *FraudDetectionConfig,
) FraudDetectionServiceInterface {
	return &FraudDetectionService{
		fraudRepo:       fraudRepo,
		transactionRepo: transactionRepo,
		enterpriseRepo:  enterpriseRepo,
		messagingClient: messagingClient,
		config:          config,
	}
}

// Configuration types
type FraudDetectionConfig struct {
	// Detection thresholds
	HighAmountThreshold float64 `json:"high_amount_threshold"` // e.g., 10000
	VelocityThreshold   int     `json:"velocity_threshold"`    // e.g., 10 transactions per hour
	RiskScoreThreshold  float64 `json:"risk_score_threshold"`  // e.g., 0.7

	// Time windows
	AnalysisWindow time.Duration `json:"analysis_window"` // e.g., 24 hours
	VelocityWindow time.Duration `json:"velocity_window"` // e.g., 1 hour
	PatternWindow  time.Duration `json:"pattern_window"`  // e.g., 7 days

	// Business hours (24-hour format)
	BusinessHoursStart int `json:"business_hours_start"` // e.g., 6
	BusinessHoursEnd   int `json:"business_hours_end"`   // e.g., 22

	// Alert configuration
	AutoAlertThreshold    float64 `json:"auto_alert_threshold"`    // e.g., 0.8
	ManualReviewThreshold float64 `json:"manual_review_threshold"` // e.g., 0.6
	BlockThreshold        float64 `json:"block_threshold"`         // e.g., 0.9
}

// Request types
type FraudAnalysisRequest struct {
	TransactionID   string                 `json:"transaction_id" validate:"required"`
	EnterpriseID    uuid.UUID              `json:"enterprise_id" validate:"required"`
	Amount          string                 `json:"amount" validate:"required"`
	CurrencyCode    string                 `json:"currency_code" validate:"required"`
	TransactionType string                 `json:"transaction_type" validate:"required"`
	Destination     string                 `json:"destination"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type FraudAnalysisResult struct {
	TransactionID   string                 `json:"transaction_id"`
	EnterpriseID    uuid.UUID              `json:"enterprise_id"`
	RiskScore       float64                `json:"risk_score"`
	RiskLevel       models.FraudRiskLevel  `json:"risk_level"`
	FraudDetected   bool                   `json:"fraud_detected"`
	AlertGenerated  bool                   `json:"alert_generated"`
	AlertID         *uuid.UUID             `json:"alert_id,omitempty"`
	RiskFactors     []string               `json:"risk_factors"`
	Recommendations []string               `json:"recommendations"`
	Evidence        map[string]interface{} `json:"evidence"`
	ProcessingTime  time.Duration          `json:"processing_time"`
}

type FraudAlertFilter struct {
	EnterpriseID *uuid.UUID               `json:"enterprise_id,omitempty"`
	Status       *models.FraudAlertStatus `json:"status,omitempty"`
	Severity     *models.FraudSeverity    `json:"severity,omitempty"`
	AlertType    *models.FraudAlertType   `json:"alert_type,omitempty"`
	StartDate    *time.Time               `json:"start_date,omitempty"`
	EndDate      *time.Time               `json:"end_date,omitempty"`
	Limit        int                      `json:"limit"`
	Offset       int                      `json:"offset"`
}

type FraudCaseRequest struct {
	EnterpriseID   uuid.UUID                `json:"enterprise_id" validate:"required"`
	Title          string                   `json:"title" validate:"required"`
	Description    string                   `json:"description"`
	Category       models.FraudCaseCategory `json:"category" validate:"required"`
	Priority       models.FraudCasePriority `json:"priority"`
	AlertIDs       []uuid.UUID              `json:"alert_ids,omitempty"`
	TransactionIDs []string                 `json:"transaction_ids,omitempty"`
}

type FraudCaseUpdate struct {
	Status             *models.FraudCaseStatus    `json:"status,omitempty"`
	Priority           *models.FraudCasePriority  `json:"priority,omitempty"`
	AssignedTo         *uuid.UUID                 `json:"assigned_to,omitempty"`
	Investigator       *uuid.UUID                 `json:"investigator,omitempty"`
	InvestigationNotes []models.InvestigationNote `json:"investigation_notes,omitempty"`
}

type FraudReportRequest struct {
	EnterpriseID *uuid.UUID `json:"enterprise_id,omitempty"`
	StartDate    time.Time  `json:"start_date" validate:"required"`
	EndDate      time.Time  `json:"end_date" validate:"required"`
	ReportType   string     `json:"report_type"` // "summary", "detailed", "trends"
}

// Response types
type FraudPattern struct {
	PatternType string                 `json:"pattern_type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Evidence    map[string]interface{} `json:"evidence"`
	RiskScore   float64                `json:"risk_score"`
	DetectedAt  time.Time              `json:"detected_at"`
}

type FraudReport struct {
	ReportID     uuid.UUID  `json:"report_id"`
	GeneratedAt  time.Time  `json:"generated_at"`
	ReportPeriod TimeWindow `json:"report_period"`
	EnterpriseID *uuid.UUID `json:"enterprise_id,omitempty"`

	// Summary statistics
	TotalTransactions int `json:"total_transactions"`
	FraudAlerts       int `json:"fraud_alerts"`
	FraudCases        int `json:"fraud_cases"`
	ConfirmedFraud    int `json:"confirmed_fraud"`
	FalsePositives    int `json:"false_positives"`

	// Risk metrics
	AverageRiskScore     float64 `json:"average_risk_score"`
	HighRiskTransactions int     `json:"high_risk_transactions"`
	RiskTrend            string  `json:"risk_trend"` // "increasing", "decreasing", "stable"

	// Alert breakdown
	AlertBySeverity map[string]int `json:"alert_by_severity"`
	AlertByType     map[string]int `json:"alert_by_type"`

	// Recommendations
	Recommendations []string `json:"recommendations"`
}

type FraudMetrics struct {
	EnterpriseID     uuid.UUID             `json:"enterprise_id"`
	CurrentRiskScore float64               `json:"current_risk_score"`
	RiskLevel        models.FraudRiskLevel `json:"risk_level"`
	ActiveAlerts     int                   `json:"active_alerts"`
	OpenCases        int                   `json:"open_cases"`
	LastAlertDate    *time.Time            `json:"last_alert_date,omitempty"`
	RiskTrend        string                `json:"risk_trend"`
	TopRiskFactors   []string              `json:"top_risk_factors"`
}

// AnalyzeTransaction analyzes a transaction for fraud
func (s *FraudDetectionService) AnalyzeTransaction(ctx context.Context, req *FraudAnalysisRequest) (*FraudAnalysisResult, error) {
	startTime := time.Now()

	result := &FraudAnalysisResult{
		TransactionID:   req.TransactionID,
		EnterpriseID:    req.EnterpriseID,
		RiskScore:       0.0,
		RiskLevel:       models.FraudRiskLevelLow,
		FraudDetected:   false,
		AlertGenerated:  false,
		RiskFactors:     []string{},
		Recommendations: []string{},
		Evidence:        make(map[string]interface{}),
	}

	// Get active fraud rules
	activeRules, err := s.GetActiveRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active rules: %w", err)
	}

	// Apply each rule
	var totalWeightedScore float64
	var totalWeight float64

	for _, rule := range activeRules {
		if !rule.IsActive() {
			continue
		}

		score, evidence, riskFactors := s.evaluateRule(rule, req)
		if score > 0 {
			totalWeightedScore += score * rule.BaseScore
			totalWeight += rule.BaseScore

			// Add evidence and risk factors
			result.Evidence[rule.Name] = evidence
			result.RiskFactors = append(result.RiskFactors, riskFactors...)
		}
	}

	// Calculate final risk score
	if totalWeight > 0 {
		result.RiskScore = totalWeightedScore / totalWeight
	}

	// Determine risk level
	result.RiskLevel = s.determineRiskLevel(result.RiskScore)

	// Check if fraud is detected
	result.FraudDetected = result.RiskScore >= s.config.RiskScoreThreshold

	// Generate alert if threshold is met
	if result.RiskScore >= s.config.AutoAlertThreshold {
		alert, err := s.generateAlert(ctx, req, result)
		if err != nil {
			return nil, fmt.Errorf("failed to generate alert: %w", err)
		}
		result.AlertGenerated = true
		result.AlertID = &alert.ID
	}

	// Generate recommendations
	result.Recommendations = s.generateRecommendations(result)

	result.ProcessingTime = time.Since(startTime)

	// Publish fraud analysis event
	s.publishFraudAnalysisEvent(ctx, result)

	return result, nil
}

// evaluateRule evaluates a single fraud rule against a transaction
func (s *FraudDetectionService) evaluateRule(rule *models.FraudRule, req *FraudAnalysisRequest) (float64, map[string]interface{}, []string) {
	score := 0.0
	evidence := make(map[string]interface{})
	riskFactors := []string{}

	switch rule.RuleType {
	case models.FraudRuleTypeThreshold:
		score, evidence, riskFactors = s.evaluateThresholdRule(rule, req)
	case models.FraudRuleTypeVelocity:
		score, evidence, riskFactors = s.evaluateVelocityRule(rule, req)
	case models.FraudRuleTypePattern:
		score, evidence, riskFactors = s.evaluatePatternRule(rule, req)
	case models.FraudRuleTypeStatistical:
		score, evidence, riskFactors = s.evaluateStatisticalRule(rule, req)
	default:
		// Default evaluation for unknown rule types
		score = 0.1
		evidence["rule_type"] = "unknown"
		riskFactors = append(riskFactors, "Unknown rule type")
	}

	return score, evidence, riskFactors
}

// evaluateThresholdRule evaluates threshold-based rules
func (s *FraudDetectionService) evaluateThresholdRule(rule *models.FraudRule, req *FraudAnalysisRequest) (float64, map[string]interface{}, []string) {
	score := 0.0
	evidence := make(map[string]interface{})
	riskFactors := []string{}

	// Parse amount (support decimals)
	amountFloat, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		// Fallback: strip non-digits and try big.Int
		clean := strings.ReplaceAll(req.Amount, ".", "")
		amount := new(big.Int)
		if _, ok := amount.SetString(clean, 10); ok {
			f, _ := amount.Float64()
			amountFloat = f
		} else {
			return 0.0, evidence, riskFactors
		}
	}

	// Determine amount threshold from thresholds or conditions
	var thresholdFloat float64
	var hasThreshold bool
	if t, exists := rule.Thresholds["amount_threshold"]; exists {
		if v, ok := t.(float64); ok {
			thresholdFloat = v
			hasThreshold = true
		}
	} else if t, exists := rule.Thresholds["max_amount"]; exists {
		if v, ok := t.(float64); ok {
			thresholdFloat = v
			hasThreshold = true
		}
	} else if t, exists := rule.Conditions["amount_threshold"]; exists {
		if v, ok := t.(float64); ok {
			thresholdFloat = v
			hasThreshold = true
		}
	}

	// Check amount threshold
	if hasThreshold {
		if amountFloat > thresholdFloat {
			score = math.Min(amountFloat/thresholdFloat, 1.0)
			evidence["amount"] = amountFloat
			evidence["threshold"] = thresholdFloat
			evidence["exceeded_by_ratio"] = amountFloat / thresholdFloat
			riskFactors = append(riskFactors, "Transaction amount exceeds threshold")
		}
	}

	// Check transaction type conditions
	if transactionTypes, exists := rule.Conditions["transaction_type"]; exists {
		if types, ok := transactionTypes.([]interface{}); ok {
			for _, t := range types {
				if t.(string) == req.TransactionType {
					score = math.Max(score, 0.3)
					evidence["transaction_type"] = req.TransactionType
					riskFactors = append(riskFactors, "Transaction type matches risk pattern")
					break
				}
			}
		}
	}

	return score, evidence, riskFactors
}

// evaluateVelocityRule evaluates velocity-based rules
func (s *FraudDetectionService) evaluateVelocityRule(rule *models.FraudRule, _ *FraudAnalysisRequest) (float64, map[string]interface{}, []string) {
	score := 0.0
	evidence := make(map[string]interface{})
	riskFactors := []string{}

	// Get recent transaction count (simulated)
	recentCount := 5 // In real implementation, this would query the database

	// Check velocity threshold
	if maxTransactions, exists := rule.Thresholds["max_transactions"]; exists {
		if max, ok := maxTransactions.(float64); ok {
			if float64(recentCount) > max {
				score = math.Min(float64(recentCount)/max, 1.0)
				evidence["recent_transactions"] = recentCount
				evidence["max_allowed"] = max
				evidence["velocity_ratio"] = float64(recentCount) / max
				riskFactors = append(riskFactors, "High transaction velocity detected")
			}
		}
	}

	return score, evidence, riskFactors
}

// evaluatePatternRule evaluates pattern-based rules
func (s *FraudDetectionService) evaluatePatternRule(rule *models.FraudRule, req *FraudAnalysisRequest) (float64, map[string]interface{}, []string) {
	score := 0.0
	evidence := make(map[string]interface{})
	riskFactors := []string{}

	// Check business hours
	if businessHours, exists := rule.Conditions["business_hours"]; exists {
		if hours, ok := businessHours.(map[string]interface{}); ok {
			if startStr, exists := hours["start"]; exists {
				if endStr, exists := hours["end"]; exists {
					startHour := int(startStr.(float64))
					endHour := int(endStr.(float64))

					currentHour := req.Timestamp.Hour()
					if currentHour < startHour || currentHour > endHour {
						score = 0.4
						evidence["current_hour"] = currentHour
						evidence["business_hours"] = fmt.Sprintf("%d:00-%d:00", startHour, endHour)
						riskFactors = append(riskFactors, "Transaction outside business hours")
					}
				}
			}
		}
	}

	// Check for suspicious patterns in destination
	if req.Destination != "" {
		suspiciousKeywords := []string{"test", "demo", "temp", "external"}
		for _, keyword := range suspiciousKeywords {
			if strings.Contains(strings.ToLower(req.Destination), keyword) {
				score = math.Max(score, 0.2)
				evidence["suspicious_destination"] = req.Destination
				evidence["suspicious_keyword"] = keyword
				riskFactors = append(riskFactors, "Suspicious destination pattern")
				break
			}
		}
	}

	return score, evidence, riskFactors
}

// evaluateStatisticalRule evaluates statistical-based rules
func (s *FraudDetectionService) evaluateStatisticalRule(_ *models.FraudRule, req *FraudAnalysisRequest) (float64, map[string]interface{}, []string) {
	score := 0.0
	evidence := make(map[string]interface{})
	riskFactors := []string{}

	// Parse amount for statistical analysis
	amount := new(big.Int)
	if _, ok := amount.SetString(req.Amount, 10); !ok {
		return 0.0, evidence, riskFactors
	}

	// Simulate statistical analysis (Z-score calculation)
	amountFloat, _ := amount.Float64()
	historicalAvg := 10000.0   // Simulated historical average
	historicalStdDev := 5000.0 // Simulated standard deviation

	if historicalStdDev > 0 {
		zScore := math.Abs((amountFloat - historicalAvg) / historicalStdDev)
		if zScore > 2.0 {
			score = math.Min(zScore/3.0, 1.0) // Normalize to 0-1
			evidence["z_score"] = zScore
			evidence["historical_avg"] = historicalAvg
			evidence["historical_std_dev"] = historicalStdDev
			riskFactors = append(riskFactors, "Transaction amount is statistically unusual")
		}
	}

	return score, evidence, riskFactors
}

// determineRiskLevel determines the risk level based on score
func (s *FraudDetectionService) determineRiskLevel(score float64) models.FraudRiskLevel {
	switch {
	case score >= 0.8:
		return models.FraudRiskLevelCritical
	case score >= 0.6:
		return models.FraudRiskLevelHigh
	case score >= 0.4:
		return models.FraudRiskLevelMedium
	default:
		return models.FraudRiskLevelLow
	}
}

// generateAlert creates a fraud alert
func (s *FraudDetectionService) generateAlert(ctx context.Context, req *FraudAnalysisRequest, result *FraudAnalysisResult) (*models.FraudAlert, error) {
	alert := &models.FraudAlert{
		ID:                   uuid.New(),
		EnterpriseID:         req.EnterpriseID,
		TransactionID:        &req.TransactionID,
		AlertType:            s.determineAlertType(req, result),
		Severity:             s.determineSeverity(result.RiskScore),
		Status:               models.FraudAlertStatusNew,
		Score:                result.RiskScore,
		Confidence:           0.8, // Default confidence
		DetectionMethod:      "rule_based",
		Evidence:             result.Evidence,
		Title:                s.generateAlertTitle(result),
		Description:          s.generateAlertDescription(req, result),
		Recommendation:       s.generateAlertRecommendation(result),
		NotificationChannels: []string{"email", "webhook"},
		DetectedAt:           time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save alert to database
	if err := s.fraudRepo.CreateFraudAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to save fraud alert: %w", err)
	}

	// Publish alert event
	s.publishFraudAlertEvent(ctx, alert)

	return alert, nil
}

// determineAlertType determines the type of alert based on analysis
func (s *FraudDetectionService) determineAlertType(_ *FraudAnalysisRequest, result *FraudAnalysisResult) models.FraudAlertType {
	if result.RiskScore >= 0.8 {
		return models.FraudAlertTypeTransactionAnomaly
	} else if len(result.RiskFactors) > 2 {
		return models.FraudAlertTypePatternAnomaly
	} else {
		return models.FraudAlertTypeSuspiciousActivity
	}
}

// determineSeverity determines alert severity based on risk score
func (s *FraudDetectionService) determineSeverity(riskScore float64) models.FraudSeverity {
	switch {
	case riskScore >= 0.8:
		return models.FraudSeverityCritical
	case riskScore >= 0.6:
		return models.FraudSeverityHigh
	case riskScore >= 0.4:
		return models.FraudSeverityMedium
	default:
		return models.FraudSeverityLow
	}
}

// generateAlertTitle generates a title for the alert
func (s *FraudDetectionService) generateAlertTitle(result *FraudAnalysisResult) string {
	switch result.RiskLevel {
	case models.FraudRiskLevelCritical:
		return "Critical Fraud Risk Detected"
	case models.FraudRiskLevelHigh:
		return "High Fraud Risk Detected"
	case models.FraudRiskLevelMedium:
		return "Medium Fraud Risk Detected"
	default:
		return "Low Fraud Risk Detected"
	}
}

// generateAlertDescription generates a description for the alert
func (s *FraudDetectionService) generateAlertDescription(req *FraudAnalysisRequest, result *FraudAnalysisResult) string {
	description := fmt.Sprintf("Transaction %s with amount %s %s has been flagged for fraud risk.",
		req.TransactionID, req.Amount, req.CurrencyCode)

	if len(result.RiskFactors) > 0 {
		description += " Risk factors: " + strings.Join(result.RiskFactors, ", ")
	}

	return description
}

// generateAlertRecommendation generates recommendations for the alert
func (s *FraudDetectionService) generateAlertRecommendation(result *FraudAnalysisResult) string {
	switch result.RiskLevel {
	case models.FraudRiskLevelCritical:
		return "Immediate review required. Consider blocking transaction and freezing account."
	case models.FraudRiskLevelHigh:
		return "Manual review recommended. Hold transaction pending investigation."
	case models.FraudRiskLevelMedium:
		return "Enhanced monitoring recommended. Flag for follow-up review."
	default:
		return "Standard monitoring. No immediate action required."
	}
}

// generateRecommendations generates recommendations based on analysis
func (s *FraudDetectionService) generateRecommendations(result *FraudAnalysisResult) []string {
	recommendations := []string{}

	if result.RiskScore >= 0.8 {
		recommendations = append(recommendations, "Consider blocking this transaction")
		recommendations = append(recommendations, "Review account for suspicious activity")
	}

	if result.RiskScore >= 0.6 {
		recommendations = append(recommendations, "Hold transaction for manual review")
		recommendations = append(recommendations, "Increase monitoring for this enterprise")
	}

	if len(result.RiskFactors) > 2 {
		recommendations = append(recommendations, "Multiple risk factors detected - investigate further")
	}

	return recommendations
}

// Publish events
func (s *FraudDetectionService) publishFraudAnalysisEvent(ctx context.Context, result *FraudAnalysisResult) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   "fraud.analysis.completed",
		Source: "fraud-detection-service",
		Data: map[string]interface{}{
			"transaction_id":  result.TransactionID,
			"enterprise_id":   result.EnterpriseID.String(),
			"risk_score":      result.RiskScore,
			"risk_level":      result.RiskLevel,
			"fraud_detected":  result.FraudDetected,
			"alert_generated": result.AlertGenerated,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish fraud analysis event: %v\n", err)
	}
}

func (s *FraudDetectionService) publishFraudAlertEvent(ctx context.Context, alert *models.FraudAlert) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   "fraud.alert.generated",
		Source: "fraud-detection-service",
		Data: map[string]interface{}{
			"alert_id":      alert.ID.String(),
			"enterprise_id": alert.EnterpriseID.String(),
			"alert_type":    alert.AlertType,
			"severity":      alert.Severity,
			"score":         alert.Score,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish fraud alert event: %v\n", err)
	}
}

// Placeholder implementations for other interface methods
func (s *FraudDetectionService) DetectFraudPatterns(ctx context.Context, enterpriseID uuid.UUID) ([]*FraudPattern, error) {
	// In a real implementation, this would analyze patterns across multiple transactions
	return []*FraudPattern{}, nil
}

func (s *FraudDetectionService) GetActiveRules(ctx context.Context) ([]*models.FraudRule, error) {
	if s.fraudRepo == nil {
		return []*models.FraudRule{}, nil
	}
	return s.fraudRepo.GetActiveFraudRules(ctx)
}

func (s *FraudDetectionService) CreateRule(ctx context.Context, rule *models.FraudRule) error {
	if s.fraudRepo == nil {
		return nil
	}
	return s.fraudRepo.CreateFraudRule(ctx, rule)
}

func (s *FraudDetectionService) UpdateRule(ctx context.Context, rule *models.FraudRule) error {
	if s.fraudRepo == nil {
		return nil
	}
	return s.fraudRepo.UpdateFraudRule(ctx, rule)
}

func (s *FraudDetectionService) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	if s.fraudRepo == nil {
		return nil
	}
	return s.fraudRepo.DeleteFraudRule(ctx, ruleID)
}

func (s *FraudDetectionService) GetAlerts(ctx context.Context, filter *FraudAlertFilter) ([]*models.FraudAlert, error) {
	if s.fraudRepo == nil {
		return []*models.FraudAlert{}, nil
	}
	// Convert to repository filter
	repoFilter := &repository.FraudAlertFilter{}
	if filter != nil {
		repoFilter.EnterpriseID = filter.EnterpriseID
		repoFilter.Status = filter.Status
		repoFilter.Severity = filter.Severity
		repoFilter.AlertType = filter.AlertType
		repoFilter.StartDate = filter.StartDate
		repoFilter.EndDate = filter.EndDate
	}
	limit := 1000
	offset := 0
	if filter != nil {
		if filter.Limit > 0 {
			limit = filter.Limit
		}
		offset = filter.Offset
	}
	return s.fraudRepo.ListFraudAlerts(ctx, repoFilter, limit, offset)
}

func (s *FraudDetectionService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error {
	// In a real implementation, this would update the alert status
	return nil
}

func (s *FraudDetectionService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error {
	// In a real implementation, this would resolve the alert
	return nil
}

func (s *FraudDetectionService) CreateCase(ctx context.Context, req *FraudCaseRequest) (*models.FraudCase, error) {
	if s.fraudRepo == nil {
		return &models.FraudCase{}, nil
	}
	caseObj := &models.FraudCase{
		ID:           uuid.New(),
		EnterpriseID: req.EnterpriseID,
		CaseNumber:   fmt.Sprintf("FC-%s", time.Now().Format("20060102-150405")),
		Status:       models.FraudCaseStatusOpen,
		Priority:     req.Priority,
		Title:        req.Title,
		Description:  req.Description,
		Category:     req.Category,
		Alerts:       req.AlertIDs,
		Transactions: req.TransactionIDs,
		OpenedAt:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.fraudRepo.CreateFraudCase(ctx, caseObj); err != nil {
		return nil, err
	}
	return caseObj, nil
}

func (s *FraudDetectionService) GetCase(ctx context.Context, caseID uuid.UUID) (*models.FraudCase, error) {
	if s.fraudRepo == nil {
		return &models.FraudCase{}, nil
	}
	return s.fraudRepo.GetFraudCaseByID(ctx, caseID)
}

func (s *FraudDetectionService) UpdateCase(ctx context.Context, caseID uuid.UUID, updates *FraudCaseUpdate) error {
	if s.fraudRepo == nil {
		return nil
	}
	fc, err := s.fraudRepo.GetFraudCaseByID(ctx, caseID)
	if err != nil {
		return err
	}
	if updates.Status != nil {
		fc.Status = *updates.Status
	}
	if updates.Priority != nil {
		fc.Priority = *updates.Priority
	}
	if updates.AssignedTo != nil {
		fc.AssignedTo = updates.AssignedTo
		now := time.Now()
		fc.AssignedAt = &now
	}
	if updates.Investigator != nil {
		fc.Investigator = updates.Investigator
	}
	if len(updates.InvestigationNotes) > 0 {
		fc.InvestigationNotes = append(fc.InvestigationNotes, updates.InvestigationNotes...)
	}
	fc.UpdatedAt = time.Now()
	return s.fraudRepo.UpdateFraudCase(ctx, fc)
}

func (s *FraudDetectionService) CloseCase(ctx context.Context, caseID uuid.UUID, resolution *models.FraudCaseResolution, _ uuid.UUID) error {
	if s.fraudRepo == nil {
		return nil
	}
	fc, err := s.fraudRepo.GetFraudCaseByID(ctx, caseID)
	if err != nil {
		return err
	}
	// Set resolution and move to resolved then closed
	fc.Resolution = resolution
	fc.Status = models.FraudCaseStatusResolved
	now := time.Now()
	fc.ResolvedAt = &now
	fc.UpdatedAt = now
	if err := s.fraudRepo.UpdateFraudCase(ctx, fc); err != nil {
		return err
	}
	if err := fc.Close(); err != nil {
		return err
	}
	return s.fraudRepo.UpdateFraudCase(ctx, fc)
}

func (s *FraudDetectionService) GetAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID) (*models.AccountFraudStatus, error) {
	if s.fraudRepo == nil {
		return &models.AccountFraudStatus{}, nil
	}
	return s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
}

func (s *FraudDetectionService) UpdateAccountFraudStatus(ctx context.Context, enterpriseID uuid.UUID, status models.AccountFraudStatusType, reason string, userID uuid.UUID) error {
	if s.fraudRepo == nil {
		return nil
	}
	current, err := s.fraudRepo.GetAccountFraudStatusByEnterprise(ctx, enterpriseID)
	if err != nil {
		return err
	}
	// Append status history if available
	change := models.FraudStatusChange{
		FromStatus: current.Status,
		ToStatus:   status,
		ChangedAt:  time.Now(),
		ChangedBy:  userID,
		Reason:     reason,
	}
	current.Status = status
	current.StatusChangedAt = time.Now()
	current.UpdatedAt = time.Now()
	current.StatusHistory = append(current.StatusHistory, change)
	if err := s.fraudRepo.UpdateAccountFraudStatus(ctx, current); err != nil {
		return err
	}
	// Publish status change event
	if s.messagingClient != nil {
		_ = s.messagingClient.PublishEvent(ctx, &messaging.Event{
			Type:   "account.fraud_status.changed",
			Source: "fraud-detection-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID.String(),
				"old_status":    string(change.FromStatus),
				"new_status":    string(change.ToStatus),
				"reason":        reason,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
	return nil
}

func (s *FraudDetectionService) AddAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restriction *models.AccountRestriction) error {
	if s.fraudRepo == nil {
		return nil
	}
	if err := s.fraudRepo.AddAccountRestriction(ctx, enterpriseID, restriction); err != nil {
		return err
	}
	// Publish restriction added event
	if s.messagingClient != nil {
		_ = s.messagingClient.PublishEvent(ctx, &messaging.Event{
			Type:   "account.restriction.added",
			Source: "fraud-detection-service",
			Data: map[string]interface{}{
				"enterprise_id": enterpriseID.String(),
				"restriction":   restriction.Type,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
	return nil
}

func (s *FraudDetectionService) RemoveAccountRestriction(ctx context.Context, enterpriseID uuid.UUID, restrictionType models.RestrictionType) error {
	if s.fraudRepo == nil {
		return nil
	}
	return s.fraudRepo.RemoveAccountRestriction(ctx, enterpriseID, restrictionType)
}

func (s *FraudDetectionService) GenerateFraudReport(ctx context.Context, req *FraudReportRequest) (*FraudReport, error) {
	// In a real implementation, this would generate a fraud report
	return &FraudReport{}, nil
}

func (s *FraudDetectionService) GetFraudMetrics(ctx context.Context, enterpriseID *uuid.UUID) (*FraudMetrics, error) {
	// In a real implementation, this would retrieve fraud metrics
	return &FraudMetrics{}, nil
}
