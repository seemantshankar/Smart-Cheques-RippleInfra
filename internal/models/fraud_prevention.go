package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FraudAlert represents a fraud detection alert
type FraudAlert struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	EnterpriseID  uuid.UUID        `json:"enterprise_id" db:"enterprise_id"`
	TransactionID *string          `json:"transaction_id,omitempty" db:"transaction_id"`
	AlertType     FraudAlertType   `json:"alert_type" db:"alert_type"`
	Severity      FraudSeverity    `json:"severity" db:"severity"`
	Status        FraudAlertStatus `json:"status" db:"status"`
	RuleID        *uuid.UUID       `json:"rule_id,omitempty" db:"rule_id"`
	CaseID        *uuid.UUID       `json:"case_id,omitempty" db:"case_id"`

	// Detection details
	Score           float64                `json:"score" db:"score"`
	Confidence      float64                `json:"confidence" db:"confidence"`
	DetectionMethod string                 `json:"detection_method" db:"detection_method"`
	Evidence        map[string]interface{} `json:"evidence" db:"evidence"`

	// Alert details
	Title          string `json:"title" db:"title"`
	Description    string `json:"description" db:"description"`
	Recommendation string `json:"recommendation" db:"recommendation"`

	// Notification
	NotifiedAt           *time.Time `json:"notified_at,omitempty" db:"notified_at"`
	NotificationChannels []string   `json:"notification_channels" db:"notification_channels"`

	// Investigation
	AssignedTo         *uuid.UUID          `json:"assigned_to,omitempty" db:"assigned_to"`
	InvestigationNotes []InvestigationNote `json:"investigation_notes" db:"investigation_notes"`

	// Timestamps
	DetectedAt     time.Time  `json:"detected_at" db:"detected_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// FraudRule represents a configurable fraud detection rule
type FraudRule struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	Category    FraudRuleCategory `json:"category" db:"category"`
	RuleType    FraudRuleType     `json:"rule_type" db:"rule_type"`

	// Rule configuration
	Conditions map[string]interface{} `json:"conditions" db:"conditions"`
	Thresholds map[string]interface{} `json:"thresholds" db:"thresholds"`
	Actions    []FraudAction          `json:"actions" db:"actions"`

	// Scoring and severity
	BaseScore  float64       `json:"base_score" db:"base_score"`
	Severity   FraudSeverity `json:"severity" db:"severity"`
	Confidence float64       `json:"confidence" db:"confidence"`

	// Status and versioning
	Status      FraudRuleStatus `json:"status" db:"status"`
	Version     int             `json:"version" db:"version"`
	EffectiveAt time.Time       `json:"effective_at" db:"effective_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty" db:"expires_at"`

	// Metadata
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// FraudCase represents a fraud investigation case
type FraudCase struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	EnterpriseID uuid.UUID         `json:"enterprise_id" db:"enterprise_id"`
	CaseNumber   string            `json:"case_number" db:"case_number"`
	Status       FraudCaseStatus   `json:"status" db:"status"`
	Priority     FraudCasePriority `json:"priority" db:"priority"`

	// Case details
	Title       string            `json:"title" db:"title"`
	Description string            `json:"description" db:"description"`
	Category    FraudCaseCategory `json:"category" db:"category"`

	// Investigation
	AssignedTo         *uuid.UUID          `json:"assigned_to,omitempty" db:"assigned_to"`
	Investigator       *uuid.UUID          `json:"investigator,omitempty" db:"investigator"`
	InvestigationNotes []InvestigationNote `json:"investigation_notes" db:"investigation_notes"`

	// Related entities
	Alerts       []uuid.UUID `json:"alerts" db:"alerts"`
	Transactions []string    `json:"transactions" db:"transactions"`

	// Outcome
	Resolution *FraudCaseResolution `json:"resolution,omitempty" db:"resolution"`
	Outcome    FraudCaseOutcome     `json:"outcome" db:"outcome"`

	// Timestamps
	OpenedAt   time.Time  `json:"opened_at" db:"opened_at"`
	AssignedAt *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt   *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AccountFraudStatus represents enterprise fraud status management
type AccountFraudStatus struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	EnterpriseID uuid.UUID              `json:"enterprise_id" db:"enterprise_id"`
	Status       AccountFraudStatusType `json:"status" db:"status"`

	// Risk assessment
	RiskScore   float64        `json:"risk_score" db:"risk_score"`
	RiskLevel   FraudRiskLevel `json:"risk_level" db:"risk_level"`
	RiskFactors []string       `json:"risk_factors" db:"risk_factors"`

	// Restrictions
	Restrictions []AccountRestriction   `json:"restrictions" db:"restrictions"`
	Limits       map[string]interface{} `json:"limits" db:"limits"`

	// Monitoring
	MonitoringLevel MonitoringLevel `json:"monitoring_level" db:"monitoring_level"`
	ReviewFrequency time.Duration   `json:"review_frequency" db:"review_frequency"`
	NextReviewDate  time.Time       `json:"next_review_date" db:"next_review_date"`

	// History
	StatusHistory []FraudStatusChange `json:"status_history" db:"status_history"`

	// Timestamps
	StatusChangedAt time.Time `json:"status_changed_at" db:"status_changed_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Supporting types

// InvestigationNote represents a note in fraud investigation
type InvestigationNote struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AuthorID  uuid.UUID `json:"author_id" db:"author_id"`
	Content   string    `json:"content" db:"content"`
	NoteType  string    `json:"note_type" db:"note_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FraudCaseResolution represents the resolution of a fraud case
type FraudCaseResolution struct {
	ResolutionType string                 `json:"resolution_type" db:"resolution_type"`
	Description    string                 `json:"description" db:"description"`
	Actions        []string               `json:"actions" db:"actions"`
	Evidence       map[string]interface{} `json:"evidence" db:"evidence"`
	ResolvedBy     uuid.UUID              `json:"resolved_by" db:"resolved_by"`
	ResolvedAt     time.Time              `json:"resolved_at" db:"resolved_at"`
}

// FraudStatusChange represents a change in fraud status
type FraudStatusChange struct {
	FromStatus AccountFraudStatusType `json:"from_status" db:"from_status"`
	ToStatus   AccountFraudStatusType `json:"to_status" db:"to_status"`
	Reason     string                 `json:"reason" db:"reason"`
	ChangedBy  uuid.UUID              `json:"changed_by" db:"changed_by"`
	ChangedAt  time.Time              `json:"changed_at" db:"changed_at"`
}

// AccountRestriction represents restrictions on an account
type AccountRestriction struct {
	Type        RestrictionType        `json:"type" db:"type"`
	Description string                 `json:"description" db:"description"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	EffectiveAt time.Time              `json:"effective_at" db:"effective_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	AppliedBy   uuid.UUID              `json:"applied_by" db:"applied_by"`
}

// Enums

type FraudAlertType string

const (
	FraudAlertTypeTransactionAnomaly  FraudAlertType = "transaction_anomaly"
	FraudAlertTypeVelocitySpike       FraudAlertType = "velocity_spike"
	FraudAlertTypeAmountOutlier       FraudAlertType = "amount_outlier"
	FraudAlertTypePatternAnomaly      FraudAlertType = "pattern_anomaly"
	FraudAlertTypeBehavioralChange    FraudAlertType = "behavioral_change"
	FraudAlertTypeComplianceViolation FraudAlertType = "compliance_violation"
	FraudAlertTypeAccountTakeover     FraudAlertType = "account_takeover"
	FraudAlertTypeSuspiciousActivity  FraudAlertType = "suspicious_activity"
)

type FraudSeverity string

const (
	FraudSeverityLow      FraudSeverity = "low"
	FraudSeverityMedium   FraudSeverity = "medium"
	FraudSeverityHigh     FraudSeverity = "high"
	FraudSeverityCritical FraudSeverity = "critical"
)

type FraudAlertStatus string

const (
	FraudAlertStatusNew           FraudAlertStatus = "new"
	FraudAlertStatusAcknowledged  FraudAlertStatus = "acknowledged"
	FraudAlertStatusInvestigating FraudAlertStatus = "investigating"
	FraudAlertStatusResolved      FraudAlertStatus = "resolved"
	FraudAlertStatusFalsePositive FraudAlertStatus = "false_positive"
)

type FraudRuleCategory string

const (
	FraudRuleCategoryTransaction FraudRuleCategory = "transaction"
	FraudRuleCategoryBehavioral  FraudRuleCategory = "behavioral"
	FraudRuleCategoryCompliance  FraudRuleCategory = "compliance"
	FraudRuleCategoryAccount     FraudRuleCategory = "account"
	FraudRuleCategoryNetwork     FraudRuleCategory = "network"
)

type FraudRuleType string

const (
	FraudRuleTypeThreshold   FraudRuleType = "threshold"
	FraudRuleTypePattern     FraudRuleType = "pattern"
	FraudRuleTypeVelocity    FraudRuleType = "velocity"
	FraudRuleTypeStatistical FraudRuleType = "statistical"
	FraudRuleTypeML          FraudRuleType = "ml"
	FraudRuleTypeCustom      FraudRuleType = "custom"
)

type FraudAction string

const (
	FraudActionAlert       FraudAction = "alert"
	FraudActionHold        FraudAction = "hold"
	FraudActionBlock       FraudAction = "block"
	FraudActionFreeze      FraudAction = "freeze"
	FraudActionInvestigate FraudAction = "investigate"
	FraudActionMonitor     FraudAction = "monitor"
)

type FraudRuleStatus string

const (
	FraudRuleStatusActive   FraudRuleStatus = "active"
	FraudRuleStatusInactive FraudRuleStatus = "inactive"
	FraudRuleStatusDraft    FraudRuleStatus = "draft"
)

type FraudCaseStatus string

const (
	FraudCaseStatusOpen          FraudCaseStatus = "open"
	FraudCaseStatusAssigned      FraudCaseStatus = "assigned"
	FraudCaseStatusInvestigating FraudCaseStatus = "investigating"
	FraudCaseStatusResolved      FraudCaseStatus = "resolved"
	FraudCaseStatusClosed        FraudCaseStatus = "closed"
)

type FraudCasePriority string

const (
	FraudCasePriorityLow      FraudCasePriority = "low"
	FraudCasePriorityMedium   FraudCasePriority = "medium"
	FraudCasePriorityHigh     FraudCasePriority = "high"
	FraudCasePriorityCritical FraudCasePriority = "critical"
)

type FraudCaseCategory string

const (
	FraudCaseCategoryTransactionFraud FraudCaseCategory = "transaction_fraud"
	FraudCaseCategoryAccountTakeover  FraudCaseCategory = "account_takeover"
	FraudCaseCategoryCompliance       FraudCaseCategory = "compliance"
	FraudCaseCategoryMoneyLaundering  FraudCaseCategory = "money_laundering"
	FraudCaseCategoryIdentityTheft    FraudCaseCategory = "identity_theft"
)

type FraudCaseOutcome string

const (
	FraudCaseOutcomeConfirmedFraud FraudCaseOutcome = "confirmed_fraud"
	FraudCaseOutcomeFalsePositive  FraudCaseOutcome = "false_positive"
	FraudCaseOutcomeInconclusive   FraudCaseOutcome = "inconclusive"
	FraudCaseOutcomeSystemError    FraudCaseOutcome = "system_error"
)

type AccountFraudStatusType string

const (
	AccountFraudStatusNormal      AccountFraudStatusType = "normal"
	AccountFraudStatusUnderReview AccountFraudStatusType = "under_review"
	AccountFraudStatusRestricted  AccountFraudStatusType = "restricted"
	AccountFraudStatusSuspended   AccountFraudStatusType = "suspended"
	AccountFraudStatusFrozen      AccountFraudStatusType = "frozen"
	AccountFraudStatusTerminated  AccountFraudStatusType = "terminated"
)

type FraudRiskLevel string

const (
	FraudRiskLevelLow      FraudRiskLevel = "low"
	FraudRiskLevelMedium   FraudRiskLevel = "medium"
	FraudRiskLevelHigh     FraudRiskLevel = "high"
	FraudRiskLevelCritical FraudRiskLevel = "critical"
)

type MonitoringLevel string

const (
	MonitoringLevelStandard  MonitoringLevel = "standard"
	MonitoringLevelEnhanced  MonitoringLevel = "enhanced"
	MonitoringLevelIntensive MonitoringLevel = "intensive"
)

type RestrictionType string

const (
	RestrictionTypeTransactionLimit RestrictionType = "transaction_limit"
	RestrictionTypeDailyLimit       RestrictionType = "daily_limit"
	RestrictionTypeVelocityLimit    RestrictionType = "velocity_limit"
	RestrictionTypeGeographic       RestrictionType = "geographic"
	RestrictionTypeTimeBased        RestrictionType = "time_based"
	RestrictionTypeManualApproval   RestrictionType = "manual_approval"
)

// Methods

// IsActive checks if the fraud rule is currently active
func (r *FraudRule) IsActive() bool {
	now := time.Now()
	return r.Status == FraudRuleStatusActive &&
		!r.EffectiveAt.After(now) &&
		(r.ExpiresAt == nil || r.ExpiresAt.After(now))
}

// CanTransition checks if the fraud alert can transition to the new status
func (a *FraudAlert) CanTransition(newStatus FraudAlertStatus) error {
	validTransitions := map[FraudAlertStatus][]FraudAlertStatus{
		FraudAlertStatusNew: {
			FraudAlertStatusAcknowledged,
			FraudAlertStatusInvestigating,
			FraudAlertStatusResolved,
			FraudAlertStatusFalsePositive,
		},
		FraudAlertStatusAcknowledged: {
			FraudAlertStatusInvestigating,
			FraudAlertStatusResolved,
			FraudAlertStatusFalsePositive,
		},
		FraudAlertStatusInvestigating: {
			FraudAlertStatusResolved,
			FraudAlertStatusFalsePositive,
		},
	}

	if validStatuses, exists := validTransitions[a.Status]; exists {
		for _, validStatus := range validStatuses {
			if validStatus == newStatus {
				return nil
			}
		}
		return fmt.Errorf("invalid transition from %s to %s", a.Status, newStatus)
	}

	return fmt.Errorf("unknown status: %s", a.Status)
}

// TransitionTo changes the alert status with validation
func (a *FraudAlert) TransitionTo(newStatus FraudAlertStatus) error {
	if err := a.CanTransition(newStatus); err != nil {
		return err
	}

	now := time.Now()
	switch newStatus {
	case FraudAlertStatusAcknowledged:
		a.AcknowledgedAt = &now
	case FraudAlertStatusResolved:
		a.ResolvedAt = &now
	}

	a.Status = newStatus
	a.UpdatedAt = now
	return nil
}

// IsHighPriority checks if the fraud case is high priority
func (c *FraudCase) IsHighPriority() bool {
	return c.Priority == FraudCasePriorityHigh || c.Priority == FraudCasePriorityCritical
}

// CanClose checks if the fraud case can be closed
func (c *FraudCase) CanClose() bool {
	return c.Status == FraudCaseStatusResolved && c.Resolution != nil
}

// Close closes the fraud case
func (c *FraudCase) Close() error {
	if !c.CanClose() {
		return fmt.Errorf("cannot close case in status %s", c.Status)
	}

	now := time.Now()
	c.Status = FraudCaseStatusClosed
	c.ClosedAt = &now
	c.UpdatedAt = now
	return nil
}

// IsRestricted checks if the account has restrictions
func (s *AccountFraudStatus) IsRestricted() bool {
	return s.Status == AccountFraudStatusRestricted ||
		s.Status == AccountFraudStatusSuspended ||
		s.Status == AccountFraudStatusFrozen
}

// CanTransition checks if the account fraud status can transition
func (s *AccountFraudStatus) CanTransition(newStatus AccountFraudStatusType) error {
	validTransitions := map[AccountFraudStatusType][]AccountFraudStatusType{
		AccountFraudStatusNormal: {
			AccountFraudStatusUnderReview,
			AccountFraudStatusRestricted,
		},
		AccountFraudStatusUnderReview: {
			AccountFraudStatusNormal,
			AccountFraudStatusRestricted,
			AccountFraudStatusSuspended,
		},
		AccountFraudStatusRestricted: {
			AccountFraudStatusNormal,
			AccountFraudStatusSuspended,
			AccountFraudStatusFrozen,
		},
		AccountFraudStatusSuspended: {
			AccountFraudStatusRestricted,
			AccountFraudStatusFrozen,
			AccountFraudStatusTerminated,
		},
		AccountFraudStatusFrozen: {
			AccountFraudStatusSuspended,
			AccountFraudStatusTerminated,
		},
	}

	if validStatuses, exists := validTransitions[s.Status]; exists {
		for _, validStatus := range validStatuses {
			if validStatus == newStatus {
				return nil
			}
		}
		return fmt.Errorf("invalid transition from %s to %s", s.Status, newStatus)
	}

	return fmt.Errorf("unknown status: %s", s.Status)
}

// TransitionTo changes the account fraud status with validation
func (s *AccountFraudStatus) TransitionTo(newStatus AccountFraudStatusType, reason string, changedBy uuid.UUID) error {
	if err := s.CanTransition(newStatus); err != nil {
		return err
	}

	now := time.Now()

	// Record status change
	statusChange := FraudStatusChange{
		FromStatus: s.Status,
		ToStatus:   newStatus,
		Reason:     reason,
		ChangedBy:  changedBy,
		ChangedAt:  now,
	}

	s.StatusHistory = append(s.StatusHistory, statusChange)

	// Update status
	s.Status = newStatus
	s.StatusChangedAt = now
	s.UpdatedAt = now

	return nil
}

// GetActiveRestrictions returns currently active restrictions
func (s *AccountFraudStatus) GetActiveRestrictions() []AccountRestriction {
	var active []AccountRestriction
	now := time.Now()

	for _, restriction := range s.Restrictions {
		if restriction.EffectiveAt.Before(now) &&
			(restriction.ExpiresAt == nil || restriction.ExpiresAt.After(now)) {
			active = append(active, restriction)
		}
	}

	return active
}
