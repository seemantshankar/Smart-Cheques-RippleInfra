package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
)

// DisputeInAppNotificationService handles real-time in-app notifications for dispute events
type DisputeInAppNotificationService struct {
	connections         map[string][]*NotificationConnection // userID -> connections
	broadcastChan       chan *NotificationMessage
	mu                  sync.RWMutex
	maxHistory          int
	notificationHistory map[string][]*NotificationMessage // userID -> notifications
}

// NotificationConnection represents a WebSocket or similar real-time connection
type NotificationConnection struct {
	ID        string
	UserID    string
	SendChan  chan *NotificationMessage
	Active    bool
	CreatedAt time.Time
}

// NotificationMessage represents a real-time notification message
type NotificationMessage struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Type       string                 `json:"type"` // dispute_created, status_changed, etc.
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	Priority   string                 `json:"priority"` // low, normal, high, urgent
	Read       bool                   `json:"read"`
	ReadAt     *time.Time             `json:"read_at,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	ActionURL  string                 `json:"action_url,omitempty"`
	ActionText string                 `json:"action_text,omitempty"`
	Category   string                 `json:"category"` // dispute, system, alert
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

// DisputeNotificationPreferences represents user dispute notification preferences
type DisputeNotificationPreferences struct {
	UserID            string          `json:"user_id"`
	EmailEnabled      bool            `json:"email_enabled"`
	WebhooksEnabled   bool            `json:"webhooks_enabled"`
	InAppEnabled      bool            `json:"in_app_enabled"`
	PushEnabled       bool            `json:"push_enabled"`
	QuietHoursStart   *time.Time      `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd     *time.Time      `json:"quiet_hours_end,omitempty"`
	Categories        map[string]bool `json:"categories"`         // category -> enabled
	PriorityThreshold string          `json:"priority_threshold"` // minimum priority to notify
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// NewDisputeInAppNotificationService creates a new in-app notification service
func NewDisputeInAppNotificationService() *DisputeInAppNotificationService {
	service := &DisputeInAppNotificationService{
		connections:         make(map[string][]*NotificationConnection),
		broadcastChan:       make(chan *NotificationMessage, 1000),
		maxHistory:          100,
		notificationHistory: make(map[string][]*NotificationMessage),
	}

	// Start the broadcaster
	go service.notificationBroadcaster()

	return service
}

// RegisterConnection registers a new real-time connection for a user
func (s *DisputeInAppNotificationService) RegisterConnection(userID string, sendChan chan *NotificationMessage) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	connectionID := uuid.New().String()
	connection := &NotificationConnection{
		ID:        connectionID,
		UserID:    userID,
		SendChan:  sendChan,
		Active:    true,
		CreatedAt: time.Now(),
	}

	if s.connections[userID] == nil {
		s.connections[userID] = make([]*NotificationConnection, 0)
	}

	s.connections[userID] = append(s.connections[userID], connection)

	// Send recent notifications to the new connection
	s.sendRecentNotifications(connection)

	return connectionID
}

// UnregisterConnection removes a connection
func (s *DisputeInAppNotificationService) UnregisterConnection(userID, connectionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	connections := s.connections[userID]
	for i, conn := range connections {
		if conn.ID == connectionID {
			conn.Active = false
			close(conn.SendChan)
			// Remove from slice
			s.connections[userID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}

	// Clean up empty connection lists
	if len(s.connections[userID]) == 0 {
		delete(s.connections, userID)
	}
}

// SendNotification sends a notification to a specific user
func (s *DisputeInAppNotificationService) SendNotification(ctx context.Context, userID string, notification *NotificationMessage) error {
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	// Add to user's notification history
	s.addToHistory(userID, notification)

	// Send to active connections
	s.broadcastChan <- notification

	return nil
}

// SendBulkNotifications sends notifications to multiple users
func (s *DisputeInAppNotificationService) SendBulkNotifications(ctx context.Context, userIDs []string, notification *NotificationMessage) error {
	for _, userID := range userIDs {
		notificationCopy := *notification // Copy the notification
		notificationCopy.UserID = userID
		if err := s.SendNotification(ctx, userID, &notificationCopy); err != nil {
			// Log error but continue
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
		}
	}
	return nil
}

// SendDisputeNotification sends a dispute-specific notification
func (s *DisputeInAppNotificationService) SendDisputeNotification(ctx context.Context, dispute *models.Dispute, recipientID, eventType string, additionalData map[string]interface{}) error {
	notification := s.createDisputeNotification(dispute, eventType, additionalData)
	return s.SendNotification(ctx, recipientID, notification)
}

// MarkAsRead marks a notification as read
func (s *DisputeInAppNotificationService) MarkAsRead(userID, notificationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	history, exists := s.notificationHistory[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	for _, notification := range history {
		if notification.ID == notificationID {
			if !notification.Read {
				now := time.Now()
				notification.Read = true
				notification.ReadAt = &now
			}
			return nil
		}
	}

	return fmt.Errorf("notification not found")
}

// MarkAllAsRead marks all notifications as read for a user
func (s *DisputeInAppNotificationService) MarkAllAsRead(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	history, exists := s.notificationHistory[userID]
	if !exists {
		return nil
	}

	now := time.Now()
	for _, notification := range history {
		if !notification.Read {
			notification.Read = true
			notification.ReadAt = &now
		}
	}

	return nil
}

// GetNotifications retrieves notifications for a user
func (s *DisputeInAppNotificationService) GetNotifications(userID string, limit, offset int, includeRead bool) ([]*NotificationMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.notificationHistory[userID]
	if !exists {
		return []*NotificationMessage{}, nil
	}

	var filtered []*NotificationMessage
	for _, notification := range history {
		if includeRead || !notification.Read {
			filtered = append(filtered, notification)
		}
	}

	// Sort by creation date (most recent first)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].CreatedAt.After(filtered[i].CreatedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Apply pagination
	start := offset
	end := offset + limit

	if start >= len(filtered) {
		return []*NotificationMessage{}, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *DisputeInAppNotificationService) GetUnreadCount(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.notificationHistory[userID]
	if !exists {
		return 0
	}

	count := 0
	for _, notification := range history {
		if !notification.Read {
			count++
		}
	}

	return count
}

// UpdateNotificationPreferences updates user notification preferences
func (s *DisputeInAppNotificationService) UpdateNotificationPreferences(userID string, preferences *DisputeNotificationPreferences) error {
	preferences.UserID = userID
	preferences.UpdatedAt = time.Now()
	if preferences.CreatedAt.IsZero() {
		preferences.CreatedAt = time.Now()
	}

	// This would typically save to a database
	// For now, just validate the preferences
	if preferences.Categories == nil {
		preferences.Categories = make(map[string]bool)
	}

	return nil
}

// GetNotificationPreferences retrieves user notification preferences
func (s *DisputeInAppNotificationService) GetNotificationPreferences(userID string) (*DisputeNotificationPreferences, error) {
	// This would typically load from a database
	// For now, return default preferences
	return &DisputeNotificationPreferences{
		UserID:            userID,
		EmailEnabled:      true,
		WebhooksEnabled:   true,
		InAppEnabled:      true,
		PushEnabled:       true,
		Categories:        map[string]bool{"dispute": true, "system": true, "alert": true},
		PriorityThreshold: "normal",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}, nil
}

// Private methods

func (s *DisputeInAppNotificationService) notificationBroadcaster() {
	for notification := range s.broadcastChan {
		s.mu.RLock()
		connections := s.connections[notification.UserID]
		s.mu.RUnlock()

		// Send to all active connections for this user
		for _, conn := range connections {
			if conn.Active {
				select {
				case conn.SendChan <- notification:
					// Successfully sent
				default:
					// Channel is full, mark as inactive
					conn.Active = false
				}
			}
		}
	}
}

func (s *DisputeInAppNotificationService) addToHistory(userID string, notification *NotificationMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.notificationHistory[userID] == nil {
		s.notificationHistory[userID] = make([]*NotificationMessage, 0)
	}

	history := s.notificationHistory[userID]

	// Add new notification
	history = append(history, notification)

	// Keep only the most recent notifications
	if len(history) > s.maxHistory {
		// Remove oldest notifications
		history = history[len(history)-s.maxHistory:]
	}

	s.notificationHistory[userID] = history
}

func (s *DisputeInAppNotificationService) sendRecentNotifications(connection *NotificationConnection) {
	s.mu.RLock()
	history := s.notificationHistory[connection.UserID]
	s.mu.RUnlock()

	// Send the 5 most recent unread notifications
	count := 0
	for i := len(history) - 1; i >= 0 && count < 5; i-- {
		notification := history[i]
		if !notification.Read {
			select {
			case connection.SendChan <- notification:
				count++
			default:
				// Channel is full
				break
			}
		}
	}
}

func (s *DisputeInAppNotificationService) createDisputeNotification(dispute *models.Dispute, eventType string, additionalData map[string]interface{}) *NotificationMessage {
	notification := &NotificationMessage{
		UserID:   "", // Will be set by caller
		Type:     eventType,
		Priority: string(dispute.Priority),
		Category: "dispute",
		Data: map[string]interface{}{
			"dispute_id": dispute.ID,
			"title":      dispute.Title,
			"status":     dispute.Status,
			"category":   dispute.Category,
		},
	}

	// Set notification content based on event type
	switch eventType {
	case "created":
		notification.Title = "New Dispute Created"
		notification.Message = fmt.Sprintf("A new dispute has been created: %s", dispute.Title)
		notification.ActionURL = fmt.Sprintf("/disputes/%s", dispute.ID)
		notification.ActionText = "View Dispute"

	case "status_changed":
		oldStatus := additionalData["old_status"]
		newStatus := additionalData["new_status"]
		notification.Title = "Dispute Status Changed"
		notification.Message = fmt.Sprintf("Dispute '%s' status changed from %s to %s", dispute.Title, oldStatus, newStatus)
		notification.ActionURL = fmt.Sprintf("/disputes/%s", dispute.ID)
		notification.ActionText = "View Details"

	case "evidence_added":
		notification.Title = "New Evidence Added"
		notification.Message = fmt.Sprintf("New evidence has been added to dispute: %s", dispute.Title)
		notification.ActionURL = fmt.Sprintf("/disputes/%s", dispute.ID)
		notification.ActionText = "Review Evidence"

	case "resolution_proposed":
		notification.Title = "Resolution Proposed"
		notification.Message = fmt.Sprintf("A resolution has been proposed for dispute: %s", dispute.Title)
		notification.ActionURL = fmt.Sprintf("/disputes/%s", dispute.ID)
		notification.ActionText = "Review Resolution"

	case "resolution_accepted":
		notification.Title = "Resolution Accepted"
		notification.Message = fmt.Sprintf("The resolution has been accepted for dispute: %s", dispute.Title)

	case "escalated":
		notification.Title = "Dispute Escalated"
		notification.Message = fmt.Sprintf("Dispute '%s' has been escalated and requires immediate attention", dispute.Title)
		notification.Priority = "urgent"
		notification.ActionURL = fmt.Sprintf("/disputes/%s", dispute.ID)
		notification.ActionText = "Handle Escalation"

	case "resolved":
		notification.Title = "Dispute Resolved"
		notification.Message = fmt.Sprintf("Dispute '%s' has been resolved", dispute.Title)

	case "closed":
		notification.Title = "Dispute Closed"
		notification.Message = fmt.Sprintf("Dispute '%s' has been closed", dispute.Title)

	default:
		notification.Title = "Dispute Update"
		notification.Message = fmt.Sprintf("Update for dispute: %s", dispute.Title)
	}

	// Add additional data
	for key, value := range additionalData {
		notification.Data[key] = value
	}

	// Set expiration for low-priority notifications
	if notification.Priority == "low" {
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days
		notification.ExpiresAt = &expiresAt
	}

	return notification
}

// CleanupExpiredNotifications removes expired notifications
func (s *DisputeInAppNotificationService) CleanupExpiredNotifications() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for userID, history := range s.notificationHistory {
		var activeNotifications []*NotificationMessage
		for _, notification := range history {
			if notification.ExpiresAt == nil || notification.ExpiresAt.After(now) {
				activeNotifications = append(activeNotifications, notification)
			}
		}
		s.notificationHistory[userID] = activeNotifications
	}
}

// GetNotificationStats returns statistics about notifications
func (s *DisputeInAppNotificationService) GetNotificationStats(userID string) (*DisputeNotificationStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.notificationHistory[userID]
	stats := &DisputeNotificationStats{
		UserID:       userID,
		TotalCount:   len(history),
		UnreadCount:  0,
		ReadCount:    0,
		UrgentCount:  0,
		DisputeCount: 0,
		SystemCount:  0,
		AlertCount:   0,
	}

	for _, notification := range history {
		if !notification.Read {
			stats.UnreadCount++
		} else {
			stats.ReadCount++
		}

		if notification.Priority == "urgent" {
			stats.UrgentCount++
		}

		switch notification.Category {
		case "dispute":
			stats.DisputeCount++
		case "system":
			stats.SystemCount++
		case "alert":
			stats.AlertCount++
		}
	}

	return stats, nil
}

// DisputeNotificationStats represents dispute notification statistics for a user
type DisputeNotificationStats struct {
	UserID       string `json:"user_id"`
	TotalCount   int    `json:"total_count"`
	UnreadCount  int    `json:"unread_count"`
	ReadCount    int    `json:"read_count"`
	UrgentCount  int    `json:"urgent_count"`
	DisputeCount int    `json:"dispute_count"`
	SystemCount  int    `json:"system_count"`
	AlertCount   int    `json:"alert_count"`
}

// GetActiveConnections returns the count of active connections
func (s *DisputeInAppNotificationService) GetActiveConnections() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int)
	for userID, connections := range s.connections {
		activeCount := 0
		for _, conn := range connections {
			if conn.Active {
				activeCount++
			}
		}
		if activeCount > 0 {
			result[userID] = activeCount
		}
	}

	return result
}

// BroadcastSystemNotification sends a system-wide notification
func (s *DisputeInAppNotificationService) BroadcastSystemNotification(title, message string, priority string) error {
	notification := &NotificationMessage{
		Type:      "system_broadcast",
		Title:     title,
		Message:   message,
		Priority:  priority,
		Category:  "system",
		CreatedAt: time.Now(),
	}

	// Send to all connected users
	s.mu.RLock()
	userIDs := make([]string, 0, len(s.connections))
	for userID := range s.connections {
		userIDs = append(userIDs, userID)
	}
	s.mu.RUnlock()

	return s.SendBulkNotifications(context.Background(), userIDs, notification)
}
