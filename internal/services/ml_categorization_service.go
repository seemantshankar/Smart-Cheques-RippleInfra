package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// MLCategorizationService provides machine learning-based dispute categorization
type MLCategorizationService struct {
	modelRepo       repository.CategorizationMLModelRepositoryInterface
	dataRepo        repository.CategorizationTrainingDataRepositoryInterface
	predRepo        repository.CategorizationPredictionRepositoryInterface
	metricsRepo     repository.MLModelMetricsRepositoryInterface
	contentAnalyzer *ContentAnalyzer
}

// NewMLCategorizationService creates a new ML categorization service
func NewMLCategorizationService(
	modelRepo repository.CategorizationMLModelRepositoryInterface,
	dataRepo repository.CategorizationTrainingDataRepositoryInterface,
	predRepo repository.CategorizationPredictionRepositoryInterface,
	metricsRepo repository.MLModelMetricsRepositoryInterface,
	contentAnalyzer *ContentAnalyzer,
) *MLCategorizationService {
	return &MLCategorizationService{
		modelRepo:       modelRepo,
		dataRepo:        dataRepo,
		predRepo:        predRepo,
		metricsRepo:     metricsRepo,
		contentAnalyzer: contentAnalyzer,
	}
}

// FeatureVector represents extracted features for ML prediction
type FeatureVector struct {
	// Text features
	TitleLength       int     `json:"title_length"`
	DescriptionLength int     `json:"description_length"`
	WordCount         int     `json:"word_count"`
	SentenceCount     int     `json:"sentence_count"`
	AvgWordLength     float64 `json:"avg_word_length"`

	// Sentiment features
	SentimentScore float64 `json:"sentiment_score"`

	// Entity features
	HasCurrency     bool `json:"has_currency"`
	HasDate         bool `json:"has_date"`
	HasContractRef  bool `json:"has_contract_ref"`
	HasMilestoneRef bool `json:"has_milestone_ref"`
	EntityCount     int  `json:"entity_count"`

	// Keyword features (TF-IDF like scores)
	PaymentKeywords   float64 `json:"payment_keywords"`
	ContractKeywords  float64 `json:"contract_keywords"`
	FraudKeywords     float64 `json:"fraud_keywords"`
	MilestoneKeywords float64 `json:"milestone_keywords"`
	TechnicalKeywords float64 `json:"technical_keywords"`

	// Semantic features
	PaymentSemantic   float64 `json:"payment_semantic"`
	ContractSemantic  float64 `json:"contract_semantic"`
	FraudSemantic     float64 `json:"fraud_semantic"`
	MilestoneSemantic float64 `json:"milestone_semantic"`
	TechnicalSemantic float64 `json:"technical_semantic"`

	// Urgency features
	UrgencyIndicators int  `json:"urgency_indicators"`
	HasUrgent         bool `json:"has_urgent"`
	HasCritical       bool `json:"has_critical"`

	// Complexity features
	ComplexityScore float64 `json:"complexity_score"`
	KeyPhraseCount  int     `json:"key_phrase_count"`
}

// MLPredictionResult represents the result of ML-based categorization
type MLPredictionResult struct {
	PredictedCategory models.DisputeCategory `json:"predicted_category"`
	PredictedPriority models.DisputePriority `json:"predicted_priority"`
	Confidence        float64                `json:"confidence"`
	PredictionScores  map[string]float64     `json:"prediction_scores"`
	ModelVersion      string                 `json:"model_version"`
	ResponseTime      int64                  `json:"response_time"`
}

// CreateTrainingData extracts and stores training data from existing disputes
func (s *MLCategorizationService) CreateTrainingData(ctx context.Context, dispute *models.Dispute, userID string) error {
	// Extract features from the dispute
	features, err := s.extractFeatures(dispute)
	if err != nil {
		return fmt.Errorf("failed to extract features: %w", err)
	}

	trainingData := &models.CategorizationTrainingData{
		DisputeID:   dispute.ID,
		Title:       dispute.Title,
		Description: dispute.Description,
		Category:    dispute.Category,
		Priority:    dispute.Priority,
		Features:    features,
		IsValidated: true,
		ValidatedBy: &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	trainingData.ID = uuid.New().String()
	now := time.Now()
	trainingData.ValidatedAt = &now

	return s.dataRepo.CreateTrainingData(ctx, trainingData)
}

// ExtractFeatures extracts ML features from a dispute
func (s *MLCategorizationService) extractFeatures(dispute *models.Dispute) (map[string]interface{}, error) {
	content := dispute.Title + " " + dispute.Description
	words := strings.Fields(content)
	sentences := strings.Split(content, ".")

	// Analyze content
	analysis := s.contentAnalyzer.AnalyzeContent(dispute.Title, dispute.Description)

	// Calculate basic text features
	totalWords := len(words)
	totalSentences := len(sentences)

	avgWordLength := 0.0
	if totalWords > 0 {
		for _, word := range words {
			avgWordLength += float64(len(word))
		}
		avgWordLength /= float64(totalWords)
	}

	// Calculate keyword scores (simple TF-IDF approximation)
	keywordScores := s.calculateKeywordScores(content)

	// Calculate semantic scores
	semanticScores := s.calculateSemanticScores(analysis)

	// Create feature vector
	features := &FeatureVector{
		TitleLength:       len(dispute.Title),
		DescriptionLength: len(dispute.Description),
		WordCount:         totalWords,
		SentenceCount:     totalSentences,
		AvgWordLength:     avgWordLength,
		SentimentScore:    analysis.SentimentScore,
		HasCurrency:       len(analysis.Entities["currency"]) > 0,
		HasDate:           len(analysis.Entities["date"]) > 0,
		HasContractRef:    len(analysis.Entities["contract"]) > 0,
		HasMilestoneRef:   len(analysis.Entities["milestone"]) > 0,
		EntityCount:       len(analysis.Entities),
		PaymentKeywords:   keywordScores["payment"],
		ContractKeywords:  keywordScores["contract"],
		FraudKeywords:     keywordScores["fraud"],
		MilestoneKeywords: keywordScores["milestone"],
		TechnicalKeywords: keywordScores["technical"],
		PaymentSemantic:   semanticScores["payment"],
		ContractSemantic:  semanticScores["contract"],
		FraudSemantic:     semanticScores["fraud"],
		MilestoneSemantic: semanticScores["milestone"],
		TechnicalSemantic: semanticScores["technical"],
		UrgencyIndicators: len(analysis.UrgencyIndicators),
		HasUrgent:         s.containsUrgencyWord(analysis.UrgencyIndicators, "urgent"),
		HasCritical:       s.containsUrgencyWord(analysis.UrgencyIndicators, "critical"),
		ComplexityScore:   analysis.ComplexityScore,
		KeyPhraseCount:    len(analysis.KeyPhrases),
	}

	// Convert to map for storage
	featureMap := make(map[string]interface{})
	featureJSON, err := json.Marshal(features)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(featureJSON, &featureMap)
	if err != nil {
		return nil, err
	}

	return featureMap, nil
}

// calculateKeywordScores calculates keyword density scores
func (s *MLCategorizationService) calculateKeywordScores(content string) map[string]float64 {
	content = strings.ToLower(content)
	words := strings.Fields(content)
	totalWords := float64(len(words))

	keywordSets := map[string][]string{
		"payment":   {"payment", "paid", "unpaid", "refund", "charge", "fee", "billing", "invoice", "money", "currency", "transaction", "transfer", "amount", "balance"},
		"contract":  {"contract", "agreement", "breach", "violation", "terms", "clause", "provision", "obligation", "commitment", "promise"},
		"fraud":     {"fraud", "scam", "fake", "forgery", "unauthorized", "theft", "stolen", "suspicious", "illegal", "criminal", "deceptive"},
		"milestone": {"milestone", "deliverable", "delivery", "completion", "progress", "stage", "deadline", "due", "schedule", "timeline", "delay", "overdue"},
		"technical": {"error", "bug", "system", "technical", "glitch", "failure", "crash", "timeout", "connection", "server", "database", "api", "integration"},
	}

	scores := make(map[string]float64)
	for category, keywords := range keywordSets {
		score := 0.0
		for _, keyword := range keywords {
			count := strings.Count(content, keyword)
			score += float64(count)
		}
		if totalWords > 0 {
			scores[category] = score / totalWords
		} else {
			scores[category] = 0.0
		}
	}

	return scores
}

// calculateSemanticScores calculates semantic similarity scores
func (s *MLCategorizationService) calculateSemanticScores(analysis *ContentAnalysisResult) map[string]float64 {
	scores := make(map[string]float64)

	// Use semantic categories from content analysis
	if score, exists := analysis.SemanticCategories["payment_dispute"]; exists {
		scores["payment"] = score
	}
	if score, exists := analysis.SemanticCategories["contract_dispute"]; exists {
		scores["contract"] = score
	}
	if score, exists := analysis.SemanticCategories["fraud_dispute"]; exists {
		scores["fraud"] = score
	}
	if score, exists := analysis.SemanticCategories["milestone_dispute"]; exists {
		scores["milestone"] = score
	}
	if score, exists := analysis.SemanticCategories["technical_dispute"]; exists {
		scores["technical"] = score
	}

	return scores
}

// containsUrgencyWord checks if urgency indicators contain a specific word
func (s *MLCategorizationService) containsUrgencyWord(indicators []string, word string) bool {
	for _, indicator := range indicators {
		if strings.Contains(strings.ToLower(indicator), word) {
			return true
		}
	}
	return false
}

// PredictCategorization predicts dispute category using ML models
func (s *MLCategorizationService) PredictCategorization(ctx context.Context, dispute *models.Dispute) (*MLPredictionResult, error) {
	startTime := time.Now()

	// Get the latest deployed model
	model, err := s.getLatestDeployedModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployed model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("no deployed ML model found")
	}

	// Extract features
	features, err := s.extractFeatures(dispute)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Make prediction (simplified - in real implementation, this would call actual ML model)
	prediction, err := s.makePrediction(model, features, dispute)
	if err != nil {
		return nil, fmt.Errorf("failed to make prediction: %w", err)
	}

	responseTime := time.Since(startTime).Milliseconds()

	result := &MLPredictionResult{
		PredictedCategory: prediction.Category,
		PredictedPriority: prediction.Priority,
		Confidence:        prediction.Confidence,
		PredictionScores:  prediction.Scores,
		ModelVersion:      model.Version,
		ResponseTime:      responseTime,
	}

	// Store prediction for tracking (best-effort)
	_ = s.storePrediction(ctx, dispute.ID, model.ID, result, features)

	return result, nil
}

// getLatestDeployedModel retrieves the latest deployed ML model
func (s *MLCategorizationService) getLatestDeployedModel(ctx context.Context) (*models.CategorizationMLModel, error) {
	// In a real implementation, this would query for the latest deployed model
	// For now, we'll return a mock model
	return &models.CategorizationMLModel{
		ID:        "mock-model-id",
		Name:      "Mock Categorization Model",
		Version:   "1.0.0",
		Status:    models.MLModelStatusDeployed,
		Algorithm: "mock_algorithm",
		Accuracy:  0.85,
		CreatedAt: time.Now(),
	}, nil
}

// Prediction holds prediction data
type Prediction struct {
	Category   models.DisputeCategory
	Priority   models.DisputePriority
	Confidence float64
	Scores     map[string]float64
}

// makePrediction makes a prediction using the ML model (simplified implementation)
func (s *MLCategorizationService) makePrediction(_ *models.CategorizationMLModel, features map[string]interface{}, dispute *models.Dispute) (*Prediction, error) {
	// This is a simplified implementation
	// In a real system, this would use the actual ML model to make predictions

	// Analyze content to determine category
	content := strings.ToLower(dispute.Title + " " + dispute.Description)

	// Calculate scores for each category based on features and content analysis
	scores := make(map[string]float64)

	// Payment category scoring
	paymentScore := 0.0
	if strings.Contains(content, "payment") || strings.Contains(content, "paid") {
		paymentScore += 0.8
	}
	if strings.Contains(content, "refund") || strings.Contains(content, "unpaid") {
		paymentScore += 0.6
	}
	if features["has_currency"].(bool) {
		paymentScore += 0.4
	}
	if features["payment_keywords"].(float64) > 0.01 {
		paymentScore += 0.3
	}
	scores["payment"] = paymentScore

	// Contract category scoring
	contractScore := 0.0
	if strings.Contains(content, "contract") || strings.Contains(content, "breach") {
		contractScore += 0.8
	}
	if strings.Contains(content, "violation") || strings.Contains(content, "agreement") {
		contractScore += 0.6
	}
	if features["has_contract_ref"].(bool) {
		contractScore += 0.4
	}
	if features["contract_keywords"].(float64) > 0.01 {
		contractScore += 0.3
	}
	scores["contract_breach"] = contractScore

	// Fraud category scoring
	fraudScore := 0.0
	if strings.Contains(content, "fraud") || strings.Contains(content, "scam") {
		fraudScore += 0.9
	}
	if strings.Contains(content, "unauthorized") || strings.Contains(content, "theft") {
		fraudScore += 0.7
	}
	if features["fraud_keywords"].(float64) > 0.005 {
		fraudScore += 0.4
	}
	scores["fraud"] = fraudScore

	// Milestone category scoring
	milestoneScore := 0.0
	if strings.Contains(content, "milestone") || strings.Contains(content, "delivery") {
		milestoneScore += 0.7
	}
	if strings.Contains(content, "deadline") || strings.Contains(content, "overdue") {
		milestoneScore += 0.5
	}
	if features["has_milestone_ref"].(bool) {
		milestoneScore += 0.4
	}
	if features["milestone_keywords"].(float64) > 0.01 {
		milestoneScore += 0.3
	}
	scores["milestone"] = milestoneScore

	// Technical category scoring
	technicalScore := 0.0
	if strings.Contains(content, "error") || strings.Contains(content, "system") {
		technicalScore += 0.6
	}
	if strings.Contains(content, "bug") || strings.Contains(content, "failure") {
		technicalScore += 0.5
	}
	if features["technical_keywords"].(float64) > 0.005 {
		technicalScore += 0.4
	}
	scores["technical"] = technicalScore

	// Find the category with the highest score
	maxScore := 0.0
	bestCategory := models.DisputeCategoryOther
	for categoryStr, score := range scores {
		if score > maxScore {
			maxScore = score
			switch categoryStr {
			case "payment":
				bestCategory = models.DisputeCategoryPayment
			case "contract_breach":
				bestCategory = models.DisputeCategoryContractBreach
			case "fraud":
				bestCategory = models.DisputeCategoryFraud
			case "milestone":
				bestCategory = models.DisputeCategoryMilestone
			case "technical":
				bestCategory = models.DisputeCategoryTechnical
			}
		}
	}

	// Determine priority based on urgency and amount
	priority := models.DisputePriorityNormal
	if features["urgency_indicators"].(float64) > 0 {
		priority = models.DisputePriorityHigh
	}
	if features["has_critical"].(bool) {
		priority = models.DisputePriorityUrgent
	}

	// Calculate confidence based on the score and feature strength
	confidence := math.Min(maxScore*0.8+0.2, 0.95) // Cap at 95%

	return &Prediction{
		Category:   bestCategory,
		Priority:   priority,
		Confidence: confidence,
		Scores:     scores,
	}, nil
}

// storePrediction stores a prediction for tracking and validation
func (s *MLCategorizationService) storePrediction(ctx context.Context, disputeID, modelID string, result *MLPredictionResult, features map[string]interface{}) error {
	prediction := &models.CategorizationPrediction{
		DisputeID:         disputeID,
		ModelID:           modelID,
		PredictedCategory: result.PredictedCategory,
		PredictedPriority: result.PredictedPriority,
		Confidence:        result.Confidence,
		PredictionScores:  result.PredictionScores,
		Features:          features,
		ResponseTime:      result.ResponseTime,
		CreatedAt:         time.Now(),
	}

	prediction.ID = uuid.New().String()

	return s.predRepo.CreatePrediction(ctx, prediction)
}

// ValidatePrediction validates a prediction against the actual outcome
func (s *MLCategorizationService) ValidatePrediction(ctx context.Context, predictionID string, isCorrect bool, actualCategory models.DisputeCategory, actualPriority models.DisputePriority, userID string) error {
	prediction, err := s.predRepo.GetPredictionByID(ctx, predictionID)
	if err != nil {
		return fmt.Errorf("failed to get prediction: %w", err)
	}

	now := time.Now()
	prediction.IsCorrect = &isCorrect
	prediction.CorrectCategory = &actualCategory
	prediction.CorrectPriority = &actualPriority
	prediction.ValidatedBy = &userID
	prediction.ValidatedAt = &now

	return s.predRepo.UpdatePrediction(ctx, prediction)
}

// GetModelPerformance retrieves performance metrics for ML models
func (s *MLCategorizationService) GetModelPerformance(ctx context.Context, modelID string, startDate, endDate time.Time) (*models.MLModelMetrics, error) {
	return s.metricsRepo.GetModelMetrics(ctx, modelID, startDate, endDate)
}

// TrainModel initiates model training (simplified implementation)
func (s *MLCategorizationService) TrainModel(ctx context.Context, modelName, algorithm string, userID string) error {
	// Get training data
	trainingData, err := s.dataRepo.GetValidatedTrainingData(ctx, 1000) // Get up to 1000 validated samples
	if err != nil {
		return fmt.Errorf("failed to get training data: %w", err)
	}

	if len(trainingData) < 100 {
		return fmt.Errorf("insufficient training data: need at least 100 samples, got %d", len(trainingData))
	}

	// Create new model
	model := &models.CategorizationMLModel{
		Name:             modelName,
		Description:      fmt.Sprintf("Trained model using %s algorithm", algorithm),
		Version:          s.generateModelVersion(),
		Algorithm:        algorithm,
		Status:           models.MLModelStatusTraining,
		TrainingDataSize: len(trainingData),
		CreatedBy:        userID,
		UpdatedBy:        userID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	model.ID = uuid.New().String()

	// In a real implementation, this would start an async training job
	// For now, we'll simulate training completion
	go s.simulateModelTraining(model, trainingData)

	return s.modelRepo.CreateModel(ctx, model)
}

// simulateModelTraining simulates model training (for demonstration)
func (s *MLCategorizationService) simulateModelTraining(model *models.CategorizationMLModel, _ []*models.CategorizationTrainingData) {
	// Simulate training time
	time.Sleep(5 * time.Second)

	ctx := context.Background()

	// Simulate training results
	model.Status = models.MLModelStatusTrained
	model.Accuracy = 0.82 + (0.1 * math.Sin(float64(time.Now().Unix()))) // Simulate varying accuracy
	model.Precision = 0.78
	model.Recall = 0.85
	model.F1Score = 2 * model.Precision * model.Recall / (model.Precision + model.Recall)
	model.TrainedAt = &[]time.Time{time.Now()}[0]
	model.TrainingTime = 300 // 5 minutes
	model.UpdatedAt = time.Now()

	// Update model
	if err := s.modelRepo.UpdateModel(ctx, model); err != nil {
		// Log error but don't fail the operation as this is not critical
		// TODO: Add proper logging here
		_ = err
	}
}

// generateModelVersion generates a new model version string
func (s *MLCategorizationService) generateModelVersion() string {
	now := time.Now()
	return fmt.Sprintf("v%d.%d.%d", now.Year()%100, now.Month(), now.Day())
}

// DeployModel deploys a trained model for production use
func (s *MLCategorizationService) DeployModel(ctx context.Context, modelID string, userID string) error {
	model, err := s.modelRepo.GetModelByID(ctx, modelID)
	if err != nil {
		return fmt.Errorf("failed to get model: %w", err)
	}

	if model.Status != models.MLModelStatusTrained {
		return fmt.Errorf("model is not trained and ready for deployment")
	}

	// Update model status
	now := time.Now()
	model.Status = models.MLModelStatusDeployed
	model.DeployedAt = &now
	model.UpdatedBy = userID
	model.UpdatedAt = now

	return s.modelRepo.UpdateModel(ctx, model)
}

// GetTrainingDataStats returns statistics about available training data
func (s *MLCategorizationService) GetTrainingDataStats(ctx context.Context) (*TrainingDataStats, error) {
	totalCount, err := s.dataRepo.GetTrainingDataCount(ctx)
	if err != nil {
		return nil, err
	}

	validatedCount, err := s.dataRepo.GetValidatedTrainingDataCount(ctx)
	if err != nil {
		return nil, err
	}

	categoryDistribution, err := s.dataRepo.GetTrainingDataCategoryDistribution(ctx)
	if err != nil {
		return nil, err
	}

	return &TrainingDataStats{
		TotalCount:           totalCount,
		ValidatedCount:       validatedCount,
		ValidationRate:       float64(validatedCount) / float64(totalCount) * 100,
		CategoryDistribution: categoryDistribution,
	}, nil
}

// TrainingDataStats represents statistics about training data
type TrainingDataStats struct {
	TotalCount           int64            `json:"total_count"`
	ValidatedCount       int64            `json:"validated_count"`
	ValidationRate       float64          `json:"validation_rate"`
	CategoryDistribution map[string]int64 `json:"category_distribution"`
}
