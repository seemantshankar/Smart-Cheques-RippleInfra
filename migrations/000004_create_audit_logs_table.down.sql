-- Drop indexes
DROP INDEX IF EXISTS idx_audit_logs_action_resource;
DROP INDEX IF EXISTS idx_audit_logs_enterprise_created;
DROP INDEX IF EXISTS idx_audit_logs_user_created;
DROP INDEX IF EXISTS idx_audit_logs_success;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_enterprise_id;
DROP INDEX IF EXISTS idx_audit_logs_user_id;

-- Drop table
DROP TABLE IF EXISTS audit_logs;