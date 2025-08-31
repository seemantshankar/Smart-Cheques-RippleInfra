package models

import (
	"time"

	"github.com/google/uuid"
)

// OracleType represents the type of oracle provider
type OracleType string

const (
	OracleTypeAPI     OracleType = "api"
	OracleTypeWebhook OracleType = "webhook"
	OracleTypeManual  OracleType = "manual"
)

// OracleProvider represents a configured oracle service provider
type OracleProvider struct {
	ID              uuid.UUID             `json:"id" db:"id"`
	Name            string                `json:"name" db:"name"`
	Description     string                `json:"description" db:"description"`
	Type            OracleType            `json:"type" db:"type"`
	Endpoint        string                `json:"endpoint" db:"endpoint"`
	AuthConfig      OracleAuthConfig      `json:"auth_config" db:"auth_config"`
	RateLimitConfig OracleRateLimitConfig `json:"rate_limit_config" db:"rate_limit_config"`
	IsActive        bool                  `json:"is_active" db:"is_active"`
	Reliability     float64               `json:"reliability" db:"reliability"` // 0.0 - 1.0
	ResponseTime    time.Duration         `json:"response_time" db:"response_time"`
	Capabilities    []string              `json:"capabilities" db:"capabilities"`
	CreatedAt       time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at" db:"updated_at"`
}

// OracleAuthConfig represents authentication configuration for oracle providers
type OracleAuthConfig struct {
	Type       string            `json:"type"` // bearer, api_key, oauth, none
	ConfigData map[string]string `json:"config_data"`
}

// OracleRateLimitConfig represents rate limiting configuration for oracle providers
type OracleRateLimitConfig struct {
	RequestsPerSecond int `json:"requests_per_second"`
	BurstLimit        int `json:"burst_limit"`
}

// RequestStatus represents the status of an oracle request
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
	RequestStatusCached     RequestStatus = "cached"
)

// OracleRequest represents a request made to an oracle provider
type OracleRequest struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	ProviderID   uuid.UUID     `json:"provider_id" db:"provider_id"`
	Condition    string        `json:"condition" db:"condition"`
	ContextData  interface{}   `json:"context_data" db:"context_data"`
	Status       RequestStatus `json:"status" db:"status"`
	Result       *bool         `json:"result" db:"result"`
	Confidence   *float64      `json:"confidence" db:"confidence"`
	Evidence     []byte        `json:"evidence" db:"evidence"`
	Metadata     interface{}   `json:"metadata" db:"metadata"`
	VerifiedAt   *time.Time    `json:"verified_at" db:"verified_at"`
	ProofHash    *string       `json:"proof_hash" db:"proof_hash"`
	RetryCount   int           `json:"retry_count" db:"retry_count"`
	ErrorMessage *string       `json:"error_message" db:"error_message"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	CachedUntil  *time.Time    `json:"cached_until" db:"cached_until"`
}

// OracleResponse represents the response from an oracle verification
type OracleResponse struct {
	RequestID  uuid.UUID   `json:"request_id"`
	Condition  string      `json:"condition"`
	Result     bool        `json:"result"`
	Confidence float64     `json:"confidence"`  // confidence score (0.0 - 1.0)
	Evidence   []byte      `json:"evidence"`    // verification evidence
	Metadata   interface{} `json:"metadata"`    // additional metadata
	VerifiedAt time.Time   `json:"verified_at"` // verification timestamp
	ProofHash  string      `json:"proof_hash"`  // hash of the proof for integrity
}

// OracleStatus represents the health and availability status of an oracle
type OracleStatus struct {
	IsHealthy    bool          `json:"is_healthy"`
	ResponseTime time.Duration `json:"response_time"`
	LastChecked  time.Time     `json:"last_checked"`
	ErrorRate    float64       `json:"error_rate"`
	Capacity     int           `json:"capacity"`     // available capacity
	Load         int           `json:"load"`         // current load
	Version      string        `json:"version"`      // oracle version
	Capabilities []string      `json:"capabilities"` // supported capabilities
	Reliability  float64       `json:"reliability"`  // 0.0 - 1.0
}

// OracleMetrics represents performance metrics for an oracle
type OracleMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
}

// OracleConfigStatus represents the configuration status of an oracle
type OracleConfigStatus struct {
	IsValid     bool      `json:"is_valid"`
	Errors      []string  `json:"errors"`
	Warnings    []string  `json:"warnings"`
	LastUpdated time.Time `json:"last_updated"`
}
