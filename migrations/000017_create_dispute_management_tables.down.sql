-- Drop dispute management tables
-- Migration: 000015_create_dispute_management_tables.down.sql

-- Drop views
DROP VIEW IF EXISTS dispute_statistics;
DROP VIEW IF EXISTS dispute_participant_summary;
DROP VIEW IF EXISTS dispute_resolution_summary;
DROP VIEW IF EXISTS urgent_disputes;
DROP VIEW IF EXISTS resolved_disputes;
DROP VIEW IF EXISTS active_disputes;

-- Drop triggers
DROP TRIGGER IF EXISTS update_dispute_notifications_updated_at ON dispute_notifications;
DROP TRIGGER IF EXISTS update_dispute_comments_updated_at ON dispute_comments;
DROP TRIGGER IF EXISTS update_dispute_resolutions_updated_at ON dispute_resolutions;
DROP TRIGGER IF EXISTS update_dispute_evidence_updated_at ON dispute_evidence;
DROP TRIGGER IF EXISTS update_disputes_updated_at ON disputes;

-- Drop function
DROP FUNCTION IF EXISTS update_dispute_management_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_dispute_notifications_sent_at;
DROP INDEX IF EXISTS idx_dispute_notifications_status;
DROP INDEX IF EXISTS idx_dispute_notifications_channel;
DROP INDEX IF EXISTS idx_dispute_notifications_type;
DROP INDEX IF EXISTS idx_dispute_notifications_recipient;
DROP INDEX IF EXISTS idx_dispute_notifications_dispute_id;

DROP INDEX IF EXISTS idx_dispute_audit_logs_created_at;
DROP INDEX IF EXISTS idx_dispute_audit_logs_user_id;
DROP INDEX IF EXISTS idx_dispute_audit_logs_action;
DROP INDEX IF EXISTS idx_dispute_audit_logs_dispute_id;

DROP INDEX IF EXISTS idx_dispute_comments_is_internal;
DROP INDEX IF EXISTS idx_dispute_comments_author_type;
DROP INDEX IF EXISTS idx_dispute_comments_author_id;
DROP INDEX IF EXISTS idx_dispute_comments_dispute_id;

DROP INDEX IF EXISTS idx_dispute_resolutions_executed_at;
DROP INDEX IF EXISTS idx_dispute_resolutions_respondent_accepted;
DROP INDEX IF EXISTS idx_dispute_resolutions_initiator_accepted;
DROP INDEX IF EXISTS idx_dispute_resolutions_is_executed;
DROP INDEX IF EXISTS idx_dispute_resolutions_method;
DROP INDEX IF EXISTS idx_dispute_resolutions_dispute_id;

DROP INDEX IF EXISTS idx_dispute_evidence_is_public;
DROP INDEX IF EXISTS idx_dispute_evidence_file_type;
DROP INDEX IF EXISTS idx_dispute_evidence_uploaded_by;
DROP INDEX IF EXISTS idx_dispute_evidence_dispute_id;

DROP INDEX IF EXISTS idx_disputes_updated_by;
DROP INDEX IF EXISTS idx_disputes_created_by;
DROP INDEX IF EXISTS idx_disputes_last_activity_at;
DROP INDEX IF EXISTS idx_disputes_initiated_at;
DROP INDEX IF EXISTS idx_disputes_transaction_id;
DROP INDEX IF EXISTS idx_disputes_contract_id;
DROP INDEX IF EXISTS idx_disputes_milestone_id;
DROP INDEX IF EXISTS idx_disputes_smart_cheque_id;
DROP INDEX IF EXISTS idx_disputes_priority;
DROP INDEX IF EXISTS idx_disputes_category;
DROP INDEX IF EXISTS idx_disputes_status;
DROP INDEX IF EXISTS idx_disputes_respondent_id;
DROP INDEX IF EXISTS idx_disputes_initiator_id;

-- Drop GIN indexes
DROP INDEX IF EXISTS idx_dispute_notifications_metadata_gin;
DROP INDEX IF EXISTS idx_dispute_audit_logs_new_value_gin;
DROP INDEX IF EXISTS idx_dispute_audit_logs_old_value_gin;
DROP INDEX IF EXISTS idx_dispute_resolutions_outcome_gin;
DROP INDEX IF EXISTS idx_disputes_metadata_gin;
DROP INDEX IF EXISTS idx_disputes_tags_gin;

-- Drop full-text search indexes
DROP INDEX IF EXISTS idx_dispute_notifications_message_fts;
DROP INDEX IF EXISTS idx_dispute_comments_content_fts;
DROP INDEX IF EXISTS idx_disputes_title_desc_fts;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS dispute_notifications;
DROP TABLE IF EXISTS dispute_audit_logs;
DROP TABLE IF EXISTS dispute_comments;
DROP TABLE IF EXISTS dispute_resolutions;
DROP TABLE IF EXISTS dispute_evidence;
DROP TABLE IF EXISTS disputes;
