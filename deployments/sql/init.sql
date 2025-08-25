-- Smart Payment Infrastructure Database Schema
-- Initial setup for PostgreSQL

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enterprises table
CREATE TABLE IF NOT EXISTS enterprises (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    legal_name VARCHAR(255) NOT NULL,
    jurisdiction VARCHAR(100) NOT NULL,
    kyb_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    xrpl_wallet VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT kyb_status_check CHECK (kyb_status IN ('pending', 'verified', 'rejected'))
);

-- Smart Cheques table
CREATE TABLE IF NOT EXISTS smart_cheques (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payer_id UUID NOT NULL REFERENCES enterprises(id),
    payee_id UUID NOT NULL REFERENCES enterprises(id),
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    escrow_address VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    contract_hash VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT currency_check CHECK (currency IN ('USDT', 'USDC', 'e₹')),
    CONSTRAINT status_check CHECK (status IN ('created', 'locked', 'in_progress', 'completed', 'disputed'))
);

-- Contracts table
CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parties TEXT[] NOT NULL,
    contract_hash VARCHAR(100) UNIQUE,
    ai_analysis_confidence DECIMAL(3, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Milestones table
CREATE TABLE IF NOT EXISTS milestones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    smart_cheque_id UUID NOT NULL REFERENCES smart_cheques(id),
    description TEXT NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    verification_method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT verification_method_check CHECK (verification_method IN ('oracle', 'manual', 'hybrid')),
    CONSTRAINT milestone_status_check CHECK (status IN ('pending', 'verified', 'failed'))
);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_enterprises_kyb_status ON enterprises(kyb_status);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_payer_id ON smart_cheques(payer_id);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_payee_id ON smart_cheques(payee_id);
CREATE INDEX IF NOT EXISTS idx_smart_cheques_status ON smart_cheques(status);
CREATE INDEX IF NOT EXISTS idx_milestones_smart_cheque_id ON milestones(smart_cheque_id);
CREATE INDEX IF NOT EXISTS idx_milestones_status ON milestones(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_enterprises_updated_at BEFORE UPDATE ON enterprises FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_smart_cheques_updated_at BEFORE UPDATE ON smart_cheques FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_contracts_updated_at BEFORE UPDATE ON contracts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_milestones_updated_at BEFORE UPDATE ON milestones FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();