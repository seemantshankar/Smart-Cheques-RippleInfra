package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
)

// DisputeEmailService handles email notifications for dispute events
type DisputeEmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromAddress  string
	baseURL      string
	templates    map[string]*template.Template
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// NewDisputeEmailService creates a new email service instance
func NewDisputeEmailService(smtpHost, smtpPort, smtpUsername, smtpPassword, fromAddress, baseURL string) *DisputeEmailService {
	service := &DisputeEmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		baseURL:      baseURL,
		templates:    make(map[string]*template.Template),
	}

	service.initializeTemplates()
	return service
}

// SendDisputeNotification sends an email notification for a dispute event
func (s *DisputeEmailService) SendDisputeNotification(ctx context.Context, dispute *models.Dispute, recipientEmail, eventType string, additionalData map[string]interface{}) error {
	// Prepare template data
	templateData := s.prepareTemplateData(dispute, eventType, additionalData)

	// Get the appropriate template
	tmpl, exists := s.templates[eventType]
	if !exists {
		return fmt.Errorf("email template not found for event type: %s", eventType)
	}

	// Render the email
	var subjectBuf, htmlBuf, textBuf bytes.Buffer

	// Render subject
	subjectTmpl := template.Must(template.New("subject").Parse(s.getSubjectTemplate(eventType)))
	if err := subjectTmpl.Execute(&subjectBuf, templateData); err != nil {
		return fmt.Errorf("failed to render email subject: %w", err)
	}

	// Render HTML body
	if err := tmpl.ExecuteTemplate(&htmlBuf, "html", templateData); err != nil {
		return fmt.Errorf("failed to render HTML email body: %w", err)
	}

	// Render text body
	textTmpl := template.Must(template.New("text").Parse(s.getTextTemplate(eventType)))
	if err := textTmpl.Execute(&textBuf, templateData); err != nil {
		return fmt.Errorf("failed to render text email body: %w", err)
	}

	// Send the email
	return s.sendEmail(recipientEmail, subjectBuf.String(), htmlBuf.String(), textBuf.String())
}

// SendBulkDisputeNotifications sends notifications to multiple recipients
func (s *DisputeEmailService) SendBulkDisputeNotifications(ctx context.Context, dispute *models.Dispute, recipientEmails []string, eventType string, additionalData map[string]interface{}) error {
	for _, email := range recipientEmails {
		if err := s.SendDisputeNotification(ctx, dispute, email, eventType, additionalData); err != nil {
			// Log error but continue with other recipients
			fmt.Printf("Failed to send email to %s: %v\n", email, err)
		}
		// Add small delay to avoid overwhelming SMTP server
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// SendDisputeStatusUpdate sends a status update notification
func (s *DisputeEmailService) SendDisputeStatusUpdate(ctx context.Context, dispute *models.Dispute, recipientEmail, oldStatus, newStatus, reason string) error {
	additionalData := map[string]interface{}{
		"old_status": oldStatus,
		"new_status": newStatus,
		"reason":     reason,
	}
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "status_changed", additionalData)
}

// SendEvidenceAddedNotification sends notification when evidence is added
func (s *DisputeEmailService) SendEvidenceAddedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail, evidenceFileName string) error {
	additionalData := map[string]interface{}{
		"evidence_file_name": evidenceFileName,
	}
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "evidence_added", additionalData)
}

// SendResolutionProposedNotification sends notification when resolution is proposed
func (s *DisputeEmailService) SendResolutionProposedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail, resolutionMethod string) error {
	additionalData := map[string]interface{}{
		"resolution_method": resolutionMethod,
	}
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "resolution_proposed", additionalData)
}

// SendResolutionAcceptedNotification sends notification when resolution is accepted
func (s *DisputeEmailService) SendResolutionAcceptedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail string) error {
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "resolution_accepted", nil)
}

// SendDisputeEscalatedNotification sends notification when dispute is escalated
func (s *DisputeEmailService) SendDisputeEscalatedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail string) error {
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "escalated", nil)
}

// SendDisputeResolvedNotification sends notification when dispute is resolved
func (s *DisputeEmailService) SendDisputeResolvedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail string) error {
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "resolved", nil)
}

// SendDisputeClosedNotification sends notification when dispute is closed
func (s *DisputeEmailService) SendDisputeClosedNotification(ctx context.Context, dispute *models.Dispute, recipientEmail string) error {
	return s.SendDisputeNotification(ctx, dispute, recipientEmail, "closed", nil)
}

// Private methods

func (s *DisputeEmailService) initializeTemplates() {
	// Initialize HTML templates
	s.templates["created"] = template.Must(template.New("created").Parse(`
{{define "html"}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Dispute Created</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .content { margin: 20px 0; }
        .dispute-details { background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .action-button { display: inline-block; background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>New Dispute Created</h1>
        <p>A new dispute has been initiated and requires your attention.</p>
    </div>

    <div class="content">
        <h2>Dispute Details</h2>
        <div class="dispute-details">
            <p><strong>Title:</strong> {{.DisputeTitle}}</p>
            <p><strong>Category:</strong> {{.DisputeCategory}}</p>
            <p><strong>Priority:</strong> {{.DisputePriority}}</p>
            <p><strong>Description:</strong> {{.DisputeDescription}}</p>
            {{if .DisputedAmount}}<p><strong>Disputed Amount:</strong> ${{.DisputedAmount}}</p>{{end}}
            <p><strong>Initiated:</strong> {{.InitiatedAt}}</p>
        </div>

        <p>Please review this dispute and take appropriate action.</p>

        <a href="{{.DisputeURL}}" class="action-button">View Dispute Details</a>

        <div class="footer">
            <p>This is an automated notification from the Smart Payment Infrastructure dispute resolution system.</p>
            <p>If you have any questions, please contact support.</p>
        </div>
    </div>
</body>
</html>
{{end}}
`))

	s.templates["status_changed"] = template.Must(template.New("status_changed").Parse(`
{{define "html"}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Dispute Status Updated</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .status-change { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .old-status { color: #dc3545; }
        .new-status { color: #28a745; }
        .action-button { display: inline-block; background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Dispute Status Updated</h1>
        <p>The status of dispute #{{.DisputeID}} has been changed.</p>
    </div>

    <div class="status-change">
        <p><strong>Status Change:</strong></p>
        <p><span class="old-status">{{.OldStatus}}</span> → <span class="new-status">{{.NewStatus}}</span></p>
        {{if .Reason}}<p><strong>Reason:</strong> {{.Reason}}</p>{{end}}
    </div>

    <div class="content">
        <h2>Dispute Summary</h2>
        <p><strong>Title:</strong> {{.DisputeTitle}}</p>
        <p><strong>Category:</strong> {{.DisputeCategory}}</p>

        <a href="{{.DisputeURL}}" class="action-button">View Full Details</a>
    </div>
</body>
</html>
{{end}}
`))
}

func (s *DisputeEmailService) getSubjectTemplate(eventType string) string {
	switch eventType {
	case DisputeEventCreated:
		return "New Dispute Created: {{.DisputeTitle}}"
	case DisputeEventStatusChanged:
		return "Dispute Status Changed: {{.DisputeTitle}}"
	case DisputeEventEvidenceAdded:
		return "New Evidence Added to Dispute: {{.DisputeTitle}}"
	case DisputeEventResolutionProposed:
		return "Resolution Proposed for Dispute: {{.DisputeTitle}}"
	case DisputeEventResolutionAccepted:
		return "Resolution Accepted for Dispute: {{.DisputeTitle}}"
	case DisputeEventEscalated:
		return "Dispute Escalated: {{.DisputeTitle}}"
	case DisputeEventResolved:
		return "Dispute Resolved: {{.DisputeTitle}}"
	case DisputeEventClosed:
		return "Dispute Closed: {{.DisputeTitle}}"
	default:
		return "Dispute Update: {{.DisputeTitle}}"
	}
}

func (s *DisputeEmailService) getTextTemplate(eventType string) string {
	switch eventType {
	case "created":
		return `A new dispute has been created.

Title: {{.DisputeTitle}}
Category: {{.DisputeCategory}}
Priority: {{.DisputePriority}}
{{if .DisputedAmount}}Disputed Amount: ${{.DisputedAmount}}{{end}}

Description: {{.DisputeDescription}}

Please review this dispute at: {{.DisputeURL}}`
	case "status_changed":
		return `Dispute status has been updated.

Title: {{.DisputeTitle}}
Status Change: {{.OldStatus}} → {{.NewStatus}}
{{if .Reason}}Reason: {{.Reason}}{{end}}

View details at: {{.DisputeURL}}`
	default:
		return `Dispute update for: {{.DisputeTitle}}

View details at: {{.DisputeURL}}`
	}
}

func (s *DisputeEmailService) prepareTemplateData(dispute *models.Dispute, _ string, additionalData map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"DisputeID":          dispute.ID,
		"DisputeTitle":       dispute.Title,
		"DisputeCategory":    strings.Title(strings.ReplaceAll(string(dispute.Category), "_", " ")),
		"DisputePriority":    strings.Title(string(dispute.Priority)),
		"DisputeDescription": dispute.Description,
		"DisputeURL":         fmt.Sprintf("%s/disputes/%s", s.baseURL, dispute.ID),
		"InitiatedAt":        dispute.InitiatedAt.Format("January 2, 2006 at 3:04 PM"),
		"LastActivityAt":     dispute.LastActivityAt.Format("January 2, 2006 at 3:04 PM"),
	}

	if dispute.DisputedAmount != nil {
		data["DisputedAmount"] = fmt.Sprintf("%.2f", *dispute.DisputedAmount)
	}

	// Add additional data
	for key, value := range additionalData {
		data[key] = value
	}

	return data
}

func (s *DisputeEmailService) sendEmail(to, subject, htmlBody, textBody string) error {
	// Prepare the email message
	var message bytes.Buffer

	// Email headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", s.fromAddress))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
	message.WriteString("\r\n")

	// Text part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML part
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")
	message.WriteString("--boundary123--")

	// SMTP authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.fromAddress, []string{to}, message.Bytes())
}

// EmailDeliveryResult represents the result of an email delivery attempt
type EmailDeliveryResult struct {
	MessageID   string     `json:"message_id"`
	Recipient   string     `json:"recipient"`
	Status      string     `json:"status"` // sent, failed, bounced
	SentAt      time.Time  `json:"sent_at"`
	ErrorMsg    *string    `json:"error_msg,omitempty"`
	RetryCount  int        `json:"retry_count"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`
}

// TrackEmailDelivery tracks email delivery status (for future implementation)
func (s *DisputeEmailService) TrackEmailDelivery(result *EmailDeliveryResult) error {
	// This would integrate with email delivery tracking systems
	// For now, just log the result
	fmt.Printf("Email delivery tracked: %s to %s - Status: %s\n",
		result.MessageID, result.Recipient, result.Status)
	return nil
}

// ValidateEmailConfiguration validates SMTP configuration
func (s *DisputeEmailService) ValidateEmailConfiguration() error {
	if s.smtpHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if s.smtpPort == "" {
		return fmt.Errorf("SMTP port is required")
	}
	if s.fromAddress == "" {
		return fmt.Errorf("from address is required")
	}
	return nil
}

// GetEmailTemplatePreview generates a preview of an email template
func (s *DisputeEmailService) GetEmailTemplatePreview(eventType string, sampleData map[string]interface{}) (string, error) {
	tmpl, exists := s.templates[eventType]
	if !exists {
		return "", fmt.Errorf("email template not found for event type: %s", eventType)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "html", sampleData); err != nil {
		return "", fmt.Errorf("failed to render template preview: %w", err)
	}

	return buf.String(), nil
}
