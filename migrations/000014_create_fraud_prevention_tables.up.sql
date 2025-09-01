-- Create fraud prevention tables
-- Migration: 000014_create_fraud_prevention_tables.up.sql

-- Create fraud_alerts table
CREATE TABLE fraud_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    transaction_id VARCHAR(255),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'new',
    rule_id UUID,
    case_id UUID,
    
    -- Detection details
    score DECIMAL(5,4) NOT NULL,
    confidence DECIMAL(5,4) NOT NULL,
    detection_method VARCHAR(100) NOT NULL,
    evidence JSONB DEFAULT '{}',
    
    -- Alert details
    title VARCHAR(255) NOT NULL,
    description TEXT,
    recommendation TEXT,
    
    -- Notification
    notified_at TIMESTAMP WITH TIME ZONE,
    notification_channels TEXT[] DEFAULT '{}',
    
    -- Investigation
    assigned_to UUID,
    investigation_notes JSONB DEFAULT '[]',
    
    -- Timestamps
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fraud_alerts_severity_check CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT fraud_alerts_status_check CHECK (status IN ('new', 'acknowledged', 'investigating', 'resolved', 'false_positive')),
    CONSTRAINT fraud_alerts_score_check CHECK (score >= 0 AND score <= 1),
    CONSTRAINT fraud_alerts_confidence_check CHECK (confidence >= 0 AND confidence <= 1)
);

-- Create fraud_rules table
CREATE TABLE fraud_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(30) NOT NULL,
    rule_type VARCHAR(30) NOT NULL,
    
    -- Rule configuration
    conditions JSONB DEFAULT '{}',
    thresholds JSONB DEFAULT '{}',
    actions TEXT[] DEFAULT '{}',
    
    -- Scoring and severity
    base_score DECIMAL(5,4) NOT NULL DEFAULT 0.5,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0.8,
    
    -- Status and versioning
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    version INTEGER NOT NULL DEFAULT 1,
    effective_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fraud_rules_category_check CHECK (category IN ('transaction', 'behavioral', 'compliance', 'account', 'network')),
    CONSTRAINT fraud_rules_type_check CHECK (rule_type IN ('threshold', 'pattern', 'velocity', 'statistical', 'ml', 'custom')),
    CONSTRAINT fraud_rules_status_check CHECK (status IN ('active', 'inactive', 'draft')),
    CONSTRAINT fraud_rules_severity_check CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT fraud_rules_score_check CHECK (base_score >= 0 AND base_score <= 1),
    CONSTRAINT fraud_rules_confidence_check CHECK (confidence >= 0 AND confidence <= 1)
);

-- Create fraud_cases table
CREATE TABLE fraud_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE,
    case_number VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'open',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    
    -- Case details
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    
    -- Investigation
    assigned_to UUID,
    investigator UUID,
    investigation_notes JSONB DEFAULT '[]',
    
    -- Related entities
    alerts UUID[] DEFAULT '{}',
    transactions TEXT[] DEFAULT '{}',
    
    -- Outcome
    resolution JSONB,
    outcome VARCHAR(30) DEFAULT 'inconclusive',
    
    -- Timestamps
    opened_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    assigned_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT fraud_cases_status_check CHECK (status IN ('open', 'assigned', 'investigating', 'resolved', 'closed')),
    CONSTRAINT fraud_cases_priority_check CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT fraud_cases_category_check CHECK (category IN ('transaction_fraud', 'account_takeover', 'compliance', 'money_laundering', 'identity_theft')),
    CONSTRAINT fraud_cases_outcome_check CHECK (outcome IN ('confirmed_fraud', 'false_positive', 'inconclusive', 'system_error'))
);

-- Create account_fraud_status table
CREATE TABLE account_fraud_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enterprise_id UUID NOT NULL REFERENCES enterprises(id) ON DELETE CASCADE UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'normal',
    
    -- Risk assessment
    risk_score DECIMAL(5,4) NOT NULL DEFAULT 0.0,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    risk_factors TEXT[] DEFAULT '{}',
    
    -- Restrictions
    restrictions JSONB DEFAULT '[]',
    limits JSONB DEFAULT '{}',
    
    -- Monitoring
    monitoring_level VARCHAR(20) NOT NULL DEFAULT 'standard',
    review_frequency INTERVAL NOT NULL DEFAULT '30 days',
    next_review_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() + INTERVAL '30 days'),
    
    -- History
    status_history JSONB DEFAULT '[]',
    
    -- Timestamps
    status_changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT account_fraud_status_status_check CHECK (status IN ('normal', 'under_review', 'restricted', 'suspended', 'frozen', 'terminated')),
    CONSTRAINT account_fraud_status_risk_level_check CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT account_fraud_status_monitoring_level_check CHECK (monitoring_level IN ('standard', 'enhanced', 'intensive')),
    CONSTRAINT account_fraud_status_risk_score_check CHECK (risk_score >= 0 AND risk_score <= 1)
);

-- Create indexes for performance
CREATE INDEX idx_fraud_alerts_enterprise_id ON fraud_alerts(enterprise_id);
CREATE INDEX idx_fraud_alerts_status ON fraud_alerts(status);
CREATE INDEX idx_fraud_alerts_severity ON fraud_alerts(severity);
CREATE INDEX idx_fraud_alerts_detected_at ON fraud_alerts(detected_at);
CREATE INDEX idx_fraud_alerts_alert_type ON fraud_alerts(alert_type);
CREATE INDEX idx_fraud_alerts_case_id ON fraud_alerts(case_id) WHERE case_id IS NOT NULL;
CREATE INDEX idx_fraud_alerts_rule_id ON fraud_alerts(rule_id) WHERE rule_id IS NOT NULL;

CREATE INDEX idx_fraud_rules_status ON fraud_rules(status);
CREATE INDEX idx_fraud_rules_category ON fraud_rules(category);
CREATE INDEX idx_fraud_rules_effective_at ON fraud_rules(effective_at);
CREATE INDEX idx_fraud_rules_created_by ON fraud_rules(created_by);
CREATE INDEX idx_fraud_rules_active ON fraud_rules(id) WHERE status = 'active' AND effective_at <= NOW() AND (expires_at IS NULL OR expires_at > NOW());

CREATE INDEX idx_fraud_cases_enterprise_id ON fraud_cases(enterprise_id);
CREATE INDEX idx_fraud_cases_status ON fraud_cases(status);
CREATE INDEX idx_fraud_cases_priority ON fraud_cases(priority);
CREATE INDEX idx_fraud_cases_category ON fraud_cases(category);
CREATE INDEX idx_fraud_cases_opened_at ON fraud_cases(opened_at);
CREATE INDEX idx_fraud_cases_case_number ON fraud_cases(case_number);

CREATE INDEX idx_account_fraud_status_status ON account_fraud_status(status);
CREATE INDEX idx_account_fraud_status_risk_level ON account_fraud_status(risk_level);
CREATE INDEX idx_account_fraud_status_next_review_date ON account_fraud_status(next_review_date);

-- Create GIN indexes for JSON fields
CREATE INDEX idx_fraud_alerts_evidence_gin ON fraud_alerts USING GIN (evidence);
CREATE INDEX idx_fraud_rules_conditions_gin ON fraud_rules USING GIN (conditions);
CREATE INDEX idx_fraud_rules_thresholds_gin ON fraud_rules USING GIN (thresholds);
CREATE INDEX idx_fraud_cases_resolution_gin ON fraud_cases USING GIN (resolution);
CREATE INDEX idx_account_fraud_status_restrictions_gin ON account_fraud_status USING GIN (restrictions);
CREATE INDEX idx_account_fraud_status_limits_gin ON account_fraud_status USING GIN (limits);

-- Create full-text search indexes
CREATE INDEX idx_fraud_alerts_title_desc_fts ON fraud_alerts USING GIN (to_tsvector('english', title || ' ' || COALESCE(description, '')));
CREATE INDEX idx_fraud_rules_name_desc_fts ON fraud_rules USING GIN (to_tsvector('english', name || ' ' || COALESCE(description, '')));
CREATE INDEX idx_fraud_cases_title_desc_fts ON fraud_cases USING GIN (to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Create triggers to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_fraud_prevention_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_fraud_alerts_updated_at 
    BEFORE UPDATE ON fraud_alerts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_fraud_prevention_updated_at();

CREATE TRIGGER update_fraud_rules_updated_at 
    BEFORE UPDATE ON fraud_rules 
    FOR EACH ROW 
    EXECUTE FUNCTION update_fraud_prevention_updated_at();

CREATE TRIGGER update_fraud_cases_updated_at 
    BEFORE UPDATE ON fraud_cases 
    FOR EACH ROW 
    EXECUTE FUNCTION update_fraud_prevention_updated_at();

CREATE TRIGGER update_account_fraud_status_updated_at 
    BEFORE UPDATE ON account_fraud_status 
    FOR EACH ROW 
    EXECUTE FUNCTION update_fraud_prevention_updated_at();

-- Create views for common queries
CREATE VIEW active_fraud_rules AS
SELECT * FROM fraud_rules 
WHERE status = 'active' 
  AND effective_at <= NOW() 
  AND (expires_at IS NULL OR expires_at > NOW());

CREATE VIEW high_priority_fraud_cases AS
SELECT * FROM fraud_cases 
WHERE priority IN ('high', 'critical') 
  AND status IN ('open', 'assigned', 'investigating');

CREATE VIEW restricted_accounts AS
SELECT e.id, e.legal_name, afs.status, afs.risk_level, afs.risk_score
FROM enterprises e
JOIN account_fraud_status afs ON e.id = afs.enterprise_id
WHERE afs.status IN ('restricted', 'suspended', 'frozen');

CREATE VIEW fraud_alert_summary AS
SELECT 
    enterprise_id,
    COUNT(*) as total_alerts,
    COUNT(*) FILTER (WHERE status = 'new') as new_alerts,
    COUNT(*) FILTER (WHERE severity = 'critical') as critical_alerts,
    COUNT(*) FILTER (WHERE severity = 'high') as high_alerts,
    AVG(score) as average_score,
    MAX(detected_at) as latest_alert
FROM fraud_alerts 
GROUP BY enterprise_id;

-- Insert default fraud rules
INSERT INTO fraud_rules (
    name, description, category, rule_type, 
    conditions, thresholds, actions, base_score, severity, confidence,
    status, effective_at, created_by, updated_by
) VALUES 
(
    'High Amount Transaction',
    'Detects transactions with unusually high amounts',
    'transaction',
    'threshold',
    '{"transaction_type": ["withdrawal", "transfer"]}',
    '{"amount_threshold": 10000}',
    '{"alert", "investigate"}',
    0.7,
    'high',
    0.8,
    'active',
    NOW(),
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000'
),
(
    'Velocity Spike',
    'Detects unusual transaction frequency',
    'behavioral',
    'velocity',
    '{"time_window": "1 hour"}',
    '{"max_transactions": 10}',
    '{"alert", "monitor"}',
    0.6,
    'medium',
    0.7,
    'active',
    NOW(),
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000'
),
(
    'Off-Hours Activity',
    'Detects transactions outside normal business hours',
    'behavioral',
    'pattern',
    '{"business_hours": {"start": "06:00", "end": "22:00"}}',
    '{"risk_multiplier": 2.0}',
    '{"alert"}',
    0.4,
    'low',
    0.6,
    'active',
    NOW(),
    '00000000-0000-0000-0000-000000000000',
    '00000000-0000-0000-0000-000000000000'
);
