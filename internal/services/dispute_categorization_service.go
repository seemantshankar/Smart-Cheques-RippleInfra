package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// DisputeCategorizationService provides intelligent categorization and priority assignment for disputes
type DisputeCategorizationService struct {
	// Categorization rules and patterns
	categoryRules   map[models.DisputeCategory][]CategorizationRule
	priorityRules   []PriorityRule
	resolutionRules []ResolutionMethodRule
	// Enhanced content analysis
	contentAnalyzer *ContentAnalyzer
	// Dynamic rules engine
	ruleService *CategorizationRuleService
	// Optional repository for audit logging
	disputeRepo repository.DisputeRepositoryInterface
	// Configurable thresholds for confidence and fallbacks
	config *CategorizationConfig
}

// CategorizationRule defines a rule for automatic categorization
type CategorizationRule struct {
	Pattern     *regexp.Regexp
	Keywords    []string
	Category    models.DisputeCategory
	Priority    models.DisputePriority
	Confidence  float64
	Description string
}

// PriorityRule defines rules for dynamic priority assignment
type PriorityRule struct {
	Condition func(*models.Dispute) bool
	Priority  models.DisputePriority
	Reason    string
	Weight    float64
}

// ResolutionMethodRule defines rules for resolution method selection
type ResolutionMethodRule struct {
	Condition        func(*models.Dispute) bool
	Method           models.DisputeResolutionMethod
	Reason           string
	SuitabilityScore float64
}

// ContentAnalyzer provides advanced content analysis capabilities
type ContentAnalyzer struct {
	// Entity patterns for recognition
	entityPatterns map[string]*regexp.Regexp
	// Sentiment analysis patterns
	sentimentPatterns map[string]float64
	// Semantic analysis patterns
	semanticPatterns map[string][]string
}

// ContentAnalysisResult represents the result of advanced content analysis
type ContentAnalysisResult struct {
	// Extracted entities
	Entities map[string][]string
	// Sentiment score (-1.0 to 1.0)
	SentimentScore float64
	// Semantic categories with confidence scores
	SemanticCategories map[string]float64
	// Key phrases extracted
	KeyPhrases []string
	// Content complexity score
	ComplexityScore float64
	// Urgency indicators
	UrgencyIndicators []string
}

// NewContentAnalyzer creates a new content analyzer with predefined patterns
func NewContentAnalyzer() *ContentAnalyzer {
	analyzer := &ContentAnalyzer{
		entityPatterns:    make(map[string]*regexp.Regexp),
		sentimentPatterns: make(map[string]float64),
		semanticPatterns:  make(map[string][]string),
	}

	analyzer.initializeEntityPatterns()
	analyzer.initializeSentimentPatterns()
	analyzer.initializeSemanticPatterns()

	return analyzer
}

// NewDisputeCategorizationService creates a new categorization service with predefined rules
func NewDisputeCategorizationService() *DisputeCategorizationService {
	service := &DisputeCategorizationService{
		categoryRules:   make(map[models.DisputeCategory][]CategorizationRule),
		contentAnalyzer: NewContentAnalyzer(),
		config: &CategorizationConfig{
			MinConfidenceAutoApply:       0.6,
			EntityFallbackMinConfidence:  0.4,
			DefaultFallbackMinConfidence: 0.25,
			SLA:                          SLAPriorityConfig{FirstEscalationDays: 7, SecondEscalationDays: 14, FinalEscalationDays: 30},
		},
	}

	service.initializeCategorizationRules()
	service.initializePriorityRules()
	service.initializeResolutionRules()

	return service
}

// NewDisputeCategorizationServiceWithRules creates a new categorization service with dynamic rules support
func NewDisputeCategorizationServiceWithRules(ruleService *CategorizationRuleService) *DisputeCategorizationService {
	service := &DisputeCategorizationService{
		categoryRules:   make(map[models.DisputeCategory][]CategorizationRule),
		contentAnalyzer: NewContentAnalyzer(),
		ruleService:     ruleService,
		config: &CategorizationConfig{
			MinConfidenceAutoApply:       0.6,
			EntityFallbackMinConfidence:  0.4,
			DefaultFallbackMinConfidence: 0.25,
			SLA:                          SLAPriorityConfig{FirstEscalationDays: 7, SecondEscalationDays: 14, FinalEscalationDays: 30},
		},
	}

	service.initializeCategorizationRules()
	service.initializePriorityRules()
	service.initializeResolutionRules()

	return service
}

// WithDisputeRepository injects the dispute repository for audit logging
func (s *DisputeCategorizationService) WithDisputeRepository(repo repository.DisputeRepositoryInterface) *DisputeCategorizationService {
	s.disputeRepo = repo
	return s
}

// WithConfig overrides the default categorization configuration
func (s *DisputeCategorizationService) WithConfig(cfg *CategorizationConfig) *DisputeCategorizationService {
	if cfg != nil {
		s.config = cfg
	}
	return s
}

// initializeEntityPatterns sets up regex patterns for entity recognition
func (ca *ContentAnalyzer) initializeEntityPatterns() {
	// Currency amounts
	ca.entityPatterns["currency"] = regexp.MustCompile(`(?i)(\$|usd|eur|gbp|inr|e₹|₹|usdt|usdc)\s*\d{1,3}(?:,\d{3})*(?:\.\d{2})?|\d{1,3}(?:,\d{3})*(?:\.\d{2})?\s*(\$|usd|eur|gbp|inr|e₹|₹|usdt|usdc)`)

	// Dates
	ca.entityPatterns["date"] = regexp.MustCompile(`(?i)\b\d{1,2}[/-]\d{1,2}[/-]\d{2,4}|\b\d{2,4}[/-]\d{1,2}[/-]\d{1,2}|\b(january|february|march|april|may|june|july|august|september|october|november|december|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\s+\d{1,2}(?:st|nd|rd|th)?,?\s+\d{2,4}`)

	// Contract references
	ca.entityPatterns["contract"] = regexp.MustCompile(`(?i)(contract|agreement|po|purchase order|order)\s+#?\w+-\d+|\w+-\d+\s+(contract|agreement|po|purchase order|order)`)

	// Milestone references
	ca.entityPatterns["milestone"] = regexp.MustCompile(`(?i)(milestone|phase|stage|deliverable)\s+#?\d+|\d+(?:st|nd|rd|th)?\s+(milestone|phase|stage|deliverable)`)

	// Time periods
	ca.entityPatterns["time_period"] = regexp.MustCompile(`(?i)\d+\s+(day|week|month|hour)s?|within\s+\d+\s+(day|week|month|hour)s?`)
}

// initializeSentimentPatterns sets up patterns for sentiment analysis
func (ca *ContentAnalyzer) initializeSentimentPatterns() {
	// Positive sentiment indicators
	positivePatterns := map[string]float64{
		"excellent": 0.8, "satisfactory": 0.6, "good": 0.5, "completed": 0.4,
		"delivered": 0.4, "successful": 0.6, "resolved": 0.5, "happy": 0.7,
		"pleased": 0.6, "grateful": 0.5, "thank": 0.4, "appreciate": 0.4,
	}

	// Negative sentiment indicators
	negativePatterns := map[string]float64{
		"terrible": -0.8, "awful": -0.8, "horrible": -0.7, "bad": -0.6,
		"disappointed": -0.7, "failed": -0.6, "breach": -0.5, "violation": -0.5,
		"delay": -0.4, "late": -0.4, "missing": -0.4, "incomplete": -0.4,
		"refuse": -0.5, "deny": -0.5, "reject": -0.5, "cancel": -0.4,
		"fraud": -0.9, "scam": -0.9, "unauthorized": -0.7, "stolen": -0.8,
	}

	// Urgency indicators
	urgencyPatterns := map[string]float64{
		"urgent": -0.3, "immediate": -0.4, "emergency": -0.5, "critical": -0.4,
		"asap": -0.3, "deadline": -0.2, "overdue": -0.3, "past due": -0.3,
	}

	// Combine all patterns
	for pattern, score := range positivePatterns {
		ca.sentimentPatterns[pattern] = score
	}
	for pattern, score := range negativePatterns {
		ca.sentimentPatterns[pattern] = score
	}
	for pattern, score := range urgencyPatterns {
		ca.sentimentPatterns[pattern] = score
	}
}

// initializeSemanticPatterns sets up patterns for semantic category analysis
func (ca *ContentAnalyzer) initializeSemanticPatterns() {
	ca.semanticPatterns["payment_dispute"] = []string{
		"payment", "paid", "unpaid", "refund", "charge", "fee", "billing", "invoice",
		"money", "currency", "transaction", "transfer", "amount", "balance",
	}

	ca.semanticPatterns["milestone_dispute"] = []string{
		"milestone", "deliverable", "delivery", "completion", "progress", "stage",
		"deadline", "due", "schedule", "timeline", "delay", "overdue",
	}

	ca.semanticPatterns["contract_dispute"] = []string{
		"contract", "agreement", "terms", "breach", "violation", "non-compliance",
		"obligation", "commitment", "promise", "clause", "provision",
	}

	ca.semanticPatterns["fraud_dispute"] = []string{
		"fraud", "scam", "fake", "forgery", "unauthorized", "theft", "stolen",
		"suspicious", "illegal", "criminal", "deceptive", "misleading",
	}

	ca.semanticPatterns["technical_dispute"] = []string{
		"error", "bug", "system", "technical", "glitch", "failure", "crash",
		"timeout", "connection", "server", "database", "api", "integration",
	}
}

// AnalyzeContent performs advanced content analysis on dispute text
func (ca *ContentAnalyzer) AnalyzeContent(title, description string) *ContentAnalysisResult {
	content := strings.ToLower(title + " " + description)

	result := &ContentAnalysisResult{
		Entities:           make(map[string][]string),
		SemanticCategories: make(map[string]float64),
		KeyPhrases:         make([]string, 0),
		UrgencyIndicators:  make([]string, 0),
	}

	// Extract entities
	result.Entities = ca.extractEntities(content)

	// Analyze sentiment
	result.SentimentScore = ca.analyzeSentiment(content)

	// Perform semantic analysis
	result.SemanticCategories = ca.analyzeSemantics(content)

	// Extract key phrases
	result.KeyPhrases = ca.extractKeyPhrases(title, description)

	// Calculate complexity
	result.ComplexityScore = ca.calculateComplexity(content)

	// Find urgency indicators
	result.UrgencyIndicators = ca.findUrgencyIndicators(content)

	return result
}

// extractEntities extracts named entities from the content
func (ca *ContentAnalyzer) extractEntities(content string) map[string][]string {
	entities := make(map[string][]string)

	for entityType, pattern := range ca.entityPatterns {
		matches := pattern.FindAllString(content, -1)
		if len(matches) > 0 {
			// Remove duplicates
			uniqueMatches := make(map[string]bool)
			for _, match := range matches {
				uniqueMatches[strings.TrimSpace(match)] = true
			}
			for match := range uniqueMatches {
				entities[entityType] = append(entities[entityType], match)
			}
		}
	}

	return entities
}

// analyzeSentiment performs basic sentiment analysis
func (ca *ContentAnalyzer) analyzeSentiment(content string) float64 {
	words := strings.Fields(content)
	score := 0.0
	wordCount := 0

	// Look for sentiment-bearing words
	for _, word := range words {
		// Remove punctuation and convert to lowercase
		word = strings.ToLower(strings.Trim(word, ".,!?;:\"'()[]{}"))

		if sentimentScore, exists := ca.sentimentPatterns[word]; exists {
			score += sentimentScore
			wordCount++
		}
	}

	// If no explicit sentiment words found, look for contextual indicators
	if wordCount == 0 {
		// Check for urgency indicators which often indicate negative sentiment
		urgencyCount := 0
		for _, word := range words {
			word = strings.ToLower(strings.Trim(word, ".,!?;:\"'()[]{}"))
			if word == "urgent" || word == "critical" || word == "immediate" || word == "blocking" {
				urgencyCount++
			}
		}

		if urgencyCount > 0 {
			score = -0.3 * float64(urgencyCount) // Negative sentiment for urgency
			wordCount = urgencyCount
		} else {
			// Check for positive indicators
			positiveCount := 0
			for _, word := range words {
				word = strings.ToLower(strings.Trim(word, ".,!?;:\"'()[]{}"))
				if word == "satisfactory" || word == "completed" || word == "successful" {
					positiveCount++
				}
			}
			if positiveCount > 0 {
				score = 0.2 * float64(positiveCount)
				wordCount = positiveCount
			}
		}
	}

	if wordCount == 0 {
		return 0.0
	}

	// Normalize score to [-1, 1] range
	normalizedScore := score / float64(wordCount)

	// Cap the score
	if normalizedScore > 1.0 {
		normalizedScore = 1.0
	} else if normalizedScore < -1.0 {
		normalizedScore = -1.0
	}

	return normalizedScore
}

// analyzeSemantics performs semantic category analysis
func (ca *ContentAnalyzer) analyzeSemantics(content string) map[string]float64 {
	categories := make(map[string]float64)
	words := strings.Fields(content)

	for category, keywords := range ca.semanticPatterns {
		score := 0.0
		matches := 0

		for _, word := range words {
			word = strings.Trim(word, ".,!?;:\"'()[]{}")
			for _, keyword := range keywords {
				if strings.Contains(word, keyword) || strings.Contains(keyword, word) {
					score += 1.0
					matches++
					break
				}
			}
		}

		if matches > 0 {
			// Normalize by keyword count and content length
			categories[category] = score / float64(len(keywords)) / float64(len(words)) * 1000
		}
	}

	return categories
}

// extractKeyPhrases extracts important phrases from the content
func (ca *ContentAnalyzer) extractKeyPhrases(title, description string) []string {
	phrases := make([]string, 0)

	// Extract phrases from title (usually most important)
	titleWords := strings.Fields(title)
	if len(titleWords) > 0 {
		phrases = append(phrases, strings.Join(titleWords, " "))
	}

	// Look for noun phrases and important terms in description (case-insensitive)
	descriptionLower := strings.ToLower(description)
	importantPatterns := []string{
		"breach of contract", "payment not received", "goods not delivered",
		"service not performed", "quality issues", "delivery delay",
		"incorrect amount", "unauthorized transaction", "system error",
		"technical failure", "communication breakdown", "non-compliance",
	}

	for _, pattern := range importantPatterns {
		if strings.Contains(descriptionLower, pattern) {
			phrases = append(phrases, pattern)
		}
	}

	return phrases
}

// calculateComplexity calculates content complexity score
func (ca *ContentAnalyzer) calculateComplexity(content string) float64 {
	words := strings.Fields(content)
	sentences := strings.Split(content, ".")

	// Factors: word count, sentence count, average word length
	avgWordLength := 0.0
	for _, word := range words {
		avgWordLength += float64(len(word))
	}
	if len(words) > 0 {
		avgWordLength /= float64(len(words))
	}

	complexity := float64(len(words))*0.1 +
		float64(len(sentences))*0.2 +
		avgWordLength*0.05

	// Normalize to 0-1 scale
	return complexity / 10.0
}

// findUrgencyIndicators identifies urgency-related terms
func (ca *ContentAnalyzer) findUrgencyIndicators(content string) []string {
	indicators := []string{
		"urgent", "immediate", "emergency", "critical", "asap",
		"deadline", "overdue", "past due", "time sensitive",
		"blocking", "show stopper", "cannot proceed",
	}

	found := make([]string, 0)
	for _, indicator := range indicators {
		if strings.Contains(strings.ToLower(content), indicator) {
			found = append(found, indicator)
		}
	}

	return found
}

// determineCategoryWithAnalysis determines the best category using both rule-based and content analysis
func (s *DisputeCategorizationService) determineCategoryWithAnalysis(dispute *models.Dispute, analysis *ContentAnalysisResult, matchedRules *[]string) models.DisputeCategory {
	content := strings.ToLower(dispute.Title + " " + dispute.Description)

	// Get traditional rule-based matches
	ruleMatches := s.findCategoryMatches(content)

	// Get semantic analysis results
	semanticMatches := make([]CategorizationRule, 0)
	for semanticCategory, score := range analysis.SemanticCategories {
		if score > 0.5 { // Threshold for considering semantic match
			var category models.DisputeCategory
			switch semanticCategory {
			case "payment_dispute":
				category = models.DisputeCategoryPayment
			case "milestone_dispute":
				category = models.DisputeCategoryMilestone
			case "contract_dispute":
				category = models.DisputeCategoryContractBreach
			case "fraud_dispute":
				category = models.DisputeCategoryFraud
			case "technical_dispute":
				category = models.DisputeCategoryTechnical
			default:
				continue
			}

			semanticMatches = append(semanticMatches, CategorizationRule{
				Category:    category,
				Confidence:  score / 10.0, // Normalize semantic score to confidence
				Description: fmt.Sprintf("Semantic analysis suggests %s (score: %.2f)", category, score),
			})
		}
	}

	// Combine rule-based and semantic matches
	allMatches := append(ruleMatches, semanticMatches...)

	// Consider key phrases for additional categorization hints
	for _, phrase := range analysis.KeyPhrases {
		if strings.Contains(phrase, "breach") || strings.Contains(phrase, "violation") {
			allMatches = append(allMatches, CategorizationRule{
				Category:    models.DisputeCategoryContractBreach,
				Confidence:  0.7,
				Description: fmt.Sprintf("Key phrase '%s' suggests contract dispute", phrase),
			})
		}
		if strings.Contains(phrase, "payment not received") || strings.Contains(phrase, "unpaid") {
			allMatches = append(allMatches, CategorizationRule{
				Category:    models.DisputeCategoryPayment,
				Confidence:  0.8,
				Description: fmt.Sprintf("Key phrase '%s' suggests payment dispute", phrase),
			})
		}
	}

	// Sort by confidence and return the best match
	for i := 0; i < len(allMatches)-1; i++ {
		for j := i + 1; j < len(allMatches); j++ {
			if allMatches[j].Confidence > allMatches[i].Confidence {
				allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
			}
		}
	}

	if len(allMatches) > 0 {
		*matchedRules = append(*matchedRules, allMatches[0].Description)
		return allMatches[0].Category
	}

	// Enhanced fallback using content analysis
	if len(analysis.Entities["currency"]) > 0 {
		*matchedRules = append(*matchedRules, "Contains currency entities, suggesting payment dispute")
		return models.DisputeCategoryPayment
	}

	if len(analysis.Entities["contract"]) > 0 {
		*matchedRules = append(*matchedRules, "Contains contract references, suggesting contract dispute")
		return models.DisputeCategoryContractBreach
	}

	// Default fallback
	return models.DisputeCategoryOther
}

// calculateEnhancedConfidence calculates confidence using multiple factors
func (s *DisputeCategorizationService) calculateEnhancedConfidence(dispute *models.Dispute, analysis *ContentAnalysisResult, category models.DisputeCategory) float64 {
	baseConfidence := 0.0

	// Base confidence from rule matching
	content := strings.ToLower(dispute.Title + " " + dispute.Description)
	ruleMatches := s.findCategoryMatches(content)
	if len(ruleMatches) > 0 && ruleMatches[0].Category == category {
		baseConfidence = ruleMatches[0].Confidence
	}

	// Boost confidence based on semantic analysis
	if semanticScore, exists := analysis.SemanticCategories[s.categoryToSemanticKey(category)]; exists {
		baseConfidence += semanticScore / 20.0 // Normalize and add
	}

	// Boost confidence based on entities
	if len(analysis.Entities) > 0 {
		entityBoost := float64(len(analysis.Entities)) * 0.1
		if entityBoost > 0.3 {
			entityBoost = 0.3 // Cap entity boost
		}
		baseConfidence += entityBoost
	}

	// Boost confidence based on key phrases
	phraseBoost := float64(len(analysis.KeyPhrases)) * 0.05
	if phraseBoost > 0.2 {
		phraseBoost = 0.2 // Cap phrase boost
	}
	baseConfidence += phraseBoost

	// Consider sentiment - extreme sentiment can indicate clearer disputes
	sentimentAbs := analysis.SentimentScore
	if sentimentAbs < 0 {
		sentimentAbs = -sentimentAbs
	}
	if sentimentAbs > 0.5 {
		baseConfidence += 0.1 // Clear sentiment boosts confidence
	}

	// Cap confidence at 0.95
	if baseConfidence > 0.95 {
		baseConfidence = 0.95
	}

	// Minimum confidence
	if baseConfidence < 0.2 {
		baseConfidence = 0.2
	}

	return baseConfidence
}

// calculatePriorityWithAnalysis calculates priority using content analysis insights
func (s *DisputeCategorizationService) calculatePriorityWithAnalysis(dispute *models.Dispute, analysis *ContentAnalysisResult, matchedRules *[]string) models.DisputePriority {
	// Compute a risk-based baseline and combine with rules-based
	riskPriority := s.calculateRiskBasedPriority(dispute, analysis, matchedRules)
	rulesPriority := s.calculatePriority(dispute, matchedRules)
	priority := maxPriority(riskPriority, rulesPriority)

	// Adjust based on content analysis
	if len(analysis.UrgencyIndicators) > 0 {
		// Multiple urgency indicators suggest higher priority
		if len(analysis.UrgencyIndicators) >= 2 {
			if priority == models.DisputePriorityNormal {
				priority = models.DisputePriorityHigh
				*matchedRules = append(*matchedRules, "Multiple urgency indicators detected")
			} else if priority == models.DisputePriorityHigh {
				priority = models.DisputePriorityUrgent
				*matchedRules = append(*matchedRules, "High urgency indicators with elevated priority")
			}
		}
	}

	// Consider sentiment - very negative sentiment might indicate urgency
	if analysis.SentimentScore < -0.6 {
		if priority == models.DisputePriorityNormal {
			priority = models.DisputePriorityHigh
			*matchedRules = append(*matchedRules, "Strong negative sentiment suggests higher priority")
		}
	}

	// High complexity might require more attention
	if analysis.ComplexityScore > 0.7 {
		if priority == models.DisputePriorityNormal {
			priority = models.DisputePriorityHigh
			*matchedRules = append(*matchedRules, "High content complexity suggests detailed review needed")
		}
	}

	// SLA-based escalation: escalate priority as the dispute ages (configurable)
	// Thresholds: >First:+1, >Second:+1 (cumulative), >Final: urgent
	now := time.Now()
	age := now.Sub(dispute.InitiatedAt)
	// Final escalation hard floor to urgent
	if s.config != nil && age.Hours() >= 24*float64(s.config.SLA.FinalEscalationDays) {
		if priority != models.DisputePriorityUrgent {
			priority = models.DisputePriorityUrgent
			*matchedRules = append(*matchedRules, "SLA escalation: dispute age > final threshold -> urgent")
		}
	} else {
		// staged escalations
		if s.config != nil && age.Hours() >= 24*float64(s.config.SLA.SecondEscalationDays) {
			// second escalation
			old := priority
			priority = increasePriority(priority)
			if old != priority {
				*matchedRules = append(*matchedRules, "SLA escalation: dispute age > second threshold")
			}
		}
		if s.config != nil && age.Hours() >= 24*float64(s.config.SLA.FirstEscalationDays) {
			// first escalation
			old := priority
			priority = increasePriority(priority)
			if old != priority {
				*matchedRules = append(*matchedRules, "SLA escalation: dispute age > first threshold")
			}
		}
	}

	// Timeline urgency detection from content
	content := strings.ToLower(dispute.Title + " " + dispute.Description)
	if strings.Contains(content, "overdue") || strings.Contains(content, "past due") {
		old := priority
		priority = increasePriority(priority)
		if old != priority {
			*matchedRules = append(*matchedRules, "Timeline indicators detected (overdue/past due)")
		}
	}

	// Stakeholder impact analysis: key accounts, regulatory impact, linked milestones/cheques with amount
	hasKeyImpact := false
	for _, tag := range dispute.Tags {
		lower := strings.ToLower(tag)
		if lower == "key_account" || lower == "vip" || lower == "regulatory" {
			hasKeyImpact = true
			break
		}
	}
	if !hasKeyImpact && dispute.Metadata != nil {
		if v, ok := dispute.Metadata["regulatory_impact"].(bool); ok && v {
			hasKeyImpact = true
		}
	}
	if !hasKeyImpact {
		if (dispute.SmartChequeID != nil || dispute.MilestoneID != nil) && dispute.DisputedAmount != nil && *dispute.DisputedAmount > 10000 {
			hasKeyImpact = true
		}
	}
	if hasKeyImpact {
		old := priority
		priority = increasePriority(priority)
		if old != priority {
			*matchedRules = append(*matchedRules, "Stakeholder impact escalation (key account/regulatory/high value milestone)")
		}
	}

	return priority
}

// increasePriority moves priority up one level, capping at urgent
func increasePriority(p models.DisputePriority) models.DisputePriority {
	switch p {
	case models.DisputePriorityLow:
		return models.DisputePriorityNormal
	case models.DisputePriorityNormal:
		return models.DisputePriorityHigh
	case models.DisputePriorityHigh:
		return models.DisputePriorityUrgent
	default:
		return p
	}
}

// addContentAnalysisInsights adds content analysis findings to matched rules
func (s *DisputeCategorizationService) addContentAnalysisInsights(analysis *ContentAnalysisResult, matchedRules *[]string) {
	// Add entity insights
	for entityType, entities := range analysis.Entities {
		if len(entities) > 0 {
			*matchedRules = append(*matchedRules, fmt.Sprintf("Found %d %s entities", len(entities), entityType))
		}
	}

	// Add sentiment insights
	if analysis.SentimentScore < -0.3 {
		*matchedRules = append(*matchedRules, fmt.Sprintf("Strong negative sentiment (%.2f)", analysis.SentimentScore))
	} else if analysis.SentimentScore > 0.3 {
		*matchedRules = append(*matchedRules, fmt.Sprintf("Strong positive sentiment (%.2f)", analysis.SentimentScore))
	}

	// Add urgency insights - check individual indicators
	urgencyFound := false
	for _, indicator := range analysis.UrgencyIndicators {
		if strings.Contains(strings.ToLower(indicator), "urgent") {
			*matchedRules = append(*matchedRules, "urgent")
			urgencyFound = true
			break
		}
	}
	if !urgencyFound && len(analysis.UrgencyIndicators) > 0 {
		*matchedRules = append(*matchedRules, fmt.Sprintf("Urgency indicators: %s", strings.Join(analysis.UrgencyIndicators, ", ")))
	}

	// Add key phrase insights
	if len(analysis.KeyPhrases) > 0 {
		*matchedRules = append(*matchedRules, fmt.Sprintf("Key phrases identified: %s", strings.Join(analysis.KeyPhrases, "; ")))
	}
}

// calculateRiskBasedPriority computes a composite risk score and maps it to a baseline priority
func (s *DisputeCategorizationService) calculateRiskBasedPriority(dispute *models.Dispute, analysis *ContentAnalysisResult, matchedRules *[]string) models.DisputePriority {
	score := 0.0

	// Amount factor
	if dispute.DisputedAmount != nil {
		amt := *dispute.DisputedAmount
		switch {
		case amt >= 100000:
			score += 0.45
			if matchedRules != nil {
				*matchedRules = append(*matchedRules, "High amount factor (>=100k)")
			}
		case amt >= 50000:
			score += 0.30
			if matchedRules != nil {
				*matchedRules = append(*matchedRules, "Medium-high amount factor (>=50k)")
			}
		case amt >= 10000:
			score += 0.18
			if matchedRules != nil {
				*matchedRules = append(*matchedRules, "Medium amount factor (>=10k)")
			}
		case amt >= 1000:
			score += 0.08
		}
	}

	// Category severity factor
	switch dispute.Category {
	case models.DisputeCategoryFraud:
		score += 0.35
		if matchedRules != nil {
			*matchedRules = append(*matchedRules, "Category severity: fraud")
		}
	case models.DisputeCategoryContractBreach:
		score += 0.22
	case models.DisputeCategoryMilestone:
		score += 0.18
	case models.DisputeCategoryPayment:
		score += 0.15
	case models.DisputeCategoryTechnical:
		score += 0.10
	}

	// Recurrence factor (from metadata or tags)
	recurrence := 0.0
	if dispute.Metadata != nil {
		if v, ok := dispute.Metadata["recurrence_count"].(float64); ok && v >= 2 {
			recurrence = 0.12
		}
	}
	for _, t := range dispute.Tags {
		if strings.EqualFold(t, "recurring") || strings.EqualFold(t, "repeat_issue") {
			recurrence = 0.12
			break
		}
	}
	score += recurrence

	// Linkage factor: linked milestone or smart cheque, boost for larger amounts
	if dispute.SmartChequeID != nil || dispute.MilestoneID != nil {
		linkBoost := 0.10
		if dispute.DisputedAmount != nil && *dispute.DisputedAmount > 10000 {
			linkBoost = 0.18
		}
		score += linkBoost
		if matchedRules != nil {
			*matchedRules = append(*matchedRules, "Linked to milestone/smart cheque")
		}
	}

	// Urgency indicators in content
	if analysis != nil && len(analysis.UrgencyIndicators) > 0 {
		if len(analysis.UrgencyIndicators) >= 2 {
			score += 0.15
		} else {
			score += 0.08
		}
	}

	// Cap score to 1.0
	if score > 1.0 {
		score = 1.0
	}

	// Map score bands to priority
	switch {
	case score >= 0.6:
		return models.DisputePriorityUrgent
	case score >= 0.35:
		return models.DisputePriorityHigh
	case score >= 0.15:
		return models.DisputePriorityNormal
	default:
		return models.DisputePriorityLow
	}
}

// RecomputeAndApplyPriority recomputes priority and applies it if changed; emits audit log
func (s *DisputeCategorizationService) RecomputeAndApplyPriority(ctx context.Context, disputeID string, userID string) (bool, models.DisputePriority, error) {
	if s.disputeRepo == nil {
		return false, "", fmt.Errorf("dispute repository not configured")
	}

	dispute, err := s.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return false, "", err
	}
	if dispute == nil {
		return false, "", fmt.Errorf("dispute not found: %s", disputeID)
	}

	analysis := s.contentAnalyzer.AnalyzeContent(dispute.Title, dispute.Description)
	matched := make([]string, 0)
	newPriority := s.calculatePriorityWithAnalysis(dispute, analysis, &matched)

	// Enforce SLA floor
	floor := s.computeSLAFloorPriority(dispute)
	if priorityBelowFloor(newPriority, floor) {
		newPriority = floor
	}

	if newPriority == dispute.Priority {
		// Update next review metadata even if unchanged
		if dispute.Metadata == nil {
			dispute.Metadata = make(map[string]interface{})
		}
		next := s.computeNextPriorityReview(dispute)
		if next != nil {
			dispute.Metadata["next_priority_review_at"] = next.Format(time.RFC3339)
		} else {
			dispute.Metadata["next_priority_review_at"] = nil
		}
		if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
			return false, dispute.Priority, err
		}
		return false, dispute.Priority, nil
	}

	old := dispute.Priority
	dispute.Priority = newPriority
	dispute.UpdatedBy = userID
	if dispute.Metadata == nil {
		dispute.Metadata = make(map[string]interface{})
	}
	next := s.computeNextPriorityReview(dispute)
	if next != nil {
		dispute.Metadata["next_priority_review_at"] = next.Format(time.RFC3339)
	} else {
		dispute.Metadata["next_priority_review_at"] = nil
	}

	if err := s.disputeRepo.UpdateDispute(ctx, dispute); err != nil {
		return false, old, err
	}

	// Audit log
	_ = s.disputeRepo.CreateAuditLog(ctx, &models.DisputeAuditLog{
		DisputeID: dispute.ID,
		Action:    "priority_recomputed",
		UserID:    userID,
		UserType:  "system",
		Details:   fmt.Sprintf("priority changed from %s to %s", old, newPriority),
		CreatedAt: time.Now(),
	})

	return true, newPriority, nil
}

// buildEnhancedCategorizationReason builds a comprehensive reason string
func (s *DisputeCategorizationService) buildEnhancedCategorizationReason(result *CategorizationResult, analysis *ContentAnalysisResult) string {
	reason := fmt.Sprintf("Automatically categorized as %s priority %s dispute",
		result.SuggestedPriority, result.SuggestedCategory)

	if len(result.MatchedRules) > 0 {
		reason += fmt.Sprintf(" based on: %s", strings.Join(result.MatchedRules, ", "))
	}

	reason += fmt.Sprintf(" (confidence: %.1f%%)", result.Confidence*100)

	// Add content analysis summary
	if analysis.SentimentScore != 0 {
		sentiment := "neutral"
		if analysis.SentimentScore > 0.2 {
			sentiment = "positive"
		} else if analysis.SentimentScore < -0.2 {
			sentiment = "negative"
		}
		reason += fmt.Sprintf(", sentiment: %s", sentiment)
	}

	if len(analysis.Entities) > 0 {
		reason += fmt.Sprintf(", entities detected: %d types", len(analysis.Entities))
	}

	return reason
}

// categoryToSemanticKey converts a dispute category to semantic analysis key
func (s *DisputeCategorizationService) categoryToSemanticKey(category models.DisputeCategory) string {
	switch category {
	case models.DisputeCategoryPayment:
		return "payment_dispute"
	case models.DisputeCategoryMilestone:
		return "milestone_dispute"
	case models.DisputeCategoryContractBreach:
		return "contract_dispute"
	case models.DisputeCategoryFraud:
		return "fraud_dispute"
	case models.DisputeCategoryTechnical:
		return "technical_dispute"
	default:
		return ""
	}
}

// determineCategoryHybrid determines category using both static and dynamic rules
func (s *DisputeCategorizationService) determineCategoryHybrid(dispute *models.Dispute, analysis *ContentAnalysisResult, dynamicMatches []*RuleMatchResult, matchedRules *[]string) models.DisputeCategory {
	// First, try dynamic rules if we have high-confidence matches
	if len(dynamicMatches) > 0 {
		for _, match := range dynamicMatches {
			if match.IsMatch && match.Confidence > 0.8 {
				*matchedRules = append(*matchedRules, fmt.Sprintf("High-confidence dynamic rule match: %s", match.Rule.Name))
				return match.Rule.Category
			}
		}
	}

	// Fall back to enhanced static analysis
	return s.determineCategoryWithAnalysis(dispute, analysis, matchedRules)
}

// calculateHybridConfidence calculates confidence using both static and dynamic approaches
func (s *DisputeCategorizationService) calculateHybridConfidence(dispute *models.Dispute, analysis *ContentAnalysisResult, dynamicMatches []*RuleMatchResult, category models.DisputeCategory) float64 {
	baseConfidence := s.calculateEnhancedConfidence(dispute, analysis, category)

	// Boost confidence if we have dynamic rule matches
	if len(dynamicMatches) > 0 {
		dynamicBoost := 0.0
		validMatches := 0

		for _, match := range dynamicMatches {
			if match.IsMatch && match.Rule.Category == category {
				dynamicBoost += match.Confidence * match.Rule.Weight
				validMatches++
			}
		}

		if validMatches > 0 {
			// Average dynamic boost weighted by rule performance
			avgDynamicBoost := dynamicBoost / float64(validMatches)
			baseConfidence = baseConfidence*0.7 + avgDynamicBoost*0.3 // 70% static, 30% dynamic
		}
	}

	return baseConfidence
}

// buildHybridCategorizationReason builds a comprehensive reason string for hybrid categorization
func (s *DisputeCategorizationService) buildHybridCategorizationReason(result *CategorizationResult, analysis *ContentAnalysisResult, dynamicMatches []*RuleMatchResult) string {
	reason := fmt.Sprintf("Automatically categorized as %s priority %s dispute",
		result.SuggestedPriority, result.SuggestedCategory)

	if len(result.MatchedRules) > 0 {
		reason += fmt.Sprintf(" based on: %s", strings.Join(result.MatchedRules, ", "))
	}

	reason += fmt.Sprintf(" (confidence: %.1f%%)", result.Confidence*100)

	// Add content analysis summary
	if analysis.SentimentScore != 0 {
		sentiment := "neutral"
		if analysis.SentimentScore > 0.2 {
			sentiment = "positive"
		} else if analysis.SentimentScore < -0.2 {
			sentiment = "negative"
		}
		reason += fmt.Sprintf(", sentiment: %s", sentiment)
	}

	if len(analysis.Entities) > 0 {
		reason += fmt.Sprintf(", entities detected: %d types", len(analysis.Entities))
	}

	// Add dynamic rules summary
	if len(dynamicMatches) > 0 {
		highConfidenceMatches := 0
		for _, match := range dynamicMatches {
			if match.IsMatch && match.Confidence > 0.7 {
				highConfidenceMatches++
			}
		}
		if highConfidenceMatches > 0 {
			reason += fmt.Sprintf(", %d high-confidence dynamic rules matched", highConfidenceMatches)
		}
	}

	return reason
}

// CategorizationResult represents the result of automatic categorization
type CategorizationResult struct {
	SuggestedCategory models.DisputeCategory         `json:"suggested_category"`
	SuggestedPriority models.DisputePriority         `json:"suggested_priority"`
	SuggestedMethod   models.DisputeResolutionMethod `json:"suggested_method"`
	Confidence        float64                        `json:"confidence"`
	MatchedRules      []string                       `json:"matched_rules"`
	Reason            string                         `json:"reason"`
}

// AutoCategorizeDispute automatically categorizes a dispute using hybrid approach (static + dynamic rules)
func (s *DisputeCategorizationService) AutoCategorizeDispute(ctx context.Context, dispute *models.Dispute) (*CategorizationResult, error) {
	result := &CategorizationResult{
		Confidence:   0.0,
		MatchedRules: make([]string, 0),
	}

	// Perform advanced content analysis
	contentAnalysis := s.contentAnalyzer.AnalyzeContent(dispute.Title, dispute.Description)

	// Try dynamic rules first if available
	var dynamicMatches []*RuleMatchResult
	if s.ruleService != nil {
		var err error
		dynamicMatches, err = s.ruleService.ApplyRules(ctx, dispute)
		if err != nil {
			// Log error but continue with static rules
			result.MatchedRules = append(result.MatchedRules, "Dynamic rules evaluation failed, falling back to static rules")
		}
	}

	// Determine category using hybrid approach
	result.SuggestedCategory = s.determineCategoryHybrid(dispute, contentAnalysis, dynamicMatches, &result.MatchedRules)
	result.Confidence = s.calculateHybridConfidence(dispute, contentAnalysis, dynamicMatches, result.SuggestedCategory)

	// Apply multi-stage fallback strategies based on configured thresholds
	s.applyFallbackStrategies(dispute, contentAnalysis, &result)

	// Enhanced priority calculation with content analysis
	result.SuggestedPriority = s.calculatePriorityWithAnalysis(dispute, contentAnalysis, &result.MatchedRules)

	// Suggest resolution method
	result.SuggestedMethod = s.suggestResolutionMethod(dispute, result.SuggestedCategory, result.SuggestedPriority)

	// Add content analysis insights to matched rules
	s.addContentAnalysisInsights(contentAnalysis, &result.MatchedRules)

	// Add dynamic rule insights
	if len(dynamicMatches) > 0 {
		result.MatchedRules = append(result.MatchedRules, fmt.Sprintf("Applied %d dynamic rules", len(dynamicMatches)))
		for _, match := range dynamicMatches {
			if match.IsMatch && match.Confidence > 0.5 {
				result.MatchedRules = append(result.MatchedRules, fmt.Sprintf("Dynamic: %s (%.1f%%)", match.Rule.Name, match.Confidence*100))
			}
		}
	}

	// Build reason string
	result.Reason = s.buildHybridCategorizationReason(result, contentAnalysis, dynamicMatches)

	return result, nil
}

// applyFallbackStrategies adjusts result based on confidence thresholds using multi-stage fallbacks
func (s *DisputeCategorizationService) applyFallbackStrategies(dispute *models.Dispute, analysis *ContentAnalysisResult, result **CategorizationResult) {
	r := *result
	if s.config == nil {
		return
	}

	// If confidence is extremely low, fallback to default 'other' and flag for manual review
	if r.Confidence < s.config.DefaultFallbackMinConfidence {
		r.SuggestedCategory = models.DisputeCategoryOther
		r.MatchedRules = append(r.MatchedRules, "Fallback: default category due to very low confidence")
		// Give a minimal baseline confidence
		r.Confidence = s.config.DefaultFallbackMinConfidence
		*result = r
		return
	}

	// If confidence is low, but entities indicate likely category, fallback to entity-based heuristic
	if r.Confidence < s.config.EntityFallbackMinConfidence {
		if len(analysis.Entities["currency"]) > 0 {
			r.SuggestedCategory = models.DisputeCategoryPayment
			r.MatchedRules = append(r.MatchedRules, "Fallback: entity-based payment due to low confidence and currency entities")
			r.Confidence = s.config.EntityFallbackMinConfidence
		} else if len(analysis.Entities["contract"]) > 0 {
			r.SuggestedCategory = models.DisputeCategoryContractBreach
			r.MatchedRules = append(r.MatchedRules, "Fallback: entity-based contract due to low confidence and contract references")
			r.Confidence = s.config.EntityFallbackMinConfidence
		}
		*result = r
		return
	}

	// If below auto-apply threshold, annotate for review but keep categorization
	if r.Confidence < s.config.MinConfidenceAutoApply {
		r.MatchedRules = append(r.MatchedRules, fmt.Sprintf("Low-confidence categorization (%.2f) — flagged for review", r.Confidence))
		*result = r
	}
}

// ApplyCategorization applies the suggested categorization to a dispute
func (s *DisputeCategorizationService) ApplyCategorization(ctx context.Context, dispute *models.Dispute, result *CategorizationResult, userID string) error {
	// Update dispute with suggested values
	dispute.Category = result.SuggestedCategory
	dispute.Priority = result.SuggestedPriority
	dispute.UpdatedBy = userID
	dispute.UpdatedAt = time.Now()

	// Add metadata about automatic categorization
	if dispute.Metadata == nil {
		dispute.Metadata = make(map[string]interface{})
	}

	dispute.Metadata["auto_categorized"] = true
	dispute.Metadata["categorization_confidence"] = result.Confidence
	dispute.Metadata["categorization_rules"] = result.MatchedRules
	dispute.Metadata["suggested_resolution_method"] = result.SuggestedMethod
	dispute.Metadata["categorized_at"] = time.Now().Format(time.RFC3339)
	dispute.Metadata["categorized_by"] = userID

	// Record next review timestamp based on SLA thresholds
	if s.config != nil {
		next := s.computeNextPriorityReview(dispute)
		if next != nil {
			dispute.Metadata["next_priority_review_at"] = next.Format(time.RFC3339)
		} else {
			dispute.Metadata["next_priority_review_at"] = nil
		}
	}

	// Audit logging (best-effort)
	if s.disputeRepo != nil && dispute.ID != "" {
		details := fmt.Sprintf("category=%s, priority=%s, confidence=%.2f", result.SuggestedCategory, result.SuggestedPriority, result.Confidence)
		audit := &models.DisputeAuditLog{
			DisputeID: dispute.ID,
			Action:    "categorized",
			UserID:    userID,
			UserType:  "user",
			Details:   details,
			CreatedAt: time.Now(),
		}
		_ = s.disputeRepo.CreateAuditLog(ctx, audit)
	}

	return nil
}

// OverrideCategorization allows manual override of automatic categorization
func (s *DisputeCategorizationService) OverrideCategorization(ctx context.Context, dispute *models.Dispute, category models.DisputeCategory, priority models.DisputePriority, reason string, userID string) error {
	// Enforce SLA floor - do not allow lowering below the floor
	floor := s.computeSLAFloorPriority(dispute)
	if priorityBelowFloor(priority, floor) {
		return fmt.Errorf("cannot set priority below SLA floor (%s)", floor)
	}

	dispute.Category = category
	dispute.Priority = priority
	dispute.UpdatedBy = userID

	// Add override metadata
	if dispute.Metadata == nil {
		dispute.Metadata = make(map[string]interface{})
	}

	dispute.Metadata["manual_override"] = true
	dispute.Metadata["override_reason"] = reason
	dispute.Metadata["overridden_at"] = time.Now().Format(time.RFC3339)
	dispute.Metadata["overridden_by"] = userID

	// Audit logging (best-effort)
	if s.disputeRepo != nil && dispute.ID != "" {
		details := fmt.Sprintf("override to category=%s, priority=%s, reason=%s", category, priority, reason)
		audit := &models.DisputeAuditLog{
			DisputeID: dispute.ID,
			Action:    "categorization_overridden",
			UserID:    userID,
			UserType:  "user",
			Details:   details,
			CreatedAt: time.Now(),
		}
		_ = s.disputeRepo.CreateAuditLog(ctx, audit)
	}

	return nil
}

// computeNextPriorityReview returns the next review timestamp or nil if none
func (s *DisputeCategorizationService) computeNextPriorityReview(dispute *models.Dispute) *time.Time {
	now := time.Now()
	thresholds := []int{s.config.SLA.FirstEscalationDays, s.config.SLA.SecondEscalationDays, s.config.SLA.FinalEscalationDays}
	for _, d := range thresholds {
		thresholdTime := dispute.InitiatedAt.Add(time.Duration(d) * 24 * time.Hour)
		if now.Before(thresholdTime) {
			return &thresholdTime
		}
	}
	return nil
}

// computeSLAFloorPriority returns the minimum allowed priority based on age thresholds
func (s *DisputeCategorizationService) computeSLAFloorPriority(dispute *models.Dispute) models.DisputePriority {
	now := time.Now()
	age := now.Sub(dispute.InitiatedAt)
	if s.config == nil {
		return models.DisputePriorityLow
	}
	if age.Hours() >= 24*float64(s.config.SLA.FinalEscalationDays) {
		return models.DisputePriorityUrgent
	}
	if age.Hours() >= 24*float64(s.config.SLA.SecondEscalationDays) {
		return models.DisputePriorityHigh
	}
	if age.Hours() >= 24*float64(s.config.SLA.FirstEscalationDays) {
		return models.DisputePriorityNormal
	}
	return models.DisputePriorityLow
}

// ValidateCategorization checks if the current categorization is appropriate
func (s *DisputeCategorizationService) ValidateCategorization(ctx context.Context, dispute *models.Dispute) (*CategorizationValidation, error) {
	validation := &CategorizationValidation{
		IsValid:     true,
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Check if category matches dispute content
	content := strings.ToLower(dispute.Title + " " + dispute.Description)
	categoryMatches := s.findCategoryMatches(content)

	if len(categoryMatches) > 0 {
		bestMatch := categoryMatches[0]
		if bestMatch.Category != dispute.Category && bestMatch.Confidence > 0.7 {
			validation.IsValid = false
			validation.Issues = append(validation.Issues,
				fmt.Sprintf("Content suggests %s category (confidence: %.1f%%) but currently categorized as %s",
					bestMatch.Category, bestMatch.Confidence*100, dispute.Category))
			validation.Suggestions = append(validation.Suggestions,
				fmt.Sprintf("Consider changing category to %s", bestMatch.Category))
		}
	}

	// Check priority appropriateness
	suggestedPriority := s.calculatePriority(dispute, nil)
	if suggestedPriority != dispute.Priority {
		validation.Suggestions = append(validation.Suggestions,
			fmt.Sprintf("Consider changing priority from %s to %s based on dispute characteristics",
				dispute.Priority, suggestedPriority))
	}

	return validation, nil
}

// CategorizationValidation represents validation results for dispute categorization
type CategorizationValidation struct {
	IsValid     bool     `json:"is_valid"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
}

// Private methods for rule-based categorization

func (s *DisputeCategorizationService) initializeCategorizationRules() {
	// Payment-related patterns
	s.categoryRules[models.DisputeCategoryPayment] = []CategorizationRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(payment|pay|paid|unpaid|refund|charge|fee|billing)`),
			Keywords:    []string{"payment", "paid", "unpaid", "refund", "charge", "fee", "billing", "invoice"},
			Category:    models.DisputeCategoryPayment,
			Priority:    models.DisputePriorityHigh,
			Confidence:  0.8,
			Description: "Contains payment-related keywords",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(amount|money|currency|transaction|transfer)`),
			Keywords:    []string{"amount", "money", "currency", "transaction", "transfer"},
			Category:    models.DisputeCategoryPayment,
			Priority:    models.DisputePriorityNormal,
			Confidence:  0.6,
			Description: "Contains financial transaction keywords",
		},
	}

	// Milestone-related patterns
	s.categoryRules[models.DisputeCategoryMilestone] = []CategorizationRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(milestone|deliverable|delivery|completion|progress|stage)`),
			Keywords:    []string{"milestone", "deliverable", "delivery", "completion", "progress", "stage"},
			Category:    models.DisputeCategoryMilestone,
			Priority:    models.DisputePriorityHigh,
			Confidence:  0.8,
			Description: "Contains milestone or delivery-related keywords",
		},
		{
			Pattern:     regexp.MustCompile(`(?i)(deadline|due|schedule|timeline|delay|overdue)`),
			Keywords:    []string{"deadline", "due", "schedule", "timeline", "delay", "overdue"},
			Category:    models.DisputeCategoryMilestone,
			Priority:    models.DisputePriorityUrgent,
			Confidence:  0.7,
			Description: "Contains timeline or deadline-related keywords",
		},
	}

	// Contract breach patterns
	s.categoryRules[models.DisputeCategoryContractBreach] = []CategorizationRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(breach|violation|non-compliance|contract|agreement|terms)`),
			Keywords:    []string{"breach", "violation", "non-compliance", "contract", "agreement", "terms"},
			Category:    models.DisputeCategoryContractBreach,
			Priority:    models.DisputePriorityHigh,
			Confidence:  0.9,
			Description: "Contains contract breach or violation keywords",
		},
	}

	// Fraud patterns
	s.categoryRules[models.DisputeCategoryFraud] = []CategorizationRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(fraud|scam|fake|forgery|unauthorized|theft|stolen)`),
			Keywords:    []string{"fraud", "scam", "fake", "forgery", "unauthorized", "theft", "stolen"},
			Category:    models.DisputeCategoryFraud,
			Priority:    models.DisputePriorityUrgent,
			Confidence:  0.9,
			Description: "Contains fraud or unauthorized activity keywords",
		},
	}

	// Technical issues patterns
	s.categoryRules[models.DisputeCategoryTechnical] = []CategorizationRule{
		{
			Pattern:     regexp.MustCompile(`(?i)(error|bug|system|technical|glitch|failure|crash|timeout)`),
			Keywords:    []string{"error", "bug", "system", "technical", "glitch", "failure", "crash", "timeout"},
			Category:    models.DisputeCategoryTechnical,
			Priority:    models.DisputePriorityNormal,
			Confidence:  0.7,
			Description: "Contains technical issue keywords",
		},
	}
}

func (s *DisputeCategorizationService) initializePriorityRules() {
	s.priorityRules = []PriorityRule{
		// High amount disputes
		{
			Condition: func(d *models.Dispute) bool {
				return d.DisputedAmount != nil && *d.DisputedAmount > 50000
			},
			Priority: models.DisputePriorityUrgent,
			Reason:   "High disputed amount (> $50,000)",
			Weight:   1.0,
		},
		{
			Condition: func(d *models.Dispute) bool {
				return d.DisputedAmount != nil && *d.DisputedAmount > 10000
			},
			Priority: models.DisputePriorityHigh,
			Reason:   "Medium disputed amount (> $10,000)",
			Weight:   0.8,
		},
		// Time-sensitive disputes
		{
			Condition: func(d *models.Dispute) bool {
				return strings.Contains(strings.ToLower(d.Description), "urgent") ||
					strings.Contains(strings.ToLower(d.Description), "immediate") ||
					strings.Contains(strings.ToLower(d.Title), "urgent")
			},
			Priority: models.DisputePriorityUrgent,
			Reason:   "Contains urgent keywords",
			Weight:   0.9,
		},
		// Fraud-related disputes
		{
			Condition: func(d *models.Dispute) bool {
				return d.Category == models.DisputeCategoryFraud
			},
			Priority: models.DisputePriorityUrgent,
			Reason:   "Fraud-related dispute",
			Weight:   1.0,
		},
	}
}

func (s *DisputeCategorizationService) initializeResolutionRules() {
	s.resolutionRules = []ResolutionMethodRule{
		// Low amount disputes - mutual agreement
		{
			Condition: func(d *models.Dispute) bool {
				return d.DisputedAmount != nil && *d.DisputedAmount < 1000
			},
			Method:           models.DisputeResolutionMethodMutualAgreement,
			Reason:           "Low disputed amount suitable for mutual agreement",
			SuitabilityScore: 0.9,
		},
		// High amount or complex disputes - arbitration
		{
			Condition: func(d *models.Dispute) bool {
				return (d.DisputedAmount != nil && *d.DisputedAmount > 50000) ||
					d.Category == models.DisputeCategoryContractBreach
			},
			Method:           models.DisputeResolutionMethodArbitration,
			Reason:           "High amount or complex dispute requiring arbitration",
			SuitabilityScore: 0.8,
		},
		// Fraud disputes - may require legal action
		{
			Condition: func(d *models.Dispute) bool {
				return d.Category == models.DisputeCategoryFraud
			},
			Method:           models.DisputeResolutionMethodCourt,
			Reason:           "Fraud disputes may require court intervention",
			SuitabilityScore: 0.7,
		},
		// Default to mediation for most cases
		{
			Condition: func(d *models.Dispute) bool {
				return true // Default rule
			},
			Method:           models.DisputeResolutionMethodMediation,
			Reason:           "Mediation suitable for most dispute types",
			SuitabilityScore: 0.6,
		},
	}
}

func (s *DisputeCategorizationService) findCategoryMatches(content string) []CategorizationRule {
	var matches []CategorizationRule

	for _, rules := range s.categoryRules {
		for _, rule := range rules {
			score := 0.0

			// Check regex pattern
			if rule.Pattern != nil && rule.Pattern.MatchString(content) {
				score += rule.Confidence
			}

			// Check keywords
			for _, keyword := range rule.Keywords {
				if strings.Contains(content, keyword) {
					score += 0.1 // Small boost for each keyword match
				}
			}

			if score > 0 {
				ruleWithScore := rule
				ruleWithScore.Confidence = score
				matches = append(matches, ruleWithScore)
			}
		}
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Confidence > matches[i].Confidence {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

func (s *DisputeCategorizationService) calculatePriority(dispute *models.Dispute, matchedRules *[]string) models.DisputePriority {
	priorityScores := make(map[models.DisputePriority]float64)

	// Apply priority rules
	for _, rule := range s.priorityRules {
		if rule.Condition(dispute) {
			priorityScores[rule.Priority] += rule.Weight
			if matchedRules != nil {
				*matchedRules = append(*matchedRules, rule.Reason)
			}
		}
	}

	// Find priority with highest score
	maxScore := 0.0
	bestPriority := models.DisputePriorityNormal

	for priority, score := range priorityScores {
		if score > maxScore {
			maxScore = score
			bestPriority = priority
		}
	}

	return bestPriority
}

func (s *DisputeCategorizationService) suggestResolutionMethod(dispute *models.Dispute, category models.DisputeCategory, priority models.DisputePriority) models.DisputeResolutionMethod {
	var bestMethod models.DisputeResolutionMethod
	maxScore := -1.0

	// Score base on predefined rules
	for _, rule := range s.resolutionRules {
		if !rule.Condition(dispute) {
			continue
		}
		score := rule.SuitabilityScore

		// Priority-based adjustments
		switch priority {
		case models.DisputePriorityUrgent:
			if rule.Method == models.DisputeResolutionMethodCourt || rule.Method == models.DisputeResolutionMethodArbitration {
				score += 0.1
			}
		case models.DisputePriorityHigh:
			if rule.Method == models.DisputeResolutionMethodArbitration || rule.Method == models.DisputeResolutionMethodMediation {
				score += 0.05
			}
		case models.DisputePriorityNormal, models.DisputePriorityLow:
			if rule.Method == models.DisputeResolutionMethodMutualAgreement || rule.Method == models.DisputeResolutionMethodMediation {
				score += 0.1
			}
		}

		// Category-based adjustments
		switch category {
		case models.DisputeCategoryFraud:
			if rule.Method == models.DisputeResolutionMethodCourt {
				score += 0.2
			}
		case models.DisputeCategoryContractBreach:
			if rule.Method == models.DisputeResolutionMethodArbitration {
				score += 0.15
			}
		case models.DisputeCategoryPayment:
			if rule.Method == models.DisputeResolutionMethodMutualAgreement || rule.Method == models.DisputeResolutionMethodMediation {
				score += 0.1
			}
		case models.DisputeCategoryTechnical:
			if rule.Method == models.DisputeResolutionMethodMediation {
				score += 0.1
			}
		}

		// Amount-based adjustment: very high amounts favor arbitration
		if dispute.DisputedAmount != nil && *dispute.DisputedAmount > 100000 && rule.Method == models.DisputeResolutionMethodArbitration {
			score += 0.1
		}

		if score > maxScore {
			maxScore = score
			bestMethod = rule.Method
		}
	}

	// Fallback
	if maxScore < 0 {
		return models.DisputeResolutionMethodMediation
	}
	return bestMethod
}

func (s *DisputeCategorizationService) buildCategorizationReason(result *CategorizationResult) string {
	reason := fmt.Sprintf("Automatically categorized as %s priority %s dispute",
		result.SuggestedPriority, result.SuggestedCategory)

	if len(result.MatchedRules) > 0 {
		reason += fmt.Sprintf(" based on: %s", strings.Join(result.MatchedRules, ", "))
	}

	reason += fmt.Sprintf(" (confidence: %.1f%%)", result.Confidence*100)

	return reason
}

// CategorizationConfig defines scoring thresholds and fallback behavior
type CategorizationConfig struct {
	// Minimum confidence to auto-apply categorization without flags
	MinConfidenceAutoApply float64
	// If below this, prefer entity-based fallback suggestions
	EntityFallbackMinConfidence float64
	// If below this, default to 'other' category and flag for review
	DefaultFallbackMinConfidence float64
	// SLA thresholds for priority escalation
	SLA SLAPriorityConfig
}

// SLAPriorityConfig defines thresholds for priority escalation in days
type SLAPriorityConfig struct {
	FirstEscalationDays  int
	SecondEscalationDays int
	FinalEscalationDays  int
}

// priorityBelowFloor checks if a proposed priority is below a floor
func priorityBelowFloor(p models.DisputePriority, floor models.DisputePriority) bool {
	order := map[models.DisputePriority]int{
		models.DisputePriorityLow:    0,
		models.DisputePriorityNormal: 1,
		models.DisputePriorityHigh:   2,
		models.DisputePriorityUrgent: 3,
	}
	return order[p] < order[floor]
}

// maxPriority returns the higher of two priorities in severity order
func maxPriority(a, b models.DisputePriority) models.DisputePriority {
	order := map[models.DisputePriority]int{
		models.DisputePriorityLow:    0,
		models.DisputePriorityNormal: 1,
		models.DisputePriorityHigh:   2,
		models.DisputePriorityUrgent: 3,
	}
	if order[a] >= order[b] {
		return a
	}
	return b
}
