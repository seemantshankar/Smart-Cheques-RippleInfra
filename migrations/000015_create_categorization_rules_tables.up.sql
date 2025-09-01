-- Create categorization rules tables

-- Categorization rules table
CREATE TABLE categorization_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('keyword', 'pattern', 'semantic', 'entity', 'composite')),
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    keywords TEXT[],
    patterns TEXT[],
    entities TEXT[],
    semantic_keys TEXT[],
    conditions JSONB,
    base_confidence DECIMAL(3,2) NOT NULL DEFAULT 0.5 CHECK (base_confidence >= 0 AND base_confidence <= 1),
    weight DECIMAL(3,2) NOT NULL DEFAULT 1.0 CHECK (weight >= 0 AND weight <= 1),
    min_confidence DECIMAL(3,2) NOT NULL DEFAULT 0.1 CHECK (min_confidence >= 0 AND min_confidence <= 1),
    max_confidence DECIMAL(3,2) NOT NULL DEFAULT 0.95 CHECK (max_confidence >= 0 AND max_confidence <= 1),
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority_order INTEGER NOT NULL DEFAULT 0,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    use_count BIGINT NOT NULL DEFAULT 0,
    success_count BIGINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    performance_score DECIMAL(5,4) NOT NULL DEFAULT 0.0 CHECK (performance_score >= 0 AND performance_score <= 1),
    CONSTRAINT chk_confidence_bounds CHECK (min_confidence <= max_confidence)
);

-- Create indexes for performance
CREATE INDEX idx_categorization_rules_category ON categorization_rules(category);
CREATE INDEX idx_categorization_rules_type ON categorization_rules(type);
CREATE INDEX idx_categorization_rules_is_active ON categorization_rules(is_active);
CREATE INDEX idx_categorization_rules_priority_order ON categorization_rules(priority_order);
CREATE INDEX idx_categorization_rules_performance ON categorization_rules(performance_score DESC);
CREATE INDEX idx_categorization_rules_created_by ON categorization_rules(created_by);

-- Full-text search index on keywords
CREATE INDEX idx_categorization_rules_keywords_gin ON categorization_rules USING GIN(keywords);

-- Rule groups table
CREATE TABLE categorization_rule_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority_order INTEGER NOT NULL DEFAULT 0,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for rule groups
CREATE INDEX idx_categorization_rule_groups_category ON categorization_rule_groups(category);
CREATE INDEX idx_categorization_rule_groups_is_active ON categorization_rule_groups(is_active);

-- Rule performance tracking table
CREATE TABLE categorization_rule_performance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES categorization_rules(id) ON DELETE CASCADE,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    total_applications BIGINT NOT NULL DEFAULT 0,
    successful_matches BIGINT NOT NULL DEFAULT 0,
    accuracy_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0 CHECK (accuracy_rate >= 0 AND accuracy_rate <= 1),
    average_confidence DECIMAL(5,4) NOT NULL DEFAULT 0.0 CHECK (average_confidence >= 0 AND average_confidence <= 1),
    false_positives BIGINT NOT NULL DEFAULT 0,
    false_negatives BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(rule_id, period_start, period_end)
);

-- Create indexes for performance tracking
CREATE INDEX idx_categorization_rule_performance_rule_id ON categorization_rule_performance(rule_id);
CREATE INDEX idx_categorization_rule_performance_period ON categorization_rule_performance(period_start, period_end);
CREATE INDEX idx_categorization_rule_performance_accuracy ON categorization_rule_performance(accuracy_rate DESC);

-- Rule templates table
CREATE TABLE categorization_rule_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('keyword', 'pattern', 'semantic', 'entity', 'composite')),
    template JSONB NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    use_count BIGINT NOT NULL DEFAULT 0
);

-- Create indexes for templates
CREATE INDEX idx_categorization_rule_templates_category ON categorization_rule_templates(category);
CREATE INDEX idx_categorization_rule_templates_type ON categorization_rule_templates(type);
CREATE INDEX idx_categorization_rule_templates_is_public ON categorization_rule_templates(is_public);
CREATE INDEX idx_categorization_rule_templates_created_by ON categorization_rule_templates(created_by);

-- Insert some default categorization rules
INSERT INTO categorization_rules (
    name, description, type, category, priority, keywords, base_confidence,
    min_confidence, max_confidence, is_active, priority_order, created_by, updated_by
) VALUES
(
    'Payment Dispute Keywords',
    'Detects payment-related disputes using common keywords',
    'keyword',
    'payment',
    'high',
    ARRAY['payment', 'paid', 'unpaid', 'refund', 'charge', 'fee', 'billing', 'invoice', 'money', 'currency'],
    0.8, 0.3, 0.95, true, 1, 'system', 'system'
),
(
    'Contract Breach Detection',
    'Identifies contract breach disputes',
    'keyword',
    'contract_breach',
    'high',
    ARRAY['breach', 'violation', 'non-compliance', 'contract', 'agreement', 'terms', 'clause'],
    0.9, 0.4, 0.95, true, 2, 'system', 'system'
),
(
    'Fraud Pattern Detection',
    'Detects potential fraud disputes',
    'keyword',
    'fraud',
    'urgent',
    ARRAY['fraud', 'scam', 'fake', 'forgery', 'unauthorized', 'theft', 'stolen', 'suspicious'],
    0.9, 0.5, 0.95, true, 3, 'system', 'system'
),
(
    'Milestone Disputes',
    'Identifies milestone and delivery-related disputes',
    'keyword',
    'milestone',
    'high',
    ARRAY['milestone', 'deliverable', 'delivery', 'completion', 'progress', 'stage', 'deadline', 'overdue'],
    0.8, 0.3, 0.9, true, 4, 'system', 'system'
),
(
    'Technical Issues',
    'Detects technical problem disputes',
    'keyword',
    'technical',
    'normal',
    ARRAY['error', 'bug', 'system', 'technical', 'glitch', 'failure', 'crash', 'timeout', 'connection'],
    0.7, 0.2, 0.85, true, 5, 'system', 'system'
);

-- Insert some semantic patterns as rules
INSERT INTO categorization_rules (
    name, description, type, category, priority, semantic_keys, base_confidence,
    min_confidence, max_confidence, is_active, priority_order, created_by, updated_by
) VALUES
(
    'Payment Semantic Analysis',
    'Semantic analysis for payment disputes',
    'semantic',
    'payment',
    'high',
    ARRAY['payment_dispute'],
    0.6, 0.2, 0.8, true, 6, 'system', 'system'
),
(
    'Contract Semantic Analysis',
    'Semantic analysis for contract disputes',
    'semantic',
    'contract_breach',
    'high',
    ARRAY['contract_dispute'],
    0.7, 0.3, 0.85, true, 7, 'system', 'system'
);

-- Insert entity-based rules
INSERT INTO categorization_rules (
    name, description, type, category, priority, entities, base_confidence,
    min_confidence, max_confidence, is_active, priority_order, created_by, updated_by
) VALUES
(
    'Currency Entity Detection',
    'Detects disputes containing currency amounts',
    'entity',
    'payment',
    'normal',
    ARRAY['currency'],
    0.6, 0.2, 0.8, true, 8, 'system', 'system'
),
(
    'Contract Reference Detection',
    'Detects disputes with contract references',
    'entity',
    'contract_breach',
    'high',
    ARRAY['contract_ref'],
    0.7, 0.3, 0.9, true, 9, 'system', 'system'
);

-- Create trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_categorization_rules_updated_at BEFORE UPDATE ON categorization_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categorization_rule_groups_updated_at BEFORE UPDATE ON categorization_rule_groups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categorization_rule_templates_updated_at BEFORE UPDATE ON categorization_rule_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
