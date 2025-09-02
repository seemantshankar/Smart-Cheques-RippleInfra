package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// FraudAlertingServiceInterface defines the interface for fraud alerting operations
type FraudAlertingServiceInterface interface {
	// Alert processing
	ProcessAlert(ctx context.Context, alert *models.FraudAlert) error
	SendNotifications(ctx context.Context, alert *models.FraudAlert) error
	EscalateAlert(ctx context.Context, alertID uuid.UUID, reason string) error

	// Alert management
	AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID, notes string) error
	ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error
	AssignAlert(ctx context.Context, alertID uuid.UUID, assignedTo uuid.UUID) error

	// Notification configuration
	ConfigureNotificationChannels(ctx context.Context, enterpriseID uuid.UUID, channels []FraudNotificationChannel) error
	GetNotificationChannels(ctx context.Context, enterpriseID uuid.UUID) ([]FraudNotificationChannel, error)

	// Alert correlation
	CorrelateAlerts(ctx context.Context, enterpriseID uuid.UUID, timeWindow time.Duration) ([]*AlertCorrelation, error)
	DeduplicateAlerts(ctx context.Context, alerts []*models.FraudAlert) ([]*models.FraudAlert, error)

	// Escalation management
	GetEscalationRules(ctx context.Context) ([]*AlertEscalationRule, error)
	CreateEscalationRule(ctx context.Context, rule *AlertEscalationRule) error
	UpdateEscalationRule(ctx context.Context, rule *AlertEscalationRule) error
	DeleteEscalationRule(ctx context.Context, ruleID uuid.UUID) error
}

// FraudAlertingService implements the fraud alerting service interface
type FraudAlertingService struct {
	fraudRepo       repository.FraudRepositoryInterface
	enterpriseRepo  repository.EnterpriseRepositoryInterface
	messagingClient messaging.EventBus
	config          *FraudAlertingConfig
}

// NewFraudAlertingService creates a new fraud alerting service instance
func NewFraudAlertingService(
	fraudRepo repository.FraudRepositoryInterface,
	enterpriseRepo repository.EnterpriseRepositoryInterface,
	messagingClient messaging.EventBus,
	config *FraudAlertingConfig,
) FraudAlertingServiceInterface {
	return &FraudAlertingService{
		fraudRepo:       fraudRepo,
		enterpriseRepo:  enterpriseRepo,
		messagingClient: messagingClient,
		config:          config,
	}
}

// Configuration types
type FraudAlertingConfig struct {
	// Notification settings
	DefaultChannels     []string      `json:"default_channels"`      // ["email", "webhook"]
	EscalationDelay     time.Duration `json:"escalation_delay"`      // e.g., 30 minutes
	MaxEscalationLevels int           `json:"max_escalation_levels"` // e.g., 3

	// Email configuration
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`

	// Webhook configuration
	WebhookTimeout time.Duration `json:"webhook_timeout"` // e.g., 10 seconds
	WebhookRetries int           `json:"webhook_retries"` // e.g., 3

	// SMS configuration
	SMSProvider   string `json:"sms_provider"` // e.g., "twilio"
	SMSAccountSID string `json:"sms_account_sid"`
	SMSAuthToken  string `json:"sms_auth_token"`
	SMSFromNumber string `json:"sms_from_number"`

	// Alert correlation
	CorrelationWindow   time.Duration `json:"correlation_window"`   // e.g., 1 hour
	DeduplicationWindow time.Duration `json:"deduplication_window"` // e.g., 5 minutes
}

// Request/Response types
type FraudNotificationChannel struct {
	Type      string                 `json:"type"`   // "email", "sms", "webhook"
	Config    map[string]interface{} `json:"config"` // Channel-specific configuration
	IsActive  bool                   `json:"is_active"`
	Priority  int                    `json:"priority"` // Lower number = higher priority
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type AlertCorrelation struct {
	CorrelationID uuid.UUID   `json:"correlation_id"`
	EnterpriseID  uuid.UUID   `json:"enterprise_id"`
	AlertIDs      []uuid.UUID `json:"alert_ids"`
	Pattern       string      `json:"pattern"`
	Confidence    float64     `json:"confidence"`
	RiskScore     float64     `json:"risk_score"`
	CreatedAt     time.Time   `json:"created_at"`
}

type AlertEscalationRule struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []EscalationAction     `json:"actions"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type EscalationAction struct {
	Type       string                 `json:"type"` // "notify", "assign", "escalate", "create_case"
	Parameters map[string]interface{} `json:"parameters"`
	Delay      time.Duration          `json:"delay"`
}

// ProcessAlert processes a fraud alert and sends notifications
func (s *FraudAlertingService) ProcessAlert(ctx context.Context, alert *models.FraudAlert) error {
	// Update alert status
	alert.Status = models.FraudAlertStatusNew
	alert.DetectedAt = time.Now()
	alert.UpdatedAt = time.Now()

	// Save alert to database
	if err := s.fraudRepo.CreateFraudAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to save fraud alert: %w", err)
	}

	// Send notifications
	if err := s.SendNotifications(ctx, alert); err != nil {
		return fmt.Errorf("failed to send notifications: %w", err)
	}

	// Check for escalation
	if err := s.checkEscalation(ctx, alert); err != nil {
		return fmt.Errorf("failed to check escalation: %w", err)
	}

	// Publish alert event
	s.publishAlertProcessedEvent(ctx, alert)

	return nil
}

// SendNotifications sends notifications through configured channels
func (s *FraudAlertingService) SendNotifications(ctx context.Context, alert *models.FraudAlert) error {
	// Get notification channels for the enterprise
	channels, err := s.GetNotificationChannels(ctx, alert.EnterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get notification channels: %w", err)
	}

	// If no channels configured, use defaults
	if len(channels) == 0 {
		channels = s.getDefaultChannels()
	}

	// Sort channels by priority
	sortChannelsByPriority(channels)

	// Send notifications through each channel
	for _, channel := range channels {
		if !channel.IsActive {
			continue
		}

		if err := s.sendNotification(ctx, alert, channel); err != nil {
			// Log error but continue with other channels
			fmt.Printf("Warning: Failed to send notification via %s: %v\n", channel.Type, err)
		}
	}

	// Update alert notification status
	alert.NotifiedAt = &time.Time{}
	*alert.NotifiedAt = time.Now()
	alert.NotificationChannels = s.getChannelTypes(channels)

	if err := s.fraudRepo.UpdateFraudAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update alert notification status: %w", err)
	}

	return nil
}

// sendNotification sends a notification through a specific channel
func (s *FraudAlertingService) sendNotification(ctx context.Context, alert *models.FraudAlert, channel FraudNotificationChannel) error {
	switch channel.Type {
	case "email":
		return s.sendEmailNotification(ctx, alert, channel)
	case "sms":
		return s.sendSMSNotification(ctx, alert, channel)
	case "webhook":
		return s.sendWebhookNotification(ctx, alert, channel)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// sendEmailNotification sends an email notification
func (s *FraudAlertingService) sendEmailNotification(ctx context.Context, alert *models.FraudAlert, channel FraudNotificationChannel) error {
	// Get enterprise details
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(alert.EnterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get enterprise details: %w", err)
	}

	// Build email content
	subject := fmt.Sprintf("Fraud Alert: %s", alert.Title)
	body := s.buildEmailBody(alert, enterprise)

	// In a real implementation, this would use an email service
	// For now, we'll just log the email
	fmt.Printf("Sending email notification:\nSubject: %s\nTo: %s\nBody: %s\n",
		subject, enterprise.Email, body)

	return nil
}

// sendSMSNotification sends an SMS notification
func (s *FraudAlertingService) sendSMSNotification(ctx context.Context, alert *models.FraudAlert, channel FraudNotificationChannel) error {
	// Get enterprise details
	enterprise, err := s.enterpriseRepo.GetEnterpriseByID(alert.EnterpriseID)
	if err != nil {
		return fmt.Errorf("failed to get enterprise details: %w", err)
	}

	// Build SMS content
	message := s.buildSMSBody(alert)

	// In a real implementation, this would use an SMS service
	// For now, we'll just log the SMS
	fmt.Printf("Sending SMS notification:\nTo: %s\nMessage: %s\n",
		enterprise.Phone, message)

	return nil
}

// sendWebhookNotification sends a webhook notification
func (s *FraudAlertingService) sendWebhookNotification(ctx context.Context, alert *models.FraudAlert, channel FraudNotificationChannel) error {
	// Get webhook URL from channel config
	webhookURL, exists := channel.Config["url"]
	if !exists {
		return fmt.Errorf("webhook URL not configured")
	}

	// Build webhook payload
	payload := s.buildWebhookPayload(alert)

	// In a real implementation, this would make an HTTP POST request
	// For now, we'll just log the webhook
	fmt.Printf("Sending webhook notification:\nURL: %s\nPayload: %+v\n",
		webhookURL, payload)

	return nil
}

// buildEmailBody builds the email body content
func (s *FraudAlertingService) buildEmailBody(alert *models.FraudAlert, enterprise *models.Enterprise) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("Dear %s,\n\n", enterprise.LegalName))
	body.WriteString("A fraud alert has been detected for your account.\n\n")
	body.WriteString("Alert Details:\n")
	body.WriteString(fmt.Sprintf("- Type: %s\n", alert.AlertType))
	body.WriteString(fmt.Sprintf("- Severity: %s\n", alert.Severity))
	body.WriteString(fmt.Sprintf("- Score: %.2f\n", alert.Score))
	body.WriteString(fmt.Sprintf("- Description: %s\n\n", alert.Description))
	body.WriteString(fmt.Sprintf("Recommendation: %s\n\n", alert.Recommendation))
	body.WriteString("Please review this alert and take appropriate action.\n\n")
	body.WriteString("Best regards,\nSmart Payment Infrastructure Team")

	return body.String()
}

// buildSMSBody builds the SMS body content
func (s *FraudAlertingService) buildSMSBody(alert *models.FraudAlert) string {
	return fmt.Sprintf("FRAUD ALERT: %s - %s. Score: %.1f. %s",
		alert.AlertType, alert.Severity, alert.Score, alert.Recommendation)
}

// buildWebhookPayload builds the webhook payload
func (s *FraudAlertingService) buildWebhookPayload(alert *models.FraudAlert) map[string]interface{} {
	return map[string]interface{}{
		"alert_id":       alert.ID.String(),
		"enterprise_id":  alert.EnterpriseID.String(),
		"alert_type":     alert.AlertType,
		"severity":       alert.Severity,
		"score":          alert.Score,
		"title":          alert.Title,
		"description":    alert.Description,
		"recommendation": alert.Recommendation,
		"detected_at":    alert.DetectedAt.Format(time.RFC3339),
		"timestamp":      time.Now().Format(time.RFC3339),
	}
}

// checkEscalation checks if the alert should be escalated
func (s *FraudAlertingService) checkEscalation(ctx context.Context, alert *models.FraudAlert) error {
	// Get escalation rules
	rules, err := s.GetEscalationRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to get escalation rules: %w", err)
	}

	// Check each rule
	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		if s.matchesEscalationConditions(alert, rule.Conditions) {
			// Execute escalation actions
			for _, action := range rule.Actions {
				if err := s.executeEscalationAction(ctx, alert, action); err != nil {
					fmt.Printf("Warning: Failed to execute escalation action: %v\n", err)
				}
			}
		}
	}

	return nil
}

// matchesEscalationConditions checks if alert matches escalation conditions
func (s *FraudAlertingService) matchesEscalationConditions(alert *models.FraudAlert, conditions map[string]interface{}) bool {
	// Check severity condition
	if severity, exists := conditions["severity"]; exists {
		if alert.Severity != severity.(models.FraudSeverity) {
			return false
		}
	}

	// Check score condition
	if minScore, exists := conditions["min_score"]; exists {
		if alert.Score < minScore.(float64) {
			return false
		}
	}

	// Check alert type condition
	if alertType, exists := conditions["alert_type"]; exists {
		if alert.AlertType != alertType.(models.FraudAlertType) {
			return false
		}
	}

	return true
}

// executeEscalationAction executes an escalation action
func (s *FraudAlertingService) executeEscalationAction(ctx context.Context, alert *models.FraudAlert, action EscalationAction) error {
	switch action.Type {
	case "notify":
		return s.executeNotifyAction(ctx, alert, action)
	case "assign":
		return s.executeAssignAction(ctx, alert, action)
	case "escalate":
		return s.executeEscalateAction(ctx, alert, action)
	case "create_case":
		return s.executeCreateCaseAction(ctx, alert, action)
	default:
		return fmt.Errorf("unknown escalation action type: %s", action.Type)
	}
}

// executeNotifyAction executes a notify escalation action
func (s *FraudAlertingService) executeNotifyAction(ctx context.Context, alert *models.FraudAlert, action EscalationAction) error {
	// Get notification channels from action parameters
	if channels, exists := action.Parameters["channels"]; exists {
		if channelList, ok := channels.([]string); ok {
			// Send notifications to specific channels
			for _, channelType := range channelList {
				channel := FraudNotificationChannel{
					Type:     channelType,
					IsActive: true,
					Priority: 1,
				}
				if err := s.sendNotification(ctx, alert, channel); err != nil {
					fmt.Printf("Warning: Failed to send escalation notification via %s: %v\n", channelType, err)
				}
			}
		}
	}

	return nil
}

// executeAssignAction executes an assign escalation action
func (s *FraudAlertingService) executeAssignAction(ctx context.Context, alert *models.FraudAlert, action EscalationAction) error {
	if assignedTo, exists := action.Parameters["assigned_to"]; exists {
		if userID, ok := assignedTo.(string); ok {
			if uuid, err := uuid.Parse(userID); err == nil {
				alert.AssignedTo = &uuid
				alert.Status = models.FraudAlertStatusInvestigating
				alert.UpdatedAt = time.Now()

				return s.fraudRepo.UpdateFraudAlert(ctx, alert)
			}
		}
	}

	return nil
}

// executeEscalateAction executes an escalate escalation action
func (s *FraudAlertingService) executeEscalateAction(ctx context.Context, alert *models.FraudAlert, action EscalationAction) error {
	// Increase alert severity
	switch alert.Severity {
	case models.FraudSeverityLow:
		alert.Severity = models.FraudSeverityMedium
	case models.FraudSeverityMedium:
		alert.Severity = models.FraudSeverityHigh
	case models.FraudSeverityHigh:
		alert.Severity = models.FraudSeverityCritical
	}

	alert.UpdatedAt = time.Now()
	return s.fraudRepo.UpdateFraudAlert(ctx, alert)
}

// executeCreateCaseAction executes a create case escalation action
func (s *FraudAlertingService) executeCreateCaseAction(ctx context.Context, alert *models.FraudAlert, action EscalationAction) error {
	// Create a fraud case
	fraudCase := &models.FraudCase{
		ID:           uuid.New(),
		EnterpriseID: alert.EnterpriseID,
		CaseNumber:   fmt.Sprintf("FC-%s", time.Now().Format("20060102-0001")),
		Status:       models.FraudCaseStatusOpen,
		Priority:     s.determineCasePriority(alert.Severity),
		Title:        fmt.Sprintf("Fraud Case: %s", alert.Title),
		Description:  alert.Description,
		Category:     s.determineCaseCategory(alert.AlertType),
		Alerts:       []uuid.UUID{alert.ID},
		OpenedAt:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.fraudRepo.CreateFraudCase(ctx, fraudCase)
}

// determineCasePriority determines case priority based on alert severity
func (s *FraudAlertingService) determineCasePriority(severity models.FraudSeverity) models.FraudCasePriority {
	switch severity {
	case models.FraudSeverityCritical:
		return models.FraudCasePriorityCritical
	case models.FraudSeverityHigh:
		return models.FraudCasePriorityHigh
	case models.FraudSeverityMedium:
		return models.FraudCasePriorityMedium
	default:
		return models.FraudCasePriorityLow
	}
}

// determineCaseCategory determines case category based on alert type
func (s *FraudAlertingService) determineCaseCategory(alertType models.FraudAlertType) models.FraudCaseCategory {
	switch alertType {
	case models.FraudAlertTypeAccountTakeover:
		return models.FraudCaseCategoryAccountTakeover
	case models.FraudAlertTypeComplianceViolation:
		return models.FraudCaseCategoryCompliance
	default:
		return models.FraudCaseCategoryTransactionFraud
	}
}

// Helper methods
func (s *FraudAlertingService) getDefaultChannels() []FraudNotificationChannel {
	var channels []FraudNotificationChannel

	for _, channelType := range s.config.DefaultChannels {
		channels = append(channels, FraudNotificationChannel{
			Type:      channelType,
			IsActive:  true,
			Priority:  1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	return channels
}

func sortChannelsByPriority(channels []FraudNotificationChannel) {
	// In a real implementation, this would sort channels by priority
	// For now, we'll keep the order as is
}

func (s *FraudAlertingService) getChannelTypes(channels []FraudNotificationChannel) []string {
	var types []string
	for _, channel := range channels {
		types = append(types, channel.Type)
	}
	return types
}

// Publish events
func (s *FraudAlertingService) publishAlertProcessedEvent(ctx context.Context, alert *models.FraudAlert) {
	if s.messagingClient == nil {
		return
	}

	event := &messaging.Event{
		Type:   "fraud.alert.processed",
		Source: "fraud-alerting-service",
		Data: map[string]interface{}{
			"alert_id":      alert.ID.String(),
			"enterprise_id": alert.EnterpriseID.String(),
			"alert_type":    alert.AlertType,
			"severity":      alert.Severity,
			"status":        alert.Status,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := s.messagingClient.PublishEvent(ctx, event); err != nil {
		fmt.Printf("Warning: Failed to publish alert processed event: %v\n", err)
	}
}

// Placeholder implementations for other interface methods
func (s *FraudAlertingService) EscalateAlert(ctx context.Context, alertID uuid.UUID, reason string) error {
	// In a real implementation, this would escalate an alert
	return nil
}

func (s *FraudAlertingService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID uuid.UUID, notes string) error {
	alert, err := s.fraudRepo.GetFraudAlertByID(ctx, alertID)
	if err != nil {
		return err
	}
	if err := alert.TransitionTo(models.FraudAlertStatusAcknowledged); err != nil {
		return err
	}
	now := time.Now()
	alert.AcknowledgedAt = &now
	alert.UpdatedAt = now
	return s.fraudRepo.UpdateFraudAlert(ctx, alert)
}

func (s *FraudAlertingService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string, userID uuid.UUID) error {
	alert, err := s.fraudRepo.GetFraudAlertByID(ctx, alertID)
	if err != nil {
		return err
	}
	if err := alert.TransitionTo(models.FraudAlertStatusResolved); err != nil {
		return err
	}
	now := time.Now()
	alert.ResolvedAt = &now
	alert.UpdatedAt = now
	return s.fraudRepo.UpdateFraudAlert(ctx, alert)
}

func (s *FraudAlertingService) AssignAlert(ctx context.Context, alertID uuid.UUID, assignedTo uuid.UUID) error {
	alert, err := s.fraudRepo.GetFraudAlertByID(ctx, alertID)
	if err != nil {
		return err
	}
	alert.AssignedTo = &assignedTo
	alert.Status = models.FraudAlertStatusInvestigating
	alert.UpdatedAt = time.Now()
	return s.fraudRepo.UpdateFraudAlert(ctx, alert)
}

func (s *FraudAlertingService) ConfigureNotificationChannels(ctx context.Context, enterpriseID uuid.UUID, channels []FraudNotificationChannel) error {
	// In a real implementation, this would configure notification channels
	return nil
}

func (s *FraudAlertingService) GetNotificationChannels(ctx context.Context, enterpriseID uuid.UUID) ([]FraudNotificationChannel, error) {
	// In a real implementation, this would get notification channels
	return []FraudNotificationChannel{}, nil
}

func (s *FraudAlertingService) CorrelateAlerts(ctx context.Context, enterpriseID uuid.UUID, timeWindow time.Duration) ([]*AlertCorrelation, error) {
	// In a real implementation, this would correlate alerts
	return []*AlertCorrelation{}, nil
}

func (s *FraudAlertingService) DeduplicateAlerts(ctx context.Context, alerts []*models.FraudAlert) ([]*models.FraudAlert, error) {
	// In a real implementation, this would deduplicate alerts
	return alerts, nil
}

func (s *FraudAlertingService) GetEscalationRules(ctx context.Context) ([]*AlertEscalationRule, error) {
	// In a real implementation, this would get escalation rules
	return []*AlertEscalationRule{}, nil
}

func (s *FraudAlertingService) CreateEscalationRule(ctx context.Context, rule *AlertEscalationRule) error {
	// In a real implementation, this would create an escalation rule
	return nil
}

func (s *FraudAlertingService) UpdateEscalationRule(ctx context.Context, rule *AlertEscalationRule) error {
	// In a real implementation, this would update an escalation rule
	return nil
}

func (s *FraudAlertingService) DeleteEscalationRule(ctx context.Context, ruleID uuid.UUID) error {
	// In a real implementation, this would delete an escalation rule
	return nil
}
