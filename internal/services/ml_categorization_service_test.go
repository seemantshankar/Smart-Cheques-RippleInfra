package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

type MLCategorizationServiceTestSuite struct {
	suite.Suite
	service         *MLCategorizationService
	mockModelRepo   *MockMLModelRepository
	mockDataRepo    *MockTrainingDataRepository
	mockPredRepo    *MockPredictionRepository
	mockMetricsRepo *MockMetricsRepository
	contentAnalyzer *ContentAnalyzer
	ctx             context.Context
}

type MockMLModelRepository struct {
	models map[string]*models.CategorizationMLModel
}

func NewMockMLModelRepository() *MockMLModelRepository {
	return &MockMLModelRepository{
		models: make(map[string]*models.CategorizationMLModel),
	}
}

func (m *MockMLModelRepository) CreateModel(ctx context.Context, model *models.CategorizationMLModel) error {
	model.ID = uuid.New().String()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()
	m.models[model.ID] = model
	return nil
}

func (m *MockMLModelRepository) GetModelByID(ctx context.Context, id string) (*models.CategorizationMLModel, error) {
	if model, exists := m.models[id]; exists {
		return model, nil
	}
	return nil, nil
}

func (m *MockMLModelRepository) UpdateModel(ctx context.Context, model *models.CategorizationMLModel) error {
	if _, exists := m.models[model.ID]; exists {
		model.UpdatedAt = time.Now()
		m.models[model.ID] = model
		return nil
	}
	return nil
}

func (m *MockMLModelRepository) DeleteModel(ctx context.Context, id string) error {
	delete(m.models, id)
	return nil
}

func (m *MockMLModelRepository) GetModels(ctx context.Context, filter *repository.MLModelFilter, limit, offset int) ([]*models.CategorizationMLModel, error) {
	var models []*models.CategorizationMLModel
	for _, model := range m.models {
		models = append(models, model)
	}
	return models, nil
}

func (m *MockMLModelRepository) GetLatestDeployedModel(ctx context.Context) (*models.CategorizationMLModel, error) {
	for _, model := range m.models {
		if model.Status == models.MLModelStatusDeployed {
			return model, nil
		}
	}
	// Return mock model if none exists - this simulates the behavior in the service
	return &models.CategorizationMLModel{
		ID:        "mock-model-id",
		Name:      "Mock Categorization Model",
		Version:   "1.0.0",
		Status:    models.MLModelStatusDeployed,
		Accuracy:  0.85,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockMLModelRepository) GetModelsByStatus(ctx context.Context, status models.MLModelStatus, limit, offset int) ([]*models.CategorizationMLModel, error) {
	var models []*models.CategorizationMLModel
	for _, model := range m.models {
		if model.Status == status {
			models = append(models, model)
		}
	}
	return models, nil
}

// Mock Training Data Repository
type MockTrainingDataRepository struct {
	data map[string]*models.CategorizationTrainingData
}

func NewMockTrainingDataRepository() *MockTrainingDataRepository {
	return &MockTrainingDataRepository{
		data: make(map[string]*models.CategorizationTrainingData),
	}
}

func (m *MockTrainingDataRepository) CreateTrainingData(ctx context.Context, data *models.CategorizationTrainingData) error {
	data.ID = uuid.New().String()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	m.data[data.ID] = data
	return nil
}

func (m *MockTrainingDataRepository) GetTrainingDataByID(ctx context.Context, id string) (*models.CategorizationTrainingData, error) {
	if data, exists := m.data[id]; exists {
		return data, nil
	}
	return nil, nil
}

func (m *MockTrainingDataRepository) UpdateTrainingData(ctx context.Context, data *models.CategorizationTrainingData) error {
	if _, exists := m.data[data.ID]; exists {
		data.UpdatedAt = time.Now()
		m.data[data.ID] = data
		return nil
	}
	return nil
}

func (m *MockTrainingDataRepository) DeleteTrainingData(ctx context.Context, id string) error {
	delete(m.data, id)
	return nil
}

func (m *MockTrainingDataRepository) GetTrainingData(ctx context.Context, filter *repository.TrainingDataFilter, limit, offset int) ([]*models.CategorizationTrainingData, error) {
	var data []*models.CategorizationTrainingData
	for _, item := range m.data {
		data = append(data, item)
	}
	return data, nil
}

func (m *MockTrainingDataRepository) GetValidatedTrainingData(ctx context.Context, limit int) ([]*models.CategorizationTrainingData, error) {
	var validated []*models.CategorizationTrainingData
	for _, item := range m.data {
		if item.IsValidated {
			validated = append(validated, item)
		}
	}
	return validated, nil
}

func (m *MockTrainingDataRepository) GetTrainingDataCount(ctx context.Context) (int64, error) {
	return int64(len(m.data)), nil
}

func (m *MockTrainingDataRepository) GetValidatedTrainingDataCount(ctx context.Context) (int64, error) {
	count := 0
	for _, item := range m.data {
		if item.IsValidated {
			count++
		}
	}
	return int64(count), nil
}

func (m *MockTrainingDataRepository) GetTrainingDataCategoryDistribution(ctx context.Context) (map[string]int64, error) {
	distribution := make(map[string]int64)
	for _, item := range m.data {
		distribution[string(item.Category)]++
	}
	return distribution, nil
}

func (m *MockTrainingDataRepository) BulkCreateTrainingData(ctx context.Context, data []*models.CategorizationTrainingData) error {
	for _, item := range data {
		item.ID = uuid.New().String()
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
		m.data[item.ID] = item
	}
	return nil
}

// Mock Prediction Repository
type MockPredictionRepository struct {
	predictions map[string]*models.CategorizationPrediction
}

func NewMockPredictionRepository() *MockPredictionRepository {
	return &MockPredictionRepository{
		predictions: make(map[string]*models.CategorizationPrediction),
	}
}

func (m *MockPredictionRepository) CreatePrediction(ctx context.Context, prediction *models.CategorizationPrediction) error {
	prediction.ID = uuid.New().String()
	prediction.CreatedAt = time.Now()
	m.predictions[prediction.ID] = prediction
	return nil
}

func (m *MockPredictionRepository) GetPredictionByID(ctx context.Context, id string) (*models.CategorizationPrediction, error) {
	if prediction, exists := m.predictions[id]; exists {
		return prediction, nil
	}
	return nil, nil
}

func (m *MockPredictionRepository) UpdatePrediction(ctx context.Context, prediction *models.CategorizationPrediction) error {
	if _, exists := m.predictions[prediction.ID]; exists {
		m.predictions[prediction.ID] = prediction
		return nil
	}
	return nil
}

func (m *MockPredictionRepository) GetPredictionsByDisputeID(ctx context.Context, disputeID string, limit, offset int) ([]*models.CategorizationPrediction, error) {
	var predictions []*models.CategorizationPrediction
	for _, prediction := range m.predictions {
		if prediction.DisputeID == disputeID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *MockPredictionRepository) GetPredictionsByModelID(ctx context.Context, modelID string, limit, offset int) ([]*models.CategorizationPrediction, error) {
	var predictions []*models.CategorizationPrediction
	for _, prediction := range m.predictions {
		if prediction.ModelID == modelID {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *MockPredictionRepository) GetUnvalidatedPredictions(ctx context.Context, limit, offset int) ([]*models.CategorizationPrediction, error) {
	var predictions []*models.CategorizationPrediction
	for _, prediction := range m.predictions {
		if prediction.IsCorrect == nil {
			predictions = append(predictions, prediction)
		}
	}
	return predictions, nil
}

func (m *MockPredictionRepository) GetPredictionAccuracy(ctx context.Context, modelID string, startDate, endDate time.Time) (float64, error) {
	total := 0
	correct := 0
	for _, prediction := range m.predictions {
		if prediction.ModelID == modelID && prediction.CreatedAt.After(startDate) && prediction.CreatedAt.Before(endDate) {
			total++
			if prediction.IsCorrect != nil && *prediction.IsCorrect {
				correct++
			}
		}
	}
	if total == 0 {
		return 0.0, nil
	}
	return float64(correct) / float64(total), nil
}

// Mock Metrics Repository
type MockMetricsRepository struct {
	metrics map[string]*models.MLModelMetrics
}

func NewMockMetricsRepository() *MockMetricsRepository {
	return &MockMetricsRepository{
		metrics: make(map[string]*models.MLModelMetrics),
	}
}

func (m *MockMetricsRepository) CreateModelMetrics(ctx context.Context, metrics *models.MLModelMetrics) error {
	metrics.ID = uuid.New().String()
	metrics.CreatedAt = time.Now()
	m.metrics[metrics.ID] = metrics
	return nil
}

func (m *MockMetricsRepository) GetModelMetrics(ctx context.Context, modelID string, startDate, endDate time.Time) (*models.MLModelMetrics, error) {
	for _, metrics := range m.metrics {
		if metrics.ModelID == modelID && metrics.PeriodStart.Equal(startDate) && metrics.PeriodEnd.Equal(endDate) {
			return metrics, nil
		}
	}
	return nil, nil
}

func (m *MockMetricsRepository) UpdateModelMetrics(ctx context.Context, metrics *models.MLModelMetrics) error {
	if _, exists := m.metrics[metrics.ID]; exists {
		m.metrics[metrics.ID] = metrics
		return nil
	}
	return nil
}

func (m *MockMetricsRepository) GetModelMetricsHistory(ctx context.Context, modelID string, limit, offset int) ([]*models.MLModelMetrics, error) {
	var metrics []*models.MLModelMetrics
	for _, metric := range m.metrics {
		if metric.ModelID == modelID {
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}

func (m *MockMetricsRepository) GetLatestModelMetrics(ctx context.Context, modelID string) (*models.MLModelMetrics, error) {
	var latest *models.MLModelMetrics
	for _, metrics := range m.metrics {
		if metrics.ModelID == modelID {
			if latest == nil || metrics.CreatedAt.After(latest.CreatedAt) {
				latest = metrics
			}
		}
	}
	return latest, nil
}

func (suite *MLCategorizationServiceTestSuite) SetupTest() {
	suite.mockModelRepo = NewMockMLModelRepository()
	suite.mockDataRepo = NewMockTrainingDataRepository()
	suite.mockPredRepo = NewMockPredictionRepository()
	suite.mockMetricsRepo = NewMockMetricsRepository()
	suite.contentAnalyzer = NewContentAnalyzer()
	suite.service = NewMLCategorizationService(
		suite.mockModelRepo,
		suite.mockDataRepo,
		suite.mockPredRepo,
		suite.mockMetricsRepo,
		suite.contentAnalyzer,
	)
	suite.ctx = context.Background()
}

func (suite *MLCategorizationServiceTestSuite) TestCreateTrainingData() {
	dispute := &models.Dispute{
		ID:          "test-dispute-1",
		Title:       "Payment Issue",
		Description: "We haven't received payment for our services",
		Category:    models.DisputeCategoryPayment,
		Priority:    models.DisputePriorityHigh,
	}

	err := suite.service.CreateTrainingData(suite.ctx, dispute, "test-user")

	assert.NoError(suite.T(), err)

	// Verify training data was created
	trainingData, err := suite.mockDataRepo.GetTrainingData(suite.ctx, nil, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), trainingData, 1)
	assert.Equal(suite.T(), dispute.ID, trainingData[0].DisputeID)
	assert.True(suite.T(), trainingData[0].IsValidated)
}

func (suite *MLCategorizationServiceTestSuite) TestExtractFeatures() {
	dispute := &models.Dispute{
		Title:       "Payment Not Received",
		Description: "We have not received the payment of $50,000 for services rendered. This is urgent.",
	}

	features, err := suite.service.extractFeatures(dispute)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), features)

	// Check that features contain expected keys
	assert.Contains(suite.T(), features, "title_length")
	assert.Contains(suite.T(), features, "description_length")
	assert.Contains(suite.T(), features, "has_currency")
	assert.Contains(suite.T(), features, "urgency_indicators")
}

func (suite *MLCategorizationServiceTestSuite) TestPredictCategorization() {
	// Create and deploy a mock model
	model := &models.CategorizationMLModel{
		Name:      "Test Model",
		Version:   "v1.0.0",
		Algorithm: "test_algorithm",
		Status:    models.MLModelStatusDeployed,
		Accuracy:  0.85,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.mockModelRepo.CreateModel(suite.ctx, model)
	assert.NoError(suite.T(), err)

	dispute := &models.Dispute{
		ID:          "test-dispute-1",
		Title:       "Payment Not Received",
		Description: "We have not received payment for services rendered. This is causing cash flow issues.",
	}

	result, err := suite.service.PredictCategorization(suite.ctx, dispute)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), models.DisputeCategoryPayment, result.PredictedCategory)
	assert.Greater(suite.T(), result.Confidence, 0.0)
	assert.LessOrEqual(suite.T(), result.Confidence, 1.0)
	assert.NotEmpty(suite.T(), result.ModelVersion)
	assert.GreaterOrEqual(suite.T(), result.ResponseTime, int64(0)) // ResponseTime can be 0 in tests
}

func (suite *MLCategorizationServiceTestSuite) TestPredictCategorizationNoModel() {
	// Clear any existing models
	suite.mockModelRepo = NewMockMLModelRepository()
	suite.service = NewMLCategorizationService(
		suite.mockModelRepo,
		suite.mockDataRepo,
		suite.mockPredRepo,
		suite.mockMetricsRepo,
		suite.contentAnalyzer,
	)

	dispute := &models.Dispute{
		Title:       "Test Dispute",
		Description: "Test description",
	}

	// Test without any deployed model - should still work with mock fallback
	result, err := suite.service.PredictCategorization(suite.ctx, dispute)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), models.DisputeCategoryOther, result.PredictedCategory) // Default fallback
	assert.GreaterOrEqual(suite.T(), result.Confidence, 0.0)
}

func (suite *MLCategorizationServiceTestSuite) TestValidatePrediction() {
	// Create a prediction
	prediction := &models.CategorizationPrediction{
		DisputeID:         "test-dispute-1",
		ModelID:           "test-model-1",
		PredictedCategory: models.DisputeCategoryPayment,
		PredictedPriority: models.DisputePriorityHigh,
		Confidence:        0.8,
		ResponseTime:      100,
	}

	err := suite.mockPredRepo.CreatePrediction(suite.ctx, prediction)
	assert.NoError(suite.T(), err)

	// Validate the prediction
	err = suite.service.ValidatePrediction(suite.ctx, prediction.ID, true, models.DisputeCategoryPayment, models.DisputePriorityHigh, "validator-user")

	assert.NoError(suite.T(), err)

	// Verify validation
	validatedPrediction, err := suite.mockPredRepo.GetPredictionByID(suite.ctx, prediction.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), validatedPrediction.IsCorrect)
	assert.True(suite.T(), *validatedPrediction.IsCorrect)
	assert.Equal(suite.T(), "validator-user", *validatedPrediction.ValidatedBy)
}

func (suite *MLCategorizationServiceTestSuite) TestGetTrainingDataStats() {
	// Create some training data
	data1 := &models.CategorizationTrainingData{
		DisputeID:   "dispute-1",
		Title:       "Payment Issue",
		Description: "Payment not received",
		Category:    models.DisputeCategoryPayment,
		Priority:    models.DisputePriorityHigh,
		IsValidated: true,
		Features:    map[string]interface{}{"test": "value"},
	}

	data2 := &models.CategorizationTrainingData{
		DisputeID:   "dispute-2",
		Title:       "Contract Breach",
		Description: "Breach of contract",
		Category:    models.DisputeCategoryContractBreach,
		Priority:    models.DisputePriorityHigh,
		IsValidated: false,
		Features:    map[string]interface{}{"test": "value"},
	}

	err := suite.mockDataRepo.CreateTrainingData(suite.ctx, data1)
	assert.NoError(suite.T(), err)
	err = suite.mockDataRepo.CreateTrainingData(suite.ctx, data2)
	assert.NoError(suite.T(), err)

	stats, err := suite.service.GetTrainingDataStats(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.Equal(suite.T(), int64(2), stats.TotalCount)
	assert.Equal(suite.T(), int64(1), stats.ValidatedCount)
	assert.Equal(suite.T(), 50.0, stats.ValidationRate)
	assert.Contains(suite.T(), stats.CategoryDistribution, "payment")
	assert.Contains(suite.T(), stats.CategoryDistribution, "contract_breach")
}

func (suite *MLCategorizationServiceTestSuite) TestTrainModelInsufficientData() {
	err := suite.service.TrainModel(suite.ctx, "Test Model", "test_algorithm", "test-user")

	// Should fail due to insufficient training data
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "insufficient training data")
}

func (suite *MLCategorizationServiceTestSuite) TestTrainModel() {
	// Create sufficient training data
	for i := 0; i < 110; i++ {
		data := &models.CategorizationTrainingData{
			DisputeID:   uuid.New().String(),
			Title:       "Test Dispute",
			Description: "Test description",
			Category:    models.DisputeCategoryPayment,
			Priority:    models.DisputePriorityNormal,
			IsValidated: true,
			Features:    map[string]interface{}{"test": "value"},
		}
		err := suite.mockDataRepo.CreateTrainingData(suite.ctx, data)
		assert.NoError(suite.T(), err)
	}

	err := suite.service.TrainModel(suite.ctx, "Test Model", "test_algorithm", "test-user")

	assert.NoError(suite.T(), err)

	// Verify model was created
	models, err := suite.mockModelRepo.GetModels(suite.ctx, nil, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), models, 1)
	assert.Equal(suite.T(), "Test Model", models[0].Name)
	assert.Equal(suite.T(), "training", string(models[0].Status))
}

func (suite *MLCategorizationServiceTestSuite) TestDeployModel() {
	// Create a trained model
	model := &models.CategorizationMLModel{
		Name:      "Test Model",
		Version:   "v1.0.0",
		Algorithm: "test_algorithm",
		Status:    models.MLModelStatusTrained,
		Accuracy:  0.85,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.mockModelRepo.CreateModel(suite.ctx, model)
	assert.NoError(suite.T(), err)

	// Deploy the model
	err = suite.service.DeployModel(suite.ctx, model.ID, "deploy-user")

	assert.NoError(suite.T(), err)

	// Verify deployment
	deployedModel, err := suite.mockModelRepo.GetModelByID(suite.ctx, model.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), models.MLModelStatusDeployed, deployedModel.Status)
	assert.NotNil(suite.T(), deployedModel.DeployedAt)
}

func (suite *MLCategorizationServiceTestSuite) TestDeployModelNotTrained() {
	// Create a model that's not trained
	model := &models.CategorizationMLModel{
		Name:      "Test Model",
		Version:   "v1.0.0",
		Algorithm: "test_algorithm",
		Status:    models.MLModelStatusTraining,
		CreatedBy: "test-user",
		UpdatedBy: "test-user",
	}

	err := suite.mockModelRepo.CreateModel(suite.ctx, model)
	assert.NoError(suite.T(), err)

	// Try to deploy untrained model
	err = suite.service.DeployModel(suite.ctx, model.ID, "deploy-user")

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not trained and ready for deployment")
}

func (suite *MLCategorizationServiceTestSuite) TestCalculateKeywordScores() {
	content := "payment not received unpaid invoice billing charge"

	scores := suite.service.calculateKeywordScores(content)

	// Payment keywords should be found
	assert.Greater(suite.T(), scores["payment"], 0.0)
	// Contract keywords might have some matches due to "invoice"
	// Fraud keywords should be 0 for this content
	assert.Equal(suite.T(), 0.0, scores["fraud"])
}

func (suite *MLCategorizationServiceTestSuite) TestContainsUrgencyWord() {
	indicators := []string{"urgent", "critical", "blocking"}

	assert.True(suite.T(), suite.service.containsUrgencyWord(indicators, "urgent"))
	assert.True(suite.T(), suite.service.containsUrgencyWord(indicators, "critical"))
	assert.False(suite.T(), suite.service.containsUrgencyWord(indicators, "normal"))
}

func TestMLCategorizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MLCategorizationServiceTestSuite))
}
