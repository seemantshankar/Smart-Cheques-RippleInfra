package models

import (
	"time"
)

// DisputeStatus represents the current status of a dispute
type DisputeStatus string

const (
	DisputeStatusInitiated   DisputeStatus = "initiated"
	DisputeStatusUnderReview DisputeStatus = "under_review"
	DisputeStatusEscalated   DisputeStatus = "escalated"
	DisputeStatusResolved    DisputeStatus = "resolved"
	DisputeStatusClosed      DisputeStatus = "closed"
	DisputeStatusCancelled   DisputeStatus = "cancelled"
)

// DisputeCategory represents the type of dispute
type DisputeCategory string

const (
	DisputeCategoryPayment        DisputeCategory = "payment"
	DisputeCategoryMilestone      DisputeCategory = "milestone"
	DisputeCategoryContractBreach DisputeCategory = "contract_breach"
	DisputeCategoryFraud          DisputeCategory = "fraud"
	DisputeCategoryTechnical      DisputeCategory = "technical"
	DisputeCategoryOther          DisputeCategory = "other"
)

// DisputePriority represents the priority level of a dispute
type DisputePriority string

const (
	DisputePriorityLow    DisputePriority = "low"
	DisputePriorityNormal DisputePriority = "normal"
	DisputePriorityHigh   DisputePriority = "high"
	DisputePriorityUrgent DisputePriority = "urgent"
)

// DisputeResolutionMethod represents how the dispute was resolved
type DisputeResolutionMethod string

const (
	DisputeResolutionMethodMutualAgreement DisputeResolutionMethod = "mutual_agreement"
	DisputeResolutionMethodMediation       DisputeResolutionMethod = "mediation"
	DisputeResolutionMethodArbitration     DisputeResolutionMethod = "arbitration"
	DisputeResolutionMethodCourt           DisputeResolutionMethod = "court"
	DisputeResolutionMethodAdministrative  DisputeResolutionMethod = "administrative"
)

// Dispute represents a dispute in the system
type Dispute struct {
	ID          string          `json:"id" db:"id" gorm:"primaryKey"`
	Title       string          `json:"title" db:"title" validate:"required,min=5,max=200"`
	Description string          `json:"description" db:"description" validate:"required,min=10,max=2000"`
	Category    DisputeCategory `json:"category" db:"category" validate:"required"`
	Priority    DisputePriority `json:"priority" db:"priority" validate:"required"`
	Status      DisputeStatus   `json:"status" db:"status" validate:"required"`

	// Related entities
	SmartChequeID *string `json:"smart_cheque_id,omitempty" db:"smart_cheque_id"`
	MilestoneID   *string `json:"milestone_id,omitempty" db:"milestone_id"`
	ContractID    *string `json:"contract_id,omitempty" db:"contract_id"`
	TransactionID *string `json:"transaction_id,omitempty" db:"transaction_id"`

	// Parties involved
	InitiatorID    string `json:"initiator_id" db:"initiator_id" validate:"required"`
	InitiatorType  string `json:"initiator_type" db:"initiator_type" validate:"required"` // "enterprise" or "user"
	RespondentID   string `json:"respondent_id" db:"respondent_id" validate:"required"`
	RespondentType string `json:"respondent_type" db:"respondent_type" validate:"required"` // "enterprise" or "user"

	// Financial impact
	DisputedAmount *float64  `json:"disputed_amount,omitempty" db:"disputed_amount"`
	Currency       *Currency `json:"currency,omitempty" db:"currency"`

	// Resolution details
	Resolution *DisputeResolution `json:"resolution,omitempty" db:"-" gorm:"foreignKey:DisputeID"`

	// Evidence and documents
	Evidence []DisputeEvidence `json:"evidence,omitempty" db:"-" gorm:"foreignKey:DisputeID"`

	// Timestamps and tracking
	InitiatedAt    time.Time  `json:"initiated_at" db:"initiated_at"`
	LastActivityAt time.Time  `json:"last_activity_at" db:"last_activity_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty" db:"closed_at"`

	// Metadata
	Tags     []string               `json:"tags,omitempty" db:"tags" gorm:"type:jsonb"`
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata" gorm:"type:jsonb"`

	// Audit fields
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DisputeEvidence represents evidence attached to a dispute
type DisputeEvidence struct {
	ID          string    `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID   string    `json:"dispute_id" db:"dispute_id" validate:"required"`
	FileName    string    `json:"file_name" db:"file_name" validate:"required"`
	FileType    string    `json:"file_type" db:"file_type" validate:"required"`
	FileSize    int64     `json:"file_size" db:"file_size" validate:"required"`
	FilePath    string    `json:"file_path" db:"file_path" validate:"required"`
	Description string    `json:"description,omitempty" db:"description"`
	UploadedBy  string    `json:"uploaded_by" db:"uploaded_by" validate:"required"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DisputeResolution represents the resolution of a dispute
type DisputeResolution struct {
	ID                 string                  `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID          string                  `json:"dispute_id" db:"dispute_id" validate:"required"`
	Method             DisputeResolutionMethod `json:"method" db:"method" validate:"required"`
	ResolutionDetails  string                  `json:"resolution_details" db:"resolution_details" validate:"required"`
	OutcomeAmount      *float64                `json:"outcome_amount,omitempty" db:"outcome_amount"`
	OutcomeDescription string                  `json:"outcome_description" db:"outcome_description"`

	// Resolution parties
	MediatorID      *string `json:"mediator_id,omitempty" db:"mediator_id"`
	ArbitratorID    *string `json:"arbitrator_id,omitempty" db:"arbitrator_id"`
	CourtCaseNumber *string `json:"court_case_number,omitempty" db:"court_case_number"`

	// Agreement tracking
	InitiatorAccepted  bool       `json:"initiator_accepted" db:"initiator_accepted"`
	RespondentAccepted bool       `json:"respondent_accepted" db:"respondent_accepted"`
	AcceptanceDeadline *time.Time `json:"acceptance_deadline,omitempty" db:"acceptance_deadline"`

	// Execution details
	IsExecuted bool       `json:"is_executed" db:"is_executed"`
	ExecutedAt *time.Time `json:"executed_at,omitempty" db:"executed_at"`
	ExecutedBy *string    `json:"executed_by,omitempty" db:"executed_by"`

	// Audit fields
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DisputeComment represents a comment or note on a dispute
type DisputeComment struct {
	ID         string    `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID  string    `json:"dispute_id" db:"dispute_id" validate:"required"`
	AuthorID   string    `json:"author_id" db:"author_id" validate:"required"`
	AuthorType string    `json:"author_type" db:"author_type" validate:"required"` // "enterprise", "user", "mediator", "admin"
	Content    string    `json:"content" db:"content" validate:"required,min=1,max=1000"`
	IsInternal bool      `json:"is_internal" db:"is_internal"` // Internal notes not visible to all parties
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// DisputeAuditLog represents an audit log entry for dispute activities
type DisputeAuditLog struct {
	ID        string                 `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID string                 `json:"dispute_id" db:"dispute_id"`
	Action    string                 `json:"action" db:"action"` // created, updated, status_changed, evidence_added, resolved, etc.
	UserID    string                 `json:"user_id" db:"user_id"`
	UserType  string                 `json:"user_type" db:"user_type"`
	Details   string                 `json:"details,omitempty" db:"details"`
	OldValue  map[string]interface{} `json:"old_value,omitempty" db:"old_value" gorm:"type:jsonb"`
	NewValue  map[string]interface{} `json:"new_value,omitempty" db:"new_value" gorm:"type:jsonb"`
	IPAddress string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// DisputeFilter represents filter criteria for dispute queries
type DisputeFilter struct {
	InitiatorID     *string          `json:"initiator_id,omitempty"`
	RespondentID    *string          `json:"respondent_id,omitempty"`
	Category        *DisputeCategory `json:"category,omitempty"`
	Priority        *DisputePriority `json:"priority,omitempty"`
	Status          *DisputeStatus   `json:"status,omitempty"`
	SmartChequeID   *string          `json:"smart_cheque_id,omitempty"`
	MilestoneID     *string          `json:"milestone_id,omitempty"`
	ContractID      *string          `json:"contract_id,omitempty"`
	DateFrom        *time.Time       `json:"date_from,omitempty"`
	DateTo          *time.Time       `json:"date_to,omitempty"`
	MinAmount       *float64         `json:"min_amount,omitempty"`
	MaxAmount       *float64         `json:"max_amount,omitempty"`
	Tags            []string         `json:"tags,omitempty"`
	SearchText      *string          `json:"search_text,omitempty"`
	ExcludeStatuses []DisputeStatus  `json:"exclude_statuses,omitempty"`
}

// DisputeStats represents dispute statistics
type DisputeStats struct {
	TotalDisputes         int64                     `json:"total_disputes"`
	ActiveDisputes        int64                     `json:"active_disputes"`
	ResolvedDisputes      int64                     `json:"resolved_disputes"`
	AverageResolutionTime time.Duration             `json:"average_resolution_time"`
	DisputesByCategory    map[DisputeCategory]int64 `json:"disputes_by_category"`
	DisputesByStatus      map[DisputeStatus]int64   `json:"disputes_by_status"`
	DisputesByPriority    map[DisputePriority]int64 `json:"disputes_by_priority"`
}

// DisputeNotification represents a notification for dispute events
type DisputeNotification struct {
	ID        string                 `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID string                 `json:"dispute_id" db:"dispute_id"`
	Recipient string                 `json:"recipient" db:"recipient"`
	Type      string                 `json:"type" db:"type"`       // email, sms, webhook, in_app
	Channel   string                 `json:"channel" db:"channel"` // email, sms, webhook, in_app
	Subject   string                 `json:"subject" db:"subject"`
	Message   string                 `json:"message" db:"message"`
	Status    string                 `json:"status" db:"status"` // pending, sent, failed
	SentAt    *time.Time             `json:"sent_at,omitempty" db:"sent_at"`
	ErrorMsg  *string                `json:"error_msg,omitempty" db:"error_msg"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// CategorizationRuleType represents the type of categorization rule
type CategorizationRuleType string

const (
	CategorizationRuleTypeKeyword   CategorizationRuleType = "keyword"
	CategorizationRuleTypePattern   CategorizationRuleType = "pattern"
	CategorizationRuleTypeSemantic  CategorizationRuleType = "semantic"
	CategorizationRuleTypeEntity    CategorizationRuleType = "entity"
	CategorizationRuleTypeComposite CategorizationRuleType = "composite"
)

// CategorizationRule represents a configurable categorization rule
type CategorizationRule struct {
	ID          string                 `json:"id" db:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" db:"name" validate:"required,min=3,max=100"`
	Description string                 `json:"description" db:"description" validate:"max=500"`
	Type        CategorizationRuleType `json:"type" db:"type" validate:"required"`
	Category    DisputeCategory        `json:"category" db:"category" validate:"required"`
	Priority    DisputePriority        `json:"priority" db:"priority" validate:"required"`

	// Rule configuration
	Keywords     []string               `json:"keywords,omitempty" db:"keywords" gorm:"type:jsonb"`
	Patterns     []string               `json:"patterns,omitempty" db:"patterns" gorm:"type:jsonb"`
	Entities     []string               `json:"entities,omitempty" db:"entities" gorm:"type:jsonb"`
	SemanticKeys []string               `json:"semantic_keys,omitempty" db:"semantic_keys" gorm:"type:jsonb"`
	Conditions   map[string]interface{} `json:"conditions,omitempty" db:"conditions" gorm:"type:jsonb"`

	// Scoring and thresholds
	BaseConfidence float64 `json:"base_confidence" db:"base_confidence" validate:"min=0,max=1"`
	Weight         float64 `json:"weight" db:"weight" validate:"min=0,max=1"`
	MinConfidence  float64 `json:"min_confidence" db:"min_confidence" validate:"min=0,max=1"`
	MaxConfidence  float64 `json:"max_confidence" db:"max_confidence" validate:"min=0,max=1"`

	// Rule management
	IsActive      bool      `json:"is_active" db:"is_active"`
	PriorityOrder int       `json:"priority_order" db:"priority_order"`
	CreatedBy     string    `json:"created_by" db:"created_by"`
	UpdatedBy     string    `json:"updated_by" db:"updated_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// Performance tracking
	UseCount         int64      `json:"use_count" db:"use_count"`
	SuccessCount     int64      `json:"success_count" db:"success_count"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	PerformanceScore float64    `json:"performance_score" db:"performance_score"`
}

// CategorizationRuleGroup represents a group of related categorization rules
type CategorizationRuleGroup struct {
	ID            string               `json:"id" db:"id" gorm:"primaryKey"`
	Name          string               `json:"name" db:"name" validate:"required,min=3,max=100"`
	Description   string               `json:"description" db:"description" validate:"max=500"`
	Category      DisputeCategory      `json:"category" db:"category"`
	IsActive      bool                 `json:"is_active" db:"is_active"`
	PriorityOrder int                  `json:"priority_order" db:"priority_order"`
	Rules         []CategorizationRule `json:"rules,omitempty" db:"-" gorm:"foreignKey:RuleGroupID"`
	CreatedBy     string               `json:"created_by" db:"created_by"`
	UpdatedBy     string               `json:"updated_by" db:"updated_by"`
	CreatedAt     time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" db:"updated_at"`
}

// CategorizationRulePerformance tracks rule performance metrics
type CategorizationRulePerformance struct {
	ID                string    `json:"id" db:"id" gorm:"primaryKey"`
	RuleID            string    `json:"rule_id" db:"rule_id"`
	PeriodStart       time.Time `json:"period_start" db:"period_start"`
	PeriodEnd         time.Time `json:"period_end" db:"period_end"`
	TotalApplications int64     `json:"total_applications" db:"total_applications"`
	SuccessfulMatches int64     `json:"successful_matches" db:"successful_matches"`
	AccuracyRate      float64   `json:"accuracy_rate" db:"accuracy_rate"`
	AverageConfidence float64   `json:"average_confidence" db:"average_confidence"`
	FalsePositives    int64     `json:"false_positives" db:"false_positives"`
	FalseNegatives    int64     `json:"false_negatives" db:"false_negatives"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// CategorizationRuleTemplate represents reusable rule templates
type CategorizationRuleTemplate struct {
	ID          string                 `json:"id" db:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" db:"name" validate:"required,min=3,max=100"`
	Description string                 `json:"description" db:"description" validate:"max=500"`
	Category    DisputeCategory        `json:"category" db:"category"`
	Type        CategorizationRuleType `json:"type" db:"type"`
	Template    map[string]interface{} `json:"template" db:"template" gorm:"type:jsonb"`
	IsPublic    bool                   `json:"is_public" db:"is_public"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	UseCount    int64                  `json:"use_count" db:"use_count"`
}

// MLModelStatus represents the status of an ML model
type MLModelStatus string

const (
	MLModelStatusTraining MLModelStatus = "training"
	MLModelStatusTrained  MLModelStatus = "trained"
	MLModelStatusFailed   MLModelStatus = "failed"
	MLModelStatusDeployed MLModelStatus = "deployed"
	MLModelStatusRetired  MLModelStatus = "retired"
)

// CategorizationMLModel represents a machine learning model for dispute categorization
type CategorizationMLModel struct {
	ID          string        `json:"id" db:"id" gorm:"primaryKey"`
	Name        string        `json:"name" db:"name" validate:"required,min=3,max=100"`
	Description string        `json:"description" db:"description" validate:"max=500"`
	Version     string        `json:"version" db:"version" validate:"required"`
	Algorithm   string        `json:"algorithm" db:"algorithm" validate:"required"` // e.g., "random_forest", "svm", "neural_network"
	Status      MLModelStatus `json:"status" db:"status" validate:"required"`

	// Model metadata
	TrainingDataSize int     `json:"training_data_size" db:"training_data_size"`
	Accuracy         float64 `json:"accuracy" db:"accuracy"`
	Precision        float64 `json:"precision" db:"precision"`
	Recall           float64 `json:"recall" db:"recall"`
	F1Score          float64 `json:"f1_score" db:"f1_score"`

	// Model parameters
	Parameters   map[string]interface{} `json:"parameters" db:"parameters" gorm:"type:jsonb"`
	FeatureNames []string               `json:"feature_names" db:"feature_names" gorm:"type:jsonb"`

	// Model artifacts (serialized model data)
	ModelData []byte `json:"model_data,omitempty" db:"model_data"`
	ModelPath string `json:"model_path" db:"model_path"`

	// Training information
	TrainedAt    *time.Time `json:"trained_at,omitempty" db:"trained_at"`
	TrainingTime int64      `json:"training_time" db:"training_time"` // in seconds
	DeployedAt   *time.Time `json:"deployed_at,omitempty" db:"deployed_at"`

	// Performance tracking
	UseCount     int64      `json:"use_count" db:"use_count"`
	SuccessCount int64      `json:"success_count" db:"success_count"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`

	// Audit fields
	CreatedBy string    `json:"created_by" db:"created_by"`
	UpdatedBy string    `json:"updated_by" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CategorizationTrainingData represents training data for ML models
type CategorizationTrainingData struct {
	ID          string          `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID   string          `json:"dispute_id" db:"dispute_id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description" db:"description"`
	Category    DisputeCategory `json:"category" db:"category"`
	Priority    DisputePriority `json:"priority" db:"priority"`

	// Extracted features
	Features map[string]interface{} `json:"features" db:"features" gorm:"type:jsonb"`

	// Training metadata
	IsValidated bool       `json:"is_validated" db:"is_validated"`
	ValidatedBy *string    `json:"validated_by,omitempty" db:"validated_by"`
	ValidatedAt *time.Time `json:"validated_at,omitempty" db:"validated_at"`

	// Usage tracking
	UseCount   int64      `json:"use_count" db:"use_count"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CategorizationPrediction represents a prediction made by an ML model
type CategorizationPrediction struct {
	ID                string          `json:"id" db:"id" gorm:"primaryKey"`
	DisputeID         string          `json:"dispute_id" db:"dispute_id"`
	ModelID           string          `json:"model_id" db:"model_id"`
	PredictedCategory DisputeCategory `json:"predicted_category" db:"predicted_category"`
	PredictedPriority DisputePriority `json:"predicted_priority" db:"predicted_priority"`
	Confidence        float64         `json:"confidence" db:"confidence"`

	// Prediction details
	PredictionScores map[string]float64     `json:"prediction_scores" db:"prediction_scores" gorm:"type:jsonb"`
	Features         map[string]interface{} `json:"features" db:"features" gorm:"type:jsonb"`

	// Validation
	IsCorrect       *bool            `json:"is_correct,omitempty" db:"is_correct"`
	CorrectCategory *DisputeCategory `json:"correct_category,omitempty" db:"correct_category"`
	CorrectPriority *DisputePriority `json:"correct_priority,omitempty" db:"correct_priority"`
	ValidatedBy     *string          `json:"validated_by,omitempty" db:"validated_by"`
	ValidatedAt     *time.Time       `json:"validated_at,omitempty" db:"validated_at"`

	// Performance tracking
	ResponseTime int64     `json:"response_time" db:"response_time"` // in milliseconds
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// MLModelMetrics represents performance metrics for ML models
type MLModelMetrics struct {
	ID          string    `json:"id" db:"id" gorm:"primaryKey"`
	ModelID     string    `json:"model_id" db:"model_id"`
	PeriodStart time.Time `json:"period_start" db:"period_start"`
	PeriodEnd   time.Time `json:"period_end" db:"period_end"`

	// Overall metrics
	TotalPredictions   int64   `json:"total_predictions" db:"total_predictions"`
	CorrectPredictions int64   `json:"correct_predictions" db:"correct_predictions"`
	Accuracy           float64 `json:"accuracy" db:"accuracy"`

	// Category-specific metrics
	CategoryAccuracy  map[string]float64 `json:"category_accuracy" db:"category_accuracy" gorm:"type:jsonb"`
	CategoryPrecision map[string]float64 `json:"category_precision" db:"category_precision" gorm:"type:jsonb"`
	CategoryRecall    map[string]float64 `json:"category_recall" db:"category_recall" gorm:"type:jsonb"`

	// Performance metrics
	AvgResponseTime int64 `json:"avg_response_time" db:"avg_response_time"`
	MinResponseTime int64 `json:"min_response_time" db:"min_response_time"`
	MaxResponseTime int64 `json:"max_response_time" db:"max_response_time"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
