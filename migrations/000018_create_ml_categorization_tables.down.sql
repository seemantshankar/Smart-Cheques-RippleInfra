-- Drop ML categorization tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_categorization_ml_models_updated_at ON categorization_ml_models;
DROP TRIGGER IF EXISTS update_categorization_training_data_updated_at ON categorization_training_data;

-- Drop function
DROP FUNCTION IF EXISTS update_ml_tables_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS ml_model_metrics;
DROP TABLE IF EXISTS categorization_predictions;
DROP TABLE IF EXISTS categorization_training_data;
DROP TABLE IF EXISTS categorization_ml_models;
