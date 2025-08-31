-- +goose Up
-- +goose StatementBegin

-- Add new milestone columns to contract_milestones
ALTER TABLE contract_milestones
ADD COLUMN sequence_number INTEGER,
ADD COLUMN category VARCHAR(50),
ADD COLUMN priority INTEGER,
ADD COLUMN critical_path BOOLEAN DEFAULT FALSE,
ADD COLUMN estimated_start_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN estimated_end_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN actual_start_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN actual_end_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN percentage_complete DOUBLE PRECISION DEFAULT 0,
ADD COLUMN contingency_plans JSONB;

-- Create milestone_templates table for reusable milestone patterns
CREATE TABLE IF NOT EXISTS milestone_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_category VARCHAR(50),
    default_priority INTEGER,
    variables JSONB,
    version VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Create milestone_dependencies to represent relationships between milestones
CREATE TABLE IF NOT EXISTS milestone_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    milestone_id UUID NOT NULL REFERENCES contract_milestones(id),
    depends_on_id UUID NOT NULL REFERENCES contract_milestones(id),
    dependency_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_contract_milestones_sequence_number ON contract_milestones(sequence_number);
CREATE INDEX IF NOT EXISTS idx_milestone_templates_name ON milestone_templates(name);

-- +goose StatementEnd



