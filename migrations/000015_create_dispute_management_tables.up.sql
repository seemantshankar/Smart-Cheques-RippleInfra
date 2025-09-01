-- Create dispute management tables
-- Migration: 000015_create_dispute_management_tables.up.sql

-- Create disputes table
CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(30) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(30) NOT NULL DEFAULT 'initiated',

    -- Related entities
    smart_cheque_id VARCHAR(255),
    milestone_id VARCHAR(255),
    contract_id VARCHAR(255),
    transaction_id VARCHAR(255),

    -- Parties involved
    initiator_id VARCHAR(255) NOT NULL,
    initiator_type VARCHAR(20) NOT NULL,
    respondent_id VARCHAR(255) NOT NULL,
    respondent_type VARCHAR(20) NOT NULL,

    -- Financial impact
    disputed_amount DECIMAL(20,8),
    currency VARCHAR(10),

    -- Timestamps and tracking
    initiated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT disputes_category_check CHECK (category IN ('payment', 'milestone', 'contract_breach', 'fraud', 'technical', 'other')),
    CONSTRAINT disputes_priority_check CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    CONSTRAINT disputes_status_check CHECK (status IN ('initiated', 'under_review', 'escalated', 'resolved', 'closed', 'cancelled')),
    CONSTRAINT disputes_initiator_type_check CHECK (initiator_type IN ('enterprise', 'user')),
    CONSTRAINT disputes_respondent_type_check CHECK (respondent_type IN ('enterprise', 'user')),
    CONSTRAINT disputes_disputed_amount_check CHECK (disputed_amount >= 0)
);

-- Create dispute_evidence table
CREATE TABLE dispute_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_path TEXT NOT NULL,
    description TEXT,
    uploaded_by VARCHAR(255) NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dispute_evidence_file_size_check CHECK (file_size >= 0 AND file_size <= 104857600), -- Max 100MB
    CONSTRAINT dispute_evidence_file_type_check CHECK (file_type IN ('pdf', 'doc', 'docx', 'txt', 'png', 'jpg', 'jpeg', 'gif', 'csv', 'xlsx', 'other'))
);

-- Create dispute_resolutions table
CREATE TABLE dispute_resolutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    method VARCHAR(30) NOT NULL,
    resolution_details TEXT NOT NULL,
    outcome_amount DECIMAL(20,8),
    outcome_description TEXT,

    -- Resolution parties
    mediator_id VARCHAR(255),
    arbitrator_id VARCHAR(255),
    court_case_number VARCHAR(255),

    -- Agreement tracking
    initiator_accepted BOOLEAN NOT NULL DEFAULT false,
    respondent_accepted BOOLEAN NOT NULL DEFAULT false,
    acceptance_deadline TIMESTAMP WITH TIME ZONE,

    -- Execution details
    is_executed BOOLEAN NOT NULL DEFAULT false,
    executed_at TIMESTAMP WITH TIME ZONE,
    executed_by VARCHAR(255),

    -- Audit fields
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dispute_resolutions_method_check CHECK (method IN ('mutual_agreement', 'mediation', 'arbitration', 'court', 'administrative')),
    CONSTRAINT dispute_resolutions_outcome_amount_check CHECK (outcome_amount >= 0)
);

-- Create dispute_comments table
CREATE TABLE dispute_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    author_id VARCHAR(255) NOT NULL,
    author_type VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dispute_comments_author_type_check CHECK (author_type IN ('enterprise', 'user', 'mediator', 'admin')),
    CONSTRAINT dispute_comments_content_length_check CHECK (char_length(content) BETWEEN 1 AND 1000)
);

-- Create dispute_audit_logs table
CREATE TABLE dispute_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    details TEXT,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dispute_audit_logs_user_type_check CHECK (user_type IN ('enterprise', 'user', 'mediator', 'admin'))
);

-- Create dispute_notifications table
CREATE TABLE dispute_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES disputes(id) ON DELETE CASCADE,
    recipient VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMP WITH TIME ZONE,
    error_msg TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT dispute_notifications_type_check CHECK (type IN ('email', 'sms', 'webhook', 'in_app')),
    CONSTRAINT dispute_notifications_channel_check CHECK (channel IN ('email', 'sms', 'webhook', 'in_app')),
    CONSTRAINT dispute_notifications_status_check CHECK (status IN ('pending', 'sent', 'failed'))
);

-- Create indexes for performance
CREATE INDEX idx_disputes_initiator_id ON disputes(initiator_id);
CREATE INDEX idx_disputes_respondent_id ON disputes(respondent_id);
CREATE INDEX idx_disputes_status ON disputes(status);
CREATE INDEX idx_disputes_category ON disputes(category);
CREATE INDEX idx_disputes_priority ON disputes(priority);
CREATE INDEX idx_disputes_smart_cheque_id ON disputes(smart_cheque_id) WHERE smart_cheque_id IS NOT NULL;
CREATE INDEX idx_disputes_milestone_id ON disputes(milestone_id) WHERE milestone_id IS NOT NULL;
CREATE INDEX idx_disputes_contract_id ON disputes(contract_id) WHERE contract_id IS NOT NULL;
CREATE INDEX idx_disputes_transaction_id ON disputes(transaction_id) WHERE transaction_id IS NOT NULL;
CREATE INDEX idx_disputes_initiated_at ON disputes(initiated_at);
CREATE INDEX idx_disputes_last_activity_at ON disputes(last_activity_at);
CREATE INDEX idx_disputes_created_by ON disputes(created_by);
CREATE INDEX idx_disputes_updated_by ON disputes(updated_by);

CREATE INDEX idx_dispute_evidence_dispute_id ON dispute_evidence(dispute_id);
CREATE INDEX idx_dispute_evidence_uploaded_by ON dispute_evidence(uploaded_by);
CREATE INDEX idx_dispute_evidence_file_type ON dispute_evidence(file_type);
CREATE INDEX idx_dispute_evidence_is_public ON dispute_evidence(is_public);

CREATE INDEX idx_dispute_resolutions_dispute_id ON dispute_resolutions(dispute_id);
CREATE INDEX idx_dispute_resolutions_method ON dispute_resolutions(method);
CREATE INDEX idx_dispute_resolutions_is_executed ON dispute_resolutions(is_executed);
CREATE INDEX idx_dispute_resolutions_initiator_accepted ON dispute_resolutions(initiator_accepted);
CREATE INDEX idx_dispute_resolutions_respondent_accepted ON dispute_resolutions(respondent_accepted);
CREATE INDEX idx_dispute_resolutions_executed_at ON dispute_resolutions(executed_at) WHERE executed_at IS NOT NULL;

CREATE INDEX idx_dispute_comments_dispute_id ON dispute_comments(dispute_id);
CREATE INDEX idx_dispute_comments_author_id ON dispute_comments(author_id);
CREATE INDEX idx_dispute_comments_author_type ON dispute_comments(author_type);
CREATE INDEX idx_dispute_comments_is_internal ON dispute_comments(is_internal);

CREATE INDEX idx_dispute_audit_logs_dispute_id ON dispute_audit_logs(dispute_id);
CREATE INDEX idx_dispute_audit_logs_action ON dispute_audit_logs(action);
CREATE INDEX idx_dispute_audit_logs_user_id ON dispute_audit_logs(user_id);
CREATE INDEX idx_dispute_audit_logs_created_at ON dispute_audit_logs(created_at);

CREATE INDEX idx_dispute_notifications_dispute_id ON dispute_notifications(dispute_id);
CREATE INDEX idx_dispute_notifications_recipient ON dispute_notifications(recipient);
CREATE INDEX idx_dispute_notifications_type ON dispute_notifications(type);
CREATE INDEX idx_dispute_notifications_channel ON dispute_notifications(channel);
CREATE INDEX idx_dispute_notifications_status ON dispute_notifications(status);
CREATE INDEX idx_dispute_notifications_sent_at ON dispute_notifications(sent_at) WHERE sent_at IS NOT NULL;

-- Create GIN indexes for JSON and array fields
CREATE INDEX idx_disputes_tags_gin ON disputes USING GIN (tags);
CREATE INDEX idx_disputes_metadata_gin ON disputes USING GIN (metadata);
CREATE INDEX idx_dispute_resolutions_outcome_gin ON dispute_resolutions USING GIN ((jsonb_build_object('outcome_amount', outcome_amount, 'outcome_description', outcome_description)));
CREATE INDEX idx_dispute_audit_logs_old_value_gin ON dispute_audit_logs USING GIN (old_value);
CREATE INDEX idx_dispute_audit_logs_new_value_gin ON dispute_audit_logs USING GIN (new_value);
CREATE INDEX idx_dispute_notifications_metadata_gin ON dispute_notifications USING GIN (metadata);

-- Create full-text search indexes
CREATE INDEX idx_disputes_title_desc_fts ON disputes USING GIN (to_tsvector('english', title || ' ' || description));
CREATE INDEX idx_dispute_comments_content_fts ON dispute_comments USING GIN (to_tsvector('english', content));
CREATE INDEX idx_dispute_notifications_message_fts ON dispute_notifications USING GIN (to_tsvector('english', message));

-- Create triggers to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_dispute_management_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_disputes_updated_at
    BEFORE UPDATE ON disputes
    FOR EACH ROW
    EXECUTE FUNCTION update_dispute_management_updated_at();

CREATE TRIGGER update_dispute_evidence_updated_at
    BEFORE UPDATE ON dispute_evidence
    FOR EACH ROW
    EXECUTE FUNCTION update_dispute_management_updated_at();

CREATE TRIGGER update_dispute_resolutions_updated_at
    BEFORE UPDATE ON dispute_resolutions
    FOR EACH ROW
    EXECUTE FUNCTION update_dispute_management_updated_at();

CREATE TRIGGER update_dispute_comments_updated_at
    BEFORE UPDATE ON dispute_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_dispute_management_updated_at();

CREATE TRIGGER update_dispute_notifications_updated_at
    BEFORE UPDATE ON dispute_notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_dispute_management_updated_at();

-- Create views for common queries
CREATE VIEW active_disputes AS
SELECT * FROM disputes
WHERE status IN ('initiated', 'under_review', 'escalated');

CREATE VIEW resolved_disputes AS
SELECT * FROM disputes
WHERE status IN ('resolved', 'closed')
  AND resolved_at IS NOT NULL;

CREATE VIEW urgent_disputes AS
SELECT * FROM disputes
WHERE priority IN ('high', 'urgent')
  AND status NOT IN ('resolved', 'closed', 'cancelled');

CREATE VIEW dispute_resolution_summary AS
SELECT
    d.id,
    d.title,
    d.category,
    d.priority,
    d.status,
    d.disputed_amount,
    d.currency,
    dr.method as resolution_method,
    dr.is_executed,
    dr.executed_at,
    EXTRACT(EPOCH FROM (d.resolved_at - d.initiated_at))/86400 as resolution_days
FROM disputes d
LEFT JOIN dispute_resolutions dr ON d.id = dr.dispute_id
WHERE d.status IN ('resolved', 'closed');

CREATE VIEW dispute_participant_summary AS
SELECT
    participant_id,
    participant_type,
    COUNT(*) as total_disputes,
    COUNT(*) FILTER (WHERE status = 'resolved') as resolved_disputes,
    COUNT(*) FILTER (WHERE status = 'closed') as closed_disputes,
    COUNT(*) FILTER (WHERE priority IN ('high', 'urgent')) as urgent_disputes,
    AVG(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as avg_resolution_days
FROM (
    SELECT initiator_id as participant_id, initiator_type as participant_type, status, resolved_at, initiated_at, priority FROM disputes
    UNION ALL
    SELECT respondent_id as participant_id, respondent_type as participant_type, status, resolved_at, initiated_at, priority FROM disputes
) participants
GROUP BY participant_id, participant_type;

CREATE VIEW dispute_statistics AS
SELECT
    COUNT(*) as total_disputes,
    COUNT(*) FILTER (WHERE status = 'initiated') as initiated_disputes,
    COUNT(*) FILTER (WHERE status = 'under_review') as under_review_disputes,
    COUNT(*) FILTER (WHERE status = 'escalated') as escalated_disputes,
    COUNT(*) FILTER (WHERE status = 'resolved') as resolved_disputes,
    COUNT(*) FILTER (WHERE status = 'closed') as closed_disputes,
    COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled_disputes,
    AVG(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as avg_resolution_time_days,
    MIN(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as min_resolution_time_days,
    MAX(EXTRACT(EPOCH FROM (resolved_at - initiated_at))/86400) FILTER (WHERE resolved_at IS NOT NULL) as max_resolution_time_days
FROM disputes;

-- Insert default dispute categories and priorities (for reference)
-- This is just for documentation; the actual enum constraints are handled by CHECK constraints above
COMMENT ON TABLE disputes IS 'Main disputes table containing all dispute information and metadata';
COMMENT ON TABLE dispute_evidence IS 'Evidence files and documents attached to disputes';
COMMENT ON TABLE dispute_resolutions IS 'Resolution details and outcomes for disputes';
COMMENT ON TABLE dispute_comments IS 'Comments and notes on disputes from various participants';
COMMENT ON TABLE dispute_audit_logs IS 'Audit trail for all dispute-related activities';
COMMENT ON TABLE dispute_notifications IS 'Notification records for dispute events';
