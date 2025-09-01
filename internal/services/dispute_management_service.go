package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// DisputeManagementService provides comprehensive dispute lifecycle management
type DisputeManagementService struct {
	disputeRepo     repository.DisputeRepositoryInterface
	smartChequeRepo repository.SmartChequeRepositoryInterface
	messagingClient *messaging.Service
	auditRepo       repository.AuditRepositoryInterface
}

// NewDisputeManagementService creates a new dispute management service
func NewDisputeManagementService(
	disputeRepo repository.DisputeRepositoryInterface,
	smartChequeRepo repository.SmartChequeRepositoryInterface,
	messagingClient *messaging.Service,
	auditRepo repository.AuditRepositoryInterface,
) *DisputeManagementService {
	return &DisputeManagementService{
		disputeRepo:     disputeRepo,
		smartChequeRepo: smartChequeRepo,
		messagingClient: messagingClient,
		auditRepo:       auditRepo,
	}
}

// CreateDisputeRequest represents the request to create a new dispute
type CreateDisputeRequest struct {
	Title       string                 `json:"title" validate:"required,min=5,max=200"`
	Description string                 `json:"description" validate:"required,min=10,max=2000"`
	Category    models.DisputeCategory `json:"category" validate:"required"`
	Priority    models.DisputePriority `json:"priority" validate:"required"`

	// Related entities
	SmartChequeID *string `json:"smart_cheque_id,omitempty"`
	MilestoneID   *string `json:"milestone_id,omitempty"`
	ContractID    *string `json:"contract_id,omitempty"`
	TransactionID *string `json:"transaction_id,omitempty"`

	// Parties involved
	InitiatorID    string `json:"initiator_id" validate:"required"`
	InitiatorType  string `json:"initiator_type" validate:"required,oneof=enterprise user"`
	RespondentID   string `json:"respondent_id" validate:"required"`
	RespondentType string `json:"respondent_type" validate:"required,oneof=enterprise user"`

	// Financial impact
	DisputedAmount *float64         `json:"disputed_amount,omitempty"`
	Currency       *models.Currency `json:"currency,omitempty"`

	// Metadata
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Evidence files (optional at creation)
	EvidenceFiles []EvidenceFileRequest `json:"evidence_files,omitempty"`
}

// EvidenceFileRequest represents evidence file information for dispute creation
type EvidenceFileRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	FileType    string `json:"file_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required"`
	FilePath    string `json:"file_path" validate:"required"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

// CreateDispute creates a new dispute with full validation and lifecycle initialization
func (s *DisputeManagementService) CreateDispute(ctx context.Context, request *CreateDisputeRequest, createdBy string) (*models.Dispute, error) {
	// Validate the request
	if err := s.validateCreateDisputeRequest(request); err != nil {
		return nil, fmt.Errorf("invalid dispute request: %w", err)
	}

	// Check if related entities exist and are valid
	if err := s.validateRelatedEntities(ctx, request); err != nil {
		return nil, fmt.Errorf("invalid related entities: %w", err)
	}

	// Create dispute model
	dispute := &models.Dispute{
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Priority:    request.Priority,
		Status:      models.DisputeStatusInitiated,

		// Related entities
		SmartChequeID: request.SmartChequeID,
		MilestoneID:   request.MilestoneID,
		ContractID:    request.ContractID,
		TransactionID: request.TransactionID,

		// Parties
		InitiatorID:    request.InitiatorID,
		InitiatorType:  request.InitiatorType,
		RespondentID:   request.RespondentID,
		RespondentType: request.RespondentType,

		// Financial impact
		DisputedAmount: request.DisputedAmount,
		Currency:       request.Currency,

		// Metadata
		Tags:     request.Tags,
		Metadata: request.Metadata,

		// Audit
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}

	// Create dispute in database
	if err := s.disputeRepo.CreateDispute(ctx, dispute); err != nil {
		return nil, fmt.Errorf("failed to create dispute: %w", err)
	}

	// Create evidence files if provided
	if len(request.EvidenceFiles) > 0 {
		for _, evidenceReq := range request.EvidenceFiles {
			evidence := &models.DisputeEvidence{
				DisputeID:   dispute.ID,
				FileName:    evidenceReq.FileName,
				FileType:    evidenceReq.FileType,
				FileSize:    evidenceReq.FileSize,
				FilePath:    evidenceReq.FilePath,
				Description: evidenceReq.Description,
				UploadedBy:  createdBy,
				IsPublic:    evidenceReq.IsPublic,
			}

			if err := s.disputeRepo.CreateEvidence(ctx, evidence); err != nil {
				log.Printf("Warning: Failed to create evidence for dispute %s: %v", dispute.ID, err)
				// Don't fail the entire operation for evidence creation issues
			}
		}
	}

	// Create audit log
	auditLog := &models.DisputeAuditLog{
		DisputeID: dispute.ID,
		Action:    "created",
		UserID:    createdBy,
		UserType:  "user", // TODO: Determine user type from context
		Details:   "Dispute created with initial evidence",
		NewValue:  map[string]interface{}{"status": dispute.Status},
	}
	if err := s.disputeRepo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Warning: Failed to create audit log for dispute %s: %v", dispute.ID, err)
	}

	// Send notifications
	if err := s.sendDisputeNotifications(ctx, dispute, "created"); err != nil {
		log.Printf("Warning: Failed to send dispute notifications for %s: %v", dispute.ID, err)
	}

	// Publish event
	if err := s.publishDisputeEvent(dispute, "dispute_created"); err != nil {
		log.Printf("Warning: Failed to publish dispute event for %s: %v", dispute.ID, err)
	}

	return dispute, nil
}

// UpdateDisputeStatus updates the status of a dispute with proper validation and notifications
func (s *DisputeManagementService) UpdateDisputeStatus(ctx context.Context, disputeID string, newStatus models.DisputeStatus, updatedBy string, reason string) error {
	// Get current dispute
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Validate status transition
	if err := s.validateStatusTransition(dispute.Status, newStatus); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	oldStatus := dispute.Status
	dispute.Status = newStatus
	dispute.UpdatedBy = updatedBy

	// Set timestamps based on status
	now := time.Now()
	if newStatus == models.DisputeStatusResolved {
		dispute.ResolvedAt = &now
	} else if newStatus == models.DisputeStatusClosed {
		dispute.ClosedAt = &now
	}

	// Update dispute
	if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
		return fmt.Errorf("failed to update dispute status: %w", err)
	}

	// Create audit log
	auditLog := &models.DisputeAuditLog{
		DisputeID: disputeID,
		Action:    "status_changed",
		UserID:    updatedBy,
		UserType:  "user", // TODO: Determine user type from context
		Details:   fmt.Sprintf("Status changed from %s to %s: %s", oldStatus, newStatus, reason),
		OldValue:  map[string]interface{}{"status": oldStatus},
		NewValue:  map[string]interface{}{"status": newStatus},
	}
	if err := s.disputeRepo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Warning: Failed to create audit log for status change: %v", err)
	}

	// Handle status-specific actions
	if err := s.handleStatusTransition(ctx, dispute, oldStatus, newStatus); err != nil {
		log.Printf("Warning: Failed to handle status transition: %v", err)
	}

	// Send notifications
	if err := s.sendDisputeNotifications(ctx, dispute, "status_changed"); err != nil {
		log.Printf("Warning: Failed to send status change notifications: %v", err)
	}

	// Publish event
	if err := s.publishDisputeEvent(dispute, "dispute_status_changed"); err != nil {
		log.Printf("Warning: Failed to publish status change event: %v", err)
	}

	return nil
}

// AddDisputeEvidence adds evidence to an existing dispute
func (s *DisputeManagementService) AddDisputeEvidence(ctx context.Context, disputeID string, evidence *models.DisputeEvidence) error {
	// Validate dispute exists
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	// Set dispute ID and create evidence
	evidence.DisputeID = disputeID
	if err := s.disputeRepo.CreateEvidence(ctx, evidence); err != nil {
		return fmt.Errorf("failed to create evidence: %w", err)
	}

	// Update dispute last activity
	dispute.LastActivityAt = time.Now()
	dispute.UpdatedBy = evidence.UploadedBy
	if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
		log.Printf("Warning: Failed to update dispute last activity: %v", err)
	}

	// Create audit log
	auditLog := &models.DisputeAuditLog{
		DisputeID: disputeID,
		Action:    "evidence_added",
		UserID:    evidence.UploadedBy,
		UserType:  "user", // TODO: Determine user type from context
		Details:   fmt.Sprintf("Evidence added: %s", evidence.FileName),
		NewValue:  map[string]interface{}{"evidence_id": evidence.ID},
	}
	if err := s.disputeRepo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Warning: Failed to create audit log for evidence: %v", err)
	}

	// Send notifications
	if err := s.sendDisputeNotifications(ctx, dispute, "evidence_added"); err != nil {
		log.Printf("Warning: Failed to send evidence notifications: %v", err)
	}

	return nil
}

// CreateDisputeResolution creates a resolution for a dispute
func (s *DisputeManagementService) CreateDisputeResolution(ctx context.Context, resolution *models.DisputeResolution, createdBy string) error {
	// Validate dispute exists and is in appropriate status
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, resolution.DisputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", resolution.DisputeID)
	}

	if dispute.Status != models.DisputeStatusUnderReview && dispute.Status != models.DisputeStatusEscalated {
		return fmt.Errorf("dispute must be in under_review or escalated status to create resolution")
	}

	// Create resolution
	resolution.CreatedBy = createdBy
	resolution.UpdatedBy = createdBy
	if err := s.disputeRepo.CreateResolution(ctx, resolution); err != nil {
		return fmt.Errorf("failed to create resolution: %w", err)
	}

	// Update dispute status to resolved
	if err := s.UpdateDisputeStatus(ctx, resolution.DisputeID, models.DisputeStatusResolved, createdBy, "Resolution created"); err != nil {
		log.Printf("Warning: Failed to update dispute status to resolved: %v", err)
	}

	// Create audit log
	auditLog := &models.DisputeAuditLog{
		DisputeID: resolution.DisputeID,
		Action:    "resolution_created",
		UserID:    createdBy,
		UserType:  "user", // TODO: Determine user type from context
		Details:   fmt.Sprintf("Resolution created using method: %s", resolution.Method),
		NewValue:  map[string]interface{}{"resolution_id": resolution.ID, "method": resolution.Method},
	}
	if err := s.disputeRepo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Warning: Failed to create audit log for resolution: %v", err)
	}

	return nil
}

// ExecuteDisputeResolution executes the approved dispute resolution
func (s *DisputeManagementService) ExecuteDisputeResolution(ctx context.Context, disputeID string, executedBy string) error {
	// Get dispute and resolution
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	resolution, err := s.disputeRepo.GetResolutionByDisputeID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("failed to get resolution: %w", err)
	}
	if resolution == nil {
		return fmt.Errorf("no resolution found for dispute: %s", disputeID)
	}

	// Check if both parties have accepted
	if !resolution.InitiatorAccepted || !resolution.RespondentAccepted {
		return fmt.Errorf("both parties must accept the resolution before execution")
	}

	// Mark resolution as executed
	now := time.Now()
	resolution.IsExecuted = true
	resolution.ExecutedAt = &now
	resolution.ExecutedBy = &executedBy
	resolution.UpdatedBy = executedBy

	if err := s.disputeRepo.UpdateResolution(ctx, resolution); err != nil {
		return fmt.Errorf("failed to update resolution: %w", err)
	}

	// Update dispute status to closed
	if err := s.UpdateDisputeStatus(ctx, disputeID, models.DisputeStatusClosed, executedBy, "Resolution executed"); err != nil {
		log.Printf("Warning: Failed to close dispute: %v", err)
	}

	// Execute the actual resolution (fund transfers, etc.)
	if err := s.executeResolutionOutcome(ctx, dispute, resolution); err != nil {
		log.Printf("Warning: Failed to execute resolution outcome: %v", err)
		// Don't fail the entire operation - resolution is still marked as executed
	}

	return nil
}

// Helper methods

func (s *DisputeManagementService) validateCreateDisputeRequest(request *CreateDisputeRequest) error {
	if request.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(request.Title) < 5 || len(request.Title) > 200 {
		return fmt.Errorf("title must be between 5 and 200 characters")
	}
	if request.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(request.Description) < 10 || len(request.Description) > 2000 {
		return fmt.Errorf("description must be between 10 and 2000 characters")
	}
	if request.InitiatorID == "" || request.RespondentID == "" {
		return fmt.Errorf("initiator and respondent IDs are required")
	}
	if request.InitiatorID == request.RespondentID {
		return fmt.Errorf("initiator and respondent cannot be the same")
	}
	return nil
}

func (s *DisputeManagementService) validateRelatedEntities(ctx context.Context, request *CreateDisputeRequest) error {
	// Validate SmartCheque if provided
	if request.SmartChequeID != nil {
		smartCheque, err := s.smartChequeRepo.GetSmartChequeByID(ctx, *request.SmartChequeID)
		if err != nil {
			return fmt.Errorf("failed to validate smart cheque: %w", err)
		}
		if smartCheque == nil {
			return fmt.Errorf("smart cheque not found: %s", *request.SmartChequeID)
		}
		// Validate that initiator/respondent match smart cheque parties
		if smartCheque.PayerID != request.InitiatorID && smartCheque.PayerID != request.RespondentID {
			return fmt.Errorf("initiator/respondent must be a party to the smart cheque")
		}
		if smartCheque.PayeeID != request.InitiatorID && smartCheque.PayeeID != request.RespondentID {
			return fmt.Errorf("initiator/respondent must be a party to the smart cheque")
		}
	}
	return nil
}

func (s *DisputeManagementService) validateStatusTransition(currentStatus, newStatus models.DisputeStatus) error {
	validTransitions := map[models.DisputeStatus][]models.DisputeStatus{
		models.DisputeStatusInitiated:   {models.DisputeStatusUnderReview, models.DisputeStatusCancelled},
		models.DisputeStatusUnderReview: {models.DisputeStatusEscalated, models.DisputeStatusResolved, models.DisputeStatusClosed, models.DisputeStatusCancelled},
		models.DisputeStatusEscalated:   {models.DisputeStatusUnderReview, models.DisputeStatusResolved, models.DisputeStatusClosed, models.DisputeStatusCancelled},
		models.DisputeStatusResolved:    {models.DisputeStatusClosed},
		models.DisputeStatusClosed:      {}, // Terminal state
		models.DisputeStatusCancelled:   {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
}

func (s *DisputeManagementService) handleStatusTransition(ctx context.Context, dispute *models.Dispute, oldStatus, newStatus models.DisputeStatus) error {
	switch newStatus {
	case models.DisputeStatusUnderReview:
		// Notify relevant parties that review has started
		return s.sendDisputeNotifications(ctx, dispute, "review_started")
	case models.DisputeStatusEscalated:
		// Escalate to higher authority or mediation
		return s.handleEscalation(ctx, dispute)
	case models.DisputeStatusResolved:
		// Resolution process initiated
		return s.sendDisputeNotifications(ctx, dispute, "resolved")
	case models.DisputeStatusClosed:
		// Final closure notifications
		return s.sendDisputeNotifications(ctx, dispute, "closed")
	}
	return nil
}

func (s *DisputeManagementService) handleEscalation(ctx context.Context, dispute *models.Dispute) error {
	// Implementation for escalation logic (e.g., assign to mediator, notify authorities, etc.)
	log.Printf("Handling escalation for dispute %s", dispute.ID)
	return nil
}

func (s *DisputeManagementService) executeResolutionOutcome(ctx context.Context, dispute *models.Dispute, resolution *models.DisputeResolution) error {
	// Implementation for executing resolution outcomes (fund transfers, status updates, etc.)
	log.Printf("Executing resolution outcome for dispute %s", dispute.ID)
	return nil
}

func (s *DisputeManagementService) sendDisputeNotifications(ctx context.Context, dispute *models.Dispute, eventType string) error {
	// Implementation for sending notifications to relevant parties
	log.Printf("Sending %s notifications for dispute %s", eventType, dispute.ID)
	return nil
}

func (s *DisputeManagementService) publishDisputeEvent(dispute *models.Dispute, eventType string) error {
	if s.messagingClient == nil {
		return nil
	}

	event := &messaging.Event{
		Type:   eventType,
		Source: "dispute_management_service",
		Data: map[string]interface{}{
			"dispute_id": dispute.ID,
			"title":      dispute.Title,
			"status":     dispute.Status,
			"category":   dispute.Category,
			"priority":   dispute.Priority,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return s.messagingClient.PublishEvent(event)
}

// Legacy methods for backward compatibility

// InitiateMilestoneDispute initiates a dispute for a milestone (legacy method)
func (s *DisputeManagementService) InitiateMilestoneDispute(ctx context.Context, milestoneID string, reason string) error {
	request := &CreateDisputeRequest{
		Title:          fmt.Sprintf("Milestone Dispute: %s", milestoneID),
		Description:    reason,
		Category:       models.DisputeCategoryMilestone,
		Priority:       models.DisputePriorityNormal,
		MilestoneID:    &milestoneID,
		InitiatorID:    "system", // TODO: Get from context
		InitiatorType:  "user",
		RespondentID:   "system", // TODO: Get from milestone/contract
		RespondentType: "user",
	}

	_, err := s.CreateDispute(ctx, request, "system")
	return err
}

// HoldMilestoneFunds holds funds when dispute is initiated (legacy method)
func (s *DisputeManagementService) HoldMilestoneFunds(ctx context.Context, milestoneID string) error {
	// Get dispute by milestone ID
	disputes, err := s.disputeRepo.GetDisputesByMilestone(ctx, milestoneID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to get disputes for milestone: %w", err)
	}
	if len(disputes) == 0 {
		return fmt.Errorf("no active dispute found for milestone: %s", milestoneID)
	}

	dispute := disputes[0]

	// Update dispute status to hold funds
	return s.UpdateDisputeStatus(ctx, dispute.ID, models.DisputeStatusUnderReview, "system", "Funds held due to dispute")
}

// ExecuteDisputeResolutionLegacy executes the dispute resolution workflow (legacy method)
func (s *DisputeManagementService) ExecuteDisputeResolutionLegacy(ctx context.Context, disputeID string) error {
	return s.ExecuteDisputeResolution(ctx, disputeID, "system")
}

// EnforceDisputeOutcome enforces the outcome of dispute resolution (legacy method)
func (s *DisputeManagementService) EnforceDisputeOutcome(ctx context.Context, disputeID string, outcome string) error {
	// Create resolution if it doesn't exist
	resolution := &models.DisputeResolution{
		DisputeID:          disputeID,
		Method:             models.DisputeResolutionMethodMutualAgreement,
		ResolutionDetails:  outcome,
		OutcomeDescription: outcome,
		InitiatorAccepted:  true,
		RespondentAccepted: true,
	}

	return s.CreateDisputeResolution(ctx, resolution, "system")
}
