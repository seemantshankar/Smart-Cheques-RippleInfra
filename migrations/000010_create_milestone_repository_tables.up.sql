-- Create milestone dependencies table
CREATE TABLE IF NOT EXISTS milestone_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    milestone_id UUID NOT NULL,
    depends_on_id UUID NOT NULL,
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'prerequisite',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_milestone_dependencies_milestone FOREIGN KEY (milestone_id)
        REFERENCES contract_milestones(id) ON DELETE CASCADE,
    CONSTRAINT fk_milestone_dependencies_depends_on FOREIGN KEY (depends_on_id)
        REFERENCES contract_milestones(id) ON DELETE CASCADE,
    CONSTRAINT chk_milestone_dependencies_type CHECK (dependency_type IN ('prerequisite', 'parallel', 'conditional')),
    CONSTRAINT uq_milestone_dependencies UNIQUE (milestone_id, depends_on_id)
);

-- Create milestone progress history table
CREATE TABLE IF NOT EXISTS milestone_progress_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    milestone_id UUID NOT NULL,
    percentage_complete DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    notes TEXT,
    recorded_by VARCHAR(255),
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_milestone_progress_milestone FOREIGN KEY (milestone_id)
        REFERENCES contract_milestones(id) ON DELETE CASCADE,
    CONSTRAINT chk_milestone_progress_percentage CHECK (percentage_complete >= 0 AND percentage_complete <= 100),
    CONSTRAINT chk_milestone_progress_status CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'on_hold'))
);

-- Create milestone templates table
CREATE TABLE IF NOT EXISTS milestone_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_category VARCHAR(100) DEFAULT 'delivery',
    default_priority INTEGER DEFAULT 3,
    variables TEXT[] DEFAULT '{}',
    version VARCHAR(20) DEFAULT '1.0',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT chk_milestone_template_priority CHECK (default_priority >= 1 AND default_priority <= 10),
    CONSTRAINT chk_milestone_template_category CHECK (default_category IN ('delivery', 'payment', 'approval', 'compliance', 'review', 'milestone'))
);

-- Create milestone template versions table
CREATE TABLE IF NOT EXISTS milestone_template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,
    version VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_category VARCHAR(100) DEFAULT 'delivery',
    default_priority INTEGER DEFAULT 3,
    variables TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255),

    CONSTRAINT fk_milestone_template_versions_template FOREIGN KEY (template_id)
        REFERENCES milestone_templates(id) ON DELETE CASCADE,
    CONSTRAINT uq_milestone_template_versions UNIQUE (template_id, version),
    CONSTRAINT chk_milestone_template_version_priority CHECK (default_priority >= 1 AND default_priority <= 10)
);

-- Create milestone template shares table
CREATE TABLE IF NOT EXISTS milestone_template_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL,
    shared_with VARCHAR(255) NOT NULL,
    shared_by VARCHAR(255) NOT NULL,
    permissions TEXT[] DEFAULT '{"read"}',
    shared_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_milestone_template_shares_template FOREIGN KEY (template_id)
        REFERENCES milestone_templates(id) ON DELETE CASCADE,
    CONSTRAINT uq_milestone_template_shares UNIQUE (template_id, shared_with)
);

-- Add indexes for milestone dependencies
CREATE INDEX IF NOT EXISTS idx_milestone_dependencies_milestone_id ON milestone_dependencies(milestone_id);
CREATE INDEX IF NOT EXISTS idx_milestone_dependencies_depends_on_id ON milestone_dependencies(depends_on_id);
CREATE INDEX IF NOT EXISTS idx_milestone_dependencies_type ON milestone_dependencies(dependency_type);

-- Add indexes for milestone progress history
CREATE INDEX IF NOT EXISTS idx_milestone_progress_milestone_id ON milestone_progress_history(milestone_id);
CREATE INDEX IF NOT EXISTS idx_milestone_progress_recorded_at ON milestone_progress_history(recorded_at);
CREATE INDEX IF NOT EXISTS idx_milestone_progress_status ON milestone_progress_history(status);
CREATE INDEX IF NOT EXISTS idx_milestone_progress_recorded_by ON milestone_progress_history(recorded_by);

-- Add indexes for milestone templates
CREATE INDEX IF NOT EXISTS idx_milestone_templates_name ON milestone_templates(name);
CREATE INDEX IF NOT EXISTS idx_milestone_templates_category ON milestone_templates(default_category);
CREATE INDEX IF NOT EXISTS idx_milestone_templates_active ON milestone_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_milestone_templates_created_at ON milestone_templates(created_at);

-- Add indexes for milestone template versions
CREATE INDEX IF NOT EXISTS idx_milestone_template_versions_template_id ON milestone_template_versions(template_id);
CREATE INDEX IF NOT EXISTS idx_milestone_template_versions_version ON milestone_template_versions(version);
CREATE INDEX IF NOT EXISTS idx_milestone_template_versions_created_at ON milestone_template_versions(created_at);

-- Add indexes for milestone template shares
CREATE INDEX IF NOT EXISTS idx_milestone_template_shares_template_id ON milestone_template_shares(template_id);
CREATE INDEX IF NOT EXISTS idx_milestone_template_shares_shared_with ON milestone_template_shares(shared_with);
CREATE INDEX IF NOT EXISTS idx_milestone_template_shares_shared_by ON milestone_template_shares(shared_by);
CREATE INDEX IF NOT EXISTS idx_milestone_template_shares_expires_at ON milestone_template_shares(expires_at);

-- Enhance existing contract_milestones table with additional indexes for new functionality
CREATE INDEX IF NOT EXISTS idx_contract_milestones_contract_id_sequence ON contract_milestones(contract_id, sequence_number);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_category ON contract_milestones(category);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_priority ON contract_milestones(priority);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_critical_path ON contract_milestones(critical_path);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_risk_level ON contract_milestones(risk_level);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_percentage_complete ON contract_milestones(percentage_complete);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_estimated_end_date ON contract_milestones(estimated_end_date);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_actual_end_date ON contract_milestones(actual_end_date);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_criticality_score ON contract_milestones(criticality_score);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_overdue ON contract_milestones(estimated_end_date, percentage_complete)
    WHERE percentage_complete < 100;

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_contract_milestones_contract_status ON contract_milestones(contract_id, percentage_complete);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_risk_priority ON contract_milestones(risk_level, priority);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_timeline ON contract_milestones(estimated_start_date, estimated_end_date);

-- Add GIN indexes for array fields
CREATE INDEX IF NOT EXISTS idx_contract_milestones_dependencies_gin ON contract_milestones USING gin(dependencies);
CREATE INDEX IF NOT EXISTS idx_contract_milestones_contingency_plans_gin ON contract_milestones USING gin(contingency_plans);
CREATE INDEX IF NOT EXISTS idx_milestone_templates_variables_gin ON milestone_templates USING gin(variables);

-- Add full-text search indexes
CREATE INDEX IF NOT EXISTS idx_contract_milestones_text_search ON contract_milestones USING gin(
    to_tsvector('english', coalesce(trigger_conditions, '') || ' ' || coalesce(verification_criteria, ''))
);

CREATE INDEX IF NOT EXISTS idx_milestone_templates_text_search ON milestone_templates USING gin(
    to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
);

-- Create triggers for automatic updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers to relevant tables
CREATE TRIGGER update_milestone_templates_updated_at
    BEFORE UPDATE ON milestone_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create views for common queries

-- View for milestone completion statistics
CREATE OR REPLACE VIEW milestone_completion_stats AS
SELECT
    contract_id,
    COUNT(*) as total_milestones,
    COUNT(CASE WHEN percentage_complete = 100 THEN 1 END) as completed_milestones,
    COUNT(CASE WHEN percentage_complete > 0 AND percentage_complete < 100 THEN 1 END) as in_progress_milestones,
    COUNT(CASE WHEN percentage_complete = 0 THEN 1 END) as pending_milestones,
    COUNT(CASE WHEN estimated_end_date < NOW() AND percentage_complete < 100 THEN 1 END) as overdue_milestones,
    AVG(percentage_complete) as average_completion,
    CASE
        WHEN COUNT(*) > 0 THEN
            COUNT(CASE WHEN percentage_complete = 100 THEN 1 END) * 100.0 / COUNT(*)
        ELSE 0
    END as completion_rate
FROM contract_milestones
GROUP BY contract_id;

-- View for critical path milestones
CREATE OR REPLACE VIEW critical_path_milestones AS
SELECT
    cm.*,
    c.status as contract_status,
    CASE
        WHEN cm.estimated_end_date < NOW() AND cm.percentage_complete < 100 THEN 'overdue'
        WHEN cm.percentage_complete = 100 THEN 'completed'
        WHEN cm.percentage_complete > 0 THEN 'in_progress'
        ELSE 'pending'
    END as milestone_status
FROM contract_milestones cm
JOIN contracts c ON cm.contract_id = c.id
WHERE cm.critical_path = true;

-- View for milestone dependencies with details
CREATE OR REPLACE VIEW milestone_dependency_details AS
SELECT
    md.id,
    md.milestone_id,
    md.depends_on_id,
    md.dependency_type,
    cm1.trigger_conditions as milestone_description,
    cm2.trigger_conditions as dependency_description,
    cm1.contract_id,
    cm1.percentage_complete as milestone_progress,
    cm2.percentage_complete as dependency_progress,
    CASE
        WHEN cm2.percentage_complete = 100 THEN 'satisfied'
        WHEN cm2.percentage_complete > 0 THEN 'in_progress'
        ELSE 'pending'
    END as dependency_status
FROM milestone_dependencies md
JOIN contract_milestones cm1 ON md.milestone_id = cm1.id
JOIN contract_milestones cm2 ON md.depends_on_id = cm2.id;

-- View for overdue milestones with impact analysis
CREATE OR REPLACE VIEW overdue_milestones_report AS
SELECT
    cm.id,
    cm.contract_id,
    cm.trigger_conditions as description,
    cm.estimated_end_date,
    cm.actual_end_date,
    cm.percentage_complete,
    cm.risk_level,
    cm.critical_path,
    cm.criticality_score,
    EXTRACT(EPOCH FROM (COALESCE(cm.actual_end_date, NOW()) - cm.estimated_end_date)) / 86400.0 as days_overdue,
    CASE
        WHEN cm.critical_path AND cm.risk_level = 'high' THEN 'critical'
        WHEN cm.critical_path OR cm.risk_level = 'high' THEN 'high'
        WHEN cm.risk_level = 'medium' THEN 'medium'
        ELSE 'low'
    END as impact_level,
    c.status as contract_status
FROM contract_milestones cm
JOIN contracts c ON cm.contract_id = c.id
WHERE cm.estimated_end_date < NOW()
  AND cm.percentage_complete < 100
ORDER BY cm.criticality_score DESC, cm.estimated_end_date ASC;

-- Add comments for documentation
COMMENT ON TABLE milestone_dependencies IS 'Stores dependencies between milestones for workflow orchestration';
COMMENT ON TABLE milestone_progress_history IS 'Tracks historical progress updates for milestones';
COMMENT ON TABLE milestone_templates IS 'Stores reusable milestone templates for contract creation';
COMMENT ON TABLE milestone_template_versions IS 'Stores versions of milestone templates for change tracking';
COMMENT ON TABLE milestone_template_shares IS 'Manages sharing permissions for milestone templates';

COMMENT ON VIEW milestone_completion_stats IS 'Provides aggregated completion statistics per contract';
COMMENT ON VIEW critical_path_milestones IS 'Shows all critical path milestones with current status';
COMMENT ON VIEW milestone_dependency_details IS 'Detailed view of milestone dependencies with progress info';
COMMENT ON VIEW overdue_milestones_report IS 'Comprehensive report of overdue milestones with impact analysis';
