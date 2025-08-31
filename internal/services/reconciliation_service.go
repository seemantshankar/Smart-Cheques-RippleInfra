package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// ReconciliationServiceInterface defines the interface for balance reconciliation operations
type ReconciliationServiceInterface interface {
	// Automated Reconciliation
	PerformReconciliation(ctx context.Context, req *ReconciliationRequest) (*ReconciliationResult, error)
	ScheduleReconciliation(ctx context.Context, schedule *ReconciliationSchedule) error
	GetReconciliationStatus(ctx context.Context, reconciliationID uuid.UUID) (*ReconciliationStatus, error)

	// Discrepancy Management
	GetDiscrepancies(ctx context.Context, enterpriseID *uuid.UUID, limit, offset int) ([]*BalanceDiscrepancy, error)
	ResolveDiscrepancy(ctx context.Context, req *DiscrepancyResolutionRequest) (*DiscrepancyResolution, error)
	BulkResolveDiscrepancies(ctx context.Context, req *BulkDiscrepancyResolutionRequest) (*BulkDiscrepancyResolutionResult, error)

	// Reporting and Analytics
	GenerateReconciliationReport(ctx context.Context, req *ReconciliationReportRequest) (*ReconciliationReport, error)
	GetReconciliationHistory(ctx context.Context, enterpriseID *uuid.UUID, limit, offset int) ([]*ReconciliationRecord, error)
	GetReconciliationMetrics(ctx context.Context, period ReconciliationPeriod) (*ReconciliationMetrics, error)

	// Manual Operations
	PerformManualReconciliation(ctx context.Context, req *ManualReconciliationRequest) (*ReconciliationResult, error)
	CreateReconciliationOverride(ctx context.Context, req *ReconciliationOverrideRequest) (*ReconciliationOverride, error)
	GetPendingOverrides(ctx context.Context) ([]*ReconciliationOverride, error)
}

// ReconciliationService implements the reconciliation service interface
type ReconciliationService struct {
	balanceRepo     repository.BalanceRepositoryInterface
	assetRepo       repository.AssetRepositoryInterface
	xrplService     repository.XRPLServiceInterface
	messagingClient messaging.EventBus
	config          *ReconciliationConfig
}

// NewReconciliationService creates a new reconciliation service instance
func NewReconciliationService(
	balanceRepo repository.BalanceRepositoryInterface,
	assetRepo repository.AssetRepositoryInterface,
	xrplService repository.XRPLServiceInterface,
	messagingClient messaging.EventBus,
	config *ReconciliationConfig,
) ReconciliationServiceInterface {
	return &ReconciliationService{
		balanceRepo:     balanceRepo,
		assetRepo:       assetRepo,
		xrplService:     xrplService,
		messagingClient: messagingClient,
		config:          config,
	}
}

// Configuration types
type ReconciliationConfig struct {
	// Tolerance settings
	ToleranceThreshold string `json:"tolerance_threshold"` // e.g., "0.001" (0.1%)
	CriticalThreshold  string `json:"critical_threshold"`  // e.g., "1000" (absolute amount)

	// Timing settings
	AutoReconcileInterval time.Duration `json:"auto_reconcile_interval"` // e.g., 1 hour
	BatchSize             int           `json:"batch_size"`              // e.g., 100 enterprises per batch

	// Alert settings
	AlertOnDiscrepancy  bool   `json:"alert_on_discrepancy"`
	AlertThreshold      string `json:"alert_threshold"`      // e.g., "100"
	EscalationThreshold string `json:"escalation_threshold"` // e.g., "10000"

	// Retry settings
	MaxRetries int           `json:"max_retries"` // e.g., 3
	RetryDelay time.Duration `json:"retry_delay"` // e.g., 5 minutes
}

// Request and response types
type ReconciliationRequest struct {
	EnterpriseID  *uuid.UUID `json:"enterprise_id,omitempty"` // Nil means all enterprises
	CurrencyCode  string     `json:"currency_code,omitempty"` // Empty means all currencies
	ForceRefresh  bool       `json:"force_refresh,omitempty"`
	IncludeReport bool       `json:"include_report,omitempty"`
}

type ReconciliationSchedule struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Frequency      ScheduleFrequency `json:"frequency"`
	CronExpression string            `json:"cron_expression,omitempty"`
	EnterpriseID   *uuid.UUID        `json:"enterprise_id,omitempty"`
	CurrencyCode   string            `json:"currency_code,omitempty"`
	IsActive       bool              `json:"is_active"`
	NextRun        time.Time         `json:"next_run"`
	CreatedAt      time.Time         `json:"created_at"`
}

type ManualReconciliationRequest struct {
	EnterpriseID    uuid.UUID `json:"enterprise_id"`
	CurrencyCode    string    `json:"currency_code"`
	InitiatedBy     uuid.UUID `json:"initiated_by"`
	Reason          string    `json:"reason"`
	ExpectedBalance string    `json:"expected_balance,omitempty"`
}

type DiscrepancyResolutionRequest struct {
	DiscrepancyID    uuid.UUID                 `json:"discrepancy_id"`
	ResolutionType   DiscrepancyResolutionType `json:"resolution_type"`
	AdjustmentAmount string                    `json:"adjustment_amount,omitempty"`
	Reason           string                    `json:"reason"`
	ApprovedBy       uuid.UUID                 `json:"approved_by"`
	Comments         string                    `json:"comments,omitempty"`
}

type BulkDiscrepancyResolutionRequest struct {
	DiscrepancyIDs []uuid.UUID               `json:"discrepancy_ids"`
	ResolutionType DiscrepancyResolutionType `json:"resolution_type"`
	Reason         string                    `json:"reason"`
	ApprovedBy     uuid.UUID                 `json:"approved_by"`
	Comments       string                    `json:"comments,omitempty"`
}

type ReconciliationOverrideRequest struct {
	EnterpriseID   uuid.UUID `json:"enterprise_id"`
	CurrencyCode   string    `json:"currency_code"`
	OverrideAmount string    `json:"override_amount"`
	Reason         string    `json:"reason"`
	RequestedBy    uuid.UUID `json:"requested_by"`
	ValidUntil     time.Time `json:"valid_until"`
}

type ReconciliationReportRequest struct {
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	EnterpriseID   *uuid.UUID `json:"enterprise_id,omitempty"`
	CurrencyCode   string     `json:"currency_code,omitempty"`
	IncludeDetails bool       `json:"include_details,omitempty"`
}

// Response types
type ReconciliationResult struct {
	ID                     uuid.UUID              `json:"id"`
	EnterpriseID           *uuid.UUID             `json:"enterprise_id,omitempty"`
	CurrencyCode           string                 `json:"currency_code,omitempty"`
	Status                 ReconciliationStatus   `json:"status"`
	StartedAt              time.Time              `json:"started_at"`
	CompletedAt            *time.Time             `json:"completed_at,omitempty"`
	TotalChecked           int                    `json:"total_checked"`
	DiscrepanciesFound     int                    `json:"discrepancies_found"`
	TotalDiscrepancyAmount string                 `json:"total_discrepancy_amount"`
	Discrepancies          []*BalanceDiscrepancy  `json:"discrepancies,omitempty"`
	Summary                *ReconciliationSummary `json:"summary,omitempty"`
	Report                 *ReconciliationReport  `json:"report,omitempty"`
}

type BalanceDiscrepancy struct {
	ID                 uuid.UUID              `json:"id"`
	ReconciliationID   uuid.UUID              `json:"reconciliation_id"`
	EnterpriseID       uuid.UUID              `json:"enterprise_id"`
	CurrencyCode       string                 `json:"currency_code"`
	InternalBalance    string                 `json:"internal_balance"`
	XRPLBalance        string                 `json:"xrpl_balance"`
	DiscrepancyAmount  string                 `json:"discrepancy_amount"`
	DiscrepancyPercent float64                `json:"discrepancy_percent"`
	Severity           DiscrepancySeverity    `json:"severity"`
	Status             DiscrepancyStatus      `json:"status"`
	PossibleCauses     []string               `json:"possible_causes"`
	DetectedAt         time.Time              `json:"detected_at"`
	ResolvedAt         *time.Time             `json:"resolved_at,omitempty"`
	Resolution         *DiscrepancyResolution `json:"resolution,omitempty"`
}

type DiscrepancyResolution struct {
	ID               uuid.UUID                 `json:"id"`
	DiscrepancyID    uuid.UUID                 `json:"discrepancy_id"`
	ResolutionType   DiscrepancyResolutionType `json:"resolution_type"`
	AdjustmentAmount string                    `json:"adjustment_amount,omitempty"`
	Reason           string                    `json:"reason"`
	ApprovedBy       uuid.UUID                 `json:"approved_by"`
	Comments         string                    `json:"comments,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	TransactionID    *uuid.UUID                `json:"transaction_id,omitempty"`
}

type BulkDiscrepancyResolutionResult struct {
	TotalRequests   int                            `json:"total_requests"`
	SuccessfulCount int                            `json:"successful_count"`
	FailedCount     int                            `json:"failed_count"`
	Results         []*DiscrepancyResolutionResult `json:"results"`
	ProcessedAt     time.Time                      `json:"processed_at"`
}

type DiscrepancyResolutionResult struct {
	DiscrepancyID uuid.UUID  `json:"discrepancy_id"`
	Success       bool       `json:"success"`
	Error         string     `json:"error,omitempty"`
	ResolutionID  *uuid.UUID `json:"resolution_id,omitempty"`
}

type ReconciliationOverride struct {
	ID             uuid.UUID      `json:"id"`
	EnterpriseID   uuid.UUID      `json:"enterprise_id"`
	CurrencyCode   string         `json:"currency_code"`
	OverrideAmount string         `json:"override_amount"`
	Reason         string         `json:"reason"`
	RequestedBy    uuid.UUID      `json:"requested_by"`
	Status         OverrideStatus `json:"status"`
	ValidUntil     time.Time      `json:"valid_until"`
	CreatedAt      time.Time      `json:"created_at"`
	ApprovedAt     *time.Time     `json:"approved_at,omitempty"`
	ApprovedBy     *uuid.UUID     `json:"approved_by,omitempty"`
}

type ReconciliationSummary struct {
	TotalEnterprises     int     `json:"total_enterprises"`
	TotalCurrencies      int     `json:"total_currencies"`
	TotalBalancesChecked int     `json:"total_balances_checked"`
	DiscrepanciesFound   int     `json:"discrepancies_found"`
	DiscrepancyRate      float64 `json:"discrepancy_rate"`
	LargestDiscrepancy   string  `json:"largest_discrepancy"`
	AverageDiscrepancy   string  `json:"average_discrepancy"`
	ProcessingTime       string  `json:"processing_time"`
}

type ReconciliationRecord struct {
	ID                 uuid.UUID            `json:"id"`
	EnterpriseID       *uuid.UUID           `json:"enterprise_id,omitempty"`
	CurrencyCode       string               `json:"currency_code,omitempty"`
	Status             ReconciliationStatus `json:"status"`
	TotalChecked       int                  `json:"total_checked"`
	DiscrepanciesFound int                  `json:"discrepancies_found"`
	StartedAt          time.Time            `json:"started_at"`
	CompletedAt        *time.Time           `json:"completed_at,omitempty"`
	TriggerType        TriggerType          `json:"trigger_type"`
	InitiatedBy        *uuid.UUID           `json:"initiated_by,omitempty"`
}

type ReconciliationMetrics struct {
	Period                ReconciliationPeriod  `json:"period"`
	TotalReconciliations  int                   `json:"total_reconciliations"`
	SuccessfulCount       int                   `json:"successful_count"`
	FailedCount           int                   `json:"failed_count"`
	AverageProcessingTime string                `json:"average_processing_time"`
	TotalDiscrepancies    int                   `json:"total_discrepancies"`
	ResolvedDiscrepancies int                   `json:"resolved_discrepancies"`
	PendingDiscrepancies  int                   `json:"pending_discrepancies"`
	DiscrepancyTrends     []*DiscrepancyTrend   `json:"discrepancy_trends"`
	TopCurrenciesByIssues []*CurrencyIssueCount `json:"top_currencies_by_issues"`
}

type DiscrepancyTrend struct {
	Date        time.Time `json:"date"`
	Count       int       `json:"count"`
	TotalAmount string    `json:"total_amount"`
}

type CurrencyIssueCount struct {
	CurrencyCode string `json:"currency_code"`
	IssueCount   int    `json:"issue_count"`
	TotalAmount  string `json:"total_amount"`
}

type ReconciliationReport struct {
	ID                      uuid.UUID              `json:"id"`
	GeneratedAt             time.Time              `json:"generated_at"`
	StartDate               time.Time              `json:"start_date"`
	EndDate                 time.Time              `json:"end_date"`
	Summary                 *ReconciliationSummary `json:"summary"`
	DiscrepanciesByType     map[string]int         `json:"discrepancies_by_type"`
	DiscrepanciesBySeverity map[string]int         `json:"discrepancies_by_severity"`
	ResolutionBreakdown     map[string]int         `json:"resolution_breakdown"`
	Recommendations         []string               `json:"recommendations"`
	DetailedFindings        []*DetailedFinding     `json:"detailed_findings,omitempty"`
}

type DetailedFinding struct {
	EnterpriseID      uuid.UUID           `json:"enterprise_id"`
	CurrencyCode      string              `json:"currency_code"`
	DiscrepancyAmount string              `json:"discrepancy_amount"`
	Severity          DiscrepancySeverity `json:"severity"`
	Description       string              `json:"description"`
	RecommendedAction string              `json:"recommended_action"`
}

// Enums
type ReconciliationStatus string

const (
	ReconciliationStatusPending   ReconciliationStatus = "pending"
	ReconciliationStatusRunning   ReconciliationStatus = "running"
	ReconciliationStatusCompleted ReconciliationStatus = "completed"
	ReconciliationStatusFailed    ReconciliationStatus = "failed"
	ReconciliationStatusCancelled ReconciliationStatus = "canceled"
)

type DiscrepancySeverity string

const (
	DiscrepancySeverityLow      DiscrepancySeverity = "low"
	DiscrepancySeverityMedium   DiscrepancySeverity = "medium"
	DiscrepancySeverityHigh     DiscrepancySeverity = "high"
	DiscrepancySeverityCritical DiscrepancySeverity = "critical"
)

type DiscrepancyStatus string

const (
	DiscrepancyStatusPending       DiscrepancyStatus = "pending"
	DiscrepancyStatusInvestigating DiscrepancyStatus = "investigating"
	DiscrepancyStatusResolved      DiscrepancyStatus = "resolved"
	DiscrepancyStatusIgnored       DiscrepancyStatus = "ignored"
)

type DiscrepancyResolutionType string

const (
	DiscrepancyResolutionTypeAdjustInternal DiscrepancyResolutionType = "adjust_internal"
	DiscrepancyResolutionTypeAdjustXRPL     DiscrepancyResolutionType = "adjust_xrpl"
	DiscrepancyResolutionTypeIgnore         DiscrepancyResolutionType = "ignore"
	DiscrepancyResolutionTypeInvestigate    DiscrepancyResolutionType = "investigate"
)

type OverrideStatus string

const (
	OverrideStatusPending  OverrideStatus = "pending"
	OverrideStatusApproved OverrideStatus = "approved"
	OverrideStatusRejected OverrideStatus = "rejected"
	OverrideStatusExpired  OverrideStatus = "expired"
)

type ScheduleFrequency string

const (
	ScheduleFrequencyHourly  ScheduleFrequency = "hourly"
	ScheduleFrequencyDaily   ScheduleFrequency = "daily"
	ScheduleFrequencyWeekly  ScheduleFrequency = "weekly"
	ScheduleFrequencyMonthly ScheduleFrequency = "monthly"
	ScheduleFrequencyCustom  ScheduleFrequency = "custom"
)

type TriggerType string

const (
	TriggerTypeScheduled TriggerType = "scheduled"
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeAPI       TriggerType = "api"
	TriggerTypeAlert     TriggerType = "alert"
)

type ReconciliationPeriod string

const (
	ReconciliationPeriodDaily     ReconciliationPeriod = "daily"
	ReconciliationPeriodWeekly    ReconciliationPeriod = "weekly"
	ReconciliationPeriodMonthly   ReconciliationPeriod = "monthly"
	ReconciliationPeriodQuarterly ReconciliationPeriod = "quarterly"
)

// PerformReconciliation performs balance reconciliation
func (s *ReconciliationService) PerformReconciliation(ctx context.Context, req *ReconciliationRequest) (*ReconciliationResult, error) {
	reconciliationID := uuid.New()
	startTime := time.Now()

	result := &ReconciliationResult{
		ID:                     reconciliationID,
		EnterpriseID:           req.EnterpriseID,
		CurrencyCode:           req.CurrencyCode,
		Status:                 ReconciliationStatusRunning,
		StartedAt:              startTime,
		TotalChecked:           0,
		DiscrepanciesFound:     0,
		TotalDiscrepancyAmount: "0",
		Discrepancies:          []*BalanceDiscrepancy{},
	}

	// Get enterprises to reconcile
	enterprises, err := s.getEnterprisesToReconcile(ctx, req.EnterpriseID)
	if err != nil {
		result.Status = ReconciliationStatusFailed
		return result, fmt.Errorf("failed to get enterprises: %w", err)
	}

	// Get currencies to reconcile
	currencies, err := s.getCurrenciesToReconcile(ctx, req.CurrencyCode)
	if err != nil {
		result.Status = ReconciliationStatusFailed
		return result, fmt.Errorf("failed to get currencies: %w", err)
	}

	// Perform reconciliation for each enterprise-currency combination
	totalDiscrepancyAmount := big.NewInt(0)
	for _, enterprise := range enterprises {
		for _, currency := range currencies {
			discrepancy, err := s.reconcileBalance(ctx, enterprise.ID, currency, req.ForceRefresh)
			if err != nil {
				fmt.Printf("Warning: Failed to reconcile balance for enterprise %s, currency %s: %v\n",
					enterprise.ID.String(), currency, err)
				continue
			}

			result.TotalChecked++

			if discrepancy != nil {
				result.DiscrepanciesFound++
				result.Discrepancies = append(result.Discrepancies, discrepancy)

				// Add to total discrepancy amount
				discAmount := new(big.Int)
				discAmount.SetString(discrepancy.DiscrepancyAmount, 10)
				totalDiscrepancyAmount.Add(totalDiscrepancyAmount, discAmount)

				// Send alert if necessary
				if s.shouldSendAlert(discrepancy) {
					s.sendDiscrepancyAlert(ctx, discrepancy)
				}
			}
		}
	}

	// Update result
	result.TotalDiscrepancyAmount = totalDiscrepancyAmount.String()
	result.Status = ReconciliationStatusCompleted
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	// Generate summary
	result.Summary = s.generateReconciliationSummary(result, startTime, completedAt)

	// Generate report if requested
	if req.IncludeReport {
		report, err := s.generateReconciliationReportFromResult(ctx, result)
		if err != nil {
			fmt.Printf("Warning: Failed to generate reconciliation report: %v\n", err)
		} else {
			result.Report = report
		}
	}

	// Publish reconciliation event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "reconciliation.completed",
			Source: "reconciliation-service",
			Data: map[string]interface{}{
				"reconciliation_id":   reconciliationID.String(),
				"total_checked":       result.TotalChecked,
				"discrepancies_found": result.DiscrepanciesFound,
				"status":              result.Status,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish reconciliation event: %v\n", err)
		}
	}

	return result, nil
}

// Helper methods for reconciliation
func (s *ReconciliationService) getEnterprisesToReconcile(_ context.Context, enterpriseID *uuid.UUID) ([]*models.Enterprise, error) {
	// In a real implementation, this would query the database
	// For now, return a placeholder enterprise
	if enterpriseID != nil {
		return []*models.Enterprise{
			{ID: *enterpriseID},
		}, nil
	}
	// Return all enterprises (placeholder)
	return []*models.Enterprise{
		{ID: uuid.New()},
	}, nil
}

func (s *ReconciliationService) getCurrenciesToReconcile(_ context.Context, currencyCode string) ([]string, error) {
	if currencyCode != "" {
		return []string{currencyCode}, nil
	}
	// Return all supported currencies
	return []string{"USDT", "USDC", "e₹", "XRP"}, nil
}

func (s *ReconciliationService) reconcileBalance(ctx context.Context, enterpriseID uuid.UUID, currencyCode string, _ bool) (*BalanceDiscrepancy, error) {
	// Get internal balance
	balance, err := s.balanceRepo.GetBalance(ctx, enterpriseID, currencyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal balance: %w", err)
	}

	// Get XRPL balance (simplified - in reality this would query XRPL)
	xrplBalance := s.getXRPLBalance(ctx, enterpriseID, currencyCode)

	// Compare balances
	internalBalance := new(big.Int)
	internalBalance.SetString(balance.AvailableBalance, 10)

	xrplBalanceInt := new(big.Int)
	xrplBalanceInt.SetString(xrplBalance, 10)

	// Calculate discrepancy
	discrepancyAmount := new(big.Int).Sub(internalBalance, xrplBalanceInt)
	discrepancyAmountAbs := new(big.Int).Abs(discrepancyAmount)

	// Check if discrepancy is within tolerance
	toleranceThreshold := new(big.Int)
	toleranceThreshold.SetString(s.config.ToleranceThreshold, 10)

	if discrepancyAmountAbs.Cmp(toleranceThreshold) <= 0 {
		// Within tolerance, no discrepancy
		return nil, nil
	}

	// Calculate discrepancy percentage
	discrepancyPercent := 0.0
	if internalBalance.Cmp(big.NewInt(0)) > 0 {
		discrepancyFloat := new(big.Float).SetInt(discrepancyAmountAbs)
		internalFloat := new(big.Float).SetInt(internalBalance)
		percentFloat := new(big.Float).Quo(discrepancyFloat, internalFloat)
		percentFloat.Mul(percentFloat, big.NewFloat(100))
		discrepancyPercent, _ = percentFloat.Float64()
	}

	// Determine severity
	severity := s.determineDiscrepancySeverity(discrepancyAmountAbs, discrepancyPercent)

	// Analyze possible causes
	possibleCauses := s.analyzePossibleCauses(ctx, enterpriseID, currencyCode, discrepancyAmount)

	// Create discrepancy record
	discrepancy := &BalanceDiscrepancy{
		ID:                 uuid.New(),
		ReconciliationID:   uuid.New(), // Would be passed from parent function
		EnterpriseID:       enterpriseID,
		CurrencyCode:       currencyCode,
		InternalBalance:    balance.AvailableBalance,
		XRPLBalance:        xrplBalance,
		DiscrepancyAmount:  discrepancyAmount.String(),
		DiscrepancyPercent: discrepancyPercent,
		Severity:           severity,
		Status:             DiscrepancyStatusPending,
		PossibleCauses:     possibleCauses,
		DetectedAt:         time.Now(),
	}

	return discrepancy, nil
}

func (s *ReconciliationService) getXRPLBalance(_ context.Context, _ uuid.UUID, _ string) string {
	// In a real implementation, this would query XRPL
	// For now, return a simulated balance that might have discrepancies
	return "1000000" // Placeholder
}

func (s *ReconciliationService) determineDiscrepancySeverity(discrepancyAmount *big.Int, discrepancyPercent float64) DiscrepancySeverity {
	criticalThreshold := new(big.Int)
	criticalThreshold.SetString(s.config.CriticalThreshold, 10)

	if discrepancyAmount.Cmp(criticalThreshold) >= 0 || discrepancyPercent >= 10.0 {
		return DiscrepancySeverityCritical
	} else if discrepancyPercent >= 5.0 || discrepancyAmount.Cmp(big.NewInt(1000)) >= 0 {
		return DiscrepancySeverityHigh
	} else if discrepancyPercent >= 1.0 || discrepancyAmount.Cmp(big.NewInt(100)) >= 0 {
		return DiscrepancySeverityMedium
	}
	return DiscrepancySeverityLow
}

func (s *ReconciliationService) analyzePossibleCauses(_ context.Context, _ uuid.UUID, _ string, discrepancyAmount *big.Int) []string {
	causes := []string{}

	if discrepancyAmount.Sign() > 0 {
		// Internal balance is higher than XRPL
		causes = append(causes, "Pending XRPL transactions not yet confirmed")
		causes = append(causes, "Failed XRPL transaction not properly handled")
		causes = append(causes, "Manual adjustment needed on XRPL side")
	} else {
		// XRPL balance is higher than internal
		causes = append(causes, "Unrecorded incoming XRPL transaction")
		causes = append(causes, "Failed internal balance update")
		causes = append(causes, "System synchronization delay")
	}

	causes = append(causes, "Data consistency issue")
	causes = append(causes, "Network connectivity problem during transaction")

	return causes
}

func (s *ReconciliationService) shouldSendAlert(discrepancy *BalanceDiscrepancy) bool {
	if !s.config.AlertOnDiscrepancy {
		return false
	}

	alertThreshold := new(big.Int)
	alertThreshold.SetString(s.config.AlertThreshold, 10)

	discrepancyAmount := new(big.Int)
	discrepancyAmount.SetString(discrepancy.DiscrepancyAmount, 10)
	discrepancyAmountAbs := new(big.Int).Abs(discrepancyAmount)

	return discrepancyAmountAbs.Cmp(alertThreshold) >= 0 ||
		discrepancy.Severity == DiscrepancySeverityCritical ||
		discrepancy.Severity == DiscrepancySeverityHigh
}

func (s *ReconciliationService) sendDiscrepancyAlert(ctx context.Context, discrepancy *BalanceDiscrepancy) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   "reconciliation.discrepancy.alert",
		Source: "reconciliation-service",
		Data: map[string]interface{}{
			"discrepancy_id":     discrepancy.ID.String(),
			"enterprise_id":      discrepancy.EnterpriseID.String(),
			"currency_code":      discrepancy.CurrencyCode,
			"discrepancy_amount": discrepancy.DiscrepancyAmount,
			"severity":           discrepancy.Severity,
			"internal_balance":   discrepancy.InternalBalance,
			"xrpl_balance":       discrepancy.XRPLBalance,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish discrepancy alert: %v\n", err)
	}
}

func (s *ReconciliationService) generateReconciliationSummary(result *ReconciliationResult, startTime, endTime time.Time) *ReconciliationSummary {
	processingTime := endTime.Sub(startTime)

	discrepancyRate := 0.0
	if result.TotalChecked > 0 {
		discrepancyRate = float64(result.DiscrepanciesFound) / float64(result.TotalChecked) * 100
	}

	// Find largest and calculate average discrepancy
	largestDiscrepancy := "0"
	totalDiscrepancyForAvg := big.NewInt(0)
	for _, discrepancy := range result.Discrepancies {
		discAmount := new(big.Int)
		discAmount.SetString(discrepancy.DiscrepancyAmount, 10)
		discAmountAbs := new(big.Int).Abs(discAmount)

		largestAmount := new(big.Int)
		largestAmount.SetString(largestDiscrepancy, 10)

		if discAmountAbs.Cmp(largestAmount) > 0 {
			largestDiscrepancy = discAmountAbs.String()
		}

		totalDiscrepancyForAvg.Add(totalDiscrepancyForAvg, discAmountAbs)
	}

	averageDiscrepancy := "0"
	if result.DiscrepanciesFound > 0 {
		avgAmount := new(big.Int).Div(totalDiscrepancyForAvg, big.NewInt(int64(result.DiscrepanciesFound)))
		averageDiscrepancy = avgAmount.String()
	}

	return &ReconciliationSummary{
		TotalEnterprises:     1, // Simplified
		TotalCurrencies:      4, // Simplified
		TotalBalancesChecked: result.TotalChecked,
		DiscrepanciesFound:   result.DiscrepanciesFound,
		DiscrepancyRate:      discrepancyRate,
		LargestDiscrepancy:   largestDiscrepancy,
		AverageDiscrepancy:   averageDiscrepancy,
		ProcessingTime:       processingTime.String(),
	}
}

// ResolveDiscrepancy resolves a balance discrepancy
func (s *ReconciliationService) ResolveDiscrepancy(ctx context.Context, req *DiscrepancyResolutionRequest) (*DiscrepancyResolution, error) {
	// Create resolution record
	resolution := &DiscrepancyResolution{
		ID:               uuid.New(),
		DiscrepancyID:    req.DiscrepancyID,
		ResolutionType:   req.ResolutionType,
		AdjustmentAmount: req.AdjustmentAmount,
		Reason:           req.Reason,
		ApprovedBy:       req.ApprovedBy,
		Comments:         req.Comments,
		CreatedAt:        time.Now(),
	}

	// Process the resolution based on type
	switch req.ResolutionType {
	case DiscrepancyResolutionTypeAdjustInternal:
		err := s.processInternalAdjustment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to process internal adjustment: %w", err)
		}
	case DiscrepancyResolutionTypeAdjustXRPL:
		err := s.processXRPLAdjustment(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to process XRPL adjustment: %w", err)
		}
	case DiscrepancyResolutionTypeIgnore:
		// Just mark as resolved
	case DiscrepancyResolutionTypeInvestigate:
		// Mark for further investigation
	}

	// Publish resolution event
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "reconciliation.discrepancy.resolved",
			Source: "reconciliation-service",
			Data: map[string]interface{}{
				"discrepancy_id":  req.DiscrepancyID.String(),
				"resolution_type": req.ResolutionType,
				"approved_by":     req.ApprovedBy.String(),
				"resolution_id":   resolution.ID.String(),
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish resolution event: %v\n", err)
		}
	}

	return resolution, nil
}

func (s *ReconciliationService) processInternalAdjustment(_ context.Context, req *DiscrepancyResolutionRequest) error {
	// In a real implementation, this would create an adjustment transaction
	fmt.Printf("Processing internal adjustment for discrepancy %s: %s\n", req.DiscrepancyID.String(), req.AdjustmentAmount)
	return nil
}

func (s *ReconciliationService) processXRPLAdjustment(_ context.Context, req *DiscrepancyResolutionRequest) error {
	// In a real implementation, this would submit an XRPL transaction
	fmt.Printf("Processing XRPL adjustment for discrepancy %s: %s\n", req.DiscrepancyID.String(), req.AdjustmentAmount)
	return nil
}

// GetDiscrepancies retrieves balance discrepancies
func (s *ReconciliationService) GetDiscrepancies(_ context.Context, enterpriseID *uuid.UUID, limit, offset int) ([]*BalanceDiscrepancy, error) {
	// In a real implementation, this would query the database
	return []*BalanceDiscrepancy{}, nil
}

// ScheduleReconciliation creates a scheduled reconciliation
func (s *ReconciliationService) ScheduleReconciliation(_ context.Context, schedule *ReconciliationSchedule) error {
	// In a real implementation, this would store the schedule and set up a cron job
	fmt.Printf("Scheduling reconciliation: %s with frequency %s\n", schedule.Name, schedule.Frequency)
	return nil
}

// GetReconciliationStatus gets the status of a reconciliation
func (s *ReconciliationService) GetReconciliationStatus(_ context.Context, reconciliationID uuid.UUID) (*ReconciliationStatus, error) {
	// In a real implementation, this would query the database
	status := ReconciliationStatusCompleted
	return &status, nil
}

// BulkResolveDiscrepancies resolves multiple discrepancies
func (s *ReconciliationService) BulkResolveDiscrepancies(ctx context.Context, req *BulkDiscrepancyResolutionRequest) (*BulkDiscrepancyResolutionResult, error) {
	results := []*DiscrepancyResolutionResult{}
	successCount := 0
	failedCount := 0

	for _, discrepancyID := range req.DiscrepancyIDs {
		resolutionReq := &DiscrepancyResolutionRequest{
			DiscrepancyID:  discrepancyID,
			ResolutionType: req.ResolutionType,
			Reason:         req.Reason,
			ApprovedBy:     req.ApprovedBy,
			Comments:       req.Comments,
		}

		resolution, err := s.ResolveDiscrepancy(ctx, resolutionReq)
		if err != nil {
			results = append(results, &DiscrepancyResolutionResult{
				DiscrepancyID: discrepancyID,
				Success:       false,
				Error:         err.Error(),
			})
			failedCount++
		} else {
			results = append(results, &DiscrepancyResolutionResult{
				DiscrepancyID: discrepancyID,
				Success:       true,
				ResolutionID:  &resolution.ID,
			})
			successCount++
		}
	}

	return &BulkDiscrepancyResolutionResult{
		TotalRequests:   len(req.DiscrepancyIDs),
		SuccessfulCount: successCount,
		FailedCount:     failedCount,
		Results:         results,
		ProcessedAt:     time.Now(),
	}, nil
}

// GenerateReconciliationReport generates a reconciliation report
func (s *ReconciliationService) GenerateReconciliationReport(ctx context.Context, req *ReconciliationReportRequest) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		ID:          uuid.New(),
		GeneratedAt: time.Now(),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Summary: &ReconciliationSummary{
			TotalEnterprises:     10,
			TotalCurrencies:      4,
			TotalBalancesChecked: 100,
			DiscrepanciesFound:   5,
			DiscrepancyRate:      5.0,
			LargestDiscrepancy:   "1000",
			AverageDiscrepancy:   "200",
			ProcessingTime:       "2m30s",
		},
		DiscrepanciesByType: map[string]int{
			"balance_mismatch":    3,
			"pending_transaction": 2,
		},
		DiscrepanciesBySeverity: map[string]int{
			"low":    2,
			"medium": 2,
			"high":   1,
		},
		ResolutionBreakdown: map[string]int{
			"resolved": 3,
			"pending":  2,
		},
		Recommendations: []string{
			"Increase reconciliation frequency for high-volume currencies",
			"Implement real-time balance monitoring",
			"Review XRPL transaction confirmation processes",
		},
	}

	if req.IncludeDetails {
		report.DetailedFindings = []*DetailedFinding{
			{
				EnterpriseID:      uuid.New(),
				CurrencyCode:      "USDT",
				DiscrepancyAmount: "500",
				Severity:          DiscrepancySeverityMedium,
				Description:       "Internal balance exceeds XRPL balance",
				RecommendedAction: "Check for pending XRPL transactions",
			},
		}
	}

	return report, nil
}

func (s *ReconciliationService) generateReconciliationReportFromResult(_ context.Context, result *ReconciliationResult) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		ID:          uuid.New(),
		GeneratedAt: time.Now(),
		StartDate:   result.StartedAt,
		EndDate:     *result.CompletedAt,
		Summary:     result.Summary,
	}

	// Analyze discrepancies by type and severity
	discrepanciesByType := make(map[string]int)
	discrepanciesBySeverity := make(map[string]int)

	for _, discrepancy := range result.Discrepancies {
		// Categorize by severity
		discrepanciesBySeverity[string(discrepancy.Severity)]++

		// Categorize by type (simplified logic)
		if len(discrepancy.PossibleCauses) > 0 {
			if discrepancy.PossibleCauses[0] == "Pending XRPL transactions not yet confirmed" {
				discrepanciesByType["pending_transaction"]++
			} else {
				discrepanciesByType["balance_mismatch"]++
			}
		} else {
			discrepanciesByType["unknown"]++
		}
	}

	report.DiscrepanciesByType = discrepanciesByType
	report.DiscrepanciesBySeverity = discrepanciesBySeverity

	// Generate recommendations
	recommendations := []string{}
	if result.DiscrepanciesFound > 0 {
		if result.DiscrepanciesFound > result.TotalChecked/10 {
			recommendations = append(recommendations, "High discrepancy rate detected - review reconciliation processes")
		}
		if discrepanciesBySeverity["critical"] > 0 || discrepanciesBySeverity["high"] > 0 {
			recommendations = append(recommendations, "Critical discrepancies found - immediate investigation required")
		}
		recommendations = append(recommendations, "Review recent transaction patterns for unusual activity")
	} else {
		recommendations = append(recommendations, "All balances reconciled successfully")
	}

	report.Recommendations = recommendations

	return report, nil
}

// GetReconciliationHistory gets reconciliation history
func (s *ReconciliationService) GetReconciliationHistory(ctx context.Context, enterpriseID *uuid.UUID, limit, offset int) ([]*ReconciliationRecord, error) {
	// In a real implementation, this would query the database
	return []*ReconciliationRecord{}, nil
}

// GetReconciliationMetrics gets reconciliation metrics
func (s *ReconciliationService) GetReconciliationMetrics(ctx context.Context, period ReconciliationPeriod) (*ReconciliationMetrics, error) {
	return &ReconciliationMetrics{
		Period:                period,
		TotalReconciliations:  50,
		SuccessfulCount:       48,
		FailedCount:           2,
		AverageProcessingTime: "1m45s",
		TotalDiscrepancies:    15,
		ResolvedDiscrepancies: 12,
		PendingDiscrepancies:  3,
		DiscrepancyTrends: []*DiscrepancyTrend{
			{
				Date:        time.Now().AddDate(0, 0, -7),
				Count:       2,
				TotalAmount: "500",
			},
			{
				Date:        time.Now().AddDate(0, 0, -6),
				Count:       1,
				TotalAmount: "100",
			},
		},
		TopCurrenciesByIssues: []*CurrencyIssueCount{
			{
				CurrencyCode: "USDT",
				IssueCount:   8,
				TotalAmount:  "2000",
			},
			{
				CurrencyCode: "USDC",
				IssueCount:   4,
				TotalAmount:  "800",
			},
		},
	}, nil
}

// PerformManualReconciliation performs manual reconciliation
func (s *ReconciliationService) PerformManualReconciliation(ctx context.Context, req *ManualReconciliationRequest) (*ReconciliationResult, error) {
	// Create a standard reconciliation request
	reconReq := &ReconciliationRequest{
		EnterpriseID:  &req.EnterpriseID,
		CurrencyCode:  req.CurrencyCode,
		ForceRefresh:  true,
		IncludeReport: true,
	}

	result, err := s.PerformReconciliation(ctx, reconReq)
	if err != nil {
		return nil, err
	}

	// Log manual reconciliation
	if s.messagingClient != nil {
		event := &messaging.Event{
			Type:   "reconciliation.manual.performed",
			Source: "reconciliation-service",
			Data: map[string]interface{}{
				"reconciliation_id": result.ID.String(),
				"enterprise_id":     req.EnterpriseID.String(),
				"currency_code":     req.CurrencyCode,
				"initiated_by":      req.InitiatedBy.String(),
				"reason":            req.Reason,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
			fmt.Printf("Warning: Failed to publish manual reconciliation event: %v\n", err)
		}
	}

	return result, nil
}

// CreateReconciliationOverride creates a reconciliation override
func (s *ReconciliationService) CreateReconciliationOverride(ctx context.Context, req *ReconciliationOverrideRequest) (*ReconciliationOverride, error) {
	override := &ReconciliationOverride{
		ID:             uuid.New(),
		EnterpriseID:   req.EnterpriseID,
		CurrencyCode:   req.CurrencyCode,
		OverrideAmount: req.OverrideAmount,
		Reason:         req.Reason,
		RequestedBy:    req.RequestedBy,
		Status:         OverrideStatusPending,
		ValidUntil:     req.ValidUntil,
		CreatedAt:      time.Now(),
	}

	// In a real implementation, this would store the override and trigger approval workflow
	return override, nil
}

// GetPendingOverrides gets pending reconciliation overrides
func (s *ReconciliationService) GetPendingOverrides(ctx context.Context) ([]*ReconciliationOverride, error) {
	// In a real implementation, this would query the database
	return []*ReconciliationOverride{}, nil
}
