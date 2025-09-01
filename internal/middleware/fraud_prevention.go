package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/services"
)

// FraudPreventionMiddleware provides fraud detection for API endpoints
type FraudPreventionMiddleware struct {
	fraudDetectionService services.FraudDetectionServiceInterface
	config                *FraudPreventionMiddlewareConfig
}

// FraudPreventionMiddlewareConfig defines configuration for fraud prevention middleware
type FraudPreventionMiddlewareConfig struct {
	// Enable fraud detection on specific endpoints
	EnabledEndpoints []string `json:"enabled_endpoints"`

	// Skip fraud detection for specific enterprise IDs (whitelist)
	WhitelistedEnterprises []uuid.UUID `json:"whitelisted_enterprises"`

	// Risk thresholds
	BlockThreshold float64 `json:"block_threshold"` // Block requests above this threshold
	AlertThreshold float64 `json:"alert_threshold"` // Alert for requests above this threshold
	LogThreshold   float64 `json:"log_threshold"`   // Log requests above this threshold

	// Rate limiting
	MaxRequestsPerMinute int           `json:"max_requests_per_minute"`
	RateLimitWindow      time.Duration `json:"rate_limit_window"`

	// Response configuration
	BlockResponseCode    int    `json:"block_response_code"`    // HTTP status code for blocked requests
	BlockResponseMessage string `json:"block_response_message"` // Message for blocked requests
}

// NewFraudPreventionMiddleware creates a new fraud prevention middleware
func NewFraudPreventionMiddleware(
	fraudDetectionService services.FraudDetectionServiceInterface,
	config *FraudPreventionMiddlewareConfig,
) *FraudPreventionMiddleware {
	if config == nil {
		config = &FraudPreventionMiddlewareConfig{
			BlockThreshold:       0.8,
			AlertThreshold:       0.6,
			LogThreshold:         0.4,
			MaxRequestsPerMinute: 100,
			RateLimitWindow:      1 * time.Minute,
			BlockResponseCode:    http.StatusForbidden,
			BlockResponseMessage: "Request blocked due to fraud detection",
		}
	}

	return &FraudPreventionMiddleware{
		fraudDetectionService: fraudDetectionService,
		config:                config,
	}
}

// FraudDetectionMiddleware returns a Gin middleware function for fraud detection
func (m *FraudPreventionMiddleware) FraudDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip fraud detection if service is not available
		if m.fraudDetectionService == nil {
			c.Next()
			return
		}

		// Check if endpoint is enabled for fraud detection
		if !m.isEndpointEnabled(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract enterprise ID from request
		enterpriseID, err := m.extractEnterpriseID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid or missing enterprise ID",
			})
			c.Abort()
			return
		}

		// Skip fraud detection for whitelisted enterprises
		if m.isEnterpriseWhitelisted(enterpriseID) {
			c.Next()
			return
		}

		// Perform fraud analysis on the request
		fraudResult, err := m.analyzeRequest(c, enterpriseID)
		if err != nil {
			// Log error but don't block the request
			_ = c.Error(err)
			c.Next()
			return
		}

		// Handle fraud detection results
		m.handleFraudResult(c, fraudResult)
	}
}

// isEndpointEnabled checks if fraud detection is enabled for the given endpoint
func (m *FraudPreventionMiddleware) isEndpointEnabled(path string) bool {
	if len(m.config.EnabledEndpoints) == 0 {
		return true // Enable for all endpoints if none specified
	}

	for _, endpoint := range m.config.EnabledEndpoints {
		if endpoint == path {
			return true
		}
	}
	return false
}

// extractEnterpriseID extracts enterprise ID from the request
func (m *FraudPreventionMiddleware) extractEnterpriseID(c *gin.Context) (uuid.UUID, error) {
	// Try to get from header first
	if enterpriseIDStr := c.GetHeader("X-Enterprise-ID"); enterpriseIDStr != "" {
		return uuid.Parse(enterpriseIDStr)
	}

	// Try to get from query parameter
	if enterpriseIDStr := c.Query("enterprise_id"); enterpriseIDStr != "" {
		return uuid.Parse(enterpriseIDStr)
	}

	// Try to get from request body (for POST/PUT requests)
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err == nil {
			if enterpriseIDStr, ok := requestBody["enterprise_id"].(string); ok {
				return uuid.Parse(enterpriseIDStr)
			}
		}
	}

	// Try to get from URL parameter
	if enterpriseIDStr := c.Param("enterpriseID"); enterpriseIDStr != "" {
		return uuid.Parse(enterpriseIDStr)
	}

	return uuid.Nil, &FraudPreventionError{
		Message: "Enterprise ID not found in request",
		Code:    "MISSING_ENTERPRISE_ID",
	}
}

// isEnterpriseWhitelisted checks if the enterprise is whitelisted
func (m *FraudPreventionMiddleware) isEnterpriseWhitelisted(enterpriseID uuid.UUID) bool {
	for _, whitelistedID := range m.config.WhitelistedEnterprises {
		if whitelistedID == enterpriseID {
			return true
		}
	}
	return false
}

// analyzeRequest performs fraud analysis on the incoming request
func (m *FraudPreventionMiddleware) analyzeRequest(c *gin.Context, enterpriseID uuid.UUID) (*services.FraudAnalysisResult, error) {
	// Create fraud analysis request
	fraudRequest := &services.FraudAnalysisRequest{
		TransactionID:   m.generateRequestID(c),
		EnterpriseID:    enterpriseID,
		Amount:          m.extractAmount(c),
		CurrencyCode:    m.extractCurrency(c),
		TransactionType: m.extractTransactionType(c),
		Destination:     m.extractDestination(c),
		Timestamp:       time.Now(),
		Metadata:        m.extractRequestMetadata(c),
	}

	// Perform fraud analysis
	return m.fraudDetectionService.AnalyzeTransaction(c.Request.Context(), fraudRequest)
}

// generateRequestID generates a unique ID for the request
func (m *FraudPreventionMiddleware) generateRequestID(c *gin.Context) string {
	// Try to get existing request ID
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate new request ID
	return uuid.New().String()
}

// extractAmount extracts amount from the request
func (m *FraudPreventionMiddleware) extractAmount(c *gin.Context) string {
	// Try to get from query parameter
	if amount := c.Query("amount"); amount != "" {
		return amount
	}

	// Try to get from request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		if amount, ok := requestBody["amount"].(string); ok {
			return amount
		}
		if amount, ok := requestBody["amount"].(float64); ok {
			return fmt.Sprintf("%.2f", amount)
		}
	}

	return "0"
}

// extractCurrency extracts currency from the request
func (m *FraudPreventionMiddleware) extractCurrency(c *gin.Context) string {
	// Try to get from query parameter
	if currency := c.Query("currency"); currency != "" {
		return currency
	}

	// Try to get from request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		if currency, ok := requestBody["currency"].(string); ok {
			return currency
		}
		if currency, ok := requestBody["currency_code"].(string); ok {
			return currency
		}
	}

	return "USD"
}

// extractTransactionType extracts transaction type from the request
func (m *FraudPreventionMiddleware) extractTransactionType(c *gin.Context) string {
	// Try to get from query parameter
	if txType := c.Query("transaction_type"); txType != "" {
		return txType
	}

	// Try to get from request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		if txType, ok := requestBody["transaction_type"].(string); ok {
			return txType
		}
	}

	// Infer from HTTP method and path
	switch c.Request.Method {
	case "POST":
		if c.Request.URL.Path == "/api/v1/payments" {
			return "payment"
		}
		if c.Request.URL.Path == "/api/v1/escrows" {
			return "escrow_create"
		}
	case "PUT":
		if c.Request.URL.Path == "/api/v1/escrows" {
			return "escrow_finish"
		}
	case "DELETE":
		if c.Request.URL.Path == "/api/v1/escrows" {
			return "escrow_cancel"
		}
	}

	return "unknown"
}

// extractDestination extracts destination from the request
func (m *FraudPreventionMiddleware) extractDestination(c *gin.Context) string {
	// Try to get from query parameter
	if destination := c.Query("destination"); destination != "" {
		return destination
	}

	// Try to get from request body
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		if destination, ok := requestBody["destination"].(string); ok {
			return destination
		}
		if destination, ok := requestBody["to_address"].(string); ok {
			return destination
		}
	}

	return ""
}

// extractRequestMetadata extracts metadata from the request
func (m *FraudPreventionMiddleware) extractRequestMetadata(c *gin.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add request metadata
	metadata["method"] = c.Request.Method
	metadata["path"] = c.Request.URL.Path
	metadata["user_agent"] = c.Request.UserAgent()
	metadata["remote_addr"] = c.ClientIP()
	metadata["content_type"] = c.GetHeader("Content-Type")

	// Add request body metadata if available
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err == nil {
		// Add relevant fields from request body
		if referenceID, ok := requestBody["reference_id"].(string); ok {
			metadata["reference_id"] = referenceID
		}
		if description, ok := requestBody["description"].(string); ok {
			metadata["description"] = description
		}
	}

	return metadata
}

// handleFraudResult handles the fraud detection result
func (m *FraudPreventionMiddleware) handleFraudResult(c *gin.Context, fraudResult *services.FraudAnalysisResult) {
	// Log high-risk requests
	if fraudResult.RiskScore >= m.config.LogThreshold {
		_ = c.Error(&FraudPreventionError{
			Message:   fmt.Sprintf("High-risk request detected: risk score %.2f", fraudResult.RiskScore),
			Code:      "HIGH_RISK_REQUEST",
			RiskScore: fraudResult.RiskScore,
		})
	}

	// Block requests above block threshold
	if fraudResult.RiskScore >= m.config.BlockThreshold {
		c.JSON(m.config.BlockResponseCode, gin.H{
			"error":        m.config.BlockResponseMessage,
			"risk_score":   fraudResult.RiskScore,
			"risk_factors": fraudResult.RiskFactors,
		})
		c.Abort()
		return
	}

	// Add fraud analysis result to context for downstream handlers
	c.Set("fraud_analysis", fraudResult)
	c.Set("risk_score", fraudResult.RiskScore)

	// Continue to next handler
	c.Next()
}

// FraudPreventionError represents a fraud prevention error
type FraudPreventionError struct {
	Message   string  `json:"message"`
	Code      string  `json:"code"`
	RiskScore float64 `json:"risk_score,omitempty"`
}

// Error implements the error interface
func (e *FraudPreventionError) Error() string {
	return e.Message
}
