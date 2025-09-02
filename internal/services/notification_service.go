package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// NotificationServiceInterface defines the interface for notification operations
type NotificationServiceInterface interface {
	// Notification Management
	SendNotification(ctx context.Context, notification *NotificationRequest) error
	BatchSendNotifications(ctx context.Context, notifications []*NotificationRequest) error

	// Channel Management
	SendEmail(ctx context.Context, email *EmailNotification) error
	SendWebhook(ctx context.Context, webhook *WebhookNotification) error
	SendInAppNotification(ctx context.Context, inApp *InAppNotification) error

	// Subscription Management
	SubscribeToPaymentEvents(ctx context.Context) error
	UnsubscribeFromPaymentEvents(ctx context.Context) error

	// Template Management
	GetNotificationTemplate(ctx context.Context, templateType NotificationType) (*NotificationTemplate, error)
	UpdateNotificationTemplate(ctx context.Context, template *NotificationTemplate) error

	// Preference Management
	GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, preferences *NotificationPreferences) error

	// History and Analytics
	GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*NotificationHistory, error)
	GetNotificationStats(ctx context.Context, userID uuid.UUID) (*NotificationStats, error)
}

// NotificationService implements the notification service interface
type NotificationService struct {
	emailService       EmailServiceInterface
	webhookService     WebhookServiceInterface
	inAppService       InAppServiceInterface
	messagingClient    messaging.EventBus
	notificationConfig *NotificationConfig
	activeSubscribers  map[string]context.CancelFunc
	subscriberMutex    sync.RWMutex
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(
	emailService EmailServiceInterface,
	webhookService WebhookServiceInterface,
	inAppService InAppServiceInterface,
	messagingClient messaging.EventBus,
	config *NotificationConfig,
) NotificationServiceInterface {
	service := &NotificationService{
		emailService:       emailService,
		webhookService:     webhookService,
		inAppService:       inAppService,
		messagingClient:    messagingClient,
		notificationConfig: config,
		activeSubscribers:  make(map[string]context.CancelFunc),
	}

	// Auto-subscribe to payment events if enabled
	if config.AutoSubscribeToEvents {
		go func() {
			ctx := context.Background()
			if err := service.SubscribeToPaymentEvents(ctx); err != nil {
				log.Printf("Failed to auto-subscribe to payment events: %v", err)
			}
		}()
	}

	return service
}

// NotificationConfig defines configuration for the notification service
type NotificationConfig struct {
	// Service settings
	AutoSubscribeToEvents      bool          `json:"auto_subscribe_to_events"`
	MaxConcurrentNotifications int           `json:"max_concurrent_notifications"`
	RetryAttempts              int           `json:"retry_attempts"`
	RetryDelay                 time.Duration `json:"retry_delay"`

	// Email settings
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`

	// Webhook settings
	WebhookTimeout    time.Duration `json:"webhook_timeout"`
	MaxWebhookRetries int           `json:"max_webhook_retries"`

	// In-app settings
	InAppRetentionDays int `json:"in_app_retention_days"`
}

// NotificationRequest represents a notification request
type NotificationRequest struct {
	ID          uuid.UUID              `json:"id"`
	Type        NotificationType       `json:"type"`
	UserID      uuid.UUID              `json:"user_id"`
	Recipient   string                 `json:"recipient"`
	Channels    []NotificationChannel  `json:"channels"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Priority    NotificationPriority   `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypePaymentAuthorized    NotificationType = "payment.authorized"
	NotificationTypePaymentRejected      NotificationType = "payment.rejected"
	NotificationTypePaymentExecuted      NotificationType = "payment.executed"
	NotificationTypePaymentConfirmed     NotificationType = "payment.confirmed"
	NotificationTypePaymentFailed        NotificationType = "payment.failed"
	NotificationTypeMilestoneCompleted   NotificationType = "milestone.completed"
	NotificationTypeSmartChequeCreated   NotificationType = "smart_cheque.created"
	NotificationTypeSmartChequeCompleted NotificationType = "smart_cheque.completed"
)

// NotificationChannel represents notification delivery channels
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelWebhook NotificationChannel = "webhook"
	NotificationChannelInApp   NotificationChannel = "in_app"
)

// NotificationPriority represents notification priority levels
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// EmailNotification represents an email notification
type EmailNotification struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
}

// WebhookNotification represents a webhook notification
type WebhookNotification struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body"`
}

// InAppNotification represents an in-app notification
type InAppNotification struct {
	UserID  uuid.UUID              `json:"user_id"`
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Type    string                 `json:"type"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID        uuid.UUID           `json:"id"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Subject   string              `json:"subject"`
	Template  string              `json:"template"`
	IsHTML    bool                `json:"is_html"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID      uuid.UUID                                       `json:"user_id"`
	Preferences map[NotificationType]UserNotificationPreference `json:"preferences"`
	Channels    []NotificationChannel                           `json:"channels"`
	IsActive    bool                                            `json:"is_active"`
	UpdatedAt   time.Time                                       `json:"updated_at"`
}

// UserNotificationPreference represents user preference for a notification type
type UserNotificationPreference struct {
	EmailEnabled   bool `json:"email_enabled"`
	WebhookEnabled bool `json:"webhook_enabled"`
	InAppEnabled   bool `json:"in_app_enabled"`
}

// NotificationHistory represents notification delivery history
type NotificationHistory struct {
	ID          uuid.UUID           `json:"id"`
	UserID      uuid.UUID           `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Channel     NotificationChannel `json:"channel"`
	Recipient   string              `json:"recipient"`
	Status      string              `json:"status"`
	Error       string              `json:"error,omitempty"`
	SentAt      time.Time           `json:"sent_at"`
	DeliveredAt *time.Time          `json:"delivered_at,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	UserID             uuid.UUID  `json:"user_id"`
	TotalSent          int        `json:"total_sent"`
	TotalDelivered     int        `json:"total_delivered"`
	TotalFailed        int        `json:"total_failed"`
	EmailSent          int        `json:"email_sent"`
	WebhookSent        int        `json:"webhook_sent"`
	InAppSent          int        `json:"in_app_sent"`
	LastNotificationAt *time.Time `json:"last_notification_at,omitempty"`
}

// EmailServiceInterface defines the interface for email operations
type EmailServiceInterface interface {
	SendEmail(ctx context.Context, email *EmailNotification) error
}

// WebhookServiceInterface defines the interface for webhook operations
type WebhookServiceInterface interface {
	SendWebhook(ctx context.Context, webhook *WebhookNotification) error
}

// InAppServiceInterface defines the interface for in-app notification operations
type InAppServiceInterface interface {
	SendInAppNotification(ctx context.Context, notification *InAppNotification) error
}

// SendNotification sends a notification through configured channels
func (s *NotificationService) SendNotification(ctx context.Context, notification *NotificationRequest) error {
	log.Printf("Sending notification: %s to %s", notification.Type, notification.Recipient)

	// Get user preferences
	preferences, err := s.GetNotificationPreferences(ctx, notification.UserID)
	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		// Continue with default channels if preferences can't be loaded
	}

	// Send through each configured channel
	var lastError error
	for _, channel := range notification.Channels {
		if s.shouldSendToChannel(channel, notification.Type, preferences) {
			if err := s.sendToChannel(ctx, channel, notification); err != nil {
				log.Printf("Failed to send notification via %s: %v", channel, err)
				lastError = err
				// Continue with other channels even if one fails
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("failed to send notification through one or more channels: %w", lastError)
	}

	// Record notification history
	s.recordNotificationHistory(ctx, notification, "sent", "")

	log.Printf("Notification sent successfully: %s", notification.ID)
	return nil
}

// BatchSendNotifications sends multiple notifications concurrently
func (s *NotificationService) BatchSendNotifications(ctx context.Context, notifications []*NotificationRequest) error {
	log.Printf("Sending batch notifications: %d notifications", len(notifications))

	// Use semaphore to limit concurrent notifications
	semaphore := make(chan struct{}, s.notificationConfig.MaxConcurrentNotifications)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, notification := range notifications {
		wg.Add(1)
		go func(notif *NotificationRequest) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := s.SendNotification(ctx, notif); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to send notification %s: %w", notif.ID, err))
				mu.Unlock()
			}
		}(notification)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("batch notification errors: %v", errors)
	}

	log.Printf("Batch notifications completed successfully")
	return nil
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(ctx context.Context, email *EmailNotification) error {
	return s.emailService.SendEmail(ctx, email)
}

// SendWebhook sends a webhook notification
func (s *NotificationService) SendWebhook(ctx context.Context, webhook *WebhookNotification) error {
	return s.webhookService.SendWebhook(ctx, webhook)
}

// SendInAppNotification sends an in-app notification
func (s *NotificationService) SendInAppNotification(ctx context.Context, inApp *InAppNotification) error {
	return s.inAppService.SendInAppNotification(ctx, inApp)
}

// SubscribeToPaymentEvents subscribes to payment-related events
func (s *NotificationService) SubscribeToPaymentEvents(ctx context.Context) error {
	log.Printf("Subscribing to payment events")

	s.subscriberMutex.Lock()
	defer s.subscriberMutex.Unlock()

	// Subscribe to various payment events
	eventTypes := []string{
		"payment.execution.started",
		"payment.execution.completed",
		"payment.execution.failed",
		"payment.confirmed",
		"milestone.completed",
		"smartcheque.created",
		"smartcheque.completed",
	}

	for _, eventType := range eventTypes {
		if cancel, exists := s.activeSubscribers[eventType]; exists {
			cancel()
		}

		subCtx, cancel := context.WithCancel(ctx)
		s.activeSubscribers[eventType] = cancel

		go s.handleEventSubscription(subCtx, eventType)
	}

	log.Printf("Successfully subscribed to payment events")
	return nil
}

// UnsubscribeFromPaymentEvents unsubscribes from payment-related events
func (s *NotificationService) UnsubscribeFromPaymentEvents(ctx context.Context) error {
	log.Printf("Unsubscribing from payment events")

	s.subscriberMutex.Lock()
	defer s.subscriberMutex.Unlock()

	for eventType, cancel := range s.activeSubscribers {
		cancel()
		delete(s.activeSubscribers, eventType)
		log.Printf("Unsubscribed from event: %s", eventType)
	}

	return nil
}

// GetNotificationTemplate gets a notification template
func (s *NotificationService) GetNotificationTemplate(ctx context.Context, templateType NotificationType) (*NotificationTemplate, error) {
	// TODO: Implement template retrieval from repository
	// For now, return a default template
	template := &NotificationTemplate{
		ID:        uuid.New(),
		Type:      templateType,
		Channel:   NotificationChannelEmail,
		Subject:   cases.Title(language.English).String(string(templateType)) + " Notification",
		Template:  "Your {{.Type}} notification has been processed. Details: {{.Data}}",
		IsHTML:    false,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return template, nil
}

// UpdateNotificationTemplate updates a notification template
func (s *NotificationService) UpdateNotificationTemplate(ctx context.Context, template *NotificationTemplate) error {
	// TODO: Implement template update in repository
	log.Printf("Template updated: %s", template.ID)
	return nil
}

// GetNotificationPreferences gets notification preferences for a user
func (s *NotificationService) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	// TODO: Implement preference retrieval from repository
	// For now, return default preferences
	preferences := &NotificationPreferences{
		UserID: userID,
		Preferences: map[NotificationType]UserNotificationPreference{
			NotificationTypePaymentAuthorized: {
				EmailEnabled:   true,
				WebhookEnabled: false,
				InAppEnabled:   true,
			},
			NotificationTypePaymentExecuted: {
				EmailEnabled:   true,
				WebhookEnabled: true,
				InAppEnabled:   true,
			},
			NotificationTypePaymentConfirmed: {
				EmailEnabled:   false,
				WebhookEnabled: true,
				InAppEnabled:   true,
			},
		},
		Channels:  []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		IsActive:  true,
		UpdatedAt: time.Now(),
	}

	return preferences, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (s *NotificationService) UpdateNotificationPreferences(ctx context.Context, preferences *NotificationPreferences) error {
	// TODO: Implement preference update in repository
	log.Printf("Preferences updated for user: %s", preferences.UserID)
	return nil
}

// GetNotificationHistory gets notification history for a user
func (s *NotificationService) GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*NotificationHistory, error) {
	// TODO: Implement history retrieval from repository
	return []*NotificationHistory{}, nil
}

// GetNotificationStats gets notification statistics for a user
func (s *NotificationService) GetNotificationStats(ctx context.Context, userID uuid.UUID) (*NotificationStats, error) {
	// TODO: Implement stats calculation from repository
	stats := &NotificationStats{
		UserID:         userID,
		TotalSent:      0,
		TotalDelivered: 0,
		TotalFailed:    0,
		EmailSent:      0,
		WebhookSent:    0,
		InAppSent:      0,
	}

	return stats, nil
}

// Helper methods

func (s *NotificationService) shouldSendToChannel(channel NotificationChannel, notificationType NotificationType, preferences *NotificationPreferences) bool {
	if preferences == nil {
		return true // Send to all channels by default
	}

	pref, exists := preferences.Preferences[notificationType]
	if !exists {
		return true // Send by default if no specific preference
	}

	switch channel {
	case NotificationChannelEmail:
		return pref.EmailEnabled
	case NotificationChannelWebhook:
		return pref.WebhookEnabled
	case NotificationChannelInApp:
		return pref.InAppEnabled
	default:
		return false
	}
}

func (s *NotificationService) sendToChannel(ctx context.Context, channel NotificationChannel, notification *NotificationRequest) error {
	switch channel {
	case NotificationChannelEmail:
		return s.sendEmailNotification(ctx, notification)
	case NotificationChannelWebhook:
		return s.sendWebhookNotification(ctx, notification)
	case NotificationChannelInApp:
		return s.sendInAppNotification(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification channel: %s", channel)
	}
}

func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *NotificationRequest) error {
	email := &EmailNotification{
		To:      []string{notification.Recipient},
		Subject: notification.Subject,
		Body:    notification.Message,
		IsHTML:  strings.Contains(notification.Message, "<"),
	}

	return s.SendEmail(ctx, email)
}

func (s *NotificationService) sendWebhookNotification(ctx context.Context, notification *NotificationRequest) error {
	webhook := &WebhookNotification{
		URL:    notification.Recipient, // Assuming recipient is webhook URL
		Method: "POST",
		Body: map[string]interface{}{
			"type":    notification.Type,
			"subject": notification.Subject,
			"message": notification.Message,
			"data":    notification.Data,
		},
	}

	return s.SendWebhook(ctx, webhook)
}

func (s *NotificationService) sendInAppNotification(ctx context.Context, notification *NotificationRequest) error {
	inApp := &InAppNotification{
		UserID:  notification.UserID,
		Title:   notification.Subject,
		Message: notification.Message,
		Type:    string(notification.Type),
		Data:    notification.Data,
	}

	return s.SendInAppNotification(ctx, inApp)
}

func (s *NotificationService) handleEventSubscription(ctx context.Context, eventType string) {
	log.Printf("Starting event subscription handler for: %s", eventType)

	handler := func(event *messaging.Event) error {
		return s.processPaymentEvent(ctx, event)
	}

	// Subscribe to the event type
	if err := s.messagingClient.SubscribeToEvent(ctx, eventType, handler); err != nil {
		log.Printf("Failed to subscribe to event %s: %v", eventType, err)
		return
	}

	<-ctx.Done()
	log.Printf("Event subscription handler stopped for: %s", eventType)
}

func (s *NotificationService) processPaymentEvent(ctx context.Context, event *messaging.Event) error {
	log.Printf("Processing payment event: %s", event.Type)

	// Extract user ID from event data
	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("No user_id found in event data")
		return nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user_id format: %v", err)
		return nil
	}

	// Create notification request
	notification := &NotificationRequest{
		ID:        uuid.New(),
		UserID:    userUUID,
		Type:      NotificationType(event.Type),
		Recipient: userID, // This should be replaced with actual recipient info
		Channels:  []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		Subject:   s.getEventSubject(event.Type),
		Message:   s.getEventMessage(event),
		Data:      event.Data,
		Priority:  NotificationPriorityNormal,
	}

	// Send notification
	return s.SendNotification(ctx, notification)
}

func (s *NotificationService) getEventSubject(eventType string) string {
	switch eventType {
	case "payment.execution.started":
		return "Payment Execution Started"
	case "payment.execution.completed":
		return "Payment Completed Successfully"
	case "payment.execution.failed":
		return "Payment Execution Failed"
	case "payment.confirmed":
		return "Payment Confirmed on Blockchain"
	case "milestone.completed":
		return "Milestone Completed"
	case "smartcheque.created":
		return "Smart Check Created"
	case "smartcheque.completed":
		return "Smart Check Completed"
	default:
		return "Payment Notification"
	}
}

func (s *NotificationService) getEventMessage(event *messaging.Event) string {
	switch event.Type {
	case "payment.execution.started":
		return "Your payment execution has started and is being processed."
	case "payment.execution.completed":
		return "Your payment has been executed successfully."
	case "payment.execution.failed":
		return "Your payment execution has failed. Please check the details."
	case "payment.confirmed":
		return "Your payment has been confirmed on the blockchain."
	case "milestone.completed":
		return "A milestone in your Smart Check has been completed."
	case "smartcheque.created":
		return "A new Smart Check has been created."
	case "smartcheque.completed":
		return "Your Smart Check has been completed successfully."
	default:
		return "You have a new payment notification."
	}
}

func (s *NotificationService) recordNotificationHistory(ctx context.Context, notification *NotificationRequest, status, errorMsg string) {
	// Check if context is canceled
	select {
	case <-ctx.Done():
		log.Printf("Context canceled, skipping notification history recording for: %s", notification.ID)
		return
	default:
	}

	// TODO: Implement history recording in repository
	if errorMsg != "" {
		log.Printf("Notification history recorded: %s - %s (Error: %s)", notification.ID, status, errorMsg)
	} else {
		log.Printf("Notification history recorded: %s - %s", notification.ID, status)
	}
}

// EmailService implementation
type emailService struct {
	config *NotificationConfig
}

func NewEmailService(config *NotificationConfig) EmailServiceInterface {
	return &emailService{config: config}
}

func (e *emailService) SendEmail(ctx context.Context, email *EmailNotification) error {
	// SMTP email implementation
	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)

	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.config.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	if email.IsHTML {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
	} else {
		msg.WriteString("\r\n")
	}

	msg.WriteString(email.Body)

	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)
	return smtp.SendMail(addr, auth, e.config.FromEmail, email.To, msg.Bytes())
}

// WebhookService implementation
type webhookService struct {
	config *NotificationConfig
	client *http.Client
}

func NewWebhookService(config *NotificationConfig) WebhookServiceInterface {
	return &webhookService{
		config: config,
		client: &http.Client{
			Timeout: config.WebhookTimeout,
		},
	}
}

func (w *webhookService) SendWebhook(ctx context.Context, webhook *WebhookNotification) error {
	jsonData, err := json.Marshal(webhook.Body)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, webhook.Method, webhook.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// InAppService implementation
type inAppService struct {
	// TODO: Implement in-app notification storage and delivery
}

func NewInAppService() InAppServiceInterface {
	return &inAppService{}
}

func (i *inAppService) SendInAppNotification(ctx context.Context, notification *InAppNotification) error {
	// TODO: Implement in-app notification storage and real-time delivery
	log.Printf("In-app notification sent to user %s: %s", notification.UserID, notification.Title)
	return nil
}
