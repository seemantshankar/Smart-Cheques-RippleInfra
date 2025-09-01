package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// ResolutionRoute represents the suggested path to resolve a dispute
type ResolutionRoute struct {
	Method       models.DisputeResolutionMethod `json:"method"`
	Confidence   float64                        `json:"confidence"`
	Reasons      []string                       `json:"reasons"`
	NextActions  []string                       `json:"next_actions"`
	SLABy        *time.Time                     `json:"sla_by,omitempty"`
	Alternatives []AlternativeRoute             `json:"alternatives,omitempty"`
}

// AlternativeRoute represents an alternative method with a score and reason
type AlternativeRoute struct {
	Method models.DisputeResolutionMethod `json:"method"`
	Score  float64                        `json:"score"`
	Reason string                         `json:"reason"`
}

// ResolutionStrategy allows pluggable decision heuristics
type ResolutionStrategy interface {
	Name() string
	// Evaluate returns (applicable, score, method, reason)
	Evaluate(dispute *models.Dispute, categorization *CategorizationResult) (bool, float64, models.DisputeResolutionMethod, string)
}

// ResolutionRoutingService selects an appropriate resolution route using multiple strategies
type ResolutionRoutingService struct {
	disputeRepo    repository.DisputeRepositoryInterface
	categorization *DisputeCategorizationService
	strategies     []ResolutionStrategy
	methodStats    map[models.DisputeResolutionMethod]*methodPerformance
}

type methodPerformance struct {
	TotalCount     int64
	SuccessCount   int64
	TotalCycleDays float64
}

// NewResolutionRoutingService constructs the routing service
func NewResolutionRoutingService(
	disputeRepo repository.DisputeRepositoryInterface,
	categorization *DisputeCategorizationService,
) *ResolutionRoutingService {
	service := &ResolutionRoutingService{
		disputeRepo:    disputeRepo,
		categorization: categorization,
		strategies:     make([]ResolutionStrategy, 0),
		methodStats:    make(map[models.DisputeResolutionMethod]*methodPerformance),
	}

	// Register default strategies (ordered by specificity)
	service.RegisterStrategy(&fraudCourtStrategy{})
	service.RegisterStrategy(&highAmountArbitrationStrategy{})
	service.RegisterStrategy(&lowAmountMutualAgreementStrategy{})
	service.RegisterStrategy(&defaultMediationStrategy{})

	return service
}

// RegisterStrategy adds a strategy to the pipeline
func (s *ResolutionRoutingService) RegisterStrategy(strategy ResolutionStrategy) {
	s.strategies = append(s.strategies, strategy)
}

// SuggestRoute evaluates available strategies and returns the best route for a dispute ID
func (s *ResolutionRoutingService) SuggestRoute(ctx context.Context, disputeID string) (*ResolutionRoute, error) {
	if s.disputeRepo == nil {
		return nil, fmt.Errorf("dispute repository not configured")
	}
	if s.categorization == nil {
		return nil, fmt.Errorf("categorization service not configured")
	}

	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to load dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found: %s", disputeID)
	}

	return s.SuggestForDispute(ctx, dispute)
}

// SuggestForDispute evaluates strategies for a loaded dispute
func (s *ResolutionRoutingService) SuggestForDispute(ctx context.Context, dispute *models.Dispute) (*ResolutionRoute, error) {
	// Reuse categorization pipeline to get suggested method and rich signals
	categorization, err := s.categorization.AutoCategorizeDispute(ctx, dispute)
	if err != nil {
		return nil, fmt.Errorf("categorization failed: %w", err)
	}

	// Collect candidate suggestions from strategies
	type candidate struct {
		name   string
		score  float64
		method models.DisputeResolutionMethod
		reason string
	}
	candidates := make([]candidate, 0, len(s.strategies)+1)

	// Baseline candidate from categorization suggestion
	candidates = append(candidates, candidate{
		name:   "categorization_suggested",
		score:  clamp01(0.5 + categorization.Confidence*0.5), // weight confidence into baseline
		method: categorization.SuggestedMethod,
		reason: "Baseline suggestion from categorization pipeline",
	})

	for _, strategy := range s.strategies {
		if applicable, score, method, reason := strategy.Evaluate(dispute, categorization); applicable {
			candidates = append(candidates, candidate{name: strategy.Name(), score: score, method: method, reason: reason})
		}
	}

	// Additional analysis for complexity and other adjustments
	analysis := s.categorization.contentAnalyzer.AnalyzeContent(dispute.Title, dispute.Description)
	evidenceCount := 0
	if s.disputeRepo != nil && dispute.ID != "" {
		if evidence, err := s.disputeRepo.GetEvidenceByDisputeID(ctx, dispute.ID); err == nil {
			evidenceCount = len(evidence)
		}
	}
	partiesCount := 2
	if dispute.Metadata != nil {
		if v, ok := dispute.Metadata["parties_count"].(float64); ok && v > 0 {
			partiesCount = int(v)
		}
	}
	jurisdiction := ""
	if dispute.Metadata != nil {
		if v, ok := dispute.Metadata["jurisdiction"].(string); ok {
			jurisdiction = strings.ToUpper(strings.TrimSpace(v))
		}
	}

	for i := range candidates {
		base := candidates[i].score
		if analysis != nil && analysis.ComplexityScore > 0.7 {
			switch candidates[i].method {
			case models.DisputeResolutionMethodArbitration:
				base += 0.06
			case models.DisputeResolutionMethodMediation:
				base += 0.04
			case models.DisputeResolutionMethodCourt:
				base += 0.03
			}
		}
		if partiesCount > 2 || evidenceCount >= 5 {
			switch candidates[i].method {
			case models.DisputeResolutionMethodMediation, models.DisputeResolutionMethodArbitration:
				base += 0.05
			}
		}
		if jurisdiction != "" {
			switch jurisdiction {
			case "EU", "IN":
				if candidates[i].method == models.DisputeResolutionMethodCourt {
					base -= 0.05
				}
				if candidates[i].method == models.DisputeResolutionMethodMediation || candidates[i].method == models.DisputeResolutionMethodArbitration {
					base += 0.05
				}
			}
		}
		if stats, ok := s.methodStats[candidates[i].method]; ok && stats.TotalCount > 10 {
			successRate := float64(stats.SuccessCount) / float64(stats.TotalCount)
			base += 0.06 * (successRate - 0.5)
			avgDays := stats.TotalCycleDays / float64(stats.TotalCount)
			if avgDays > 0 {
				base -= 0.01 * (avgDays / 30.0)
			}
		}
		cost, timeCost := methodCostTime(candidates[i].method)
		base -= 0.10*cost + 0.08*timeCost
		candidates[i].score = clamp01(base)
	}

	// Pick best candidate by score
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	best := candidates[0]

	// Compose response
	reasons := []string{fmt.Sprintf("selected=%s score=%.2f method=%s", best.name, best.score, best.method)}
	reasons = append(reasons, categorization.MatchedRules...)

	// Compute SLA next action date using the categorization config helper
	var slaBy *time.Time
	if s.categorization != nil {
		next := s.categorization.computeNextPriorityReview(dispute)
		slaBy = next
	}

	// Build alternatives locally to avoid leaking the local candidate type
	alts := make([]AlternativeRoute, 0, len(candidates))
	for _, c := range candidates {
		if c.method == best.method {
			continue
		}
		alts = append(alts, AlternativeRoute{Method: c.method, Score: clamp01(c.score), Reason: c.reason})
	}

	route := &ResolutionRoute{
		Method:       best.method,
		Confidence:   clamp01(best.score),
		Reasons:      reasons,
		NextActions:  buildNextActions(best.method, dispute),
		SLABy:        slaBy,
		Alternatives: alts,
	}
	return route, nil
}

// --- Default strategies ---

type lowAmountMutualAgreementStrategy struct{}

func (s *lowAmountMutualAgreementStrategy) Name() string { return "low_amount_mutual_agreement" }
func (s *lowAmountMutualAgreementStrategy) Evaluate(d *models.Dispute, _ *CategorizationResult) (bool, float64, models.DisputeResolutionMethod, string) {
	if d.DisputedAmount != nil && *d.DisputedAmount < 1000 {
		return true, 0.85, models.DisputeResolutionMethodMutualAgreement, "Low amount < 1000 suggests mutual agreement"
	}
	return false, 0, "", ""
}

type highAmountArbitrationStrategy struct{}

func (s *highAmountArbitrationStrategy) Name() string { return "high_amount_arbitration" }
func (s *highAmountArbitrationStrategy) Evaluate(d *models.Dispute, c *CategorizationResult) (bool, float64, models.DisputeResolutionMethod, string) {
	if d.DisputedAmount != nil && *d.DisputedAmount > 50000 {
		return true, 0.8, models.DisputeResolutionMethodArbitration, "High amount > 50k favors arbitration"
	}
	if c != nil && c.SuggestedCategory == models.DisputeCategoryContractBreach {
		return true, 0.72, models.DisputeResolutionMethodArbitration, "Contract breach category favors arbitration"
	}
	return false, 0, "", ""
}

type fraudCourtStrategy struct{}

func (s *fraudCourtStrategy) Name() string { return "fraud_court" }
func (s *fraudCourtStrategy) Evaluate(_ *models.Dispute, c *CategorizationResult) (bool, float64, models.DisputeResolutionMethod, string) {
	if c != nil && c.SuggestedCategory == models.DisputeCategoryFraud {
		return true, 0.9, models.DisputeResolutionMethodCourt, "Fraud disputes often require legal action"
	}
	return false, 0, "", ""
}

type defaultMediationStrategy struct{}

func (s *defaultMediationStrategy) Name() string { return "default_mediation" }
func (s *defaultMediationStrategy) Evaluate(_ *models.Dispute, _ *CategorizationResult) (bool, float64, models.DisputeResolutionMethod, string) {
	return true, 0.6, models.DisputeResolutionMethodMediation, "Default to mediation when no higher-signal strategy applies"
}

// --- Helpers ---

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func buildNextActions(method models.DisputeResolutionMethod, _ *models.Dispute) []string {
	switch method {
	case models.DisputeResolutionMethodMutualAgreement:
		return []string{"Invite both parties to negotiate", "Set acceptance deadline", "Record agreed terms"}
	case models.DisputeResolutionMethodMediation:
		return []string{"Assign mediator", "Schedule mediation session", "Collect statements/evidence"}
	case models.DisputeResolutionMethodArbitration:
		return []string{"Select arbitration provider", "Notify parties", "Prepare arbitration briefs"}
	case models.DisputeResolutionMethodCourt:
		return []string{"Engage legal counsel", "Preserve evidence", "Initiate legal proceedings if required"}
	case models.DisputeResolutionMethodAdministrative:
		return []string{"Route to administrative review board", "Notify compliance", "Prepare admin findings"}
	default:
		return []string{"Review dispute details", "Confirm parties and scope"}
	}
}

func methodCostTime(m models.DisputeResolutionMethod) (float64, float64) {
	switch m {
	case models.DisputeResolutionMethodMutualAgreement:
		return 0.10, 0.10
	case models.DisputeResolutionMethodMediation:
		return 0.30, 0.30
	case models.DisputeResolutionMethodArbitration:
		return 0.60, 0.60
	case models.DisputeResolutionMethodCourt:
		return 0.90, 0.90
	case models.DisputeResolutionMethodAdministrative:
		return 0.40, 0.50
	default:
		return 0.50, 0.50
	}
}

// OverrideResolutionMethod allows manual override request for resolution method with audit logging
func (s *ResolutionRoutingService) OverrideResolutionMethod(ctx context.Context, disputeID string, method models.DisputeResolutionMethod, reason string, userID string) error {
	if s.disputeRepo == nil {
		return fmt.Errorf("dispute repository not configured")
	}
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return err
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}
	if dispute.Metadata == nil {
		dispute.Metadata = make(map[string]interface{})
	}
	dispute.Metadata["resolution_method_override"] = map[string]interface{}{
		"requested_method": string(method),
		"reason":           reason,
		"requested_by":     userID,
		"requested_at":     time.Now().Format(time.RFC3339),
		"status":           "pending",
	}
	if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
		return err
	}
	_ = s.disputeRepo.CreateAuditLog(ctx, &models.DisputeAuditLog{
		DisputeID: disputeID,
		Action:    "resolution_method_override_requested",
		UserID:    userID,
		UserType:  "user",
		Details:   fmt.Sprintf("requested_method=%s reason=%s", method, reason),
		CreatedAt: time.Now(),
	})
	return nil
}

// ApproveResolutionMethodOverride approves or rejects a resolution method override request with audit
func (s *ResolutionRoutingService) ApproveResolutionMethodOverride(ctx context.Context, disputeID string, approverID string, approve bool, approvalReason string) error {
	if s.disputeRepo == nil {
		return fmt.Errorf("dispute repository not configured")
	}
	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return err
	}
	if dispute == nil {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}
	if dispute.Metadata == nil {
		dispute.Metadata = make(map[string]interface{})
	}
	status := "rejected"
	if approve {
		status = "approved"
	}
	dispute.Metadata["resolution_method_override_status"] = status
	dispute.Metadata["resolution_method_override_reviewed_by"] = approverID
	dispute.Metadata["resolution_method_override_reviewed_at"] = time.Now().Format(time.RFC3339)
	dispute.Metadata["resolution_method_override_review_notes"] = approvalReason
	if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
		return err
	}
	_ = s.disputeRepo.CreateAuditLog(ctx, &models.DisputeAuditLog{
		DisputeID: disputeID,
		Action:    "resolution_method_override_reviewed",
		UserID:    approverID,
		UserType:  "admin",
		Details:   fmt.Sprintf("status=%s notes=%s", status, approvalReason),
		CreatedAt: time.Now(),
	})
	return nil
}

// UpdateResolutionPerformance updates in-memory performance metrics for routing decisions
func (s *ResolutionRoutingService) UpdateResolutionPerformance(method models.DisputeResolutionMethod, success bool, cycleDays float64) {
	mp, ok := s.methodStats[method]
	if !ok {
		mp = &methodPerformance{}
		s.methodStats[method] = mp
	}
	mp.TotalCount++
	if success {
		mp.SuccessCount++
	}
	if cycleDays > 0 {
		mp.TotalCycleDays += cycleDays
	}
}
