package services

// Dispute event types
const (
	DisputeEventCreated            = "created"
	DisputeEventStatusChanged      = "status_changed"
	DisputeEventEvidenceAdded      = "evidence_added"
	DisputeEventResolutionProposed = "resolution_proposed"
	DisputeEventResolutionAccepted = "resolution_accepted"
	DisputeEventEscalated          = "escalated"
	DisputeEventResolved           = "resolved"
	DisputeEventClosed             = "closed"
	DisputeEventReopened           = "reopened"
)

// Urgency indicators for dispute categorization
const (
	UrgencyUrgent    = "urgent"
	UrgencyCritical  = "critical"
	UrgencyImmediate = "immediate"
	UrgencyBlocking  = "blocking"
)

// Positive sentiment indicators
const (
	SentimentSatisfactory = "satisfactory"
	SentimentCompleted    = "completed"
	SentimentSuccessful   = "successful"
	SentimentResolved     = "resolved"
	SentimentFixed        = "fixed"
)

// Priority levels
const (
	PriorityKeyAccount = "key_account"
	PriorityVIP        = "vip"
	PriorityRegulatory = "regulatory"
)

// Risk levels (some may be defined elsewhere)
const (
	RiskLevelMedium = "medium"
)

// Status values
const (
	StatusStable     = "stable"
	StatusUnknown    = "unknown"
	StatusActive     = "active"
	StatusFlagged    = "flagged"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusInProgress = "in_progress"
)

// Risk trend values
const (
	RiskTrendIncreasing = "increasing"
	RiskTrendDecreasing = "decreasing"
	RiskTrendStable     = "stable"
	RiskTrendUnknown    = "unknown"
)

// Transaction statuses
const (
	TransactionStatusApproved = "approved"
)

// Alert severity levels (defined in test_utils.go)
const (
// TestAlertSeverityMedium defined in test_utils.go
)

// API paths
const (
	APIPathEscrows = "/api/v1/escrows"
)
