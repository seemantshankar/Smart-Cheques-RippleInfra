-- Create ML categorization tables

-- ML Models table
CREATE TABLE categorization_ml_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version VARCHAR(20) NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'training' CHECK (status IN ('training', 'trained', 'failed', 'deployed', 'retired')),
    training_data_size INTEGER NOT NULL DEFAULT 0,
    accuracy DECIMAL(5,4),
    precision DECIMAL(5,4),
    recall DECIMAL(5,4),
    f1_score DECIMAL(5,4),
    parameters JSONB,
    feature_names TEXT[],
    model_data BYTEA,
    model_path VARCHAR(500),
    trained_at TIMESTAMP WITH TIME ZONE,
    training_time BIGINT DEFAULT 0,
    deployed_at TIMESTAMP WITH TIME ZONE,
    use_count BIGINT NOT NULL DEFAULT 0,
    success_count BIGINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for ML models
CREATE INDEX idx_categorization_ml_models_status ON categorization_ml_models(status);
CREATE INDEX idx_categorization_ml_models_algorithm ON categorization_ml_models(algorithm);
CREATE INDEX idx_categorization_ml_models_accuracy ON categorization_ml_models(accuracy DESC);
CREATE INDEX idx_categorization_ml_models_created_at ON categorization_ml_models(created_at DESC);
CREATE INDEX idx_categorization_ml_models_deployed_at ON categorization_ml_models(deployed_at DESC);

-- Training Data table
CREATE TABLE categorization_training_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    features JSONB NOT NULL,
    is_validated BOOLEAN NOT NULL DEFAULT false,
    validated_by VARCHAR(255),
    validated_at TIMESTAMP WITH TIME ZONE,
    use_count BIGINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(dispute_id)
);

-- Create indexes for training data
CREATE INDEX idx_categorization_training_data_category ON categorization_training_data(category);
CREATE INDEX idx_categorization_training_data_is_validated ON categorization_training_data(is_validated);
CREATE INDEX idx_categorization_training_data_validated_at ON categorization_training_data(validated_at);
CREATE INDEX idx_categorization_training_data_created_at ON categorization_training_data(created_at DESC);
CREATE INDEX idx_categorization_training_data_features_gin ON categorization_training_data USING GIN(features);

-- Predictions table
CREATE TABLE categorization_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id VARCHAR(255) NOT NULL,
    model_id UUID NOT NULL REFERENCES categorization_ml_models(id) ON DELETE CASCADE,
    predicted_category VARCHAR(50) NOT NULL,
    predicted_priority VARCHAR(20) NOT NULL,
    confidence DECIMAL(5,4) NOT NULL,
    prediction_scores JSONB,
    features JSONB,
    is_correct BOOLEAN,
    correct_category VARCHAR(50),
    correct_priority VARCHAR(20),
    validated_by VARCHAR(255),
    validated_at TIMESTAMP WITH TIME ZONE,
    response_time BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for predictions
CREATE INDEX idx_categorization_predictions_dispute_id ON categorization_predictions(dispute_id);
CREATE INDEX idx_categorization_predictions_model_id ON categorization_predictions(model_id);
CREATE INDEX idx_categorization_predictions_confidence ON categorization_predictions(confidence DESC);
CREATE INDEX idx_categorization_predictions_is_correct ON categorization_predictions(is_correct);
CREATE INDEX idx_categorization_predictions_created_at ON categorization_predictions(created_at DESC);

-- ML Model Metrics table
CREATE TABLE ml_model_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES categorization_ml_models(id) ON DELETE CASCADE,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    total_predictions BIGINT NOT NULL DEFAULT 0,
    correct_predictions BIGINT NOT NULL DEFAULT 0,
    accuracy DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    category_accuracy JSONB,
    category_precision JSONB,
    category_recall JSONB,
    avg_response_time BIGINT NOT NULL DEFAULT 0,
    min_response_time BIGINT NOT NULL DEFAULT 0,
    max_response_time BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, period_start, period_end)
);

-- Create indexes for model metrics
CREATE INDEX idx_ml_model_metrics_model_id ON ml_model_metrics(model_id);
CREATE INDEX idx_ml_model_metrics_period ON ml_model_metrics(period_start, period_end);
CREATE INDEX idx_ml_model_metrics_accuracy ON ml_model_metrics(accuracy DESC);
CREATE INDEX idx_ml_model_metrics_created_at ON ml_model_metrics(created_at DESC);

-- Insert sample ML model
INSERT INTO categorization_ml_models (
    name, description, version, algorithm, status, training_data_size,
    accuracy, precision, recall, f1_score, created_by, updated_by
) VALUES (
    'Baseline Categorization Model',
    'Initial ML model for dispute categorization using rule-based features',
    'v1.0.0',
    'rule_based_ensemble',
    'deployed',
    1000,
    0.82,
    0.79,
    0.84,
    0.81,
    'system',
    'system'
);

-- Insert sample training data (mock data for demonstration)
INSERT INTO categorization_training_data (
    dispute_id, title, description, category, priority, features, is_validated, validated_by
) VALUES
(
    'sample-dispute-1',
    'Payment Not Received',
    'We have not received the payment of $50,000 for the goods delivered as per contract ABC-123. The payment was due on January 15th, 2024, and we are experiencing cash flow issues.',
    'payment',
    'high',
    '{"title_length": 19, "description_length": 150, "word_count": 35, "has_currency": true, "has_contract_ref": true, "has_date": true, "payment_keywords": 0.15, "contract_keywords": 0.08, "urgency_indicators": 1}',
    true,
    'system'
),
(
    'sample-dispute-2',
    'Contract Breach by Vendor',
    'The vendor failed to deliver the contracted services within the agreed 30-day timeframe. This constitutes a material breach of contract terms requiring immediate resolution.',
    'contract_breach',
    'urgent',
    '{"title_length": 25, "description_length": 140, "word_count": 32, "has_contract_ref": false, "contract_keywords": 0.22, "urgency_indicators": 2, "complexity_score": 0.75}',
    true,
    'system'
),
(
    'sample-dispute-3',
    'Unauthorized Transaction Detected',
    'We detected suspicious activity on our account with unauthorized transactions totaling $25,000. This appears to be fraud and requires immediate investigation.',
    'fraud',
    'urgent',
    '{"title_length": 32, "description_length": 135, "word_count": 28, "has_currency": true, "fraud_keywords": 0.12, "urgency_indicators": 2, "sentiment_score": -0.8}',
    true,
    'system'
);

-- Create trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_ml_tables_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_categorization_ml_models_updated_at BEFORE UPDATE ON categorization_ml_models FOR EACH ROW EXECUTE FUNCTION update_ml_tables_updated_at_column();
CREATE TRIGGER update_categorization_training_data_updated_at BEFORE UPDATE ON categorization_training_data FOR EACH ROW EXECUTE FUNCTION update_ml_tables_updated_at_column();
