-- +goose Up
-- +goose StatementBegin

-- Add new columns to contracts table
ALTER TABLE contracts
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft',
ADD COLUMN contract_type VARCHAR(20) NOT NULL DEFAULT 'service_agreement',
ADD COLUMN version VARCHAR(20),
ADD COLUMN parent_contract_id UUID REFERENCES contracts(id),
ADD COLUMN original_filename VARCHAR(255),
ADD COLUMN file_size BIGINT,
ADD COLUMN mime_type VARCHAR(100),
ADD COLUMN expiration_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN renewal_terms TEXT;

-- Create contract_milestones table
CREATE TABLE contract_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    milestone_id UUID NOT NULL REFERENCES milestones(id),
    sequence_order INTEGER NOT NULL,
    trigger_conditions TEXT,
    verification_criteria TEXT,
    estimated_duration BIGINT, -- in nanoseconds
    actual_duration BIGINT,
    risk_level VARCHAR(20),
    criticality_score INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX idx_contract_milestones_contract_id ON contract_milestones(contract_id);
CREATE INDEX idx_contract_milestones_milestone_id ON contract_milestones(milestone_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_contract_type ON contracts(contract_type);

-- +goose StatementEnd
